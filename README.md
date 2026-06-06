# xLab Blog

A single-author full-stack technical blog organized as a Unix-like Content Tree. It supports Markdown and sandboxed HTML Files, Reader comments/Likes, per-File Assets, hybrid search, and an Author workspace.

## Current development status

The initial A–J implementation is complete and has passed native PostgreSQL/API/browser smoke. The active work is the staged second development described in [`docs/plans/SECOND_DEVELOPMENT.md`](docs/plans/SECOND_DEVELOPMENT.md).

Read [`PROGRESS.md`](PROGRESS.md) first when resuming work.

## Repository layout

```text
.
├── AGENTS.md                 # agent/developer operating rules
├── PROGRESS.md               # durable current breakpoint
├── README.md
├── api/                      # Go API and SQL migrations
├── web/                      # React/Vite SPA
├── docs/
│   ├── README.md             # documentation index
│   ├── plans/                # active implementation plan
│   ├── specs/                # active product/flow/backend/design/stack specs
│   ├── api/                  # OpenAPI contract
│   ├── adr/                  # durable architectural decisions
│   ├── design/               # approved visual reference
│   ├── verification/         # current baseline and acceptance evidence
│   └── archive/              # compact initial-build summary
├── environment.yml
├── docker-compose.yml
└── Caddyfile
```

## Local environment

```bash
conda env create -f environment.yml
conda run -n blogenv npm install -g npm@10.9.8
```

Expected versions:

```text
Node.js 22.22.3
npm 10.9.8
Go 1.26.4
PostgreSQL 17.10
```

Install frontend dependencies when needed:

```bash
cd web
conda run -n blogenv npm ci
```

## Local acceptance stack

The existing persistent local stack lives outside the repository under `~/.local/share/xlab-blog`.

Start or recover it with:

```bash
~/.local/share/xlab-blog/start-local.sh
```

Endpoints:

- Frontend: `http://127.0.0.1:5173`
- API health: `http://127.0.0.1:8080/api/health`
- PostgreSQL: `127.0.0.1:55432`, database `xlab_blog`

Local credentials remain outside Git.

## Verification

```bash
conda run -n blogenv bash -lc \
  'cd api && CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./... && CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go vet ./... && test -z "$(gofmt -l .)"'

conda run -n blogenv bash -lc \
  'cd web && node --test tests/render-safety.test.mjs && npm run lint && npm run build'
```

See [`docs/verification/BASELINE.md`](docs/verification/BASELINE.md) for the compact verified baseline and [`docs/verification/native-local-full-stack-smoke-20260606.md`](docs/verification/native-local-full-stack-smoke-20260606.md) for detailed native acceptance evidence.

## Deployment boundary

Native local acceptance must pass before Docker/server deployment. Docker Compose live smoke remains pending Docker availability in this WSL environment.
