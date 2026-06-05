# Packet H — Hybrid Search Verification

Date: 2026-06-05 22:44 CST
Scope: Packet H backend hybrid search package/provider/handlers, route/server wiring, and frontend search-page static contract.

## Implementation evidence

Integrated commits on `main`:

- `ff3ad42` — Packet H durable start breakpoint.
- `013e8c8` — Packet H hybrid search implementation.

Key acceptance mappings:

- Public `GET /api/search?q=` is routed and backed by `api/internal/search`.
- Full-text retrieval uses `websearch_to_tsquery('simple', q)`, rank, and `ts_headline` snippets over published File rows only.
- Qwen/DashScope provider posts to OpenAI-compatible `/embeddings` with `model: text-embedding-v4`, `dimensions: 1024`, and `encoding_format: "float"`.
- Semantic retrieval uses stored `embedding vector(1024)` values with pgvector cosine distance (`<=>`) for `embedding_status='ready'` published Files only.
- RRF fusion uses default `k=60`, combines text/semantic candidates, and surfaces source badges (`text`, `semantic`, `keyword`).
- Admin routes are wired: `POST /api/admin/files/{file_id}/refresh-embedding` and `POST /api/admin/search-index/rebuild`.
- Embedding provider failures are persisted as `embedding_status='failed'` while public search still falls back to full-text results.
- Frontend `/search?q=` scaffold calls the search API and renders path, snippet, and source badges.

## Verification commands and results

All commands were run from repository root unless noted.

### Backend targeted gate

```bash
cd api && PATH=/tmp/omx-go-1.26.4/go/bin:$PATH GOCACHE=/tmp/omx-go-cache go test -count=1 ./internal/search ./internal/http/handlers ./internal/http
```

Result: PASS.

### Packet H terminal gate

```bash
cd api && PATH=/tmp/omx-go-1.26.4/go/bin:$PATH GOCACHE=/tmp/omx-go-cache go test -count=1 ./...
cd api && PATH=/tmp/omx-go-1.26.4/go/bin:$PATH GOCACHE=/tmp/omx-go-cache go vet ./...
find api -name '*.go' -not -path '*/vendor/*' -print0 | xargs -0 /tmp/omx-go-1.26.4/go/bin/gofmt -l
node --test web/tests/render-safety.test.mjs
cd web && npm run lint
cd web && npm run build
ruby -ryaml -e '<OpenAPI local ref walk>'
grep -R "websearch_to_tsquery('simple'" -n api/internal/search/repository.go
grep -R '"dimensions".*p.dimensions\|"encoding_format".*"float"' -n api/internal/search/provider.go
grep -R "DefaultRRFK.*60" -n api/internal/search/types.go
! grep -R "allow-same-origin" -n web/src
git diff --check
git status --short --branch
```

Result: PASS.

Observed evidence:

- Full backend tests passed for all packages.
- `go vet ./...` passed.
- `gofmt` scan returned no unformatted Go files.
- Frontend render/static tests passed: 6/6, including Packet H search API/source-badge checks.
- `npm run lint` passed.
- `npm run build` passed.
- OpenAPI local ref walk passed: `paths=22 schemas=33 refs=100`.
- Static guards found `websearch_to_tsquery('simple', $1)`, Qwen `dimensions`/`encoding_format`, and `DefaultRRFK = 60`.
- No `allow-same-origin` was present under `web/src`.
- `git diff --check` passed.
- Pre-documentation git status was clean on `main...origin/main [ahead 104]`.

## Known gaps / risks

- Live Postgres query execution was not smoke-tested in this runtime, so SQL shape is compile/static/unit verified but not integration-tested against a running pgvector database.
- Live DashScope network calls were not made; provider behavior is covered with an `httptest` server that verifies request body and response parsing.
- Docker and live browser/backend E2E smoke were not run in this runtime.
- Rebuild currently performs synchronous per-file refresh work and returns `accepted`; it is acceptable for this foundation but can be moved to a background worker if production data size grows.

## Terminal decision

Packet H is terminally verified for the local implementation gate. Next plan-aligned breakpoint is Packet I — Admin Tree Manager.
