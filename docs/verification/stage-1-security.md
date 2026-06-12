# Stage 1 Security Verification

Status: complete
Stage: 1 — Reliability, navigation, and identity
Independent verifier: replacement for OMX Team task 9 after worker-5 quota exhaustion
Integrated product SHA under test: `8b343880d58a1b3a562a80afc1f84cab666933c3`
Verdict: **FAIL**

## Decision

Stage 1 security does not pass. `RequireAdmin`, token rejection, Reader behavior,
network/auth classification, and Draft File/search/Asset isolation passed, but three
security-contract defects remain:

1. **HIGH — deployment-configured Author elevation preserves the prior Reader
   password.**
2. **MEDIUM — unexpected authentication service errors are returned verbatim to
   unauthenticated clients.**
3. **MEDIUM — a backslash-form return target passes validation and breaks navigation
   after successful login.**

The HIGH finding prevents approval. The two MEDIUM findings also violate explicit
Stage 1 requirements for sanitized errors and safe return targets.

## Findings

### [HIGH] ADMIN_EMAIL elevation does not establish the configured Author credential

Files:

- `api/internal/users/repository.go:72-79`
- `api/internal/auth/service.go:70-86`

`SeedAdmin` hashes `ADMIN_PASSWORD`, but `UpsertAdmin` handles an existing email with
`do update set role = 'admin'` only. It does not update `password_hash` or restore the
configured local provider. A Reader who already controls the configured email is
promoted while retaining the Reader password.

Live disposable-user evidence:

```text
role_after_seed_equivalent=admin
configured_password_applied=f
original_reader_password_login=200
```

This is within the Stage 1 identity/authorization boundary and blocks Stage 1.

Required repair: make the deployment seed authoritative for both role and configured
credential (and provider as applicable), then add a PostgreSQL-backed regression proving
the previous Reader password no longer authenticates after elevation.

### [MEDIUM] Auth handlers expose unexpected internal service errors

File: `api/internal/http/handlers/auth.go:39-46,57-64`

Registration and login return `err.Error()` for every unexpected repository/service
failure. An unauthenticated client can therefore receive database connection,
topology, or SQL details.

A read-only Go overlay probe injected an internal repository error and proved that the
sentinel was copied into the HTTP response:

```text
=== RUN   TestStage1ProbeAuthHandlerLeaksUnexpectedServiceError
--- PASS: TestStage1ProbeAuthHandlerLeaksUnexpectedServiceError
PASS
```

The probe passed only when the response contained
`internal database topology sentinel`.

This violates the requested sanitized-error boundary and blocks Stage 1.

Required repair: map known validation/conflict/credential errors explicitly and return
a generic server error for unexpected failures; preserve detailed diagnostics only in
sanitized server logs.

### [MEDIUM] Backslash return target is accepted and breaks post-login navigation

File: `web/src/lib/auth.ts:15-25`

`sanitizeReturnTo` rejects absolute and `//` URLs but accepts `/\\attacker.example/...`.
Browsers interpret backslashes as URL separators during URL resolution. React Router's
same-origin history guard prevented an actual cross-origin transition in this probe,
but navigation threw after the token had already been stored. The login page remained
open and incorrectly displayed a network failure.

Browser evidence:

```text
stage1-sec-return-backslash
url=http://127.0.0.1:5173/login?return_to=%2F%5C%5Cattacker.example%2Fsteal
body_state=navigation_error
```

Screenshot:

`~/.local/share/xlab-blog/verification/stage1-security-8b34388/backslash-return-navigation-error.png`

This violates the Stage 1 requirement to reject backslash-form return targets and
blocks Stage 1.

Required repair: accept only a normalized same-origin application pathname, reject
backslashes/control characters and auth-loop destinations, and add behavioral tests
for encoded and decoded variants.

## Reassessed prior observations

### UpsertAdmin role elevation by ADMIN_EMAIL

Classification: **HIGH, Stage 1 blocker**.

Promoting the deployment-configured email is specified behavior. Preserving the old
Reader password instead of applying `ADMIN_PASSWORD` is not. The live probe above
demonstrates the account-takeover condition.

### Placeholder production secrets/credentials guard

Files:

- `docker-compose.yml:5-24`
- `api/internal/config/config.go:63-71`

Classification: **HIGH for deployment, not an independent native Stage 1 blocker**.

Observed:

```text
compose_has_active_placeholder_defaults=true
config_has_placeholder_rejection=false
```

No real local credential or bearer token was found in tracked files. The committed
values are placeholders, but Compose activates them as working database, JWT, and
Author credentials when environment variables are absent, and configuration accepts
them.

The active plan forbids Docker/server deployment before later native stages, so this
does not independently fail the native Stage 1 gate. It **must block any production or
Compose deployment claim** until startup fails closed on placeholder/default secrets.

### Multipart request cap

File: `api/internal/http/handlers/assets.go:18,34-55`

Classification: **MEDIUM, not an independent Stage 1 blocker; release hardening
required**.

`ParseMultipartForm(8 << 20)` limits in-memory buffering, not total request size. There
is no `http.MaxBytesReader` envelope. A request containing more than 25 MiB of ignored
multipart padding plus a small valid File was accepted:

```text
request_bytes_gt_25MiB_status=201 accepted=true
```

The endpoint is protected by `RequireAdmin`, and a total multipart envelope is not an
explicit Stage 1 identity requirement, so this observation does not independently
block Stage 1. It must be fixed before treating Asset upload as production-hardened.

## Passed boundaries

### RequireAdmin and token semantics

`api/internal/http/router.go:131-153` places every current Author route behind
`RequireAdmin`.

Live role matrix:

```text
nodes_create             anonymous=401 reader=403
node_get                 anonymous=401 reader=403
node_patch               anonymous=401 reader=403
node_delete              anonymous=401 reader=403
content_put              anonymous=401 reader=403
publish                  anonymous=401 reader=403
unpublish                anonymous=401 reader=403
asset_upload             anonymous=401 reader=403
asset_delete             anonymous=401 reader=403
refresh_embedding        anonymous=401 reader=403
rebuild                  anonymous=401 reader=403
```

`GET /api/auth/me` results:

```text
anonymous=401
wrong_scheme=401
malformed=401
tampered=401
expired=401
role_forged=401
reader_valid=200
author_valid=200
deleted_subject=401
signed_admin_claim_for_database_reader=403
```

The final result proves authorization uses the database-loaded role rather than a
signed but stale/elevated role claim.

### Reader 403 without logout loop

Reader login with `return_to=/admin` ended at `/admin`, retained the Reader session,
and displayed `Author access required` plus `Return to Recent`.

Screenshot:

`~/.local/share/xlab-blog/verification/stage1-security-8b34388/reader-admin-author-required.png`

### Invalid-token versus network-failure classification

Invalid token:

```text
token_cleared=true
login_visible=true
```

Network failure on `/api/auth/me`:

```text
url=http://127.0.0.1:5173/admin
retry_visible=true
token_preserved=true
login_decision_absent=true
```

Screenshots:

- `~/.local/share/xlab-blog/verification/stage1-security-8b34388/invalid-token-anonymous.png`
- `~/.local/share/xlab-blog/verification/stage1-security-8b34388/network-failure-retry-2.png`

### Safe return-target cases that passed

Absolute external, scheme-relative, and recursive Login targets all resolved to the
safe `/recent` default. Anonymous `/admin` redirected to
`/login?return_to=%2Fadmin`.

The backslash variant remains the failing case documented above.

### Draft File, search, interaction, and Asset isolation

A disposable Draft File received a unique Name/body/keyword marker and Draft Asset.
Anonymous/Reader probes observed:

```text
draft_file_resolve_status=404
recent_draft_file_matches=0
search_draft_file_matches=0
draft_asset_anonymous=404
draft_asset_reader=404
draft_file_like_reader=404
draft_file_comment_reader=404
```

Directories do not have Draft state in the current Stage 1 model. Empty Directory
visibility was therefore not classified as Draft File disclosure.

### Credential/error echo checks that passed

A failed login containing a unique password sentinel produced:

```text
bad_login_status=401
response_echoed_password=false
api_log_contains_password=false
api_log_contains_bearer=false
```

The separate raw-unexpected-error finding remains applicable.

## Verification gates

Environment:

```text
Node.js v22.22.3
npm 10.9.8
Go 1.26.4
PostgreSQL client 17.10
```

Backend targeted and full gates:

```text
go test -count=1 ./internal/auth ./internal/http/middleware ./internal/http ./internal/assets ./internal/search ./internal/tree
PASS

CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./...
PASS

CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go vet ./...
PASS

test -z "$(gofmt -l .)"
PASS
```

Frontend/security regression gates:

```text
node --test tests/identity-navigation.test.mjs tests/admin-create-flow.test.mjs
7/7 PASS

node --test tests/render-safety.test.mjs
7/7 PASS

npm run lint
PASS

npm run build
PASS
```

Type diagnostics:

- `npx tsc --noEmit` reported zero errors.
- Per-file diagnostics reported zero TypeScript errors for all modified frontend files.
- The available LSP wrapper has no Go backend; Go type/static validation was provided
  by fresh `go test` and `go vet`.
- AST searches found no `console.log`, empty catch block, or hardcoded frontend API-key
  pattern in `web/src`.

## Runtime evidence and cleanup

Native health:

```text
GET http://127.0.0.1:8080/api/health -> 200
GET http://127.0.0.1:5173/ -> 200
```

Pre-probe database backup:

```text
/home/zephry_xzx/.local/share/xlab-blog/stage1-security-preprobe-20260612-102917.dump
sha256 50147bc5abf825b5ed4c0f52407dcc466b26d42fcf1b51391ea63f5b1060ebf4
```

All disposable users, nodes, Assets, and upload files were removed:

```text
probe_users=0
probe_nodes=0
```

The shared branch advanced after the security probes. This report's commit parent is
`525e65ccfc9c1242c4f1e85707d31013c75e0707`; the intervening product change
`58df9f6` only reorders Admin Directory-create conflict message classification in
`web/src/pages/AdminPage.tsx` and adds its regression test. It was not part of the
requested integrated SHA and was not included in this verdict. This report remains
strictly scoped to `8b343880`. Acceptance-seat artifacts were not modified, reverted,
staged, or included.

## Tested

- Integrated product SHA `8b343880d58a1b3a562a80afc1f84cab666933c3`.
- Missing, malformed, tampered, expired, forged-role, deleted-user, Reader, and Author
  token behavior.
- Every current `/api/admin/**` route as Anonymous Visitor and Reader.
- Database role authority over JWT claims.
- Reader `/admin`, Anonymous `/admin`, invalid-token, and current-user network-failure
  browser behavior.
- Absolute, scheme-relative, recursive, Author-only, and backslash return targets.
- Draft File public resolve, Recent, search, Asset, comment, and Like isolation.
- Credential/token echo in failed-login responses and API logs.
- Prior UpsertAdmin, placeholder deployment credential, and multipart-envelope
  observations.

## Not tested

- Real DashScope boundary; no API key was used.
- Docker Compose runtime; Docker is unavailable in this WSL environment and deployment
  is outside the native Stage 1 boundary.
- External production ingress, TLS, proxy, and rate limiting.

## Required next action

Create development fixes and a new independent security retest task. Stage 1 may not
close until the Author seed applies the configured credential, unexpected auth errors
are sanitized, and return targets reject backslash/control-character variants.
