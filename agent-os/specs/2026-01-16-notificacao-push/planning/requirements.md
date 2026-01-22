# Spec Requirements: Notificacao Push

## Initial Description

Implementar push notifications via web e/ou app mobile para alertas em tempo real mesmo com app fechado.

**Contexto do Produto:**
- Sistema SIDOT para notificacao de captacao de orgaos e tecidos
- Janela critica de 6 horas para captacao de corneas
- Equipes de plantao 24h que precisam de alertas imediatos
- Complementa canais existentes (email, SMS, dashboard)

## Requirements Discussion

### First Round Questions

**Q1:** Qual o escopo do MVP - Web Push (PWA), App Nativo ou ambos?
**Answer:** Web Push (PWA). Justificativa: Apps nativos demoram para aprovar. PWA instalado no celular/desktop funciona para o video e atende mobilidade.

**Q2:** Qual tecnologia para Push - Firebase Cloud Messaging (FCM), OneSignal, ou Web Push API nativa?
**Answer:** Firebase Cloud Messaging (FCM). Justificativa: Abstrai complexidade de subscriptions, console de testes util para desenvolvimento.

**Q3:** Quais eventos disparam notificacoes push - Nova Ocorrencia, Mudanca de Status, Tempo Critico, todos?
**Answer:** Prioridade: Nova Ocorrencia (Criacao). Secundario: Se sobrar tempo, notificar mudanca para "Concluida" (feedback), mas foco e tempo de resposta inicial.

**Q4:** Como sera solicitada a permissao do usuario - automatico no login, via botao no dashboard, ou configuracao explicita?
**Answer:** Solicitacao via Botao no Dashboard. UX: Componente de alerta no topo: "Ative as notificacoes para nao perder alertas criticos" [Botao: Ativar]. Motivo: Navegadores bloqueiam solicitacoes automaticas no login. Fluxo: Clicar -> Browser pede permissao -> Obter Token FCM -> Enviar para Backend (tabela push_subscriptions).

**Q5:** Qual o conteudo da notificacao - formato simples ou dados completos do caso?
**Answer:** Alerta Visual Forte. Titulo: "ALERTA DE TRANSPLANTE" (emoji vermelho se FCM permitir). Corpo: "Hosp: {hospital_name} | Paciente: {age} anos. Toque para iniciar protocolo." Icone: Logo SIDOT (fundo branco para contraste).

**Q6:** Service Workers serao necessarios - ja existe um ou precisamos criar?
**Answer:** Sim, obrigatorio. Implementacao: next-pwa ou firebase-messaging-sw.js manual na pasta public. Requisito tecnico inevitavel para Push.

**Q7:** O que acontece quando o usuario clica na notificacao - abre dashboard ou vai direto para a ocorrencia?
**Answer:** Deep Link (Direto para Ocorrencia). Caminho: /ocorrencias/{id_ocorrencia}. Justificativa: Reduz tempo de acao, usuario nao precisa procurar o caso.

**Q8:** Ha funcionalidades que devemos explicitamente deixar de fora do MVP?
**Answer:** Fora do escopo:
- Acoes Inline (Rich Notifications com botoes Aceitar/Recusar)
- Notificacao Silenciosa (Data-only messages) - focar em visiveis
- Agrupamento Personalizado - deixar SO agrupar por padrao

### Existing Code to Reference

No similar existing features identified for reference. This is a new capability being added to the system.

**Nota:** O sistema ja possui:
- Servico de Notificacao por email implementado (MVP completo)
- API de Ocorrencias com endpoints REST
- Dashboard de Ocorrencias em Next.js com shadcn/ui
- Sistema de autenticacao JWT

Estes componentes existentes serao a base para integracao do push.

### Follow-up Questions

Nenhuma pergunta adicional necessaria. As respostas foram completas e detalhadas.

## Visual Assets

### Files Provided:
Nenhum arquivo visual encontrado.

### Visual Insights:
Nao aplicavel - nenhum asset visual fornecido.

## Requirements Summary

### Functional Requirements

**Core - Push Notification:**
- Implementar Web Push via PWA usando Firebase Cloud Messaging (FCM)
- Criar Service Worker (firebase-messaging-sw.js ou via next-pwa) na pasta public
- Registrar tokens FCM dos dispositivos no backend (tabela push_subscriptions)
- Disparar notificacao push quando nova ocorrencia for criada
- Redirecionar para /ocorrencias/{id} ao clicar na notificacao

**UI - Solicitacao de Permissao:**
- Componente de banner/alerta no topo do dashboard
- Texto: "Ative as notificacoes para nao perder alertas criticos"
- Botao: "Ativar" que dispara fluxo de permissao do browser
- Feedback visual apos ativacao (sucesso ou erro)

**Conteudo da Notificacao:**
- Titulo: "ALERTA DE TRANSPLANTE" (com emoji vermelho se suportado)
- Corpo: "Hosp: {hospital_name} | Paciente: {age} anos. Toque para iniciar protocolo."
- Icone: Logo SIDOT com fundo branco
- Deep link: /ocorrencias/{id_ocorrencia}

**Backend:**
- Endpoint para registrar subscription/token FCM do usuario
- Integracao do servico de notificacao existente com FCM Admin SDK (Go)
- Armazenar tokens na tabela push_subscriptions vinculada ao usuario

### Reusability Opportunities

- Servico de Notificacao existente pode ser estendido para incluir canal push
- Dashboard de Ocorrencias ja implementado recebera o banner de ativacao
- Sistema de autenticacao JWT existente vincula tokens push ao usuario logado
- Estrutura de filas (Redis Streams) pode ser usada para enfileirar push notifications

### Scope Boundaries

**In Scope:**
- Web Push via PWA com Firebase Cloud Messaging
- Service Worker para receber notificacoes em background
- Banner de solicitacao de permissao no dashboard
- Notificacao ao criar nova ocorrencia (evento principal)
- Deep link para pagina da ocorrencia especifica
- Tabela push_subscriptions no banco de dados
- Integracao FCM Admin SDK no backend Go

**Out of Scope:**
- App nativo iOS/Android (deixado para Fase 3 do roadmap)
- Rich Notifications com botoes de acao inline (Aceitar/Recusar)
- Notificacoes silenciosas (data-only messages)
- Agrupamento personalizado de notificacoes
- Notificacao de mudanca de status (apenas se sobrar tempo)
- Configuracoes avancadas de preferencias de notificacao por usuario

### Technical Considerations

**Stack Definida:**
- Frontend: Next.js 14+ com React 18+, Tailwind CSS, shadcn/ui
- Backend: Go com Gin/Echo, PostgreSQL, Redis
- Push: Firebase Cloud Messaging (FCM)
- Service Worker: next-pwa ou firebase-messaging-sw.js manual

**Integracao FCM:**
- Usar FCM Admin SDK para Go no backend
- Configurar projeto Firebase com Web Push certificates (VAPID keys)
- Token FCM armazenado por usuario (pode ter multiplos dispositivos)

**Service Worker:**
- Arquivo na pasta public/ do Next.js
- Tratamento de evento push para exibir notificacao
- Tratamento de evento notificationclick para navegacao

**Banco de Dados:**
- Nova tabela: push_subscriptions
- Campos: id, user_id, fcm_token, device_info, created_at, updated_at
- Relacionamento: usuario pode ter multiplas subscriptions (dispositivos)

**PWA Considerations:**
- Manifest.json ja deve existir ou sera criado
- HTTPS obrigatorio para Service Workers
- Testar em Chrome, Firefox, Edge (Safari tem limitacoes)
