# xLab Blog OMX Implementation Plan

This repo-root file is the execution plan for implementing the xLab blog. It is subordinate to `AGENTS.md` and the product specs listed below, and it supersedes the earlier micro-step plan at `docs/archive/plans/superpowers/2026-06-03-xlab-personal-blog.md` for day-to-day OMX execution.

> This replaces the overly heavy step-by-step Superpowers plan with an OMX-native delivery plan. Keep the original product philosophy, but let OMX own orchestration, dispatch, internal communication, verification loops, and recovery. Individual agents may invoke Matt/Superpowers skills when useful for their local task; those skills are tools, not global process law.

## 0. Target Outcome

Build the full-stack xLab Personal Blog described by the repo specs:

- A single-author blog / knowledge space, not a portfolio or flat post list.
- Unix-like Content Tree: nested Directory/File paths under `/`.
- File render formats: sanitized Markdown reading pages and sandboxed full HTML Documents.
- Reader/admin auth, comments, likes, per-file assets, hybrid search, admin Tree Manager.
- Vite React SPA + Go/chi API + PostgreSQL/pgvector + Caddy + Docker Compose.
- Glass Ricepaper visual identity from `docs/specs/DESIGN.md`: warm paper, one frosted-glass material, single Action Blue, no dark mode.

Success means the acceptance signals in `docs/specs/PRD.md` pass, especially Docker startup, admin content creation, public path resolution, Markdown/HTML rendering safety, assets, search fallback, auth redirects, and high-risk tests.

## 1. Source of Truth

Read these before implementation or review:

1. `docs/specs/PRD.md` — product scope and acceptance.
2. `docs/specs/BLOG_FLOW.md` — routes and user flows.
3. `docs/specs/TECH_STACK.md` — exact versions and prohibited substitutions.
4. `docs/specs/BACKEND_STRUCTURE.md` — backend layout, DB schema, rules.
5. `docs/api/openapi.yaml` — API contract; update this before route changes.
6. `docs/specs/DESIGN.md` and `docs/design/glass-light-v2.html` — visual system.
7. `docs/specs/CONTEXT.md` — product vocabulary.
8. `docs/adr/*.md` — architectural decisions.

Do not resurrect the old flat article/category model. Use Directory, File, Reader, Anonymous Visitor, Content Tree, HTML Document, Hybrid Search, and Glass Ricepaper consistently.


Architecture constraints that are not negotiable:

- Backend dependency direction is `handler -> service -> repository -> db`; SQL belongs in repositories.
- Frontend routes are `/`, `/recent`, `/search`, `/login`, `/register`, `/admin`, and catch-all `/*`.
- Public/admin API routes must stay aligned to `docs/api/openapi.yaml` before implementation changes land.


## 2. Local Environment Contract

Local development uses Conda. The environment name is fixed:

```bash
conda create -n blogenv
conda activate blogenv
```

If an `environment.yml` is created, it must declare:

```yaml
name: blogenv
```

Then install or expose the exact toolchain required by `docs/specs/TECH_STACK.md` inside `blogenv`:

- Node.js `22.22.3`
- npm `10.9.8`
- Go `1.26.4`
- Docker Compose v2 CLI available to the shell

If Conda cannot resolve exact Node/Go versions, stop that lane and report the solver error. Do not silently substitute versions; update `docs/specs/TECH_STACK.md` only after an explicit decision.

Stack substitutions disallowed by `docs/specs/TECH_STACK.md` stay disallowed unless the spec changes: no Next.js/Remix/SSR, no Redux, no CRA, no gin/echo/gorilla, no ORM/GORM, no `dgrijalva/jwt-go`, no Compose v1 `docker-compose`, no nginx edge, and no first-release object storage implementation.

## 3. Execution Philosophy

The previous plan encoded every small action as a Matt skill checkbox. This plan uses a lighter rule:

- OMX owns the workflow: task decomposition, multi-agent dispatch, ownership, communication, state, verification, and retry.
- Agents own bounded work packets with explicit file scopes and acceptance criteria.
- Matt/Superpowers skills are invoked inside a packet only when they materially help:
  - `superpowers:using-git-worktrees` for isolated execution branches.
  - `superpowers:test-driven-development` for security/business-rule code.
  - `superpowers:systematic-debugging` for failing tests or regressions.
  - `superpowers:subagent-driven-development` or `superpowers:executing-plans` for a packet that needs local task-by-task execution.
  - `superpowers:requesting-code-review`, `superpowers:receiving-code-review`, and `superpowers:finishing-a-development-branch` near integration.
- Do not require every agent to use every skill. Pick the smallest useful skill surface.
- Prefer vertical, verifiable slices over giant layer-only work.
- Keep diffs reviewable. Use the Lore commit protocol from `AGENTS.md`.

## 4. OMX Agent Roster

OMX should run at least these lanes. More can be added only when parallelism is real and file ownership is clear.

### 4.1 Orchestrator / Dispatcher

**Primary role:** `planner` or Ralph leader.

Owns:

- Reads specs and keeps the shared goal current.
- Splits work into packets with file ownership.
- Dispatches agents, tracks status, resolves conflicts.
- Maintains `.omx/state/*`, `.omx/notepad.md`, and handoff notes.
- Integrates work and decides when to verify or debug.

Does not own:

- Large implementation patches unless a lane is blocked.
- Silent scope cuts.

### 4.2 Architecture / Contract Agent

**Primary role:** `architect`.

Owns:

- DB schema and API boundary review.
- OpenAPI-first checks.
- Package/version compliance with `docs/specs/TECH_STACK.md`.
- Security boundaries: JWT, sandbox iframe, asset rules, search fallback.

Matt skills likely used:

- `superpowers:requesting-code-review` for contract reviews.
- `superpowers:receiving-code-review` when findings return.

### 4.3 Development Agents

**Primary role:** `executor` / `team-executor`.

Suggested independent lanes:

- Backend foundation and auth.
- Content tree, file content, redirects.
- Comments, likes, assets.
- Search and embeddings.
- Frontend shell, routes, UI components.
- Admin UX and upload/edit flows.
- Docker/Caddy deployment.

Owns:

- Code in assigned file scopes only.
- Unit/integration tests for changed behavior.
- Clear final report: files changed, tests run, blockers.

Matt skills likely used:

- `superpowers:test-driven-development` for auth, likes, render safety, assets, redirects, and search failure handling.
- `superpowers:executing-plans` for packet-local checklists.

### 4.4 Acceptance / Verification Agent

**Primary role:** `verifier` or `test-engineer`.

Owns:

- Test adequacy and acceptance evidence.
- PRD acceptance matrix.
- Fresh verification commands and output reading.
- Regression checks after integration/deslop.

Must verify at minimum:

- `cd api && go test ./...`
- `cd web && npm run lint && npm run build`
- `docker compose config`
- `docker compose up -d --build` smoke when Docker is available
- `curl -fsS http://localhost:8080/api/health`
- Browser/manual checks for admin creation, path resolve, iframe sandbox, comments/likes, assets, search.

### 4.5 Debug Agent

**Primary role:** `debugger`.

Activated when:

- Tests fail after a normal fix attempt.
- Docker/build dependency issues are unclear.
- Search, path redirects, auth, asset serving, or iframe behavior diverges from specs.

Owns:

- Root-cause analysis.
- Minimal reproduction.
- Fix recommendation or patch within an assigned scope.

Matt skills likely used:

- `superpowers:systematic-debugging`.

### 4.6 Design QA Agent

**Primary role:** `designer` or `vision` when screenshots exist.

Owns:

- Glass Ricepaper fidelity.
- UI consistency across nav, cards, reading page, comments, drawer, admin forms.
- No dark mode, no extra accent colors, no cold glow/orbs.

## 5. Communication Protocol

Use OMX as the team bus. Every agent report should be short and structured:

```txt
Role:
Packet:
Files owned:
Status: done | blocked | needs-review
Changed files:
Verification run:
Evidence:
Blockers / risks:
Next recommended handoff:
```

Rules:

- One writer per file at a time.
- Shared contracts (`openapi.yaml`, migrations, generated types, core API client) require orchestrator approval before edits.
- Agents must not revert other agents' changes.
- If a task needs a dependency/version change, pause that packet and route through Architecture / Contract Agent.
- If a packet changes API response shapes, update `docs/api/openapi.yaml` first, then backend, then frontend types/usages.
- Keep handoff notes in `.omx/notepad.md` or packet-local markdown under `.omx/plans/` when useful.


## 6. Lean Phase Map

Use phases for orchestration and packets for ownership. OMX may run non-conflicting packets in parallel inside a phase.

| Phase | Purpose | Main packets |
|---|---|---|
| Phase 0 | Contracts, environment, repo guardrails | A |
| Phase 1 | Backend foundation, DB schema, auth/admin seed | B, C |
| Phase 2 | Content tree, path resolution, publish/unpublish, redirects | D |
| Phase 3 | Render safety, comments/likes, assets, hybrid search | E, F, G, H |
| Phase 4 | Public frontend shell, routes, auth flows, Glass Ricepaper UI | D frontend, E frontend, F frontend |
| Phase 5 | Admin tree/file/assets/search UI | I |
| Phase 6 | Docker Compose + Caddy deployment | J |
| Phase 7 | Acceptance matrix, smoke walkthrough, debug loop | Verification + Debug lanes |

Phase gates:

- Phase 1 cannot complete until backend tests compile and migrations reflect `docs/specs/BACKEND_STRUCTURE.md`.
- Phase 2 cannot complete until public path resolution and redirect behavior match `docs/specs/BLOG_FLOW.md`.
- Phase 4 cannot complete until iframe sandbox and Markdown sanitization are verified.
- Phase 7 cannot complete until the verification matrix in Section 8 has fresh evidence.

## 7. Implementation Packets

Packets are intentionally larger than the old plan's 2–5 minute steps. OMX should decompose them internally when dispatching.

### Packet A — GitHub Repo Structure and Environment Foundation

Owner: Orchestrator + Development Agent.

Deliverables:

- GitHub-friendly root: keep `README.md`, `IMPLEMENTATION_PLAN.md`, `.gitignore`, and future runtime entry files at root.
- Active docs under `docs/specs/`, `docs/api/`, `docs/adr/`, and `docs/design/`; historical material under `docs/archive/`.
- `.gitignore` for env/build/data/runtime artifacts.
- `environment.yml` or documented Conda setup for `blogenv`.
- Root `README.md` with setup, verification, source-doc index, and repository layout.
- Initial `docker-compose.yml`, `Caddyfile`, `.env.example` skeleton can land here or in Packet H.

Acceptance:

- Root directory is not crowded with historical planning/spec files.
- Active implementation specs are discoverable from `docs/README.md` and root `README.md`.
- `conda activate blogenv` is the documented local entrypoint.
- No implementation artifact fights `.gitignore`.
- README points to the active source docs and states `docs/archive/` is historical only.

### Packet B — Backend Foundation and Schema

Owner: Backend Development Agent. Reviewer: Architecture Agent.

Deliverables:

- `api/go.mod` with exact versions from `docs/specs/TECH_STACK.md`.
- Go server bootstrap under `api/cmd/server`.
- Config loading with required `DATABASE_URL`, `JWT_SECRET`, admin seed envs, asset/search envs.
- pgx pool and migrations.
- Migration for users, nodes, file_contents, path_redirects, comments, likes, file_assets, extensions, indexes.
- Health endpoint.

Acceptance:

- `cd api && go test ./...` passes for available unit tests.
- Migration matches `docs/specs/BACKEND_STRUCTURE.md` unless a documented ADR/spec update exists.
- No SQL in handlers.

### Packet C — Auth and Role Guards

Owner: Backend Development Agent. Reviewer: Verifier.

Deliverables:

- Register/login/me endpoints.
- bcrypt password hashing.
- JWT issue/parse with `sub`, `role`, `email`, `exp`.
- Optional auth, require auth, require admin middleware.
- Admin seed from env creates or upgrades admin.

Acceptance:

- Public register always creates `reader`.
- Invalid/tampered/expired JWT is rejected.
- Admin routes reject non-admin.
- Passwords and secrets are never logged.

Recommended Matt skill: `superpowers:test-driven-development`.

### Packet D — Content Tree and File Lifecycle

Owner: Backend Development Agent + Frontend Development Agent as separate file scopes.

Backend deliverables:

- Node create/read/update/delete services.
- Public `GET /api/tree`, `GET /api/tree/resolve`, `GET /api/tree/{node_id}/children`.
- Admin node CRUD.
- File content upsert, publish, unpublish.
- Path normalization and redirect rules for published File or published descendant path changes.
- Deletion rules: published File and Directory with published descendants cannot hard-delete.

Frontend deliverables:

- Root Directory Page.
- Catch-all Content Path Resolver.
- Breadcrumb.
- Directory/File `content-entry-card`.
- Redirect replace navigation.

Acceptance:

- Root `/` shows published root entries only.
- Nested Directory and File paths resolve through SPA routing.
- Draft Files are not public.
- Reserved root slugs are rejected.
- Published path changes create non-chained redirects.

Recommended Matt skill: `superpowers:test-driven-development` for redirect/deletion rules.

### Packet E — Rendering and Glass Ricepaper Public UI

Owner: Frontend Development Agent + Design QA Agent. Backend support for render/search text.

Deliverables:

- Shared glass tokens and `.glass` primitive.
- `GlassNav`, Directory Drawer, `ContentEntryCard`, `FilePage`, reading card.
- Markdown rendering via `marked` + `DOMPurify`.
- HTML Document iframe via `sandbox="allow-scripts"` with no `allow-same-origin`.
- Backend Markdown `search_text` and optional sanitized `body_html`.
- Backend HTML visible text extraction excluding script/style/meta/link/noscript/hidden.

Acceptance:

- UI follows `docs/specs/DESIGN.md`: warm paper, one frosted-glass material, Action Blue only, no dark mode.
- Markdown XSS test passes.
- HTML Document never enters main React DOM.
- Iframe sandbox attribute is exact and safe.

### Packet F — Reader Interactions: Comments and Likes

Owner: Backend + Frontend Development Agents.

Deliverables:

- Two-level comment thread endpoints and UI.
- Reply-to-reply normalization to top-level parent with `reply_to_user_id`.
- Soft delete for user-owned comments and admin deletes.
- Idempotent likes/unlikes for Files and Comments.
- Anonymous read and login redirect for write actions.

Acceptance:

- Anonymous visitors can read comments/like counts.
- Anonymous write actions go to `/login?return_to=current_path`.
- Repeated like/unlike is stable.
- Comment bodies are plain text only.

Recommended Matt skill: `superpowers:test-driven-development`.

### Packet G — Per-File Assets

Owner: Backend Development Agent + Frontend/Admin Development Agent.

Deliverables:

- `AssetStorage` interface and `LocalAssetStorage`.
- Admin upload/delete endpoints.
- Public immutable endpoint `/api/assets/{asset_id}/{filename}` for published File assets.
- MIME and size allowlist.
- SVG safety rejection for script/event/javascript/external/foreignObject.
- Admin Asset Manager UI.

Acceptance:

- Published assets serve with strong immutable cache.
- Draft assets are not publicly exposed.
- Storage keys are provider-neutral and local absolute paths are never exposed.
- Malicious SVG fixtures are rejected.

### Packet H — Hybrid Search

Owner: Search Development Agent. Reviewer: Architecture + Verifier.

Deliverables:

- Full-text search with `websearch_to_tsquery('simple', q)`, rank, snippet.
- Qwen/DashScope OpenAI-compatible embedding provider.
- pgvector semantic search using `vector(1024)`.
- RRF fusion with default `k=60`.
- Admin refresh/rebuild endpoints.
- Frontend `/search?q=` result page with path/snippet/source badges.

Acceptance:

- Search only returns published Files.
- Embedding request sends `dimensions: 1024` and `encoding_format: "float"`.
- Qwen failure records failed state but does not fail content save.
- Full-text fallback still returns results when embeddings are unavailable.

### Packet I — Admin Tree Manager

Owner: Frontend Development Agent + Backend Admin endpoints owner.

Deliverables:

- `/admin` route protected for admin.
- File-manager style tree manager.
- Create/edit/move/delete Directory/File.
- File editor for Markdown and HTML Document.
- Publish/unpublish controls.
- Asset upload/delete panel.
- Embedding refresh/rebuild controls.
- Impact prompts for dangerous operations.

Acceptance:

- Admin can create Directory and File, edit content, upload assets, publish, unpublish.
- Moving/renaming published content prompts and creates redirects.
- Published File cannot directly change `content_format`.
- No drag-and-drop requirement.

### Packet J — Deployment and Smoke

Owner: Infra Development Agent. Reviewer: Verifier.

Deliverables:

- `docker-compose.yml` with `db`, `api`, `web`.
- Caddy serving SPA and reverse-proxying `/api`.
- API and web Dockerfiles.
- Uploads and Postgres named volumes.
- `.env.example`.

Acceptance:

- `docker compose config` passes.
- `docker compose up -d --build` starts db/api/web.
- `curl -fsS http://localhost:8080/api/health` returns OK.
- SPA routes are served by Caddy fallback.

## 8. Required Verification Matrix

Verifier owns the final matrix. Development agents own local proof for their packets.

| Area | Required evidence |
|---|---|
| Backend compile/tests | `cd api && go test ./...` |
| Frontend lint/build | `cd web && npm run lint && npm run build` |
| Docker config | `docker compose config` |
| Stack smoke | `docker compose up -d --build`; health curl; SPA curl |
| JWT/security | tests for tampered/expired token and role guard |
| Likes | idempotent like/unlike tests |
| Markdown | XSS sanitization test |
| HTML Document | iframe sandbox check: `allow-scripts`, no `allow-same-origin` |
| Assets | malicious SVG rejection tests; MIME/size tests |
| Tree | reserved root slug, path resolve, path redirect, deletion rules |
| Search | RRF unit test; Qwen failure does not block save; full-text fallback |
| UI | manual/design QA against Glass Ricepaper tokens and prototype |

Do not claim completion from “looks good.” Completion needs fresh command output and an acceptance matrix mapping PRD signals to evidence.

## 9. Integration Order

Recommended OMX execution order:

1. Packet A — repo/environment.
2. Packet B — backend foundation/schema.
3. Packet C — auth.
4. Packet D backend half — tree/content lifecycle.
5. Packet E backend render/search-text half.
6. Packet D/E frontend half — public shell, tree, file rendering.
7. Packet F — comments/likes.
8. Packet G — assets.
9. Packet H — search.
10. Packet I — admin.
11. Packet J — deployment.
12. Full acceptance and debug loop.

Parallelism rule:

- Backend schema/API and frontend visual shell can run in parallel after Packet A, but shared API shape changes must go through OpenAPI first.
- Comments/likes, assets, and search can run in parallel once auth + file IDs + content lifecycle are stable.
- Admin UI should wait until backend admin endpoints are stable enough to avoid rework.

## 10. Debug and Recovery Protocol

When a packet fails:

1. Orchestrator records failure summary and assigns Debug Agent.
2. Debug Agent uses `superpowers:systematic-debugging` when root cause is not obvious.
3. Debug Agent produces:
   - failing command,
   - smallest repro,
   - suspected layer,
   - patch or recommended owner handoff.
4. Original owner fixes within assigned scope.
5. Verifier reruns the failed command and any nearby regression checks.

Do not delete tests to make a lane pass. Do not bypass security checks to unblock UI.

## 11. Commit and Review Protocol

Use small, packet-level commits. Commit messages must follow the Lore protocol from `AGENTS.md`:

```txt
<why this change was made>

Constraint: <external constraint>
Confidence: <low|medium|high>
Scope-risk: <narrow|moderate|broad>
Directive: <future warning>
Tested: <commands/evidence>
Not-tested: <known gaps>
```

Before integration:

- Orchestrator reviews changed files and ownership conflicts.
- Architecture Agent reviews API/schema/security changes.
- Verifier reviews tests and acceptance evidence.
- Design QA reviews visual pages when frontend UI changes.

## 12. Done Definition

The project is complete only when all are true:

- Docker Compose starts web/api/db and Caddy serves SPA + `/api` proxy.
- Admin seed account can log in and manage Directory/File/content/assets/publish/search refresh.
- Public `/`, `/recent`, `/search`, nested Directory paths, and File paths work.
- Published path changes return redirect through `/api/tree/resolve`.
- Markdown renders sanitized content in the reading card.
- HTML Document renders in fixed-height sandbox iframe with JS allowed and same-origin denied.
- Assets upload, serve, cache, and reject unsafe content per spec.
- Comments and likes work with anonymous read and login redirect for writes.
- Hybrid search returns path/snippet/source badges and degrades to full-text without Qwen.
- Required high-risk tests pass.
- UI preserves Glass Ricepaper identity.
- Verifier has fresh evidence for the matrix in Section 8.

## 13. What This Plan Deliberately Removes

Compared with `docs/archive/plans/superpowers/2026-06-03-xlab-personal-blog.md`, this plan removes:

- Mandatory code blocks for every micro-step.
- Universal Matt skill usage.
- Overly prescriptive 2–5 minute task granularity.
- Premature exact implementation snippets for files that agents should design after reading current state.
- The assumption that one linear session should execute the whole product.

It keeps:

- Product scope.
- Tech stack constraints.
- Conda `blogenv` requirement.
- Glass Ricepaper design philosophy.
- OpenAPI-first backend/frontend contract.
- Security and acceptance gates.
- Frequent verification and small commits.
