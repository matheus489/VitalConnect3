# Spec Requirements: Gestao de Plantoes (Shift Management)

## Initial Description

Feature para gerenciamento de escalas de plantao no VitalConnect - sistema de notificacao de doacao de orgaos.

### Funcionalidades Solicitadas:
- Cadastro de escalas de plantao com horarios e responsaveis
- Direcionamento de notificacoes para a equipe correta do turno
- Criar e gerenciar escalas de plantao
- Definir slots de tempo e funcionarios responsaveis por cada turno
- Rotear notificacoes para a equipe de plantao correta baseado no turno atual
- Suportar multiplos hospitais com diferentes configuracoes de plantao

### Contexto de Features Relacionadas:
- Gestao de Usuarios (spec recem completada) tera relacao N:N com hospitais
- Usuarios possuem roles: operador, gestor, admin
- Sistema de notificacao ja existe (email, dashboard)
- Esta feature determina QUEM recebe notificacao baseado no horario/turno atual

## Requirements Discussion

### First Round Questions

**Q1:** Formato dos Turnos - Turnos fixos pre-definidos (Manha/Tarde/Noite) ou horarios customizaveis (start_time/end_time)?
**Answer:** Horarios Customizaveis (Start/End Time). Banco salva start_time e end_time (ex: "07:00", "19:00"). Frontend tera botoes de atalho para "Diurno (07-19)" e "Noturno (19-07)".

**Q2:** Responsaveis por Turno - Responsavel unico por turno (chain of responsibility) ou broadcast para todos do turno?
**Answer:** Broadcast (Todos do Turno). Todos os operadores escalados recebem alerta simultaneamente. Primeiro que visualizar/aceitar assume (muda status para "Em Analise"). Sem escalacao complexa no MVP.

**Q3:** Gaps na Escala (Resiliencia) - O que fazer se nao houver ninguem escalado para um horario?
**Answer:** Notificar Grupo de Gestores (Fallback). Se nao houver ninguem na escala, alerta todos os Gestores do hospital. UX: exibir alerta visual se houver horarios descobertos. Fallback no backend e obrigatorio para seguranca.

**Q4:** Recorrencia das Escalas - Escalas unicas por data ou com recorrencia semanal?
**Answer:** Recorrencia Semanal Simples. Gestor define: "Segunda-feira: Dr. Joao (07:00-19:00)". Sistema replica para todas as segundas. Sem calendario de feriados ou trocas complexas - alteracao manual se necessario.

**Q5:** Transicao de Turnos - Como tratar notificacoes que chegam no limite entre turnos?
**Answer:** Baseado na Hora do Evento. Obito as 18:55 notifica equipe vigente (que sai as 19:00). Proximo turno ve caso como "Pendente" na listagem. Nao duplicar notificacao para evitar ruido.

**Q6:** Multiplos Setores - Escalas segmentadas por setor (UTI, Emergencia) ou escala unica por hospital?
**Answer:** Escala Unica por Hospital (Global). Equipe CIHDOTT responde pelo hospital inteiro. Nao segmentar por UTI/Emergencia nesta versao.

**Q7:** Permissoes - Quais acoes cada role pode executar?
**Answer:**
- Gestor/Admin: Criar, Editar e Excluir escalas
- Operador: Apenas visualiza propria escala ("Meu Plantao") e escala geral do dia

**Q8:** O que esta fora do escopo do MVP?
**Answer:**
- Integracao com RH/Ponto
- Troca de Plantao Automatizada ("solicitar troca" fica para V2)
- Calculo de Horas (nao e sistema de folha de pagamento)

### Existing Code to Reference

No similar existing features identified for reference by the user.

**Note:** The system already has:
- User management with roles (operador, gestor, admin)
- Hospital management (CRUD completo)
- Notification system (email, dashboard)
- Authentication with JWT

These existing features will be integrated with the shift management system.

### Follow-up Questions

No follow-up questions were necessary - all requirements were clearly defined.

## Visual Assets

### Files Provided:
No visual assets provided.

### Visual Insights:
N/A

## Requirements Summary

### Functional Requirements

**Cadastro de Escalas:**
- Criar escalas com horarios customizaveis (start_time, end_time)
- Atalhos de UX para turnos comuns: "Diurno (07:00-19:00)" e "Noturno (19:00-07:00)"
- Recorrencia semanal automatica (ex: toda segunda-feira)
- Escala unica por hospital (sem segmentacao por setor)

**Atribuicao de Responsaveis:**
- Vincular multiplos operadores a um turno
- Broadcast de notificacoes para todos os operadores do turno ativo
- Primeiro operador que aceitar assume o caso (status "Em Analise")

**Resiliencia e Fallback:**
- Detectar gaps na escala (horarios sem cobertura)
- Exibir alerta visual para gestores sobre horarios descobertos
- Fallback obrigatorio: notificar todos os Gestores do hospital se nao houver escala

**Roteamento de Notificacoes:**
- Determinar equipe de plantao baseado na hora do evento (obito)
- Notificacao unica para equipe vigente (sem duplicacao)
- Casos pendentes visiveis para proximo turno na listagem

**Visualizacao:**
- Gestor: ver e gerenciar todas as escalas do hospital
- Operador: ver "Meu Plantao" (proprias escalas) e escala geral do dia

### Reusability Opportunities

- Sistema de usuarios existente (roles, vinculos com hospitais)
- Sistema de notificacao ja implementado (email, dashboard)
- CRUD de hospitais existente
- Padrao de autorizacao por roles ja estabelecido

### Scope Boundaries

**In Scope:**
- CRUD de escalas de plantao
- Horarios customizaveis com atalhos para turnos comuns
- Recorrencia semanal simples
- Broadcast de notificacoes para equipe do turno
- Fallback para gestores quando nao ha cobertura
- Deteccao e alerta de gaps na escala
- Roteamento de notificacoes baseado na hora do evento
- Visualizacao de "Meu Plantao" para operadores
- Permissoes diferenciadas por role

**Out of Scope:**
- Integracao com sistemas de RH/Ponto
- Troca de plantao automatizada (feature para V2)
- Calculo de horas trabalhadas
- Escalas por setor/unidade (UTI, Emergencia)
- Calendario de feriados
- Escalacao complexa (chain of responsibility)
- Notificacao SMS/Push (usar canais existentes: email, dashboard)

### Technical Considerations

**Modelo de Dados:**
- Tabela de escalas com: hospital_id, user_id, day_of_week, start_time, end_time
- start_time e end_time como TIME (ex: "07:00", "19:00")
- Suporte a turnos que cruzam meia-noite (19:00-07:00)
- Recorrencia baseada em day_of_week (0-6)

**Integracao:**
- Conectar ao sistema de notificacoes existente
- Usar tabela de usuarios existente com roles
- Respeitar vinculos usuario-hospital existentes

**Logica de Roteamento:**
- Query para encontrar operadores de plantao no momento do evento
- Fallback query para buscar gestores do hospital
- Timestamp do evento determina turno (nao timestamp da notificacao)

**Performance:**
- Cache da escala atual (Redis) para lookups rapidos
- Invalidar cache ao modificar escalas

**UX Frontend:**
- Botoes de atalho "Diurno" e "Noturno" no formulario
- Visualizacao semanal da escala (grade de horarios)
- Indicador visual de gaps/descobertas na escala
- Aba "Meu Plantao" para operadores
