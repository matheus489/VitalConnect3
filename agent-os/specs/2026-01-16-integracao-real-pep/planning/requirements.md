# Spec Requirements: Integracao Real com PEP

## Initial Description

Adaptar o listener existente para conectar a bancos de dados hospitalares reais (PostgreSQL/MySQL/Oracle), substituindo a tabela simulada por integracao real com sistemas de PEP. Incluir mapeamento configuravel de campos para suportar diferentes esquemas de diferentes hospitais.

**Requisitos iniciais:**
- Conectar a bancos de dados hospitalares reais (PostgreSQL, MySQL, Oracle)
- Suportar multiplos tipos de banco de dados
- Permitir mapeamento configuravel de campos (hospitais diferentes tem esquemas diferentes)
- Substituir o listener simulado por integracao real com hospital
- Principais sistemas PEP no Brasil: MV, Tasy, Philips

**Contexto:**
- Sistema atual tem death listener monitorando tabela simulada (`obitos_simulados`)
- Feature #11 da Fase 2 do roadmap (estimativa: L - 2 semanas)
- Habilita piloto real com hospital parceiro
- Deadline da demo: 26 de Janeiro
- Seguranca e privacidade de dados sao criticos

## Requirements Discussion

### First Round Questions

**Q1:** Hospital Parceiro Confirmado?
**Answer:** Nao (Simulado). Container Docker `postgres-pep-sim` simula esquema Tasy (Philips) ou MV. Narrativa para video: "Agente conectado a instancia espelhada do Tasy do HGG".

**Q2:** Metodo de Conexao ao PEP?
**Answer:** Conexao Direta ao Banco (Read-Only). Metodo "guerrilha" comum em GovTech Brasil. APIs de PEP sao caras ou inexistentes.

**Q3:** Onde o Listener Roda?
**Answer:** On-Premise (Agente Instalado Localmente). Binario Go roda no servidor do hospital. Le banco local e Push via HTTPS (Outbound Only). Nao exige abrir portas no firewall (seguranca).

**Q4:** Campos Obrigatorios para Captar do PEP?
**Answer:** Campos confirmados + CNS/CPF. Adicionar: `cns_paciente` (Cartao Nacional de Saude) ou `cpf` (mascarado). Crucial para Central de Transplantes validar identidade unica.

**Q5:** Interface de Mapeamento de Campos?
**Answer:** Arquivo de Configuracao (YAML/JSON). Nao fazer UI de "De-Para". Admin edita `mapping.yaml`:
```yaml
source_table: "TASY.PACIENTES"
fields:
  name: "NM_PACIENTE"
  death_time: "DT_OBITO"
```
Mais rapido e mostra flexibilidade tecnica.

**Q6:** Fallback se Conexao com PEP Cair?
**Answer:** Retry automatico com Backoff + Alertas. Tenta reconectar: 10s, 30s, 1min... Offline >10min gera alerta para Health Check (spec 20).

**Q7:** Requisitos de Seguranca para Conexao?
**Answer:** Outbound Only & Read-Only User. Agente nunca aceita conexoes de fora. Usuario do banco: permissao SELECT apenas nas views necessarias.

**Q8:** Abordagem para Demo de 26/Janeiro?
**Answer:** Arquitetura production-ready com dados simulados. Listener real conectando em banco mock. Insere obito a cada 2 minutos. Prova TRL 5 sem depender de terceiros.

**Q9:** O que esta Fora do Escopo?
**Answer:**
- HL7/FHIR (deixar para V2, foco em SQL direto)
- Escrita (Write-back) - jamais escrever no PEP
- Migracao de Historico - olha apenas "daqui para frente" (D+0)

### Existing Code to Reference

**Similar Features Identified:**
- Feature: Death Listener atual - Path: `services/listener/` (servico Go existente)
- Componentes a reutilizar: Polling mechanism, Redis Streams publisher
- Backend logic: Hospital model com campo `config_conexao` (JSONB)

### Follow-up Questions

Nao foram necessarias perguntas de follow-up. As respostas do usuario foram completas e detalhadas.

## Visual Assets

### Files Provided:
Nenhum arquivo visual fornecido.

### Visual Insights:
N/A - Verificacao via bash confirmou ausencia de arquivos na pasta `visuals/`.

## Requirements Summary

### Functional Requirements

**Agente PEP Listener (Go):**
- Binario Go standalone que roda on-premise no servidor do hospital
- Conexao direta read-only a bancos PostgreSQL/MySQL/Oracle
- Polling periodico de tabela/view configurada
- Mapeamento de campos via arquivo YAML (`mapping.yaml`)
- Push de eventos via HTTPS para servidor central (Outbound Only)
- Retry com backoff exponencial (10s, 30s, 1min...)
- Alerta para Health Check quando offline >10min

**Campos a Captar:**
- Nome do paciente
- Data/hora do obito
- Causa mortis
- Idade
- CNS (Cartao Nacional de Saude) ou CPF (mascarado)
- Hospital de origem
- Identificador unico do registro

**Container de Simulacao (postgres-pep-sim):**
- Simula esquema de banco Tasy/MV
- Insere registro de obito automaticamente a cada 2 minutos
- Dados ficticios mas realistas para demo
- Estrutura de tabelas compativel com mapeamento configuravel

### Reusability Opportunities

- **Death Listener existente:** Reutilizar logica de polling e publicacao no Redis Streams
- **Hospital model:** Campo `config_conexao` (JSONB) ja existe para armazenar configuracoes
- **Redis Streams:** Infraestrutura de mensageria ja implementada no MVP

### Scope Boundaries

**In Scope:**
- Agente Go para conexao direta a banco de dados
- Suporte a PostgreSQL, MySQL, Oracle
- Arquivo de configuracao YAML para mapeamento de campos
- Container Docker simulando PEP (postgres-pep-sim)
- Retry automatico com backoff exponencial
- Integracao com Health Check existente (alertas)
- Modelo de seguranca: Outbound Only + Read-Only User

**Out of Scope:**
- Integracao via HL7/FHIR (deixar para V2)
- Qualquer operacao de escrita (Write-back) no PEP
- Migracao de historico de obitos (apenas D+0 em diante)
- Interface grafica para mapeamento de campos (apenas YAML)
- Conexao com hospital real (demo usa simulacao)

### Technical Considerations

**Arquitetura:**
- Agente roda on-premise, nunca aceita conexoes externas
- Comunicacao unidirecional: Agente -> Servidor Central (HTTPS)
- Nao requer abertura de portas no firewall do hospital

**Seguranca:**
- Usuario de banco com permissao SELECT apenas
- Acesso restrito a views/tabelas especificas
- CPF mascarado nos dados transmitidos
- Conformidade LGPD para dados de saude

**Configuracao (mapping.yaml):**
```yaml
source_table: "TASY.PACIENTES"
fields:
  name: "NM_PACIENTE"
  death_time: "DT_OBITO"
  cause: "DS_CAUSA_MORTIS"
  age: "NR_IDADE"
  cns: "NR_CNS"
  cpf: "NR_CPF"
```

**Resiliencia:**
- Retry com backoff: 10s -> 30s -> 1min -> ...
- Alerta automatico apos 10min offline
- Integracao com spec #20 (Health Check e Monitoramento)

**Demo (26/Janeiro):**
- Arquitetura production-ready com dados simulados
- Narrativa: "Agente conectado a instancia espelhada do Tasy do HGG"
- Prova de conceito TRL 5 (Technology Readiness Level)
- Insercao automatica de obito a cada 2 minutos para demonstracao ao vivo

**Stack Tecnica:**
- Linguagem: Go (binario standalone)
- Bancos suportados: PostgreSQL, MySQL, Oracle
- Mensageria: Redis Streams (existente)
- Containerizacao: Docker (simulador PEP)
- Configuracao: YAML
