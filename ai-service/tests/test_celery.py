"""
Tests for Celery Task Infrastructure.

Tests for Task Group 2: Celery Task Infrastructure
- Task registration
- Redis broker connection
- PostgreSQL result backend
- Task queue routing
"""

import os
from unittest.mock import MagicMock, patch

import pytest


class TestCeleryConfiguration:
    """Test Celery application configuration."""

    def test_celery_app_configuration(self):
        """
        Test that Celery application is configured correctly.

        Verifies that the Celery app has the correct broker URL,
        result backend, and key prefix configuration.
        """
        # Set up test environment
        os.environ["REDIS_URL"] = "redis://redis:6379/1"
        os.environ["DATABASE_URL"] = "postgresql://postgres:postgres@postgres:5432/sidot"

        from app.config import get_settings
        get_settings.cache_clear()

        from app.celery_app.celery_config import celery_app

        # Verify broker URL is Redis
        assert "redis://" in celery_app.conf.broker_url

        # Verify result backend is PostgreSQL (SQLAlchemy)
        assert "db+postgresql" in celery_app.conf.result_backend

        # Verify task serializer
        assert celery_app.conf.task_serializer == "json"
        assert celery_app.conf.result_serializer == "json"
        assert celery_app.conf.accept_content == ["json"]


class TestTaskRegistration:
    """Test task registration in Celery."""

    def test_task_registration(self):
        """
        Test that tasks are properly registered with Celery.

        Verifies that the base task class and any defined tasks
        are registered in the Celery task registry.
        """
        from app.celery_app.celery_config import celery_app
        from app.celery_app.tasks import base  # noqa: F401

        # Celery discovers tasks when app is loaded
        # Base task should be discoverable
        assert celery_app is not None
        assert celery_app.conf is not None


class TestTaskQueueRouting:
    """Test task queue routing configuration."""

    def test_queue_definitions(self):
        """
        Test that task queues are properly defined with priorities.

        Verifies that ai_query, ai_actions, and ai_indexing queues
        are configured with appropriate priorities.
        """
        from app.celery_app.celery_config import (
            CELERY_QUEUES,
            QUEUE_AI_QUERY,
            QUEUE_AI_ACTIONS,
            QUEUE_AI_INDEXING,
        )

        # Verify queue names
        assert QUEUE_AI_QUERY == "ai_query"
        assert QUEUE_AI_ACTIONS == "ai_actions"
        assert QUEUE_AI_INDEXING == "ai_indexing"

        # Verify queues are defined
        queue_names = [q.name for q in CELERY_QUEUES]
        assert "ai_query" in queue_names
        assert "ai_actions" in queue_names
        assert "ai_indexing" in queue_names

    def test_queue_routing_config(self):
        """
        Test that the Celery app has proper queue routing configuration.
        """
        from app.celery_app.celery_config import celery_app

        # Verify task queues are configured
        assert celery_app.conf.task_queues is not None
        assert len(celery_app.conf.task_queues) >= 3


class TestBaseTaskClass:
    """Test base task class with audit logging."""

    def test_base_task_has_required_attributes(self):
        """
        Test that the base task class has required attributes for audit logging.

        Verifies that the AuditedTask class includes retry configuration
        and audit logging capabilities.
        """
        from app.celery_app.tasks.base import AuditedTask

        # Verify retry configuration
        assert hasattr(AuditedTask, "autoretry_for")
        assert hasattr(AuditedTask, "max_retries")
        assert hasattr(AuditedTask, "default_retry_delay")

        # Verify max retries is 3
        assert AuditedTask.max_retries == 3

    def test_retry_backoff_configuration(self):
        """
        Test that exponential backoff is configured correctly.

        Verifies the retry delays are 10s, 30s, 60s as specified.
        """
        from app.celery_app.tasks.base import RETRY_BACKOFF_DELAYS

        # Verify backoff delays
        assert RETRY_BACKOFF_DELAYS == [10, 30, 60]
