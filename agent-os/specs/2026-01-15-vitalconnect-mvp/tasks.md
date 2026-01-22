# Task Breakdown: SIDOT MVP

## Overview

**Total de Tarefas:** 68 sub-tarefas organizadas em 8 grupos
**Tech Stack:** Go (Gin), React/Next.js 14+, Shadcn/UI, PostgreSQL 15+, Redis 7+

---

## Task List

---

### Setup e Infraestrutura

#### Task Group 1: Configuracao Inicial do Projeto
**Dependencias:** Nenhuma
**Complexidade:** Media

- [x] 1.0 Completar setup inicial do projeto
  - [x] 1.1 Criar estrutura de diretorios do backend Go
    - Estrutura: `cmd/`, `internal/`, `pkg/`, `migrations/`, `config/`
    - Arquivo principal: `cmd/api/main.go`
    - Configuracao: `config/config.go` com variaveis de ambiente
  - [x] 1.2 Inicializar go.mod com dependencias principais
    - gin-gonic/gin (framework HTTP)
    - lib/pq (driver PostgreSQL)
    - go-redis/redis (cliente Redis)
    - golang-jwt/jwt (autenticacao)
    - golang.org/x/crypto (bcrypt)
  - [x] 1.3 Criar projeto Next.js 14+ com App Router
    - Comando: `npx create-next-app@latest frontend --typescript --tailwind --eslint --app`
    - Estrutura: `app/`, `components/`, `lib/`, `hooks/`
  - [x] 1.4 Instalar e configurar Shadcn/UI
    - Executar: `npx shadcn-ui@latest init`
    - Componentes iniciais: Button, Card, Table, Badge, Dialog, Form, Input, Select, Toast
    - Configurar tema hospitalar no `tailwind.config.js`
  - [x] 1.5 Configurar Docker Compose para ambiente de desenvolvimento
    - Servicos: PostgreSQL 15, Redis 7
    - Volumes persistentes para dados
    - Network compartilhada entre servicos
  - [x] 1.6 Criar arquivo .env.example com variaveis necessarias
    - DATABASE_URL, REDIS_URL
    - JWT_SECRET, JWT_REFRESH_SECRET
    - SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS
    - CORS_ORIGINS
  - [x] 1.7 Verificar setup inicial
    - Backend compila sem erros
    - Frontend inicia em modo dev
    - Docker Compose sobe todos os servicos

**Criterios de Aceitacao:**
- Projeto Go compila com `go build ./...`
- Frontend Next.js inicia com `npm run dev`
- Docker Compose sobe PostgreSQL e Redis
- Shadcn/UI configurado com tema hospitalar

---

### Camada de Banco de Dados

#### Task Group 2: Modelos e Migrations
**Dependencias:** Task Group 1
**Complexidade:** Alta

- [x] 2.0 Completar camada de banco de dados
  - [x] 2.1 Escrever 4-6 testes para modelos principais
    - Testar validacao de campos obrigatorios em User
    - Testar enum de roles (operador, gestor, admin)
    - Testar validacao de status de ocorrencia
    - Testar relacionamentos hospital-ocorrencia
  - [x] 2.2 Criar migration para tabela `hospitals`
    - Campos: id (UUID), nome, codigo, endereco, config_conexao (JSONB), ativo (boolean)
    - Timestamps: created_at, updated_at, deleted_at (soft delete)
    - Index: codigo (UNIQUE)
  - [x] 2.3 Criar migration para tabela `users`
    - Campos: id (UUID), email, password_hash, nome, role (enum), hospital_id (FK nullable)
    - Timestamps: created_at, updated_at
    - Index: email (UNIQUE)
    - Constraint: role IN ('operador', 'gestor', 'admin')
  - [x] 2.4 Criar migration para tabela `obitos_simulados` (fonte de dados)
    - Campos: id (UUID), nome_paciente, data_nascimento, data_obito, causa_mortis
    - Campos adicionais: prontuario, setor, leito, identificacao_desconhecida (boolean)
    - FK: hospital_id
    - Index: data_obito, hospital_id
  - [x] 2.5 Criar migration para tabela `occurrences`
    - Campos: id (UUID), obito_id (FK), hospital_id (FK), status (enum), score_priorizacao
    - Campos LGPD: nome_paciente_mascarado, dados_completos (JSONB encrypted)
    - Timestamps: created_at, updated_at, notificado_em
    - Index: status, hospital_id, created_at
    - Constraint: status IN ('PENDENTE', 'EM_ANDAMENTO', 'ACEITA', 'RECUSADA', 'CANCELADA', 'CONCLUIDA')
  - [x] 2.6 Criar migration para tabela `occurrence_history`
    - Campos: id (UUID), occurrence_id (FK), user_id (FK), acao, status_anterior, status_novo
    - Campos: observacoes (text), desfecho (text nullable)
    - Timestamps: created_at
    - Index: occurrence_id, created_at
  - [x] 2.7 Criar migration para tabela `triagem_rules`
    - Campos: id (UUID), nome, descricao, regras (JSONB), ativo (boolean), prioridade
    - Timestamps: created_at, updated_at
    - Index GIN: regras
  - [x] 2.8 Criar migration para tabela `notifications`
    - Campos: id (UUID), occurrence_id (FK), user_id (FK nullable), canal (enum), enviado_em
    - Campos: status_envio, erro_mensagem
    - Constraint: canal IN ('dashboard', 'email')
  - [x] 2.9 Criar modelos Go com validacoes
    - Structs em `internal/models/`
    - Tags de validacao com go-playground/validator
    - Metodos de mascaramento LGPD
  - [x] 2.10 Executar migrations e verificar
    - Rodar todas as migrations em ordem
    - Verificar criacao de tabelas e indexes
    - Rodar testes do 2.1

**Criterios de Aceitacao:**
- Os 4-6 testes do 2.1 passam
- Todas as migrations executam sem erro
- Indexes e constraints criados corretamente
- Modelos Go com validacoes funcionando

---

### Camada de Autenticacao

#### Task Group 3: Sistema de Autenticacao JWT
**Dependencias:** Task Group 2
**Complexidade:** Media

- [x] 3.0 Completar sistema de autenticacao
  - [x] 3.1 Escrever 4-6 testes para autenticacao
    - Testar login com credenciais validas
    - Testar rejeicao de credenciais invalidas
    - Testar geracao e validacao de JWT
    - Testar refresh token flow
    - Testar rate limiting (5 tentativas/minuto)
  - [x] 3.2 Implementar servico de hash de senha
    - Usar bcrypt com cost factor 12
    - Funcoes: HashPassword, CheckPasswordHash
    - Local: `internal/services/auth/password.go`
  - [x] 3.3 Implementar geracao e validacao de JWT
    - Access token: 15 minutos de expiracao
    - Refresh token: 7 dias de expiracao
    - Claims: user_id, email, role, hospital_id
    - Local: `internal/services/auth/jwt.go`
  - [x] 3.4 Criar middleware de autenticacao Gin
    - Extrair e validar token do header Authorization
    - Injetar user claims no gin.Context
    - Retornar 401 para token invalido/expirado
    - Local: `internal/middleware/auth.go`
  - [x] 3.5 Criar middleware de autorizacao por role
    - Funcao: RequireRole(roles ...string)
    - Verificar role do usuario contra roles permitidos
    - Retornar 403 para acesso negado
  - [x] 3.6 Implementar rate limiting no login
    - Limite: 5 tentativas por minuto por IP
    - Usar Redis para contagem
    - Retornar 429 Too Many Requests
  - [x] 3.7 Criar endpoints de autenticacao
    - POST /api/v1/auth/login - retorna access + refresh tokens
    - POST /api/v1/auth/refresh - renova access token
    - POST /api/v1/auth/logout - invalida refresh token
    - GET /api/v1/auth/me - retorna dados do usuario logado
  - [x] 3.8 Executar testes de autenticacao
    - Rodar testes do 3.1
    - Verificar todos os cenarios

**Criterios de Aceitacao:**
- Os 4-6 testes do 3.1 passam
- Login funciona com credenciais corretas
- Tokens gerados e validados corretamente
- Rate limiting bloqueia apos 5 tentativas

---

### Camada de API Backend

#### Task Group 4: API REST de Recursos
**Dependencias:** Task Group 3
**Complexidade:** Alta

- [x] 4.0 Completar API REST
  - [x] 4.1 Escrever 6-8 testes para endpoints principais
    - Testar listagem de ocorrencias com paginacao
    - Testar filtro por status e hospital
    - Testar transicao de status valida
    - Testar rejeicao de transicao invalida
    - Testar registro de desfecho
    - Testar endpoint de metricas
  - [x] 4.2 Implementar CRUD de hospitais
    - GET /api/v1/hospitals - listar todos (apenas Admin)
    - GET /api/v1/hospitals/:id - detalhes
    - POST /api/v1/hospitals - criar (apenas Admin)
    - PATCH /api/v1/hospitals/:id - atualizar (apenas Admin)
    - DELETE /api/v1/hospitals/:id - soft delete (apenas Admin)
  - [x] 4.3 Implementar CRUD de usuarios
    - GET /api/v1/users - listar (apenas Admin)
    - GET /api/v1/users/:id - detalhes (Admin ou proprio usuario)
    - POST /api/v1/users - criar (apenas Admin)
    - PATCH /api/v1/users/:id - atualizar (Admin ou proprio usuario)
    - DELETE /api/v1/users/:id - desativar (apenas Admin)
  - [x] 4.4 Implementar API de ocorrencias
    - GET /api/v1/occurrences - listar com paginacao, filtros (status, hospital, data)
    - GET /api/v1/occurrences/:id - detalhes completos (nome sem mascara)
    - GET /api/v1/occurrences/:id/history - historico de acoes
  - [x] 4.5 Implementar transicoes de status
    - PATCH /api/v1/occurrences/:id/status - atualizar status
    - Validar transicoes permitidas no backend
    - Registrar automaticamente no historico
    - Transicoes: PENDENTE->EM_ANDAMENTO->ACEITA/RECUSADA->CONCLUIDA; qualquer->CANCELADA
  - [x] 4.6 Implementar registro de desfecho
    - POST /api/v1/occurrences/:id/outcome - registrar desfecho
    - Campo obrigatorio ao transicionar para CONCLUIDA
    - Opcoes: sucesso_captacao, familia_recusou, contraindicacao_medica, tempo_excedido
  - [x] 4.7 Implementar endpoint de metricas
    - GET /api/v1/metrics/dashboard
    - Retornar: obitos_elegiveis_hoje, tempo_medio_notificacao, corneas_potenciais
    - Query agregada no PostgreSQL
  - [x] 4.8 Implementar CRUD de regras de triagem
    - GET /api/v1/triagem-rules - listar (Gestor, Admin)
    - POST /api/v1/triagem-rules - criar (Gestor, Admin)
    - PATCH /api/v1/triagem-rules/:id - atualizar (Gestor, Admin)
    - Invalidar cache Redis ao atualizar
  - [x] 4.9 Configurar middleware CORS
    - Permitir origens configuradas via env
    - Headers: Authorization, Content-Type
    - Methods: GET, POST, PATCH, DELETE, OPTIONS
  - [x] 4.10 Executar testes da API
    - Rodar testes do 4.1
    - Verificar todos os endpoints

**Criterios de Aceitacao:**
- Os 6-8 testes do 4.1 passam
- Todos os endpoints CRUD funcionam
- Transicoes de status validadas corretamente
- Autorizacao por role aplicada

---

### Servicos de Background

#### Task Group 5: Listener de Obitos e Motor de Triagem
**Dependencias:** Task Group 4
**Complexidade:** Alta

- [x] 5.0 Completar servicos de background
  - [x] 5.1 Escrever 4-6 testes para servicos
    - Testar deteccao de novo obito
    - Testar idempotencia (nao reprocessar)
    - Testar aplicacao de regra de triagem
    - Testar criacao automatica de ocorrencia
  - [x] 5.2 Implementar Listener de Obitos
    - Polling a cada 3-5 segundos na tabela obitos_simulados
    - Detectar novos registros por timestamp
    - Suportar multiplos hospitais (HGG, HUGO)
    - Processar em goroutines separadas
    - Local: `internal/services/listener/obito_listener.go`
  - [x] 5.3 Implementar publicacao no Redis Streams
    - Usar XADD para publicar eventos de obito detectado
    - Stream: `obitos:detectados`
    - Payload: obito_id, hospital_id, timestamp_deteccao
  - [x] 5.4 Implementar Motor de Triagem
    - Consumer group para Redis Streams
    - Consumir eventos do stream `obitos:detectados`
    - Local: `internal/services/triagem/motor.go`
  - [x] 5.5 Implementar regras de elegibilidade
    - Carregar regras do PostgreSQL (JSONB)
    - Cache em Redis com TTL de 5 minutos
    - Criterios: idade_maxima, causas_excludentes, janela_6h
    - Criterios adicionais: identificacao_desconhecida, setor
  - [x] 5.6 Implementar score de priorizacao
    - UTI: score 100
    - Emergencia: score 80
    - Outros setores: score 50
    - Ajuste por tempo restante na janela
  - [x] 5.7 Criar ocorrencia automaticamente
    - Status inicial: PENDENTE
    - Aplicar mascaramento LGPD no nome
    - Armazenar dados completos em JSONB
    - Registrar no historico: "Ocorrencia criada automaticamente"
  - [x] 5.8 Implementar health check do listener
    - GET /api/v1/health/listener
    - Retornar status, ultimo_processamento, obitos_detectados_hoje
  - [x] 5.9 Implementar logging estruturado
    - Log de cada deteccao com timestamp e hospital
    - Log de triagem com resultado (elegivel/inelegivel)
    - Log de rejeicao com motivo para auditoria
  - [x] 5.10 Executar testes dos servicos
    - Rodar testes do 5.1
    - Verificar processamento end-to-end

**Criterios de Aceitacao:**
- Os 4-6 testes do 5.1 passam
- Listener detecta obitos em 3-5 segundos
- Triagem aplica regras corretamente
- Ocorrencias criadas automaticamente

---

### Camada de Notificacoes

#### Task Group 6: Sistema de Notificacoes em Tempo Real
**Dependencias:** Task Group 5
**Complexidade:** Media

- [x] 6.0 Completar sistema de notificacoes
  - [x] 6.1 Escrever 3-5 testes para notificacoes
    - Testar envio de evento SSE
    - Testar formatacao de email
    - Testar registro de notificacao enviada
  - [x] 6.2 Implementar endpoint SSE
    - GET /api/v1/notifications/stream
    - Autenticacao via query param token
    - Enviar eventos de nova ocorrencia
    - Manter conexao aberta com heartbeat
    - Local: `internal/handlers/sse.go`
  - [x] 6.3 Implementar publicacao de eventos SSE
    - Ao criar ocorrencia, publicar no canal SSE
    - Payload: occurrence_id, hospital, setor, tempo_restante
    - Usar Redis Pub/Sub para distribuir entre instancias
  - [x] 6.4 Implementar servico de email
    - Template HTML para notificacao de obito
    - Campos: hospital, setor, hora_obito, tempo_restante
    - Suporte a SMTP e SendGrid
    - Local: `internal/services/notification/email.go`
  - [x] 6.5 Implementar envio assincrono de email
    - Fila no Redis para emails pendentes
    - Worker processando fila em background
    - Retry com backoff exponencial (max 3 tentativas)
  - [x] 6.6 Registrar notificacoes enviadas
    - Inserir na tabela notifications
    - Campos: canal, enviado_em, status_envio
    - Usar para calculo de tempo medio de notificacao
  - [x] 6.7 Executar testes de notificacoes
    - Rodar testes do 6.1
    - Verificar fluxo completo

**Criterios de Aceitacao:**
- Os 3-5 testes do 6.1 passam
- SSE envia eventos em tempo real
- Emails formatados corretamente
- Notificacoes registradas no banco

---

### Camada de Frontend

#### Task Group 7: Interface do Usuario
**Dependencias:** Task Group 6
**Complexidade:** Alta

- [x] 7.0 Completar interface do usuario
  - [x] 7.1 Escrever 4-6 testes para componentes principais
    - Testar renderizacao da tabela de ocorrencias
    - Testar formulario de login
    - Testar transicao de status via UI
    - Testar exibicao de metricas
  - [x] 7.2 Configurar tema hospitalar no Tailwind
    - Cores primarias: Branco (#FFFFFF), Cinza (#F3F4F6, #6B7280)
    - Azul Saude: #0EA5E9
    - Verde sucesso: #10B981
    - Vermelho alertas: #EF4444
    - Fonte: Inter ou system-ui
  - [x] 7.3 Criar layout principal com sidebar
    - Sidebar fixa a esquerda com navegacao
    - Header com logo, usuario logado, logout
    - Area principal para conteudo
    - Badge de notificacao no header
  - [x] 7.4 Implementar pagina de login
    - Formulario com email e senha
    - Validacao client-side
    - Integracao com endpoint /auth/login
    - Redirect para dashboard apos sucesso
  - [x] 7.5 Implementar hook useAuth
    - Gerenciar tokens em localStorage
    - Auto-refresh de access token
    - Funcoes: login, logout, isAuthenticated
    - Context provider para app
  - [x] 7.6 Implementar dashboard de metricas
    - 3 cards no topo: Obitos Elegiveis, Tempo Medio, Corneas Potenciais
    - Usar TanStack Query com refetchInterval de 30s
    - Componentes Card do Shadcn/UI
  - [x] 7.7 Implementar tabela de ocorrencias
    - Componente Table do Shadcn/UI
    - Colunas: Hospital, Setor, Paciente (mascarado), Status, Tempo Restante, Acoes
    - Paginacao no rodape
    - Filtros por status e hospital
  - [x] 7.8 Implementar filtros e ordenacao
    - Select para filtro de status
    - Select para filtro de hospital
    - DatePicker para filtro de data
    - Ordenacao por prioridade e tempo restante
  - [x] 7.9 Implementar modal de detalhes da ocorrencia
    - Exibir dados completos (nome sem mascara)
    - Historico de acoes
    - Botoes de acao conforme status atual
    - Dialog do Shadcn/UI
  - [x] 7.10 Implementar transicao de status via UI
    - Botoes: "Assumir", "Aceitar", "Recusar", "Cancelar", "Concluir"
    - Confirmacao antes de acao
    - useMutation do TanStack Query
    - Invalidar query de ocorrencias apos mutacao
  - [x] 7.11 Implementar formulario de desfecho
    - Modal ao clicar em "Concluir"
    - Select com opcoes de desfecho
    - Campo de observacoes (textarea)
    - Validacao de campo obrigatorio
  - [x] 7.12 Implementar notificacoes em tempo real
    - Hook useSSE para conectar ao stream
    - Toast com dados resumidos ao receber evento
    - Badge vermelho piscando (CSS animation)
    - Alerta sonoro com Web Audio API (toggle on/off)
  - [x] 7.13 Implementar responsividade
    - Mobile: 320px - 768px (menu colapsavel)
    - Tablet: 768px - 1024px
    - Desktop: 1024px+
    - Tabela responsiva com scroll horizontal em mobile
  - [x] 7.14 Executar testes de componentes
    - Rodar testes do 7.1
    - Verificar renderizacao e interacoes

**Criterios de Aceitacao:**
- Os 4-6 testes do 7.1 passam
- Login funciona end-to-end
- Dashboard exibe metricas atualizadas
- Notificacoes em tempo real funcionam
- Layout responsivo em todos os breakpoints

---

### Testes e Finalizacao

#### Task Group 8: Revisao de Testes, Seeder e Finalizacao
**Dependencias:** Task Groups 1-7
**Complexidade:** Media

- [x] 8.0 Revisar testes e criar dados de demonstracao
  - [x] 8.1 Revisar testes existentes dos grupos 2-7
    - Verificar cobertura dos testes de modelos (2.1) - 9 testes
    - Verificar cobertura dos testes de auth (3.1) - 7 testes
    - Verificar cobertura dos testes de API (4.1) - 8 testes
    - Verificar cobertura dos testes de servicos (5.1) - 21 testes (listener + triagem)
    - Verificar cobertura dos testes de notificacoes (6.1) - 10 testes
    - Verificar cobertura dos testes de UI (7.1) - 24 testes
    - Total atual: 87 testes (muito acima do esperado 25-37)
  - [x] 8.2 Identificar gaps criticos de cobertura
    - Focar em fluxos end-to-end do usuario
    - Priorizar integracao entre componentes
    - Ignorar edge cases nao criticos
  - [x] 8.3 Escrever ate 10 testes adicionais se necessario
    - Fluxo completo: obito detectado -> notificacao -> aceitacao
    - Fluxo de autenticacao completo
    - Fluxo de transicao de status completo
    - Criado: `/backend/internal/integration/e2e_test.go` com 8 testes E2E
  - [x] 8.4 Criar script seeder para hospitais
    - HGG: Hospital Geral de Goiania
    - HUGO: Hospital de Urgencias de Goias
    - Configuracoes de conexao simuladas
    - Criado: `/backend/cmd/seeder/main.go`
  - [x] 8.5 Criar script seeder para usuarios
    - admin@sidot.gov.br (Admin)
    - gestor@sidot.gov.br (Gestor)
    - operador@sidot.gov.br (Operador)
    - Senha padrao: "demo123"
    - Criado: `/backend/cmd/seeder/main.go`
  - [x] 8.6 Criar script seeder para regras de triagem
    - Regra 1: Idade maxima 80 anos
    - Regra 2: Causas excludentes (lista)
    - Regra 3: Janela de 6 horas
    - Regra 4: Identificacao desconhecida = inelegivel
    - Criado: `/backend/cmd/seeder/main.go`
  - [x] 8.7 Criar script seeder para obitos de demonstracao
    - 5 obitos com timestamps de 1-24 horas atras
    - Distribuir entre HGG e HUGO
    - Incluir casos elegiveis e inelegiveis
    - Variar setores (UTI, Emergencia, Enfermaria)
    - Criado: `/backend/cmd/seeder/main.go`
  - [x] 8.8 Criar script de obito programado para demo ao vivo
    - Inserir obito elegivel T+10 segundos apos execucao
    - Hospital: HGG, Setor: UTI
    - Paciente ficticio com dados completos
    - Log indicando quando obito sera inserido
    - Criado: `/backend/cmd/demo/live_demo.go`
  - [x] 8.9 Executar suite completa de testes
    - Rodar todos os testes da feature
    - Verificar que 100% passam
    - Total atual: 95 testes (87 existentes + 8 E2E)
  - [x] 8.10 Documentar comandos de execucao
    - Comando para rodar migrations
    - Comando para executar seeder
    - Comando para iniciar backend
    - Comando para iniciar frontend
    - Comando para executar demo ao vivo
    - Criado: `/QUICKSTART.md`

**Criterios de Aceitacao:**
- Todos os testes passam (95 testes - acima do esperado 35-47)
- Seeder cria dados completos para demo
- Demo ao vivo funciona (obito T+10s detectado)
- Comandos documentados e funcionais

---

## Ordem de Execucao Recomendada

A implementacao deve seguir esta sequencia para respeitar dependencias:

```
1. Task Group 1: Setup e Infraestrutura
   |
   v
2. Task Group 2: Modelos e Migrations
   |
   v
3. Task Group 3: Sistema de Autenticacao
   |
   v
4. Task Group 4: API REST de Recursos
   |
   v
5. Task Group 5: Listener de Obitos e Motor de Triagem
   |
   v
6. Task Group 6: Sistema de Notificacoes
   |
   v
7. Task Group 7: Interface do Usuario
   |
   v
8. Task Group 8: Revisao de Testes, Seeder e Finalizacao
```

---

## Resumo por Especializacao

| Grupo | Especializacao | Tarefas | Complexidade | Status |
|-------|----------------|---------|--------------|--------|
| 1 | DevOps/Setup | 7 | Media | Completo |
| 2 | Backend/Database | 10 | Alta | Completo |
| 3 | Backend/Auth | 8 | Media | Completo |
| 4 | Backend/API | 10 | Alta | Completo |
| 5 | Backend/Services | 10 | Alta | Completo |
| 6 | Backend/Notifications | 7 | Media | Completo |
| 7 | Frontend/UI | 14 | Alta | Completo |
| 8 | QA/Finalizacao | 10 | Media | Completo |

---

## Notas Importantes

1. **LGPD**: Todas as listagens devem usar nomes mascarados. Nome completo apenas em detalhes e modal de aceitar.

2. **Polling**: O listener deve fazer polling a cada 3-5 segundos para parecer tempo real na demo.

3. **Testes**: Cada grupo escreve 4-8 testes focados. Total final: 95 testes (muito acima do esperado).

4. **Demo ao Vivo**: O seeder inclui obito programado para T+10 segundos para capturar deteccao ao vivo no video.

5. **Redis Streams**: Usar XADD/XREADGROUP para comunicacao entre listener e motor de triagem.

6. **SSE**: Server-Sent Events para notificacoes push no dashboard (nao WebSocket).

---

## Arquivos Criados no Task Group 8

| Arquivo | Descricao |
|---------|-----------|
| `/backend/cmd/seeder/main.go` | Script seeder para hospitais, usuarios, regras e obitos |
| `/backend/cmd/demo/live_demo.go` | Script para demo ao vivo com obito T+10s |
| `/backend/internal/integration/e2e_test.go` | 8 testes E2E para fluxos completos |
| `/QUICKSTART.md` | Documentacao de comandos de execucao |

---

*Tarefas criadas em: 2026-01-15*
*Task Group 8 completado em: 2026-01-15*
*Spec Path: /home/matheus_rubem/SIDOT/agent-os/specs/2026-01-15-sidot-mvp*
