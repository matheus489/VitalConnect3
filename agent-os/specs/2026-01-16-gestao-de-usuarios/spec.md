# Specification: Gestao de Usuarios

## Goal

Implementar CRUD completo de usuarios com atribuicao de multiplos hospitais (N:N), preferencias de notificacao por email e edicao de perfil proprio, permitindo que administradores gerenciem a equipe do sistema SIDOT.

## User Stories

- Como administrador, quero criar, editar, desativar e reativar usuarios para gerenciar a equipe que opera o sistema de notificacao de doacao de orgaos
- Como usuario do sistema, quero editar meu nome e senha para manter meus dados atualizados sem depender do administrador

## Specific Requirements

**Relacao N:N entre Usuario e Hospital**
- Criar tabela de juncao `user_hospitals` com campos: `user_id`, `hospital_id`, `created_at`
- Remover campo `hospital_id` da tabela `users` (migracao de dados necessaria)
- Indices compostos em `(user_id, hospital_id)` como chave primaria
- Foreign keys com `ON DELETE CASCADE` para user e hospital
- Atualizar model `User` em Go para incluir slice de `Hospitals`
- Atualizar `UserResponse` para retornar array de hospitais vinculados

**Preferencias de Notificacao**
- Adicionar campo `email_notifications` (boolean, default true) na tabela `users`
- Dashboard e sempre ativo e obrigatorio (nao configuravel)
- Toggle simples no formulario de criacao/edicao de usuario
- Incluir preferencia na resposta da API (`email_notifications: boolean`)

**Listagem de Usuarios com Paginacao e Busca**
- Endpoint `GET /api/v1/users` com parametros: `page`, `per_page`, `search`, `status`
- Busca server-side por nome e email (ILIKE no PostgreSQL)
- Filtro por status: `all`, `active`, `inactive`
- Retornar metadados de paginacao conforme `PaginatedResponse<T>` existente
- Ordenacao padrao por nome (ASC)

**Criacao de Usuario (Admin)**
- Campos obrigatorios: nome, email, senha provisoria, role
- Campos opcionais: hospitais vinculados, preferencia de email
- Validacao de senha: minimo 8 caracteres (usar `auth.ValidatePasswordStrength` existente)
- Hash de senha com bcrypt (usar `auth.HashPassword` existente)
- Verificar unicidade de email antes de criar
- Retornar 201 Created com dados do usuario (sem password_hash)

**Edicao de Usuario (Admin)**
- Admin pode editar: nome, role, hospitais vinculados, preferencia de email, status ativo/inativo
- Admin NAO pode editar email (identificador unico do sistema)
- Endpoint `PATCH /api/v1/users/:id` com campos opcionais
- Reativar usuario: setar `ativo = true` via campo no update

**Soft Delete (Desativar Usuario)**
- Endpoint `DELETE /api/v1/users/:id` seta `ativo = false`
- Admin nao pode desativar a si mesmo
- Usuario desativado nao consegue fazer login (validar no `auth.Service`)
- Manter todos os dados para historico e auditoria

**Edicao de Perfil Proprio**
- Endpoint `PATCH /api/v1/users/me` para usuario autenticado
- Campos permitidos: nome, senha (com confirmacao de senha atual)
- Nao pode alterar: email, role, hospitais, status
- Validar senha atual antes de permitir troca de senha

**Controle de Acesso**
- Apenas role `admin` acessa endpoints de gestao de usuarios
- Usar middleware `RequireRole("admin")` existente
- Gestor e Operador nao veem o modulo no menu/sidebar
- Qualquer usuario autenticado acessa `/api/v1/users/me`

## Visual Design

Nenhum asset visual fornecido. Seguir padroes visuais estabelecidos nas paginas de Hospitais e Dashboard de Ocorrencias.

## Existing Code to Leverage

**`backend/internal/handlers/users.go`**
- Handler CRUD basico ja implementado (ListUsers, GetUser, CreateUser, UpdateUser, DeleteUser)
- Seguir padrao de validacao com `validator.New()` e `ShouldBindJSON`
- Manter estrutura de resposta com `gin.H{"data": ..., "total": ...}`
- Extender para suportar paginacao e busca

**`backend/internal/repository/user_repository.go`**
- Repositorio com metodos List, GetByID, CreateUser, UpdateUser, DeactivateUser
- Extender `List` para aceitar filtros de paginacao e busca
- Adicionar metodos para gerenciar relacao N:N com hospitais

**`backend/internal/middleware/auth.go`**
- Middleware `RequireRole(roles ...string)` pronto para uso
- Funcao `GetUserClaims(c *gin.Context)` para obter usuario autenticado
- Usar para proteger endpoints e identificar usuario em `/me`

**`frontend/src/components/dashboard/Pagination.tsx`**
- Componente de paginacao reutilizavel com selecao de itens por pagina
- Props: currentPage, totalPages, perPage, totalItems, onPageChange, onPerPageChange

**`frontend/src/app/dashboard/hospitals/page.tsx`**
- Padrao de pagina de listagem com cards e estados de loading/error
- Reutilizar estrutura de layout e componentes Card/Badge

## Out of Scope

- SSO com Gov.br ou Active Directory (Fase 2)
- Autenticacao de dois fatores - 2FA (Fase 2)
- Grupos ou roles personalizados (manter apenas admin/gestor/operador)
- Preferencias granulares de notificacao por tipo de alerta
- Configuracao de horarios de silencio para notificacoes
- Recuperacao de senha via email (apenas futuro)
- Hard delete de usuarios (manter soft delete)
- Edicao de email do usuario (email e identificador imutavel)
- Upload de foto de perfil
- Logs de auditoria detalhados de alteracoes (usar apenas updated_at)
