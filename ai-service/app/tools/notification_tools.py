"""
Notification Tools for VitalConnect AI Assistant.

Provides tools for sending notifications to team members based on
shift schedules and notification preferences.

Tools:
- send_team_notification: Send push/SMS notifications to shift team
"""

import logging
from typing import Any, List, Optional

from app.tools.base import (
    BaseTool,
    ToolContext,
    ToolResult,
    ToolExecutionError,
)


logger = logging.getLogger(__name__)


class SendTeamNotificationTool(BaseTool):
    """
    Tool for sending notifications to team members.

    Queries the shift schedule from the Go backend and sends
    push/SMS notifications via the existing notification service.

    Permissions: gestor+ (gestor, admin)
    """

    name = "send_team_notification"
    description = """Envia notificacoes para membros da equipe de plantao.

    Parametros:
    - message: Mensagem a ser enviada (obrigatorio)
    - notification_type: Tipo de notificacao (push, sms, ambos) - padrao: push
    - team_role: Filtrar por funcao da equipe (medico, enfermeiro, coordenador, todos)
    - hospital_id: ID do hospital para filtrar equipe (opcional, usa hospital do contexto)
    - priority: Prioridade da mensagem (normal, alta, urgente) - padrao: normal
    - occurrence_id: ID da ocorrencia relacionada (opcional)

    A equipe sera identificada automaticamente com base no plantao atual.
    Notificacoes urgentes serao enviadas via SMS e push simultaneamente."""

    requires_confirmation = True  # Sending notifications requires confirmation
    audit_severity = "WARN"

    VALID_NOTIFICATION_TYPES = ["push", "sms", "ambos"]
    VALID_TEAM_ROLES = ["medico", "enfermeiro", "coordenador", "todos"]
    VALID_PRIORITIES = ["normal", "alta", "urgente"]

    async def execute(
        self,
        context: ToolContext,
        message: str,
        notification_type: str = "push",
        team_role: str = "todos",
        hospital_id: Optional[str] = None,
        priority: str = "normal",
        occurrence_id: Optional[str] = None,
        **kwargs: Any,
    ) -> ToolResult:
        """
        Execute the notification send.

        Args:
            context: Tool execution context.
            message: Message content to send.
            notification_type: Type of notification (push, sms, ambos).
            team_role: Filter by team role.
            hospital_id: Hospital ID to filter team.
            priority: Message priority.
            occurrence_id: Related occurrence ID.

        Returns:
            ToolResult with notification outcome.
        """
        # Validate inputs
        if not message or not message.strip():
            return ToolResult(
                success=False,
                message="A mensagem nao pode estar vazia.",
            )

        if notification_type not in self.VALID_NOTIFICATION_TYPES:
            return ToolResult(
                success=False,
                message=f"Tipo de notificacao invalido. Tipos validos: {', '.join(self.VALID_NOTIFICATION_TYPES)}",
            )

        if team_role not in self.VALID_TEAM_ROLES:
            return ToolResult(
                success=False,
                message=f"Funcao da equipe invalida. Funcoes validas: {', '.join(self.VALID_TEAM_ROLES)}",
            )

        if priority not in self.VALID_PRIORITIES:
            return ToolResult(
                success=False,
                message=f"Prioridade invalida. Prioridades validas: {', '.join(self.VALID_PRIORITIES)}",
            )

        # For urgent messages, always send both types
        if priority == "urgente":
            notification_type = "ambos"

        try:
            # First, get the current shift schedule
            shift_params = {}
            if hospital_id:
                shift_params["hospital_id"] = hospital_id
            if team_role != "todos":
                shift_params["role"] = team_role

            shift_response = await self.call_backend_api(
                method="GET",
                endpoint="/api/v1/plantoes/atual",
                context=context,
                params=shift_params,
            )

            team_members = shift_response.get("data", [])

            if not team_members:
                return ToolResult(
                    success=True,
                    data={"recipients": [], "message_sent": False},
                    message="Nenhum membro da equipe encontrado no plantao atual.",
                )

            # Prepare notification payload
            notification_data = {
                "message": message,
                "title": self._get_notification_title(priority, occurrence_id),
                "type": notification_type,
                "priority": priority,
                "recipients": [m.get("user_id") for m in team_members],
                "sender_id": context.user_id,
                "metadata": {
                    "sent_via": "ai_assistant",
                    "occurrence_id": occurrence_id,
                },
            }

            # Send notifications via backend API
            send_response = await self.call_backend_api(
                method="POST",
                endpoint="/api/v1/notifications/send",
                context=context,
                data=notification_data,
            )

            sent_count = send_response.get("sent_count", len(team_members))
            failed_count = send_response.get("failed_count", 0)

            logger.info(
                f"Team notification sent: {sent_count} recipients "
                f"by user {context.user_id} "
                f"type={notification_type} priority={priority}"
            )

            return ToolResult(
                success=True,
                data={
                    "recipients_count": sent_count,
                    "failed_count": failed_count,
                    "notification_type": notification_type,
                    "priority": priority,
                    "recipients": [
                        {
                            "name": m.get("nome"),
                            "role": m.get("funcao"),
                        }
                        for m in team_members
                    ],
                },
                message=f"Notificacao enviada para {sent_count} membros da equipe.",
            )

        except Exception as e:
            logger.error(f"Error sending team notification: {e}")
            raise ToolExecutionError(
                message="Erro ao enviar notificacao para a equipe.",
                details={"error": str(e)}
            )

    def _get_notification_title(
        self,
        priority: str,
        occurrence_id: Optional[str] = None
    ) -> str:
        """
        Generate notification title based on priority.

        Args:
            priority: Notification priority.
            occurrence_id: Related occurrence ID.

        Returns:
            Notification title string.
        """
        prefix = ""
        if priority == "urgente":
            prefix = "[URGENTE] "
        elif priority == "alta":
            prefix = "[IMPORTANTE] "

        if occurrence_id:
            return f"{prefix}VitalConnect - Ocorrencia #{occurrence_id[:8]}"

        return f"{prefix}VitalConnect - Mensagem da Equipe"

    def get_confirmation_details(
        self,
        context: ToolContext,
        message: str = "",
        notification_type: str = "push",
        team_role: str = "todos",
        priority: str = "normal",
        **kwargs: Any,
    ) -> dict:
        """
        Get confirmation details for notification send.

        Args:
            context: Tool execution context.
            message: Message content.
            notification_type: Type of notification.
            team_role: Target team role.
            priority: Message priority.

        Returns:
            Dictionary with confirmation message and details.
        """
        type_display = {
            "push": "notificacao push",
            "sms": "SMS",
            "ambos": "notificacao push e SMS",
        }

        role_display = {
            "todos": "toda a equipe",
            "medico": "medicos",
            "enfermeiro": "enfermeiros",
            "coordenador": "coordenadores",
        }

        return {
            "message": f"Deseja enviar {type_display.get(notification_type, notification_type)} para {role_display.get(team_role, team_role)}?",
            "action": "Enviar Notificacao",
            "warning": "A mensagem sera enviada para todos os membros do plantao atual.",
            "details": {
                "preview_message": message[:100] + "..." if len(message) > 100 else message,
                "notification_type": notification_type,
                "target_team": team_role,
                "priority": priority,
            },
        }


# Tool instance for registration
send_team_notification_tool = SendTeamNotificationTool()
