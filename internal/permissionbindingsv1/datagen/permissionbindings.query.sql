-- name: GetPermissionsForRole :many
SELECT * FROM permissionbindingsv1 WHERE role_id = $1;

-- name: GetPermissions :many
SELECT * FROM permissionbindingsv1 LIMIT $1 OFFSET $2;

-- name: AssignPermission :exec
INSERT INTO permissionbindingsv1 (
    role_id,
    permission_id,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    now(),
    now()
);

-- name: UnassignPermission :exec
DELETE FROM permissionbindingsv1 WHERE permission_id = $1 AND role_id = $2;

-- name: UnassignAllPermissions :exec
DELETE FROM permissionbindingsv1 WHERE role_id = $1;

-- name: UnassignFromAllRoles :exec
DELETE FROM permissionbindingsv1 WHERE permission_id = $1;