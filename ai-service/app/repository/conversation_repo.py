"""
AI Conversation Repository.

Data access layer for AI conversation history with tenant isolation.
"""

from datetime import datetime
from typing import Optional
from uuid import UUID

from sqlalchemy import delete, func, select
from sqlalchemy.ext.asyncio import AsyncSession

from app.models.conversation import AIConversation, MessageRole
from app.repository.base import BaseRepository


class ConversationRepository(BaseRepository[AIConversation]):
    """
    Repository for AI conversation history operations.

    Provides CRUD operations and queries for conversation messages
    with mandatory tenant isolation.
    """

    def __init__(self, session: AsyncSession):
        """Initialize repository with database session."""
        super().__init__(session, AIConversation)

    async def get_session_messages(
        self,
        session_id: UUID,
        tenant_id: UUID,
        limit: int = 50,
        offset: int = 0,
    ) -> list[AIConversation]:
        """
        Get all messages in a conversation session.

        Args:
            session_id: Conversation session UUID
            tenant_id: Tenant UUID for isolation
            limit: Maximum messages to return (default 50)
            offset: Number of messages to skip

        Returns:
            List of conversation messages ordered by creation time
        """
        stmt = (
            select(AIConversation)
            .where(
                AIConversation.session_id == session_id,
                AIConversation.tenant_id == tenant_id,
            )
            .order_by(AIConversation.created_at.asc())
            .limit(limit)
            .offset(offset)
        )
        result = await self.session.execute(stmt)
        return list(result.scalars().all())

    async def get_user_sessions(
        self,
        user_id: UUID,
        tenant_id: UUID,
        limit: int = 20,
        offset: int = 0,
    ) -> list[dict]:
        """
        Get unique conversation sessions for a user.

        Args:
            user_id: User UUID
            tenant_id: Tenant UUID for isolation
            limit: Maximum sessions to return
            offset: Number of sessions to skip

        Returns:
            List of session info dicts with session_id, last_message_at, message_count
        """
        # Subquery to get session stats
        stmt = (
            select(
                AIConversation.session_id,
                func.max(AIConversation.created_at).label("last_message_at"),
                func.count(AIConversation.id).label("message_count"),
            )
            .where(
                AIConversation.user_id == user_id,
                AIConversation.tenant_id == tenant_id,
            )
            .group_by(AIConversation.session_id)
            .order_by(func.max(AIConversation.created_at).desc())
            .limit(limit)
            .offset(offset)
        )
        result = await self.session.execute(stmt)
        rows = result.all()

        return [
            {
                "session_id": str(row.session_id),
                "last_message_at": row.last_message_at.isoformat(),
                "message_count": row.message_count,
            }
            for row in rows
        ]

    async def create_user_message(
        self,
        tenant_id: UUID,
        user_id: UUID,
        session_id: UUID,
        content: str,
        message_metadata: Optional[dict] = None,
    ) -> AIConversation:
        """
        Create a user message in the conversation.

        Args:
            tenant_id: Tenant UUID
            user_id: User UUID
            session_id: Session UUID
            content: Message content
            message_metadata: Optional message_metadata dict

        Returns:
            Created conversation message
        """
        message = AIConversation.create_user_message(
            tenant_id=tenant_id,
            user_id=user_id,
            session_id=session_id,
            content=content,
            message_metadata=message_metadata,
        )
        return await self.create(message)

    async def create_assistant_message(
        self,
        tenant_id: UUID,
        user_id: UUID,
        session_id: UUID,
        content: str,
        tool_calls: Optional[dict] = None,
        message_metadata: Optional[dict] = None,
    ) -> AIConversation:
        """
        Create an assistant response in the conversation.

        Args:
            tenant_id: Tenant UUID
            user_id: User UUID
            session_id: Session UUID
            content: Response content
            tool_calls: Optional tool calls made
            message_metadata: Optional message_metadata dict

        Returns:
            Created conversation message
        """
        message = AIConversation.create_assistant_message(
            tenant_id=tenant_id,
            user_id=user_id,
            session_id=session_id,
            content=content,
            tool_calls=tool_calls,
            message_metadata=message_metadata,
        )
        return await self.create(message)

    async def create_system_message(
        self,
        tenant_id: UUID,
        user_id: UUID,
        session_id: UUID,
        content: str,
        message_metadata: Optional[dict] = None,
    ) -> AIConversation:
        """
        Create a system message in the conversation.

        Args:
            tenant_id: Tenant UUID
            user_id: User UUID
            session_id: Session UUID
            content: System message content
            message_metadata: Optional message_metadata dict

        Returns:
            Created conversation message
        """
        message = AIConversation.create_system_message(
            tenant_id=tenant_id,
            user_id=user_id,
            session_id=session_id,
            content=content,
            message_metadata=message_metadata,
        )
        return await self.create(message)

    async def delete_session(
        self,
        session_id: UUID,
        tenant_id: UUID,
    ) -> int:
        """
        Delete all messages in a conversation session.

        Args:
            session_id: Session UUID to delete
            tenant_id: Tenant UUID for isolation

        Returns:
            Number of messages deleted
        """
        stmt = delete(AIConversation).where(
            AIConversation.session_id == session_id,
            AIConversation.tenant_id == tenant_id,
        )
        result = await self.session.execute(stmt)
        return result.rowcount

    async def count_user_messages(
        self,
        user_id: UUID,
        tenant_id: UUID,
        since: Optional[datetime] = None,
    ) -> int:
        """
        Count messages for a user (for rate limiting).

        Args:
            user_id: User UUID
            tenant_id: Tenant UUID for isolation
            since: Optional datetime to count from

        Returns:
            Number of messages
        """
        stmt = select(func.count(AIConversation.id)).where(
            AIConversation.user_id == user_id,
            AIConversation.tenant_id == tenant_id,
        )
        if since:
            stmt = stmt.where(AIConversation.created_at >= since)

        result = await self.session.execute(stmt)
        return result.scalar() or 0

    async def get_recent_context(
        self,
        session_id: UUID,
        tenant_id: UUID,
        limit: int = 10,
    ) -> list[dict]:
        """
        Get recent messages for context injection.

        Args:
            session_id: Session UUID
            tenant_id: Tenant UUID for isolation
            limit: Maximum messages to return

        Returns:
            List of message dicts with role and content
        """
        stmt = (
            select(AIConversation.role, AIConversation.content)
            .where(
                AIConversation.session_id == session_id,
                AIConversation.tenant_id == tenant_id,
            )
            .order_by(AIConversation.created_at.desc())
            .limit(limit)
        )
        result = await self.session.execute(stmt)
        rows = result.all()

        # Reverse to get chronological order
        return [
            {"role": row.role, "content": row.content}
            for row in reversed(rows)
        ]
