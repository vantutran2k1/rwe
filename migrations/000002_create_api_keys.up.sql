ALTER TABLE tenants
    DROP COLUMN api_key_hash;

CREATE TABLE IF NOT EXISTS api_keys
(
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    key_hash     TEXT NOT NULL,
    key_prefix   TEXT NOT NULL,
    name         TEXT,
    created_at   timestamptz      DEFAULT now(),
    expires_at   timestamptz,
    revoked      BOOLEAN          DEFAULT false,
    last_used_at timestamptz
);

CREATE INDEX idx_api_keys_hash ON api_keys (key_hash);