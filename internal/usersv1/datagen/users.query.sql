-- name: GetUser :one
SELECT * FROM usersv1 WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM usersv1 WHERE username = $1;

-- name: ListUsers :many
SELECT * FROM usersv1 ORDER BY username LIMIT $1 OFFSET $2;

-- name: SearchUsers :many
SELECT * FROM usersv1 WHERE username ILIKE $1 ORDER BY username LIMIT $2 OFFSET $3;

-- name: CreateUser :one
INSERT INTO usersv1 (
  id,
  username,
  email,
  salt,
  key
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5
) RETURNING *;

-- name: UpdateUser :exec
UPDATE usersv1 SET
  username = COALESCE(sqlc.narg('username'), username),
  email = COALESCE(sqlc.narg('email'), email),
  key = COALESCE(sqlc.narg('key'), key),
  updated_at = now()
WHERE id = $1 RETURNING *;

-- name: DeleteUser :exec
DELETE FROM usersv1 WHERE id = $1;


-- name: GetUserAgg :many
-- SELECT users.id as user_id, 
--   users.username as username, 
--   users.email as email, 
--   users.created_at as created_at, 
--   users.updated_at as updated_at, 
--   roles.id as role_id, 
--   roles.name as name, 
--   roles.hierarchy as hierarchy, 
--   roles.permissions as permissions FROM usersv1 users
--   LEFT JOIN authv1_role_bindings role_bindings ON users.id = role_bindings.user_id
--   LEFT JOIN authv1_roles roles ON role_bindings.role_id = roles.id
-- WHERE users.id = $1;