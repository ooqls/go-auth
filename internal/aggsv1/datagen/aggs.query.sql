-- name: GetUserAggregate :many
SELECT
  u.id          AS user_id,
  u.username    AS username,
  u.email       AS email,
  u.created_at  AS user_created_at,
  u.updated_at  AS user_updated_at,
  r.id          AS role_id,
  r.name        AS role_name,
  r.hierarchy   AS hierarchy,
  r.description AS role_description,
  r.created_at  AS role_created_at,
  r.updated_at  AS role_updated_at,
  p.permission     AS permission
FROM usersv1 u
LEFT JOIN rolebindingsv1 rb ON u.id = rb.user_id
LEFT JOIN rolesv1 r ON rb.role_id = r.id
LEFT JOIN permissionbindingsv1 pb ON r.id = pb.role_id
LEFT JOIN permissionsv1 p ON pb.permission = p.permission
WHERE u.id = $1;

-- name: GetRoleAggregate :many
SELECT
  r.id          AS role_id,
  r.name        AS role_name,
  r.hierarchy   AS hierarchy,
  r.description AS description,
  r.created_at  AS role_created_at,
  r.updated_at  AS role_updated_at,
  p.permission     AS permission
FROM rolesv1 r
INNER JOIN permissionbindingsv1 pb ON r.id = pb.role_id
INNER JOIN permissionsv1 p ON pb.permission = p.permission
WHERE r.id = $1;


-- name: CompareUserRoleHierarchy :one
SELECT SIGN(actor.hierarchy - target.hierarchy)::int
FROM user_highest_role actor
JOIN user_highest_role target ON target.user_id = $2
WHERE actor.user_id = $1;

-- name: GetUserHierarchy :one
SELECT hierarchy FROM user_highest_role WHERE user_id = $1;