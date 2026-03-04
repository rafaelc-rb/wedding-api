# Arquitetura

## Clean Architecture

O projeto segue Clean Architecture com dependências apontando para o centro (domínio).

```
┌─────────────────────────────────────────┐
│  infra/web (handlers, middlewares)      │
│  infra/database (repositórios SQLite)   │
│  infra/gateway (Mercado Pago)           │
│  infra/config                           │
│  ┌───────────────────────────────────┐  │
│  │  usecase (regras de aplicação)    │  │
│  │  dto (request/response)           │  │
│  │  ┌─────────────────────────────┐  │  │
│  │  │  domain (entidades,         │  │  │
│  │  │  interfaces de repositório) │  │  │
│  │  └─────────────────────────────┘  │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
```

### Regra de dependência

- **domain/** não importa nenhum outro pacote interno.
- **usecase/** importa apenas `domain/`.
- **dto/** não importa nenhum outro pacote interno (tipos puros de transporte).
- **infra/** importa `domain/`, `usecase/` e `dto/`.

### Camadas

| Camada | Pacote | Responsabilidade |
|--------|--------|------------------|
| Domain | `internal/domain/entity` | Entidades e erros de domínio |
| Domain | `internal/domain/repository` | Interfaces dos repositórios (contratos) |
| Use Case | `internal/usecase/*` | Orquestração de regras de negócio por contexto |
| DTO | `internal/dto` | Structs de request/response para a camada HTTP |
| Infra | `internal/infra/database` | Conexão SQLite, migrações e implementação dos repositórios |
| Infra | `internal/infra/gateway` | Clientes de serviços externos (Mercado Pago) |
| Infra | `internal/infra/web/handler` | Handlers HTTP + helpers (response JSON, validação) |
| Infra | `internal/infra/web/middleware` | Auth JWT, TenantResolver, Logger, Recovery |
| Infra | `internal/infra/config` | Leitura de variáveis de ambiente |
| Entrypoint | `cmd/api` | Bootstrap: config → DB → repos → use cases → handlers → router → server |

## Multi-tenancy

### Estratégia: shared schema com discriminador

Banco único, todas as tabelas possuem `wedding_id`. Simples, eficiente, e suficiente para a escala esperada. Se necessário no futuro, migrar para PostgreSQL com schema-per-tenant ou database-per-tenant.

### Resolução do tenant

| Contexto | Como resolve | Onde |
|----------|-------------|------|
| Endpoints públicos | UUID na URL: `/api/v1/w/{weddingId}/...` | Middleware `TenantResolver` |
| Endpoints admin | `wedding_id` no JWT claims | Middleware `Auth` |
| Webhook de pagamento | `payment → gift → wedding` no banco | Handler direto |

O `TenantResolver` busca o wedding por UUID via `WeddingRepository.FindByID`, valida que está ativo, e injeta o `wedding_id` no context da request.

### Fluxo de uma request pública

```
GET /api/v1/w/{weddingId}/gifts
  → TenantResolver middleware
    → FindByID(weddingId)
    → Valida active == true
    → Injeta wedding_id no context
  → Handler
    → Use Case (recebe wedding_id via context)
      → Repository (filtra por wedding_id)
```

### Fluxo de uma request admin

```
GET /api/v1/admin/guests
  → Auth middleware
    → Valida JWT (HMAC-SHA256)
    → Extrai wedding_id dos claims
    → Injeta no context
  → Handler
    → Use Case (recebe wedding_id via context)
      → Repository (filtra por wedding_id)
```

### Isolamento

- Repositórios **sempre** recebem `weddingID` como parâmetro.
- Nunca existe query sem filtro de tenant (exceto busca de wedding por ID/email).
- Use cases não decidem o tenant — recebem do handler via context.

## Decisões Técnicas

### Go 1.26

Desempenho, simplicidade, tipagem estática e excelente stdlib para HTTP.

### chi v5 (router)

Compatível com `net/http`, middleware chain, agrupamento de rotas, parâmetros de URL. Leve e idiomático.

### SQLite

Banco embutido, zero configuração. Ideal para o escopo inicial. Multi-tenancy via coluna `wedding_id` em todas as tabelas. WAL mode habilitado, foreign keys ativadas na connection string.

### golang-migrate

Migrações versionadas em SQL puro (up/down), executadas automaticamente no boot da aplicação.

### envconfig + godotenv

`godotenv` carrega o `.env` no boot (ignora silenciosamente se não existir). `envconfig` lê as variáveis para uma struct `Config` tipada com suporte a defaults e validação de campos required.

### JWT para autenticação admin

Token stateless com `wedding_id` e `email` nos claims, assinado com HMAC-SHA256. Cada casamento tem seu próprio admin (email + senha bcrypt). Expiração configurável via `JWT_EXPIRATION_HOURS`.

### Mercado Pago (pagamentos)

Gateway para lista de presentes. SDK oficial em Go (v1.8.0). PIX (~0.5% taxa) e cartão (~4-5%). Checkout Transparente. Credenciais globais via env (futuro: por tenant). Gateway com graceful degradation — se `MP_ACCESS_TOKEN` não estiver configurado, endpoints de pagamento retornam `503 Service Unavailable`. Detalhes em [gift-list.md](gift-list.md).

### Validação com go-playground/validator

Validação declarativa via struct tags (`validate:"required,email"`). Helper `decodeAndValidate` faz decode do JSON body + validação em um passo.

### slog (stdlib)

Logger estruturado nativo. Nível configurável via `LOG_LEVEL`. Output texto em desenvolvimento.

## CORS

Configurável via env, separado por vírgula:

```
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://manurafa.com.br
```

## Estrutura de Diretórios

```
mr-wedding-api/
├── cmd/
│   └── api/
│       └── main.go                    # Bootstrap, seed, graceful shutdown
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   ├── wedding.go             # Entidade Wedding (tenant)
│   │   │   ├── invitation.go          # Entidade Invitation (convite)
│   │   │   ├── guest.go               # Entidade Guest + GuestStatus enum
│   │   │   ├── gift.go                # Entidade Gift + GiftStatus enum
│   │   │   ├── payment.go             # Entidade Payment + PaymentStatus/Method enums
│   │   │   └── errors.go              # Erros de domínio
│   │   └── repository/
│   │       ├── wedding.go             # Interface WeddingRepository
│   │       ├── invitation.go          # Interface InvitationRepository
│   │       ├── guest.go               # Interface GuestRepository
│   │       ├── gift.go                # Interface GiftRepository
│   │       └── payment.go             # Interface PaymentRepository
│   ├── usecase/
│   │   ├── wedding/
│   │   │   └── wedding.go             # Authenticate, Seed
│   │   ├── rsvp/
│   │   │   └── rsvp.go                # Confirm, LookupInvitation
│   │   ├── invitation/
│   │   │   └── invitation.go          # CRUD + AddGuest
│   │   ├── guest/
│   │   │   └── guest.go               # CRUD + Dashboard RSVP
│   │   ├── gift/
│   │   │   └── gift.go                # CRUD + Dashboard gifts
│   │   └── payment/
│   │       └── payment.go             # Purchase, HandleWebhook, GetStatus
│   ├── dto/
│   │   ├── request.go                 # Login, RSVP, Invitation, Guest, Gift, Payment requests
│   │   └── response.go                # Todas as responses + PaginatedResponse
│   └── infra/
│       ├── config/
│       │   └── config.go              # Struct Config + Load()
│       ├── database/
│       │   ├── sqlite.go              # Open() + RunMigrations()
│       │   ├── wedding_repository.go  # Implementação WeddingRepository
│       │   ├── invitation_repository.go # Implementação InvitationRepository
│       │   ├── guest_repository.go    # Implementação GuestRepository
│       │   ├── gift_repository.go     # Implementação GiftRepository
│       │   └── payment_repository.go  # Implementação PaymentRepository
│       ├── gateway/
│       │   └── mercadopago.go         # SDK Mercado Pago (PIX + cartão)
│       └── web/
│           ├── handler/
│           │   ├── auth.go            # Login admin
│           │   ├── health.go          # Health check
│           │   ├── rsvp.go            # Confirm, LookupInvitation (público)
│           │   ├── invitation.go      # CRUD invitations + AddGuest (admin)
│           │   ├── guest.go           # CRUD guests (admin)
│           │   ├── gift.go            # CRUD gifts (admin) + listagem pública
│           │   ├── payment.go         # Purchase, GetStatus, Webhook, admin list/detail
│           │   ├── dashboard.go       # Estatísticas RSVP + gifts (admin)
│           │   ├── response.go        # respondJSON, respondError
│           │   └── validator.go       # decodeAndValidate
│           ├── middleware/
│           │   ├── auth.go            # JWT + injeta wedding_id
│           │   ├── tenant.go          # Resolve weddingId da URL
│           │   ├── logger.go          # Request logging
│           │   └── recovery.go        # Panic recovery
│           └── router.go             # Setup chi com rotas e middleware groups
├── migrations/
│   ├── 001_create_weddings.up.sql
│   ├── 001_create_weddings.down.sql
│   ├── 002_create_invitations.up.sql
│   ├── 002_create_invitations.down.sql
│   ├── 003_create_guests.up.sql
│   ├── 003_create_guests.down.sql
│   ├── 004_create_gifts.up.sql
│   ├── 004_create_gifts.down.sql
│   ├── 005_create_payments.up.sql
│   └── 005_create_payments.down.sql
├── docs/
├── .cursor/rules/
├── .env.example
├── .env                               # (gitignored)
├── .gitignore
├── Makefile
├── go.mod
└── go.sum
```
