"""
API Routers module.

Contains FastAPI router definitions for AI service endpoints.

Routers:
- chat: Chat endpoints for AI assistant interaction
- documents: Document management endpoints (admin only)
"""

from app.routers.chat import router as chat_router
from app.routers.documents import router as documents_router

__all__ = [
    "chat_router",
    "documents_router",
]
