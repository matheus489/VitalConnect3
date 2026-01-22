# Specification: Logs e Auditoria

## Goal

Implementar um sistema completo de logs de auditoria para o SIDOT, garantindo rastreabilidade de acoes de usuarios e do sistema, conformidade com LGPD e normas do CFM, com tela de consulta para Admin e Gestor, e timeline de eventos dentro de ocorrencias para Operadores.

## User Stories

- Como Admin, quero visualizar todos os logs de auditoria do sistema com filtros por data, usuario, acao e severidade, para garantir governanca e compliance
- Como Gestor, quero ver os logs de auditoria apenas do meu hospital, para auditar as acoes da minha equipe

## Specific Requirements

**Tabela audit_logs no PostgreSQL**
- Campos obrigatorios: id (UUID), timestamp, usuario_id (nullable para acoes do sistema), actor_name (nome do usuario ou "SIDOT Bot")
- Campo acao como VARCHAR descritivo (ex: "regra.update", "ocorrencia.aceitar", "auth.login")
- Campos entidade_tipo e entidade_id para identificar o registro afetado
- Campo hospital_id para filtrar por hospital (visao do Gestor)
- Enum severity com valores INFO, WARN, CRITICAL
- Campo detalhes como JSONB para dados contextuais (sem PII conforme LGPD)
- Campos ip_address e user_agent para rastreabilidade de seguranca
- Constraint de imutabilidade: tabela sem permissao de UPDATE ou DELETE

**Indices para Performance**
- Indice em timestamp DESC para ordenacao padrao
- Indices em usuario_id, entidade_tipo, entidade_id, severity, hospital_id
- Indice composto para queries frequentes (hospital_id, timestamp DESC)
- Considerar particionamento por data para volume de 5 anos

**API REST - GET /api/v1/audit-logs**
- Retorno paginado seguindo padrao PaginatedResponse existente
- Filtros via query params: data_inicio, data_fim, usuario_id, acao, entidade_tipo, entidade_id, severity, hospital_id
- Ordenacao por timestamp DESC como padrao
- Admin ve todos os logs; Gestor ve apenas logs com hospital_id do seu hospital
- Operador NAO tem acesso a este endpoint

**Servico de Auditoria Reutilizavel**
- Criar servico/package interno para registrar eventos de qualquer parte do sistema
- Assinatura: LogAuditEvent(ctx, acao, entidadeTipo, entidadeID, userID, severity, detalhes)
- Para acoes do sistema: usuario_id = nil, actor_name = "SIDOT Bot"
- Integracao com handlers existentes de Ocorrencias, Regras de Triagem e Auth

**Eventos a Rastrear com Severidade**
- auth.login (INFO), auth.logout (INFO), auth.login_failed (WARN)
- regra.create (INFO), regra.update (CRITICAL), regra.delete (CRITICAL)
- ocorrencia.visualizar (INFO), ocorrencia.aceitar (INFO), ocorrencia.recusar (INFO)
- ocorrencia.status_change (INFO), triagem.rejeicao (INFO para acoes do sistema)
- usuario.create (INFO), usuario.update (INFO), usuario.desativar (WARN)

**Frontend - Tela de Logs para Admin/Gestor**
- Pagina /logs acessivel apenas para roles admin e gestor
- Tabela com colunas: Data/Hora, Usuario, Acao, Entidade, Severidade
- Filtros: periodo (date range), usuario (select), tipo de acao (select), severidade (select)
- Paginacao usando componente Pagination existente
- Badge colorido por severidade: INFO=cinza, WARN=amarelo, CRITICAL=vermelho

**Frontend - Timeline em Ocorrencia para Operador**
- Componente OccurrenceTimeline dentro da tela de detalhes da ocorrencia
- Exibe apenas eventos relacionados aquela ocorrencia especifica
- Formato de timeline vertical cronologica
- Mostrar: horario, acao, nome do usuario (ou "Sistema"), observacoes

**Seguranca e LGPD**
- Nunca armazenar dados sensiveis de pacientes no campo detalhes
- Usar apenas IDs de referencia para entidades
- Controle de acesso rigido por role e hospital_id no backend
- Validar permissoes antes de retornar dados na API

## Visual Design

Nenhum mockup fornecido. Interface deve seguir padroes visuais existentes do Dashboard de Ocorrencias.

## Existing Code to Leverage

**OccurrencesTable e OccurrenceFilters (/frontend/src/components/dashboard/)**
- Reutilizar estrutura de tabela com Table, TableHeader, TableBody do shadcn/ui
- Seguir padrao de filtros com Select, Input type="date" e Button para limpar
- Adaptar props e tipos para contexto de AuditLog

**Pagination Component (/frontend/src/components/dashboard/Pagination.tsx)**
- Componente pronto para paginacao com navegacao e selecao de itens por pagina
- Reutilizar diretamente passando props de currentPage, totalPages, etc.

**StatusBadge Pattern (/frontend/src/components/dashboard/StatusBadge.tsx)**
- Usar mesmo padrao de mapeamento de valores para cores/variantes do Badge
- Criar SeverityBadge seguindo estrutura de getStatusConfig

**Handler Pattern (/backend/internal/handlers/occurrences.go)**
- Seguir padrao de parsing de query params para filtros
- Usar validator para input validation
- Retornar respostas no formato PaginatedResponse

**Repository Pattern (/backend/internal/repository/occurrence_history_repository.go)**
- Seguir estrutura de Create e List com contexto e tratamento de erros
- Usar sql.NullString para campos opcionais
- Pattern de conversao para Response structs

## Out of Scope

- Integracao com Loki ou outro sistema de agregacao de logs (previsto para Fase 3)
- Dashboard com graficos ou metricas de logs (apenas listagem tabular)
- Alertas de anomalia ou integracao com SIEM
- Blockchain ou certificacao digital para imutabilidade
- Exportacao de relatorios de auditoria em PDF/CSV (feature separada)
- Arquivamento automatico de logs antigos (apenas conceito de retencao)
- Escrita assincrona com fila Redis (implementar sincrono primeiro)
- Edicao ou exclusao de logs pela interface
- Logs de acesso a dados de pacientes alem do ja rastreado em ocorrencias
