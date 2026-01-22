"""
SIDOT AI Service - FastAPI Application Entry Point

A hybrid AI assistant (Q&A via RAG + Function Calling) that serves as an
operational co-pilot for SIDOT users.
"""

import logging
from contextlib import asynccontextmanager

import redis.asyncio as redis
from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from sqlalchemy import text

from app.config import get_settings, validate_settings
from app.database import (
    init_redis,
    close_redis,
    init_database,
    close_database,
    redis_pool,
    db_engine,
)

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """
    Application lifespan context manager.

    Handles startup and shutdown events for initializing and closing
    database and Redis connections.
    """
    settings = get_settings()

    # Validate settings on startup
    try:
        validate_settings(settings)
    except ValueError as e:
        logger.error(f"Configuration validation failed: {e}")
        if settings.environment == "production":
            raise

    logger.info(f"Starting AI Service in {settings.environment} mode")

    # Initialize connections
    try:
        await init_redis()
    except Exception as e:
        logger.warning(f"Failed to connect to Redis: {e}")
        if settings.environment == "production":
            raise

    try:
        await init_database()
    except Exception as e:
        logger.warning(f"Failed to connect to PostgreSQL: {e}")
        if settings.environment == "production":
            raise

    logger.info("AI Service started successfully")

    yield

    # Cleanup on shutdown
    logger.info("Shutting down AI Service")
    await close_redis()
    await close_database()
    logger.info("AI Service shutdown complete")


def create_app() -> FastAPI:
    """
    Create and configure the FastAPI application.

    Returns:
        Configured FastAPI application instance.
    """
    settings = get_settings()

    app = FastAPI(
        title="SIDOT AI Service",
        description="AI Assistant Co-Pilot for SIDOT - RAG + Function Calling",
        version="1.0.0",
        docs_url="/docs" if settings.debug else None,
        redoc_url="/redoc" if settings.debug else None,
        lifespan=lifespan,
    )

    # Configure CORS
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"] if settings.debug else [settings.go_backend_url],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    # Register API routers
    from app.routers import chat_router, documents_router

    app.include_router(chat_router)
    app.include_router(documents_router)

    # Register health and root routes
    @app.get("/health", tags=["Health"])
    async def health_check() -> dict:
        """
        Health check endpoint.

        Returns service health status and connection states.
        """
        health_status = {
            "status": "healthy",
            "service": "ai-service",
            "version": "1.0.0",
        }

        # Check Redis connection
        try:
            if redis_pool:
                client = redis.Redis(connection_pool=redis_pool)
                await client.ping()
                health_status["redis"] = "connected"
            else:
                health_status["redis"] = "not_initialized"
        except Exception as e:
            health_status["redis"] = f"error: {str(e)}"
            health_status["status"] = "degraded"

        # Check PostgreSQL connection
        try:
            if db_engine:
                async with db_engine.begin() as conn:
                    await conn.execute(text("SELECT 1"))
                health_status["postgres"] = "connected"
            else:
                health_status["postgres"] = "not_initialized"
        except Exception as e:
            health_status["postgres"] = f"error: {str(e)}"
            health_status["status"] = "degraded"

        return health_status

    @app.get("/", tags=["Root"])
    async def root() -> dict:
        """Root endpoint with service information."""
        return {
            "service": "SIDOT AI Service",
            "version": "1.0.0",
            "status": "running",
            "docs": "/docs" if settings.debug else "disabled",
        }

    # Global exception handler
    @app.exception_handler(Exception)
    async def global_exception_handler(request: Request, exc: Exception):
        """Handle unhandled exceptions globally."""
        logger.error(f"Unhandled exception: {exc}", exc_info=True)

        if settings.debug:
            return JSONResponse(
                status_code=500,
                content={
                    "error": "internal_server_error",
                    "detail": str(exc),
                }
            )

        return JSONResponse(
            status_code=500,
            content={
                "error": "internal_server_error",
                "detail": "An unexpected error occurred",
            }
        )

    return app


# Create the application instance
app = create_app()


if __name__ == "__main__":
    import uvicorn

    settings = get_settings()
    uvicorn.run(
        "app.main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.debug,
    )
