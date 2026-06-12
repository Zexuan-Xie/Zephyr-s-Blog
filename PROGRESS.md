# xLab Blog Progress

Last updated: 2026-06-12 23:42 CST

This is the durable resume point. Keep it concise and update it after every key milestone.

## Current state

- Branch: `main`; local commits are ahead of `origin/main`.
- Initial Packets A–J are complete and natively verified.
- Active plan: `docs/plans/SECOND_DEVELOPMENT.md`.
- Current breakpoint: Stage 1 engineering closeout is complete at product SHA `f0877d0` with Architect `CLEAR` and Code Review `APPROVE`; Stage 1 is ready for explicit user acceptance and Stage 2 must not start until that acceptance is given.
- Current integrated product commit: `f0877d09608a8b58a38f51f5a62cd02ec8cdcd81`.
- Completed Stage 1 packets: control checkpoint, backend Red contract, frontend Red contract, precise auth/create API errors, acceptance/security preparation, truthful identity/minimal navigation, Directory creation result repair, security repairs, closeout identity repair, integrated acceptance/security, and independent review.
- Remaining Stage 1 work: explicit user acceptance only.
- Runtime services: API `:8080` and web `:5173` are healthy as of 23:42 CST.
- Recovery Team terminal task snapshot: 16 total; 16 completed, 0 in progress, 0 pending, 0 failed; all five tmux workers remain dead and the leader completed closeout via the durable ledger.
- Worker adaptation: all five tmux workers remain dead after quota exhaustion. Native independent Architect/code-reviewer agents completed review; the durable Team ledger remains the source of truth.
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

The native stack is currently running from this session for user acceptance. If it is offline later, restart it with `~/.local/share/xlab-blog/start-local.sh`.

## Cleanup checkpoint — complete

- Removed old build/runtime caches, obsolete soft-link index, stale generated team instructions, historical packet plans, and monitor logs from the working tree.
- Removed unused frontend dependencies `@types/dompurify` and `prettier`, and unused backend dependency `pgvector-go`.
- Compacted 7,489 removed lines into current entry docs, `docs/archive/INITIAL_BUILD_SUMMARY.md`, and `docs/verification/BASELINE.md`; net working-tree reduction is 7,184 lines.
- Moved the active plan to `docs/plans/SECOND_DEVELOPMENT.md` and synchronized README, agent guide, PRD read order, and technical stack.
- Full Go tests/vet/gofmt, frontend 7/7 tests/lint/build, OpenAPI local refs, Markdown links, package lock, module tidy, and host-network local health checks pass.
- `npm audit` remains externally blocked: the configured mirror does not implement the audit endpoint and the official registry request failed from the current environment.

## Immediate next steps

1. Ask the user to perform Stage 1 acceptance at `http://127.0.0.1:5173` using the handoff checklist in the assistant response.
2. Do not start Stage 2 until the user explicitly accepts Stage 1.
3. Preserve the current local database until the Stage 2 pre-stage backup/fixture cleanup step.
4. If services stop, recover with `~/.local/share/xlab-blog/start-local.sh`.

## Recent milestones

- **2026-06-12 23:42 CST** — Stage 1 independent review and coordinator closeout completed; terminal Team snapshot 16/16 completed. Initial Architect/Code Review found identity closeout blockers; `f0877d0` repaired role-aware Author login, Reader logout staying on public pages, Author logout from Admin to `/`, and status-specific auth UI errors. Full backend gates, 19/19 frontend tests, lint, build, 12-step native identity/API smoke, and desktop/mobile browser identity closeout passed. Evidence commit `a76dd89` records `docs/verification/stage-1-code-review.md`, closeout browser evidence, and screenshots. Architect `CLEAR`; Code Review `APPROVE`. Stage 1 now waits only for explicit user acceptance.

- **2026-06-12 18:26 CST** — integrated acceptance/security retests completed with `PASS` against product SHA `d6c7d09`; evidence commit `ee6a5c2` records the 21-step PostgreSQL/API smoke, full backend/frontend gates, desktop/mobile role and creation scenarios, security blocker retests, explicit-ID cleanup, and redacted artifacts. Team tasks are 14/16 complete; independent architecture/code review is active.
- **2026-06-12 17:54 CST** — security fixes integrated: `673650e` hardens Login return targets (17/17 frontend tests, lint, build PASS); `d6c7d09` makes configured Author credentials authoritative and sanitizes unexpected auth errors (full Go tests/vet/gofmt and PostgreSQL rollback regression PASS). Independent security and acceptance retests started.
- **2026-06-12 17:45 CST** — independent security report `1877c25` recorded `FAIL`: existing Reader elevation retained the old password, unexpected auth errors leaked internal messages, and backslash-form return targets broke post-login navigation. Backend task `14`, frontend task `15`, and security retest task `16` were added; native agents Dewey/Planck are active because all tmux workers are dead.
- **2026-06-12 10:38 CST** — integrated browser acceptance exposed incorrect duplicate-URL-Path messaging because parent wording was matched before HTTP 409. Fix task `12` used Red→Green regression coverage and completed at `58df9f6`; full frontend tests 15/15, lint, and build pass. Retest task `13` is active.
- **2026-06-12 10:24 CST** — task `7` completed at integrated SHA `8b34388`; leader verification passed 14/14 frontend tests, lint, build, and `git show --check`. Tasks `8`/`9` are in progress. Because their tmux owners reached usage limits, native independent acceptance/security agents were assigned the unchanged verification contracts. API and web services are healthy.
- **2026-06-12 09:57 CST** — recovery DAG reconciled: tasks `6`–`11` now encode the remaining frontend, acceptance, security, independent review, and closeout dependency chain. Backend bootstrap task `2` completed; four roots remain active. Services remain offline and no final gate is claimed.
- **2026-06-12 09:56 CST** — recovery Team `execute-approved-xlab-31e581d5` launched with five live/reporting workers and all five coarse seat tasks claimed. Coordinator detected missing packet dependencies and notified the leader.
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
