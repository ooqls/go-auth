--- name: GetPermissionsForUser :many
SELECT u.id as userId,
       u.username,
       p.id as permission_id,
       r.id as role_id,
       p.resource_name,
       p.resource_kind,
       p.resource_group,
       p.actions,
       r.role_name as role_name,
       r.role_hierarchy as role_hierarchy FROM authv1_users u
  LEFT JOIN authv1_roles r ON p.role_id = r.id
  LEFT JOIN authv1_permissions p ON u.role_id = p.role_id
  WHERE u.id = $1;

-- name: GetActionsForUserByResource :many
SELECT p.actions FROM authv1_permissions p
  LEFT JOIN authv1_role_permissions rp ON p.id = rp.permission_id
  LEFT JOIN authv1_user_roles ur ON rp.role_id = ur.role_id
  WHERE ur.user_id = $1 AND p.resource_group = $2 AND p.resource_kind = $3 AND p.resource_name = $4;