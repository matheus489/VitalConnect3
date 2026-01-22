# Specification: Notificacao Push via PWA

## Goal

Implementar notificacoes push via Web Push (PWA) usando Firebase Cloud Messaging para alertar equipes de plantao em tempo real sobre novas ocorrencias de transplante, mesmo com o navegador fechado, reduzindo o tempo de resposta na janela critica de 6 horas para captacao de corneas.

## User Stories

- Como operador de plantao, quero receber alertas push no meu dispositivo quando uma nova ocorrencia for criada, para poder iniciar o protocolo de captacao imediatamente mesmo sem estar com o dashboard aberto
- Como coordenador, quero que minha equipe possa ativar notificacoes de forma simples atraves de um botao no dashboard, para garantir que todos recebam alertas criticos

## Specific Requirements

**Integracao Firebase Cloud Messaging (FCM)**
- Configurar projeto Firebase com credenciais Web Push (VAPID keys) no console Firebase
- Instalar e configurar FCM Admin SDK para Go no backend (`firebase.google.com/go/v4/messaging`)
- Armazenar credenciais Firebase em variaveis de ambiente (FIREBASE_PROJECT_ID, FIREBASE_SERVICE_ACCOUNT)
- Criar servico `PushService` em `backend/internal/services/notification/push.go` seguindo padrao do `EmailService` existente
- Implementar metodo `SendPushNotification` que recebe occurrence_id e lista de tokens FCM
- Tratar erros de tokens invalidos/expirados removendo-os automaticamente da tabela

**Tabela push_subscriptions no PostgreSQL**
- Criar migration `008_create_push_subscriptions.sql` seguindo padrao existente das migrations
- Campos: id (UUID), user_id (FK para users), fcm_token (VARCHAR UNIQUE), device_info (JSONB), created_at, updated_at
- Index composto em (user_id, fcm_token) para busca eficiente
- Constraint UNIQUE em fcm_token para evitar duplicatas (mesmo dispositivo so registra uma vez)
- Relacionamento: um usuario pode ter multiplos tokens (multiplos dispositivos)

**API Endpoint para Registro de Token**
- Criar endpoint POST `/api/v1/push/subscribe` para registrar token FCM
- Body: `{ "fcm_token": "string", "device_info": { "browser": "string", "os": "string" } }`
- Autenticacao JWT obrigatoria (extrair user_id do token)
- Validar fcm_token nao vazio e formato basico (string nao vazia, tamanho maximo 255)
- Retornar 201 Created em sucesso, 409 Conflict se token ja existe para outro usuario

**Service Worker para Push**
- Criar arquivo `frontend/public/firebase-messaging-sw.js` com configuracao FCM
- Registrar Service Worker no entry point do Next.js (verificar suporte do browser)
- Implementar handler `messaging.onBackgroundMessage` para exibir notificacao quando app em background
- Implementar handler `self.addEventListener('notificationclick')` para deep link

**Componente de Banner de Ativacao**
- Criar `frontend/src/components/notifications/PushNotificationBanner.tsx` usando shadcn/ui Alert
- Exibir banner no topo do DashboardLayout quando permissao for "default" (nao solicitada ainda)
- Texto: "Ative as notificacoes para nao perder alertas criticos" com botao "Ativar"
- Ao clicar: solicitar permissao do browser -> obter token FCM -> enviar para backend
- Esconder banner permanentemente apos ativacao bem-sucedida (localStorage flag)
- Mostrar estado de loading no botao durante processo de ativacao

**Conteudo da Notificacao Push**
- Titulo: "ALERTA DE TRANSPLANTE" (emoji vermelho se suportado pelo FCM: usar unicode)
- Corpo: "Hosp: {hospital_name} | Paciente: {idade} anos. Toque para iniciar protocolo."
- Icone: Usar logo SIDOT (criar `/public/icons/notification-icon.png` 192x192 fundo branco)
- Data payload: incluir `occurrence_id` e `click_action` URL para deep link
- Deep link ao clicar: `/dashboard/occurrences/{occurrence_id}`

**Disparo de Push ao Criar Ocorrencia**
- Modificar `obito_listener.go` para chamar `PushService.NotifyNewOccurrence` apos criar ocorrencia
- Buscar todos os tokens FCM ativos de usuarios com role 'operador' ou 'gestor'
- Enviar push para todos os tokens encontrados de forma assincrona (nao bloquear fluxo principal)
- Registrar envio na tabela `notifications` com canal='push' para auditoria

**Hook usePushNotifications no Frontend**
- Criar `frontend/src/hooks/usePushNotifications.ts` para encapsular logica FCM
- Expor: `permissionStatus`, `isSupported`, `requestPermission()`, `fcmToken`
- Verificar suporte do browser (navigator.serviceWorker && 'PushManager' in window)
- Gerenciar estado da permissao (granted, denied, default)
- Armazenar token em estado apos registro bem-sucedido

## Visual Design

Nenhum asset visual fornecido. Seguir padroes visuais existentes do dashboard:
- Banner deve usar componente Alert do shadcn/ui com variant "default" ou cor de destaque
- Botao "Ativar" deve usar variante "default" do Button com icone de sino (Bell do lucide-react)
- Feedback de sucesso via toast do sonner (ja implementado no projeto)

## Existing Code to Leverage

**backend/internal/services/notification/email.go**
- Padrao de service (EmailService struct, NewEmailService constructor, IsConfigured method)
- Replicar estrutura para criar PushService com mesma organizacao
- Usar padrao de erros definidos (ErrSMTPNotConfigured -> ErrFCMNotConfigured)

**backend/internal/services/notification/email_queue.go**
- Padrao de worker com Redis para processamento assincrono (EmailQueueWorker)
- Considerar criar PushQueueWorker similar se volume de notificacoes for alto
- Reutilizar constantes de Redis keys com prefixo "sidot:push_queue"

**backend/migrations/007_create_notifications.sql**
- Padrao de migration com UP/DOWN, indexes, comments
- Canal "push" ja pode ser adicionado ao enum notification_channel existente (ALTER TYPE)
- Seguir mesmo padrao de indices para push_subscriptions

**frontend/src/components/layout/DashboardLayout.tsx**
- Local onde o PushNotificationBanner deve ser inserido (acima do main content)
- Ja possui integracao com useSSE e toast - usar mesmo padrao de notificacao visual
- Verificar isAuthenticated antes de mostrar banner (ja disponivel via useAuth)

**frontend/src/hooks/useSSE.tsx**
- Padrao de hook com estados (isConnected, enabled)
- Padrao de verificacao de suporte do browser (typeof window check)
- Padrao de localStorage para persistencia de preferencias do usuario

## Out of Scope

- App nativo iOS/Android (planejado para Fase 3 do roadmap)
- Rich Notifications com botoes de acao inline (Aceitar/Recusar na notificacao)
- Notificacoes silenciosas (data-only messages para sync em background)
- Agrupamento personalizado de notificacoes (deixar SO agrupar automaticamente)
- Notificacao de mudanca de status da ocorrencia (apenas criacao no MVP)
- Configuracoes avancadas de preferencias de notificacao por usuario
- Notificacoes agendadas ou recorrentes
- Integracao com OneSignal ou outros provedores alem do FCM
- Suporte a Safari/iOS (limitacoes conhecidas do Web Push no iOS)
- Testes A/B de conteudo de notificacao
