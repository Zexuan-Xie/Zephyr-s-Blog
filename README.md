# Zephyr-s-Blog

xLab single-author full-stack personal blog / knowledge space.

The implementation plan is `IMPLEMENTATION_PLAN.md`. Project-level agent orientation is `AGENT.md`. Product and architecture specs live under `docs/` so the repo root stays GitHub-friendly.

## Repository layout

```txt
.
├── AGENT.md
├── AGENTS.md -> AGENT.md
├── README.md
├── IMPLEMENTATION_PLAN.md
├── environment.yml      # Conda environment contract for blogenv
├── .env.example         # local configuration template; copy to .env
├── docker-compose.yml   # local full-stack deployment skeleton
├── Caddyfile            # Caddy edge proxy skeleton
├── docs/
│   ├── README.md
│   ├── specs/          # active product, flow, stack, backend, design, vocabulary specs
│   ├── api/            # OpenAPI contract
│   ├── adr/            # active architectural decisions
│   ├── design/         # visual prototypes
│   └── archive/        # historical plans/roadmaps/decisions, not implementation source
├── api/                # Go API, created during implementation
└── web/                # Vite React SPA, created during implementation
```

## Active source documents

Read these before implementation or review:

1. `docs/specs/PRD.md`
2. `docs/specs/BLOG_FLOW.md`
3. `docs/specs/TECH_STACK.md`
4. `docs/specs/BACKEND_STRUCTURE.md`
5. `docs/api/openapi.yaml`
6. `docs/specs/DESIGN.md`
7. `docs/specs/CONTEXT.md`
8. `docs/adr/`

Archived files under `docs/archive/` are historical context only.

## Development environment

Use Conda for local development. The environment name is fixed and declared in `environment.yml`:

```bash
conda env create -f environment.yml
conda run -n blogenv npm install -g npm@10.9.8
conda activate blogenv
node --version   # expected 22.22.3
npm --version    # expected 10.9.8
go version       # expected 1.26.4
postgres --version  # expected 17.10 in the current local environment
```

`environment.yml` pins Node.js and Go through Conda. The exact npm version is
installed into `blogenv` with npm itself because the standalone `npm=10.9.8`
Conda package is not available from the current conda-forge/defaults channels.
If the solver cannot provide the exact versions in `docs/specs/TECH_STACK.md`,
stop and update the spec only after an explicit decision; do not silently
substitute versions.

## Local configuration

```bash
cp .env.example .env
# edit .env secrets before running the API or Docker Compose
```

`docker-compose.yml` uses pgvector, the Go API image, the Vite static-build image, Caddy SPA fallback, `postgres_data`, `uploads`, and `web_dist` foundations.

## Planned verification

Backend:

```bash
cd api
go test ./...
```

Frontend:

```bash
cd web
npm run lint
npm run build
```

Full stack:

```bash
docker compose config
docker compose up -d --build
curl -fsS http://localhost:8080/api/health
```

For a native pre-container smoke run, `blogenv` also contains PostgreSQL and
pgvector. Initialize a disposable cluster outside the repo, run the API with
`CGO_ENABLED=0`, then start Vite:

```bash
conda run -n blogenv initdb -D /tmp/xlab-blog-local-pg --auth-local=trust --auth-host=trust --no-locale
conda run -n blogenv pg_ctl -D /tmp/xlab-blog-local-pg -l /tmp/xlab-blog-postgres.log \
  -o "-p 55432 -h 127.0.0.1 -k /tmp" start
conda run -n blogenv createdb -h 127.0.0.1 -p 55432 xlab_blog

cd api
CGO_ENABLED=0 HTTP_ADDR=127.0.0.1:8080 \
DATABASE_URL='postgres://YOUR_LOCAL_USER@127.0.0.1:55432/xlab_blog?sslmode=disable' \
JWT_SECRET='replace-with-at-least-32-bytes' \
ADMIN_EMAIL='admin@example.com' ADMIN_PASSWORD='replace-me' \
ASSET_UPLOAD_DIR='/tmp/xlab-blog-uploads' \
conda run -n blogenv go run ./cmd/server

cd ../web
conda run -n blogenv npm run dev -- --host 127.0.0.1
```
