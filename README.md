# mr-wedding-api

API multi-tenant para sites de casamento. Fornece confirmação de presença (RSVP), lista de presentes com pagamento integrado (PIX/cartão via Mercado Pago) e painel administrativo.

Cada casamento é um tenant isolado. O primeiro tenant é o casamento **Manoela & Rafael — 07.07.2026** ([frontend](../mr-wedding/)).

## Stack

| Componente | Tecnologia | Versão |
|------------|------------|--------|
| Linguagem | Go | 1.26 |
| Arquitetura | Clean Architecture, multi-tenant | — |
| Router HTTP | [chi](https://github.com/go-chi/chi) | v5.2.5 |
| CORS | [go-chi/cors](https://github.com/go-chi/cors) | v1.2.2 |
| Banco de dados | SQLite (via [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)) | v1.14.34 |
| Migrações | [golang-migrate](https://github.com/golang-migrate/migrate) | v4.19.1 |
| Configuração | [envconfig](https://github.com/kelseyhightower/envconfig) + [godotenv](https://github.com/joho/godotenv) | v1.4.0 / v1.5.1 |
| Autenticação admin | JWT ([golang-jwt](https://github.com/golang-jwt/jwt)) | v5.3.1 |
| Validação | [go-playground/validator](https://github.com/go-playground/validator) | v10.30.1 |
| Pagamentos | [Mercado Pago SDK Go](https://github.com/mercadopago/sdk-go) (PIX + cartão) | v1.8.0 |
| UUID | [google/uuid](https://github.com/google/uuid) | v1.6.0 |
| Crypto | [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) (bcrypt) | v0.48.0 |
| Logging | `log/slog` (stdlib) | — |
| Testes | `testing` (stdlib) + [testify](https://github.com/stretchr/testify) | v1.11.1 |

## Quick Start

```bash
cp .env.example .env    # preencher variáveis — ver docs/configuration.md
make setup              # go mod tidy + copia .env se não existir
make run                # sobe o servidor (carrega .env automaticamente)
```

**Antes de rodar**, gere valores seguros para `JWT_SECRET` e `SEED_ADMIN_PASSWORD`:

```bash
openssl rand -base64 32    # gera JWT_SECRET
openssl rand -base64 20    # gera senha para SEED_ADMIN_PASSWORD
```

Consulte [docs/configuration.md](docs/configuration.md) para instruções detalhadas sobre cada variável, incluindo como obter credenciais do Mercado Pago.

O servidor sobe em `http://localhost:8080`. O arquivo `.env` é carregado automaticamente via godotenv.

## Estrutura do Projeto

```
├── cmd/api/              # Entrypoint (bootstrap, graceful shutdown)
├── internal/
│   ├── domain/           # Entidades e interfaces de repositório
│   │   ├── entity/       # Wedding, Invitation, Guest, Gift, Payment, erros
│   │   └── repository/   # Interfaces (Wedding, Invitation, Guest, Gift, Payment)
│   ├── usecase/          # Casos de uso (wedding, rsvp, invitation, guest, gift, payment)
│   ├── dto/              # Objetos de transferência (request/response)
│   └── infra/
│       ├── database/     # Conexão SQLite, migrações, repositórios
│       ├── gateway/      # Mercado Pago SDK (PIX + cartão)
│       ├── web/
│       │   ├── handler/  # auth, rsvp, invitation, guest, gift, payment, dashboard
│       │   └── middleware/ # Auth JWT, TenantResolver, Logger, Recovery
│       └── config/       # Leitura de env vars
├── migrations/           # SQL migrations (001-005)
├── docs/                 # Documentação detalhada
└── .cursor/rules/        # Convenções para o Cursor AI
```

## Documentação

| Documento | Conteúdo |
|-----------|----------|
| [docs/configuration.md](docs/configuration.md) | Guia de configuração do `.env` (JWT, senhas, Mercado Pago) |
| [docs/roadmap.md](docs/roadmap.md) | Roadmap por fases e prioridades |
| [docs/architecture.md](docs/architecture.md) | Arquitetura, multi-tenancy e decisões técnicas |
| [docs/api.md](docs/api.md) | Endpoints, contratos e exemplos |
| [docs/database.md](docs/database.md) | Modelo de dados e migrações |
| [docs/gift-list.md](docs/gift-list.md) | Lista de presentes: fluxos e integração Mercado Pago |
