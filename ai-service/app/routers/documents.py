"""
Document Management API Router for AI Service.

Provides admin endpoints for managing indexed documents:
- POST /api/v1/ai/documents/index - Trigger document indexing
- GET /api/v1/ai/documents - List indexed documents
- DELETE /api/v1/ai/documents/{id} - Remove document from index
"""

import logging
from typing import Optional
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, status
from pydantic import BaseModel, Field
from sqlalchemy.ext.asyncio import AsyncSession

from app.dependencies import require_roles, RequestContext
from app.database import get_db_session


logger = logging.getLogger(__name__)


# Request/Response Models
class IndexDocumentRequest(BaseModel):
    """Request body for document indexing."""

    document_path: Optional[str] = Field(
        None,
        description="Path to specific document or directory to index"
    )
    document_type: Optional[str] = Field(
        None,
        description="Type filter (markdown, pdf, all)"
    )
    force_reindex: bool = Field(
        False,
        description="Force reindexing even if document already indexed"
    )


class IndexDocumentResponse(BaseModel):
    """Response body for document indexing."""

    task_id: str
    status: str
    message: str
    documents_queued: int


class DocumentInfo(BaseModel):
    """Information about an indexed document."""

    id: str
    filename: str
    document_type: str
    indexed_at: str
    chunk_count: int
    metadata: dict = Field(default_factory=dict)


class DocumentListResponse(BaseModel):
    """Response body for listing documents."""

    documents: list[DocumentInfo]
    total: int
    page: int
    page_size: int


class DeleteDocumentResponse(BaseModel):
    """Response body for deleting a document."""

    id: str
    deleted: bool
    message: str


# Create router with admin-only access
router = APIRouter(prefix="/api/v1/ai/documents", tags=["AI Documents (Admin)"])


@router.post("/index", response_model=IndexDocumentResponse)
async def trigger_document_indexing(
    request: IndexDocumentRequest,
    request_ctx: RequestContext = Depends(require_roles("admin", "gestor")),
    db: AsyncSession = Depends(get_db_session),
) -> IndexDocumentResponse:
    """
    Trigger document indexing process.

    This endpoint queues documents for indexing in the vector store.
    Only admin and gestor roles can access this endpoint.

    Args:
        request: Indexing request with optional path and type filters.
        request_ctx: Request context with user and tenant info.
        db: Database session.

    Returns:
        IndexDocumentResponse with task information.
    """
    tenant_id = UUID(request_ctx.effective_tenant_id)

    try:
        # Queue indexing task via Celery
        from app.celery_app.tasks.indexing import index_documents_task

        task_result = index_documents_task.apply_async(
            kwargs={
                "tenant_id": str(tenant_id),
                "user_id": request_ctx.user_id,
                "document_path": request.document_path,
                "document_type": request.document_type,
                "force_reindex": request.force_reindex,
            },
            queue="ai_indexing",
        )

        return IndexDocumentResponse(
            task_id=task_result.id,
            status="queued",
            message="Processo de indexacao iniciado",
            documents_queued=1 if request.document_path else 0,
        )

    except ImportError:
        # Celery task not yet implemented - return placeholder
        logger.warning("Indexing task not yet implemented")
        return IndexDocumentResponse(
            task_id="pending-implementation",
            status="pending",
            message="Sistema de indexacao em implementacao",
            documents_queued=0,
        )

    except Exception as e:
        logger.error(f"Document indexing error: {e}", exc_info=True)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={
                "error": "indexing_failed",
                "message": "Erro ao iniciar processo de indexacao",
            }
        )


@router.get("", response_model=DocumentListResponse)
async def list_indexed_documents(
    page: int = 1,
    page_size: int = 20,
    document_type: Optional[str] = None,
    request_ctx: RequestContext = Depends(require_roles("admin", "gestor")),
    db: AsyncSession = Depends(get_db_session),
) -> DocumentListResponse:
    """
    List all indexed documents.

    Only admin and gestor roles can access this endpoint.

    Args:
        page: Page number (1-indexed).
        page_size: Number of documents per page.
        document_type: Optional filter by document type.
        request_ctx: Request context with user and tenant info.
        db: Database session.

    Returns:
        DocumentListResponse with paginated document list.
    """
    tenant_id = UUID(request_ctx.effective_tenant_id)

    try:
        # Try to get documents from document manager
        from app.rag.doc_manager import DocumentManager

        doc_manager = DocumentManager(tenant_id=str(tenant_id))

        documents = await doc_manager.list_documents(
            page=page,
            page_size=page_size,
            document_type=document_type,
        )

        return DocumentListResponse(
            documents=[
                DocumentInfo(
                    id=doc["id"],
                    filename=doc["filename"],
                    document_type=doc["document_type"],
                    indexed_at=doc["indexed_at"],
                    chunk_count=doc.get("chunk_count", 0),
                    metadata=doc.get("metadata", {}),
                )
                for doc in documents.get("items", [])
            ],
            total=documents.get("total", 0),
            page=page,
            page_size=page_size,
        )

    except ImportError:
        # Document manager not yet fully implemented
        logger.warning("Document manager not fully implemented")
        return DocumentListResponse(
            documents=[],
            total=0,
            page=page,
            page_size=page_size,
        )

    except Exception as e:
        logger.error(f"List documents error: {e}", exc_info=True)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={
                "error": "list_failed",
                "message": "Erro ao listar documentos",
            }
        )


@router.delete("/{document_id}", response_model=DeleteDocumentResponse)
async def delete_indexed_document(
    document_id: str,
    request_ctx: RequestContext = Depends(require_roles("admin", "gestor")),
    db: AsyncSession = Depends(get_db_session),
) -> DeleteDocumentResponse:
    """
    Remove a document from the index.

    Only admin and gestor roles can access this endpoint.

    Args:
        document_id: ID of the document to remove.
        request_ctx: Request context with user and tenant info.
        db: Database session.

    Returns:
        DeleteDocumentResponse with deletion result.
    """
    tenant_id = UUID(request_ctx.effective_tenant_id)

    try:
        # Try to delete document
        from app.rag.doc_manager import DocumentManager

        doc_manager = DocumentManager(tenant_id=str(tenant_id))

        deleted = await doc_manager.delete_document(document_id)

        return DeleteDocumentResponse(
            id=document_id,
            deleted=deleted,
            message="Documento removido do indice" if deleted
                    else "Documento nao encontrado",
        )

    except ImportError:
        # Document manager not yet fully implemented
        logger.warning("Document manager not fully implemented")
        return DeleteDocumentResponse(
            id=document_id,
            deleted=False,
            message="Sistema de gerenciamento de documentos em implementacao",
        )

    except Exception as e:
        logger.error(f"Delete document error: {e}", exc_info=True)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={
                "error": "delete_failed",
                "message": "Erro ao remover documento",
            }
        )
