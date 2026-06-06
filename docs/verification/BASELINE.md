# Verified Initial Baseline

Date: 2026-06-06

This file replaces packet-by-packet monitor logs and stale early acceptance matrices. Detailed command/output evidence remains in Git history and in `native-local-full-stack-smoke-20260606.md`.

## Static and unit gates

- Go `1.26.4`: full `go test -count=1 ./...` pass.
- Go vet: pass.
- Go formatting scan: pass.
- Frontend render-safety/static tests: 7/7 pass.
- Frontend lint: pass.
- TypeScript/Vite production build: pass.
- OpenAPI local references: previously verified.
- HTML iframe sandbox: `allow-scripts`, never `allow-same-origin`.

## Native runtime gates

Environment:

- Node.js `22.22.3`
- npm `10.9.8`
- Go `1.26.4`
- PostgreSQL `17.10`
- pgvector `0.8.1`

Fresh-database business smoke passed 21 checks covering authentication, Content Tree creation, content save/publish, public resolution, search fallback, Assets, Reader registration, comments, Likes, redirects, Unpublish, Draft isolation, and search exclusion.

Browser acceptance passed on desktop and mobile for Admin protection/login return, Recent, Search, File Assets/comments/Likes, responsive rendering, and console errors.

## Known boundaries

- Real DashScope embeddings were not tested without an API key; failure state and full-text/keyword fallback pass.
- Docker Compose live smoke is pending Docker availability in WSL; Docker/Caddy configuration passed static checks.
- Initial Admin UX is intentionally superseded by the staged second-development plan.

## Current gate

All second-development stages must preserve this baseline unless an active spec explicitly changes the behavior.
