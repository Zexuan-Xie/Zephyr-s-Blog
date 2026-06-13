# Stage 3 Security and Abuse Review

Status: Gateway 0/1 planning complete; implementation review pending

Reviewed at Stage 3 start on `73bcc9e` (`checkpoint: stage 2 polish before stage 3`).

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

## Current verdict

Gateway 0/1 may proceed if the coordinator, backend, and verifier incorporate the
requirements above into OpenAPI and red tests before implementation. Stage 3 is
not security-approved until later gates provide migration, API, browser, asset,
and MCP evidence against this checklist.

## Gateway 1 contract review — OpenAPI and red tests

Reviewed after backend task 3 completed on `91d5f57` with Gateway 1 red-test
commits reported as `66d5829`, `8b03e62`, and `91d5f57`.

### Positive findings

- OpenAPI now defines Author-only Current/Previous/Published version surfaces:
  `GET/PUT /admin/files/{file_id}/content`, Previous restore, publish summary,
  publish, unpublish, and Draft Preview.
- Save and Publish requests require `expected_revision`, and the save conflict
  response documents a machine-readable `revision_conflict` with
  `current_revision`.
- Draft Preview is explicitly admin-routed and documents 401/403 denial for
  Anonymous Visitor and Reader access.
- Draft/Published Asset surfaces are now modeled, including an Author-only draft
  asset byte route and copy that public `/assets` routes serve only Published
  Assets.
- Expected-red tests cover key missing implementation seams: revision fields,
  lifecycle repository methods, migration tokens, public tree/search source switch
  to `published_file_contents`, draft/published asset repository methods, admin
  route exposure, Reader denial for Draft Preview, and lost-update error mapping.

### Security findings to carry into Gateway 2/3

1. **Public file schema still references mutable `FileContent`.**
   `FilePage.content` still points at `FileContent`, which is now the Current
   Content schema and includes revision/autosave/embedding fields. Public file
   responses should use `PublishedContent` or a dedicated public content DTO so
   Readers never receive Current Content metadata or an implementation contract
   that encourages public reads from mutable content.
2. **Unpublish lacks an optimistic-concurrency request.**
   Publish has `expected_revision`, but Unpublish has no request body. Because
   Unpublish is a public-visibility destructive operation, Gateway 2/3 should
   either require an expected revision/version or document why it is safe without
   one.
3. **Draft asset DTO requires `public_url`.**
   `FileAsset.public_url` is required for every asset state, including draft
   assets. This risks client misuse or accidental leakage. Draft assets should
   expose either no URL or an explicitly protected admin preview URL that cannot
   be fetched by Anonymous/Reader.
4. **Draft asset byte route needs denial/path tests.**
   OpenAPI defines `GET /admin/assets/{asset_id}/{filename}`, but Gateway 1 tests
   do not yet prove Anonymous/Reader denial, filename mismatch handling, or path
   traversal rejection for this route.
5. **Conflict implementation must return structured details.**
   The new handler test correctly expects `ErrLostUpdate` to map to 409 with
   `revision_conflict`. Gateway 2/3 should also include `current_revision` in the
   response so the frontend can implement Reload latest / Copy my changes without
   guessing.
6. **Search/tree red tests are source-token guards.**
   They are useful as early red tests, but Gateway 2 acceptance still needs DB/API
   tests proving autosaved Current changes do not appear in public file/tree,
   recent, search, comments/likes existence checks, or asset responses until
   Publish.
7. **MCP remains unreviewed.**
   Task 3 did not implement or contract MCP. The initial MCP requirements above
   remain open for the later MCP gateway.

### Gateway 1 security verdict

PASS for proceeding to Gateway 2 with follow-up findings. The contracts and red
suite now cover the core Stage 3 abuse cases, but the seven findings above must
be resolved or explicitly risk-accepted before final Stage 3 security approval.

## Gateway 2/3 backend implementation security review

Reviewed after backend tasks 9 and 10 completed on `main` at `db1633c`
(latest backend checkpoint reported by task 10: `657e669`). Scope was backend
migration/core model plus HTTP/Draft Preview/draft-asset routes. Frontend and MCP
security remain pending.

### Verdict: REVISE

The backend moved public tree/search/recent/file reads substantially toward the
Published Content model, and Draft Preview routing is protected by `RequireAdmin`.
However, the implementation still has security/contract blockers before this gate
can pass.

### Blocking findings

1. **HIGH — Draft assets for visible published files can be fetched through the
   public asset route before Publish.**
   `api/internal/assets/repository.go:55-61` only checks that the owning File has
   visible Published Content; it does not require the asset itself to be in a
   published state or in `published_file_assets`. Newly uploaded draft assets on a
   published File receive `PublicURL` in `scanAsset` (`api/internal/assets/repository.go:150`)
   and are returned from Author draft state. A guessed or leaked `/api/assets/{id}/{filename}`
   URL can therefore expose draft bytes before Publish. Public asset serving must
   join the published asset snapshot/state, not only visible Published Content.

2. **HIGH — Publish can bypass the required `expected_revision`.**
   `api/internal/http/handlers/tree_lifecycle.go:158-176` silently ignores JSON
   decode errors and falls back to `PublishFile` when `expected_revision` is absent
   or zero. That defeats the OpenAPI contract and allows stale clients to publish
   the latest Current Content without proving they observed the current revision.
   This undermines the stale-tab protection Stage 3 requires. Invalid JSON should
   be 400, and Publish should require a positive expected revision.

3. **HIGH — Public comment/like existence checks still use mutable
   `file_contents.status='published'`.**
   `api/internal/comments/repository.go:24-30` and
   `api/internal/likes/repository.go:21-27` were not migrated to
   `published_file_contents.visible`. After autosaving changes to a published File,
   Current status becomes `unpublished_changes` while Published Content remains
   visible, so Readers may be incorrectly blocked from commenting/liking public
   content. Conversely, these checks are no longer the single public visibility
   source of truth. They must use Published Content visibility.

4. **MEDIUM — `revision_conflict` responses omit `current_revision`.**
   `api/internal/http/handlers/tree_lifecycle.go:221-222` returns structured
   `reason: revision_conflict`, but omits the documented `current_revision`. The
   frontend conflict flow cannot reliably implement Reload latest / Copy my
   changes without fetching extra state, and the OpenAPI example promises this
   field.

5. **MEDIUM — Public File DTO still uses the mutable Current `FileContent` type.**
   Public `FilePage.Content` remains `FileContent` (`api/internal/tree/types.go:185-188`).
   The repository currently fills it from `published_file_contents`
   (`api/internal/tree/repository.go:91-120`), so this is not a direct draft-body
   leak in the reviewed code, but it exposes Current-only fields such as revision,
   last_saved_at, and embedding metadata on public responses and keeps API
   semantics coupled to the Author Current model. Use `PublishedContent` or a
   dedicated public DTO.

6. **MEDIUM — Unpublish has no concurrency guard and is not atomic.**
   `api/internal/http/handlers/tree_lifecycle.go:179-189` accepts no expected
   revision/version, and `api/internal/tree/lifecycle_repository.go:283-298`
   updates `file_contents` and `published_file_contents` in separate statements
   outside a transaction. This can race with publish/save flows and can leave
   inconsistent Current vs Published visibility if the second statement fails.

### Non-blocking positive findings

- Public tree, recent, resolve redirects, and search now read visible
  `published_file_contents` rather than mutable Current content.
- Draft Preview is mounted under `/api/admin` with `RequireAdmin`, and the router
  preserves Reader denial even when lifecycle dependencies are absent.
- Draft Preview reports `iframe_sandbox: allow-scripts`, preserving the no
  `allow-same-origin` contract at the API layer.
- Draft asset bytes are mounted only under `/api/admin/assets/...` and use
  `Cache-Control: no-store`.

### Required repair evidence

Before this backend security slice can pass, add/repair tests proving:

- public `/api/assets/{asset_id}/{filename}` denies draft-only assets on published
  Files until Publish promotes them;
- Publish without a positive `expected_revision` returns 400/409 and cannot fall
  back to stale publish;
- comments and likes accept visible Published Content even after Current has
  unpublished changes, and deny only when `published_file_contents.visible=false`;
- `revision_conflict` includes `current_revision`;
- Unpublish either requires an expected revision/version or documents and tests a
  safe atomic alternative.
