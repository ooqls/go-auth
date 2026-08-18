-- name: CreateRole :one
INSERT INTO rolesv1 (
  name,
  description,
  hierarchy
) VALUES (
  $1,
  $2,
  $3
) RETURNING *;

-- name: UpdateRole :one
UPDATE rolesv1 SET
  name = COALESCE(sqlc.narg('name'), name),
  description = COALESCE(sqlc.narg('description'), description),
  hierarchy = COALESCE(sqlc.narg('hierarchy'), hierarchy)
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM rolesv1 WHERE id = $1;

-- name: CountRoles :one
SELECT COUNT(*) FROM rolesv1;

-- name: GetRole :one
SELECT * FROM rolesv1 WHERE id = $1;

-- name: GetRoles :many
SELECT * FROM rolesv1 ORDER BY name LIMIT $1 OFFSET $2;

-- name: GetRolesByName :one
SELECT * FROM rolesv1 WHERE name = $1 ORDER BY name LIMIT 1;



