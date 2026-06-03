# Project Agent Guide — Zephyr-s-Blog

This file is the project-level orientation guide for agents working in this repository. It records the current GitHub-friendly repo structure, the active source-of-truth documents, the archive policy, and the document soft-link index.

## Working Rule

Read `IMPLEMENTATION_PLAN.md` first, then read the active source documents under `docs/specs/`, `docs/api/`, `docs/adr/`, and `docs/design/` as needed for the assigned packet. Archived files under `docs/archive/` are historical context only and must not override active specs.

## Repo Structure

```txt
.
├── AGENT.md                 # project-level agent orientation guide
├── AGENTS.md -> AGENT.md    # compatibility symlink for agents that look for AGENTS.md
├── README.md                # human-facing GitHub entrypoint
├── IMPLEMENTATION_PLAN.md   # OMX-native execution plan
├── .gitignore               # env/build/runtime ignore rules
├── docs/
│   ├── README.md            # documentation index and archive policy
│   ├── links/               # soft-link index to docs and root markdown files
│   ├── specs/               # active product / flow / stack / backend / design specs
│   ├── api/                 # OpenAPI contract
│   ├── adr/                 # active architecture decision records
│   ├── design/              # visual prototype(s)
│   └── archive/             # historical plans, roadmaps, and decisions
├── api/                     # planned Go API implementation
├── web/                     # planned Vite React SPA implementation
├── docker-compose.yml       # planned full-stack local/deploy composition
└── Caddyfile                # planned SPA serving + /api reverse proxy
```

`api/`, `web/`, `docker-compose.yml`, and `Caddyfile` are planned implementation artifacts and may not exist until their implementation packets run.

## Soft-Link Index

Soft links live under `docs/links/`. They provide stable grouped shortcuts without crowding the repository root.

### Root project documents

| Soft link | Target | Purpose |
|---|---|---|
| `docs/links/root/AGENT.md` | `AGENT.md` | This agent guide and repo structure map. |
| `docs/links/root/README.md` | `README.md` | Human-facing GitHub overview, active docs list, and setup notes. |
| `docs/links/root/IMPLEMENTATION_PLAN.md` | `IMPLEMENTATION_PLAN.md` | OMX-native multi-agent implementation plan and packet map. |

### Active source-of-truth documents

| Soft link | Target | Purpose |
|---|---|---|
| `docs/links/active/docs-index.md` | `docs/README.md` | Documentation index and archive policy. |
| `docs/links/active/specs/PRD.md` | `docs/specs/PRD.md` | Product requirements, user roles, scope, non-goals, and acceptance signals. |
| `docs/links/active/specs/BLOG_FLOW.md` | `docs/specs/BLOG_FLOW.md` | Public/admin routes, user flows, UI states, and navigation behavior. |
| `docs/links/active/specs/TECH_STACK.md` | `docs/specs/TECH_STACK.md` | Exact runtime, dependency, Docker image, and prohibited substitution rules. |
| `docs/links/active/specs/BACKEND_STRUCTURE.md` | `docs/specs/BACKEND_STRUCTURE.md` | Go backend layout, database schema, repository/service rules, and edge cases. |
| `docs/links/active/specs/DESIGN.md` | `docs/specs/DESIGN.md` | Glass Ricepaper visual system, tokens, component language, and UI constraints. |
| `docs/links/active/specs/CONTEXT.md` | `docs/specs/CONTEXT.md` | Canonical product vocabulary and forbidden old terminology. |
| `docs/links/active/api/openapi.yaml` | `docs/api/openapi.yaml` | OpenAPI API contract; update before route/shape changes. |
| `docs/links/active/design/glass-light-v2.html` | `docs/design/glass-light-v2.html` | Approved visual prototype backing `DESIGN.md`. |

### Active ADR documents

| Soft link | Target | Purpose |
|---|---|---|
| `docs/links/active/adr/0001-nested-comment-threads.md` | `docs/adr/0001-nested-comment-threads.md` | Decision to support two-level comment threads. |
| `docs/links/active/adr/0002-hybrid-search-as-core-feature.md` | `docs/adr/0002-hybrid-search-as-core-feature.md` | Decision to ship full-text + Qwen/pgvector hybrid search with RRF. |
| `docs/links/active/adr/0003-unix-like-content-tree.md` | `docs/adr/0003-unix-like-content-tree.md` | Decision to organize content as nested Directory/File tree. |
| `docs/links/active/adr/0004-path-redirects-for-tree-rewrites.md` | `docs/adr/0004-path-redirects-for-tree-rewrites.md` | Decision to preserve links with path redirects on published path changes. |
| `docs/links/active/adr/0005-sandbox-full-html-documents.md` | `docs/adr/0005-sandbox-full-html-documents.md` | Decision to render full HTML Documents in sandboxed iframes. |
| `docs/links/active/adr/0006-per-file-assets.md` | `docs/adr/0006-per-file-assets.md` | Decision to scope uploaded assets to individual Files. |

### Archived historical documents

| Soft link | Target | Purpose |
|---|---|---|
| `docs/links/archive/decisions/blog-design-decisions-2026-06-02.md` | `docs/archive/decisions/blog-design-decisions-2026-06-02.md` | Historical decision summary from the earlier planning phase. |
| `docs/links/archive/plans/blog-implementation-plan-2026-06-02.md` | `docs/archive/plans/blog-implementation-plan-2026-06-02.md` | Historical implementation plan superseded by `IMPLEMENTATION_PLAN.md`. |
| `docs/links/archive/plans/superpowers/2026-06-03-xlab-personal-blog.md` | `docs/archive/plans/superpowers/2026-06-03-xlab-personal-blog.md` | Heavy Superpowers micro-step plan superseded by the OMX-native plan. |
| `docs/links/archive/roadmaps/blog-learning-roadmap.md` | `docs/archive/roadmaps/blog-learning-roadmap.md` | Historical learning roadmap, not an implementation source. |

## Implementation Invariants

- Local development uses Conda environment `blogenv`.
- Backend follows `handler -> service -> repository -> db`; SQL belongs in repositories.
- API changes update `docs/api/openapi.yaml` first.
- Frontend is Vite React SPA; no Next.js/SSR, Redux, or dark mode.
- HTML Documents render only in iframe sandbox with `allow-scripts` and without `allow-same-origin`.
- Glass Ricepaper design stays warm, light-only, with one frosted-glass material and Action Blue as the only accent.
- Use Directory/File/Reader/Anonymous Visitor/Content Tree vocabulary from `docs/specs/CONTEXT.md`.
