"""
Tests for RAG (Retrieval-Augmented Generation) functionality.

Tests cover:
- Document indexing
- Vector search with tenant filter
- Query response generation
"""

import os
import uuid
from unittest.mock import Mock, patch, MagicMock, AsyncMock
from datetime import datetime

import pytest


# Test fixtures
@pytest.fixture
def tenant_id():
    """Provide a test tenant ID."""
    return str(uuid.uuid4())


@pytest.fixture
def user_id():
    """Provide a test user ID."""
    return str(uuid.uuid4())


@pytest.fixture
def mock_settings():
    """Mock settings for testing."""
    with patch("app.config.get_settings") as mock:
        settings = Mock()
        settings.qdrant_url = "http://localhost:6333"
        settings.qdrant_collection_prefix = "test_"
        settings.embedding_model = "multilingual-e5-large"
        settings.openai_api_key = "test-key"
        settings.ai_model = "gpt-4o"
        mock.return_value = settings
        yield settings


class TestDocumentIndexing:
    """Tests for document indexing functionality."""

    def test_create_document_nodes_from_markdown(self, tenant_id):
        """Test creating document nodes from markdown content."""
        from app.rag.ingestion import DocumentIngestionPipeline

        pipeline = DocumentIngestionPipeline()

        markdown_content = """
        # Test Document

        This is a test document for VitalConnect.

        ## Section 1

        Content about procedures.

        ## Section 2

        More content about protocols.
        """

        metadata = {
            "tenant_id": tenant_id,
            "doc_type": "markdown",
            "source_path": "/docs/test.md",
        }

        nodes = pipeline.create_nodes_from_text(
            content=markdown_content,
            metadata=metadata,
        )

        assert len(nodes) > 0
        for node in nodes:
            assert node.metadata["tenant_id"] == tenant_id
            assert node.metadata["doc_type"] == "markdown"
            assert node.metadata["source_path"] == "/docs/test.md"

    def test_chunk_documents_with_overlap(self, tenant_id):
        """Test document chunking with proper overlap."""
        from app.rag.ingestion import DocumentIngestionPipeline

        pipeline = DocumentIngestionPipeline(
            chunk_size=100,
            chunk_overlap=20,
        )

        long_content = "A" * 50 + " " + "B" * 50 + " " + "C" * 50 + " " + "D" * 50

        metadata = {"tenant_id": tenant_id, "doc_type": "text", "source_path": "test.txt"}
        nodes = pipeline.create_nodes_from_text(content=long_content, metadata=metadata)

        assert len(nodes) >= 2
        assert all(node.metadata["tenant_id"] == tenant_id for node in nodes)

    def test_document_metadata_includes_required_fields(self, tenant_id):
        """Test that document metadata includes all required fields."""
        from app.rag.ingestion import DocumentIngestionPipeline

        pipeline = DocumentIngestionPipeline()

        content = "Test content for indexing."
        metadata = {
            "tenant_id": tenant_id,
            "doc_type": "pdf",
            "source_path": "/uploads/document.pdf",
        }

        nodes = pipeline.create_nodes_from_text(content=content, metadata=metadata)

        assert len(nodes) > 0
        node = nodes[0]

        assert "tenant_id" in node.metadata
        assert "doc_type" in node.metadata
        assert "source_path" in node.metadata
        assert "indexed_at" in node.metadata


class TestVectorSearchWithTenantFilter:
    """Tests for vector search with tenant isolation."""

    def test_vector_store_connection_config(self, mock_settings):
        """Test vector store connection configuration."""
        from app.rag.vector_store import VectorStoreManager

        manager = VectorStoreManager()

        assert manager.collection_prefix == mock_settings.qdrant_collection_prefix
        assert manager.qdrant_url == mock_settings.qdrant_url

    def test_tenant_filter_applied_to_search(self, tenant_id, mock_settings):
        """Test that tenant filter is applied to search queries."""
        from app.rag.vector_store import VectorStoreManager
        from qdrant_client.models import Filter, FieldCondition, MatchValue

        manager = VectorStoreManager()

        tenant_filter = manager.create_tenant_filter(tenant_id)

        assert tenant_filter is not None
        assert isinstance(tenant_filter, Filter)

    def test_collection_name_includes_prefix(self, mock_settings):
        """Test that collection name includes the configured prefix."""
        from app.rag.vector_store import VectorStoreManager

        manager = VectorStoreManager()

        collection_name = manager.get_collection_name("documents")

        assert collection_name.startswith(mock_settings.qdrant_collection_prefix)
        assert "documents" in collection_name


class TestQueryResponseGeneration:
    """Tests for RAG query response generation."""

    @pytest.mark.asyncio
    async def test_query_engine_creation(self, tenant_id, mock_settings):
        """Test RAG query engine creation with tenant context."""
        from app.rag.query_engine import RAGQueryEngine

        with patch("app.rag.query_engine.VectorStoreManager") as mock_manager:
            mock_manager.return_value.get_retriever.return_value = Mock()

            engine = RAGQueryEngine(tenant_id=tenant_id)

            assert engine.tenant_id == tenant_id

    @pytest.mark.asyncio
    async def test_query_with_context_retrieval(self, tenant_id, mock_settings):
        """Test query execution with context retrieval."""
        from app.rag.query_engine import RAGQueryEngine

        with patch("app.rag.query_engine.VectorStoreManager") as mock_vs, \
             patch("app.rag.query_engine.OpenAI") as mock_llm:

            mock_retriever = Mock()
            mock_retriever.retrieve.return_value = [
                Mock(text="Context 1", score=0.9),
                Mock(text="Context 2", score=0.8),
            ]
            mock_vs.return_value.get_retriever.return_value = mock_retriever

            mock_llm_instance = Mock()
            mock_llm_instance.complete.return_value = Mock(text="Generated response")
            mock_llm.return_value = mock_llm_instance

            engine = RAGQueryEngine(tenant_id=tenant_id)
            response = await engine.query("What are the procedures?")

            assert response is not None
            assert "response" in response or isinstance(response, str)

    def test_hybrid_search_configuration(self, tenant_id, mock_settings):
        """Test that hybrid search (vector + keyword) is configured."""
        from app.rag.query_engine import RAGQueryEngine

        with patch("app.rag.query_engine.VectorStoreManager"):
            engine = RAGQueryEngine(tenant_id=tenant_id)

            assert hasattr(engine, "enable_hybrid_search")
            assert engine.enable_hybrid_search is True
