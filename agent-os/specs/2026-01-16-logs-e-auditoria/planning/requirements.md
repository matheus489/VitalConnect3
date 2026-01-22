# Spec Requirements: Logs e Auditoria

## Initial Description

Sistema de logs e auditoria para o SIDOT, conforme item 18 do roadmap (Fase 2): "Registro detalhado de todas as acoes do sistema e usuarios, com tela de consulta para rastreabilidade."

Este sistema e critico para:
- Conformidade com LGPD e normas do CFM
- Rastreabilidade completa de acoes em ocorrencias de captacao
- Auditoria de mudancas em regras de triagem
- Demonstracao de governanca para stakeholders (demo de 5 minutos)

## Requirements Discussion

### First Round Questions

**Q1:** Quais entidades do sistema devem ser rastreadas?
**Answer:** Foco em "Ocorrencias" e "Regras". Prioridade para demo:
- Regras de Triagem (Critico: quem mudou idade limite)
- Ocorrencias (Quem visualizou? Quem aceitou/recusou?)
- Login/Logout (Seguranca basica)
- Secundario: Criacao de Usuarios/Hospitais (pode ficar generico)

**Q2:** Qual estrutura de log utilizar e quais campos sao necessarios?
**Answer:** Estrutura aprovada com adicao de campo `severity`:
- Niveis: INFO, WARN, CRITICAL
- Exemplos: Login = INFO, Exclusao de Regra = CRITICAL
- LGPD: `entidade_id` suficiente, evitar dados sensiveis no JSON de detalhes
- Acesso limitado conforme perfil

**Q3:** Quais perfis de usuario podem acessar a tela de logs?
**Answer:**
- Admin: Visao global de todos os logs
- Gestor: Visao local (apenas seu hospital)
- Operador: NAO acessa logs gerais - ve apenas "Linha do Tempo" dentro de ocorrencia especifica

**Q4:** Quais filtros de busca sao necessarios?
**Answer:** Padrao Admin + Severidade:
- Filtros padroes (data, usuario, entidade, acao)
- Filtro por severidade para localizar rapidamente eventos criticos
- Importante para demonstracao em video

**Q5:** Qual a politica de retencao de dados de auditoria?
**Answer:** 5 Anos com imutabilidade:
- Configurar/simular como WORM (Write Once, Read Many)
- Logs nunca deletados, apenas arquivados
- Garantia de trilha de auditoria integra

**Q6:** Devemos integrar com Loki para agregacao de logs?
**Answer:** Apenas Banco de Dados (PostgreSQL) nesta fase:
- Tabela visual no Frontend e mais impactante para demo de 5 minutos
- Loki fica para infraestrutura de producao (Fase 3)

**Q7:** Como tratar eventos automaticos do sistema (ex: triagem automatica)?
**Answer:** Tabela unificada (audit_logs) com autor "System":
- usuario_id: null
- actor_name: "SIDOT Bot"
- acao: ex. "triagem.rejeicao"
- Permite timeline completa: "Bot rejeitou -> Humano forcou aprovacao"

**Q8:** O que esta fora do escopo desta feature?
**Answer:**
- Alertas de Anomalia (SIEM)
- Dashboard de Logs (graficos) - apenas lista tabular
- Blockchain - nao prometer, apenas mencionar "Trilha de Auditoria Imutavel"

### Existing Code to Reference

Nenhuma feature similar foi explicitamente indicada pelo usuario. Porem, features existentes que podem servir de referencia:

- **Dashboard de Ocorrencias** (item 9 do roadmap - concluido): Padrao de listagem com filtros
- **API de Ocorrencias** (item 6 do roadmap - concluido): Padrao de endpoints REST
- **Autenticacao e Autorizacao** (item 7 do roadmap - concluido): Sistema de roles e permissoes

### Follow-up Questions

Nenhuma pergunta de follow-up foi necessaria. As respostas do usuario foram completas e objetivas.

## Visual Assets

### Files Provided:
Nenhum arquivo visual fornecido.

### Visual Insights:
N/A - Usuario confirmou que nao ha mockups ou wireframes para esta feature.

## Requirements Summary

### Functional Requirements

**Backend - Modelo de Dados:**
- Tabela `audit_logs` com campos:
  - id, timestamp, usuario_id (nullable para sistema), actor_name
  - acao (ex: "regra.update", "ocorrencia.aceitar", "auth.login")
  - entidade_tipo (ex: "Regra", "Ocorrencia", "Usuario")
  - entidade_id (UUID/ID da entidade afetada)
  - hospital_id (para filtro por hospital do gestor)
  - severity (enum: INFO, WARN, CRITICAL)
  - detalhes (JSONB - dados contextuais, sem PII)
  - ip_address, user_agent (metadados de seguranca)
- Constraints de imutabilidade (sem UPDATE/DELETE)
- Indices para queries por data, usuario, entidade, severidade

**Backend - API:**
- GET /api/audit-logs (listagem paginada com filtros)
- Filtros: data_inicio, data_fim, usuario_id, acao, entidade_tipo, entidade_id, severity, hospital_id
- Ordenacao por timestamp DESC (mais recentes primeiro)
- Controle de acesso: Admin ve tudo, Gestor ve apenas seu hospital

**Backend - Servico de Auditoria:**
- Funcao reutilizavel para registrar eventos em qualquer parte do sistema
- Parametros: acao, entidade, usuario (ou null para sistema), severity, detalhes
- Integracao com endpoints existentes de Ocorrencias e Regras

**Frontend - Tela de Logs (Admin/Gestor):**
- Tabela com colunas: Data/Hora, Usuario, Acao, Entidade, Severidade
- Filtros: periodo, usuario, tipo de acao, severidade
- Paginacao com scroll infinito ou paginas
- Badge colorido por severidade (INFO=cinza, WARN=amarelo, CRITICAL=vermelho)
- Drill-down para ver detalhes do evento

**Frontend - Linha do Tempo em Ocorrencia (Operador):**
- Componente de timeline dentro da tela de detalhes da ocorrencia
- Mostra apenas eventos relacionados aquela ocorrencia especifica
- Formato cronologico visual (timeline vertical)

**Eventos a Rastrear:**
- Regras de Triagem: criar, atualizar, excluir (CRITICAL para alteracoes)
- Ocorrencias: visualizar, aceitar, recusar, atualizar status
- Autenticacao: login, logout, falha de login (WARN)
- Usuarios: criar, atualizar, desativar
- Hospitais: criar, atualizar

### Reusability Opportunities

- Reutilizar componentes de tabela e filtros do Dashboard de Ocorrencias
- Seguir padrao de API REST ja estabelecido
- Aproveitar sistema de permissoes por role existente
- Componente de Badge para severidade pode ser adicionado ao shadcn/ui local

### Scope Boundaries

**In Scope:**
- Tabela audit_logs no PostgreSQL
- API REST para consulta de logs com filtros
- Tela de listagem para Admin e Gestor
- Timeline de eventos dentro de Ocorrencia para Operador
- Registro automatico de eventos do sistema ("SIDOT Bot")
- Campo de severidade (INFO, WARN, CRITICAL)
- Politica de imutabilidade (WORM simulado)
- Retencao de 5 anos

**Out of Scope:**
- Integracao com Loki (Fase 3)
- Dashboard com graficos de logs
- Alertas de anomalia/SIEM
- Blockchain ou certificacao digital
- Exportacao de relatorios de auditoria (pode ser feature separada)
- Arquivamento automatico (apenas conceito, sem implementacao)

### Technical Considerations

**Banco de Dados:**
- PostgreSQL com tabela imutavel (sem triggers de UPDATE/DELETE)
- JSONB para campo de detalhes (flexibilidade sem PII)
- Indices compostos para queries frequentes
- Particao por data para performance em 5 anos de dados

**Seguranca e LGPD:**
- Nao armazenar dados sensiveis de pacientes no campo detalhes
- Apenas IDs de referencia para entidades
- Acesso controlado por role e hospital_id
- IP e user_agent para rastreabilidade de seguranca

**Performance:**
- Paginacao obrigatoria (nunca retornar todos os logs)
- Limite de registros por pagina
- Filtro de data obrigatorio ou range maximo padrao

**Arquitetura:**
- Servico de auditoria desacoplado (pode ser chamado de qualquer modulo)
- Escrita assincrona se necessario (fila Redis) para nao impactar performance
- actor_name permite identificar acoes do sistema vs usuario
