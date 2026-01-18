"""
AI Action Audit Log Repository.

Data access layer for AI action audit logs with tenant isolation.
"""

from datetime import datetime
from typing import Optional
from uuid import UUID

from sqlalchemy import func, select
from sqlalchemy.ext.asyncio import AsyncSession

from app.models.audit_log import (
    AIActionAuditLog,
    ActionStatus,
    ActionType,
    Severity,
)
from app.repository.base import BaseRepository


class AuditLogRepository(BaseRepository[AIActionAuditLog]):
    """
    Repository for AI action audit log operations.

    Provides CRUD operations and queries for audit logs
    with mandatory tenant isolation.
    """

    def __init__(self, session: AsyncSession):
        """Initialize repository with database session."""
        super().__init__(session, AIActionAuditLog)

    async def create_query_log(
        self,
        tenant_id: UUID,
        user_id: UUID,
        conversation_id: Optional[UUID],
        input_params: dict,
        status: ActionStatus = ActionStatus.PENDING,
        output_result: Optional[dict] = None,
        execution_time_ms: Optional[int] = None,
        error_message: Optional[str] = None,
    ) -> AIActionAuditLog:
        """
        Create an audit log for a query action.

        Args:
            tenant_id: Tenant UUID
            user_id: User UUID
            conversation_id: Related conversation message UUID
            input_params: Query input parameters
            status: Execution status
            output_result: Query result
            execution_time_ms: Execution time
            error_message: Error message if failed

        Returns:
            Created audit log entry
        """
        log = AIActionAuditLog.create_query_log(
            tenant_id=tenant_id,
            user_id=user_id,
            conversation_id=conversation_id,
            input_params=input_params,
            status=status,
            output_result=output_result,
            execution_time_ms=execution_time_ms,
            error_message=error_message,
        )
        return await self.create(log)

    async def create_tool_execution_log(
        self,
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
    ) -> AIActionAuditLog:
        """
        Create an audit log for a tool execution.

        Args:
            tenant_id: Tenant UUID
            user_id: User UUID
            conversation_id: Related conversation message UUID
            tool_name: Name of the tool executed
            input_params: Tool input parameters
            status: Execution status
            output_result: Tool execution result
            execution_time_ms: Execution time
            error_message: Error message if failed
            severity: Log severity level

        Returns:
            Created audit log entry
        """
        log = AIActionAuditLog.create_tool_execution_log(
            tenant_id=tenant_id,
            user_id=user_id,
            conversation_id=conversation_id,
            tool_name=tool_name,
            input_params=input_params,
            status=status,
            output_result=output_result,
            execution_time_ms=execution_time_ms,
            error_message=error_message,
            severity=severity,
        )
        return await self.create(log)

    async def create_confirmation_log(
        self,
        tenant_id: UUID,
        user_id: UUID,
        conversation_id: Optional[UUID],
        tool_name: str,
        input_params: dict,
        confirmed: bool,
        execution_time_ms: Optional[int] = None,
    ) -> AIActionAuditLog:
        """
        Create an audit log for a human-in-the-loop confirmation.

        Args:
            tenant_id: Tenant UUID
            user_id: User UUID
            conversation_id: Related conversation message UUID
            tool_name: Name of the tool being confirmed
            input_params: Action parameters that were confirmed/rejected
            confirmed: Whether the action was confirmed
            execution_time_ms: Time to confirmation

        Returns:
            Created audit log entry
        """
        log = AIActionAuditLog.create_confirmation_log(
            tenant_id=tenant_id,
            user_id=user_id,
            conversation_id=conversation_id,
            tool_name=tool_name,
            input_params=input_params,
            confirmed=confirmed,
            execution_time_ms=execution_time_ms,
        )
        return await self.create(log)

    async def get_by_conversation(
        self,
        conversation_id: UUID,
        tenant_id: UUID,
    ) -> list[AIActionAuditLog]:
        """
        Get all audit logs for a conversation message.

        Args:
            conversation_id: Conversation message UUID
            tenant_id: Tenant UUID for isolation

        Returns:
            List of audit logs
        """
        stmt = (
            select(AIActionAuditLog)
            .where(
                AIActionAuditLog.conversation_id == conversation_id,
                AIActionAuditLog.tenant_id == tenant_id,
            )
            .order_by(AIActionAuditLog.created_at.asc())
        )
        result = await self.session.execute(stmt)
        return list(result.scalars().all())

    async def get_pending_actions(
        self,
        user_id: UUID,
        tenant_id: UUID,
    ) -> list[AIActionAuditLog]:
        """
        Get pending actions awaiting confirmation.

        Args:
            user_id: User UUID
            tenant_id: Tenant UUID for isolation

        Returns:
            List of pending audit logs
        """
        stmt = (
            select(AIActionAuditLog)
            .where(
                AIActionAuditLog.user_id == user_id,
                AIActionAuditLog.tenant_id == tenant_id,
                AIActionAuditLog.status == ActionStatus.PENDING.value,
            )
            .order_by(AIActionAuditLog.created_at.desc())
        )
        result = await self.session.execute(stmt)
        return list(result.scalars().all())

    async def get_user_audit_history(
        self,
        user_id: UUID,
        tenant_id: UUID,
        action_type: Optional[ActionType] = None,
        status: Optional[ActionStatus] = None,
        since: Optional[datetime] = None,
        limit: int = 50,
        offset: int = 0,
    ) -> list[AIActionAuditLog]:
        """
        Get audit history for a user with filters.

        Args:
            user_id: User UUID
            tenant_id: Tenant UUID for isolation
            action_type: Filter by action type
            status: Filter by status
            since: Filter by creation date
            limit: Maximum records to return
            offset: Number of records to skip

        Returns:
            List of audit logs
        """
        stmt = select(AIActionAuditLog).where(
            AIActionAuditLog.user_id == user_id,
            AIActionAuditLog.tenant_id == tenant_id,
        )

        if action_type:
            stmt = stmt.where(AIActionAuditLog.action_type == action_type.value)

        if status:
            stmt = stmt.where(AIActionAuditLog.status == status.value)

        if since:
            stmt = stmt.where(AIActionAuditLog.created_at >= since)

        stmt = (
            stmt.order_by(AIActionAuditLog.created_at.desc())
            .limit(limit)
            .offset(offset)
        )

        result = await self.session.execute(stmt)
        return list(result.scalars().all())

    async def get_tenant_audit_summary(
        self,
        tenant_id: UUID,
        since: Optional[datetime] = None,
    ) -> dict:
        """
        Get audit summary statistics for a tenant.

        Args:
            tenant_id: Tenant UUID
            since: Optional datetime to count from

        Returns:
            Summary dict with counts by action_type and status
        """
        base_filter = [AIActionAuditLog.tenant_id == tenant_id]
        if since:
            base_filter.append(AIActionAuditLog.created_at >= since)

        # Count by action type
        action_stmt = (
            select(
                AIActionAuditLog.action_type,
                func.count(AIActionAuditLog.id).label("count"),
            )
            .where(*base_filter)
            .group_by(AIActionAuditLog.action_type)
        )
        action_result = await self.session.execute(action_stmt)
        action_counts = {row.action_type: row.count for row in action_result.all()}

        # Count by status
        status_stmt = (
            select(
                AIActionAuditLog.status,
                func.count(AIActionAuditLog.id).label("count"),
            )
            .where(*base_filter)
            .group_by(AIActionAuditLog.status)
        )
        status_result = await self.session.execute(status_stmt)
        status_counts = {row.status: row.count for row in status_result.all()}

        # Average execution time
        avg_time_stmt = (
            select(func.avg(AIActionAuditLog.execution_time_ms))
            .where(*base_filter)
            .where(AIActionAuditLog.execution_time_ms.isnot(None))
        )
        avg_result = await self.session.execute(avg_time_stmt)
        avg_time = avg_result.scalar()

        return {
            "by_action_type": action_counts,
            "by_status": status_counts,
            "avg_execution_time_ms": round(avg_time) if avg_time else None,
        }

    async def update_status(
        self,
        log_id: UUID,
        tenant_id: UUID,
        status: ActionStatus,
        output_result: Optional[dict] = None,
        execution_time_ms: Optional[int] = None,
        error_message: Optional[str] = None,
    ) -> Optional[AIActionAuditLog]:
        """
        Update the status of an audit log.

        Args:
            log_id: Audit log UUID
            tenant_id: Tenant UUID for isolation
            status: New status
            output_result: Optional output result
            execution_time_ms: Optional execution time
            error_message: Optional error message

        Returns:
            Updated audit log or None if not found
        """
        log = await self.get_by_id(log_id, tenant_id)
        if not log:
            return None

        log.status = status.value
        if output_result is not None:
            log.output_result = output_result
        if execution_time_ms is not None:
            log.execution_time_ms = execution_time_ms
        if error_message is not None:
            log.error_message = error_message
            log.severity = Severity.WARN.value

        return await self.update(log)
