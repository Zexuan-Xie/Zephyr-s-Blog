# Stage 2 Polish Verification — 2026-06-13

Scope:

- Rename product brand in navigation from `xLab Blog` to `Aeolian`.
- Replace Author Workspace copy with simple English.
- Improve Author Workspace layout and reduce visible controls.
- Add graphical same-parent drag sorting with card-based pointer drag.
- Keep slug/internal IDs hidden from product UI.

Browser smoke:

- `/admin` shows `Aeolian`, `Author Workspace`, `Tree`, and simple English controls.
- Directory workspace shows graphical Create segment (`Directory` / `File`) and `Arrange` card board.
- Same-parent pointer drag from the second card onto the first card sent:
  `PUT /api/admin/nodes/:id/children/order`.
- Page showed `Order saved.` and DB order changed to `test:0`, `drag-check:1` during smoke.
- The temporary empty `drag-check` Directory was deleted after smoke.

Gates:

```bash
cd api
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./...
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go vet ./...
test -z "$(gofmt -l .)"

cd web
node --test tests/*.test.mjs
npm run lint
npm run build
```

Result: PASS.


## Aeolian reference-style visual polish rerun — 2026-06-13

User request: keep the existing layout unchanged, but reference `/mnt/c/Users/Alex-/Downloads/Aeolian (fixed).html` for UI style.

Implemented visual changes:

- Preserved current page structure and Author Workspace layout.
- Switched the visual system from warm ricepaper to cool Aeolian paper: soft blue/grey background, subtle radial light pools, and matte grain.
- Refined frosted-glass material for nav, panels, cards, drawers, inputs, and buttons.
- Added a crystal aqua accent system for active nav, chips, primary buttons, drag/drop states, and brand dot.
- Replaced the nav book icon with a small crystal dot and kept the visible brand as `Aeolian`.
- Self-hosted `Instrument Serif` under `web/src/assets/fonts/` only for the `Aeolian` brand mark; all other UI text uses the original app font stack.
- Updated document title and theme color to `Aeolian` / cool paper tone.

Browser smoke:

- Opened `http://127.0.0.1:5173/admin`.
- Confirmed page title: `Aeolian`.
- Confirmed primary nav still exposes `Aeolian`, `Recent`, `Search`, `Directory`, `Author`.
- Confirmed Author Workspace keeps the same two-column layout and existing controls.
- Screenshot evidence: `docs/verification/stage-2-polish-20260613/visual-admin.png`.

Frontend gates rerun after visual polish:

```bash
cd web
node --test tests/*.test.mjs
npm run lint
npm run build
```

Result: PASS.


## Admin create-form overlap hotfix — 2026-06-13

Issue: on `/admin`, the hidden `kind` input inside the compact Directory create form was still participating in CSS Grid placement. This pushed `Name`, `URL Path preview`, and action buttons into cramped columns, making the two visible inputs appear partially overlapped.

Fix:

- Hide `.admin-form input[type="hidden"]` from layout with `display: none`.
- Make the compact create form use two stable visible columns.
- Force `URL Path preview` and the action row to span the full form width.

Browser smoke:

- Opened `/admin`, scrolled to the Directory create panel.
- Confirmed `Name` is on its own row/column and `URL Path preview` spans a separate full-width row.
- Screenshot: `/tmp/admin-overlap-form-after.png` during local verification.

Gates:

```bash
cd web
node --test tests/*.test.mjs
npm run lint
npm run build
```

Result: PASS.


## Root sibling create + expandable Directory drawer hotfix — 2026-06-13

Issues:

1. After creating the first root Directory, Author Workspace only created children inside the selected Directory and had no obvious way to create another root sibling under `/`.
2. The top `Directory` drawer only listed root-level entries and did not let users expand subdirectories/files in place.

Fix:

- Added `New root` in the Author Workspace tree panel. It opens the same minimal root create form used by the empty-tree state and creates with `parent_id: null`.
- Root and child create now invalidate public/admin tree queries so the navigation drawer updates without a manual refresh.
- Added `fetchDirectoryChildren()` for public child expansion.
- Rebuilt `DirectoryDrawer` as an expandable tree. It shows all top-level entries, lets Directory rows expand/collapse, and links every visible Directory/File to its path.
- When the current user is Author, the drawer uses the protected admin tree so draft Directories and Files are visible; Reader/Anonymous still use the public tree.

Browser smoke:

- `/admin` shows `New root` while an existing root Directory is selected.
- Clicking `New root` opens a root create form with preview `/New directory`.
- Creating a root sibling makes it appear at the same level as the existing root Directory.
- Top `Directory` drawer shows multiple top-level Directories with expand controls.

Gates:

```bash
cd web
node --test tests/*.test.mjs
npm run lint
npm run build
```

Result: PASS.


## Content Tree hierarchy + tree-based drag hotfix — 2026-06-13

User request:

1. Distinguish Directory/File levels in `Content Tree`; parent Directory should be slightly larger than child Directory/File, and level 4+ should use the same size.
2. `Content Tree` should continue to fold/collapse Directory contents.
3. Remove right-side `Drag cards`; drag directly inside `Content Tree`.

Fix:

- Added depth classes `tree-depth-0` through `tree-depth-3`; depth 3 is reused for level 4+.
- Preserved Directory expand/collapse in the left tree.
- Added same-parent drag/drop directly on tree rows. Dragging does not reparent; it only reorders siblings under the same parent Directory.
- Removed the right-side Arrange/Drag cards panel from Directory detail.
- Added tree row drag handle, dragging, and drop-target visual states.

Gates:

```bash
cd web
node --test tests/*.test.mjs
npm run lint
npm run build
```

Result: PASS.

Browser smoke:

- `/admin` renders left Content Tree with hierarchy styling and collapsible Directory rows.
- Right-side Directory detail no longer shows `Drag cards`/Arrange panel.
- Screenshot: `/tmp/admin-tree-drag-polish.png` during local verification.
