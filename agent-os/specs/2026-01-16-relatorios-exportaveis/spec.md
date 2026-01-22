# Specification: Relatorios Exportaveis

## Goal

Implementar geracao de relatorios em PDF e CSV com dados de ocorrencias e metricas para prestacao de contas a Secretaria Estadual de Saude (SES), permitindo que gestores e administradores exportem dados filtrados de forma segura e em conformidade com a LGPD.

## User Stories

- Como gestor, quero exportar um relatorio CSV com dados de ocorrencias filtradas para analise em planilhas e prestacao de contas a SES
- Como administrador, quero gerar um relatorio PDF institucional com metricas agregadas para apresentacao em reunioes e auditorias

## Specific Requirements

**Exportacao CSV com Streaming**
- Utilizar biblioteca nativa `encoding/csv` do Go para geracao performatica
- Implementar streaming via `io.Writer` para evitar carregar todos os dados em memoria
- Encoding UTF-8 com BOM (`\xEF\xBB\xBF`) no inicio do arquivo para compatibilidade com Excel brasileiro
- Colunas obrigatorias: Hospital, Data/Hora Obito, Iniciais Paciente, Idade, Status Final, Tempo de Reacao (min), Usuario Responsavel
- Header Content-Disposition com nome do arquivo incluindo periodo filtrado

**Exportacao PDF Institucional**
- Utilizar biblioteca `gofpdf` ou equivalente leve para geracao do PDF
- Cabecalho: Logo SIDOT (esquerda) + texto "Governo do Estado de Goias - SES" (direita)
- Corpo: Titulo "Relatorio de Ocorrencias", Periodo filtrado, Tabela zebrada com dados
- Rodape: "Gerado automaticamente por SIDOT em {data_hora}" com numero de pagina
- Tabela com mesmas colunas do CSV, formatacao zebrada para legibilidade

**Metricas Agregadas no Relatorio**
- Total de ocorrencias por desfecho (Captado, Recusa Familiar, Contraindicacao Medica, Expirado)
- Taxa de Perda Operacional: percentual de notificacoes que expiraram apos 6h sem acao (status CANCELADA por tempo)
- Tempo medio de reacao em minutos (diferenca entre created_at e primeira mudanca de status)
- Exibir metricas no topo do PDF antes da tabela de dados

**Filtros de Relatorio**
- Periodo: campos date_from e date_to no formato YYYY-MM-DD
- Hospital: dropdown com lista de hospitais do sistema (reutilizar `useHospitals` hook)
- Desfecho: multi-select com opcoes "Captado", "Recusa Familiar", "Contraindicacao Medica", "Expirado"
- Todos os filtros sao opcionais; sem filtros retorna todos os dados

**Endpoints REST para Exportacao**
- `GET /api/v1/reports/csv` - gera e retorna arquivo CSV com streaming
- `GET /api/v1/reports/pdf` - gera e retorna arquivo PDF
- Query params: `date_from`, `date_to`, `hospital_id`, `desfecho[]`
- Resposta com Content-Type apropriado e Content-Disposition attachment

**Controle de Acesso e Autorizacao**
- Middleware `RequireRole("admin", "gestor")` nos endpoints de relatorio
- Operadores (role "operador") nao podem acessar funcionalidade de exportacao
- Registrar log de auditoria para cada exportacao realizada (user_id, timestamp, filtros aplicados)

**Conformidade LGPD**
- Nome do paciente exportado apenas como iniciais (campo `nome_paciente_mascarado` ja existente)
- Nao expor dados_completos nos relatorios, apenas dados anonimizados
- Log de auditoria deve registrar quem exportou e quando

**Interface de Usuario**
- Nova pagina `/relatorios` no dashboard com layout consistente
- Componentes de filtro reutilizando padrao de `OccurrenceFilters`
- Botoes "Exportar CSV" e "Exportar PDF" com icones apropriados
- Loading state durante geracao com feedback visual
- Download automatico do arquivo apos geracao bem-sucedida

## Visual Design

Nenhum arquivo visual fornecido. Seguir padrao de design existente no dashboard.

## Existing Code to Leverage

**`/backend/internal/middleware/auth.go` - Middleware de Autorizacao**
- Funcao `RequireRole(roles ...string)` ja implementada para restringir acesso por role
- Funcao `GetUserClaims(c *gin.Context)` para obter dados do usuario autenticado
- Reutilizar para proteger endpoints de relatorio com roles admin/gestor

**`/backend/internal/repository/occurrence_repository.go` - Repository de Ocorrencias**
- Metodo `List(ctx, filters)` com filtros por status, hospital_id, date_from, date_to
- Padrao de query builder com parametros dinamicos para reutilizar
- Estrutura `OccurrenceListFilters` para modelar filtros do relatorio

**`/backend/internal/models/occurrence.go` - Modelo de Ocorrencia**
- Campo `NomePacienteMascarado` ja anonimizado para LGPD
- Metodo `ToListResponse()` para transformar dados para exportacao
- Struct `OccurrenceCompleteData` com campo `Idade` para relatorio

**`/backend/internal/models/occurrence_history.go` - Desfechos**
- Enum `OutcomeType` com valores: sucesso_captacao, familia_recusou, contraindicacao_medica, tempo_excedido
- Metodo `DisplayName()` para labels legives no relatorio
- Usar para filtro e agrupamento por desfecho

**`/frontend/src/components/dashboard/OccurrenceFilters.tsx` - Filtros de UI**
- Componente com Select para hospital e status, Input para datas
- Hook `useHospitals()` para carregar lista de hospitais
- Padrao de props `filters/onFiltersChange` para replicar na tela de relatorios

## Out of Scope

- Graficos e visualizacoes no PDF (planejado para V2)
- Exportacao Excel (.xlsx) - apenas CSV no MVP
- Design elaborado do PDF com cores e formatacao avancada
- Agendamento automatico de relatorios
- Envio de relatorios por email
- Relatorios personalizaveis pelo usuario (escolher colunas)
- Preview do relatorio antes de exportar
- Cache de relatorios gerados
- Paginacao nos relatorios (exporta todos os dados filtrados)
- Relatorios por periodo maior que 1 ano
