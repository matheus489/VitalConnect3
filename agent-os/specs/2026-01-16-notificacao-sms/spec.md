# Specification: Notificacao SMS

## Goal
Implementar integracao com gateway SMS (Twilio) para envio de alertas criticos de obitos PCR para celulares da equipe de plantao, garantindo resiliencia com fila Redis e backoff exponencial.

## User Stories
- Como operador de plantao, quero receber SMS imediato quando uma ocorrencia elegivel for criada, para agir dentro da janela critica de 6 horas
- Como gestor, quero que o sistema seja resiliente a falhas de envio, para garantir que alertas criticos nao sejam perdidos

## Specific Requirements

**Integracao com Twilio API**
- Utilizar SDK oficial `twilio-go` para envio de SMS
- Configurar credenciais via variaveis de ambiente: `TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN`, `TWILIO_PHONE_NUMBER`
- Criar servico `SMSService` no pacote `internal/services/notification/` seguindo padrao do `EmailService` existente
- Implementar metodo `IsConfigured()` para verificar disponibilidade do servico
- Tratar erros especificos da API Twilio (rate limit, numero invalido, credenciais expiradas)

**Campo de Telefone no Usuario**
- Adicionar coluna `mobile_phone VARCHAR(16)` na tabela `users` (formato E.164: +5511999999999)
- Validacao backend com regex: `^\+[1-9]\d{10,14}$`
- Atualizar model `User` em `internal/models/user.go` com novo campo
- Atualizar `CreateUserInput` e `UpdateUserInput` para incluir telefone
- Mascarar telefone em logs (+55119****9999)

**Preferencias de Notificacao**
- Criar tabela `user_notification_preferences` com colunas: `user_id (FK)`, `sms_enabled`, `email_enabled`, `dashboard_enabled`
- Default: `sms_enabled=TRUE` (se telefone presente), `email_enabled=TRUE`, `dashboard_enabled=TRUE` (nao editavel)
- Criar model `UserNotificationPreferences` em `internal/models/`
- Criar repository `user_notification_preferences_repository.go`

**Worker SMS com Fila Redis**
- Criar `SMSQueueWorker` seguindo exatamente o padrao de `EmailQueueWorker` existente
- Usar Redis key `sidot:sms_queue` para fila e `sidot:sms_processing` para processamento
- Implementar backoff exponencial: 1s, 2s, 4s, 8s, 16s (5 tentativas antes de DLQ)
- Criar struct `SMSQueueItem` com campos: `id`, `occurrence_id`, `user_id`, `phone_number`, `message`, `retries`, `created_at`, `last_attempt_at`, `next_retry_at`, `error`

**Template de Mensagem SMS**
- Template fixo: `[SIDOT] ALERTA CRITICO: Obito PCR detectado. Hosp: {hospital_name} Idade: {age} Janela: {hours_left}h restantes. Acao: {short_link}`
- Criar funcao `BuildSMSMessage()` que recebe dados da ocorrencia e retorna string formatada
- Gerar short link para URL da ocorrencia (ex: `/ocorrencias/{id}`)
- Limite de 160 caracteres para evitar fragmentacao

**Disparo de Notificacoes SMS**
- Trigger apenas na criacao de nova ocorrencia elegivel (hook existente em triagem)
- Buscar usuarios com `role` apropriado + `sms_enabled=TRUE` + `mobile_phone` preenchido
- Limite de 1 SMS por ocorrencia por usuario (verificar em `notifications` antes de enfileirar)
- Operacao 24/7 sem restricao de horario

**Registro de Logs de Envio**
- Adicionar constante `ChannelSMS NotificationChannel = "sms"` no model `notification.go`
- Atualizar `ValidChannels` para incluir canal SMS
- Criar metodo `CreateNotificationFromSMS()` no `NotificationRepository`
- Registrar em `notifications`: `canal='sms'`, `status_envio='enviado'` ou `'falha'`, `erro_mensagem` se aplicavel
- Adicionar campos SMS no `NotificationMetadata`: `sms_to`, `sms_message`

**API Endpoints**
- `PATCH /api/v1/users/{id}` - Atualizar telefone do usuario (endpoint existente, apenas adicionar campo)
- `GET /api/v1/users/{id}/notification-preferences` - Buscar preferencias
- `PUT /api/v1/users/{id}/notification-preferences` - Atualizar preferencias (apenas `sms_enabled` e `email_enabled`)

## Visual Design
Nenhum asset visual fornecido. Interface de preferencias no frontend e opcional para o MVP.

## Existing Code to Leverage

**`backend/internal/services/notification/email.go`**
- Estrutura de `EmailService` como modelo para `SMSService`
- Padrao de configuracao com struct `EmailConfig` -> criar `SMSConfig`
- Metodo `IsConfigured()` para verificar disponibilidade
- Padrao de error handling com erros customizados (`ErrSMTPNotConfigured` -> `ErrTwilioNotConfigured`)

**`backend/internal/services/notification/email_queue.go`**
- `EmailQueueWorker` como modelo exato para `SMSQueueWorker`
- Padrao de `EmailQueueItem` -> `SMSQueueItem` com mesma estrutura
- Implementacao de backoff exponencial ja existente (copiar logica)
- Metodos `EnqueueEmail()` -> `EnqueueSMS()`, `processEmail()` -> `processSMS()`
- Constantes de Redis keys e retry configuradas

**`backend/internal/services/listener/obito_listener.go`**
- Padrao de Redis Streams para publicacao de eventos
- Estrutura `ObitoEvent` contem dados necessarios para template SMS (hospital, idade)
- Integracao com motor de triagem que dispara notificacoes

**`backend/internal/models/notification.go`**
- `NotificationChannel` enum ja define `ChannelDashboard` e `ChannelEmail`
- `NotificationMetadata` pode ser extendido para SMS
- `NotificationStatus` ja define estados `enviado`, `falha`, `pendente`

**`backend/internal/repository/notification_repository.go`**
- `CreateNotificationFromEmail()` como modelo para `CreateNotificationFromSMS()`
- Queries de contagem e busca por canal ja implementadas

## Out of Scope
- WhatsApp Business API (requer aprovacao Meta)
- SMS em massa ou marketing
- Templates editaveis pelo usuario
- Webhooks de delivery report da Twilio (status_callback)
- Gateway Zenvia (roadmap para producao estadual)
- Interface de preferencias completa no frontend (opcional se der tempo)
- Notificacao SMS para atualizacoes de status de ocorrencia (apenas na criacao)
- Shortener de URL customizado (usar path direto)
- Internacionalizacao de mensagens SMS
- Suporte a multiplos numeros de telefone por usuario
