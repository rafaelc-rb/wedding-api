CREATE TABLE IF NOT EXISTS gifts (
    id          TEXT PRIMARY KEY,
    wedding_id  TEXT NOT NULL REFERENCES weddings(id),
    name        TEXT NOT NULL,
    description TEXT DEFAULT '',
    price       REAL NOT NULL,
    image_url   TEXT DEFAULT '',
    category    TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'available',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_gifts_wedding_id ON gifts(wedding_id);
CREATE INDEX IF NOT EXISTS idx_gifts_category ON gifts(wedding_id, category);
CREATE INDEX IF NOT EXISTS idx_gifts_status ON gifts(wedding_id, status);
