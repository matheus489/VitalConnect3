# Task Breakdown: Integracao Real com PEP

## Visao Geral
**Total de Tarefas:** 5 grupos, 28 sub-tarefas

**Objetivo:** Criar agente Go standalone para conexao direta read-only com bancos hospitalares reais, com mapeamento configuravel via YAML e container de simulacao para demo.

**Stack:** Go (agente/backend), PostgreSQL, Docker, YAML, Redis Streams

## Lista de Tarefas

---

### Infraestrutura Docker

#### Grupo 1: Container de Simulacao PEP
**Dependencias:** Nenhuma

- [ ] 1.0 Completar container postgres-pep-sim
  - [ ] 1.1 Criar Dockerfile para postgres-pep-sim
    - Base: `postgres:15-alpine`
    - Copiar scripts de inicializacao
    - Expor porta 5433 (evitar conflito com PostgreSQL principal)
  - [ ] 1.2 Criar script SQL de schema simulando Tasy/MV
    - Tabela `TASY.TB_PACIENTE_OBITO` com nomenclatura realista
    - Campos: `CD_PACIENTE`, `NM_PACIENTE`, `DT_OBITO`, `DS_CAUSA_MORTIS`, `DT_NASCIMENTO`, `NR_CNS`, `NR_CPF`, `CD_SETOR`, `NR_LEITO`, `NR_PRONTUARIO`
    - Indexes para campos de consulta frequente
  - [ ] 1.3 Criar script de seed com dados ficticios brasileiros
    - Nomes brasileiros realistas
    - CIDs validos (ex: I21.0, J18.9)
    - CNS em formato correto (15 digitos)
    - CPF validos (com digitos verificadores)
  - [ ] 1.4 Criar script de auto-insercao de obitos
    - Inserir novo obito a cada 2 minutos via pg_cron ou script externo
    - Variar dados para demonstracao realista
  - [ ] 1.5 Adicionar servico ao docker-compose.yml
    - Profile `dev` (igual adminer/redis-commander)
    - Porta 5433:5432
    - Healthcheck configurado
    - Volume para persistencia
    - Rede `sidot-network`

**Criterios de Aceite:**
- Container inicia sem erros
- Schema Tasy simulado criado corretamente
- Dados seed inseridos na inicializacao
- Novo obito inserido automaticamente a cada 2 minutos
- Conexao possivel via porta 5433

---

### Configuracao YAML

#### Grupo 2: Sistema de Mapeamento Configuravel
**Dependencias:** Grupo 1

- [ ] 2.0 Completar sistema de configuracao YAML
  - [ ] 2.1 Definir estrutura do arquivo `mapping.yaml`
    - Conexao: `driver`, `host`, `port`, `database`, `user`, `password` (via env var)
    - Mapeamento: `source_table`, `fields` (de-para)
    - Filtros: `filter_column`, `filter_value`
    - Polling: `poll_interval` (default: 3s)
    - Endpoint central: `central_url`, `api_key` (via env var)
  - [ ] 2.2 Criar parser YAML em Go
    - Package: `pep-agent/internal/config`
    - Structs: `AgentConfig`, `DatabaseConfig`, `FieldMapping`, `CentralConfig`
    - Validacao de campos obrigatorios
    - Suporte a env vars com sintaxe `${VAR_NAME}`
  - [ ] 2.3 Criar arquivo de exemplo `mapping.example.yaml`
    - Documentar todos os campos
    - Exemplo para schema Tasy simulado
    - Comentarios explicativos em portugues

**Criterios de Aceite:**
- Estrutura YAML clara e documentada
- Parser le e valida configuracao corretamente
- Suporte a substituicao de env vars para credenciais
- Arquivo de exemplo completo e funcional

---

### Agente PEP (Go Standalone)

#### Grupo 3: Binario Go do Agente
**Dependencias:** Grupos 1 e 2

- [ ] 3.0 Completar agente PEP standalone
  - [ ] 3.1 Escrever 4-6 testes focados para funcionalidades criticas do agente
    - Teste de conexao com banco (mock)
    - Teste de mapeamento de campos
    - Teste de retry com backoff
    - Teste de mascaramento de CPF
  - [ ] 3.2 Criar estrutura do projeto `pep-agent/`
    - `cmd/pep-agent/main.go` - entrypoint
    - `internal/config/` - parser YAML
    - `internal/database/` - conexoes SQL
    - `internal/poller/` - logica de polling
    - `internal/pusher/` - envio para servidor central
    - `internal/models/` - structs de eventos
  - [ ] 3.3 Implementar conexao multi-driver
    - Suporte a PostgreSQL (`lib/pq`)
    - Suporte a MySQL (`go-sql-driver/mysql`)
    - Suporte a Oracle (`godror`)
    - Factory pattern para selecao de driver via config
  - [ ] 3.4 Implementar polling periodico
    - Ticker configuravel (default: 3s, igual listener atual)
    - Query dinamica baseada em `source_table` e `fields`
    - Controle de ultimo registro processado (watermark)
    - Reutilizar padroes do `obito_listener.go`
  - [ ] 3.5 Implementar mapeamento dinamico de campos
    - Mapear campos do PEP para estrutura `ObitoEvent` padrao
    - Campos obrigatorios: `nome_paciente`, `data_obito`, `causa_mortis`, `data_nascimento`/`idade`, `cns`/`cpf`
    - Campos opcionais: `setor`, `leito`, `prontuario`
  - [ ] 3.6 Implementar mascaramento LGPD
    - Funcao `MaskCPF()`: `***.***.123-45`
    - CNS transmitido completo (identificador de saude)
    - Logs sem dados pessoais identificaveis
  - [ ] 3.7 Implementar push HTTPS para servidor central
    - POST para endpoint configurado
    - Header `X-API-Key` para autenticacao
    - Payload JSON com estrutura `ObitoEvent`
    - TLS obrigatorio em producao (flag `--insecure` para dev)
  - [ ] 3.8 Implementar retry com backoff exponencial
    - Intervalos: 10s, 30s, 1min, 2min, 5min (cap)
    - Aplicar para falhas de conexao com PEP
    - Aplicar para falhas de push ao servidor central
    - Gerar log de alerta quando offline >10min
  - [ ] 3.9 Implementar persistencia de estado (watermark)
    - Salvar ultimo registro processado em arquivo local
    - Retomar do ponto de parada apos reinicio
    - Evitar reprocessamento duplicado
  - [ ] 3.10 Compilar binario standalone
    - Build com `CGO_ENABLED=0` para portabilidade
    - Cross-compile para Linux AMD64 (ambiente hospitalar)
    - Makefile com targets: `build`, `test`, `run`
  - [ ] 3.11 Garantir que testes do agente passam
    - Executar APENAS os 4-6 testes escritos em 3.1
    - Verificar polling, mapeamento e mascaramento
    - NAO executar suite completa de testes

**Criterios de Aceite:**
- Os 4-6 testes escritos em 3.1 passam
- Agente conecta ao postgres-pep-sim
- Polling detecta novos obitos corretamente
- Campos mapeados para estrutura padrao
- CPF mascarado, CNS completo
- Push funcional para servidor central
- Retry com backoff funcionando
- Binario compila como standalone

---

### Backend - Endpoint de Recepcao

#### Grupo 4: API para Receber Eventos do Agente
**Dependencias:** Grupo 3

- [ ] 4.0 Completar endpoint de recepcao no servidor central
  - [ ] 4.1 Escrever 3-5 testes focados para o endpoint
    - Teste de autenticacao via API Key
    - Teste de validacao de payload
    - Teste de publicacao no Redis Streams
  - [ ] 4.2 Criar endpoint `POST /api/v1/pep/eventos`
    - Handler em `backend/internal/handlers/pep_handler.go`
    - Rota registrada no router existente
  - [ ] 4.3 Implementar autenticacao via API Key
    - Header `X-API-Key` obrigatorio
    - Validar API Key contra tabela `hospitais` (campo em `config_conexao`)
    - Retornar 401 se invalido
  - [ ] 4.4 Implementar validacao de payload
    - Validar campos obrigatorios do `ObitoEvent`
    - Retornar 400 com erros especificos se invalido
  - [ ] 4.5 Inserir evento recebido na base de dados
    - Inserir em `obitos_simulados` ou tabela dedicada `obitos_pep`
    - Marcar origem como "pep" vs "simulado"
    - Manter compatibilidade com `GetUnprocessed()`
  - [ ] 4.6 Publicar no Redis Streams
    - Reutilizar logica do `obito_listener.go`
    - Stream: `obitos:detectados`
    - Formato: mesmo `ObitoEvent` do listener atual
  - [ ] 4.7 Garantir que testes do endpoint passam
    - Executar APENAS os 3-5 testes escritos em 4.1
    - Verificar autenticacao, validacao e publicacao
    - NAO executar suite completa de testes

**Criterios de Aceite:**
- Os 3-5 testes escritos em 4.1 passam
- Endpoint autenticado via API Key
- Payload validado corretamente
- Evento inserido no banco de dados
- Evento publicado no Redis Streams
- Motor de triagem processa eventos do PEP

---

### Testes e Integracao

#### Grupo 5: Revisao de Testes e Integracao End-to-End
**Dependencias:** Grupos 1-4

- [ ] 5.0 Revisar testes e validar integracao completa
  - [ ] 5.1 Revisar testes existentes dos Grupos 3-4
    - Revisar os 4-6 testes do agente (Tarefa 3.1)
    - Revisar os 3-5 testes do endpoint (Tarefa 4.1)
    - Total existente: aproximadamente 7-11 testes
  - [ ] 5.2 Analisar gaps criticos de cobertura
    - Identificar fluxos de integracao nao testados
    - Focar APENAS em gaps relacionados a esta feature
    - Priorizar fluxo end-to-end: PEP -> Agente -> Backend -> Redis
  - [ ] 5.3 Escrever ate 5 testes adicionais para gaps criticos
    - Teste de integracao: agente + postgres-pep-sim
    - Teste de integracao: agente + endpoint central
    - Teste de resiliencia: retry apos falha de conexao
    - NAO escrever testes exaustivos para todos os cenarios
  - [ ] 5.4 Executar testes especificos da feature
    - Executar APENAS testes relacionados a esta spec
    - Total esperado: aproximadamente 12-16 testes
    - NAO executar suite completa da aplicacao
  - [ ] 5.5 Validar demo end-to-end
    - Iniciar postgres-pep-sim
    - Iniciar agente conectado ao simulador
    - Verificar deteccao de obito inserido automaticamente
    - Verificar evento recebido no backend
    - Verificar publicacao no Redis Streams
    - Verificar processamento pelo motor de triagem

**Criterios de Aceite:**
- Todos os testes especificos da feature passam (12-16 testes)
- Fluxo end-to-end funcional para demo
- Obito detectado em <5s apos insercao no PEP simulado
- Nenhum dado pessoal em logs

---

## Ordem de Execucao

Sequencia recomendada de implementacao:

1. **Grupo 1: Container de Simulacao PEP** - Base para desenvolvimento e testes
2. **Grupo 2: Sistema de Configuracao YAML** - Define contrato de configuracao
3. **Grupo 3: Agente PEP (Go)** - Core da feature, binario standalone
4. **Grupo 4: Endpoint de Recepcao** - Backend para receber eventos do agente
5. **Grupo 5: Testes e Integracao** - Validacao final e preparacao para demo

## Notas Tecnicas

### Arquivos a Criar
- `/pep-agent/` - Novo modulo Go standalone
- `/pep-agent/cmd/pep-agent/main.go` - Entrypoint do agente
- `/pep-agent/internal/config/config.go` - Parser YAML
- `/pep-agent/internal/database/drivers.go` - Conexoes SQL
- `/pep-agent/internal/poller/poller.go` - Logica de polling
- `/pep-agent/internal/pusher/pusher.go` - Push para central
- `/pep-agent/mapping.example.yaml` - Exemplo de configuracao
- `/docker/postgres-pep-sim/` - Dockerfile e scripts
- `/backend/internal/handlers/pep_handler.go` - Novo endpoint

### Arquivos a Modificar
- `/docker-compose.yml` - Adicionar servico postgres-pep-sim
- `/backend/internal/models/hospital.go` - Adicionar campo `api_key` em `HospitalConfig`
- `/backend/internal/router/router.go` - Registrar nova rota

### Codigo a Reutilizar
- `backend/internal/services/listener/obito_listener.go` - Estrutura `ObitoEvent`, padrao de polling
- `backend/internal/models/models.go` - Funcao `MaskName()` como referencia para `MaskCPF()`
- `backend/config/config.go` - Padroes de leitura de env vars

### Seguranca
- Usuario do PEP: permissao SELECT apenas
- Agente: Outbound Only, nunca aceita conexoes externas
- API Key rotacionavel por hospital
- CPF mascarado em transmissao e logs
- HTTPS obrigatorio em producao
