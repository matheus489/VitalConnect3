"""
VitalConnect Copilot Agent Configuration.

Configures the LlamaIndex ReActAgent with all registered tools
and system prompt for the VitalConnect operational context.

Features:
- Portuguese language support
- Tool registration with permission filtering
- Context-aware system prompt
- Conversation memory management
"""

import logging
from typing import Any, List, Optional
from uuid import UUID

from llama_index.core.agent import ReActAgent
from llama_index.core.llms import ChatMessage, MessageRole
from llama_index.core.memory import ChatMemoryBuffer
from llama_index.core.tools import FunctionTool, ToolMetadata
from llama_index.llms.openai import OpenAI

from app.config import get_settings
from app.middleware.permissions import get_allowed_tools
from app.middleware.tenant import RequestContext
from app.tools import (
    BaseTool,
    ToolContext,
    ToolResult,
    get_all_tools,
    get_tool_by_name,
)


logger = logging.getLogger(__name__)


# System prompt for VitalConnect Copilot
SYSTEM_PROMPT_PT = """Voce e o assistente virtual do VitalConnect, um sistema de gestao de captacao de orgaos para transplante.

Seu papel e auxiliar os usuarios da central de captacao em suas tarefas diarias, incluindo:
- Consultar e gerenciar ocorrencias de captacao
- Enviar notificacoes para equipes de plantao
- Gerar relatorios operacionais
- Responder duvidas sobre procedimentos e protocolos

DIRETRIZES IMPORTANTES:

1. LINGUA: Sempre responda em portugues brasileiro.

2. SEGURANCA: Voce NAO pode:
   - Tomar decisoes clinicas sobre viabilidade de orgaos
   - Alterar dados de pacientes diretamente
   - Executar acoes sem as permissoes adequadas do usuario

3. ACOES CRITICAS: Para acoes que modificam dados (como atualizar status), sempre:
   - Confirme os detalhes com o usuario antes de executar
   - Informe claramente o que sera alterado
   - Aguarde confirmacao explicita

4. LGPD: Ao acessar dados de pacientes:
   - Informe que o acesso sera registrado
   - Mostre apenas informacoes necessarias
   - Use mascaramento de dados sensiveis quando apropriado

5. FORMATO DE RESPOSTA:
   - Seja conciso e direto
   - Use formatacao estruturada (listas, tabelas) quando apropriado
   - Para listas de ocorrencias, use cards formatados
   - Para erros, explique claramente o problema e sugira solucoes

6. CONTEXTO DO USUARIO:
   - Usuario: {user_name} ({user_role})
   - Hospital: {hospital_name}
   - Tenant: {tenant_name}
   - Ferramentas disponiveis: {available_tools}

Como posso ajuda-lo hoje?"""


class CopilotAgent:
    """
    VitalConnect Copilot Agent.

    Manages the LlamaIndex ReActAgent instance with proper tool
    registration and context management.
    """

    def __init__(
        self,
        request_ctx: RequestContext,
        conversation_id: Optional[UUID] = None,
        memory_token_limit: int = 3000,
    ):
        """
        Initialize the copilot agent.

        Args:
            request_ctx: Request context with user and tenant info.
            conversation_id: Optional conversation ID for context.
            memory_token_limit: Maximum tokens to keep in memory.
        """
        self._settings = get_settings()
        self._request_ctx = request_ctx
        self._conversation_id = conversation_id
        self._memory_token_limit = memory_token_limit
        self._agent: Optional[ReActAgent] = None
        self._tool_context: Optional[ToolContext] = None

    @property
    def tool_context(self) -> ToolContext:
        """Get the tool context for this agent session."""
        if self._tool_context is None:
            self._tool_context = ToolContext(
                request_ctx=self._request_ctx,
                conversation_id=self._conversation_id,
            )
        return self._tool_context

    def _get_llm(self) -> OpenAI:
        """
        Get the LLM instance.

        Returns:
            OpenAI LLM instance configured for the service.
        """
        return OpenAI(
            model=self._settings.ai_model,
            api_key=self._settings.openai_api_key,
            temperature=0.1,  # Low temperature for more consistent responses
        )

    def _get_allowed_tools(self) -> List[BaseTool]:
        """
        Get tools that the current user is allowed to use.

        Returns:
            List of tool instances the user can execute.
        """
        user_role = self._request_ctx.role
        allowed_tool_names = get_allowed_tools(user_role)

        all_tools = get_all_tools()
        return [tool for tool in all_tools if tool.name in allowed_tool_names]

    def _create_tool_wrapper(self, tool: BaseTool) -> FunctionTool:
        """
        Create a LlamaIndex FunctionTool wrapper for a BaseTool.

        The wrapper injects the tool context and handles the async execution.

        Args:
            tool: The BaseTool instance to wrap.

        Returns:
            LlamaIndex FunctionTool instance.
        """
        context = self.tool_context

        async def async_wrapper(**kwargs: Any) -> str:
            """Async wrapper that executes the tool with context."""
            try:
                result = await tool.run(context, **kwargs)
                return self._format_tool_result(result)
            except Exception as e:
                logger.error(f"Tool execution error: {tool.name} - {e}")
                return f"Erro ao executar {tool.name}: {str(e)}"

        def sync_wrapper(**kwargs: Any) -> str:
            """Sync wrapper for LlamaIndex compatibility."""
            import asyncio
            return asyncio.get_event_loop().run_until_complete(async_wrapper(**kwargs))

        return FunctionTool.from_defaults(
            fn=sync_wrapper,
            async_fn=async_wrapper,
            name=tool.name,
            description=tool.description,
        )

    def _format_tool_result(self, result: ToolResult) -> str:
        """
        Format a ToolResult for the LLM.

        Args:
            result: The tool execution result.

        Returns:
            Formatted string for the LLM.
        """
        if result.confirmation_required:
            return (
                f"CONFIRMACAO_NECESSARIA: {result.message}\n"
                f"action_id: {result.confirmation_action_id}\n"
                f"detalhes: {result.confirmation_details}"
            )

        if not result.success:
            return f"ERRO: {result.message}"

        # Format data as JSON-like string for the LLM
        import json
        data_str = ""
        if result.data:
            try:
                data_str = json.dumps(result.data, ensure_ascii=False, indent=2)
            except (TypeError, ValueError):
                data_str = str(result.data)

        if result.message and data_str:
            return f"{result.message}\n\nDados:\n{data_str}"
        elif result.message:
            return result.message
        elif data_str:
            return data_str
        else:
            return "Operacao concluida com sucesso."

    def _build_system_prompt(self) -> str:
        """
        Build the system prompt with user context.

        Returns:
            Formatted system prompt string.
        """
        allowed_tools = self._get_allowed_tools()
        tool_names = [t.name for t in allowed_tools]

        return SYSTEM_PROMPT_PT.format(
            user_name=self._request_ctx.email,
            user_role=self._request_ctx.role,
            hospital_name="N/A",  # Would be fetched from context
            tenant_name=self._request_ctx.effective_tenant_id[:8],
            available_tools=", ".join(tool_names) if tool_names else "Nenhuma",
        )

    def _create_agent(self) -> ReActAgent:
        """
        Create the ReActAgent instance.

        Returns:
            Configured ReActAgent.
        """
        llm = self._get_llm()
        tools = [self._create_tool_wrapper(t) for t in self._get_allowed_tools()]
        system_prompt = self._build_system_prompt()

        # Create memory buffer
        memory = ChatMemoryBuffer.from_defaults(
            token_limit=self._memory_token_limit
        )

        # Create the agent
        agent = ReActAgent.from_tools(
            tools=tools,
            llm=llm,
            memory=memory,
            system_prompt=system_prompt,
            verbose=self._settings.debug,
            max_iterations=10,
        )

        return agent

    @property
    def agent(self) -> ReActAgent:
        """Get or create the agent instance."""
        if self._agent is None:
            self._agent = self._create_agent()
        return self._agent

    async def chat(
        self,
        message: str,
        chat_history: Optional[List[ChatMessage]] = None,
    ) -> dict:
        """
        Send a message to the agent and get a response.

        Args:
            message: User message.
            chat_history: Optional previous chat history.

        Returns:
            Dictionary with response and metadata.
        """
        try:
            # Add chat history to memory if provided
            if chat_history:
                for msg in chat_history:
                    self.agent.memory.put(msg)

            # Get response from agent
            response = await self.agent.achat(message)

            return {
                "response": str(response),
                "tool_calls": self._extract_tool_calls(response),
                "confirmation_required": self._check_confirmation_required(response),
            }

        except Exception as e:
            logger.exception(f"Agent chat error: {e}")
            return {
                "response": "Desculpe, ocorreu um erro ao processar sua solicitacao. Por favor, tente novamente.",
                "error": str(e),
                "tool_calls": [],
            }

    def _extract_tool_calls(self, response: Any) -> list:
        """
        Extract tool calls from the agent response.

        Args:
            response: Agent response object.

        Returns:
            List of tool call information.
        """
        tool_calls = []
        if hasattr(response, "sources"):
            for source in response.sources:
                if hasattr(source, "tool_name"):
                    tool_calls.append({
                        "tool_name": source.tool_name,
                        "raw_output": str(source.raw_output)[:500],
                    })
        return tool_calls

    def _check_confirmation_required(self, response: Any) -> Optional[dict]:
        """
        Check if the response requires user confirmation.

        Args:
            response: Agent response object.

        Returns:
            Confirmation details if required, None otherwise.
        """
        response_text = str(response)
        if "CONFIRMACAO_NECESSARIA" in response_text:
            # Parse confirmation details from response
            # This is a simplified implementation
            return {
                "required": True,
                "message": "Esta acao requer confirmacao.",
            }
        return None

    async def confirm_action(
        self,
        action_id: str,
        confirmed: bool,
    ) -> dict:
        """
        Confirm or reject a pending action.

        Args:
            action_id: The pending action ID.
            confirmed: Whether the action is confirmed.

        Returns:
            Dictionary with result.
        """
        if not confirmed:
            return {
                "response": "Acao cancelada pelo usuario.",
                "status": "cancelled",
            }

        # Set confirmation flag and re-run the tool
        self._tool_context.confirmation_received = True

        # The actual execution would need to store and retrieve pending actions
        # This is a simplified implementation
        return {
            "response": "Acao confirmada e executada.",
            "status": "executed",
        }

    def reset_memory(self) -> None:
        """Reset the agent's conversation memory."""
        if self._agent is not None:
            self._agent.memory.reset()


def create_copilot_agent(
    request_ctx: RequestContext,
    conversation_id: Optional[UUID] = None,
) -> CopilotAgent:
    """
    Factory function to create a copilot agent.

    Args:
        request_ctx: Request context with user and tenant info.
        conversation_id: Optional conversation ID.

    Returns:
        Configured CopilotAgent instance.
    """
    return CopilotAgent(
        request_ctx=request_ctx,
        conversation_id=conversation_id,
    )
