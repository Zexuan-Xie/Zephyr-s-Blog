# Stage 2 Author Workspace — OMX Team Execution Plan

Date: 2026-06-13 CST
Status: planning artifact; do not implement until user explicitly approves execution.

## 0. Evidence base

This plan is grounded in the current repository state:

- `AGENTS.md:4-14` defines read order and requires OpenAPI before shared API changes.
- `AGENTS.md:16-31` defines product language and current Stage 2 scope as Chinese Author Workspace / protected Content Tree.
- `AGENTS.md:33-45` requires SQL in repositories, iframe sandbox preservation, full-text fallback, backups, progress updates, and verification evidence.
- `PROGRESS.md:7-12` says Stage 2 is planned and must not start until explicit user instruction.
- `PROGRESS.md:23-38` summarizes current Stage 2 scope.
- `docs/plans/SECOND_DEVELOPMENT.md:78-195` is the controlling Stage 2 product plan.
- `docs/plans/SECOND_DEVELOPMENT.md:266-292` defines the Team evidence model and closeout gates.
- Current frontend blocker evidence: `web/src/pages/AdminPage.tsx:220-238` still renders old `ADMIN / Tree Manager`, `Node id`, and `Load selected node`; `web/src/pages/AdminPage.tsx:276-289` exposes Parent id, URL Path, Sort order, and English create UI.
- Current API starting point: `docs/api/openapi.yaml:306-380` defines existing `/admin/nodes` create/detail/update/delete contract; `api/internal/http/router.go:131-153` wires current protected admin endpoints.

## 1. Requirements summary

Deliver Stage 2 as a desktop-first Chinese **Author Workspace** replacing the current form-heavy Admin page.

Must deliver:

1. Protected Author Workspace Content Tree containing all Directories, Draft Files, and Published Files. `有未发布修改` is deferred to Stage 3 because it requires separate Published Content snapshots.
2. Minimal graphical creation flow: Directory = `名称`; File = `名称` + `格式`; generated URL Path with Chinese preserved.
3. Immediate post-create tree/navigation refresh, parent expansion, new-node selection/opening, and Chinese toast/path feedback.
4. Directory overview workspace with child cards and new Directory/File actions.
5. File workspace shell with `内容` / `资源` / `设置`, manual save, publish/unpublish controls, assets, and settings.
6. Settings with `基础信息`, `位置`, `危险操作`, graphical Directory Picker, explicit return buttons, and no Parent ID / Node ID / Sort order / `slug` in primary UI.
7. Author-only public Directory/File actions: `管理此目录` and `编辑文件`, returning to the workspace with the node selected.
8. Same-parent desktop drag sorting only; drag never reparents.
9. Dedicated Stage 2 fixture, desktop acceptance path, mobile no-regression sanity, security review, architect CLEAR, code-reviewer APPROVE.
10. Presentation-grade code quality: readable, extensible, layered, no duplicated SQL/business logic.

Minimum OpenAPI contract to define before backend implementation:

| Capability | Minimum protected contract | Notes |
|---|---|---|
| Author Content Tree | `GET /admin/tree` | Returns a complete protected tree in one call for the Stage 2 left tree: nested nodes or a flat list with parent IDs, but OpenAPI must choose one exact shape. It includes all Directories, Draft Files, and Published Files; status is `draft`/`published` only. Public APIs remain published-only. |
| Workspace detail | `GET /admin/nodes/{node_id}` | Returns Directory overview data or File content/settings/assets metadata needed by the right workspace. OpenAPI must specify exact Directory vs File response fields. |
| Minimal create | `POST /admin/nodes` | Request is context-aware: selected parent Directory + `kind` + `name` and File `content_format`; no client-supplied `slug` in minimal create. Backend generates URL Path segment, preserves Chinese, normalizes Latin text, and suffixes initial conflicts. Response returns final `url_path`. |
| Explicit settings update | `PATCH /admin/nodes/{node_id}` | Explicit URL Path edits use `url_path` / path segment fields defined by OpenAPI, return validation/conflict errors, and must not be silently rewritten. Reserved root segments are rejected. |
| Same-parent reorder | `PUT /admin/nodes/{parent_id}/children/order` | Accepts ordered child IDs for one parent only plus an explicit `version`/`etag` or documented lost-update guard; transactional; rejects reparent attempts. |
| Move preview/commit | `POST /admin/nodes/{node_id}/move-preview` and commit endpoint, or one clearly named pair | Directory Picker support: cycle prevention, destination path preview, affected published File descendant paths, redirects for formerly public File paths, subtree rewrite impact, and machine-readable blocked reasons. Directory paths themselves are not public redirect targets unless their published File descendants are affected. |
| Delete constraints | `DELETE /admin/nodes/{node_id}` response contract | Stage 2 blocks every non-empty Directory, including draft-only subtrees, and blocks Published Files. Return machine-readable reason plus Chinese-safe message source for UI. No recursive subtree deletion in Stage 2. |
| Publication state | existing publish/unpublish/content endpoints, documented if reused | Enough for manual save, publish, current `已发布` status, and secondary `撤回发布`; no `有未发布修改`, `发布更新`, or Stage 3 snapshots. |

Backend Red tests must explicitly cover Chinese-preserving URL Path generation, reserved root rejection, initial create conflict suffixes, explicit URL Path conflict without silent rewrite, exact protected-tree response shape, draft/published-only status fields, same-parent reorder transaction/lost-update behavior, move preview affected paths/redirects/blocked reasons, move cycle/path traversal rejection, delete machine-readable reasons, non-empty Directory deletion block, and Anonymous Visitor/Reader denial.

Out of Stage 2:

- Autosave, Content Version history, Draft Preview, Draft/Published Asset split, independent Published Content snapshots.
- Complete mobile Author workflow.
- Public homepage / Recent / public reading / comments / Likes / Glass Ricepaper redesign except regression repair and required Author entry.

Stage boundary precedence for execution:

- For Stage 2 implementation, `docs/plans/SECOND_DEVELOPMENT.md` Section 4 is the controlling scope.
- Any broader Author Workspace material in active specs that describes Draft Preview, autosave, Content Version history, Draft/Published Asset split, independent Published Content snapshots, or no-manual-save behavior is Stage 3 unless repeated in Section 4.
- Stage 2 File editing remains manual-save and uses the existing asset model. It exposes only `草稿` / `已发布`; Stage 3 will add `有未发布修改`, `发布更新`, autosave, version history, Draft Preview, and Draft/Published Assets.

## 2. Execution model

Use **OMX Team mode** with one coordinator and multiple role-specific lanes. The Team executes; the leader integrates; acceptance/security validate only integrated SHAs. Do not let feature developers self-approve.

Recommended launch shape after user approval:

```bash
# activate the approved Stage 2 DAG template first
export STAGE=2
rm -f .omx/plans/team-dag-second-development-active-*.json
cp .omx/plans/stages/stage-2-team-dag.json .omx/plans/team-dag-second-development-active.json
cmp .omx/plans/stages/stage-2-team-dag.json .omx/plans/team-dag-second-development-active.json

# exact approved launch hint; do not substitute free-form wording
omx team 5 "Execute approved xLab Blog second-development Stage 2 DAG"
```

Hard launch guard:

- The exact approved launch hint is required because OMX only imports the repo DAG when the invocation matches the approved PRD/test-spec launch text.
- Immediately after startup, the Coordinator must verify `.omx/state/team/<team>/decomposition-report.json` has `decomposition_source` indicating approved DAG/sidecar import, not `legacy_text`.
- If startup produces `legacy_text`, wrong worker mapping, all tasks assigned to one worker, or a task set unrelated to the Stage 2 DAG, stop before implementation, shutdown that Team, record the abort, and relaunch with the exact approved hint.

Why five persistent seats plus on-demand repair:

- The current Stage DAG is a strict five-seat bootstrap (`coordinator`, `backend`, `frontend`, `acceptance`, `security`), so Stage 2 must launch with five persistent workers.
- Repair remains an explicit **on-demand agent role/task lane**: when a gateway fails, the Coordinator creates a targeted repair task and routes it to the relevant development worker or spawns/assigns a temporary `debugger` / `code-simplifier` repair agent if the runtime supports it.
- This preserves the user-required repair responsibility without fighting the active Team DAG or causing seat-routing ambiguity.

## 3. Agent roster, responsibilities, intake, and deliverables

| Seat | OMX role / agent_type | Posture | May use Matt/skills | Primary intake | Primary deliverables | Must not do |
|---|---|---|---|---|---|---|
| 1. Coordinator / Gateway | `planner` or `writer` | Real-time scheduler, integration ledger, progress owner | `oh-my-codex:plan`, `to-issues`, `markdown-mermaid-writing` if needed | Approved plan, `PROGRESS.md`, Team status/events | Stage DAG/tasks, gateway decisions, `PROGRESS.md`, `docs/verification/stage-2-team-log.md`, integration ledger, stop/go notices | Feature code, self-approval |
| 2. Backend Developer | `executor` or `team-executor` | OpenAPI + Go API/repository/service implementation | `tdd`, `diagnose` for backend bugs | `docs/api/openapi.yaml`, `api/internal/tree/**`, router, tests | OpenAPI update, protected Author tree/detail/create/reorder/move/delete APIs, backend tests, migration/fixture helpers if needed | Frontend UI, acceptance signoff |
| 3. Frontend Developer | `designer` or `team-executor` | React implementation | `design-taste-frontend`, `tdd`, `image-to-code` only if visual reference appears | `web/src/pages/AdminPage.tsx`, `web/src/lib/api.ts`, `web/src/lib/types.ts`, styles/tests | Chinese Author Workspace shell, Content Tree, Directory/File workspaces, public Author entry, frontend tests | Backend contracts without OpenAPI first |
| 4. Acceptance Agent | `test-engineer` or `verifier` | Black-box acceptance and fixture verification | `tdd` for test authoring, `agent-browser`/browser verification when needed | Integrated SHA, fixture spec, acceptance criteria | `docs/verification/stage-2-acceptance.md`, desktop browser evidence, mobile no-regression sanity, API smoke, screenshots/logs | Feature implementation |
| 5. Security Agent | `code-reviewer` | Threat review and abuse tests | `code-review`, `diagnose` for security repros | Integrated SHA, protected API/public entry changes | `docs/verification/stage-2-security.md`, Draft leakage tests, auth denial tests, destructive bypass review | Feature implementation |
| 6. Repair / Stabilization Agent (on demand) | `debugger` or `code-simplifier` | Fixes issues found by gateways without derailing feature lanes | `diagnose`, `ai-slop-cleaner`, `code-simplifier` | Failed gateway reports, minimal repros | Small targeted repair commits, regression tests, cleanup patches | Broad redesign, unassigned feature expansion |
| 7. Independent Architect Review (after integration) | `architect` | Architecture review | `improve-codebase-architecture`, `code-review` | Integrated candidate and evidence | Architect `CLEAR` / required changes in `docs/verification/stage-2-code-review.md` | Authored feature code |
| 8. Independent Code Review (after architecture) | `code-reviewer` | Code-quality review | `code-review` | Integrated candidate and architect result | Code Review `APPROVE` / required changes in `docs/verification/stage-2-code-review.md` | Authored feature code |

Notes:

- The named skills are optional aids for each lane, not substitutes for project specs.
- Coordinator is the only lane that edits `PROGRESS.md` and the stage team log while Team is active.
- Acceptance/security must reset to integrated SHAs before testing.
- Repair is not a persistent sixth seat in the default launch; it is an explicit on-demand gateway-failure lane. Repair only acts on coordinator-created failure reports with a minimal repro, a bounded write scope, and a required regression test/evidence update.

## 4. Gateway model

A **Gateway** is a required stop/go checkpoint. No downstream packet may claim done until its gateway evidence is recorded.

### Gateway 0 — Launch readiness

Owner: Coordinator.

Entry conditions:

- User explicitly approves Stage 2 execution.
- Working tree is clean or intentional docs-only changes are committed.
- `PROGRESS.md`, `SECOND_DEVELOPMENT.md`, `CONTEXT.md`, relevant specs, and `openapi.yaml` have been read.
- Local services/database state is known.

Evidence:

- `docs/verification/stage-2-team-log.md` created with Team name, exact approved launch command, active DAG `cmp` output, `decomposition-report.json` path, `decomposition_source` not `legacy_text`, worker count=5, node/task mapping, worker/seat mapping, event cursor, and task table.
- Gateway 0 explicitly records that no `5:executor` or other role override was used.
- `PROGRESS.md` updated with Team name, current commit, and next gateway.

Exit:

- Team tasks created and assigned.
- Gateway 1 tasks unblocked.

### Gateway 1 — Data safety and fixture

Owner: Coordinator + Acceptance.

Required:

- Back up local database/uploads before fixture cleanup or schema changes.
- If cleanup, reset, or schema work is performed, prove disposable restore from the backup; if not applicable, record explicit `Restore proof: N/A` with reason.
- Create or refresh `/stage-2-acceptance` fixture.
- Record fixture IDs/paths and cleanup policy.
- Prove no accidental deletion of preserved baseline content.

Evidence:

- `docs/verification/stage-2-backup-and-fixture.md` or a section in `stage-2-acceptance.md`.
- `PROGRESS.md` checkpoint.

Exit:

- Backend contract work can proceed with known fixture assumptions.

### Gateway 2 — OpenAPI and backend Red contract

Owner: Backend, reviewed by Coordinator.

Required:

- Update `docs/api/openapi.yaml` first for protected Author Workspace APIs.
- Add failing backend tests for protected complete tree, Chinese-preserving generated URL Path, reserved root rejection, create conflict suffix, explicit URL Path conflict without silent rewrite, reorder transaction/lost-update behavior, move cycle/path traversal constraints, delete constraints, and auth denial.

Evidence:

- Test output showing intended Red failures or explicit baseline assertions.
- OpenAPI diff referenced in team log.

Exit:

- Backend implementation packets can proceed.

### Gateway 3 — Backend Green and service boundary

Owner: Backend; Repair may assist if Red remains.

Required:

- Implement repository/service/handler changes.
- SQL remains in repositories.
- No duplicated business logic for future MCP.
- Backend tests pass:

```bash
cd api
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./...
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go vet ./...
test -z "$(gofmt -l .)"
```

Evidence:

- Backend source SHA and output summary in stage team log.
- API smoke plan for integrated testing.

Exit:

- Frontend can bind to stable Author Workspace contracts.

### Gateway 4 — Frontend Red/UI contract

Owner: Frontend + Acceptance.

Required:

- Add or update frontend tests that fail against the old Admin page behavior:
  - no `ADMIN / Tree Manager`, `Node id`, Parent ID, Sort order, or `slug` primary UI;
  - Chinese Author Workspace labels;
  - create success refresh/select/open behavior mocked at API layer;
  - public Author edit/manage entry visibility by role;
  - Stage 2 does not show `有未发布修改` or `发布更新`;
  - return controls in subflows.

Evidence:

- Frontend Red output or explicit old-behavior reproduction notes.

Exit:

- Frontend implementation can proceed.

### Gateway 5 — Frontend Green and UX structure

Owner: Frontend; Repair may assist.

Required:

- Implement Chinese Author Workspace shell/tree/directory/file/settings/public-entry flows.
- Keep public reading surfaces stable except required Author entry.
- Required frontend gates pass:

```bash
cd web
node --test tests/render-safety.test.mjs
npm run lint
npm run build
```

Evidence:

- Frontend source SHA and test output in team log.
- Screenshots or browser notes if visually relevant.

Exit:

- Integrated acceptance/security can begin.

### Gateway 6 — Integrated desktop acceptance

Owner: Acceptance.

Required integrated SHA path:

1. Author login → Chinese Author Workspace.
2. Create Directory/File using minimal forms.
3. Tree refreshes, expands parent, selects/opens new node, shows Chinese toast/path.
4. Edit File, manual save, publish.
5. Public File opens; `编辑文件` returns to workspace with File selected.
6. Public Directory shows Author-only `管理此目录`, and it returns to workspace with that Directory selected.
7. Anonymous Visitor and Reader do not see `编辑文件` or `管理此目录`.
8. `撤回发布` hides public File after confirmation.
9. Settings move/URL Path/delete-constrained scenarios show clear Chinese prompts, including non-empty Directory deletion blocked even for draft-only children.
10. Same-parent drag reorder persists and never reparents.
11. Mobile no-regression sanity: no major layout break, orientation/exit present.
12. Public homepage/Recent/public reading/comments/Likes not redesigned.

Evidence:

- `docs/verification/stage-2-acceptance.md`.
- Browser logs/screenshots under `docs/verification/stage-2-browser-<date>/`.
- Native API smoke logs.

Exit:

- Security gateway can finalize; repair tasks are created for failures.

### Gateway 7 — Security and abuse review

Owner: Security.

Required:

- Anonymous Visitor/Reader denied protected APIs.
- Draft Files and draft-only branches do not leak via public tree/search/recent/assets/public Author entry logic.
- Invalid parent, cycles, path traversal, reorder lost update, redirect loop/chain, destructive bypass attempts are rejected.
- iframe sandbox and full-text fallback are preserved.

Evidence:

- `docs/verification/stage-2-security.md` with PASS/FAIL and exact repros.

Exit:

- Independent architecture/code review can begin if PASS; otherwise repair tasks.

### Gateway 8 — Independent review and closeout

Owner: Architect + Code Reviewer + Coordinator.

Required:

- Architect `CLEAR`.
- Code reviewer `APPROVE`.
- Full backend/frontend gates rerun on final integrated SHA.
- Stage 2 team log closed with pending=0, in_progress=0, failed=0.
- Rollback instructions and user验收 instructions recorded.

Evidence:

- `docs/verification/stage-2-code-review.md`.
- Updated `PROGRESS.md`.

Exit:

- Ask user to验收 Stage 2; do not start Stage 3 before explicit user acceptance.

## 5. OMX Team task graph

Coordinator should create/read concrete tasks from the Stage 2 packet DAG after Team launch. The tracked product plan and ignored runtime packet graph must stay aligned with this table. Runtime copy: `.omx/plans/stages/stage-2-packet-dag.json`.

| ID | Owner seat | Depends on | Task | Receives | Delivers |
|---|---|---|---|---|---|
| s2-00-launch-control | Coordinator | none | Gateway 0 launch control and decomposition audit | exact approved launch hint, active DAG | team log, `decomposition_source` audit, five-seat mapping, PROGRESS checkpoint |
| s2-01-data-fixture | Acceptance + Coordinator | s2-00-launch-control | Gateway 1 backup/restore/fixture | DB/uploads, fixture requirements | backup evidence, disposable restore proof or N/A, `/stage-2-acceptance` IDs |
| s2-02-backend-red-openapi | Backend | s2-01-data-fixture | Gateway 2 OpenAPI and backend Red contracts | OpenAPI, tree/admin code | OpenAPI diff, failing/targeted backend tests |
| s2-03-backend-tree-create | Backend | s2-02-backend-red-openapi | Protected Author tree/detail/minimal create APIs | Red tests | implementation + passing backend tests |
| s2-04-backend-reorder-move-delete | Backend | s2-03-backend-tree-create | Same-parent reorder, move preview/commit, delete constraints | tree services/repos | implementation + tests |
| s2-05-frontend-red-contracts | Frontend + Acceptance | s2-02-backend-red-openapi | Gateway 4 frontend Red/UI contracts | old AdminPage, API schemas | failing/current-behavior tests |
| s2-06-frontend-shell-tree | Frontend | s2-03-backend-tree-create, s2-05-frontend-red-contracts | Chinese Author Workspace shell + protected Content Tree | API contracts, UI spec | shell/tree, no raw IDs/slug primary UI, tests |
| s2-07-frontend-directory-create | Frontend | s2-06-frontend-shell-tree | Directory overview + minimal create flow | Author tree/create APIs | overview/create/toast/select/open behavior |
| s2-08-frontend-file-settings-assets | Frontend | s2-06-frontend-shell-tree, s2-04-backend-reorder-move-delete | File workspace + settings + assets + reorder UI | detail/content/publish/assets/reorder APIs | manual save, publish, secondary unpublish, settings, Directory Picker |
| s2-09-frontend-public-author-entry | Frontend | s2-08-frontend-file-settings-assets | Author-only public manage/edit entry | public resolver/file/directory pages | `管理此目录` / `编辑文件`, role-safe visibility, target selection routing |
| s2-11-acceptance | Acceptance | s2-04-backend-reorder-move-delete, s2-09-frontend-public-author-entry | Gateway 6 integrated desktop acceptance and mobile sanity | integrated SHA, fixture | acceptance evidence PASS/FAIL |
| s2-12-security | Security | s2-04-backend-reorder-move-delete, s2-09-frontend-public-author-entry | Gateway 7 integrated security and abuse review | integrated SHA | security evidence PASS/FAIL |
| s2-13-architect-review | Architect | s2-11-acceptance, s2-12-security | Independent architecture review gate | final candidate + evidence | CLEAR or required changes |
| s2-14-code-review | Code Reviewer | s2-13-architect-review | Independent code review gate | final candidate + architect result | APPROVE or required changes |
| s2-15-closeout | Coordinator | s2-14-code-review | Gateway 8 closeout and user acceptance handoff | all evidence | PROGRESS, team log, rollback/user验收 instructions |

On-demand repair routing:

- There is no default ready repair packet in the packet DAG. Repair is created only after a failed gateway or review finding.
- Backend failures → Coordinator creates a targeted repair task and assigns it to Backend or an on-demand `debugger`/`code-simplifier` repair agent with Backend consultation; create a new backend subtask if broad.
- Frontend failures → Coordinator creates a targeted repair task and assigns it to Frontend or an on-demand `debugger`/`code-simplifier` repair agent with Frontend consultation; create a new frontend subtask if broad.
- Acceptance/security FAIL → Coordinator creates explicit fix task(s), then reruns the relevant gateway.
- Review rejection → Coordinator creates targeted refactor/fix tasks; no self-approval.

## 6. Concrete file boundaries

Backend primary files:

- `docs/api/openapi.yaml`
- `api/internal/http/router.go`
- `api/internal/http/handlers/**`
- `api/internal/tree/admin_repository.go`
- `api/internal/tree/admin_service.go`
- `api/internal/tree/lifecycle_repository.go`
- `api/internal/tree/lifecycle_service_test.go`
- `api/internal/tree/*_test.go`
- `api/internal/search/**` only if search-index behavior is touched

Frontend primary files:

- `web/src/pages/AdminPage.tsx` — likely replaced/split rather than patched in place.
- `web/src/lib/api.ts`
- `web/src/lib/types.ts`
- `web/src/pages/ContentResolverPage.tsx`, `RootPage.tsx`, `DirectoryPage.tsx`, `components/FilePage.tsx` for public Author entry.
- `web/src/components/**` new Author Workspace components.
- `web/src/styles/glass.css` or new scoped styles for Author Workspace.
- `web/tests/**` for render-safety/API behavior tests.

Docs/evidence files:

- `PROGRESS.md`
- `docs/verification/stage-2-team-log.md`
- `docs/verification/stage-2-acceptance.md`
- `docs/verification/stage-2-security.md`
- `docs/verification/stage-2-code-review.md`
- `docs/verification/stage-2-browser-<date>/**`

Do not edit unrelated public UI unless required by Author entry or regression repair.

Public Author entry data-flow constraint:

- `管理此目录` / `编辑文件` visibility must be based on confirmed Author role, not merely token presence. If public components currently receive only token state, route `currentUser.role === "admin"` or an equivalent role-safe prop through the public page/component boundary.

## 7. Acceptance criteria

A Stage 2 candidate is acceptable only when all are true:

1. Author can complete the primary desktop path in `SECOND_DEVELOPMENT.md:181-193`.
2. No old primary UI labels: `ADMIN / Tree Manager`, `Node id`, Parent ID, Sort order, or `slug` in Author primary UI.
3. Newly created Directory/File appears immediately without manual refresh.
4. Generated File can be selected, edited, manually saved, published, and unpublish reached; Stage 2 does not expose `有未发布修改` / `发布更新`.
5. Public File `编辑文件` and public Directory `管理此目录` are visible only to Author and return to the workspace with the target selected.
6. Protected Author tree includes Draft Files that public tree/search/recent do not expose.
7. Same-parent reorder persists and cannot reparent.
8. Destructive/blocked operations have clear Chinese prompts, including every non-empty Directory deletion blocked.
9. Mobile width has no major layout break and a clear exit/orientation path.
10. Backend, frontend, API smoke, acceptance, security, architect, and code review gates all pass.


## 7.1 Stage 2 Chinese UI vocabulary

Use these Author-facing labels unless a route/API/code identifier requires English internally:

| Concept | UI label | Notes |
|---|---|---|
| Author Workspace | 作者工作区 | Do not show `Admin` / `Tree Manager` in product UI. |
| Content Tree | 内容树 | Left protected tree. |
| Directory | 目录 | Buttons: `新建目录`, public Author action `管理此目录`. |
| File | 文件 | Buttons: `新建文件`, public Author action `编辑文件`. |
| URL Path | 访问路径 | Never show `slug`. |
| Markdown | Markdown | Format option may remain technical label. |
| HTML Document | HTML 文档 | Format option. |
| Content | 内容 | File tab. |
| Assets | 资源 | File tab. |
| Settings | 设置 | File/Directory settings. |
| Publish / Unpublish | 发布 / 撤回发布 | `撤回发布` is secondary/danger. |
| View public page | 查看公开页面 | Use only when File is published. |
| Draft / Published | 草稿 / 已发布 | Stage 2 has only these two publication states. |

## 8. Verification commands

When shell environment is uncertain, wrap commands with `conda run -n blogenv bash -lc ...` as required by `AGENTS.md`.

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

Runtime/API/browser:

```bash
~/.local/share/xlab-blog/start-local.sh
curl -fsS http://127.0.0.1:8080/api/health
curl -fsS http://127.0.0.1:5173/ >/dev/null
```

Then run the Stage 2 fixture/API smoke and desktop/mobile-no-regression browser scripts; record exact commands and outputs in `docs/verification/`.

## 9. Risks and mitigations

| Risk | Mitigation |
|---|---|
| Protected tree leaks Draft content publicly | Separate Author Workspace APIs; security tests for public tree/search/recent/assets |
| UI becomes a patch over old AdminPage | Frontend Red tests reject old labels and raw IDs; split components by workspace area |
| Backend business logic duplicated for UI convenience | Service/repository boundary review at Gateway 3 and Architect gate |
| Drag sorting accidentally reparents | API schema only accepts same-parent order; browser test attempts invalid drag/reorder |
| Stage 3 semantics creep into Stage 2 | Gateway rejects autosave/version/Draft Preview implementation unless explicitly planned |
| Mobile scope expands silently | Acceptance defines mobile as no-regression sanity only |
| Team claims success without integrated proof | Coordinator ledger requires integrated SHA and gateway evidence before progress advances |

## 10. Goal / Team follow-up suggestions

Recommended execution after user approval:

- Use **Team + Ultragoal** if available: Ultragoal owns durable objective/checkpoints; Team returns gateway evidence.
- Use **Team alone** if user wants this session to manage execution manually.
- Do not use `$ralph` unless the user explicitly requests a sequential single-owner fallback.
- Do not use `$autoresearch-goal` or `$performance-goal`; this is implementation delivery, not research or measurable performance optimization.

Suggested command handoff:

```bash
export STAGE=2
rm -f .omx/plans/team-dag-second-development-active-*.json
cp .omx/plans/stages/stage-2-team-dag.json .omx/plans/team-dag-second-development-active.json
cmp .omx/plans/stages/stage-2-team-dag.json .omx/plans/team-dag-second-development-active.json
omx team 5 "Execute approved xLab Blog second-development Stage 2 DAG"
# then immediately verify status/list-tasks/decomposition-report before any implementation
```

Team must prove before shutdown:

- all S2 tasks terminal;
- all gateways PASS;
- all evidence files present;
- Stage 2 user验收 instructions ready;
- rollback instructions documented;
- `PROGRESS.md` updated.

## 11. Stop rules

Stop and ask/repair rather than broadening scope when:

- OpenAPI and implementation diverge;
- protected tree leaks Drafts;
- public UI is redesigned outside allowed Author entry/regression repair;
- Stage 2 starts implementing Stage 3 autosave/version semantics;
- database backup/fixture state is uncertain;
- acceptance/security/review fails;
- Team state cannot be reconciled.
