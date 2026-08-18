
-- name: GetPermissions :many
SELECT * FROM permissionsv1 ORDER BY permission LIMIT $1 OFFSET $2;

-- name: GetPermissionById :one
SELECT * FROM permissionsv1 WHERE permission = $1;

-- name: GetPermissionByName :one
SELECT * FROM permissionsv1 WHERE permission = $1;

-- name: CountPermissions :one
SELECT COUNT(*) FROM permissionsv1;

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

-- name: HasResourcePermission :one
-- Stored permissions are encoded as "group:kind:action". A stored
-- permission may use '*' in any component as a wildcard. The request
-- (group, kind, action) is matched component-by-component, where a
-- stored '*' matches anything. A bare "*" permission grants everything.
SELECT EXISTS (
  SELECT 1
  FROM rolebindingsv1 rb
  INNER JOIN permissionbindingsv1 pb ON rb.role_id = pb.role_id
  INNER JOIN permissionsv1 p ON pb.permission = p.permission
  WHERE rb.user_id = sqlc.arg(user_id)
    AND (
      p.permission = 'resources:*'
      OR (
        (split_part(p.permission, ':', 1) = 'resource' OR split_part(p.permission, ':', 1) = '*')
        AND (split_part(p.permission, ':', 2) = sqlc.arg(rGroup) OR split_part(p.permission, ':', 2) = '*')
        AND (split_part(p.permission, ':', 3) = sqlc.arg(kind) OR split_part(p.permission, ':', 3) = '*')
        AND (split_part(p.permission, ':', 4) = sqlc.arg(action) OR split_part(p.permission, ':', 4) = '*')
      )
    )
);

-- name: HasCorePermission :one
-- Stored permissions are encoded as "group:kind:action". A stored
-- permission may use '*' in any component as a wildcard. The request
-- (group, kind, action) is matched component-by-component, where a
-- stored '*' matches anything. A bare "*" permission grants everything.
SELECT EXISTS (
  SELECT 1
  FROM rolebindingsv1 rb
  INNER JOIN permissionbindingsv1 pb ON rb.role_id = pb.role_id
  INNER JOIN permissionsv1 p ON pb.permission = p.permission
  WHERE rb.user_id = sqlc.arg(user_id)
    AND (
      p.permission = 'core:*'
      OR (
        (split_part(p.permission, ':', 1) = 'core' OR split_part(p.permission, ':', 1) = '*')
        AND (split_part(p.permission, ':', 2) = sqlc.arg(rGroup) OR split_part(p.permission, ':', 2) = '*')
        AND (split_part(p.permission, ':', 3) = sqlc.arg(kind) OR split_part(p.permission, ':', 3) = '*')
        AND (split_part(p.permission, ':', 4) = sqlc.arg(action) OR split_part(p.permission, ':', 4) = '*')
      )
    )
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

-- name: SearchPermissions :many
SELECT
  p.permission AS permission
FROM permissionsv1 p
WHERE p.permission LIKE $1;
