-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS usersv1 (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
  username VARCHAR(32) NOT NULL,
  email VARCHAR(128) NOT NULL,
  key BYTEA NOT NULL,
  salt BYTEA NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  UNIQUE (username, email) -- Ensure usernames are unique
);

CREATE TABLE IF NOT EXISTS rolesv1 (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
  name VARCHAR(32) NOT NULL,
  hierarchy INT NOT NULL DEFAULT 0,
  description TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now ()
);

CREATE TABLE IF NOT EXISTS permissionsv1 (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(32) NOT NULL,
  "group" VARCHAR(32) NOT NULL,
  kind VARCHAR(32) NOT NULL,
  actions TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  UNIQUE (name, "group", kind) -- Ensure permission uniqueness
);

CREATE TABLE IF NOT EXISTS permissionbindingsv1 (
  role_id uuid NOT NULL REFERENCES rolesv1 (id) ON DELETE CASCADE,
  permission_id uuid NOT NULL REFERENCES permissionsv1 (id) ON DELETE CASCADE,
  PRIMARY KEY (role_id, permission_id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  UNIQUE (role_id, permission_id) -- Ensure role-permission pairs are unique
);

CREATE TABLE IF NOT EXISTS rolebindingsv1 (
  user_id uuid NOT NULL REFERENCES usersv1 (id) ON DELETE CASCADE,
  role_id uuid NOT NULL REFERENCES rolesv1 (id) ON DELETE CASCADE,
  PRIMARY KEY (role_id, user_id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  UNIQUE (role_id, user_id) -- Ensure role_id and user_id are unique
);


CREATE TABLE IF NOT EXISTS resourcesv1 (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
  resource_group VARCHAR(32) NOT NULL,
  resource_kind VARCHAR(32) NOT NULL,
  resource_name VARCHAR(32) NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS challengesv1 (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
  user_id uuid NOT NULL REFERENCES usersv1 (id) ON DELETE CASCADE,
  challenge TEXT NOT NULL,
  salt BYTEA NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS challengeattemptsv1 (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
  challenge_id uuid NOT NULL REFERENCES challengesv1 (id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES usersv1 (id) ON DELETE CASCADE,
  success BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now ()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- DROP TABLE IF EXISTS permissionbindingsv1;
-- DROP TABLE IF EXISTS permissionsv1;
-- DROP TABLE IF EXISTS rolebindingsv1;
-- DROP TABLE IF EXISTS rolesv1;
-- DROP TABLE IF EXISTS usersv1;
-- DROP TABLE IF EXISTS resourcesv1;
-- DROP TABLE IF EXISTS challengesv1;
-- DROP TABLE IF EXISTS challengeattemptsv1;

-- +goose StatementEnd

