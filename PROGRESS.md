# xLab Blog Progress

Last updated: 2026-06-13 13:42 CST

This is the durable resume point. Keep it concise and update it after every key milestone.

## Current breakpoint

- Branch: `main`; local commits are ahead of `origin/main`.
- Active plan: `docs/plans/SECOND_DEVELOPMENT.md`.
- Active Team: `execute-approved-xlab-015f30a9` launched with the exact approved command `omx team 5 "Execute approved xLab Blog second-development Stage 2 DAG"`.
- Current breakpoint: Gateway 0 launch/decomposition audit is PASS at `cabf9a4`; `decomposition_source=dag_sidecar`, active DAG `cmp` passed, five-seat mapping is recorded, and no launch role override such as `5:executor` was used.
- Stage 2 implementation has not started. Next gate: worker-4 Gateway 1 backup/restore proof and `/stage-2-acceptance` fixture after task 6 completes.
- Current product baseline before Stage 2 implementation: `cabf9a497a7ce1253e99824b6eb8605ba029d813`.
- Runtime services at Gateway 0: API `127.0.0.1:8080` and web `127.0.0.1:5173` are offline; restart before fixture/API/browser verification.

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
- Required persistent seats: coordinator/gateway, backend developer, frontend developer, acceptance, security. Repair/stabilization is an on-demand gateway-failure lane; independent architect and code-reviewer gates remain required.
- Gateway flow: launch readiness → data/fixture safety → OpenAPI/backend Red → backend Green → frontend Red → frontend Green → integrated acceptance → security → independent review/closeout.

## Latest planning review


- 2026-06-13: Stage 2 Team `execute-approved-xlab-015f30a9` Gateway 0 launch/decomposition audit passed. Evidence is in `docs/verification/stage-2-team-log.md`; next gate is Gateway 1 backup/fixture, and product implementation remains blocked until the gate chain clears.

- 2026-06-13: Second Critic probe round returned `REVISE`; fixes applied in working tree: clarified Stage 2 has only `草稿` / `已发布` and defers `有未发布修改` / `发布更新`; chose conservative deletion rule blocking every non-empty Directory; made `GET /admin/tree` the complete protected Stage 2 tree; strengthened OpenAPI field/error requirements; aligned active specs; rewrote ignored `.omx` Stage 2 PRD/packet DAG runtime copies to current Author Workspace scope.
- 2026-06-13: Wrong Team launch abort: a free-form Stage 2 launch produced `legacy_text` decomposition and was force-shutdown before integration; worker diffs were empty/noop. Plan launch guard now requires the exact approved Stage 2 DAG hint and decomposition-source verification before implementation.
- 2026-06-13: Ran a Critic probe on the Stage 2 plan. Verdict was `REVISE`, not implementation-blocking after fixes.
- Applied fixes: unified Team launch to five persistent seats plus on-demand repair; added Stage 2/Stage 3 boundary precedence; added minimum OpenAPI contract table and stronger backend Red cases; added public Directory `管理此目录` acceptance/security checks; added disposable restore proof condition; clarified Author-role-only public entry data flow.

## Immediate next steps

1. Complete coordinator task 6 and hand off Gateway 1 to worker-4; do not implement feature code before Gateway 1 passes.
2. Worker-4 Gateway 1: record database/uploads backup state, disposable restore proof or explicit N/A reason, `/stage-2-acceptance` fixture IDs/paths, and baseline content preservation.
3. After Gateway 1, backend task 8 must update OpenAPI first and add backend Red contracts before any Green implementation.
4. Before coding, reread `AGENTS.md`, this file, `docs/plans/SECOND_DEVELOPMENT.md`, `docs/specs/CONTEXT.md`, relevant specs, and `docs/api/openapi.yaml` for API changes.
5. Update `PROGRESS.md` and `docs/verification/` at every key milestone and before stopping.

## Key evidence and history

- Baseline evidence: `docs/verification/BASELINE.md`, `docs/verification/native-local-full-stack-smoke-20260606.md`.
- Stage 1 evidence: `docs/verification/stage-1-*` and `docs/verification/stage-1-browser-20260612/*`.
- Historical implementation summary: `docs/archive/INITIAL_BUILD_SUMMARY.md`.
- Detailed older Team/task history is in Git history; do not revive stale OMX runtime state.
