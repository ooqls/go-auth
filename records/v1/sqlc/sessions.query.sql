-- name: GetSession :one
SELECT * FROM authv1_sessions WHERE id = $1;

-- name: ListSessions :many
SELECT * FROM authv1_sessions ORDER BY created_at LIMIT $1 OFFSET $2;

-- name: CreateSession :one
INSERT INTO authv1_sessions (
  user_id,
  token,
  expires_at
) VALUES (
  $1,
  $2,
  $3
) RETURNING *;

-- name: UpdateSession :exec
UPDATE authv1_sessions SET
  user_id = $1,
  token = $2,
  expires_at = $3
WHERE id = $4;

-- name: DeleteSession :exec
DELETE FROM authv1_sessions WHERE id = $1;