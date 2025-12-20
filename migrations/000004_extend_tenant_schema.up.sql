CREATE TYPE tenant_status AS ENUM ('active', 'suspended', 'archived', 'pending');

ALTER TABLE tenants
    DROP COLUMN billing_plan;
ALTER TABLE tenants
    DROP COLUMN quota;

ALTER TABLE tenants
    ADD COLUMN parent_id UUID REFERENCES tenants (id);
ALTER TABLE tenants
    ADD COLUMN slug TEXT UNIQUE NOT NULL DEFAULT '-';
ALTER TABLE tenants
    ADD COLUMN domain TEXT UNIQUE;
ALTER TABLE tenants
    ADD COLUMN status tenant_status DEFAULT 'active';
ALTER TABLE tenants
    ADD COLUMN region TEXT DEFAULT 'us-east-1';
ALTER TABLE tenants
    ADD COLUMN tier TEXT DEFAULT 'free';
ALTER TABLE tenants
    ADD COLUMN settings JSONB DEFAULT '{}';
ALTER TABLE tenants
    ADD COLUMN contact_email TEXT;
ALTER TABLE tenants
    ADD COLUMN updated_at timestamptz DEFAULT now();

CREATE INDEX idx_tenants_slug ON tenants (slug);