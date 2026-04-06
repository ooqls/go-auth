-- name: GetRoleBindingsForUser :many
SELECT * FROM rolebindingsv1 WHERE user_id = $1;


-- name: AddRoleToUser :exec
INSERT INTO rolebindingsv1 (user_id, role_id) VALUES ($1, $2);

-- name: RemoveRoleFromUser :exec
DELETE FROM rolebindingsv1 WHERE user_id = $1 AND role_id = $2;

-- name: UserHasRole :one
SELECT EXISTS(SELECT 1 FROM rolebindingsv1 WHERE user_id = $1 AND role_id = $2);
