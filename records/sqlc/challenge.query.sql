-- name: GetChallenge :one
SELECT id, user_id, challenge, salt, created_at, expires_at
FROM authv1_challenges
WHERE id = $1 AND expires_at > NOW()
ORDER BY created_at DESC;

-- name: CreateChallenge :one
INSERT INTO authv1_challenges (user_id, challenge, salt, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, challenge, salt, created_at, expires_at;