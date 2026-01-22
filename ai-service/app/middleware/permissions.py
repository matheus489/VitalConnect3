"""
Role-Based Permission Checker for AI Service.

Defines the permission matrix for tool execution based on user roles.
This module ensures that users can only execute actions appropriate
to their access level.

Permission Matrix (from spec):
| Tool                      | admin | gestor | operador | medico |
|---------------------------|-------|--------|----------|--------|
| list_occurrences          | Yes   | Yes    | Yes      | Yes    |
| get_occurrence_details    | Yes   | Yes    | Yes      | Yes    |
| update_occurrence_status  | Yes   | Yes    | Yes      | No     |
| send_team_notification    | Yes   | Yes    | No       | No     |
| generate_report           | Yes   | Yes    | No       | No     |
| search_documentation      | Yes   | Yes    | Yes      | Yes    |
"""

from enum import Enum
from typing import Set


class Role(str, Enum):
    """User roles in SIDOT."""

    ADMIN = "admin"
    GESTOR = "gestor"
    OPERADOR = "operador"
    MEDICO = "medico"


class PermissionDeniedError(Exception):
    """Raised when a user attempts an action without required permissions."""

    def __init__(
        self,
        message: str = "Insufficient permissions",
        required_role: str = None,
        user_role: str = None,
        tool_name: str = None
    ):
        self.message = message
        self.code = "PERMISSION_DENIED"
        self.required_role = required_role
        self.user_role = user_role
        self.tool_name = tool_name
        super().__init__(self.message)


# Permission matrix: tool_name -> set of allowed roles
PERMISSION_MATRIX: dict[str, Set[str]] = {
    # All authenticated users can access these
    "list_occurrences": {
        Role.ADMIN, Role.GESTOR, Role.OPERADOR, Role.MEDICO
    },
    "get_occurrence_details": {
        Role.ADMIN, Role.GESTOR, Role.OPERADOR, Role.MEDICO
    },
    "search_documentation": {
        Role.ADMIN, Role.GESTOR, Role.OPERADOR, Role.MEDICO
    },

    # operador+ (operador, gestor, admin)
    "update_occurrence_status": {
        Role.ADMIN, Role.GESTOR, Role.OPERADOR
    },

    # gestor+ (gestor, admin)
    "send_team_notification": {
        Role.ADMIN, Role.GESTOR
    },
    "generate_report": {
        Role.ADMIN, Role.GESTOR
    },
}


# Role hierarchy for error messages
ROLE_HIERARCHY = {
    Role.ADMIN: 4,
    Role.GESTOR: 3,
    Role.OPERADOR: 2,
    Role.MEDICO: 1,
}


def get_minimum_required_role(tool_name: str) -> str:
    """
    Get the minimum role required to execute a tool.

    Args:
        tool_name: The name of the tool.

    Returns:
        The minimum role name required, or "admin" if tool is unknown.
    """
    allowed_roles = PERMISSION_MATRIX.get(tool_name, set())

    if not allowed_roles:
        return "admin"  # Default to admin-only for unknown tools

    # Find the role with lowest hierarchy level
    min_level = 5
    min_role = "admin"

    for role in allowed_roles:
        level = ROLE_HIERARCHY.get(role, 5)
        if level < min_level:
            min_level = level
            min_role = role.value if isinstance(role, Role) else role

    return min_role


def check_permission(role: str, tool_name: str) -> None:
    """
    Check if a user role has permission to execute a tool.

    Args:
        role: The user's role.
        tool_name: The name of the tool to execute.

    Raises:
        PermissionDeniedError: If the user doesn't have permission.
    """
    allowed_roles = PERMISSION_MATRIX.get(tool_name)

    # Unknown tools default to admin-only
    if allowed_roles is None:
        allowed_roles = {Role.ADMIN}

    # Normalize role to enum if possible
    try:
        role_enum = Role(role)
    except ValueError:
        # Unknown role - deny access
        raise PermissionDeniedError(
            message=f"Unknown role: {role}",
            user_role=role,
            tool_name=tool_name,
        )

    if role_enum not in allowed_roles:
        min_required = get_minimum_required_role(tool_name)
        raise PermissionDeniedError(
            message=f"Insufficient permissions to execute {tool_name}",
            required_role=min_required,
            user_role=role,
            tool_name=tool_name,
        )


def has_permission(role: str, tool_name: str) -> bool:
    """
    Check if a user role has permission to execute a tool.

    This is a non-throwing version of check_permission.

    Args:
        role: The user's role.
        tool_name: The name of the tool to execute.

    Returns:
        True if the user has permission, False otherwise.
    """
    try:
        check_permission(role, tool_name)
        return True
    except PermissionDeniedError:
        return False


def get_allowed_tools(role: str) -> Set[str]:
    """
    Get the set of tools a user role is allowed to execute.

    Args:
        role: The user's role.

    Returns:
        Set of tool names the user can execute.
    """
    try:
        role_enum = Role(role)
    except ValueError:
        return set()

    allowed_tools = set()
    for tool_name, allowed_roles in PERMISSION_MATRIX.items():
        if role_enum in allowed_roles:
            allowed_tools.add(tool_name)

    return allowed_tools


def require_any_role(*allowed_roles: str):
    """
    Decorator factory to require any of the specified roles.

    Usage:
        @require_any_role("admin", "gestor")
        async def some_handler(...):
            ...

    Args:
        allowed_roles: Variable number of role strings that are allowed.

    Returns:
        A decorator that checks the user's role.
    """
    def decorator(func):
        from functools import wraps

        @wraps(func)
        async def wrapper(*args, **kwargs):
            # Find the request context in kwargs
            from app.middleware.tenant import RequestContext

            request_ctx = kwargs.get("request_ctx")
            if not request_ctx or not isinstance(request_ctx, RequestContext):
                raise PermissionDeniedError(
                    message="Request context required for permission check"
                )

            user_role = request_ctx.role

            if user_role not in allowed_roles:
                raise PermissionDeniedError(
                    message=f"This action requires one of these roles: {', '.join(allowed_roles)}",
                    required_role=", ".join(allowed_roles),
                    user_role=user_role,
                )

            return await func(*args, **kwargs)

        return wrapper
    return decorator
