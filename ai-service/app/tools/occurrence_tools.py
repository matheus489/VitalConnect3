"""
Occurrence Tools for SIDOT AI Assistant.

Provides tools for querying and managing occurrences (organ capture events).
These tools communicate with the Go backend API to perform operations.

Tools:
- list_occurrences: Query occurrences with filters
- get_occurrence_details: Get specific occurrence with LGPD data
- update_occurrence_status: Change occurrence status (requires confirmation)
"""

import logging
from datetime import datetime
from typing import Any, Optional

from app.tools.base import (
    BaseTool,
    ToolContext,
    ToolResult,
    ToolExecutionError,
)


logger = logging.getLogger(__name__)


class ListOccurrencesTool(BaseTool):
    """
    Tool for listing occurrences with various filters.

    Queries the Go backend API to retrieve occurrences filtered by
    status, hospital, date range, etc.

    Permissions: All authenticated users (admin, gestor, operador, medico)
    """

    name = "list_occurrences"
    description = """Lista ocorrencias de captacao de orgaos com filtros opcionais.

    Parametros:
    - status: Filtrar por status (aberta, em_andamento, concluida, cancelada)
    - hospital_id: Filtrar por ID do hospital
    - start_date: Data inicial (formato ISO: YYYY-MM-DD)
    - end_date: Data final (formato ISO: YYYY-MM-DD)
    - limit: Numero maximo de resultados (padrao: 20)

    Retorna lista de ocorrencias com informacoes basicas para visualizacao."""

    requires_confirmation = False
    audit_severity = "INFO"

    async def execute(
        self,
        context: ToolContext,
        status: Optional[str] = None,
        hospital_id: Optional[str] = None,
        start_date: Optional[str] = None,
        end_date: Optional[str] = None,
        limit: int = 20,
        **kwargs: Any,
    ) -> ToolResult:
        """
        Execute the list occurrences query.

        Args:
            context: Tool execution context.
            status: Filter by occurrence status.
            hospital_id: Filter by hospital ID.
            start_date: Filter by start date (ISO format).
            end_date: Filter by end date (ISO format).
            limit: Maximum number of results.

        Returns:
            ToolResult with list of occurrences.
        """
        # Build query parameters
        params = {"limit": min(limit, 100)}  # Cap at 100

        if status:
            params["status"] = status
        if hospital_id:
            params["hospital_id"] = hospital_id
        if start_date:
            params["start_date"] = start_date
        if end_date:
            params["end_date"] = end_date

        try:
            # Call backend API
            response = await self.call_backend_api(
                method="GET",
                endpoint="/api/v1/ocorrencias",
                context=context,
                params=params,
            )

            occurrences = response.get("data", [])
            total = response.get("total", len(occurrences))

            # Format occurrences for display
            formatted_occurrences = []
            for occ in occurrences:
                formatted_occurrences.append({
                    "id": occ.get("id"),
                    "hospital_nome": occ.get("hospital", {}).get("nome", "N/A"),
                    "status": occ.get("status"),
                    "data_abertura": occ.get("data_abertura"),
                    "tempo_restante": self._calculate_remaining_time(occ),
                    "orgaos": [o.get("tipo") for o in occ.get("orgaos", [])],
                })

            return ToolResult(
                success=True,
                data={
                    "occurrences": formatted_occurrences,
                    "total": total,
                    "filters_applied": {
                        "status": status,
                        "hospital_id": hospital_id,
                        "start_date": start_date,
                        "end_date": end_date,
                    },
                },
                message=f"Encontradas {len(formatted_occurrences)} ocorrencias.",
            )

        except Exception as e:
            logger.error(f"Error listing occurrences: {e}")
            raise ToolExecutionError(
                message="Erro ao buscar ocorrencias. Tente novamente.",
                details={"error": str(e)}
            )

    def _calculate_remaining_time(self, occurrence: dict) -> Optional[str]:
        """
        Calculate remaining time for the occurrence based on organ viability.

        Args:
            occurrence: Occurrence data dictionary.

        Returns:
            Human-readable remaining time string or None.
        """
        # This would typically calculate based on organ types and their viability windows
        # Simplified implementation for now
        deadline = occurrence.get("prazo_limite")
        if not deadline:
            return None

        try:
            deadline_dt = datetime.fromisoformat(deadline.replace("Z", "+00:00"))
            now = datetime.now(deadline_dt.tzinfo)
            delta = deadline_dt - now

            if delta.total_seconds() < 0:
                return "Expirado"

            hours = int(delta.total_seconds() // 3600)
            minutes = int((delta.total_seconds() % 3600) // 60)

            if hours > 0:
                return f"{hours}h {minutes}min"
            return f"{minutes}min"

        except (ValueError, TypeError):
            return None


class GetOccurrenceDetailsTool(BaseTool):
    """
    Tool for retrieving detailed occurrence information.

    Returns full occurrence details including LGPD-protected patient data.
    Access to sensitive data is logged for audit compliance.

    Permissions: All authenticated users, but LGPD data access is audited
    """

    name = "get_occurrence_details"
    description = """Obtem detalhes completos de uma ocorrencia especifica.

    Parametros:
    - occurrence_id: ID da ocorrencia (obrigatorio)

    Retorna informacoes detalhadas incluindo:
    - Dados do paciente (sujeito a LGPD)
    - Status atual e historico
    - Orgaos disponibilizados
    - Equipes envolvidas
    - Timeline de eventos

    ATENCAO: O acesso a dados de paciente e registrado para auditoria LGPD."""

    requires_confirmation = False
    audit_severity = "WARN"  # Higher severity due to LGPD data access

    async def execute(
        self,
        context: ToolContext,
        occurrence_id: str,
        **kwargs: Any,
    ) -> ToolResult:
        """
        Execute the get occurrence details query.

        Args:
            context: Tool execution context.
            occurrence_id: ID of the occurrence to retrieve.

        Returns:
            ToolResult with occurrence details.
        """
        if not occurrence_id:
            return ToolResult(
                success=False,
                message="ID da ocorrencia e obrigatorio.",
            )

        try:
            # Call backend API
            response = await self.call_backend_api(
                method="GET",
                endpoint=f"/api/v1/ocorrencias/{occurrence_id}",
                context=context,
            )

            occurrence = response.get("data", response)

            # Log LGPD data access
            logger.info(
                f"LGPD data access: occurrence={occurrence_id} "
                f"user={context.user_id} tenant={context.tenant_id}"
            )

            # Format response with masked sensitive data for display
            formatted_occurrence = {
                "id": occurrence.get("id"),
                "status": occurrence.get("status"),
                "hospital": {
                    "id": occurrence.get("hospital", {}).get("id"),
                    "nome": occurrence.get("hospital", {}).get("nome"),
                },
                "paciente": {
                    # Mask patient data for display (full data in LGPD-compliant format)
                    "nome_inicial": self._mask_name(occurrence.get("paciente", {}).get("nome")),
                    "idade": occurrence.get("paciente", {}).get("idade"),
                    "sexo": occurrence.get("paciente", {}).get("sexo"),
                    "tipo_sanguineo": occurrence.get("paciente", {}).get("tipo_sanguineo"),
                },
                "orgaos": occurrence.get("orgaos", []),
                "data_abertura": occurrence.get("data_abertura"),
                "data_atualizacao": occurrence.get("data_atualizacao"),
                "prazo_limite": occurrence.get("prazo_limite"),
                "equipes": occurrence.get("equipes", []),
                "timeline": occurrence.get("timeline", []),
                "observacoes": occurrence.get("observacoes"),
            }

            return ToolResult(
                success=True,
                data={"occurrence": formatted_occurrence},
                message=f"Detalhes da ocorrencia {occurrence_id} carregados.",
            )

        except Exception as e:
            logger.error(f"Error getting occurrence details: {e}")
            raise ToolExecutionError(
                message="Erro ao buscar detalhes da ocorrencia.",
                details={"occurrence_id": occurrence_id, "error": str(e)}
            )

    def _mask_name(self, name: Optional[str]) -> str:
        """
        Mask patient name for display while preserving identifiability.

        Args:
            name: Full patient name.

        Returns:
            Masked name (e.g., "Joao S.").
        """
        if not name:
            return "N/A"

        parts = name.split()
        if len(parts) == 1:
            return f"{parts[0][0]}."

        # First name + last initial
        return f"{parts[0]} {parts[-1][0]}."


class UpdateOccurrenceStatusTool(BaseTool):
    """
    Tool for updating the status of an occurrence.

    This tool requires human-in-the-loop confirmation before execution
    to prevent accidental status changes.

    Permissions: operador+ (operador, gestor, admin)
    """

    name = "update_occurrence_status"
    description = """Atualiza o status de uma ocorrencia.

    Parametros:
    - occurrence_id: ID da ocorrencia (obrigatorio)
    - new_status: Novo status (aberta, em_andamento, concluida, cancelada)
    - observacao: Observacao sobre a mudanca de status (opcional)

    IMPORTANTE: Esta acao requer confirmacao do usuario antes de ser executada.
    Apos submeter, o usuario recebera uma solicitacao de confirmacao."""

    requires_confirmation = True
    audit_severity = "WARN"

    # Valid status transitions
    VALID_STATUSES = ["aberta", "em_andamento", "concluida", "cancelada"]

    async def execute(
        self,
        context: ToolContext,
        occurrence_id: str,
        new_status: str,
        observacao: Optional[str] = None,
        **kwargs: Any,
    ) -> ToolResult:
        """
        Execute the status update.

        This method is only called after user confirmation.

        Args:
            context: Tool execution context.
            occurrence_id: ID of the occurrence to update.
            new_status: New status to set.
            observacao: Optional observation about the change.

        Returns:
            ToolResult with update outcome.
        """
        # Validate inputs
        if not occurrence_id:
            return ToolResult(
                success=False,
                message="ID da ocorrencia e obrigatorio.",
            )

        if new_status not in self.VALID_STATUSES:
            return ToolResult(
                success=False,
                message=f"Status invalido. Status validos: {', '.join(self.VALID_STATUSES)}",
            )

        try:
            # Prepare update data
            update_data = {
                "status": new_status,
                "updated_by": context.user_id,
            }
            if observacao:
                update_data["observacao"] = observacao

            # Call backend API
            response = await self.call_backend_api(
                method="PUT",
                endpoint=f"/api/v1/ocorrencias/{occurrence_id}/status",
                context=context,
                data=update_data,
            )

            logger.info(
                f"Occurrence status updated: {occurrence_id} -> {new_status} "
                f"by user {context.user_id}"
            )

            return ToolResult(
                success=True,
                data={
                    "occurrence_id": occurrence_id,
                    "previous_status": response.get("previous_status"),
                    "new_status": new_status,
                },
                message=f"Status da ocorrencia {occurrence_id} atualizado para '{new_status}'.",
            )

        except Exception as e:
            logger.error(f"Error updating occurrence status: {e}")
            raise ToolExecutionError(
                message="Erro ao atualizar status da ocorrencia.",
                details={"occurrence_id": occurrence_id, "error": str(e)}
            )

    def get_confirmation_details(
        self,
        context: ToolContext,
        occurrence_id: str = "",
        new_status: str = "",
        observacao: Optional[str] = None,
        **kwargs: Any,
    ) -> dict:
        """
        Get confirmation details for status update.

        Args:
            context: Tool execution context.
            occurrence_id: ID of the occurrence.
            new_status: New status to set.
            observacao: Optional observation.

        Returns:
            Dictionary with confirmation message and details.
        """
        return {
            "message": f"Deseja realmente alterar o status da ocorrencia {occurrence_id} para '{new_status}'?",
            "action": "Atualizar Status",
            "warning": "Esta acao sera registrada no historico da ocorrencia.",
            "details": {
                "occurrence_id": occurrence_id,
                "new_status": new_status,
                "observacao": observacao,
            },
        }


# Tool instances for registration
list_occurrences_tool = ListOccurrencesTool()
get_occurrence_details_tool = GetOccurrenceDetailsTool()
update_occurrence_status_tool = UpdateOccurrenceStatusTool()
