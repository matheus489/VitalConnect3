"""
Report Tools for SIDOT AI Assistant.

Provides tools for generating reports about organ capture operations.

Tools:
- generate_report: Trigger report generation and return download URL
- search_documentation: RAG-based documentation search
"""

import logging
from datetime import datetime, timedelta
from typing import Any, Optional

from app.tools.base import (
    BaseTool,
    ToolContext,
    ToolResult,
    ToolExecutionError,
)


logger = logging.getLogger(__name__)


class GenerateReportTool(BaseTool):
    """
    Tool for generating operational reports.

    Triggers report generation via the Go backend and returns
    a download URL for the generated PDF.

    Permissions: gestor+ (gestor, admin)
    """

    name = "generate_report"
    description = """Gera relatorios operacionais em formato PDF.

    Parametros:
    - report_type: Tipo de relatorio (diario, semanal, mensal, customizado)
    - start_date: Data inicial (formato ISO: YYYY-MM-DD) - obrigatorio para customizado
    - end_date: Data final (formato ISO: YYYY-MM-DD) - obrigatorio para customizado
    - hospital_id: Filtrar por hospital (opcional)
    - include_sections: Secoes a incluir (resumo, ocorrencias, equipes, metricas) - lista

    Tipos de relatorio:
    - diario: Ultimas 24 horas
    - semanal: Ultimos 7 dias
    - mensal: Ultimos 30 dias
    - customizado: Periodo definido por start_date e end_date

    Retorna URL para download do PDF gerado."""

    requires_confirmation = False
    audit_severity = "INFO"

    VALID_REPORT_TYPES = ["diario", "semanal", "mensal", "customizado"]
    VALID_SECTIONS = ["resumo", "ocorrencias", "equipes", "metricas"]

    async def execute(
        self,
        context: ToolContext,
        report_type: str = "diario",
        start_date: Optional[str] = None,
        end_date: Optional[str] = None,
        hospital_id: Optional[str] = None,
        include_sections: Optional[list] = None,
        **kwargs: Any,
    ) -> ToolResult:
        """
        Execute report generation.

        Args:
            context: Tool execution context.
            report_type: Type of report to generate.
            start_date: Start date for custom report.
            end_date: End date for custom report.
            hospital_id: Filter by hospital ID.
            include_sections: List of sections to include.

        Returns:
            ToolResult with report download URL.
        """
        # Validate report type
        if report_type not in self.VALID_REPORT_TYPES:
            return ToolResult(
                success=False,
                message=f"Tipo de relatorio invalido. Tipos validos: {', '.join(self.VALID_REPORT_TYPES)}",
            )

        # Calculate date range based on report type
        now = datetime.now()
        if report_type == "diario":
            calculated_start = (now - timedelta(days=1)).strftime("%Y-%m-%d")
            calculated_end = now.strftime("%Y-%m-%d")
        elif report_type == "semanal":
            calculated_start = (now - timedelta(days=7)).strftime("%Y-%m-%d")
            calculated_end = now.strftime("%Y-%m-%d")
        elif report_type == "mensal":
            calculated_start = (now - timedelta(days=30)).strftime("%Y-%m-%d")
            calculated_end = now.strftime("%Y-%m-%d")
        elif report_type == "customizado":
            if not start_date or not end_date:
                return ToolResult(
                    success=False,
                    message="Para relatorio customizado, start_date e end_date sao obrigatorios.",
                )
            calculated_start = start_date
            calculated_end = end_date

        # Validate sections
        sections = include_sections or self.VALID_SECTIONS
        invalid_sections = [s for s in sections if s not in self.VALID_SECTIONS]
        if invalid_sections:
            return ToolResult(
                success=False,
                message=f"Secoes invalidas: {', '.join(invalid_sections)}. Validas: {', '.join(self.VALID_SECTIONS)}",
            )

        try:
            # Prepare report request
            report_data = {
                "report_type": report_type,
                "start_date": calculated_start,
                "end_date": calculated_end,
                "sections": sections,
                "format": "pdf",
                "requested_by": context.user_id,
            }

            if hospital_id:
                report_data["hospital_id"] = hospital_id

            # Call backend API to trigger report generation
            response = await self.call_backend_api(
                method="POST",
                endpoint="/api/v1/relatorios/gerar",
                context=context,
                data=report_data,
            )

            report_id = response.get("report_id")
            download_url = response.get("download_url")
            status = response.get("status", "processing")

            logger.info(
                f"Report generation requested: {report_id} "
                f"type={report_type} "
                f"user={context.user_id}"
            )

            # If report is ready immediately
            if status == "completed" and download_url:
                return ToolResult(
                    success=True,
                    data={
                        "report_id": report_id,
                        "download_url": download_url,
                        "status": "completed",
                        "report_type": report_type,
                        "period": {
                            "start": calculated_start,
                            "end": calculated_end,
                        },
                        "sections": sections,
                    },
                    message=f"Relatorio {report_type} gerado com sucesso. Clique para baixar.",
                )

            # Report is being generated asynchronously
            return ToolResult(
                success=True,
                data={
                    "report_id": report_id,
                    "status": "processing",
                    "report_type": report_type,
                    "period": {
                        "start": calculated_start,
                        "end": calculated_end,
                    },
                    "estimated_time": "1-2 minutos",
                },
                message=f"Relatorio {report_type} esta sendo gerado. Voce sera notificado quando estiver pronto.",
            )

        except Exception as e:
            logger.error(f"Error generating report: {e}")
            raise ToolExecutionError(
                message="Erro ao gerar relatorio.",
                details={"report_type": report_type, "error": str(e)}
            )


class SearchDocumentationTool(BaseTool):
    """
    Tool for searching documentation using RAG.

    Wrapper around the RAG query engine that returns relevant
    documentation chunks for the user's query.

    Permissions: All authenticated users
    """

    name = "search_documentation"
    description = """Pesquisa na base de conhecimento e documentacao do SIDOT.

    Parametros:
    - query: Pergunta ou termos de busca (obrigatorio)
    - doc_type: Filtrar por tipo de documento (procedimento, protocolo, manual, faq)
    - limit: Numero maximo de resultados (padrao: 5)

    Retorna trechos relevantes da documentacao que respondem a pergunta.
    Use para duvidas sobre procedimentos, protocolos e uso do sistema."""

    requires_confirmation = False
    audit_severity = "INFO"

    VALID_DOC_TYPES = ["procedimento", "protocolo", "manual", "faq"]

    async def execute(
        self,
        context: ToolContext,
        query: str,
        doc_type: Optional[str] = None,
        limit: int = 5,
        **kwargs: Any,
    ) -> ToolResult:
        """
        Execute documentation search using RAG.

        Args:
            context: Tool execution context.
            query: Search query string.
            doc_type: Filter by document type.
            limit: Maximum number of results.

        Returns:
            ToolResult with relevant documentation chunks.
        """
        if not query or not query.strip():
            return ToolResult(
                success=False,
                message="A consulta de busca nao pode estar vazia.",
            )

        if doc_type and doc_type not in self.VALID_DOC_TYPES:
            return ToolResult(
                success=False,
                message=f"Tipo de documento invalido. Tipos validos: {', '.join(self.VALID_DOC_TYPES)}",
            )

        try:
            # Try to use the RAG query engine if available
            results = await self._search_with_rag(context, query, doc_type, limit)

            if not results:
                return ToolResult(
                    success=True,
                    data={"results": [], "query": query},
                    message="Nenhum documento encontrado para sua busca. Tente reformular a pergunta.",
                )

            return ToolResult(
                success=True,
                data={
                    "results": results,
                    "query": query,
                    "total_results": len(results),
                },
                message=f"Encontrados {len(results)} documentos relevantes.",
            )

        except Exception as e:
            logger.error(f"Error searching documentation: {e}")
            # Fallback to backend API if RAG is not available
            return await self._search_via_backend(context, query, doc_type, limit)

    async def _search_with_rag(
        self,
        context: ToolContext,
        query: str,
        doc_type: Optional[str],
        limit: int,
    ) -> list:
        """
        Search documentation using the RAG query engine.

        Args:
            context: Tool execution context.
            query: Search query.
            doc_type: Document type filter.
            limit: Maximum results.

        Returns:
            List of relevant documentation chunks.
        """
        # Import RAG query engine (may not be available in all environments)
        try:
            from app.rag.query_engine import get_query_engine

            query_engine = get_query_engine(
                tenant_id=context.tenant_id,
                doc_type=doc_type,
            )

            # Execute query
            response = await query_engine.aquery(query)

            # Format results
            results = []
            for node in response.source_nodes[:limit]:
                results.append({
                    "content": node.text,
                    "score": node.score,
                    "metadata": {
                        "doc_type": node.metadata.get("doc_type"),
                        "title": node.metadata.get("title"),
                        "source": node.metadata.get("source_path"),
                    },
                })

            return results

        except ImportError:
            logger.warning("RAG query engine not available, using backend fallback")
            raise

    async def _search_via_backend(
        self,
        context: ToolContext,
        query: str,
        doc_type: Optional[str],
        limit: int,
    ) -> ToolResult:
        """
        Fallback search via Go backend API.

        Args:
            context: Tool execution context.
            query: Search query.
            doc_type: Document type filter.
            limit: Maximum results.

        Returns:
            ToolResult with search results.
        """
        try:
            params = {
                "q": query,
                "limit": limit,
            }
            if doc_type:
                params["type"] = doc_type

            response = await self.call_backend_api(
                method="GET",
                endpoint="/api/v1/documentacao/buscar",
                context=context,
                params=params,
            )

            results = response.get("data", [])

            return ToolResult(
                success=True,
                data={
                    "results": results,
                    "query": query,
                    "total_results": len(results),
                    "source": "backend_search",
                },
                message=f"Encontrados {len(results)} documentos relevantes.",
            )

        except Exception as e:
            logger.error(f"Error in backend documentation search: {e}")
            raise ToolExecutionError(
                message="Erro ao buscar documentacao.",
                details={"query": query, "error": str(e)}
            )


# Tool instances for registration
generate_report_tool = GenerateReportTool()
search_documentation_tool = SearchDocumentationTool()
