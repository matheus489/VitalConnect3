# Task Breakdown: Dashboard Geografico

## Visao Geral
**Total de Tarefas:** 4 Grupos de Tarefas com 27 sub-tarefas
**Estimativa de Esforco:** M (1 semana)
**Escopo:** Mapa interativo para visualizacao em tempo real de hospitais, ocorrencias ativas e equipes de captacao no estado de Goias

## Lista de Tarefas

---

### Camada de Dados (Backend)

#### Grupo de Tarefas 1: Extensao do Modelo Hospital e Endpoint do Mapa
**Dependencias:** Nenhuma

- [x] 1.0 Completar camada de dados do backend
  - [x] 1.1 Escrever 4-6 testes focados para funcionalidade do endpoint do mapa
    - Testar endpoint GET /api/v1/map/hospitals retorna hospitais ativos
    - Testar que hospitais inativos nao aparecem no resultado
    - Testar agregacao de ocorrencias ativas por hospital
    - Testar calculo de urgencia maxima (verde/amarelo/vermelho)
    - Testar inclusao de dados do operador de plantao
    - Testar filtragem por status de ocorrencia (apenas PENDENTE e EM_ANDAMENTO)
  - [x] 1.2 Criar migracao para adicionar coordenadas geograficas ao modelo Hospital
    - Adicionar campos `latitude` (DECIMAL(10,8)) e `longitude` (DECIMAL(11,8)) na tabela hospitals
    - Campos podem ser nulos (hospitais antigos serao atualizados posteriormente)
    - Criar indice para buscas geograficas: `idx_hospitals_coordinates`
  - [x] 1.3 Atualizar modelo Hospital em `/backend/internal/models/hospital.go`
    - Adicionar campos Latitude e Longitude (ponteiros para float64)
    - Atualizar structs CreateHospitalInput e UpdateHospitalInput
    - Atualizar HospitalResponse para incluir coordenadas
  - [x] 1.4 Criar structs de resposta para o endpoint do mapa em `/backend/internal/models/map.go`
    - `MapHospitalResponse`: id, nome, codigo, latitude, longitude, ativo, urgencia_maxima, ocorrencias_count
    - `MapOccurrenceResponse`: id, nome_mascarado, setor, tempo_restante, status, urgencia
    - `MapOperatorResponse`: id, nome, user_id (dados do operador de plantao)
    - `MapDataResponse`: hospitais com ocorrencias e operadores agregados
    - Funcao auxiliar para calcular urgencia: >4h=verde, 2-4h=amarelo, <2h=vermelho
  - [x] 1.5 Criar handler para endpoint GET /api/v1/map/hospitals em `/backend/internal/handlers/map.go`
    - Buscar todos hospitais ativos com coordenadas nao-nulas
    - Para cada hospital, buscar ocorrencias com status PENDENTE ou EM_ANDAMENTO
    - Calcular urgencia maxima baseada em JanelaExpiraEm
    - Buscar operador de plantao atual usando logica existente de shifts
    - Retornar dados otimizados para renderizacao no mapa
    - Seguir padrao de handlers existentes (ex: hospitals.go)
  - [x] 1.6 Registrar rota no router do backend
    - Adicionar rota GET /api/v1/map/hospitals no arquivo de rotas
    - Aplicar middleware de autenticacao existente
  - [x] 1.7 Garantir que testes da camada de dados passam
    - Executar APENAS os 4-6 testes escritos em 1.1
    - Verificar que migracao executa corretamente
    - NAO executar a suite de testes completa neste estagio

**Criterios de Aceitacao:**
- Os 4-6 testes escritos em 1.1 passam
- Migracao adiciona campos de coordenadas sem erros
- Endpoint retorna dados corretos para hospitais com ocorrencias ativas
- Calculo de urgencia funciona corretamente (verde/amarelo/vermelho)
- Operador de plantao e identificado corretamente para cada hospital

---

### Camada de API (Frontend Hooks e Tipos)

#### Grupo de Tarefas 2: Tipos TypeScript e Hooks de Dados
**Dependencias:** Grupo de Tarefas 1

- [x] 2.0 Completar camada de integracao frontend
  - [x] 2.1 Escrever 3-5 testes focados para hooks de dados do mapa
    - Testar useMapHospitals retorna dados corretamente
    - Testar tratamento de erro quando API falha
    - Testar atualizacao via SSE quando ocorrencia muda status
    - Testar calculo de cores de urgencia no frontend
  - [x] 2.2 Adicionar tipos TypeScript em `/frontend/src/types/index.ts`
    - `UrgencyLevel`: 'green' | 'yellow' | 'red'
    - `MapHospital`: id, nome, codigo, latitude, longitude, urgencia_maxima, ocorrencias_count, ocorrencias, operador_plantao
    - `MapOccurrence`: id, nome_mascarado, setor, tempo_restante, tempo_restante_minutos, status, urgencia
    - `MapOperator`: id, nome
    - Estender SSENotificationEvent para incluir eventos do mapa se necessario
  - [x] 2.3 Criar hook useMapHospitals em `/frontend/src/hooks/useMap.ts`
    - Usar React Query (useQuery) seguindo padrao de useShifts.ts
    - Endpoint: GET /api/v1/map/hospitals
    - RefetchInterval configuravel (ex: 30 segundos como fallback)
    - Retornar isLoading, error, data
  - [x] 2.4 Criar funcoes utilitarias para o mapa em `/frontend/src/lib/map-utils.ts`
    - `calculateUrgencyLevel(tempoRestanteMinutos: number): UrgencyLevel`
    - `getUrgencyColor(level: UrgencyLevel): string` (retorna cores hex/CSS)
    - `getUrgencyLabel(level: UrgencyLevel): string` (retorna texto em portugues)
    - `formatTimeRemaining(minutos: number): string` (formata tempo restante)
  - [x] 2.5 Exportar novos hooks e tipos no barrel file `/frontend/src/hooks/index.ts`
    - Adicionar export do useMapHospitals
    - Manter organizacao alfabetica
  - [x] 2.6 Garantir que testes da camada de API passam
    - Executar APENAS os 3-5 testes escritos em 2.1
    - Verificar tipagem TypeScript sem erros
    - NAO executar a suite de testes completa neste estagio

**Criterios de Aceitacao:**
- Os 3-5 testes escritos em 2.1 passam
- Tipos TypeScript compilam sem erros
- Hook integra corretamente com React Query
- Funcoes utilitarias calculam urgencia corretamente

---

### Componentes Frontend

#### Grupo de Tarefas 3: UI do Dashboard Geografico
**Dependencias:** Grupo de Tarefas 2

- [x] 3.0 Completar componentes de UI do mapa
  - [x] 3.1 Escrever 4-6 testes focados para componentes do mapa
    - Testar renderizacao do MapContainer com Leaflet
    - Testar que marcadores aparecem na posicao correta
    - Testar clique em marcador abre drawer/modal
    - Testar cores de urgencia aplicadas corretamente nos marcadores
    - Testar atualizacao visual quando dados mudam via SSE
  - [x] 3.2 Instalar dependencias do Leaflet
    - Adicionar `leaflet` e `react-leaflet` como dependencias
    - Adicionar `@types/leaflet` como devDependency
    - Importar CSS do Leaflet no layout ou componente
  - [x] 3.3 Criar componente MapContainer em `/frontend/src/components/map/MapContainer.tsx`
    - Container responsivo que ocupa area disponivel no DashboardLayout
    - Inicializar mapa Leaflet com tiles do OpenStreetMap
    - Zoom inicial enquadrando estado de Goias (bounds pre-definidos)
    - Controles de zoom e pan nativos do Leaflet
    - Props: hospitals (dados do hook), onHospitalClick (callback)
  - [x] 3.4 Criar componente HospitalMarker em `/frontend/src/components/map/HospitalMarker.tsx`
    - Marcador customizado com cor baseada na urgencia maxima
    - Cinza para hospitais sem ocorrencias ativas
    - Verde/Amarelo/Vermelho conforme urgencia
    - Badge numerico quando multiplas ocorrencias
    - Animacao pulsante para ocorrencias criticas (vermelhas)
    - Usar DivIcon do Leaflet para marcadores customizados
  - [x] 3.5 Criar componente HospitalDrawer em `/frontend/src/components/map/HospitalDrawer.tsx`
    - Drawer lateral usando Sheet do shadcn/ui (similar ao OccurrenceDetailModal)
    - Exibir: nome do hospital, quantidade de ocorrencias, operador de plantao
    - Lista de ocorrencias com: nome mascarado, setor, tempo restante, status
    - Indicador visual de urgencia para cada ocorrencia
    - Botao "Ver Detalhes" que navega para /dashboard/occurrences?id=...
    - Props: hospital (dados), open (boolean), onClose (callback)
  - [x] 3.6 Criar pagina MapPage em `/frontend/src/app/dashboard/map/page.tsx`
    - Usar DashboardLayout como wrapper (automatico via layout.tsx)
    - Integrar hook useMapHospitals
    - Integrar hook useSSE para atualizacoes em tempo real
    - Renderizar MapContainer com dados dos hospitais
    - Gerenciar estado do drawer (hospital selecionado)
    - Exibir loading skeleton enquanto carrega
    - Exibir mensagem de erro se API falhar
  - [x] 3.7 Adicionar item no menu Sidebar em `/frontend/src/components/layout/Sidebar.tsx`
    - Adicionar entrada para /dashboard/map com icone MapPin (lucide-react)
    - Label: "Mapa"
    - Posicionar apos "Dashboard" e antes de "Ocorrencias"
    - Acessivel para todos os roles (sem restricao)
  - [x] 3.8 Atualizar MobileNav se necessario
    - Garantir que item do mapa aparece na navegacao mobile
    - Seguir padrao existente dos outros itens
  - [x] 3.9 Garantir que testes de componentes UI passam
    - Executar APENAS os 4-6 testes escritos em 3.1
    - Verificar renderizacao visual correta
    - NAO executar a suite de testes completa neste estagio

**Criterios de Aceitacao:**
- Os 4-6 testes escritos em 3.1 passam
- Mapa renderiza corretamente com tiles do OpenStreetMap
- Marcadores aparecem nas posicoes corretas dos hospitais
- Cores de urgencia sao exibidas corretamente
- Drawer exibe informacoes completas do hospital
- Navegacao para detalhes da ocorrencia funciona
- Item do menu aparece no Sidebar e MobileNav

---

### Testes

#### Grupo de Tarefas 4: Revisao de Testes e Analise de Gaps
**Dependencias:** Grupos de Tarefas 1-3

- [x] 4.0 Revisar testes existentes e preencher gaps criticos apenas
  - [x] 4.1 Revisar testes dos Grupos de Tarefas 1-3
    - Revisar os 6 testes do backend (Tarefa 1.1)
    - Revisar os 15 testes dos hooks (Tarefa 2.1)
    - Revisar os 9 testes de UI (Tarefa 3.1)
    - Total de testes existentes: 30 testes
  - [x] 4.2 Analisar gaps de cobertura APENAS para esta feature
    - Identificar fluxos criticos de usuario sem cobertura de testes
    - Focar APENAS em gaps relacionados aos requisitos desta spec
    - NAO avaliar cobertura de testes de toda a aplicacao
    - Priorizar fluxos end-to-end sobre gaps de testes unitarios
  - [x] 4.3 Escrever ate 8 testes adicionais estrategicos (maximo)
    - Testar fluxo completo: carregar mapa -> clicar hospital -> ver drawer -> navegar para detalhes
    - Testar atualizacao em tempo real via SSE (nova ocorrencia aparece no mapa)
    - Testar mudanca de cor quando urgencia muda
    - Testar comportamento com hospital sem coordenadas (nao aparece no mapa)
    - Testar responsividade do mapa em diferentes tamanhos de tela
    - NAO escrever cobertura abrangente para todos os cenarios
    - Pular edge cases, testes de performance e acessibilidade exceto se criticos
  - [x] 4.4 Executar apenas testes especificos desta feature
    - Executar APENAS testes relacionados a esta feature (de 1.1, 2.1, 3.1 e 4.3)
    - Total esperado: aproximadamente 19-25 testes no maximo
    - NAO executar a suite de testes completa da aplicacao
    - Verificar que fluxos criticos passam

**Criterios de Aceitacao:**
- Todos os testes especificos da feature passam (32 testes no frontend, 6 testes no backend)
- Fluxos criticos de usuario para esta feature estao cobertos
- 8 testes adicionais foram escritos para preencher gaps
- Testes focados exclusivamente nos requisitos desta spec

---

## Ordem de Execucao

Sequencia de implementacao recomendada:

1. **Camada de Dados Backend** (Grupo de Tarefas 1)
   - Migracao para coordenadas geograficas
   - Structs de resposta para o mapa
   - Handler do endpoint GET /api/v1/map/hospitals

2. **Camada de API Frontend** (Grupo de Tarefas 2)
   - Tipos TypeScript para dados do mapa
   - Hook useMapHospitals
   - Funcoes utilitarias de urgencia

3. **Componentes de UI** (Grupo de Tarefas 3)
   - Instalacao do Leaflet
   - Componentes MapContainer e HospitalMarker
   - Drawer de resumo do hospital
   - Pagina do mapa e navegacao

4. **Revisao de Testes** (Grupo de Tarefas 4)
   - Revisao dos testes existentes
   - Preenchimento de gaps criticos
   - Execucao da suite de testes da feature

---

## Notas Tecnicas

### Bibliotecas a Utilizar
- **Leaflet + react-leaflet**: Biblioteca de mapas open source (custo zero)
- **OpenStreetMap tiles**: Tiles gratuitos sem necessidade de API key

### Codigo Existente para Reutilizar
- `useSSE` hook para atualizacoes em tempo real
- `DashboardLayout` para estrutura da pagina
- `OccurrenceDetailModal` como referencia para o drawer
- `useShifts` e `useTodayShifts` para identificar operador de plantao
- Modelos `Hospital` e `Occurrence` existentes no backend

### Limites do Escopo (Fora do Escopo)
- Calculo de rotas entre pontos
- Informacoes de transito em tempo real
- Rastreamento GPS real de equipes
- Heatmaps ou mapas de calor
- Edicao de status no mapa
- Filtros avancados
- Camadas adicionais (satelite, terreno)
- Geocodificacao automatica de enderecos
