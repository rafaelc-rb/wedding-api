CREATE TABLE IF NOT EXISTS payments (
    id              TEXT PRIMARY KEY,
    gift_id         TEXT NOT NULL REFERENCES gifts(id),
    wedding_id      TEXT NOT NULL REFERENCES weddings(id),
    provider_id     TEXT DEFAULT '',
    amount          REAL NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    payment_method  TEXT NOT NULL,
    payer_name      TEXT NOT NULL,
    payer_email     TEXT DEFAULT '',
    message         TEXT DEFAULT '',
    pix_qr_code     TEXT DEFAULT '',
    pix_expiration  DATETIME,
    paid_at         DATETIME,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_payments_gift_id ON payments(gift_id);
CREATE INDEX IF NOT EXISTS idx_payments_wedding_id ON payments(wedding_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(wedding_id, status);
CREATE INDEX IF NOT EXISTS idx_payments_provider_id ON payments(provider_id);
