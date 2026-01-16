# Task Breakdown: Editor de Regras de Triagem

## Visao Geral
Total de Tasks: 4 Grupos de Tarefas

## Lista de Tarefas

### Camada de Backend

#### Grupo de Tarefas 1: Repository e Handlers
**Dependencias:** Nenhuma (model e repository base ja existem)

- [ ] 1.0 Completar camada de backend
  - [ ] 1.1 Escrever 4-6 testes focados para funcionalidades do repository e handlers
    - Testar metodo `SoftDelete` do repository (setar `ativo = false`)
    - Testar endpoint PATCH para toggle de status
    - Testar endpoint de listagem filtrando apenas regras ativas
    - Testar invalidacao de cache Redis apos mutacao
  - [ ] 1.2 Adicionar metodo `SoftDelete` no TriagemRuleRepository
    - Arquivo: `backend/internal/repository/triagem_rule_repository.go`
    - Implementar soft delete setando `ativo = false`
    - Chamar `InvalidateRulesCache()` apos operacao
    - Seguir padrao existente dos outros metodos do repository
  - [ ] 1.3 Criar endpoint PATCH para toggle de status
    - Arquivo: `backend/internal/handlers/triagem.go`
    - Rota: `PATCH /api/v1/triagem-rules/:id`
    - Body: `{ "ativo": boolean }`
    - Validacao com `validator.v10`
    - Retornar regra atualizada com status 200
  - [ ] 1.4 Criar endpoint para soft delete de regra
    - Arquivo: `backend/internal/handlers/triagem.go`
    - Rota: `DELETE /api/v1/triagem-rules/:id`
    - Chamar metodo `SoftDelete` do repository
    - Retornar status 204 No Content
  - [ ] 1.5 Atualizar endpoint de listagem para filtrar regras ativas
    - Modificar handler `ListTriagemRules` para usar `ListActive`
    - Garantir que regras com `ativo = false` nao aparecem
  - [ ] 1.6 Registrar novas rotas no router
    - Arquivo: `backend/internal/routes/routes.go` ou similar
    - Adicionar rotas PATCH e DELETE com middleware de autenticacao
  - [ ] 1.7 Garantir que testes do backend passam
    - Executar APENAS os 4-6 testes escritos em 1.1
    - Verificar que operacoes CRUD funcionam corretamente
    - NAO executar suite de testes completa neste momento

**Criterios de Aceitacao:**
- Os 4-6 testes escritos em 1.1 passam
- Metodo SoftDelete funciona corretamente
- Toggle de status atualiza regra e invalida cache
- Listagem retorna apenas regras ativas
- Rotas protegidas por autenticacao JWT

---

### Camada de Frontend - Tipos e API

#### Grupo de Tarefas 2: Tipos TypeScript e Servico de API
**Dependencias:** Grupo de Tarefas 1

- [ ] 2.0 Completar tipos e servico de API
  - [ ] 2.1 Escrever 3-4 testes focados para servicos de API
    - Testar funcao de listagem de regras
    - Testar funcao de criacao de regra
    - Testar funcao de toggle de status
    - Testar funcao de soft delete
  - [ ] 2.2 Estender tipos TypeScript existentes
    - Arquivo: `frontend/src/types/index.ts`
    - Garantir que `TriagemRule` inclui campo `ativo: boolean`
    - Adicionar tipos para inputs de criacao/edicao
    - Tipos de regra: `idade_maxima`, `janela_horas`, `causas_excludentes`
  - [ ] 2.3 Criar schema de validacao Zod
    - Arquivo: `frontend/src/lib/validations/triagem-rule.ts`
    - Schema para Idade Limite: `.number().int().min(0).max(120)`
    - Schema para Janela de Tempo: `.number().int().min(1).max(48)`
    - Schema para CIDs Excludentes: `.string().min(1)`
    - Schema para nome: `.string().min(2).max(255)`
  - [ ] 2.4 Criar servico de API para regras de triagem
    - Arquivo: `frontend/src/services/triagem-rules.ts`
    - Funcao `listTriagemRules()`: GET `/api/v1/triagem-rules`
    - Funcao `createTriagemRule(data)`: POST `/api/v1/triagem-rules`
    - Funcao `updateTriagemRule(id, data)`: PATCH `/api/v1/triagem-rules/:id`
    - Funcao `toggleTriagemRuleStatus(id, ativo)`: PATCH `/api/v1/triagem-rules/:id`
    - Funcao `deleteTriagemRule(id)`: DELETE `/api/v1/triagem-rules/:id`
  - [ ] 2.5 Criar hooks TanStack Query
    - Arquivo: `frontend/src/hooks/use-triagem-rules.ts`
    - Hook `useTriagemRules()` para listagem com cache
    - Hook `useCreateTriagemRule()` com mutation e invalidacao
    - Hook `useUpdateTriagemRule()` com mutation e invalidacao
    - Hook `useDeleteTriagemRule()` com mutation e invalidacao
  - [ ] 2.6 Garantir que testes de API passam
    - Executar APENAS os 3-4 testes escritos em 2.1
    - Verificar que funcoes de API funcionam corretamente
    - NAO executar suite de testes completa neste momento

**Criterios de Aceitacao:**
- Os 3-4 testes escritos em 2.1 passam
- Tipos TypeScript cobrem todos os casos de uso
- Schemas Zod validam corretamente os inputs
- Servico de API comunica com backend
- Hooks TanStack Query gerenciam cache corretamente

---

### Camada de Frontend - Componentes UI

#### Grupo de Tarefas 3: Componentes de Interface
**Dependencias:** Grupo de Tarefas 2

- [ ] 3.0 Completar componentes de UI
  - [ ] 3.1 Escrever 4-6 testes focados para componentes UI
    - Testar renderizacao da tabela de regras
    - Testar toggle de status inline
    - Testar abertura e submissao do modal de criacao
    - Testar confirmacao de exclusao
  - [ ] 3.2 Criar componente RulesTable
    - Arquivo: `frontend/src/components/dashboard/RulesTable.tsx`
    - Seguir padrao de `OccurrencesTable.tsx`
    - Colunas: Nome, Tipo de Regra, Parametro, Status, Acoes
    - Estados de loading skeleton e empty state
    - Usar componentes Table do shadcn/ui
  - [ ] 3.3 Implementar toggle de status inline
    - Usar componente Switch do shadcn/ui
    - Chamar hook `useUpdateTriagemRule` ao mudar
    - Feedback visual com toast Sonner (sucesso/erro)
    - Invalidar cache apos mutacao
  - [ ] 3.4 Criar componente RuleFormModal
    - Arquivo: `frontend/src/components/dashboard/RuleFormModal.tsx`
    - Usar Dialog do shadcn/ui existente
    - React Hook Form para gerenciamento do formulario
    - Integracao com schemas Zod de validacao
    - Modo criacao e edicao (reutilizar mesmo componente)
  - [ ] 3.5 Implementar campos dinamicos por tipo de regra
    - Select para tipo: Idade Limite, Janela de Tempo, CIDs Excludentes
    - Idade Limite: Input numerico (0-120)
    - Janela de Tempo: Input numerico (1-48 horas)
    - CIDs Excludentes: Input de texto livre (virgula separada)
    - Mostrar/ocultar campos baseado no tipo selecionado
  - [ ] 3.6 Criar Dialog de confirmacao de exclusao
    - Arquivo: `frontend/src/components/dashboard/DeleteRuleDialog.tsx`
    - Usar AlertDialog do shadcn/ui
    - Texto de confirmacao claro
    - Botoes Cancelar e Confirmar Exclusao
    - Chamar hook `useDeleteTriagemRule` ao confirmar
  - [ ] 3.7 Criar pagina de listagem de regras
    - Arquivo: `frontend/src/app/dashboard/rules/page.tsx`
    - Rota: `/dashboard/rules`
    - Header com titulo e botao "Nova Regra"
    - Integrar RulesTable com dados do hook `useTriagemRules`
    - Seguir layout existente do dashboard
  - [ ] 3.8 Garantir que testes de componentes passam
    - Executar APENAS os 4-6 testes escritos em 3.1
    - Verificar que componentes renderizam corretamente
    - NAO executar suite de testes completa neste momento

**Criterios de Aceitacao:**
- Os 4-6 testes escritos em 3.1 passam
- Tabela exibe regras corretamente
- Toggle de status funciona com feedback visual
- Modal de criacao/edicao valida e salva
- Exclusao funciona com confirmacao
- Interface consistente com dashboard existente

---

### Testes e Integracao

#### Grupo de Tarefas 4: Revisao de Testes e Analise de Gaps
**Dependencias:** Grupos de Tarefas 1-3

- [ ] 4.0 Revisar testes existentes e preencher gaps criticos apenas
  - [ ] 4.1 Revisar testes dos Grupos de Tarefas 1-3
    - Revisar os 4-6 testes escritos pelo backend (Tarefa 1.1)
    - Revisar os 3-4 testes escritos para API (Tarefa 2.1)
    - Revisar os 4-6 testes escritos para UI (Tarefa 3.1)
    - Total de testes existentes: aproximadamente 11-16 testes
  - [ ] 4.2 Analisar gaps de cobertura APENAS para esta feature
    - Identificar workflows criticos de usuario sem cobertura
    - Focar APENAS em gaps relacionados aos requisitos desta spec
    - NAO avaliar cobertura de testes da aplicacao inteira
    - Priorizar fluxos end-to-end sobre gaps de testes unitarios
  - [ ] 4.3 Escrever ate 8 testes adicionais estrategicos no maximo
    - Testar fluxo completo: criar regra -> aparecer na lista
    - Testar fluxo: toggle status -> cache invalidado
    - Testar fluxo: excluir regra -> desaparecer da lista
    - Testar validacao de campos obrigatorios no formulario
    - NAO escrever cobertura exaustiva para todos os cenarios
    - Pular testes de edge cases, performance e acessibilidade
  - [ ] 4.4 Executar testes especificos da feature apenas
    - Executar APENAS testes relacionados a esta spec (testes de 1.1, 2.1, 3.1 e 4.3)
    - Total esperado: aproximadamente 19-24 testes no maximo
    - NAO executar suite de testes completa da aplicacao
    - Verificar que workflows criticos passam

**Criterios de Aceitacao:**
- Todos os testes especificos da feature passam (aproximadamente 19-24 testes)
- Workflows criticos de usuario estao cobertos
- No maximo 8 testes adicionais escritos para preencher gaps
- Testes focados exclusivamente nos requisitos desta spec

---

## Ordem de Execucao

Sequencia recomendada de implementacao:

1. **Grupo 1: Backend (Repository e Handlers)** - Base para todas as operacoes
2. **Grupo 2: Frontend Tipos e API** - Comunicacao com backend
3. **Grupo 3: Frontend Componentes UI** - Interface do usuario
4. **Grupo 4: Revisao de Testes** - Validacao final e gaps criticos

---

## Notas Tecnicas

### Stack Utilizada
- **Backend:** Go com Gin, PostgreSQL, Redis
- **Frontend:** Next.js 14+, React 18+, TypeScript
- **UI:** shadcn/ui, Tailwind CSS
- **Formularios:** React Hook Form + Zod
- **Estado:** TanStack Query

### Arquivos Existentes para Referencia
- Model: `backend/internal/models/triagem_rule.go`
- Repository: `backend/internal/repository/triagem_rule_repository.go`
- Handlers: `backend/internal/handlers/triagem.go`
- Tabela: `frontend/src/components/dashboard/OccurrencesTable.tsx`
- Tipos: `frontend/src/types/index.ts`

### Padroes a Seguir
- RESTful API com versionamento `/api/v1/`
- Soft delete com campo `ativo` ao inves de exclusao fisica
- Cache Redis com invalidacao automatica
- Componentes shadcn/ui para consistencia visual
- Validacao client-side com Zod, server-side com validator.v10
