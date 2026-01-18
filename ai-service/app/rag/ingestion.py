"""
Document ingestion pipeline for RAG system.

Supports markdown and PDF documents with chunking, metadata enrichment,
and multilingual embeddings for Portuguese language support.
"""

import logging
import uuid
from datetime import datetime, timezone
from pathlib import Path
from typing import List, Optional, Dict, Any, BinaryIO

from llama_index.core.schema import TextNode
from llama_index.core.node_parser import SentenceSplitter

from app.rag.vector_store import get_vector_store_manager, VectorStoreManager


logger = logging.getLogger(__name__)

# Default chunking parameters
DEFAULT_CHUNK_SIZE = 512
DEFAULT_CHUNK_OVERLAP = 50


class DocumentIngestionError(Exception):
    """Raised when document ingestion fails."""

    def __init__(self, message: str, source: str = "unknown"):
        self.message = message
        self.source = source
        super().__init__(self.message)


class DocumentIngestionPipeline:
    """
    Pipeline for ingesting documents into the vector store.

    Handles:
    - Markdown document parsing
    - PDF document extraction
    - Text chunking with overlap
    - Metadata enrichment
    - Vector store insertion
    """

    def __init__(
        self,
        chunk_size: int = DEFAULT_CHUNK_SIZE,
        chunk_overlap: int = DEFAULT_CHUNK_OVERLAP,
        vector_store_manager: Optional[VectorStoreManager] = None,
    ):
        """
        Initialize the ingestion pipeline.

        Args:
            chunk_size: Size of text chunks in characters.
            chunk_overlap: Overlap between consecutive chunks.
            vector_store_manager: Optional vector store manager instance.
        """
        self.chunk_size = chunk_size
        self.chunk_overlap = chunk_overlap
        self._vector_store_manager = vector_store_manager

        self.node_parser = SentenceSplitter(
            chunk_size=chunk_size,
            chunk_overlap=chunk_overlap,
            paragraph_separator="\n\n",
            secondary_chunking_regex="[^,.;?!]+[,.;?!]?",
        )

    @property
    def vector_store_manager(self) -> VectorStoreManager:
        """Get or create vector store manager."""
        if self._vector_store_manager is None:
            self._vector_store_manager = get_vector_store_manager()
        return self._vector_store_manager

    def create_nodes_from_text(
        self,
        content: str,
        metadata: Dict[str, Any],
    ) -> List[TextNode]:
        """
        Create nodes from text content with metadata.

        Args:
            content: Text content to chunk.
            metadata: Metadata to attach to each node (must include tenant_id).

        Returns:
            List of TextNode objects ready for indexing.
        """
        if not content or not content.strip():
            return []

        required_fields = ["tenant_id", "doc_type", "source_path"]
        for field in required_fields:
            if field not in metadata:
                raise DocumentIngestionError(
                    f"Missing required metadata field: {field}",
                    source=metadata.get("source_path", "unknown"),
                )

        enriched_metadata = {
            **metadata,
            "indexed_at": datetime.now(timezone.utc).isoformat(),
        }

        nodes = []
        chunks = self._chunk_text(content)

        for i, chunk in enumerate(chunks):
            node_id = str(uuid.uuid4())
            node = TextNode(
                id_=node_id,
                text=chunk,
                metadata={
                    **enriched_metadata,
                    "chunk_index": i,
                    "total_chunks": len(chunks),
                },
            )
            nodes.append(node)

        logger.debug(
            f"Created {len(nodes)} nodes from content "
            f"[source={metadata.get('source_path')}]"
        )
        return nodes

    def _chunk_text(self, text: str) -> List[str]:
        """
        Chunk text into smaller pieces with overlap.

        Args:
            text: Text to chunk.

        Returns:
            List of text chunks.
        """
        text = text.strip()
        if len(text) <= self.chunk_size:
            return [text]

        chunks = []
        start = 0

        while start < len(text):
            end = start + self.chunk_size

            if end < len(text):
                boundary = self._find_sentence_boundary(text, end)
                if boundary > start:
                    end = boundary

            chunk = text[start:end].strip()
            if chunk:
                chunks.append(chunk)

            start = end - self.chunk_overlap
            if start < 0:
                start = 0
            if start >= len(text):
                break

        return chunks if chunks else [text]

    def _find_sentence_boundary(self, text: str, position: int) -> int:
        """
        Find the nearest sentence boundary before position.

        Args:
            text: Full text content.
            position: Target position.

        Returns:
            Position of nearest sentence boundary.
        """
        search_start = max(0, position - 100)
        search_text = text[search_start:position]

        for marker in [".", "!", "?", "\n"]:
            last_pos = search_text.rfind(marker)
            if last_pos != -1:
                return search_start + last_pos + 1

        for marker in [",", ";", ":"]:
            last_pos = search_text.rfind(marker)
            if last_pos != -1:
                return search_start + last_pos + 1

        last_space = search_text.rfind(" ")
        if last_space != -1:
            return search_start + last_space + 1

        return position

    def ingest_markdown(
        self,
        content: str,
        tenant_id: str,
        source_path: str,
        additional_metadata: Optional[Dict[str, Any]] = None,
    ) -> List[str]:
        """
        Ingest a markdown document.

        Args:
            content: Markdown content.
            tenant_id: Tenant ID for isolation.
            source_path: Path or identifier for the source document.
            additional_metadata: Optional extra metadata fields.

        Returns:
            List of node IDs that were indexed.
        """
        metadata = {
            "tenant_id": tenant_id,
            "doc_type": "markdown",
            "source_path": source_path,
            **(additional_metadata or {}),
        }

        nodes = self.create_nodes_from_text(content, metadata)

        if not nodes:
            logger.warning(f"No nodes created from markdown: {source_path}")
            return []

        node_ids = self.vector_store_manager.add_nodes(nodes)
        logger.info(
            f"Ingested markdown document: {source_path} "
            f"[nodes={len(node_ids)}, tenant={tenant_id}]"
        )
        return node_ids

    def ingest_pdf(
        self,
        file_content: bytes,
        tenant_id: str,
        source_path: str,
        additional_metadata: Optional[Dict[str, Any]] = None,
    ) -> List[str]:
        """
        Ingest a PDF document.

        Args:
            file_content: PDF file content as bytes.
            tenant_id: Tenant ID for isolation.
            source_path: Path or identifier for the source document.
            additional_metadata: Optional extra metadata fields.

        Returns:
            List of node IDs that were indexed.
        """
        text = self._extract_pdf_text(file_content)

        if not text:
            raise DocumentIngestionError(
                "Failed to extract text from PDF",
                source=source_path,
            )

        metadata = {
            "tenant_id": tenant_id,
            "doc_type": "pdf",
            "source_path": source_path,
            **(additional_metadata or {}),
        }

        nodes = self.create_nodes_from_text(text, metadata)

        if not nodes:
            logger.warning(f"No nodes created from PDF: {source_path}")
            return []

        node_ids = self.vector_store_manager.add_nodes(nodes)
        logger.info(
            f"Ingested PDF document: {source_path} "
            f"[nodes={len(node_ids)}, tenant={tenant_id}]"
        )
        return node_ids

    def _extract_pdf_text(self, file_content: bytes) -> str:
        """
        Extract text content from PDF bytes.

        Args:
            file_content: PDF file content as bytes.

        Returns:
            Extracted text content.
        """
        try:
            import io
            from PyPDF2 import PdfReader

            reader = PdfReader(io.BytesIO(file_content))
            text_parts = []

            for page in reader.pages:
                page_text = page.extract_text()
                if page_text:
                    text_parts.append(page_text)

            return "\n\n".join(text_parts)

        except ImportError:
            logger.warning("PyPDF2 not available, attempting with pdfminer")
            return self._extract_pdf_with_pdfminer(file_content)
        except Exception as e:
            raise DocumentIngestionError(
                f"PDF extraction failed: {e}",
                source="pdf",
            )

    def _extract_pdf_with_pdfminer(self, file_content: bytes) -> str:
        """
        Fallback PDF extraction using pdfminer.

        Args:
            file_content: PDF file content as bytes.

        Returns:
            Extracted text content.
        """
        try:
            import io
            from pdfminer.high_level import extract_text

            return extract_text(io.BytesIO(file_content))
        except ImportError:
            raise DocumentIngestionError(
                "No PDF library available (PyPDF2 or pdfminer.six)",
                source="pdf",
            )
        except Exception as e:
            raise DocumentIngestionError(
                f"PDF extraction with pdfminer failed: {e}",
                source="pdf",
            )

    def ingest_text(
        self,
        content: str,
        tenant_id: str,
        source_path: str,
        doc_type: str = "text",
        additional_metadata: Optional[Dict[str, Any]] = None,
    ) -> List[str]:
        """
        Ingest a plain text document.

        Args:
            content: Text content.
            tenant_id: Tenant ID for isolation.
            source_path: Path or identifier for the source document.
            doc_type: Document type label.
            additional_metadata: Optional extra metadata fields.

        Returns:
            List of node IDs that were indexed.
        """
        metadata = {
            "tenant_id": tenant_id,
            "doc_type": doc_type,
            "source_path": source_path,
            **(additional_metadata or {}),
        }

        nodes = self.create_nodes_from_text(content, metadata)

        if not nodes:
            logger.warning(f"No nodes created from text: {source_path}")
            return []

        node_ids = self.vector_store_manager.add_nodes(nodes)
        logger.info(
            f"Ingested text document: {source_path} "
            f"[nodes={len(node_ids)}, tenant={tenant_id}]"
        )
        return node_ids

    def ingest_from_file(
        self,
        file_path: str,
        tenant_id: str,
        additional_metadata: Optional[Dict[str, Any]] = None,
    ) -> List[str]:
        """
        Ingest a document from file path.

        Automatically detects file type from extension.

        Args:
            file_path: Path to the file.
            tenant_id: Tenant ID for isolation.
            additional_metadata: Optional extra metadata fields.

        Returns:
            List of node IDs that were indexed.
        """
        path = Path(file_path)

        if not path.exists():
            raise DocumentIngestionError(
                f"File not found: {file_path}",
                source=file_path,
            )

        extension = path.suffix.lower()

        if extension == ".pdf":
            with open(path, "rb") as f:
                return self.ingest_pdf(
                    f.read(),
                    tenant_id,
                    str(path),
                    additional_metadata,
                )

        elif extension in [".md", ".markdown"]:
            with open(path, "r", encoding="utf-8") as f:
                return self.ingest_markdown(
                    f.read(),
                    tenant_id,
                    str(path),
                    additional_metadata,
                )

        elif extension in [".txt", ".text"]:
            with open(path, "r", encoding="utf-8") as f:
                return self.ingest_text(
                    f.read(),
                    tenant_id,
                    str(path),
                    doc_type="text",
                    additional_metadata=additional_metadata,
                )

        else:
            with open(path, "r", encoding="utf-8") as f:
                return self.ingest_text(
                    f.read(),
                    tenant_id,
                    str(path),
                    doc_type=extension.lstrip(".") or "unknown",
                    additional_metadata=additional_metadata,
                )

    def reindex_document(
        self,
        content: str,
        tenant_id: str,
        source_path: str,
        doc_type: str,
        additional_metadata: Optional[Dict[str, Any]] = None,
    ) -> List[str]:
        """
        Re-index a document by deleting existing and re-ingesting.

        Args:
            content: Document content.
            tenant_id: Tenant ID for isolation.
            source_path: Path or identifier for the source document.
            doc_type: Document type (markdown, pdf, text).
            additional_metadata: Optional extra metadata fields.

        Returns:
            List of new node IDs.
        """
        self.vector_store_manager.delete_by_source(
            tenant_id=tenant_id,
            source_path=source_path,
        )

        if doc_type == "markdown":
            return self.ingest_markdown(
                content,
                tenant_id,
                source_path,
                additional_metadata,
            )
        elif doc_type == "pdf":
            return self.ingest_pdf(
                content.encode() if isinstance(content, str) else content,
                tenant_id,
                source_path,
                additional_metadata,
            )
        else:
            return self.ingest_text(
                content,
                tenant_id,
                source_path,
                doc_type=doc_type,
                additional_metadata=additional_metadata,
            )


def get_ingestion_pipeline(
    chunk_size: int = DEFAULT_CHUNK_SIZE,
    chunk_overlap: int = DEFAULT_CHUNK_OVERLAP,
) -> DocumentIngestionPipeline:
    """
    Get a configured DocumentIngestionPipeline instance.

    Args:
        chunk_size: Size of text chunks.
        chunk_overlap: Overlap between chunks.

    Returns:
        Configured DocumentIngestionPipeline instance.
    """
    return DocumentIngestionPipeline(
        chunk_size=chunk_size,
        chunk_overlap=chunk_overlap,
    )
