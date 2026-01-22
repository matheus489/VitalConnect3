# Specification: Dashboard de Metricas

## Goal
Criar um dashboard de indicadores para gestores visualizarem metricas de desempenho do sistema SIDOT, incluindo taxa de conversao, tempos de resposta e ranking de hospitais, com dados reais consumidos via API e seed de 30 dias para demonstracao.

## User Stories
- Como gestor da SES, quero visualizar metricas consolidadas de todos os hospitais para identificar oportunidades de melhoria no processo de captacao.
- Como operador de hospital, quero ver as metricas do meu hospital para acompanhar meu desempenho e comparar com a media geral.

## Specific Requirements

**Cards de Metricas Principais**
- Card "Taxa de Conversao": formula `(Captacoes Realizadas / Notificacoes Validas) * 100` com valor em percentual
- Card "Latencia do Sistema": tempo medio entre deteccao e notificacao (em segundos) - prova inovacao tecnologica
- Card "Tempo de Resposta Operacional": tempo medio entre notificacao e aceite do operador (em minutos)
- Todos os cards devem exibir icone, titulo descritivo e valor formatado
- Adicionar indicador visual (cor verde/amarelo/vermelho) baseado em thresholds definidos

**Grafico de Linha do Tempo (30 dias)**
- Grafico de linha (LineChart) com eixo X representando os ultimos 30 dias
- Duas series de dados: "Obitos Totais" e "Captados"
- O gap entre as linhas representa a oportunidade de melhoria
- Usar Recharts como biblioteca de graficos (ja compativel com shadcn/ui)
- Tooltip interativo ao passar o mouse mostrando valores de cada serie
- Legenda identificando cada linha

**Ranking de Hospitais (Top 5)**
- Lista ordenada dos 5 hospitais com maior volume absoluto de captacoes
- Visualizacao com barras horizontais proporcionais ao valor
- Formato: "1. Hospital HGG - 15 Captacoes"
- Para operadores, mostrar posicao do proprio hospital destacada

**Filtro por Hospital (Admin/Gestor)**
- Select dropdown com lista de hospitais ativos
- Opcao "Todos os Hospitais" como padrao para Admin/Gestor
- Ao selecionar hospital, todos os cards e graficos atualizam
- Filtro nao aparece para Operador (dados ja filtrados pelo hospital_id)

**Sistema de Permissoes**
- Admin/Gestor: acesso a visao global com filtro por hospital
- Operador: acesso apenas aos dados do proprio hospital (filtrado automaticamente pelo hospital_id do usuario)
- Validar permissoes tanto no frontend (UI) quanto no backend (API)

**Novo Item de Menu "Indicadores"**
- Adicionar item no Sidebar existente com icone BarChart3 (lucide-react)
- Posicionar apos "Ocorrencias" e antes de "Hospitais"
- Rota: `/dashboard/indicadores`
- Acessivel para roles: admin, gestor, operador

**Endpoint GET /api/v1/metrics/indicators**
- Retorna todas as metricas necessarias para o dashboard em uma unica chamada
- Query params: `hospital_id` (opcional, UUID)
- Response inclui: taxa_conversao, latencia_sistema, tempo_resposta, series_30_dias, ranking_hospitais
- Aplicar filtro por hospital_id do usuario para role operador automaticamente

**Script Seeder para Demo**
- Gerar 30 dias de historico com dados coerentes e realistas
- Criar ocorrencias com diferentes status finais (CONCLUIDA com sucesso_captacao, familia_recusou, etc)
- Distribuir dados entre hospitais existentes de forma proporcional
- Garantir que graficos nao aparecem vazios na demonstracao

## Visual Design
Nenhum arquivo visual foi fornecido na pasta `planning/visuals/`.

## Existing Code to Leverage

**`/backend/internal/models/metrics.go`**
- Estrutura DashboardMetrics ja existente com campos basicos
- Metodo FormatTempoMedioNotificacao para formatacao de tempo
- Pattern de MetricsResponse para resposta de API
- Extender para incluir novos campos do dashboard de indicadores

**`/frontend/src/components/dashboard/MetricsCards.tsx`**
- Componente existente que renderiza cards de metricas
- Usa shadcn/ui Card components e lucide-react icons
- Hook useMetrics com React Query para fetch de dados
- Replicar pattern para os novos cards de indicadores

**`/frontend/src/components/layout/Sidebar.tsx`**
- Sistema de navegacao com suporte a roles
- Array navItems com estrutura {href, label, icon, roles}
- Adicionar novo item "Indicadores" seguindo mesmo pattern
- Filtro automatico por role ja implementado

**`/backend/internal/models/occurrence.go`**
- Status de ocorrencia: PENDENTE, EM_ANDAMENTO, ACEITA, RECUSADA, CANCELADA, CONCLUIDA
- OutcomeType para desfechos: sucesso_captacao, familia_recusou, etc
- Usar status CONCLUIDA com desfecho sucesso_captacao para calcular captacoes
- Campo NotificadoEm para calcular tempo de notificacao

**`/frontend/src/hooks/useAuth.tsx`**
- Context de autenticacao com user.role e user.hospital_id
- Usar para filtrar dados no frontend baseado em permissao
- Pattern para verificar isAuthenticated antes de renderizar

## Out of Scope
- Exportacao para PDF ou Excel (spec futura: "Relatorios Exportaveis")
- Comparativo entre periodos (mes atual vs mes anterior)
- Drill-down clicando em barras para ver lista de pacientes
- Datepicker customizado para selecao de periodo (fixo em 30 dias)
- Graficos de pizza ou outros tipos alem de linha e barras horizontais
- Metricas em tempo real via WebSocket (usar polling com React Query)
- Dashboard mobile-first com layout diferenciado
- Alertas ou notificacoes baseados em metricas
- Cache Redis para metricas agregadas (usar query direta no PostgreSQL)
- Testes de carga ou performance do endpoint de metricas
