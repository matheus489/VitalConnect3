# Spec Requirements: Multi-Tenant

## Initial Description

Implementar suporte a Multi-Tenant no SIDOT para permitir que multiplas Centrais de Transplante operem na mesma instancia, com isolamento de dados e configuracoes independentes. Este item corresponde ao item #21 do Roadmap (Fase 3: Expansao e Melhorias v2.0).

## Requirements Discussion

### First Round Questions

**Q1:** O que define um "tenant" no contexto do SIDOT? (Central de Transplantes Estadual, Regional, ou outro nivel?)
**Answer:** Tenant = Central de Transplantes (Estadual/Regional). Exemplos: "SES-GO", "SES-PE", "SES-SP". Hospitais sao filhos de um Tenant. Ocorrencias pertencem a um Tenant.

**Q2:** Como os usuarios serao associados a tenants e como sera feita a identificacao do tenant no momento do acesso?
**Answer:** Identificacao Hibrida (JWT Claims + Header Opcional):
- `tenant_id` embutido no Token JWT apos login
- Middleware Go extrai tenant_id do contexto do usuario
- Injeta automaticamente em todas as queries
- Evita complexidade de DNS, garante seguranca

**Q3:** Qual estrategia de isolamento de dados deve ser utilizada (banco separado, schema separado, ou coluna tenant_id)?
**Answer:** Coluna `tenant_id` (Shared Database, Shared Schema):
- Coluna `tenant_id` (UUID) em todas as tabelas relevantes
- Scopes do GORM ou middleware de repositorio para WHERE tenant_id = ?
- Controle na camada de aplicacao Go (nao RLS do Postgres)

**Q4:** Usuarios podem pertencer a multiplos tenants ou apenas a um?
**Answer:**
- Usuarios comuns: pertencem a UM unico tenant
- Super-Admin: flag `is_super_admin`, pode trocar contexto ou ver dados agregados

**Q5:** Quais entidades do sistema precisam de tenant_id e quais sao globais?
**Answer:**
- COM tenant_id: hospitals, users, occurrences, obitos_simulados, triagem_rules, shifts, notifications, audit_logs
- GLOBAL (sem tenant_id): tenants (tabela cadastro), feature_flags

**Q6:** As regras de triagem serao compartilhadas entre tenants ou cada um tera suas proprias regras?
**Answer:** Independentes com copia no cadastro:
- Novo tenant recebe copia de "Template de Regras Padrao" (legislacao federal)
- Gestor local pode editar sem afetar outros estados
- Sem heranca de regras

**Q7:** Como sera feita a migracao dos dados existentes para o modelo multi-tenant?
**Answer:** Tenant "Legacy" (SES-GO):
- Criar tabela tenants
- Inserir "SES-GO" (primeiro tenant)
- Migration adiciona tenant_id com DEFAULT para o UUID do SES-GO
- Tornar NOT NULL apos migracao

**Q8:** Sera necessario um dashboard de Super-Admin para gerenciamento de tenants ou criacao sera via scripts/banco?
**Answer:** Via banco/scripts (SEM UI nesta fase):
- Criacao de tenants e evento raro/comercial
- Usar seed ou INSERT direto

**Q9:** O que explicitamente NAO deve fazer parte desta implementacao?
**Answer:** Out of Scope:
- Custom Branding/White-label
- Billing automatizado
- Dados cross-tenant para usuarios comuns

### Existing Code to Reference

**Similar Features Identified:**
- Autenticacao JWT existente: middleware de autenticacao em `/backend/internal/middleware/auth.go`
- Sistema de roles (operador, gestor, admin): modelo de permissoes existente
- CRUD de hospitais: padrao de repositorio e handlers
- Audit logs existentes: estrutura de logging para auditoria

### Follow-up Questions

Nenhuma pergunta de follow-up necessaria - usuario forneceu decisoes completas e detalhadas.

## Visual Assets

### Files Provided:
No visual assets provided.

### Visual Insights:
N/A - Implementacao puramente de backend/infraestrutura, sem novos componentes visuais.

## Requirements Summary

### Functional Requirements

**Modelo de Tenant:**
- Criar tabela `tenants` com campos: id (UUID), name, slug, created_at, updated_at
- Cada tenant representa uma Central de Transplantes (ex: SES-GO, SES-PE, SES-SP)
- Hospitais, usuarios e todas as ocorrencias pertencem a um tenant especifico

**Identificacao de Tenant:**
- Adicionar `tenant_id` ao payload do JWT durante autenticacao
- Criar middleware Go que extrai tenant_id do contexto JWT
- Injetar tenant_id automaticamente em todas as queries via GORM scopes ou middleware de repositorio

**Isolamento de Dados:**
- Adicionar coluna `tenant_id` (UUID, NOT NULL) nas tabelas: hospitals, users, occurrences, obitos_simulados, triagem_rules, shifts, notifications, audit_logs
- Implementar scopes GORM para filtrar automaticamente por tenant_id
- Manter tabelas globais sem tenant_id: tenants, feature_flags

**Usuarios e Permissoes:**
- Usuarios comuns pertencem a exatamente um tenant
- Adicionar campo `is_super_admin` (boolean) na tabela users
- Super-admins podem alternar contexto entre tenants ou visualizar dados agregados

**Regras de Triagem:**
- Criar template padrao de regras de triagem (legislacao federal)
- Ao criar novo tenant, copiar template para regras proprias do tenant
- Gestores podem editar regras de seu tenant sem afetar outros

**Migracao de Dados:**
- Criar tenant "SES-GO" como primeiro registro
- Adicionar coluna tenant_id com DEFAULT = UUID do SES-GO
- Atualizar todos os registros existentes
- Remover DEFAULT e tornar coluna NOT NULL

**Gerenciamento de Tenants:**
- Sem interface grafica para criacao de tenants nesta fase
- Criacao via seed files ou INSERT direto no banco
- Documentar processo de criacao de novo tenant

### Reusability Opportunities

- Middleware de autenticacao existente: estender para extrair e validar tenant_id
- Padrao de repositorio existente: adicionar tenant scope como decorator/wrapper
- Sistema de roles: adicionar nivel super_admin como extensao natural
- Audit logs: ja possui estrutura, apenas adicionar tenant_id

### Scope Boundaries

**In Scope:**
- Tabela e modelo de tenants
- Middleware de identificacao de tenant via JWT
- Coluna tenant_id em todas as entidades relevantes
- Scopes GORM para isolamento automatico
- Flag is_super_admin para usuarios administrativos
- Migracao de dados existentes para tenant SES-GO
- Template de regras de triagem para novos tenants
- Documentacao do processo de criacao de tenant

**Out of Scope:**
- Interface grafica para gerenciamento de tenants
- Custom branding/white-label por tenant
- Billing ou cobranca automatizada por tenant
- Visualizacao cross-tenant para usuarios comuns
- Subdominio ou DNS customizado por tenant
- Row-Level Security (RLS) do PostgreSQL

### Technical Considerations

**Arquitetura:**
- Shared Database, Shared Schema com coluna discriminadora (tenant_id)
- Controle de acesso na camada de aplicacao Go (nao RLS)
- UUID para tenant_id garantindo unicidade global

**Middleware:**
- Extrair tenant_id do JWT claims
- Injetar no contexto da requisicao (context.Context)
- Validar que usuario pertence ao tenant solicitado

**GORM Scopes:**
- Criar scope global que adiciona WHERE tenant_id = ? automaticamente
- Garantir que scope seja aplicado em todas as queries (SELECT, UPDATE, DELETE)
- Bypass controlado apenas para super_admin

**Migracao:**
- Executar em transacao atomica
- Rollback seguro em caso de falha
- Manter compatibilidade com dados existentes

**Seguranca:**
- Tenant_id nunca deve vir do cliente (sempre do JWT)
- Validar que usuario so acessa dados de seu tenant
- Log de auditoria para acessos cross-tenant de super_admin
