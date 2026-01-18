"""
Celery tasks for AI query processing.

High priority tasks for processing user chat messages asynchronously.
Routed to the ai_query queue.
"""

import logging
import time
from datetime import datetime, timezone
from typing import Any, Optional
from uuid import UUID

from celery import shared_task

from app.celery_app import celery_app, QUEUE_AI_QUERY
from app.celery_app.tasks.base import AuditedTask


logger = logging.getLogger(__name__)


@celery_app.task(
    bind=True,
    base=AuditedTask,
    name="app.celery_app.tasks.query.process_chat_message",
    queue=QUEUE_AI_QUERY,
    max_retries=3,
    default_retry_delay=10,
)
def process_chat_message(
    self,
    tenant_id: str,
    user_id: str,
    session_id: str,
    message: str,
    context: Optional[list] = None,
    user_role: str = "operador",
    **kwargs,
) -> dict:
    """
    Process a chat message asynchronously.

    This task handles AI query processing in the background,
    allowing for long-running operations without blocking the API.

    Args:
        tenant_id: Tenant UUID string.
        user_id: User UUID string.
        session_id: Conversation session UUID string.
        message: User message to process.
        context: Optional list of recent conversation context.
        user_role: User's role for permission checking.

    Returns:
        Dictionary with response and metadata.
    """
    start_time = time.time()

    logger.info(
        f"Processing chat message: tenant={tenant_id}, user={user_id}, "
        f"session={session_id}, message_length={len(message)}"
    )

    try:
        # Process with AI agent
        # This will be replaced with actual LlamaIndex agent call
        response = _process_message_with_agent(
            message=message,
            context=context or [],
            user_role=user_role,
            tenant_id=tenant_id,
        )

        execution_time_ms = int((time.time() - start_time) * 1000)

        logger.info(
            f"Chat message processed successfully: "
            f"execution_time_ms={execution_time_ms}"
        )

        return {
            "status": "success",
            "response": response["text"],
            "tool_calls": response.get("tool_calls"),
            "confirmation_required": response.get("confirmation_required"),
            "metadata": {
                "execution_time_ms": execution_time_ms,
                "processed_at": datetime.now(timezone.utc).isoformat(),
                "model_used": response.get("model", "placeholder"),
            },
        }

    except Exception as e:
        logger.error(f"Chat message processing failed: {e}", exc_info=True)

        # Retry with exponential backoff
        retry_count = self.request.retries
        if retry_count < self.max_retries:
            raise self.retry(exc=e)

        return {
            "status": "failed",
            "error": str(e),
            "metadata": {
                "execution_time_ms": int((time.time() - start_time) * 1000),
                "failed_at": datetime.now(timezone.utc).isoformat(),
                "retries_exhausted": True,
            },
        }


@celery_app.task(
    bind=True,
    base=AuditedTask,
    name="app.celery_app.tasks.query.process_rag_query",
    queue=QUEUE_AI_QUERY,
    max_retries=2,
)
def process_rag_query(
    self,
    tenant_id: str,
    user_id: str,
    query: str,
    top_k: int = 5,
    **kwargs,
) -> dict:
    """
    Process a RAG query to search documentation.

    This task searches the vector store for relevant documentation
    chunks and returns the results.

    Args:
        tenant_id: Tenant UUID string for isolation.
        user_id: User UUID string.
        query: Search query text.
        top_k: Number of results to return.

    Returns:
        Dictionary with search results.
    """
    start_time = time.time()

    logger.info(
        f"Processing RAG query: tenant={tenant_id}, query_length={len(query)}"
    )

    try:
        # Search documentation
        results = _search_documentation(
            tenant_id=tenant_id,
            query=query,
            top_k=top_k,
        )

        execution_time_ms = int((time.time() - start_time) * 1000)

        return {
            "status": "success",
            "results": results,
            "metadata": {
                "execution_time_ms": execution_time_ms,
                "result_count": len(results),
                "query": query[:100],  # Truncate for logging
            },
        }

    except Exception as e:
        logger.error(f"RAG query failed: {e}", exc_info=True)

        if self.request.retries < self.max_retries:
            raise self.retry(exc=e)

        return {
            "status": "failed",
            "error": str(e),
            "results": [],
        }


@celery_app.task(
    bind=True,
    base=AuditedTask,
    name="app.celery_app.tasks.query.execute_tool",
    queue=QUEUE_AI_QUERY,
    max_retries=2,
)
def execute_tool(
    self,
    tenant_id: str,
    user_id: str,
    tool_name: str,
    tool_params: dict,
    user_role: str = "operador",
    **kwargs,
) -> dict:
    """
    Execute a tool asynchronously.

    This task handles tool execution in the background,
    useful for long-running tool operations.

    Args:
        tenant_id: Tenant UUID string.
        user_id: User UUID string.
        tool_name: Name of the tool to execute.
        tool_params: Tool input parameters.
        user_role: User's role for permission checking.

    Returns:
        Dictionary with tool execution result.
    """
    start_time = time.time()

    logger.info(
        f"Executing tool: tenant={tenant_id}, user={user_id}, tool={tool_name}"
    )

    try:
        # Check permissions
        from app.middleware.permissions import check_permission, PermissionDeniedError

        try:
            check_permission(user_role, tool_name)
        except PermissionDeniedError as e:
            return {
                "status": "failed",
                "error": f"Permission denied: {e.message}",
                "tool_name": tool_name,
            }

        # Execute tool
        result = _execute_tool_impl(
            tenant_id=tenant_id,
            tool_name=tool_name,
            tool_params=tool_params,
        )

        execution_time_ms = int((time.time() - start_time) * 1000)

        return {
            "status": "success",
            "tool_name": tool_name,
            "result": result,
            "metadata": {
                "execution_time_ms": execution_time_ms,
            },
        }

    except Exception as e:
        logger.error(f"Tool execution failed: {e}", exc_info=True)

        if self.request.retries < self.max_retries:
            raise self.retry(exc=e)

        return {
            "status": "failed",
            "tool_name": tool_name,
            "error": str(e),
        }


# Helper functions

def _process_message_with_agent(
    message: str,
    context: list,
    user_role: str,
    tenant_id: str,
) -> dict:
    """
    Process message with AI agent.

    Placeholder implementation - will be replaced with actual
    LlamaIndex agent integration from Task Group 6.

    Args:
        message: User message.
        context: Conversation context.
        user_role: User's role.
        tenant_id: Tenant ID.

    Returns:
        Response dictionary.
    """
    # This will be replaced with actual agent call
    # from app.agents.copilot_agent import CopilotAgent

    context_info = f" (com {len(context)} mensagens de contexto)" if context else ""

    return {
        "text": (
            f"Mensagem processada assincronamente{context_info}. "
            f"O agente de IA sera integrado em breve."
        ),
        "tool_calls": None,
        "confirmation_required": None,
        "model": "placeholder",
    }


def _search_documentation(
    tenant_id: str,
    query: str,
    top_k: int,
) -> list:
    """
    Search documentation via RAG.

    Placeholder implementation - will be replaced with actual
    RAG query engine integration from Task Group 5.

    Args:
        tenant_id: Tenant ID for isolation.
        query: Search query.
        top_k: Number of results.

    Returns:
        List of search results.
    """
    # This will be replaced with actual RAG query
    # from app.rag.query_engine import RAGQueryEngine

    return [
        {
            "content": "Resultado de busca placeholder",
            "source": "documentation",
            "score": 0.0,
        }
    ]


def _execute_tool_impl(
    tenant_id: str,
    tool_name: str,
    tool_params: dict,
) -> dict:
    """
    Execute a tool implementation.

    Placeholder implementation - will be replaced with actual
    tool implementations from Task Group 6.

    Args:
        tenant_id: Tenant ID.
        tool_name: Tool name.
        tool_params: Tool parameters.

    Returns:
        Tool execution result.
    """
    # This will be replaced with actual tool execution
    # from app.tools import get_tool_by_name

    return {
        "executed": True,
        "tool_name": tool_name,
        "message": "Ferramenta executada em modo de teste",
    }
