# Native Local Full-Stack Acceptance Verification

Date: 2026-06-06 14:18 CST

Scope: Native Conda/PostgreSQL/API/Vite validation before Docker or server deployment.

## Acceptance environment

- Conda environment: `blogenv`
- Node.js: `22.22.3`
- npm: `10.9.8`
- Go: `1.26.4`
- PostgreSQL: `17.10`
- pgvector: `0.8.1`
- Main PostgreSQL: `127.0.0.1:55432`, database `xlab_blog`
- Main API: `http://127.0.0.1:8080`
- Main frontend: `http://127.0.0.1:5173`

The main services were intentionally left running for user acceptance.

## Runtime defects found and repaired

Real PostgreSQL and browser use exposed issues that static checks did not:

- PostgreSQL 17 rejected the generated `search_vector` expression as non-immutable; it is now a normal `tsvector` maintained by a trigger.
- Comment insert and soft-delete queries supplied too many pgx bind arguments.
- The frontend expected `/api/recent`, but the API route and repository query were missing.
- `/admin` lacked a frontend authentication guard and login return flow.
- Admin panel headings were clipped because the glass panels lacked internal padding.

The repairs are integrated in commit `22dfa71`.

## Fresh-database API smoke

A disposable `xlab_blog_verify` database and a temporary API on port `18080` were created, tested, stopped, and removed. The main acceptance database was not modified by this rerun.

Result: `SMOKE_OK checks=21`.

Covered behavior:

1. Admin login.
2. Directory creation.
3. File creation.
4. Content save and search-text generation.
5. Publish.
6. Public path resolution.
7. Full-text/keyword search fallback.
8. Asset upload.
9. Public immutable asset serving.
10. Missing-provider embedding failure state.
11. Reader registration.
12. Comment creation.
13. File like.
14. Comment like.
15. Anonymous comment read.
16. Published-file rename.
17. Old-path redirect.
18. Unpublish.
19. Draft public-path isolation.
20. Draft asset isolation.
21. Draft search exclusion.

## Browser acceptance

The Browser plugin was not available in this runtime, so the documented fallback used headless Playwright Chromium.

Desktop result: `BROWSER_ACCEPTANCE_OK`.

- Anonymous `/admin` redirects to `/login?return_to=%2Fadmin`.
- Login returns to `/admin`.
- Admin Tree Manager renders.
- Recent page renders the published sample file.
- Search returns the sample file.
- File page renders assets, comments, and likes.
- No relevant browser console errors were observed.

Mobile result: `MOBILE_SMOKE_OK`.

- Viewport: `390x844`.
- The recent page rendered successfully without the prior heading clipping.

Temporary screenshots:

- `/tmp/xlab-admin-acceptance.png`
- `/tmp/xlab-recent-acceptance.png`
- `/tmp/xlab-mobile-recent.png`

## Regression gate

All commands passed in the exact Conda environment:

```bash
conda run -n blogenv bash -lc 'node --version; npm --version; go version; postgres --version'
cd api
conda run -n blogenv bash -lc \
  'export CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache; go test -count=1 ./... && go vet ./... && test -z "$(gofmt -l .)"'
cd ../web
conda run -n blogenv bash -lc \
  'node --test tests/render-safety.test.mjs && npm run lint && npm run build'
node /tmp/xlab_browser_acceptance.cjs
node /tmp/xlab_mobile_smoke.cjs
```

Observed results:

- Full backend tests: pass.
- `go vet ./...`: pass.
- Go formatting scan: pass.
- Frontend render/static tests: 7/7 pass.
- Frontend lint: pass.
- Frontend typecheck/build: pass.
- Desktop and mobile browser acceptance: pass.
- Main health endpoint after temporary smoke cleanup: `{"status":"ok","database":"ok"}`.

## Remaining boundaries

- Real DashScope embedding generation was not tested because no API key was configured. The missing-provider failure state and full-text/keyword fallback were verified.
- Docker Compose remains outside this acceptance stage because Docker is unavailable in the WSL distro. Containerization is intentionally deferred until after native user acceptance.

## Acceptance decision

The native local stack is ready for real user acceptance at `http://127.0.0.1:5173`. Keep PostgreSQL, the API, and Vite running until the user finishes acceptance. The next deployment step is Docker/WSL enablement followed by live Compose smoke.
