# Stage 3 Security and Abuse Review

Status: Gateway 0/1 planning complete; Gateway 2/3/4 implementation review completed with blockers

Reviewed at Stage 3 start on `73bcc9e` (`checkpoint: stage 2 polish before stage 3`).

Updated implementation review on `9b51d36` after backend Gateway 2/3 and
frontend Gateway 4 landed.

This document is the security lane source of truth for Stage 3. It covers the
initial Gateway 0/1 threat model and the security requirements that must be
proved by later migration, backend, frontend, asset, MCP, and acceptance gates.
It is not a final PASS for Stage 3 implementation.

## Baseline observations

- Stage 2 admin routes are mounted below `/api/admin` with `RequireAdmin` in
  `api/internal/http/router.go`.
- Public asset serving currently goes through `AssetService.OpenPublished`; Stage
  3 must preserve that public-only boundary after adding Draft Assets.
- Current asset validation already normalizes filenames and rejects obvious SVG
  active content in `api/internal/assets/validation.go`; Stage 3 must keep and
  extend filename/path traversal tests.
- HTML document rendering must preserve the iframe sandbox contract:
  `sandbox="allow-scripts"` without `allow-same-origin`.
- Existing Stage 2 note still applies: JWT in localStorage is accepted for this
  local personal-blog scope, but XSS remains high impact.

## Threat model

### Actors

| Actor | Expected access | Main abuse risks |
| --- | --- | --- |
| Anonymous Visitor | Public Published Content, Published Assets, public search/recent/tree, login/register | Draft Preview leakage, Draft Asset guessing, path redirect abuse, HTML sandbox escape |
| Reader | Anonymous access plus comments/likes as authenticated Reader | Privilege escalation to Author APIs, draft route access, CSRF-like destructive actions if token leaks |
| Author | Full Author Workspace and Stage 3 MCP-equivalent authority | Accidental destructive operations, stale-tab overwrite, unsafe asset upload, publish of wrong snapshot |
| Server-local MCP client | Full trusted Author operations only when explicitly enabled | Tool prompt abuse, destructive batch without backup/audit, operation while disabled, duplicate business logic bypassing checks |
| Network/local attacker | No direct privileged access | Asset path traversal, public route probing, stale revision replay, migration/backup exposure |

### Protected assets

- Current Content Version and Previous Content Version.
- Independent Published Content snapshot.
- Draft Preview output.
- Draft Assets and unpublished asset metadata.
- Published Assets and their stable public URLs.
- Author tokens, MCP enable/disable config, audit logs, and backups/exports.
- Content Tree path/redirect integrity.

## Authorization matrix required for Stage 3

| Surface | Anonymous | Reader | Author | MCP enabled local stdio |
| --- | ---: | ---: | ---: | ---: |
| Public tree/resolve/recent/search | Published only | Published only | Published only outside Author routes | Read through approved tool |
| Public file page | Published Content only | Published Content only | Published Content only outside Author routes | Read through approved tool |
| Public asset endpoint | Published Assets only | Published Assets only | Published Assets only outside Author routes | Read through approved tool |
| `/api/admin/*` | Deny 401/403 | Deny 403 | Allow | Not directly applicable unless API-client uses Author auth/config |
| Draft Preview `/admin/preview/{file_id}` | Deny | Deny | Allow saved Current + Draft Assets | Tool may expose equivalent only with audit |
| Draft Asset bytes/list | Deny | Deny | Allow | Allow only when enabled and audited |
| Save Current / restore Previous | Deny | Deny | Allow with expected revision | Allow only when enabled and audited |
| Publish/unpublish | Deny | Deny | Allow after required save succeeds | Allow only when enabled and audited |
| Destructive tree/assets tools | Deny | Deny | Allow with existing validation and confirmations where applicable | Allow only with audit plus backup/export where practical |

## Gateway 1 OpenAPI/security requirements

OpenAPI and red tests must make the following behavior machine-readable before
implementation proceeds:

1. **Optimistic concurrency**
   - Save Current requires `expected_revision`.
   - Stale writes return `409 conflict` and never overwrite Current.
   - Conflict response contains enough metadata for `Reload latest` / `Copy my changes`.
2. **No-op vs changed save**
   - No-op save must not rotate Previous.
   - Changed save rotates the old Current into Previous and increments revision.
3. **Published Content isolation**
   - Public file/tree/recent/search responses are backed only by Published Content.
   - Autosaved Current changes never appear publicly until Publish.
   - Unpublish hides public visibility while retaining Published Content metadata.
4. **Draft Preview protection**
   - Draft Preview requires Author.
   - Anonymous Visitor and Reader denial must be explicit 401/403, not redirect loops.
   - Draft Preview uses saved Current content and Draft Assets, not unsaved client text.
5. **Draft/Published Asset isolation**
   - Upload creates Draft Asset state only.
   - Public asset endpoints serve only Published Assets.
   - Removing a Draft reference must not break already Published Assets until Publish/Unpublish.
   - Filenames, MIME types, SVG payloads, and path traversal attempts remain rejected.
6. **Search and embeddings**
   - Search indexes Published Content only.
   - Semantic embedding failure remains non-blocking and cannot fake a failed save as success.
7. **Error mapping**
   - Use stable 400 validation, 401 unauthenticated, 403 unauthorized, 404 not found,
     409 conflict/lost update, and 500 real server error responses.

## MCP security requirements

The Blog MCP Server is high trust, but it must not silently widen the attack
surface.

- Server-local stdio only for Stage 3; no public HTTP/SSE transport.
- Explicit opt-in such as `BLOG_MCP_ENABLED=true`.
- Emergency disable/kill switch checked before every tool call, not only at startup.
- JSONL audit entry for every tool call with timestamp, tool name, argument
  summary, result/error, and destructive-operation marker.
- Backup/export before destructive batches where practical: delete, move subtree,
  bulk publish/unpublish, rebuilds, or maintenance operations.
- MCP handlers must reuse backend service/API-client boundaries. No direct SQL or
  duplicate authorization/business rules in MCP handlers.
- Disabled MCP smoke must prove every tool refuses operation.
- Enabled MCP smoke must prove read, content update, publish/unpublish, reorder,
  asset list, and export backup paths are audited.

## Red-test checklist for later gates

- Anonymous/Reader cannot access `/api/admin/*`, Draft Preview, draft asset bytes,
  or any MCP operation.
- Public page/search/recent still show old Published Content after autosave and
  change only after Publish.
- Two-tab stale revision save returns 409 and preserves both server and client text.
- Required save failure blocks publish, logout/leave, and node switch.
- Restore Previous changes Current only; public content remains stable until Publish.
- Unpublish hides public resolve/search/recent while retaining Published Content.
- Draft-only asset upload is not public by guessed URL or listed public metadata.
- Published Asset remains available after draft-only removal until next explicit
  Publish/Unpublish rule changes it.
- SVG/script/path traversal payloads remain rejected.
- Redirect loops/cycles/path rewrite attacks remain rejected.
- MCP disabled state refuses all tools; MCP enabled state audits and backs up
  destructive operations.

## Gateway 2/3/4 implementation review

### Review verdict

**REQUEST CHANGES before Stage 3 security approval.** The implemented backend and
frontend preserve the high-level admin route boundary, but two asset/DTO issues
can leak draft implementation details or draft bytes, and MCP is not implemented
yet. Treat the frontend Gateway 4 UI as functionally complete but not as a
security closeout until the asset and MCP gates below are fixed and verified.

### Findings

#### HIGH — Public asset route can serve draft-only assets for already published files

Evidence:

- Public asset route `/api/assets/{asset_id}/{filename}` is mounted outside
  `/api/admin` and calls `AssetHandler.ServePublished`.
- `api/internal/assets/repository.go` `FindPublishedAsset` checks only that the
  asset belongs to a file with a visible `published_file_contents` row; it does
  **not** require the asset itself to be in the published snapshot:
  `join published_file_contents pfc on pfc.node_id = n.id and pfc.visible`.
- Stage 3 publish writes a separate `published_file_assets` snapshot in
  `api/internal/tree/lifecycle_repository.go`, but the public asset lookup does
  not read that table.

Risk:

- If an Author uploads a new draft asset to a file that is already visible, the
  new asset row has `state='draft'`, but the file still has visible Published
  Content. A guessed or leaked `/api/assets/<draft_asset_id>/<filename>` URL can
  be served publicly before Publish, violating Draft Asset isolation and the
  acceptance requirement that Draft uploads are not public until Publish.

Required fix:

- Change public asset lookup to join `published_file_assets` by
  `published_asset_id`/`asset_id`/filename or otherwise prove membership in the
  last Published Asset snapshot, not merely file visibility.
- Add a regression test where a published file receives a draft-only upload and
  the public asset endpoint returns 404 until Publish.

#### HIGH — Public/admin asset DTOs expose `storage_key`

Evidence:

- `api/internal/assets/types.go` and `api/internal/tree/types.go` expose
  `StorageKey` with `json:"storage_key"`.
- `scanAsset` returns that field from repositories, and handlers return the asset
  structs directly for upload, list, preview, and public file assets.
- `docs/api/openapi.yaml` also documents `storage_key` on `FileAsset`.

Risk:

- Even though `LocalStorage.pathForKey` rejects traversal and absolute paths,
  provider-neutral storage keys are internal implementation details. Returning
  them in public file payloads, draft preview payloads, and Author asset lists
  widens the information surface and makes future object-storage migration or
  signed URL policies harder to secure.

Required fix:

- Introduce public/admin DTOs that omit `storage_key`; keep storage keys internal
  to repository/service/storage layers.
- Update OpenAPI and frontend schemas to consume only `id`, `filename`,
  `mime_type`, `size_bytes`, `public_url`, and state metadata that is truly
  needed by the Author UI.

#### MEDIUM — Revision conflict response is machine-readable but lacks documented current revision

Evidence:

- OpenAPI example requires `details: { reason: revision_conflict,
  current_revision: 7 }`.
- `TreeLifecycleHandler.respondError` currently returns only
  `details.reason = revision_conflict`.

Risk:

- The UI can still offer “Reload latest” / “Copy my changes”, but clients cannot
  display or log the server-side Current revision without doing another fetch.

Recommended fix:

- Return `current_revision` when the service/repository can provide it, or update
  OpenAPI to remove the stronger contract if the extra fetch is intentional.

#### WATCH — Unpublish hides content in two statements without transactional proof

Evidence:

- `UnpublishFile` first updates `file_contents.status='draft'`, then separately
  updates `published_file_contents.visible=false`.

Risk:

- A mid-operation database error can leave author-facing Current status and public
  Published visibility inconsistent. The public path is still controlled by
  `published_file_contents.visible`, so this is not an immediate public leak, but
  it weakens DB/API snapshot proof.

Recommended fix:

- Wrap unpublish in a transaction or document the intentional consistency model
  with a focused test that proves public visibility is eventually false and
  Published Content metadata is retained.

#### WATCH — MCP implementation/evidence remains absent

Evidence:

- Repository search found no MCP server package/process, only Stage 3 plans and
  acceptance/security requirements.

Risk:

- MCP disable-by-default, kill-switch-before-every-call, audit JSONL,
  backup/export, and “no direct SQL in handlers” cannot be security-approved yet.

Required later gate:

- Repeat this security review after the MCP slice lands and attach disabled and
  enabled stdio smoke transcripts plus audit/backup evidence.

### Positive checks

- `/api/admin` routes, including Draft Preview and draft asset byte routes, are
  mounted under `RequireAdmin` in `api/internal/http/router.go`.
- Draft Preview returns `iframe_sandbox: "allow-scripts"` and frontend
  `AdminPage.tsx` contains no `allow-same-origin`.
- Local storage key resolution rejects absolute paths, `..`, backslash traversal,
  and verifies the resolved path remains under the configured asset root.
- Public tree/file/search code paths use `published_file_contents` rather than
  Current Content for public content visibility.

### Required follow-up verification

- Add and run backend tests for:
  - anonymous/Reader denial for `/api/admin/preview/{file_id}`;
  - anonymous/Reader denial for `/api/admin/assets/{asset_id}/{filename}`;
  - public 404 for draft-only asset bytes after upload to an already published
    file, followed by public 200 only after Publish;
  - public file/asset DTOs do not contain `storage_key`;
  - unpublish public visibility and DB snapshot consistency.
- Run Stage 3 black-box API/browser/MCP acceptance after MCP lands.

## Task 15 post-repair backend security re-review

Reviewed integrated repaired HEAD `97acc9e` (includes task-14 repair commits `71c9d93`/`e4e2dae`) against the prior Task 12 REVISE blockers.

### Verdict: PASS for backend Gateway 2/3 security repair

MCP-specific security remains out of scope until the MCP slice lands, but the backend HTTP/Draft Preview/public asset repair items from Task 12 and Task 14 are resolved.

### Evidence checked

- Public asset bytes now use `published_file_assets` joined to visible `published_file_contents` in `api/internal/assets/repository.go::FindPublishedAsset`; draft-only assets are no longer authorized merely because the parent File is published.
- Draft/admin asset byte route remains under `/api/admin/...` and `RequireAdmin` in `api/internal/http/router.go`; public route calls `OpenPublished`, admin route calls `OpenDraft`.
- `CreateAsset` now returns `state` and `published_asset_id`, matching `scanAsset` and removing the post-insert scan mismatch/orphan-row failure path.
- Publish and unpublish handlers require valid JSON and positive `expected_revision`; invalid JSON and missing/zero/negative revisions return 400.
- Unpublish now takes `expected_revision`, locks Current Content with `for update`, and updates Current status plus Published visibility inside one transaction.
- Comment and like visibility checks use `published_file_contents.visible` instead of mutable `file_contents.status`.
- `revision_conflict` responses include `details.current_revision` when version state can be read.
- Public `FilePage.Content` uses `PublicFileContent`, and `storage_key`/`storage_provider` are JSON-hidden and removed from OpenAPI/web schemas; focused response tests check they are not leaked.
- Draft Preview iframe sandbox remains `allow-scripts`; no `allow-same-origin` appears in `web/src/pages/AdminPage.tsx`.
- SQL for publication/asset snapshot behavior remains in repository packages.

### Verification commands

- PASS `cd api && CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./internal/assets ./internal/http ./internal/http/handlers ./internal/comments ./internal/likes ./internal/tree -run 'Stage3|Lifecycle|ServiceUpload|RouterExposesAssetRoutes|CreateAsset|Publish|Unpublish|Asset'`
- PASS `cd api && test -z "$(gofmt -l internal/assets internal/http internal/comments internal/likes internal/tree)" && CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go vet ./internal/assets ./internal/http ./internal/http/handlers ./internal/comments ./internal/likes ./internal/tree`
- PASS `grep -R 'json:"storage_key\|json:"storage_provider\|allow-same-origin\|fc.status = "'"'"'published"'"'"'' -n api web docs/api/openapi.yaml | head -120` only matched negative regression-test strings and OpenAPI prose forbidding `allow-same-origin`.
- PASS `git diff --check`.

### Remaining later gate

Repeat security review for MCP once implemented: disabled-by-default behavior, per-call kill switch, audit JSONL, destructive tool confirmation, backup/export, and no direct SQL in MCP handlers are still pending future evidence.

## Task 19 Gateway 6 MCP security review

Reviewed integrated leader HEAD `d4c7a5f` after MCP tasks 16 and 17 completed. Scope: server-local MCP package under `mcp/`, tool registration, backend API boundary, disabled/kill-switch guards, audit JSONL, destructive confirmations, backup/export, direct SQL/DB avoidance, stdio-only transport, input/path validation, stale revision parity, and protected API auth boundary.

### Verdict: REVISE

The MCP implementation satisfies several core architectural controls, but it cannot pass the security gate until backup/export path validation is fixed and tested.

### Blocking finding

1. **HIGH — `export_backup` writes to an arbitrary caller-supplied path without canonicalization or allowlist.**
   `mcp/src/tools.ts` accepts `output_dir` as any non-empty string, then `mcp/src/backendClient.ts:155-158` calls `mkdir(args.outputDir, { recursive: true })` and writes `path.join(args.outputDir, aeolian-backup-*.json)` with `createWriteStream(..., { flags: "wx" })`. There is no `path.resolve`, `realpath`, base-directory allowlist, or traversal/absolute-path denial. Any enabled MCP client can direct backup writes to arbitrary local directories writable by the MCP process. This violates the Task 19 path/input validation requirement and turns the trusted backup helper into a local file-write primitive. Required fix: constrain backups to a configured backup root such as `BLOG_MCP_BACKUP_DIR` or `~/.local/share/xlab-blog/mcp-backups`, resolve/canonicalize paths, reject traversal/absolute outputs outside that root, and add tests for `..`, absolute paths, symlink/realpath behavior where feasible, and successful safe export.

### Positive findings

- Disabled-by-default and kill-switch gates are centralized in `assertEnabled`, and `runGuardedTool` checks them before input validation or backend calls.
- Every registered tool routes through `runGuardedTool` and writes JSONL audit for `ok`, `error`, and `refused` outcomes with secret-like argument redaction.
- Server startup uses `StdioServerTransport`; grep found no HTTP/SSE MCP listener in `mcp/src`.
- Tool handlers call `BlogBackendClient`; grep found no direct DB clients, repository imports, or SQL in `mcp/src`.
- Protected API boundary is preserved through `/api/admin/...` backend HTTP calls and optional `Authorization: Bearer ${BLOG_ADMIN_TOKEN}`; revision-sensitive tools require positive `expected_revision`/`expected_version` inputs and delegate stale revision checks to backend APIs.
- `delete_node`, `delete_asset`, and `rebuild_search_index` require `confirm=true` before backend calls; missing confirmation is audited as an error.
- Unit smoke tests cover disabled refusal, kill-switch refusal, tool-surface registration, backend auth header use, and destructive delete refusal before backend mutation.

### Verification commands

- PASS `cd mcp && npm test && npm run build`
- PASS `grep -R "pgx\|database/sql\|SELECT \|INSERT \|UPDATE \|DELETE \|SQL" -n mcp/src` produced no matches.
- PASS `grep -R "listen\|createServer\|Sse\|SSE\|StreamableHTTP\|StdioServerTransport" -n mcp/src mcp/package.json mcp/README.md` found only `StdioServerTransport` and README no-HTTP/SSE prose.
- REVISE `grep -R "resolve\|normalize\|realpath\|relative\|outputDir\|output_dir\|createWriteStream\|mkdir" -n mcp/src mcp/tests mcp/README.md` showed arbitrary `outputDir` use and no canonicalization/allowlist tests.
- PASS `git diff --check` before this documentation update.

### Required repair before MCP security PASS

Repair `export_backup` path handling as above, add regression tests, and rerun this MCP security review. Task 18 black-box acceptance should also include safe backup path and rejected unsafe path transcripts.


## Task 19 repair review — MCP backup path hardening

Reviewed integrated leader HEAD `47bea64`, which repairs the Task 19 `export_backup` arbitrary output path finding.

### Verdict: PASS for MCP Gateway 6 security

The previous blocking finding is resolved. `export_backup` no longer accepts a caller-supplied output directory. The tool now uses configured `BLOG_MCP_BACKUP_DIR` (default `~/.local/share/xlab-blog/mcp-backups`) plus an optional relative `label` subdirectory. The implementation rejects absolute labels, `.`/empty/`..` path segments, traversal outside the root, and symlink labels that resolve outside the canonical backup root before any backend calls or backup writes. Backup files are still created with `wx` exclusive-create semantics.

### Evidence checked

- `mcp/src/config.ts` adds `backupDir` from `BLOG_MCP_BACKUP_DIR` or the server-local default.
- `mcp/src/tools.ts` changes `export_backup` input to optional `label` only and passes the configured backup root, not a caller-supplied output directory.
- `mcp/src/backendClient.ts` resolves/canonicalizes the backup root, validates relative labels, creates path segments one at a time, checks `realpath` after every existing/created segment, rejects symlink escapes, and writes the final backup with `flags: "wx"`.
- `mcp/tests/skeleton.test.mjs` covers safe export, traversal rejection, absolute-label rejection, and symlink escape rejection, including no backend call on rejected labels.
- `mcp/tests/stdio-smoke.test.mjs` covers disabled-by-default stdio refusal, kill-switch refusal, safe/rejected `export_backup` over real MCP stdio, and all required tool names calling the backend API boundary with `Authorization` when enabled.

### Verification commands

```text
PASS cd mcp && npm test  # 15/15
PASS cd mcp && npm run build
PASS git diff --check
PASS grep -R "pgx\|database/sql\|SELECT \|INSERT \|UPDATE \|DELETE \|SQL" -n mcp/src -> no matches
PASS grep -R "listen\|createServer\|Sse\|SSE\|StreamableHTTP\|StdioServerTransport" -n mcp/src mcp/package.json mcp/README.md -> production MCP src uses StdioServerTransport only; README states no HTTP/SSE
```

### Remaining note

This is a high-trust, server-local Author MCP surface by design. It remains disabled by default, kill-switch guarded per call, JSONL-audited, stdio-only, and routed through the backend HTTP API boundary rather than direct SQL.

## Post-review MCP security hardening — 2026-06-14

Reviewed repair commit `92c345c` after independent review feedback.

### Verdict: PASS / APPROVE

Resolved review findings:

- Audit availability is now a precondition for enabled MCP tool execution. `runGuardedTool` writes a `started` audit JSONL event before input validation/backend calls; if that append fails, the tool returns an error before mutation. Regression test: `enabled mutating tool refuses before backend call when audit log cannot be opened` uses `/dev/full` and verifies no backend call.
- Backup filesystem policy now lives in `mcp/src/backup.ts` `BackupExportService`, not in `BlogBackendClient`; MCP keeps a clean backend API-client boundary for blog state changes.
- Backup finalization uses a temp file plus hard-link-to-final inside the canonical backup Directory and removes the temp file, preserving exclusive creation and reducing validate-then-open exposure.
- README startup command now matches the actual TypeScript stdio entrypoint.

Final evidence:

```text
PASS cd mcp && npm test  # 16/16
PASS cd mcp && npm run build
PASS MCP direct SQL/DB grep over mcp/src: no matches
PASS MCP transport grep: production source uses StdioServerTransport only
PASS final integrated post-review gates at 92c345c
```
