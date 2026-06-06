# xLab Blog Progress

Last updated: 2026-06-06 21:12 CST

This is the durable resume point. Keep it concise and update it after every key milestone.

## Current state

- Branch: `main`; local commits are ahead of `origin/main`.
- Initial Packets A–J are complete and natively verified.
- Active plan: `docs/plans/SECOND_DEVELOPMENT.md`.
- Current breakpoint: clean, verified Stage 1-ready baseline; Stage 1 implementation has not started.
- Cleanup checkpoint: `453515d`.
- No active OMX team. Do not revive old team state or detached worker commits.
- Product code and acceptance data have not yet been changed for Stage 1.

## Locked delivery stages

1. **Reliability/navigation/identity** — false create failure, actionable errors, single search input, identity-aware navigation, Reader/Author access behavior.
2. **Graphical Admin workspace** — complete protected Content Tree, graphical creation, ordering, moves, node settings, Content/Assets/Settings workspace.
3. **Autosave/publication model** — Current/Previous Content Versions, independent Published Content, Draft Preview, optimistic concurrency, Draft/Published Assets.

Each stage must finish runnable, reversible, fully tested, documented, and user-accepted before the next begins.

## Scope lock

Do not proactively redesign public homepage, Recent cards, public Directory/File reading, comments/Likes, or Glass Ricepaper. Only repair regressions in those areas.

## Verified baseline

- Go full tests/vet/gofmt: pass.
- Frontend static tests 7/7, lint, build: pass.
- Fresh PostgreSQL migration and 21-step API smoke: pass.
- Desktop/mobile browser acceptance: pass.
- Evidence: `docs/verification/BASELINE.md` and `docs/verification/native-local-full-stack-smoke-20260606.md`.
- Remaining external boundary: real DashScope embeddings and live Docker Compose.

## Local environment and recovery

Conda environment `blogenv`:

- Node.js `22.22.3`
- npm `10.9.8`
- Go `1.26.4`
- PostgreSQL `17.10`
- pgvector `0.8.1`

Persistent local state is outside Git at `~/.local/share/xlab-blog`.

Recover services:

```bash
~/.local/share/xlab-blog/start-local.sh
curl -fsS http://127.0.0.1:8080/api/health
curl -fsS http://127.0.0.1:5173/ >/dev/null
```

Current acceptance services are running in tmux session `xlab-blog-local`.

## Cleanup checkpoint — complete

- Removed old build/runtime caches, obsolete soft-link index, stale generated team instructions, historical packet plans, and monitor logs from the working tree.
- Removed unused frontend dependencies `@types/dompurify` and `prettier`, and unused backend dependency `pgvector-go`.
- Compacted 7,489 removed lines into current entry docs, `docs/archive/INITIAL_BUILD_SUMMARY.md`, and `docs/verification/BASELINE.md`; net working-tree reduction is 7,184 lines.
- Moved the active plan to `docs/plans/SECOND_DEVELOPMENT.md` and synchronized README, agent guide, PRD read order, and technical stack.
- Full Go tests/vet/gofmt, frontend 7/7 tests/lint/build, OpenAPI local refs, Markdown links, package lock, module tidy, and host-network local health checks pass.
- `npm audit` remains externally blocked: the configured mirror does not implement the audit endpoint and the official registry request failed from the current environment.

## Immediate next steps

1. Begin Stage 1 with failing regression tests for the successful-create/false-error bug.
2. Preserve the current local database until the Stage 2 pre-stage backup/fixture cleanup step.

## Recent milestones

- **2026-06-06 21:13 CST** — clean Stage 1-ready repository baseline committed as `453515d`.
- **2026-06-06 21:12 CST** — repository cleanup completed: stale agent/runtime/doc artifacts removed, unused dependencies removed, active documentation reorganized and compacted, all local quality gates passed, and the native acceptance stack was restored in `xlab-blog-local`.
- **2026-06-06 20:58 CST** — requirements interview reached ~95% shared understanding; three-stage plan and ADR 0007 committed as `075d2f3`.
- **2026-06-06 17:00–20:58 CST** — navigation, identity, graphical Admin, autosave/version/publication, Draft Preview, Asset lifecycle, migration, and stage-gate decisions recorded in active specs.
- **2026-06-06 14:17 CST** — persistent PostgreSQL/uploads recovery state established outside `/tmp`.
- **2026-06-06 14:16 CST** — native full-stack acceptance candidate passed.

Historical team/task detail is available in Git history and summarized in `docs/archive/INITIAL_BUILD_SUMMARY.md`.
