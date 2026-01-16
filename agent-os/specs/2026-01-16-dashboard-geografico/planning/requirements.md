# Requisitos da Spec: Dashboard Geografico

## Descricao Inicial
Mapa interativo mostrando hospitais, ocorrencias ativas e equipes de captacao em tempo real. Esta funcionalidade faz parte da Fase 3 (Expansao e Melhorias v2.0) do produto VitalConnect - um sistema de deteccao e notificacao para captacao de orgaos para transplante.

**Contexto do Roadmap:**
- Item #28 do roadmap
- Estimativa de esforco: M (1 semana)
- O sistema ja possui: dashboard de ocorrencias, gestao de hospitais, gestao de plantoes, sistema de notificacoes (email, SMS, push), metricas e relatorios

## Discussao de Requisitos

### Primeira Rodada de Perguntas

**Q1:** Marcadores e Visualizacao - Como devem ser exibidos hospitais e ocorrencias no mapa?
**Resposta:** Hospitais sao marcadores fixos (pinos). Ocorrencias ativas devem ser indicadores visuais pulsantes ou badges numericos sobre o pino do hospital. Isso evita poluicao visual se houver multiplas ocorrencias no mesmo local. Status do Hospital: Icone cinza (sem ocorrencia) vs. Icone colorido (com ocorrencia ativa).

**Q2:** Equipes de Captacao - Como determinar a localizacao das equipes?
**Resposta:** Assumir a localizacao baseada na Escala de Plantao (Item 14), nao em GPS real. O rastreamento GPS em tempo real depende do App Mobile (Item 23), que e uma funcionalidade futura (XL). Para esta etapa, o mapa deve mostrar "Quem e o responsavel por este hospital agora", baseando-se no cadastro de plantoes, situando a equipe na base da Central ou no hospital de referencia.

**Q3:** Biblioteca de Mapas - Qual tecnologia utilizar?
**Resposta:** Leaflet com OpenStreetMap. Motivo Estrategico (Crucial): O Edital pontua alto (nota 10 em Viabilidade Economica) para solucoes sem custos recorrentes. Usar Google Maps ou Mapbox geraria custos em dolar por visualizacao (API). O Leaflet e gratuito, leve e Open Source, alinhando-se perfeitamente a proposta de "Independencia Tecnologica" e "Custo Zero de Licenciamento".

**Q4:** Interatividade - O que acontece ao clicar em um marcador?
**Resposta:** Visualizacao com link para detalhes (Drawer/Modal lateral). Nao implementar acoes complexas (como editar status) direto no pino do mapa agora. O clique abre um resumo (Card) com botao "Ver Detalhes", que leva para a interface de gestao ja existente. Manter o mapa como ferramenta de monitoramento, nao de edicao.

**Q5:** Escopo Geografico - Qual a abrangencia do mapa?
**Resposta:** Foco estadual (Goias). O zoom inicial deve enquadrar todos os hospitais cadastrados e ativos no estado.

**Q6:** Atualizacao em Tempo Real - Como atualizar os dados no mapa?
**Resposta:** Server-Sent Events (SSE). O sistema ja usa essa arquitetura para o Dashboard de Ocorrencias principal (para garantir o "Tempo Real"). Deve-se reutilizar a mesma conexao ou mecanismo para o mapa. Polling e ineficiente e cria "delays" que contradizem o pitch de "milissegundos".

**Q7:** Filtros Especificos - Quais filtros sao necessarios?
**Resposta:** O filtro mais importante e o Tempo Restante (Janela de Isquemia). Os marcadores dos hospitais devem mudar de cor baseados na urgencia da ocorrencia mais critica ali:
- Verde: > 4 horas restantes
- Amarelo: 2 a 4 horas restantes
- Vermelho: < 2 horas (Critico)

**Q8:** O que NAO incluir (Limites)?
**Resposta:** Excluir calculo de rotas (routing) e transito em tempo real. Isso exige APIs pagas ou pesadas. O foco agora e saber onde esta a notificacao, nao como chegar la. Heatmaps tambem sao desnecessarios para a operacao em tempo real (sao ferramentas de gestao/analise historica, Item 16).

### Codigo Existente para Referencia

**Features Similares Identificadas:**

- Feature: Hook useSSE - Path: `/home/matheus_rubem/VitalConnect/frontend/src/hooks/useSSE.tsx`
  - Reutilizar a arquitetura SSE existente para atualizacoes em tempo real do mapa

- Feature: Dashboard de Ocorrencias - Path: `/home/matheus_rubem/VitalConnect/frontend/src/app/dashboard/`
  - Interface existente de gestao de ocorrencias (destino do botao "Ver Detalhes")

- Feature: Gestao de Plantoes (Item 14) - Path: backend e frontend existentes
  - Base de dados de escalas para determinar "quem esta de plantao agora" em cada hospital

- Feature: DashboardLayout - Path: `/home/matheus_rubem/VitalConnect/frontend/src/components/layout/DashboardLayout.tsx`
  - Layout base do dashboard para manter consistencia visual

### Perguntas de Follow-up
Nao foram necessarias perguntas de follow-up. As respostas fornecidas foram completas e detalhadas.

## Ativos Visuais

### Arquivos Fornecidos:
Nenhum arquivo visual foi fornecido.

### Insights Visuais:
Nao aplicavel - nenhum ativo visual disponivel.

## Resumo dos Requisitos

### Requisitos Funcionais
- Exibir mapa interativo com hospitais como marcadores fixos (pinos)
- Mostrar ocorrencias ativas como indicadores visuais pulsantes ou badges numericos sobre o pino do hospital
- Diferenciar hospitais visualmente: icone cinza (sem ocorrencia) vs. icone colorido (com ocorrencia ativa)
- Mostrar localizacao das equipes de captacao baseada na Escala de Plantao (nao GPS real)
- Indicar responsavel atual por cada hospital baseado no cadastro de plantoes
- Ao clicar em marcador: abrir Drawer/Modal lateral com resumo da ocorrencia
- Incluir botao "Ver Detalhes" que direciona para interface de gestao existente
- Zoom inicial enquadra todos os hospitais cadastrados e ativos no estado de Goias
- Atualizacao em tempo real via Server-Sent Events (SSE)
- Codificacao por cores baseada na urgencia (Janela de Isquemia):
  - Verde: > 4 horas restantes
  - Amarelo: 2 a 4 horas restantes
  - Vermelho: < 2 horas (Critico)

### Oportunidades de Reutilizacao
- Hook `useSSE` existente para conexao de eventos em tempo real
- Layout e componentes do Dashboard existente
- Dados de plantoes ja gerenciados pelo sistema (Item 14)
- Interface de detalhes de ocorrencias ja implementada

### Limites do Escopo

**Dentro do Escopo:**
- Mapa interativo com Leaflet + OpenStreetMap
- Visualizacao de hospitais e ocorrencias ativas
- Localizacao de equipes baseada em plantoes
- Indicacao visual de urgencia por cores
- Modal/Drawer de resumo com link para detalhes
- Atualizacao em tempo real via SSE
- Foco estadual (Goias)

**Fora do Escopo:**
- Calculo de rotas (routing)
- Informacoes de transito em tempo real
- Rastreamento GPS real de equipes (depende do App Mobile - Item 23)
- Heatmaps (ferramenta de analise historica - Item 16)
- Edicao de status diretamente no mapa
- Acoes complexas nos marcadores

### Consideracoes Tecnicas
- **Biblioteca de Mapas:** Leaflet com OpenStreetMap (custo zero, open source)
- **Atualizacao em Tempo Real:** SSE (reutilizar arquitetura existente)
- **Integracao:** Reutilizar hook useSSE e conexao existente
- **Navegacao:** Link para interfaces de gestao ja implementadas
- **Justificativa Estrategica:** Edital pontua alto (nota 10) para solucoes sem custos recorrentes - evitar APIs pagas como Google Maps ou Mapbox
