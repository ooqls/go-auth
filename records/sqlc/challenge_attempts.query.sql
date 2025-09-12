-- name: GetChallengeAttempts :many
SELECT id, challenge_id, user_id, success, created_at
FROM authv1_challenge_attempts
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CreateChallengeAttempt :exec
INSERT INTO authv1_challenge_attempts (challenge_id, user_id, success) VALUES ($1, $2, $3);

-- name: GetFailedAttempts :many
SELECT id, challenge_id, user_id, success, created_at
FROM authv1_challenge_attempts
WHERE user_id = $1 AND success = FALSE AND created_at >= NOW() - (sqlc.arg(minutes) || ' minutes')::interval
ORDER BY created_at DESC;