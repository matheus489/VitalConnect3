# Specification: Gestao de Plantoes (Shift Management)

## Goal

Implementar sistema de gerenciamento de escalas de plantao para o SIDOT, permitindo que hospitais configurem turnos de trabalho e direcionem notificacoes de obitos elegiveis para a equipe correta baseado no horario do evento.

## User Stories

- Como Gestor, quero criar e gerenciar escalas de plantao com horarios customizaveis para que as notificacoes sejam enviadas para a equipe correta de cada turno
- Como Operador, quero visualizar "Meu Plantao" para saber quando estou escalado e ver a escala geral do dia

## Specific Requirements

**Modelo de Dados - Tabela shifts**
- Criar tabela `shifts` com campos: id (UUID), hospital_id (FK), user_id (FK), day_of_week (INTEGER 0-6), start_time (TIME), end_time (TIME)
- Suportar turnos que cruzam meia-noite (ex: 19:00-07:00) com logica de comparacao especial
- Incluir campos created_at, updated_at seguindo padrao existente
- Criar indices em hospital_id, user_id e day_of_week para queries performaticas
- Constraint UNIQUE em (hospital_id, user_id, day_of_week, start_time) para evitar duplicatas

**API REST - CRUD de Escalas**
- GET /api/v1/hospitals/:hospital_id/shifts - listar escalas do hospital (gestor/admin)
- POST /api/v1/hospitals/:hospital_id/shifts - criar escala (gestor/admin)
- PATCH /api/v1/shifts/:id - atualizar escala (gestor/admin)
- DELETE /api/v1/shifts/:id - remover escala (gestor/admin)
- GET /api/v1/shifts/my-schedule - listar proprias escalas (operador)
- GET /api/v1/shifts/today - listar escalados do dia atual (todos autenticados)

**Servico de Roteamento de Notificacoes**
- Criar ShiftRoutingService que determina operadores de plantao baseado em hospital_id e timestamp do evento
- Query deve considerar turnos que cruzam meia-noite corretamente
- Implementar fallback obrigatorio: se nenhum operador escalado, retornar todos Gestores do hospital
- Usar timestamp do obito (nao da notificacao) para determinar turno
- Integrar com NotificationService existente para broadcast

**Cache Redis para Lookups Rapidos**
- Cachear escala atual por hospital_id com TTL de 5 minutos
- Invalidar cache ao criar/atualizar/deletar escalas
- Key pattern: `shift:hospital:{hospital_id}:current`
- Fallback para database se cache miss

**Deteccao de Gaps na Escala**
- Criar endpoint GET /api/v1/hospitals/:hospital_id/shifts/coverage para analisar cobertura
- Retornar lista de horarios descobertos por dia da semana
- Considerar gaps como intervalos sem nenhum operador escalado

**Permissoes por Role**
- Admin: acesso total a todas as escalas de todos os hospitais
- Gestor: CRUD de escalas apenas do proprio hospital
- Operador: somente leitura (visualiza propria escala e escala do dia)
- Validar hospital_id do usuario no middleware para gestor/operador

**Frontend - Pagina de Gestao de Escalas**
- Criar rota /dashboard/shifts para gestores/admins
- Componente ShiftScheduleGrid com visualizacao semanal (grade de horarios)
- Formulario de criacao com atalhos "Diurno (07:00-19:00)" e "Noturno (19:00-07:00)"
- Select de operadores filtrado por hospital (operadores ativos vinculados)
- Indicador visual (badge vermelho) em celulas com gaps de cobertura

**Frontend - Componente Meu Plantao**
- Adicionar aba/card "Meu Plantao" na sidebar ou dashboard do operador
- Exibir proximos plantoes do usuario logado
- Mostrar escala geral do dia atual com destaque para o turno ativo

## Visual Design

Nenhum asset visual foi fornecido. Seguir padroes visuais existentes do sistema (shadcn/ui, cores do SIDOT).

## Existing Code to Leverage

**models/user.go - Modelo de Usuario com Roles**
- Reutilizar UserRole enum (operador, gestor, admin) para validacoes de permissao
- Seguir padrao de struct com tags json/db/validate
- Implementar metodos CanManageShifts() e CanViewShifts() no modelo User

**handlers/users.go - Padrao de Handlers**
- Seguir estrutura de handlers com gin.Context
- Usar middleware.GetUserClaims() para autorizacao
- Retornar responses com gin.H{"data": ..., "total": ...}
- Tratar erros com codigos HTTP apropriados (400, 401, 403, 404, 500)

**repository/user_repository.go - Padrao de Repository**
- Seguir padrao de repository com *sql.DB
- Usar parameterized queries para prevenir SQL injection
- Implementar metodos List, GetByID, Create, Update, Delete
- Tratar sql.ErrNoRows retornando erro especifico

**migrations/002_create_users.sql - Padrao de Migrations**
- Seguir formato com comentarios UP/DOWN
- Criar indices para colunas de lookup frequente
- Adicionar COMMENT ON TABLE e COMMENT ON COLUMN

**frontend/src/app/dashboard/page.tsx - Padrao de Pagina**
- Usar hooks customizados para data fetching (useShifts, useCoverage)
- Componentes com props tipadas e handlers
- Layout com espacamento consistente (space-y-6)

## Out of Scope

- Integracao com sistemas de RH ou Ponto Eletronico
- Funcionalidade de troca de plantao automatizada entre operadores (feature para V2)
- Calculo de horas trabalhadas ou relatorios de banco de horas
- Escalas segmentadas por setor (UTI, Emergencia) - escala e global por hospital
- Calendario de feriados com escalas especiais
- Escalacao em cadeia (chain of responsibility) - apenas broadcast para todos do turno
- Canais de notificacao SMS ou Push - usar email e dashboard existentes
- Recorrencia complexa (quinzenal, mensal) - apenas recorrencia semanal simples
- Interface de drag-and-drop para mover escalas - usar formulario tradicional
- Historico de alteracoes de escala (audit log) - implementar em versao futura
