"""
Pytest configuration and fixtures for AI Service tests.
"""

import os
import sys
from pathlib import Path

import pytest

# Add the app directory to the Python path
app_path = Path(__file__).parent.parent
sys.path.insert(0, str(app_path))

# Set default test environment
os.environ.setdefault("ENVIRONMENT", "testing")
os.environ.setdefault("DEBUG", "true")


@pytest.fixture(autouse=True)
def reset_settings_cache():
    """Reset settings cache before each test."""
    from app.config import get_settings
    get_settings.cache_clear()
    yield
    get_settings.cache_clear()


@pytest.fixture
def test_settings():
    """Provide test settings with defaults overridden for testing."""
    from app.config import get_settings

    # Set test environment variables
    os.environ["ENVIRONMENT"] = "testing"
    os.environ["DEBUG"] = "true"
    os.environ["DATABASE_URL"] = "postgresql://test:test@localhost:5432/test"
    os.environ["REDIS_URL"] = "redis://localhost:6379/1"
    os.environ["JWT_SECRET"] = "test-jwt-secret"

    get_settings.cache_clear()
    settings = get_settings()
    yield settings
    get_settings.cache_clear()
