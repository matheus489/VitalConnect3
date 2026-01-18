"""
FastAPI Dependency Injection for AI Service.

Provides dependencies for extracting authentication context,
tenant context, and database sessions throughout the application.

Usage:
    from app.dependencies import get_request_context, get_db_session

    @router.get("/example")
    async def example_endpoint(
        request_ctx: RequestContext = Depends(get_request_context),
        db: AsyncSession = Depends(get_db_session),
    ):
        # request_ctx contains user claims and tenant context
        user_id = request_ctx.user_id
        tenant_id = request_ctx.effective_tenant_id
        ...
"""

from typing import AsyncGenerator, Optional

from fastapi import Depends, Request, HTTPException, status
from sqlalchemy.ext.asyncio import AsyncSession

from app.middleware.auth import (
    UserClaims,
    get_current_user,
    get_current_user_optional,
)
from app.middleware.tenant import (
    TenantContext,
    RequestContext,
    get_tenant_context_from_request,
    require_tenant_context,
)
from app.middleware.permissions import (
    check_permission,
    PermissionDeniedError,
    get_allowed_tools,
)


async def get_user_claims(
    user_claims: UserClaims = Depends(get_current_user)
) -> UserClaims:
    """
    Dependency that provides the current user's claims.

    Requires a valid JWT token in the Authorization header.

    Args:
        user_claims: Injected by get_current_user dependency.

    Returns:
        UserClaims for the authenticated user.

    Raises:
        HTTPException: If authentication fails.
    """
    return user_claims


async def get_user_claims_optional(
    user_claims: Optional[UserClaims] = Depends(get_current_user_optional)
) -> Optional[UserClaims]:
    """
    Dependency that optionally provides the current user's claims.

    Does not require authentication - returns None if not authenticated.

    Args:
        user_claims: Injected by get_current_user_optional dependency.

    Returns:
        UserClaims if authenticated, None otherwise.
    """
    return user_claims


async def get_tenant_context(
    request: Request,
    user_claims: UserClaims = Depends(get_current_user)
) -> TenantContext:
    """
    Dependency that provides the tenant context for the current request.

    Extracts tenant information from user claims and optional
    X-Tenant-Context header (for super-admin context switching).

    Args:
        request: The FastAPI request object.
        user_claims: Injected user claims from authentication.

    Returns:
        TenantContext with effective_tenant_id.

    Raises:
        HTTPException: If tenant context cannot be established.
    """
    return get_tenant_context_from_request(user_claims, request)


async def get_request_context(
    request: Request,
    user_claims: UserClaims = Depends(get_current_user)
) -> RequestContext:
    """
    Dependency that provides combined user and tenant context.

    This is the primary dependency for protected endpoints,
    providing both authentication and tenant isolation.

    Args:
        request: The FastAPI request object.
        user_claims: Injected user claims from authentication.

    Returns:
        RequestContext containing user claims and tenant context.

    Raises:
        HTTPException: If authentication or tenant context fails.
    """
    tenant_ctx = get_tenant_context_from_request(user_claims, request)
    return RequestContext(user=user_claims, tenant=tenant_ctx)


async def require_tenant(
    tenant_ctx: TenantContext = Depends(get_tenant_context)
) -> TenantContext:
    """
    Dependency that requires a valid tenant context.

    Ensures that the effective_tenant_id is present and valid.

    Args:
        tenant_ctx: Injected tenant context.

    Returns:
        Validated TenantContext.

    Raises:
        HTTPException: If tenant context is missing or invalid.
    """
    require_tenant_context(tenant_ctx)
    return tenant_ctx


def require_permission(tool_name: str):
    """
    Dependency factory that checks permission for a specific tool.

    Usage:
        @router.post("/update-status")
        async def update_status(
            request_ctx: RequestContext = Depends(require_permission("update_occurrence_status"))
        ):
            ...

    Args:
        tool_name: The name of the tool requiring permission.

    Returns:
        A dependency that validates permission and returns RequestContext.
    """
    async def dependency(
        request: Request,
        user_claims: UserClaims = Depends(get_current_user)
    ) -> RequestContext:
        # Create request context first
        tenant_ctx = get_tenant_context_from_request(user_claims, request)
        request_ctx = RequestContext(user=user_claims, tenant=tenant_ctx)

        # Check permission for the tool
        try:
            check_permission(user_claims.role, tool_name)
        except PermissionDeniedError as e:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail={
                    "error": "insufficient permissions",
                    "code": "PERMISSION_DENIED",
                    "required_role": e.required_role,
                    "user_role": e.user_role,
                    "tool": tool_name,
                }
            )

        return request_ctx

    return dependency


def require_roles(*allowed_roles: str):
    """
    Dependency factory that requires any of the specified roles.

    Usage:
        @router.post("/admin-action")
        async def admin_action(
            request_ctx: RequestContext = Depends(require_roles("admin", "gestor"))
        ):
            ...

    Args:
        allowed_roles: Variable number of role strings that are allowed.

    Returns:
        A dependency that validates role and returns RequestContext.
    """
    async def dependency(
        request: Request,
        user_claims: UserClaims = Depends(get_current_user)
    ) -> RequestContext:
        # Create request context first
        tenant_ctx = get_tenant_context_from_request(user_claims, request)
        request_ctx = RequestContext(user=user_claims, tenant=tenant_ctx)

        # Check if user has any of the allowed roles
        if user_claims.role not in allowed_roles:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail={
                    "error": "insufficient permissions",
                    "code": "PERMISSION_DENIED",
                    "required_roles": list(allowed_roles),
                    "user_role": user_claims.role,
                }
            )

        return request_ctx

    return dependency


async def get_allowed_tools_for_user(
    user_claims: UserClaims = Depends(get_current_user)
) -> set:
    """
    Dependency that returns the set of tools the current user can execute.

    Useful for informing the user what actions are available to them.

    Args:
        user_claims: Injected user claims.

    Returns:
        Set of tool names the user can execute.
    """
    return get_allowed_tools(user_claims.role)


# Re-export commonly used items for convenience
__all__ = [
    # User authentication
    "get_user_claims",
    "get_user_claims_optional",
    # Tenant context
    "get_tenant_context",
    "require_tenant",
    # Combined context
    "get_request_context",
    # Permission checks
    "require_permission",
    "require_roles",
    "get_allowed_tools_for_user",
    # Types
    "UserClaims",
    "TenantContext",
    "RequestContext",
]
