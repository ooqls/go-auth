-- name: GetResourceByName :one
SELECT * FROM resourcesv1 WHERE name = $1;

-- name: GetResources :many
SELECT * FROM resourcesv1 ORDER BY name LIMIT $1 OFFSET $2;

-- name: GetResourceByID :one
SELECT * FROM resourcesv1 WHERE id = $1;

-- name: GetResourcesByGroup :many
SELECT * FROM resourcesv1 WHERE rGroup = $1 ORDER BY name LIMIT $2 OFFSET $3;

-- name: GetResourcesByGroupAndKind :many
SELECT * FROM resourcesv1 WHERE rGroup = $1 AND kind = $2 ORDER BY name LIMIT $3 OFFSET $4;

-- name: CountAllResources :one
SELECT COUNT(*) FROM resourcesv1;

-- name: CountResourcesByGroup :one
SELECT COUNT(*) FROM resourcesv1 WHERE rGroup = $1;

-- name: CountResources :one
SELECT COUNT(*) FROM resourcesv1 WHERE rGroup = $1 AND kind = $2;

-- name: CreateResource :one
INSERT INTO resourcesv1 (
  name,
  rGroup,
  kind,
  created_at,
  updated_at,
  id
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6
) RETURNING *;

-- name: UpdateResource :one
UPDATE resourcesv1 SET
  name = $4,
  updated_at = now()
WHERE rGroup = $1 AND kind = $2 AND name = $3
RETURNING *;

-- name: DeleteResource :exec
DELETE FROM resourcesv1 WHERE name = $1 and rGroup = $2 and kind = $3;

-- name: DeleteResourceById :exec
DELETE FROM resourcesv1 WHERE id = $1;

-- name: SearchResources :many
SELECT * FROM resourcesv1
WHERE (rGroup = sqlc.narg('rgroup') OR sqlc.narg('rgroup') IS NULL)
  AND (kind = sqlc.narg('kind') OR sqlc.narg('kind') IS NULL)
  AND (name = sqlc.narg('name') OR sqlc.narg('name') IS NULL)
  AND (sqlc.narg('query')::text IS NULL OR name ILIKE '%' || sqlc.narg('query') || '%')
ORDER BY name
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountSearchResources :one
SELECT COUNT(*) FROM resourcesv1
WHERE (rGroup = sqlc.narg('rgroup') OR sqlc.narg('rgroup') IS NULL)
  AND (kind = sqlc.narg('kind') OR sqlc.narg('kind') IS NULL)
  AND (name = sqlc.narg('name') OR sqlc.narg('name') IS NULL)
  AND (sqlc.narg('query')::text IS NULL OR name ILIKE '%' || sqlc.narg('query') || '%');
