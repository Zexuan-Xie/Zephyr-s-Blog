# Aeolian Blog

A single-Author full-stack personal blog organized as a Unix-like **Content Tree**. It supports Markdown and sandboxed HTML Files, Reader comments/Likes, per-File Assets, hybrid search, autosaved Content Versions, Published Content snapshots, Draft Preview, and a server-local stdio Blog MCP Server.

## Status

Stage 3 engineering is complete and ready for first-version release/acceptance:

- Stage 1: reliability, navigation, and identity — complete.
- Stage 2: simple-English Author Workspace and protected Content Tree — complete.
- Stage 3: autosave, Content Versions, Published Content, Draft Preview, Draft/Published Assets, and local stdio MCP — complete.

Read [`PROGRESS.md`](PROGRESS.md) first when resuming development.

## Repository layout

```text
.
├── AGENTS.md / AGENT.md      # agent/developer operating rules
├── PROGRESS.md               # durable current breakpoint
├── README.md                 # project overview and commands
├── api/                      # Go API, services, repositories, SQL migrations
├── web/                      # React/Vite SPA
├── mcp/                      # server-local stdio Blog MCP Server
├── docs/
│   ├── README.md             # documentation index
│   ├── specs/                # product, route, backend, design, stack specs
│   ├── api/                  # OpenAPI contract
│   ├── adr/                  # durable architectural decisions
│   ├── plans/                # staged implementation plan
│   ├── verification/         # acceptance/security/review evidence
│   └── archive/              # compact historical implementation summary
├── environment.yml           # local Conda environment
├── docker-compose.yml        # deployment/local container entry point
└── Caddyfile                 # web/API reverse proxy config
```

Ignored local artifacts include dependency folders, build output, databases/uploads/backups, logs, and agent runtime state (`.omx/`, `.code-review-graph/`, etc.).

## Local environment

Use the Conda environment declared in `environment.yml`:

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
pgvector 0.8.1
```

Install dependencies when needed:

```bash
conda run -n blogenv bash -lc 'cd web && npm ci'
conda run -n blogenv bash -lc 'cd mcp && npm ci'
```

## Local acceptance stack

The persistent local stack lives outside this repository under `~/.local/share/xlab-blog`.

```bash
~/.local/share/xlab-blog/start-local.sh
curl -fsS http://127.0.0.1:8080/api/health
curl -fsS http://127.0.0.1:5173/ >/dev/null
```

Endpoints:

- Frontend: `http://127.0.0.1:5173`
- API health: `http://127.0.0.1:8080/api/health`
- PostgreSQL: `127.0.0.1:55432`, database `xlab_blog`

Local development credentials remain outside Git. The local recovery script seeds the Author account documented in `PROGRESS.md` for acceptance.

## Verification

Backend:

```bash
conda run -n blogenv bash -lc '
  cd api &&
  CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./... &&
  CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go vet ./... &&
  test -z "$(gofmt -l .)"
'
```

Frontend:

```bash
conda run -n blogenv bash -lc '
  cd web &&
  node --test tests/*.test.mjs &&
  npm run lint &&
  npm run build
'
```

MCP:

```bash
conda run -n blogenv bash -lc '
  cd mcp &&
  npm test &&
  npm run build
'
```

Stage evidence is under [`docs/verification/`](docs/verification/), especially Stage 3 acceptance, security, code-review, and browser/API artifacts.

## MCP posture

The Blog MCP Server in `mcp/` is intended for trusted AI agents running on the server/local machine.

- stdio transport only;
- disabled by default via `BLOG_MCP_ENABLED`;
- per-call kill switch via `BLOG_MCP_KILL_SWITCH`;
- JSONL audit trail;
- backend HTTP API boundary, no direct SQL in MCP handlers.

See [`mcp/README.md`](mcp/README.md) for usage.
