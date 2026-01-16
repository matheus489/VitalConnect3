# Task Breakdown: Gestao de Usuarios

## Visao Geral
**Total de Tarefas:** 4 grupos principais com sub-tarefas

**Resumo da Feature:**
Implementar CRUD completo de usuarios com relacao N:N entre usuarios e hospitais, preferencias de notificacao por email, e edicao de perfil proprio. Apenas administradores podem gerenciar usuarios, enquanto todos os usuarios podem editar seu proprio perfil (nome e senha).

---

## Lista de Tarefas

### Camada de Banco de Dados

#### Grupo de Tarefas 1: Migracao e Modelos de Dados
**Dependencias:** Nenhuma

- [ ] 1.0 Completar camada de banco de dados
  - [ ] 1.1 Escrever 4-6 testes focados para a relacao N:N Usuario-Hospital
    - Limitar a 4-6 testes focados no maximo
    - Testar comportamentos criticos: criacao da tabela de juncao, associacao usuario-hospitais, cascade delete
    - Ignorar cobertura exaustiva de todos os cenarios
  - [ ] 1.2 Criar migracao para tabela de juncao `user_hospitals`
    - Campos: `user_id` (UUID, FK), `hospital_id` (UUID, FK), `created_at` (timestamp)
    - Chave primaria composta em `(user_id, hospital_id)`
    - Foreign keys com `ON DELETE CASCADE` para users e hospitals
    - Indices nas colunas de FK para performance
    - Seguir padrao de migracao reversivel (up/down)
  - [ ] 1.3 Criar migracao para adicionar campo `email_notifications` na tabela `users`
    - Tipo: boolean, default true
    - Migracao separada da tabela de juncao (separar schema changes)
  - [ ] 1.4 Criar migracao para remover campo `hospital_id` da tabela `users`
    - Migracao de dados: transferir vinculos existentes para `user_hospitals`
    - Executar apos 1.2 para garantir integridade dos dados
  - [ ] 1.5 Atualizar model `User` em Go
    - Adicionar campo `EmailNotifications bool`
    - Substituir `HospitalID *uuid.UUID` por `Hospitals []Hospital`
    - Atualizar metodo `ToResponse()` para incluir array de hospitais e preferencia de email
  - [ ] 1.6 Atualizar structs `CreateUserInput` e `UpdateUserInput`
    - Substituir `HospitalID` por `HospitalIDs []uuid.UUID`
    - Adicionar campo `EmailNotifications *bool`
  - [ ] 1.7 Garantir que testes da camada de dados passem
    - Executar APENAS os 4-6 testes escritos em 1.1
    - Verificar que migracoes executam com sucesso
    - NAO executar suite de testes completa neste estagio

**Criterios de Aceite:**
- Os 4-6 testes escritos em 1.1 passam
- Tabela `user_hospitals` criada corretamente com constraints
- Campo `email_notifications` adicionado a tabela `users`
- Campo `hospital_id` removido da tabela `users` com dados migrados
- Models Go atualizados com novas estruturas

---

### Camada de API (Backend)

#### Grupo de Tarefas 2: Endpoints de API
**Dependencias:** Grupo de Tarefas 1

- [ ] 2.0 Completar camada de API
  - [ ] 2.1 Escrever 6-8 testes focados para os endpoints de API
    - Limitar a 6-8 testes focados no maximo
    - Testar operacoes criticas: listagem paginada, criacao com hospitais, update de usuario, endpoint /me
    - Incluir teste de autorizacao (admin vs nao-admin)
    - Ignorar testes exaustivos de todos os cenarios de erro
  - [ ] 2.2 Atualizar `UserRepository` para suportar relacao N:N
    - Criar metodo `SetUserHospitals(ctx, userID, hospitalIDs []uuid.UUID)` para gerenciar vinculos
    - Criar metodo `GetUserHospitals(ctx, userID uuid.UUID)` para buscar hospitais vinculados
    - Atualizar `List()` para carregar hospitais via JOIN na tabela de juncao
    - Atualizar `GetModelByID()` para carregar hospitais vinculados
    - Atualizar `CreateUser()` para criar vinculos na tabela de juncao
    - Atualizar `UpdateUser()` para atualizar vinculos e preferencia de email
  - [ ] 2.3 Implementar paginacao e busca no endpoint `GET /api/v1/users`
    - Parametros: `page` (default 1), `per_page` (default 10, max 50), `search`, `status`
    - Busca server-side por nome e email usando ILIKE
    - Filtro por status: `all`, `active`, `inactive`
    - Retornar metadados de paginacao conforme `PaginatedResponse<T>` existente
    - Ordenacao padrao por nome (ASC)
  - [ ] 2.4 Atualizar handler `CreateUser` para suportar multiplos hospitais
    - Aceitar array `hospital_ids` no body
    - Chamar `SetUserHospitals` apos criar usuario
    - Incluir campo `email_notifications` (default true)
    - Manter validacoes existentes (email unico, senha forte)
  - [ ] 2.5 Atualizar handler `UpdateUser` para admin
    - Admin pode editar: nome, role, hospitais, email_notifications, ativo
    - Admin NAO pode editar email (remover essa possibilidade)
    - Chamar `SetUserHospitals` para atualizar vinculos
  - [ ] 2.6 Criar endpoint `PATCH /api/v1/users/me` para edicao de perfil proprio
    - Acessivel por qualquer usuario autenticado
    - Campos permitidos: nome, senha (com confirmacao de senha atual via `current_password`)
    - Validar senha atual antes de permitir troca
    - NAO pode alterar: email, role, hospitais, status
    - Usar `GetUserClaims()` para identificar usuario
  - [ ] 2.7 Atualizar validacao de login no `auth.Service`
    - Verificar se `ativo = false` e retornar erro apropriado
    - Usuario desativado nao consegue fazer login
  - [ ] 2.8 Garantir que testes da camada de API passem
    - Executar APENAS os 6-8 testes escritos em 2.1
    - Verificar operacoes CRUD com multiplos hospitais
    - NAO executar suite de testes completa neste estagio

**Criterios de Aceite:**
- Os 6-8 testes escritos em 2.1 passam
- Listagem com paginacao, busca e filtro funciona corretamente
- CRUD de usuarios funciona com relacao N:N de hospitais
- Endpoint `/me` permite edicao de perfil proprio
- Autorizacao por role implementada (admin para gestao, qualquer usuario para /me)
- Usuario desativado nao consegue fazer login

---

### Camada de Frontend

#### Grupo de Tarefas 3: Componentes de UI
**Dependencias:** Grupo de Tarefas 2

- [ ] 3.0 Completar componentes de UI
  - [ ] 3.1 Escrever 4-6 testes focados para componentes de UI
    - Limitar a 4-6 testes focados no maximo
    - Testar comportamentos criticos: renderizacao da listagem, submit do formulario, selecao multipla de hospitais
    - Ignorar testes exaustivos de todos os estados e interacoes
  - [ ] 3.2 Criar hook `useUsers` para gerenciamento de estado
    - Usar React Query para cache e sincronizacao
    - Implementar funcoes: `listUsers`, `getUser`, `createUser`, `updateUser`, `deleteUser`
    - Suportar parametros de paginacao e busca
    - Seguir padrao do hook `useHospitals` existente
  - [ ] 3.3 Criar pagina de listagem de usuarios `/dashboard/users`
    - Tabela com colunas: Nome, Email, Role, Hospitais, Status, Acoes
    - Componente de paginacao reutilizando `Pagination.tsx` existente
    - Campo de busca com debounce (300ms)
    - Filtro de status (Todos, Ativos, Inativos)
    - Badge para status (Ativo/Inativo) e role
    - Botao "Novo Usuario" no header
    - Estados de loading e error seguindo padrao de `hospitals/page.tsx`
  - [ ] 3.4 Criar componente `UserForm` para criacao e edicao
    - Campos: nome, email (apenas criacao), senha (apenas criacao), role (select), hospitais (multi-select), email_notifications (toggle)
    - Validacao client-side com react-hook-form + zod
    - Multi-select de hospitais com checkboxes ou combobox shadcn/ui
    - Toggle para preferencia de notificacao por email
    - Modo criacao vs edicao baseado em prop
  - [ ] 3.5 Criar modal/dialog para formulario de usuario
    - Usar Dialog do shadcn/ui
    - Abrir via botao "Novo Usuario" ou icone de edicao na tabela
    - Feedback de sucesso/erro com toast
  - [ ] 3.6 Implementar confirmacao de desativacao
    - AlertDialog do shadcn/ui antes de desativar
    - Mensagem clara sobre soft delete
    - Opcao de reativar usuario inativo
  - [ ] 3.7 Criar pagina/modal de perfil proprio
    - Acessivel por qualquer usuario via menu/avatar
    - Formulario para editar nome
    - Formulario separado para troca de senha (senha atual + nova senha + confirmacao)
    - Validacao de senha forte client-side
  - [ ] 3.8 Proteger rota e menu para admin apenas
    - Esconder item "Usuarios" no sidebar para roles nao-admin
    - Redirect para dashboard se usuario nao-admin tentar acessar `/dashboard/users`
    - Usar informacoes de role do contexto de autenticacao
  - [ ] 3.9 Garantir que testes de UI passem
    - Executar APENAS os 4-6 testes escritos em 3.1
    - Verificar comportamentos criticos dos componentes
    - NAO executar suite de testes completa neste estagio

**Criterios de Aceite:**
- Os 4-6 testes escritos em 3.1 passam
- Listagem de usuarios com paginacao, busca e filtro funciona
- Formulario de criacao/edicao valida e submete corretamente
- Multi-select de hospitais funciona
- Pagina de perfil proprio permite editar nome e senha
- Menu/rota protegido para admin apenas
- Design consistente com outras paginas do dashboard

---

### Revisao de Testes

#### Grupo de Tarefas 4: Revisao de Testes e Analise de Lacunas
**Dependencias:** Grupos de Tarefas 1-3

- [ ] 4.0 Revisar testes existentes e preencher lacunas criticas
  - [ ] 4.1 Revisar testes dos Grupos 1-3
    - Revisar os 4-6 testes escritos na camada de dados (Tarefa 1.1)
    - Revisar os 6-8 testes escritos na camada de API (Tarefa 2.1)
    - Revisar os 4-6 testes escritos na camada de UI (Tarefa 3.1)
    - Total de testes existentes: aproximadamente 14-20 testes
  - [ ] 4.2 Analisar lacunas de cobertura APENAS para esta feature
    - Identificar fluxos criticos de usuario sem cobertura de teste
    - Focar APENAS em lacunas relacionadas aos requisitos desta spec
    - NAO avaliar cobertura de toda a aplicacao
    - Priorizar fluxos end-to-end sobre testes unitarios
  - [ ] 4.3 Escrever ate 10 testes adicionais estrategicos
    - Adicionar no maximo 10 novos testes para preencher lacunas criticas identificadas
    - Focar em pontos de integracao e fluxos end-to-end
    - NAO escrever cobertura abrangente para todos os cenarios
    - Ignorar edge cases, testes de performance e acessibilidade a menos que sejam criticos para o negocio
    - Priorizar:
      - Fluxo completo de criacao de usuario com multiplos hospitais
      - Fluxo de desativacao e tentativa de login
      - Fluxo de edicao de perfil proprio
      - Fluxo de listagem com paginacao e busca
  - [ ] 4.4 Executar testes especificos da feature
    - Executar APENAS testes relacionados a esta feature (testes de 1.1, 2.1, 3.1 e 4.3)
    - Total esperado: aproximadamente 24-30 testes no maximo
    - NAO executar suite de testes completa da aplicacao
    - Verificar que fluxos criticos passam

**Criterios de Aceite:**
- Todos os testes especificos da feature passam (aproximadamente 24-30 testes no total)
- Fluxos criticos de usuario para esta feature estao cobertos
- No maximo 10 testes adicionais escritos ao preencher lacunas
- Testes focados exclusivamente nos requisitos desta spec

---

## Ordem de Execucao

Sequencia recomendada de implementacao:

1. **Camada de Banco de Dados** (Grupo 1) - Migracoes e models
2. **Camada de API** (Grupo 2) - Repository, handlers e endpoints
3. **Camada de Frontend** (Grupo 3) - Hooks, paginas e componentes
4. **Revisao de Testes** (Grupo 4) - Analise e preenchimento de lacunas

---

## Referencias Tecnicas

### Arquivos Existentes a Reutilizar
- `backend/internal/handlers/users.go` - Handler CRUD basico existente
- `backend/internal/repository/user_repository.go` - Repositorio com metodos base
- `backend/internal/middleware/auth.go` - Middleware `RequireRole()` e `GetUserClaims()`
- `backend/internal/services/auth/password.go` - `ValidatePasswordStrength()` e `HashPassword()`
- `frontend/src/components/dashboard/Pagination.tsx` - Componente de paginacao
- `frontend/src/app/dashboard/hospitals/page.tsx` - Padrao de listagem com cards

### Padroes a Seguir
- API RESTful com versionamento `/api/v1/`
- Respostas JSON com estrutura `{ "data": ..., "total": ... }` ou `PaginatedResponse<T>`
- Validacao com `validator` no backend e `zod` no frontend
- Componentes shadcn/ui para UI
- React Query para gerenciamento de estado
- Migracoes reversiveis com up/down
