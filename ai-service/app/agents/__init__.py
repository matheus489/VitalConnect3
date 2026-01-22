"""
AI Agents module for SIDOT AI Assistant.

Contains agent configurations for the SIDOT co-pilot.

The SimpleAgent provides:
- Portuguese language support
- Context-aware system prompt
- Direct OpenAI integration (avoids llama_index Python 3.14 issues)
"""

from app.agents.simple_agent import (
    SimpleAgent,
    create_simple_agent,
    SYSTEM_PROMPT_PT,
)


__all__ = [
    "SimpleAgent",
    "create_simple_agent",
    "SYSTEM_PROMPT_PT",
]
