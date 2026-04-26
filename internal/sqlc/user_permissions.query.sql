--- name: GetPermissionsForUser :many
SELECT u.id as userId,
       u.username,
       p.id as permission_id,
       r.id as role_id,
       p.permission,
       r.role_name as role_name,
       r.role_hierarchy as role_hierarchy FROM authv1_users
  LEFT JOIN authv1_roles r ON p.role_id = r.id
  WHERE u.id = $1;
