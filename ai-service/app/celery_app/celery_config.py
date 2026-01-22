"""
Celery application configuration for SIDOT AI Service.

Configures Celery with:
- Redis as message broker
- PostgreSQL as result backend
- Task queues with priorities
- Retry logic with exponential backoff
"""

from kombu import Exchange, Queue
from celery import Celery

from app.config import get_settings


# Queue names
QUEUE_AI_QUERY = "ai_query"
QUEUE_AI_ACTIONS = "ai_actions"
QUEUE_AI_INDEXING = "ai_indexing"

# Default exchange
default_exchange = Exchange("sidot_ai", type="direct")

# Define task queues with priorities
# Higher x-max-priority = higher priority support
CELERY_QUEUES = (
    Queue(
        QUEUE_AI_QUERY,
        exchange=default_exchange,
        routing_key="ai.query",
        queue_arguments={"x-max-priority": 10},  # High priority
    ),
    Queue(
        QUEUE_AI_ACTIONS,
        exchange=default_exchange,
        routing_key="ai.actions",
        queue_arguments={"x-max-priority": 5},  # Normal priority
    ),
    Queue(
        QUEUE_AI_INDEXING,
        exchange=default_exchange,
        routing_key="ai.indexing",
        queue_arguments={"x-max-priority": 1},  # Low priority
    ),
)


def create_celery_app() -> Celery:
    """
    Create and configure the Celery application.

    Returns:
        Configured Celery application instance.
    """
    settings = get_settings()

    # Create Celery app with Redis broker
    app = Celery(
        "sidot_ai",
        broker=settings.celery_broker_url,
        include=[
            "app.celery_app.tasks.base",
            "app.celery_app.tasks.query",
            "app.celery_app.tasks.indexing",
        ],
    )

    # Build SQLAlchemy result backend URL
    # Convert postgresql:// to db+postgresql://
    result_backend = settings.celery_result_backend
    if result_backend.startswith("postgresql://"):
        result_backend = result_backend.replace("postgresql://", "db+postgresql://", 1)

    # Configure Celery
    app.conf.update(
        # Result backend (PostgreSQL via SQLAlchemy)
        result_backend=result_backend,

        # Task serialization
        task_serializer="json",
        result_serializer="json",
        accept_content=["json"],

        # Timezone
        timezone="America/Sao_Paulo",
        enable_utc=True,

        # Task queues
        task_queues=CELERY_QUEUES,
        task_default_queue=QUEUE_AI_ACTIONS,
        task_default_exchange=default_exchange.name,
        task_default_routing_key="ai.actions",

        # Task routing
        task_routes={
            "app.celery_app.tasks.query.*": {
                "queue": QUEUE_AI_QUERY,
                "routing_key": "ai.query",
            },
            "app.celery_app.tasks.actions.*": {
                "queue": QUEUE_AI_ACTIONS,
                "routing_key": "ai.actions",
            },
            "app.celery_app.tasks.indexing.*": {
                "queue": QUEUE_AI_INDEXING,
                "routing_key": "ai.indexing",
            },
        },

        # Broker settings
        broker_connection_retry_on_startup=True,
        broker_connection_max_retries=10,
        broker_transport_options={
            "visibility_timeout": 3600,  # 1 hour
            "priority_steps": list(range(10)),
        },

        # Result settings
        result_expires=86400,  # 24 hours
        result_extended=True,

        # Task settings
        task_track_started=True,
        task_time_limit=600,  # 10 minutes max per task
        task_soft_time_limit=540,  # 9 minutes soft limit

        # Worker settings
        worker_prefetch_multiplier=1,  # Fair scheduling
        worker_concurrency=4,

        # Key prefix for Redis
        broker_transport_options_key_prefix=settings.redis_key_prefix,

        # Retry settings (global defaults)
        task_default_retry_delay=10,
        task_max_retries=3,

        # Acknowledgment settings
        task_acks_late=True,
        task_reject_on_worker_lost=True,
    )

    return app


# Create the Celery application instance
celery_app = create_celery_app()
