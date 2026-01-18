"""
Security tests for AI Service.

Tests for Task Group 4: Authentication and Multi-Tenant Middleware
- JWT token validation
- Tenant context extraction
- Permission checking
- Super-admin context switch
"""

import os
from datetime import datetime, timedelta, timezone
from unittest.mock import AsyncMock, MagicMock, patch

import jwt
import pytest
from fastapi import FastAPI, Request
from fastapi.testclient import TestClient

# Ensure test environment
os.environ.setdefault("ENVIRONMENT", "testing")
os.environ.setdefault("JWT_SECRET", "test-jwt-secret-key")


class TestJWTValidation:
    """Test JWT token validation middleware."""

    def setup_method(self):
        """Set up test fixtures."""
        self.jwt_secret = "test-jwt-secret-key"
        self.jwt_algorithm = "HS256"

    def _create_token(
        self,
        user_id: str = "123e4567-e89b-12d3-a456-426614174000",
        email: str = "test@example.com",
        role: str = "operador",
        tenant_id: str = "987fcdeb-51a2-3b4c-d567-890123456789",
        is_super_admin: bool = False,
        expired: bool = False,
    ) -> str:
        """Create a test JWT token."""
        exp = datetime.now(timezone.utc) + timedelta(hours=1)
        if expired:
            exp = datetime.now(timezone.utc) - timedelta(hours=1)

        payload = {
            "user_id": user_id,
            "email": email,
            "role": role,
            "tenant_id": tenant_id,
            "is_super_admin": is_super_admin,
            "exp": exp,
            "iat": datetime.now(timezone.utc),
        }
        return jwt.encode(payload, self.jwt_secret, algorithm=self.jwt_algorithm)

    def test_valid_jwt_token_extracts_user_claims(self):
        """
        Test that a valid JWT token is properly validated and claims extracted.

        The middleware should decode the token and make user claims available
        in the request context.
        """
        from app.middleware.auth import validate_jwt_token, UserClaims

        token = self._create_token(
            user_id="123e4567-e89b-12d3-a456-426614174000",
            email="test@example.com",
            role="operador",
            tenant_id="987fcdeb-51a2-3b4c-d567-890123456789",
        )

        claims = validate_jwt_token(token, self.jwt_secret, self.jwt_algorithm)

        assert claims is not None
        assert claims.user_id == "123e4567-e89b-12d3-a456-426614174000"
        assert claims.email == "test@example.com"
        assert claims.role == "operador"
        assert claims.tenant_id == "987fcdeb-51a2-3b4c-d567-890123456789"
        assert claims.is_super_admin is False

    def test_expired_token_raises_token_expired_error(self):
        """
        Test that an expired JWT token raises TOKEN_EXPIRED error.

        Following the Go backend pattern, expired tokens should return
        a specific error code for the client to handle token refresh.
        """
        from app.middleware.auth import validate_jwt_token, TokenExpiredError

        expired_token = self._create_token(expired=True)

        with pytest.raises(TokenExpiredError):
            validate_jwt_token(expired_token, self.jwt_secret, self.jwt_algorithm)

    def test_invalid_token_raises_invalid_token_error(self):
        """
        Test that an invalid JWT token raises INVALID_TOKEN error.

        Malformed or tampered tokens should be rejected with a clear error.
        """
        from app.middleware.auth import validate_jwt_token, InvalidTokenError

        invalid_token = "invalid.token.here"

        with pytest.raises(InvalidTokenError):
            validate_jwt_token(invalid_token, self.jwt_secret, self.jwt_algorithm)


class TestTenantContext:
    """Test tenant context extraction middleware."""

    def test_tenant_context_extraction_from_claims(self):
        """
        Test that tenant context is correctly extracted from user claims.

        The middleware should create a TenantContext with the user's
        assigned tenant ID as the effective tenant ID.
        """
        from app.middleware.auth import UserClaims
        from app.middleware.tenant import TenantContext, create_tenant_context

        user_claims = UserClaims(
            user_id="123e4567-e89b-12d3-a456-426614174000",
            email="test@example.com",
            role="operador",
            tenant_id="987fcdeb-51a2-3b4c-d567-890123456789",
            is_super_admin=False,
        )

        tenant_ctx = create_tenant_context(user_claims, header_tenant_id=None)

        assert tenant_ctx.tenant_id == "987fcdeb-51a2-3b4c-d567-890123456789"
        assert tenant_ctx.effective_tenant_id == "987fcdeb-51a2-3b4c-d567-890123456789"
        assert tenant_ctx.is_super_admin is False

    def test_super_admin_context_switch_via_header(self):
        """
        Test that super-admin can switch tenant context via X-Tenant-Context header.

        Super-admin users should be able to impersonate different tenants
        by providing the X-Tenant-Context header.
        """
        from app.middleware.auth import UserClaims
        from app.middleware.tenant import TenantContext, create_tenant_context

        super_admin_claims = UserClaims(
            user_id="admin-uuid-here",
            email="admin@example.com",
            role="admin",
            tenant_id="admin-tenant-id",
            is_super_admin=True,
        )

        target_tenant_id = "target-tenant-uuid-123"
        tenant_ctx = create_tenant_context(
            super_admin_claims,
            header_tenant_id=target_tenant_id
        )

        assert tenant_ctx.tenant_id == "admin-tenant-id"
        assert tenant_ctx.effective_tenant_id == target_tenant_id
        assert tenant_ctx.is_super_admin is True

    def test_non_super_admin_cannot_switch_context(self):
        """
        Test that non-super-admin users cannot switch tenant context.

        Regular users attempting to use X-Tenant-Context header should
        get an error (TENANT_CONTEXT_DENIED).
        """
        from app.middleware.auth import UserClaims
        from app.middleware.tenant import (
            TenantContext,
            create_tenant_context,
            TenantContextDeniedError,
        )

        regular_user_claims = UserClaims(
            user_id="user-uuid",
            email="user@example.com",
            role="operador",
            tenant_id="user-tenant-id",
            is_super_admin=False,
        )

        with pytest.raises(TenantContextDeniedError):
            create_tenant_context(
                regular_user_claims,
                header_tenant_id="other-tenant-id"
            )


class TestPermissionChecker:
    """Test role-based permission checking."""

    def test_permission_check_allows_authorized_role(self):
        """
        Test that users with proper role can execute allowed tools.

        Based on the permission matrix, operador+ should be able to
        update_occurrence_status, while gestor+ can send_team_notification.
        """
        from app.middleware.permissions import check_permission, PermissionDeniedError

        # operador can update_occurrence_status
        check_permission("operador", "update_occurrence_status")  # Should not raise

        # gestor can send_team_notification
        check_permission("gestor", "send_team_notification")  # Should not raise

        # admin can do everything
        check_permission("admin", "generate_report")  # Should not raise

    def test_permission_check_denies_unauthorized_role(self):
        """
        Test that users without proper role are denied tool execution.

        medico should not be able to update_occurrence_status,
        and operador should not be able to send_team_notification.
        """
        from app.middleware.permissions import check_permission, PermissionDeniedError

        # medico cannot update_occurrence_status
        with pytest.raises(PermissionDeniedError):
            check_permission("medico", "update_occurrence_status")

        # operador cannot send_team_notification
        with pytest.raises(PermissionDeniedError):
            check_permission("operador", "send_team_notification")
