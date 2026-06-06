# Initial Build Summary

The first implementation cycle completed Packets A–J between 2026-06-03 and 2026-06-06. Detailed team plans, recovery logs, and packet-by-packet monitor files were removed from the working tree during the second-development cleanup; Git history remains the archival source.

## Delivered baseline

- Exact Conda toolchain and deployment configuration.
- Go API with PostgreSQL/pgvector migrations.
- Local registration/login, seeded Author account, JWT role guards.
- Unix-like Directory/File Content Tree with Draft/Published lifecycle.
- Atomic path updates and redirects.
- Markdown rendering and isolated full HTML Documents.
- Reader comments, replies, soft deletion, File/Comment Likes.
- Per-File local Assets with validation and immutable public serving.
- Full-text and semantic search with RRF and full-text fallback.
- Initial Admin Tree Manager.
- Dockerfiles, Compose, Caddy SPA fallback, and static deployment checks.

## Important runtime repairs

Native PostgreSQL/browser testing found and repaired:

- PostgreSQL 17 incompatibility with the original generated search-vector expression.
- Incorrect pgx argument counts in comment insert/delete paths.
- Missing `/api/recent` implementation.
- Missing frontend Admin authentication guard.
- Admin panel heading clipping.

## Verified baseline

- Full Go tests, vet, and formatting pass.
- Frontend render-safety tests, lint, and build pass.
- Fresh PostgreSQL migration and 21-step API business smoke pass.
- Desktop and mobile browser acceptance pass.
- Docker live smoke remains pending Docker availability in WSL.

## Superseding plan

The active work is no longer the packet plan. Use [`../plans/SECOND_DEVELOPMENT.md`](../plans/SECOND_DEVELOPMENT.md).
