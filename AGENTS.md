# xLab Blog Agent Guide

## Read order

Before changing code, read:

1. `PROGRESS.md` — current breakpoint, environment, and next actions.
2. `docs/plans/SECOND_DEVELOPMENT.md` — active three-stage plan.
3. `docs/specs/CONTEXT.md` — canonical product language.
4. The relevant active specs under `docs/specs/`.
5. `docs/api/openapi.yaml` before changing shared API behavior.

Historical implementation detail is compacted in `docs/archive/INITIAL_BUILD_SUMMARY.md` and Git history. Do not revive old OMX task state or detached worker changes.

## Product language

Use `Author`, `Reader`, `Anonymous Visitor`, `Content Tree`, `Directory`, `File`, `URL Path`, `Content Version`, and `Published Content` consistently. `Admin` describes privileges/routes, not the person. Do not expose the implementation term `slug` in product UI.

## Current scope

The active work is staged:

1. Reliability, navigation, and identity.
2. Graphical Admin Content Tree and workspace.
3. Autosave, version history, publication snapshots, Draft Preview, and Draft/Published Assets.

Do not redesign public homepage, Recent cards, public Directory/File reading, comments/Likes, or the Glass Ricepaper system except to repair regressions.

## Engineering rules

- Keep each stage runnable, reversible, and independently testable.
- Update OpenAPI first for shared API contract changes.
- Keep SQL in repositories, not HTTP handlers.
- Preserve iframe `sandbox="allow-scripts"` without `allow-same-origin`.
- Preserve full-text search fallback when semantic indexing is unavailable.
- Back up the local database before cleanup or schema migration.
- Do not commit credentials, local database files, uploads, caches, build output, or agent runtime state.
- Update `PROGRESS.md` at every key milestone and before stopping.

## Exact local environment

Use Conda environment `blogenv`:

- Node.js `22.22.3`
- npm `10.9.8`
- Go `1.26.4`
- PostgreSQL `17.10`
- pgvector `0.8.1` locally

Run tools through `conda run -n blogenv ...` when the shell environment is uncertain.

## Required verification

Backend:

```bash
cd api
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./...
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go vet ./...
test -z "$(gofmt -l .)"
```

Frontend:

```bash
cd web
node --test tests/render-safety.test.mjs
npm run lint
npm run build
```

For runtime/auth/tree/publication changes, also run native PostgreSQL API smoke and desktop/mobile browser acceptance. Record evidence under `docs/verification/`.
