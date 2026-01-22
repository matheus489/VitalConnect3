"""
Base Tool Class for SIDOT AI Assistant.

Provides abstract base class for all tools with:
- Permission validation before execution
- Audit logging integration
- Human-in-the-loop confirmation support
- Error handling standardization
"""

import logging
import time
from abc import ABC, abstractmethod
from dataclasses import dataclass
from datetime import datetime, timezone
from enum import Enum
from typing import Any, Optional
from uuid import UUID

import httpx

from app.config import get_settings
from app.middleware.permissions import (
    check_permission,
    has_permission,
    PermissionDeniedError,
)
from app.middleware.tenant import RequestContext


logger = logging.getLogger(__name__)


class ToolError(Exception):
    """Base exception for tool execution errors."""

    def __init__(
        self,
        message: str,
        code: str = "TOOL_ERROR",
        details: Optional[dict] = None
    ):
        self.message = message
        self.code = code
        self.details = details or {}
        super().__init__(self.message)


class ToolPermissionError(ToolError):
    """Raised when user lacks permission to execute a tool."""

    def __init__(
        self,
        message: str = "Insufficient permissions",
        required_role: Optional[str] = None,
        user_role: Optional[str] = None,
        tool_name: Optional[str] = None,
    ):
        super().__init__(
            message=message,
            code="PERMISSION_DENIED",
            details={
                "required_role": required_role,
                "user_role": user_role,
                "tool_name": tool_name,
            }
        )


class ToolExecutionError(ToolError):
    """Raised when tool execution fails."""

    def __init__(self, message: str, details: Optional[dict] = None):
        super().__init__(
            message=message,
            code="TOOL_EXECUTION_ERROR",
            details=details
        )


class BackendConnectionError(ToolError):
    """Raised when connection to Go backend fails."""

    def __init__(self, message: str = "Failed to connect to backend service"):
        super().__init__(
            message=message,
            code="BACKEND_CONNECTION_ERROR"
        )


@dataclass
class ToolResult:
    """
    Standardized result from tool execution.

    Attributes:
        success: Whether the tool executed successfully.
        data: The tool's output data.
        message: Human-readable message describing the result.
        confirmation_required: Whether human confirmation is needed.
        confirmation_action_id: ID for the pending confirmation action.
        confirmation_details: Details to show user for confirmation.
        execution_time_ms: Time taken to execute the tool.
    """

    success: bool
    data: Optional[dict] = None
    message: Optional[str] = None
    confirmation_required: bool = False
    confirmation_action_id: Optional[str] = None
    confirmation_details: Optional[dict] = None
    execution_time_ms: Optional[int] = None

    def to_dict(self) -> dict:
        """Convert result to dictionary for serialization."""
        result = {
            "success": self.success,
            "data": self.data,
            "message": self.message,
        }
        if self.confirmation_required:
            result["confirmation_required"] = True
            result["confirmation_action_id"] = self.confirmation_action_id
            result["confirmation_details"] = self.confirmation_details
        if self.execution_time_ms is not None:
            result["execution_time_ms"] = self.execution_time_ms
        return result


@dataclass
class ToolContext:
    """
    Context information passed to tool execution.

    Attributes:
        request_ctx: The request context with user and tenant info.
        conversation_id: Optional conversation message ID for audit.
        confirmation_received: Whether user has confirmed the action.
    """

    request_ctx: RequestContext
    conversation_id: Optional[UUID] = None
    confirmation_received: bool = False

    @property
    def user_id(self) -> str:
        return self.request_ctx.user_id

    @property
    def tenant_id(self) -> str:
        return self.request_ctx.effective_tenant_id

    @property
    def role(self) -> str:
        return self.request_ctx.role


class BaseTool(ABC):
    """
    Abstract base class for all SIDOT AI tools.

    Provides common functionality for:
    - Permission validation
    - Audit logging
    - Error handling
    - Backend API communication

    Subclasses must implement:
    - name: Tool name for registration
    - description: Tool description for LLM
    - execute(): Actual tool logic
    """

    # Tool registration properties
    name: str = ""
    description: str = ""

    # Whether this tool requires human confirmation before execution
    requires_confirmation: bool = False

    # Severity level for audit logging
    audit_severity: str = "INFO"

    def __init__(self):
        """Initialize the tool with settings."""
        self._settings = get_settings()
        self._http_client: Optional[httpx.AsyncClient] = None

    @property
    def http_client(self) -> httpx.AsyncClient:
        """Get or create HTTP client for backend communication."""
        if self._http_client is None:
            self._http_client = httpx.AsyncClient(
                base_url=self._settings.go_backend_url,
                timeout=30.0,
            )
        return self._http_client

    async def close(self) -> None:
        """Close HTTP client connection."""
        if self._http_client is not None:
            await self._http_client.aclose()
            self._http_client = None

    def check_permission(self, role: str) -> None:
        """
        Check if the user role has permission to execute this tool.

        Args:
            role: The user's role.

        Raises:
            ToolPermissionError: If permission is denied.
        """
        try:
            check_permission(role, self.name)
        except PermissionDeniedError as e:
            raise ToolPermissionError(
                message=e.message,
                required_role=e.required_role,
                user_role=e.user_role,
                tool_name=self.name,
            )

    def has_permission(self, role: str) -> bool:
        """
        Check if the user role has permission without raising.

        Args:
            role: The user's role.

        Returns:
            True if permitted, False otherwise.
        """
        return has_permission(role, self.name)

    async def run(
        self,
        context: ToolContext,
        **kwargs: Any
    ) -> ToolResult:
        """
        Execute the tool with permission checking and audit logging.

        This is the main entry point for tool execution. It:
        1. Validates permissions
        2. Checks for confirmation requirement
        3. Executes the tool
        4. Logs the result

        Args:
            context: Tool execution context with user/tenant info.
            **kwargs: Tool-specific parameters.

        Returns:
            ToolResult with execution outcome.
        """
        start_time = time.time()

        try:
            # Check permissions
            self.check_permission(context.role)

            # Check if confirmation is required and not yet received
            if self.requires_confirmation and not context.confirmation_received:
                return await self._request_confirmation(context, **kwargs)

            # Execute the tool
            result = await self.execute(context, **kwargs)

            # Add execution time
            execution_time_ms = int((time.time() - start_time) * 1000)
            result.execution_time_ms = execution_time_ms

            # Log success
            logger.info(
                f"Tool executed: {self.name} "
                f"[user_id={context.user_id}] "
                f"[tenant_id={context.tenant_id}] "
                f"[execution_time_ms={execution_time_ms}]"
            )

            return result

        except ToolPermissionError:
            raise

        except ToolError as e:
            execution_time_ms = int((time.time() - start_time) * 1000)
            logger.error(
                f"Tool error: {self.name} "
                f"[user_id={context.user_id}] "
                f"[error={e.message}] "
                f"[code={e.code}]"
            )
            return ToolResult(
                success=False,
                message=e.message,
                data={"error_code": e.code, **e.details},
                execution_time_ms=execution_time_ms,
            )

        except Exception as e:
            execution_time_ms = int((time.time() - start_time) * 1000)
            logger.exception(
                f"Unexpected tool error: {self.name} "
                f"[user_id={context.user_id}] "
                f"[error={str(e)}]"
            )
            return ToolResult(
                success=False,
                message="An unexpected error occurred while executing the tool.",
                data={"error": str(e)},
                execution_time_ms=execution_time_ms,
            )

    async def _request_confirmation(
        self,
        context: ToolContext,
        **kwargs: Any
    ) -> ToolResult:
        """
        Create a confirmation request for human-in-the-loop.

        Args:
            context: Tool execution context.
            **kwargs: Tool parameters to confirm.

        Returns:
            ToolResult with confirmation_required=True.
        """
        import uuid

        action_id = str(uuid.uuid4())
        confirmation_details = self.get_confirmation_details(context, **kwargs)

        logger.info(
            f"Confirmation requested: {self.name} "
            f"[user_id={context.user_id}] "
            f"[action_id={action_id}]"
        )

        return ToolResult(
            success=True,
            message=f"Esta acao requer confirmacao. {confirmation_details.get('message', '')}",
            confirmation_required=True,
            confirmation_action_id=action_id,
            confirmation_details={
                "tool_name": self.name,
                "parameters": kwargs,
                **confirmation_details,
            },
        )

    def get_confirmation_details(
        self,
        context: ToolContext,
        **kwargs: Any
    ) -> dict:
        """
        Get details to display for confirmation dialog.

        Override in subclass to provide tool-specific details.

        Args:
            context: Tool execution context.
            **kwargs: Tool parameters.

        Returns:
            Dictionary with confirmation message and details.
        """
        return {
            "message": f"Deseja executar {self.name}?",
            "action": self.name,
        }

    @abstractmethod
    async def execute(
        self,
        context: ToolContext,
        **kwargs: Any
    ) -> ToolResult:
        """
        Execute the actual tool logic.

        Subclasses must implement this method.

        Args:
            context: Tool execution context with user/tenant info.
            **kwargs: Tool-specific parameters.

        Returns:
            ToolResult with execution outcome.
        """
        pass

    async def call_backend_api(
        self,
        method: str,
        endpoint: str,
        context: ToolContext,
        data: Optional[dict] = None,
        params: Optional[dict] = None,
    ) -> dict:
        """
        Call the Go backend API with proper authentication headers.

        Args:
            method: HTTP method (GET, POST, PUT, DELETE).
            endpoint: API endpoint path.
            context: Tool context for authentication headers.
            data: Request body data (for POST/PUT).
            params: Query parameters.

        Returns:
            Response data as dictionary.

        Raises:
            BackendConnectionError: If connection fails.
            ToolExecutionError: If API returns an error.
        """
        headers = {
            "X-Tenant-Context": context.tenant_id,
            "X-User-ID": context.user_id,
            "Content-Type": "application/json",
        }

        try:
            response = await self.http_client.request(
                method=method,
                url=endpoint,
                headers=headers,
                json=data,
                params=params,
            )

            if response.status_code >= 400:
                error_data = response.json() if response.text else {}
                raise ToolExecutionError(
                    message=error_data.get("error", f"Backend API error: {response.status_code}"),
                    details={"status_code": response.status_code, "response": error_data}
                )

            return response.json() if response.text else {}

        except httpx.RequestError as e:
            logger.error(f"Backend connection error: {e}")
            raise BackendConnectionError(f"Failed to connect to backend: {str(e)}")

    def to_llama_index_tool(self):
        """
        Convert this tool to a LlamaIndex FunctionTool.

        Returns:
            LlamaIndex FunctionTool instance.
        """
        from llama_index.core.tools import FunctionTool

        async def tool_wrapper(**kwargs):
            # Context will be injected at runtime by the agent
            context = kwargs.pop("_context", None)
            if context is None:
                raise ToolError("Tool context is required")
            return await self.run(context, **kwargs)

        return FunctionTool.from_defaults(
            fn=tool_wrapper,
            name=self.name,
            description=self.description,
        )
