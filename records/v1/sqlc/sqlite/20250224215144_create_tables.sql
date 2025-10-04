-- +goose Up
-- +goose StatementBegin

-- NOTE: All UUIDs and timestamps must be set in application code, not as SQL defaults, for sqlc compatibility.

CREATE TABLE IF NOT EXISTS authv1_roles (
  id TEXT PRIMARY KEY,
  domain TEXT NOT NULL,
  role_name TEXT NOT NULL,
  role_hierarchy INTEGER NOT NULL,
  description TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE (role_name)
);

CREATE TABLE IF NOT EXISTS authv1_resources (
  resource_group TEXT NOT NULL,
  resource_kind TEXT NOT NULL,
  resource_name TEXT NOT NULL,
  description TEXT,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  PRIMARY KEY (resource_group, resource_kind, resource_name)
);

CREATE TABLE IF NOT EXISTS authv1_permissions (
  id TEXT PRIMARY KEY,
  resource_kind TEXT NOT NULL,
  resource_group TEXT NOT NULL,
  resource_name TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  actions TEXT NOT NULL,
  FOREIGN KEY (resource_group, resource_kind, resource_name) REFERENCES authv1_resources (resource_group, resource_kind, resource_name)
);

CREATE TABLE IF NOT EXISTS authv1_role_permissions (
  role_id TEXT NOT NULL,
  permission_id TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  PRIMARY KEY (role_id, permission_id),
  UNIQUE (role_id, permission_id),
  FOREIGN KEY (role_id) REFERENCES authv1_roles (id),
  FOREIGN KEY (permission_id) REFERENCES authv1_permissions (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS authv1_users (
  id TEXT PRIMARY KEY,
  username TEXT NOT NULL,
  email TEXT NOT NULL,
  key BLOB NOT NULL,
  salt BLOB NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE (username, email)
);

CREATE TABLE IF NOT EXISTS authv1_user_roles (
  user_id TEXT NOT NULL,
  role_id TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  PRIMARY KEY (user_id, role_id),
  UNIQUE (user_id, role_id),
  FOREIGN KEY (user_id) REFERENCES authv1_users (id) ON DELETE CASCADE,
  FOREIGN KEY (role_id) REFERENCES authv1_roles (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS authv1_sessions (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  token TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  expires_at DATETIME NOT NULL,
  UNIQUE (token)
);

CREATE TABLE IF NOT EXISTS authv1_challenges (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  challenge TEXT NOT NULL,
  salt BLOB NOT NULL,
  created_at DATETIME NOT NULL,
  expires_at DATETIME NOT NULL,
  FOREIGN KEY (user_id) REFERENCES authv1_users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS authv1_challenge_attempts (
  id TEXT PRIMARY KEY,
  challenge_id TEXT NOT NULL,
  user_id TEXT NOT NULL,
  success BOOLEAN NOT NULL,
  created_at DATETIME NOT NULL,
  FOREIGN KEY (challenge_id) REFERENCES authv1_challenges (id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES authv1_users (id) ON DELETE CASCADE
);

-- +goose StatementEnd
