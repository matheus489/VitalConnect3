# Specification: Cadastro de Hospitais com Integracao ao Mapa

## Goal

Permitir que usuarios Admin e Gestor cadastrem e editem hospitais com localizacao geografica via formulario em drawer lateral, utilizando entrada hibrida (autocomplete de endereco ou clique/arraste no mapa interativo) integrado ao mapa ja existente no sistema.

## User Stories

- Como um Gestor, eu quero cadastrar novos hospitais com localizacao precisa para que eles aparecam corretamente no mapa do sistema
- Como um Admin, eu quero editar hospitais existentes e ajustar sua posicao no mapa arrastando o pin para corrigir coordenadas imprecisas

## Specific Requirements

**Formulario de Cadastro em Drawer Lateral**
- Utilizar componente Sheet do shadcn/ui com side="right" (padrao do HospitalDrawer existente)
- Abrir drawer ao clicar em botao "Novo Hospital" na pagina /dashboard/hospitals
- Usar react-hook-form com zod para validacao (padrao do LoginForm)
- Campos: nome, codigo, endereco, telefone, latitude, longitude, ativo (switch)
- Coordenadas sao obrigatorias - desabilitar botao salvar ate que coordenadas estejam preenchidas
- Exibir mapa interativo dentro do drawer para selecao visual de localizacao

**Mapa Interativo no Formulario**
- Usar react-leaflet com tiles OpenStreetMap (ja configurado no projeto)
- Dynamic import do componente de mapa para evitar problemas de SSR
- Mapa limpo sem outros hospitais - apenas o pin do cadastro atual
- Utilizar GOIAS_BOUNDS de /lib/map-utils para posicao inicial
- Pin arrastavel para ajuste fino de posicao (usar evento dragend do Leaflet Marker)
- Clique no mapa posiciona ou reposiciona o pin

**Geocodificacao via Nominatim**
- Utilizar API Nominatim do OpenStreetMap (gratuito, sem API key)
- Implementar autocomplete no campo de endereco com debounce de 300ms
- Ao selecionar sugestao: preencher endereco, atualizar coordenadas, centralizar mapa e posicionar pin
- Endpoint: https://nominatim.openstreetmap.org/search?format=json&q={endereco}&limit=5

**Sincronizacao Bidirecional Mapa-Formulario**
- Ao digitar/selecionar endereco: atualizar coordenadas e mover pin no mapa
- Ao clicar/arrastar pin no mapa: atualizar campos de latitude e longitude em tempo real
- Campos de coordenadas visiveis mas readonly (apenas para visualizacao)

**Edicao de Hospitais Existentes**
- Ao clicar em hospital na listagem, abrir drawer com dados preenchidos
- Mapa inicializado na posicao atual do hospital com pin ja posicionado
- Mesmo comportamento do cadastro para ajuste de localizacao

**Controle de Permissoes**
- Botao "Novo Hospital" visivel apenas para roles Admin e Gestor
- Opcao de editar hospital visivel apenas para Admin e Gestor
- Verificar permissao no frontend via hook useAuth (verificar user.role)
- Backend ja possui endpoints protegidos - manter consistencia

**Campo Telefone (Novo)**
- Adicionar campo telefone na tabela hospitals via migration
- Campo opcional, formato brasileiro (ex: (62) 3333-4444)
- Mascara de input para formatacao automatica
- Adicionar no modelo Go, inputs e response

**Validacoes de Formulario**
- Nome: obrigatorio, minimo 3 caracteres
- Codigo: obrigatorio, unico (validar no backend), alfanumerico
- Endereco: obrigatorio
- Latitude: obrigatorio, range -90 a 90
- Longitude: obrigatorio, range -180 a 180
- Telefone: opcional, validar formato brasileiro se preenchido

## Existing Code to Leverage

**HospitalDrawer (`/frontend/src/components/map/HospitalDrawer.tsx`)**
- Padrao de drawer lateral com Sheet, SheetContent, SheetHeader, SheetTitle
- Usar mesma estrutura side="right" e className="w-full sm:max-w-md overflow-y-auto"
- Seguir padrao de props: open, onClose, e dados do hospital

**MapContainer (`/frontend/src/components/map/MapContainer.tsx`)**
- Reutilizar estrutura de LeafletMapContainer com TileLayer do OpenStreetMap
- Usar constante GOIAS_BOUNDS de /lib/map-utils para configuracao inicial
- Seguir padrao de dynamic import para evitar SSR issues
- Adaptar para modo "preview" sem outros hospitais

**LoginForm (`/frontend/src/components/forms/LoginForm.tsx`)**
- Seguir padrao de react-hook-form com zodResolver
- Usar componentes Form, FormField, FormControl, FormItem, FormLabel, FormMessage
- Seguir padrao de loading state com isLoading e Loader2 icon
- Usar toast (sonner) para feedback de sucesso/erro

**useHospitals Hook (`/frontend/src/hooks/useHospitals.ts`)**
- Estender hook para incluir mutacoes (create, update)
- Usar useMutation do @tanstack/react-query
- Invalidar queryKey ['hospitals'] apos mutacao bem-sucedida

**Backend Hospitals Handler (`/backend/internal/handlers/hospitals.go`)**
- Endpoints POST e PATCH ja existem e funcionais
- Seguir padrao de validacao com validator/v10
- Retornar HospitalResponse consistente

## Out of Scope

- Codigo CNES (Sistema Nacional de Estabelecimentos de Saude)
- Importacao em lote de hospitais via CSV ou Excel
- Upload de logos ou imagens dos hospitais
- Geocodificacao reversa automatica (coordenadas -> endereco via Nominatim)
- Integracao com outros servicos de geocodificacao (Google Maps, Mapbox)
- Busca/filtro avancado na listagem de hospitais
- Exclusao de hospitais (soft delete ja existe no backend, mas UI nao e escopo)
- Validacao de endereco contra banco de CEPs
- Historico de alteracoes de localizacao
- Notificacoes quando hospital e cadastrado/editado
