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
