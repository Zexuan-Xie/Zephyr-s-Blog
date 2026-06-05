# Packet J — Deployment and Smoke Verification

Date: 2026-06-05 22:52 CST
Scope: Packet J Docker/Caddy/env deployment foundation and local verification attempts.

## Implementation evidence

Integrated commit on `main`:

- `976b5c1` — Packet J deployment foundation.

Key delivery mappings:

- `api/Dockerfile` builds the Go API with `golang:1.26.4-alpine`, copies migrations, and runs `/app/xlab-blog-api` on port 8080.
- `web/Dockerfile` builds the Vite SPA with `node:22.22.3-alpine`, `npm ci`, and `npm run build`, then copies `dist` into a shared `/srv` volume.
- `docker-compose.yml` defines `db`, `api`, `web`, and `caddy` services with `postgres_data`, `uploads`, `web_dist`, `caddy_data`, and `caddy_config` volumes.
- API config now uses `HTTP_ADDR=:8080` in Compose/.env, matching `api/internal/config`.
- Caddy reverse-proxies `/api/*` to `api:8080` and serves the SPA from `/srv` with `try_files {path} /index.html` and `file_server` fallback for client-side routes.
- API and web services include healthchecks; Caddy waits for both before serving.

## Verification commands and results

All commands were run from repository root unless noted.

### Packet J terminal/local gate

```bash
cd api && PATH=/tmp/omx-go-1.26.4/go/bin:$PATH GOCACHE=/tmp/omx-go-cache go test -count=1 ./...
cd api && PATH=/tmp/omx-go-1.26.4/go/bin:$PATH GOCACHE=/tmp/omx-go-cache go vet ./...
find api -name '*.go' -not -path '*/vendor/*' -print0 | xargs -0 /tmp/omx-go-1.26.4/go/bin/gofmt -l
node --test web/tests/render-safety.test.mjs
cd web && npm run lint
cd web && npm run build
ruby -ryaml -e '<OpenAPI local ref walk>'
ruby -ryaml -e '<Compose/Caddy static service-volume-fallback validation>'
grep -R "FROM golang:1.26.4-alpine\|go build" -n api/Dockerfile
grep -R "FROM node:22.22.3-alpine\|npm ci\|npm run build" -n web/Dockerfile
grep -R "try_files {path} /index.html\|reverse_proxy api:8080\|file_server" -n Caddyfile
! grep -R "allow-same-origin" -n web/src
git diff --check
git status --short --branch
```

Result: PASS.

Observed evidence:

- Full backend tests passed for all packages.
- `go vet ./...` passed.
- `gofmt` scan returned no unformatted Go files.
- Frontend render/static tests passed: 7/7.
- `npm run lint` passed.
- `npm run build` passed.
- OpenAPI local ref walk passed: `paths=22 schemas=33 refs=100`.
- Compose/Caddy static validation passed: `services=db,api,web,caddy`; volumes include `postgres_data,uploads,web_dist,caddy_data,caddy_config`.
- Dockerfile/Caddy static guards found exact Go/Node base images, build commands, API reverse proxy, and SPA fallback.
- No `allow-same-origin` was present under `web/src`.
- `git diff --check` passed.
- Pre-documentation git status was clean on `main...origin/main [ahead 109]`.

### Docker probe

```bash
docker compose config
```

Result: BLOCKED by environment.

Output:

```txt
The command 'docker' could not be found in this WSL 2 distro.
We recommend to activate the WSL integration in Docker Desktop settings.
```

Because Docker is unavailable, the following Packet J acceptance checks could not be executed in this runtime:

- `docker compose config`
- `docker compose up -d --build`
- `curl -fsS http://localhost:8080/api/health` against the Compose stack
- SPA route curl through the Caddy container

## Known gaps / risks

- Docker daemon/CLI is not installed or enabled in this WSL distro, so live Compose build/smoke remains external-environment blocked.
- Docker image tags are aligned with the project version spec but were not pulled/built in this runtime.
- The `web` service is a static asset publisher into the `web_dist` volume; Caddy is the public SPA server.

## Terminal decision

Packet J implementation is locally complete and statically verified, but live Docker smoke acceptance is blocked until Docker is available in the runtime. The durable resume point is: enable Docker/WSL integration, then run `docker compose config`, `docker compose up -d --build`, API health curl, and SPA fallback curl.
