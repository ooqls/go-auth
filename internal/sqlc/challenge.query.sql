-- name: GetChallenge :one
SELECT id, user_id, challenge, salt, created_at, expires_at
FROM challengesv1
WHERE id = $1 AND expires_at > NOW()
ORDER BY created_at DESC;

-- name: CreateChallenge :one
INSERT INTO challengesv1 (user_id, challenge, salt, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, challenge, salt, created_at, expires_at;