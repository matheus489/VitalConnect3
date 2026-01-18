"""
Document Index Management for RAG pipeline.

Provides CRUD operations for indexed documents including
listing, retrieval, deletion, and re-indexing capabilities.
"""

import logging
import uuid
from datetime import datetime, timezone
from typing import Optional, List, Dict, Any
from functools import lru_cache

from app.rag.vector_store import VectorStoreManager, get_vector_store_manager
from app.rag.ingestion import DocumentIngestionPipeline, get_ingestion_pipeline


logger = logging.getLogger(__name__)


class DocumentNotFoundError(Exception):
    """Raised when a document is not found."""

    def __init__(self, document_id: str, tenant_id: str):
        self.document_id = document_id
        self.tenant_id = tenant_id
        self.message = f"Document {document_id} not found for tenant {tenant_id}"
        super().__init__(self.message)


class DocumentManagerError(Exception):
    """Raised when document management operations fail."""

    def __init__(self, message: str, operation: str = "unknown"):
        self.message = message
        self.operation = operation
        super().__init__(self.message)


class DocumentManager:
    """
    Manages indexed documents in the RAG system.

    Provides:
    - Listing indexed documents per tenant
    - Document metadata retrieval
    - Document deletion
    - Re-indexing capabilities
    - Index statistics
    """

    def __init__(
        self,
        vector_store_manager: Optional[VectorStoreManager] = None,
        ingestion_pipeline: Optional[DocumentIngestionPipeline] = None,
    ):
        """
        Initialize the document manager.

        Args:
            vector_store_manager: Optional vector store manager instance.
            ingestion_pipeline: Optional ingestion pipeline instance.
        """
        self._vector_store_manager = vector_store_manager
        self._ingestion_pipeline = ingestion_pipeline
        self._document_registry: Dict[str, Dict[str, Any]] = {}

    @property
    def vector_store_manager(self) -> VectorStoreManager:
        """Get or create vector store manager."""
        if self._vector_store_manager is None:
            self._vector_store_manager = get_vector_store_manager()
        return self._vector_store_manager

    @property
    def ingestion_pipeline(self) -> DocumentIngestionPipeline:
        """Get or create ingestion pipeline."""
        if self._ingestion_pipeline is None:
            self._ingestion_pipeline = get_ingestion_pipeline()
        return self._ingestion_pipeline

    def register_document(
        self,
        tenant_id: str,
        source_path: str,
        doc_type: str,
        content: Optional[str] = None,
        content_base64: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None,
    ) -> str:
        """
        Register a document in the local registry.

        This tracks documents for re-indexing and management purposes.

        Args:
            tenant_id: Tenant ID for isolation.
            source_path: Path or identifier for the document.
            doc_type: Document type (markdown, pdf, text).
            content: Optional text content for text documents.
            content_base64: Optional base64 content for binary documents.
            metadata: Optional additional metadata.

        Returns:
            Document ID.
        """
        document_id = str(uuid.uuid4())

        doc_info = {
            "id": document_id,
            "tenant_id": tenant_id,
            "source_path": source_path,
            "doc_type": doc_type,
            "content": content,
            "content_base64": content_base64,
            "metadata": metadata or {},
            "registered_at": datetime.now(timezone.utc).isoformat(),
            "last_indexed_at": None,
            "node_count": 0,
        }

        registry_key = f"{tenant_id}:{document_id}"
        self._document_registry[registry_key] = doc_info

        logger.info(
            f"Registered document: {source_path} [id={document_id}, tenant={tenant_id}]"
        )

        return document_id

    def get_document(
        self,
        document_id: str,
        tenant_id: str,
    ) -> Optional[Dict[str, Any]]:
        """
        Get document information by ID.

        Args:
            document_id: Document ID.
            tenant_id: Tenant ID for isolation.

        Returns:
            Document information dictionary or None if not found.
        """
        registry_key = f"{tenant_id}:{document_id}"
        return self._document_registry.get(registry_key)

    def list_documents(
        self,
        tenant_id: str,
        doc_type: Optional[str] = None,
        limit: int = 100,
        offset: int = 0,
    ) -> List[Dict[str, Any]]:
        """
        List documents for a tenant.

        Args:
            tenant_id: Tenant ID for isolation.
            doc_type: Optional filter by document type.
            limit: Maximum number of documents to return.
            offset: Number of documents to skip.

        Returns:
            List of document information dictionaries.
        """
        tenant_docs = []

        for key, doc in self._document_registry.items():
            if doc["tenant_id"] == tenant_id:
                if doc_type is None or doc["doc_type"] == doc_type:
                    doc_summary = {
                        "id": doc["id"],
                        "source_path": doc["source_path"],
                        "doc_type": doc["doc_type"],
                        "registered_at": doc["registered_at"],
                        "last_indexed_at": doc["last_indexed_at"],
                        "node_count": doc["node_count"],
                    }
                    tenant_docs.append(doc_summary)

        tenant_docs.sort(key=lambda x: x["registered_at"], reverse=True)

        return tenant_docs[offset:offset + limit]

    def update_index_status(
        self,
        document_id: str,
        tenant_id: str,
        node_count: int,
    ) -> None:
        """
        Update the index status of a document.

        Args:
            document_id: Document ID.
            tenant_id: Tenant ID for isolation.
            node_count: Number of nodes indexed.
        """
        registry_key = f"{tenant_id}:{document_id}"

        if registry_key in self._document_registry:
            self._document_registry[registry_key]["last_indexed_at"] = (
                datetime.now(timezone.utc).isoformat()
            )
            self._document_registry[registry_key]["node_count"] = node_count

    def delete_document(
        self,
        document_id: str,
        tenant_id: str,
    ) -> bool:
        """
        Delete a document and its vectors from the index.

        Args:
            document_id: Document ID.
            tenant_id: Tenant ID for isolation.

        Returns:
            True if document was deleted successfully.
        """
        registry_key = f"{tenant_id}:{document_id}"

        doc_info = self._document_registry.get(registry_key)
        if not doc_info:
            raise DocumentNotFoundError(document_id, tenant_id)

        try:
            self.vector_store_manager.delete_by_source(
                tenant_id=tenant_id,
                source_path=doc_info["source_path"],
            )

            del self._document_registry[registry_key]

            logger.info(
                f"Deleted document: {doc_info['source_path']} "
                f"[id={document_id}, tenant={tenant_id}]"
            )

            return True

        except Exception as e:
            raise DocumentManagerError(
                f"Failed to delete document: {e}",
                operation="delete",
            )

    def delete_by_source_path(
        self,
        source_path: str,
        tenant_id: str,
    ) -> bool:
        """
        Delete a document by its source path.

        Args:
            source_path: Path or identifier of the document.
            tenant_id: Tenant ID for isolation.

        Returns:
            True if document was deleted successfully.
        """
        try:
            self.vector_store_manager.delete_by_source(
                tenant_id=tenant_id,
                source_path=source_path,
            )

            for key, doc in list(self._document_registry.items()):
                if (doc["tenant_id"] == tenant_id and
                    doc["source_path"] == source_path):
                    del self._document_registry[key]
                    break

            logger.info(
                f"Deleted document by source: {source_path} [tenant={tenant_id}]"
            )

            return True

        except Exception as e:
            raise DocumentManagerError(
                f"Failed to delete document by source: {e}",
                operation="delete_by_source",
            )

    async def reindex_document(
        self,
        document_id: str,
        tenant_id: str,
    ) -> Dict[str, Any]:
        """
        Re-index a document from stored content.

        Args:
            document_id: Document ID.
            tenant_id: Tenant ID for isolation.

        Returns:
            Dictionary with re-indexing results.
        """
        doc_info = self.get_document(document_id, tenant_id)
        if not doc_info:
            raise DocumentNotFoundError(document_id, tenant_id)

        try:
            self.vector_store_manager.delete_by_source(
                tenant_id=tenant_id,
                source_path=doc_info["source_path"],
            )

            if doc_info["doc_type"] == "markdown":
                node_ids = self.ingestion_pipeline.ingest_markdown(
                    content=doc_info["content"],
                    tenant_id=tenant_id,
                    source_path=doc_info["source_path"],
                )
            elif doc_info["doc_type"] == "pdf":
                import base64
                file_content = base64.b64decode(doc_info["content_base64"])
                node_ids = self.ingestion_pipeline.ingest_pdf(
                    file_content=file_content,
                    tenant_id=tenant_id,
                    source_path=doc_info["source_path"],
                )
            else:
                node_ids = self.ingestion_pipeline.ingest_text(
                    content=doc_info["content"],
                    tenant_id=tenant_id,
                    source_path=doc_info["source_path"],
                    doc_type=doc_info["doc_type"],
                )

            self.update_index_status(document_id, tenant_id, len(node_ids))

            return {
                "status": "success",
                "document_id": document_id,
                "source_path": doc_info["source_path"],
                "node_count": len(node_ids),
                "reindexed_at": datetime.now(timezone.utc).isoformat(),
            }

        except Exception as e:
            raise DocumentManagerError(
                f"Re-indexing failed: {e}",
                operation="reindex",
            )

    async def reindex_all_for_tenant(
        self,
        tenant_id: str,
    ) -> Dict[str, Any]:
        """
        Re-index all documents for a tenant.

        Args:
            tenant_id: Tenant ID for isolation.

        Returns:
            Dictionary with batch re-indexing results.
        """
        documents = self.list_documents(tenant_id, limit=1000)

        results = {
            "total": len(documents),
            "successful": 0,
            "failed": 0,
            "details": [],
        }

        for doc in documents:
            try:
                result = await self.reindex_document(doc["id"], tenant_id)
                results["successful"] += 1
                results["details"].append({
                    "document_id": doc["id"],
                    "source_path": doc["source_path"],
                    "status": "success",
                })
            except Exception as e:
                results["failed"] += 1
                results["details"].append({
                    "document_id": doc["id"],
                    "source_path": doc["source_path"],
                    "status": "failed",
                    "error": str(e),
                })

        return results

    def get_index_stats(
        self,
        tenant_id: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        Get statistics about the document index.

        Args:
            tenant_id: Optional tenant ID to filter stats.

        Returns:
            Dictionary with index statistics.
        """
        try:
            collection_info = self.vector_store_manager.get_collection_info()

            tenant_docs = []
            if tenant_id:
                tenant_docs = self.list_documents(tenant_id, limit=10000)

            stats = {
                "collection": collection_info,
                "tenant_document_count": len(tenant_docs) if tenant_id else None,
                "total_registered_documents": len(self._document_registry),
            }

            if tenant_docs:
                by_type = {}
                total_nodes = 0
                for doc in tenant_docs:
                    doc_type = doc["doc_type"]
                    by_type[doc_type] = by_type.get(doc_type, 0) + 1
                    total_nodes += doc.get("node_count", 0)

                stats["tenant_stats"] = {
                    "by_type": by_type,
                    "total_nodes": total_nodes,
                }

            return stats

        except Exception as e:
            raise DocumentManagerError(
                f"Failed to get index stats: {e}",
                operation="get_stats",
            )

    def clear_tenant_index(
        self,
        tenant_id: str,
    ) -> Dict[str, Any]:
        """
        Clear all indexed documents for a tenant.

        Args:
            tenant_id: Tenant ID to clear.

        Returns:
            Dictionary with clearing results.
        """
        try:
            self.vector_store_manager.delete_by_tenant(tenant_id)

            removed_count = 0
            for key in list(self._document_registry.keys()):
                if self._document_registry[key]["tenant_id"] == tenant_id:
                    del self._document_registry[key]
                    removed_count += 1

            logger.info(
                f"Cleared tenant index: {tenant_id} "
                f"[documents_removed={removed_count}]"
            )

            return {
                "status": "success",
                "tenant_id": tenant_id,
                "documents_removed": removed_count,
                "cleared_at": datetime.now(timezone.utc).isoformat(),
            }

        except Exception as e:
            raise DocumentManagerError(
                f"Failed to clear tenant index: {e}",
                operation="clear_tenant",
            )


_document_manager_instance: Optional[DocumentManager] = None


def get_document_manager() -> DocumentManager:
    """
    Get singleton DocumentManager instance.

    Returns:
        DocumentManager instance.
    """
    global _document_manager_instance

    if _document_manager_instance is None:
        _document_manager_instance = DocumentManager()

    return _document_manager_instance
