# Stack Tecnologico

## Visao Geral da Arquitetura

O VitalConnect utiliza uma arquitetura orientada a eventos com servicos desacoplados, priorizando alta disponibilidade, baixa latencia e facilidade de implantacao em ambientes hospitalares com recursos limitados.

---

## Backend e Core

### Linguagem Principal
- **Go (Golang) 1.21+**
- **Justificativa:** Alta performance com baixo consumo de memoria, compilacao para binario unico sem dependencias externas, excelente suporte a concorrencia nativa (goroutines), ideal para agentes que rodam 24/7 em servidores hospitalares. Binarios pequenos facilitam distribuicao e atualizacao.

### Framework Web
- **Gin ou Echo**
- **Justificativa:** Frameworks HTTP leves e de alta performance para Go, com roteamento rapido, middleware flexivel e boa documentacao. Evitamos frameworks pesados para manter binarios compactos.

### ORM / Acesso a Dados
- **sqlx ou GORM**
- **Justificativa:** sqlx para queries SQL com tipagem forte quando precisamos de controle fino; GORM para operacoes CRUD padrao com migrations automaticas. Ambos suportam PostgreSQL nativamente.

---

## Banco de Dados

### Banco Principal
- **PostgreSQL 15+**
- **Justificativa:** Banco relacional robusto com excelente suporte a JSONB (para configuracoes flexiveis), transacoes ACID, replicacao nativa e ecossistema maduro. Padrao em ambientes governamentais e de saude.

### Cache e Sessoes
- **Redis 7+**
- **Justificativa:** Armazenamento em memoria para cache de regras de triagem, sessoes de usuario e rate limiting. Tambem utilizado como broker de mensagens leve quando RabbitMQ nao for necessario.

---

## Mensageria e Filas

### Message Broker
- **Redis Streams** (MVP) / **RabbitMQ** (Producao)
- **Justificativa:** Redis Streams para simplicidade no MVP com fila basica. RabbitMQ para producao com dead-letter queues, retry policies e garantias de entrega. Ambos suportam o padrao de eventos necessario para desacoplar listener, triagem e notificacao.

---

## Frontend

### Framework JavaScript
- **React 18+ com Next.js 14+**
- **Justificativa:** Next.js oferece SSR/SSG para performance, App Router para organizacao, e API Routes para BFF quando necessario. React e o padrao de mercado com maior disponibilidade de desenvolvedores e bibliotecas.

### Estilizacao
- **Tailwind CSS 3+**
- **Justificativa:** Utility-first CSS que acelera desenvolvimento, garante consistencia visual e produz bundles otimizados. Evita CSS customizado fragil e facilita manutencao.

### Componentes UI
- **shadcn/ui**
- **Justificativa:** Componentes acessiveis e bem estilizados baseados em Radix UI, com codigo copiado para o projeto (sem dependencia de pacote). Permite customizacao total mantendo qualidade e acessibilidade.

### Gerenciamento de Estado
- **TanStack Query (React Query)**
- **Justificativa:** Cache e sincronizacao de dados do servidor com invalidacao automatica, retry e loading states. Evita boilerplate de Redux para dados que vem da API.

### Validacao de Formularios
- **React Hook Form + Zod**
- **Justificativa:** React Hook Form para performance em formularios grandes; Zod para validacao type-safe compartilhavel entre frontend e backend (quando usando Node).

---

## Infraestrutura

### Containerizacao
- **Docker + Docker Compose**
- **Justificativa:** Padrao de mercado para empacotamento e deploy. Compose para ambiente de desenvolvimento local e orquestracao simples em producao. Imagens Go sao extremamente leves (Alpine < 20MB).

### Orquestracao (Producao)
- **Docker Swarm** (inicial) / **Kubernetes** (escala)
- **Justificativa:** Swarm para deploys simples em infraestrutura governamental com poucos servidores. Migracao para K8s quando escalar para multiplas centrais de transplante.

### CI/CD
- **GitHub Actions**
- **Justificativa:** Integracao nativa com GitHub, runners gratuitos para projetos publicos, workflows declarativos em YAML. Suporta build de binarios Go e imagens Docker.

### Hospedagem
- **On-Premise** (Datacenter SES) / **Cloud Hibrida**
- **Justificativa:** Dados de saude exigem conformidade com LGPD e normas do CFM. Prioridade para deploy em infraestrutura da Secretaria de Saude com opcao de cloud governamental (GovCloud) para backup.

---

## Servicos Externos

### Autenticacao
- **JWT com refresh tokens**
- **Justificativa:** Stateless auth que escala horizontalmente. Refresh tokens para sessoes longas sem comprometer seguranca. Possivel integracao futura com SSO governamental (GovBR).

### Envio de Email
- **SMTP Institucional** / **SendGrid**
- **Justificativa:** Prioridade para SMTP da propria SES por conformidade. SendGrid como fallback para garantia de entrega e analytics.

### Envio de SMS
- **Zenvia** / **Twilio**
- **Justificativa:** Zenvia como primeira opcao por ser brasileira com bom suporte local. Twilio como alternativa com API mais robusta e documentacao superior.

### Monitoramento e Logs
- **Prometheus + Grafana** (metricas) / **Loki** (logs)
- **Justificativa:** Stack open-source padrao para observabilidade. Prometheus coleta metricas do Go nativamente; Grafana para dashboards; Loki para agregacao de logs sem custo de licenciamento.

### Error Tracking
- **Sentry**
- **Justificativa:** Captura de erros em tempo real com stack traces, contexto de usuario e alertas. Tier gratuito suficiente para MVP.

---

## Testes e Qualidade

### Testes Unitarios e Integracao
- **Go:** `testing` nativo + `testify`
- **React:** `Vitest` + `React Testing Library`
- **Justificativa:** Ferramentas nativas e leves que cobrem a maioria dos casos. Testify adiciona assertions expressivas em Go.

### Testes E2E
- **Playwright**
- **Justificativa:** Mais rapido e confiavel que Cypress para testes de interface. Suporta multiplos browsers e tem boa integracao com CI.

### Linting e Formatacao
- **Go:** `golangci-lint` + `gofmt`
- **JS/TS:** `ESLint` + `Prettier`
- **Justificativa:** Padrao de mercado para cada linguagem. golangci-lint agrega multiplos linters Go em uma ferramenta.

---

## Seguranca

### Criptografia
- **TLS 1.3** para comunicacao
- **bcrypt/argon2** para senhas
- **AES-256** para dados sensiveis em repouso

### Conformidade
- **LGPD:** Anonimizacao de dados pessoais em logs e metricas
- **CFM:** Registro de acesso a dados de pacientes para auditoria
- **HIPAA-like:** Principios de minimo privilegio e segregacao de dados

---

> **Notas**
> - Stack escolhida priorizando: performance, simplicidade de deploy, conformidade governamental
> - Go no backend permite binarios que rodam em servidores hospitalares antigos
> - PostgreSQL e padrao em sistemas de saude brasileiros
> - Toda a stack e open-source, evitando vendor lock-in
