"""
Chat API Router for AI Service.

Provides endpoints for interacting with the AI assistant:
- POST /api/v1/ai/chat - Send message to agent
- POST /api/v1/ai/chat/stream - SSE streaming response
- POST /api/v1/ai/confirm/{action_id} - Confirm pending action
- GET /api/v1/ai/conversations - List user's conversations
- GET /api/v1/ai/conversations/{session_id} - Get conversation messages
- DELETE /api/v1/ai/conversations/{session_id} - Clear conversation
"""

import asyncio
import json
import logging
import time
from typing import Any, Optional
from uuid import UUID, uuid4

from fastapi import APIRouter, Depends, HTTPException, Request, status
from fastapi.responses import StreamingResponse
from pydantic import BaseModel, Field
from sqlalchemy.ext.asyncio import AsyncSession

from app.dependencies import (
    get_request_context,
    require_permission,
    RequestContext,
)
from app.database import get_db_session
from app.models.audit_log import ActionStatus, ActionType, Severity
from app.repository.conversation_repo import ConversationRepository
from app.repository.audit_repo import AuditLogRepository
from app.agents.simple_agent import create_simple_agent


logger = logging.getLogger(__name__)


# Request/Response Models
class ChatRequest(BaseModel):
    """Request body for chat endpoint."""

    message: str = Field(..., min_length=1, max_length=10000)
    session_id: Optional[str] = Field(
        None,
        description="Session ID for continuing a conversation"
    )


class ToolCall(BaseModel):
    """Tool call information in response."""

    tool_name: str
    arguments: dict = Field(default_factory=dict)
    result: Optional[Any] = None


class ConfirmationRequired(BaseModel):
    """Confirmation requirement in response."""

    action_id: str
    action_type: str
    description: str
    details: dict = Field(default_factory=dict)


class ChatResponse(BaseModel):
    """Response body for chat endpoint."""

    response: str
    session_id: str
    message_id: str
    tool_calls: Optional[list[ToolCall]] = None
    confirmation_required: Optional[ConfirmationRequired] = None


class ConfirmRequest(BaseModel):
    """Request body for confirmation endpoint."""

    confirmed: bool


class ConfirmResponse(BaseModel):
    """Response body for confirmation endpoint."""

    action_id: str
    confirmed: bool
    result: Optional[dict] = None
    message: str


class ConversationSummary(BaseModel):
    """Summary of a conversation session."""

    session_id: str
    last_message_at: str
    message_count: int


class ConversationsListResponse(BaseModel):
    """Response body for listing conversations."""

    conversations: list[ConversationSummary]
    total: int


class MessageResponse(BaseModel):
    """Single message in conversation."""

    id: str
    role: str
    content: str
    tool_calls: Optional[dict] = None
    created_at: str


class ConversationMessagesResponse(BaseModel):
    """Response body for conversation messages."""

    session_id: str
    messages: list[MessageResponse]


class DeleteConversationResponse(BaseModel):
    """Response body for deleting conversation."""

    session_id: str
    deleted_count: int
    message: str


# Create router
router = APIRouter(prefix="/api/v1/ai", tags=["AI Chat"])


@router.post("/chat", response_model=ChatResponse)
async def send_chat_message(
    request: ChatRequest,
    request_ctx: RequestContext = Depends(get_request_context),
    db: AsyncSession = Depends(get_db_session),
) -> ChatResponse:
    """
    Send a message to the AI assistant.

    This endpoint receives a user message and returns the AI's response.
    May include tool calls and confirmation requirements for certain actions.

    Args:
        request: Chat request with message and optional session_id.
        request_ctx: Request context with user and tenant info.
        db: Database session.

    Returns:
        ChatResponse with AI response and optional tool calls/confirmations.
    """
    start_time = time.time()

    # Generate or use provided session ID
    session_id = UUID(request.session_id) if request.session_id else uuid4()
    user_id = UUID(request_ctx.user_id)
    tenant_id = UUID(request_ctx.effective_tenant_id)

    # Initialize repositories
    conv_repo = ConversationRepository(db)
    audit_repo = AuditLogRepository(db)

    try:
        # Store user message
        user_message = await conv_repo.create_user_message(
            tenant_id=tenant_id,
            user_id=user_id,
            session_id=session_id,
            content=request.message,
        )

        # Create audit log for query
        query_log = await audit_repo.create_query_log(
            tenant_id=tenant_id,
            user_id=user_id,
            conversation_id=user_message.id,
            input_params={"message": request.message},
            status=ActionStatus.PENDING,
        )

        # Get conversation context
        context = await conv_repo.get_recent_context(
            session_id=session_id,
            tenant_id=tenant_id,
            limit=10,
        )

        # Process with AI agent (LlamaIndex CopilotAgent)
        agent_result = await _process_with_agent(
            message=request.message,
            context=context,
            request_ctx=request_ctx,
            session_id=session_id,
        )

        response_text = agent_result.get("response", "")
        tool_calls = agent_result.get("tool_calls", [])
        agent_confirmation = agent_result.get("confirmation_required")

        # Check for confirmation requirements
        confirmation_required = None
        if agent_confirmation or _requires_confirmation(response_text, request.message):
            action_id = str(uuid4())
            confirmation_required = ConfirmationRequired(
                action_id=action_id,
                action_type="update_occurrence_status",
                description=agent_confirmation.get("message") if agent_confirmation else "Esta acao requer sua confirmacao",
                details={"original_message": request.message},
            )

            # Create pending action audit log
            await audit_repo.create_tool_execution_log(
                tenant_id=tenant_id,
                user_id=user_id,
                conversation_id=user_message.id,
                tool_name="update_occurrence_status",
                input_params={"message": request.message},
                status=ActionStatus.PENDING,
                severity=Severity.WARN,
            )

        # Store assistant response
        assistant_message = await conv_repo.create_assistant_message(
            tenant_id=tenant_id,
            user_id=user_id,
            session_id=session_id,
            content=response_text,
            tool_calls={"calls": tool_calls} if tool_calls else None,
        )

        # Update audit log with success
        execution_time_ms = int((time.time() - start_time) * 1000)
        await audit_repo.update_status(
            log_id=query_log.id,
            tenant_id=tenant_id,
            status=ActionStatus.SUCCESS,
            output_result={"response": response_text[:1000]},
            execution_time_ms=execution_time_ms,
        )

        return ChatResponse(
            response=response_text,
            session_id=str(session_id),
            message_id=str(assistant_message.id),
            tool_calls=[ToolCall(**tc) for tc in tool_calls] if tool_calls else None,
            confirmation_required=confirmation_required,
        )

    except Exception as e:
        logger.error(f"Chat endpoint error: {e}", exc_info=True)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={
                "error": "internal_server_error",
                "message": "Ocorreu um erro ao processar sua mensagem",
            }
        )


@router.post("/chat/stream")
async def stream_chat_message(
    request: ChatRequest,
    request_ctx: RequestContext = Depends(get_request_context),
    db: AsyncSession = Depends(get_db_session),
) -> StreamingResponse:
    """
    Send a message with SSE streaming response.

    Event types:
    - thinking: AI is processing
    - tool_call: Tool execution notification
    - response_chunk: Partial response text
    - done: Streaming complete

    Args:
        request: Chat request with message and optional session_id.
        request_ctx: Request context with user and tenant info.
        db: Database session.

    Returns:
        StreamingResponse with SSE events.
    """
    session_id = UUID(request.session_id) if request.session_id else uuid4()
    user_id = UUID(request_ctx.user_id)
    tenant_id = UUID(request_ctx.effective_tenant_id)

    async def generate_events():
        """Generate SSE events for streaming response."""
        try:
            # Send thinking event
            yield _format_sse_event("thinking", {
                "status": "processing",
                "message": "Analisando sua mensagem...",
            })

            await asyncio.sleep(0.1)  # Small delay for UX

            # Initialize repositories (in streaming context)
            conv_repo = ConversationRepository(db)

            # Store user message
            user_message = await conv_repo.create_user_message(
                tenant_id=tenant_id,
                user_id=user_id,
                session_id=session_id,
                content=request.message,
            )

            # Get context
            context = await conv_repo.get_recent_context(
                session_id=session_id,
                tenant_id=tenant_id,
                limit=10,
            )

            # Send context retrieval event
            yield _format_sse_event("thinking", {
                "status": "retrieving_context",
                "message": "Consultando historico da conversa...",
            })

            await asyncio.sleep(0.1)

            # Process with agent (LlamaIndex CopilotAgent)
            agent_result = await _process_with_agent(
                message=request.message,
                context=context,
                request_ctx=request_ctx,
                session_id=session_id,
            )

            response_text = agent_result.get("response", "")
            tool_calls = agent_result.get("tool_calls", [])

            # Stream response in chunks
            chunk_size = 50
            for i in range(0, len(response_text), chunk_size):
                chunk = response_text[i:i + chunk_size]
                yield _format_sse_event("response_chunk", {
                    "content": chunk,
                    "index": i // chunk_size,
                })
                await asyncio.sleep(0.02)  # Simulate streaming delay

            # Store assistant response
            assistant_message = await conv_repo.create_assistant_message(
                tenant_id=tenant_id,
                user_id=user_id,
                session_id=session_id,
                content=response_text,
                tool_calls={"calls": tool_calls} if tool_calls else None,
            )

            # Send done event
            yield _format_sse_event("done", {
                "session_id": str(session_id),
                "message_id": str(assistant_message.id),
                "complete_response": response_text,
                "tool_calls": tool_calls,
            })

        except Exception as e:
            logger.error(f"Streaming error: {e}", exc_info=True)
            yield _format_sse_event("error", {
                "message": "Ocorreu um erro ao processar sua mensagem",
                "code": "STREAMING_ERROR",
            })

    return StreamingResponse(
        generate_events(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no",
        },
    )


@router.post("/confirm/{action_id}", response_model=ConfirmResponse)
async def confirm_action(
    action_id: str,
    request: ConfirmRequest,
    request_ctx: RequestContext = Depends(get_request_context),
    db: AsyncSession = Depends(get_db_session),
) -> ConfirmResponse:
    """
    Confirm or cancel a pending action.

    This endpoint handles human-in-the-loop confirmation for critical actions.

    Args:
        action_id: UUID of the pending action.
        request: Confirmation request with confirmed flag.
        request_ctx: Request context with user and tenant info.
        db: Database session.

    Returns:
        ConfirmResponse with action result.
    """
    try:
        action_uuid = UUID(action_id)
    except ValueError:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={"error": "invalid_action_id", "message": "ID da acao invalido"},
        )

    user_id = UUID(request_ctx.user_id)
    tenant_id = UUID(request_ctx.effective_tenant_id)

    audit_repo = AuditLogRepository(db)

    # Get the pending action
    pending_action = await audit_repo.get_by_id(action_uuid, tenant_id)

    if not pending_action:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={"error": "action_not_found", "message": "Acao nao encontrada"},
        )

    if pending_action.status != ActionStatus.PENDING.value:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={
                "error": "action_not_pending",
                "message": "Esta acao ja foi processada",
            },
        )

    # Verify user owns this action
    if str(pending_action.user_id) != str(user_id):
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail={
                "error": "unauthorized",
                "message": "Voce nao tem permissao para confirmar esta acao",
            },
        )

    # Create confirmation audit log
    await audit_repo.create_confirmation_log(
        tenant_id=tenant_id,
        user_id=user_id,
        conversation_id=pending_action.conversation_id,
        tool_name=pending_action.tool_name,
        input_params=pending_action.input_params,
        confirmed=request.confirmed,
    )

    result = None
    if request.confirmed:
        # Execute the action
        try:
            result = await _execute_confirmed_action(
                tool_name=pending_action.tool_name,
                input_params=pending_action.input_params,
                user_role=request_ctx.role,
                tenant_id=str(tenant_id),
            )

            await audit_repo.update_status(
                log_id=action_uuid,
                tenant_id=tenant_id,
                status=ActionStatus.SUCCESS,
                output_result=result,
            )

            message = "Acao executada com sucesso"
        except Exception as e:
            logger.error(f"Action execution error: {e}", exc_info=True)
            await audit_repo.update_status(
                log_id=action_uuid,
                tenant_id=tenant_id,
                status=ActionStatus.FAILED,
                error_message=str(e),
            )
            message = f"Erro ao executar acao: {str(e)}"
    else:
        await audit_repo.update_status(
            log_id=action_uuid,
            tenant_id=tenant_id,
            status=ActionStatus.CANCELLED,
        )
        message = "Acao cancelada"

    return ConfirmResponse(
        action_id=action_id,
        confirmed=request.confirmed,
        result=result,
        message=message,
    )


@router.get("/conversations", response_model=ConversationsListResponse)
async def list_conversations(
    limit: int = 20,
    offset: int = 0,
    request_ctx: RequestContext = Depends(get_request_context),
    db: AsyncSession = Depends(get_db_session),
) -> ConversationsListResponse:
    """
    List user's conversation sessions.

    Args:
        limit: Maximum number of conversations to return.
        offset: Number of conversations to skip.
        request_ctx: Request context with user and tenant info.
        db: Database session.

    Returns:
        List of conversation summaries.
    """
    user_id = UUID(request_ctx.user_id)
    tenant_id = UUID(request_ctx.effective_tenant_id)

    conv_repo = ConversationRepository(db)

    sessions = await conv_repo.get_user_sessions(
        user_id=user_id,
        tenant_id=tenant_id,
        limit=limit,
        offset=offset,
    )

    conversations = [
        ConversationSummary(
            session_id=session["session_id"],
            last_message_at=session["last_message_at"],
            message_count=session["message_count"],
        )
        for session in sessions
    ]

    return ConversationsListResponse(
        conversations=conversations,
        total=len(conversations),
    )


@router.get("/conversations/{session_id}", response_model=ConversationMessagesResponse)
async def get_conversation_messages(
    session_id: str,
    limit: int = 50,
    offset: int = 0,
    request_ctx: RequestContext = Depends(get_request_context),
    db: AsyncSession = Depends(get_db_session),
) -> ConversationMessagesResponse:
    """
    Get messages from a specific conversation session.

    Args:
        session_id: UUID of the conversation session.
        limit: Maximum number of messages to return.
        offset: Number of messages to skip.
        request_ctx: Request context with user and tenant info.
        db: Database session.

    Returns:
        Conversation messages.
    """
    try:
        session_uuid = UUID(session_id)
    except ValueError:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={"error": "invalid_session_id", "message": "ID da sessao invalido"},
        )

    tenant_id = UUID(request_ctx.effective_tenant_id)

    conv_repo = ConversationRepository(db)

    messages = await conv_repo.get_session_messages(
        session_id=session_uuid,
        tenant_id=tenant_id,
        limit=limit,
        offset=offset,
    )

    return ConversationMessagesResponse(
        session_id=session_id,
        messages=[
            MessageResponse(
                id=str(msg.id),
                role=msg.role,
                content=msg.content,
                tool_calls=msg.tool_calls,
                created_at=msg.created_at.isoformat() if msg.created_at else "",
            )
            for msg in messages
        ],
    )


@router.delete("/conversations/{session_id}", response_model=DeleteConversationResponse)
async def delete_conversation(
    session_id: str,
    request_ctx: RequestContext = Depends(get_request_context),
    db: AsyncSession = Depends(get_db_session),
) -> DeleteConversationResponse:
    """
    Delete a conversation session and all its messages.

    Args:
        session_id: UUID of the conversation session to delete.
        request_ctx: Request context with user and tenant info.
        db: Database session.

    Returns:
        Deletion result with count of deleted messages.
    """
    try:
        session_uuid = UUID(session_id)
    except ValueError:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={"error": "invalid_session_id", "message": "ID da sessao invalido"},
        )

    tenant_id = UUID(request_ctx.effective_tenant_id)

    conv_repo = ConversationRepository(db)

    deleted_count = await conv_repo.delete_session(
        session_id=session_uuid,
        tenant_id=tenant_id,
    )

    return DeleteConversationResponse(
        session_id=session_id,
        deleted_count=deleted_count,
        message=f"{deleted_count} mensagens removidas" if deleted_count > 0
                else "Nenhuma mensagem encontrada para remover",
    )


# Helper functions

def _format_sse_event(event_type: str, data: dict) -> str:
    """Format data as SSE event string."""
    return f"event: {event_type}\ndata: {json.dumps(data, ensure_ascii=False)}\n\n"


async def _process_with_agent(
    message: str,
    context: list[dict],
    request_ctx: RequestContext,
    session_id: Optional[UUID] = None,
) -> dict:
    """
    Process message with AI agent using LlamaIndex CopilotAgent.

    Args:
        message: User message to process.
        context: Recent conversation context.
        request_ctx: Request context with user and tenant info.
        session_id: Optional conversation session ID.

    Returns:
        Dictionary with response text, tool_calls, and confirmation_required.
    """
    try:
        # Create the simple agent (uses OpenAI directly)
        agent = create_simple_agent(
            request_ctx=request_ctx,
            conversation_id=session_id,
        )

        # Context is already in dict format for the simple agent
        chat_history = context

        # Process message with the agent
        result = await agent.chat(message, chat_history=chat_history)

        return result

    except Exception as e:
        logger.exception(f"Error processing with agent: {e}")
        return {
            "response": "Desculpe, ocorreu um erro ao processar sua solicitacao. Por favor, tente novamente.",
            "error": str(e),
            "tool_calls": [],
        }


def _requires_confirmation(response_text: str, original_message: str) -> bool:
    """
    Check if the action requires human confirmation.

    Args:
        response_text: AI response text.
        original_message: Original user message.

    Returns:
        True if confirmation is required.
    """
    # Keywords that trigger confirmation requirement
    confirmation_keywords = [
        "atualizar status",
        "alterar status",
        "mudar status",
        "update status",
        "confirmar",
    ]

    message_lower = original_message.lower()
    return any(keyword in message_lower for keyword in confirmation_keywords)


async def _execute_confirmed_action(
    tool_name: str,
    input_params: dict,
    user_role: str,
    tenant_id: str,
) -> dict:
    """
    Execute a confirmed action.

    This is a placeholder that will be replaced with actual tool execution
    from Task Group 6.

    Args:
        tool_name: Name of the tool to execute.
        input_params: Tool input parameters.
        user_role: User's role for permission checking.
        tenant_id: Tenant ID for data isolation.

    Returns:
        Execution result dictionary.
    """
    # Placeholder - will be replaced with actual tool execution
    return {
        "status": "executed",
        "tool_name": tool_name,
        "message": "Acao executada com sucesso (modo de teste)",
    }
