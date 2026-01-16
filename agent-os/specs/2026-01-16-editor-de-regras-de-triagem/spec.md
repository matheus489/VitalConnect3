# Specification: Editor de Regras de Triagem

## Goal
Permitir que gestores criem, editem e gerenciem regras de elegibilidade de triagem atraves de uma interface visual, sem necessidade de alteracao de codigo. As regras determinam se um evento de obito PCR dispara uma notificacao para potencial doacao de orgaos.

## User Stories
- Como gestor, quero criar e editar regras de triagem visualmente para que eu possa ajustar criterios de elegibilidade sem depender de desenvolvedores.
- Como gestor, quero ativar/desativar regras temporariamente para demonstrar a flexibilidade do sistema e testar diferentes configuracoes.

## Specific Requirements

**Pagina de Listagem de Regras**
- Nova rota `/dashboard/rules` seguindo padrao Admin Dashboard existente
- Tabela com colunas: Nome, Tipo de Regra, Parametro, Status (Ativo/Inativo), Acoes
- Botao "Nova Regra" no header da pagina abrindo modal de criacao
- Cada linha exibe o valor configurado da regra de forma legivel (ex: "Idade > 70 anos")
- Estados de loading e empty state consistentes com `OccurrencesTable.tsx`

**Toggle de Ativacao Inline**
- Switch component na coluna de status para ativar/desativar regra
- Chamada PATCH para `/api/v1/triagem-rules/:id` com `{ ativo: boolean }`
- Feedback visual imediato com toast de sucesso/erro usando Sonner
- Invalidar cache do TanStack Query apos mutacao bem-sucedida

**Modal de Criacao/Edicao de Regra**
- Usar componente Dialog do shadcn/ui existente em `components/ui/dialog.tsx`
- Formulario com React Hook Form + Zod para validacao
- Campo Select para tipo de regra (hardcoded): Idade Limite, Janela de Tempo, CIDs Excludentes
- Campos dinamicos baseados no tipo selecionado
- Botoes Cancelar e Salvar no footer do modal

**Tipos de Regras e Campos**
- Idade Limite: Input numerico (0-120), acao sempre "rejeitar"
- Janela de Tempo: Input numerico (1-48 horas), acao sempre "rejeitar"
- CIDs Excludentes: Input de texto livre para lista de codigos separados por virgula
- Cada tipo mapeia para o campo JSONB `regras` com estrutura `{ tipo, valor, acao }`

**Validacao Client-side**
- Idade: numero inteiro entre 0 e 120 (Zod `.number().int().min(0).max(120)`)
- Janela de Tempo: numero inteiro entre 1 e 48 (Zod `.number().int().min(1).max(48)`)
- CIDs: formato livre, apenas validar que nao esta vazio
- Nome da regra: obrigatorio, minimo 2 caracteres, maximo 255

**Soft Delete de Regras**
- Botao "Excluir" nas acoes da tabela
- Confirmar com Dialog de confirmacao antes de executar
- Chamada PATCH para `/api/v1/triagem-rules/:id` setando `{ ativo: false }`
- Regra excluida nao aparece mais na listagem (filtrar por `ativo = true`)

**Integracao com Motor de Triagem**
- Apos criar/editar/excluir regra, invalidar cache Redis das regras ativas
- Motor de Triagem ja possui metodo `InvalidateRulesCache()` que deve ser chamado
- Endpoint existente `ListActive` no repositorio ja filtra regras ativas

## Visual Design
Nenhum asset visual fornecido. Interface deve seguir padroes existentes do dashboard.

## Existing Code to Leverage

**Backend: Model e Tipos de Regras (`backend/internal/models/triagem_rule.go`)**
- Model `TriagemRule` ja definido com campos: ID, Nome, Descricao, Regras (JSONB), Ativo, Prioridade
- Tipos de regra ja definidos: `RuleTypeIdadeMaxima`, `RuleTypeJanelaHoras`, `RuleTypeCausasExcludentes`
- Structs de input `CreateTriagemRuleInput` e `UpdateTriagemRuleInput` prontos para uso
- Reutilizar `RuleAction = "rejeitar"` como acao padrao

**Backend: Repository (`backend/internal/repository/triagem_rule_repository.go`)**
- CRUD completo ja implementado: `List`, `GetByID`, `Create`, `Update`
- Cache Redis com TTL de 5 minutos e invalidacao automatica
- Metodo `ListActive` para buscar apenas regras ativas (usar na listagem do frontend)
- Adicionar metodo `SoftDelete` que seta `ativo = false`

**Backend: Handlers (`backend/internal/handlers/triagem.go`)**
- Endpoints `ListTriagemRules`, `CreateTriagemRule`, `UpdateTriagemRule` ja implementados
- Padrao de validacao com `validator.v10` ja configurado
- Adicionar endpoint para toggle de status e soft delete

**Frontend: Componentes de Tabela (`frontend/src/components/dashboard/OccurrencesTable.tsx`)**
- Seguir mesmo padrao de estrutura de tabela com loading skeleton
- Reutilizar imports de `Table`, `TableHeader`, `TableBody`, `TableRow`, `TableCell`
- Copiar padrao de estados de loading e empty state

**Frontend: Types (`frontend/src/types/index.ts`)**
- Interface `TriagemRule` e `TriagemRuleConfig` ja definidas
- Estender para incluir tipos de regra MVP: `idade_maxima`, `janela_horas`, `causas_excludentes`

## Out of Scope
- Historico de versoes e auditoria de alteracoes de regras
- Editor visual drag-and-drop para criar regras
- Regras combinadas com operadores logicos (AND/OR entre multiplos criterios)
- Preview ou simulacao de impacto de regras antes de salvar
- Tipos de regras adicionais alem dos 3 do MVP (peso, comorbidades, etc.)
- Acao "Priorizar" - apenas acao "Rejeitar" sera suportada
- Ordenacao manual de prioridade de regras pelo usuario
- Paginacao na listagem de regras (assumir volume pequeno no MVP)
- Busca/filtro na listagem de regras
- Validacao de codigo CID contra tabela oficial
