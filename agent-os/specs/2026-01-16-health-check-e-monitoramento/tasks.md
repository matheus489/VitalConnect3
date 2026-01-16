# Detalhamento de Tarefas: Health Check e Monitoramento

## Visao Geral
**Total de Tarefas:** 28
**Estimativa:** 2-3 dias (tamanho S)
**Deadline:** 26 de Janeiro de 2026

## Lista de Tarefas

### Camada Backend - Servicos

#### Grupo de Tarefas 1: Sistema de Heartbeat do Listener
**Dependencias:** Nenhuma

- [ ] 1.0 Completar sistema de heartbeat do ObitoListener
  - [ ] 1.1 Escrever 2-4 testes focados para funcionalidade de heartbeat
    - Testar publicacao de heartbeat no Redis com TTL correto (15s)
    - Testar deteccao de listener DOWN apos ausencia de heartbeat por 20s
    - Limitar a 4 testes maximos nesta etapa
  - [ ] 1.2 Implementar goroutine de heartbeat no ObitoListener
    - Arquivo: `/backend/internal/services/listener/obito_listener.go`
    - Publicar heartbeat a cada 5 segundos
    - Chave Redis: `vitalconnect:listener:heartbeat`
    - TTL: 15 segundos
    - Incluir timestamp do ultimo heartbeat
  - [ ] 1.3 Adicionar metodo para verificar status do heartbeat
    - Verificar existencia da chave no Redis
    - Retornar tempo desde ultimo heartbeat
    - Considerar DOWN se ausente por 2 ciclos (20s)
  - [ ] 1.4 Garantir que os testes do heartbeat passem
    - Executar APENAS os 2-4 testes escritos em 1.1
    - NAO executar suite completa de testes

**Criterios de Aceitacao:**
- Os 2-4 testes escritos em 1.1 passam
- Heartbeat publicado no Redis a cada 5 segundos
- TTL configurado para 15 segundos
- Deteccao correta de estado DOWN

---

#### Grupo de Tarefas 2: Health Monitor Service (Background)
**Dependencias:** Grupo de Tarefas 1

- [ ] 2.0 Completar servico de monitoramento em background
  - [ ] 2.1 Escrever 3-5 testes focados para HealthMonitorService
    - Testar verificacao de status do PostgreSQL (ping com timeout 2s)
    - Testar verificacao de status do Redis (PING command)
    - Testar deteccao de transicao de estado (UP -> DOWN)
    - Testar disparo de alerta na transicao
    - Limitar a 5 testes maximos
  - [ ] 2.2 Criar estrutura do HealthMonitorService
    - Arquivo: `/backend/internal/services/health/monitor.go`
    - Seguir padrao de inicializacao do ObitoListener
    - Campos: db, redis, emailService, lastKnownStates, alertCooldowns
  - [ ] 2.3 Implementar verificadores individuais de servicos
    - `checkDatabase()` - ping com timeout 2s
    - `checkRedis()` - PING command com timeout 2s
    - `checkListener()` - verificar chave de heartbeat
    - `checkTriagemMotor()` - verificar via globalTriagemMotor
    - `checkSSEHub()` - verificar via globalSSEHub
    - Cada verificador retorna: status (up/degraded/down), latencia em ms
  - [ ] 2.4 Implementar loop de verificacao periodica
    - Executar a cada 10 segundos
    - Armazenar ultimo estado conhecido no Redis
    - Detectar transicoes de estado
    - Disparar alertas quando Listener transicionar para DOWN
  - [ ] 2.5 Implementar controle de cooldown de alertas
    - Cooldown de 5 minutos entre alertas do mesmo tipo
    - Armazenar timestamp do ultimo alerta por servico
    - Apenas alertar na transicao UP -> DOWN
  - [ ] 2.6 Garantir que os testes do HealthMonitorService passem
    - Executar APENAS os 3-5 testes escritos em 2.1
    - NAO executar suite completa

**Criterios de Aceitacao:**
- Os 3-5 testes escritos em 2.1 passam
- Verificacoes executam a cada 10 segundos
- Transicoes de estado detectadas corretamente
- Cooldown de alertas funciona conforme especificado

---

### Camada Backend - API

#### Grupo de Tarefas 3: Endpoint Agregado de Health Check
**Dependencias:** Grupo de Tarefas 2

- [ ] 3.0 Completar endpoint agregado de health check
  - [ ] 3.1 Escrever 3-5 testes focados para o endpoint /api/health/summary
    - Testar resposta JSON com todos os componentes
    - Testar status geral "healthy" quando todos up
    - Testar status geral "unhealthy" quando algum down
    - Testar que endpoint e publico (sem autenticacao)
    - Limitar a 5 testes maximos
  - [ ] 3.2 Criar structs de resposta para health summary
    - Arquivo: `/backend/internal/handlers/health.go`
    - Estender structs existentes (ListenerHealthResponse)
    - `HealthSummaryResponse`: status geral, timestamp, componentes
    - `ComponentStatus`: nome, status, latencia, ultima verificacao
  - [ ] 3.3 Implementar handler HealthSummary
    - GET /api/health/summary (publico, sem autenticacao)
    - Coletar status de: Database, Redis, Listener, TriagemMotor, SSEHub, API
    - Calcular status geral: healthy/degraded/unhealthy
    - Retornar JSON estruturado
  - [ ] 3.4 Registrar rota no main.go
    - Adicionar rota publica (fora do grupo protected)
    - Seguir padrao do endpoint /health existente
  - [ ] 3.5 Garantir que os testes do endpoint passem
    - Executar APENAS os 3-5 testes escritos em 3.1
    - NAO executar suite completa

**Criterios de Aceitacao:**
- Os 3-5 testes escritos em 3.1 passam
- Endpoint retorna JSON com todos os componentes
- Status codes apropriados (200 OK, 503 Service Unavailable)
- Endpoint acessivel sem autenticacao

---

#### Grupo de Tarefas 4: Sistema de Alertas por Email
**Dependencias:** Grupo de Tarefas 2

- [ ] 4.0 Completar sistema de alertas de infraestrutura
  - [ ] 4.1 Escrever 2-3 testes focados para alertas de email
    - Testar envio de email quando Listener cair
    - Testar que cooldown impede emails repetidos
    - Limitar a 3 testes maximos
  - [ ] 4.2 Criar template HTML para alerta de infraestrutura
    - Arquivo: `/backend/internal/services/notification/email.go`
    - Seguir padrao do obitoNotificationTemplate existente
    - Incluir: nome do servico, hora da falha, acoes recomendadas
    - Cores de destaque para urgencia (vermelho)
  - [ ] 4.3 Implementar metodo SendInfrastructureAlert
    - Reutilizar EmailService existente
    - Parametros: servico afetado, status anterior, novo status
    - Enviar para email do admin configurado
  - [ ] 4.4 Integrar alerta com HealthMonitorService
    - Chamar SendInfrastructureAlert na transicao Listener UP -> DOWN
    - Respeitar cooldown de 5 minutos
  - [ ] 4.5 Garantir que os testes de alertas passem
    - Executar APENAS os 2-3 testes escritos em 4.1

**Criterios de Aceitacao:**
- Os 2-3 testes escritos em 4.1 passam
- Email enviado apenas na transicao de estado
- Template HTML segue padrao visual existente
- Cooldown funciona corretamente

---

### Camada Backend - Inicializacao

#### Grupo de Tarefas 5: Integracao com main.go
**Dependencias:** Grupos de Tarefas 1, 2, 3, 4

- [ ] 5.0 Completar integracao do HealthMonitorService
  - [ ] 5.1 Inicializar HealthMonitorService no main.go
    - Seguir padrao de inicializacao do ObitoListener/TriagemMotor
    - Passar dependencias: db, redisClient, emailService
    - Configurar email do admin via variavel de ambiente
  - [ ] 5.2 Registrar Start/Stop no ciclo de vida da aplicacao
    - Chamar Start() junto com outros servicos background
    - Chamar Stop() no graceful shutdown
    - Adicionar log de inicializacao
  - [ ] 5.3 Expor instancia global para handlers
    - Criar SetGlobalHealthMonitor() em handlers/health.go
    - Permitir acesso ao ultimo estado conhecido

**Criterios de Aceitacao:**
- HealthMonitorService inicia com a aplicacao
- Graceful shutdown funciona corretamente
- Logs de inicializacao presentes

---

### Camada Frontend - Componentes

#### Grupo de Tarefas 6: Componentes de Status
**Dependencias:** Grupo de Tarefas 3 (endpoint disponivel)

- [ ] 6.0 Completar componentes de UI para status
  - [ ] 6.1 Escrever 2-4 testes focados para componentes de status
    - Testar renderizacao do ServiceStatusCard com diferentes estados
    - Testar mudanca de cor do indicador conforme status
    - Testar animacao de pulse quando status vermelho
    - Limitar a 4 testes maximos
  - [ ] 6.2 Criar componente ServiceStatusIndicator
    - Arquivo: `/frontend/src/components/dashboard/ServiceStatusIndicator.tsx`
    - Usar StatusBadge.tsx como referencia
    - Props: status ('up' | 'degraded' | 'down'), latency
    - Indicador circular colorido (24px)
    - Verde: < 500ms, Amarelo: 500-2000ms, Vermelho: down/timeout
  - [ ] 6.3 Criar componente ServiceStatusCard
    - Arquivo: `/frontend/src/components/dashboard/ServiceStatusCard.tsx`
    - Usar componentes Card, Badge do shadcn/ui
    - Props: serviceName, icon, status, latency, lastCheck
    - Icone do servico (Database, Server, Activity, Radio, etc)
    - Borda do card muda de cor conforme status
    - Animacao de pulse quando status vermelho
    - Tooltip com detalhes adicionais
  - [ ] 6.4 Criar componente SystemAlertBanner
    - Arquivo: `/frontend/src/components/dashboard/SystemAlertBanner.tsx`
    - Exibir quando algum servico estiver DOWN
    - Estilo de alerta vermelho no topo da pagina
    - Lista de servicos afetados
  - [ ] 6.5 Garantir que os testes dos componentes passem
    - Executar APENAS os 2-4 testes escritos em 6.1

**Criterios de Aceitacao:**
- Os 2-4 testes escritos em 6.1 passam
- Componentes renderizam corretamente
- Cores do semaforo aplicadas conforme regras
- Animacoes funcionam para status criticos

---

#### Grupo de Tarefas 7: Hook de Health Status
**Dependencias:** Grupo de Tarefas 3

- [ ] 7.0 Completar hook para consumo de status
  - [ ] 7.1 Escrever 2-3 testes focados para useHealthStatus
    - Testar fetch inicial de dados
    - Testar polling a cada 10 segundos
    - Limitar a 3 testes maximos
  - [ ] 7.2 Criar hook useHealthStatus
    - Arquivo: `/frontend/src/hooks/useHealthStatus.tsx`
    - Usar useSSE.tsx como referencia de padrao
    - Fazer fetch em /api/health/summary
    - Polling automatico a cada 10 segundos
    - Estados: isLoading, error, data, lastUpdated
    - Funcao para refresh manual
  - [ ] 7.3 Criar tipos TypeScript para resposta
    - Arquivo: `/frontend/src/types/health.ts`
    - Interface HealthSummary
    - Interface ComponentStatus
    - Type ServiceStatus = 'up' | 'degraded' | 'down'
  - [ ] 7.4 Garantir que os testes do hook passem
    - Executar APENAS os 2-3 testes escritos em 7.1

**Criterios de Aceitacao:**
- Os 2-3 testes escritos em 7.1 passam
- Hook busca dados automaticamente
- Polling funciona a cada 10 segundos
- Refresh manual disponivel

---

### Camada Frontend - Pagina

#### Grupo de Tarefas 8: Pagina de Status do Sistema
**Dependencias:** Grupos de Tarefas 6, 7

- [ ] 8.0 Completar pagina de dashboard de status
  - [ ] 8.1 Escrever 2-3 testes focados para a pagina
    - Testar renderizacao dos cards de servico
    - Testar restricao de acesso (apenas admin)
    - Limitar a 3 testes maximos
  - [ ] 8.2 Criar pagina /dashboard/status
    - Arquivo: `/frontend/src/app/dashboard/status/page.tsx`
    - Titulo: "Status do Sistema"
    - Badge de ultima atualizacao no header
    - Botao de refresh manual
    - Grid responsivo (2 colunas mobile, 3 colunas desktop)
  - [ ] 8.3 Implementar grid de ServiceStatusCards
    - Cards para: Database, Redis, Listener, TriagemMotor, SSEHub, API
    - Icones apropriados para cada servico (Lucide icons)
    - Ordenar por prioridade (servicos criticos primeiro)
  - [ ] 8.4 Integrar SystemAlertBanner
    - Exibir banner no topo se algum servico DOWN
    - Listar servicos afetados
    - Link para documentacao de troubleshooting (opcional)
  - [ ] 8.5 Implementar restricao de acesso
    - Verificar role "admin" do usuario logado
    - Redirecionar para /dashboard se nao autorizado
    - Usar middleware/hook de autorizacao existente
  - [ ] 8.6 Garantir que os testes da pagina passem
    - Executar APENAS os 2-3 testes escritos em 8.1

**Criterios de Aceitacao:**
- Os 2-3 testes escritos em 8.1 passam
- Pagina acessivel apenas para admin
- Cards exibem status correto de cada servico
- Atualizacao automatica a cada 10 segundos
- Layout responsivo funciona em mobile/desktop

---

### Integracao SSE (Opcional)

#### Grupo de Tarefas 9: Notificacoes em Tempo Real via SSE
**Dependencias:** Grupos de Tarefas 2, 8

- [ ] 9.0 Completar integracao SSE para status em tempo real (OPCIONAL)
  - [ ] 9.1 Publicar eventos de mudanca de status via SSE Hub
    - Tipo de evento: "system_status_change"
    - Payload: servico, status anterior, novo status, timestamp
    - Publicar apenas para usuarios admin
  - [ ] 9.2 Atualizar useHealthStatus para receber SSE
    - Combinar polling com eventos SSE
    - Atualizar estado imediatamente ao receber evento
    - Manter polling como fallback
  - [ ] 9.3 Exibir notificacao toast na mudanca de status
    - Toast vermelho quando servico cair
    - Toast verde quando servico recuperar

**Criterios de Aceitacao:**
- Eventos SSE publicados corretamente
- Dashboard atualiza instantaneamente
- Toasts exibidos nas transicoes de estado

---

### Testes e Validacao

#### Grupo de Tarefas 10: Revisao de Testes e Analise de Gaps
**Dependencias:** Grupos de Tarefas 1-9

- [ ] 10.0 Revisar testes existentes e preencher gaps criticos
  - [ ] 10.1 Revisar testes dos Grupos de Tarefas anteriores
    - Revisar 2-4 testes do heartbeat (Grupo 1)
    - Revisar 3-5 testes do HealthMonitorService (Grupo 2)
    - Revisar 3-5 testes do endpoint (Grupo 3)
    - Revisar 2-3 testes de alertas (Grupo 4)
    - Revisar 2-4 testes de componentes (Grupo 6)
    - Revisar 2-3 testes do hook (Grupo 7)
    - Revisar 2-3 testes da pagina (Grupo 8)
    - Total esperado: ~16-27 testes
  - [ ] 10.2 Analisar gaps de cobertura APENAS para esta feature
    - Identificar fluxos criticos sem cobertura
    - Focar em integracao ponta-a-ponta
    - NAO avaliar cobertura geral da aplicacao
  - [ ] 10.3 Escrever ate 8 testes adicionais para gaps criticos
    - Testar fluxo completo: servico cai -> alerta enviado
    - Testar fluxo: servico cai -> dashboard atualiza
    - Testar cenario de recuperacao de servico
    - NAO escrever testes exaustivos para todos os cenarios
  - [ ] 10.4 Executar testes especificos desta feature
    - Executar APENAS testes relacionados a esta spec
    - Total esperado: ~20-35 testes
    - NAO executar suite completa da aplicacao
    - Verificar que fluxos criticos passam

**Criterios de Aceitacao:**
- Todos os testes especificos da feature passam
- Fluxos criticos de usuario cobertos
- Maximo de 8 testes adicionais escritos
- Testes focados exclusivamente nesta feature

---

## Ordem de Execucao

Sequencia recomendada de implementacao:

```
1. Grupo 1: Sistema de Heartbeat do Listener
   |
2. Grupo 2: Health Monitor Service (Background)
   |
   +---> 3. Grupo 3: Endpoint Agregado de Health Check
   |
   +---> 4. Grupo 4: Sistema de Alertas por Email
   |
5. Grupo 5: Integracao com main.go
   |
   +---> 6. Grupo 6: Componentes de Status
   |     |
   |     +---> 8. Grupo 8: Pagina de Status do Sistema
   |
   +---> 7. Grupo 7: Hook de Health Status
         |
         +---> 8. Grupo 8: Pagina de Status do Sistema
   |
9. Grupo 9: Integracao SSE (Opcional)
   |
10. Grupo 10: Revisao de Testes e Analise de Gaps
```

### Dependencias Paralelas

Os seguintes grupos podem ser executados em paralelo:
- **Grupos 3 e 4** - ambos dependem apenas do Grupo 2
- **Grupos 6 e 7** - ambos dependem apenas do Grupo 3

---

## Resumo de Arquivos a Criar/Modificar

### Backend (Go)

**Novos Arquivos:**
- `/backend/internal/services/health/monitor.go` - HealthMonitorService

**Arquivos a Modificar:**
- `/backend/internal/services/listener/obito_listener.go` - adicionar heartbeat
- `/backend/internal/handlers/health.go` - adicionar HealthSummary
- `/backend/internal/services/notification/email.go` - adicionar template de alerta
- `/backend/cmd/api/main.go` - registrar HealthMonitorService e rota

### Frontend (Next.js/React)

**Novos Arquivos:**
- `/frontend/src/components/dashboard/ServiceStatusIndicator.tsx`
- `/frontend/src/components/dashboard/ServiceStatusCard.tsx`
- `/frontend/src/components/dashboard/SystemAlertBanner.tsx`
- `/frontend/src/hooks/useHealthStatus.tsx`
- `/frontend/src/types/health.ts`
- `/frontend/src/app/dashboard/status/page.tsx`

### Testes

**Backend:**
- `/backend/internal/services/health/monitor_test.go`
- `/backend/internal/handlers/health_test.go` (estender)

**Frontend:**
- `/frontend/src/components/dashboard/ServiceStatusCard.test.tsx`
- `/frontend/src/hooks/useHealthStatus.test.tsx`
- `/frontend/src/app/dashboard/status/page.test.tsx`

---

## Notas Tecnicas

### Chaves Redis Utilizadas
- `vitalconnect:listener:heartbeat` - heartbeat do Listener (TTL 15s)
- `vitalconnect:health:last_states` - ultimo estado conhecido de cada servico
- `vitalconnect:health:alert_cooldowns` - timestamps de ultimos alertas

### Thresholds de Status
- **Up (Verde):** latencia < 500ms
- **Degraded (Amarelo):** latencia 500ms - 2000ms
- **Down (Vermelho):** timeout (> 2000ms) ou servico nao responde

### Configuracoes de Ambiente
- `ADMIN_ALERT_EMAIL` - email para receber alertas de infraestrutura
- `HEALTH_CHECK_INTERVAL` - intervalo de verificacao (default: 10s)
- `ALERT_COOLDOWN_MINUTES` - cooldown entre alertas (default: 5min)
