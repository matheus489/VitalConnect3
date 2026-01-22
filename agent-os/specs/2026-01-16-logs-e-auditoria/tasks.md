# Task Breakdown: Logs e Auditoria

## Visao Geral
Total de Tarefas: 4 Grupos de Tarefas | ~32 Sub-tarefas

## Lista de Tarefas

### Camada de Banco de Dados

#### Grupo de Tarefas 1: Modelo de Dados e Migracao
**Dependencias:** Nenhuma

- [ ] 1.0 Completar camada de banco de dados
  - [ ] 1.1 Escrever 4-6 testes focados para funcionalidade do modelo AuditLog
    - Limitar a 4-6 testes altamente focados
    - Testar apenas comportamentos criticos: validacao de severity, criacao de log, associacoes
    - Pular cobertura exaustiva de todos os metodos e casos de borda
  - [ ] 1.2 Criar modelo AuditLog em `/backend/internal/models/audit_log.go`
    - Campos: id (UUID), timestamp, usuario_id (nullable), actor_name, acao, entidade_tipo, entidade_id, hospital_id, severity, detalhes (JSONB), ip_address, user_agent
    - Enum Severity: INFO, WARN, CRITICAL
    - Structs: AuditLog, AuditLogResponse, AuditLogFilter
    - Seguir padrao de `/backend/internal/models/occurrence_history.go`
  - [ ] 1.3 Criar migracao `/backend/migrations/008_create_audit_logs.sql`
    - Tabela `audit_logs` com todos os campos especificados
    - Constraint de imutabilidade: REVOKE UPDATE, DELETE na tabela
    - Indices: timestamp DESC, usuario_id, entidade_tipo, entidade_id, severity, hospital_id
    - Indice composto: (hospital_id, timestamp DESC) para queries de Gestor
    - Comentario sobre particionamento por data para volume de 5 anos (implementacao futura)
  - [ ] 1.4 Criar repository `/backend/internal/repository/audit_log_repository.go`
    - Metodo Create(ctx, auditLog) para inserir log
    - Metodo List(ctx, filters, pagination) com filtros: data_inicio, data_fim, usuario_id, acao, entidade_tipo, entidade_id, severity, hospital_id
    - Metodo GetByEntityID(ctx, entidadeTipo, entidadeID) para timeline de ocorrencia
    - Seguir padrao de `/backend/internal/repository/occurrence_history_repository.go`
  - [ ] 1.5 Garantir que testes da camada de banco passem
    - Executar APENAS os 4-6 testes escritos em 1.1
    - Verificar que migracoes rodam com sucesso
    - NAO executar suite de testes completa neste estagio

**Criterios de Aceite:**
- Os 4-6 testes escritos em 1.1 passam
- Modelo passa nos testes de validacao
- Migracao executa com sucesso
- Repository insere e consulta logs corretamente
- Constraints de imutabilidade funcionam (UPDATE/DELETE bloqueados)

---

### Camada de Servico

#### Grupo de Tarefas 2: Servico de Auditoria Reutilizavel
**Dependencias:** Grupo de Tarefas 1

- [ ] 2.0 Completar servico de auditoria
  - [ ] 2.1 Escrever 4-6 testes focados para o servico de auditoria
    - Limitar a 4-6 testes altamente focados
    - Testar: criacao de log com usuario, criacao de log do sistema ("SIDOT Bot"), extracao de IP/User-Agent do contexto
    - Pular testes exaustivos de todas as combinacoes
  - [ ] 2.2 Criar servico `/backend/internal/services/audit/audit_service.go`
    - Interface AuditService com metodo LogEvent
    - Assinatura: LogEvent(ctx, acao, entidadeTipo, entidadeID, hospitalID, severity, detalhes)
    - Para usuario autenticado: extrair userID e nome do contexto
    - Para acoes do sistema: usuario_id = nil, actor_name = "SIDOT Bot"
    - Extrair ip_address e user_agent do contexto HTTP
  - [ ] 2.3 Integrar servico nos handlers de Autenticacao
    - Em `/backend/internal/handlers/auth.go`
    - Eventos: auth.login (INFO), auth.logout (INFO), auth.login_failed (WARN)
    - Capturar IP e User-Agent da requisicao
  - [ ] 2.4 Integrar servico nos handlers de Regras de Triagem
    - Em `/backend/internal/handlers/triagem.go`
    - Eventos: regra.create (INFO), regra.update (CRITICAL), regra.delete (CRITICAL)
    - Incluir hospital_id da regra
  - [ ] 2.5 Integrar servico nos handlers de Ocorrencias
    - Em `/backend/internal/handlers/occurrences.go`
    - Eventos: ocorrencia.visualizar (INFO), ocorrencia.aceitar (INFO), ocorrencia.recusar (INFO), ocorrencia.status_change (INFO)
    - Incluir hospital_id da ocorrencia
  - [ ] 2.6 Integrar servico nos handlers de Usuarios
    - Em `/backend/internal/handlers/users.go`
    - Eventos: usuario.create (INFO), usuario.update (INFO), usuario.desativar (WARN)
  - [ ] 2.7 Adicionar log de triagem automatica (sistema)
    - Onde triagem automatica ocorre, chamar servico com actor_name = "SIDOT Bot"
    - Evento: triagem.rejeicao (INFO)
  - [ ] 2.8 Garantir que testes do servico passem
    - Executar APENAS os 4-6 testes escritos em 2.1
    - Verificar integracoes manuais nos handlers
    - NAO executar suite de testes completa neste estagio

**Criterios de Aceite:**
- Os 4-6 testes escritos em 2.1 passam
- Servico registra eventos corretamente com todos os campos
- Acoes de usuario incluem usuario_id e actor_name do usuario
- Acoes do sistema usam "SIDOT Bot" como actor_name
- IP e User-Agent sao capturados corretamente

---

### Camada de API

#### Grupo de Tarefas 3: Endpoint de API REST
**Dependencias:** Grupo de Tarefas 2

- [ ] 3.0 Completar camada de API
  - [ ] 3.1 Escrever 4-6 testes focados para endpoint de API
    - Limitar a 4-6 testes altamente focados
    - Testar: listagem paginada, filtros basicos, controle de acesso (Admin vs Gestor vs Operador negado)
    - Pular testes exaustivos de todas as combinacoes de filtros
  - [ ] 3.2 Criar handler `/backend/internal/handlers/audit_logs.go`
    - GET /api/v1/audit-logs com paginacao (padrao PaginatedResponse)
    - Query params: data_inicio, data_fim, usuario_id, acao, entidade_tipo, entidade_id, severity, hospital_id
    - Ordenacao: timestamp DESC (padrao)
    - Seguir padrao de `/backend/internal/handlers/occurrences.go`
  - [ ] 3.3 Implementar controle de acesso no handler
    - Admin: acesso a todos os logs
    - Gestor: acesso apenas a logs com hospital_id do seu hospital
    - Operador: acesso NEGADO (retornar 403 Forbidden)
    - Usar middleware de autenticacao existente
  - [ ] 3.4 Criar endpoint para timeline de ocorrencia
    - GET /api/v1/occurrences/:id/timeline
    - Retorna logs onde entidade_tipo = "Ocorrencia" e entidade_id = :id
    - Acessivel por Admin, Gestor (mesmo hospital) e Operador (apenas suas ocorrencias)
    - Ordenacao: timestamp ASC (cronologico)
  - [ ] 3.5 Registrar rotas em `/backend/cmd/server/main.go` ou router
    - Adicionar rota GET /api/v1/audit-logs
    - Adicionar rota GET /api/v1/occurrences/:id/timeline
    - Aplicar middlewares de autenticacao
  - [ ] 3.6 Garantir que testes da API passem
    - Executar APENAS os 4-6 testes escritos em 3.1
    - Verificar respostas da API manualmente
    - NAO executar suite de testes completa neste estagio

**Criterios de Aceite:**
- Os 4-6 testes escritos em 3.1 passam
- Endpoint retorna dados paginados corretamente
- Filtros funcionam individualmente e combinados
- Controle de acesso por role funciona corretamente
- Timeline de ocorrencia retorna eventos ordenados cronologicamente

---

### Componentes de Frontend

#### Grupo de Tarefas 4: Interface de Usuario
**Dependencias:** Grupo de Tarefas 3

- [ ] 4.0 Completar componentes de UI
  - [ ] 4.1 Escrever 4-6 testes focados para componentes de UI
    - Limitar a 4-6 testes altamente focados
    - Testar: renderizacao da tabela de logs, filtragem basica, exibicao do SeverityBadge
    - Pular testes exaustivos de todos os estados e interacoes
  - [ ] 4.2 Criar tipo TypeScript para AuditLog em `/frontend/src/types/`
    - Interface AuditLog com todos os campos
    - Enum Severity: INFO, WARN, CRITICAL
    - Interface AuditLogFilters para estado dos filtros
  - [ ] 4.3 Criar componente SeverityBadge em `/frontend/src/components/dashboard/SeverityBadge.tsx`
    - Seguir padrao de `/frontend/src/components/dashboard/StatusBadge.tsx`
    - Mapeamento: INFO = cinza, WARN = amarelo, CRITICAL = vermelho
    - Usar componente Badge do shadcn/ui
  - [ ] 4.4 Criar componente AuditLogFilters em `/frontend/src/components/audit/AuditLogFilters.tsx`
    - Filtros: periodo (date range), usuario (Select), tipo de acao (Select), severidade (Select)
    - Botao para limpar filtros
    - Seguir padrao de `/frontend/src/components/dashboard/OccurrenceFilters.tsx`
  - [ ] 4.5 Criar componente AuditLogsTable em `/frontend/src/components/audit/AuditLogsTable.tsx`
    - Colunas: Data/Hora, Usuario, Acao, Entidade, Severidade
    - Usar Table, TableHeader, TableBody do shadcn/ui
    - Incluir SeverityBadge para coluna de severidade
    - Seguir padrao de `/frontend/src/components/dashboard/OccurrencesTable.tsx`
  - [ ] 4.6 Criar pagina /logs em `/frontend/src/app/logs/page.tsx`
    - Layout com filtros no topo e tabela abaixo
    - Integracao com API GET /api/v1/audit-logs
    - Paginacao usando componente Pagination existente
    - Controle de acesso: redirecionar Operador para dashboard
    - Loading state e empty state
  - [ ] 4.7 Criar componente OccurrenceTimeline em `/frontend/src/components/dashboard/OccurrenceTimeline.tsx`
    - Timeline vertical cronologica
    - Mostrar: horario, acao, nome do usuario (ou "Sistema"), observacoes
    - Consumir API GET /api/v1/occurrences/:id/timeline
    - Estilo visual de timeline com linha conectando eventos
  - [ ] 4.8 Integrar OccurrenceTimeline no OccurrenceDetailModal
    - Adicionar aba ou secao "Historico" no modal existente
    - Em `/frontend/src/components/dashboard/OccurrenceDetailModal.tsx`
    - Carregar timeline quando modal abrir
  - [ ] 4.9 Adicionar link para /logs no menu de navegacao
    - Visivel apenas para Admin e Gestor
    - Icone e label apropriados
  - [ ] 4.10 Garantir que testes de UI passem
    - Executar APENAS os 4-6 testes escritos em 4.1
    - Verificar comportamentos criticos dos componentes
    - NAO executar suite de testes completa neste estagio

**Criterios de Aceite:**
- Os 4-6 testes escritos em 4.1 passam
- Pagina /logs renderiza corretamente para Admin e Gestor
- Operador e redirecionado ao tentar acessar /logs
- Filtros atualizam a listagem corretamente
- Paginacao funciona
- SeverityBadge exibe cores corretas
- Timeline aparece no modal de detalhes da ocorrencia
- Timeline exibe eventos em ordem cronologica

---

### Testes e Validacao

#### Grupo de Tarefas 5: Revisao de Testes e Analise de Gaps
**Dependencias:** Grupos de Tarefas 1-4

- [ ] 5.0 Revisar testes existentes e preencher gaps criticos
  - [ ] 5.1 Revisar testes dos Grupos de Tarefas 1-4
    - Revisar os 4-6 testes escritos pelo grupo de banco (Tarefa 1.1)
    - Revisar os 4-6 testes escritos pelo grupo de servico (Tarefa 2.1)
    - Revisar os 4-6 testes escritos pelo grupo de API (Tarefa 3.1)
    - Revisar os 4-6 testes escritos pelo grupo de UI (Tarefa 4.1)
    - Total de testes existentes: aproximadamente 16-24 testes
  - [ ] 5.2 Analisar gaps de cobertura de teste para ESTA feature apenas
    - Identificar fluxos criticos de usuario sem cobertura de teste
    - Focar APENAS em gaps relacionados aos requisitos desta spec
    - NAO avaliar cobertura de toda a aplicacao
    - Priorizar fluxos end-to-end sobre gaps de testes unitarios
  - [ ] 5.3 Escrever ate 10 testes adicionais estrategicos no maximo
    - Adicionar maximo de 10 novos testes para preencher gaps criticos identificados
    - Focar em pontos de integracao e fluxos end-to-end
    - NAO escrever cobertura abrangente para todos os cenarios
    - Pular casos de borda, testes de performance e testes de acessibilidade a menos que sejam criticos para o negocio
  - [ ] 5.4 Executar testes especificos da feature apenas
    - Executar APENAS testes relacionados a esta spec (testes de 1.1, 2.1, 3.1, 4.1 e 5.3)
    - Total esperado: aproximadamente 26-34 testes no maximo
    - NAO executar suite de testes completa da aplicacao
    - Verificar que fluxos criticos passam

**Criterios de Aceite:**
- Todos os testes especificos da feature passam (aproximadamente 26-34 testes no total)
- Fluxos criticos de usuario para esta feature estao cobertos
- No maximo 10 testes adicionais foram adicionados ao preencher gaps
- Testes focados exclusivamente nos requisitos desta spec

---

## Ordem de Execucao

Sequencia de implementacao recomendada:

1. **Camada de Banco de Dados** (Grupo de Tarefas 1)
   - Fundacao para todo o sistema de auditoria
   - Modelo, migracao e repository

2. **Camada de Servico** (Grupo de Tarefas 2)
   - Servico reutilizavel para registrar eventos
   - Integracoes com handlers existentes

3. **Camada de API** (Grupo de Tarefas 3)
   - Endpoints REST para consulta de logs
   - Controle de acesso por role

4. **Componentes de Frontend** (Grupo de Tarefas 4)
   - Tela de logs para Admin/Gestor
   - Timeline de eventos para Operador

5. **Revisao de Testes** (Grupo de Tarefas 5)
   - Revisao e validacao final
   - Preenchimento de gaps criticos

---

## Notas Tecnicas

### Padroes a Seguir
- **Backend:** Seguir padroes existentes em `/backend/internal/` para handlers, repository e models
- **Frontend:** Seguir padroes de `/frontend/src/components/dashboard/` para componentes
- **API:** Usar formato PaginatedResponse existente
- **Testes:** Manter testes focados e minimos conforme orientacao

### Consideracoes de Seguranca e LGPD
- Nunca armazenar dados sensiveis de pacientes no campo `detalhes`
- Usar apenas IDs de referencia para entidades
- Validar permissoes antes de retornar dados na API
- Controle de acesso rigido por role e hospital_id

### Fora do Escopo (Confirmar antes de implementar)
- Integracao com Loki (Fase 3)
- Dashboard com graficos
- Alertas de anomalia/SIEM
- Exportacao PDF/CSV
- Particionamento automatico de tabela
