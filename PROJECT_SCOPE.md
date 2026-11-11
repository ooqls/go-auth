# Project Scope

Status: agreed product requirements, not a statement of implementation completeness.

## Purpose

This project is the central authentication and authorization authority for applications controlled by its operator. It establishes who a user is and decides whether that user may perform an action on a resource.

Applications integrate through an HTTP API using clients generated from the published OpenAPI specification. Applications retain their business data and enforce access decisions before performing protected operations.

The initial goal is reliable integration with one application: an administrator provisions a user, the user signs in, and the application's backend checks permission before acting on a protected resource. One deployment may serve multiple applications within a single security realm.

## Guidance for models and contributors

- Use this file as the source of truth for intended product scope. Use `AGENTS.md` for repository-specific development guidance and OpenAPI specifications for existing API contracts.
- Distinguish required behavior, implemented behavior, and verified behavior. Existing handlers, schemas, or documentation do not demonstrate that a requirement is complete.
- Where existing code or documents conflict with this scope, identify the discrepancy before making changes. This document does not authorize an unrequested rewrite, API break, or credential migration.
- Obtain explicit approval before expanding scope. Deferred capabilities are not implicit requirements.
- Resolve the open design choices below before implementing behavior that depends on them. Record the decision rather than silently treating an assumption as a requirement.

## Responsibility boundary

| This service owns | Applications own |
| --- | --- |
| Security identities and account status | Business profiles and business objects |
| Credentials and authenticated sessions | Browser-facing application sessions/cookies |
| Roles, permission grants, and assignments | Resource types, identifiers, and available actions |
| Access decisions and administrative delegation safeguards | Enforcement of decisions against the exact operation and target |
| Security audit events | Ownership rules and other business conditions |

Resource metadata held here identifies security targets; it is not a copy of application business data. Business profiles, subscriptions, and automatic business-data synchronization are outside this service's responsibility.

## Required behavior

### HTTP integration and trust

- HTTP is the supported integration boundary. The OpenAPI specification must accurately describe supported operations, authentication, request/response schemas, and errors so other applications can generate clients.
- Application backends act as intermediaries between browsers and this service. Browsers receive secure, HTTP-only application cookies; backends use the authenticated user's service session credential when calling this API on that user's behalf.
- Establish user identity from validated credentials. Arbitrary caller-supplied user IDs or unsigned user-ID headers are not proof of identity.
- An application must enforce the returned authorization decision before performing the protected operation. Authentication alone is insufficient.
- Go embedding and authentication proxies are optional integrations, not equivalent primary product boundaries.

### Accounts and credentials

- Administrators create accounts. Public self-registration is outside the initial release.
- Support password login over HTTPS using salted Argon2id password hashing with reviewed libraries. Use configurable, documented hashing parameters appropriate to deployment resources.
- Passwords may be received transiently for verification; they must not be persisted in plaintext, reversibly stored, or logged. Keeping the password hidden from the authentication server is not a requirement.
- Replace preservation of custom challenge-response cryptography as a product goal with established authentication mechanisms. The exact migration from existing credentials remains a separate design decision.
- Support authenticated credential changes and an audited administrator-assisted recovery process.
- Support disabling accounts. Disabled users cannot authenticate or continue using existing sessions.
- Include login rate limiting and abuse protections. Recovery must not introduce an unauthenticated bypass.

### Sessions

- Support multiple concurrent sessions per user, logout of the current session, and logout of all sessions.
- Credential resets and account disabling revoke all of the user's sessions.
- Enforce configurable idle and absolute session expiration.
- Logout and revocation must invalidate server-side session authority, not merely clear browser cookies.
- Credentials, session secrets, and usable recovery secrets must not appear in logs or audit records.
- Session/token format and any renewal mechanism are implementation choices subject to these guarantees; JWTs are not a product requirement.

### Authorization

- Deny access by default. A matching permission grant is required to allow an operation.
- Roles group additive permission grants; users receive permissions through role assignments. The initial model has neither explicit deny rules nor role inheritance.
- Identify permission targets by application namespace, resource type, action, and optional resource-instance identifier.
- Support resource-type grants and grants limited to a particular instance. Application namespaces must prevent an identically named resource in another application from matching accidentally.
- Applications define resource identifiers and actions. Only authorized administrators manage grants and assignments.
- Evaluate the exact target requested by the application. Missing matching grants deny access.
- Ownership and other business conditions remain application responsibilities. No implicit owner permissions or general-purpose policy language is required.

### Administrative delegation

- Administrative hierarchy limits management authority; it does not imply application permissions or role inheritance.
- Administrators may manage only lower-ranked users and roles, subject to their administrative permissions.
- Administrators may grant only permissions they themselves possess. Administrative mutation paths must preserve this restriction, including role changes and assignments.
- Protect the last active bootstrap administrator from removal or disabling.
- Initial bootstrap and recovery from administrative lockout require an explicit design; ordinary API callers must not acquire bootstrap privileges implicitly.

### Freshness and failure handling

- Application backends perform online authorization checks for protected operations.
- After an access-changing update succeeds, the next authorization check must reflect that update. Internal caches must preserve this guarantee across service instances.
- Session revocation and account disabling must also take effect on the next relevant validation after a successful update.
- Fail closed when identity or permission cannot be established. Distinguish service failure from an ordinary permission denial in the API while granting access in neither case.
- Application-side decision caching and offline authorization are deferred. A previously returned decision does not authorize future operations indefinitely.
- The initial guarantee concerns subsequent checks, not retroactive cancellation of operations already authorized or in progress.

### Security auditing

- Record important authentication, session, recovery, account, role, permission, and assignment events, including relevant failures and rejected administrative actions.
- Events must identify the actor when known, action, target, time, and outcome without recording usable secrets.
- Audit storage, retention, access controls, and failure behavior require an explicit implementation design. The presence of logging scaffolding alone does not satisfy auditing requirements.

## Initial release acceptance criteria

The initial release is complete only when automated tests and an application integration example demonstrate all of the following:

1. A client generated from OpenAPI can perform the supported authentication and authorization workflow without importing internal Go packages.
2. An administrator can provision a user, configure a role and grants, and assign that role within the delegation rules.
3. Valid credentials establish a session; invalid credentials do not. Disabled accounts cannot sign in.
4. A user with a matching grant can perform an allowed action. An unassigned user, wrong application namespace, wrong action, and wrong resource instance are denied.
5. Resource-type and instance-specific grants behave distinctly and consistently.
6. Equal- or higher-ranked administration and grants exceeding the actor's authority are rejected. The last active bootstrap administrator remains protected.
7. Current-session logout leaves other sessions intact; logout-all, credential reset, and account disabling invalidate every affected session on the next validation.
8. Idle and absolute session expiration are enforced.
9. Grant removal and account/security changes are visible on the next check after a successful update, including with caching and multiple service instances.
10. Service or dependency failures cannot produce an allow decision without established authority.
11. Security events are auditable, rate limiting is exercised, and logs/audit records exclude credentials and session secrets.
12. OpenAPI schemas, documented errors, and runtime behavior agree for the initial supported API surface.

These criteria describe required evidence, not tests that have already passed.

## Explicit non-goals for the initial release

- Public registration, email verification, automated email recovery, and invitations.
- Organization management, tenant-specific administrators, multiple realms within one deployment, and cross-tenant sharing.
- Independent service-account authentication and machine-to-machine grants.
- MFA, passkeys, OPAQUE/PAKE login, social login, and identity federation such as OAuth/OIDC/SAML provider features.
- Shared-browser SSO and direct cross-domain browser integration.
- Explicit deny policies, inherited roles, automatic ownership permissions, and a general policy language.
- Offline authorization and application-side authorization-decision caching.
- A custom management dashboard; administrative workflows must be available through the API.
- Public-product packaging as a driver of the initial requirements.

Deferred means outside the initial commitment, not permanently prohibited.

## Open implementation decisions

Resolve these separately without changing the agreed responsibility boundary:

- Session credential format, transport from backends, renewal, expiration defaults, and secure handling of browser-to-backend sessions.
- Credential migration from the current authentication mechanism and exact credential-change/recovery procedures.
- Numeric rank direction, bootstrap privileges, safe initial provisioning, and administrator recovery.
- Exact permission syntax, wildcard/containment rules, resource lifecycle handling, and administrative API permissions.
- Audit persistence/retention, audit failure behavior, rate-limit policies, and cache invalidation/concurrency mechanisms.
- API versioning and compatibility policy, deployment readiness criteria, and operational availability/performance targets.

## Relationship to the current repository

The repository contains custom challenge-response authentication, token/session-related code, hierarchical RBAC, resource metadata, and OpenAPI contracts. Those are implementation starting points, not requirements to preserve unchanged.

`README.md`, `model.md`, and existing APIs may describe broader or different behavior. Before each implementation task, inspect the relevant code and tests against this scope and report gaps. This scope definition does not certify production readiness or completeness of any current feature.
