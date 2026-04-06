
-- name: GetPermissions :many
SELECT * FROM permissionsv1 ORDER BY name LIMIT $1 OFFSET $2;

-- name: GetPermission :one
SELECT * FROM permissionsv1 WHERE name = $1 AND "group" = $2 AND kind = $3;

-- name: CreatePermission :one
INSERT INTO permissionsv1 (
  name,
  "group",
  kind,
  actions,
  created_at,
  updated_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  now(),
  now()
) RETURNING *;

-- name: DeletePermission :exec
DELETE FROM permissionsv1 WHERE name = $1 AND "group" = $2 AND kind = $3;

-- name: HasPermission :one
SELECT EXISTS (
  SELECT 1
  FROM rolebindingsv1 rb
  INNER JOIN permissionbindingsv1 pb ON rb.role_id = pb.role_id
  INNER JOIN permissionsv1 p ON pb.permission_id = p.id
  WHERE rb.user_id = $1
    AND (p.name    = $2 OR p.name    = '*')
    AND (p."group" = $3 OR p."group" = '*')
    AND (p.kind    = $4 OR p.kind    = '*')
    AND (
      p.actions = '*'
      OR p.actions = $5
      OR p.actions LIKE $5 || ',%'
      OR p.actions LIKE '%,' || $5
      OR p.actions LIKE '%,' || $5 || ',%'
    )
) AS has_permission;

-- name: GetPermissionsForUser :many
SELECT
  p.id        AS permission_id,
  p.name      AS permission_name,
  p."group"   AS permission_group,
  p.kind      AS permission_kind,
  p.actions   AS actions
FROM rolebindingsv1 rb
INNER JOIN permissionbindingsv1 pb ON rb.role_id = pb.role_id
INNER JOIN permissionsv1 p ON pb.permission_id = p.id
WHERE rb.user_id = $1
  AND (p."group" = $2 OR p."group" = '*')
  AND (p.kind    = $3 OR p.kind    = '*');

-- name: GetPermissionsForUserByGroup :many
SELECT
  p.id        AS permission_id,
  p.name      AS permission_name,
  p."group"   AS permission_group,
  p.kind      AS permission_kind,
  p.actions   AS actions
FROM rolebindingsv1 rb
INNER JOIN permissionbindingsv1 pb ON rb.role_id = pb.role_id
INNER JOIN permissionsv1 p ON pb.permission_id = p.id
WHERE rb.user_id = $1
  AND (p."group" = $2 OR p."group" = '*');

