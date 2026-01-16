# Task Breakdown: Dashboard de Metricas

## Visao Geral
Total de Tarefas: 5 Grupos | ~28 Sub-tarefas

## Lista de Tarefas

### Camada de Dados

#### Grupo de Tarefas 1: Script Seeder para Dados de Demonstracao
**Dependencias:** Nenhuma

- [ ] 1.0 Completar script seeder para dados de demo
  - [ ] 1.1 Escrever 2-4 testes focados para o seeder
    - Testar geracao de ocorrencias com status CONCLUIDA e desfecho sucesso_captacao
    - Testar distribuicao de dados entre hospitais existentes
    - Testar coerencia temporal dos 30 dias de historico
  - [ ] 1.2 Criar script seeder em Go
    - Localizar em: `/backend/cmd/seeder/metrics_seed.go`
    - Gerar 30 dias de historico com dados realistas
    - Criar ocorrencias com diferentes status finais (CONCLUIDA com sucesso_captacao, familia_recusou, etc)
    - Distribuir dados proporcionalmente entre hospitais cadastrados
    - Garantir campos NotificadoEm e timestamps coerentes para calculos de tempo
  - [ ] 1.3 Adicionar comando de execucao do seeder
    - Integrar ao Makefile ou CLI existente
    - Permitir execucao idempotente (limpar dados anteriores de seed)
  - [ ] 1.4 Garantir que os testes do seeder passam
    - Executar APENAS os 2-4 testes escritos em 1.1
    - Verificar que dados gerados sao coerentes
    - NAO executar a suite completa de testes nesta etapa

**Criterios de Aceite:**
- Os 2-4 testes escritos em 1.1 passam
- Seeder gera dados para 30 dias
- Graficos nao aparecem vazios na demonstracao
- Dados distribuidos entre hospitais existentes

---

### Camada de Backend/API

#### Grupo de Tarefas 2: Estruturas de Dados e Calculos de Metricas
**Dependencias:** Grupo de Tarefas 1

- [ ] 2.0 Completar estruturas de dados para metricas
  - [ ] 2.1 Escrever 2-4 testes focados para calculos de metricas
    - Testar calculo de taxa de conversao: `(Captacoes / Notificacoes Validas) * 100`
    - Testar calculo de latencia do sistema (tempo entre deteccao e notificacao)
    - Testar calculo de tempo de resposta operacional (tempo entre notificacao e aceite)
  - [ ] 2.2 Estender modelo DashboardMetrics em `/backend/internal/models/metrics.go`
    - Adicionar campo TaxaConversao (float64)
    - Adicionar campo LatenciaSistema (segundos)
    - Adicionar campo TempoRespostaOperacional (minutos)
    - Adicionar struct Series30Dias com arrays de datas e valores
    - Adicionar struct RankingHospitais com top 5
    - Reutilizar pattern MetricsResponse existente
  - [ ] 2.3 Implementar funcoes de calculo de metricas
    - Funcao CalcularTaxaConversao(hospitalID *uuid.UUID) float64
    - Funcao CalcularLatenciaSistema(hospitalID *uuid.UUID) float64
    - Funcao CalcularTempoResposta(hospitalID *uuid.UUID) float64
    - Funcao ObterSeries30Dias(hospitalID *uuid.UUID) Series30Dias
    - Funcao ObterRankingHospitais(limit int) []RankingItem
  - [ ] 2.4 Garantir que os testes de calculos passam
    - Executar APENAS os 2-4 testes escritos em 2.1
    - Verificar precisao dos calculos
    - NAO executar a suite completa de testes nesta etapa

**Criterios de Aceite:**
- Os 2-4 testes escritos em 2.1 passam
- Calculos retornam valores corretos baseados nos dados do seeder
- Estruturas de dados prontas para serializar em JSON

---

#### Grupo de Tarefas 3: Endpoint de API
**Dependencias:** Grupo de Tarefas 2

- [ ] 3.0 Completar endpoint GET /api/v1/metrics/indicators
  - [ ] 3.1 Escrever 3-5 testes focados para o endpoint
    - Testar resposta completa com todas as metricas para Admin/Gestor
    - Testar filtragem por query param hospital_id
    - Testar filtragem automatica por hospital_id para role Operador
    - Testar resposta 401 para usuario nao autenticado
  - [ ] 3.2 Criar controller de metricas
    - Localizar em: `/backend/internal/handlers/metrics_handler.go`
    - Handler GetIndicators para GET /api/v1/metrics/indicators
    - Seguir pattern de controllers existentes
  - [ ] 3.3 Implementar logica de permissoes no handler
    - Extrair user do contexto JWT
    - Se role == "operador": filtrar automaticamente por user.hospital_id
    - Se role == "admin" ou "gestor": usar hospital_id do query param (opcional)
    - Validar que hospital_id e UUID valido quando fornecido
  - [ ] 3.4 Estruturar response JSON
    - Campos: taxa_conversao, latencia_sistema, tempo_resposta_operacional
    - Campos: series_30_dias (array de objetos com data, obitos_totais, captados)
    - Campos: ranking_hospitais (array de objetos com posicao, nome, captacoes)
    - Adicionar indicadores de threshold (status: verde/amarelo/vermelho)
  - [ ] 3.5 Registrar rota no router
    - Adicionar em `/backend/internal/routes/routes.go`
    - Aplicar middleware de autenticacao JWT
    - Permitir acesso para roles: admin, gestor, operador
  - [ ] 3.6 Garantir que os testes do endpoint passam
    - Executar APENAS os 3-5 testes escritos em 3.1
    - Verificar responses corretos para cada cenario
    - NAO executar a suite completa de testes nesta etapa

**Criterios de Aceite:**
- Os 3-5 testes escritos em 3.1 passam
- Endpoint retorna todas as metricas em uma unica chamada
- Permissoes aplicadas corretamente por role
- Response JSON segue estrutura definida

---

### Camada de Frontend

#### Grupo de Tarefas 4: Componentes UI e Pagina de Indicadores
**Dependencias:** Grupo de Tarefas 3

- [ ] 4.0 Completar componentes UI do dashboard
  - [ ] 4.1 Escrever 3-5 testes focados para componentes UI
    - Testar renderizacao dos MetricCards com valores formatados
    - Testar renderizacao do grafico de linha com duas series
    - Testar renderizacao do ranking com barras horizontais
    - Testar visibilidade do filtro baseado em role
  - [ ] 4.2 Criar hook useIndicators para fetch de dados
    - Localizar em: `/frontend/src/hooks/useIndicators.ts`
    - Usar React Query seguindo pattern de useMetrics existente
    - Receber hospitalId opcional como parametro
    - Tratar loading, error e data states
  - [ ] 4.3 Criar componente IndicatorCards
    - Localizar em: `/frontend/src/components/dashboard/IndicatorCards.tsx`
    - Card "Taxa de Conversao" com valor percentual e icone TrendingUp
    - Card "Latencia do Sistema" com valor em segundos e icone Zap
    - Card "Tempo de Resposta" com valor em minutos e icone Clock
    - Indicador visual de cor (verde/amarelo/vermelho) baseado em thresholds
    - Reutilizar pattern de MetricsCards.tsx e shadcn/ui Card
  - [ ] 4.4 Criar componente TimeSeriesChart
    - Localizar em: `/frontend/src/components/dashboard/TimeSeriesChart.tsx`
    - Usar Recharts LineChart
    - Eixo X: ultimos 30 dias (formatado pt-BR)
    - Duas series: "Obitos Totais" (linha cinza) e "Captados" (linha verde)
    - Tooltip interativo com valores de cada serie
    - Legenda identificando cada linha
  - [ ] 4.5 Criar componente HospitalRanking
    - Localizar em: `/frontend/src/components/dashboard/HospitalRanking.tsx`
    - Lista ordenada Top 5 hospitais
    - Barras horizontais proporcionais ao valor
    - Formato: "1. Hospital HGG - 15 Captacoes"
    - Destacar posicao do proprio hospital para Operador
  - [ ] 4.6 Criar componente HospitalFilter
    - Localizar em: `/frontend/src/components/dashboard/HospitalFilter.tsx`
    - Select dropdown com lista de hospitais ativos
    - Opcao "Todos os Hospitais" como padrao
    - Visivel apenas para roles admin e gestor
    - Usar useAuth para verificar role do usuario
  - [ ] 4.7 Criar pagina de Indicadores
    - Localizar em: `/frontend/src/app/dashboard/indicadores/page.tsx`
    - Layout responsivo com grid
    - Compor: HospitalFilter (se permitido), IndicatorCards, TimeSeriesChart, HospitalRanking
    - Usar useAuth para obter user.hospital_id quando Operador
  - [ ] 4.8 Adicionar item "Indicadores" no Sidebar
    - Editar: `/frontend/src/components/layout/Sidebar.tsx`
    - Icone: BarChart3 (lucide-react)
    - Posicao: apos "Ocorrencias" e antes de "Hospitais"
    - Rota: `/dashboard/indicadores`
    - Roles permitidos: admin, gestor, operador
    - Seguir pattern existente do array navItems
  - [ ] 4.9 Garantir que os testes de UI passam
    - Executar APENAS os 3-5 testes escritos em 4.1
    - Verificar renderizacao correta dos componentes
    - NAO executar a suite completa de testes nesta etapa

**Criterios de Aceite:**
- Os 3-5 testes escritos em 4.1 passam
- Cards exibem metricas formatadas corretamente
- Grafico mostra duas series com tooltip interativo
- Ranking exibe Top 5 hospitais com barras
- Filtro visivel apenas para Admin/Gestor
- Item de menu aparece no Sidebar para todos os roles

---

### Testes e Validacao

#### Grupo de Tarefas 5: Revisao de Testes e Analise de Gaps
**Dependencias:** Grupos de Tarefas 1-4

- [ ] 5.0 Revisar testes existentes e preencher gaps criticos
  - [ ] 5.1 Revisar testes dos Grupos 1-4
    - Revisar os 2-4 testes do seeder (Tarefa 1.1)
    - Revisar os 2-4 testes de calculos (Tarefa 2.1)
    - Revisar os 3-5 testes do endpoint (Tarefa 3.1)
    - Revisar os 3-5 testes de UI (Tarefa 4.1)
    - Total de testes existentes: aproximadamente 10-18 testes
  - [ ] 5.2 Analisar gaps de cobertura APENAS para esta feature
    - Identificar fluxos criticos de usuario sem cobertura de testes
    - Focar APENAS em gaps relacionados ao Dashboard de Metricas
    - NAO avaliar cobertura de testes da aplicacao inteira
    - Priorizar fluxos end-to-end sobre gaps de testes unitarios
  - [ ] 5.3 Escrever ate 8 testes adicionais estrategicos (maximo)
    - Adicionar no maximo 8 testes para preencher gaps criticos
    - Focar em integracao: Frontend -> API -> Banco
    - Testar fluxo completo: usuario Operador ve apenas dados do seu hospital
    - Testar fluxo completo: usuario Gestor filtra por hospital
    - NAO escrever testes exaustivos para todos os cenarios
    - Pular testes de edge cases, performance e acessibilidade
  - [ ] 5.4 Executar testes especificos da feature
    - Executar APENAS testes relacionados ao Dashboard de Metricas
    - Total esperado: aproximadamente 18-26 testes
    - NAO executar suite completa de testes da aplicacao
    - Verificar que fluxos criticos passam

**Criterios de Aceite:**
- Todos os testes especificos da feature passam (aproximadamente 18-26 testes)
- Fluxos criticos de usuario para esta feature estao cobertos
- No maximo 8 testes adicionais foram escritos
- Testes focados exclusivamente nos requisitos do Dashboard de Metricas

---

## Ordem de Execucao

Sequencia recomendada de implementacao:

```
1. Camada de Dados (Grupo 1: Seeder)
   - Prioridade: Alta - dados sao necessarios para testar tudo
   - Estimativa: 2-3 horas

2. Backend - Estruturas de Dados (Grupo 2: Modelos e Calculos)
   - Prioridade: Alta - fundacao para a API
   - Estimativa: 2-3 horas

3. Backend - Endpoint de API (Grupo 3: Handler e Rotas)
   - Prioridade: Alta - necessario antes do frontend
   - Estimativa: 3-4 horas

4. Frontend - Componentes UI (Grupo 4: Pagina e Componentes)
   - Prioridade: Alta - entrega visual para demo
   - Estimativa: 4-6 horas

5. Testes e Validacao (Grupo 5: Revisao de Gaps)
   - Prioridade: Media - garantir qualidade
   - Estimativa: 2-3 horas
```

**Tempo Total Estimado:** 13-19 horas

---

## Arquivos a Serem Criados/Modificados

### Novos Arquivos:
- `/backend/cmd/seeder/metrics_seed.go`
- `/backend/internal/handlers/metrics_handler.go`
- `/frontend/src/hooks/useIndicators.ts`
- `/frontend/src/components/dashboard/IndicatorCards.tsx`
- `/frontend/src/components/dashboard/TimeSeriesChart.tsx`
- `/frontend/src/components/dashboard/HospitalRanking.tsx`
- `/frontend/src/components/dashboard/HospitalFilter.tsx`
- `/frontend/src/app/dashboard/indicadores/page.tsx`

### Arquivos a Modificar:
- `/backend/internal/models/metrics.go` - Estender estruturas existentes
- `/backend/internal/routes/routes.go` - Adicionar nova rota
- `/frontend/src/components/layout/Sidebar.tsx` - Adicionar item de menu

---

## Dependencias Externas

- **Recharts:** Biblioteca de graficos (verificar se ja esta instalada no frontend)
- **lucide-react:** Icones (ja utilizado no projeto)
- **shadcn/ui:** Componentes UI (ja configurado no projeto)
- **React Query:** Gerenciamento de estado async (ja utilizado no projeto)

---

## Notas Importantes

1. **Demo em 26 de Janeiro:** Priorizar entrega funcional sobre perfeicao
2. **Dados do Seeder:** Executar seeder antes de testar manualmente o dashboard
3. **Permissoes:** Testar com usuarios de diferentes roles (admin, gestor, operador)
4. **Graficos:** Garantir que Recharts esta instalado antes de iniciar Grupo 4
5. **Thresholds:** Definir valores de corte para indicadores verde/amarelo/vermelho com o time
