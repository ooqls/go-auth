-- name: GetResourceByID :one
SELECT * FROM resourcesv1 WHERE resource_name = $1 AND resource_group = $2 AND resource_kind = $3;


-- name: GetResources :many
SELECT * FROM resourcesv1 WHERE resource_group = $1 AND ($2::text = '*' OR resource_kind = $2) ORDER BY resource_name LIMIT $3 OFFSET $4;

-- name: GetResourcesById :one
SELECT * FROM resourcesv1 WHERE id = $1;

-- name: CreateResource :one
INSERT INTO resourcesv1 (
  resource_kind,
  resource_group,
  resource_name,
  description,
  created_at,
  updated_at
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
  resource_name = COALESCE(NULLIF(@name, ''), resource_name),
  description = COALESCE(NULLIF(@description, ''), description),
  updated_at = now()
WHERE id = @id
RETURNING *;

-- name: DeleteResource :exec
DELETE FROM resourcesv1 WHERE resource_name = $1 AND resource_group = $2 AND resource_kind = $3;

-- name: DeleteResourceById :exec
DELETE FROM resourcesv1 WHERE id = $1;
