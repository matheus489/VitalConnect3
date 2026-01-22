"""
Infrastructure tests for AI Service.

Tests for Task Group 1: Python Microservice Setup
- Health endpoint returns correct status
- Configuration loading from environment
- Redis connection initialization
- PostgreSQL connection initialization
"""

import os
from unittest.mock import AsyncMock, MagicMock, patch

import pytest
from fastapi.testclient import TestClient


class TestHealthEndpoint:
    """Test the health endpoint returns correct status."""

    def test_health_endpoint_returns_healthy_status(self):
        """
        Test that the health endpoint returns a healthy status.

        The health endpoint should return a JSON response with status 'healthy'
        or 'degraded' depending on connection states.
        """
        # Mock Redis and database connections before importing app
        with patch("app.database.redis_pool", None), \
             patch("app.database.db_engine", None):

            from app.main import create_app

            app = create_app()
            client = TestClient(app, raise_server_exceptions=False)

            response = client.get("/health")

            assert response.status_code == 200
            data = response.json()
            assert "status" in data
            assert data["status"] in ["healthy", "degraded"]
            assert data["service"] == "ai-service"
            assert data["version"] == "1.0.0"


class TestConfiguration:
    """Test configuration loading from environment variables."""

    def test_configuration_loads_defaults_in_development(self):
        """
        Test that configuration loads with sensible defaults in development.

        In development mode, the service should start with default values
        even when environment variables are not set.
        """
        # Clear relevant environment variables and set development mode
        env_vars_to_clear = [
            "DATABASE_URL", "REDIS_URL", "JWT_SECRET",
            "OPENAI_API_KEY", "QDRANT_URL"
        ]

        original_values = {}
        for var in env_vars_to_clear:
            original_values[var] = os.environ.get(var)
            if var in os.environ:
                del os.environ[var]

        os.environ["ENVIRONMENT"] = "development"

        try:
            # Clear the cache to reload settings
            from app.config import get_settings
            get_settings.cache_clear()

            settings = get_settings()

            assert settings.environment == "development"
            assert settings.debug is True
            assert settings.host == "0.0.0.0"
            assert settings.port == 8000
            assert settings.database_url is not None
            assert settings.redis_url is not None
            assert settings.jwt_secret == "dev-jwt-secret-change-in-production"
            assert settings.redis_key_prefix == "sidot:ai:"
            assert settings.redis_max_connections == 10

        finally:
            # Restore original environment
            for var, value in original_values.items():
                if value is not None:
                    os.environ[var] = value
                elif var in os.environ:
                    del os.environ[var]

            get_settings.cache_clear()

    def test_configuration_loads_from_environment(self):
        """
        Test that configuration correctly loads custom environment variables.
        """
        from app.config import get_settings

        # Set custom environment variables
        test_env = {
            "ENVIRONMENT": "testing",
            "DEBUG": "false",
            "HOST": "127.0.0.1",
            "PORT": "9000",
            "DATABASE_URL": "postgresql://test:test@testdb:5432/testdb",
            "REDIS_URL": "redis://testredis:6379/0",
            "JWT_SECRET": "test-secret-key",
            "REDIS_KEY_PREFIX": "test:ai:",
            "REDIS_MAX_CONNECTIONS": "20",
        }

        original_values = {}
        for var, value in test_env.items():
            original_values[var] = os.environ.get(var)
            os.environ[var] = value

        try:
            get_settings.cache_clear()
            settings = get_settings()

            assert settings.environment == "testing"
            assert settings.debug is False
            assert settings.host == "127.0.0.1"
            assert settings.port == 9000
            assert settings.database_url == "postgresql://test:test@testdb:5432/testdb"
            assert settings.redis_url == "redis://testredis:6379/0"
            assert settings.jwt_secret == "test-secret-key"
            assert settings.redis_key_prefix == "test:ai:"
            assert settings.redis_max_connections == 20

        finally:
            # Restore original environment
            for var, value in original_values.items():
                if value is not None:
                    os.environ[var] = value
                elif var in os.environ:
                    del os.environ[var]

            get_settings.cache_clear()


class TestRedisConnection:
    """Test Redis connection initialization."""

    @pytest.mark.asyncio
    async def test_redis_connection_initialization(self):
        """
        Test that Redis connection can be initialized and ping succeeds.

        This test mocks the Redis connection to verify the initialization
        flow works correctly without requiring a real Redis instance.
        """
        mock_redis_client = AsyncMock()
        mock_redis_client.ping = AsyncMock(return_value=True)

        mock_pool = MagicMock()
        mock_pool.disconnect = AsyncMock()

        with patch("redis.asyncio.ConnectionPool.from_url", return_value=mock_pool), \
             patch("redis.asyncio.Redis", return_value=mock_redis_client):

            from app.database import init_redis, close_redis

            # Initialize Redis
            client = await init_redis()

            # Verify ping was called
            mock_redis_client.ping.assert_called_once()

            # Cleanup
            await close_redis()
            mock_pool.disconnect.assert_called_once()


class TestPostgreSQLConnection:
    """Test PostgreSQL connection initialization."""

    @pytest.mark.asyncio
    async def test_postgresql_connection_initialization(self):
        """
        Test that PostgreSQL connection can be initialized.

        This test mocks the database engine to verify the initialization
        flow works correctly without requiring a real PostgreSQL instance.
        """
        # Create a mock async context manager for begin()
        mock_connection = AsyncMock()
        mock_connection.execute = AsyncMock(return_value=MagicMock())

        class MockAsyncContextManager:
            async def __aenter__(self):
                return mock_connection

            async def __aexit__(self, exc_type, exc_val, exc_tb):
                pass

        mock_engine = MagicMock()
        mock_engine.begin = MagicMock(return_value=MockAsyncContextManager())
        mock_engine.dispose = AsyncMock()

        # Patch at the module level where it is imported
        with patch("app.database.create_async_engine", return_value=mock_engine):
            # We need to reload the database module to use the patched import
            import importlib
            import app.database as db_module

            # Reset module state
            db_module.db_engine = None
            db_module.async_session_maker = None

            # Initialize database
            await db_module.init_database()

            # Verify engine was set
            assert db_module.db_engine is not None
            assert db_module.async_session_maker is not None

            # Cleanup
            await db_module.close_database()
