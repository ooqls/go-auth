-- name: GetRole :one
SELECT * FROM authv1_roles WHERE id = $1;

-- name: GetRoles :many
SELECT * FROM authv1_roles WHERE domain = $1 ORDER BY role_name LIMIT $1 OFFSET $2;

-- name: GetRoleAggregate :many
SELECT roles.role_name, roles.role_hierarchy, perm.*, user_roles.role_id FROM authv1_roles roles
  LEFT JOIN authv1_user_roles user_roles ON roles.id = user_roles.role_id
  LEFT JOIN authv1_role_permissions role_perm ON roles.id = role_perm.role_id
  LEFT JOIN authv1_permissions perm ON role_perm.permission_id = perm.id
WHERE user_roles.user_id = $1;

-- name: CreateRole :one
INSERT INTO authv1_roles (
  role_name,
  description,
  created_at,
  updated_at
) VALUES (
  $1,
  $2,
  $3,
  $4
) RETURNING *;

-- name: UpdateRole :one
UPDATE authv1_roles SET
  role_name = $1,
  description = $2,
  updated_at = now()
WHERE id = $3
RETURNING *;

-- name: DeleteRole :one
DELETE FROM authv1_roles WHERE id = $1 RETURNING *;

-- name: GetRolesByName :many
SELECT * FROM authv1_roles WHERE role_name = $1;
