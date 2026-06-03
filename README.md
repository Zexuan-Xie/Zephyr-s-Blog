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
├── docs/
│   ├── README.md
│   ├── specs/          # active product, flow, stack, backend, design, vocabulary specs
│   ├── api/            # OpenAPI contract
│   ├── adr/            # active architectural decisions
│   ├── design/         # visual prototypes
│   └── archive/        # historical plans/roadmaps/decisions, not implementation source
├── api/                # Go API, created during implementation
├── web/                # Vite React SPA, created during implementation
├── docker-compose.yml  # created during implementation
└── Caddyfile           # created during implementation
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

Use Conda for local development. The environment name is fixed:

```bash
conda activate blogenv
```

If an `environment.yml` is added, it must declare `name: blogenv` and keep versions aligned with `docs/specs/TECH_STACK.md`.

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
