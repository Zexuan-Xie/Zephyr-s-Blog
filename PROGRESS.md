# xLab Blog Progress

Last updated: 2026-06-13 02:34 CST

This is the durable resume point. Keep it concise and update it after every key milestone.

## Current breakpoint

- Branch: `main`; local commits are ahead of `origin/main`.
- Active plan: `docs/plans/SECOND_DEVELOPMENT.md`.
- Current breakpoint: Stage 2 is planned and ready for execution planning/Team launch when the user explicitly instructs. Do not start implementation yet.
- Stage 1 engineering is complete and reviewed, but user acceptance did not pass because Author Workspace UX was inadequate; its blockers are now Stage 2 scope.
- Current product baseline before Stage 2 implementation: Stage 1 closeout commits through `9b45f6b` documentation updates; recheck `git log` before starting.
- Runtime services were healthy at the Stage 1 handoff; recheck before verification.

## Active delivery stages

1. **Stage 1 — Reliability/navigation/identity**: engineering complete; acceptance feedback folded into Stage 2.
2. **Stage 2 — Chinese Author Workspace and protected Content Tree**: current active target. Desktop-first; mobile no-regression sanity only.
3. **Stage 3 — Autosave/publication model, Draft Preview, Draft/Published Assets, and Blog MCP Server**: final product stage.

Each stage must finish runnable, reversible, tested, documented, independently reviewed, and user-accepted before the next begins.

## Stage 2 scope summary

Stage 2 replaces the current form-heavy Admin page with a Chinese **Author Workspace**:

- protected complete Content Tree showing Directories, Draft Files, Published Files, and Files with unpublished changes;
- minimal create flow: Directory = `名称`; File = `名称` + `格式`; URL Path generated automatically with Chinese preserved;
- create success immediately refreshes tree/navigation, expands parent, selects/opens new node, and shows clear Chinese toast/path feedback;
- Directory overview workspace with child cards and new Directory/File actions;
- File workspace shell with `内容` / `资源` / `设置`, manual save, single primary publish action, and secondary `撤回发布`;
- Settings sections: `基础信息`, `位置`, `危险操作`; no Parent ID, Node ID, Sort order, or `slug` in primary UI;
- explicit return buttons and lightweight breadcrumbs for subflows;
- Author-only public entries: `管理此目录` and `编辑文件` return to the workspace with the target selected;
- same-parent desktop drag sorting only; drag never reparents;
- dedicated fixture root such as `/stage-2-acceptance`;
- presentation/defense quality: readable, extensible, clearly layered code and documentation.

See `docs/plans/SECOND_DEVELOPMENT.md` Section 4 for the controlling Stage 2 plan.

## Stage 3 / MCP summary

Stage 3 adds autosave, Current/Previous Content Versions, independent Published Content, Draft Preview, Draft/Published Assets, and a separate server-local stdio Blog MCP Server.

MCP decisions:

- independent MCP Server process/package, not embedded in Web UI;
- server-local stdio for trusted AI agents; no public HTTP/SSE transport initially;
- high-trust full Author permissions;
- tools: read, content, publish, tree, assets, maintenance;
- safeguards: explicit enablement config, operation audit logs, backup/export before destructive batches where practical, emergency disable;
- reuse backend service/API-client capabilities; do not duplicate business logic or SQL.

## Scope lock

Do not proactively redesign public homepage, Recent cards, public Directory/File reading, comments/Likes, or Glass Ricepaper. Only repair regressions or add the required Author-only public manage/edit entry.

Preserve:

- iframe `sandbox="allow-scripts"` without `allow-same-origin`;
- full-text search fallback when semantic indexing is unavailable;
- product language from `docs/specs/CONTEXT.md`.

## Local environment and recovery

Use Conda environment `blogenv`:

- Node.js `22.22.3`
- npm `10.9.8`
- Go `1.26.4`
- PostgreSQL `17.10`
- pgvector `0.8.1`

Persistent local state is outside Git at `~/.local/share/xlab-blog`.

Recover services:

```bash
~/.local/share/xlab-blog/start-local.sh
curl -fsS http://127.0.0.1:8080/api/health
curl -fsS http://127.0.0.1:5173/ >/dev/null
```

## Required verification

Backend:

```bash
cd api
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./...
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go vet ./...
test -z "$(gofmt -l .)"
```

Frontend:

```bash
cd web
node --test tests/render-safety.test.mjs
npm run lint
npm run build
```

For runtime/auth/tree/publication changes, also run native PostgreSQL API smoke and browser acceptance. Stage 2 requires desktop Author workflow acceptance plus mobile no-regression sanity. Record evidence under `docs/verification/`.


## Current execution plan artifact

- Stage 2 Team execution plan: `docs/plans/STAGE_2_TEAM_EXECUTION.md` (tracked canonical copy).
- Required seats: coordinator/gateway, backend developer, frontend developer, acceptance, security, repair/stabilization, plus independent architect and code-reviewer gates.
- Gateway flow: launch readiness → data/fixture safety → OpenAPI/backend Red → backend Green → frontend Red → frontend Green → integrated acceptance → security → independent review/closeout.

## Immediate next steps

1. Review `docs/plans/STAGE_2_TEAM_EXECUTION.md`, then await explicit user instruction before Stage 2 implementation.
2. If the user approves execution, launch OMX Team from that plan, starting with Gateway 0/1: clean status check, local database/uploads backup, and `/stage-2-acceptance` fixture setup.
3. Before coding, reread `AGENTS.md`, this file, `docs/plans/SECOND_DEVELOPMENT.md`, `docs/specs/CONTEXT.md`, relevant specs, and `docs/api/openapi.yaml` for API changes.
4. Update `PROGRESS.md` and `docs/verification/` at every key milestone and before stopping.

## Key evidence and history

- Baseline evidence: `docs/verification/BASELINE.md`, `docs/verification/native-local-full-stack-smoke-20260606.md`.
- Stage 1 evidence: `docs/verification/stage-1-*` and `docs/verification/stage-1-browser-20260612/*`.
- Historical implementation summary: `docs/archive/INITIAL_BUILD_SUMMARY.md`.
- Detailed older Team/task history is in Git history; do not revive stale OMX runtime state.
