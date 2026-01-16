# Missao do Produto

## Pitch

**VitalConnect** e uma solucao GovTech de interoperabilidade ativa (middleware) que ajuda equipes de Centrais de Transplantes e Bancos de Olhos a captar orgaos e tecidos viaveis dentro da janela de oportunidade critica, conectando automaticamente os sistemas hospitalares (PEP/EMR) aos processos de notificacao de obitos.

O nome reflete a conexao critica ("Connect") de dados para preservar a vida ou a qualidade de vida ("Vital"), alinhando-se ao proposito de saude publica e transplantes.

## O Problema

### Subnotificacao de Obitos e Perda de Tecidos Viaveis

A captacao de corneas possui uma janela de oportunidade de apenas **6 horas** apos o obito por Parada Cardiorrespiratoria (PCR). Atualmente, a notificacao depende de processos manuais (ligacoes telefonicas) que frequentemente falham ou atrasam.

**Consequencias:**
- Tecidos viaveis sao perdidos diariamente por notificacoes tardias
- Falta de dados estatisticos sobre a real taxa de subnotificacao
- Sobrecarga da equipe medica hospitalar com tarefas administrativas
- Filas de transplante permanecem maiores do que o necessario

**Nossa Solucao:** Um middleware baseado em arquitetura orientada a eventos que detecta automaticamente registros de obito nos sistemas hospitalares, aplica regras de triagem inteligente e notifica a equipe de captacao em tempo real, eliminando a dependencia humana no processo de notificacao.

## Usuarios

### Clientes Primarios
- **Centrais de Transplantes Estaduais:** Orgaos publicos responsaveis pela coordenacao de captacao e distribuicao de orgaos
- **Bancos de Olhos:** Instituicoes especializadas na captacao e preservacao de corneas
- **Secretarias de Saude:** Gestores publicos que necessitam de metricas e auditoria do processo

### Personas de Usuario

**Enfermeiro da Central de Transplantes** (30-50 anos)
- **Funcao:** Coordenador de captacao em regime de plantao 24h
- **Contexto:** Trabalha sob alta pressao com multiplas notificacoes simultaneas, precisa mobilizar equipes de captacao rapidamente
- **Dores:** Recebe notificacoes atrasadas (frequentemente apos as 6h), dados incompletos sobre causa do obito, informacoes de contato da familia incorretas ou ausentes
- **Objetivos:** Receber alertas imediatos com dados completos e confiaveis para maximizar captacoes bem-sucedidas

**Gestor da SVO/SES** (40-60 anos)
- **Funcao:** Coordenador do Servico de Verificacao de Obito ou gestor da Secretaria Estadual de Saude
- **Contexto:** Responsavel por definir politicas e medir eficiencia do sistema de captacao
- **Dores:** Falta de dados estatisticos sobre obitos reais vs. notificados, impossibilidade de medir taxa de subnotificacao, dificuldade em justificar investimentos
- **Objetivos:** Ter dashboards com metricas claras de eficiencia e rastreabilidade completa do processo

**Equipe Medica Hospitalar** (25-65 anos)
- **Funcao:** Medicos, enfermeiros e tecnicos que registram obitos no prontuario eletronico
- **Contexto:** Alta carga de trabalho assistencial, notificacao manual e uma tarefa adicional facilmente esquecida
- **Dores:** Pressao para lembrar de notificar a central, formularios manuais, tempo perdido em ligacoes
- **Objetivos:** Sistema invisivel que elimine a carga administrativa sem alterar o fluxo de trabalho atual

## Diferenciais

### Deteccao em Tempo Real (Zero-Delay)
Diferente de processos manuais que dependem da memoria e disponibilidade da equipe hospitalar, o VitalConnect detecta o evento de obito no exato momento em que e registrado no prontuario. Isso garante o inicio do cronometro de 6 horas com precisao e maximiza o tempo disponivel para captacao.

### Triagem Inteligente Automatizada
Diferente de sistemas que simplesmente repassam todas as notificacoes, aplicamos regras de negocio configuráveis (idade, causa mortis, intervalo de tempo) para filtrar casos nao elegíveis. Isso reduz falsos positivos e evita sobrecarga da equipe de captacao.

### Auditoria e Rastreabilidade Completa
Cada notificacao gera um "ticket" de captacao com historico completo de acoes e desfecho. Isso permite medir a eficiencia real do processo e identificar gargalos, algo impossivel com notificacoes telefonicas.

### Implantacao Nao-Invasiva
O sistema opera como um agente de escuta que se conecta aos bancos de dados existentes sem exigir modificacoes nos sistemas hospitalares. A equipe medica continua usando os mesmos sistemas de sempre.

## Funcionalidades Principais

### Funcionalidades Core
- **Listener de Eventos:** Monitoramento em tempo real de registros de obito nos sistemas hospitalares via conexao direta ao banco ou API
- **Motor de Triagem:** Aplicacao automatica de regras de elegibilidade (idade, causa mortis, tempo decorrido)
- **Sistema de Alertas:** Notificacoes push, SMS e email para a equipe de plantao

### Funcionalidades de Gestao
- **Gestao de Ocorrencias:** Tickets de captacao com workflow de acompanhamento e registro de desfecho
- **Dashboard Operacional:** Visao em tempo real de notificacoes ativas, pendentes e concluidas
- **Relatorios e Metricas:** Estatisticas de eficiencia, tempo medio de resposta, taxa de conversao

### Funcionalidades Avancadas
- **Configuracao de Regras:** Interface para gestores ajustarem criterios de elegibilidade sem codigo
- **Integracao Multicanal:** Suporte a diferentes protocolos de integracao com sistemas hospitalares
- **API para Terceiros:** Endpoints para integracao com outros sistemas da rede de saude
