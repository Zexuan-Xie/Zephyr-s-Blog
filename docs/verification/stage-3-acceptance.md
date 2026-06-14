# Stage 3 Acceptance Plan

Status: **Gateway 0/1 acceptance plan draft**

Owner: `worker-4` acceptance lane  
Scope: black-box database, API, browser, and MCP verification for Stage 3 after coordinator integration.  
Rule: acceptance tests run only against integrated leader branch SHAs, never isolated worker branches.

## Purpose

Stage 3 must prove that autosave, content history, Published Content snapshots, Draft Preview, Draft/Published Assets, and the server-local Blog MCP Server work together without weakening the Stage 2 Author Workspace. This document defines the acceptance matrix, fixture data, evidence files, and smoke-script plan that will be used once backend/frontend/MCP slices are integrated.

## Test baseline and SHA discipline

For every execution pass, record:

- integrated SHA under test: `git rev-parse HEAD`;
- migration version list applied to the local PostgreSQL database;
- backup directory used before any destructive or migration test;
- API and web process start commands and health checks;
- browser viewport/device used for each smoke;
- exact PASS/FAIL summary with command output or artifact paths.

Acceptance must stop and report to coordinator if:

- the database migration cannot be restored, re-run on a disposable database, or rolled forward from the Stage 2 shape;
- a public page/search/recent result changes from a draft autosave before Publish;
- Reader/Anonymous can reach Draft Preview, draft assets, protected Author APIs, or MCP operations;
- any failed required save allows unsafe navigation, logout, or publication;
- MCP destructive operations lack audit evidence or the required backup/export evidence.

## Fixture data needs

Create or reuse fixtures only after a fresh backup. Prefer a dedicated root path so cleanup and comparisons are deterministic.

| Fixture | Required state | Purpose |
|---|---|---|
| `/stage-3-acceptance` root directory | directory, public children under it only as specified | isolates all Stage 3 checks from user content |
| `published-stability` file | Published Content snapshot body: `Published v1`; Current body initially matches; one published asset | proves autosave edits do not alter public file/search/recent/assets until Publish |
| `draft-preview-only` file | draft Current body with one draft asset; no Published Content | proves Author-only preview and public denial |
| `conflict-target` file | Current revision known to test harness | proves optimistic concurrency and two-tab conflict behavior |
| `previous-rotation` file | Current body `current-A`; later save to `current-B` | proves no-op save does not rotate Previous, changed save does rotate, restore swaps Current/Previous |
| `asset-promotion` file | Published body references `published.png`; draft upload references `draft.png` | proves Draft/Published asset isolation and publish promotion |
| Reader account | role `reader` | role-matrix checks for Draft Preview/protected APIs |
| Author account | role `admin` | all Author Workspace/API/MCP checks |

Record fixture IDs, paths, initial revisions, asset IDs, and content hashes in the final evidence section. If the local database is intentionally empty at Stage 3 start, fixture creation must also prove the empty-root Author Workspace still works.

## Black-box acceptance matrix

### A. Migration, backup, and restore

| ID | Scenario | Method | Expected result | Evidence |
|---|---|---|---|---|
| A1 | Pre-migration backup | `pg_dump` custom-format dump plus uploads archive | dump, uploads archive, checksums exist before schema changes | `stage-3-backup-*.txt`, `SHA256SUMS.txt` |
| A2 | Disposable restore before fixture | restore dump into disposable DB and count core tables | restore succeeds; counts readable; disposable DB dropped | restore transcript |
| A3 | Stage 2 -> Stage 3 migration | run migrations on a database with Stage 2-style `nodes`, `file_contents`, `file_assets` | existing published files copied to Published Content; drafts remain Current-only | SQL count/hash report |
| A4 | Migration re-run safety | run migration tool a second time on disposable DB | no duplicate Published Content/assets; no destructive drift | migration transcript |
| A5 | Rollback/restore rehearsal | restore pre-migration dump after test run | application can return to pre-test data state | restore command and health check |

### B. API content-version contract

| ID | Scenario | Method | Expected result | Evidence |
|---|---|---|---|---|
| B1 | Current read includes revision | Author API fetches file detail/current content | response includes monotonic `revision`, timestamps, Current metadata | API transcript |
| B2 | No-op save | save identical body with matching revision | HTTP 200; revision unchanged or explicitly no-op; Previous slot unchanged | DB/API before-after diff |
| B3 | Changed save rotation | save changed body with matching revision | HTTP 200; revision increments; old Current becomes Previous | API + DB row hashes |
| B4 | Stale revision conflict | save with older expected revision | HTTP 409; machine-readable conflict/lost-update error; Current unchanged | API transcript |
| B5 | Restore Previous | call restore endpoint/action | Current/Previous swap; revision increments; public Published Content unchanged until Publish | API + public resolver diff |
| B6 | Publish snapshot | publish Current | Published Content body/search fields match Current; published timestamp updated | API + DB hash report |
| B7 | Unpublish | unpublish published file | public resolver/search/recent hide file; Published Content metadata retained for compare/republish | API transcript + DB check |
| B8 | Error mapping | invalid format, missing file, non-file target | 400/404/409 as documented; no 500 for client errors | API transcript |

### C. Published Content public stability

| ID | Scenario | Method | Expected result | Evidence |
|---|---|---|---|---|
| C1 | Public page stable after draft autosave | edit `published-stability` Current to `Draft v2`; query public path | public page still shows `Published v1` | browser/API screenshots/transcripts |
| C2 | Search source is Published Content | search for `Draft v2` before Publish and `Published v1` before Publish | draft term absent; published term present | API transcript |
| C3 | Recent source is Published Content | inspect Recent/home cards after draft edit | card metadata/content stays at published snapshot | browser screenshot/API transcript |
| C4 | Publish updates public snapshot | publish `Draft v2` | public page/search/recent now show `Draft v2` | browser/API evidence |
| C5 | Comments/likes still attach safely | comment/like existing published file, then draft-edit | comments/likes remain on public published file; no draft leakage | API/browser evidence |
| C6 | iframe sandbox preserved | render published HTML document | iframe sandbox is `allow-scripts` and does not include `allow-same-origin` | DOM assertion screenshot/log |

### D. Browser autosave and Author Workspace

| ID | Scenario | Method | Expected result | Evidence |
|---|---|---|---|---|
| D1 | Debounced autosave | type in editor and wait 15 seconds after input stops | status transitions Editing -> Saving -> Saved; API save observed | browser log/screenshot |
| D2 | Blur save | edit then blur editor before 15 seconds | immediate save starts and succeeds | browser network log |
| D3 | Node-change required save | edit file then select another node | save runs before switch; switch blocked on failure | browser log |
| D4 | Publish required save | edit then click Publish/Publish changes | pending save completes before publish; failed save blocks publish | browser log/API transcript |
| D5 | Logout/leave required save | edit then logout or navigate away | save runs first; failed save blocks unsafe action and preserves text | browser log |
| D6 | Save failure state | force API failure/network failure | status shows Save failed; text remains in editor; unsafe actions blocked | browser screenshot/log |
| D7 | Conflict UI | simulate stale tab save | status Conflict; Reload latest and Copy my changes both work; no auto-merge | two-context browser evidence |
| D8 | Current/Previous UI | perform changed save then open compare panel | timestamps, compare, and restore controls visible and correct | screenshot/log |
| D9 | Publish controls | published, unpublished-changes, and draft states | labels are exactly Publish / Publish changes / Published; Unpublish is secondary/danger | screenshots |
| D10 | Regression Stage 2 workspace | create root directory/file, settings path edit, move preview, reorder, delete blocked non-empty | Stage 2 flows still pass with simple English Aeolian copy | browser/API evidence |

### E. Draft Preview role matrix

| ID | Actor | Method | Expected result | Evidence |
|---|---|---|---|---|
| E1 | Author | open `/admin/preview/{file_id}` for draft file | saved Current content and draft assets render | screenshot/network log |
| E2 | Author | preview published file with unpublished Current edits | preview shows Current draft, public path shows Published Content | side-by-side screenshot/API transcript |
| E3 | Reader | open Draft Preview URL | denied with 401/403 or redirected to login; no body/asset leakage | browser/API transcript |
| E4 | Anonymous | open Draft Preview URL | denied; no body/asset leakage | browser/API transcript |
| E5 | Reader/Anonymous | request preview API directly | denied; no draft JSON or asset URL in response | curl transcript |

### F. Draft/Published Assets

| ID | Scenario | Method | Expected result | Evidence |
|---|---|---|---|---|
| F1 | Draft upload is private | upload asset to published file after public snapshot exists | public asset endpoint denies draft asset until Publish | API transcript |
| F2 | Published asset remains stable | remove/replace draft asset while public content references previous published asset | public page and published asset URL still work | browser/API evidence |
| F3 | Publish promotes asset state | publish after draft asset changes | public page and asset endpoint expose promoted assets only | API/browser evidence |
| F4 | Draft-only delete | delete an unpublished draft asset | asset disappears from draft list and never becomes public | API transcript |
| F5 | Published snapshot delete semantics | remove an asset still used by Published Content before next Publish | existing public Published Content does not break before next Publish/Unpublish | browser/API evidence |
| F6 | Filename/path manipulation | upload `../`, backslash, duplicate, oversized, unsafe SVG | rejected with precise client error; storage path remains inside uploads root | API transcript + storage check |

### G. Blog MCP Server

| ID | Scenario | Method | Expected result | Evidence |
|---|---|---|---|---|
| G1 | Disabled by default | run MCP tool with enable flag absent/false | every tool refuses before doing work; audit records refusal if implemented | stdio transcript/audit log |
| G2 | Enabled local stdio only | run server with explicit enablement | stdio transport works locally; no public HTTP/SSE listener added for Stage 3 | process/netstat transcript |
| G3 | Read/search tools | `list_content_tree`, `get_file`, `search_files` | returns only authorized Author-view data through backend service/API-client boundary | MCP transcript |
| G4 | Content tools | `create_directory`, `create_file`, `update_file_content`, `update_file_settings` | changes appear in Author API; revision/conflict rules apply | MCP + API transcript |
| G5 | Publish tools | `publish_file`, `unpublish_file` | public snapshot changes only on publish; unpublish hides public visibility | MCP + public API transcript |
| G6 | Tree tools | `move_node`, `reorder_children`, `delete_node` | same constraints as Author API; destructive delete blocked when unsafe | MCP transcript |
| G7 | Asset tools | `upload_asset`, `delete_asset`, `list_assets` | draft/published asset rules are identical to Author API | MCP transcript |
| G8 | Maintenance tools | `rebuild_search_index`, `export_backup` | backup/export artifact recorded before destructive/batch operations where practical | MCP + filesystem evidence |
| G9 | Audit log | run representative successful and failed tools | JSONL contains timestamp, tool name, args summary, result/error, no secrets | audit excerpt |
| G10 | Kill switch | toggle emergency disable before tool call | tool refuses and does not mutate database/files | MCP transcript + DB diff |
| G11 | No direct SQL in handlers | static grep/review of MCP package | MCP handlers call backend service/API-client; no ad hoc SQL in handler package | grep/review transcript |

## Smoke-script plan

The acceptance lane should keep scripts black-box and disposable. If scripts are added, place them under a verification-focused path and make them parameterized by base URLs and database URL; do not embed secrets.

Recommended scripts/artifacts after implementation integration:

| Script/artifact | Purpose | Inputs | Output |
|---|---|---|---|
| `stage3_api_smoke` | fixture create, revision saves, publish/unpublish, public stability, role denial | API base, Author token, Reader token | `docs/verification/stage-3-browser-YYYYMMDD/stage3-api-smoke-<sha>.txt` |
| `stage3_db_snapshot.sql` or equivalent command block | migration count/hash checks for Current/Previous/Published/Assets | database URL | SQL transcript with row counts and content hashes |
| `stage3_browser_smoke` | autosave, conflict, preview, assets, regression Stage 2 flows | web base, Author/Reader credentials | screenshots and browser network logs |
| `stage3_mcp_smoke` | disabled/enabled stdio, CRUD/publish/tree/assets/export/audit | MCP command, enable flag, API/database config | stdio transcript, audit log path, backup/export path |
| `stage3_security_probe` | role denial, path traversal, unsafe SVG, stale revision overwrite | API base/tokens | PASS/FAIL transcript |

Until scripts are committed, the command blocks in this document are the source of truth for manual/agent-driven acceptance.

## Required command gates

Run from the integrated leader worktree after each acceptance candidate SHA:

```bash
cd api
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./...
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go vet ./...
test -z "$(gofmt -l .)"

cd ../web
node --test tests/*.test.mjs
npm run lint
npm run build
```

Expected final result: all PASS. If Stage 3 introduces new targeted tests or MCP package tests, add them to this list before closeout.

## Evidence layout

Use this layout for Stage 3 evidence:

```text
docs/verification/stage-3-acceptance.md
docs/verification/stage-3-security.md
docs/verification/stage-3-team-log.md
docs/verification/stage-3-browser-20260613/
  backend-go-test-<sha>.txt
  backend-go-vet-<sha>.txt
  backend-gofmt-<sha>.txt
  frontend-node-test-<sha>.txt
  frontend-lint-<sha>.txt
  frontend-build-<sha>.txt
  stage3-api-smoke-<sha>.txt
  stage3-db-migration-restore-<sha>.txt
  stage3-browser-autosave-<sha>.txt/png
  stage3-browser-conflict-<sha>.txt/png
  stage3-browser-preview-role-matrix-<sha>.txt/png
  stage3-browser-assets-<sha>.txt/png
  stage3-mcp-smoke-<sha>.txt
  stage3-mcp-audit-<sha>.jsonl
```

## Gateway 0/1 readiness checklist

- [ ] Coordinator records Stage 2 checkpoint SHA and backup plan.
- [ ] Backend OpenAPI defines revision, conflict errors, Current/Previous, Published Content, Draft Preview, assets, and any MCP-support endpoints.
- [ ] Backend red tests include revision conflict, rotation, restore, published snapshot stability, search over Published Content, Draft Preview denial, and asset isolation.
- [ ] Frontend contract tests include autosave status machine, required-save blocking, conflict actions, publish labels, version restore, preview, and asset states.
- [ ] Acceptance fixture IDs and expected public/draft terms are recorded before Gateway 2 migration work.
- [ ] Security lane has reviewed denial cases and destructive MCP expectations before implementation slices merge.

## Initial task-5 verification

This draft is documentation-only and does not implement feature code. Verification for this task should therefore prove that the repo still builds/tests at the current Stage 2 checkpoint and that the acceptance plan exists for downstream integrated SHAs.

## Gateway 1 contract acceptance review — 2026-06-13

Verdict: **PASS for Gateway 1 contract readiness, with downstream implementation watch items**.

Reviewed integrated leader-branch artifacts:

- `docs/api/openapi.yaml` contains Stage 3 endpoints and schemas for Current/Previous/Published Content, `expected_revision`, revision conflict errors, publish summary, Draft Preview, and draft/published asset state.
- Backend expected-red contract tests cover Current revision fields, Previous restore, Published Content snapshot queries, Draft Preview denial, revision-conflict mapping, search over `published_file_contents`, and draft/published asset isolation.
- Frontend expected-red contract test covers revision typing, autosave status states and triggers, required-save blocking, conflict actions, publish state labels, Draft Preview, editor/preview split, and Draft/Published asset UX.
- `docs/verification/stage-3-security.md` covers Draft Preview denial, draft asset leakage, stale revision overwrite, unsafe assets, and MCP enable/disable/audit/backup/kill-switch expectations.
- This acceptance plan covers DB/API/browser/MCP black-box scenarios and evidence layout.

Fixture gaps to close before Gateway 2/3/4/6 acceptance execution:

1. Record concrete Stage 3 fixture IDs only after the migration lands: root path, file IDs, starting Current revisions, Previous slot state, Published Content source revisions, draft asset IDs, and published asset IDs.
2. Add exact SQL hash/count probes after the final migration names and table names are stable; acceptance should compare Current/Previous/Published rows and draft/published asset rows before and after Publish.
3. Add exact MCP stdio command and audit-log path after the MCP package/process exists; until then the MCP matrix remains the required black-box checklist.
4. Re-run focused expected-red backend tests on a leader SHA that is not mid-Gateway-2 partial integration. The review run found expected-red coverage present, but the current leader working tree also had downstream build errors around `listFileAssetsByState`, so that failure should be treated as a Gateway 2 implementation watch item rather than a Gateway 1 contract gap.
5. Keep dependency artifacts out of acceptance commits: no `web/node_modules`, symlinks, caches, `dist`, local DB/uploads, or `.omx` runtime state.

Gateway 1 acceptance can proceed to implementation gates once the coordinator records the reviewed integrated SHA and backend/frontend lanes acknowledge the expected-red tests as intentional pre-implementation failures.


## Gateway 6 MCP acceptance smoke — 2026-06-14

Verdict: **PASS for MCP acceptance smoke on integrated repair SHA `47bea64`**.

Evidence source: `mcp/tests/stdio-smoke.test.mjs` plus direct MCP unit tests in `mcp/tests/skeleton.test.mjs`.

Covered acceptance matrix items:

- G1 disabled by default: stdio client can list tools, but `health_check` refuses with `Blog MCP disabled`; JSONL audit records `refused`.
- G2 enabled local stdio only: tests connect with `StdioClientTransport`; production grep finds only `StdioServerTransport` and no HTTP/SSE listener.
- G3-G8 tools: stdio smoke calls all required read/search/content/publish/tree/assets/maintenance tools against a disposable local HTTP backend stub, proving backend-boundary requests and `Authorization: Bearer ...` propagation when configured.
- G8 backup/export: safe `export_backup { "label": "acceptance" }` writes under `BLOG_MCP_BACKUP_DIR`; traversal, absolute path, and symlink escape labels are rejected before backend calls or writes.
- G9 audit: representative ok/error/refused calls produce JSONL entries with timestamp/tool/result and redacted or summarized args.
- G10 kill switch: `BLOG_MCP_KILL_SWITCH=true` refuses an enabled stdio call.
- G11 no direct SQL: static grep over `mcp/src` has no DB/SQL matches; tool handlers use `BlogBackendClient`.

Transcript summary:

```text
PASS cd mcp && npm test
  tests: 15 pass / 0 fail
  includes stdio disabled, kill switch, safe/rejected backup, full tool surface backend-boundary smoke
PASS cd mcp && npm run build
PASS git diff --check
PASS grep -R "pgx\|database/sql\|SELECT \|INSERT \|UPDATE \|DELETE \|SQL" -n mcp/src -> no matches
PASS grep -R "listen\|createServer\|Sse\|SSE\|StreamableHTTP\|StdioServerTransport" -n mcp/src mcp/package.json mcp/README.md -> no production public MCP transport; stdio only
```

## Final integrated acceptance evidence — 2026-06-14

Verdict: **PASS for automated gates, API smoke, browser smoke, and MCP smoke on integrated SHA `6224134`**.

Evidence artifacts under `docs/verification/stage-3-browser-20260614/`:

- `final-gates-6224134.txt` — backend `go test`/`go vet`/gofmt, frontend node tests/lint/build, MCP tests/build, and `git diff --check` all PASS.
- `stage3-api-smoke-6224134.txt` — local PostgreSQL API smoke covering Author login, create Directory/File, revision save, stale revision conflict, Draft Preview `iframe_sandbox=allow-scripts`, Anonymous Preview 401, publish, public resolve, unpublish, and public 404 after unpublish.
- `stage3-author-workspace-6224134.png` — browser screenshot of Author Workspace with Content Tree, selected File, Publish summary, Current/Previous, and Draft Preview.
- `stage3-author-after-unpublish-6224134.png` and `stage3-browser-after-unpublish-6224134.txt` — browser evidence after unpublish showing File state as Draft and Publish action available.
- `stage3-browser-unpublish-api-state-6224134.txt` — browser/network/API evidence for `POST /api/admin/files/{file_id}/unpublish`, `current_status=draft`, `published_visible=false`, and public resolve 404.

Final gate transcript summary:

```text
PASS cd api && CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./...
PASS cd api && CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go vet ./...
PASS cd api && test -z "$(gofmt -l .)"
PASS cd web && node --test tests/*.test.mjs  # 43/43
PASS cd web && npm run lint
PASS cd web && npm run build
PASS cd mcp && npm test  # 15/15
PASS cd mcp && npm run build
PASS git diff --check
```

Notes:

- Frontend dependencies were installed locally to run TypeScript/build gates; `web/node_modules/` and `web/dist/` remain ignored and untracked.
- The browser smoke used the local Author account configured by `~/.local/share/xlab-blog/start-local.sh` and disposable Stage 3 smoke content in the local database.
- The MCP acceptance/security PASS is separately recorded in this document and `docs/verification/stage-3-security.md`; the final gate transcript includes MCP tests/build after the backup path hardening repair.
