-- name: CreateWorkflow :one
INSERT INTO workflows (id, tenant_id, name, definition)
VALUES (gen_random_uuid(), $1, $2, $3)
RETURNING id, tenant_id, name;

-- name: GetWorkflowByID :one
SELECT id,
       tenant_id,
       name,
       version,
       definition,
       created_at,
       updated_at,
       archived
FROM workflows
WHERE id = $1;

-- name: ListWorkflowsByTenantID :many
SELECT id,
       tenant_id,
       name,
       version,
       definition,
       created_at,
       updated_at,
       archived
FROM workflows
WHERE tenant_id = $1
  AND ((updated_at < $2)
    OR (updated_at = $2 AND id < $3))
ORDER BY updated_at DESC
LIMIT $4;
