"""
Vector Store configuration and management for RAG pipeline.

Manages Qdrant vector store connections with tenant-isolated collections
using metadata filtering for multi-tenant support.
"""

import logging
from typing import Optional, List, Any
from functools import lru_cache

from qdrant_client import QdrantClient
from qdrant_client.models import (
    Distance,
    VectorParams,
    Filter,
    FieldCondition,
    MatchValue,
    PointStruct,
    PayloadSchemaType,
)
from llama_index.core import VectorStoreIndex, StorageContext
from llama_index.core.schema import TextNode
from llama_index.vector_stores.qdrant import QdrantVectorStore
from llama_index.embeddings.huggingface import HuggingFaceEmbedding

from app.config import get_settings


logger = logging.getLogger(__name__)

# Default collection name for documents
DEFAULT_COLLECTION = "documents"

# Vector dimension for multilingual-e5-large embeddings
EMBEDDING_DIMENSION = 1024


class VectorStoreError(Exception):
    """Raised when vector store operations fail."""

    def __init__(self, message: str, operation: str = "unknown"):
        self.message = message
        self.operation = operation
        super().__init__(self.message)


class VectorStoreManager:
    """
    Manages Qdrant vector store connections and operations.

    Provides tenant isolation through metadata filtering and supports
    CRUD operations on vector collections.
    """

    def __init__(self):
        """Initialize vector store manager with configuration."""
        settings = get_settings()
        self.qdrant_url = settings.qdrant_url
        self.collection_prefix = settings.qdrant_collection_prefix
        self.embedding_model_name = settings.embedding_model
        self._client: Optional[QdrantClient] = None
        self._embedding_model: Optional[HuggingFaceEmbedding] = None

    @property
    def client(self) -> QdrantClient:
        """Get or create Qdrant client connection."""
        if self._client is None:
            self._client = self._create_client()
        return self._client

    @property
    def embedding_model(self) -> HuggingFaceEmbedding:
        """Get or create embedding model for Portuguese text."""
        if self._embedding_model is None:
            self._embedding_model = self._create_embedding_model()
        return self._embedding_model

    def _create_client(self) -> QdrantClient:
        """Create Qdrant client connection."""
        try:
            client = QdrantClient(url=self.qdrant_url, timeout=30)
            logger.info(f"Connected to Qdrant at {self.qdrant_url}")
            return client
        except Exception as e:
            raise VectorStoreError(
                f"Failed to connect to Qdrant: {e}",
                operation="connect",
            )

    def _create_embedding_model(self) -> HuggingFaceEmbedding:
        """"
        Create HuggingFace embedding model for Portuguese language support.

        Uses multilingual-e5-large for high-quality Portuguese embeddings.
        """
        try:
            model = HuggingFaceEmbedding(
                model_name=f"intfloat/{self.embedding_model_name}",
                trust_remote_code=True,
            )
            logger.info(f"Loaded embedding model: {self.embedding_model_name}")
            return model
        except Exception as e:
            raise VectorStoreError(
                f"Failed to load embedding model: {e}",
                operation="load_embedding",
            )

    def get_collection_name(self, name: str = DEFAULT_COLLECTION) -> str:
        """
        Get full collection name with prefix.

        Args:
            name: Base collection name.

        Returns:
            Full collection name with prefix.
        """
        return f"{self.collection_prefix}{name}"

    def ensure_collection_exists(
        self,
        collection_name: str = DEFAULT_COLLECTION,
        dimension: int = EMBEDDING_DIMENSION,
    ) -> bool:
        """
        Ensure collection exists, create if not.

        Args:
            collection_name: Name of the collection.
            dimension: Vector dimension size.

        Returns:
            True if collection exists or was created successfully.
        """
        full_name = self.get_collection_name(collection_name)

        try:
            collections = self.client.get_collections()
            existing_names = [c.name for c in collections.collections]

            if full_name in existing_names:
                logger.debug(f"Collection {full_name} already exists")
                return True

            self.client.create_collection(
                collection_name=full_name,
                vectors_config=VectorParams(
                    size=dimension,
                    distance=Distance.COSINE,
                ),
            )

            self.client.create_payload_index(
                collection_name=full_name,
                field_name="tenant_id",
                field_schema=PayloadSchemaType.KEYWORD,
            )

            logger.info(f"Created collection {full_name} with tenant_id index")
            return True

        except Exception as e:
            raise VectorStoreError(
                f"Failed to ensure collection {full_name}: {e}",
                operation="ensure_collection",
            )

    def create_tenant_filter(self, tenant_id: str) -> Filter:
        """
        Create a filter for tenant isolation.

        Args:
            tenant_id: The tenant ID to filter by.

        Returns:
            Qdrant Filter object for tenant isolation.
        """
        return Filter(
            must=[
                FieldCondition(
                    key="tenant_id",
                    match=MatchValue(value=tenant_id),
                )
            ]
        )

    def get_vector_store(
        self,
        collection_name: str = DEFAULT_COLLECTION,
        tenant_id: Optional[str] = None,
    ) -> QdrantVectorStore:
        """
        Get LlamaIndex QdrantVectorStore instance.

        Args:
            collection_name: Name of the collection.
            tenant_id: Optional tenant ID for filtering.

        Returns:
            QdrantVectorStore instance configured for the collection.
        """
        full_name = self.get_collection_name(collection_name)
        self.ensure_collection_exists(collection_name)

        return QdrantVectorStore(
            client=self.client,
            collection_name=full_name,
            enable_hybrid=True,
        )

    def get_retriever(
        self,
        tenant_id: str,
        collection_name: str = DEFAULT_COLLECTION,
        similarity_top_k: int = 5,
    ):
        """
        Get a retriever configured with tenant filter.

        Args:
            tenant_id: Tenant ID for filtering results.
            collection_name: Name of the collection.
            similarity_top_k: Number of results to retrieve.

        Returns:
            Configured retriever instance.
        """
        vector_store = self.get_vector_store(collection_name)

        index = VectorStoreIndex.from_vector_store(
            vector_store=vector_store,
            embed_model=self.embedding_model,
        )

        tenant_filter = self.create_tenant_filter(tenant_id)

        return index.as_retriever(
            similarity_top_k=similarity_top_k,
            vector_store_kwargs={
                "qdrant_filters": tenant_filter,
            },
        )

    def add_nodes(
        self,
        nodes: List[TextNode],
        collection_name: str = DEFAULT_COLLECTION,
    ) -> List[str]:
        """
        Add nodes to the vector store.

        Args:
            nodes: List of TextNode objects to add.
            collection_name: Name of the collection.

        Returns:
            List of node IDs that were added.
        """
        vector_store = self.get_vector_store(collection_name)

        storage_context = StorageContext.from_defaults(
            vector_store=vector_store,
        )

        index = VectorStoreIndex(
            nodes=nodes,
            storage_context=storage_context,
            embed_model=self.embedding_model,
        )

        node_ids = [node.node_id for node in nodes]
        logger.info(f"Added {len(nodes)} nodes to collection {collection_name}")
        return node_ids

    def delete_by_tenant(
        self,
        tenant_id: str,
        collection_name: str = DEFAULT_COLLECTION,
    ) -> int:
        """
        Delete all documents for a specific tenant.

        Args:
            tenant_id: Tenant ID whose documents should be deleted.
            collection_name: Name of the collection.

        Returns:
            Number of points deleted.
        """
        full_name = self.get_collection_name(collection_name)
        tenant_filter = self.create_tenant_filter(tenant_id)

        try:
            result = self.client.delete(
                collection_name=full_name,
                points_selector=tenant_filter,
            )
            logger.info(f"Deleted documents for tenant {tenant_id} from {full_name}")
            return result
        except Exception as e:
            raise VectorStoreError(
                f"Failed to delete tenant documents: {e}",
                operation="delete_by_tenant",
            )

    def delete_by_source(
        self,
        tenant_id: str,
        source_path: str,
        collection_name: str = DEFAULT_COLLECTION,
    ) -> int:
        """
        Delete documents by source path within a tenant.

        Args:
            tenant_id: Tenant ID for isolation.
            source_path: Source path of documents to delete.
            collection_name: Name of the collection.

        Returns:
            Number of points deleted.
        """
        full_name = self.get_collection_name(collection_name)

        source_filter = Filter(
            must=[
                FieldCondition(
                    key="tenant_id",
                    match=MatchValue(value=tenant_id),
                ),
                FieldCondition(
                    key="source_path",
                    match=MatchValue(value=source_path),
                ),
            ]
        )

        try:
            result = self.client.delete(
                collection_name=full_name,
                points_selector=source_filter,
            )
            logger.info(
                f"Deleted documents from {source_path} for tenant {tenant_id}"
            )
            return result
        except Exception as e:
            raise VectorStoreError(
                f"Failed to delete source documents: {e}",
                operation="delete_by_source",
            )

    def search(
        self,
        query_text: str,
        tenant_id: str,
        collection_name: str = DEFAULT_COLLECTION,
        limit: int = 5,
    ) -> List[dict]:
        """
        Search for similar documents with tenant filter.

        Args:
            query_text: Text to search for.
            tenant_id: Tenant ID for filtering.
            collection_name: Name of the collection.
            limit: Maximum number of results.

        Returns:
            List of search results with scores and metadata.
        """
        full_name = self.get_collection_name(collection_name)
        tenant_filter = self.create_tenant_filter(tenant_id)

        try:
            query_embedding = self.embedding_model.get_text_embedding(query_text)

            results = self.client.search(
                collection_name=full_name,
                query_vector=query_embedding,
                query_filter=tenant_filter,
                limit=limit,
            )

            return [
                {
                    "id": str(hit.id),
                    "score": hit.score,
                    "text": hit.payload.get("text", ""),
                    "metadata": {
                        k: v for k, v in hit.payload.items() if k != "text"
                    },
                }
                for hit in results
            ]

        except Exception as e:
            raise VectorStoreError(
                f"Search failed: {e}",
                operation="search",
            )

    def get_collection_info(
        self,
        collection_name: str = DEFAULT_COLLECTION,
    ) -> dict:
        """
        Get information about a collection.

        Args:
            collection_name: Name of the collection.

        Returns:
            Dictionary with collection statistics.
        """
        full_name = self.get_collection_name(collection_name)

        try:
            info = self.client.get_collection(collection_name=full_name)
            return {
                "name": full_name,
                "vectors_count": info.vectors_count,
                "points_count": info.points_count,
                "status": info.status.name,
            }
        except Exception as e:
            raise VectorStoreError(
                f"Failed to get collection info: {e}",
                operation="get_info",
            )

    def close(self) -> None:
        """Close the Qdrant client connection."""
        if self._client is not None:
            self._client.close()
            self._client = None
            logger.info("Closed Qdrant client connection")


@lru_cache()
def get_vector_store_manager() -> VectorStoreManager:
    """
    Get singleton VectorStoreManager instance.

    Returns:
        Cached VectorStoreManager instance.
    """
    return VectorStoreManager()
