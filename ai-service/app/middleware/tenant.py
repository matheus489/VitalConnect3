"""
Multi-Tenant Context Middleware for AI Service.

Replicates the TenantContext logic from the Go backend:
/backend/internal/middleware/tenant.go

Handles tenant isolation and supports super-admin context switching
via the X-Tenant-Context header.
"""

from dataclasses import dataclass
from typing import Optional
from uuid import UUID

from fastapi import Request, HTTPException, status

from app.middleware.auth import UserClaims


# Header key for tenant context switching (super-admin only)
TENANT_CONTEXT_HEADER = "X-Tenant-Context"


class TenantContextDeniedError(Exception):
    """Raised when a non-super-admin attempts to switch tenant context."""

    def __init__(self, message: str = "Tenant context switch requires super admin privileges"):
        self.message = message
        self.code = "TENANT_CONTEXT_DENIED"
        super().__init__(self.message)


class InvalidTenantIdError(Exception):
    """Raised when an invalid tenant ID is provided."""

    def __init__(self, message: str = "Invalid tenant ID format"):
        self.message = message
        self.code = "INVALID_TENANT_ID"
        super().__init__(self.message)


class MissingTenantContextError(Exception):
    """Raised when tenant context is missing but required."""

    def __init__(self, message: str = "Tenant context not found"):
        self.message = message
        self.code = "TENANT_REQUIRED"
        super().__init__(self.message)


@dataclass
class TenantContext:
    """
    Holds tenant information for the current request.

    Mirrors the TenantContext structure from the Go backend:
    /backend/internal/middleware/tenant.go

    Attributes:
        tenant_id: The tenant ID from JWT claims (user's assigned tenant).
        is_super_admin: Whether the user is a super admin.
        effective_tenant_id: The tenant ID to use for queries (may differ
                            for super-admin context switch).
    """

    tenant_id: str
    is_super_admin: bool
    effective_tenant_id: str


def validate_uuid(value: str) -> bool:
    """
    Validate that a string is a valid UUID format.

    Args:
        value: The string to validate.

    Returns:
        True if valid UUID format, False otherwise.
    """
    try:
        UUID(value)
        return True
    except (ValueError, TypeError):
        return False


def create_tenant_context(
    user_claims: UserClaims,
    header_tenant_id: Optional[str] = None
) -> TenantContext:
    """
    Create a TenantContext from user claims and optional header.

    This function replicates the logic from TenantContextMiddleware
    in the Go backend.

    Args:
        user_claims: The authenticated user's claims from JWT.
        header_tenant_id: Optional tenant ID from X-Tenant-Context header.

    Returns:
        TenantContext with appropriate effective_tenant_id.

    Raises:
        TenantContextDeniedError: If non-super-admin tries to switch context.
        InvalidTenantIdError: If header contains invalid tenant ID format.
    """
    # Default effective tenant ID is the user's assigned tenant
    effective_tenant_id = user_claims.tenant_id

    # Check for X-Tenant-Context header (super-admin only)
    if header_tenant_id:
        # Only super-admins can use X-Tenant-Context header
        if not user_claims.is_super_admin:
            raise TenantContextDeniedError()

        # Validate the header tenant ID format
        if not validate_uuid(header_tenant_id):
            raise InvalidTenantIdError(
                "X-Tenant-Context header must contain a valid UUID"
            )

        # Set effective tenant ID to the one from header
        effective_tenant_id = header_tenant_id

    return TenantContext(
        tenant_id=user_claims.tenant_id,
        is_super_admin=user_claims.is_super_admin,
        effective_tenant_id=effective_tenant_id,
    )


def get_tenant_context_from_request(
    user_claims: UserClaims,
    request: Request
) -> TenantContext:
    """
    Extract tenant context from request and user claims.

    Args:
        user_claims: The authenticated user's claims.
        request: The FastAPI request object.

    Returns:
        TenantContext for the current request.

    Raises:
        HTTPException: If tenant context cannot be established.
    """
    header_tenant_id = request.headers.get(TENANT_CONTEXT_HEADER)

    try:
        return create_tenant_context(user_claims, header_tenant_id)
    except TenantContextDeniedError:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail={
                "error": "tenant context switch requires super admin privileges",
                "code": "TENANT_CONTEXT_DENIED",
            }
        )
    except InvalidTenantIdError:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={
                "error": "invalid X-Tenant-Context header",
                "details": "tenant ID must be a valid UUID",
            }
        )


def require_tenant_context(tenant_ctx: TenantContext) -> None:
    """
    Verify that tenant context is present and valid.

    Args:
        tenant_ctx: The tenant context to validate.

    Raises:
        HTTPException: If tenant context is missing or invalid.
    """
    if not tenant_ctx.effective_tenant_id:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail={
                "error": "tenant ID is required for this operation",
                "code": "TENANT_ID_MISSING",
            }
        )

    if not validate_uuid(tenant_ctx.effective_tenant_id):
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={
                "error": "invalid tenant ID format",
                "code": "INVALID_TENANT_ID",
            }
        )


@dataclass
class RequestContext:
    """
    Combined context for the current request.

    This dataclass combines user claims and tenant context
    into a single object for convenience.
    """

    user: UserClaims
    tenant: TenantContext

    @property
    def user_id(self) -> str:
        """Get the current user's ID."""
        return self.user.user_id

    @property
    def email(self) -> str:
        """Get the current user's email."""
        return self.user.email

    @property
    def role(self) -> str:
        """Get the current user's role."""
        return self.user.role

    @property
    def effective_tenant_id(self) -> str:
        """Get the effective tenant ID for queries."""
        return self.tenant.effective_tenant_id

    @property
    def is_super_admin(self) -> bool:
        """Check if the current user is a super admin."""
        return self.tenant.is_super_admin
