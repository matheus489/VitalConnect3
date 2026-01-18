"""
Database models module.

Contains SQLAlchemy models for AI conversation history and audit logs.
"""

from app.models.base import Base, UUIDMixin, TimestampMixin, TenantMixin, UserMixin
from app.models.conversation import AIConversation, MessageRole
from app.models.audit_log import (
    AIActionAuditLog,
    ActionType,
    ActionStatus,
    Severity,
)

__all__ = [
    # Base classes and mixins
    "Base",
    "UUIDMixin",
    "TimestampMixin",
    "TenantMixin",
    "UserMixin",
    # Conversation model
    "AIConversation",
    "MessageRole",
    # Audit log model
    "AIActionAuditLog",
    "ActionType",
    "ActionStatus",
    "Severity",
]
