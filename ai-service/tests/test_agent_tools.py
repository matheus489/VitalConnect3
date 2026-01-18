"""
Tests for AI Agent and Tools Layer.

Tests for Task Group 6: LlamaIndex Agent and Tool Definitions
- Tool execution with permission check
- Agent response generation
- Human-in-the-loop flow
"""

import os
from datetime import datetime, timedelta, timezone
from unittest.mock import AsyncMock, MagicMock, patch
from uuid import uuid4

import pytest

# Ensure test environment
os.environ.setdefault("ENVIRONMENT", "testing")
os.environ.setdefault("JWT_SECRET", "test-jwt-secret-key")
os.environ.setdefault("OPENAI_API_KEY", "test-openai-key")


class TestToolExecutionWithPermissions:
    """Test tool execution with permission checking."""

    def _create_request_context(
        self,
        role: str = "operador",
        user_id: str = None,
        tenant_id: str = None,
    ):
        """Create a mock RequestContext."""
        from app.middleware.auth import UserClaims
        from app.middleware.tenant import TenantContext, RequestContext

        user_id = user_id or str(uuid4())
        tenant_id = tenant_id or str(uuid4())

        user_claims = UserClaims(
            user_id=user_id,
            email="test@example.com",
            role=role,
            tenant_id=tenant_id,
            is_super_admin=False,
        )

        tenant_context = TenantContext(
            tenant_id=tenant_id,
            is_super_admin=False,
            effective_tenant_id=tenant_id,
        )

        return RequestContext(user=user_claims, tenant=tenant_context)

    def test_tool_permission_check_allows_authorized_role(self):
        """
        Test that tools correctly allow execution for authorized roles.

        The permission checking should pass for roles that are in the
        tool's allowed roles set according to the permission matrix.
        """
        from app.tools.base import ToolContext
        from app.tools.occurrence_tools import ListOccurrencesTool

        # operador should be allowed to list occurrences
        request_ctx = self._create_request_context(role="operador")
        tool = ListOccurrencesTool()

        # This should not raise
        tool.check_permission(request_ctx.role)

    def test_tool_permission_check_denies_unauthorized_role(self):
        """
        Test that tools correctly deny execution for unauthorized roles.

        For example, medico should not be able to update_occurrence_status.
        """
        from app.tools.base import ToolContext, ToolPermissionError
        from app.tools.occurrence_tools import UpdateOccurrenceStatusTool

        # medico should NOT be allowed to update occurrence status
        request_ctx = self._create_request_context(role="medico")
        tool = UpdateOccurrenceStatusTool()

        with pytest.raises(ToolPermissionError):
            tool.check_permission(request_ctx.role)

    def test_tool_permission_check_for_notification_tool(self):
        """
        Test permission checking for send_team_notification tool.

        Only gestor+ (gestor, admin) should be able to send notifications.
        """
        from app.tools.base import ToolPermissionError
        from app.tools.notification_tools import SendTeamNotificationTool

        tool = SendTeamNotificationTool()

        # gestor should be allowed
        tool.check_permission("gestor")

        # operador should NOT be allowed
        with pytest.raises(ToolPermissionError):
            tool.check_permission("operador")


class TestAgentResponseGeneration:
    """Test agent response generation capabilities."""

    def _create_request_context(self, role: str = "operador"):
        """Create a mock RequestContext."""
        from app.middleware.auth import UserClaims
        from app.middleware.tenant import TenantContext, RequestContext

        user_id = str(uuid4())
        tenant_id = str(uuid4())

        user_claims = UserClaims(
            user_id=user_id,
            email="test@example.com",
            role=role,
            tenant_id=tenant_id,
            is_super_admin=False,
        )

        tenant_context = TenantContext(
            tenant_id=tenant_id,
            is_super_admin=False,
            effective_tenant_id=tenant_id,
        )

        return RequestContext(user=user_claims, tenant=tenant_context)

    def test_copilot_agent_creation_with_valid_context(self):
        """
        Test that CopilotAgent can be created with a valid request context.

        The agent should initialize with the correct tools based on
        the user's role permissions.
        """
        from app.agents.copilot_agent import CopilotAgent

        request_ctx = self._create_request_context(role="gestor")

        agent = CopilotAgent(request_ctx=request_ctx)

        # Agent should be created successfully
        assert agent is not None
        assert agent._request_ctx == request_ctx

    def test_copilot_agent_filters_tools_by_role(self):
        """
        Test that the agent only registers tools the user is allowed to use.

        A medico should have fewer tools available than a gestor.
        """
        from app.agents.copilot_agent import CopilotAgent

        # medico has limited permissions
        medico_ctx = self._create_request_context(role="medico")
        medico_agent = CopilotAgent(request_ctx=medico_ctx)
        medico_tools = medico_agent._get_allowed_tools()

        # gestor has more permissions
        gestor_ctx = self._create_request_context(role="gestor")
        gestor_agent = CopilotAgent(request_ctx=gestor_ctx)
        gestor_tools = gestor_agent._get_allowed_tools()

        # gestor should have more tools than medico
        assert len(gestor_tools) > len(medico_tools)

        # medico should NOT have send_team_notification
        medico_tool_names = [t.name for t in medico_tools]
        assert "send_team_notification" not in medico_tool_names

        # gestor should have send_team_notification
        gestor_tool_names = [t.name for t in gestor_tools]
        assert "send_team_notification" in gestor_tool_names

    def test_system_prompt_includes_user_context(self):
        """
        Test that the system prompt includes user-specific context.

        The prompt should include the user's email, role, and available tools.
        """
        from app.agents.copilot_agent import CopilotAgent

        request_ctx = self._create_request_context(role="operador")
        agent = CopilotAgent(request_ctx=request_ctx)

        system_prompt = agent._build_system_prompt()

        # Prompt should include user info
        assert "test@example.com" in system_prompt
        assert "operador" in system_prompt


class TestHumanInTheLoopFlow:
    """Test human-in-the-loop confirmation flow."""

    def _create_request_context(self, role: str = "operador"):
        """Create a mock RequestContext."""
        from app.middleware.auth import UserClaims
        from app.middleware.tenant import TenantContext, RequestContext

        user_id = str(uuid4())
        tenant_id = str(uuid4())

        user_claims = UserClaims(
            user_id=user_id,
            email="test@example.com",
            role=role,
            tenant_id=tenant_id,
            is_super_admin=False,
        )

        tenant_context = TenantContext(
            tenant_id=tenant_id,
            is_super_admin=False,
            effective_tenant_id=tenant_id,
        )

        return RequestContext(user=user_claims, tenant=tenant_context)

    @pytest.mark.asyncio
    async def test_tool_requiring_confirmation_returns_confirmation_request(self):
        """
        Test that tools requiring confirmation return confirmation_required=true.

        When a tool like update_occurrence_status is called without
        prior confirmation, it should return a confirmation request
        instead of executing the action.
        """
        from app.tools.base import ToolContext
        from app.tools.occurrence_tools import UpdateOccurrenceStatusTool

        request_ctx = self._create_request_context(role="operador")
        tool_ctx = ToolContext(
            request_ctx=request_ctx,
            confirmation_received=False,  # No confirmation yet
        )

        tool = UpdateOccurrenceStatusTool()

        # Run the tool without confirmation
        result = await tool.run(
            context=tool_ctx,
            occurrence_id="occ-123",
            new_status="em_andamento",
        )

        # Should return confirmation request
        assert result.success is True
        assert result.confirmation_required is True
        assert result.confirmation_action_id is not None
        assert "confirmation_details" in result.to_dict()

    @pytest.mark.asyncio
    async def test_tool_executes_after_confirmation_received(self):
        """
        Test that tools execute after confirmation is received.

        When confirmation_received=True, the tool should proceed with
        the actual execution instead of returning a confirmation request.
        """
        from app.tools.base import ToolContext
        from app.tools.occurrence_tools import UpdateOccurrenceStatusTool

        request_ctx = self._create_request_context(role="operador")
        tool_ctx = ToolContext(
            request_ctx=request_ctx,
            confirmation_received=True,  # Confirmation already given
        )

        tool = UpdateOccurrenceStatusTool()

        # Mock the backend API call
        with patch.object(tool, 'call_backend_api', new_callable=AsyncMock) as mock_api:
            mock_api.return_value = {
                "previous_status": "aberta",
                "new_status": "em_andamento",
            }

            result = await tool.run(
                context=tool_ctx,
                occurrence_id="occ-123",
                new_status="em_andamento",
            )

            # Should execute and return success
            assert result.success is True
            assert result.confirmation_required is False
            assert mock_api.called

    @pytest.mark.asyncio
    async def test_tool_without_confirmation_requirement_executes_directly(self):
        """
        Test that tools not requiring confirmation execute directly.

        Tools like list_occurrences should execute without asking
        for confirmation.
        """
        from app.tools.base import ToolContext
        from app.tools.occurrence_tools import ListOccurrencesTool

        request_ctx = self._create_request_context(role="operador")
        tool_ctx = ToolContext(
            request_ctx=request_ctx,
            confirmation_received=False,
        )

        tool = ListOccurrencesTool()

        # Mock the backend API call
        with patch.object(tool, 'call_backend_api', new_callable=AsyncMock) as mock_api:
            mock_api.return_value = {
                "data": [
                    {"id": "occ-1", "status": "aberta"},
                    {"id": "occ-2", "status": "em_andamento"},
                ],
                "total": 2,
            }

            result = await tool.run(context=tool_ctx)

            # Should execute directly without confirmation
            assert result.success is True
            assert result.confirmation_required is False
            assert result.data["total"] == 2
            assert mock_api.called
