"""
Celery tasks for document indexing in the RAG pipeline.

Handles background document processing and vector store updates,
routed to the low-priority ai_indexing queue.
"""

import logging
from typing import Optional, Dict, Any, List
from datetime import datetime, timezone

from app.celery_app.celery_config import celery_app, QUEUE_AI_INDEXING
from app.celery_app.tasks.base import AuditedTask
from app.rag.ingestion import get_ingestion_pipeline, DocumentIngestionError
from app.rag.vector_store import get_vector_store_manager, VectorStoreError
from app.rag.doc_manager import DocumentManager, get_document_manager


logger = logging.getLogger(__name__)


@celery_app.task(
    bind=True,
    base=AuditedTask,
    name="app.celery_app.tasks.indexing.index_markdown_document",
    queue=QUEUE_AI_INDEXING,
    routing_key="ai.indexing",
)
def index_markdown_document(
    self,
    content: str,
    source_path: str,
    tenant_id: str,
    user_id: str,
    additional_metadata: Optional[Dict[str, Any]] = None,
) -> Dict[str, Any]:
    """
    Background task to index a markdown document.

    Args:
        content: Markdown content to index.
        source_path: Path or identifier for the document.
        tenant_id: Tenant ID for isolation.
        user_id: User ID who initiated the indexing.
        additional_metadata: Optional extra metadata.

    Returns:
        Dictionary with indexing results including node_ids.
    """
    logger.info(
        f"Starting markdown indexing: {source_path} "
        f"[tenant={tenant_id}, user={user_id}]"
    )

    try:
        pipeline = get_ingestion_pipeline()

        node_ids = pipeline.ingest_markdown(
            content=content,
            tenant_id=tenant_id,
            source_path=source_path,
            additional_metadata={
                **(additional_metadata or {}),
                "indexed_by": user_id,
            },
        )

        result = {
            "status": "success",
            "source_path": source_path,
            "tenant_id": tenant_id,
            "node_count": len(node_ids),
            "node_ids": node_ids,
            "indexed_at": datetime.now(timezone.utc).isoformat(),
        }

        logger.info(
            f"Completed markdown indexing: {source_path} "
            f"[nodes={len(node_ids)}]"
        )
        return result

    except DocumentIngestionError as e:
        logger.error(f"Ingestion error: {e.message} [source={e.source}]")
        raise
    except Exception as e:
        logger.error(f"Unexpected indexing error: {e}")
        raise


@celery_app.task(
    bind=True,
    base=AuditedTask,
    name="app.celery_app.tasks.indexing.index_pdf_document",
    queue=QUEUE_AI_INDEXING,
    routing_key="ai.indexing",
)
def index_pdf_document(
    self,
    file_content_base64: str,
    source_path: str,
    tenant_id: str,
    user_id: str,
    additional_metadata: Optional[Dict[str, Any]] = None,
) -> Dict[str, Any]:
    """
    Background task to index a PDF document.

    Args:
        file_content_base64: Base64 encoded PDF content.
        source_path: Path or identifier for the document.
        tenant_id: Tenant ID for isolation.
        user_id: User ID who initiated the indexing.
        additional_metadata: Optional extra metadata.

    Returns:
        Dictionary with indexing results including node_ids.
    """
    import base64

    logger.info(
        f"Starting PDF indexing: {source_path} "
        f"[tenant={tenant_id}, user={user_id}]"
    )

    try:
        file_content = base64.b64decode(file_content_base64)

        pipeline = get_ingestion_pipeline()

        node_ids = pipeline.ingest_pdf(
            file_content=file_content,
            tenant_id=tenant_id,
            source_path=source_path,
            additional_metadata={
                **(additional_metadata or {}),
                "indexed_by": user_id,
            },
        )

        result = {
            "status": "success",
            "source_path": source_path,
            "tenant_id": tenant_id,
            "node_count": len(node_ids),
            "node_ids": node_ids,
            "indexed_at": datetime.now(timezone.utc).isoformat(),
        }

        logger.info(
            f"Completed PDF indexing: {source_path} "
            f"[nodes={len(node_ids)}]"
        )
        return result

    except DocumentIngestionError as e:
        logger.error(f"PDF ingestion error: {e.message} [source={e.source}]")
        raise
    except Exception as e:
        logger.error(f"Unexpected PDF indexing error: {e}")
        raise


@celery_app.task(
    bind=True,
    base=AuditedTask,
    name="app.celery_app.tasks.indexing.index_text_document",
    queue=QUEUE_AI_INDEXING,
    routing_key="ai.indexing",
)
def index_text_document(
    self,
    content: str,
    source_path: str,
    tenant_id: str,
    user_id: str,
    doc_type: str = "text",
    additional_metadata: Optional[Dict[str, Any]] = None,
) -> Dict[str, Any]:
    """
    Background task to index a plain text document.

    Args:
        content: Text content to index.
        source_path: Path or identifier for the document.
        tenant_id: Tenant ID for isolation.
        user_id: User ID who initiated the indexing.
        doc_type: Document type label.
        additional_metadata: Optional extra metadata.

    Returns:
        Dictionary with indexing results including node_ids.
    """
    logger.info(
        f"Starting text indexing: {source_path} "
        f"[tenant={tenant_id}, user={user_id}]"
    )

    try:
        pipeline = get_ingestion_pipeline()

        node_ids = pipeline.ingest_text(
            content=content,
            tenant_id=tenant_id,
            source_path=source_path,
            doc_type=doc_type,
            additional_metadata={
                **(additional_metadata or {}),
                "indexed_by": user_id,
            },
        )

        result = {
            "status": "success",
            "source_path": source_path,
            "tenant_id": tenant_id,
            "doc_type": doc_type,
            "node_count": len(node_ids),
            "node_ids": node_ids,
            "indexed_at": datetime.now(timezone.utc).isoformat(),
        }

        logger.info(
            f"Completed text indexing: {source_path} "
            f"[nodes={len(node_ids)}]"
        )
        return result

    except DocumentIngestionError as e:
        logger.error(f"Text ingestion error: {e.message} [source={e.source}]")
        raise
    except Exception as e:
        logger.error(f"Unexpected text indexing error: {e}")
        raise


@celery_app.task(
    bind=True,
    base=AuditedTask,
    name="app.celery_app.tasks.indexing.reindex_document",
    queue=QUEUE_AI_INDEXING,
    routing_key="ai.indexing",
)
def reindex_document(
    self,
    document_id: str,
    tenant_id: str,
    user_id: str,
) -> Dict[str, Any]:
    """
    Background task to re-index an existing document.

    Deletes existing vectors and re-indexes from stored content.

    Args:
        document_id: ID of the document to re-index.
        tenant_id: Tenant ID for isolation.
        user_id: User ID who initiated the re-indexing.

    Returns:
        Dictionary with re-indexing results.
    """
    logger.info(
        f"Starting document re-indexing: {document_id} "
        f"[tenant={tenant_id}, user={user_id}]"
    )

    try:
        doc_manager = get_document_manager()

        doc_info = doc_manager.get_document(document_id, tenant_id)
        if not doc_info:
            return {
                "status": "error",
                "message": f"Document not found: {document_id}",
                "document_id": document_id,
            }

        vector_store = get_vector_store_manager()
        vector_store.delete_by_source(
            tenant_id=tenant_id,
            source_path=doc_info["source_path"],
        )

        pipeline = get_ingestion_pipeline()

        if doc_info["doc_type"] == "markdown":
            node_ids = pipeline.ingest_markdown(
                content=doc_info["content"],
                tenant_id=tenant_id,
                source_path=doc_info["source_path"],
                additional_metadata={"reindexed_by": user_id},
            )
        elif doc_info["doc_type"] == "pdf":
            import base64
            file_content = base64.b64decode(doc_info["content_base64"])
            node_ids = pipeline.ingest_pdf(
                file_content=file_content,
                tenant_id=tenant_id,
                source_path=doc_info["source_path"],
                additional_metadata={"reindexed_by": user_id},
            )
        else:
            node_ids = pipeline.ingest_text(
                content=doc_info["content"],
                tenant_id=tenant_id,
                source_path=doc_info["source_path"],
                doc_type=doc_info["doc_type"],
                additional_metadata={"reindexed_by": user_id},
            )

        result = {
            "status": "success",
            "document_id": document_id,
            "source_path": doc_info["source_path"],
            "tenant_id": tenant_id,
            "node_count": len(node_ids),
            "node_ids": node_ids,
            "reindexed_at": datetime.now(timezone.utc).isoformat(),
        }

        logger.info(
            f"Completed document re-indexing: {document_id} "
            f"[nodes={len(node_ids)}]"
        )
        return result

    except Exception as e:
        logger.error(f"Re-indexing error: {e}")
        raise


@celery_app.task(
    bind=True,
    base=AuditedTask,
    name="app.celery_app.tasks.indexing.delete_document_index",
    queue=QUEUE_AI_INDEXING,
    routing_key="ai.indexing",
)
def delete_document_index(
    self,
    source_path: str,
    tenant_id: str,
    user_id: str,
) -> Dict[str, Any]:
    """
    Background task to delete a document's index.

    Args:
        source_path: Path or identifier of the document.
        tenant_id: Tenant ID for isolation.
        user_id: User ID who initiated the deletion.

    Returns:
        Dictionary with deletion results.
    """
    logger.info(
        f"Starting index deletion: {source_path} "
        f"[tenant={tenant_id}, user={user_id}]"
    )

    try:
        vector_store = get_vector_store_manager()

        vector_store.delete_by_source(
            tenant_id=tenant_id,
            source_path=source_path,
        )

        result = {
            "status": "success",
            "source_path": source_path,
            "tenant_id": tenant_id,
            "deleted_at": datetime.now(timezone.utc).isoformat(),
            "deleted_by": user_id,
        }

        logger.info(f"Completed index deletion: {source_path}")
        return result

    except VectorStoreError as e:
        logger.error(f"Vector store error during deletion: {e.message}")
        raise
    except Exception as e:
        logger.error(f"Unexpected deletion error: {e}")
        raise


@celery_app.task(
    bind=True,
    base=AuditedTask,
    name="app.celery_app.tasks.indexing.batch_index_documents",
    queue=QUEUE_AI_INDEXING,
    routing_key="ai.indexing",
)
def batch_index_documents(
    self,
    documents: List[Dict[str, Any]],
    tenant_id: str,
    user_id: str,
) -> Dict[str, Any]:
    """
    Background task to index multiple documents in batch.

    Args:
        documents: List of document dictionaries with keys:
            - content: Document content
            - source_path: Path or identifier
            - doc_type: Document type (markdown, pdf, text)
        tenant_id: Tenant ID for isolation.
        user_id: User ID who initiated the batch indexing.

    Returns:
        Dictionary with batch indexing results.
    """
    logger.info(
        f"Starting batch indexing: {len(documents)} documents "
        f"[tenant={tenant_id}, user={user_id}]"
    )

    results = {
        "status": "success",
        "total": len(documents),
        "successful": 0,
        "failed": 0,
        "details": [],
    }

    pipeline = get_ingestion_pipeline()

    for doc in documents:
        try:
            doc_type = doc.get("doc_type", "text")
            source_path = doc.get("source_path", "unknown")
            content = doc.get("content", "")

            if doc_type == "markdown":
                node_ids = pipeline.ingest_markdown(
                    content=content,
                    tenant_id=tenant_id,
                    source_path=source_path,
                    additional_metadata={"indexed_by": user_id},
                )
            elif doc_type == "pdf":
                import base64
                file_content = base64.b64decode(content)
                node_ids = pipeline.ingest_pdf(
                    file_content=file_content,
                    tenant_id=tenant_id,
                    source_path=source_path,
                    additional_metadata={"indexed_by": user_id},
                )
            else:
                node_ids = pipeline.ingest_text(
                    content=content,
                    tenant_id=tenant_id,
                    source_path=source_path,
                    doc_type=doc_type,
                    additional_metadata={"indexed_by": user_id},
                )

            results["successful"] += 1
            results["details"].append({
                "source_path": source_path,
                "status": "success",
                "node_count": len(node_ids),
            })

        except Exception as e:
            results["failed"] += 1
            results["details"].append({
                "source_path": doc.get("source_path", "unknown"),
                "status": "failed",
                "error": str(e),
            })
            logger.error(f"Failed to index document: {e}")

    if results["failed"] > 0:
        results["status"] = "partial" if results["successful"] > 0 else "failed"

    results["indexed_at"] = datetime.now(timezone.utc).isoformat()

    logger.info(
        f"Completed batch indexing: {results['successful']}/{results['total']} successful"
    )
    return results
