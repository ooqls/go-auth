
-- name: GetPermissions :many
SELECT * FROM permissionsv1 ORDER BY permission LIMIT $1 OFFSET $2;

-- name: GetPermissionById :one
SELECT * FROM permissionsv1 WHERE permission = $1;

-- name: GetPermissionByName :one
SELECT * FROM permissionsv1 WHERE permission = $1;

-- name: CreatePermission :one
INSERT INTO permissionsv1 (
  permission
) VALUES (
  $1
) RETURNING *;

-- name: GetOrCreatePermission :one
INSERT INTO permissionsv1 (permission)
VALUES ($1)
ON CONFLICT (permission) DO NOTHING
RETURNING *;

-- name: DeletePermission :exec
DELETE FROM permissionsv1 WHERE permission = $1;

-- name: HasPermission :one
SELECT EXISTS (
  SELECT 1
  FROM rolebindingsv1 rb
  INNER JOIN permissionbindingsv1 pb ON rb.role_id = pb.role_id
  INNER JOIN permissionsv1 p ON pb.permission = p.permission
  WHERE rb.user_id = $1
    AND (p.permission = $2 OR p.permission = '*')
);

-- name: GetPermissionsForUser :many
SELECT
  p.permission AS permission
FROM rolebindingsv1 rb
INNER JOIN permissionbindingsv1 pb ON rb.role_id = pb.role_id
INNER JOIN permissionsv1 p ON pb.permission = p.permission
WHERE rb.user_id = $1;

-- name: GetPermissionsForUserByGroup :many
SELECT
  p.permission AS permission
FROM rolebindingsv1 rb
INNER JOIN permissionbindingsv1 pb ON rb.role_id = pb.role_id
INNER JOIN permissionsv1 p ON pb.permission = p.permission
WHERE rb.user_id = $1;

