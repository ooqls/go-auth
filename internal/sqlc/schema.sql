START TRANSACTION;

CREATE TABLE IF NOT EXISTS authv1_roles (
  id uuid PRIMARY KEY,
  role_name TEXT NOT NULL,
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (role_name) -- Ensure role names are unique
);

CREATE TABLE IF NOT EXISTS authv1_resources (
  id uuid PRIMARY KEY,
  resource_name TEXT NOT NULL,
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  UNIQUE (resource_name) -- Ensure resource names are unique
);

CREATE TABLE IF NOT EXISTS authv1_role_permissions (
  id uuid PRIMARY KEY,
  role_id uuid REFERENCES authv1_roles (id),
  permission_id uuid REFERENCES authv1_permissions (id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  UNIQUE (role_id, permission_id) -- Ensure role_id and permission_id are unique
);

CREATE TABLE IF NOT EXISTS authv1_permissions (
  id uuid PRIMARY KEY,
  resource_id uuid REFERENCES authv1_resources (id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  actions TEXT NOT NULL,
  UNIQUE (role_id, resource_id) -- Ensure role_id and resource_id are unique
);

CREATE TABLE IF NOT EXISTS authv1_users (
  id uuid PRIMARY KEY,
  username TEXT NOT NULL,
  email TEXT NOT NULL,
  pw_hash TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  UNIQUE (username, email) -- Ensure usernames and emails are unique
);

CREATE TABLE IF NOT EXISTS authv1_user_roles (
  id uuid PRIMARY KEY,
  user_id uuid REFERENCES authv1_users (id),
  role_id uuid REFERENCES authv1_roles (id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  UNIQUE (user_id, role_id) -- Ensure user_id and role_id are unique
);

CREATE TABLE IF NOT EXISTS authv1_sessions (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL,
  token TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  expires_at TIMESTAMPTZ NOT NULL,
  UNIQUE (token)
);

CREATE TABLE IF NOT EXISTS authv1_audits (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL,
  action TEXT NOT NULL,
  resource_id INTEGER NOT NULL REFERENCES authv1_resources (id),
  timestamp TIMESTAMPTZ NOT NULL DEFAULT now ()
);

COMMIT;
