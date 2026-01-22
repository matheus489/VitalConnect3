# Divisao de Tarefas: Relatorios Exportaveis

## Visao Geral
Total de Tarefas: 4 grupos principais com subtarefas

## Lista de Tarefas

### Camada Backend - Servicos de Relatorio

#### Grupo de Tarefas 1: Servico de Geracao de Relatorios
**Dependencias:** Nenhuma

- [ ] 1.0 Completar servico de geracao de relatorios
  - [ ] 1.1 Escrever 2-8 testes focados para o servico de relatorios
    - Limitar a 2-8 testes altamente focados no maximo
    - Testar apenas comportamentos criticos (geracao CSV com streaming, geracao PDF, calculo de metricas)
    - Pular cobertura exaustiva de todos os cenarios e casos de borda
  - [ ] 1.2 Criar struct `ReportFilters` para parametros de filtro
    - Campos: `DateFrom`, `DateTo`, `HospitalID`, `Desfechos []string`
    - Validacao de formato de data (YYYY-MM-DD)
    - Reutilizar padrao de `OccurrenceListFilters` existente
  - [ ] 1.3 Implementar servico `ReportService` com metodos de geracao
    - Metodo `GenerateCSV(ctx, filters, writer io.Writer)` com streaming
    - Metodo `GeneratePDF(ctx, filters) ([]byte, error)`
    - Metodo `CalculateMetrics(ctx, filters) (*ReportMetrics, error)`
    - Injetar `OccurrenceRepository` como dependencia
  - [ ] 1.4 Implementar geracao CSV com streaming
    - Usar biblioteca nativa `encoding/csv` do Go
    - Escrever diretamente no `io.Writer` para evitar carregar dados em memoria
    - Adicionar BOM UTF-8 (`\xEF\xBB\xBF`) no inicio para compatibilidade Excel brasileiro
    - Colunas: Hospital, Data/Hora Obito, Iniciais Paciente, Idade, Status Final, Tempo de Reacao (min), Usuario Responsavel
  - [ ] 1.5 Implementar geracao PDF institucional
    - Usar biblioteca `gofpdf` ou `go-pdf` para geracao leve
    - Cabecalho: Logo SIDOT (esquerda) + "Governo do Estado de Goias - SES" (direita)
    - Corpo: Titulo "Relatorio de Ocorrencias", Periodo filtrado, Metricas agregadas, Tabela zebrada
    - Rodape: "Gerado automaticamente por SIDOT em {data_hora}" + numero de pagina
  - [ ] 1.6 Implementar calculo de metricas agregadas
    - Total de ocorrencias por desfecho (Captado, Recusa Familiar, Contraindicacao Medica, Expirado)
    - Taxa de Perda Operacional: % de notificacoes CANCELADAS por tempo apos 6h sem acao
    - Tempo medio de reacao em minutos (diferenca entre `created_at` e primeira mudanca de status)
  - [ ] 1.7 Implementar log de auditoria para exportacoes
    - Registrar: user_id, timestamp, tipo de relatorio (CSV/PDF), filtros aplicados
    - Conformidade LGPD: rastrear quem exportou dados e quando
  - [ ] 1.8 Garantir que testes do servico de relatorios passem
    - Executar APENAS os 2-8 testes escritos em 1.1
    - Verificar geracao CSV com streaming funciona corretamente
    - Verificar geracao PDF inclui todos elementos obrigatorios
    - NAO executar toda a suite de testes nesta etapa

**Criterios de Aceite:**
- Os 2-8 testes escritos em 1.1 passam
- CSV e gerado com streaming sem carregar todos dados em memoria
- PDF contem cabecalho, metricas, tabela zebrada e rodape
- Metricas agregadas calculadas corretamente
- Log de auditoria registra cada exportacao

---

### Camada API - Endpoints REST

#### Grupo de Tarefas 2: Endpoints de Exportacao
**Dependencias:** Grupo de Tarefas 1

- [ ] 2.0 Completar camada de API para relatorios
  - [ ] 2.1 Escrever 2-8 testes focados para endpoints de API
    - Limitar a 2-8 testes altamente focados no maximo
    - Testar apenas comportamentos criticos (autorizacao por role, download CSV, download PDF)
    - Pular testes exaustivos de todos cenarios de filtros
  - [ ] 2.2 Criar handler `ReportHandler` com endpoints de exportacao
    - `GET /api/v1/reports/csv` - gera e retorna arquivo CSV com streaming
    - `GET /api/v1/reports/pdf` - gera e retorna arquivo PDF
    - Query params: `date_from`, `date_to`, `hospital_id`, `desfecho[]`
  - [ ] 2.3 Implementar parsing e validacao de query params
    - Validar formato de datas (YYYY-MM-DD)
    - Validar `hospital_id` como UUID valido (se fornecido)
    - Validar valores de `desfecho[]` contra enum permitido
    - Retornar erro 400 com mensagem descritiva para parametros invalidos
  - [ ] 2.4 Configurar headers de resposta para download
    - Content-Type: `text/csv; charset=utf-8` para CSV
    - Content-Type: `application/pdf` para PDF
    - Content-Disposition: `attachment; filename="relatorio_YYYY-MM-DD_YYYY-MM-DD.csv"` (periodo filtrado)
  - [ ] 2.5 Implementar middleware de autorizacao nos endpoints
    - Usar `RequireRole("admin", "gestor")` existente em `/backend/internal/middleware/auth.go`
    - Operadores (role "operador") recebem erro 403 Forbidden
    - Extrair user_id de `GetUserClaims(c *gin.Context)` para log de auditoria
  - [ ] 2.6 Registrar rotas no router principal
    - Adicionar grupo `/api/v1/reports` com middleware de autenticacao
    - Aplicar middleware `RequireRole` nas rotas de CSV e PDF
  - [ ] 2.7 Garantir que testes de API passem
    - Executar APENAS os 2-8 testes escritos em 2.1
    - Verificar que admin/gestor conseguem baixar relatorios
    - Verificar que operador recebe 403
    - NAO executar toda a suite de testes nesta etapa

**Criterios de Aceite:**
- Os 2-8 testes escritos em 2.1 passam
- Endpoints retornam arquivos com headers corretos
- Autorizacao bloqueia operadores corretamente
- Parametros invalidos retornam erro 400 descritivo

---

### Camada Frontend - Interface de Usuario

#### Grupo de Tarefas 3: Pagina de Relatorios
**Dependencias:** Grupo de Tarefas 2

- [ ] 3.0 Completar interface de usuario para relatorios
  - [ ] 3.1 Escrever 2-8 testes focados para componentes de UI
    - Limitar a 2-8 testes altamente focados no maximo
    - Testar apenas comportamentos criticos (renderizacao de filtros, clique em botoes de exportacao, estados de loading)
    - Pular testes exaustivos de todas interacoes e estados
  - [ ] 3.2 Criar pagina `/relatorios` no dashboard
    - Arquivo: `/frontend/src/app/(dashboard)/relatorios/page.tsx`
    - Layout consistente com outras paginas do dashboard
    - Titulo: "Relatorios" com descricao "Exporte dados de ocorrencias para prestacao de contas"
  - [ ] 3.3 Implementar componente `ReportFilters`
    - Reutilizar padrao de props `filters/onFiltersChange` de `OccurrenceFilters.tsx`
    - DatePicker para `date_from` e `date_to` (shadcn/ui)
    - Select para Hospital usando hook `useHospitals()` existente
    - MultiSelect para Desfecho com opcoes: Captado, Recusa Familiar, Contraindicacao Medica, Expirado
  - [ ] 3.4 Implementar botoes de exportacao
    - Botao "Exportar CSV" com icone de arquivo/planilha
    - Botao "Exportar PDF" com icone de documento
    - Desabilitar botoes durante loading
    - Estilo consistente com botoes existentes no sistema
  - [ ] 3.5 Implementar hook `useReportExport` para download de arquivos
    - Funcao `exportCSV(filters)` que faz GET em `/api/v1/reports/csv`
    - Funcao `exportPDF(filters)` que faz GET em `/api/v1/reports/pdf`
    - Tratar resposta como blob e disparar download automatico
    - Gerenciar estados: `isLoading`, `error`
  - [ ] 3.6 Implementar feedback visual durante exportacao
    - Loading spinner nos botoes durante geracao
    - Toast de sucesso apos download iniciar
    - Toast de erro com mensagem descritiva em caso de falha
    - Desabilitar interacao com filtros durante loading
  - [ ] 3.7 Adicionar item de navegacao no menu do dashboard
    - Adicionar link "Relatorios" na sidebar
    - Icone apropriado (documento/grafico)
    - Visivel apenas para roles admin/gestor (verificar permissao no frontend)
  - [ ] 3.8 Garantir que testes de UI passem
    - Executar APENAS os 2-8 testes escritos em 3.1
    - Verificar renderizacao correta dos filtros
    - Verificar comportamento dos botoes de exportacao
    - NAO executar toda a suite de testes nesta etapa

**Criterios de Aceite:**
- Os 2-8 testes escritos em 3.1 passam
- Pagina de relatorios acessivel via navegacao do dashboard
- Filtros funcionam corretamente e atualizam estado
- Download de CSV e PDF funciona automaticamente
- Estados de loading e erro exibidos adequadamente

---

### Testes - Revisao e Complemento

#### Grupo de Tarefas 4: Revisao de Testes e Analise de Gaps
**Dependencias:** Grupos de Tarefas 1-3

- [ ] 4.0 Revisar testes existentes e preencher gaps criticos apenas
  - [ ] 4.1 Revisar testes dos Grupos de Tarefas 1-3
    - Revisar os 2-8 testes escritos pelo backend (Tarefa 1.1)
    - Revisar os 2-8 testes escritos pela API (Tarefa 2.1)
    - Revisar os 2-8 testes escritos pelo frontend (Tarefa 3.1)
    - Total de testes existentes: aproximadamente 6-24 testes
  - [ ] 4.2 Analisar gaps de cobertura APENAS para esta feature
    - Identificar fluxos criticos de usuario sem cobertura de teste
    - Focar APENAS em gaps relacionados aos requisitos desta spec
    - NAO avaliar cobertura de testes de toda a aplicacao
    - Priorizar fluxos end-to-end sobre gaps de testes unitarios
  - [ ] 4.3 Escrever ate 10 testes adicionais estrategicos no maximo
    - Adicionar maximo de 10 novos testes para preencher gaps criticos identificados
    - Focar em pontos de integracao e fluxos end-to-end
    - NAO escrever cobertura abrangente para todos cenarios
    - Pular casos de borda, testes de performance e acessibilidade exceto se criticos para o negocio
  - [ ] 4.4 Executar apenas testes especificos desta feature
    - Executar APENAS testes relacionados a esta spec (testes de 1.1, 2.1, 3.1 e 4.3)
    - Total esperado: aproximadamente 16-34 testes no maximo
    - NAO executar toda a suite de testes da aplicacao
    - Verificar que fluxos criticos passam

**Criterios de Aceite:**
- Todos os testes especificos da feature passam (aproximadamente 16-34 testes total)
- Fluxos criticos de usuario para esta feature estao cobertos
- No maximo 10 testes adicionais ao preencher gaps
- Testes focados exclusivamente nos requisitos desta spec

---

## Ordem de Execucao

Sequencia recomendada de implementacao:

1. **Camada Backend - Servicos** (Grupo de Tarefas 1)
   - Fundacao: servico de geracao sem dependencia de outras camadas
   - Permite testar logica de negocio isoladamente

2. **Camada API - Endpoints** (Grupo de Tarefas 2)
   - Depende do servico de relatorios estar funcional
   - Expoe funcionalidade para consumo pelo frontend

3. **Camada Frontend - Interface** (Grupo de Tarefas 3)
   - Depende dos endpoints de API estarem disponiveis
   - Integra filtros, botoes e download de arquivos

4. **Revisao de Testes** (Grupo de Tarefas 4)
   - Revisao final apos todas camadas implementadas
   - Preenche gaps criticos identificados

---

## Codigo Existente a Reutilizar

| Arquivo | O que reutilizar |
|---------|------------------|
| `/backend/internal/middleware/auth.go` | Funcoes `RequireRole()` e `GetUserClaims()` |
| `/backend/internal/repository/occurrence_repository.go` | Metodo `List(ctx, filters)` e struct `OccurrenceListFilters` |
| `/backend/internal/models/occurrence.go` | Campo `NomePacienteMascarado` e metodo `ToListResponse()` |
| `/backend/internal/models/occurrence_history.go` | Enum `OutcomeType` e metodo `DisplayName()` |
| `/frontend/src/components/dashboard/OccurrenceFilters.tsx` | Padrao de componente de filtros e hook `useHospitals()` |

---

## Consideracoes de Conformidade LGPD

- **Dados de Paciente**: Exportar apenas iniciais (campo `NomePacienteMascarado`)
- **Sem Dados Completos**: Nao expor `dados_completos` nos relatorios
- **Auditoria**: Log de cada exportacao com user_id, timestamp e filtros
- **Controle de Acesso**: Apenas admin/gestor podem exportar dados em massa
