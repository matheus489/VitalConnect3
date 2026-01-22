# SIDOT - Sistema Inteligente de Doação de Órgãos e Tecidos

Sistema de Captacao de Orgaos e Tecidos - Middleware GovTech para deteccao automatica de obitos e notificacao de equipes de captacao em tempo real.

## Requisitos

- Go 1.21+
- Node.js 18+
- Docker e Docker Compose
- PostgreSQL 15+ (ou via Docker)
- Redis 7+ (ou via Docker)

## Setup Rapido

### 1. Configurar variaveis de ambiente

```bash
cp .env.example .env
# Edite o arquivo .env com suas configuracoes
```

### 2. Iniciar servicos de banco de dados

```bash
docker compose up -d postgres redis
```

### 3. Backend

```bash
cd backend
go mod download
go run cmd/api/main.go
```

O backend estara disponivel em: http://localhost:8080

### 4. Frontend

```bash
cd frontend
npm install
npm run dev
```

O frontend estara disponivel em: http://localhost:3000

## Estrutura do Projeto

```
SIDOT/
├── backend/
│   ├── cmd/
│   │   └── api/          # Ponto de entrada da API
│   ├── config/           # Configuracao
│   ├── internal/
│   │   ├── handlers/     # HTTP handlers
│   │   ├── middleware/   # Middlewares Gin
│   │   ├── models/       # Modelos de dados
│   │   ├── repository/   # Acesso a dados
│   │   └── services/     # Logica de negocio
│   ├── migrations/       # Migrations SQL
│   └── pkg/              # Pacotes compartilhados
├── frontend/
│   ├── src/
│   │   ├── app/          # App Router (Next.js 14+)
│   │   ├── components/   # Componentes React
│   │   ├── hooks/        # Custom hooks
│   │   ├── lib/          # Utilitarios
│   │   ├── services/     # API services
│   │   └── types/        # TypeScript types
│   └── public/           # Assets estaticos
├── docker-compose.yml    # Servicos Docker
└── Makefile             # Comandos de desenvolvimento
```

## Tecnologias

### Backend
- Go (Golang) com Gin Framework
- PostgreSQL 15
- Redis 7 (cache e message queue)
- JWT para autenticacao

### Frontend
- Next.js 14+ com App Router
- React 18+
- TypeScript
- Tailwind CSS
- Shadcn/UI
- TanStack Query

## Licenca

Projeto desenvolvido para a Central Estadual de Transplantes de Goias.
