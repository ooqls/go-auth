-- name: CompareUserRoleHierarchy :one
SELECT SIGN(actor.hierarchy - target.hierarchy)::int
FROM user_highest_role actor
JOIN user_highest_role target ON target.user_id = $2
WHERE actor.user_id = $1;