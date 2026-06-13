# Aeolian Blog MCP Server

Server-local stdio Model Context Protocol server for trusted Author automation.

Gateway 6 starts with a disabled-by-default skeleton only. Later tasks add the full read/content/publish/tree/assets/maintenance tool set.

## Safety contract

- Separate process/package; not mounted in the public web app and not exposed over HTTP/SSE.
- `BLOG_MCP_ENABLED=true` is required before any tool mutates or reads blog state.
- `BLOG_MCP_KILL_SWITCH=true` is checked before every tool call and refuses all operations.
- Every tool call writes a JSONL audit event to `BLOG_MCP_AUDIT_LOG` or `~/.local/share/xlab-blog/mcp-audit.jsonl`.
- Destructive tools must request/record a backup/export step before mutation when implemented.
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

The first implemented tool is `health_check`, intentionally non-destructive. It proves the registration/audit/guard pattern.
