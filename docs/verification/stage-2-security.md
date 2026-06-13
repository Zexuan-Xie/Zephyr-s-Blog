# Stage 2 Security and Abuse Review

Status: PASS with follow-up notes

Reviewed on repaired Stage 2 code ending at `7ba0d2921acf22448164d39f2c7c5550aa5f3398`.

## Checks

- Protected `/api/admin/*` routes are behind `RequireAdmin`.
- Reader/Anonymous cannot access Author Workspace APIs without an admin token.
- Public tree/resolve/recent/search continue to use published content paths; draft fixture resolver returns 404 in recorded fixture smoke.
- File HTML rendering preserves iframe sandbox contract: `sandbox="allow-scripts"` without `allow-same-origin`.
- Full-text search fallback is preserved by tests and search service behavior.
- Directory deletion blocks non-empty Directories with `non_empty_directory` detail.
- Published File deletion remains blocked until unpublish.
- Move/reorder APIs are in service/repository layers, not ad hoc SQL in handlers.

## Follow-up notes

- JWT remains in localStorage; acceptable for local personal blog scope but increases XSS blast radius.
- SVG validation is blacklist-oriented; keep strict SVG review in Stage 3 assets work.
- Stage 3 MCP must include explicit enablement, audit logs, backup/export before destructive batches, and kill switch.

## Verdict

No Stage 2 blocking security issue found in the repaired scope.
