-- +goose Up
-- +goose StatementBegin

CREATE OR REPLACE VIEW user_highest_role AS
SELECT
  rb.user_id AS user_id,
  r.id   AS role_id,
  r.name AS role_name,
  r.hierarchy as hierarchy
FROM rolebindingsv1 rb
JOIN rolesv1 r ON r.id = rb.role_id
WHERE hierarchy = (
  SELECT MAX(r2.hierarchy)
  FROM rolebindingsv1 rb2
  JOIN rolesv1 r2 ON r2.id = rb2.role_id
  WHERE rb2.user_id = rb.user_id
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- DROP VIEW IF EXISTS user_highest_role;

-- +goose StatementEnd
