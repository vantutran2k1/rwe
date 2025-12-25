-- name: CreateTenant :one
INSERT INTO tenants (name, slug, contact_email, tier, region, status)
VALUES ($1, $2, $3, $4, $5, 'active')
RETURNING *;

-- name: GetTenantBySlug :one
SELECT id
FROM tenants
WHERE slug = $1;

-- name: AddTenantMember :one
INSERT INTO tenant_members (tenant_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING *;