# Phase 0/1 Verification Acceptance Matrix

Updated by leader on 2026-06-03 after all `follow-implementation-b973ccd0` tasks reached terminal completion.

## Scope and Sources

- Implementation plan: `IMPLEMENTATION_PLAN.md`
- Product acceptance: `docs/specs/PRD.md`
- Stack/version contract: `docs/specs/TECH_STACK.md`
- Backend/API contract: `docs/specs/BACKEND_STRUCTURE.md`, `docs/api/openapi.yaml`
- Latest progress ledger: `PROGRESS.md`
- Prior verification refresh: `docs/verification/task-5-leader-verification-pass.md`

## Summary

Phase 0/1 foundation artifacts are now present. Root/deploy skeleton, backend health/auth foundation, frontend SPA foundation, and OpenAPI static checks are integrated. Backend tests pass with exact Go `1.26.4`. Conda environment creation required a spec adjustment because the standalone `npm=10.9.8` Conda package was unavailable from the current channels; `environment.yml` now pins Node.js and Go through Conda and documents installing exact npm inside `blogenv`.

Full PRD acceptance remains incomplete because content tree/search/comments/likes/assets/admin APIs and end-to-end Docker/browser flows are not implemented/proven yet.

## Contract Checks

| Check | Result | Evidence | Gap / Next Action |
|---|---:|---|---|
| OpenAPI YAML parses | PASS | Python YAML/ref checks in prior verification reports; `paths=22`, `schemas=33` | Keep `docs/api/openapi.yaml` authoritative before API shape changes. |
| OpenAPI local `$ref` targets resolve | PASS | Prior custom Python ref walk found all local refs resolve | Run a formal OpenAPI validator when available. |
| Root environment contract | PASS | `environment.yml` exists with `name: blogenv`; pins `nodejs=22.22.3` and `go=1.26.4`; README documents `npm install -g npm@10.9.8` inside env | Retry `blogenv` creation when network stabilizes, then verify npm exact version. |
| Deployment skeleton | PASS static | `docker-compose.yml` and `Caddyfile` exist | Docker runtime unavailable in current environment; run `docker compose config` once Docker works. |
| Backend exact Go module contract | PASS | `api/go.mod` uses `go 1.26.4` and required direct module pins | Continue implementing full OpenAPI surface after auth/health foundation. |
| Backend tests | PASS | `/tmp/omx-go-1.26.4/go/bin/go version` -> `go1.26.4`; `(cd api && PATH="/tmp/omx-go-1.26.4/go/bin:$PATH" go test ./...)` passed | Re-run with `conda run -n blogenv go test ./...` once `blogenv` completes. |
| Frontend exact package contract | PASS static | `web/package.json` pins required dependencies/devDependencies exactly | Re-run lint/build with exact `blogenv` Node/npm after env setup. |
| HTML iframe sandbox anchor | PASS static | `web/src/components/FilePage.tsx` uses `sandbox="allow-scripts"`; no `allow-same-origin` found in prior check | Add browser/runtime checks later. |
| Markdown sanitization anchor | PASS static | `web/src/lib/renderMarkdown.ts` uses `marked` then `DOMPurify.sanitize` | Add XSS regression tests later. |

## Artifact Presence Matrix

| Required artifact | Required by | Current state | Status |
|---|---|---|---:|
| `environment.yml` with `name: blogenv` | `IMPLEMENTATION_PLAN.md`, `TECH_STACK.md` | Present; Conda pins Node/Go and documents exact npm install step | PASS |
| `.env.example` | Backend/env bootstrap | Present | PASS |
| `api/go.mod` | Go/chi backend foundation | Present; exact required pins | PASS |
| `api/internal/**` | Backend layering and services | Present for config/db/auth/http/users foundation | PASS |
| `api/migrations/**` | DB schema and pgvector/extensions | Present: `000001_initial_schema.sql` | PASS |
| `web/package.json` | Vite React SPA foundation | Present; exact pins | PASS |
| `web/src/**` | SPA routes and UI shell | Present | PASS |
| `docker-compose.yml` | Deployability acceptance | Present | PASS static |
| `Caddyfile` | SPA/API edge acceptance | Present | PASS static |
| `docs/api/openapi.yaml` | API contract source of truth | Present; parse/ref checks pass | PASS |

## Local Toolchain Matrix

| Tool | Required | Observed | Status |
|---|---:|---:|---:|
| Conda | local env manager | `conda 26.1.1` | PASS |
| Go | `1.26.4` | Temporary exact toolchain `/tmp/omx-go-1.26.4/go/bin/go` -> `go1.26.4`; Conda env creation blocked by network download failures | PASS for backend tests / PENDING for durable env |
| Node.js | `22.22.3` | Conda can solve `nodejs=22.22.3`; env creation blocked by network download failures | PENDING |
| npm | `10.9.8` | Standalone Conda package unavailable; README/environment notes install exact npm inside `blogenv` using npm | PENDING |
| Docker Compose | v2 via `docker compose` | Docker unavailable in current WSL/Docker Desktop setup | BLOCKED |

## PRD Acceptance Matrix

| PRD acceptance signal | Current result | Evidence / blocker |
|---|---:|---|
| Docker Compose starts web/api/db; Caddy serves SPA and proxies `/api` | BLOCKED | Static Compose/Caddy files exist; Docker runtime unavailable. |
| Admin seed login and `/admin` content creation | PARTIAL/BLOCKED | Auth foundation exists; admin/content APIs and UI are not implemented. |
| Root/nested/File SPA path resolution | PARTIAL/BLOCKED | SPA route shell exists; `/api/tree/resolve` implementation missing. |
| Published path changes create `path_redirects` and resolve redirects | BLOCKED | Migration table exists; service/API implementation missing. |
| Markdown File renders body, keyword chips, likes, comments | PARTIAL/BLOCKED | Markdown sanitizer/UI anchor exists; comments/likes APIs incomplete. |
| HTML Document uses iframe sandbox `allow-scripts` without `allow-same-origin` | PASS static / BLOCKED runtime | Static iframe anchor passes; no browser/runtime test yet. |
| Assets upload/public access/cache/MIME/SVG/PDF rules | BLOCKED | Migration table exists; asset API/storage validation missing. |
| `/search` returns path/snippet/source badge; embedding failure falls back to full-text | BLOCKED | OpenAPI paths exist; backend/frontend search behavior incomplete. |
| Anonymous read; interaction redirects to login and returns to target or `/recent` | PARTIAL/BLOCKED | SPA/auth foundations exist; end-to-end behavior unproven. |
| Five high-risk tests: JWT, like idempotency, Markdown XSS, iframe sandbox, SVG safety | PARTIAL/BLOCKED | JWT/auth tests pass; like/XSS/sandbox/SVG tests still needed. |

## Verification Commands Run

```text
Backend exact Go version:
/tmp/omx-go-1.26.4/go/bin/go version -> go version go1.26.4 linux/amd64

Backend tests:
(cd api && PATH="/tmp/omx-go-1.26.4/go/bin:$PATH" go test ./...) -> PASS

OMX team reconciliation:
omx team status follow-implementation-b973ccd0 --json --tail-lines 100 -> phase=complete, tasks completed=5 pending=0 in_progress=0 failed=0

Conda solver:
conda env create -f environment.yml before spec adjustment -> FAIL, standalone npm=10.9.8 unavailable from current conda-forge/defaults channels
conda create -n blogenv-probe -c conda-forge nodejs=22.22.3 go=1.26.4 --dry-run -> PASS
conda env create -f environment.yml after spec adjustment -> FAIL twice during package download: first ConnectionResetError(104), then CondaHTTPError HTTP 000 for go-1.26.4 package
```

## Recommended Next Feasible Work

1. Finish durable `blogenv` creation and exact npm installation, then verify `node --version`, `npm --version`, `go version`.
2. Re-run `cd api && go test ./...` with `blogenv` Go.
3. Run frontend `npm install`, `npm run lint`, and `npm run build` with exact `blogenv` Node/npm.
4. Run `docker compose config` once Docker is available.
5. Continue the next implementation packet for OpenAPI routes beyond auth/health: public tree, search, comments, likes, assets, and admin APIs.
