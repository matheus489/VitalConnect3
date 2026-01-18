# VitalConnect - Documentacao Completa

Sistema GovTech para deteccao automatica de obitos e notificacao em tempo real de equipes de captacao de orgaos.

**Versao**: 1.0.0
**Ultima atualizacao**: Janeiro 2026

---

## Indice

1. [Visao Geral](#visao-geral)
2. [Stack Tecnologica](#stack-tecnologica)
3. [Funcionalidades da Plataforma](#funcionalidades-da-plataforma)
4. [Endpoints da API](#endpoints-da-api)
5. [Modelos de Dados](#modelos-de-dados)
6. [Configuracao de Deploy](#configuracao-de-deploy)
7. [Variaveis de Ambiente](#variaveis-de-ambiente)
8. [Guia de Deploy - Render.com](#guia-de-deploy-rendercom)

---

## Visao Geral

O VitalConnect e um middleware que integra multiplas fontes de dados hospitalares (PEP - Plataforma de Notificacao de Obitos) e fornece um dashboard completo para gerenciamento de hospitais e rastreamento de ocorrencias de potenciais doadores de orgaos.

### Objetivos Principais

- Detectar automaticamente obitos em hospitais integrados
- Notificar equipes de captacao em tempo real
- Gerenciar o fluxo de trabalho de ocorrencias
- Garantir conformidade com LGPD
- Fornecer metricas e relatorios

---

## Stack Tecnologica

### Backend
- **Linguagem**: Go 1.21+
- **Framework**: Gin
- **Banco de Dados**: PostgreSQL 15+
- **Cache/Filas**: Redis 7+
- **Autenticacao**: JWT (Access + Refresh tokens)
- **Tempo Real**: SSE (Server-Sent Events)

### Frontend
- **Framework**: Next.js 14+
- **Linguagem**: TypeScript
- **UI**: React 18 + Tailwind CSS
- **Mapas**: Leaflet
- **Estado**: React Query

---

## Funcionalidades da Plataforma

### 1. Autenticacao e Autorizacao

#### Funcionalidades
- Login com email e senha
- Tokens JWT com refresh automatico
- Rate limiting no login (protecao DDoS)
- Logout com revogacao de token
- Auditoria de tentativas de login

#### Papeis de Usuario (RBAC)
| Papel | Descricao | Permissoes |
|-------|-----------|------------|
| `admin` | Administrador do sistema | Acesso total a todas funcionalidades |
| `gestor` | Gestor de hospital | Gerencia seu hospital e usuarios |
| `operador` | Operador de captacao | Visualiza e gerencia ocorrencias |

---

### 2. Gerenciamento de Usuarios

#### Funcionalidades
- Criar, editar, desativar usuarios
- Atribuir papeis e hospitais
- Validacao de senha forte (8+ caracteres, especiais, numeros)
- Validacao de telefone (formato E.164)
- Busca e filtragem de usuarios
- Paginacao

#### Regras de Acesso
- Admin: gerencia todos usuarios
- Gestor: gerencia usuarios do seu hospital
- Operador: apenas visualiza seu perfil

---

### 3. Gerenciamento de Hospitais

#### Funcionalidades
- Cadastro de hospitais com coordenadas geograficas
- Codigo unico por hospital
- Informacoes de contato
- Status ativo/inativo
- Integracao com mapa interativo

#### Campos
- Nome, codigo (unico)
- Endereco, telefone, email
- Latitude e longitude (para mapa)
- Tenant ID (multi-tenant)

---

### 4. Gerenciamento de Ocorrencias

#### Status do Fluxo de Trabalho
```
PENDENTE (inicial)
  |-> EM_ANDAMENTO (em processamento)
  |-> RECUSADA (recusada)
  |-> CANCELADA (cancelada)

EM_ANDAMENTO
  |-> ACEITA (aceita para captacao)
  |-> RECUSADA
  |-> CANCELADA

ACEITA -> CONCLUIDA (apos registro de desfecho)
RECUSADA -> CONCLUIDA (apos registro de desfecho)
```

#### Tipos de Desfecho
- `doacao_realizada` - Doacao efetivada
- `nao_autorizado_familia` - Familia nao autorizou
- `contraindicacao_medica` - Contraindicacao medica
- `janela_expirada` - Janela de tempo expirou
- `outros` - Outros motivos

#### Funcionalidades
- Listagem com filtros avancados (status, hospital, data)
- Visualizacao detalhada com dados completos
- Historico de acoes (timeline)
- Transicao de status com validacao
- Registro de desfecho
- Score de priorizacao automatico
- Mascara de dados pessoais (LGPD)

---

### 5. Regras de Triagem

#### Funcionalidades
- Definir criterios de elegibilidade para doadores
- Prioridade de regras
- Ativacao/desativacao de regras
- Motor de triagem automatico

#### Motor de Triagem
- Processa eventos PEP automaticamente
- Calcula score de priorizacao
- Cria ocorrencias quando criterios sao atendidos
- Dispara notificacoes em tempo real

---

### 6. Gerenciamento de Plantoes

#### Funcionalidades
- Criar escalas de plantao
- Atribuir operadores a hospitais
- Verificar conflitos de horario
- Visualizacao semanal
- Analise de cobertura (gaps)

#### Campos
- Hospital, usuario
- Dia da semana (0-6)
- Hora inicio, hora fim

---

### 7. Dashboard e Metricas

#### KPIs Exibidos
- Obitos elegiveis hoje
- Tempo medio de notificacao
- Corneas potenciais
- Ocorrencias pendentes
- Ocorrencias em andamento

#### Dashboard Geografico (Mapa)
- Mapa interativo com hospitais
- Marcadores com indicadores de urgencia
- Ocorrencias ativas por hospital
- Operador de plantao atual
- Atualizacao em tempo real

---

### 8. Notificacoes

#### Canais Suportados
| Canal | Descricao |
|-------|-----------|
| SSE | Notificacoes em tempo real no browser |
| Email | Notificacoes por email (SMTP) |
| SMS | Notificacoes por SMS |
| Push | Notificacoes push (Firebase FCM) |

#### Eventos Notificados
- Nova ocorrencia criada
- Mudanca de status
- Desfecho registrado
- Alertas do sistema

---

### 9. Relatorios

#### Formatos
- **CSV**: Exportacao tabular
- **PDF**: Relatorio formatado

#### Filtros Disponiveis
- Periodo (data inicio/fim)
- Hospital
- Tipo de desfecho

#### Conformidade LGPD
- Dados anonimizados
- Acesso registrado em auditoria

---

### 10. Auditoria e Conformidade

#### Funcionalidades
- Log de todas as acoes do sistema
- Filtros por usuario, acao, entidade, data
- Niveis de severidade (INFO, WARN, CRITICAL)
- Timeline de ocorrencias
- Exportacao de logs

#### Eventos Auditados
- Login/logout
- CRUD de usuarios
- CRUD de hospitais
- Acoes em ocorrencias
- Visualizacao de dados sensiveis
- Exportacao de relatorios
- Alteracoes em regras de triagem

---

### 11. Multi-Tenant

#### Funcionalidades
- Suporte a multiplas centrais de transplante
- Isolamento completo de dados
- Tenant identificado por slug
- Usuarios vinculados a tenant
- Hospitais vinculados a tenant

---

### 12. Monitoramento de Saude

#### Componentes Monitorados
- API (servidor HTTP)
- Banco de dados (PostgreSQL)
- Cache (Redis)
- Listener de obitos
- Motor de triagem
- Fila de emails
- Hub SSE

#### Status
- `UP` - Funcionando normalmente
- `DEGRADED` - Funcionando com problemas
- `DOWN` - Fora do ar

---

## Endpoints da API

### Autenticacao
| Metodo | Endpoint | Descricao |
|--------|----------|-----------|
| POST | `/api/v1/auth/login` | Login |
| POST | `/api/v1/auth/refresh` | Renovar token |
| POST | `/api/v1/auth/logout` | Logout |
| GET | `/api/v1/auth/me` | Usuario atual |

### Usuarios
| Metodo | Endpoint | Descricao |
|--------|----------|-----------|
| GET | `/api/v1/users` | Listar usuarios |
| GET | `/api/v1/users/:id` | Detalhes do usuario |
| POST | `/api/v1/users` | Criar usuario |
| PATCH | `/api/v1/users/:id` | Atualizar usuario |
| DELETE | `/api/v1/users/:id` | Desativar usuario |
| PATCH | `/api/v1/users/me` | Atualizar perfil proprio |

### Hospitais
| Metodo | Endpoint | Descricao |
|--------|----------|-----------|
| GET | `/api/v1/hospitals` | Listar hospitais |
| GET | `/api/v1/hospitals/:id` | Detalhes do hospital |
| POST | `/api/v1/hospitals` | Criar hospital |
| PATCH | `/api/v1/hospitals/:id` | Atualizar hospital |
| DELETE | `/api/v1/hospitals/:id` | Remover hospital |

### Ocorrencias
| Metodo | Endpoint | Descricao |
|--------|----------|-----------|
| GET | `/api/v1/occurrences` | Listar ocorrencias |
| GET | `/api/v1/occurrences/:id` | Detalhes da ocorrencia |
| GET | `/api/v1/occurrences/:id/history` | Historico |
| PATCH | `/api/v1/occurrences/:id/status` | Atualizar status |
| POST | `/api/v1/occurrences/:id/outcome` | Registrar desfecho |

### Regras de Triagem
| Metodo | Endpoint | Descricao |
|--------|----------|-----------|
| GET | `/api/v1/triagem-rules` | Listar regras |
| POST | `/api/v1/triagem-rules` | Criar regra |
| PATCH | `/api/v1/triagem-rules/:id` | Atualizar regra |
| DELETE | `/api/v1/triagem-rules/:id` | Remover regra |

### Plantoes
| Metodo | Endpoint | Descricao |
|--------|----------|-----------|
| POST | `/api/v1/shifts` | Criar plantao |
| GET | `/api/v1/shifts/:id` | Detalhes do plantao |
| PUT | `/api/v1/shifts/:id` | Atualizar plantao |
| DELETE | `/api/v1/shifts/:id` | Remover plantao |
| GET | `/api/v1/shifts/me` | Meus plantoes |
| GET | `/api/v1/hospitals/:id/shifts` | Plantoes do hospital |
| GET | `/api/v1/hospitals/:id/shifts/today` | Plantoes de hoje |
| GET | `/api/v1/hospitals/:id/shifts/coverage` | Analise de cobertura |

### Metricas
| Metodo | Endpoint | Descricao |
|--------|----------|-----------|
| GET | `/api/v1/metrics/dashboard` | KPIs do dashboard |
| GET | `/api/v1/metrics/indicators` | Indicadores detalhados |

### Mapa
| Metodo | Endpoint | Descricao |
|--------|----------|-----------|
| GET | `/api/v1/map/hospitals` | Hospitais para mapa |

### Relatorios
| Metodo | Endpoint | Descricao |
|--------|----------|-----------|
| GET | `/api/v1/reports/csv` | Exportar CSV |
| GET | `/api/v1/reports/pdf` | Exportar PDF |

### Auditoria
| Metodo | Endpoint | Descricao |
|--------|----------|-----------|
| GET | `/api/v1/audit-logs` | Listar logs |
| GET | `/api/v1/occurrences/:id/timeline` | Timeline da ocorrencia |

### Push Notifications
| Metodo | Endpoint | Descricao |
|--------|----------|-----------|
| POST | `/api/v1/push/subscribe` | Registrar dispositivo |
| DELETE | `/api/v1/push/unsubscribe` | Remover dispositivo |
| GET | `/api/v1/push/subscriptions` | Minhas inscricoes |
| GET | `/api/v1/push/status` | Status do servico |

### SSE (Tempo Real)
| Metodo | Endpoint | Descricao |
|--------|----------|-----------|
| GET | `/api/v1/notifications/stream` | Stream de eventos |

### Saude
| Metodo | Endpoint | Descricao |
|--------|----------|-----------|
| GET | `/health` | Health check basico |
| GET | `/api/v1/health/summary` | Status de todos componentes |
| GET | `/api/v1/health/listener` | Status do listener |
| GET | `/api/v1/health/sse` | Status do SSE |

### Integracao PEP
| Metodo | Endpoint | Descricao |
|--------|----------|-----------|
| POST | `/api/v1/pep/eventos` | Receber evento de obito |
| GET | `/api/v1/pep/status` | Status da integracao |

---

## Modelos de Dados

### Tenant
```sql
CREATE TABLE tenants (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);
```

### User
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nome VARCHAR(255) NOT NULL,
    role user_role NOT NULL DEFAULT 'operador',
    mobile_phone VARCHAR(16),
    email_notifications BOOLEAN DEFAULT true,
    is_super_admin BOOLEAN DEFAULT false,
    ativo BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);
```

### Hospital
```sql
CREATE TABLE hospitals (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    nome VARCHAR(255) NOT NULL,
    codigo VARCHAR(50) UNIQUE NOT NULL,
    endereco TEXT,
    telefone VARCHAR(20),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    config_conexao JSONB DEFAULT '{}',
    ativo BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);
```

### Occurrence
```sql
CREATE TABLE occurrences (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    obito_id UUID NOT NULL,
    hospital_id UUID NOT NULL,
    status occurrence_status DEFAULT 'PENDENTE',
    score_priorizacao INTEGER DEFAULT 50,
    nome_paciente_mascarado VARCHAR(255) NOT NULL,
    dados_completos JSONB NOT NULL,
    data_obito TIMESTAMP WITH TIME ZONE NOT NULL,
    janela_expira_em TIMESTAMP WITH TIME ZONE NOT NULL,
    notificado_em TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);
```

### Tipos ENUM
```sql
-- Papeis de usuario
CREATE TYPE user_role AS ENUM ('operador', 'gestor', 'admin');

-- Status de ocorrencia
CREATE TYPE occurrence_status AS ENUM (
    'PENDENTE', 'EM_ANDAMENTO', 'ACEITA',
    'RECUSADA', 'CANCELADA', 'CONCLUIDA'
);

-- Tipos de desfecho
CREATE TYPE outcome_type AS ENUM (
    'doacao_realizada', 'nao_autorizado_familia',
    'contraindicacao_medica', 'janela_expirada', 'outros'
);

-- Canais de notificacao
CREATE TYPE notification_channel AS ENUM (
    'dashboard', 'email', 'sms', 'push'
);

-- Severidade de auditoria
CREATE TYPE audit_severity AS ENUM ('INFO', 'WARN', 'CRITICAL');
```

---

## Configuracao de Deploy

### Estrutura de Arquivos

```
VitalConnect/
├── backend/
│   ├── Dockerfile
│   ├── cmd/api/main.go
│   ├── internal/
│   │   ├── handlers/
│   │   ├── services/
│   │   ├── repository/
│   │   ├── models/
│   │   └── middleware/
│   ├── migrations/
│   │   └── RAILWAY_INIT.sql
│   └── go.mod
├── frontend/
│   ├── Dockerfile
│   ├── next.config.ts
│   ├── src/app/
│   └── package.json
└── DOCUMENTACAO_COMPLETA.md
```

### Dockerfile Backend

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

### Dockerfile Frontend

```dockerfile
FROM node:18-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM node:18-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build

FROM node:18-alpine AS runner
WORKDIR /app
ENV NODE_ENV production
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static
USER nextjs
EXPOSE 3000
ENV PORT 3000
CMD ["node", "server.js"]
```

---

## Variaveis de Ambiente

### Backend

| Variavel | Descricao | Exemplo |
|----------|-----------|---------|
| `DATABASE_URL` | URL do PostgreSQL | `postgresql://user:pass@host:5432/db` |
| `REDIS_URL` | URL do Redis | `redis://host:6379` |
| `JWT_SECRET` | Chave secreta JWT | (gerar com `openssl rand -base64 32`) |
| `JWT_REFRESH_SECRET` | Chave refresh JWT | (gerar com `openssl rand -base64 32`) |
| `JWT_ACCESS_DURATION` | Duracao access token | `15m` |
| `JWT_REFRESH_DURATION` | Duracao refresh token | `168h` |
| `SERVER_PORT` | Porta do servidor | `8080` |
| `ENVIRONMENT` | Ambiente | `production` |
| `CORS_ORIGINS` | Origens CORS permitidas | `https://frontend.render.com` |
| `LOGIN_RATE_LIMIT` | Limite de tentativas login | `5` |
| `HEALTH_CHECK_INTERVAL` | Intervalo health check | `60s` |
| `ALERT_COOLDOWN_MINUTES` | Cooldown de alertas | `30` |
| `ADMIN_ALERT_EMAIL` | Email para alertas | `admin@example.com` |
| `FCM_SERVER_KEY` | Chave Firebase (opcional) | `...` |
| `SMTP_HOST` | Host SMTP (opcional) | `smtp.gmail.com` |
| `SMTP_PORT` | Porta SMTP | `587` |
| `SMTP_USER` | Usuario SMTP | `user@gmail.com` |
| `SMTP_PASSWORD` | Senha SMTP | `...` |
| `SMTP_FROM` | Email remetente | `noreply@vitalconnect.com` |

### Frontend

| Variavel | Descricao | Exemplo |
|----------|-----------|---------|
| `NEXT_PUBLIC_API_URL` | URL da API | `https://backend.render.com/api/v1` |

---

## Guia de Deploy - Render.com

### 1. Criar Banco de Dados PostgreSQL

1. Acesse [render.com](https://render.com)
2. Clique em **New +** > **PostgreSQL**
3. Configure:
   - Name: `vitalconnect-db`
   - Database: `vitalconnect_db`
   - User: `vitalconnect_db_user`
   - Region: Ohio (US East)
   - Instance Type: Free

### 2. Criar Redis

1. Clique em **New +** > **Redis**
2. Configure:
   - Name: `vitalconnect-redis`
   - Region: Ohio (US East)
   - Instance Type: Free

### 3. Deploy Backend

1. Clique em **New +** > **Web Service**
2. Conecte ao repositorio GitHub
3. Configure:
   - Name: `vitalconnect-backend`
   - Root Directory: `backend`
   - Runtime: Docker
   - Dockerfile Path: `./Dockerfile`
   - Instance Type: Free

4. Adicione variaveis de ambiente:
```
DATABASE_URL=<Internal Database URL do PostgreSQL>
REDIS_URL=<Internal Redis URL>
JWT_SECRET=<gerar com openssl rand -base64 32>
JWT_REFRESH_SECRET=<gerar com openssl rand -base64 32>
JWT_ACCESS_DURATION=15m
JWT_REFRESH_DURATION=168h
SERVER_PORT=8080
ENVIRONMENT=production
CORS_ORIGINS=https://<frontend-url>.onrender.com
LOGIN_RATE_LIMIT=5
HEALTH_CHECK_INTERVAL=60s
ALERT_COOLDOWN_MINUTES=30
```

### 4. Executar Migrations

Apos deploy do backend, conecte ao PostgreSQL:

```bash
# Copie o PSQL Command do painel do PostgreSQL no Render
PGPASSWORD=<senha> psql -h <host> -U <user> <database>
```

Execute o script de migracao:
```sql
-- Cole o conteudo de backend/migrations/RAILWAY_INIT.sql
```

Ou execute via arquivo local:
```bash
PGPASSWORD=<senha> psql -h <host> -U <user> <database> -f backend/migrations/RAILWAY_INIT.sql
```

### 5. Deploy Frontend

1. Clique em **New +** > **Web Service**
2. Conecte ao mesmo repositorio GitHub
3. Configure:
   - Name: `vitalconnect-frontend`
   - Root Directory: `frontend`
   - Runtime: Docker
   - Dockerfile Path: `./Dockerfile`
   - Instance Type: Free

4. Adicione variavel de ambiente:
```
NEXT_PUBLIC_API_URL=https://vitalconnect-backend.onrender.com/api/v1
```

### 6. Atualizar CORS no Backend

Apos obter a URL do frontend, atualize `CORS_ORIGINS` no backend:
```
CORS_ORIGINS=https://vitalconnect-frontend.onrender.com
```

### 7. Acessar o Sistema

**URL**: `https://vitalconnect-frontend.onrender.com`

**Credenciais padrao**:
- Email: `admin@vitalconnect.gov.br`
- Senha: `admin123`

**IMPORTANTE**: Altere a senha apos primeiro acesso!

---

## Comandos Uteis

### Gerar JWT Secrets
```bash
openssl rand -base64 32
```

### Gerar Hash de Senha (bcrypt)
```bash
# No diretorio backend
go run -e 'package main; import ("fmt"; "golang.org/x/crypto/bcrypt"); func main() { h, _ := bcrypt.GenerateFromPassword([]byte("suaSenha"), 10); fmt.Println(string(h)) }'
```

### Atualizar Senha no Banco
```sql
UPDATE users
SET password_hash = '$2a$10$...'
WHERE email = 'admin@vitalconnect.gov.br';
```

### Criar Ocorrencia de Teste
```sql
-- 1. Criar obito simulado
INSERT INTO obitos_simulados (
  id, tenant_id, hospital_id, nome_paciente, data_nascimento,
  data_obito, causa_mortis, prontuario, setor, leito
) VALUES (
  gen_random_uuid(),
  '00000000-0000-0000-0000-000000000001',
  '<hospital_id>',
  'Nome do Paciente',
  '1965-03-15',
  NOW() - INTERVAL '2 hours',
  'Causa do obito',
  'PRONT-001',
  'UTI',
  'Leito 1'
) RETURNING id;

-- 2. Criar ocorrencia (use o id retornado)
INSERT INTO occurrences (
  tenant_id, obito_id, hospital_id, status, score_priorizacao,
  nome_paciente_mascarado, dados_completos, data_obito, janela_expira_em
) VALUES (
  '00000000-0000-0000-0000-000000000001',
  '<obito_id>',
  '<hospital_id>',
  'PENDENTE',
  85,
  'N*** P***',
  '{"nome": "Nome Paciente", "idade": 59}',
  NOW() - INTERVAL '2 hours',
  NOW() + INTERVAL '22 hours'
);
```

---

## Suporte

- **Documentacao Render**: https://render.com/docs
- **Codigo Fonte**: GitHub (repositorio privado)

---

*Documentacao gerada automaticamente - VitalConnect v1.0.0*
