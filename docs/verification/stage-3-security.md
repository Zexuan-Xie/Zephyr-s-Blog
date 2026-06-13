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
