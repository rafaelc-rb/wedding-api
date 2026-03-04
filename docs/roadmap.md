# Roadmap

## Fase 1 — Fundação ✅

Setup do projeto, estrutura base e infraestrutura mínima.

- [x] Bootstrap Go (go mod, main.go, Makefile)
- [x] Configuração via env (envconfig + godotenv)
- [x] Conexão SQLite + migrações automáticas no boot
- [x] Entidade Wedding + migração + repositório
- [x] Router chi com middlewares base (CORS, logging, recovery)
- [x] Middleware de autenticação JWT (extrai wedding_id dos claims)
- [x] Middleware TenantResolver (resolve weddingId → wedding_id no context)
- [x] Endpoint de login admin (`POST /api/v1/admin/auth`)
- [x] Health check (`GET /api/v1/health`)
- [x] Seed do primeiro wedding via env vars no boot
- [x] Estrutura de resposta padronizada (sucesso e erro)
- [x] Helpers de validação (decode + validator)
- [x] Erros de domínio (ErrNotFound, ErrUnauthorized, etc.)
- [ ] Dockerfile

## Fase 2 — Confirmação de Presença (RSVP) ✅

Feature principal. Sem isso, os convidados não conseguem confirmar presença.

- [x] Entidades Invitation + Guest
- [x] Migrações SQL (002_create_invitations, 003_create_guests)
- [x] Repositórios SQLite (scoped por wedding_id)
- [x] Use case RSVP (buscar por nome no tenant, confirmar, idempotência)
- [x] Use case CRUD de convites (admin, scoped por wedding_id)
- [x] Use case CRUD de convidados (admin, scoped)
- [x] Handlers públicos: `POST /w/{weddingId}/rsvp`, `GET /w/{weddingId}/rsvp/invitation`
- [x] Handlers admin: CRUD invitations, CRUD guests, dashboard
- [ ] Testes unitários dos use cases
- [ ] Integração com o frontend (ajustar form action + JS de submit)

## Fase 3 — Lista de Presentes ✅

Substituir o Casar.com por solução própria com PIX e cartão.

- [x] Entidades Gift + Payment
- [x] Migrações SQL (004_create_gifts, 005_create_payments)
- [x] Repositórios SQLite (scoped por wedding_id)
- [x] Integração Mercado Pago (SDK Go v1.8.0) para PIX e cartão
- [x] Gateway com graceful degradation (503 se MP_ACCESS_TOKEN não configurado)
- [x] Use case: listar presentes (público, por tenant)
- [x] Use case: iniciar pagamento (público — PIX + cartão)
- [x] Use case: webhook de confirmação (Mercado Pago → API)
- [x] Use case: CRUD de presentes (admin, scoped)
- [x] Use case: dashboard com stats de gifts + receita (admin, scoped)
- [x] Handlers públicos: `/w/{weddingId}/gifts`, `/w/{weddingId}/gifts/{id}/purchase`, `/payments/{id}/status`
- [x] Handlers admin: CRUD gifts, listagem/detalhe payments
- [x] Webhook: `POST /payments/webhook`
- [ ] Testes unitários dos use cases
- [ ] Integração com o frontend

## Fase 4 — Polimento

- [ ] Dockerfile
- [ ] Seed de dados para desenvolvimento
- [ ] Rate limiting nos endpoints públicos
- [ ] Logs estruturados em produção (JSON)
- [ ] CI/CD básico
- [ ] Deploy (VPS, Fly.io, Railway ou similar)
- [ ] Monitoramento básico (uptime, erros)

## Fase 5 — Plataforma (Futuro)

Evoluções para oferecer o serviço a outros casais.

- [ ] Fluxo de cadastro de novo wedding (self-service ou admin global)
- [ ] Credenciais Mercado Pago por tenant (cada casal recebe na própria conta)
- [ ] Painel super-admin (gestão de todos os weddings)
- [ ] Domínios customizados ou subdomínios por tenant
- [ ] Limites e planos (free, premium)

## Prioridades e Datas

| Fase | Prioridade | Meta |
|------|-----------|------|
| Fase 1 | ~~Bloqueante~~ ✅ | — |
| Fase 2 | ~~Urgente~~ ✅ | — |
| Fase 3 | ~~Urgente~~ ✅ | — |
| Fase 4 | Importante | Antes do casamento (07.07.2026) |
| Fase 5 | Futuro | Pós-casamento |
