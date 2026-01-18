"""
JWT Authentication Middleware for AI Service.

Validates JWT tokens using the same secret as the Go backend and extracts
user claims for use throughout the request lifecycle.

This module replicates the authentication logic from:
/backend/internal/middleware/auth.go
"""

from dataclasses import dataclass
from typing import Optional

import jwt
from fastapi import Request, HTTPException, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials

from app.config import get_settings


class TokenExpiredError(Exception):
    """Raised when the JWT token has expired."""

    def __init__(self, message: str = "Token has expired"):
        self.message = message
        self.code = "TOKEN_EXPIRED"
        super().__init__(self.message)


class InvalidTokenError(Exception):
    """Raised when the JWT token is invalid or malformed."""

    def __init__(self, message: str = "Invalid token"):
        self.message = message
        self.code = "INVALID_TOKEN"
        super().__init__(self.message)


class InvalidClaimsError(Exception):
    """Raised when JWT claims are invalid or missing required fields."""

    def __init__(self, message: str = "Invalid token claims"):
        self.message = message
        self.code = "INVALID_CLAIMS"
        super().__init__(self.message)


@dataclass
class UserClaims:
    """
    User claims extracted from JWT token.

    Mirrors the UserClaims structure from the Go backend:
    /backend/internal/middleware/auth.go
    """

    user_id: str
    email: str
    role: str
    tenant_id: str
    is_super_admin: bool = False
    hospital_id: Optional[str] = None


def validate_jwt_token(
    token: str,
    secret: str,
    algorithm: str = "HS256"
) -> UserClaims:
    """
    Validate a JWT token and extract user claims.

    Args:
        token: The JWT token string to validate.
        secret: The secret key used to sign the token.
        algorithm: The algorithm used for signing (default: HS256).

    Returns:
        UserClaims object with extracted user information.

    Raises:
        TokenExpiredError: If the token has expired.
        InvalidTokenError: If the token is invalid or malformed.
        InvalidClaimsError: If required claims are missing.
    """
    try:
        payload = jwt.decode(
            token,
            secret,
            algorithms=[algorithm],
            options={"require": ["exp", "iat"]}
        )
    except jwt.ExpiredSignatureError:
        raise TokenExpiredError()
    except jwt.InvalidTokenError as e:
        raise InvalidTokenError(str(e))

    # Extract and validate required claims
    user_id = payload.get("user_id")
    email = payload.get("email")
    role = payload.get("role")
    tenant_id = payload.get("tenant_id")

    if not user_id:
        raise InvalidClaimsError("Missing user_id in token")
    if not email:
        raise InvalidClaimsError("Missing email in token")
    if not role:
        raise InvalidClaimsError("Missing role in token")

    return UserClaims(
        user_id=user_id,
        email=email,
        role=role,
        tenant_id=tenant_id or "",
        is_super_admin=payload.get("is_super_admin", False),
        hospital_id=payload.get("hospital_id"),
    )


class JWTBearer(HTTPBearer):
    """
    FastAPI security scheme for JWT Bearer token authentication.

    This class validates the Authorization header and extracts user claims
    from the JWT token.
    """

    def __init__(self, auto_error: bool = True):
        super().__init__(auto_error=auto_error)

    async def __call__(self, request: Request) -> Optional[UserClaims]:
        """
        Validate the JWT token from the Authorization header.

        Args:
            request: The FastAPI request object.

        Returns:
            UserClaims if token is valid.

        Raises:
            HTTPException: If token is missing, expired, or invalid.
        """
        credentials: Optional[HTTPAuthorizationCredentials] = await super().__call__(request)

        if not credentials:
            if self.auto_error:
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail={
                        "error": "authorization header required",
                    }
                )
            return None

        if credentials.scheme.lower() != "bearer":
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail={
                    "error": "invalid authorization header format",
                }
            )

        token = credentials.credentials
        if not token:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail={
                    "error": "token required",
                }
            )

        settings = get_settings()

        try:
            claims = validate_jwt_token(
                token,
                settings.jwt_secret,
                settings.jwt_algorithm
            )
            return claims

        except TokenExpiredError:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail={
                    "error": "token has expired",
                    "code": "TOKEN_EXPIRED",
                }
            )
        except InvalidTokenError:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail={
                    "error": "invalid token",
                    "code": "INVALID_TOKEN",
                }
            )
        except InvalidClaimsError as e:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail={
                    "error": e.message,
                    "code": "INVALID_CLAIMS",
                }
            )


# Singleton instances for dependency injection
jwt_bearer = JWTBearer()
jwt_bearer_optional = JWTBearer(auto_error=False)


async def get_current_user(request: Request) -> UserClaims:
    """
    Dependency that extracts and validates the current user from JWT.

    Args:
        request: The FastAPI request object.

    Returns:
        UserClaims for the authenticated user.

    Raises:
        HTTPException: If authentication fails.
    """
    return await jwt_bearer(request)


async def get_current_user_optional(request: Request) -> Optional[UserClaims]:
    """
    Dependency that optionally extracts the current user from JWT.

    Does not raise an error if the token is missing.

    Args:
        request: The FastAPI request object.

    Returns:
        UserClaims if token is valid, None otherwise.
    """
    try:
        return await jwt_bearer_optional(request)
    except HTTPException:
        return None
