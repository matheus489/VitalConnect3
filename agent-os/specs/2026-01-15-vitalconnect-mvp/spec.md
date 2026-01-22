# Specification: SIDOT MVP

## Goal

Desenvolver o MVP do SIDOT, um middleware GovTech que detecta automaticamente obitos em sistemas hospitalares, aplica triagem inteligente para elegibilidade de doacao de corneas, e notifica equipes de captacao em tempo real dentro da janela critica de 6 horas.

## User Stories

- Como um Operador da Central de Transplantes, quero receber alertas visuais e sonoros imediatamente quando um obito elegivel for detectado, para que eu possa iniciar o processo de captacao dentro da janela de 6 horas
- Como um Gestor da Central, quero visualizar metricas de obitos detectados e tempo de notificacao, para que eu possa avaliar a eficiencia do sistema
- Como um Admin do sistema, quero gerenciar hospitais e usuarios com diferentes niveis de acesso, para que eu possa controlar quem opera o sistema

## Specific Requirements

**Servico Listener de Obitos**
- Implementar polling a cada 3-5 segundos na tabela de obitos simulada
- Detectar novos registros comparando timestamp do ultimo processamento
- Suportar multiplos hospitais (HGG e HUGO) com configuracoes independentes
- Processar eventos em goroutines para nao bloquear o loop principal
- Publicar eventos detectados no Redis Streams para processamento assincrono
- Implementar health check endpoint para monitoramento do listener
- Garantir idempotencia no processamento (nao reprocessar obitos ja detectados)
- Logar todas as deteccoes com timestamp e hospital de origem

**Motor de Triagem Inteligente**
- Consumir eventos do Redis Streams publicados pelo listener
- Aplicar regras de elegibilidade configuraveis armazenadas em JSONB no PostgreSQL
- Criterios base: idade maxima (ex: 80 anos), causas de morte excludentes, janela de 6 horas
- Criterios adicionais: identificacao desconhecida (indigentes = inelegivel), setor do obito (priorizar UTI/Emergencia)
- Gerar score de priorizacao baseado no setor (UTI > Emergencia > Outros)
- Criar ocorrencia automaticamente para obitos elegiveis com status PENDENTE
- Rejeitar silenciosamente obitos inelegiveis (logar motivo para auditoria)
- Permitir atualizacao de regras sem restart do servico via cache invalidation

**Sistema de Notificacoes em Tempo Real**
- Dashboard Web com Server-Sent Events (SSE) para push de novas ocorrencias
- Badge vermelho piscando (CSS animation) para indicar ocorrencias pendentes
- Alerta sonoro (Web Audio API) ativavel/desativavel pelo usuario
- Popup/toast com dados resumidos do obito ao detectar nova ocorrencia
- Email como canal secundario usando SMTP institucional ou SendGrid
- Template de email com: hospital, setor, hora do obito, tempo restante da janela
- Registrar timestamp de cada notificacao enviada para metricas

**Workflow de Ocorrencias**
- Status disponiveis: PENDENTE, EM_ANDAMENTO, ACEITA, RECUSADA, CANCELADA, CONCLUIDA
- Transicoes permitidas: PENDENTE -> EM_ANDAMENTO -> ACEITA/RECUSADA -> CONCLUIDA; qualquer status -> CANCELADA
- Registro de desfecho obrigatorio ao concluir (sucesso na captacao, familia recusou, etc)
- Historico de acoes com usuario, timestamp e observacoes
- Filtros por status, hospital e data na listagem
- Ordenacao por prioridade (score de triagem) e tempo restante da janela

**Anonimizacao LGPD**
- Nomes mascarados em todas as listagens usando regex (ex: "Joao Silva" -> "Jo** Si***")
- Nome completo visivel apenas em: tela de detalhes da ocorrencia, modal de aceitar ocorrencia
- Implementar funcao de mascaramento reutilizavel no backend e frontend
- Dados completos acessiveis apenas para usuarios autenticados com role apropriado
- Logar acesso a dados completos para auditoria

**Dashboard de Metricas**
- Card 1: Obitos Elegiveis Detectados (Hoje) - contador simples
- Card 2: Tempo Medio de Notificacao - media em segundos/minutos
- Card 3: Corneas Potenciais (Estimativa) - obitos elegiveis x 2
- Atualizacao automatica a cada 30 segundos via polling ou SSE
- Dados calculados via query agregada no PostgreSQL

**Sistema de Autenticacao**
- Login com email/senha usando JWT com refresh tokens
- Access token expira em 15 minutos, refresh token em 7 dias
- Tres perfis: Operador (opera ocorrencias), Gestor (configura regras, ve metricas), Admin (gerencia usuarios/hospitais)
- Middleware de autorizacao por role em cada endpoint
- Senhas hasheadas com bcrypt (cost factor 12)
- Rate limiting de 5 tentativas por minuto no login

**API REST de Ocorrencias**
- GET /api/v1/occurrences - listar com paginacao, filtros (status, hospital, data)
- GET /api/v1/occurrences/:id - detalhes completos (nome sem mascara)
- PATCH /api/v1/occurrences/:id/status - atualizar status com validacao de transicao
- POST /api/v1/occurrences/:id/outcome - registrar desfecho
- GET /api/v1/occurrences/:id/history - historico de acoes
- GET /api/v1/metrics/dashboard - dados dos 3 cards de metricas

**CRUD de Hospitais**
- Cadastrar HGG (Hospital Geral de Goiania) e HUGO (Hospital de Urgencias de Goias)
- Campos: nome, codigo, endereco, configuracao de conexao (simulada), ativo/inativo
- Apenas Admin pode criar/editar/desativar hospitais
- Soft delete para manter historico de ocorrencias vinculadas

**Script Seeder para Demo**
- 5 obitos com timestamps de 1-24 horas atras (historico variado)
- 1 obito programado para T+10 segundos apos inicio (captura ao vivo no video)
- Distribuir obitos entre HGG e HUGO
- Incluir casos elegiveis e inelegiveis para demonstrar triagem
- Usuarios de teste: admin@sidot.gov.br, gestor@sidot.gov.br, operador@sidot.gov.br
- Senha padrao para todos: "demo123" (apenas ambiente de desenvolvimento)

## Visual Design

Nenhum arquivo visual fornecido. Seguir diretrizes definidas:

**Tema Clean/Hospitalar**
- Cores primarias: Branco (#FFFFFF), Cinza (#F3F4F6, #6B7280), Azul Saude (#0EA5E9)
- Verde secundario para sucesso: #10B981
- Vermelho APENAS para alertas criticos e badges de notificacao (#EF4444)
- Tipografia: Inter ou system-ui para legibilidade
- Espacamento generoso para ambiente hospitalar com telas grandes
- Componentes Shadcn/UI com tema customizado

**Layout do Dashboard**
- Sidebar fixa a esquerda com navegacao principal
- Header com logo, nome do usuario logado e botao de logout
- Area principal com cards de metricas no topo
- Tabela de ocorrencias ocupando o restante da tela
- Badge de notificacao no header piscando quando ha pendentes

## Existing Code to Leverage

**Shadcn/UI Components**
- Utilizar componentes pre-construidos: Table, Card, Badge, Button, Dialog, Form, Input, Select
- Copiar componentes para o projeto seguindo padrao do shadcn/ui
- Customizar tema de cores no tailwind.config.js para paleta hospitalar
- Usar Radix UI primitives subjacentes para acessibilidade

**TanStack Query**
- Usar useQuery para fetching de ocorrencias e metricas com cache automatico
- Usar useMutation para acoes de atualizacao de status
- Configurar staleTime e refetchInterval para dados em tempo real
- Invalidar queries apos mutacoes para sincronizacao

**Redis Streams para Mensageria**
- Usar XADD para publicar eventos de obito detectado
- Usar XREADGROUP com consumer groups para processamento distribuido
- Implementar XACK para confirmar processamento
- Usar XPENDING para reprocessar mensagens nao confirmadas

**Go Gin Framework**
- Estruturar API com gin.RouterGroup para versionamento (/api/v1)
- Usar gin.Context para injecao de dependencias
- Middleware de autenticacao JWT com gin.HandlerFunc
- Middleware de CORS para permitir frontend em porta diferente

**PostgreSQL JSONB para Regras**
- Armazenar regras de triagem em coluna JSONB para flexibilidade
- Usar operadores JSONB (@>, ?, ->>) para consultas eficientes
- Indexar campos JSONB frequentemente consultados com GIN index

## Out of Scope

- Integracao real com sistemas hospitalares (PEP/EMR) - usar tabela simulada
- Notificacoes via SMS (Zenvia/Twilio) - apenas email como secundario
- Notificacoes Push (PWA/mobile) - apenas web com SSE
- Gestao de plantoes e escalas - notificar todos os operadores ativos
- Editor visual drag-and-drop de regras de triagem - editar JSON diretamente
- Relatorios exportaveis em PDF/Excel - apenas dashboard visual
- Aplicativo mobile nativo iOS/Android
- Multi-tenant para multiplas Centrais de Transplante - single tenant
- Alta disponibilidade com redundancia e failover automatico
- Integracao com Sistema Nacional de Transplantes (SNT)
