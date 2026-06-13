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

1. Protected Author Workspace Content Tree containing all Directories, Draft Files, Published Files, and Files with unpublished changes.
2. Minimal graphical creation flow: Directory = `名称`; File = `名称` + `格式`; generated URL Path with Chinese preserved.
3. Immediate post-create tree/navigation refresh, parent expansion, new-node selection/opening, and Chinese toast/path feedback.
4. Directory overview workspace with child cards and new Directory/File actions.
5. File workspace shell with `内容` / `资源` / `设置`, manual save, publish/unpublish controls, assets, and settings.
6. Settings with `基础信息`, `位置`, `危险操作`, graphical Directory Picker, explicit return buttons, and no Parent ID / Node ID / Sort order / `slug` in primary UI.
7. Author-only public Directory/File actions: `管理此目录` and `编辑文件`, returning to the workspace with the node selected.
8. Same-parent desktop drag sorting only; drag never reparents.
9. Dedicated Stage 2 fixture, desktop acceptance path, mobile no-regression sanity, security review, architect CLEAR, code-reviewer APPROVE.
10. Presentation-grade code quality: readable, extensible, layered, no duplicated SQL/business logic.

Out of Stage 2:

- Autosave, Content Version history, Draft Preview, Draft/Published Asset split, independent Published Content snapshots.
- Complete mobile Author workflow.
- Public homepage / Recent / public reading / comments / Likes / Glass Ricepaper redesign except regression repair and required Author entry.

## 2. Execution model

Use **OMX Team mode** with one coordinator and multiple role-specific lanes. The Team executes; the leader integrates; acceptance/security validate only integrated SHAs. Do not let feature developers self-approve.

Recommended launch shape after user approval:

```bash
# activate Stage 2 DAG template first if using existing DAG files
cp .omx/plans/stages/stage-2-team-dag.json .omx/plans/team-dag-second-development-active.json
cmp .omx/plans/stages/stage-2-team-dag.json .omx/plans/team-dag-second-development-active.json

omx team 6 "Execute Stage 2 Chinese Author Workspace plan from .omx/plans/stage-2-author-workspace-team-execution-plan.md"
```

Why six seats instead of the old five:

- Stage 2 now has enough UI/product risk to separate frontend feature development from design/UX implementation review, and enough expected fixes to dedicate a repair lane.
- If runtime only supports five reliable seats, merge `repair` into backend/frontend owners but keep repair tasks as explicit dependent tasks.

## 3. Agent roster, responsibilities, intake, and deliverables

| Seat | OMX role / agent_type | Posture | May use Matt/skills | Primary intake | Primary deliverables | Must not do |
|---|---|---|---|---|---|---|
| 1. Coordinator / Gateway | `planner` or `writer` | Real-time scheduler, integration ledger, progress owner | `oh-my-codex:plan`, `to-issues`, `markdown-mermaid-writing` if needed | Approved plan, `PROGRESS.md`, Team status/events | Stage DAG/tasks, gateway decisions, `PROGRESS.md`, `docs/verification/stage-2-team-log.md`, integration ledger, stop/go notices | Feature code, self-approval |
| 2. Backend Developer | `executor` or `team-executor` | OpenAPI + Go API/repository/service implementation | `tdd`, `diagnose` for backend bugs | `docs/api/openapi.yaml`, `api/internal/tree/**`, router, tests | OpenAPI update, protected Author tree/detail/create/reorder/move/delete APIs, backend tests, migration/fixture helpers if needed | Frontend UI, acceptance signoff |
| 3. Frontend Developer | `designer` or `team-executor` | React implementation | `design-taste-frontend`, `tdd`, `image-to-code` only if visual reference appears | `web/src/pages/AdminPage.tsx`, `web/src/lib/api.ts`, `web/src/lib/types.ts`, styles/tests | Chinese Author Workspace shell, Content Tree, Directory/File workspaces, public Author entry, frontend tests | Backend contracts without OpenAPI first |
| 4. Acceptance Agent | `test-engineer` or `verifier` | Black-box acceptance and fixture verification | `tdd` for test authoring, `agent-browser`/browser verification when needed | Integrated SHA, fixture spec, acceptance criteria | `docs/verification/stage-2-acceptance.md`, desktop browser evidence, mobile no-regression sanity, API smoke, screenshots/logs | Feature implementation |
| 5. Security Agent | `code-reviewer` | Threat review and abuse tests | `code-review`, `diagnose` for security repros | Integrated SHA, protected API/public entry changes | `docs/verification/stage-2-security.md`, Draft leakage tests, auth denial tests, destructive bypass review | Feature implementation |
| 6. Repair / Stabilization Agent | `debugger` or `code-simplifier` | Fixes issues found by gateways without derailing feature lanes | `diagnose`, `ai-slop-cleaner`, `code-simplifier` | Failed gateway reports, minimal repros | Small targeted repair commits, regression tests, cleanup patches | Broad redesign, unassigned feature expansion |
| 7. Independent Architect Review (after integration) | `architect` | Architecture review | `improve-codebase-architecture`, `code-review` | Integrated candidate and evidence | Architect `CLEAR` / required changes in `docs/verification/stage-2-code-review.md` | Authored feature code |
| 8. Independent Code Review (after architecture) | `code-reviewer` | Code-quality review | `code-review` | Integrated candidate and architect result | Code Review `APPROVE` / required changes in `docs/verification/stage-2-code-review.md` | Authored feature code |

Notes:

- The named skills are optional aids for each lane, not substitutes for project specs.
- Coordinator is the only lane that edits `PROGRESS.md` and the stage team log while Team is active.
- Acceptance/security must reset to integrated SHAs before testing.
- Repair lane only acts on gateway failures or explicit coordinator assignment.

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

- `docs/verification/stage-2-team-log.md` created with Team name, launch command, worker/seat mapping, event cursor, task table.
- `PROGRESS.md` updated with Team name, current commit, and next gateway.

Exit:

- Team tasks created and assigned.
- Gateway 1 tasks unblocked.

### Gateway 1 — Data safety and fixture

Owner: Coordinator + Acceptance.

Required:

- Back up local database/uploads before fixture cleanup or schema changes.
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
- Add failing backend tests for protected complete tree, generated URL Path, create conflict suffix, explicit URL Path conflict, reorder/move/delete constraints, and auth denial.

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
6. `撤回发布` hides public File after confirmation.
7. Settings move/URL Path/delete-constrained scenarios show clear Chinese prompts.
8. Same-parent drag reorder persists and never reparents.
9. Mobile no-regression sanity: no major layout break, orientation/exit present.
10. Public homepage/Recent/public reading/comments/Likes not redesigned.

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

Coordinator should create concrete tasks after Team launch. Suggested packet DAG:

| ID | Owner seat | Depends on | Task | Receives | Delivers |
|---|---|---|---|---|---|
| S2-00 | Coordinator | none | Launch/control task | plan, specs, current commit | team log, task ledger, status cadence, gateway enforcement |
| S2-01 | Coordinator + Acceptance | S2-00 | Data backup and fixture gate | DB/uploads, fixture requirements | backup/fixture evidence, `/stage-2-acceptance` IDs |
| S2-02 | Backend | S2-01 | OpenAPI + backend Red contracts | OpenAPI, tree/admin code | OpenAPI diff, failing/targeted tests |
| S2-03 | Backend | S2-02 | Protected Author tree/detail/create APIs | Red tests | implementation + passing backend tests |
| S2-04 | Backend | S2-03 | Reorder/move/delete constraints | tree services/repos | implementation + tests |
| S2-05 | Frontend + Acceptance | S2-02 | Frontend Red/UI contracts | old AdminPage, API schemas | failing/current-behavior tests |
| S2-06 | Frontend | S2-03,S2-05 | Author Workspace shell + Content Tree | API contracts, UI spec | Chinese shell/tree, tests |
| S2-07 | Frontend | S2-06 | Directory workspace + minimal create flow | Author tree/create APIs | overview/create/toast/select behavior |
| S2-08 | Frontend | S2-06 | File workspace + settings + assets | detail/content/publish/assets APIs | manual save, publish, secondary unpublish, settings |
| S2-09 | Frontend | S2-08 | Public Author manage/edit entry | public resolver/file/directory pages | `管理此目录` / `编辑文件`, node selection routing |
| S2-10 | Repair | any failed gateway | Targeted repair packet | failure report + repro | minimal fix + regression test |
| S2-11 | Acceptance | S2-04,S2-09 | Integrated desktop acceptance | integrated SHA, fixture | acceptance evidence PASS/FAIL |
| S2-12 | Security | S2-04,S2-09 | Security review | integrated SHA | security evidence PASS/FAIL |
| S2-13 | Architect | S2-11,S2-12 | Architecture review | final candidate + evidence | CLEAR or required changes |
| S2-14 | Code Reviewer | S2-13 | Code review | final candidate + architect result | APPROVE or required changes |
| S2-15 | Coordinator | S2-14 | Closeout | all evidence | PROGRESS, team log, user验收 instructions |

Repair routing:

- Backend failures → S2-10 assigned to Repair with Backend consultation, or new backend subtask if broad.
- Frontend failures → S2-10 assigned to Repair with Frontend consultation, or new frontend subtask if broad.
- Acceptance/security FAIL → coordinator creates explicit fix task(s), then reruns relevant gateway.
- Review rejection → coordinator creates targeted refactor/fix tasks; no self-approval.

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

## 7. Acceptance criteria

A Stage 2 candidate is acceptable only when all are true:

1. Author can complete the primary desktop path in `SECOND_DEVELOPMENT.md:181-193`.
2. No old primary UI labels: `ADMIN / Tree Manager`, `Node id`, Parent ID, Sort order, or `slug` in Author primary UI.
3. Newly created Directory/File appears immediately without manual refresh.
4. Generated File can be selected, edited, published, and unpublish reached.
5. Public File `编辑文件` returns to the workspace with the File selected.
6. Protected Author tree includes Draft Files that public tree/search/recent do not expose.
7. Same-parent reorder persists and cannot reparent.
8. Destructive/blocked operations have clear Chinese prompts.
9. Mobile width has no major layout break and a clear exit/orientation path.
10. Backend, frontend, API smoke, acceptance, security, architect, and code review gates all pass.

## 8. Verification commands

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
omx team 6 "Execute Stage 2 Chinese Author Workspace plan from .omx/plans/stage-2-author-workspace-team-execution-plan.md"
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
