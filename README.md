# go-auth
> A batteries-included authentication and authorization service for Go teams. Drop it into existing apps as a module, run it as a sidecar/proxy, or deploy it as a standalone identity service backed by Postgres, Redis, and OpenAPI-first contracts.

## Overview
go-auth packages the primitives you need to issue credentials, enforce RBAC, and audit access flows without rebuilding the same stack for every project. The repository hosts:
- HTTP services for authentication, roles, permissions, and resources (`api/v1/*`) built with Gin and generated OpenAPI contracts.
- Domain libraries (`domain/authentication`, `domain/authorization`) that you can import directly into any Go project.
- SQLC-powered data access layers (`authv1/sqlc`, `authv1/gen`) targeting Postgres, plus Redis-backed challenge caches for low-latency auth flows.
- A lightweight dashboard served from `/api/v1/docs` so you can exercise endpoints via an embedded Swagger UI without extra tooling.

Use cases include:
1. **Embedded library** – import `github.com/ooqls/go-auth/domain/v1/...` and compose just the pieces you need inside an existing server.
2. **Auth proxy** – run the Docker image in front of legacy services and rely on cookies/JWTs for upstream identity.
3. **Full identity provider** – deploy both the authentication and role-management APIs and integrate via the published OpenAPI specs.

## Key Capabilities
- **End-to-end login flow** – challenge-based login, registration, refresh, token verification, and cookie management exposed under `/auth/*` (`api/v1/authentication/docs/openapi.yaml`).
- **JWT issuance** – short-lived auth tokens and refresh tokens signed via RSA keys configured in `jwt` blocks (`api/v1/config.go`, `devops/compose/auth/config/config.yaml:1`).
- **Role-based access control** – role, permission, and resource services (see `api/v1/roles`, `api/v1/permissions`, `api/v1/resources`) backed by Postgres migrations in `authv1/sqlc/migrations`.
- **Redis-backed challenges and caching** – challenge secrets plus read-mostly views stored in Redis via `github.com/ooqls/getset/cache`.
- **UI dashboard & docs** – enable docs in config to automatically serve Swagger UI for every OpenAPI doc set (`DocsConfig` in `api/v1/config.go`).
- **Production tooling** – configurable TLS, centralized registry for infra secrets, structured logging (zap), and Makefile targets for multi-arch Docker builds.

## Architecture At A Glance
```
┌───────────┐        ┌─────────────────────┐        ┌──────────────────────┐
│  Clients  │ ─────► │ Gin HTTP Endpoints  │ ─────► │ Domain Auth/Z Logic  │
└───────────┘        │  /auth /roles ...   │        │ (tokens, RBAC, audit)│
                      └─────────┬──────────┘        └──────────┬───────────┘
                                │                              │
                                ▼                              ▼
                       ┌────────────────┐              ┌──────────────────┐
                       │ Postgres (SQLC │◄────────────►│ Redis (challenges│
                       │  migrations)   │              │  + caches)       │
                       └────────────────┘              └──────────────────┘
```

## Repository Layout
- `api/v1/authentication|roles|permissions|resources` – service entrypoints, OpenAPI configs, generated clients/servers.
- `domain/authentication`, `domain/authorization` – reusable logic for token issuance, RBAC evaluation, and operation contexts.
- `authv1/sqlc`, `authv1/gen`, `authv1/types.go` – SQLC configuration, generated models, and helpers for Postgres access.
- `devops/compose` – ready-to-run Docker Compose stack with TLS, JWT keys, migrations, and registry examples.
- `dockerfiles/` – build contexts for `auth`, `roles`, and shared `base` images used by the Makefile targets.
- `cli/` – example consumer that uses the generated API client (`cli/main.go`) to register and log in a user.

## Getting Started

### Prerequisites
- Go **1.24.3+** (module file currently targets Go 1.25, but the code compiles on 1.24.3).
- Docker/Docker Compose (v2.27+) if you prefer containers.
- Access to Postgres 15+ and Redis 7+ (local or managed).
- RSA keypair (`jwt.rsa`, `jwt.pub`) for signing tokens if you override the sample config.

### Run Everything with Docker Compose (preferred)
1. Copy the sample runtime assets:
   ```bash
   cp -r devops/compose/auth devops/runtime-auth
   cp -r devops/compose/roles devops/runtime-roles
   ```
2. Update secrets inside `devops/runtime-auth` (TLS certs, JWT keys, registry credentials).
3. Start the stack:
   ```bash
   cd devops/compose
   docker compose up -d db redis
   docker compose up -d auth roles
   ```
   - Auth API is exposed on `http://localhost:8080`.
   - Roles API is exposed on `http://localhost:8081`.
4. Visit `http://localhost:8080/api/v1/docs` to open the dashboard/Swagger UI and exercise endpoints interactively.

### Run from Source
```bash
# Start Postgres & Redis first (Docker, local installs, or `docker compose up db redis`).
export APP_CONFIG=devops/compose/auth/config/config.yaml

# Authentication API
go run ./api/v1/authentication/main.go --app-config ${APP_CONFIG}

# Roles API (similar flags)
go run ./api/v1/roles/main.go --app-config devops/compose/roles/config/app.yaml
```
- Use `--standalone` if you just want the default in-memory config defined in `api/v1/config.go`.
- `--test-environment` spins up embedded Postgres/Redis containers via go-app’s test harness for integration testing.

### Import as a Go Module (Library or Proxy Mode)
```go
import (
	"github.com/ooqls/go-auth/domain/v1/authentication"
	"github.com/ooqls/go-auth/domain/v1/authorization"
	usersvc "github.com/ooqls/go-auth/domain/v1/serivce/users"
	"github.com/ooqls/go-auth/records/v1"
)

func wireAuthenticator(db *sqlx.DB, cache cache.Factory, issuers authentication.Issuers) authentication.Authenticator {
	factory := v1.NewFactory(*db, cache)
	challenger := authentication.NewChallengerV1(cache.NewStore("challenges", 15*time.Minute))
	return authentication.NewAuthenticatorV1(
		issuers.AuthIssuer,
		issuers.RefreshIssuer,
		cache,
		challenger,
		[]string{"auth"},
	)
}
```
Once wired, expose the handlers you care about (e.g., `AuthenticationServerImpl` in `api/v1/authentication/server.go`) inside your own Gin router or pass-through proxy.

## Configuration
Every binary accepts `--app-config <path>` which points to a go-app YAML file. The sample at `devops/compose/auth/config/config.yaml` illustrates:
- `gin` – port, CORS, and TLS termination for the HTTP server.
- `docs` – enables the dashboard + serves Swagger UI from `/app/docs`.
- `tls` – CA, cert, and key paths for gRPC/HTTP listeners.
- `jwt` – RSA key locations plus per-audience token lifetimes.
- `sql` – SQL migrations and directories applied during startup.
- `registry` – pointer to `registry.yaml` describing Postgres/Redis endpoints (`devops/compose/auth/registry/registry.yaml`).
- `health`, `http`, `rsa` – optional health probes or asymmetric key usage for downstream services.

To customize:
1. Update JDBC-style credentials in `registry.yaml`.
2. Drop migrations into `devops/migrations`; they are auto-applied through the `sql.sql_files_dirs` setting.
3. Supply JWT key pairs in `jwt/` and update token durations per audience.
4. Toggle the docs/dashboard per environment (disable it in production if necessary).

## APIs, Dashboard, and Sample Requests
- OpenAPI specs live under `api/v1/**/docs/openapi.yaml`. Regenerate server/client bindings with `go generate ./api/v1` (see `api/v1/tools.go`) or invoke `oapi-codegen` directly using the provided configs.
- Swagger UI is automatically mounted when `docs.enabled=true` (`/api/v1/docs` by default).
- Example `curl` workflow (assuming localhost defaults):
  ```bash
  # Register a user
  curl -X POST http://localhost:8080/auth/registration \
       -H "Content-Type: application/json" \
       -d '{"username":"demo","email":"demo@example.com","key":"<AES-GCM key derived by password>","secret":"<Username signed with the key>","salt":"<base64 SHA256 of username>"}'

  # Request a login challenge
  curl -X POST http://localhost:8080/auth/login_challenge \
       -H "Content-Type: application/json" \
       -d '{"username":"demo"}'

  # Answer the challenge and receive cookies (OKEY, RKEY, UID)
  curl -X POST http://localhost:8080/auth/login_challenge_response \
       -H "Content-Type: application/json" \
       -d '{"id":"<challenge-id>","challenge":"<challenge encrypted by the key>"}' -i
  ```
- Roles/permissions/resources services expose CRUD endpoints for RBAC state under `/roles/*`, `/permissions/*`, and `/resources/*`. Combine them to create role hierarchies that the authorization package can evaluate at runtime (see `domain/authorization/role_authorizer.go` and related files).

## Data, Migrations, and Tooling
- Database schema migrations live in `authv1/sqlc/migrations` (e.g., `20250224215144_create_tables.sql`).
- SQLC configuration is defined in `authv1/sqlc/sqlc.yaml`; generated models reside in `authv1/gen`.
- Use `sqlc generate -f authv1/sqlc/sqlc.yaml` or `go generate ./...` whenever you update queries.
- Dockerfiles for multi-stage builds live under `dockerfiles/`; invoke `make build_auth`, `make build_roles`, or `make build_all` to cut amd64 images tagged with the current git ref.

## Testing & Verification
```bash
go test ./...               # run unit + integration tests (see authv1/integrationTest/*)
GOMODCACHE=off go test ...  # if you need a fully hermetic run
```
- Integration suites (`authv1/integrationTest`) rely on Postgres/Redis. Use the `--test-environment` flag or Docker Compose services for reproducible runs.
- Static analysis plus OpenAPI/TypeScript generation commands live inside `api/v1/tools.go`.

## Roadmap
Planned features that are already on the backlog:
1. **Multi-tenant auth realms** – isolate users/roles per tenant without duplicating infrastructure.
2. **SQLite support** – developer-friendly storage for local or edge deployments.
3. **Kubernetes & Helm** – production manifests plus customizable Helm charts for cluster rollouts.
4. **Twilio/email hooks** – pluggable notification adapters for MFA and onboarding flows.
5. **First-class dashboard** – richer UI for managing users, roles, and audits beyond Swagger interactions.

## Contributing
Issues and pull requests are welcome! To contribute:
1. Fork the repo and create a feature branch.
2. Add/adjust tests for any behavior changes.
3. Run `go test ./...` and, if relevant, regenerate OpenAPI or SQLC artifacts.
4. Submit a PR that explains the change, configuration impact, and how to validate it.

If you have questions (architecture decisions, configs, roadmap ideas), open a GitHub discussion or issue—feedback is what keeps go-auth growing.
