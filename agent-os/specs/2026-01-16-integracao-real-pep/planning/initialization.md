# Initialization

## Feature Name
Integracao Real com PEP (Prontuario Eletronico do Paciente)

## Description
Adaptar o listener existente para conectar a bancos de dados hospitalares reais (PostgreSQL/MySQL/Oracle), substituindo a tabela simulada por integracao real com sistemas de PEP. Incluir mapeamento configuravel de campos para suportar diferentes esquemas de diferentes hospitais.

## Key Requirements
- Conectar a bancos de dados hospitalares reais (PostgreSQL, MySQL, Oracle)
- Suportar multiplos tipos de banco de dados
- Permitir mapeamento configuravel de campos (hospitais diferentes tem esquemas diferentes)
- Substituir o listener simulado por integracao real com hospital
- Principais sistemas PEP no Brasil: MV, Tasy, Philips

## Context
- O sistema atual tem um death listener que monitora uma tabela simulada (`obitos_simulados`)
- Marcado como "L" (Large - 2 semanas) no roadmap
- Esta e a feature #11 da Fase 2 do roadmap
- Habilita o piloto real com um hospital parceiro
- Deadline da demo: 26 de Janeiro
- Pode ser parcialmente simulado para a demo, mas arquitetura deve ser production-ready
- Seguranca e privacidade de dados sao criticos (dados hospitalares)

## Technical Notes
- Stack atual: Go (backend), PostgreSQL (banco principal), Redis Streams (mensageria)
- O listener atual usa polling com intervalo configuravel
- Hospital model ja tem campo `config_conexao` (JSONB) para configuracoes
- O listener atual processa obitos da tabela local e publica no Redis Streams
