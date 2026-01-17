# Spec Requirements: Cadastro de Hospitais com Integracao ao Mapa

## Initial Description

Cadastro de Hospitais com integracao ao mapa - permitir cadastrar multiplos hospitais com localizacao via endereco ou clique no mapa manualmente, vinculando ao mapa ja existente no sistema.

## Requirements Discussion

### First Round Questions

**Q1:** Interface de cadastro - formulario em drawer lateral ou pagina dedicada?
**Answer:** Side Drawer (Sheet) - reutilizar o padrao existente do HospitalDrawer.

**Q2:** Comportamento do pin no mapa - fixo apos selecao ou arrastavel para ajuste fino?
**Answer:** Arrastavel/interativo - usuario pode ajustar posicao do pin apos colocacao inicial.

**Q3:** Servico de geocodificacao para conversao endereco -> coordenadas?
**Answer:** Nominatim (OpenStreetMap) com debounce para evitar requisicoes excessivas.

**Q4:** Metodo de entrada de localizacao - apenas endereco, apenas clique no mapa, ou hibrido?
**Answer:** Hibrido - autocomplete de endereco + clique no mapa. Usuario pode usar qualquer um dos metodos.

**Q5:** Coordenadas (latitude/longitude) sao obrigatorias ou opcionais?
**Answer:** Obrigatorias - hospital nao pode ser cadastrado sem coordenadas validas.

**Q6:** Preview do mapa no formulario - mostrar outros hospitais ou mapa limpo?
**Answer:** Mapa limpo - apenas o pin do hospital sendo cadastrado, sem poluicao visual.

**Q7:** Permissoes de acesso - quais roles podem cadastrar hospitais?
**Answer:** Apenas Admin e Gestor podem cadastrar/editar hospitais.

**Q8:** Campos adicionais alem dos existentes no modelo?
**Answer:** Adicionar campo de telefone da recepcao/plantao para contato.

**Q9:** O que esta fora do escopo desta feature?
**Answer:** CNES (codigo nacional), importacao em lote, logos/imagens dos hospitais.

### Existing Code to Reference

**Similar Features Identified:**

- Feature: Hospital CRUD Page - Path: `/frontend/src/app/dashboard/hospitals/page.tsx`
  - Atualmente read-only, exibe hospitais em grid de cards
  - Usa hook `useHospitals()` para fetch de dados
  - Usa componentes Card do shadcn/ui

- Feature: Map Container - Path: `/frontend/src/components/map/MapContainer.tsx`
  - Usa Leaflet com react-leaflet
  - Tiles do OpenStreetMap (custo zero)
  - Constante `GOIAS_BOUNDS` para posicionamento inicial
  - Dynamic import para evitar problemas de SSR
  - Ja possui callback pattern `onHospitalClick`

- Feature: Hospital Drawer - Path: `/frontend/src/components/map/HospitalDrawer.tsx`
  - Usa componente Sheet do shadcn/ui (`@/components/ui/sheet`)
  - Padrao de drawer lateral (side="right")
  - Componentes: SheetContent, SheetHeader, SheetTitle, SheetDescription
  - **Este e o padrao a seguir para o formulario de cadastro**

- Feature: Map Hooks - Hooks `useMapHospitals` e `useMap` existentes para operacoes de mapa

- Feature: Hospital Model - Campos existentes: nome, codigo, endereco, latitude, longitude, ativo

## Visual Assets

### Files Provided:
Nenhum arquivo visual fornecido.

### Visual Insights:
N/A - seguir padroes visuais existentes no sistema.

## Requirements Summary

### Functional Requirements

**Formulario de Cadastro (Drawer)**
- Abrir drawer lateral ao clicar em botao "Novo Hospital" na pagina de hospitais
- Campos do formulario:
  - Nome do hospital (texto, obrigatorio)
  - Codigo interno (texto, obrigatorio)
  - Endereco completo (texto com autocomplete, obrigatorio)
  - Telefone recepcao/plantao (texto, novo campo, opcional)
  - Latitude (numero, preenchido automaticamente ou via mapa)
  - Longitude (numero, preenchido automaticamente ou via mapa)
  - Status ativo (toggle, default true)
- Validacao: coordenadas sao obrigatorias antes de salvar

**Integracao com Mapa**
- Exibir mapa dentro do drawer para selecao de localizacao
- Mapa limpo, sem outros hospitais (apenas pin do cadastro atual)
- Pin arrastavel para ajuste fino de posicao
- Clique no mapa posiciona/reposiciona o pin
- Coordenadas atualizadas em tempo real conforme pin se move

**Geocodificacao**
- Autocomplete de endereco usando Nominatim (OpenStreetMap)
- Debounce nas requisicoes (ex: 300ms)
- Ao selecionar endereco sugerido:
  - Preencher campo de endereco
  - Atualizar coordenadas
  - Centralizar mapa e posicionar pin

**Fluxo Reverso (Mapa -> Endereco)**
- Ao clicar/arrastar pin no mapa:
  - Atualizar campos de latitude/longitude
  - Opcionalmente fazer geocodificacao reversa para sugerir endereco

**Permissoes**
- Apenas usuarios com role Admin ou Gestor podem:
  - Ver botao "Novo Hospital"
  - Acessar formulario de cadastro
  - Editar hospitais existentes

**Edicao de Hospitais Existentes**
- Ao clicar em hospital existente, abrir drawer com dados preenchidos
- Mesmo comportamento do cadastro para edicao de localizacao
- Mesmas validacoes

### Reusability Opportunities

**Componentes a Reutilizar:**
- Sheet/Drawer do shadcn/ui (padrao do HospitalDrawer)
- Componentes de formulario do shadcn/ui (Input, Button, Label, Switch)
- MapContainer base com adaptacoes (modo preview limpo)
- Hook useHospitals para CRUD

**Padroes Backend a Seguir:**
- API REST existente para hospitais
- Modelo Hospital ja possui campos necessarios (exceto telefone)
- Adicionar migration para campo telefone

**Padroes Frontend a Seguir:**
- Dynamic import do mapa (evitar SSR)
- GOIAS_BOUNDS para posicao inicial
- Leaflet markers com eventos de drag

### Scope Boundaries

**In Scope:**
- Formulario de cadastro em drawer lateral
- Mapa interativo com pin arrastavel
- Autocomplete de endereco via Nominatim
- Geocodificacao direta (endereco -> coordenadas)
- Campo de telefone recepcao/plantao
- Validacao de coordenadas obrigatorias
- Edicao de hospitais existentes
- Controle de permissoes (Admin/Gestor)

**Out of Scope:**
- Codigo CNES (Sistema Nacional de Estabelecimentos de Saude)
- Importacao em lote de hospitais (CSV, Excel)
- Upload de logos ou imagens dos hospitais
- Geocodificacao reversa automatica (coordenadas -> endereco)
- Integracao com outros servicos de geocodificacao (Google Maps, etc)

### Technical Considerations

**Frontend:**
- Usar react-leaflet com Leaflet (ja configurado)
- Dynamic import para evitar SSR issues
- Nominatim API para geocodificacao (gratuito, sem API key)
- Debounce de 300ms para requisicoes de autocomplete
- Sheet component do shadcn/ui para drawer

**Backend:**
- Migration para adicionar campo `telefone` na tabela hospitals
- Endpoint POST /api/hospitals para criacao
- Endpoint PUT /api/hospitals/:id para edicao
- Validacao de coordenadas no backend

**Bibliotecas Existentes:**
- react-leaflet (ja instalado)
- leaflet (ja instalado)
- shadcn/ui (ja configurado)
- Nao requer novas dependencias

**Validacoes:**
- Nome: obrigatorio, min 3 caracteres
- Codigo: obrigatorio, unico
- Endereco: obrigatorio
- Latitude: obrigatorio, range valido (-90 a 90)
- Longitude: obrigatorio, range valido (-180 a 180)
- Telefone: opcional, formato brasileiro
