# xLab Blog Progress

Last updated: 2026-06-12 09:49 CST

This is the durable resume point. Keep it concise and update it after every key milestone.

## Current state

- Branch: `main`; local commits are ahead of `origin/main`.
- Initial Packets A–J are complete and natively verified.
- Active plan: `docs/plans/SECOND_DEVELOPMENT.md`.
- Current breakpoint: Stage 1 backend/API error work and the frontend Red contract are integrated. The original Team `execute-approved-xlab-d760bfbb` became non-resumable after all five worker processes exited; its state was archived under `.omx/recovery/`, and its clean detached worktrees/state were removed. A fresh five-seat Team must continue from the current `main` SHA without reviving old worker changes.
- Current integrated commit: `d636c3176d031b0d714ff8fdcd7920ea807b15fe`.
- Completed Stage 1 packets: control checkpoint, backend Red contract, frontend Red contract, precise auth/create API errors, acceptance preparation document, and security preparation document.
- Remaining Stage 1 work: truthful identity/minimal navigation, Directory creation result repair, integrated native/browser acceptance, integrated security review, independent architecture/code review, and coordinator closeout.
- Runtime services: API `:8080` and web `:5173` were not reachable at the 00:02 CST coordinator check; restart with `~/.local/share/xlab-blog/start-local.sh` before integrated browser verification.
- Cleanup checkpoint: `453515d`.
- Approved Ralplan consensus: Architect `APPROVE/CLEAR`; Critic `APPROVE` at 99%.
- Stage 1 Team will use five fixed seats: coordinator, backend, frontend, acceptance, and security.
- Only the coordinator edits `PROGRESS.md` and `docs/verification/stage-1-team-log.md` while Team is active.
- Stage 1 backend error behavior and regression tests have changed; no database schema or acceptance fixture has changed.

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

The acceptance services were offline at the latest coordinator check. Their persistent state remains outside Git; restart them before integrated browser verification.

## Cleanup checkpoint — complete

- Removed old build/runtime caches, obsolete soft-link index, stale generated team instructions, historical packet plans, and monitor logs from the working tree.
- Removed unused frontend dependencies `@types/dompurify` and `prettier`, and unused backend dependency `pgvector-go`.
- Compacted 7,489 removed lines into current entry docs, `docs/archive/INITIAL_BUILD_SUMMARY.md`, and `docs/verification/BASELINE.md`; net working-tree reduction is 7,184 lines.
- Moved the active plan to `docs/plans/SECOND_DEVELOPMENT.md` and synchronized README, agent guide, PRD read order, and technical stack.
- Full Go tests/vet/gofmt, frontend 7/7 tests/lint/build, OpenAPI local refs, Markdown links, package lock, module tidy, and host-network local health checks pass.
- `npm audit` remains externally blocked: the configured mirror does not implement the audit endpoint and the official registry request failed from the current environment.

## Immediate next steps

1. Launch a fresh five-seat recovery Team from `d636c31`; do not reuse old detached worktrees.
2. Implement and integrate truthful identity/minimal navigation, then repair the Directory creation result.
3. Restart native services and run full backend/frontend, PostgreSQL API smoke, desktop/mobile browser acceptance, and security gates.
4. Obtain architect CLEAR, code-reviewer APPROVE, and user acceptance before coordinator closeout.
5. Preserve the current local database until the Stage 2 pre-stage backup/fixture cleanup step.

## Recent milestones

- **2026-06-12 09:49 CST** — recovered from a non-resumable Team: old runtime state archived, all five clean detached worktrees removed, backend packets 2/6 verified complete, and the durable breakpoint moved to integrated SHA `d636c31`.
- **2026-06-12 00:05 CST** — all five root packets were claimed; Team state is 5 in progress, 6 pending, 0 failed.
- **2026-06-12 00:03 CST** — Stage 1 Team launched with five live workers. `dag_sidecar`, effective worker count 5, bootstrap mapping, and all 11 packet subjects/owners/file scopes/dependencies passed exact audit. Initial event cursor and integration ledger were recorded in `docs/verification/stage-1-team-log.md`.
- **2026-06-11 23:59 CST** — Stage 1 preflight baseline passed: Go tests/vet/gofmt and frontend render-safety/lint/build; approved plan applied and Stage 1 bootstrap DAG byte-verified.
- **2026-06-11 23:00 CST** — Ralplan consensus completed after five review iterations: Team launch parsing, DAG JSON, packet command syntax, file ownership, Stage 2 cleanup safety, and Stage 3 restore lifecycle passed; approved plan applied and Stage 1 launch preflight started.
- **2026-06-06 21:13 CST** — clean Stage 1-ready repository baseline committed as `453515d`.
- **2026-06-06 21:12 CST** — repository cleanup completed: stale agent/runtime/doc artifacts removed, unused dependencies removed, active documentation reorganized and compacted, all local quality gates passed, and the native acceptance stack was restored in `xlab-blog-local`.
- **2026-06-06 20:58 CST** — requirements interview reached ~95% shared understanding; three-stage plan and ADR 0007 committed as `075d2f3`.
- **2026-06-06 17:00–20:58 CST** — navigation, identity, graphical Admin, autosave/version/publication, Draft Preview, Asset lifecycle, migration, and stage-gate decisions recorded in active specs.
- **2026-06-06 14:17 CST** — persistent PostgreSQL/uploads recovery state established outside `/tmp`.
- **2026-06-06 14:16 CST** — native full-stack acceptance candidate passed.

Historical team/task detail is available in Git history and summarized in `docs/archive/INITIAL_BUILD_SUMMARY.md`.
