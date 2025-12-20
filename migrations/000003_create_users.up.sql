CREATE TABLE IF NOT EXISTS users
(
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL REFERENCES tenants (id),
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    full_name     TEXT,
    role          TEXT             DEFAULT 'admin',
    created_at    timestamptz      DEFAULT now()
);

CREATE INDEX idx_users_email ON users (email);