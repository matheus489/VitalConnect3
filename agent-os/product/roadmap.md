# Roadmap do Produto

## Visao Geral

O desenvolvimento do VitalConnect segue uma abordagem incremental, priorizando a entrega de valor o mais rapido possivel. O foco inicial e criar um fluxo funcional de ponta a ponta (deteccao -> triagem -> notificacao) antes de adicionar funcionalidades secundarias.

---

## Fase 1: MVP (Minimo Produto Viavel)

Objetivo: Fluxo completo de deteccao e notificacao funcionando em ambiente controlado.

1. [x] **Modelagem do Banco de Dados** - Definir e criar as tabelas principais: hospitais, regras de triagem, notificacoes, ocorrencias e usuarios. Incluir migrations e seeds iniciais. `S`

2. [x] **Servico Listener Base** - Implementar o agente de escuta que monitora mudancas em uma tabela de obitos simulada, detectando novos registros em tempo real via polling ou triggers. `M`

3. [x] **Motor de Triagem** - Criar o servico que aplica regras de elegibilidade (idade maxima, causas excludentes, tempo maximo decorrido) sobre os eventos detectados pelo listener. `M`

4. [x] **Sistema de Filas** - Implementar a camada de mensageria para garantir resiliencia entre deteccao, triagem e notificacao, com retry automatico e dead-letter queue. `S`

5. [x] **Servico de Notificacao** - Criar o servico que envia alertas via canais configurados (inicialmente email), incluindo template de mensagem com dados do obito. `S`

6. [x] **API de Ocorrencias** - Endpoints REST para criar, listar, atualizar status e registrar desfecho de ocorrencias de captacao. `S`

7. [x] **Autenticacao e Autorizacao** - Sistema de login com JWT, roles (operador, gestor, admin) e protecao de endpoints por permissao. `M`

8. [x] **Tela de Login e Layout Base** - Interface de autenticacao e estrutura base do dashboard (sidebar, header, navegacao) usando React/Next.js. `S`

9. [x] **Dashboard de Ocorrencias** - Tela principal listando ocorrencias ativas com filtros por status, hospital e data. Permitir visualizar detalhes e atualizar status. `M`

10. [x] **Configuracao de Hospitais** - CRUD de hospitais com dados de conexao (simulados no MVP) e configuracoes especificas de integracao. `S`

---

## Fase 2: Produto Completo para Piloto (v1.0)

Objetivo: Sistema pronto para implantacao piloto em ambiente real com um hospital parceiro.

11. [ ] **Integracao Real com PEP** - Adaptar o listener para conectar a um banco de dados hospitalar real (PostgreSQL/MySQL/Oracle), com mapeamento configuravel de campos. `L`

12. [ ] **Notificacao SMS** - Integrar com gateway SMS (ex: Twilio, Zenvia) para envio de alertas para celulares da equipe de plantao. `S`

13. [ ] **Notificacao Push** - Implementar push notifications via web e/ou app mobile para alertas em tempo real mesmo com app fechado. `M`

14. [ ] **Gestao de Plantoes** - Cadastro de escalas de plantao com horarios e responsaveis, direcionando notificacoes para a equipe correta do turno. `M`

15. [ ] **Editor de Regras de Triagem** - Interface visual para gestores criarem e editarem regras de elegibilidade sem necessidade de alteracao de codigo. `M`

16. [ ] **Dashboard de Metricas** - Tela com graficos de notificacoes por periodo, tempo medio de resposta, taxa de conversao (notificacao -> captacao) e ranking por hospital. `M`

17. [ ] **Relatorios Exportaveis** - Geracao de relatorios em PDF e Excel com dados de ocorrencias e metricas para prestacao de contas. `S`

18. [ ] **Logs e Auditoria** - Registro detalhado de todas as acoes do sistema e usuarios, com tela de consulta para rastreabilidade. `M`

19. [ ] **Gestao de Usuarios** - CRUD de usuarios com atribuicao de roles, hospitais vinculados e preferencias de notificacao. `S`

20. [ ] **Health Check e Monitoramento** - Endpoints de status dos servicos, alertas de falha do listener e dashboard de saude do sistema. `S`

---

## Fase 3: Expansao e Melhorias (v2.0)

Objetivo: Escalar para multiplos hospitais e adicionar inteligencia ao sistema.

21. [ ] **Multi-Tenant** - Suporte a multiplas Centrais de Transplante operando na mesma instancia, com isolamento de dados e configuracoes independentes. `L`

22. [ ] **Conectores Padronizados** - Biblioteca de conectores pre-configurados para os principais sistemas de PEP do mercado (MV, Tasy, Philips). `XL`

23. [ ] **App Mobile Nativo** - Aplicativo iOS/Android para equipe de captacao com notificacoes push, lista de ocorrencias e atualizacao de status em campo. `XL`

24. [ ] **Predicao de Elegibilidade** - Modelo de ML que analisa historico para sugerir priorizacao de casos com maior probabilidade de sucesso na captacao. `L`

25. [ ] **Integracao com SNT** - Conexao com o Sistema Nacional de Transplantes para reporte automatico de captacoes e sincronizacao de dados. `L`

26. [ ] **API Publica Documentada** - API REST documentada com OpenAPI/Swagger para integracao de sistemas terceiros da rede de saude. `M`

27. [ ] **Webhooks Configuraveis** - Sistema de webhooks para notificar sistemas externos sobre eventos relevantes (nova ocorrencia, captacao concluida). `S`

28. [ ] **Dashboard Geografico** - Mapa interativo mostrando hospitais, ocorrencias ativas e equipes de captacao em tempo real. `M`

---

> **Notas**
> - Itens ordenados por dependencias tecnicas e arquitetura do produto
> - Cada item representa uma funcionalidade completa (frontend + backend) e testavel
> - Estimativas de esforco: XS (1 dia), S (2-3 dias), M (1 semana), L (2 semanas), XL (3+ semanas)
> - Fase 1 foca em validar o fluxo core antes de adicionar complexidade
> - Fase 2 prepara o sistema para uso em producao com hospital real
> - Fase 3 expande para escala estadual/nacional
