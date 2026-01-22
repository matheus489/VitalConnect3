# Detalhamento de Tarefas: Gestao de Plantoes (Shift Management)

## Visao Geral
Total de Tarefas: 6 Grupos de Tarefas

## Lista de Tarefas

### Camada de Banco de Dados

#### Grupo de Tarefas 1: Modelo de Dados e Migrations
**Dependencias:** Nenhuma

- [ ] 1.0 Completar camada de banco de dados
  - [ ] 1.1 Escrever 2-8 testes focados para o modelo Shift
    - Limitar a 2-8 testes altamente focados no maximo
    - Testar apenas comportamentos criticos do modelo (validacao de horarios, associacoes com User e Hospital, logica de turno noturno)
    - Pular cobertura exaustiva de todos os metodos e casos extremos
  - [ ] 1.2 Criar modelo Shift em `models/shift.go`
    - Campos: id (UUID), hospital_id (FK), user_id (FK), day_of_week (INTEGER 0-6), start_time (TIME), end_time (TIME), created_at, updated_at
    - Validacoes: day_of_week entre 0-6, start_time e end_time validos
    - Tags: json, db, validate seguindo padrao de `models/user.go`
    - Implementar metodo IsNightShift() para detectar turnos que cruzam meia-noite
  - [ ] 1.3 Criar migration `XXX_create_shifts.sql`
    - Tabela shifts com todos os campos especificados
    - Indices em: hospital_id, user_id, day_of_week
    - Constraint UNIQUE em (hospital_id, user_id, day_of_week, start_time)
    - Foreign keys para hospitals e users
    - Incluir COMMENT ON TABLE e COMMENT ON COLUMN
    - Seguir padrao de `migrations/002_create_users.sql`
  - [ ] 1.4 Configurar associacoes
    - Shift belongs_to Hospital
    - Shift belongs_to User
    - User has_many Shifts
    - Hospital has_many Shifts
  - [ ] 1.5 Garantir que testes da camada de banco passem
    - Executar APENAS os 2-8 testes escritos em 1.1
    - Verificar que migrations rodam com sucesso
    - NAO executar suite de testes completa nesta etapa

**Criterios de Aceitacao:**
- Os 2-8 testes escritos em 1.1 passam
- Modelo passa nos testes de validacao
- Migrations executam com sucesso
- Associacoes funcionam corretamente

---

### Camada de Repositorio

#### Grupo de Tarefas 2: Repository Pattern para Shifts
**Dependencias:** Grupo de Tarefas 1

- [ ] 2.0 Completar camada de repositorio
  - [ ] 2.1 Escrever 2-8 testes focados para ShiftRepository
    - Limitar a 2-8 testes altamente focados no maximo
    - Testar apenas operacoes criticas (Create, GetByID, ListByHospital, GetActiveShifts)
    - Pular testes exaustivos de todas as queries e cenarios
  - [ ] 2.2 Criar `repository/shift_repository.go`
    - Seguir padrao de `repository/user_repository.go`
    - Metodos: Create, GetByID, Update, Delete, ListByHospitalID
    - Usar parameterized queries para prevenir SQL injection
    - Tratar sql.ErrNoRows retornando erro especifico
  - [ ] 2.3 Implementar queries especializadas
    - GetActiveShifts(hospitalID, dayOfWeek, currentTime) - retorna operadores de plantao no momento
    - Query com logica especial para turnos que cruzam meia-noite (start_time > end_time)
    - GetShiftsByUserID(userID) - retorna escalas do usuario
    - GetTodayShifts(hospitalID) - retorna escalados do dia atual
  - [ ] 2.4 Implementar query de cobertura
    - GetCoverageGaps(hospitalID) - analisa escala e retorna horarios descobertos por dia da semana
    - Considerar gaps como intervalos sem nenhum operador escalado
  - [ ] 2.5 Garantir que testes do repositorio passem
    - Executar APENAS os 2-8 testes escritos em 2.1
    - Verificar operacoes CRUD funcionam
    - NAO executar suite de testes completa nesta etapa

**Criterios de Aceitacao:**
- Os 2-8 testes escritos em 2.1 passam
- Todas as operacoes CRUD funcionam
- Queries especializadas retornam dados corretos
- Logica de turno noturno funciona corretamente

---

### Camada de Servicos

#### Grupo de Tarefas 3: Servicos de Roteamento e Cache
**Dependencias:** Grupo de Tarefas 2

- [ ] 3.0 Completar camada de servicos
  - [ ] 3.1 Escrever 2-8 testes focados para ShiftRoutingService
    - Limitar a 2-8 testes altamente focados no maximo
    - Testar apenas fluxos criticos (roteamento normal, fallback para gestores, cache hit/miss)
    - Pular testes exaustivos de todos os cenarios
  - [ ] 3.2 Criar `services/shift_routing_service.go`
    - Metodo GetOnDutyOperators(hospitalID, eventTimestamp) - determina operadores de plantao baseado no timestamp do evento
    - Implementar fallback obrigatorio: se nenhum operador escalado, retornar todos Gestores do hospital
    - Usar timestamp do obito (nao da notificacao) para determinar turno
  - [ ] 3.3 Implementar cache Redis
    - Cachear escala atual por hospital_id com TTL de 5 minutos
    - Key pattern: `shift:hospital:{hospital_id}:current`
    - Fallback para database se cache miss
    - Criar metodos auxiliares: GetFromCache, SetCache, InvalidateCache
  - [ ] 3.4 Implementar invalidacao de cache
    - Invalidar cache ao criar/atualizar/deletar escalas
    - Conectar com operacoes do repository
  - [ ] 3.5 Integrar com NotificationService existente
    - Metodo RouteNotification(hospitalID, eventTimestamp, notificationPayload)
    - Obter operadores de plantao e disparar broadcast
  - [ ] 3.6 Garantir que testes de servicos passem
    - Executar APENAS os 2-8 testes escritos em 3.1
    - Verificar roteamento e cache funcionam
    - NAO executar suite de testes completa nesta etapa

**Criterios de Aceitacao:**
- Os 2-8 testes escritos em 3.1 passam
- Roteamento retorna operadores corretos baseado no horario
- Fallback para gestores funciona quando nao ha escala
- Cache Redis funciona com invalidacao correta

---

### Camada de API

#### Grupo de Tarefas 4: Endpoints REST
**Dependencias:** Grupo de Tarefas 3

- [ ] 4.0 Completar camada de API
  - [ ] 4.1 Escrever 2-8 testes focados para endpoints de API
    - Limitar a 2-8 testes altamente focados no maximo
    - Testar apenas acoes criticas (listar escalas, criar escala, validacao de permissoes)
    - Pular testes exaustivos de todas as acoes e cenarios
  - [ ] 4.2 Criar `handlers/shifts.go`
    - Seguir estrutura de handlers com gin.Context
    - Usar middleware.GetUserClaims() para autorizacao
    - Retornar responses com gin.H{"data": ..., "total": ...}
    - Tratar erros com codigos HTTP apropriados (400, 401, 403, 404, 500)
  - [ ] 4.3 Implementar endpoints de gestao (Gestor/Admin)
    - GET /api/v1/hospitals/:hospital_id/shifts - listar escalas do hospital
    - POST /api/v1/hospitals/:hospital_id/shifts - criar escala
    - PATCH /api/v1/shifts/:id - atualizar escala
    - DELETE /api/v1/shifts/:id - remover escala
  - [ ] 4.4 Implementar endpoints de visualizacao (Operador)
    - GET /api/v1/shifts/my-schedule - listar proprias escalas
    - GET /api/v1/shifts/today - listar escalados do dia atual
  - [ ] 4.5 Implementar endpoint de cobertura
    - GET /api/v1/hospitals/:hospital_id/shifts/coverage - analisar cobertura e retornar gaps
  - [ ] 4.6 Implementar middleware de permissoes
    - Admin: acesso total a todas as escalas de todos os hospitais
    - Gestor: CRUD de escalas apenas do proprio hospital
    - Operador: somente leitura (visualiza propria escala e escala do dia)
    - Validar hospital_id do usuario no middleware para gestor/operador
  - [ ] 4.7 Adicionar metodos de permissao no modelo User
    - CanManageShifts() - retorna true para Admin e Gestor
    - CanViewShifts() - retorna true para todos os roles autenticados
  - [ ] 4.8 Registrar rotas no router principal
    - Adicionar grupo de rotas /shifts com middlewares apropriados
  - [ ] 4.9 Garantir que testes de API passem
    - Executar APENAS os 2-8 testes escritos em 4.1
    - Verificar operacoes CRUD via API funcionam
    - NAO executar suite de testes completa nesta etapa

**Criterios de Aceitacao:**
- Os 2-8 testes escritos em 4.1 passam
- Todas as operacoes CRUD funcionam via API
- Autorizacao por role funciona corretamente
- Respostas seguem formato consistente

---

### Camada de Frontend

#### Grupo de Tarefas 5: Componentes de UI
**Dependencias:** Grupo de Tarefas 4

- [ ] 5.0 Completar componentes de UI
  - [ ] 5.1 Escrever 2-8 testes focados para componentes de UI
    - Limitar a 2-8 testes altamente focados no maximo
    - Testar apenas comportamentos criticos (renderizacao do grid, submissao do formulario, exibicao de gaps)
    - Pular testes exaustivos de todos os estados e interacoes
  - [ ] 5.2 Criar hooks customizados para data fetching
    - useShifts(hospitalID) - lista escalas do hospital
    - useMySchedule() - lista escalas do usuario logado
    - useTodayShifts() - lista escalados do dia
    - useCoverage(hospitalID) - retorna analise de cobertura com gaps
  - [ ] 5.3 Criar pagina de Gestao de Escalas
    - Rota: /dashboard/shifts
    - Acessivel apenas para Gestores e Admins
    - Layout com espacamento consistente (space-y-6)
    - Seguir padrao de `frontend/src/app/dashboard/page.tsx`
  - [ ] 5.4 Criar componente ShiftScheduleGrid
    - Visualizacao semanal em formato de grade de horarios
    - Colunas: dias da semana (Dom-Sab)
    - Linhas: slots de horario
    - Celulas mostram operadores escalados
    - Usar componentes shadcn/ui (Table, Badge, Card)
  - [ ] 5.5 Criar componente ShiftForm
    - Formulario de criacao/edicao de escala
    - Campos: operador (Select), dia da semana, horario inicio, horario fim
    - Botoes de atalho "Diurno (07:00-19:00)" e "Noturno (19:00-07:00)"
    - Select de operadores filtrado por hospital (operadores ativos vinculados)
    - Validacao client-side
  - [ ] 5.6 Implementar indicador de gaps
    - Badge vermelho em celulas com gaps de cobertura
    - Tooltip explicando o horario descoberto
    - Alerta visual no topo da pagina se houver gaps
  - [ ] 5.7 Criar componente MyShiftCard
    - Card "Meu Plantao" para sidebar/dashboard do operador
    - Exibir proximos plantoes do usuario logado
    - Mostrar escala geral do dia atual
    - Destaque visual para o turno ativo
  - [ ] 5.8 Adicionar navegacao
    - Link para /dashboard/shifts no menu lateral (visivel apenas para Gestor/Admin)
    - Aba/card "Meu Plantao" no dashboard do operador
  - [ ] 5.9 Garantir que testes de UI passem
    - Executar APENAS os 2-8 testes escritos em 5.1
    - Verificar componentes renderizam corretamente
    - NAO executar suite de testes completa nesta etapa

**Criterios de Aceitacao:**
- Os 2-8 testes escritos em 5.1 passam
- Componentes renderizam corretamente
- Formularios validam e submetem
- Design segue padroes visuais do SIDOT (shadcn/ui)

---

### Testes e Validacao

#### Grupo de Tarefas 6: Revisao de Testes e Analise de Gaps
**Dependencias:** Grupos de Tarefas 1-5

- [ ] 6.0 Revisar testes existentes e preencher gaps criticos apenas
  - [ ] 6.1 Revisar testes dos Grupos de Tarefas 1-5
    - Revisar os 2-8 testes escritos pelo Grupo 1 (banco de dados)
    - Revisar os 2-8 testes escritos pelo Grupo 2 (repositorio)
    - Revisar os 2-8 testes escritos pelo Grupo 3 (servicos)
    - Revisar os 2-8 testes escritos pelo Grupo 4 (API)
    - Revisar os 2-8 testes escritos pelo Grupo 5 (UI)
    - Total de testes existentes: aproximadamente 10-40 testes
  - [ ] 6.2 Analisar gaps de cobertura APENAS para esta feature
    - Identificar fluxos criticos de usuario sem cobertura de teste
    - Focar APENAS em gaps relacionados aos requisitos desta spec
    - NAO avaliar cobertura de testes da aplicacao inteira
    - Priorizar fluxos end-to-end sobre gaps de testes unitarios
  - [ ] 6.3 Escrever ate 10 testes estrategicos adicionais no maximo
    - Adicionar maximo de 10 novos testes para preencher gaps criticos identificados
    - Focar em pontos de integracao e fluxos end-to-end
    - NAO escrever cobertura completa para todos os cenarios
    - Pular casos extremos, testes de performance e acessibilidade exceto se criticos para o negocio
    - Exemplos de testes adicionais importantes:
      - Fluxo completo: criar escala -> roteamento de notificacao -> operador recebe
      - Fallback: hospital sem escala -> gestores recebem notificacao
      - Turno noturno: evento as 02:00 -> operador do turno 19:00-07:00 recebe
  - [ ] 6.4 Executar apenas testes especificos da feature
    - Executar APENAS testes relacionados a esta feature (testes de 1.1, 2.1, 3.1, 4.1, 5.1 e 6.3)
    - Total esperado: aproximadamente 20-50 testes no maximo
    - NAO executar suite de testes completa da aplicacao
    - Verificar que fluxos criticos passam

**Criterios de Aceitacao:**
- Todos os testes especificos da feature passam (aproximadamente 20-50 testes no total)
- Fluxos criticos de usuario para esta feature estao cobertos
- Nao mais que 10 testes adicionais ao preencher gaps de testes
- Testes focados exclusivamente nos requisitos desta spec

---

## Ordem de Execucao

Sequencia recomendada de implementacao:

1. **Camada de Banco de Dados** (Grupo de Tarefas 1)
   - Fundacao para todas as outras camadas
   - Nenhuma dependencia

2. **Camada de Repositorio** (Grupo de Tarefas 2)
   - Depende do modelo e migrations
   - Queries especializadas para roteamento

3. **Camada de Servicos** (Grupo de Tarefas 3)
   - Logica de roteamento e cache
   - Integracao com sistema de notificacoes

4. **Camada de API** (Grupo de Tarefas 4)
   - Endpoints REST para CRUD e consultas
   - Middlewares de autorizacao

5. **Camada de Frontend** (Grupo de Tarefas 5)
   - Componentes de UI e paginas
   - Integracao com API

6. **Revisao de Testes** (Grupo de Tarefas 6)
   - Validacao final de cobertura
   - Testes de integracao end-to-end

---

## Notas Tecnicas

### Logica de Turno Noturno
Turnos que cruzam meia-noite (ex: 19:00-07:00) requerem logica especial:
```sql
-- Para turno noturno (start_time > end_time):
WHERE (current_time >= start_time OR current_time < end_time)

-- Para turno diurno (start_time < end_time):
WHERE current_time >= start_time AND current_time < end_time
```

### Padrao de Cache Redis
```
Key: shift:hospital:{hospital_id}:current
TTL: 5 minutos
Valor: JSON com lista de operadores do turno atual
```

### Estrutura de Diretorio Backend (Go)
```
models/shift.go
repository/shift_repository.go
services/shift_routing_service.go
handlers/shifts.go
migrations/XXX_create_shifts.sql
```

### Estrutura de Diretorio Frontend (Next.js)
```
frontend/src/app/dashboard/shifts/page.tsx
frontend/src/components/shifts/ShiftScheduleGrid.tsx
frontend/src/components/shifts/ShiftForm.tsx
frontend/src/components/shifts/MyShiftCard.tsx
frontend/src/hooks/useShifts.ts
```
