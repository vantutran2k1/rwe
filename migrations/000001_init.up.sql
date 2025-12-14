CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    api_key_hash TEXT NOT NULL,
    created_at timestamptz DEFAULT now(),
    billing_plan TEXT,
    quota JSONB DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id),
    name TEXT NOT NULL,
    version INT DEFAULT 1,
    definition JSONB NOT NULL,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    archived BOOLEAN DEFAULT false
);

CREATE TABLE IF NOT EXISTS workflow_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id),
    workflow_id UUID REFERENCES workflows(id),
    status TEXT NOT NULL,
    started_at timestamptz DEFAULT now(),
    finished_at timestamptz,
    payload JSONB,
    metadata JSONB
);

CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID REFERENCES workflow_runs(id),
    step_id TEXT NOT NULL,
    status TEXT NOT NULL,
    worker_id UUID,
    attempts INT DEFAULT 0,
    last_error TEXT,
    started_at timestamptz,
    finished_at timestamptz,
    result JSONB
);

CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID,
    event_type TEXT,
    aggregate_id UUID,
    payload JSONB,
    created_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS workers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT,
    version TEXT,
    last_heartbeat timestamptz,
    capacity JSONB,
    metadata JSONB
);

CREATE TABLE IF NOT EXISTS usage_records (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID,
    metric TEXT,
    value NUMERIC,
    sample_at timestamptz DEFAULT now()
);
