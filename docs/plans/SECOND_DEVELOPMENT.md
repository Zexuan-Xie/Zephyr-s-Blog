# xLab Blog Debugging and Admin Redesign Plan

Date: 2026-06-06

Status: Requirements locked; implementation not started.

This plan converts the user-acceptance findings and the completed requirements interview into three independently runnable, testable, reversible stages. Each stage requires user acceptance before the next begins.

## 1. Locked scope

### In scope

- Top navigation, the single search entry, and the identity-aware account entry.
- Anonymous, Reader, and Author authorization states.
- Admin Content Tree, graphical Directory/File creation, ordering, moves, and node settings.
- File Content/Assets/Settings workspace.
- Autosave, Current/Previous Content Versions, independent Published Content, Draft Preview, and optimistic concurrency.
- Draft/Published Asset lifecycle.
- Required OpenAPI, backend, database migration, frontend, test, and recovery changes.

### Out of scope except for regression fixes

- Public homepage redesign.
- Recent card redesign.
- Public Directory page redesign.
- Public File reading typography/layout redesign.
- Comment and Like visual redesign.
- Changes to the Glass Ricepaper visual language.
- Cross-Directory drag-and-drop reparenting.
- Admin tree search.
- More than Current/Previous Author history.

## 2. Non-negotiable delivery rules

- Update `docs/api/openapi.yaml` before shared API behavior changes.
- Back up the local database before cleanup or schema migration.
- Preserve existing public reading, comments, Likes, search fallback, assets, redirects, and HTML sandbox behavior.
- End every stage with a runnable local application, a clean regression gate, a Git checkpoint, an updated `PROGRESS.md`, and explicit user acceptance.
- Do not hand off a partially working intermediate state.
- Do not containerize or deploy to the server until all three native stages pass.

## 3. Stage 1 — Reliability, navigation, and identity

### Objective

Repair the observed acceptance defects and remove misleading or duplicate global navigation without changing the content model.

### Work items

1. Add failing regression coverage for the successful-create/false-error bug.
2. Repair create-form event handling so a successful API response cannot be caught as a create failure.
3. Parse API error bodies and map them to actionable Author-facing messages.
4. Introduce one application-level current-user state:
   - loading skeleton while identity resolves;
   - invalid token clears to Anonymous Visitor;
   - network failure exposes retry rather than pretending to be anonymous.
5. Replace the global navigation structure:
   - keep `Recent`;
   - remove the separate `Search` link;
   - retain the nav search as the only search input;
   - remove the permanent `Admin` link;
   - show one identity entry: `Login`, Reader display name, or `Author`.
6. Implement identity interactions:
   - Reader name opens a minimal menu containing `Logout`;
   - `Author` enters `/admin`;
   - Reader `/admin` shows `Author access required` and `Return to Recent`;
   - Anonymous Visitor `/admin` redirects to Login with return target.
7. Simplify `/search` to current query, results, empty state, and error state only.
8. Implement role-aware login/logout destinations and loop prevention.

### Tests

- Static/navigation contract tests.
- Current-user loading, invalid-token, and network-error tests.
- Anonymous `/admin`, Reader `/admin`, and Author `/admin` browser tests.
- Reader and Author login-return tests.
- Successful Directory creation must produce success state, not the generic failure.
- Search must have one input and no duplicate Search link.
- Full existing frontend/backend regression gate.

### Stage 1 acceptance

- Top bar is minimal and identity-correct.
- Anonymous protection is clear rather than misleading.
- Reader authorization failure is distinct from Login.
- Directory creation gives truthful success/error feedback.
- Existing public behavior remains intact.

## 4. Stage 2 — Graphical Admin Content Tree and workspace

### Pre-stage controlled data step

1. Back up the local PostgreSQL database.
2. Preserve `Smoke Notes / Local Smoke Renamed`.
3. Remove accidental `Acceptance 1425`, its child File, and `Acceptance 1426`.
4. Verify backup restoration instructions before continuing.

### Objective

Replace raw Node-ID forms with a graphical, context-aware Admin workspace while retaining the current content save/publication model until Stage 3.

### Backend/API work

1. Add protected Admin Content Tree endpoints that include:
   - all Directories;
   - Draft, Published, and changed Files;
   - lazy same-parent children;
   - child and attention-state metadata.
2. Add transaction-safe sibling reorder behavior for a mixed Directory/File sequence.
3. Make backend node creation authoritative for:
   - Name-to-path-segment normalization;
   - Chinese preservation and English lowercase/hyphen normalization;
   - same-parent numeric conflict resolution during initial creation;
   - final returned URL Path.
4. Add impact-preview support for URL Path changes and cross-Directory moves.
5. Preserve atomic subtree path rewrites and redirects for formerly public Directory/File paths.
6. Add read-only redirect inspection for System status.
7. Ensure explicit URL Path edits remain strict and never receive silent numeric suffixes.

### Frontend work

1. Remove the Admin hero card and build the compact Admin shell:
   - `Content` and current path;
   - context-aware `View site`;
   - page-level Rebuild search, System status, and Logout.
2. Build the complete lazy Content Tree:
   - click to select/open;
   - independent Directory disclosure arrow;
   - selected-node `···` menu;
   - per-browser selection/expansion restoration;
   - Draft/Published/Changes/Save-failed states;
   - collapsed Directory attention state.
3. Build `＋ New`:
   - Directory/File graphical type cards;
   - right-workspace flow, not a modal;
   - readable `Create in /…` context;
   - Directory requires Name only;
   - File requires Name and graphical Markdown/HTML choice;
   - live read-only final URL Path preview;
   - preservation of input while parent changes;
   - discard protection only after Name input.
4. Build same-parent ordering:
   - desktop drag-and-drop;
   - mobile Move up/Move down;
   - auto-save;
   - rollback on failure.
5. Build node `Advanced settings` in the right workspace:
   - Name;
   - URL Path;
   - graphical Move to Directory Picker;
   - sort position;
   - collapsed Technical details;
   - Delete danger zone.
6. Build File workspace tabs:
   - Content;
   - Assets;
   - Settings.
7. Use a single transient Toast for completed operations; keep actionable errors inline.

### Tests

- Admin tree includes Draft-only branches.
- Lazy expansion and restored browser state.
- Creation path normalization for English, Chinese, and mixed names.
- Concurrent same-name creation returns distinct final paths.
- Explicit URL Path conflict does not silently rename.
- Mixed sibling reorder persists atomically and rolls back in the UI on failure.
- Directory move/path change rewrites descendants atomically and creates correct redirects.
- Non-empty Directory and Published File deletion remain protected.
- Mobile tree and move controls.
- Full public regression gate.

### Stage 2 acceptance

- Author creates Directory/File without seeing a Node ID or implementation term `slug`.
- New nodes appear, select, and open automatically.
- Tree management is visual and understandable.
- Moves, ordering, URL changes, and deletion show truthful impact.
- Existing content remains available and unchanged.

## 5. Stage 3 — Autosave, versions, publication snapshots, and Draft Assets

### Objective

Separate Author editing from public publication and make editing automatically durable and reversible.

### Data model and migration

1. Back up the database.
2. Add three content states per File:
   - Current;
   - Previous;
   - independent Published Content.
3. Add monotonic content revision/version for optimistic concurrency.
4. Migrate transactionally:
   - existing Published File → Current + Published Content;
   - existing Draft File → Current only;
   - Previous initially empty.
5. Add Draft/Published Asset state.
6. Migrate existing Assets from File publication state and Published Content references.
7. Verify rollback/restoration before enabling the new frontend.

### Backend/API work

1. Autosave Current with expected revision; reject stale writes.
2. Promote old Current to Previous only after a successful changed save.
3. Do not create a version for identical content/Keywords/Render Format.
4. Atomically swap Current and Previous on restore.
5. Publish Current into independent Published Content.
6. Unpublish only changes visibility and retains Published Content.
7. Add Author-only Draft Preview APIs and Draft Asset serving.
8. Promote only Draft Assets referenced by Current content during Publish changes.
9. Block deletion of a Published Asset referenced by Published Content.
10. Keep full-text/semantic indexing aligned to Published Content, with semantic failure remaining non-blocking.

### Frontend work

1. Replace manual Save with autosave:
   - 15 seconds after input stops;
   - immediately on blur, node change, Publish, Logout, or leaving Admin;
   - `Editing…`, `Saving…`, `Saved`, `Save failed`, and `Conflict`.
2. Block navigation/logout/publication when required save fails.
3. Add optimistic-concurrency conflict UI:
   - `Reload latest`;
   - `Copy my changes`;
   - no automatic merge.
4. Add Version history:
   - Current/Previous timestamps;
   - Compare;
   - Restore previous with reversible swap.
5. Add publication model:
   - `Publish`, `Publish changes`, or read-only `Published`;
   - `Unpublished changes`;
   - confirmation with public path and differences;
   - Unpublish in overflow.
6. Build responsive editor:
   - resizable 55/45 desktop split;
   - Editor/Split/Preview modes;
   - mobile Edit/Preview switch;
   - sandboxed HTML Preview.
7. Add `/admin/preview/{file_id}`:
   - Author only;
   - Current saved content;
   - Draft Assets;
   - cross-tab refresh after successful save;
   - manual Refresh fallback.
8. Implement Asset Draft/Published presentation and publication summaries.

### Tests

- Migration fixture coverage for existing Draft/Published Files and Assets.
- Autosave timing with controlled timers.
- Blur/navigation/publish forced save.
- No-op save creates no Previous version.
- Current/Previous rotation and reversible restore.
- Published Content remains stable through editing and restore.
- Unpublish/re-publish behavior.
- Stale revision returns conflict without overwrite.
- Save failure blocks navigation and preserves text.
- Draft Preview authorization for Anonymous Visitor, Reader, and Author.
- Draft Asset isolation and reference-based promotion.
- Published Asset deletion guard.
- Cross-tab Draft Preview refresh.
- Desktop/mobile editor layouts and HTML sandbox regression.
- Full API/browser/regression smoke.

### Stage 3 acceptance

- Author can write without a manual Save button and can always see save state.
- Previous content can be restored and restoration can be undone.
- Half-written changes never become public.
- Draft Preview accurately follows successfully saved Current content.
- Published Assets and content remain stable until explicit publication.

## 6. Final native acceptance

After Stage 3:

1. Create a clean acceptance Directory and both Markdown/HTML Files.
2. Test identity states, creation, ordering, moves, autosave, conflicts, versions, Draft Preview, publication, redirects, Assets, search, comments, and Likes.
3. Run desktop and mobile browser acceptance.
4. Record evidence in `docs/verification/`.
5. Update `PROGRESS.md`.
6. Obtain user acceptance.
7. Only then proceed to Docker/WSL Compose smoke and server deployment.

## 7. Stop conditions

Pause the active stage rather than broadening scope if:

- migration cannot be demonstrated lossless and reversible;
- public behavior regresses;
- a stage cannot end in a runnable state;
- a new request changes the locked data/publication model;
- user acceptance finds a reproducible blocker.
