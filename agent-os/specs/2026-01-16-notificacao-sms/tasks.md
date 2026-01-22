# Task Breakdown: Notificacao SMS

## Visao Geral
**Total de Tarefas:** 32 sub-tarefas distribuidas em 4 grupos

**Objetivo:** Implementar integracao com gateway SMS (Twilio) para envio de alertas criticos de obitos PCR para celulares da equipe de plantao, garantindo resiliencia com fila Redis e backoff exponencial.

---

## Lista de Tarefas

### Camada de Banco de Dados

#### Grupo de Tarefas 1: Modelos de Dados e Migracoes
**Dependencias:** Nenhuma

- [ ] 1.0 Completar camada de banco de dados
  - [ ] 1.1 Escrever 4-6 testes focados para a camada de dados
    - Testar validacao de telefone no formato E.164
    - Testar criacao de `UserNotificationPreferences` com defaults corretos
    - Testar associacao entre `User` e `UserNotificationPreferences`
    - Testar que `sms_enabled` default e TRUE quando `mobile_phone` preenchido
  - [ ] 1.2 Criar migracao para adicionar campo `mobile_phone` na tabela `users`
    - Coluna: `mobile_phone VARCHAR(16)` (nullable)
    - Formato E.164: +5511999999999
    - Criar indice se necessario para buscas
  - [ ] 1.3 Criar migracao para tabela `user_notification_preferences`
    - Campos: `id (UUID PK)`, `user_id (FK)`, `sms_enabled (BOOLEAN)`, `email_enabled (BOOLEAN)`, `dashboard_enabled (BOOLEAN)`, `created_at`, `updated_at`
    - Foreign key para `users` com `ON DELETE CASCADE`
    - Indice unico em `user_id`
    - Defaults: `sms_enabled=TRUE`, `email_enabled=TRUE`, `dashboard_enabled=TRUE`
  - [ ] 1.4 Atualizar model `User` em `internal/models/user.go`
    - Adicionar campo `MobilePhone *string` com tag `db:"mobile_phone"`
    - Adicionar validacao regex: `^\+[1-9]\d{10,14}$`
    - Atualizar `CreateUserInput` com campo `MobilePhone *string`
    - Atualizar `UpdateUserInput` com campo `MobilePhone *string`
    - Atualizar `UserResponse` para incluir telefone (mascarado em logs)
  - [ ] 1.5 Criar model `UserNotificationPreferences` em `internal/models/`
    - Struct `UserNotificationPreferences` com campos: `ID`, `UserID`, `SMSEnabled`, `EmailEnabled`, `DashboardEnabled`, `CreatedAt`, `UpdatedAt`
    - Criar `CreateNotificationPreferencesInput` e `UpdateNotificationPreferencesInput`
    - Criar `NotificationPreferencesResponse` para API
    - Metodo `ToResponse()` para conversao
  - [ ] 1.6 Criar repository `user_notification_preferences_repository.go`
    - Metodos: `Create()`, `GetByUserID()`, `Update()`, `EnsureExists()` (cria com defaults se nao existir)
    - Seguir padrao dos repositories existentes
  - [ ] 1.7 Atualizar `user_repository.go` para incluir campo `mobile_phone`
    - Atualizar queries de `Create`, `Update`, `GetByID`, `GetByEmail`
    - Adicionar metodo `GetUsersWithSMSEnabled()` para buscar usuarios aptos a receber SMS
  - [ ] 1.8 Garantir que testes da camada de dados passam
    - Executar APENAS os 4-6 testes escritos em 1.1
    - Verificar que migracoes rodam com sucesso
    - NAO executar suite completa de testes neste momento

**Criterios de Aceite:**
- Os 4-6 testes escritos em 1.1 passam
- Migracoes executam sem erros
- Campo `mobile_phone` aceita formato E.164
- Preferencias de notificacao criadas com defaults corretos
- Associacao `User` <-> `UserNotificationPreferences` funciona corretamente

---

### Camada de Servicos Backend

#### Grupo de Tarefas 2: Servico SMS e Worker com Fila Redis
**Dependencias:** Grupo de Tarefas 1

- [ ] 2.0 Completar camada de servicos SMS
  - [ ] 2.1 Escrever 4-6 testes focados para servico SMS
    - Testar `SMSService.IsConfigured()` retorna false sem credenciais
    - Testar `SMSService.IsConfigured()` retorna true com credenciais validas
    - Testar `BuildSMSMessage()` gera mensagem correta com limite de 160 caracteres
    - Testar `EnqueueSMS()` adiciona item na fila Redis
    - Testar backoff exponencial calcula delays corretos (1s, 2s, 4s, 8s, 16s)
  - [ ] 2.2 Criar `SMSConfig` e `SMSService` em `internal/services/notification/sms.go`
    - Seguir padrao do `EmailService` existente
    - Struct `SMSConfig` com: `AccountSID`, `AuthToken`, `FromPhoneNumber`
    - Carregar de variaveis de ambiente: `TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN`, `TWILIO_PHONE_NUMBER`
    - Metodo `NewSMSService()` para criacao
    - Metodo `IsConfigured()` para verificar disponibilidade
    - Erro customizado `ErrTwilioNotConfigured`
  - [ ] 2.3 Implementar integracao com Twilio API
    - Usar SDK oficial `twilio-go`
    - Metodo `SendSMS(ctx, to, message)` para envio
    - Tratar erros especificos: rate limit, numero invalido, credenciais expiradas
    - Mascarar telefone em logs (+55119****9999)
  - [ ] 2.4 Criar funcao `BuildSMSMessage()` para template
    - Template: `[SIDOT] ALERTA CRITICO: Obito PCR detectado. Hosp: {hospital_name} Idade: {age} Janela: {hours_left}h restantes. Acao: {short_link}`
    - Receber dados da ocorrencia e retornar string formatada
    - Gerar short link: `/ocorrencias/{id}`
    - Garantir limite de 160 caracteres (truncar hospital_name se necessario)
  - [ ] 2.5 Criar `SMSQueueItem` e `SMSQueueWorker` em `internal/services/notification/sms_queue.go`
    - Seguir EXATAMENTE o padrao do `EmailQueueWorker` existente
    - Struct `SMSQueueItem`: `ID`, `OccurrenceID`, `UserID`, `PhoneNumber`, `Message`, `Retries`, `CreatedAt`, `LastAttemptAt`, `NextRetryAt`, `Error`
    - Redis keys: `sidot:sms_queue` e `sidot:sms_processing`
    - Constantes: `MaxRetries = 5`, `BaseBackoffDelay = 1 * time.Second`
  - [ ] 2.6 Implementar metodos do `SMSQueueWorker`
    - `NewSMSQueueWorker()` para criacao
    - `Start()`, `Stop()`, `IsRunning()` para controle de lifecycle
    - `EnqueueSMS()` para adicionar SMS na fila
    - `processLoop()`, `processQueue()`, `processSMS()` para processamento
    - `requeue()` e `removeFromProcessing()` para gestao da fila
    - `GetStats()` e `GetQueueLength()` para monitoramento
  - [ ] 2.7 Implementar backoff exponencial com 5 tentativas
    - Delays: 1s, 2s, 4s, 8s, 16s (formula: `2^retries * BaseBackoffDelay`)
    - Apos 5 tentativas, mover para DLQ (registrar como falha)
    - Log de cada tentativa com timestamp e resultado
  - [ ] 2.8 Atualizar model `notification.go` para suportar canal SMS
    - Adicionar constante `ChannelSMS NotificationChannel = "sms"`
    - Atualizar `ValidChannels` para incluir `ChannelSMS`
    - Estender `NotificationMetadata` com: `SMSTo`, `SMSMessage`
    - Atualizar validacao `Canal` em `CreateNotificationInput`
  - [ ] 2.9 Criar metodo `CreateNotificationFromSMS()` no `NotificationRepository`
    - Seguir padrao do `CreateNotificationFromEmail()` existente
    - Registrar: `canal='sms'`, `status_envio`, `erro_mensagem` se aplicavel
    - Incluir metadata SMS
  - [ ] 2.10 Integrar disparo de SMS no fluxo de triagem existente
    - Modificar `obito_listener.go` para disparar SMS alem de email
    - Buscar usuarios com: role apropriado + `sms_enabled=TRUE` + `mobile_phone` preenchido
    - Verificar limite de 1 SMS por ocorrencia por usuario na tabela `notifications`
    - Operacao 24/7 sem restricao de horario
  - [ ] 2.11 Garantir que testes da camada de servicos passam
    - Executar APENAS os 4-6 testes escritos em 2.1
    - Verificar integracao com Redis funciona
    - NAO executar suite completa de testes neste momento

**Criterios de Aceite:**
- Os 4-6 testes escritos em 2.1 passam
- `SMSService` envia SMS via Twilio quando configurado
- `SMSQueueWorker` processa fila com backoff exponencial
- Template SMS respeita limite de 160 caracteres
- Notificacoes SMS sao registradas na tabela `notifications`
- Duplicatas sao evitadas (1 SMS por ocorrencia por usuario)

---

### Camada de API

#### Grupo de Tarefas 3: Endpoints de Preferencias de Notificacao
**Dependencias:** Grupo de Tarefas 1

- [ ] 3.0 Completar camada de API
  - [ ] 3.1 Escrever 4-6 testes focados para endpoints de API
    - Testar `PATCH /api/v1/users/{id}` atualiza telefone com formato E.164 valido
    - Testar `PATCH /api/v1/users/{id}` rejeita telefone com formato invalido
    - Testar `GET /api/v1/users/{id}/notification-preferences` retorna preferencias
    - Testar `PUT /api/v1/users/{id}/notification-preferences` atualiza `sms_enabled` e `email_enabled`
    - Testar que `dashboard_enabled` nao pode ser alterado (sempre TRUE)
  - [ ] 3.2 Atualizar endpoint `PATCH /api/v1/users/{id}` para aceitar `mobile_phone`
    - Adicionar campo `mobile_phone` no handler existente
    - Validar formato E.164 com regex
    - Retornar erro 400 se formato invalido
    - Atualizar resposta para incluir telefone
  - [ ] 3.3 Criar handler para `GET /api/v1/users/{id}/notification-preferences`
    - Verificar autorizacao (apenas proprio usuario ou admin)
    - Buscar preferencias do usuario (criar com defaults se nao existir)
    - Retornar `NotificationPreferencesResponse`
  - [ ] 3.4 Criar handler para `PUT /api/v1/users/{id}/notification-preferences`
    - Verificar autorizacao (apenas proprio usuario ou admin)
    - Aceitar apenas `sms_enabled` e `email_enabled` no body
    - Ignorar tentativas de alterar `dashboard_enabled`
    - Retornar preferencias atualizadas
  - [ ] 3.5 Registrar novas rotas no router
    - Adicionar rotas em `internal/api/routes.go` ou equivalente
    - Aplicar middlewares de autenticacao e autorizacao existentes
  - [ ] 3.6 Garantir que testes da camada de API passam
    - Executar APENAS os 4-6 testes escritos em 3.1
    - Verificar respostas HTTP corretas
    - NAO executar suite completa de testes neste momento

**Criterios de Aceite:**
- Os 4-6 testes escritos em 3.1 passam
- Campo `mobile_phone` pode ser atualizado via API
- Preferencias de notificacao podem ser consultadas e atualizadas
- `dashboard_enabled` permanece sempre TRUE
- Autorizacao funciona corretamente (usuario so edita proprio perfil, admin edita todos)

---

### Testes e Validacao

#### Grupo de Tarefas 4: Revisao de Testes e Analise de Gaps
**Dependencias:** Grupos de Tarefas 1, 2 e 3

- [ ] 4.0 Revisar testes existentes e preencher gaps criticos
  - [ ] 4.1 Revisar testes dos Grupos de Tarefas 1-3
    - Revisar os 4-6 testes escritos pelo Grupo 1 (camada de dados)
    - Revisar os 4-6 testes escritos pelo Grupo 2 (servicos)
    - Revisar os 4-6 testes escritos pelo Grupo 3 (API)
    - Total de testes existentes: aproximadamente 12-18 testes
  - [ ] 4.2 Analisar gaps de cobertura para ESTA feature apenas
    - Identificar workflows criticos do usuario sem cobertura de teste
    - Focar APENAS em gaps relacionados aos requisitos desta spec
    - NAO avaliar cobertura de toda a aplicacao
    - Priorizar testes de integracao ponta-a-ponta
  - [ ] 4.3 Escrever ate 10 testes adicionais estrategicos (maximo)
    - Testar fluxo completo: criacao de ocorrencia -> enfileiramento SMS -> envio
    - Testar cenario de retry com falha de envio
    - Testar cenario de DLQ apos 5 tentativas
    - Testar que usuario sem telefone nao recebe SMS
    - Testar que usuario com `sms_enabled=FALSE` nao recebe SMS
    - Testar limite de 1 SMS por ocorrencia por usuario
    - NAO escrever cobertura abrangente para todos os cenarios
    - Pular testes de edge cases nao criticos
  - [ ] 4.4 Executar testes especificos da feature apenas
    - Executar APENAS testes relacionados a esta feature (testes de 1.1, 2.1, 3.1 e 4.3)
    - Total esperado: aproximadamente 22-28 testes
    - NAO executar suite completa de testes da aplicacao
    - Verificar que workflows criticos passam

**Criterios de Aceite:**
- Todos os testes especificos da feature passam (aproximadamente 22-28 testes)
- Workflows criticos do usuario para esta feature estao cobertos
- No maximo 10 testes adicionais foram escritos para preencher gaps
- Testes focados exclusivamente nos requisitos desta spec

---

## Ordem de Execucao

Sequencia recomendada de implementacao:

1. **Camada de Banco de Dados (Grupo de Tarefas 1)**
   - Fundacao necessaria para todos os outros grupos
   - Cria estrutura de dados para telefone e preferencias

2. **Camada de Servicos Backend (Grupo de Tarefas 2)**
   - Depende do Grupo 1 para modelos e repositories
   - Implementa logica core de envio de SMS

3. **Camada de API (Grupo de Tarefas 3)**
   - Depende do Grupo 1 para modelos
   - Pode ser desenvolvido em paralelo com Grupo 2 se necessario

4. **Revisao de Testes (Grupo de Tarefas 4)**
   - Depende de todos os grupos anteriores
   - Validacao final da feature completa

---

## Referencias de Codigo Existente

| Arquivo | Usar como modelo para |
|---------|----------------------|
| `backend/internal/services/notification/email.go` | `SMSService` - estrutura, configuracao, error handling |
| `backend/internal/services/notification/email_queue.go` | `SMSQueueWorker` - fila Redis, backoff exponencial |
| `backend/internal/models/notification.go` | Extensao de `NotificationChannel` e `NotificationMetadata` |
| `backend/internal/models/user.go` | Adicao de campo `MobilePhone` |
| `backend/internal/repository/notification_repository.go` | `CreateNotificationFromSMS()` |
| `backend/internal/services/listener/obito_listener.go` | Integracao de disparo de SMS |

---

## Variaveis de Ambiente Necessarias

```bash
TWILIO_ACCOUNT_SID=ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
TWILIO_AUTH_TOKEN=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
TWILIO_PHONE_NUMBER=+15551234567
```

---

## Fora do Escopo (Confirmar na Spec)

- WhatsApp Business API
- SMS em massa ou marketing
- Templates editaveis pelo usuario
- Webhooks de delivery report da Twilio
- Gateway Zenvia
- Interface de preferencias completa no frontend (opcional)
- Notificacao SMS para atualizacoes de status de ocorrencia
- Shortener de URL customizado
- Internacionalizacao de mensagens SMS
- Suporte a multiplos numeros de telefone por usuario
