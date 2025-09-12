-- name: GetRolesForUser :many
SELECT r.* FROM authv1_roles r LEFT JOIN authv1_user_roles u ON r.id = u.role_id WHERE u.user_id = $1;

-- name: AddRoleToUser :exec
INSERT INTO authv1_user_roles (user_id, role_id) VALUES ($1, $2);

-- name: RemoveRoleFromUser :exec
DELETE FROM authv1_user_roles WHERE user_id = $1 AND role_id = $2;