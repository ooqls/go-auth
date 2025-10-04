-- name: GetResourceByID :one
SELECT * FROM authv1_resources WHERE resource_name = $1 AND resource_group = $2 AND resource_kind = $3;

-- name: GetResources :many
SELECT * FROM authv1_resources ORDER BY resource_name LIMIT $1 OFFSET $2;

-- name: CreateResource :one
INSERT INTO authv1_resources (
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

-- name: UpdateResource :exec
UPDATE authv1_resources SET
  resource_name = $1,
  description = $2,
  updated_at = now()
WHERE resource_name = $3 AND resource_group = $4 AND resource_kind = $5;

-- name: DeleteResource :exec
DELETE FROM authv1_resources WHERE resource_name = $1 AND resource_group = $2 AND resource_kind = $3;

