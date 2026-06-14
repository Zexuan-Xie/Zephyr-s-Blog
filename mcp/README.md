# Aeolian Blog MCP Server

Server-local stdio Model Context Protocol server for trusted Author automation.

Gateway 6 exposes a server-local, disabled-by-default tool surface for trusted Author automation.

## Safety contract

- Separate process/package; not mounted in the public web app and not exposed over HTTP/SSE.
- `BLOG_MCP_ENABLED=true` is required before any tool mutates or reads blog state.
- `BLOG_MCP_KILL_SWITCH=true` is checked before every tool call and refuses all operations.
- Every tool call writes a JSONL audit event to `BLOG_MCP_AUDIT_LOG` or `~/.local/share/xlab-blog/mcp-audit.jsonl`.
- `export_backup` writes only under `BLOG_MCP_BACKUP_DIR` or `~/.local/share/xlab-blog/mcp-backups`; callers may pass only an optional relative `label` subdirectory.
- Destructive tools must request/record a backup/export step before mutation when practical.
- MCP handlers must call backend API/service boundaries. Do not import database clients, repositories, or SQL into this package.

## Gateway 6 skeleton smoke

```bash
cd mcp
BLOG_MCP_ENABLED=false node --test tests/*.test.mjs
```

To start the stdio server after dependencies are installed:

```bash
cd mcp
BLOG_MCP_ENABLED=true node src/server.mjs
```

Registered tools: `list_content_tree`, `get_file`, `search_files`, `create_directory`, `create_file`, `update_file_content`, `update_file_settings`, `publish_file`, `unpublish_file`, `move_node`, `reorder_children`, `delete_node`, `upload_asset`, `delete_asset`, `list_assets`, `rebuild_search_index`, `export_backup`, plus `health_check`.

`export_backup` input example:

```json
{ "label": "before-delete" }
```

The resulting file is created below the configured backup root with `wx` exclusive-create semantics. Absolute paths and traversal labels are refused before backend calls or writes.
