# Ideia Inicial: Health Check e Monitoramento

## Descricao

Feature de Health Check e Monitoramento para o VitalConnect - sistema de notificacao de doacao de orgaos.

## Requisitos Iniciais

### Funcionalidades Solicitadas:
- Endpoints de status dos servicos
- Alertas de falha do listener
- Dashboard de saude do sistema
- Provide health check endpoints for all services
- Alert when the death listener fails or stops responding
- Show system health dashboard
- Support ops/devops monitoring

### Contexto do Sistema:
- Backend (Go) com API REST usando Gin
- Frontend (Next.js) com React e shadcn/ui
- Redis para cache e filas
- PostgreSQL para persistencia
- Death Listener service (monitora obitos)
- SMS via Twilio, Push via Firebase
- Sistema critico para doacao de orgaos - downtime significa orgaos perdidos

### Servicos a Monitorar:
1. API Backend (Go/Gin)
2. PostgreSQL Database
3. Redis Cache/Queue
4. Obito Listener Service
5. Triagem Motor Service
6. SSE Hub (notificacoes real-time)
7. Email Queue Worker
8. Frontend (Next.js)

### Restricoes:
- Deadline de demo: 26 de Janeiro
- E para MVP mas precisa mostrar confiabilidade operacional
- Estimativa no roadmap: S (2-3 dias)

### Endpoints Existentes:
- GET /health - basico (ja existe)
- GET /api/v1/health/listener - status do listener (ja existe, protegido)
- GET /api/v1/health/sse - status do SSE hub (ja existe, protegido)
