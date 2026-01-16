# Specification: Integracao Real com PEP

## Goal

Criar um agente Go standalone para conexao direta read-only com bancos de dados hospitalares reais (PostgreSQL/MySQL/Oracle), substituindo a tabela simulada `obitos_simulados` por integracao real com sistemas PEP, com mapeamento configuravel de campos via YAML para suportar diferentes esquemas de hospitais.

## User Stories

- Como administrador do sistema, quero configurar a conexao com o banco do PEP de cada hospital via arquivo YAML, para que a integracao funcione sem necessidade de alterar codigo
- Como equipe de TI hospitalar, quero que o agente rode on-premise e faca apenas conexoes de saida (outbound only), para que nao seja necessario abrir portas no firewall

## Specific Requirements

**Agente PEP (Binario Go Standalone)**
- Compilar como binario Go unico, independente do backend principal
- Rodar on-premise no servidor do hospital (nunca aceita conexoes externas)
- Suportar drivers para PostgreSQL (`lib/pq`), MySQL (`go-sql-driver/mysql`), Oracle (`godror`)
- Implementar polling periodico configuravel (default: 3s, igual ao listener atual)
- Comunicacao unidirecional: Push via HTTPS para servidor central (Outbound Only)
- Reutilizar estrutura `ObitoEvent` do listener atual para formato de eventos

**Configuracao via YAML (mapping.yaml)**
- Definir conexao com banco: host, port, database, user, password (criptografado ou via env vars)
- Mapear campos de origem para campos padrao do VitalConnect
- Suportar query SQL customizada ou definicao de tabela/view com filtros
- Incluir configuracao do endpoint central para push dos eventos
- Exemplo de estrutura: `source_table`, `fields`, `filter_column`, `poll_interval`

**Campos Obrigatorios a Captar**
- `nome_paciente`: Nome completo do paciente
- `data_obito`: Data/hora do obito (TIMESTAMP)
- `causa_mortis`: Causa da morte
- `data_nascimento` ou `idade`: Para calcular idade
- `cns` ou `cpf`: CNS (Cartao Nacional de Saude) ou CPF - crucial para Central de Transplantes
- `hospital_id_origem`: Identificador do registro no sistema origem
- Campos opcionais: `setor`, `leito`, `prontuario`

**Mascaramento de Dados Sensiveis (LGPD)**
- CPF deve ser transmitido mascarado (ex: `***.***.123-45`)
- Reutilizar funcao `models.MaskName()` ou criar equivalente para CPF
- CNS pode ser transmitido completo (identificador unico de saude)
- Logs nao devem conter dados pessoais identificaveis

**Retry com Backoff Exponencial**
- Implementar retry para falhas de conexao com PEP: 10s, 30s, 1min, 2min, 5min (cap)
- Implementar retry para falhas de push ao servidor central (mesmo padrao)
- Gerar alerta para Health Check quando offline >10min
- Manter estado de ultimo registro processado para retomar apos falha

**Container Docker postgres-pep-sim**
- Simular esquema de banco Tasy (Philips) ou MV para demo
- Criar tabelas com nomenclatura realista (ex: `TASY.TB_PACIENTE_OBITO`)
- Script de seed que insere obito automaticamente a cada 2 minutos
- Dados ficticios mas realistas (nomes brasileiros, CIDs validos, CNS formato correto)
- Expor na porta 5433 para nao conflitar com PostgreSQL principal (5432)

**Endpoint de Recepcao no Servidor Central**
- Criar endpoint `POST /api/v1/pep/eventos` para receber eventos do agente
- Autenticar via API Key (header `X-API-Key`) associada ao hospital
- Validar payload e inserir em `obitos_simulados` ou tabela dedicada
- Publicar no Redis Streams (`obitos:detectados`) como o listener atual

**Seguranca**
- Usuario do banco PEP: permissao SELECT apenas nas views/tabelas necessarias
- Agente nunca aceita conexoes de entrada (Outbound Only)
- Conexao com servidor central via HTTPS obrigatorio em producao
- API Key rotacionavel por hospital

## Visual Design

Nenhum arquivo visual fornecido para esta especificacao.

## Existing Code to Leverage

**ObitoListener (`backend/internal/services/listener/obito_listener.go`)**
- Reutilizar estrutura `ObitoEvent` para formato padrao de eventos publicados
- Reutilizar logica de polling periodico com ticker e goroutines
- Reutilizar mecanismo de idempotencia (verificar se ja processado)
- Reutilizar publicacao no Redis Streams via `XADD`
- Adaptar `ListenerStatus` para monitoramento do agente

**Hospital Model (`backend/internal/models/hospital.go`)**
- Campo `ConfigConexao` (JSONB) ja existe para armazenar configuracoes de conexao
- Estrutura `HospitalConfig` pode ser extendida para novos campos (api_key, tipo_pep)
- Reutilizar validacoes e padroes de naming

**ObitoRepository (`backend/internal/repository/obito_repository.go`)**
- Reutilizar logica de `Create()` para inserir eventos recebidos do agente
- Adaptar queries para suportar novos campos (CNS, CPF mascarado)
- Manter compatibilidade com `GetUnprocessed()` para motor de triagem

**Config loader (`backend/config/config.go`)**
- Reutilizar padroes de leitura de env vars (`getEnv`, `getDurationEnv`)
- Adaptar para ler arquivo YAML no agente standalone
- Manter consistencia com configuracoes existentes

**Docker Compose (`docker-compose.yml`)**
- Seguir padrao de servicos existentes (healthcheck, networks, volumes)
- Usar profile `dev` para container de simulacao (igual adminer/redis-commander)

## Out of Scope

- Integracao via HL7/FHIR (deixar para V2, foco atual em SQL direto)
- Qualquer operacao de escrita (Write-back) no banco do PEP
- Migracao de historico de obitos anteriores (apenas D+0 em diante)
- Interface grafica para configuracao de mapeamento (apenas YAML)
- Conexao com hospital real para demo (usar container de simulacao)
- Suporte a bancos alem de PostgreSQL/MySQL/Oracle (ex: SQL Server)
- Descoberta automatica de esquemas de banco
- Criptografia de dados em repouso no agente
- Multi-tenancy (um agente por hospital)
- Balanceamento de carga entre multiplas instancias do agente
