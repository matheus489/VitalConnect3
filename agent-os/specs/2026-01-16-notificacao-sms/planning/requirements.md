# Spec Requirements: Notificacao SMS

## Initial Description

Integrar com gateway SMS (ex: Twilio, Zenvia) para envio de alertas para celulares da equipe de plantao.

**Contexto do Produto:**
- Sistema SIDOT para notificacao de doacao de orgaos
- Janela critica de 6 horas para captacao de corneas apos obito por PCR
- Feature faz parte da Fase 2 (Produto Completo para Piloto v1.0)
- Deve funcionar junto com notificacoes por email e dashboard ja existentes

## Requirements Discussion

### First Round Questions

**Q1:** Qual gateway SMS utilizar para o MVP?
**Answer:** Twilio. Justificativa: Mais rapido para MVP (sem burocracia CNPJ/Shortcode). Zenvia fica no roadmap para producao no Estado.

**Q2:** Como armazenar o numero de telefone do usuario?
**Answer:** Campo unico (mobile_phone) no formato E.164 (+5511999999999). Validacao obrigatoria no backend. Apenas um numero (celular pessoal).

**Q3:** Como gerenciar preferencias de notificacao?
**Answer:** Implementar estrutura no Backend agora:
- SMS: Toggle (Default: TRUE se houver telefone)
- Email: Toggle (Default: TRUE)
- Dashboard: ALWAYS_ON (Nao editavel)
- Frontend: Se der tempo, colocar no perfil. Senao, hardcoded habilitado.

**Q4:** Qual arquitetura de disparo utilizar?
**Answer:** Confirmado: Redis + Worker + Backoff Exponencial (1s, 2s, 4s...) antes de mover para DLQ. Critico para provar requisito de Resiliencia.

**Q5:** Qual conteudo incluir no SMS?
**Answer:** Foco na Acao Imediata.
- Template: "[SIDOT] ALERTA CRITICO: Obito PCR detectado. Hosp: {hospital_name} Idade: {age} Janela: {hours_left}h restantes. Acao: {short_link}"
- Link deve levar direto para login/detalhe da ocorrencia.

**Q6:** Quais limites e restricoes de envio?
**Answer:**
- Horario: SEM RESTRICAO (24/7) - Transplantes nao esperam horario comercial
- Anti-Spam: Limite de 1 SMS por Ocorrencia (apenas na criacao)
- Atualizacoes de status nao enviam SMS, apenas Push/Email

**Q7:** Como registrar logs de entrega?
**Answer:** Registro Local de Disparo apenas.
- MVP: tabela notification_logs com status: sent ou failed
- Sem webhooks de delivery report para o dia 26/01

**Q8:** O que esta explicitamente fora do escopo?
**Answer:**
- WhatsApp Business API (demora aprovacao Meta)
- SMS em Massa/Marketing
- Templates Editaveis pelo Usuario

### Existing Code to Reference

Nenhuma feature similar identificada explicitamente pelo usuario. Porem, o sistema ja possui:
- Sistema de filas com Redis Streams (Fase 1 - item 4)
- Servico de Notificacao por email existente (Fase 1 - item 5)
- API de Ocorrencias (Fase 1 - item 6)

O spec-writer deve referenciar a arquitetura de notificacao por email existente para manter consistencia.

### Follow-up Questions

Nenhuma pergunta de follow-up foi necessaria. O usuario forneceu respostas completas e detalhadas para todas as questoes.

## Visual Assets

### Files Provided:
Nenhum arquivo visual encontrado na pasta de visuals.

### Visual Insights:
N/A - Nenhum asset visual fornecido.

## Requirements Summary

### Functional Requirements

**Integracao com Gateway SMS:**
- Integrar com Twilio API para envio de SMS
- Configurar credenciais via variaveis de ambiente (TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN, TWILIO_PHONE_NUMBER)
- Suportar formato E.164 para numeros de destino

**Armazenamento de Telefone:**
- Adicionar campo `mobile_phone` na tabela de usuarios
- Formato: E.164 (ex: +5511999999999)
- Validacao obrigatoria no backend antes de salvar
- Apenas um numero por usuario

**Preferencias de Notificacao:**
- Criar estrutura no backend para preferencias:
  - `sms_enabled`: boolean (default: TRUE se mobile_phone presente)
  - `email_enabled`: boolean (default: TRUE)
  - `dashboard_enabled`: boolean (default: TRUE, nao editavel)
- Frontend: opcional para MVP (pode ser hardcoded habilitado)

**Disparo de SMS:**
- Trigger: Apenas na criacao de nova ocorrencia elegivel
- Destinatarios: Usuarios com role apropriado + SMS habilitado + telefone cadastrado
- Frequencia: Maximo 1 SMS por ocorrencia por usuario
- Horario: 24/7 sem restricao

**Template de Mensagem:**
```
[SIDOT] ALERTA CRITICO: Obito PCR detectado. Hosp: {hospital_name} Idade: {age} Janela: {hours_left}h restantes. Acao: {short_link}
```

**Variaveis do Template:**
- `hospital_name`: Nome do hospital de origem
- `age`: Idade do paciente
- `hours_left`: Horas restantes na janela de captacao
- `short_link`: URL curta para acesso rapido a ocorrencia (deve autenticar ou levar para login)

**Arquitetura de Resiliencia:**
- Utilizar Redis Streams para fila de mensagens
- Worker dedicado para processamento de SMS
- Backoff exponencial em caso de falha: 1s, 2s, 4s, 8s...
- Dead Letter Queue (DLQ) apos esgotamento de retries
- Logging de cada tentativa com timestamp e resultado

**Registro de Logs:**
- Tabela `notification_logs` com campos:
  - id, user_id, occurrence_id, channel (sms/email), status (sent/failed), sent_at, error_message
- Sem webhooks de delivery report no MVP

### Reusability Opportunities

- Arquitetura de filas Redis Streams ja implementada na Fase 1
- Servico de Notificacao por email pode servir de modelo para o servico SMS
- API de Ocorrencias ja expoe dados necessarios para o template
- Sistema de autenticacao JWT existente para deep links

### Scope Boundaries

**In Scope:**
- Integracao Twilio para envio de SMS
- Campo de telefone no cadastro de usuario
- Estrutura de preferencias de notificacao no backend
- Worker com fila Redis e retry/backoff
- Template fixo de SMS com variaveis dinamicas
- Tabela de logs de envio
- Envio 24/7 sem restricao de horario
- Limite de 1 SMS por ocorrencia

**Out of Scope:**
- WhatsApp Business API
- SMS em massa ou marketing
- Templates editaveis pelo usuario
- Webhooks de delivery report (status_callback)
- Gateway Zenvia (roadmap para producao)
- Interface de preferencias no frontend (opcional se der tempo)
- Notificacao SMS para atualizacoes de status (apenas criacao)

### Technical Considerations

**Gateway SMS:**
- Provider: Twilio (MVP)
- SDK: twilio-go (biblioteca oficial)
- Fallback para Zenvia planejado para producao estadual

**Formato de Telefone:**
- Padrao E.164 obrigatorio
- Validacao com regex: `^\+[1-9]\d{10,14}$`
- Exemplo: +5511999999999

**Arquitetura:**
- Redis Streams para mensageria (consistente com Fase 1)
- Worker Go dedicado para consumir fila de SMS
- Backoff exponencial: 1s, 2s, 4s, 8s, 16s (5 tentativas antes de DLQ)

**Banco de Dados:**
- Adicionar coluna `mobile_phone` em tabela `users`
- Criar tabela `user_notification_preferences`
- Criar tabela `notification_logs`

**Seguranca:**
- Credenciais Twilio em variaveis de ambiente (nunca em codigo)
- Telefones mascarados em logs (+55119****9999)
- Rate limiting por usuario se necessario

**Integracao:**
- Reaproveitar trigger de notificacao existente (email)
- Adicionar canal SMS ao fluxo existente
- Deep link para ocorrencia deve funcionar com sistema de auth atual

**Prazo:**
- Target: 26/01
- Foco em MVP funcional sem delivery reports
