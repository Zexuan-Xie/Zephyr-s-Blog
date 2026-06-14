# Stage 3 Code Review

Final review date: 2026-06-14
Final reviewed HEAD: `92c345c`

## Initial independent review on HEAD `915d1af`

Two independent lanes reviewed the MCP backup hardening, stdio tests, Stage 3 evidence, and frontend unpublish guard.

### Code-reviewer lane

Recommendation before repair: **COMMENT**

Finding:

- **MEDIUM — MCP mutates backend state before proving the audit event can be written.**
  `runGuardedTool` executed the backend operation before writing the `ok` audit event. If final audit append failed after mutation, the call could mutate without a durable audit record.

Positive findings included stdio-only production MCP transport, no direct SQL/DB access in MCP source, centralized disabled/kill-switch guards, improved backup root/path hardening, frontend unpublish null guard, and passing MCP/frontend gates.

### Architect lane

Architectural status before repair: **WATCH**

Watch items:

- Backup export filesystem policy lived inside `BlogBackendClient`, mixing HTTP API boundary and local backup-service responsibility.
- Backup write flow had a validate-then-open gap.
- `mcp/README.md` startup command referenced `src/server.mjs` while the actual entrypoint is `src/server.ts` with `tsx`/package bin wiring.

## Repairs after review

Commit: `92c345c fix: harden MCP audit and backup boundaries`

Repairs:

- Added audit preflight: enabled tools now write a durable `started` JSONL event before parsing/backend calls. If the audit log cannot be opened/appended, the tool refuses before backend mutation.
- Added regression test using `/dev/full` proving `create_directory` does not call the backend when audit preflight fails.
- Moved backup export filesystem policy from `BlogBackendClient` into dedicated `mcp/src/backup.ts` `BackupExportService`; `BlogBackendClient` is again a thin HTTP API boundary.
- Backup writes now use a temp file plus hard-link-to-final within the canonical output Directory, with cleanup, preserving exclusive creation while reducing validate-then-open exposure.
- Updated `mcp/README.md` startup command to `node --import tsx src/server.ts`.
- Updated stdio smoke audit expectations for `started` + final `ok/error` events.

## Final verdict

- Code-reviewer blockers: **none remaining** after audit preflight repair.
- Architect status: **CLEAR after repair** — the prior WATCH items are resolved by separating `BackupExportService`, tightening write flow, and correcting README startup docs.
- Final recommendation: **APPROVE**.

## Final verification

Evidence: `docs/verification/stage-3-browser-20260614/final-post-review-gates-92c345c.txt`.

```text
PASS cd api && CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./...
PASS cd api && CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go vet ./...
PASS cd api && test -z "$(gofmt -l .)"
PASS cd web && node --test tests/*.test.mjs  # 43/43
PASS cd web && npm run lint
PASS cd web && npm run build
PASS cd mcp && npm test  # 16/16
PASS cd mcp && npm run build
PASS MCP direct SQL/DB grep: no matches in production source
PASS MCP transport grep: production source uses StdioServerTransport only
PASS allow-same-origin grep over web/src and api: no matches
PASS git diff --check
```
