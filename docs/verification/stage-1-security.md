# Stage 1 Security Verification

Status: complete

Verdict: **PASS**

Integrated product SHA: `d6c7d092ae41b2e37bbf1c89ea30cff4ec551ef6`

Prior failing evidence: `1877c254e310c85cc42aa68cecd4acfcf8e27460`

## Retest decision

The three Stage 1 blockers recorded by the prior independent review are repaired:

1. Configured Author seeding now authoritatively replaces the password hash, role, and provider.
2. Unexpected Register/Login service failures now return a generic HTTP 500 response.
3. Backslash/control-character Login return targets, including encoded variants, now fall back safely.

No Stage 1 security blocker remains.

## Blocker retests

### Author seed credential replacement — PASS

Implementation:

- `api/internal/users/repository.go`
- `api/internal/auth/service_test.go`
- `api/internal/users/repository_test.go`

`UpsertAdmin` now updates:

- `password_hash = excluded.password_hash`
- `role = excluded.role`
- `provider = excluded.provider`

Evidence:

```text
TEST_DATABASE_URL=postgres://.../xlab_blog
go test -count=1 -run TestUpsertAdminMakesConfiguredSeedAuthoritative ./internal/users
ok xlab-blog/api/internal/users
```

The PostgreSQL test runs inside a rollback-only transaction and proves the old
Reader password/provider are replaced by the configured local Author credential.
Service coverage also proves the configured password authenticates and the prior
Reader password no longer does.

### Unexpected authentication errors — PASS

Implementation:

- `api/internal/http/handlers/auth.go`
- `api/internal/http/handlers/auth_test.go`
- `docs/api/openapi.yaml`

Known validation/conflict/credential mappings remain explicit. Unexpected service or
repository errors now return:

```json
{"error":"internal server error"}
```

The response status is HTTP 500 and the internal sentinel is absent. OpenAPI documents
the 400/401/409/500 contracts.

### Unsafe return targets — PASS

Implementation:

- `web/src/lib/auth.ts`
- `web/tests/identity-navigation.test.mjs`

Regression cases cover decoded and repeatedly encoded backslashes, ASCII controls,
scheme-relative/external targets, and Login/Register loops. Valid application
paths/query/hash remain valid.

Real Chromium evidence:

```text
input return_to=/%5C%5Cattacker.example/steal
final_url=http://127.0.0.1:5173/recent
browser_errors=none
```

Screenshot:

- `stage-1-browser-20260612/backslash-return-fixed.png`

## Preserved boundaries

Previously verified and unchanged:

- every current Author API route is protected by `RequireAdmin`;
- Anonymous Visitor receives 401 and Reader receives 403;
- Reader denial retains the session and shows `Author access required`;
- invalid tokens are cleared, while transport failure preserves the session and offers Retry;
- Draft File resolution, search, Assets, comments, and Likes remain isolated;
- iframe sandbox remains exactly `allow-scripts` without `allow-same-origin`;
- full-text search fallback remains available without semantic indexing.

Supporting browser evidence:

- `stage-1-browser-20260612/reader-desktop-admin-denied.png`
- `stage-1-browser-20260612/reader-mobile-admin-denied.png`
- `stage-1-browser-20260612/invalid-token-cleared.png`
- `stage-1-browser-20260612/identity-network-failure.png`

## Required gates

Backend:

```text
go test -count=1 ./...  PASS
go vet ./...            PASS
gofmt scan              PASS
```

Frontend:

```text
Node tests              17/17 PASS
npm run lint            PASS
npm run build           PASS
```

Evidence:

- `stage-1-backend-gates-20260612.log`
- `stage-1-frontend-gates-20260612.log`
- `stage-1-browser-20260612/backend-gates-final.log`
- `stage-1-browser-20260612/frontend-gates-final.log`

## Non-blocking release hardening

These findings do not block native Stage 1 acceptance, but block production-hardening
claims until addressed:

- Docker Compose placeholder JWT/database/Author defaults must fail closed before deployment.
- Multipart Asset upload needs a hard total request-size envelope, not only an in-memory form limit.

Docker/server deployment remains intentionally deferred until after native user acceptance.
