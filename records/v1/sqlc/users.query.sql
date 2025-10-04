-- name: GetUser :one
SELECT * FROM authv1_users WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM authv1_users WHERE username = $1;

-- name: ListUsers :many
SELECT * FROM authv1_users ORDER BY username LIMIT $1 OFFSET $2;

-- name: SearchUsers :many
SELECT * FROM authv1_users WHERE username ILIKE $1 ORDER BY username LIMIT $2 OFFSET $3;

-- name: CreateUser :one
INSERT INTO authv1_users (
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
UPDATE authv1_users SET
  username = $1,
  email = $2,
  key = $3,
  updated_at = now()
WHERE id = $4;

-- name: DeleteUser :exec
DELETE FROM authv1_users WHERE id = $1;
