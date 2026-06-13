# Stage 3 Frontend Readiness — Gateway 4 Planning

Date: 2026-06-13 CST  
Owner: worker-3 frontend lane  
Status: **readiness plan only; production UI blocked until backend Gateway 2/3 runtime APIs are green**

## Scope boundary

- Task 8 already added the expected-red frontend contract test:
  `web/tests/stage3-author-workspace-contract-red.test.mjs`.
- This note maps the integrated Gateway 1 OpenAPI contract to frontend code
  changes for Gateway 4.
- No production UI changes were made here. Gateway 4 implementation remains
  blocked by backend Gateway 2/3 runtime APIs.

## API/type mapping

Planned `web/src/lib/types.ts` additions:

- `FileContentVersion`: `node_id`, `revision`, `content_format`, `keywords`,
  `body_raw`, optional `body_html`, `search_text`, `status`, `last_saved_at`,
  embedding fields.
- `PublishedContentSnapshot`: `node_id`, `source_revision`, `content_format`,
  `keywords`, `body_raw`, optional `body_html`, `search_text`, `published_at`,
  `updated_at`, `visible`.
- `FileVersionState`: `current`, optional `previous`, optional `published`,
  `has_unpublished_changes`, `draft_assets`, `published_assets`.
- `PublishSummary`: `current_revision`, optional `published_source_revision`,
  `will_update_content`, `draft_assets`, `published_assets`, `asset_changes`.
- `PublishResult`: `current`, `published`, `promoted_assets`.
- `DraftPreviewPayload`: `node`, `current`, `html`, `assets`,
  `iframe_sandbox`.
- `FileAsset.state`: `draft | published | draft_and_published`, with optional
  `published_asset_id`.

Planned `web/src/lib/api.ts` helpers:

- `fetchFileVersionState(fileId)` → `GET /admin/files/{file_id}/content`.
- `saveCurrentFileContent(fileId, { expected_revision, content_format, body_raw, keywords })`
  → `PUT /admin/files/{file_id}/content`.
- `restorePreviousContent(fileId, { expected_revision })`
  → `POST /admin/files/{file_id}/previous/restore`.
- `fetchPublishSummary(fileId)` → `GET /admin/files/{file_id}/publish-summary`.
- `publishCurrentFile(fileId, { expected_revision })`
  → `POST /admin/files/{file_id}/publish`.
- `unpublishFile(fileId)` keeps current helper path, but return type becomes
  `FileVersionState`.
- `fetchDraftPreview(fileId)` → `GET /admin/preview/{file_id}`.
- `fetchFileAssetState(fileId)` → `GET /admin/files/{file_id}/assets`.
- `uploadDraftAsset(fileId, file)` keeps current upload path, but returns a
  draft `FileAsset`.

Conflict handling:

- Preserve `ApiError.status === 409`.
- Parse optional `details.reason === "revision_conflict"` and
  `details.current_revision` if backend includes them.
- UI must enter `Conflict` state and offer only `Reload latest` and
  `Copy my changes`; no auto-merge.

## Hook/component split

Planned hooks/components:

- `useAutosaveFile`
  - Owns local editor text, keywords, content format, latest `revision`,
    save status, conflict state, and pending required-save promise.
  - Debounces 15 seconds after input stops.
  - Exposes `saveNow(reason)` for blur, node change, publish, logout, and
    leaving Author Workspace.
  - On failed required save, blocks unsafe transition and preserves local text.
- `useUnsavedNavigationGuard`
  - Registers `beforeunload` and in-app guard.
  - Calls `saveNow("leave")`; blocks when save fails or conflict exists.
- `VersionPanel`
  - Shows Current/Previous timestamps.
  - Provides compare and reversible restore.
- `PublishControls`
  - States: `Publish`, `Publish changes`, `Published`.
  - Fetches publish summary before publish.
  - Keeps `Unpublish` in overflow/danger area.
- `AssetStatePanel`
  - Separates `Draft assets` and `Published assets`.
  - Explains draft uploads are not public until Publish.
  - Makes draft deletion semantics clear when a published snapshot still
    references the previous public asset set.
- `PreviewSplit`
  - Desktop editor/preview split.
  - Draft Preview uses saved Current content and draft assets.
  - HTML document iframe sandbox remains `allow-scripts` only, without
    `allow-same-origin`.

## State machine

Required visible states:

1. `Editing` — local draft differs from last saved Current Content.
2. `Saving` — save request in flight.
3. `Saved` — Current Content persisted; display `last_saved_at`.
4. `Save failed` — network/server failure; preserve typed text and block
   required transitions.
5. `Conflict` — stale revision 409; block publish/navigation until user chooses
   `Reload latest` or `Copy my changes`.
6. `Unpublished changes` — Current revision/content or draft assets differ from
   Published Content snapshot.

Required save triggers:

- 15-second debounce after input stops.
- Blur from editor fields.
- Selecting another node.
- Publish / Publish changes.
- Logout.
- Leaving Author Workspace / browser unload.

## Test expectations

Existing red contract test should turn green only after production Gateway 4
implementation:

```bash
cd web
node --test tests/stage3-author-workspace-contract-red.test.mjs
```

Gateway 4 should also add behavior-level tests once runtime APIs are available:

- fake timers for 15-second debounce;
- blur save;
- node-change required save block;
- publish required save block;
- logout/leave required save block;
- 409 conflict actions;
- Current/Previous restore flow;
- publish summary and button state transitions;
- Draft Preview role/API wiring;
- Draft/Published asset separation.

## Current blockers

- Backend Gateway 2/3 runtime APIs are not green yet.
- Frontend must not implement production UI against missing runtime behavior.
- `web/node_modules`, `dist`, caches, local DB/uploads, and `.omx` runtime state
  must not be added or committed.
