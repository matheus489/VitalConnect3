# Deploy SIDOT no Railway

Este guia explica como fazer deploy do SIDOT no Railway.app.

## Pre-requisitos

1. Conta no [Railway.app](https://railway.app) (pode usar GitHub para login)
2. Projeto no GitHub (vamos fazer push do codigo)
3. Gerar secrets JWT para producao

## Passo 1: Preparar o Repositorio GitHub

Se ainda nao tem o projeto no GitHub:

```bash
# Na pasta do projeto
cd /home/matheus_rubem/SIDOT

# Criar repositorio no GitHub e fazer push
git add .
git commit -m "Prepare for Railway deployment"
git push origin main
```

## Passo 2: Criar Projeto no Railway

1. Acesse [railway.app](https://railway.app)
2. Clique em **"New Project"**
3. Selecione **"Deploy from GitHub repo"**
4. Autorize o Railway a acessar seu GitHub
5. Selecione o repositorio **SIDOT**

## Passo 3: Adicionar Servicos

No projeto Railway, voce vai adicionar 4 servicos:

### 3.1 PostgreSQL

1. Clique em **"+ New"** > **"Database"** > **"PostgreSQL"**
2. Railway cria automaticamente e fornece a `DATABASE_URL`

### 3.2 Redis

1. Clique em **"+ New"** > **"Database"** > **"Redis"**
2. Railway cria automaticamente e fornece a `REDIS_URL`

### 3.3 Backend (API)

1. Clique em **"+ New"** > **"GitHub Repo"**
2. Selecione o mesmo repositorio
3. Nas configuracoes do servico:
   - **Root Directory**: `backend`
   - **Watch Paths**: `/backend/**`

4. Configure as **Variables** (clique em "Variables"):

```
DATABASE_URL=${{Postgres.DATABASE_URL}}
REDIS_URL=${{Redis.REDIS_URL}}
JWT_SECRET=<gerar-com-openssl-rand-base64-32>
JWT_REFRESH_SECRET=<gerar-outro-com-openssl-rand-base64-32>
JWT_ACCESS_DURATION=15m
JWT_REFRESH_DURATION=168h
SERVER_PORT=8080
ENVIRONMENT=production
CORS_ORIGINS=https://<seu-frontend>.up.railway.app
LOGIN_RATE_LIMIT=5
HEALTH_CHECK_INTERVAL=60s
ALERT_COOLDOWN_MINUTES=30
```

5. Gere os JWT secrets no terminal:
```bash
openssl rand -base64 32  # Para JWT_SECRET
openssl rand -base64 32  # Para JWT_REFRESH_SECRET
```

### 3.4 Frontend

1. Clique em **"+ New"** > **"GitHub Repo"**
2. Selecione o mesmo repositorio
3. Nas configuracoes do servico:
   - **Root Directory**: `frontend`
   - **Watch Paths**: `/frontend/**`

4. Configure as **Variables**:

```
NEXT_PUBLIC_API_URL=https://<seu-backend>.up.railway.app/api/v1
```

## Passo 4: Gerar Dominios Publicos

Para cada servico (backend e frontend):

1. Clique no servico
2. Va em **"Settings"** > **"Networking"**
3. Clique em **"Generate Domain"**
4. Anote as URLs geradas

## Passo 5: Atualizar Variaveis com URLs

Agora que voce tem as URLs, atualize:

**No Backend:**
- `CORS_ORIGINS` = URL do frontend (ex: `https://frontend-production-abc123.up.railway.app`)

**No Frontend:**
- `NEXT_PUBLIC_API_URL` = URL do backend + `/api/v1` (ex: `https://backend-production-xyz789.up.railway.app/api/v1`)

## Passo 6: Rodar Migrations

Apos o primeiro deploy do backend, voce precisa rodar as migrations.

1. No Railway, clique no servico **Backend**
2. Va na aba **"Settings"**
3. Em **"Service"**, clique em **"Shell"** ou use Railway CLI:

```bash
# Instalar Railway CLI (se necessario)
npm install -g @railway/cli

# Login
railway login

# Conectar ao projeto
railway link

# Rodar migrations
railway run -s backend -- psql $DATABASE_URL -f migrations/001_initial_schema.sql
# Repita para cada arquivo de migration em ordem
```

Ou simplesmente execute no shell do Railway:
```bash
for f in migrations/*.sql; do psql $DATABASE_URL -f $f; done
```

## Passo 7: Criar Usuario Admin Inicial

No shell do backend ou via psql:

```sql
-- Conecte ao banco e execute:
INSERT INTO tenants (id, name, slug)
VALUES ('00000000-0000-0000-0000-000000000001', 'Sua Central', 'sua-central');

INSERT INTO users (id, tenant_id, nome, email, password_hash, role, ativo)
VALUES (
  gen_random_uuid(),
  '00000000-0000-0000-0000-000000000001',
  'Administrador',
  'admin@seudominio.com.br',
  '$2a$10$rqHjJxPVVg4jHqHqHqHqHuOZfQjJxPVVg4jHqHqHqHqHuOZfQjJx', -- senha: admin123
  'admin',
  true
);
```

**IMPORTANTE:** Troque a senha depois do primeiro login!

## Passo 8: Verificar Deploy

1. Acesse a URL do frontend
2. Faca login com as credenciais criadas
3. Verifique se os dados carregam corretamente

## Custos Estimados

| Servico | Custo Railway |
|---------|---------------|
| Backend | ~$5/mes |
| Frontend | ~$5/mes |
| PostgreSQL | ~$5-10/mes |
| Redis | ~$5/mes |
| **Total** | **~$20-25/mes** |

*Valores podem variar com uso. Railway cobra por recursos usados.*

## Dominio Personalizado (Opcional)

Para usar seu proprio dominio (ex: sidot.gov.br):

1. No servico, va em **Settings** > **Networking**
2. Clique em **"+ Custom Domain"**
3. Digite seu dominio
4. Configure o DNS no seu registrador:
   - Tipo: CNAME
   - Host: @ ou www
   - Valor: (Railway fornece)

## Troubleshooting

### Build falha no backend
- Verifique se `go.mod` e `go.sum` estao commitados
- Confira logs de build no Railway

### Frontend nao conecta ao backend
- Verifique `NEXT_PUBLIC_API_URL` (deve incluir `/api/v1`)
- Verifique `CORS_ORIGINS` no backend (deve ser a URL exata do frontend)

### Erro de banco de dados
- Verifique se `DATABASE_URL` esta usando a variavel do PostgreSQL
- Confirme que migrations foram executadas

### Redis connection refused
- Verifique se `REDIS_URL` esta configurada corretamente

## Suporte

- [Documentacao Railway](https://docs.railway.app)
- [Discord Railway](https://discord.gg/railway)
