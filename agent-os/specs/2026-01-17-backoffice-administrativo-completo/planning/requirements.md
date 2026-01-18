# Requirements: Backoffice Administrativo Completo

## Decisões de Escopo

### 1. Acesso e Arquitetura
- **Decisão**: Rota `/admin` com Middleware Global
- **Acesso**: Exclusivo para `is_super_admin = true`
- **Isolamento**: Acesso total (ignora `tenant_id` nas queries)

### 2. Interface CMD - Command Palette (Coração do Backoffice)
- **Funcionalidade**: Command Palette (`Ctrl+K`) onipresente que atua como editor de código em tempo real
- **Capacidades do CMD**:
  - **Alterar Estilo**: `> Set Primary Color #FF0000`
  - **Estrutura (Sidebar)**: `> Sidebar: Add Item "Relatórios Avançados"` ou `> Sidebar: Move "Config" to Bottom`
  - **Elementos de Tela**: `> Dashboard: Add Widget "Map Preview"` ou `> Dashboard: Hide "Ocorrências Recentes"`
  - **Assets**: `> Upload Logo` (abre file picker)
- **Feedback Visual**: Preview ao lado atualiza instantaneamente ao digitar comandos

### 3. Gestão de Branding & Layout (Estrutura JSONB Expandida)
- **Decisão**: Coluna `theme_config` como Schema Completo de UI
- **Nova Estrutura**:
```json
{
  "theme": {
    "colors": { "primary": "...", "bg": "..." },
    "fonts": { "body": "Inter" }
  },
  "layout": {
    "sidebar": [
       { "label": "Dashboard", "icon": "Home", "link": "/dash" },
       { "label": "Novo Item Customizado", "icon": "Star", "link": "/custom" }
    ],
    "topbar": {
       "show_user_info": true,
       "show_tenant_logo": true
    },
    "dashboard_widgets": [
       { "type": "stats_card", "visible": true, "order": 1 },
       { "type": "map_preview", "visible": false, "order": 2 }
    ]
  }
}
```
- **Aplicação**: Frontend do Tenant renderiza Sidebar e Dashboard iterando sobre este JSON
- **Dinamismo**: Se Admin adiciona item via CMD, aparece na tela do usuário final

### 4. Gestão de Usuários Cross-Tenant
- **Decisão**: Super Admin com poderes totais
- **Impersonate**: "Logar como" qualquer usuário para verificar se alterações de layout funcionam
- **Gestão**: Resetar senhas, banir/desativar, promover admins

### 5. Configurações Globais
- **Decisão**: Gerenciamento centralizado
  - SendGrid/SMTP
  - Twilio/SMS Gateways
  - FCM (Push Notifications)

### 6. Hospitais Cross-Tenant
- **Decisão**: Lista global com filtro por Tenant
- Hospitais pertencem a um único tenant

### 7. Regras de Triagem
- **Decisão**: Templates Globais
  - Criação de "Regras Mestre"
  - Clonáveis para novos Tenants

### 8. Exclusões de Escopo
- **Decisão**: Super Admin edita a FORMA (interface), não o CONTEÚDO CLÍNICO
  - Pode ver logs para auditoria técnica
  - NÃO edita dados de pacientes
  - NÃO altera status de ocorrências

---

## Arquitetura do Sistema CMD

### Fluxo de Funcionamento
```
[Super Admin] -> [CMD Input] -> [Parser de Comandos] -> [Preview Live] -> [Salvar em theme_config]
                                       |
                                       v
                              [Tenant Frontend] -> [Lê theme_config] -> [Renderiza UI Dinâmica]
```

### Comandos Suportados
| Categoria | Comando Exemplo | Ação |
|-----------|-----------------|------|
| **Cores** | `> Set Primary Color #FF0000` | Altera cor primária |
| **Cores** | `> Set Background #F5F5F5` | Altera background |
| **Sidebar** | `> Sidebar: Add Item "Nome" icon="Star" link="/path"` | Adiciona item |
| **Sidebar** | `> Sidebar: Remove "Config"` | Remove item |
| **Sidebar** | `> Sidebar: Move "Item" to Top/Bottom` | Reordena |
| **Dashboard** | `> Dashboard: Add Widget "map_preview"` | Adiciona widget |
| **Dashboard** | `> Dashboard: Hide "stats_card"` | Oculta widget |
| **Dashboard** | `> Dashboard: Show "stats_card"` | Mostra widget |
| **Assets** | `> Upload Logo` | Abre file picker |
| **Assets** | `> Upload Favicon` | Abre file picker |
| **Typography** | `> Set Font "Roboto"` | Altera fonte |

### Componentes Necessários

#### Backend
- `POST /api/v1/admin/tenants/:id/theme` - Atualiza theme_config
- `GET /api/v1/admin/tenants/:id/theme` - Obtém theme_config
- `POST /api/v1/admin/tenants/:id/assets` - Upload de assets (logo, favicon)

#### Frontend (Backoffice)
- `CommandPalette.tsx` - Interface CMD (Ctrl+K)
- `CommandParser.ts` - Parser de comandos para ações
- `ThemePreview.tsx` - Preview live das alterações
- `TenantEditor.tsx` - Página principal de edição

#### Frontend (Tenant App)
- `DynamicSidebar.tsx` - Sidebar renderizada via JSON
- `DynamicDashboard.tsx` - Dashboard com widgets configuráveis
- `ThemeProvider.tsx` - Injeta CSS variables do theme_config

---

## Funcionalidades do Backoffice

### 1. Dashboard Global (`/admin`)
- Métricas agregadas de todos os tenants
- Health check do sistema (erros, latência)
- Alertas críticos

### 2. Gestão de Tenants (`/admin/tenants`)
- Lista de todos os tenants
- CRUD completo
- **Editor Visual com CMD** (feature principal)
- Ativar/Desativar tenants

### 3. Gestão de Usuários (`/admin/users`)
- Lista global (filtro por tenant)
- Impersonate
- Promover/rebaixar
- Reset de senha
- Ban/Desativar

### 4. Gestão de Hospitais (`/admin/hospitals`)
- Lista global com filtro por tenant
- Visualização/edição

### 5. Templates de Triagem (`/admin/triagem-templates`)
- CRUD de regras mestre
- Clonar para tenants

### 6. Configurações Globais (`/admin/settings`)
- SMTP/SendGrid
- SMS/Twilio
- Push/FCM

### 7. Logs e Auditoria (`/admin/logs`)
- Logs globais
- Filtro por tenant
- Apenas visualização

---

## Migrations Necessárias

### Tabela `tenants`
```sql
ALTER TABLE tenants
ADD COLUMN theme_config JSONB DEFAULT '{}',
ADD COLUMN is_active BOOLEAN DEFAULT true;
```

### Tabela `system_settings` (nova)
```sql
CREATE TABLE system_settings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  key VARCHAR(100) UNIQUE NOT NULL,
  value JSONB NOT NULL,
  description TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Tabela `triagem_rule_templates` (nova)
```sql
CREATE TABLE triagem_rule_templates (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  nome VARCHAR(255) NOT NULL,
  tipo VARCHAR(50) NOT NULL,
  condicao JSONB NOT NULL,
  ativo BOOLEAN DEFAULT true,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

---

## Stack Técnica

### Backend
- Go (Gin) - já existente
- PostgreSQL com JSONB
- Endpoints `/api/v1/admin/*`

### Frontend Backoffice
- Next.js 14+ (App Router)
- shadcn/ui + Tailwind CSS
- cmdk (Command Palette library)
- Monaco Editor ou similar (para preview)

### Frontend Tenant (Dinâmico)
- ThemeContext com CSS Variables
- Componentes dinâmicos baseados em JSON
