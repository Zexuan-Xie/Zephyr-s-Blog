# Stage 3 Gateway 6 MCP Skeleton Evidence

Status: **PASS for Gateway 6 skeleton slice**

Task: 16 — Gateway 6 MCP research and server skeleton.
Worker: worker-2.

## Source research

Primary references requested for Gateway 6 are the official Model Context
Protocol documentation and TypeScript SDK. The skeleton follows the current
official TypeScript SDK pattern of creating an `McpServer`, registering tools with `registerTool` and Zod-style input schemas, and connecting it to `StdioServerTransport`. Context7 query of the official TypeScript SDK docs showed this import/transport pattern from `github.com/modelcontextprotocol/typescript-sdk` docs. The package is
server-local stdio only; it does not add an HTTP/SSE listener.

## Implemented first slice

Files added under separate package `mcp/`:

- `mcp/package.json` — private package for `xlab-blog-mcp` with the official `@modelcontextprotocol/sdk` and zod dependencies.
- `mcp/src/server.ts` — stdio MCP server entrypoint using `McpServer` and `StdioServerTransport`.
- `mcp/src/config.ts` — explicit `BLOG_MCP_ENABLED` gate and per-call `BLOG_MCP_KILL_SWITCH` config.
- `mcp/src/audit.ts` — JSONL audit writer with argument redaction.
- `mcp/src/backendClient.ts` — backend API-client boundary placeholder; intentionally no DB/repository/SQL imports.
- `mcp/src/tools.ts` — guarded tool registration/handler pattern with audit on ok/error/refused outcomes.
- `mcp/tests/skeleton.test.mjs` — disabled, kill-switch, enabled health-check/audit smoke tests.
- `mcp/README.md` — operation and safety notes.

The only registered Gateway 6 skeleton tool is `health_check`. It is
non-destructive and exists to prove registration, explicit enablement, per-call
kill switch, audit JSONL, argument redaction, and backend-client boundary before
later task 17 adds real blog tools.

## Safety checklist

- Disabled by default: `BLOG_MCP_ENABLED` must be exactly `true`/`1`/`yes`; otherwise every tool call refuses.
- Kill switch: `BLOG_MCP_KILL_SWITCH=true` is checked inside `runGuardedTool` before every operation.
- Audit JSONL: every tool call writes one JSON object line with timestamp, tool, destructive flag, redacted argument summary, result, and optional message.
- No public listener: entrypoint uses `StdioServerTransport` only.
- No direct SQL in MCP package: code review/grep should show no database packages or SQL keywords in `mcp/src` other than explanatory comments.
- Backend boundary: future tools should extend `BlogBackendClient` instead of importing DB/repository/business-rule code directly.
- Backup/export design: destructive tools are not implemented in task 16; task 17 must mark destructive tools and task 18/19 must require backup/export evidence for destructive batches.

## Verification commands

```text
PASS cd mcp && npm install --no-audit --no-fund
PASS cd mcp && npm test
PASS cd mcp && npm run build
PASS grep -R "pgx\|database/sql\|SELECT \|INSERT \|UPDATE \|DELETE " -n mcp/src --exclude=backendClient.ts -> no direct SQL/DB implementation matches
PASS git diff --check
PASS git status/node_modules policy: web/node_modules not tracked; mcp/node_modules remains ignored/untracked; no web/dist/.omx runtime artifacts added
```

## Smoke transcript summary

`npm test` exercises the skeleton without starting a network listener:

1. `BLOG_MCP_ENABLED=false` refuses `health_check` and writes `result:"refused"` audit JSONL.
2. `BLOG_MCP_ENABLED=true BLOG_MCP_KILL_SWITCH=true` refuses `health_check` and writes `result:"refused"` audit JSONL.
3. `BLOG_MCP_ENABLED=true` allows `health_check` and writes `result:"ok"` audit JSONL.

Full black-box stdio transcripts for read/content/publish/tree/assets/maintenance
remain blocked until task 17 implements the real tool slices.

## Dependency repair note

Leader preflight found invalid early dependency choices during initial integration. The package now uses valid current npm metadata observed during repair: `@modelcontextprotocol/sdk` 1.29.0, `@types/node` 25.9.3, `tsx` 4.22.4, and `typescript` 6.0.3. Tests import TypeScript sources through `node --import tsx --test`, and `npm run build` runs `tsc --noEmit`.
