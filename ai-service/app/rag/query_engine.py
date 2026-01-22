"""
RAG Query Engine for SIDOT AI Assistant.

Provides context-aware query processing with tenant isolation,
hybrid search (vector + keyword), and response synthesis.
"""

import logging
from typing import Optional, List, Dict, Any

from llama_index.core import PromptTemplate
from llama_index.core.schema import NodeWithScore
from llama_index.llms.openai import OpenAI

from app.config import get_settings
from app.rag.vector_store import VectorStoreManager, get_vector_store_manager


logger = logging.getLogger(__name__)

# System prompt for RAG responses in Portuguese
SYSTEM_PROMPT_PT = """Você é um assistente especializado do SIDOT, um sistema de gestão hospitalar para captação de órgãos.

Sua função é responder perguntas dos usuários com base no contexto fornecido.

Diretrizes:
- Responda sempre em português brasileiro
- Seja preciso e conciso nas respostas
- Se a informação não estiver no contexto, diga claramente que não encontrou a informação
- Cite as fontes quando relevante
- Nunca invente informações que não estejam no contexto
- Para procedimentos médicos, sempre recomende consultar o protocolo oficial

Contexto fornecido:
{context_str}

Pergunta do usuário: {query_str}

Resposta:"""

# Default retrieval parameters
DEFAULT_TOP_K = 5
DEFAULT_SIMILARITY_THRESHOLD = 0.7


class RAGQueryError(Exception):
    """Raised when RAG query processing fails."""

    def __init__(self, message: str, query: str = ""):
        self.message = message
        self.query = query
        super().__init__(self.message)


class RAGQueryEngine:
    """
    Query engine for RAG-based document retrieval and response generation.

    Features:
    - Tenant-isolated vector search
    - Hybrid search (vector + keyword)
    - Context-aware response synthesis
    - Portuguese language support
    """

    def __init__(
        self,
        tenant_id: str,
        vector_store_manager: Optional[VectorStoreManager] = None,
        top_k: int = DEFAULT_TOP_K,
        similarity_threshold: float = DEFAULT_SIMILARITY_THRESHOLD,
    ):
        """
        Initialize the RAG query engine.

        Args:
            tenant_id: Tenant ID for query isolation.
            vector_store_manager: Optional vector store manager instance.
            top_k: Number of documents to retrieve.
            similarity_threshold: Minimum similarity score for results.
        """
        self.tenant_id = tenant_id
        self.top_k = top_k
        self.similarity_threshold = similarity_threshold
        self.enable_hybrid_search = True

        self._vector_store_manager = vector_store_manager
        self._llm: Optional[OpenAI] = None
        self._prompt_template = PromptTemplate(SYSTEM_PROMPT_PT)

    @property
    def vector_store_manager(self) -> VectorStoreManager:
        """Get or create vector store manager."""
        if self._vector_store_manager is None:
            self._vector_store_manager = get_vector_store_manager()
        return self._vector_store_manager

    @property
    def llm(self) -> OpenAI:
        """Get or create LLM instance."""
        if self._llm is None:
            settings = get_settings()
            self._llm = OpenAI(
                model=settings.ai_model,
                api_key=settings.openai_api_key,
                temperature=0.1,
            )
        return self._llm

    async def query(
        self,
        query_text: str,
        additional_context: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        Execute a RAG query with context retrieval and response generation.

        Args:
            query_text: The user's question or query.
            additional_context: Optional additional context to include.

        Returns:
            Dictionary with response, sources, and metadata.
        """
        if not query_text or not query_text.strip():
            raise RAGQueryError("Query text cannot be empty", query_text)

        logger.info(
            f"Processing RAG query [tenant={self.tenant_id}]: "
            f"{query_text[:100]}..."
        )

        try:
            retrieved_nodes = await self._retrieve_context(query_text)

            if not retrieved_nodes:
                return {
                    "response": "Não encontrei informações relevantes na base de conhecimento para responder sua pergunta.",
                    "sources": [],
                    "has_context": False,
                    "query": query_text,
                }

            context_str = self._format_context(retrieved_nodes, additional_context)

            response_text = await self._generate_response(query_text, context_str)

            sources = self._extract_sources(retrieved_nodes)

            return {
                "response": response_text,
                "sources": sources,
                "has_context": True,
                "query": query_text,
                "retrieved_count": len(retrieved_nodes),
            }

        except Exception as e:
            logger.error(f"RAG query failed: {e}")
            raise RAGQueryError(f"Query processing failed: {e}", query_text)

    async def _retrieve_context(
        self,
        query_text: str,
    ) -> List[NodeWithScore]:
        """
        Retrieve relevant context from vector store.

        Args:
            query_text: The query text for retrieval.

        Returns:
            List of retrieved nodes with scores.
        """
        try:
            retriever = self.vector_store_manager.get_retriever(
                tenant_id=self.tenant_id,
                similarity_top_k=self.top_k,
            )

            nodes = retriever.retrieve(query_text)

            filtered_nodes = [
                node for node in nodes
                if node.score >= self.similarity_threshold
            ]

            logger.debug(
                f"Retrieved {len(filtered_nodes)} nodes "
                f"(filtered from {len(nodes)})"
            )

            return filtered_nodes

        except Exception as e:
            logger.error(f"Context retrieval failed: {e}")
            return []

    def _format_context(
        self,
        nodes: List[NodeWithScore],
        additional_context: Optional[str] = None,
    ) -> str:
        """
        Format retrieved nodes into a context string.

        Args:
            nodes: List of retrieved nodes.
            additional_context: Optional additional context.

        Returns:
            Formatted context string.
        """
        context_parts = []

        for i, node in enumerate(nodes, 1):
            source = node.metadata.get("source_path", "documento")
            text = node.text.strip()

            context_parts.append(
                f"[Fonte {i}: {source}]\n{text}\n"
            )

        if additional_context:
            context_parts.append(
                f"[Contexto adicional]\n{additional_context}\n"
            )

        return "\n---\n".join(context_parts)

    async def _generate_response(
        self,
        query_text: str,
        context_str: str,
    ) -> str:
        """
        Generate response using LLM with context.

        Args:
            query_text: The user's question.
            context_str: Formatted context string.

        Returns:
            Generated response text.
        """
        try:
            formatted_prompt = self._prompt_template.format(
                context_str=context_str,
                query_str=query_text,
            )

            response = self.llm.complete(formatted_prompt)

            return response.text.strip()

        except Exception as e:
            logger.error(f"Response generation failed: {e}")
            return (
                "Desculpe, ocorreu um erro ao processar sua pergunta. "
                "Por favor, tente novamente."
            )

    def _extract_sources(
        self,
        nodes: List[NodeWithScore],
    ) -> List[Dict[str, Any]]:
        """
        Extract source information from retrieved nodes.

        Args:
            nodes: List of retrieved nodes.

        Returns:
            List of source dictionaries with metadata.
        """
        sources = []

        for node in nodes:
            source = {
                "source_path": node.metadata.get("source_path", "unknown"),
                "doc_type": node.metadata.get("doc_type", "unknown"),
                "score": round(node.score, 3),
                "chunk_index": node.metadata.get("chunk_index"),
                "indexed_at": node.metadata.get("indexed_at"),
            }
            sources.append(source)

        return sources

    async def search(
        self,
        query_text: str,
        limit: Optional[int] = None,
    ) -> List[Dict[str, Any]]:
        """
        Search for relevant documents without generating a response.

        Args:
            query_text: The search query.
            limit: Maximum number of results.

        Returns:
            List of search results with text and metadata.
        """
        effective_limit = limit or self.top_k

        try:
            results = self.vector_store_manager.search(
                query_text=query_text,
                tenant_id=self.tenant_id,
                limit=effective_limit,
            )

            return results

        except Exception as e:
            logger.error(f"Search failed: {e}")
            raise RAGQueryError(f"Search failed: {e}", query_text)

    def get_retriever(self):
        """
        Get the underlying retriever for direct use.

        Returns:
            Configured retriever instance.
        """
        return self.vector_store_manager.get_retriever(
            tenant_id=self.tenant_id,
            similarity_top_k=self.top_k,
        )


def create_query_engine(
    tenant_id: str,
    top_k: int = DEFAULT_TOP_K,
    similarity_threshold: float = DEFAULT_SIMILARITY_THRESHOLD,
) -> RAGQueryEngine:
    """
    Create a RAG query engine for a specific tenant.

    Args:
        tenant_id: Tenant ID for query isolation.
        top_k: Number of documents to retrieve.
        similarity_threshold: Minimum similarity score.

    Returns:
        Configured RAGQueryEngine instance.
    """
    return RAGQueryEngine(
        tenant_id=tenant_id,
        top_k=top_k,
        similarity_threshold=similarity_threshold,
    )
