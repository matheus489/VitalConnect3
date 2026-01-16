# Spec Requirements: Editor de Regras de Triagem

## Initial Description

Interface visual para gestores criarem e editarem regras de elegibilidade sem necessidade de alteracao de codigo.

**Contexto:** Esta funcionalidade faz parte do VitalConnect - sistema de notificacao de doacao de orgaos. A feature permitira:
- Interface visual para gestores criarem/editarem regras de elegibilidade
- Configuracao de criterios como: idade maxima, causas de obito excludentes, tempo maximo decorrido
- Nenhuma alteracao de codigo necessaria para modificar regras
- Regras determinam se um evento de obito dispara uma notificacao

---

## Requirements Discussion

### First Round Questions

**Q1:** Estrutura da Interface - como deve ser organizada a tela de edicao de regras?
**Answer:** Pagina Unica (/dashboard/rules) com Listagem + Modal. Padrao Admin Dashboard, mais rapido que wizard. Componentes: Tabela simples (Nome, Parametro, Status, Acoes) + botao "Nova Regra" que abre Modal.

**Q2:** Tipos de Regras - quais tipos de regras devem ser suportados no MVP?
**Answer:** Focar em 3 tipos essenciais (Hardcoded no Select):
- Idade Limite (ex: > 70 anos -> Descarte)
- Janela de Tempo (ex: > 6 horas -> Descarte)
- CIDs Excludentes (ex: A00, B24 -> Descarte)
Outros tipos complexos ficam para V2.

**Q3:** Acoes das Regras - quais acoes uma regra pode executar?
**Answer:** Acao Unica: REJEITAR (Descarte). Todo obito PCR e "Potencial Doador" a menos que caia em regra de rejeicao. Nao precisa configurar "Priorizar" agora.

**Q4:** Validacao de Regras - qual nivel de validacao dos parametros?
**Answer:** Validacao Basica (Client-side):
- Idade: 0-120
- Janela: 1-48h
- CIDs: Formato A00.0 (Regex simples ou livre)
- Preview: Nao implementar - testar "na pratica"

**Q5:** Ativar/Desativar - regras podem ser desativadas temporariamente?
**Answer:** Sim, Toggle Switch na listagem. Essencial para demonstracao: mostrar flexibilidade (desativar regra -> mesmo obito aceito).

**Q6:** Prioridade das Regras - como definir a ordem de avaliacao?
**Answer:** Logica Fixa (Hardcoded). Ordem: 1. Validade do Dado -> 2. Janela de Tempo -> 3. Idade -> 4. CIDs. Usuario nao precisa ordenar, sistema aplica todas.

**Q7:** Exclusao de Regras - como tratar a remocao de regras?
**Answer:** Soft Delete (Desativar/Arquivar). Botao "Excluir" seta active = false e esconde da lista. Mantem historico para auditoria.

**Q8:** Exclusoes de Escopo - o que esta fora do MVP?
**Answer:**
- Historico de Versoes (quem mudou de 70 para 80)
- Editor Visual No-Code (drag-and-drop)
- Regras Combinadas (ex: "Idade > 60 E Peso < 50") - manter atomicas

---

### Existing Code to Reference

No similar existing features identified for reference. This is a new feature for rule management.

**Nota:** O sistema ja possui:
- Motor de Triagem implementado (aplica regras de elegibilidade)
- Dashboard de Ocorrencias (padrao de listagem a seguir)
- Configuracao de Hospitais (CRUD similar)

---

### Follow-up Questions

Nenhuma pergunta de follow-up foi necessaria. As respostas do usuario foram completas e claras.

---

## Visual Assets

### Files Provided:
Nenhum arquivo visual foi encontrado na pasta de visuals.

### Visual Insights:
Nao aplicavel - nenhum asset visual fornecido.

---

## Requirements Summary

### Functional Requirements

**Interface Principal:**
- Pagina unica em `/dashboard/rules` seguindo padrao Admin Dashboard
- Tabela de listagem com colunas: Nome, Parametro, Status (ativo/inativo), Acoes
- Botao "Nova Regra" abrindo modal de criacao/edicao
- Toggle switch para ativar/desativar regras inline na tabela
- Botao de exclusao (soft delete) nas acoes

**Tipos de Regras (MVP):**
1. **Idade Limite** - Rejeitar doadores acima de idade X (ex: > 70 anos)
2. **Janela de Tempo** - Rejeitar obitos com tempo decorrido maior que X horas (ex: > 6h)
3. **CIDs Excludentes** - Rejeitar obitos com codigos CID especificos (ex: A00, B24)

**Comportamento das Regras:**
- Acao unica: REJEITAR (descarte do potencial doador)
- Logica positiva: todo obito PCR e potencial doador ate ser rejeitado por uma regra
- Todas as regras ativas sao avaliadas (nao para na primeira)

**Ordem de Avaliacao (Fixa):**
1. Validade do Dado
2. Janela de Tempo
3. Idade
4. CIDs Excludentes

**Validacao Client-side:**
- Idade: numero inteiro entre 0 e 120
- Janela de Tempo: numero inteiro entre 1 e 48 (horas)
- CIDs: formato livre ou regex simples (padrao A00.0)

**Persistencia:**
- Soft delete para exclusao (campo active = false)
- Regras excluidas ficam ocultas mas mantidas para auditoria

### Reusability Opportunities

- Padrao de tabela do Dashboard de Ocorrencias
- Padrao de modal do Configuracao de Hospitais
- Componentes shadcn/ui ja existentes no projeto
- Toggle switch padrao do sistema de design

### Scope Boundaries

**In Scope:**
- CRUD de regras via interface visual
- 3 tipos de regras: Idade, Janela de Tempo, CIDs
- Ativacao/desativacao de regras
- Soft delete de regras
- Validacao basica client-side
- Integracao com Motor de Triagem existente

**Out of Scope:**
- Historico de versoes/alteracoes de regras
- Editor visual drag-and-drop
- Regras combinadas com operadores logicos (AND/OR)
- Preview/simulacao de impacto de regras
- Tipos de regras adicionais (peso, comorbidades, etc.)
- Acao de "Priorizar" (apenas Rejeitar)
- Ordenacao manual de prioridade de regras

### Technical Considerations

**Frontend:**
- React 18+ com Next.js 14+
- Componentes shadcn/ui
- Tailwind CSS para estilizacao
- React Hook Form + Zod para validacao de formularios
- TanStack Query para gerenciamento de estado/cache

**Backend:**
- Go (Golang) com Gin/Echo
- PostgreSQL para persistencia
- API REST para operacoes CRUD
- Integracao com Motor de Triagem existente

**Modelo de Dados (Existente):**
- Tabela `screening_rules` ja existe no banco
- Campo `active` para soft delete
- Campos para tipo de regra e parametros

**Padroes a Seguir:**
- Autenticacao JWT existente
- Sistema de roles (operador, gestor, admin)
- Padrao de layout do dashboard existente
