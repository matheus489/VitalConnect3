"""
Celery tasks package for VitalConnect AI Service.

Contains task definitions organized by functionality:
- base.py: Base task class with audit logging
- query.py: User query processing tasks (high priority)
- actions.py: Tool execution tasks (normal priority)
- indexing.py: Document indexing tasks (low priority)
"""

from app.celery_app.tasks.base import AuditedTask
from app.celery_app.tasks.query import (
    process_chat_message,
    process_rag_query,
    execute_tool,
)

__all__ = [
    "AuditedTask",
    "process_chat_message",
    "process_rag_query",
    "execute_tool",
]
