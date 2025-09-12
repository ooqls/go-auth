-- name: CreatePermission :one
INSERT INTO authv1_permissions (resource_name, resource_kind, resource_group, actions)
VALUES ($1, $2, $3, $4)
RETURNING id, resource_name, actions;


-- name: GetPermissions :many
SELECT * FROM authv1_permissions ORDER BY updated_at LIMIT $1 OFFSET $2;

-- name: GetPermissionByID :one
SELECT id, resource_name, resource_kind, resource_group, actions
FROM authv1_permissions
WHERE id = $1;

-- name: GetPermissionsByFilter :many
SELECT id, resource_name, resource_kind, resource_group, actions
FROM authv1_permissions
WHERE resource_name = $1 OR resource_kind = $2 OR resource_group = $3
ORDER BY updated_at LIMIT $4 OFFSET $5;

-- name: GetPermissionsByResourceGroup :many
SELECT id, resource_name, resource_kind, resource_group, actions
FROM authv1_permissions
WHERE resource_group = $1
ORDER BY updated_at LIMIT $2 OFFSET $3;

-- name: GetPermissionsByGroupKind :many
SELECT id, resource_name, resource_kind, resource_group, actions
FROM authv1_permissions
WHERE resource_group = $1 AND resource_kind = $2
ORDER BY updated_at LIMIT $3 OFFSET $4;

-- name: GetPermissionsByRoleID :many
SELECT p.id, p.resource_name, p.resource_kind, p.resource_group, p.actions
FROM authv1_permissions p
  LEFT JOIN authv1_role_permissions r ON p.id = r.permission_id
WHERE r.role_id = $1
ORDER BY p.updated_at LIMIT $2 OFFSET $3;



-- name: UpdatePermission :one
UPDATE authv1_permissions
SET resource_name = $1,
    resource_kind = $2,
    resource_group = $3,
    actions = $4,
    updated_at = now()
WHERE id = $5
RETURNING *;

-- name: DeletePermission :exec
DELETE FROM authv1_permissions
WHERE id = $1;
