# SIDOT MVP - Guia de Execucao

Este documento contem os comandos necessarios para executar o SIDOT MVP.

## Pre-requisitos

- Go 1.21+
- Node.js 18+
- Docker e Docker Compose
- PostgreSQL 15+ (via Docker)
- Redis 7+ (via Docker)

## 1. Iniciar Servicos de Infraestrutura

```bash
# Na raiz do projeto
cd /home/matheus_rubem/SIDOT

# Iniciar PostgreSQL e Redis via Docker Compose
docker-compose up -d

# Verificar se os servicos estao rodando
docker-compose ps
```

Saida esperada:
```
NAME                COMMAND                  SERVICE             STATUS
sidot-db     "docker-entrypoint.s..." postgres            running
sidot-redis  "docker-entrypoint.s..." redis               running
```

## 2. Executar Migrations

```bash
# Entrar no diretorio do backend
cd /home/matheus_rubem/SIDOT/backend

# Executar script de migrations
./migrations/migrate.sh

# OU usando o Go runner
go run migrations/run_migrations.go
```

## 3. Executar Seeder (Dados de Demonstracao)

```bash
# Executar seeder completo (hospitais, usuarios, regras, obitos)
cd /home/matheus_rubem/SIDOT/backend
go run cmd/seeder/main.go

# Opcoes do seeder:
#   --clear         Limpar dados existentes antes de inserir
#   --hospitals     Apenas hospitais
#   --users         Apenas usuarios
#   --rules         Apenas regras de triagem
#   --obitos        Apenas obitos de demonstracao
#   --live-demo     Inserir obito para demo ao vivo (T+10s)
```

**Dados criados pelo seeder:**

Usuarios:
- admin@sidot.gov.br (Admin) - Senha: demo123
- gestor@sidot.gov.br (Gestor) - Senha: demo123
- operador@sidot.gov.br (Operador) - Senha: demo123

Hospitais:
- HGG: Hospital Geral de Goiania
- HUGO: Hospital de Urgencias de Goias

Regras de Triagem:
- Idade maxima: 80 anos
- Causas excludentes: sepse, meningite, tuberculose, etc.
- Janela de captacao: 6 horas
- Identificacao desconhecida: rejeitar

## 4. Iniciar Backend

```bash
cd /home/matheus_rubem/SIDOT/backend

# Configurar variaveis de ambiente (opcional - usa defaults em desenvolvimento)
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/sidot?sslmode=disable"
export REDIS_URL="redis://localhost:6379/0"

# Iniciar servidor
go run cmd/api/main.go
```

Saida esperada:
```
SIDOT API server starting on port 8080
```

O servidor estara disponivel em: http://localhost:8080

### Endpoints principais:

- `GET /health` - Health check
- `POST /api/v1/auth/login` - Login
- `GET /api/v1/occurrences` - Listar ocorrencias
- `GET /api/v1/metrics/dashboard` - Metricas do dashboard
- `GET /api/v1/notifications/stream` - SSE para notificacoes em tempo real

## 5. Iniciar Frontend

```bash
# Em outro terminal
cd /home/matheus_rubem/SIDOT/frontend

# Instalar dependencias (primeira vez)
npm install

# Iniciar servidor de desenvolvimento
npm run dev
```

Saida esperada:
```
   - Local:        http://localhost:3000
   - Network:      http://192.168.x.x:3000
```

Acesse o dashboard em: http://localhost:3000

## 6. Executar Demo ao Vivo

Para demonstrar a deteccao em tempo real:

```bash
# Em um terminal separado, com backend e frontend ja rodando
cd /home/matheus_rubem/SIDOT/backend

# Executar script de demo ao vivo
go run cmd/demo/live_demo.go
```

O script ira:
1. Mostrar countdown de 10 segundos
2. Inserir um obito elegivel no banco
3. O listener detectara em 3-5 segundos
4. A triagem processara e criara uma ocorrencia
5. O dashboard exibira notificacao em tempo real

**O que observar no dashboard:**
- Badge vermelho piscando no header
- Toast/popup com dados do obito
- Som de alerta (se ativado)
- Nova linha na tabela de ocorrencias

## 7. Executar Testes

### Backend (Go)

```bash
cd /home/matheus_rubem/SIDOT/backend

# Executar todos os testes
go test ./... -v

# Executar testes com cobertura
go test ./... -cover

# Executar apenas testes de integracao E2E
go test ./internal/integration/... -v

# Executar testes de um pacote especifico
go test ./internal/models/... -v
go test ./internal/services/auth/... -v
go test ./internal/handlers/... -v
```

### Frontend (TypeScript/Vitest)

```bash
cd /home/matheus_rubem/SIDOT/frontend

# Executar todos os testes
npm test

# Executar testes em modo watch
npm run test:watch

# Executar testes com cobertura
npm run test:coverage
```

## 8. Comandos Uteis

### Docker

```bash
# Parar todos os servicos
docker-compose down

# Ver logs do PostgreSQL
docker-compose logs -f postgres

# Ver logs do Redis
docker-compose logs -f redis

# Reiniciar servicos
docker-compose restart
```

### Limpeza de Dados

```bash
# Limpar dados do seeder e reinserir
cd /home/matheus_rubem/SIDOT/backend
go run cmd/seeder/main.go --clear
```

### Build de Producao

```bash
# Backend
cd /home/matheus_rubem/SIDOT/backend
go build -o sidot-api cmd/api/main.go

# Frontend
cd /home/matheus_rubem/SIDOT/frontend
npm run build
```

## Troubleshooting

### Erro de conexao com banco

```
Failed to connect to database: ...
```

Solucao:
1. Verificar se Docker Compose esta rodando: `docker-compose ps`
2. Verificar logs: `docker-compose logs postgres`
3. Verificar DATABASE_URL esta correta

### Erro de conexao com Redis

```
Warning: Redis ping failed: ...
```

Solucao:
1. Verificar se Redis esta rodando: `docker-compose ps`
2. Verificar logs: `docker-compose logs redis`

### Frontend nao conecta ao backend

Solucao:
1. Verificar se backend esta rodando na porta 8080
2. Verificar CORS_ORIGINS inclui http://localhost:3000
3. Verificar console do navegador para erros de rede

### Notificacoes nao aparecem

Solucao:
1. Verificar se esta logado no dashboard
2. Verificar console do navegador para erros SSE
3. Verificar se listener e triagem motor estao rodando (ver logs do backend)

## Arquitetura de Servicos

```
                                    +------------------+
                                    |    Frontend      |
                                    |   (Next.js)      |
                                    |   :3000          |
                                    +--------+---------+
                                             |
                                             | HTTP/SSE
                                             v
+------------------+     +-----------------+------------------+
|   PostgreSQL     |<--->|     Backend (Go/Gin)               |
|   :5432          |     |     :8080                          |
+------------------+     +----+------+-------+-------+--------+
                              |      |       |       |
                              |      |       |       |
                         +----+  +---+--+ +--+---+ +-+-------+
                         |       |      | |      | |         |
                         v       v      v v      v v         v
                    Listener  Triagem  SSE   Email    Auth
                    Service   Motor    Hub   Queue    Service
                         |       |      |      |
                         |       |      |      |
                         +---+---+------+------+
                             |
                             v
                      +------+------+
                      |    Redis    |
                      |    :6379    |
                      +-------------+
```

## Credenciais de Teste

| Usuario | Email | Senha | Permissoes |
|---------|-------|-------|------------|
| Admin | admin@sidot.gov.br | demo123 | Todas |
| Gestor | gestor@sidot.gov.br | demo123 | Regras, Metricas, Ocorrencias |
| Operador | operador@sidot.gov.br | demo123 | Operar Ocorrencias |
