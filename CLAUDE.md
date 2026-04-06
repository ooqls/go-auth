# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Tests
go test ./...

# Build Docker images (tagged with current git ref)
make build_all        # base + seed + auth + roles + users
make build_auth       # authentication service only
make build_roles      # roles service only

# Run from source (requires Postgres + Redis)
go run ./api/v1/authentication/main.go --app-config devops/compose/auth/config/config.yaml
go run ./api/v1/roles/main.go --app-config devops/compose/roles/config/app.yaml

# Flags: --standalone (hardcoded defaults), --test-environment (embedded Postgres/Redis via testcontainers)

# Code generation
go generate ./...                            # all generators
sqlc generate -f <path>/sqlc.yaml           # regenerate SQLC for a single module
# oapi-codegen is invoked via go:generate directives in tools.go files
```

## Architecture

This is an authentication + RBAC service with three deployment modes: embedded library, auth proxy, or standalone identity provider.

### Package Layering

```
v1/<entity>/          → HTTP service layer (Gin handlers, OpenAPI server impl)
  <entity>api/        → server.go implements gen_<entity>.ServerInterface
  service.go          → business logic interface + impl
internal/<entity>v1/  → domain types, Reader/Writer interfaces, cache-wrapped impls
  datagen/            → SQLC-generated code (models, queries)
  datagen/sqlc.yaml   → per-module SQLC config (engine: postgresql, sql_package: pgx/v5)
internal/datav1/      → Factory that wires all readers/writers (injected at startup)
v1/base/              → reusable Gin app bootstrap (BuildBaseGinApp[T])
```

**Data flow:** HTTP handler → `v1/<entity>/service.go` → `internal/<entity>v1/` reader/writer → SQLC-generated pgx/v5 queries → Postgres. Redis wraps readers for caching.

### Key Patterns

**ServerInterface compliance:** Every API package declares `var _ gen_<entity>.ServerInterface = (*Server)(nil)` as a compile-time check. The `gen_*` packages are produced by `oapi-codegen` and must not be edited manually.

**Authorization context:** Handlers call `authorizationv1.FromGinContext(c)` to get the `*authorizationv1.Context` (carries logger + auth claims). Return 403 if not present.

**Factory pattern:** `internal/datav1.Factory` is created once at startup and passed to all service constructors. It produces typed readers and writers:
```go
factory.NewUserReader()
factory.NewRoleWriter()
// etc.
```

**Error handling:** Use `v1.GinHandleError(ctx, err)` in handlers. Internal errors are typed (`NotFoundError`, `AlreadyExistsError`, `InternalError`, `ForbiddenError`) and map to HTTP status codes automatically.

**JSONB columns** (e.g., `actions`, `permissions`) are generated as `[]byte` by SQLC. Unmarshal to the appropriate Go type in the service layer; use sqlc `overrides` in `sqlc.yaml` if you need a custom type scanned automatically.

### SQLC Conventions

- Query files: `internal/<entity>v1/datagen/<entity>.query.sql`
- Config: `internal/<entity>v1/datagen/sqlc.yaml` (schema path points to `../../schema`)
- UUID override is always present: `db_type: uuid → github.com/google/uuid.UUID`
- Column aliases are required for aggregate queries to avoid ambiguous names (e.g., `r.id AS role_id`, `p.id AS permission_id`)

### OpenAPI Conventions

- Specs live in `v1/<entity>/<entity>api/docs/openapi.yaml`
- Config in `v1/<entity>/<entity>api/openapi_config.yaml`
- Generated output goes into `v1/<entity>/<entity>api/gen_<entity>/gen.go`
- `tools.go` in each api package carries the `//go:generate` directive

### Module Dependency

`github.com/ooqls/getset` is replaced locally via `go.mod`:
```
replace github.com/ooqls/getset => ../getset
```
This sibling module must be present on disk for builds to succeed.

### Infrastructure

- **DB:** Postgres 15+, migrations via goose (files in `internal/schema/`)
- **Cache:** Redis 7+, accessed through `github.com/ooqls/getset/cache` factory
- **Auth tokens:** RSA-signed JWTs, key paths configured in the `jwt` block of the app config YAML
- **Config format:** go-app YAML (`github.com/ooqls/getset/app`), passed via `--app-config`
