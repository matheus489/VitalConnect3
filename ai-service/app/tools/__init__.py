"""
AI Tools module for SIDOT AI Assistant.

Contains tool definitions for function calling capabilities.
All tools inherit from BaseTool and provide:
- Permission validation before execution
- Audit logging integration
- Human-in-the-loop confirmation support

Available Tools:
- list_occurrences: Query occurrences with filters
- get_occurrence_details: Get specific occurrence with LGPD data
- update_occurrence_status: Change occurrence status (requires confirmation)
- send_team_notification: Send push/SMS to shift team
- generate_report: Generate PDF reports
- search_documentation: RAG-based documentation search
"""

from app.tools.base import (
    BaseTool,
    ToolContext,
    ToolResult,
    ToolError,
    ToolPermissionError,
    ToolExecutionError,
    BackendConnectionError,
)
from app.tools.occurrence_tools import (
    ListOccurrencesTool,
    GetOccurrenceDetailsTool,
    UpdateOccurrenceStatusTool,
    list_occurrences_tool,
    get_occurrence_details_tool,
    update_occurrence_status_tool,
)
from app.tools.notification_tools import (
    SendTeamNotificationTool,
    send_team_notification_tool,
)
from app.tools.report_tools import (
    GenerateReportTool,
    SearchDocumentationTool,
    generate_report_tool,
    search_documentation_tool,
)


# All tool instances for agent registration
ALL_TOOLS = [
    list_occurrences_tool,
    get_occurrence_details_tool,
    update_occurrence_status_tool,
    send_team_notification_tool,
    generate_report_tool,
    search_documentation_tool,
]


def get_all_tools() -> list[BaseTool]:
    """
    Get all available tool instances.

    Returns:
        List of all tool instances.
    """
    return ALL_TOOLS.copy()


def get_tool_by_name(name: str) -> BaseTool:
    """
    Get a tool instance by name.

    Args:
        name: The tool name.

    Returns:
        The tool instance.

    Raises:
        ValueError: If tool is not found.
    """
    for tool in ALL_TOOLS:
        if tool.name == name:
            return tool
    raise ValueError(f"Tool not found: {name}")


__all__ = [
    # Base classes and types
    "BaseTool",
    "ToolContext",
    "ToolResult",
    "ToolError",
    "ToolPermissionError",
    "ToolExecutionError",
    "BackendConnectionError",
    # Tool classes
    "ListOccurrencesTool",
    "GetOccurrenceDetailsTool",
    "UpdateOccurrenceStatusTool",
    "SendTeamNotificationTool",
    "GenerateReportTool",
    "SearchDocumentationTool",
    # Tool instances
    "list_occurrences_tool",
    "get_occurrence_details_tool",
    "update_occurrence_status_tool",
    "send_team_notification_tool",
    "generate_report_tool",
    "search_documentation_tool",
    # Utility functions
    "get_all_tools",
    "get_tool_by_name",
    "ALL_TOOLS",
]
