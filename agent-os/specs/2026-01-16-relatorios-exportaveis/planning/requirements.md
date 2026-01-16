# Spec Requirements: Relatorios Exportaveis

## Initial Description

Geracao de relatorios em PDF e CSV com dados de ocorrencias e metricas para prestacao de contas a SES (Secretaria Estadual de Saude). Funcionalidade prevista no item 17 do roadmap (Fase 2 - Produto Completo para Piloto).

## Requirements Discussion

### First Round Questions

**Q1:** Dados do Relatorio Principal - Quais colunas devem aparecer na listagem de ocorrencias?
**Answer:** Confirmado + Auditoria. Colunas: Hospital, Data/Hora Obito, Iniciais do Paciente (LGPD), Idade, Status Final, Tempo de Reacao (Minutos), Usuario Responsavel. SES precisa saber quem atendeu e quao rapido.

**Q2:** Metricas Agregadas - Alem do total de ocorrencias por status, quais metricas sao necessarias?
**Answer:** Sim + Taxa de Perda. Adicionar: "Taxa de Perda Operacional" (Notificacoes que estouraram o tempo de 6h sem acao). Essa e a metrica de dor que justifica a contratacao.

**Q3:** Filtros Disponiveis - Quais filtros o usuario pode aplicar antes de exportar?
**Answer:** Data, Hospital e Desfecho. Dropdowns simples na UI. Desfechos: "Captado", "Recusa Familiar", "Contraindicacao Medica", "Expirado".

**Q4:** Layout do PDF - Qual formato institucional deve ser seguido?
**Answer:** Institucional Padrao. Cabecalho: Logo VitalConnect a esquerda, "Governo do Estado de Goias - SES" (simulado) a direita. Corpo: Titulo, Periodo do Filtro, Tabela Zebrada. Rodape: "Gerado automaticamente por VitalConnect em {data_hora}".

**Q5:** Excel/CSV - Qual formato de exportacao e prioritario?
**Answer:** CSV (Prioridade). Justificativa: Gerar CSV em Go e trivial e performatico. Excel (.xlsx) exige bibliotecas pesadas. MVP usa botao "Exportar CSV".

**Q6:** Escopo MVP (26/01) - Confirmacao do escopo minimo?
**Answer:** Aprovado. Foco: Gerar arquivo corretamente. Design rebuscado e graficos no PDF ficam para V2.

**Q7:** Permissoes - Quais roles podem exportar relatorios?
**Answer:** Admin e Gestor apenas. Operadores nao exportam dados em massa (seguranca/evitar vazamento).

**Q8:** Fora do Escopo / Restricoes - Ha algo que explicitamente NAO deve ser incluido?
**Answer:**
- LGPD (Critico): Nome paciente NAO sai completo - apenas iniciais ou ID prontuario
- Sem Limite de Linhas: Codigo preparado para stream do CSV (nao estourar memoria)
- Volume baixo no MVP, mas arquitetura pronta para escala

### Existing Code to Reference

No similar existing features identified for reference. Este e o primeiro modulo de exportacao de relatorios do sistema.

**Componentes potencialmente reutilizaveis do MVP existente:**
- Sistema de autenticacao e autorizacao (roles Admin/Gestor)
- API de Ocorrencias existente (endpoints REST)
- Layout base do dashboard (sidebar, header, navegacao)
- Componentes de filtro (dropdowns) ja utilizados no Dashboard de Ocorrencias

### Follow-up Questions

Nenhuma pergunta de follow-up foi necessaria. As respostas do usuario foram completas e detalhadas.

## Visual Assets

### Files Provided:
Nenhum arquivo visual fornecido.

### Visual Insights:
Nao aplicavel.

## Requirements Summary

### Functional Requirements

**Exportacao CSV:**
- Gerar arquivo CSV com dados de ocorrencias filtradas
- Colunas: Hospital, Data/Hora Obito, Iniciais Paciente, Idade, Status Final, Tempo de Reacao (min), Usuario Responsavel
- Implementar streaming para evitar problemas de memoria com grandes volumes
- Encoding UTF-8 com BOM para compatibilidade com Excel brasileiro

**Exportacao PDF:**
- Cabecalho: Logo VitalConnect (esquerda) + "Governo do Estado de Goias - SES" (direita)
- Corpo: Titulo do relatorio, Periodo filtrado, Tabela zebrada com dados
- Rodape: "Gerado automaticamente por VitalConnect em {data_hora}"
- Layout simples e funcional (sem graficos no MVP)

**Metricas Agregadas no Relatorio:**
- Total de ocorrencias por desfecho
- Taxa de Perda Operacional: % de notificacoes que expiraram apos 6h sem acao
- Tempo medio de reacao

**Filtros na Interface:**
- Periodo (data inicio / data fim)
- Hospital (dropdown com lista de hospitais cadastrados)
- Desfecho (multi-select: Captado, Recusa Familiar, Contraindicacao Medica, Expirado)

**Controle de Acesso:**
- Apenas usuarios com role "admin" ou "gestor" podem acessar
- Operadores nao tem acesso a funcionalidade de exportacao

### Reusability Opportunities

- Reutilizar componentes de filtro do Dashboard de Ocorrencias
- Reutilizar sistema de autenticacao/autorizacao existente
- Utilizar a API de Ocorrencias existente como base para queries
- Layout base do dashboard (sidebar, header) ja implementado

### Scope Boundaries

**In Scope:**
- Tela de relatorios com filtros (data, hospital, desfecho)
- Exportacao CSV com streaming
- Exportacao PDF com layout institucional simples
- Metricas agregadas: totais por desfecho, taxa de perda, tempo medio
- Restricao de acesso por role (admin/gestor)
- Conformidade LGPD (apenas iniciais do paciente)

**Out of Scope:**
- Graficos e visualizacoes no PDF (V2)
- Exportacao Excel (.xlsx) - apenas CSV no MVP
- Design elaborado do PDF
- Agendamento automatico de relatorios
- Envio de relatorios por email
- Relatorios personalizaveis pelo usuario

### Technical Considerations

**Backend (Go):**
- CSV: biblioteca nativa `encoding/csv` com streaming via `io.Writer`
- PDF: biblioteca leve como `gofpdf` ou `go-pdf`
- Queries otimizadas para nao carregar todos registros em memoria
- Endpoint REST para geracao de relatorios

**Frontend (React/Next.js):**
- Tela dedicada para relatorios em `/relatorios`
- Componentes de filtro com shadcn/ui (DatePicker, Select, MultiSelect)
- Botoes "Exportar CSV" e "Exportar PDF"
- Loading state durante geracao
- Download automatico do arquivo gerado

**Seguranca:**
- Middleware de autorizacao verificando role admin/gestor
- Dados de paciente anonimizados (apenas iniciais)
- Logs de auditoria para cada exportacao realizada

**Performance:**
- Streaming do CSV para suportar grandes volumes futuros
- Paginacao/limite na query se necessario
- Timeout adequado para geracao de relatorios grandes
