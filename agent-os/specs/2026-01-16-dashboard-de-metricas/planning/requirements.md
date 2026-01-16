# Spec Requirements: Dashboard de Metricas

## Initial Description

**Feature:** Dashboard de Metricas para VitalConnect

**Contexto:** Sistema de notificacao de doacao de orgaos com deadline de demonstracao em 26 de Janeiro (MVP/video).

**Requisitos Mencionados:**
- Tela com graficos de notificacoes por periodo
- Tempo medio de resposta
- Taxa de conversao (notificacao -> captacao)
- Ranking por hospital
- Visual insights para gestores

**Contexto do Sistema:**
- Ocorrencias possuem fluxo de status: pending -> analyzing -> completed/rejected
- Notificacoes sao enviadas via email, SMS, push, dashboard
- Hospitais sao registrados no sistema
- Usuarios possuem roles: admin, gestor, operador

## Requirements Discussion

### First Round Questions

**Q1:** Dados Dinamicos vs Mockados - O dashboard deve consumir dados reais das APIs ou trabalhar com dados mockados para a demo?
**Answer:** Hibrido (Backend Real + Seed Robusto). Dashboard consome APIs reais (GET /metrics). Script Seeder popula banco com 30 dias de historico falso mas coerente. Garante graficos nao vazios na demo + prova tecnologia (TRL 5).

**Q2:** Como calcular a Taxa de Conversao? Qual formula exata?
**Answer:** Formula Ajustada: `(Captacoes Realizadas / Notificacoes Validas Enviadas) * 100`. "Notificacoes Validas" = passaram pelo filtro de triagem. Mede eficiencia da equipe humana, nao do software.

**Q3:** Qual tipo de visualizacao e periodo para o grafico de Notificacoes por Periodo?
**Answer:** Linha do Tempo (Line Chart) - Ultimos 30 dias. Filtros: Fixar em "Ultimos 30 dias" (sem datepicker complexo). Visual: Duas linhas - "Obitos Totais" vs "Captados". Gap entre linhas visualiza oportunidade de melhoria.

**Q4:** Como medir o Tempo Medio de Resposta? Qual intervalo?
**Answer:** Ambas metricas em Cards Separados:
- Metrica A (Latencia do Sistema): Deteccao -> Notificacao (segundos) - prova inovacao tecnologica
- Metrica B (Eficiencia Operacional): Notificacao -> Aceite do Operador (minutos) - prova valor do negocio

**Q5:** Para o Ranking por Hospital, qual criterio de ordenacao?
**Answer:** Volume Absoluto de Captacoes. Visual: Lista ordenada (Top 5) com barras horizontais. Ex: "1. Hospital HGG - 15 Captacoes" (mais impactante que percentual).

**Q6:** Quais niveis de acesso e permissoes para o dashboard?
**Answer:**
- Admin/Gestor (SES): Visao Global + filtro por Hospital
- Operador: Apenas metricas do proprio hospital (auto-gestao)

**Q7:** Como sera a navegacao - parte do dashboard operacional ou separado?
**Answer:** Item Separado no Menu ("Indicadores"). Dashboard Operacional (Home) = limpo para actionable items. Dashboard de Metricas = analise/gestao.

**Q8:** O que esta fora do escopo para este MVP?
**Answer:**
- Exportacao PDF/Excel (proximo spec "Relatorios Exportaveis")
- Comparativo de Periodos (mes atual vs anterior)
- Drill-down profundo (clicar barra para ver lista pacientes)

### Existing Code to Reference

No similar existing features identified for reference by the user. However, the following existing system components are relevant:
- API de Ocorrencias (endpoints REST existentes)
- Dashboard de Ocorrencias (tela principal com filtros)
- Sistema de Autenticacao com roles (admin, gestor, operador)
- Configuracao de Hospitais (CRUD existente)

### Follow-up Questions

Nenhuma pergunta de follow-up foi necessaria. As respostas do usuario foram completas e detalhadas.

## Visual Assets

### Files Provided:
Nenhum arquivo visual encontrado na pasta `/home/matheus_rubem/VitalConnect/agent-os/specs/2026-01-16-dashboard-de-metricas/planning/visuals/`.

### Visual Insights:
Nao aplicavel - usuario confirmou que nao ha visual assets.

## Requirements Summary

### Functional Requirements

**Metricas em Cards:**
- Taxa de Conversao: `(Captacoes Realizadas / Notificacoes Validas) * 100`
- Tempo Medio de Latencia do Sistema: Deteccao -> Notificacao (em segundos)
- Tempo Medio de Resposta Operacional: Notificacao -> Aceite (em minutos)

**Grafico de Linha do Tempo:**
- Periodo fixo: Ultimos 30 dias
- Duas series: "Obitos Totais" e "Captados"
- Visualiza gap de oportunidade entre as linhas

**Ranking de Hospitais:**
- Top 5 hospitais por volume absoluto de captacoes
- Visualizacao: Barras horizontais ordenadas
- Formato: "1. Hospital HGG - 15 Captacoes"

**Sistema de Permissoes:**
- Admin/Gestor: Visao global com filtro por hospital
- Operador: Apenas dados do proprio hospital

**Navegacao:**
- Novo item de menu: "Indicadores"
- Separado do Dashboard Operacional (Home)

**Backend e Dados:**
- APIs reais (GET /metrics)
- Script Seeder com 30 dias de dados coerentes para demo

### Reusability Opportunities

- Reutilizar layout base do Dashboard de Ocorrencias (sidebar, header)
- Utilizar sistema de autenticacao e roles existente
- Consumir dados de hospitais ja cadastrados
- Seguir padrao de endpoints REST existentes

### Scope Boundaries

**In Scope:**
- Cards com metricas principais (Taxa Conversao, Tempos de Resposta)
- Grafico de linha com duas series (30 dias)
- Ranking Top 5 hospitais com barras horizontais
- Filtro por hospital para Admin/Gestor
- Restricao de dados por hospital para Operador
- Novo item de menu "Indicadores"
- Script Seeder com dados de demonstracao
- Endpoint GET /metrics

**Out of Scope:**
- Exportacao PDF/Excel (spec futura: "Relatorios Exportaveis")
- Comparativo de periodos (mes atual vs anterior)
- Drill-down para lista de pacientes
- Datepicker customizado (periodo fixo em 30 dias)

### Technical Considerations

**Frontend:**
- React 18+ com Next.js 14+
- Tailwind CSS para estilizacao
- shadcn/ui para componentes
- Biblioteca de graficos compativel (ex: Recharts, Chart.js)

**Backend:**
- Go (Golang) com Gin/Echo
- PostgreSQL para dados
- Novo endpoint: GET /api/v1/metrics

**Seguranca:**
- Autenticacao JWT existente
- Filtro de dados baseado em role e hospital_id do usuario

**Demo (26 de Janeiro):**
- Seeder deve gerar dados coerentes para 30 dias
- Graficos nao podem aparecer vazios
- Demonstrar TRL 5 (tecnologia validada)
