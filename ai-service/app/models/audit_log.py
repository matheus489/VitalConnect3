"""
AI Action Audit Log Model.

SQLAlchemy model for auditing AI assistant actions and tool executions.
"""

from datetime import datetime
from enum import Enum
from typing import Any, Optional
from uuid import UUID

from sqlalchemy import ForeignKey, Index, Integer, String, Text
from sqlalchemy.dialects.postgresql import JSONB, UUID as PGUUID
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.models.base import Base, UUIDMixin, TimestampMixin, TenantMixin, UserMixin


class ActionType(str, Enum):
    """Enum for AI action types."""

    QUERY = "ai.query"
    TOOL_EXECUTION = "ai.tool_execution"
    CONFIRMATION = "ai.confirmation"


class ActionStatus(str, Enum):
    """Enum for AI action execution status."""

    PENDING = "pending"
    SUCCESS = "success"
    FAILED = "failed"
    CANCELLED = "cancelled"


class Severity(str, Enum):
    """Enum for audit log severity levels."""

    INFO = "INFO"
    WARN = "WARN"
    CRITICAL = "CRITICAL"


class AIActionAuditLog(Base, UUIDMixin, TenantMixin, UserMixin, TimestampMixin):
    """
    Model for AI action audit logging.

    Records all AI assistant actions including queries, tool executions,
    and human-in-the-loop confirmations for compliance and debugging.
    """

    __tablename__ = "ai_action_audit_log"

    # Reference to the conversation message that triggered this action
    conversation_id: Mapped[Optional[UUID]] = mapped_column(
        PGUUID(as_uuid=True),
        ForeignKey("ai_conversation_history.id", ondelete="SET NULL"),
        nullable=True,
        index=True,
    )

    # Type of action being audited
    action_type: Mapped[str] = mapped_column(
        String(100),
        nullable=False,
        index=True,
    )

    # Name of the tool executed (if applicable)
    tool_name: Mapped[Optional[str]] = mapped_column(
        String(100),
        nullable=True,
    )

    # Input parameters for the action
    input_params: Mapped[dict] = mapped_column(
        JSONB,
        nullable=False,
        default=dict,
    )

    # Output/result of the action
    output_result: Mapped[dict] = mapped_column(
        JSONB,
        nullable=False,
        default=dict,
    )

    # Execution status
    status: Mapped[str] = mapped_column(
        String(20),
        nullable=False,
        default=ActionStatus.PENDING.value,
        index=True,
    )

    # Execution time in milliseconds
    execution_time_ms: Mapped[Optional[int]] = mapped_column(
        Integer,
        nullable=True,
    )

    # Error message if failed
    error_message: Mapped[Optional[str]] = mapped_column(
        Text,
        nullable=True,
    )

    # Severity level for filtering and alerting
    severity: Mapped[str] = mapped_column(
        String(20),
        nullable=False,
        default=Severity.INFO.value,
    )

    # Relationship to conversation
    conversation = relationship(
        "AIConversation",
        back_populates="audit_logs",
    )

    # Table indexes for optimized queries
    __table_args__ = (
        # Composite index for tenant audit queries
        Index(
            "idx_ai_audit_tenant_time",
            "tenant_id",
            "created_at",
        ),
        # Composite index for tenant-user audit history
        Index(
            "idx_ai_audit_tenant_user_time",
            "tenant_id",
            "user_id",
            "created_at",
        ),
    )

    def __repr__(self) -> str:
        return (
            f"<AIActionAuditLog(id={self.id}, action_type={self.action_type}, "
            f"status={self.status}, tenant_id={self.tenant_id})>"
        )

    def to_dict(self) -> dict[str, Any]:
        """Convert model to dictionary representation."""
        return {
            "id": str(self.id),
            "tenant_id": str(self.tenant_id),
            "user_id": str(self.user_id),
            "conversation_id": str(self.conversation_id) if self.conversation_id else None,
            "action_type": self.action_type,
            "tool_name": self.tool_name,
            "input_params": self.input_params,
            "output_result": self.output_result,
            "status": self.status,
            "execution_time_ms": self.execution_time_ms,
            "error_message": self.error_message,
            "severity": self.severity,
            "created_at": self.created_at.isoformat() if self.created_at else None,
        }

    @classmethod
    def create_query_log(
        cls,
        tenant_id: UUID,
        user_id: UUID,
        conversation_id: Optional[UUID],
        input_params: dict,
        status: ActionStatus = ActionStatus.PENDING,
        output_result: Optional[dict] = None,
        execution_time_ms: Optional[int] = None,
        error_message: Optional[str] = None,
    ) -> "AIActionAuditLog":
        """Factory method to create a query audit log."""
        severity = Severity.INFO if status == ActionStatus.SUCCESS else Severity.WARN
        if status == ActionStatus.FAILED:
            severity = Severity.WARN

        return cls(
            tenant_id=tenant_id,
            user_id=user_id,
            conversation_id=conversation_id,
            action_type=ActionType.QUERY.value,
            input_params=input_params,
            output_result=output_result or {},
            status=status.value,
            execution_time_ms=execution_time_ms,
            error_message=error_message,
            severity=severity.value,
        )

    @classmethod
    def create_tool_execution_log(
        cls,
        tenant_id: UUID,
        user_id: UUID,
        conversation_id: Optional[UUID],
        tool_name: str,
        input_params: dict,
        status: ActionStatus = ActionStatus.PENDING,
        output_result: Optional[dict] = None,
        execution_time_ms: Optional[int] = None,
        error_message: Optional[str] = None,
        severity: Severity = Severity.INFO,
    ) -> "AIActionAuditLog":
        """Factory method to create a tool execution audit log."""
        return cls(
            tenant_id=tenant_id,
            user_id=user_id,
            conversation_id=conversation_id,
            action_type=ActionType.TOOL_EXECUTION.value,
            tool_name=tool_name,
            input_params=input_params,
            output_result=output_result or {},
            status=status.value,
            execution_time_ms=execution_time_ms,
            error_message=error_message,
            severity=severity.value,
        )

    @classmethod
    def create_confirmation_log(
        cls,
        tenant_id: UUID,
        user_id: UUID,
        conversation_id: Optional[UUID],
        tool_name: str,
        input_params: dict,
        confirmed: bool,
        execution_time_ms: Optional[int] = None,
    ) -> "AIActionAuditLog":
        """Factory method to create a confirmation audit log."""
        status = ActionStatus.SUCCESS if confirmed else ActionStatus.CANCELLED
        return cls(
            tenant_id=tenant_id,
            user_id=user_id,
            conversation_id=conversation_id,
            action_type=ActionType.CONFIRMATION.value,
            tool_name=tool_name,
            input_params=input_params,
            output_result={"confirmed": confirmed},
            status=status.value,
            execution_time_ms=execution_time_ms,
            severity=Severity.INFO.value,
        )

    def mark_success(
        self,
        output_result: dict,
        execution_time_ms: int,
    ) -> None:
        """Mark the action as successful."""
        self.status = ActionStatus.SUCCESS.value
        self.output_result = output_result
        self.execution_time_ms = execution_time_ms

    def mark_failed(
        self,
        error_message: str,
        execution_time_ms: Optional[int] = None,
    ) -> None:
        """Mark the action as failed."""
        self.status = ActionStatus.FAILED.value
        self.error_message = error_message
        self.execution_time_ms = execution_time_ms
        self.severity = Severity.WARN.value

    def mark_cancelled(self) -> None:
        """Mark the action as cancelled."""
        self.status = ActionStatus.CANCELLED.value
