"""
Base Celery task class with audit logging for VitalConnect AI Service.

Provides:
- Task retry logic with exponential backoff
- Audit logging to PostgreSQL
- User and tenant context tracking
- Execution time measurement
"""

import logging
import time
import uuid
from datetime import datetime, timezone
from typing import Any, Optional

from celery import Task
from celery.exceptions import MaxRetriesExceededError, SoftTimeLimitExceeded

from app.celery_app.celery_config import celery_app


logger = logging.getLogger(__name__)

# Retry backoff delays in seconds: 10s, 30s, 60s
RETRY_BACKOFF_DELAYS = [10, 30, 60]

# Default max retries
MAX_RETRIES = 3


class TaskAuditRecord:
    """
    Data class for task audit logging.

    Stores information about task execution for audit purposes.
    """

    def __init__(
        self,
        task_id: str,
        task_name: str,
        user_id: Optional[str] = None,
        tenant_id: Optional[str] = None,
        input_params: Optional[dict] = None,
    ):
        self.id = str(uuid.uuid4())
        self.task_id = task_id
        self.task_name = task_name
        self.user_id = user_id
        self.tenant_id = tenant_id
        self.input_params = input_params or {}
        self.output_result: Optional[dict] = None
        self.status = "pending"
        self.error_message: Optional[str] = None
        self.severity = "INFO"
        self.started_at: Optional[datetime] = None
        self.completed_at: Optional[datetime] = None
        self.execution_time_ms: Optional[int] = None

    def to_dict(self) -> dict:
        """Convert audit record to dictionary for storage."""
        return {
            "id": self.id,
            "task_id": self.task_id,
            "task_name": self.task_name,
            "user_id": self.user_id,
            "tenant_id": self.tenant_id,
            "input_params": self.input_params,
            "output_result": self.output_result,
            "status": self.status,
            "error_message": self.error_message,
            "severity": self.severity,
            "started_at": self.started_at.isoformat() if self.started_at else None,
            "completed_at": self.completed_at.isoformat() if self.completed_at else None,
            "execution_time_ms": self.execution_time_ms,
        }


class AuditedTask(Task):
    """
    Base Celery task class with audit logging and exponential backoff retry.

    Features:
    - Automatic retry with exponential backoff (10s, 30s, 60s)
    - Audit logging of task start/end/error to PostgreSQL
    - User and tenant context tracking
    - Execution time measurement
    - Dead letter handling for failed tasks
    """

    # Retry configuration
    autoretry_for = (Exception,)
    max_retries = MAX_RETRIES
    default_retry_delay = RETRY_BACKOFF_DELAYS[0]

    # Don't retry on these exceptions
    dont_autoretry_for = (
        ValueError,
        TypeError,
        KeyError,
        PermissionError,
    )

    # Task tracking
    track_started = True
    acks_late = True

    # Store for audit records (will be persisted to PostgreSQL)
    _audit_record: Optional[TaskAuditRecord] = None

    def before_start(
        self,
        task_id: str,
        args: tuple,
        kwargs: dict,
    ) -> None:
        """
        Called before task execution starts.

        Initializes audit record with task information.
        """
        # Extract user_id and tenant_id from kwargs if provided
        user_id = kwargs.get("user_id")
        tenant_id = kwargs.get("tenant_id")

        # Create audit record
        self._audit_record = TaskAuditRecord(
            task_id=task_id,
            task_name=self.name,
            user_id=user_id,
            tenant_id=tenant_id,
            input_params={
                "args": list(args),
                "kwargs": {k: v for k, v in kwargs.items() if k not in ("user_id", "tenant_id")},
            },
        )

        self._audit_record.started_at = datetime.now(timezone.utc)
        self._audit_record.status = "running"

        logger.info(
            f"Task started: {self.name} [task_id={task_id}] "
            f"[user_id={user_id}] [tenant_id={tenant_id}]"
        )

        self._log_audit_to_db("start")

    def on_success(
        self,
        retval: Any,
        task_id: str,
        args: tuple,
        kwargs: dict,
    ) -> None:
        """
        Called when task completes successfully.

        Updates audit record with result and execution time.
        """
        if self._audit_record:
            self._audit_record.completed_at = datetime.now(timezone.utc)
            self._audit_record.status = "success"
            self._audit_record.severity = "INFO"

            # Store result (truncate if too large)
            if retval is not None:
                try:
                    result_str = str(retval)
                    if len(result_str) > 10000:
                        self._audit_record.output_result = {
                            "truncated": True,
                            "preview": result_str[:1000],
                        }
                    else:
                        self._audit_record.output_result = {"result": retval}
                except Exception:
                    self._audit_record.output_result = {"result": "non-serializable"}

            # Calculate execution time
            if self._audit_record.started_at:
                delta = self._audit_record.completed_at - self._audit_record.started_at
                self._audit_record.execution_time_ms = int(delta.total_seconds() * 1000)

            logger.info(
                f"Task completed: {self.name} [task_id={task_id}] "
                f"[execution_time_ms={self._audit_record.execution_time_ms}]"
            )

            self._log_audit_to_db("success")

    def on_failure(
        self,
        exc: Exception,
        task_id: str,
        args: tuple,
        kwargs: dict,
        einfo: Any,
    ) -> None:
        """
        Called when task fails after all retries exhausted.

        Updates audit record with error information for dead letter handling.
        """
        if self._audit_record:
            self._audit_record.completed_at = datetime.now(timezone.utc)
            self._audit_record.status = "failed"
            self._audit_record.severity = "CRITICAL"
            self._audit_record.error_message = str(exc)

            # Calculate execution time
            if self._audit_record.started_at:
                delta = self._audit_record.completed_at - self._audit_record.started_at
                self._audit_record.execution_time_ms = int(delta.total_seconds() * 1000)

            logger.error(
                f"Task failed: {self.name} [task_id={task_id}] "
                f"[error={exc}] [execution_time_ms={self._audit_record.execution_time_ms}]"
            )

            self._log_audit_to_db("failure")

        # Handle dead letter queue logging
        self._handle_dead_letter(task_id, exc, args, kwargs)

    def on_retry(
        self,
        exc: Exception,
        task_id: str,
        args: tuple,
        kwargs: dict,
        einfo: Any,
    ) -> None:
        """
        Called when task is being retried.

        Logs retry attempt with backoff information.
        """
        retry_count = self.request.retries
        next_delay = self._get_retry_delay(retry_count)

        logger.warning(
            f"Task retrying: {self.name} [task_id={task_id}] "
            f"[retry={retry_count}/{self.max_retries}] "
            f"[next_delay={next_delay}s] [error={exc}]"
        )

        if self._audit_record:
            self._audit_record.status = "retrying"
            self._audit_record.severity = "WARN"
            self._audit_record.error_message = f"Retry {retry_count}: {exc}"
            self._log_audit_to_db("retry")

    def retry(
        self,
        args: Optional[tuple] = None,
        kwargs: Optional[dict] = None,
        exc: Optional[Exception] = None,
        throw: bool = True,
        eta: Optional[datetime] = None,
        countdown: Optional[float] = None,
        max_retries: Optional[int] = None,
        **options: Any,
    ) -> Any:
        """
        Override retry to implement exponential backoff.

        Uses configured backoff delays: 10s, 30s, 60s
        """
        if countdown is None and eta is None:
            retry_count = self.request.retries
            countdown = self._get_retry_delay(retry_count)

        return super().retry(
            args=args,
            kwargs=kwargs,
            exc=exc,
            throw=throw,
            eta=eta,
            countdown=countdown,
            max_retries=max_retries,
            **options,
        )

    def _get_retry_delay(self, retry_count: int) -> int:
        """
        Get retry delay based on retry count using exponential backoff.

        Args:
            retry_count: Current retry attempt number (0-indexed)

        Returns:
            Delay in seconds before next retry
        """
        if retry_count < len(RETRY_BACKOFF_DELAYS):
            return RETRY_BACKOFF_DELAYS[retry_count]
        return RETRY_BACKOFF_DELAYS[-1]

    def _log_audit_to_db(self, event_type: str) -> None:
        """
        Log audit record to PostgreSQL.

        This is a placeholder that will be connected to the actual
        repository in Task Group 3 when database models are created.

        Args:
            event_type: Type of event (start, success, failure, retry)
        """
        if not self._audit_record:
            return

        # Log audit record details for now
        # This will be replaced with actual database logging in Task Group 3
        audit_data = self._audit_record.to_dict()
        logger.debug(f"Audit log [{event_type}]: {audit_data}")

    def _handle_dead_letter(
        self,
        task_id: str,
        exc: Exception,
        args: tuple,
        kwargs: dict,
    ) -> None:
        """
        Handle dead letter queue for failed tasks.

        After all retries are exhausted, store the failed task information
        for manual review and potential reprocessing.

        Args:
            task_id: ID of the failed task
            exc: Exception that caused the failure
            args: Task positional arguments
            kwargs: Task keyword arguments
        """
        dead_letter_record = {
            "task_id": task_id,
            "task_name": self.name,
            "args": list(args),
            "kwargs": kwargs,
            "error": str(exc),
            "error_type": type(exc).__name__,
            "failed_at": datetime.now(timezone.utc).isoformat(),
            "retry_count": self.request.retries,
        }

        logger.error(
            f"Dead letter: Task {self.name} failed permanently after "
            f"{self.request.retries} retries. Record: {dead_letter_record}"
        )

        # This will be connected to a dead letter table in Task Group 3


# Register base task with Celery app
@celery_app.task(bind=True, base=AuditedTask)
def health_check_task(self, **kwargs) -> dict:
    """
    Simple health check task for testing Celery worker connectivity.

    Returns:
        Dictionary with health check status
    """
    return {
        "status": "healthy",
        "worker": "celery",
        "timestamp": datetime.now(timezone.utc).isoformat(),
    }
