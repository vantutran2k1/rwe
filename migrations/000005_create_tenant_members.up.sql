ALTER TABLE users
    DROP COLUMN tenant_id;
ALTER TABLE users
    DROP COLUMN role;

CREATE TABLE tenant_members
(
    tenant_id UUID REFERENCES tenants (id) ON DELETE CASCADE,
    user_id   UUID REFERENCES users (id) ON DELETE CASCADE,
    role      TEXT        DEFAULT 'member',
    joined_at timestamptz DEFAULT now(),
    PRIMARY KEY (tenant_id, user_id)
);

CREATE INDEX idx_tenant_members_user ON tenant_members (user_id);