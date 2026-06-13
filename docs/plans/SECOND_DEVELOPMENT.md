# xLab Blog Second Development Plan

Last revised: 2026-06-13 CST

This is the active development plan. Historical implementation detail is compacted in Git history and `docs/archive/INITIAL_BUILD_SUMMARY.md`; do not revive stale OMX task state or old detached-worker plans.

## 1. Product and scope locks

The project remains a single-Author full-stack Blog organized by a Unix-like **Content Tree**. Use the product language in `docs/specs/CONTEXT.md`: `Author`, `Reader`, `Anonymous Visitor`, `Author Workspace`, `Content Tree`, `Directory`, `File`, `URL Path`, `Content Version`, and `Published Content`. `Admin` describes privileges/routes only; never expose `slug` in product UI.

Locked stages:

1. **Reliability, navigation, and identity** — engineering complete; user acceptance found Author Workspace UX blockers that are folded into Stage 2.
2. **Chinese Author Workspace and protected Content Tree** — current active implementation target.
3. **Autosave, publication snapshots, Draft Preview, Draft/Published Assets, and Blog MCP Server** — final product stage plus server-local MCP.

Scope exclusions unless repairing regressions:

- no public homepage redesign;
- no Recent card redesign;
- no public Directory/File reading redesign;
- no comments/Likes redesign;
- no Glass Ricepaper redesign;
- no cross-Directory drag-and-drop reparenting;
- no Author Workspace tree search in Stage 2;
- no container/server deployment before native stages and user acceptance pass.

Engineering quality is part of the deliverable because the project will be presented and defended: code must be readable, extensible, and structurally clear; avoid UI-specific backend hacks, duplicated business logic, hidden fallbacks, swallowed errors, fake success, and untested alternate paths.

## 2. Execution rules

- Keep each stage runnable, reversible, independently testable, documented, and user-accepted before advancing.
- Update `docs/api/openapi.yaml` before shared API behavior changes.
- Keep SQL in repositories, not HTTP handlers.
- Preserve iframe `sandbox="allow-scripts"` without `allow-same-origin`.
- Preserve full-text search fallback when semantic indexing is unavailable.
- Back up the local database before cleanup, fixture reset, or schema migration.
- Do not commit credentials, local database files, uploads, caches, build output, or agent runtime state.
- Update `PROGRESS.md` at key milestones and before stopping.
- For each stage, record evidence under `docs/verification/` and state `Tested:` / `Not tested:`.
- For runtime/auth/tree/publication changes, run native PostgreSQL API smoke and browser acceptance; Stage 2 mobile is no-regression sanity only.

Required local environment is Conda `blogenv`:

- Node.js `22.22.3`
- npm `10.9.8`
- Go `1.26.4`
- PostgreSQL `17.10`
- pgvector `0.8.1`

Baseline gates:

```bash
cd api
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./...
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go vet ./...
test -z "$(gofmt -l .)"

cd web
node --test tests/render-safety.test.mjs
npm run lint
npm run build
```

Use additional targeted tests and browser/API smoke for the stage under work.

## 3. Stage 1 status — Reliability, navigation, and identity

Stage 1 engineering is complete and independently reviewed, but final user acceptance did not pass because the current Author-facing Admin UI is not acceptable. The following feedback is now Stage 2 scope:

1. after creating a Directory/File, navigation and Content Tree surfaces must update immediately;
2. all Author flows need explicit in-app return controls;
3. generated Files must be selectable/editable/unpublishable, and Authors browsing public Directory/File pages need a direct manage/edit path;
4. the Author-facing UI must become graphical, Chinese, clear, minimal, and presentation-quality.

Stage 1 fixes already landed include role-aware login defaults, Reader logout staying on public pages, Author logout from `/admin` to `/`, and status-specific auth errors. Evidence remains in `docs/verification/stage-1-*` and Git history.

## 4. Stage 2 — Chinese Author Workspace and protected Content Tree

Stage 2 replaces the current form-heavy Admin page with a desktop-first Chinese **Author Workspace**. The route may remain `/admin`, but product UI should not say `Admin / Tree Manager`.

### 4.1 Data and fixture gate

Before implementation or destructive fixture work:

1. back up the local database/uploads;
2. prove restore on a disposable target when doing schema/cleanup work;
3. create or refresh a dedicated acceptance root such as `/stage-2-acceptance`.

The fixture must cover nested Directories, Draft Files, Published Files, Chinese/mixed URL Paths, same-parent ordering, move constraints, deletion constraints, and Author-only public manage/edit entry.

### 4.2 Backend/API scope

Update OpenAPI first. Add or reshape protected Author Workspace APIs with clear repository/service/handler boundaries:

- complete protected Author Content Tree containing all Directories, Draft Files, and Published Files. Stage 2 does not model `有未发布修改`; that state requires Stage 3 Published Content snapshots;
- node detail for Directory/File workspace loading;
- context-aware create using the selected parent Directory;
- backend-authoritative Name → URL Path generation:
  - preserve Chinese characters;
  - normalize Latin text to lowercase hyphenated segments;
  - append numeric suffixes only for initial create conflicts;
  - never silently rewrite explicit URL Path edits;
- same-parent mixed Directory/File reorder with transaction safety;
- graphical Directory Picker support for cross-Directory moves, impact preview, cycle prevention, subtree path rewrite, and redirects for formerly public paths;
- deletion constraints with clear reasons for non-empty Directories and Published Files; Stage 2 chooses the conservative rule that every non-empty Directory is blocked, including draft-only subtrees;
- publication state read/update sufficient for the Stage 2 manual-save File workspace.

Do **not** implement Stage 3 Content Version history, Draft Preview, Draft/Published Asset split, or independent Published Content snapshots in Stage 2.

### 4.3 Frontend scope

Build a desktop-first Chinese Author Workspace. Mobile Stage 2 is no-regression sanity only: it must open without major layout breakage and provide orientation/exit, but full mobile create/edit/move/delete flows are deferred.

#### Workspace shell

- Desktop two-column layout: left protected Content Tree, right contextual workspace.
- Visual direction: lightweight professional writing/management tool inside Glass Ricepaper; quiet, sparse, readable, operation-first, and minimal.
- Use stronger card/status treatment only when it improves understanding or prevents mistakes: creation success, publication state, save/error state, and danger confirmations.
- Author Workspace and Author-facing flows are Chinese. Public pages are not broadly redesigned except required Author-workflow touchpoints.

#### Content Tree

- Expand/collapse tree; no tree search in Stage 2.
- Stage 2 uses one protected complete-tree load for the Author Workspace left tree. Directory workspace child cards may be derived from that tree or detail API; do not implement public-tree or draft-leaking shortcuts.
- Show all Directories, Draft Files, and Published Files. Do not display `有未发布修改` in Stage 2 because manual saves still update the single draft/live content model inherited from Stage 1; Stage 3 adds separate Published Content snapshots.
- Restore browser-local selection/expanded state when safe.
- Creation refreshes the tree, expands the parent Directory, selects the new node, and opens the right workspace.
- Author public Directory/File entry expands ancestors and selects the target node.
- Desktop drag only reorders siblings in the same Directory; drag never reparents.

#### Directory workspace

Selecting a Directory opens a Directory overview, not settings by default:

- Directory Name and URL Path;
- clear `新建 Directory` and `新建 File` actions;
- child cards for current Directory contents;
- Settings entry.

Create flow is minimal:

- Directory: `名称` only;
- File: `名称` + `格式` (`Markdown` / `HTML Document`);
- show read-only final URL Path preview;
- never expose Parent ID, Node ID, Sort order, or `slug` in the primary UI.

Creation success must show a lightweight Chinese toast, refresh tree/navigation state, expand/select/open the new node, and display the final path clearly.

#### File workspace

Selecting a File opens a Stage-3-compatible shell, but saving remains manual:

- header: File Name, status (`草稿` or `已发布` only in Stage 2), URL Path, public-view action, and one primary publication action;
- sections/tabs: `内容`, `资源`, `设置`;
- `内容`: body editor, keywords, manual save, clear Chinese success/error messages;
- `资源`: upload/view/delete using the existing asset model;
- `设置`: Name, URL Path, move, delete constraints, and danger zone.

Publication control:

- Draft → primary `发布`;
- Stage 2 does not expose `有未发布修改` / `发布更新`; changed saved content uses the existing single-content model until Stage 3 snapshots are added.
- current Published File → status `已发布` instead of redundant publish button;
- `撤回发布` is secondary/overflow/danger, not a sibling primary button.

#### Settings, return, and public Author entry

- Settings sections: `基础信息`, `位置`, `危险操作`.
- Danger actions are bottom, visually distinct, and require Chinese second confirmation.
- Block Published File deletion and every non-empty Directory deletion, including draft-only subtrees, with clear Chinese explanations. Stage 2 does not offer recursive subtree deletion; Authors must empty a Directory first.
- Cross-Directory moves use a graphical Directory Picker and path/impact preview; never require Parent ID.
- Every right-workspace subflow has explicit return buttons such as `返回当前目录`, `返回文件内容`, or `返回设置`.
- Lightweight breadcrumbs/path indicators aid orientation but are not the only return path.
- Author-only public entries:
  - Directory: `管理此目录`;
  - File: `编辑文件`;
  - both enter the workspace with the target node selected.

### 4.4 Stage 2 acceptance

Primary desktop user acceptance path:

1. log in as Author and enter the Chinese Author Workspace;
2. create a Directory and File with minimal Chinese forms;
3. verify the Content Tree refreshes immediately, expands the parent, selects/opens the new node, and shows clear Chinese toast/path feedback;
4. edit File content and keywords, manually save, and see clear Chinese feedback;
5. publish the File and verify public access;
6. from the public File, click `编辑文件` and return to the workspace with the File selected;
7. reach `撤回发布` as a secondary/danger action and verify it hides the public File after confirmation;
8. verify Settings move/URL Path/delete-constrained cases show clear Chinese prompts and no Parent ID, Node ID, or `slug`;
9. verify same-parent desktop drag sorting persists and never reparents;
10. verify public homepage, Recent cards, public Directory/File reading, comments/Likes, and Glass Ricepaper are not redesigned except required Author entry/regression repair;
11. verify mobile no-regression sanity only.

Automated/security gates must also cover Draft leakage, Reader/Anonymous Visitor denial, URL Path conflicts, deletion protection, full-text search fallback, iframe sandbox preservation, redirect/cycle/path traversal attacks, and backup/fixture evidence.

## 5. Stage 3 — Autosave, publication model, Draft Preview, Assets, and MCP

Stage 3 extends the accepted Author Workspace with the final publication model and MCP. It must remain migration-safe and presentation-quality.

### 5.1 Autosave, Content Versions, and Published Content

Add migration and APIs for:

- Current Content Version;
- Previous Content Version;
- independent Published Content;
- monotonic revision for optimistic concurrency;
- Draft/Published Asset state.

Rules:

- existing Published File → Current + Published Content;
- existing Draft File → Current only;
- no-op save produces no Previous;
- successful changed save rotates old Current into Previous;
- restore swaps Current/Previous;
- publish snapshots Current into Published Content;
- unpublish hides public visibility but retains Published Content;
- search indexes Published Content only;
- semantic failure remains non-blocking.

### 5.2 Stage 3 frontend scope

- Autosave 15 seconds after input stops and immediately on blur, node change, Publish, Logout, or leaving Author Workspace.
- Fixed status states: Editing, Saving, Saved, Save failed, Conflict, Unpublished changes.
- Failed required save blocks unsafe navigation/logout/publication and preserves text.
- Conflict actions: Reload latest / Copy my changes; no auto-merge.
- Current/Previous timestamps, compare, and reversible restore.
- Publish / Publish changes / Published, with Unpublish in overflow.
- Responsive editor/preview split on desktop; mobile edit/preview can be completed in Stage 3 if accepted.
- Author-only Draft Preview at `/admin/preview/{file_id}` using saved Current content and Draft Assets.
- Draft/Published Asset presentation and publish summary.

### 5.3 Blog MCP Server final-stage requirement

Build a separate server-local stdio MCP Server package/process for trusted AI agents running on the server. Public HTTP/SSE MCP transport is out of initial scope and may be added later behind explicit auth and network-binding controls.

MCP posture:

- high-trust full Author permissions;
- explicit enablement configuration;
- operation audit logs;
- automatic backup/export before destructive batches where practical;
- emergency disable/kill switch;
- reuse backend service/API-client capabilities;
- do not duplicate business logic or direct SQL in ad hoc MCP handlers.

MCP tools required by Stage 3 closeout:

- read: `list_content_tree`, `get_file`, `search_files`;
- content: `create_directory`, `create_file`, `update_file_content`, `update_file_settings`;
- publish: `publish_file`, `unpublish_file`;
- tree: `move_node`, `reorder_children`, `delete_node`;
- assets: `upload_asset`, `delete_asset`, `list_assets`;
- maintenance: `rebuild_search_index`, `export_backup`.

Implement in tested slices: read/search → content → publish → tree → assets → maintenance.

### 5.4 Stage 3 acceptance and security

Acceptance must prove migration/restore, autosave timing, save failure blocking, conflict behavior, Current/Previous restore, Published Content stability, Draft Preview role matrix, Draft/Published Asset isolation/promotion, public search using Published Content, sandbox preservation, and MCP tool operation/audit/disable behavior.

Security must prove Anonymous Visitor/Reader denial for Draft Preview, Draft Assets, protected APIs, and MCP; reject filename/path manipulation, stale revision overwrite, redirect loops, and destructive bypasses.

## 6. Team execution and evidence model

Use a fresh five-seat Team per stage only when implementation begins, plus on-demand repair agents/tasks when a gateway fails:

1. coordinator — progress, task ledger, integration evidence;
2. backend — OpenAPI, Go handlers/services/repositories, migrations, backend tests;
3. frontend — React state/pages/components/styles and frontend tests;
4. acceptance — black-box/API/database/browser verification;
5. security — threat review and abuse tests, no feature code.

Repair is not a default sixth persistent seat unless the active runtime/DAG is explicitly changed; failed gateways get coordinator-created repair tasks routed to the relevant lane or an on-demand debugger/code-simplifier. Leader integrates worker commits; acceptance/security test only integrated SHAs. Coordinator owns `PROGRESS.md` and the stage team log while Team is active. Evidence ownership:

- `docs/verification/stage-<n>-team-log.md` — coordinator;
- `docs/verification/stage-<n>-acceptance.md` — acceptance;
- `docs/verification/stage-<n>-security.md` — security;
- `docs/verification/stage-<n>-code-review.md` — independent review.

Stage closeout requires:

- terminal task counts: pending=0, in_progress=0, failed=0;
- backend/frontend required gates PASS;
- acceptance PASS;
- security PASS;
- architect CLEAR;
- code-reviewer APPROVE;
- rollback instructions;
- explicit user acceptance before next stage.

## 7. Stop conditions

Pause instead of broadening scope when:

- migration is not demonstrably lossless/reversible/re-runnable;
- public behavior regresses outside the allowed repair surface;
- acceptance/security/architect/code-review verdict is not PASS/PASS/CLEAR/APPROVE;
- a stage cannot end runnable;
- Team/task/worktree/integration state cannot be reconciled;
- exact versions/contracts cannot be satisfied;
- user acceptance finds a blocker.

Never weaken/delete tests, suppress errors, silently default, or change scope merely to pass a gate.
