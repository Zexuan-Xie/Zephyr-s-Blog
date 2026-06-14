# Progress

Last updated: 2026-06-14 11:22 CST

## Current breakpoint

Stage 3 engineering is complete on `main` and is ready for user acceptance.

Current HEAD: `e446424` (`docs: record Stage 3 team shutdown`).

Active Team runtime: `execute-aeolian-blog-a98ab708` has been shut down. Task 12 remains a documented historical backend security REVISE, superseded by task 14 repair + task 15 PASS and later MCP security PASS. Stage 3 engineering closeout is complete; next step is user acceptance.

## Stage status

1. **Stage 1 — Reliability, navigation, and identity:** engineering complete.
2. **Stage 2 — Simple-English Author Workspace and protected Content Tree:** accepted baseline; Stage 2 UI polish and Content Tree behavior are preserved.
3. **Stage 3 — Autosave, Content Versions, Published Content, Draft Preview, Draft/Published Assets, and server-local stdio Blog MCP Server:** engineering complete; user acceptance pending.

## Stage 3 completion summary

Implemented and verified:

- Current/Previous Content Versions with optimistic revision checks and reversible restore.
- Independent Published Content snapshots; public reading/search use Published Content rather than draft Current content.
- Publish / Publish changes / Unpublish flow with revision protection.
- Draft Preview at Author-only protected route with iframe sandbox `allow-scripts` and no `allow-same-origin`.
- Draft/Published Asset state and publication promotion semantics.
- Author Workspace autosave/status UI, conflict handling, editor/preview shell, and Stage 2 no-regression UI tests.
- Separate `mcp/` package for server-local stdio Blog MCP Server:
  - disabled by default via `BLOG_MCP_ENABLED`;
  - per-call kill switch via `BLOG_MCP_KILL_SWITCH`;
  - audit JSONL with pre-operation `started` event before enabled tools mutate/read backend state;
  - no public HTTP/SSE MCP transport;
  - no direct SQL/DB in MCP handlers;
  - blog state changes go through `BlogBackendClient` backend HTTP API boundary;
  - backup/export filesystem policy lives in `BackupExportService`, constrained to `BLOG_MCP_BACKUP_DIR` with traversal/absolute/symlink escape rejection.

## Evidence ledger

Primary Stage 3 evidence:

- Team/coordinator log: `docs/verification/stage-3-team-log.md`.
- Acceptance: `docs/verification/stage-3-acceptance.md`.
- Security: `docs/verification/stage-3-security.md`.
- Code review: `docs/verification/stage-3-code-review.md`.
- MCP skeleton: `docs/verification/stage-3-mcp-gateway6-skeleton.md`.
- MCP tools: `docs/verification/stage-3-mcp-gateway6-tools.md`.
- Final browser/API/gate artifacts: `docs/verification/stage-3-browser-20260614/`.

Important final artifacts:

- `final-gates-6224134.txt` — full gates before independent review repair.
- `final-post-review-gates-92c345c.txt` — full gates after audit/backup boundary hardening.
- `stage3-api-smoke-6224134.txt` — local PostgreSQL API smoke for create/save/conflict/preview/publish/unpublish/public visibility.
- `stage3-author-workspace-6224134.png` — Author Workspace browser evidence.
- `stage3-browser-unpublish-api-state-6224134.txt` — browser/network/API proof for Unpublish hiding Published Content.

## Final verification snapshot

Final post-review gates at code HEAD `92c345c` passed; `bbd835a` only adds documentation/evidence.

```bash
cd api
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./...
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go vet ./...
test -z "$(gofmt -l .)"

cd ../web
node --test tests/*.test.mjs
npm run lint
npm run build

cd ../mcp
npm test
npm run build

cd ..
git diff --check
```

Results:

- Backend: PASS.
- Frontend: PASS, 43/43 node tests.
- MCP: PASS, 16/16 tests.
- Static checks: PASS; MCP source has no direct SQL/DB grep matches and production transport is stdio-only.
- iframe sandbox preservation: PASS; no `allow-same-origin` in `web/src` or `api`.

## Local environment and services

Use Conda environment `blogenv`:

- Node.js `22.22.3`
- npm `10.9.8`
- Go `1.26.4`
- PostgreSQL `17.10`
- pgvector `0.8.1`

Recover local stack:

```bash
~/.local/share/xlab-blog/start-local.sh
curl -fsS http://127.0.0.1:8080/api/health
curl -fsS http://127.0.0.1:5173/ >/dev/null
```

Local Author account seeded by the recovery script:

- Email: `admin@example.com`
- Password: `LocalSmokePass123!`

## User acceptance next step

Ask the user to run the acceptance flow below:

1. Open `http://127.0.0.1:5173/`.
2. Login as Author.
3. Enter `Author` workspace.
4. Create a Directory and File.
5. Type content; wait for autosave or click Save.
6. Confirm Current/Previous and Draft Preview are visible.
7. Publish the File; open it publicly.
8. Edit the File again and confirm the public page still shows the old Published Content until Publish changes.
9. Unpublish and confirm the public URL is hidden.
10. Optionally run MCP smoke from `mcp/README.md` if server-local AI automation is part of the demo.

## Repository hygiene

Do not commit local runtime/build artifacts:

- `.omx/`, `.code-review-graph/`, `.agents/`, `.codex/`, `.claude/`
- `web/node_modules/`, `web/dist/`
- `mcp/node_modules/`
- local DB/uploads/backups/caches/logs/secrets

Historical details remain in Git history, `docs/archive/INITIAL_BUILD_SUMMARY.md`, and `docs/verification/`.
