# Requisitos do Agente Assistente Pessoal SIDOT

## 1. Escopo e Capacidades (Agente Híbrido)

**Decisão:** Q&A (RAG) + Function Calling (Ações)

O agente será um **"co-piloto operacional"**, não apenas um chatbot de suporte.

### Exemplos de Uso:
- **Q&A:** "Como faço para rejeitar uma córnea?" → Busca na documentação
- **Ação Simples:** "Mostre as ocorrências críticas do HUGOL" → Consulta banco e renderiza tabela
- **Ação Complexa:** "Avise a equipe do plantão noturno que a captação atrasou" → Busca escala, identifica usuários, envia Push/SMS

---

## 2. Consciência de Contexto e Tenant (Segurança)

**Decisão:** Tenant-Aware & Role-Based

### Implementação:
- Middleware Python recebe Token JWT
- Consultas ao Vector Store filtram por `metadata={tenant_id: "..."}`
- Tools verificam permissões antes de executar
- Exemplo: Operador não pode "Cadastrar novo Hospital" - AI recusa baseado na role

---

## 3. Confirmação de Ações Críticas (Human-in-the-Loop)

**Decisão:** Confirmação UI Interativa

### Fluxo:
1. Usuário: "Descarte a ocorrência #123 por motivo de HIV."
2. Agente: "Entendido. Vou alterar o status da ocorrência #123 para 'Descartada'. Confirma?"
3. Interface: Renderiza botões `[Confirmar]` e `[Cancelar]`
4. Ação só é executada após clique físico no botão

---

## 4. Arquitetura Técnica (Microserviço AI)

**Decisão:** Stack Python Dedicado

### Estrutura:
- **Framework:** LlamaIndex (Gestão de contexto e RAG) + FastAPI (API)
- **Comunicação:** Backend Go atua como Gateway, repassa requisições do Frontend para Python
- **Autenticação:** Validada no Go antes de chegar ao Python

---

## 5. Message Broker & Result Backend

### Decisão:
- **Broker:** Redis (alta velocidade para filas)
- **Result Backend:** PostgreSQL

### Justificativa:
Auditoria obrigatória. Log persistente no Postgres: "O Agente X executou a ação Y às 14:00 baseada no prompt Z". Redis não persiste com segurança suficiente.

---

## 6. Provedor de LLM

**Decisão:** Provider-Agnostic (Padrão: OpenAI GPT-4o)

### Estratégia:
Arquitetura preparada para trocar para:
- Azure OpenAI (se Secretaria de Saúde exigir dados em servidor Enterprise)
- Modelos locais via Ollama (Llama 3)
- Usando abstração do LlamaIndex

---

## 7. Interface do Usuário

**Decisão:** Widget Flutuante Global (Omnipresente)

### UI:
- Botão flutuante (FAB) no canto inferior direito
- Expande painel lateral que não bloqueia tela principal
- Usuário vê Dashboard enquanto conversa com assistente

---

## 8. Memória de Conversação

**Decisão:** Memória Persistente e Semântica

### Funcionalidade:
O assistente lembra interações passadas.
- Ex: "Lembra aquele caso do HGG ontem? Gere um relatório similar para hoje."

---

## 9. Exclusões

### Fora do Escopo:
- ❌ **Voice-to-Text Streaming:** Complexidade alta de WebSocket. Áudio gravado (blob) para transcrição (Whisper), mas não stream ao vivo.
- ❌ **Intervenção Clínica Autônoma:** AI nunca decide sozinho se órgão é viável. Apenas apresenta dados para médico decidir.

---

## 10. Especificações de UI (Generative UI)

### Chat com Componentes Ricos:
- Pergunta: "Liste as ocorrências"
- Resposta: Mini-Card ou Tabela Interativa dentro do balão (não texto puro)

### Estado de "Thinking":
- Indicadores claros: "Consultando Base de Conhecimento..." ou "Executando Ação..."

### Botões de Ação:
- Balões com botões interativos: `[Ver Detalhes]`, `[Confirmar]`, `[Cancelar]`

---

## Referências do Codebase Existente

### Padrões a Reutilizar:
- Sistema SSE de notificações (`/api/v1/notifications/stream`) para comunicação real-time
- Arquitetura de serviços em `/backend/internal/services/`
- Sistema de autenticação JWT existente
- Estrutura multi-tenant com `central_id`
