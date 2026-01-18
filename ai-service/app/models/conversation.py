"""
AI Conversation History Model.

SQLAlchemy model for storing conversation messages between users and the AI assistant.
"""

from datetime import datetime
from enum import Enum
from typing import Any, Optional
from uuid import UUID

from sqlalchemy import ForeignKey, Index, String, Text
from sqlalchemy.dialects.postgresql import JSONB, UUID as PGUUID
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.models.base import Base, UUIDMixin, TimestampMixin, TenantMixin, UserMixin


class MessageRole(str, Enum):
    """Enum for conversation message roles."""

    USER = "user"
    ASSISTANT = "assistant"
    SYSTEM = "system"


class AIConversation(Base, UUIDMixin, TenantMixin, UserMixin, TimestampMixin):
    """
    Model for AI conversation history.

    Stores each message in a conversation session, including user prompts,
    assistant responses, and system messages.
    """

    __tablename__ = "ai_conversation_history"

    # Session identifier to group related messages
    session_id: Mapped[UUID] = mapped_column(
        PGUUID(as_uuid=True),
        nullable=False,
        index=True,
    )

    # Message role: user, assistant, or system
    role: Mapped[str] = mapped_column(
        String(20),
        nullable=False,
    )

    # Message content
    content: Mapped[str] = mapped_column(
        Text,
        nullable=False,
    )

    # Tool calls made by the assistant (if any)
    tool_calls: Mapped[Optional[dict]] = mapped_column(
        JSONB,
        nullable=True,
    )

    # Additional metadata (e.g., tokens used, model version)
    # Note: using 'message_metadata' as attribute name since 'metadata' is reserved by SQLAlchemy
    message_metadata: Mapped[Optional[dict]] = mapped_column(
        "metadata",  # Maps to 'metadata' column in database
        JSONB,
        nullable=True,
    )

    # Relationship to audit logs
    audit_logs = relationship(
        "AIActionAuditLog",
        back_populates="conversation",
        lazy="dynamic",
    )

    # Table indexes for optimized queries
    __table_args__ = (
        # Composite index for tenant-user queries
        Index(
            "idx_ai_conversation_tenant_user_time",
            "tenant_id",
            "user_id",
            "created_at",
        ),
        # Composite index for session timeline
        Index(
            "idx_ai_conversation_session_time",
            "session_id",
            "created_at",
        ),
    )

    def __repr__(self) -> str:
        return (
            f"<AIConversation(id={self.id}, session_id={self.session_id}, "
            f"role={self.role}, tenant_id={self.tenant_id})>"
        )

    def to_dict(self) -> dict[str, Any]:
        """Convert model to dictionary representation."""
        return {
            "id": str(self.id),
            "tenant_id": str(self.tenant_id),
            "user_id": str(self.user_id),
            "session_id": str(self.session_id),
            "role": self.role,
            "content": self.content,
            "tool_calls": self.tool_calls,
            "message_metadata": self.message_metadata,
            "created_at": self.created_at.isoformat() if self.created_at else None,
        }

    @classmethod
    def create_user_message(
        cls,
        tenant_id: UUID,
        user_id: UUID,
        session_id: UUID,
        content: str,
        message_metadata: Optional[dict] = None,
    ) -> "AIConversation":
        """Factory method to create a user message."""
        return cls(
            tenant_id=tenant_id,
            user_id=user_id,
            session_id=session_id,
            role=MessageRole.USER.value,
            content=content,
            message_metadata=message_metadata,
        )

    @classmethod
    def create_assistant_message(
        cls,
        tenant_id: UUID,
        user_id: UUID,
        session_id: UUID,
        content: str,
        tool_calls: Optional[dict] = None,
        message_metadata: Optional[dict] = None,
    ) -> "AIConversation":
        """Factory method to create an assistant message."""
        return cls(
            tenant_id=tenant_id,
            user_id=user_id,
            session_id=session_id,
            role=MessageRole.ASSISTANT.value,
            content=content,
            tool_calls=tool_calls,
            message_metadata=message_metadata,
        )

    @classmethod
    def create_system_message(
        cls,
        tenant_id: UUID,
        user_id: UUID,
        session_id: UUID,
        content: str,
        message_metadata: Optional[dict] = None,
    ) -> "AIConversation":
        """Factory method to create a system message."""
        return cls(
            tenant_id=tenant_id,
            user_id=user_id,
            session_id=session_id,
            role=MessageRole.SYSTEM.value,
            content=content,
            message_metadata=message_metadata,
        )
