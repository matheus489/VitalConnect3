# Specification: Dashboard Geografico

## Goal
Criar um mapa interativo para visualizacao em tempo real de hospitais, ocorrencias ativas e equipes de captacao, permitindo monitoramento geografico das operacoes de captacao de orgaos no estado de Goias.

## User Stories
- Como operador da central, quero visualizar no mapa a localizacao dos hospitais com ocorrencias ativas para ter uma visao espacial da situacao atual
- Como gestor, quero identificar visualmente a urgencia das ocorrencias por cor (verde/amarelo/vermelho) para priorizar o acompanhamento das mais criticas

## Specific Requirements

**Mapa Interativo com Leaflet + OpenStreetMap**
- Utilizar biblioteca Leaflet com tiles do OpenStreetMap (custo zero, open source)
- Zoom inicial deve enquadrar todos os hospitais cadastrados e ativos no estado de Goias
- Mapa responsivo que se adapta ao container do DashboardLayout
- Controles de zoom e pan nativos do Leaflet
- Nao utilizar Google Maps ou Mapbox (APIs pagas)

**Marcadores de Hospitais**
- Exibir todos os hospitais ativos cadastrados como marcadores fixos (pinos) no mapa
- Hospitais sem ocorrencias ativas: marcador em cor cinza
- Hospitais com ocorrencias ativas: marcador colorido conforme urgencia da ocorrencia mais critica
- Posicionamento baseado em coordenadas geograficas (latitude/longitude) do endereco do hospital

**Indicadores de Ocorrencias**
- Ocorrencias ativas exibidas como indicadores visuais pulsantes ou badges numericos sobre o pino do hospital
- Badge numerico quando houver multiplas ocorrencias no mesmo hospital
- Animacao pulsante para chamar atencao visual, especialmente em ocorrencias criticas
- Apenas ocorrencias com status PENDENTE ou EM_ANDAMENTO devem aparecer no mapa

**Codificacao por Cores (Urgencia)**
- Verde: tempo restante > 4 horas (janela de isquemia folgada)
- Amarelo: tempo restante entre 2 e 4 horas (atencao)
- Vermelho: tempo restante < 2 horas (critico)
- A cor do marcador do hospital reflete a ocorrencia mais urgente daquele local

**Localizacao de Equipes de Captacao**
- Mostrar equipe responsavel por cada hospital baseado na Escala de Plantao (sistema existente)
- Nao utilizar GPS real (depende do App Mobile - Item 23, fora do escopo)
- Equipe posicionada na base da Central ou no hospital de referencia conforme cadastro de plantoes
- Indicar visualmente qual operador esta de plantao para cada hospital no momento

**Drawer/Modal de Resumo ao Clicar**
- Ao clicar em um marcador de hospital: abrir Drawer ou Modal lateral
- Exibir resumo: nome do hospital, quantidade de ocorrencias ativas, operador de plantao
- Para cada ocorrencia ativa: nome mascarado, setor, tempo restante formatado, status
- Botao "Ver Detalhes" que navega para a interface de gestao existente (/dashboard/occurrences?id=...)
- Mapa como ferramenta de monitoramento apenas (nao permitir edicao de status no mapa)

**Atualizacao em Tempo Real via SSE**
- Reutilizar arquitetura SSE existente e o hook useSSE
- Atualizar marcadores automaticamente quando houver novas ocorrencias ou mudancas de status
- Reconectar automaticamente em caso de perda de conexao
- Nao utilizar polling (ineficiente e contradiz requisito de "tempo real")

**Nova Pagina no Dashboard**
- Criar rota /dashboard/map para o Dashboard Geografico
- Adicionar item no menu lateral (Sidebar) com icone de mapa
- Seguir o layout padrao do DashboardLayout existente
- Pagina deve ocupar a area principal disponivel para maximizar visualizacao do mapa

**Endpoint de API para Dados do Mapa**
- Criar endpoint GET /api/v1/map/hospitals que retorna hospitais com coordenadas e ocorrencias ativas
- Incluir dados agregados: contagem de ocorrencias, urgencia maxima, operador de plantao atual
- Resposta otimizada para renderizacao no mapa (apenas dados necessarios)
- Filtro por status de ocorrencia (apenas ativas)

## Visual Design
Nenhum arquivo visual foi fornecido.

## Existing Code to Leverage

**Hook useSSE (`/home/matheus_rubem/SIDOT/frontend/src/hooks/useSSE.tsx`)**
- Hook existente para conexao SSE com backend
- Reutilizar para receber eventos de novas ocorrencias e atualizacoes de status em tempo real
- Ja implementa reconexao automatica e gerenciamento de estado de conexao

**DashboardLayout (`/home/matheus_rubem/SIDOT/frontend/src/components/layout/DashboardLayout.tsx`)**
- Layout base do dashboard com Sidebar, Header e area de conteudo principal
- Ja integrado com useSSE para notificacoes globais
- Usar como wrapper para a pagina do mapa

**OccurrenceDetailModal (`/home/matheus_rubem/SIDOT/frontend/src/components/dashboard/OccurrenceDetailModal.tsx`)**
- Modal existente para exibir detalhes completos de uma ocorrencia
- Pode ser reutilizado ou servir de referencia para o Drawer de resumo do mapa
- Padrao de Dialog/Modal ja estabelecido com shadcn/ui

**Hooks useShifts e useTodayShifts (`/home/matheus_rubem/SIDOT/frontend/src/hooks/useShifts.ts`)**
- Hooks existentes para buscar escalas de plantao
- useTodayShifts retorna operadores escalados para o dia atual com flag is_active
- Reutilizar para identificar operador responsavel por cada hospital

**Modelo Occurrence e Hospital (Backend)**
- Modelos existentes em `/home/matheus_rubem/SIDOT/backend/internal/models/`
- Hospital ja possui campo endereco para extrair coordenadas
- Occurrence possui JanelaExpiraEm para calculo de tempo restante e urgencia

## Out of Scope
- Calculo de rotas (routing) entre pontos
- Informacoes de transito em tempo real
- Rastreamento GPS real de equipes (depende do App Mobile - Item 23)
- Heatmaps ou mapas de calor (ferramenta de analise historica - Item 16)
- Edicao de status de ocorrencias diretamente no mapa
- Acoes complexas nos marcadores (apenas visualizacao e navegacao)
- Filtros avancados no mapa (data, tipo de orgao, etc.)
- Camadas adicionais no mapa (satelite, terreno)
- Geocodificacao automatica de enderecos (coordenadas serao pre-cadastradas)
- Notificacoes sonoras especificas para o mapa (usar sistema de notificacoes global existente)
