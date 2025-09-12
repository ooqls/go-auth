-- name: GetRolePermissions :many
SELECT * FROM authv1_permissions p LEFT JOIN authv1_role_permissions r ON p.id = r.permission_id AND r.role_id = $1;


-- name: AddPermissionToRole :exec
INSERT INTO authv1_role_permissions (role_id, permission_id) VALUES ($1, $2);

-- name: RemovePermissionFromRole :exec
DELETE FROM authv1_role_permissions WHERE role_id = $1 AND permission_id = $2;
