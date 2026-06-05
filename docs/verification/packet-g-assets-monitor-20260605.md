# Packet G — Per-File Assets Verification

Date: 2026-06-05 22:32 CST
Scope: Packet G backend asset core, public/admin route wiring, public file asset payloads, and frontend/admin asset UI.

## Implementation evidence

Integrated commits on `main`:

- `46ee29f` — backend `api/internal/assets` storage/service/repository/validation foundation and asset handlers.
- `7527f78` — public/admin asset routes, server/config wiring, and tree asset payload integration.
- `c134dc2` — frontend asset types/API helpers, public file asset panel, and admin Asset Manager foundation.

Key acceptance mappings:

- Published assets are served through `GET /api/assets/{asset_id}/{filename}` only when the owning File is published.
- Public asset responses set `Cache-Control: public, max-age=31536000, immutable`.
- Draft/unpublished File assets are rejected by public serve lookup.
- Storage keys use provider-neutral `files/{file_node_id}/{asset_id}-{safe_filename}` keys; local storage rejects absolute, backslash, and path-escape keys.
- Upload validation enforces MIME/size allowlists and rejects unsafe SVG markers: `<script>`, event handlers, `javascript:`, external `href/src`, and `foreignObject`.
- Public/admin File payloads expose asset metadata and `public_url`; frontend renders assets on File pages and supports admin upload/delete by File node id.

## Verification commands and results

All commands were run from repository root unless noted.

### Backend targeted gates

```bash
cd api && PATH=/tmp/omx-go-1.26.4/go/bin:$PATH GOCACHE=/tmp/omx-go-cache go test -count=1 ./internal/assets ./internal/http/handlers
```

Result: PASS.

```bash
cd api && PATH=/tmp/omx-go-1.26.4/go/bin:$PATH GOCACHE=/tmp/omx-go-cache go test -count=1 ./internal/config ./internal/http ./internal/assets ./internal/http/handlers ./internal/tree
```

Result: PASS.

### Packet G terminal gate

```bash
cd api && PATH=/tmp/omx-go-1.26.4/go/bin:$PATH GOCACHE=/tmp/omx-go-cache go test -count=1 ./...
cd api && PATH=/tmp/omx-go-1.26.4/go/bin:$PATH GOCACHE=/tmp/omx-go-cache go vet ./...
find api -name '*.go' -not -path '*/vendor/*' -print0 | xargs -0 /tmp/omx-go-1.26.4/go/bin/gofmt -l
node --test web/tests/render-safety.test.mjs
cd web && npm run lint
cd web && npm run build
ruby -ryaml -e '<OpenAPI local ref walk>'
grep -R "Cache-Control.*public, max-age=31536000, immutable" -n api/internal/http/handlers/assets.go
grep -R "foreignobject\|javascript:\|<script" -n api/internal/assets/validation.go
! grep -R "allow-same-origin" -n web/src
git diff --check
git status --short --branch
```

Result: PASS.

Observed evidence:

- Full backend tests passed for all packages.
- `go vet ./...` passed.
- `gofmt` scan returned no unformatted Go files.
- Frontend render/static tests passed: 5/5.
- `npm run lint` passed.
- `npm run build` passed.
- OpenAPI local ref walk passed: `paths=22 schemas=33 refs=100`.
- Immutable cache guard found in `api/internal/http/handlers/assets.go`.
- SVG rejection guards found in `api/internal/assets/validation.go`.
- No `allow-same-origin` was present under `web/src`.
- `git diff --check` passed.
- Pre-documentation git status was clean on `main...origin/main [ahead 101]`.

## Known gaps / risks

- Docker and live browser/backend E2E smoke were not run in this runtime.
- `ASSET_PUBLIC_BASE_URL` is wired for uploaded asset responses; tree payloads currently expose default `/api/assets/...` public URLs, which is acceptable for the default deployment but should be revisited if a non-default public asset base is required.
- Asset Manager is a functional foundation, not the final full tree-integrated admin UI from Packet I.

## Terminal decision

Packet G is terminally verified for the local implementation gate. Next plan-aligned breakpoint is Packet H — Hybrid Search.
