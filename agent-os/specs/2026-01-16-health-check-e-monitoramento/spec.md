# Specification: Health Check e Monitoramento

## Goal
Implementar um sistema de health check agregado e dashboard de monitoramento em tempo real para o VitalConnect, permitindo que administradores visualizem o status de todos os servicos criticos (Database, Redis, Listener, API) atraves de uma interface visual tipo semaforo, com alertas automaticos por email quando o Obito Listener falhar.

## User Stories
- Como Admin do Sistema, quero visualizar o status de todos os servicos em uma unica pagina com indicadores visuais claros (verde/amarelo/vermelho), para identificar rapidamente problemas de infraestrutura.
- Como Admin do Sistema, quero receber alertas por email quando o Obito Listener cair, para poder agir imediatamente e evitar perda de notificacoes de doacao.

## Specific Requirements

**Endpoint Agregado de Health Check**
- Criar endpoint GET /api/health/summary (publico, sem autenticacao)
- Retornar JSON com status de: Database (PostgreSQL), Redis, Obito Listener, Triagem Motor, SSE Hub, API
- Cada componente deve incluir: status (up/degraded/down), latencia em ms, ultima verificacao
- Usar ping do database com timeout de 2 segundos para verificar PostgreSQL
- Usar PING command do Redis com timeout de 2 segundos
- Verificar heartbeat do Listener via Redis (chave vitalconnect:listener:heartbeat)
- Status geral: "healthy" se todos up, "degraded" se algum lento, "unhealthy" se algum down

**Sistema de Heartbeat do Listener**
- Listener deve publicar heartbeat no Redis a cada 5 segundos (SET vitalconnect:listener:heartbeat com TTL de 15s)
- Health check verifica existencia da chave - se ausente por 2 ciclos (20s), considera DOWN
- Armazenar timestamp do ultimo heartbeat para calcular tempo desde ultima atividade
- Implementar no ObitoListener.go como goroutine separada

**Dashboard de Status (/dashboard/status)**
- Criar pagina dedicada acessivel apenas para usuarios com role "admin"
- Exibir cards para cada servico com indicador visual tipo semaforo
- Verde: servico operacional (latencia < 500ms)
- Amarelo: latencia alta (500ms - 2000ms)
- Vermelho: servico fora do ar ou timeout
- Atualizar automaticamente a cada 10 segundos via polling
- Mostrar ultima atualizacao no header da pagina

**Componente ServiceStatusCard**
- Card com icone do servico, nome, indicador colorido (circulo), latencia, status textual
- Animacao de pulse quando status muda para vermelho
- Tooltip com detalhes adicionais (uptime, ultimo erro, etc)
- Usar componentes Card, Badge do shadcn/ui existentes

**Sistema de Alertas por Email**
- Enviar email para admin quando Listener transicionar de UP para DOWN
- Reutilizar EmailService existente em /backend/internal/services/notification/email.go
- Criar template HTML especifico para alerta de infraestrutura
- Nao enviar alertas repetidos - apenas na transicao de estado
- Cooldown de 5 minutos entre alertas do mesmo tipo

**Background Health Monitor Service**
- Criar servico Go que executa verificacoes a cada 10 segundos
- Armazenar ultimo estado conhecido de cada servico no Redis
- Detectar transicoes de estado e disparar alertas
- Expor metricas via endpoint /api/health/summary
- Iniciar junto com a aplicacao em main.go

**Integracao com SSE para Atualizacao Real-time**
- Opcional: publicar eventos de mudanca de status via SSE Hub existente
- Tipo de evento: "system_status_change"
- Payload: servico afetado, status anterior, novo status
- Permitir que dashboard receba atualizacoes sem polling

**Restricao de Acesso**
- Endpoint /api/health/summary deve ser publico (para load balancers)
- Pagina /dashboard/status requer autenticacao + role "admin"
- Usar middleware RequireRole("admin") existente

## Visual Design
Nenhum mockup foi fornecido. Seguir o padrao visual do dashboard existente com as seguintes diretrizes:

**Layout da Pagina /dashboard/status**
- Titulo: "Status do Sistema" com badge de ultima atualizacao
- Grid responsivo de cards (2 colunas mobile, 3 colunas desktop)
- Banner de alerta no topo se algum servico estiver DOWN
- Botao para forcar refresh manual

**Card de Status Individual**
- Icone do servico (Database, Server, Activity, etc)
- Nome do servico em texto bold
- Indicador circular colorido (24px) - verde/amarelo/vermelho
- Texto de status: "Operacional", "Latencia Alta", "Fora do Ar"
- Latencia em texto pequeno cinza (ex: "45ms")
- Borda do card muda de cor conforme status

## Existing Code to Leverage

**`/backend/internal/handlers/health.go`**
- Ja possui ListenerHealthResponse e ListenerDetails structs
- Funcao ListenerHealth() verifica status do listener e triagem motor
- Padrao de resposta JSON com status, timestamp, detalhes
- Estender para incluir verificacao de Database e Redis

**`/backend/internal/services/notification/email.go`**
- EmailService completo com suporte a TLS e templates HTML
- Metodo sendEmail() para envio via SMTP
- Reutilizar para enviar alertas de infraestrutura
- Criar novo template obitoNotificationTemplate estilo para alertas de sistema

**`/frontend/src/components/dashboard/StatusBadge.tsx`**
- Componente existente com logica de cores por status
- Padrao getStatusConfig() para mapear status -> variant/className
- Usar como referencia para criar ServiceStatusIndicator

**`/frontend/src/hooks/useSSE.tsx`**
- Hook completo para conexao SSE com reconexao automatica
- Padrao de estado: isConnected, lastEvent, etc
- Usar como referencia para criar useHealthStatus hook com polling

**`/backend/cmd/api/main.go`**
- Padrao de inicializacao de servicos background (listener, triagem, sseHub)
- Padrao de graceful shutdown com cancelBackground()
- Adicionar HealthMonitorService seguindo mesmo padrao

## Out of Scope
- Historico de longo prazo e graficos de CPU/memoria da semana
- Integracoes com PagerDuty, Slack, Discord ou outros sistemas externos
- Auto-scaling ou auto-recovery (Kubernetes/Docker Swarm)
- Metricas no formato Prometheus para integracao com Grafana
- Alertas via SMS ou Push para problemas de infraestrutura
- Acesso ao dashboard de saude para Gestores ou Operadores
- Verificacao de servicos externos (Twilio, Firebase, hospital DB)
- Dashboard de metricas de performance (requests/segundo, etc)
- Configuracao dinamica de thresholds via UI
- Testes de carga ou stress testing automatizado
