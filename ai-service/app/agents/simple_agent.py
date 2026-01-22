"""
Simple AI Agent using OpenAI directly.

A simplified implementation that bypasses llama_index compatibility issues
with Python 3.14 by using the OpenAI client directly.
"""

import logging
from typing import Any, List, Optional
from uuid import UUID

from openai import AsyncOpenAI

from app.config import get_settings
from app.middleware.tenant import RequestContext


logger = logging.getLogger(__name__)


# System prompt for SIDOT Copilot (Portuguese)
SYSTEM_PROMPT_PT = """Voce e o assistente virtual do SIDOT, um sistema de gestao de captacao de orgaos para transplante.

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
   - Usuario: {user_email} ({user_role})
   - Tenant ID: {tenant_id}

Como posso ajuda-lo hoje?"""


class SimpleAgent:
    """
    Simple AI Agent using OpenAI directly.

    Provides a minimal implementation for chat functionality
    without the complexity of LlamaIndex.
    """

    def __init__(
        self,
        request_ctx: RequestContext,
        conversation_id: Optional[UUID] = None,
    ):
        """
        Initialize the simple agent.

        Args:
            request_ctx: Request context with user and tenant info.
            conversation_id: Optional conversation ID for context.
        """
        self._settings = get_settings()
        self._request_ctx = request_ctx
        self._conversation_id = conversation_id
        self._client: Optional[AsyncOpenAI] = None

    @property
    def client(self) -> AsyncOpenAI:
        """Get or create the OpenAI client."""
        if self._client is None:
            self._client = AsyncOpenAI(api_key=self._settings.openai_api_key)
        return self._client

    def _build_system_prompt(self) -> str:
        """
        Build the system prompt with user context.

        Returns:
            Formatted system prompt string.
        """
        return SYSTEM_PROMPT_PT.format(
            user_email=self._request_ctx.email,
            user_role=self._request_ctx.role,
            tenant_id=self._request_ctx.effective_tenant_id[:8],
        )

    def _convert_history_to_messages(
        self,
        chat_history: List[dict],
    ) -> List[dict]:
        """
        Convert chat history to OpenAI message format.

        Args:
            chat_history: List of message dicts with role and content.

        Returns:
            List of OpenAI-compatible message dicts.
        """
        messages = []
        for msg in chat_history:
            role = msg.get("role", "user")
            content = msg.get("content", "")
            if role in ("user", "assistant", "system"):
                messages.append({"role": role, "content": content})
        return messages

    async def chat(
        self,
        message: str,
        chat_history: Optional[List[dict]] = None,
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
            # Build messages list
            messages = [
                {"role": "system", "content": self._build_system_prompt()}
            ]

            # Add chat history if provided
            if chat_history:
                messages.extend(self._convert_history_to_messages(chat_history))

            # Add current user message
            messages.append({"role": "user", "content": message})

            # Call OpenAI API
            response = await self.client.chat.completions.create(
                model=self._settings.ai_model or "gpt-4o-mini",
                messages=messages,
                temperature=0.1,
                max_tokens=2000,
            )

            # Extract response text
            response_text = response.choices[0].message.content or ""

            return {
                "response": response_text,
                "tool_calls": [],
                "confirmation_required": None,
            }

        except Exception as e:
            logger.exception(f"Agent chat error: {e}")
            return {
                "response": "Desculpe, ocorreu um erro ao processar sua solicitacao. Por favor, tente novamente.",
                "error": str(e),
                "tool_calls": [],
            }


def create_simple_agent(
    request_ctx: RequestContext,
    conversation_id: Optional[UUID] = None,
) -> SimpleAgent:
    """
    Factory function to create a simple agent.

    Args:
        request_ctx: Request context with user and tenant info.
        conversation_id: Optional conversation ID.

    Returns:
        Configured SimpleAgent instance.
    """
    return SimpleAgent(
        request_ctx=request_ctx,
        conversation_id=conversation_id,
    )
