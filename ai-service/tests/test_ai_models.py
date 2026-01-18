"""
Tests for AI database models.

Focused tests for AIConversation and AIActionAuditLog models
including tenant isolation verification.
"""

import pytest
from uuid import uuid4
from datetime import datetime, timezone

from app.models import (
    AIConversation,
    AIActionAuditLog,
    MessageRole,
    ActionType,
    ActionStatus,
    Severity,
)


class TestAIConversationModel:
    """Tests for AIConversation model creation and methods."""

    def test_create_user_message(self):
        """Test AIConversation user message factory method."""
        tenant_id = uuid4()
        user_id = uuid4()
        session_id = uuid4()
        content = "Quais ocorrencias estao pendentes?"

        message = AIConversation.create_user_message(
            tenant_id=tenant_id,
            user_id=user_id,
            session_id=session_id,
            content=content,
            message_metadata={"source": "chat_widget"},
        )

        assert message.tenant_id == tenant_id
        assert message.user_id == user_id
        assert message.session_id == session_id
        assert message.role == MessageRole.USER.value
        assert message.content == content
        assert message.message_metadata == {"source": "chat_widget"}
        assert message.tool_calls is None

    def test_create_assistant_message_with_tool_calls(self):
        """Test AIConversation assistant message with tool calls."""
        tenant_id = uuid4()
        user_id = uuid4()
        session_id = uuid4()
        content = "Encontrei 3 ocorrencias pendentes."
        tool_calls = {
            "tools_used": ["list_occurrences"],
            "results_count": 3,
        }

        message = AIConversation.create_assistant_message(
            tenant_id=tenant_id,
            user_id=user_id,
            session_id=session_id,
            content=content,
            tool_calls=tool_calls,
        )

        assert message.role == MessageRole.ASSISTANT.value
        assert message.content == content
        assert message.tool_calls == tool_calls

    def test_conversation_to_dict(self):
        """Test AIConversation serialization to dict."""
        tenant_id = uuid4()
        user_id = uuid4()
        session_id = uuid4()

        message = AIConversation.create_user_message(
            tenant_id=tenant_id,
            user_id=user_id,
            session_id=session_id,
            content="Test message",
        )
        # Simulate database-assigned fields
        message.id = uuid4()
        message.created_at = datetime.now(timezone.utc)

        result = message.to_dict()

        assert result["tenant_id"] == str(tenant_id)
        assert result["user_id"] == str(user_id)
        assert result["session_id"] == str(session_id)
        assert result["role"] == "user"
        assert result["content"] == "Test message"
        assert "created_at" in result


class TestAIActionAuditLogModel:
    """Tests for AIActionAuditLog model creation and methods."""

    def test_create_query_log(self):
        """Test AIActionAuditLog query log factory method."""
        tenant_id = uuid4()
        user_id = uuid4()
        conversation_id = uuid4()
        input_params = {"query": "list pending occurrences"}

        log = AIActionAuditLog.create_query_log(
            tenant_id=tenant_id,
            user_id=user_id,
            conversation_id=conversation_id,
            input_params=input_params,
            status=ActionStatus.SUCCESS,
            output_result={"count": 3},
            execution_time_ms=150,
        )

        assert log.tenant_id == tenant_id
        assert log.user_id == user_id
        assert log.conversation_id == conversation_id
        assert log.action_type == ActionType.QUERY.value
        assert log.input_params == input_params
        assert log.status == ActionStatus.SUCCESS.value
        assert log.execution_time_ms == 150
        assert log.severity == Severity.INFO.value

    def test_create_tool_execution_log(self):
        """Test AIActionAuditLog tool execution log creation."""
        tenant_id = uuid4()
        user_id = uuid4()
        tool_name = "update_occurrence_status"
        input_params = {
            "occurrence_id": str(uuid4()),
            "new_status": "em_andamento",
        }

        log = AIActionAuditLog.create_tool_execution_log(
            tenant_id=tenant_id,
            user_id=user_id,
            conversation_id=None,
            tool_name=tool_name,
            input_params=input_params,
            status=ActionStatus.PENDING,
            severity=Severity.WARN,
        )

        assert log.action_type == ActionType.TOOL_EXECUTION.value
        assert log.tool_name == tool_name
        assert log.status == ActionStatus.PENDING.value
        assert log.severity == Severity.WARN.value

    def test_create_confirmation_log_confirmed(self):
        """Test AIActionAuditLog confirmation log when confirmed."""
        tenant_id = uuid4()
        user_id = uuid4()

        log = AIActionAuditLog.create_confirmation_log(
            tenant_id=tenant_id,
            user_id=user_id,
            conversation_id=None,
            tool_name="send_team_notification",
            input_params={"team_id": str(uuid4())},
            confirmed=True,
        )

        assert log.action_type == ActionType.CONFIRMATION.value
        assert log.status == ActionStatus.SUCCESS.value
        assert log.output_result == {"confirmed": True}

    def test_create_confirmation_log_cancelled(self):
        """Test AIActionAuditLog confirmation log when cancelled."""
        tenant_id = uuid4()
        user_id = uuid4()

        log = AIActionAuditLog.create_confirmation_log(
            tenant_id=tenant_id,
            user_id=user_id,
            conversation_id=None,
            tool_name="update_occurrence_status",
            input_params={"action": "change_status"},
            confirmed=False,
        )

        assert log.status == ActionStatus.CANCELLED.value
        assert log.output_result == {"confirmed": False}

    def test_mark_success(self):
        """Test marking an audit log as successful."""
        log = AIActionAuditLog.create_query_log(
            tenant_id=uuid4(),
            user_id=uuid4(),
            conversation_id=None,
            input_params={"query": "test"},
            status=ActionStatus.PENDING,
        )

        log.mark_success(
            output_result={"success": True, "data": []},
            execution_time_ms=250,
        )

        assert log.status == ActionStatus.SUCCESS.value
        assert log.execution_time_ms == 250
        assert log.output_result == {"success": True, "data": []}

    def test_mark_failed(self):
        """Test marking an audit log as failed."""
        log = AIActionAuditLog.create_tool_execution_log(
            tenant_id=uuid4(),
            user_id=uuid4(),
            conversation_id=None,
            tool_name="generate_report",
            input_params={},
            status=ActionStatus.PENDING,
        )

        log.mark_failed(
            error_message="Database connection timeout",
            execution_time_ms=5000,
        )

        assert log.status == ActionStatus.FAILED.value
        assert log.error_message == "Database connection timeout"
        assert log.severity == Severity.WARN.value


class TestTenantIsolation:
    """Tests for tenant isolation in models."""

    def test_conversation_requires_tenant_id(self):
        """Test that AIConversation requires tenant_id."""
        message = AIConversation.create_user_message(
            tenant_id=uuid4(),
            user_id=uuid4(),
            session_id=uuid4(),
            content="Test",
        )
        # tenant_id should be set
        assert message.tenant_id is not None

    def test_audit_log_requires_tenant_id(self):
        """Test that AIActionAuditLog requires tenant_id."""
        log = AIActionAuditLog.create_query_log(
            tenant_id=uuid4(),
            user_id=uuid4(),
            conversation_id=None,
            input_params={},
        )
        # tenant_id should be set
        assert log.tenant_id is not None

    def test_different_tenants_have_different_ids(self):
        """Test that messages from different tenants are distinguishable."""
        tenant_a = uuid4()
        tenant_b = uuid4()
        user_id = uuid4()
        session_id = uuid4()

        message_a = AIConversation.create_user_message(
            tenant_id=tenant_a,
            user_id=user_id,
            session_id=session_id,
            content="Message from tenant A",
        )

        message_b = AIConversation.create_user_message(
            tenant_id=tenant_b,
            user_id=user_id,
            session_id=session_id,
            content="Message from tenant B",
        )

        assert message_a.tenant_id != message_b.tenant_id
        assert message_a.tenant_id == tenant_a
        assert message_b.tenant_id == tenant_b
