# Stage 2 Acceptance

Status: Gateway 6 integrated acceptance failed

Verdict: **FAIL** for Gateway 6 integrated desktop/mobile acceptance.

Integrated code SHA under test: `59311dd6dda990e5edde75262aa11db7335c2472`
Evidence checkpoint commits in worker-4 worktree include generated verification-only commits on top of that SHA; no feature/source code was changed by acceptance.

## Fixture under test

Gateway 1 fixture remains the acceptance baseline:

```text
/stage-2-acceptance
root:           77473f2e-6069-48ff-95a7-3d7173d090c4
draft branch:   f2f3fb74-33f0-4264-baf2-b26e5d06e83e
draft file:     5b796f40-e15a-42fa-8832-9cfbd1dcd21e
published file: a260a9f7-ecf3-4c3d-a87e-cc96b44c73bc
```

Public fixture smoke using the correct resolver route passed during Gateway 6:

- `GET /api/tree/resolve?path=/stage-2-acceptance/published-fixture` -> HTTP 200.
- `GET /api/tree/resolve?path=/stage-2-acceptance/draft-branch/draft-fixture` -> HTTP 404.

Evidence: `docs/verification/stage-2-browser-20260613/native-contract-check-706c8df.txt`.

## Gate outputs recorded

Backend and frontend static/test gates were rerun on the integrated candidate before browser acceptance:

- Backend `go test ./...`: PASS (`stage-2-browser-20260613/backend-go-test-59311dd.txt`)
- Backend `go vet ./...`: PASS (`stage-2-browser-20260613/backend-go-vet-59311dd.txt`)
- Backend gofmt scan: PASS (`stage-2-browser-20260613/backend-gofmt-59311dd.txt`)
- Frontend `node --test tests/*.test.mjs`: PASS (`stage-2-browser-20260613/frontend-node-test-59311dd.txt`)
- Frontend `npm run lint`: PASS (`stage-2-browser-20260613/frontend-lint-59311dd.txt`)
- Frontend `npm run build`: PASS (`stage-2-browser-20260613/frontend-build-59311dd.txt`)

Runtime services were started from the worker-4 worktree:

```text
API: http://127.0.0.1:8080/api/health -> {"status":"ok","database":"ok"}
Web: http://127.0.0.1:5173/ -> reachable
```

## Failure summary

Desktop Author Workspace acceptance cannot proceed because the protected content tree does not load.

Observed in browser at `/admin` with a valid seeded Author token:

```text
作者工作台
内容树
受保护内容树
内容树加载失败。请刷新或重新登录。
请选择内容树中的目录或文件。
```

Evidence:

- `docs/verification/stage-2-browser-20260613/browser-admin-contract-aa7c0b9.txt`
- `docs/verification/stage-2-browser-20260613/desktop-admin-contract-aa7c0b9.png`

Native contract check shows the root cause:

```text
/api/admin/tree top_level_keys = ['nodes']
first_node_keys = ['id', 'kind', 'name', 'parent_id', 'sort_order', 'status', 'url_path']
frontend_expected_roots_present = False
frontend_expected_path = False
backend_url_path = True
```

The current frontend schema in `web/src/lib/api.ts` expects `roots` and node
`path` fields for `fetchAdminTree()`. The current backend response returns a flat
`nodes` array and `url_path`. Zod parsing rejects the response, so the Author
Workspace tree fails before any create/select/edit/publish path is possible.

## Gateway 6 checklist

| Requirement | Result | Evidence / note |
|---|---:|---|
| Author login -> Chinese Author Workspace | FAIL | Login/identity works, but workspace tree shows `内容树加载失败` after `/api/admin/tree` response parsing fails. |
| Create Directory/File using minimal forms | NOT RUN | Blocked by tree-load failure; no selectable directory in UI. |
| Tree refresh/expand/select/open/toast/path | FAIL | Tree data cannot load. |
| Edit File, manual save, publish | NOT RUN | Blocked by tree-load failure. |
| Public File opens; `编辑文件` returns to selected file | NOT RUN | Blocked before public Author entry round-trip could be verified. |
| Public Directory `管理此目录` returns selected Directory | NOT RUN | Blocked before public Author entry round-trip could be verified. |
| Anonymous/Reader do not see Author actions | NOT RUN | Blocked by integrated desktop hard failure; defer to repair rerun. |
| `撤回发布` hides public File | NOT RUN | Native public resolver/draft isolation works for fixture, but UI workflow blocked. |
| Move/delete/reorder prompts | NOT RUN | Blocked by tree-load failure. |
| Same-parent drag reorder persists | NOT RUN | Blocked by tree-load failure. |
| Mobile no-regression sanity at 390x844 | FAIL (same blocker) | Mobile shell renders, but same `内容树加载失败`; screenshot `mobile-admin-failure-a774949.png`. |
| Public homepage/Recent/reading/comments/Likes not redesigned | NOT ASSESSED | Gateway stopped at Author Workspace blocker. |
| Native API smoke recorded | PASS with contract failure noted | `native-contract-check-706c8df.txt`, `native-api-smoke-bc5e7c7.txt`. |

## Required repair

Repair the protected tree contract drift before rerunning Gateway 6:

1. Either update backend `GET /api/admin/tree` to return the OpenAPI/frontend shape (`roots`, nested `children`, node `path`), or update frontend schemas/adapters to consume the actual backend shape (`nodes`, `url_path`) and build the protected tree.
2. Add/adjust a regression test that would fail on this mismatch (API contract test or frontend mocked `fetchAdminTree` parsing test using the integrated backend payload).
3. Rerun backend/frontend gates, native contract check, desktop acceptance, and mobile sanity.

## Not changed by acceptance

Acceptance did not patch feature code. Evidence-only files were created under
`docs/verification/stage-2-browser-20260613/` and this acceptance report.
