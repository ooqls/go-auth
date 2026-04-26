-- name: GetPermissionsForRole :many
SELECT * FROM permissionbindingsv1 WHERE role_id = $1;

-- name: GetPermissions :many
SELECT * FROM permissionbindingsv1 LIMIT $1 OFFSET $2;

-- name: AssignPermission :exec
WITH upserted_permission AS (
  INSERT INTO permissionsv1 (permission)
  VALUES ($1)
  ON CONFLICT (permission) DO UPDATE
    SET permission = EXCLUDED.permission
)
INSERT INTO permissionbindingsv1 (role_id, permission)
VALUES ($2, $1)
ON CONFLICT (role_id, permission) DO NOTHING;

-- name: UnassignPermission :exec
DELETE FROM permissionbindingsv1 WHERE permission = $1 AND role_id = $2;

-- name: UnassignAllPermissions :exec
DELETE FROM permissionbindingsv1 WHERE role_id = $1;

-- name: UnassignFromAllRoles :exec
DELETE FROM permissionbindingsv1 WHERE permission = $1;