-- +goose Up
-- +goose StatementBegin
START TRANSACTION;
--   role_id uuid NOT NULL REFERENCES authv1_roles (id),
--   permission_id uuid NOT NULL REFERENCES authv1_permissions (id) ON DELETE CASCADE,
--   PRIMARY KEY (role_id, permission_id),
--   created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
--   updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
--   UNIQUE (role_id, permission_id) -- Ensure role_id and permission_id are unique
-- );


CREATE TABLE IF NOT EXISTS authv1_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
  user_id uuid NOT NULL,
  token TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  expires_at TIMESTAMPTZ NOT NULL,
  UNIQUE (token)
);

COMMIT;

-- +goose StatementEnd
