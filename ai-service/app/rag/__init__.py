"""
RAG (Retrieval-Augmented Generation) module.

Contains vector store configuration, document ingestion pipeline,
query engine, and document management utilities.

Components:
- VectorStoreManager: Manages Qdrant vector store connections
- DocumentIngestionPipeline: Processes and indexes documents
- RAGQueryEngine: Handles queries with context retrieval
- DocumentManager: CRUD operations for indexed documents
"""

from app.rag.vector_store import VectorStoreManager
from app.rag.ingestion import DocumentIngestionPipeline
from app.rag.query_engine import RAGQueryEngine
from app.rag.doc_manager import DocumentManager

__all__ = [
    "VectorStoreManager",
    "DocumentIngestionPipeline",
    "RAGQueryEngine",
    "DocumentManager",
]
