# Stage 2 Acceptance

Status: Gateway 6 repaired rerun passed with notes

Verdict: **PASS for Stage 2 engineering acceptance smoke** on `7ba0d2921acf22448164d39f2c7c5550aa5f3398`.

## Summary

The previous Gateway 6 blocker was repaired:

- Backend protected tree returns flat `{ nodes: [...] }` with `url_path`.
- Frontend `fetchAdminTree()` now adapts the backend flat contract into nested Author Workspace roots.
- `/admin` no longer shows `内容树加载失败`.

During repaired browser acceptance, one more Stage 2 regression was found and fixed:

- Public Chinese URL Paths opened by the browser were double-encoded before `/api/tree/resolve`.
- `resolveContentPath()` now decodes the browser pathname once before encoding the API query.
- Regression coverage was added to `web/tests/stage2-author-workspace-contract.test.mjs`.

## Fixture under test

Gateway 1 fixture remains the acceptance baseline:

```text
/stage-2-acceptance
root:           77473f2e-6069-48ff-95a7-3d7173d090c4
draft branch:   f2f3fb74-33f0-4264-baf2-b26e5d06e83e
draft file:     5b796f40-e15a-42fa-8832-9cfbd1dcd21e
published file: a260a9f7-ecf3-4c3d-a87e-cc96b44c73bc
```

Additional smoke nodes were created under `/stage-2-acceptance` during acceptance.

## Required gates

Commands run after the repaired rerun and Chinese-path fix:

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

Result: **PASS**.

Earlier focused gate artifacts remain under `docs/verification/stage-2-browser-20260613/`, including:

- `backend-go-test-92dd65e.txt`
- `backend-go-vet-92dd65e.txt`
- `backend-gofmt-92dd65e.txt`
- `frontend-node-test-92dd65e.txt`
- `frontend-lint-92dd65e.txt`
- `frontend-build-92dd65e.txt`

## API acceptance smoke

Evidence: `docs/verification/stage-2-browser-20260613/stage2-api-smoke-20260613T180451+0800.txt`.

Covered:

- Author login using local smoke Author account.
- Protected `/api/admin/tree` loads flat `nodes` contract and includes `/stage-2-acceptance`.
- Minimal Directory create: only parent/kind/name.
- Minimal File create: parent/kind/name/content format.
- Tree refresh includes created Directory/File.
- Manual content save with keywords.
- Publish makes File public through `/api/tree/resolve`.
- Unpublish hides File from public resolver (`404`).
- Non-empty Directory deletion is blocked with machine-readable reason `non_empty_directory`.

Result: **PASS**.

## Browser acceptance smoke

Evidence directory: `docs/verification/stage-2-browser-20260613/`.

Key evidence:

- `browser-ui-create-dir-20260613T180620+0800.txt/png`
- `browser-ui-create-file-20260613T180639+0800.txt/png`
- `browser-ui-save-publish-20260613T180656+0800.txt/png`
- `browser-public-edit-entry-fixed-20260613T181018+0800.txt/png`
- `browser-public-edit-return-20260613T181035+0800.txt/png`
- `browser-admin-after-api-unpublish-20260613T181314+0800.txt/png`
- `browser-public-after-api-unpublish-20260613T181314+0800.txt/png`
- `browser-mobile-admin-sanity-20260613T181350+0800.txt/png`

Covered:

| Requirement | Result | Evidence / note |
|---|---:|---|
| Author login -> Chinese Author Workspace | PASS | `/admin` loads `内容树` and no longer shows `内容树加载失败`. |
| Create Directory using minimal Chinese form | PASS | New Directory appears immediately in left Content Tree and opens on the right. |
| Create File using minimal Chinese form | PASS | New File appears immediately, is selected, and opens the File workspace. |
| Tree refresh/expand/select/open feedback | PASS | Evidence shows created nodes visible without page refresh and selected workspace changed. |
| Edit File content and manual save | PASS | Browser save/publish evidence plus API smoke. |
| Publish File and public access | PASS | Browser evidence shows published state; API smoke verifies public resolver. |
| Public File `编辑文件` returns to selected workspace File | PASS | `browser-public-edit-entry-fixed-*` and `browser-public-edit-return-*`. |
| `撤回发布` hides public File | PASS via API + UI state evidence | API unpublish hides public resolver; admin UI shows draft state afterward. Browser direct click was flaky in agent-browser, so this is marked for manual acceptance focus. |
| Settings move/URL Path/delete constraints hide implementation IDs | PARTIAL SMOKE | Contract/source tests cover no Parent ID/Node ID/slug in primary UI; API smoke covers delete constraint. Full visual settings move should be manually sampled. |
| Same-parent drag sorting persists and never reparents | NOT FULLY AUTOMATED | Backend/frontend contracts and implementation exist; manual desktop acceptance should sample drag reorder. |
| Public homepage/Recent/public reading/comments/Likes not redesigned | SMOKE PASS | Public File page still shows normal reader controls and only Author edit entry was added. |
| Mobile no-regression sanity | PASS | `browser-mobile-admin-sanity-*` shows mobile opens Author Workspace with orientation/exit controls. |

## Notes for manual user acceptance

Please focus manual acceptance on:

1. clicking `撤回发布` in the File content tab and verifying the button changes back to `发布`;
2. Settings → move preview / URL Path edit / delete blocked messages;
3. same-parent drag reorder on desktop.

These are implemented and covered by source/API tests, but manual UX judgment is still recommended.

## Tested

- Backend tests, vet, gofmt.
- Frontend node tests, lint, build.
- Native local API smoke on PostgreSQL.
- Browser desktop smoke for Author Workspace, create, save/publish, public edit entry, public Chinese URL Path, unpublish visibility.
- Browser mobile no-regression sanity.

## Not tested

- Exhaustive drag-and-drop visual persistence by browser automation.
- Full manual UX acceptance by the user.

## 2026-06-13 clean-slate user acceptance rerun

Context: after the user requested clearing the blog content, the local content tables were empty (`nodes = 0`, `file_contents = 0`) while users remained. A browser acceptance rerun found one clean-slate regression and fixed it.

Fix applied:

- Empty Author Workspace now shows a right-side root creation panel (`从根目录开始` / `创建第一个内容`) instead of only `暂无内容` and `请选择内容树中的目录或文件`.
- Root creation uses the same slugless minimal contract as directory creation: `parent_id: null`, `kind`, `name`, optional `content_format`; no `slug` or `sort_order` in product UI.
- Regression contract added: `Stage 2 empty Author Workspace can create the first root content item`.

Browser rerun covered:

| Requirement | Result | Note |
|---|---:|---|
| Empty local content tree opens Author Workspace without load failure | PASS | `/admin` shows Chinese Author Workspace and empty root create panel. |
| Create first root Directory from an empty Content Tree | PASS | Created `验收目录`; left Content Tree updated immediately and selected/opened it. |
| Create File under the new Directory | PASS | Created `验收文件`; left Content Tree updated immediately and selected/opened File workspace. |
| Manual save API/UI logic | PASS with automation caveat | DOM `requestSubmit()` sent `PUT /api/admin/files/:id/content` and DB stored body/keywords. `agent-browser click` on the form submit button did not dispatch the submit in this run, so the physical click should be manually sampled. |
| Publish | PASS | Browser click sent `POST /publish`, left tree status changed to `已发布`, DB status became `published`. |
| Public Chinese URL Path | PASS | `/验收目录/验收文件` rendered the published File and Chinese path resolved correctly. |
| Public `编辑文件` entry | PASS | Returned to `/admin?target=...` with `验收文件` selected. |
| Unpublish | PASS with automation caveat | DOM click sent `POST /unpublish`, DB status became `draft`, public resolver returned 404. `agent-browser click` on the off-screen/low viewport button was flaky, so the physical click should be manually sampled. |

Additional observations for Stage 2 polish / manual acceptance:

- Login, public 404, comments, and some navigation labels still contain English (`Welcome back`, `Login`, `Path not found`, `Return root`, `Reader discussion`, etc.). This is outside the clean-slate blocker but conflicts with the broader Chinese-copy preference and should be queued if full Chinese polish is required before presentation.
- The old `start-local.sh` nohup Vite process exited during one run; running `conda run --no-capture-output -n blogenv npm run dev -- --host 127.0.0.1` directly kept Vite stable for acceptance. Recheck local launcher reliability if it recurs.

Verification after the clean-slate fix:

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

Result: **PASS**.
