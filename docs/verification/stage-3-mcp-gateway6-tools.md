# Stage 3 Gateway 6 MCP Tools Evidence

Status: **PASS (leader implementation slice)**

Task: 17 — Gateway 6 MCP tool implementation slices.
Integrated SHA: pending commit after local verification.

## Implemented tool surface

The server-local `mcp/` package now registers the full required Blog MCP tool surface:

- Read/search: `list_content_tree`, `get_file`, `search_files`
- Content: `create_directory`, `create_file`, `update_file_content`, `update_file_settings`
- Publish: `publish_file`, `unpublish_file`
- Tree: `move_node`, `reorder_children`, `delete_node`
- Assets: `upload_asset`, `delete_asset`, `list_assets`
- Maintenance: `rebuild_search_index`, `export_backup`
- Skeleton/probe: `health_check`

## Architecture and safety notes

- MCP remains a separate server-local stdio package/process under `mcp/`; no HTTP/SSE MCP transport was added.
- Tool handlers are thin orchestration wrappers in `mcp/src/tools.mjs`.
- All blog state access goes through `BlogBackendClient` in `mcp/src/backendClient.mjs`, which calls the existing protected backend HTTP API.
- No MCP handler imports backend database/repository packages.
- `BLOG_MCP_ENABLED` is checked before tool input validation or backend calls.
- `BLOG_MCP_KILL_SWITCH` is checked before every tool call.
- Every tool call writes JSONL audit with tool name, destructive flag, redacted argument summary, and `ok`/`error`/`refused` result.
- Destructive delete/rebuild operations require `confirm=true`; otherwise they refuse before backend calls and audit the refusal as an error.
- `export_backup` writes a local JSON backup of the admin tree plus current file version states to support backup-before-destructive-batch workflows.

## Verification commands

```text
PASS cd mcp && npm install --no-audit --no-fund
PASS cd mcp && npm test
PASS cd mcp && npm run build
PASS git diff --check
PASS grep -R "pgx\|database/sql\|SELECT \|INSERT \|UPDATE \|DELETE \|SQL" -n mcp/src -> no matches
```

## Unit smoke transcript summary

`npm test` now covers:

1. Disabled-by-default refusal and audit JSONL.
2. Per-call kill switch refusal and audit JSONL.
3. Enabled `health_check` success and audit JSONL.
4. Exact registration of all required Stage 3 MCP tool names.
5. Destructive tools refuse while disabled before validation/mutation.
6. Enabled content tool calls the backend boundary with `Authorization: Bearer ...` and audits `ok`.
7. Enabled destructive delete refuses without `confirm=true` before any backend call and audits `error`.

Full black-box MCP stdio acceptance is assigned to Task 18 after this implementation is integrated.
