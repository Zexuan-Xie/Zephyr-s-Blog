# Stage 2 Architecture and Code Review

Status: PASS with notes

Reviewed after repaired Stage 2 acceptance smoke on `7ba0d2921acf22448164d39f2c7c5550aa5f3398`.

## Architecture review

- Backend remains layered: handlers call services; SQL stays in repositories.
- OpenAPI was updated for protected Author Workspace APIs before implementation.
- Frontend Author Workspace centralizes Stage 2 workflow in `AdminPage.tsx` with API helpers in `web/src/lib/api.ts` and shared types in `web/src/lib/types.ts`.
- Stage 2 stays within intended publication model: `草稿` / `已发布`; no Stage 3 Content Version snapshot state exposed.
- Author-only public entry is role-gated and routes to `/admin?target=...`.
- Chinese URL Path handling now works in public browser routes by decoding browser pathname once before API query encoding.

## Code review

- Required gates pass: backend tests/vet/gofmt and frontend tests/lint/build.
- Regression tests cover protected tree adapter and Chinese public path resolver.
- iframe sandbox contract preserved.
- No credentials, local DB, uploads, caches, or build artifacts intentionally committed.

## Notes

- `AdminPage.tsx` is large. It is acceptable for Stage 2 presentation but should be split into smaller components before/while doing Stage 3 autosave/version-history work.
- Browser automation did not fully exercise drag-and-drop persistence; user acceptance should manually sample it.
- Stage 3 MCP should reuse existing API/service boundaries rather than duplicate SQL/business logic.

## Verdict

Stage 2 is ready for user acceptance.
