"""
Base repository class.

Provides common CRUD operations with tenant isolation.
"""

from typing import Generic, Optional, Type, TypeVar
from uuid import UUID

from sqlalchemy import select, delete
from sqlalchemy.ext.asyncio import AsyncSession

from app.models.base import Base

ModelT = TypeVar("ModelT", bound=Base)


class BaseRepository(Generic[ModelT]):
    """
    Base repository with common CRUD operations.

    All queries are filtered by tenant_id to ensure multi-tenant isolation.
    """

    def __init__(self, session: AsyncSession, model: Type[ModelT]):
        """
        Initialize repository with database session and model class.

        Args:
            session: SQLAlchemy async session
            model: Model class for this repository
        """
        self.session = session
        self.model = model

    async def get_by_id(
        self,
        id: UUID,
        tenant_id: UUID,
    ) -> Optional[ModelT]:
        """
        Get a record by ID with tenant isolation.

        Args:
            id: Record UUID
            tenant_id: Tenant UUID for isolation

        Returns:
            Model instance or None if not found
        """
        stmt = select(self.model).where(
            self.model.id == id,
            self.model.tenant_id == tenant_id,
        )
        result = await self.session.execute(stmt)
        return result.scalar_one_or_none()

    async def create(self, instance: ModelT) -> ModelT:
        """
        Create a new record.

        Args:
            instance: Model instance to create

        Returns:
            Created model instance with generated ID
        """
        self.session.add(instance)
        await self.session.flush()
        await self.session.refresh(instance)
        return instance

    async def update(self, instance: ModelT) -> ModelT:
        """
        Update an existing record.

        Args:
            instance: Model instance to update

        Returns:
            Updated model instance
        """
        await self.session.flush()
        await self.session.refresh(instance)
        return instance

    async def delete(
        self,
        id: UUID,
        tenant_id: UUID,
    ) -> bool:
        """
        Delete a record by ID with tenant isolation.

        Args:
            id: Record UUID
            tenant_id: Tenant UUID for isolation

        Returns:
            True if deleted, False if not found
        """
        stmt = delete(self.model).where(
            self.model.id == id,
            self.model.tenant_id == tenant_id,
        )
        result = await self.session.execute(stmt)
        return result.rowcount > 0
