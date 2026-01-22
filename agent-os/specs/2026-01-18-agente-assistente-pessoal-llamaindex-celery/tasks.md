# Task Breakdown: AI Assistant Agent (SIDOT Co-Pilot)

## Overview
Total Tasks: 11 Task Groups | ~55 Sub-tasks

This task breakdown implements a hybrid AI assistant using LlamaIndex + Celery as a Python microservice, integrated with the existing Go backend and React frontend.

## Task List

---

### Infrastructure Layer

#### Task Group 1: Python Microservice Setup
**Dependencies:** None

- [x] 1.0 Complete Python microservice infrastructure
  - [x] 1.1 Write 2-4 focused tests for service health and configuration
    - Test health endpoint returns correct status
    - Test configuration loading from environment
    - Test Redis connection initialization
    - Test PostgreSQL connection initialization
  - [x] 1.2 Create project structure for AI service
    - Create `/ai-service/` directory structure:
      ```
      /ai-service/
        /app/
          __init__.py
          main.py
          config.py
          /agents/
          /tools/
          /rag/
          /middleware/
          /models/
          /celery_app/
        /tests/
        requirements.txt
        Dockerfile
        .env.example
      ```
  - [x] 1.3 Create `requirements.txt` with dependencies
    - fastapi>=0.109.0
    - uvicorn>=0.27.0
    - llama-index>=0.10.0
    - celery>=5.3.0
    - redis>=5.0.0
    - psycopg2-binary>=2.9.9
    - sqlalchemy>=2.0.0
    - pyjwt>=2.8.0
    - python-dotenv>=1.0.0
    - httpx>=0.26.0
    - qdrant-client>=1.7.0 (or chromadb>=0.4.22)
    - openai>=1.10.0
    - pytest>=7.4.0
    - pytest-asyncio>=0.23.0
  - [x] 1.4 Create `config.py` with environment configuration
    - Load from environment variables
    - Settings: DATABASE_URL, REDIS_URL, JWT_SECRET, OPENAI_API_KEY
    - Tenant isolation settings
    - Vector store configuration
  - [x] 1.5 Create Dockerfile for AI service
    - Base: python:3.11-slim
    - Install dependencies
    - Copy application code
    - Expose port 8000
    - CMD: uvicorn app.main:app
  - [x] 1.6 Update `docker-compose.yml` with AI service
    - Add `ai-service` container
    - Add Qdrant/ChromaDB vector store container
    - Configure network connectivity with existing services
    - Environment variables for OpenAI, Redis, PostgreSQL
  - [x] 1.7 Create `main.py` FastAPI application entry point
    - Initialize FastAPI app
    - Include health endpoint at `/health`
    - Setup CORS middleware
    - Include routers (to be created)
  - [x] 1.8 Ensure infrastructure tests pass
    - Run ONLY the 2-4 tests written in 1.1
    - Verify Docker build succeeds
    - Verify health endpoint responds

**Files to Create/Modify:**
- Create: `/ai-service/app/__init__.py`
- Create: `/ai-service/app/main.py`
- Create: `/ai-service/app/config.py`
- Create: `/ai-service/requirements.txt`
- Create: `/ai-service/Dockerfile`
- Create: `/ai-service/.env.example`
- Modify: `/docker-compose.yml`

**Acceptance Criteria:**
- The 2-4 tests written in 1.1 pass
- Docker container builds successfully
- Health endpoint returns `{"status": "healthy"}`
- Service connects to Redis and PostgreSQL

---

#### Task Group 2: Celery Task Infrastructure
**Dependencies:** Task Group 1

- [x] 2.0 Complete Celery task infrastructure
  - [x] 2.1 Write 2-4 focused tests for Celery configuration
    - Test task registration
    - Test Redis broker connection
    - Test PostgreSQL result backend
    - Test task queue routing
  - [x] 2.2 Create Celery application configuration
    - Create `/ai-service/app/celery_app/__init__.py`
    - Create `/ai-service/app/celery_app/celery_config.py`
    - Configure Redis as broker: `redis://redis:6379/1`
    - Configure PostgreSQL as result backend
    - Key prefix: `sidot:ai:`
  - [x] 2.3 Configure task queues with priorities
    - `ai_query` queue (high priority) - for user queries
    - `ai_actions` queue (normal priority) - for tool executions
    - `ai_indexing` queue (low priority) - for document indexing
  - [x] 2.4 Implement task retry logic with exponential backoff
    - Max retries: 3
    - Backoff: 10s, 30s, 60s
    - Dead letter handling
  - [x] 2.5 Create base task class with audit logging
    - Log task start/end/error to PostgreSQL
    - Include user_id, tenant_id, execution_time
  - [x] 2.6 Create Celery worker Dockerfile/entrypoint
    - Separate worker process from API
    - Configure concurrency
  - [x] 2.7 Ensure Celery tests pass
    - Run ONLY the 2-4 tests written in 2.1
    - Verify worker starts successfully
    - Verify task routing works

**Files to Create:**
- Create: `/ai-service/app/celery_app/__init__.py`
- Create: `/ai-service/app/celery_app/celery_config.py`
- Create: `/ai-service/app/celery_app/tasks/__init__.py`
- Create: `/ai-service/app/celery_app/tasks/base.py`

**Acceptance Criteria:**
- The 2-4 tests written in 2.1 pass
- Celery worker connects to Redis
- Task results persist to PostgreSQL
- Queue routing works correctly

---

### Database Layer

#### Task Group 3: AI Database Models and Migrations
**Dependencies:** Task Group 1

- [x] 3.0 Complete AI database layer
  - [x] 3.1 Write 2-4 focused tests for AI models
    - Test AIConversation creation and retrieval
    - Test AIActionAuditLog creation
    - Test tenant isolation on queries
  - [x] 3.2 Create SQL migration for AI tables
    - Create migration file: `/backend/migrations/YYYYMMDD_ai_tables.sql`
    - Table: `ai_conversation_history`
      - id (UUID, PK)
      - tenant_id (UUID, FK to centrais)
      - user_id (UUID, FK to usuarios)
      - session_id (UUID, index)
      - role (VARCHAR: 'user' | 'assistant' | 'system')
      - content (TEXT)
      - tool_calls (JSONB, nullable)
      - metadata (JSONB, nullable)
      - created_at (TIMESTAMPTZ)
    - Table: `ai_action_audit_log`
      - id (UUID, PK)
      - tenant_id (UUID, FK)
      - user_id (UUID, FK)
      - conversation_id (UUID, FK to ai_conversation_history)
      - action_type (VARCHAR: 'ai.query', 'ai.tool_execution', 'ai.confirmation')
      - tool_name (VARCHAR, nullable)
      - input_params (JSONB)
      - output_result (JSONB)
      - status (VARCHAR: 'pending', 'success', 'failed', 'cancelled')
      - execution_time_ms (INTEGER)
      - error_message (TEXT, nullable)
      - severity (VARCHAR: 'INFO', 'WARN', 'CRITICAL')
      - created_at (TIMESTAMPTZ)
    - Add indexes for tenant_id, user_id, session_id, created_at
  - [x] 3.3 Create Python SQLAlchemy models
    - Create `/ai-service/app/models/conversation.py`
    - Create `/ai-service/app/models/audit_log.py`
    - Follow existing SIDOT model patterns
  - [x] 3.4 Create repository layer for AI models
    - Create `/ai-service/app/repository/conversation_repo.py`
    - Create `/ai-service/app/repository/audit_repo.py`
    - Implement CRUD with tenant filtering
  - [x] 3.5 Ensure database tests pass
    - Run ONLY the 2-4 tests written in 3.1
    - Verify migration runs successfully
    - Verify tenant isolation works

**Files to Create:**
- Create: `/backend/migrations/YYYYMMDD_ai_tables.sql`
- Create: `/ai-service/app/models/__init__.py`
- Create: `/ai-service/app/models/conversation.py`
- Create: `/ai-service/app/models/audit_log.py`
- Create: `/ai-service/app/repository/__init__.py`
- Create: `/ai-service/app/repository/conversation_repo.py`
- Create: `/ai-service/app/repository/audit_repo.py`

**Acceptance Criteria:**
- The 2-4 tests written in 3.1 pass
- Migration creates tables successfully
- Models support tenant isolation
- Audit logs capture required fields

---

### Security Layer

#### Task Group 4: Authentication and Multi-Tenant Middleware
**Dependencies:** Task Group 3

- [x] 4.0 Complete security middleware layer
  - [x] 4.1 Write 2-4 focused tests for security
    - Test JWT token validation
    - Test tenant context extraction
    - Test permission checking
    - Test super-admin context switch
  - [x] 4.2 Create JWT validation middleware
    - Create `/ai-service/app/middleware/auth.py`
    - Validate JWT token using same secret as Go backend
    - Extract UserClaims: user_id, email, role, tenant_id, is_super_admin
    - Handle TOKEN_EXPIRED, INVALID_TOKEN errors
    - Follow patterns from `/backend/internal/middleware/auth.go`
  - [x] 4.3 Create tenant context middleware
    - Create `/ai-service/app/middleware/tenant.py`
    - Replicate TenantContext logic from Go
    - Support X-Tenant-Context header for super-admin
    - Calculate effective_tenant_id
    - Follow patterns from `/backend/internal/middleware/tenant.go`
  - [x] 4.4 Create role-based permission checker
    - Create `/ai-service/app/middleware/permissions.py`
    - Define permission matrix per role
    - Check tool execution permissions
    - Roles: admin, gestor, operador, medico
  - [x] 4.5 Create request context dependency
    - Create `/ai-service/app/dependencies.py`
    - FastAPI dependency injection for current user
    - Dependency for tenant context
  - [x] 4.6 Ensure security tests pass
    - Run ONLY the 2-4 tests written in 4.1
    - Verify token validation works
    - Verify tenant isolation enforced

**Files to Create:**
- Create: `/ai-service/app/middleware/__init__.py`
- Create: `/ai-service/app/middleware/auth.py`
- Create: `/ai-service/app/middleware/tenant.py`
- Create: `/ai-service/app/middleware/permissions.py`
- Create: `/ai-service/app/dependencies.py`

**Acceptance Criteria:**
- The 2-4 tests written in 4.1 pass
- JWT validation matches Go backend behavior
- Tenant context properly isolated
- Role-based permissions enforced

---

### RAG Pipeline Layer

#### Task Group 5: LlamaIndex RAG Pipeline
**Dependencies:** Task Group 4

- [x] 5.0 Complete RAG pipeline
  - [x] 5.1 Write 2-4 focused tests for RAG functionality
    - Test document indexing
    - Test vector search with tenant filter
    - Test query response generation
  - [x] 5.2 Create vector store configuration
    - Create `/ai-service/app/rag/__init__.py`
    - Create `/ai-service/app/rag/vector_store.py`
    - Configure Qdrant/ChromaDB connection
    - Create tenant-isolated collections (one per tenant or metadata filtering)
  - [x] 5.3 Create document ingestion pipeline
    - Create `/ai-service/app/rag/ingestion.py`
    - Support markdown and PDF documents
    - Chunk documents with overlap
    - Add metadata: tenant_id, doc_type, source_path
    - Use multilingual-e5-large embeddings for Portuguese
  - [x] 5.4 Create Celery task for document indexing
    - Create `/ai-service/app/celery_app/tasks/indexing.py`
    - Background document processing
    - Update vector store asynchronously
    - Route to `ai_indexing` queue
  - [x] 5.5 Create RAG query engine
    - Create `/ai-service/app/rag/query_engine.py`
    - Configure LlamaIndex retriever
    - Filter by tenant_id metadata
    - Hybrid search (vector + keyword)
    - Response synthesis with context
  - [x] 5.6 Create documentation index management
    - Create `/ai-service/app/rag/doc_manager.py`
    - CRUD for indexed documents
    - Re-indexing capability
  - [x] 5.7 Ensure RAG tests pass
    - Run ONLY the 2-4 tests written in 5.1
    - Verify document indexing works
    - Verify tenant-filtered search works

**Files to Create:**
- Create: `/ai-service/app/rag/__init__.py`
- Create: `/ai-service/app/rag/vector_store.py`
- Create: `/ai-service/app/rag/ingestion.py`
- Create: `/ai-service/app/rag/query_engine.py`
- Create: `/ai-service/app/rag/doc_manager.py`
- Create: `/ai-service/app/celery_app/tasks/indexing.py`

**Acceptance Criteria:**
- The 2-4 tests written in 5.1 pass
- Documents indexed with tenant metadata
- Vector search respects tenant isolation
- Portuguese queries work correctly

---

### Agent & Tools Layer

#### Task Group 6: LlamaIndex Agent and Tool Definitions
**Dependencies:** Task Group 5

- [x] 6.0 Complete agent and tools layer
  - [x] 6.1 Write 2-4 focused tests for agent and tools
    - Test tool execution with permission check
    - Test agent response generation
    - Test human-in-the-loop flow
  - [x] 6.2 Create base tool class with permission checking
    - Create `/ai-service/app/tools/__init__.py`
    - Create `/ai-service/app/tools/base.py`
    - Abstract base class for all tools
    - Permission validation before execution
    - Audit logging integration
  - [x] 6.3 Implement `list_occurrences` tool
    - Create `/ai-service/app/tools/occurrence_tools.py`
    - Query occurrences via Go backend API
    - Filters: status, hospital_id, date_range
    - Return structured data for UI rendering
    - Permissions: all authenticated users
  - [x] 6.4 Implement `get_occurrence_details` tool
    - Get specific occurrence with LGPD data
    - Requires elevated role (operador+)
    - Audit log access to sensitive data
  - [x] 6.5 Implement `update_occurrence_status` tool
    - Create confirmation requirement flag
    - Return confirmation_required=true for human-in-the-loop
    - Execute only after confirmation received
    - Permissions: operador+
  - [x] 6.6 Implement `send_team_notification` tool
    - Create `/ai-service/app/tools/notification_tools.py`
    - Query shift schedule from Go backend
    - Send push/SMS via existing notification service
    - Permissions: gestor+
  - [x] 6.7 Implement `generate_report` tool
    - Create `/ai-service/app/tools/report_tools.py`
    - Trigger report generation via Go backend
    - Support daily/weekly report types
    - Return PDF download URL
    - Permissions: gestor+
  - [x] 6.8 Implement `search_documentation` tool
    - Wrapper around RAG query engine
    - Return relevant documentation chunks
    - Permissions: all authenticated users
  - [x] 6.9 Create LlamaIndex agent configuration
    - Create `/ai-service/app/agents/__init__.py`
    - Create `/ai-service/app/agents/copilot_agent.py`
    - Configure LlamaIndex ReActAgent
    - Register all tools
    - System prompt for SIDOT context
    - Portuguese language support
  - [x] 6.10 Ensure agent tests pass
    - Run ONLY the 2-4 tests written in 6.1
    - Verify tool execution works
    - Verify agent responds appropriately

**Files to Create:**
- Create: `/ai-service/app/tools/__init__.py`
- Create: `/ai-service/app/tools/base.py`
- Create: `/ai-service/app/tools/occurrence_tools.py`
- Create: `/ai-service/app/tools/notification_tools.py`
- Create: `/ai-service/app/tools/report_tools.py`
- Create: `/ai-service/app/agents/__init__.py`
- Create: `/ai-service/app/agents/copilot_agent.py`

**Acceptance Criteria:**
- The 2-4 tests written in 6.1 pass
- All 6 tools implemented and registered
- Permission checks enforced per tool
- Human-in-the-loop works for critical actions

---

### API Layer

#### Task Group 7: FastAPI Endpoints
**Dependencies:** Task Group 6

- [x] 7.0 Complete FastAPI API layer
  - [x] 7.1 Write 2-4 focused tests for API endpoints
    - Test chat endpoint with valid request
    - Test confirmation endpoint
    - Test conversation history retrieval
  - [x] 7.2 Create chat endpoint router
    - Create `/ai-service/app/routers/__init__.py`
    - Create `/ai-service/app/routers/chat.py`
    - POST `/api/v1/ai/chat` - Send message to agent
    - Request: `{ message: string, session_id?: string }`
    - Response: `{ response: string, tool_calls?: [], confirmation_required?: {} }`
  - [x] 7.3 Create streaming chat endpoint
    - POST `/api/v1/ai/chat/stream` - SSE streaming response
    - Event types: `thinking`, `tool_call`, `response_chunk`, `done`
    - Use Server-Sent Events for real-time updates
  - [x] 7.4 Create confirmation endpoint
    - POST `/api/v1/ai/confirm/{action_id}` - Confirm pending action
    - Request: `{ confirmed: boolean }`
    - Execute or cancel the pending tool call
  - [x] 7.5 Create conversation history endpoints
    - GET `/api/v1/ai/conversations` - List user's conversations
    - GET `/api/v1/ai/conversations/{session_id}` - Get conversation messages
    - DELETE `/api/v1/ai/conversations/{session_id}` - Clear conversation
  - [x] 7.6 Create document management endpoints (admin)
    - POST `/api/v1/ai/documents/index` - Trigger document indexing
    - GET `/api/v1/ai/documents` - List indexed documents
    - DELETE `/api/v1/ai/documents/{id}` - Remove document from index
  - [x] 7.7 Create Celery task for async query processing
    - Create `/ai-service/app/celery_app/tasks/query.py`
    - Process chat messages asynchronously
    - Route to `ai_query` queue
  - [x] 7.8 Register routers in main.py
    - Include chat router
    - Include document router (admin only)
    - Apply auth middleware
  - [x] 7.9 Ensure API tests pass
    - Run ONLY the 2-4 tests written in 7.1
    - Verify endpoints respond correctly
    - Verify authentication required

**Files to Create:**
- Create: `/ai-service/app/routers/__init__.py`
- Create: `/ai-service/app/routers/chat.py`
- Create: `/ai-service/app/routers/documents.py`
- Create: `/ai-service/app/celery_app/tasks/query.py`
- Modify: `/ai-service/app/main.py`

**Acceptance Criteria:**
- The 2-4 tests written in 7.1 pass
- Chat endpoints work with authentication
- Streaming responses function correctly
- Confirmation flow works end-to-end

---

### Integration Layer

#### Task Group 8: Go Backend Gateway Integration
**Dependencies:** Task Group 7

- [x] 8.0 Complete Go backend gateway
  - [x] 8.1 Write 2-4 focused tests for gateway
    - Test proxy endpoint forwards correctly
    - Test JWT token forwarding
    - Test tenant context forwarding
  - [x] 8.2 Create AI proxy handler in Go backend
    - Create `/backend/internal/handlers/ai_proxy.go`
    - Proxy requests to Python AI service
    - Forward JWT token in Authorization header
    - Forward X-Tenant-Context header
    - Handle SSE streaming proxy
  - [x] 8.3 Create AI service client
    - Create `/backend/internal/integration/ai_client.go`
    - HTTP client for AI service
    - Configuration: AI_SERVICE_URL env var
    - Timeout and retry logic
  - [x] 8.4 Register AI routes in Go backend
    - Modify `/backend/cmd/api/routes.go`
    - Route group: `/api/v1/ai/*`
    - Apply AuthRequired, TenantContext middleware
  - [x] 8.5 Extend SSE event model for AI responses
    - Modify `/backend/internal/models/notification.go` (or create new)
    - Add event type: `ai_response_chunk`
    - Add AI-specific event fields
  - [x] 8.6 Ensure gateway tests pass
    - Run ONLY the 2-4 tests written in 8.1
    - Verify proxy works correctly
    - Verify headers forwarded

**Files to Create:**
- Create: `/backend/internal/handlers/ai_proxy.go`
- Create: `/backend/internal/integration/ai_client.go`
- Modify: `/backend/cmd/api/routes.go` (if exists) or router file
- Modify: `/backend/internal/models/notification.go`

**Acceptance Criteria:**
- The 2-4 tests written in 8.1 pass
- Go backend proxies to Python service
- Authentication context preserved
- SSE streaming works through gateway

---

### Frontend Layer

#### Task Group 9: Chat Widget and Components
**Dependencies:** Task Group 8

- [x] 9.0 Complete frontend chat UI
  - [x] 9.1 Write 2-4 focused tests for chat components
    - Test ChatWidget renders and toggles
    - Test message sending
    - Test confirmation dialog interaction
  - [x] 9.2 Create AI API client extension
    - Modify `/frontend/src/lib/api.ts`
    - Add AI endpoints: sendMessage, confirmAction, getHistory
    - Add SSE connection for streaming responses
  - [x] 9.3 Create ChatWidget floating action button
    - Create `/frontend/src/components/ai/ChatWidget.tsx`
    - FAB positioned at bottom-right (fixed position)
    - Click to expand side panel
    - Minimize/maximize controls
    - Badge for unread messages
  - [x] 9.4 Create ChatPanel side panel component
    - Create `/frontend/src/components/ai/ChatPanel.tsx`
    - Width: 400px, slides from right
    - Does not block main content
    - Header with minimize button
    - Message list area
    - Input area at bottom
  - [x] 9.5 Create message bubble components
    - Create `/frontend/src/components/ai/MessageBubble.tsx`
    - User message style (right-aligned, primary color)
    - Assistant message style (left-aligned, neutral color)
    - Support for rich content rendering
    - Timestamp display
  - [x] 9.6 Create ThinkingIndicator component
    - Create `/frontend/src/components/ai/ThinkingIndicator.tsx`
    - Animated dots or spinner
    - Display current step: "Consultando Base de Conhecimento..."
    - Show tool being executed
  - [x] 9.7 Create useAIChat hook
    - Create `/frontend/src/hooks/useAIChat.ts`
    - Manage chat state
    - Handle SSE streaming
    - Manage conversation session
    - Integrate with existing `useSSE` pattern
  - [x] 9.8 Ensure UI tests pass
    - Run ONLY the 2-4 tests written in 9.1
    - Verify components render
    - Verify interactions work

**Files to Create:**
- Create: `/frontend/src/components/ai/ChatWidget.tsx`
- Create: `/frontend/src/components/ai/ChatPanel.tsx`
- Create: `/frontend/src/components/ai/MessageBubble.tsx`
- Create: `/frontend/src/components/ai/ThinkingIndicator.tsx`
- Create: `/frontend/src/hooks/useAIChat.ts`
- Modify: `/frontend/src/lib/api.ts`

**Acceptance Criteria:**
- The 2-4 tests written in 9.1 pass
- Chat widget toggles open/closed
- Messages display correctly
- Streaming responses update in real-time

---

#### Task Group 10: Generative UI Components
**Dependencies:** Task Group 9

- [x] 10.0 Complete generative UI components
  - [x] 10.1 Write 2-4 focused tests for generative components
    - Test OccurrenceCard rendering
    - Test OccurrenceTable rendering
    - Test ConfirmationDialog flow
  - [x] 10.2 Create OccurrenceCard component
    - Create `/frontend/src/components/ai/OccurrenceCard.tsx`
    - Mini card for single occurrence in chat
    - Display: hospital, status, patient (masked), time remaining
    - Click to view details action
    - Follow existing card patterns from `/frontend/src/components/ui/card.tsx`
  - [x] 10.3 Create OccurrenceTable component
    - Create `/frontend/src/components/ai/OccurrenceTable.tsx`
    - Interactive table for listing occurrences
    - Sortable columns
    - Row click for details
    - Follow existing table patterns from `/frontend/src/components/ui/table.tsx`
  - [x] 10.4 Create ConfirmationDialog component
    - Create `/frontend/src/components/ai/ConfirmationDialog.tsx`
    - Modal for human-in-the-loop confirmation
    - Display action details clearly
    - Confirm and Cancel buttons
    - Follow existing dialog patterns from `/frontend/src/components/ui/dialog.tsx`
  - [x] 10.5 Create ActionButton component
    - Create `/frontend/src/components/ai/ActionButton.tsx`
    - Styled buttons for in-message actions
    - Variants: primary (confirm), secondary (cancel), link (view details)
    - Loading state support
  - [x] 10.6 Create message content renderer
    - Create `/frontend/src/components/ai/MessageContent.tsx`
    - Parse assistant response for component markers
    - Render appropriate component (card, table, buttons)
    - Fallback to text for unknown content
  - [x] 10.7 Integrate generative components into ChatPanel
    - Update MessageBubble to use MessageContent
    - Handle component interactions
    - Route confirmation actions
  - [x] 10.8 Ensure generative UI tests pass
    - Run ONLY the 2-4 tests written in 10.1
    - Verify components render correctly
    - Verify interactions work

**Files to Create:**
- Create: `/frontend/src/components/ai/OccurrenceCard.tsx`
- Create: `/frontend/src/components/ai/OccurrenceTable.tsx`
- Create: `/frontend/src/components/ai/ConfirmationDialog.tsx`
- Create: `/frontend/src/components/ai/ActionButton.tsx`
- Create: `/frontend/src/components/ai/MessageContent.tsx`
- Modify: `/frontend/src/components/ai/MessageBubble.tsx`
- Modify: `/frontend/src/components/ai/ChatPanel.tsx`

**Acceptance Criteria:**
- The 2-4 tests written in 10.1 pass
- OccurrenceCard displays correctly in chat
- OccurrenceTable is interactive
- Confirmation dialog blocks until user action

---

### Testing & Integration

#### Task Group 11: Test Review & Integration Testing
**Dependencies:** Task Groups 1-10

- [x] 11.0 Review existing tests and fill critical gaps
  - [x] 11.1 Review tests from Task Groups 1-10
    - Review the 2-4 tests written per group (approximately 24-40 tests total)
    - Document coverage of critical paths
    - Identify end-to-end workflow gaps
  - [x] 11.2 Analyze test coverage gaps for AI feature
    - Focus ONLY on gaps related to AI assistant functionality
    - Prioritize integration points between services
    - Identify critical user workflows lacking coverage:
      - User sends message -> Agent responds
      - User triggers tool -> Confirmation -> Execution
      - Multi-turn conversation flow
  - [x] 11.3 Write up to 10 additional strategic tests
    - End-to-end: User query through full stack
    - Integration: Python service <-> Go backend
    - Integration: Frontend <-> Go gateway <-> Python
    - Security: Tenant isolation verification
    - Security: Permission enforcement across stack
    - Error handling: Service unavailability
  - [x] 11.4 Run AI feature-specific tests only
    - Run all tests from Task Groups 1-10
    - Run the 10 additional tests from 11.3
    - Expected total: approximately 34-50 tests
    - Do NOT run entire application test suite
  - [x] 11.5 Create integration test documentation
    - Document test scenarios
    - Document how to run AI-specific tests
    - Document known limitations

**Files to Create:**
- Create: `/ai-service/tests/integration/test_full_flow.py`
- Create: `/ai-service/tests/integration/test_security.py`
- Create: `/frontend/src/components/ai/__tests__/integration.test.tsx`

**Acceptance Criteria:**
- All feature-specific tests pass (approximately 34-50 tests)
- Critical user workflows covered
- Integration points verified
- Security tests confirm tenant isolation

---

## Execution Order

Recommended implementation sequence with parallelization opportunities:

```
Phase 1 - Foundation (Sequential)
└── Task Group 1: Python Microservice Setup

Phase 2 - Core Infrastructure (Parallel)
├── Task Group 2: Celery Task Infrastructure
├── Task Group 3: AI Database Models
└── Task Group 4: Security Middleware

Phase 3 - Intelligence Layer (Sequential)
├── Task Group 5: LlamaIndex RAG Pipeline
└── Task Group 6: Agent and Tool Definitions

Phase 4 - API & Integration (Sequential)
├── Task Group 7: FastAPI Endpoints
└── Task Group 8: Go Backend Gateway

Phase 5 - Frontend (Sequential)
├── Task Group 9: Chat Widget Components
└── Task Group 10: Generative UI Components

Phase 6 - Validation (Sequential)
└── Task Group 11: Test Review & Integration
```

## Technical Notes

### Environment Variables Required
```
# AI Service
OPENAI_API_KEY=sk-...
AI_MODEL=gpt-4o
EMBEDDING_MODEL=multilingual-e5-large

# Vector Store
QDRANT_URL=http://qdrant:6333
QDRANT_COLLECTION_PREFIX=sidot_

# Connections (reuse existing)
DATABASE_URL=postgresql://postgres:postgres@postgres:5432/sidot
REDIS_URL=redis://redis:6379/1
JWT_SECRET=${existing_jwt_secret}

# Go Backend
AI_SERVICE_URL=http://ai-service:8000
```

### Key Integration Points
1. **Go Backend -> Python AI Service**: HTTP proxy with JWT forwarding
2. **Python AI Service -> Go Backend**: HTTP client for data operations
3. **Frontend -> Go Backend**: Existing API client extended
4. **SSE Streaming**: Go proxies SSE from Python to Frontend

### Permission Matrix
| Tool | admin | gestor | operador | medico |
|------|-------|--------|----------|--------|
| list_occurrences | Yes | Yes | Yes | Yes |
| get_occurrence_details | Yes | Yes | Yes | Yes |
| update_occurrence_status | Yes | Yes | Yes | No |
| send_team_notification | Yes | Yes | No | No |
| generate_report | Yes | Yes | No | No |
| search_documentation | Yes | Yes | Yes | Yes |

### Files Summary

**New Files to Create:**
- `/ai-service/` - Entire Python microservice (~25 files)
- `/frontend/src/components/ai/` - 8 React components
- `/frontend/src/hooks/useAIChat.ts`
- `/backend/internal/handlers/ai_proxy.go`
- `/backend/internal/integration/ai_client.go`
- `/backend/migrations/YYYYMMDD_ai_tables.sql`

**Existing Files to Modify:**
- `/docker-compose.yml` - Add AI service and vector store
- `/frontend/src/lib/api.ts` - Add AI endpoints
- `/backend/cmd/api/routes.go` - Add AI route group
- `/backend/internal/models/notification.go` - Add AI event types
