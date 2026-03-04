# API — Endpoints e Contratos

Base URL: `http://localhost:8080/api/v1`

Todas as respostas seguem o formato JSON. Erros retornam o campo `error`.

## Multi-tenancy

A API é multi-tenant. Cada casamento é um tenant isolado.

- **Endpoints públicos**: o tenant é identificado pelo `{weddingId}` (UUID) na URL
- **Endpoints admin**: o tenant é extraído do JWT (campo `wedding_id` nos claims)
- **Webhook**: o tenant é resolvido via dados do pagamento no banco

Se o `weddingId` não existir ou o wedding estiver inativo, retorna `404`. O frontend é responsável por conhecer o UUID do seu wedding (configurado como variável de ambiente ou build-time).

## Autenticação

Endpoints admin (`/api/v1/admin/*`) exigem header `Authorization: Bearer <token>`.
O token é obtido via `POST /api/v1/admin/auth` e contém `wedding_id` nos claims.

---

## Endpoints Públicos

Prefixo: `/api/v1/w/{weddingId}`

### Confirmar Presença (RSVP)

```
POST /api/v1/w/{weddingId}/rsvp
```

**Request:**

```json
{
  "name": "João Silva"
}
```

**Response 200 — confirmação registrada:**

```json
{
  "guest": {
    "id": "uuid",
    "name": "João Silva",
    "status": "confirmed",
    "confirmed_at": "2026-03-04T10:30:00Z"
  },
  "invitation": {
    "label": "Família Silva"
  },
  "message": "Presença confirmada com sucesso!"
}
```

**Response 404 — nome não encontrado:**

```json
{
  "error": "Convidado não encontrado. Verifique se o nome está exatamente como no convite."
}
```

**Response 409 — já confirmado:**

```json
{
  "guest": {
    "id": "uuid",
    "name": "João Silva",
    "status": "confirmed",
    "confirmed_at": "2026-03-01T14:00:00Z"
  },
  "message": "Presença já estava confirmada."
}
```

### Consultar Convite

```
GET /api/v1/w/{weddingId}/rsvp/invitation?name=João+Silva
```

**Response 200:**

```json
{
  "invitation": {
    "label": "Família Silva",
    "max_guests": 4
  },
  "guests": [
    { "name": "João Silva", "status": "confirmed" },
    { "name": "Maria Silva", "status": "pending" },
    { "name": "Pedro Silva", "status": "pending" }
  ]
}
```

### Listar Presentes

```
GET /api/v1/w/{weddingId}/gifts?category=cozinha
```

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Jogo de Panelas",
      "description": "Jogo com 5 peças antiaderente",
      "price": 350.00,
      "image_url": "https://...",
      "category": "Cozinha",
      "status": "available",
      "created_at": "2026-03-01T10:00:00Z",
      "updated_at": "2026-03-01T10:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 30, "total_pages": 2 }
}
```

A listagem pública mostra apenas presentes com status `available`. Filtros opcionais: `?category=Cozinha&search=panela&page=1&per_page=20`.
```

### Detalhar Presente

```
GET /api/v1/w/{weddingId}/gifts/{id}
```

### Comprar Presente (iniciar pagamento)

```
POST /api/v1/w/{weddingId}/gifts/{id}/purchase
```

**Request — PIX:**

```json
{
  "payer_name": "Tia Maria",
  "payer_email": "maria@email.com",
  "message": "Felicidades ao casal!",
  "payment_method": "pix"
}
```

**Response 201 — PIX QR code gerado:**

```json
{
  "payment_id": "uuid",
  "provider_id": "mp-123456",
  "status": "pending",
  "qr_code": "00020126...",
  "qr_code_base64": "data:image/png;base64,...",
  "expires_at": "2026-03-04T11:00:00Z"
}
```

**Request — Cartão de crédito:**

```json
{
  "payer_name": "Tia Maria",
  "payer_email": "maria@email.com",
  "message": "Felicidades ao casal!",
  "payment_method": "credit_card",
  "card_token": "token-do-sdk-js",
  "installments": 1,
  "payment_method_id": "visa"
}
```

**Response 201 — pagamento aprovado:**

```json
{
  "payment_id": "uuid",
  "provider_id": "mp-123456",
  "status": "approved"
}
```

### Consultar Status do Pagamento

Polling enquanto aguarda pagamento PIX.

```
GET /api/v1/w/{weddingId}/payments/{id}/status
```

**Response 200:**

```json
{
  "payment_id": "uuid",
  "status": "approved",
  "gift_name": "Jogo de Panelas"
}
```

### Webhook Mercado Pago

Recebe notificações de pagamento. Não é chamado pelo frontend. O tenant é resolvido internamente via `payment → gift → wedding`.

```
POST /api/v1/payments/webhook
```

---

## Endpoints Admin

Prefixo: `/api/v1/admin`

O `wedding_id` vem do JWT — não precisa de slug na URL.

### Login

```
POST /api/v1/admin/auth
```

**Request:**

```json
{
  "email": "manu.rafa@email.com",
  "password": "senha"
}
```

**Response 200:**

```json
{
  "token": "eyJhbGciOi...",
  "wedding": {
    "id": "uuid",
    "slug": "manoela-rafael",
    "title": "Casamento Manoela & Rafael"
  }
}
```

### Dashboard

```
GET /api/v1/admin/dashboard
```

**Response 200:**

```json
{
  "rsvp": {
    "total_invitations": 80,
    "total_guests": 200,
    "confirmed": 120,
    "pending": 75,
    "declined": 5,
    "confirmation_rate": 61.5
  },
  "gifts": {
    "total_gifts": 50,
    "purchased": 18,
    "available": 32,
    "total_revenue": 6500.00,
    "total_payments": 18
  }
}
```

> O campo `gifts` só aparece quando há presentes cadastrados.

### Convites (Invitations)

```
GET    /api/v1/admin/invitations          # listar (?page=1&per_page=20&search=silva)
POST   /api/v1/admin/invitations          # criar
GET    /api/v1/admin/invitations/{id}     # detalhar (inclui guests)
PUT    /api/v1/admin/invitations/{id}     # atualizar
DELETE /api/v1/admin/invitations/{id}     # remover (cascade guests)
```

**Criar convite com convidados:**

```json
{
  "code": "SILVA-001",
  "label": "Família Silva",
  "max_guests": 4,
  "guests": [
    { "name": "João Silva" },
    { "name": "Maria Silva" },
    { "name": "Pedro Silva" }
  ]
}
```

**Detalhar convite (response):**

```json
{
  "id": "uuid",
  "code": "SILVA-001",
  "label": "Família Silva",
  "max_guests": 4,
  "guests": [
    { "id": "uuid", "name": "João Silva", "status": "confirmed", "confirmed_at": "..." },
    { "id": "uuid", "name": "Maria Silva", "status": "pending", "confirmed_at": null },
    { "id": "uuid", "name": "Pedro Silva", "status": "pending", "confirmed_at": null }
  ],
  "created_at": "2026-03-01T10:00:00Z",
  "updated_at": "2026-03-01T10:00:00Z"
}
```

### Convidados (Guests)

```
GET    /api/v1/admin/guests               # listar (?page=1&per_page=20&status=confirmed&search=joão)
GET    /api/v1/admin/guests/{id}          # detalhar
PUT    /api/v1/admin/guests/{id}          # atualizar
DELETE /api/v1/admin/guests/{id}          # remover
POST   /api/v1/admin/invitations/{id}/guests  # adicionar a convite existente
```

### Presentes (Gifts)

```
GET    /api/v1/admin/gifts                # listar (?page=1&per_page=20&category=Cozinha&status=available&search=panela)
POST   /api/v1/admin/gifts                # criar
GET    /api/v1/admin/gifts/{id}           # detalhar
PUT    /api/v1/admin/gifts/{id}           # atualizar
DELETE /api/v1/admin/gifts/{id}           # remover
```

### Pagamentos

```
GET /api/v1/admin/payments                # listar (?page=1&per_page=20&status=approved&gift_id=uuid)
GET /api/v1/admin/payments/{id}           # detalhar
```

---

## Padrão de Resposta de Erro

```json
{
  "error": "mensagem descritiva do erro"
}
```

| Status | Uso |
|--------|-----|
| 400 | Validação falhou ou request malformado |
| 401 | Token ausente ou inválido |
| 404 | Recurso não encontrado ou wedding_id inválido |
| 409 | Conflito (ex: presença já confirmada, presente indisponível) |
| 500 | Erro interno |
| 503 | Serviço indisponível (ex: pagamentos sem MP_ACCESS_TOKEN) |

---

## Paginação

Endpoints de listagem suportam paginação:

```
GET /api/v1/admin/guests?page=1&per_page=20
```

```json
{
  "data": [...],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 200,
    "total_pages": 10
  }
}
```
