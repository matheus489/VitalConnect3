# Specification: AI Assistant Agent (VitalConnect Co-Pilot)

## Goal
Build a hybrid AI assistant (Q&A via RAG + Function Calling for actions) that serves as an operational co-pilot for VitalConnect users, enabling natural language queries about documentation and executing system actions while respecting multi-tenant isolation and role-based permissions.

## User Stories
- As an operator, I want to ask natural language questions about organ capture procedures so that I can quickly find answers without searching documentation manually
- As a manager, I want to request reports and trigger notifications to my team through natural language commands so that I can work more efficiently
- As an administrator, I want the AI to respect role-based permissions so that users can only execute actions appropriate to their access level

## Specific Requirements

**Python Microservice Architecture**
- Create a standalone Python service using FastAPI + LlamaIndex + Celery
- Service communicates with Go backend via internal HTTP API (Go acts as gateway)
- Use Docker Compose for local development with Redis and PostgreSQL
- Implement health check endpoint at `/health` following existing patterns
- Structure: `/ai-service/app/main.py`, `/ai-service/app/agents/`, `/ai-service/app/tools/`, `/ai-service/app/rag/`

**LlamaIndex RAG Pipeline**
- Index VitalConnect documentation (markdown files, PDFs) into vector store
- Use Qdrant or ChromaDB as vector store with tenant-isolated collections
- Implement metadata filtering by `tenant_id` on all vector queries
- Create ingestion pipeline for document updates via Celery background tasks
- Support Portuguese language queries with appropriate embedding model (multilingual-e5-large or similar)

**Celery Task Workers**
- Configure Celery with Redis as broker and PostgreSQL as result backend
- Create task queues: `ai_query` (high priority), `ai_actions` (normal), `ai_indexing` (low)
- Implement task retry logic with exponential backoff
- All task results persisted to PostgreSQL for audit compliance

**Tool/Function Definitions for Actions**
- `list_occurrences`: Query occurrences with filters (status, hospital, date range)
- `get_occurrence_details`: Retrieve specific occurrence with LGPD data (requires role check)
- `update_occurrence_status`: Change occurrence status with human-in-the-loop confirmation
- `send_team_notification`: Send push/SMS to shift team members
- `generate_report`: Create daily/weekly reports in PDF format
- `search_documentation`: RAG-based documentation search
- Each tool must validate user permissions before execution

**Redis Broker Configuration**
- Redis connection reuses existing VitalConnect Redis instance
- Separate key prefix for AI service: `vitalconnect:ai:`
- Configure message TTL of 1 hour for task messages
- Implement connection pooling with max 10 connections

**PostgreSQL Audit Logging**
- Create `ai_conversation_history` table for conversation persistence
- Create `ai_action_audit_log` table extending existing audit pattern
- Log: user_id, tenant_id, prompt, response, tool_calls, execution_time, timestamp
- Conversations linked to user session for context retrieval

## Visual Design

No visual mockups provided. Frontend implementation should follow existing VitalConnect component patterns.

**Chat Widget Design Requirements**
- Floating action button (FAB) positioned at bottom-right corner
- Expands to side panel (400px width) that does not block main content
- Message bubbles with distinct styling for user vs assistant
- Support for rich content: tables, cards, action buttons
- "Thinking" state with animated indicator showing current step
- Minimize/maximize controls

**Generative UI Components**
- `<OccurrenceCard>`: Mini occurrence card rendered in chat responses
- `<OccurrenceTable>`: Interactive table for listing occurrences
- `<ConfirmationDialog>`: Modal for human-in-the-loop action confirmation
- `<ActionButton>`: Styled buttons for confirm/cancel actions within messages

## Existing Code to Leverage

**JWT Authentication Middleware (`/backend/internal/middleware/auth.go`)**
- Reuse `UserClaims` structure containing user_id, tenant_id, role, is_super_admin
- Python service receives validated JWT token from Go gateway
- Extract claims using same JWT validation library (PyJWT with same secret)
- Follow same error response patterns (TOKEN_EXPIRED, INVALID_TOKEN)

**Multi-Tenant Context (`/backend/internal/middleware/tenant.go`)**
- Replicate `TenantContext` logic in Python middleware
- Support `X-Tenant-Context` header for super-admin context switching
- All database queries must filter by `effective_tenant_id`
- Use existing tenant validation patterns

**SSE Notification System (`/backend/internal/services/notification/sse.go`)**
- Leverage existing SSE infrastructure for streaming AI responses
- Extend `SSEEvent` model with new type: `ai_response_chunk`
- Use same Redis Pub/Sub channel pattern with new prefix
- Frontend can reuse `useSSE` hook with extended event types

**Audit Log Model (`/backend/internal/models/audit_log.go`)**
- Follow existing `AuditLog` structure for AI action logging
- Use same severity levels: INFO, WARN, CRITICAL
- Add new action types: `ai.query`, `ai.tool_execution`, `ai.confirmation`
- Add new entity type: `AIConversation`

**Frontend API Client (`/frontend/src/lib/api.ts`)**
- Extend existing axios instance with AI service endpoints
- Reuse token management (getAccessToken, setTokens)
- Follow same interceptor patterns for token refresh

## Out of Scope
- Voice-to-text streaming (WebSocket complexity); only support audio blob upload for Whisper transcription
- Autonomous clinical decisions; AI presents data but never decides organ viability
- Direct database writes from Python service; all mutations go through Go API
- Real-time collaboration features between multiple users in same conversation
- Custom LLM fine-tuning; use base models with prompt engineering
- Mobile-specific UI; responsive design only for web
- Multi-language support beyond Portuguese
- Integration with external hospital systems (HL7, FHIR)
- Offline mode or local model deployment (Ollama is for testing only, not production)
- Agent memory across different user sessions (memory is per-user only)
