# Spec Requirements: Health Check e Monitoramento

## Initial Description

Feature de Health Check e Monitoramento para o VitalConnect - sistema de notificacao de doacao de orgaos.

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

## Requirements Discussion

### First Round Questions

**Q1:** Sobre a expansao dos endpoints de health check - devo criar endpoints individuais para cada servico (/health/db, /health/redis, /health/listener) ou um endpoint agregado que retorne o status de todos os componentes em uma unica resposta?

**Answer:** Endpoint Agregado (/api/health/summary). Retorno: JSON unico com status de todos os componentes (Database, Redis, Listener, API). Facilita Frontend (uma requisicao) e reduz carga.

**Q2:** Para os alertas de falha do listener - alem de registrar em log, qual o mecanismo de alerta preferido? Opcoes: a) Email para admin do sistema, b) Webhook para sistema externo, c) SMS/Push (reutilizando infraestrutura existente), d) Apenas log + indicador visual no dashboard.

**Answer:** Email para Ops + Alerta Visual no Dashboard. Se Listener cair, enviar email para Admin do Sistema. SMS/Push: NAO para alertas de infra - deixar exclusivo para transplantes (nao banalizar).

**Q3:** Para o dashboard de saude - deve ser uma pagina dedicada (/dashboard/status) ou integrado como widget na pagina principal do dashboard existente? Se dedicada, qual nivel de detalhe: a) Apenas status (online/offline), b) Status + metricas basicas (latencia, uptime), c) Status + metricas + graficos historicos.

**Answer:** Pagina Dedicada (/dashboard/status). Visual: "Semaforo" (Traffic Lights) - Verde: Operacional, Amarelo: Latencia Alta, Vermelho: Fora do Ar. Essencial para video: mostrar deteccao de queda em tempo real.

**Q4:** Sobre o formato das metricas - devo expor metricas no formato Prometheus (/metrics) para integracao futura com Grafana, ou apenas JSON simples para consumo do frontend?

**Answer:** Apenas JSON Simples. Frontend React consome JSON nativamente. Prometheus adiciona complexidade sem ganho visual para demo.

**Q5:** Quais usuarios devem ter acesso ao dashboard de saude? Opcoes: a) Apenas admin do sistema, b) Admin + gestores, c) Todos os usuarios autenticados.

**Answer:** Apenas Admin (Sistema). Gestores/Operadores nao precisam ver latencia Redis ou memoria. Visao de Infraestrutura apenas.

**Q6:** Qual a frequencia de verificacao dos servicos? Opcoes: a) A cada 10 segundos (mais responsivo, mais carga), b) A cada 30 segundos (equilibrado), c) A cada 1 minuto (menos carga, menos responsivo).

**Answer:** 10 Segundos (Agressivo para Demo). Real seria 30s ou 1min. Para video: luz vermelha imediatamente apos desligar servico.

**Q7:** Sobre a definicao de "falha" do listener - considerar como falha: a) Apenas se o processo parar completamente, b) Se nao processar eventos por X minutos, c) Se a conexao com o banco de dados do hospital for perdida, d) Combinacao das anteriores.

**Answer:** Perda de Conexao (Ping). Se nao responder Heartbeat (Ping Redis) por 2 ciclos (20s) = DOWN. Focar em queda de infraestrutura (mais facil de demonstrar).

**Q8:** Ha algo que explicitamente NAO deve ser incluido neste escopo (para evitar scope creep)?

**Answer:** Fora do Escopo:
- Historico de Longo Prazo (graficos de CPU da semana)
- Integracoes Externas (PagerDuty, Slack, Discord)
- Auto-Scaling/Auto-Recovery (Kubernetes/Docker Swarm)

### Existing Code to Reference

No similar existing features identified for reference by the user.

**Note:** Endpoints de health check ja existem no sistema:
- GET /health - basico
- GET /api/v1/health/listener - status do listener (protegido)
- GET /api/v1/health/sse - status do SSE hub (protegido)

Estes endpoints devem ser usados como base para a implementacao do endpoint agregado.

### Follow-up Questions

Nao foram necessarias perguntas de follow-up. As respostas do usuario foram completas e detalhadas.

## Visual Assets

### Files Provided:
Nenhum arquivo visual fornecido.

### Visual Insights:
N/A - Usuario confirmou que nao ha mockups ou wireframes para esta feature.

## Requirements Summary

### Functional Requirements

**Endpoint Agregado de Health Check:**
- Criar endpoint GET /api/health/summary
- Retornar JSON unico com status de todos os componentes:
  - Database (PostgreSQL)
  - Redis (Cache/Queue)
  - Listener (Obito Listener Service)
  - API (Backend Go/Gin)
- Incluir metricas basicas: status (up/down), latencia de conexao
- Apenas JSON simples, sem formato Prometheus

**Dashboard de Saude do Sistema:**
- Criar pagina dedicada em /dashboard/status
- Visual estilo "Semaforo" (Traffic Lights):
  - Verde: Operacional (servico respondendo normalmente)
  - Amarelo: Latencia Alta (servico lento mas respondendo)
  - Vermelho: Fora do Ar (servico nao responde)
- Atualizacao automatica a cada 10 segundos
- Mostrar deteccao de queda em tempo real (essencial para demo)

**Sistema de Alertas:**
- Enviar email para Admin do Sistema quando Listener cair
- Exibir alerta visual no Dashboard quando qualquer servico falhar
- NAO usar SMS/Push para alertas de infraestrutura (reservado para transplantes)

**Deteccao de Falha do Listener:**
- Implementar heartbeat via Redis Ping
- Considerar DOWN se nao responder por 2 ciclos consecutivos (20 segundos)
- Focar em queda de infraestrutura (perda de conexao)

**Restricoes de Acesso:**
- Dashboard de saude acessivel apenas para Admin do Sistema
- Gestores e Operadores nao devem ter acesso

### Reusability Opportunities

- Endpoints existentes de health check (/health, /api/v1/health/listener, /api/v1/health/sse) como base
- Sistema de email ja implementado para notificacoes (reutilizar para alertas de ops)
- Componentes shadcn/ui para construcao do dashboard
- Padroes de autenticacao/autorizacao JWT ja implementados

### Scope Boundaries

**In Scope:**
- Endpoint agregado de health check (/api/health/summary)
- Pagina de dashboard de status (/dashboard/status)
- Visual tipo semaforo para status dos servicos
- Alerta por email quando Listener cair
- Alerta visual no dashboard
- Verificacao a cada 10 segundos
- Deteccao de queda via heartbeat/ping (2 ciclos = 20s)
- Restricao de acesso apenas para Admin

**Out of Scope:**
- Historico de longo prazo (graficos de CPU/memoria da semana)
- Integracoes externas (PagerDuty, Slack, Discord)
- Auto-Scaling/Auto-Recovery (Kubernetes/Docker Swarm)
- Formato Prometheus para metricas
- Alertas SMS/Push para infraestrutura
- Acesso para Gestores/Operadores ao dashboard de saude

### Technical Considerations

- Backend em Go com Gin framework
- Frontend em Next.js com React e shadcn/ui
- Redis para heartbeat/ping do Listener
- PostgreSQL para persistencia
- JWT para autenticacao/autorizacao
- Email via SMTP institucional ou SendGrid
- Frequencia de verificacao: 10 segundos (otimizado para demo)
- Threshold de falha: 2 ciclos sem resposta (20 segundos)

### Demo Requirements

- Visualizacao clara de status (semaforo verde/amarelo/vermelho)
- Deteccao rapida de queda (10 segundos entre verificacoes)
- Demonstracao de mudanca de cor ao desligar servico
- Interface simples e impactante para video de apresentacao
- Deadline: 26 de Janeiro
