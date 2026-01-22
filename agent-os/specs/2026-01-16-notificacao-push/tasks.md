# Task Breakdown: Notificacao Push via PWA

## Visao Geral
Total de Tarefas: 4 grupos principais

**Objetivo:** Implementar notificacoes push via Web Push (PWA) usando Firebase Cloud Messaging para alertar equipes de plantao em tempo real sobre novas ocorrencias de transplante.

**Stack:** Go (backend), Next.js/React (frontend), PostgreSQL, Redis, Firebase Cloud Messaging

## Lista de Tarefas

### Camada de Banco de Dados

#### Grupo de Tarefas 1: Tabela push_subscriptions e Modelo
**Dependencias:** Nenhuma

- [ ] 1.0 Completar camada de banco de dados para push subscriptions
  - [ ] 1.1 Escrever 2-4 testes focados para o modelo PushSubscription
    - Testar criacao de subscription com campos obrigatorios
    - Testar constraint UNIQUE em fcm_token
    - Testar relacionamento com usuario (multiplos dispositivos por usuario)
    - Testar remocao automatica de tokens invalidos
  - [ ] 1.2 Criar migration `008_create_push_subscriptions.sql`
    - Seguir padrao existente das migrations (`007_create_notifications.sql`)
    - Campos: id (UUID), user_id (FK para users), fcm_token (VARCHAR(255) UNIQUE), device_info (JSONB), created_at, updated_at
    - Implementar UP e DOWN para reversibilidade
    - Adicionar ALTER TYPE notification_channel ADD VALUE 'push' se enum existir
  - [ ] 1.3 Criar indices para performance
    - Index composto em (user_id, fcm_token) para busca eficiente
    - Index em user_id para listar dispositivos do usuario
    - Comentarios explicativos no SQL
  - [ ] 1.4 Criar model PushSubscription em Go
    - Arquivo: `backend/internal/models/push_subscription.go`
    - Struct com tags JSON e db apropriadas
    - Validacoes: fcm_token nao vazio, tamanho maximo 255
  - [ ] 1.5 Criar repository para PushSubscription
    - Arquivo: `backend/internal/repository/push_subscription_repository.go`
    - Metodos: Create, FindByUserID, FindByToken, DeleteByToken, GetAllActiveTokens
    - Seguir padrao dos repositories existentes
  - [ ] 1.6 Garantir que testes da camada de dados passam
    - Executar APENAS os 2-4 testes escritos em 1.1
    - Verificar que migrations executam com sucesso
    - NAO executar suite completa de testes neste estagio

**Criterios de Aceitacao:**
- Os 2-4 testes escritos em 1.1 passam
- Migration executa sem erros (UP e DOWN)
- Constraint UNIQUE impede tokens duplicados
- Relacionamento usuario-dispositivos funciona corretamente

---

### Camada de Backend / API

#### Grupo de Tarefas 2: PushService e Endpoints API
**Dependencias:** Grupo de Tarefas 1

- [ ] 2.0 Completar camada de servico e API para push notifications
  - [ ] 2.1 Escrever 3-5 testes focados para PushService e endpoints
    - Testar endpoint POST `/api/v1/push/subscribe` com autenticacao JWT
    - Testar envio de notificacao via FCM (mock do SDK)
    - Testar tratamento de tokens invalidos/expirados
    - Testar busca de tokens por role (operador/gestor)
  - [ ] 2.2 Configurar Firebase Admin SDK para Go
    - Instalar dependencia: `firebase.google.com/go/v4/messaging`
    - Criar arquivo de configuracao para credenciais
    - Variaveis de ambiente: FIREBASE_PROJECT_ID, FIREBASE_SERVICE_ACCOUNT
    - Arquivo: `backend/internal/config/firebase.go`
  - [ ] 2.3 Criar PushService seguindo padrao do EmailService
    - Arquivo: `backend/internal/services/notification/push.go`
    - Struct PushService com client FCM e repository
    - Constructor NewPushService com injecao de dependencias
    - Metodo IsConfigured() para verificar credenciais
    - Erro customizado: ErrFCMNotConfigured
  - [ ] 2.4 Implementar metodo SendPushNotification
    - Recebe occurrence_id e lista de tokens FCM
    - Monta payload: titulo, corpo, icone, data (occurrence_id, click_action)
    - Trata erros de tokens invalidos removendo da tabela automaticamente
    - Retorna contagem de envios bem-sucedidos
  - [ ] 2.5 Implementar metodo NotifyNewOccurrence
    - Busca todos tokens de usuarios com role 'operador' ou 'gestor'
    - Formata mensagem: "ALERTA DE TRANSPLANTE" + dados do hospital e paciente
    - Envia para todos tokens de forma assincrona
    - Registra na tabela notifications com canal='push'
  - [ ] 2.6 Criar endpoint POST `/api/v1/push/subscribe`
    - Arquivo: `backend/internal/handlers/push_handler.go`
    - Autenticacao JWT obrigatoria (extrair user_id do token)
    - Body: `{ "fcm_token": "string", "device_info": { "browser": "string", "os": "string" } }`
    - Validar fcm_token nao vazio
    - Retornar 201 Created em sucesso
    - Retornar 409 Conflict se token ja existe para outro usuario
  - [ ] 2.7 Registrar rotas no router
    - Adicionar grupo `/api/v1/push` com middleware de autenticacao
    - Registrar handler de subscribe
  - [ ] 2.8 Modificar obito_listener para disparar push
    - Chamar PushService.NotifyNewOccurrence apos criar ocorrencia
    - Execucao assincrona (goroutine) para nao bloquear fluxo principal
    - Log de erros sem interromper processamento
  - [ ] 2.9 Garantir que testes da camada API passam
    - Executar APENAS os 3-5 testes escritos em 2.1
    - Verificar endpoints respondem corretamente
    - NAO executar suite completa de testes neste estagio

**Criterios de Aceitacao:**
- Os 3-5 testes escritos em 2.1 passam
- Endpoint de subscribe funciona com autenticacao JWT
- Push notifications sao enviadas via FCM
- Tokens invalidos sao removidos automaticamente
- Notificacoes sao registradas para auditoria

---

### Camada Frontend

#### Grupo de Tarefas 3: Service Worker, Hook e Componente de Banner
**Dependencias:** Grupo de Tarefas 2

- [ ] 3.0 Completar componentes frontend para push notifications
  - [ ] 3.1 Escrever 2-4 testes focados para componentes de push
    - Testar hook usePushNotifications retorna estados corretos
    - Testar PushNotificationBanner renderiza quando permissao e "default"
    - Testar clique no botao "Ativar" dispara fluxo de permissao
    - Testar banner esconde apos ativacao bem-sucedida
  - [ ] 3.2 Criar Service Worker para Firebase Messaging
    - Arquivo: `frontend/public/firebase-messaging-sw.js`
    - Configurar Firebase com credenciais do projeto
    - Implementar handler `messaging.onBackgroundMessage` para notificacoes em background
    - Implementar handler `self.addEventListener('notificationclick')` para deep link
    - Deep link: `/dashboard/occurrences/{occurrence_id}`
  - [ ] 3.3 Criar icone de notificacao
    - Arquivo: `frontend/public/icons/notification-icon.png`
    - Dimensoes: 192x192 pixels
    - Fundo branco com logo SIDOT
  - [ ] 3.4 Criar hook usePushNotifications
    - Arquivo: `frontend/src/hooks/usePushNotifications.ts`
    - Seguir padrao do useSSE.tsx existente
    - Estados expostos: permissionStatus, isSupported, fcmToken, isLoading
    - Metodo: requestPermission() que solicita permissao e registra token
    - Verificar suporte: navigator.serviceWorker && 'PushManager' in window
    - Gerenciar estados: granted, denied, default
    - Persistir flag de ativacao em localStorage
  - [ ] 3.5 Registrar Service Worker no entry point
    - Modificar `frontend/src/app/layout.tsx` ou criar provider
    - Verificar suporte do browser antes de registrar
    - Registrar `firebase-messaging-sw.js` da pasta public
  - [ ] 3.6 Criar componente PushNotificationBanner
    - Arquivo: `frontend/src/components/notifications/PushNotificationBanner.tsx`
    - Usar componente Alert do shadcn/ui
    - Texto: "Ative as notificacoes para nao perder alertas criticos"
    - Botao "Ativar" com icone Bell do lucide-react
    - Estado de loading durante ativacao
    - Feedback via toast do sonner (sucesso/erro)
    - Esconder quando permissao != "default" ou apos ativacao
  - [ ] 3.7 Integrar PushNotificationBanner no DashboardLayout
    - Arquivo: `frontend/src/components/layout/DashboardLayout.tsx`
    - Inserir banner acima do conteudo principal
    - Verificar isAuthenticated antes de mostrar (usar useAuth)
    - Renderizar condicionalmente baseado em permissionStatus
  - [ ] 3.8 Implementar chamada API para registrar token
    - Criar funcao em `frontend/src/lib/api/push.ts`
    - POST para `/api/v1/push/subscribe` com fcm_token e device_info
    - Tratar erros 409 (token ja existe)
    - Incluir JWT no header Authorization
  - [ ] 3.9 Garantir que testes de componentes passam
    - Executar APENAS os 2-4 testes escritos em 3.1
    - Verificar comportamentos criticos funcionam
    - NAO executar suite completa de testes neste estagio

**Criterios de Aceitacao:**
- Os 2-4 testes escritos em 3.1 passam
- Service Worker registra e recebe notificacoes em background
- Hook gerencia estados de permissao corretamente
- Banner aparece apenas quando permissao e "default"
- Clique na notificacao navega para ocorrencia correta

---

### Testes e Validacao

#### Grupo de Tarefas 4: Revisao de Testes e Analise de Lacunas
**Dependencias:** Grupos de Tarefas 1-3

- [ ] 4.0 Revisar testes existentes e preencher lacunas criticas
  - [ ] 4.1 Revisar testes dos Grupos de Tarefas 1-3
    - Revisar os 2-4 testes escritos para banco de dados (Tarefa 1.1)
    - Revisar os 3-5 testes escritos para API (Tarefa 2.1)
    - Revisar os 2-4 testes escritos para frontend (Tarefa 3.1)
    - Total existente: aproximadamente 7-13 testes
  - [ ] 4.2 Analisar lacunas de cobertura APENAS para esta feature
    - Identificar workflows criticos de usuario sem cobertura
    - Focar APENAS em lacunas relacionadas aos requisitos desta spec
    - NAO avaliar cobertura de toda a aplicacao
    - Priorizar fluxos end-to-end sobre testes unitarios
  - [ ] 4.3 Escrever ate 8 testes adicionais estrategicos (se necessario)
    - Teste E2E: Fluxo completo de ativacao de push (usuario clica -> token registrado)
    - Teste E2E: Nova ocorrencia dispara notificacao para operadores
    - Teste integracao: PushService + Repository funcionam juntos
    - Teste: Notificacao click leva para ocorrencia correta
    - Teste: Tratamento de browser sem suporte a Push
    - NAO escrever cobertura exaustiva para todos os cenarios
    - Pular edge cases e testes de performance
  - [ ] 4.4 Executar apenas testes relacionados a esta feature
    - Executar APENAS testes relacionados a push notifications
    - Total esperado: aproximadamente 15-21 testes
    - NAO executar suite completa de testes da aplicacao
    - Verificar que workflows criticos passam

**Criterios de Aceitacao:**
- Todos os testes especificos da feature passam (aproximadamente 15-21 testes)
- Fluxos criticos de usuario estao cobertos
- Maximo de 8 testes adicionais ao preencher lacunas
- Testes focados exclusivamente nos requisitos desta spec

---

## Ordem de Execucao

Sequencia recomendada de implementacao:

1. **Camada de Banco de Dados** (Grupo 1) - Fundacao para armazenar subscriptions
2. **Camada Backend/API** (Grupo 2) - Servico FCM e endpoints
3. **Camada Frontend** (Grupo 3) - Service Worker, hook e UI
4. **Revisao de Testes** (Grupo 4) - Validacao e cobertura

---

## Arquivos a Serem Criados/Modificados

### Novos Arquivos:
- `backend/migrations/008_create_push_subscriptions.sql`
- `backend/internal/models/push_subscription.go`
- `backend/internal/repository/push_subscription_repository.go`
- `backend/internal/config/firebase.go`
- `backend/internal/services/notification/push.go`
- `backend/internal/handlers/push_handler.go`
- `frontend/public/firebase-messaging-sw.js`
- `frontend/public/icons/notification-icon.png`
- `frontend/src/hooks/usePushNotifications.ts`
- `frontend/src/components/notifications/PushNotificationBanner.tsx`
- `frontend/src/lib/api/push.ts`

### Arquivos a Modificar:
- `backend/internal/routes/routes.go` - Adicionar rotas de push
- `backend/internal/services/obito_listener.go` - Disparar push ao criar ocorrencia
- `frontend/src/components/layout/DashboardLayout.tsx` - Integrar banner
- `frontend/src/hooks/index.ts` - Exportar novo hook
- `frontend/src/app/layout.tsx` - Registrar Service Worker

---

## Notas Tecnicas

**Firebase Cloud Messaging:**
- VAPID keys devem ser configuradas no console Firebase
- Service Account JSON armazenado como variavel de ambiente
- Tokens FCM tem validade e podem expirar

**Service Worker:**
- Deve estar na pasta public/ para escopo correto
- HTTPS obrigatorio em producao
- Testar em Chrome, Firefox, Edge (Safari tem limitacoes)

**PWA:**
- Verificar se manifest.json existe e esta configurado
- Notificacoes funcionam mesmo com app fechado (background)

**Seguranca:**
- JWT obrigatorio para registrar tokens
- Validar que token pertence ao usuario autenticado
- Nao expor credenciais Firebase no frontend
