# Progress

Last updated: 2026-06-14 00:04 CST

## Current breakpoint

Stage 3 is active on Team `execute-aeolian-blog-a98ab708` from protected Stage 2 checkpoint `73bcc9e` (`checkpoint: stage 2 polish before stage 3`). Current coordinator worktree HEAD observed while reconciling the ledger: `8189832`; leader-reported integrated backend security PASS: `dd2b493`.

Gateway status:

- **Gateway 0 — PASS.** Stage 2 checkpoint protected; pre-schema backup/restore proof recorded in `docs/verification/stage-3-team-log.md`.
- **Gateway 1 — PASS.** Contract acceptance review task 11 completed with downstream fixture/watch items recorded in `docs/verification/stage-3-acceptance.md`.
- **Gateway 2 — PASS.** Worker-2 task 9 completed migration/core publication model; security REVISE repair task 14 is now completed and accepted by post-repair review task 15.
- **Gateway 3 — PASS.** Worker-2 task 10 completed version-state, restore, publish summary/publish/unpublish, Draft Preview, and asset-state HTTP surfaces; backend security repair is accepted after task 15.
- **Gateway 4 — PASS for frontend implementation.** Worker-3 task 4 completed autosave/version/restore/publish/Draft Preview/draft-published assets UI at commit `9b51d36`; frontend lint/build/contracts/node tests passed.
- **Security implementation review — backend repair accepted.** Task 12 remains a historical REVISE verdict with findings recorded in `docs/verification/stage-3-security.md`; task 14 completed the backend repair and task 15 post-repair security re-review passed at leader-reported integrated SHA `dd2b493`. MCP-specific security remains a later Gateway 6 gate.
- **Gateway 6 — MCP in progress.** Task 16 server-local stdio MCP skeleton is now in progress on worker-2; tasks 17-19 track MCP tools, acceptance smoke, and security review; task 20 is final integrated acceptance/closeout after MCP gates pass.

Current verification note (2026-06-13 23:58 CST): frontend Gateway 4 verification passed on worker-3 (`npm run lint`, `npm run build`, Stage 3 frontend contract, and all web node tests). Backend security REVISE repair is complete: task 14 records backend `go test`/`go vet`/gofmt PASS evidence, and task 15 records post-repair security review PASS at `dd2b493`. MCP implementation remains pending and must be accepted only after integrated MCP SHAs land.

Current task reconciliation (2026-06-13 23:48 CST): tasks 1,3,4,5,6,7,8,9,10,11,13,14,15 are completed; task 12 is terminal failed with the security REVISE that spawned task 14; task 2 remains the only active coordinator ledger task. No implementation/acceptance closeout task for MCP is complete yet, so Stage 3 is **not** final-closeout complete.

Evidence ledger:

- Coordinator log: `docs/verification/stage-3-team-log.md`.
- MCP Gateway 6 skeleton evidence: `docs/verification/stage-3-mcp-gateway6-skeleton.md`.
- Acceptance plan: `docs/verification/stage-3-acceptance.md`.
- Security plan/review: `docs/verification/stage-3-security.md`.
- Frontend readiness plan: `docs/verification/stage-3-frontend-readiness.md`.
- OpenAPI contract: `docs/api/openapi.yaml`.
- Backend Gateway 1 expected-red tests: `api/internal/tree/stage3_gateway1_contract_test.go`, `api/internal/http/handlers/stage3_gateway1_contract_test.go`, `api/internal/http/stage3_gateway1_contract_test.go`, `api/internal/search/stage3_gateway1_contract_test.go`, `api/internal/assets/stage3_gateway1_contract_test.go`.
- Backend security regression tests added by task 14: `api/internal/http/handlers/stage3_security_regression_test.go`, `api/internal/comments/stage3_security_regression_test.go`, `api/internal/likes/stage3_security_regression_test.go`.
- Frontend Gateway 1/Gateway 4 contract test: `web/tests/stage3-author-workspace-contract-red.test.mjs`.

Gateway 0 backup before schema work:

- Backup directory: `~/.local/share/xlab-blog/backups/stage-3-gateway0-20260613T221830+0800`.
- Artifacts: `xlab_blog.dump`, `uploads.tgz`, `SHA256SUMS.txt`.
- Disposable restore proof: `xlab_blog_restore_stage3_20260613222147` restored successfully and was dropped; restored row counts matched current empty local DB state (`nodes=0`, `file_contents=0`, `file_assets=0`).

Current coordination constraints:

- Do not add/commit `web/node_modules`, node_modules symlinks, caches, `web/dist`, local DB/uploads, or `.omx` runtime state.
- Production frontend Gateway 4 UI is implemented; do not run acceptance/security closeout against non-integrated branches.
- Acceptance/security tests run only against integrated leader SHAs, not isolated worker branches.
- Gateway 6 MCP work is split across tasks 16-19; keep final closeout blocked until MCP implementation, acceptance, and security review are integrated and passed.
- Preserve iframe sandbox, full-text fallback, Author-only protected surfaces, draft/public isolation, and the Stage 2 simple-English Aeolian UI baseline.
- Keep MCP security/acceptance gates blocked until an actual server-local stdio MCP package exists with explicit enablement, per-call kill switch, audit JSONL, backup/export, and no direct SQL in handlers.

Required verification baseline for every non-red integration checkpoint remains:

```bash
cd api
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./...
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go vet ./...
test -z "$(gofmt -l .)"

cd web
node --test tests/*.test.mjs
npm run lint
npm run build
```

## Immediate next steps

1. Monitor MCP Gateway 6 task chain: task 16 skeleton is in progress on worker-2; task 17 tool slices, task 18 acceptance smoke/evidence, and task 19 security review remain pending/blocked as designed.
2. Keep MCP security/acceptance gates blocked until an actual server-local stdio MCP package exists with explicit enablement, per-call kill switch, audit JSONL, backup/export, no public HTTP/SSE listener, and no direct SQL in handlers.
3. After MCP acceptance/security pass, let task 20 run final integrated acceptance and closeout; Task 2 remains open until terminal task counts and closeout gates pass.
4. Keep `PROGRESS.md` and `docs/verification/stage-3-team-log.md` synchronized after each integration or gate decision.

## Previous notes
# xLab Blog Progress

Last updated: 2026-06-13 15:13 CST

This is the durable resume point. Keep it concise and update it after every key milestone.

## Current breakpoint

- Branch: `main`; local commits are ahead of `origin/main`.
- Active plan: `docs/plans/SECOND_DEVELOPMENT.md`.
- Active Team: `execute-approved-xlab-015f30a9` launched with the exact approved command `omx team 5 "Execute approved xLab Blog second-development Stage 2 DAG"`.
- Stage 2 implementation tasks 9-15 are integrated. A small integration repair `59311dd` fixed a duplicate lifecycle test fake method and restored backend `go test`/`go vet`/gofmt.
- Gateway 6 integrated acceptance (task 16) is **FAILED** and documented in `docs/verification/stage-2-acceptance.md`. Backend/frontend gates and native runtime smoke passed, but Author Workspace desktop/mobile acceptance is blocked because `/api/admin/tree` returns `{ nodes, url_path }` while the frontend expects `{ roots, path, children }`, causing `内容树加载失败`.
- Active repair: task 21 `Repair Gateway 6 admin tree contract drift` is claimed by worker-2. After repair integration, create/reopen a fresh Gateway 6 rerun task for worker-4 because task 16 is terminal failed.
- Active review: task 17 `Gateway 7 integrated security and abuse review` is still in progress on worker-5.
- Blocked downstream: tasks 18-20 remain pending until the repaired Gateway 6 rerun and Gateway 7 both pass.
- Current product baseline before Stage 2 implementation: `cabf9a497a7ce1253e99824b6eb8605ba029d813`; Gateway 1 fixture root `/stage-2-acceptance` remains the native acceptance fixture.

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


- 2026-06-13: Gateway 2 OpenAPI/backend Red contracts passed and was integrated. OpenAPI was updated first (`83dc5f4`) with `/admin/tree`, minimal slugless create, `url_path` settings update, reorder, move preview/commit, delete reason contracts, and Author 401/403 semantics. Backend Red tests were added (`020c85d`) and currently fail as intended on missing Stage 2 API/types/errors; task 9 and task 11 are now active.

- 2026-06-13: Gateway 1 backup/restore/fixture passed and was merged. Evidence: `docs/verification/stage-2-backup-and-fixture.md` and `docs/verification/stage-2-acceptance.md`. Backup directory: `~/.local/share/xlab-blog/backups/stage-2-gateway1-20260613T134736+0800`; disposable restore passed; fixture root `/stage-2-acceptance` recorded; public draft isolation smoke passed. Gateway 2 backend OpenAPI/Red contracts is now active.

- 2026-06-13: Stage 2 Team `execute-approved-xlab-015f30a9` Gateway 0 launch/decomposition audit passed. Evidence is in `docs/verification/stage-2-team-log.md`; next gate is Gateway 1 backup/fixture, and product implementation remains blocked until the gate chain clears.

- 2026-06-13: Second Critic probe round returned `REVISE`; fixes applied in working tree: clarified Stage 2 has only `草稿` / `已发布` and defers `有未发布修改` / `发布更新`; chose conservative deletion rule blocking every non-empty Directory; made `GET /admin/tree` the complete protected Stage 2 tree; strengthened OpenAPI field/error requirements; aligned active specs; rewrote ignored `.omx` Stage 2 PRD/packet DAG runtime copies to current Author Workspace scope.
- 2026-06-13: Wrong Team launch abort: a free-form Stage 2 launch produced `legacy_text` decomposition and was force-shutdown before integration; worker diffs were empty/noop. Plan launch guard now requires the exact approved Stage 2 DAG hint and decomposition-source verification before implementation.
- 2026-06-13: Ran a Critic probe on the Stage 2 plan. Verdict was `REVISE`, not implementation-blocking after fixes.
- Applied fixes: unified Team launch to five persistent seats plus on-demand repair; added Stage 2/Stage 3 boundary precedence; added minimum OpenAPI contract table and stronger backend Red cases; added public Directory `管理此目录` acceptance/security checks; added disposable restore proof condition; clarified Author-role-only public entry data flow.

## Immediate next steps

1. Monitor worker-2 task 21: fix the protected Author tree contract drift (`/api/admin/tree` response shape vs frontend parser), add regression coverage, and run relevant backend/frontend gates.
2. Monitor worker-5 task 17: finish integrated security/abuse review and record `docs/verification/stage-2-security.md`.
3. After task 21 is integrated, create a fresh Gateway 6 rerun task for worker-4, rerun desktop/mobile acceptance on the repaired HEAD, and replace the current failed acceptance result with rerun evidence.
4. Only after repaired Gateway 6 and Gateway 7 pass, unblock worker-1 tasks 18-20 for architecture review, code review, final gates, progress closeout, and user acceptance handoff.
5. Keep `PROGRESS.md` and `docs/verification/` updated at every key milestone and before stopping.

## Key evidence and history

- Baseline evidence: `docs/verification/BASELINE.md`, `docs/verification/native-local-full-stack-smoke-20260606.md`.
- Stage 1 evidence: `docs/verification/stage-1-*` and `docs/verification/stage-1-browser-20260612/*`.
- Historical implementation summary: `docs/archive/INITIAL_BUILD_SUMMARY.md`.
- Detailed older Team/task history is in Git history; do not revive stale OMX runtime state.
