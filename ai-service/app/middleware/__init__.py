"""
Middleware module.

Contains authentication, tenant context, and permission middleware.

Exports:
    - Authentication: UserClaims, validate_jwt_token, JWTBearer
    - Tenant: TenantContext, create_tenant_context, RequestContext
    - Permissions: check_permission, has_permission, PermissionDeniedError
"""

from app.middleware.auth import (
    UserClaims,
    validate_jwt_token,
    JWTBearer,
    TokenExpiredError,
    InvalidTokenError,
    InvalidClaimsError,
    get_current_user,
    get_current_user_optional,
)

from app.middleware.tenant import (
    TenantContext,
    RequestContext,
    create_tenant_context,
    get_tenant_context_from_request,
    require_tenant_context,
    TenantContextDeniedError,
    InvalidTenantIdError,
    MissingTenantContextError,
    TENANT_CONTEXT_HEADER,
)

from app.middleware.permissions import (
    Role,
    check_permission,
    has_permission,
    get_allowed_tools,
    get_minimum_required_role,
    PermissionDeniedError,
    PERMISSION_MATRIX,
)


__all__ = [
    # Authentication
    "UserClaims",
    "validate_jwt_token",
    "JWTBearer",
    "TokenExpiredError",
    "InvalidTokenError",
    "InvalidClaimsError",
    "get_current_user",
    "get_current_user_optional",
    # Tenant Context
    "TenantContext",
    "RequestContext",
    "create_tenant_context",
    "get_tenant_context_from_request",
    "require_tenant_context",
    "TenantContextDeniedError",
    "InvalidTenantIdError",
    "MissingTenantContextError",
    "TENANT_CONTEXT_HEADER",
    # Permissions
    "Role",
    "check_permission",
    "has_permission",
    "get_allowed_tools",
    "get_minimum_required_role",
    "PermissionDeniedError",
    "PERMISSION_MATRIX",
]
