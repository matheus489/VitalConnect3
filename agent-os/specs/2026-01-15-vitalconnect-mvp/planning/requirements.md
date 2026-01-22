# Spec Requirements: SIDOT MVP

## Initial Description

MVP completo do SIDOT incluindo:

1. Deteccao automatica de obitos via integracao com sistemas hospitalares
2. Triagem inteligente para elegibilidade de doacao de corneas
3. Sistema de notificacoes em tempo real para equipes de captacao

O produto e uma solucao GovTech de middleware para Centrais de Transplantes e Bancos de Olhos, focando na janela critica de 6 horas para captacao de corneas.

---

## Requirements Discussion

### First Round Questions

**Q1:** Qual deve ser o intervalo de polling para deteccao de obitos?
**Answer:** Intervalo de 3-5 segundos (nao 30s) - precisa parecer "tempo real" no video de demonstracao.

**Q2:** Quais criterios de triagem devem ser implementados no MVP?
**Answer:** Adicionar aos criterios basicos:
- Identificacao Desconhecida (indigentes - impede contato familiar)
- Local do Obito (Setor) - priorizar UTIs e Emergencias
- Manter regras configuraveis (JSON ou tabela) para demonstrar flexibilidade

**Q3:** Quais canais de notificacao sao prioritarios?
**Answer:** Dashboard Web com alerta visual/sonoro e OBRIGATORIO (pop-up ou badge vermelho piscando). Email como canal secundario/backup.

**Q4:** Quais status devem existir no workflow de ocorrencias?
**Answer:** Adicionar status `CANCELADA` (para obitos registrados por erro) aos status existentes.

**Q5:** Qual estrutura de perfis de usuario deve ser implementada?
**Answer:** Operador, Gestor, Admin - estrutura aprovada.

**Q6:** O sistema deve suportar multiplos hospitais?
**Answer:** Sim, cadastrar HGG (Hospital Geral de Goiania) e HUGO (Hospital de Urgencias de Goias) para veracidade.

**Q7:** Quais dados do obito devem ser capturados?
**Answer:** Adicionar aos dados basicos:
- Numero do Prontuario (Patient ID)
- Setor/Leito (Ex: UTI 3, Leito 4)

**Q8:** O MVP precisa de dashboard de metricas?
**Answer:** Sim, necessario no MVP com cards simples:
- Obitos Elegiveis Detectados (Hoje)
- Tempo Medio de Notificacao
- Corneas Potenciais (Estimativa)

**Q9:** Quais sao os requisitos de disponibilidade?
**Answer:** Codigo robusto mas sem redundancia complexa. Nao quebrar com banco vazio.

**Q10:** Qual o periodo de retencao de dados?
**Answer:** 5 anos (padrao legal para prontuarios).

**Q11:** Como tratar LGPD no MVP?
**Answer:** Implementar anonimizacao no MVP - nome mascarado nas listagens (ex: Jo** Sil**), nome completo so em "Aceitar Ocorrencia" ou "Detalhes".

**Q12:** Qual deve ser o tema visual da interface?
**Answer:** Usar Shadcn/UI com tema "Clean/Hospitalar" (Branco, Cinza, Azul/Verde Saude). Vermelho apenas para alertas criticos.

**Q13:** Como simular dados para demonstracao?
**Answer:** Script seeder com:
- 5 obitos "antigos" (historico)
- 1 obito programado para 10 segundos apos inicio da demo (captura ao vivo no video)

### Existing Code to Reference

No similar existing features identified for reference - este e o primeiro desenvolvimento do produto.

### Follow-up Questions

Nenhuma pergunta de follow-up necessaria. As respostas foram completas e detalhadas.

---

## Visual Assets

### Files Provided:
Nenhum arquivo visual encontrado na pasta `/home/matheus_rubem/SIDOT/agent-os/specs/2026-01-15-sidot-mvp/planning/visuals/`.

### Visual Insights:
Nao aplicavel - seguir guidelines de design definidas nas respostas:
- Tema: Clean/Hospitalar
- Cores: Branco, Cinza, Azul/Verde Saude
- Vermelho: Apenas alertas criticos
- Componentes: Shadcn/UI

---

## Requirements Summary

### Functional Requirements

**Deteccao de Obitos:**
- Polling a cada 3-5 segundos para simular tempo real
- Integracao com tabela de obitos simulada
- Suporte a multiplos hospitais (HGG e HUGO)

**Criterios de Triagem:**
- Idade maxima para elegibilidade
- Causas de morte excludentes
- Tempo maximo desde o obito (janela de 6 horas)
- Identificacao desconhecida (flag para indigentes)
- Local do obito (priorizacao UTI/Emergencia)
- Regras configuraveis via JSON ou tabela no banco

**Sistema de Notificacoes:**
- Dashboard Web com alertas visuais (badge vermelho piscando)
- Alerta sonoro no dashboard
- Email como canal secundario/backup
- Notificacoes em tempo real para equipe de plantao

**Workflow de Ocorrencias:**
- Status: PENDENTE, EM_ANDAMENTO, ACEITA, RECUSADA, CANCELADA, CONCLUIDA
- Registro de desfecho de cada ocorrencia
- Historico completo de acoes

**Dados Capturados do Obito:**
- Nome do paciente (com anonimizacao em listagens)
- Data/hora do obito
- Causa mortis
- Idade
- Numero do prontuario (Patient ID)
- Setor/Leito (Ex: UTI 3, Leito 4)
- Hospital de origem

**Dashboard de Metricas (Cards Simples):**
- Obitos Elegiveis Detectados (Hoje)
- Tempo Medio de Notificacao
- Corneas Potenciais (Estimativa)

**Gestao de Usuarios:**
- Perfil Operador: visualiza e opera ocorrencias
- Perfil Gestor: configura regras, visualiza metricas
- Perfil Admin: gerencia usuarios e hospitais

**Gestao de Hospitais:**
- Cadastro de HGG (Hospital Geral de Goiania)
- Cadastro de HUGO (Hospital de Urgencias de Goias)
- Configuracoes de integracao por hospital

### Reusability Opportunities

- Shadcn/UI: componentes pre-construidos para interface
- TanStack Query: gerenciamento de estado e cache
- Redis Streams: fila de mensagens para MVP
- Stack definida no tech-stack.md do produto

### Scope Boundaries

**In Scope:**
- Listener de obitos com polling 3-5s
- Motor de triagem com regras configuraveis
- Dashboard de ocorrencias com alertas visuais/sonoros
- CRUD de hospitais (HGG e HUGO)
- Sistema de autenticacao com 3 perfis
- Dashboard de metricas simples (3 cards)
- Notificacao por email (secundario)
- Anonimizacao LGPD em listagens
- Seeder para demonstracao

**Out of Scope:**
- Integracao real com sistemas hospitalares (PEP)
- Notificacoes SMS e Push
- Gestao de plantoes
- Editor visual de regras de triagem
- Relatorios exportaveis (PDF/Excel)
- App mobile
- Multi-tenant para multiplas centrais
- Alta disponibilidade com redundancia

### Technical Considerations

**Stack Confirmada:**
- Backend: Go (Golang) com Gin/Echo
- Frontend: React 18+ com Next.js 14+
- Banco: PostgreSQL 15+
- Cache/Filas: Redis 7+ (Streams para MVP)
- UI: Shadcn/UI + Tailwind CSS
- Auth: JWT com refresh tokens

**Requisitos de Resiliencia:**
- Sistema nao deve quebrar com banco vazio
- Codigo robusto sem complexidade desnecessaria
- Retencao de dados: 5 anos

**Conformidade LGPD:**
- Nomes mascarados em listagens (Jo** Sil**)
- Nome completo visivel apenas em:
  - Tela "Aceitar Ocorrencia"
  - Tela "Detalhes da Ocorrencia"

**Dados de Demonstracao (Seeder):**
- 5 obitos com timestamps antigos (historico)
- 1 obito programado para T+10 segundos (demo ao vivo)

**UI/UX:**
- Tema: Clean/Hospitalar
- Cores primarias: Branco, Cinza, Azul/Verde Saude
- Vermelho: APENAS para alertas criticos
- Badge piscando para novas notificacoes
- Alerta sonoro ativavel

---

*Requisitos documentados em: 2026-01-15*
*Spec Path: /home/matheus_rubem/SIDOT/agent-os/specs/2026-01-15-sidot-mvp*
