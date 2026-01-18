"""
Database and Redis Connection Management.

Provides connection pools and session factories for PostgreSQL and Redis.
This module should be imported instead of main.py to avoid circular imports.
"""

import logging
from typing import Any, AsyncGenerator

import redis.asyncio as redis
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
from sqlalchemy import text

from app.config import get_settings


logger = logging.getLogger(__name__)


# Global connection pools
redis_pool: redis.ConnectionPool | None = None
db_engine: Any = None
async_session_maker: sessionmaker | None = None


async def init_redis() -> redis.Redis:
    """Initialize Redis connection pool."""
    global redis_pool
    settings = get_settings()

    redis_pool = redis.ConnectionPool.from_url(
        settings.redis_url,
        max_connections=settings.redis_max_connections,
        decode_responses=True,
    )

    # Test connection
    client = redis.Redis(connection_pool=redis_pool)
    await client.ping()
    logger.info("Redis connection established successfully")

    return client


async def close_redis() -> None:
    """Close Redis connection pool."""
    global redis_pool
    if redis_pool:
        await redis_pool.disconnect()
        logger.info("Redis connection pool closed")


async def init_database() -> None:
    """Initialize database connection pool."""
    global db_engine, async_session_maker
    settings = get_settings()

    # Convert postgresql:// to postgresql+asyncpg:// for async support
    database_url = settings.database_url
    if database_url.startswith("postgresql://"):
        database_url = database_url.replace("postgresql://", "postgresql+asyncpg://", 1)

    db_engine = create_async_engine(
        database_url,
        pool_size=settings.database_pool_size,
        max_overflow=settings.database_max_overflow,
        echo=settings.debug,
    )

    async_session_maker = sessionmaker(
        db_engine,
        class_=AsyncSession,
        expire_on_commit=False,
    )

    # Test connection
    async with db_engine.begin() as conn:
        await conn.execute(text("SELECT 1"))

    logger.info("PostgreSQL connection established successfully")


async def close_database() -> None:
    """Close database connection pool."""
    global db_engine
    if db_engine:
        await db_engine.dispose()
        logger.info("PostgreSQL connection pool closed")


def get_redis_client() -> redis.Redis:
    """Get Redis client from the connection pool."""
    if not redis_pool:
        raise RuntimeError("Redis pool not initialized")
    return redis.Redis(connection_pool=redis_pool)


async def get_db_session() -> AsyncGenerator[AsyncSession, None]:
    """
    FastAPI dependency for database sessions.

    Yields:
        AsyncSession: Database session with automatic cleanup.
    """
    if not async_session_maker:
        raise RuntimeError("Database not initialized")

    async with async_session_maker() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise


__all__ = [
    "init_redis",
    "close_redis",
    "init_database",
    "close_database",
    "get_redis_client",
    "get_db_session",
    "redis_pool",
    "db_engine",
    "async_session_maker",
]
