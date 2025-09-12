-- +goose Up
-- +goose StatementBegin
START TRANSACTION;

CREATE TABLE IF NOT EXISTS authv1_roles (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
  domain TEXT NOT NULL,
  role_name TEXT NOT NULL,
  role_hierarchy INT NOT NULL DEFAULT 0,
  description TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  UNIQUE (role_name) -- Ensure role names are unique
);

CREATE TABLE IF NOT EXISTS authv1_resources (
  resource_group VARCHAR(64) NOT NULL,
  resource_kind VARCHAR(64) NOT NULL,
  resource_name VARCHAR(64) NOT NULL,
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (resource_group, resource_kind, resource_name)
);

CREATE TABLE IF NOT EXISTS authv1_permissions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
  resource_kind VARCHAR(64) NOT NULL,
  resource_group VARCHAR(64) NOT NULL,
  resource_name VARCHAR(64) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  actions TEXT NOT NULL,
  FOREIGN KEY (resource_group, resource_kind, resource_name) REFERENCES authv1_resources (resource_group, resource_kind, resource_name)
);

CREATE TABLE IF NOT EXISTS authv1_role_permissions (
  role_id uuid NOT NULL REFERENCES authv1_roles (id),
  permission_id uuid NOT NULL REFERENCES authv1_permissions (id) ON DELETE CASCADE,
  PRIMARY KEY (role_id, permission_id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  UNIQUE (role_id, permission_id) -- Ensure role_id and permission_id are unique
);

CREATE TABLE IF NOT EXISTS authv1_users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
  username TEXT NOT NULL,
  email TEXT NOT NULL,
  key BYTEA NOT NULL,
  salt BYTEA NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  UNIQUE (username, email) -- Ensure usernames are unique
);

CREATE TABLE IF NOT EXISTS authv1_user_roles (
  user_id uuid NOT NULL REFERENCES authv1_users (id) ON DELETE CASCADE,
  role_id uuid NOT NULL REFERENCES authv1_roles (id) ON DELETE CASCADE,
  PRIMARY KEY (user_id, role_id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  UNIQUE (user_id, role_id) -- Ensure user_id and role_id are unique
);

CREATE TABLE IF NOT EXISTS authv1_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
  user_id uuid NOT NULL,
  token TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  expires_at TIMESTAMPTZ NOT NULL,
  UNIQUE (token)
);

CREATE TABLE IF NOT EXISTS authv1_challenges (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
  user_id uuid NOT NULL REFERENCES authv1_users (id) ON DELETE CASCADE,
  challenge TEXT NOT NULL,
  salt BYTEA NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS authv1_challenge_attempts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
  challenge_id uuid NOT NULL REFERENCES authv1_challenges (id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES authv1_users (id) ON DELETE CASCADE,
  success BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now ()
);

COMMIT;

-- +goose StatementEnd
