# Spec Requirements: Gestao de Usuarios

## Initial Description

CRUD de usuarios com atribuicao de roles, hospitais vinculados e preferencias de notificacao.

Contexto: Esta feature faz parte do SIDOT - sistema de notificacao de doacao de orgaos. A funcionalidade permitira:
- Criar, visualizar, atualizar e excluir usuarios
- Atribuir roles (operador, gestor, admin)
- Vincular usuarios a hospitais especificos
- Configurar preferencias de notificacao por usuario

## Requirements Discussion

### First Round Questions

**Q1:** Vinculo com Hospitais - Relacao 1:N (usuario pertence a um hospital) ou N:N (usuario pode atuar em multiplos hospitais)?
**Answer:** N:N (Muitos-para-Muitos). Justificativa: Medicos/enfermeiros da central podem atuar em multiplos hospitais ou cobrir uma regional.

**Q2:** Preferencias de Notificacao - Quais canais (Email/SMS/Push/Dashboard) e quais tipos de alerta (Novos obitos, Status de captacao, Alertas de sistema)? Usuario pode configurar horarios de silencio?
**Answer:**
- Canais: Toggle para habilitar/desabilitar Email. Dashboard e mandatorio (nao pode desativar)
- Tipos: Nao configurar agora (assumir "Todas" para o MVP)
- Silencio: Nao implementar - sistema de missao critica, nao pode silenciar alertas

**Q3:** Escopo do Frontend - Quais telas sao necessarias?
**Answer:** Confirmado:
- Listagem com paginacao e busca
- Modal/Tela de Adicionar Usuario
- Modal/Tela de Editar Usuario
- Soft Delete (apenas desativar acesso, manter historico)

**Q4:** Permissoes de Acesso - Quem pode gerenciar usuarios? Gestor de hospital apenas seus usuarios ou todos?
**Answer:**
- Apenas Admin pode criar/editar/excluir usuarios
- Gestor e Operador nao visualizam o modulo, apenas seus proprios perfis

**Q5:** Criacao de Senha - Admin define senha provisoria ou sistema envia link por email?
**Answer:** Admin define senha provisoria (ex: Mudar@123). Motivo: MVP/demo nao depende de servico de email transacional. Futuro: "Esqueci minha senha" via email.

**Q6:** Paginacao - Quantos usuarios sao esperados (dezenas, centenas, milhares)?
**Answer:** Centenas de usuarios esperados. Paginacao padrao (10/20/50 itens por pagina). Volume inicial: ~20 (central) + ~10 (hospitais piloto).

**Q7:** Perfil Proprio - Usuario comum pode editar apenas seu proprio perfil (nome, senha)?
**Answer:** Sim - qualquer usuario pode alterar Nome e Senha. Nao pode alterar Role ou Hospitais vinculados.

**Q8:** Fora do Escopo - SSO, 2FA ou grupos personalizados para MVP?
**Answer:** Fora do Escopo (MVP):
- SSO (Gov.br, Active Directory) - Fase 2
- 2FA - Fase 2 (essencial mas complexo para demo dia 26/01)
- Grupos personalizados - manter roles fixas (admin, gestor, operador)

### Existing Code to Reference

O usuario nao forneceu caminhos especificos para features similares. Porem, com base no roadmap do produto, as seguintes features ja implementadas devem servir como referencia:

**Features Relacionadas no Roadmap (Fase 1 - Concluidas):**
- Item 7: Autenticacao e Autorizacao - Sistema de login com JWT, roles e protecao de endpoints
- Item 8: Tela de Login e Layout Base - Interface de autenticacao e estrutura do dashboard
- Item 9: Dashboard de Ocorrencias - Listagem com filtros, padrao de CRUD no frontend
- Item 10: Configuracao de Hospitais - CRUD de hospitais (padrao similar ao de usuarios)

**Componentes Potencialmente Reutilizaveis:**
- Sistema de roles existente (admin, gestor, operador)
- Layout base do dashboard (sidebar, header, navegacao)
- Padroes de listagem com filtros e paginacao
- Padroes de modal/formulario de criacao/edicao
- Middleware de autorizacao por role

### Follow-up Questions

Nenhuma pergunta de follow-up necessaria. As respostas foram completas e claras.

## Visual Assets

### Files Provided:
Nenhum arquivo visual encontrado na pasta `/home/matheus_rubem/SIDOT/agent-os/specs/2026-01-16-gestao-de-usuarios/planning/visuals/`.

### Visual Insights:
Nao aplicavel - sem assets visuais fornecidos.

## Requirements Summary

### Functional Requirements

**Gestao de Usuarios (Apenas Admin):**
- Listar usuarios com paginacao (10/20/50 por pagina) e busca
- Criar usuario com: nome, email, senha provisoria, role, hospitais vinculados
- Editar usuario: todos os campos exceto email (identificador unico)
- Desativar usuario (soft delete) - manter historico, remover acesso
- Reativar usuario desativado

**Modelo de Dados:**
- Relacao N:N entre Usuario e Hospital (tabela de juncao)
- Roles fixas: admin, gestor, operador
- Preferencias de notificacao: toggle para email (dashboard sempre ativo)
- Campo de status: ativo/inativo (soft delete)

**Perfil Proprio (Todos os Usuarios):**
- Visualizar proprio perfil
- Alterar nome
- Alterar senha (com confirmacao de senha atual)
- Nao pode alterar: role, hospitais vinculados, email

**Sistema de Senha:**
- Admin define senha provisoria na criacao
- Senha deve seguir requisitos minimos de seguranca
- Futuro: recuperacao de senha via email (fora do MVP)

### Reusability Opportunities

- Reutilizar sistema de autenticacao JWT existente
- Seguir padrao de CRUD do modulo de Hospitais
- Reutilizar componentes de listagem/paginacao do Dashboard de Ocorrencias
- Reutilizar middleware de autorizacao por role
- Utilizar componentes shadcn/ui ja configurados no projeto

### Scope Boundaries

**In Scope:**
- CRUD de usuarios (apenas admin)
- Vinculo N:N usuario-hospital
- Preferencia de notificacao por email (toggle)
- Soft delete (desativar/reativar)
- Edicao de perfil proprio (nome, senha)
- Listagem com paginacao e busca

**Out of Scope:**
- SSO (Gov.br, Active Directory) - Fase 2
- 2FA (autenticacao de dois fatores) - Fase 2
- Grupos/roles personalizados - manter apenas admin/gestor/operador
- Preferencias granulares de notificacao por tipo de alerta
- Horarios de silencio para notificacoes
- Recuperacao de senha via email (apenas para futuro)
- Hard delete de usuarios

### Technical Considerations

**Backend (Go):**
- Endpoints REST para CRUD de usuarios
- Tabela de juncao para relacao N:N com hospitais
- Middleware de autorizacao: apenas role "admin" acessa endpoints de gestao
- Endpoint separado para edicao de perfil proprio
- Hash de senha com bcrypt/argon2 (conforme tech-stack)

**Frontend (React/Next.js):**
- Pagina de listagem de usuarios com DataTable
- Modal ou pagina para criar/editar usuario
- Formulario com selecao multipla de hospitais
- Toggle para preferencia de email
- Pagina/modal de perfil proprio
- Protecao de rota: apenas admin ve o modulo

**Seguranca:**
- Validacao de senha provisoria (requisitos minimos)
- Apenas admin pode alterar roles e vinculos
- Usuario comum so edita proprio perfil
- Audit log de alteracoes (se existente no sistema)

**Volume e Performance:**
- Esperado: centenas de usuarios
- Paginacao backend obrigatoria
- Busca server-side para performance
