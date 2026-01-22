"""
Configuration module for AI Service.

Loads configuration from environment variables with sensible defaults for development.
"""

import os
from dataclasses import dataclass
from typing import Optional
from functools import lru_cache


@dataclass
class Settings:
    """Application settings loaded from environment variables."""

    # Environment
    environment: str
    debug: bool

    # Server
    host: str
    port: int

    # Database
    database_url: str
    database_pool_size: int
    database_max_overflow: int

    # Redis
    redis_url: str
    redis_key_prefix: str
    redis_max_connections: int
    redis_message_ttl: int

    # JWT Authentication (must match Go backend)
    jwt_secret: str
    jwt_algorithm: str

    # OpenAI / LLM
    openai_api_key: Optional[str]
    ai_model: str
    embedding_model: str

    # Vector Store (Qdrant)
    qdrant_url: str
    qdrant_collection_prefix: str

    # Go Backend Integration
    go_backend_url: str

    # Celery
    celery_broker_url: str
    celery_result_backend: str


def get_env(key: str, default: str = "") -> str:
    """Get environment variable with default value."""
    return os.getenv(key, default)


def get_env_bool(key: str, default: bool = False) -> bool:
    """Get boolean environment variable."""
    value = os.getenv(key, str(default)).lower()
    return value in ("true", "1", "yes", "on")


def get_env_int(key: str, default: int = 0) -> int:
    """Get integer environment variable with default value."""
    value = os.getenv(key, str(default))
    try:
        return int(value)
    except ValueError:
        return default


@lru_cache()
def get_settings() -> Settings:
    """
    Load and cache application settings from environment variables.

    Uses lru_cache to ensure settings are only loaded once.
    """
    environment = get_env("ENVIRONMENT", "development")
    is_dev = environment == "development"

    # Database URL for AI service
    database_url = get_env(
        "DATABASE_URL",
        "postgresql://postgres:postgres@postgres:5432/sidot"
    )

    # Redis URL (use database 1 for AI service to separate from main app)
    redis_url = get_env("REDIS_URL", "redis://redis:6379/1")

    # JWT secret (must match Go backend)
    jwt_secret = get_env("JWT_SECRET", "")
    if not jwt_secret and is_dev:
        jwt_secret = "dev-jwt-secret-change-in-production"

    return Settings(
        # Environment
        environment=environment,
        debug=get_env_bool("DEBUG", is_dev),

        # Server
        host=get_env("HOST", "0.0.0.0"),
        port=get_env_int("PORT", 8000),

        # Database
        database_url=database_url,
        database_pool_size=get_env_int("DATABASE_POOL_SIZE", 5),
        database_max_overflow=get_env_int("DATABASE_MAX_OVERFLOW", 10),

        # Redis
        redis_url=redis_url,
        redis_key_prefix=get_env("REDIS_KEY_PREFIX", "sidot:ai:"),
        redis_max_connections=get_env_int("REDIS_MAX_CONNECTIONS", 10),
        redis_message_ttl=get_env_int("REDIS_MESSAGE_TTL", 3600),  # 1 hour

        # JWT Authentication
        jwt_secret=jwt_secret,
        jwt_algorithm=get_env("JWT_ALGORITHM", "HS256"),

        # OpenAI / LLM
        openai_api_key=get_env("OPENAI_API_KEY") or None,
        ai_model=get_env("AI_MODEL", "gpt-4o"),
        embedding_model=get_env("EMBEDDING_MODEL", "multilingual-e5-large"),

        # Vector Store
        qdrant_url=get_env("QDRANT_URL", "http://qdrant:6333"),
        qdrant_collection_prefix=get_env("QDRANT_COLLECTION_PREFIX", "sidot_"),

        # Go Backend Integration
        go_backend_url=get_env("GO_BACKEND_URL", "http://backend:8080"),

        # Celery
        celery_broker_url=redis_url,
        celery_result_backend=database_url,
    )


def validate_settings(settings: Settings) -> None:
    """
    Validate that required settings are present in production.

    Raises:
        ValueError: If required production settings are missing.
    """
    if settings.environment == "production":
        errors = []

        if not settings.jwt_secret:
            errors.append("JWT_SECRET is required in production")

        if not settings.openai_api_key:
            errors.append("OPENAI_API_KEY is required in production")

        if not settings.database_url:
            errors.append("DATABASE_URL is required in production")

        if errors:
            raise ValueError("Configuration errors: " + "; ".join(errors))


# Export commonly used instances
settings = get_settings()
