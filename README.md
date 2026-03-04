# Weddo API

API multi-tenant para sites de casamento. Fornece confirmação de presença (RSVP), lista de presentes com pagamento integrado (PIX/cartão via InfinitePay ou Mercado Pago) e painel administrativo.

Cada casamento é um tenant isolado. O primeiro tenant é o casamento **Manoela & Rafael — 07.07.2026**.

## Stack

| Componente | Tecnologia | Versão |
|------------|------------|--------|
| Linguagem | Go | 1.26 |
| Arquitetura | Clean Architecture, multi-tenant | — |
| Router HTTP | [chi](https://github.com/go-chi/chi) | v5.2.5 |
| CORS | [go-chi/cors](https://github.com/go-chi/cors) | v1.2.2 |
| Banco de dados | PostgreSQL (via [pgx](https://github.com/jackc/pgx)) — Supabase, Neon, local | v5.5.4 |
| Migrações | [golang-migrate](https://github.com/golang-migrate/migrate) | v4.19.1 |
| Configuração | [envconfig](https://github.com/kelseyhightower/envconfig) + [godotenv](https://github.com/joho/godotenv) | v1.4.0 / v1.5.1 |
| Autenticação admin | JWT ([golang-jwt](https://github.com/golang-jwt/jwt)) | v5.3.1 |
| Validação | [go-playground/validator](https://github.com/go-playground/validator) | v10.30.1 |
| Pagamentos | [InfinitePay](https://www.infinitepay.io/) (checkout) ou [Mercado Pago](https://github.com/mercadopago/sdk-go) (transparente) | — / v1.8.0 |
| UUID | [google/uuid](https://github.com/google/uuid) | v1.6.0 |
| Crypto | [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) (bcrypt) | v0.48.0 |
| Rate Limiting | [httprate](https://github.com/go-chi/httprate) | v0.15.0 |
| Logging | `log/slog` (stdlib) — text (dev) / JSON (prod) | — |
| Integração planilha | [Google Sheets API](https://pkg.go.dev/google.golang.org/api/sheets/v4) | v4 |
| Testes | `testing` (stdlib) + [testify](https://github.com/stretchr/testify) | v1.11.1 |
| Container | Docker (multi-stage build) | — |
| CI | GitHub Actions (build, vet, test, Postman push) | — |

## Quick Start

```bash
cp .env.example .env    # preencher variáveis — ver docs/configuration.md
make setup              # go mod tidy + copia .env se não existir
make seed-dev           # (opcional) insere dados fictícios para dev
make run                # sobe o servidor (carrega .env automaticamente)
```

**Antes de rodar**, gere valores seguros para `JWT_SECRET` e `SEED_ADMIN_PASSWORD`:

```bash
openssl rand -base64 32    # gera JWT_SECRET
openssl rand -base64 20    # gera senha para SEED_ADMIN_PASSWORD
```

Consulte [docs/configuration.md](docs/configuration.md) para instruções detalhadas sobre cada variável, incluindo como configurar InfinitePay ou Mercado Pago.

O servidor sobe em `http://localhost:8080`. O arquivo `.env` é carregado automaticamente via godotenv.

### Docker

```bash
make docker-build       # build da imagem
make docker-run         # sobe container com .env
make docker-stop        # para e remove o container
```

## Estrutura do Projeto

```
├── .github/workflows/    # CI (GitHub Actions)
├── cmd/api/              # Entrypoint (bootstrap, graceful shutdown)
├── internal/
│   ├── domain/           # Entidades e interfaces de repositório
│   │   ├── entity/       # Wedding, Invitation, Guest, Gift, Payment, erros
│   │   └── repository/   # Interfaces (Wedding, Invitation, Guest, Gift, Payment)
│   ├── usecase/          # Casos de uso (wedding, rsvp, invitation, guest, gift, payment)
│   ├── dto/              # Objetos de transferência (request/response)
│   └── infra/
│       ├── config/       # Leitura de env vars
│       ├── database/     # Conexão PostgreSQL, migrações, repositórios
│       ├── gateway/      # InfinitePay + Mercado Pago (PIX + cartão)
│       ├── sheets/       # Cliente Google Sheets (push/pull)
│       ├── seed/         # Dados fictícios para desenvolvimento
│       └── web/
│           ├── handler/  # auth, rsvp, invitation, guest, gift, payment, dashboard, sheets
│           └── middleware/ # Auth JWT, TenantResolver, Logger, Recovery
├── migrations/           # SQL migrations (001-005)
├── postman/              # Collection e environment do Postman
├── docs/                 # Documentação detalhada
├── Dockerfile            # Multi-stage build
└── .cursor/rules/        # Convenções para o Cursor AI
```

## Documentação

| Documento | Conteúdo |
|-----------|----------|
| [docs/configuration.md](docs/configuration.md) | Guia de configuração do `.env` (JWT, senhas, pagamentos) |
| [docs/roadmap.md](docs/roadmap.md) | Roadmap por fases e prioridades |
| [docs/architecture.md](docs/architecture.md) | Arquitetura, multi-tenancy e decisões técnicas |
| [docs/api.md](docs/api.md) | Endpoints, contratos e exemplos |
| [docs/database.md](docs/database.md) | Modelo de dados e migrações |
| [docs/gift-list.md](docs/gift-list.md) | Lista de presentes: fluxos e integração de pagamentos |

### Postman

A pasta `postman/` é um workspace local do Postman. Para usar:

- **VS Code Extension**: abra a pasta `postman/` pela extensão — collection e environment já estão registrados
- **App/Web**: importe `postman/collections/wedding-api.postman_collection.json` e `postman/environments/local.postman_environment.json`

Ajuste `adminEmail` e `adminPassword` no environment, execute **admin > login** e todas as variáveis (`token`, `weddingId`, etc.) serão populadas automaticamente.

### Google Sheets

Integração por tenant via OAuth:

1. Configure `GOOGLE_OAUTH_CLIENT_ID`, `GOOGLE_OAUTH_CLIENT_SECRET`, `GOOGLE_OAUTH_REDIRECT_URL` e `GOOGLE_OAUTH_TOKEN_CIPHER_KEY` no `.env`
2. Faça login admin
3. Execute `admin > sheets > connect-sheets-start` e abra `auth_url`
4. Após callback, use `admin > sheets > push-sheets` e `pull-sheets`

### CI

O workflow `.github/workflows/ci.yml` roda automaticamente em push/PR para `main`:

1. **build-and-test** — `go build`, `go vet`, `go test -race`
2. **postman-push** — sincroniza a collection e environment com o Postman Cloud (apenas em push para `main`)

Para que o push do Postman funcione, adicione o secret `POSTMAN_API_KEY` nas configurações do repositório GitHub (Settings → Secrets → Actions).

Para push local: `POSTMAN_API_KEY=pmak_xxx make postman-push`
