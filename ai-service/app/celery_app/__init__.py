"""
Celery application module for SIDOT AI Service.

Contains Celery configuration, task queues, and task definitions.

Usage:
    from app.celery_app import celery_app

    # Run worker
    celery -A app.celery_app worker --loglevel=info --queues=ai_query,ai_actions,ai_indexing

Task Queues:
    - ai_query: High priority queue for user queries
    - ai_actions: Normal priority queue for tool executions
    - ai_indexing: Low priority queue for document indexing
"""

from app.celery_app.celery_config import (
    celery_app,
    QUEUE_AI_QUERY,
    QUEUE_AI_ACTIONS,
    QUEUE_AI_INDEXING,
    CELERY_QUEUES,
)

__all__ = [
    "celery_app",
    "QUEUE_AI_QUERY",
    "QUEUE_AI_ACTIONS",
    "QUEUE_AI_INDEXING",
    "CELERY_QUEUES",
]
