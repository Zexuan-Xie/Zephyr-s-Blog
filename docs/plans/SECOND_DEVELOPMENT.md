# xLab Blog Second Development — Team Execution Plan

Date: 2026-06-07

Status: RALPLAN revision 5 for Architect → Critic consensus. This is the proposed replacement for `docs/plans/SECOND_DEVELOPMENT.md`, formerly `DEBUGGING_IMPLEMENTATION_PLAN.md`. It preserves product requirements and adds execution governance.

## 1. Outcome and locked scope

The second development remains three independently runnable, reversible, testable stages:

1. Reliability, navigation, and identity.
2. Graphical Admin Content Tree and workspace.
3. Autosave, version history, publication snapshots, Draft Preview, and Draft/Published Assets.

The active product contract remains in `docs/specs/CONTEXT.md`, `docs/specs/PRD.md`, `docs/specs/BLOG_FLOW.md`, `docs/specs/DESIGN.md`, `docs/specs/BACKEND_STRUCTURE.md`, and `docs/api/openapi.yaml`.

Use `Author`, `Reader`, `Anonymous Visitor`, `Content Tree`, `Directory`, `File`, `URL Path`, `Content Version`, and `Published Content`. Never expose `slug` in product UI.

Scope exclusions remain locked:

- no public homepage, Recent card, public Directory/File reading, comments/Likes, or Glass Ricepaper redesign except regression repair;
- no cross-Directory drag-and-drop reparenting;
- no Admin tree search;
- no more than Current/Previous Author history;
- no container/server deployment before all native stages and user acceptance pass.

### Drag interaction contract

The PRD's older broad “no drag” sentence means **no drag reparenting**. The controlling detailed contract is:

- desktop may drag siblings only to change the mixed same-parent order;
- mobile uses explicit Move up / Move down;
- drag never changes parent;
- cross-Directory movement uses the graphical Directory Picker in Advanced settings.

Stage 2 must first make `docs/specs/PRD.md` wording match `BLOG_FLOW.md` and `DESIGN.md` before implementation, without changing this locked behavior.

## 2. RALPLAN-DR decision summary

### Principles

1. **Small verified packets:** each packet has Red or explicit baseline evidence, a minimal Green change, refactor only under tests, exact commands, and one reversible integration checkpoint.
2. **Evidence before claims:** no success without observed output, source/integration SHAs, and `Tested` / `Not tested` declarations.
3. **Independent gates:** development Agents cannot approve their own acceptance, security, architecture, or code review.
4. **Runnable boundaries:** every stage ends usable, reversible, documented, user-accepted, and safe to resume after power loss.
5. **No speculative complexity:** apply YAGNI/DRY; reject placeholders, swallowed errors, hidden fallbacks, fake success, and untested alternate paths.

### Top decision drivers

1. Contain auth, destructive fixture cleanup, migration, publication, and Draft/public security risk.
2. Preserve the locked product model while enabling parallel implementation.
3. Make interruption recovery unambiguous under OMX's automatic detached worker worktrees.

### Option A — one long-lived Team

Pros:
- lower startup overhead;
- more worker context continuity.

Cons:
- stale state across stage boundaries;
- harder rollback and user-acceptance isolation;
- larger merge/conflict and scope-creep surface.

### Option B — fresh Team per stage with fixed functional seats

Pros:
- clean stage boundaries and independent acceptance/security signoff;
- easier rollback, shutdown, and recovery;
- task ownership can match each stage.

Cons:
- repeated launch/ACK/integration/shutdown overhead;
- cross-stage memory must be maintained deliberately.

### Decision

Choose **Option B**. Launch a fresh five-worker Team for each stage. Cross-stage memory lives in integrated Git commits, concise `PROGRESS.md`, and detailed `docs/verification/` evidence—not in stale Team state.

## 3. Superpowers-derived execution doctrine

Adopt the useful design philosophy of the historical Superpowers plan without reviving its runtime artifacts:

- Red → Green → Refactor → Verify for every behavior change.
- Vertical packets small enough to review, integrate, and revert independently.
- Exact files, commands, expected failure, expected passing result, and stop condition.
- Stop on version or contract mismatch; never silently substitute behavior, dependency, API, or data model.
- Commit only after fresh verification.
- Every packet and stage reports `Tested:` and `Not tested:`.
- Prefer the smallest solution satisfying the locked requirement; remove duplication after behavior is protected.
- No placeholders, dead routes, fake success, swallowed errors, or masking fallback branches.
- Independent review follows integration and tests; it never replaces them.
- Every stage closes with a requirement-to-evidence self-review.

## 4. Available Agent roster and fixed Team seats

### 4.1 Installed native roles and runtime identities

Confirmed callable role files under `/home/zephry_xzx/.codex/agents/*.toml`:

- planning/review: `analyst`, `planner`, `architect`, `critic`, `scholastic`, `prometheus-strict-metis`, `prometheus-strict-momus`, `prometheus-strict-oracle`;
- implementation: `executor`, `team-executor`, `debugger`, `dependency-expert`, `designer`, `code-simplifier`;
- assurance: `test-engineer`, `verifier`, `code-reviewer`, `git-master`;
- research/docs: `researcher`, `writer`;
- inspection: `explore`, `vision`.

`worker-1` … `worker-5` are Team runtime identities. `worker`, `explorer`, and `default` are native-tool aliases/surfaces, not installed OMX role files and not Team DAG roles. Installed `explore` currently selects an unavailable Spark model, so use normal repository inspection or another installed role. There is no dedicated security role; the security lane uses `code-reviewer` with a threat-review posture.

### 4.2 Five functional seats per stage

`omx team 5` preserves DAG roles. Never launch this handoff as `5:executor`, because an explicit role overrides every DAG role. The following are **functional task postures assigned through task/inbox state**, not per-worker launch-time agent types:

| Worker | Functional seat | Owned responsibility | Reasoning guidance |
|---|---|---|---|
| 1 | Coordinator / monitor Agent | Team/task/mailbox reconciliation, integration ledger, `PROGRESS.md`, breakpoint, team log | high |
| 2 | Backend development Agent | OpenAPI, Go handlers/services/repositories, migrations, backend tests | high; xhigh for migration concurrency |
| 3 | Frontend development Agent | React state/pages/components/styles, frontend tests | high |
| 4 | Acceptance/results Agent | Red acceptance specification, integrated black-box/API/database/browser verification | high |
| 5 | Security detection Agent | independent threat review, abuse tests, security report; no feature code | xhigh for Stages 2–3 |

Additional native `architect` and `code-reviewer` signoff is invoked by the leader after integration; neither signoff context may have authored the reviewed code.

## 5. Automatic detached worktrees and integration discipline

OMX Team workers use dedicated detached worktrees automatically. Therefore:

- workers edit and commit only inside their worktrees;
- the **leader**, not a worker, is the sole integrator into the leader branch;
- Worker 1 owns the progress documents in its worktree and supplies integration-ready progress commits;
- acceptance/security test only a leader-integrated SHA, never an unintegrated developer worktree.

### 5.1 Per-packet integration flow

1. Worker makes one integration-ready commit and reports:
   - task ID;
   - `source_sha` and parent SHA;
   - changed files;
   - verification output summary;
   - evidence path;
   - `Tested` / `Not tested`.
2. Worker 1 records the proposed source SHA in `docs/verification/stage-<n>-team-log.md` and reports its own progress-document source SHA.
3. Leader checks ownership and dependency order, then integrates with:

```bash
git cherry-pick -x <source_sha>
```

4. Leader records the resulting `integration_sha` in the Team log and sends it to Worker 4/5.
5. Before acceptance/security work, Worker 4/5 must have a clean worktree and reset its detached worktree to the integrated commit:

```bash
git status --short
git checkout --detach <integration_sha>
```

If the worktree is dirty, the worker commits its evidence first or stops and reports the conflict; it must not discard work.

6. Run verification from the integrated SHA. Cached success from a developer worktree does not count.
7. If integration verification fails, leader reverts the integration commit; the gate completes with `verdict=FAIL`, new fix/retest tasks are created, and Worker 1 records the exact rollback breakpoint.
8. `PROGRESS.md` advances only after the integrated SHA and fresh verification exist.

### 5.2 Integration ledger

Worker 1 maintains this table in `stage-<n>-team-log.md`:

| Task | Owner | Source SHA | Integration SHA | Verification reset | Evidence | Status |
|---|---|---|---|---|---|---|

The stage cannot enter acceptance/security closeout while any required source SHA lacks an integrated SHA or fresh verification result.

## 6. File ownership and conflict prevention

- Worker 1 alone edits `PROGRESS.md` and `docs/verification/stage-<n>-team-log.md` while Team is active.
- Worker 4 owns `docs/verification/stage-<n>-acceptance.md`.
- Worker 5 owns `docs/verification/stage-<n>-security.md`.
- Leader or an independent reviewer owns `docs/verification/stage-<n>-code-review.md`.
- Worker 2 owns `docs/api/openapi.yaml`, `api/**`, and migrations unless the leader assigns a disjoint backend file.
- Worker 3 owns `web/**`.
- Active spec corrections get one named owner before editing.
- No concurrent same-file editing. No worker reverts another worker's work.
- Evidence workers do not implement feature code. A failed gate creates a new development task for Worker 2/3.

## 7. Coordinator / real-time monitor contract

Worker 1 has a persistent control task from startup until closeout and closes last.

It must:

1. ACK the stage goal, owned files, monitoring cadence, and stop condition.
2. Run `omx team status <team-name> --json` after startup, each lifecycle nudge/transition, and at least every 30 minutes during active work.
3. Track a durable event cursor:

```bash
omx team await <team-name> --timeout-ms 30000 --after-event-id <last_event_id> --json
```

or:

```bash
omx team api await-event --input '{"team_name":"<team-name>","after_event_id":"<last_event_id>","timeout_ms":30000,"wakeable_only":true}' --json
```

Store the returned last processed event ID in the stage Team log and use it on the next wait.

4. Reconcile task JSON, worker status, mailbox, integration ledger, and evidence. Idle panes are not completion proof.
5. Keep `PROGRESS.md` concise; detailed commands live in verification docs.
6. Update `PROGRESS.md` after:
   - Team launch and all ACKs;
   - Red reproduction;
   - backup/restore or migration rehearsal;
   - each verified integration checkpoint;
   - blocker/rejection/reassignment;
   - acceptance/security/code-review verdict change;
   - user acceptance;
   - before stopping or shutdown.
7. Each update states current integrated commit, services, task counts/IDs, blocker, exact next command, rollback point, and evidence links.
8. Immediately notify the leader about ownership overlap, stale assumptions, merge conflict, failed verification, or acceptance/security rejection.

## 8. Approved repo-aware Team DAG bootstrap

### 8.1 Artifacts, activation, and launch

The approved execution handoff consists of:

- `.omx/plans/prd-second-development-active.md` and matching `.omx/plans/test-spec-second-development-active.md`;
- immutable five-seat bootstrap templates `.omx/plans/stages/stage-{1,2,3}-team-dag.json`;
- immutable detailed packet graphs `.omx/plans/stages/stage-{1,2,3}-packet-dag.json`;
- active import `.omx/plans/team-dag-second-development-active.json`.

Before each stage, activate exactly one approved template and prove byte identity:

```bash
export STAGE=<1|2|3>
rm -f .omx/plans/team-dag-second-development-active-*.json
cp ".omx/plans/stages/stage-$STAGE-team-dag.json" .omx/plans/team-dag-second-development-active.json
cmp ".omx/plans/stages/stage-$STAGE-team-dag.json" .omx/plans/team-dag-second-development-active.json
```

Approved launch hints (use the one matching the active template):

Launch via omx team 5 "Execute approved xLab Blog second-development Stage 1 DAG"

Launch via omx team 5 "Execute approved xLab Blog second-development Stage 2 DAG"

Launch via omx team 5 "Execute approved xLab Blog second-development Stage 3 DAG"

A normal unrelated Team command must not import this DAG.

### 8.2 Deterministic functional lanes and graph audit

DAG JSON cannot set an owner directly. The bootstrap has exactly five independent, deliberately low-overlap nodes, in order: `writer` coordination, `executor` backend, `designer` frontend, `test-engineer` acceptance, `code-reviewer` security. Repository allocation simulation must map them to workers 1–5 respectively. The detailed packet graph is not imported automatically: after this mapping is verified, the coordinator creates every packet with explicit `owner` from `owner_seat` and concrete `blocked_by` IDs, then reads every task back before dispatch.

Immediately after launch and before any edit:

```bash
omx team status <team-name> --json
omx team api list-tasks --input '{"team_name":"<team-name>"}' --json
cat ".omx/state/team/<team-name>/decomposition-report.json"
omx team api read-task --input '{"team_name":"<team-name>","task_id":"<each-concrete-id>"}' --json
```

First compare the active bootstrap against `decomposition_source=dag_sidecar`, artifact path, effective worker count `5`, `node_id_to_task_id`, and the exact 1→5 role mapping. Then traverse the approved stage packet graph topologically, call `create-task` with the mapped explicit owner and already-created concrete dependencies, record symbolic→concrete IDs, and run `list-tasks` plus `read-task` for every packet. For Stage 2, the coordinator first completes `s2-prd-drag-preflight` and records its PASS only in `docs/verification/stage-2-team-log.md`; Worker 4 then mirrors the checkpoint into `docs/verification/stage-2-acceptance.md` while preparing acceptance, and only then does the coordinator claim the persistent control packet; `s2-backend-red` and `s2-frontend-red` remain blocked by that preflight, so every implementation descendant is transitively blocked. Compare every subject, owner, file scope, and `blocked_by`; save the full transcript in the stage Team log. No worker receives an implementation packet before the graph audit and, for Stage 2, the PRD preflight both pass.

**Stop and shut down before edits** if import falls back to legacy text, the stage/count differs, a lane collides, any bootstrap/packet dependency differs, packet ownership differs, or an unexpected task exists. Correct and approve the artifact, then launch a fresh Team; do not patch ownership ad hoc.

### 8.3 Claims, gate rejection, and failure

Workers ACK/read back the imported concrete task before claiming. Normal terminal result is `completed` with `verdict=PASS`, source/integration SHA, evidence, `Tested`, and `Not tested`.

An expected acceptance/security/architecture/code-review rejection also transitions to `completed`, but with `verdict=FAIL` and exact findings. The coordinator then uses `create-task` to add a development fix task and a gate retest blocked by that fix. Terminal tasks are never reopened. Retest runs on a new integrated SHA and all dependent reviews repeat.

`release-task-claim` is only for returning non-terminal work to pending. Runtime `failed` is reserved for unrecoverable execution failure. Because OMX has no failed-task reopen API, any terminal failed task requires breakpoint capture and a fresh Team from the last verified SHA. Stage completion still requires the final Team to report `failed=0`.

Stage 2 closeout is not complete until Worker 1 has ACKed the PRD preflight PASS and the same checkpoint appears in both the stage Team log and the acceptance note.

```bash
omx team api create-task --input '{"team_name":"<team-name>","subject":"Fix <gate> <finding>","description":"Red=<command>; Green=<command>; fix only finding against <sha>","owner":"<development-worker>","blocked_by":[],"requires_code_change":true}' --json
omx team api create-task --input '{"team_name":"<team-name>","subject":"Retest <gate> <finding>","description":"Reset to new integrated SHA; complete verdict PASS or FAIL","owner":"<gate-worker>","blocked_by":["<fix-task-id>"],"requires_code_change":false}' --json
```

## 9. Team lifecycle and mandatory reviews

### 9.1 Preflight

Before launch:

1. Read `PROGRESS.md`, this plan, `CONTEXT.md`, relevant specs, and OpenAPI.
2. Require a clean leader worktree; automatic Team worktrees reject dirty leader state.
3. Confirm exact `blogenv` versions.
4. Run baseline verification or record a reproducible pre-existing failure.
5. Back up before Stage 2 cleanup and Stage 3 migration.
6. Define disjoint tasks and evidence paths.

### 9.2 Active control

```bash
omx team status <team-name> --json
omx team await <team-name> --timeout-ms 30000 --after-event-id <last_event_id> --json
omx team resume <team-name>
```

A stale/all-idle/worker-stop/merge-conflict nudge triggers status and state reconciliation before manual intervention.

### 9.3 Independent code-review gate

After development commits are integrated and acceptance/security reports exist, the leader invokes fresh native reviews against the integrated SHA:

- `architect` must return **CLEAR**;
- `code-reviewer` must return **APPROVE**.

Record both prompts/context, reviewed SHA, findings/resolutions, verdicts, and residual external boundaries in:

- `docs/verification/stage-<n>-code-review.md`.

A known product/security defect cannot be waived as `Not tested`. If either verdict is not CLEAR/APPROVE, record `verdict=FAIL`, create new fix/retest tasks under Section 8.3, integrate, rerun acceptance/security, and request fresh reviews.

### 9.4 Completion and shutdown

A stage is complete only when:

- all required changes are integrated and fresh tests pass;
- acceptance Agent says PASS;
- security Agent says PASS;
- architect says CLEAR;
- code-reviewer says APPROVE;
- Team log, acceptance, security, code-review, backup/migration evidence exist;
- `PROGRESS.md` names the rollback point and exact next step;
- user accepts the runnable stage;
- `pending=0`, `in_progress=0`, `failed=0`.

Coordinator closes last. Then:

```bash
omx team status <team-name> --json
omx team shutdown <team-name>
```

Verify panes/state are gone. Never revive old Team state for the next stage.

## 10. Skill-routing policy

| Skill | Invoke when | Do not invoke when |
|---|---|---|
| `tdd` | behavior changes need a failing test and Red/Green loop | docs-only closeout or an existing precise failing test already provides Red |
| `diagnose` | defect crosses UI/API/DB, is intermittent, or two materially different fixes failed | narrow proven root cause |
| `code-review` | mandatory after each integrated stage; auth, migration, publication, assets, sandbox, redirects | before a reviewable integrated diff or instead of tests |
| `improve-codebase-architecture` | a demonstrated boundary violation blocks Stage 2/3 | speculative refactor or unrelated cleanup |
| `setup-matt-pocock-skills` | only after explicit user approval to configure issue tracker, triage labels, and domain-doc mapping | silently during implementation; `docs/agents/` is absent |
| Matt `tdd` / `diagnose` | bounded engineering packets once repo context is sufficient | issue/triage workflows before setup approval |
| `designer` / design skill | verify Stage 2 interactions against existing Design spec | change Glass Ricepaper or locked UX direction |
| `dependency-expert` | exact external API/version/migration behavior is uncertain | generic advice replacing repository evidence |
| `code-simplifier` / AI slop cleaner | after review on files changed in a packet, with behavior locked | broad rewrite or before tests |

If Matt setup is approved later, the setup interview should propose GitHub Issues from the existing remote, canonical triage labels, and single-context domain docs at `docs/specs/CONTEXT.md` plus `docs/adr/`; it must still ask the user one decision at a time before writing.

## 11. Required baseline commands

Backend:

```bash
conda run -n blogenv bash -lc '
  cd api &&
  CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./... &&
  CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go vet ./... &&
  test -z "$(gofmt -l .)"
'
```

Frontend:

```bash
conda run -n blogenv bash -lc '
  cd web &&
  node --test tests/render-safety.test.mjs &&
  npm run lint &&
  npm run build
'
```

Runtime/auth/tree/migration/publication/asset work additionally requires native PostgreSQL API smoke and desktop/mobile browser acceptance.

Every evidence document contains timestamp, environment, command, exit status, observed result, integrated SHA, `Tested`, and `Not tested`.

## 12. Database and upload safety commands

Backups are local runtime artifacts outside Git. Only their path, checksum, restore result, and sanitized inventory enter `docs/verification/`.

### 12.1 Backup and checksum

```bash
export XLAB_STATE="$HOME/.local/share/xlab-blog"
export DB_HOST="127.0.0.1"
export DB_PORT="55432"
export DB_USER="zephry_xzx"
export DB_NAME="xlab_blog"
export STAGE="stage-<n>"
export TS="$(date +%Y%m%d-%H%M%S)"
export BACKUP_DIR="$XLAB_STATE/backups/$STAGE-$TS"
mkdir -p "$BACKUP_DIR"

conda run -n blogenv pg_dump \
  -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
  --format=custom --file "$BACKUP_DIR/database.dump"
sha256sum "$BACKUP_DIR/database.dump" | tee "$BACKUP_DIR/database.dump.sha256"

tar -C "$XLAB_STATE" -czf "$BACKUP_DIR/uploads.tar.gz" uploads
sha256sum "$BACKUP_DIR/uploads.tar.gz" | tee "$BACKUP_DIR/uploads.tar.gz.sha256"
```

### 12.2 Disposable restore rehearsal

```bash
export RESTORE_DB="xlab_blog_restore_${STAGE//-/_}_$TS"
conda run -n blogenv createdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$RESTORE_DB"
conda run -n blogenv pg_restore \
  -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$RESTORE_DB" \
  --no-owner "$BACKUP_DIR/database.dump"
conda run -n blogenv psql \
  -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$RESTORE_DB" \
  -v ON_ERROR_STOP=1 -Atc \
  "select 'users='||count(*) from users union all select 'nodes='||count(*) from nodes union all select 'file_contents='||count(*) from file_contents union all select 'path_redirects='||count(*) from path_redirects union all select 'comments='||count(*) from comments union all select 'likes='||count(*) from likes union all select 'file_assets='||count(*) from file_assets;"
if [ "$STAGE" = "stage-3" ]; then
  printf 'Retaining disposable database %s for Stage 3 migration-twice and API health verification.\n' "$RESTORE_DB"
else
  conda run -n blogenv dropdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$RESTORE_DB"
fi
```

Failure to restore or unexpected counts stops cleanup/migration. Stage 3 alone retains `$RESTORE_DB` until all Section 12.4 migration-twice, API health, inventory, and restore-proof checks pass; Section 12.4 then drops it explicitly.

### 12.3 Canonical Stage 2 inventory and explicit-ID cleanup

This is the only authorized Stage 2 cleanup path. Any earlier or generated weaker cleanup snippet is non-authoritative and must not be run.

Inventory exact identities, parent relationships, paths, and recursive subtree counts before writing the reviewed SQL file:

```bash
conda run -n blogenv psql \
  -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
  -v ON_ERROR_STOP=1 -P pager=off <<'SQL'
WITH RECURSIVE tree AS (
  SELECT id, parent_id, kind, name, slug, ('/' || slug)::text AS url_path, 0 AS depth
  FROM nodes WHERE parent_id IS NULL
  UNION ALL
  SELECT n.id, n.parent_id, n.kind, n.name, n.slug,
         (tree.url_path || '/' || n.slug)::text, tree.depth + 1
  FROM nodes n JOIN tree ON n.parent_id = tree.id
), targets AS (
  SELECT * FROM tree
  WHERE name IN ('Smoke Notes','Local Smoke Renamed','Acceptance 1425','Acceptance 1426')
)
SELECT id, parent_id, kind, name, url_path, depth,
       (SELECT count(*) FROM tree d
        WHERE d.url_path = targets.url_path OR d.url_path LIKE targets.url_path || '/%') AS subtree_count
FROM targets ORDER BY url_path;
SQL
```

Copy only the confirmed UUIDs into a reviewed temporary SQL file outside Git, export its path as `CLEANUP_SQL`, and use this exact transaction:

```sql
BEGIN;
DO $$
DECLARE
  a1425 uuid := '<acceptance-1425-id>'::uuid;
  child uuid := '<confirmed-child-file-id>'::uuid;
  a1426 uuid := '<acceptance-1426-id>'::uuid;
  smoke uuid := '<smoke-notes-id>'::uuid;
  renamed uuid := '<local-smoke-renamed-id>'::uuid;
  n integer;
BEGIN
  IF (SELECT count(*) FROM nodes WHERE id=a1425 AND name='Acceptance 1425' AND kind='directory') <> 1 THEN RAISE EXCEPTION '1425 identity'; END IF;
  IF (SELECT count(*) FROM nodes WHERE id=child AND parent_id=a1425 AND kind='file') <> 1 THEN RAISE EXCEPTION '1425 child'; END IF;
  WITH RECURSIVE d AS (SELECT id FROM nodes WHERE id=a1425 UNION ALL SELECT x.id FROM nodes x JOIN d ON x.parent_id=d.id) SELECT count(*) INTO n FROM d;
  IF n <> 2 THEN RAISE EXCEPTION '1425 subtree %, expected 2',n; END IF;
  IF (SELECT count(*) FROM nodes WHERE id=a1426 AND name='Acceptance 1426' AND kind='directory') <> 1 THEN RAISE EXCEPTION '1426 identity'; END IF;
  WITH RECURSIVE d AS (SELECT id FROM nodes WHERE id=a1426 UNION ALL SELECT x.id FROM nodes x JOIN d ON x.parent_id=d.id) SELECT count(*) INTO n FROM d;
  IF n <> 1 THEN RAISE EXCEPTION '1426 subtree %, expected 1',n; END IF;
  IF (SELECT count(*) FROM nodes WHERE id=smoke AND name='Smoke Notes' AND kind='directory') <> 1 OR
     (SELECT count(*) FROM nodes WHERE id=renamed AND parent_id=smoke AND name='Local Smoke Renamed' AND kind='file') <> 1 THEN RAISE EXCEPTION 'preserved relationship'; END IF;
  DELETE FROM nodes WHERE id=child;
  DELETE FROM nodes WHERE id IN (a1425,a1426);
  IF EXISTS (SELECT 1 FROM nodes WHERE id IN (a1425,child,a1426)) THEN RAISE EXCEPTION 'target remains'; END IF;
  IF NOT EXISTS (SELECT 1 FROM nodes WHERE id=renamed AND parent_id=smoke) THEN RAISE EXCEPTION 'preserved relationship changed'; END IF;
END $$;
COMMIT;
```

Execute only after backup checksum and disposable restore PASS:

```bash
test -n "$CLEANUP_SQL" && test -f "$CLEANUP_SQL"
conda run -n blogenv psql \
  -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
  -v ON_ERROR_STOP=1 -f "$CLEANUP_SQL"
```

Rerun the inventory query. Any assertion failure aborts the transaction; any later mismatch uses the verified backup restore, never manual repair.

### 12.4 Canonical Stage 3 migration rehearsal and rollback

Add `api/internal/db/second_development_migration_test.go`. It accepts `TEST_DATABASE_URL`, invokes the real `db.RunMigrations` twice, and asserts schema, Current/Previous/Published Content mapping, Asset state, counts, references, and idempotence:

```bash
export RESTORE_URL="postgres://$DB_USER@$DB_HOST:$DB_PORT/$RESTORE_DB?sslmode=disable"
conda run -n blogenv psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$RESTORE_DB" -v ON_ERROR_STOP=1 -Atc 'select 1'
conda run -n blogenv env TEST_DATABASE_URL="$RESTORE_URL" bash -lc \
  'cd api && test -n "$TEST_DATABASE_URL" && CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./internal/db -run "^TestSecondDevelopmentMigrationTwice$" -v'
```

Start the actual API against the same retained disposable database and require `/health` success before enabling Stage 3 frontend behavior. After migration-twice, API health, inventory, and restore-proof assertions all pass, clean up explicitly:

```bash
conda run -n blogenv dropdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$RESTORE_DB"
unset RESTORE_URL RESTORE_DB
```

For a destructive local rollback, services must already be stopped. First archive failed state, then verify the known-good backup checksums **before** terminating connections or changing the database:

```bash
export FAILED_DIR="$XLAB_STATE/failed-stage-3-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$FAILED_DIR"
conda run -n blogenv pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" --format=custom --file "$FAILED_DIR/database.dump"
tar -C "$XLAB_STATE" -czf "$FAILED_DIR/uploads.tar.gz" uploads

(cd "$BACKUP_DIR" && sha256sum -c database.dump.sha256 && sha256sum -c uploads.tar.gz.sha256)

conda run -n blogenv psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 \
  -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='$DB_NAME' AND pid<>pg_backend_pid();"
conda run -n blogenv dropdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" --if-exists "$DB_NAME"
conda run -n blogenv createdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME"
conda run -n blogenv pg_restore -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" --no-owner --exit-on-error "$BACKUP_DIR/database.dump"
rm -rf "$XLAB_STATE/uploads.restore-tmp"; mkdir -p "$XLAB_STATE/uploads.restore-tmp"
tar -C "$XLAB_STATE/uploads.restore-tmp" -xzf "$BACKUP_DIR/uploads.tar.gz"
test -d "$XLAB_STATE/uploads.restore-tmp/uploads"
mv "$XLAB_STATE/uploads" "$FAILED_DIR/uploads-before-restore"
mv "$XLAB_STATE/uploads.restore-tmp/uploads" "$XLAB_STATE/uploads"
rmdir "$XLAB_STATE/uploads.restore-tmp"
```

Rerun inventory, migration fixture assertions, and native smoke before restarting normal services. Never improvise reverse SQL.

## 13. Requirement-to-test matrix

| Requirement | Test/evidence target | Command or method | Pass condition |
|---|---|---|---|
| Stage 1 successful create feedback | frontend regression + browser | targeted Node test, frontend gate, browser create | success never enters generic failure path; actionable API errors shown |
| Stage 1 identity/nav | current-user tests + browser | Anonymous/Reader/Author, invalid token, offline/network error, return target | one search input; one identity entry; auth/network states truthful |
| Stage 2 Admin tree | Go integration + browser | protected tree/lazy children/Draft branch tests | complete Admin tree without public Draft leakage |
| Stage 2 creation/order/move | Go repository/service + browser | concurrent names, mixed reorder, same-parent drag, mobile move, Directory Picker | atomic persistence/rollback; no drag reparenting |
| Stage 2 cleanup | backup/restore/inventory | Section 12 commands | only confirmed accidental IDs removed; preserved fixture remains |
| Stage 3 migration | migration fixtures + disposable DB | migration path twice, state/count assertions, restore proof | lossless, transactional, safely rerunnable |
| Stage 3 autosave/version/conflict | unit/integration/browser | controlled 15s timer, forced save, stale revision, restore | no silent overwrite; Current/Previous rules exact |
| Stage 3 publication/preview/assets | API/browser/security | role matrix, publish/unpublish, direct Asset URL, search | Draft private; Published stable until explicit publish |
| Existing public behavior | full API/browser regression | reading/search/comments/Likes/assets/redirects | no regression |
| Security invariants | security + code review reports | abuse cases plus independent reviewers | security PASS, architect CLEAR, code-reviewer APPROVE |

Stage-specific evidence paths are mandatory:

- `docs/verification/stage-<n>-team-log.md`;
- `docs/verification/stage-<n>-acceptance.md`;
- `docs/verification/stage-<n>-security.md`;
- `docs/verification/stage-<n>-code-review.md`;
- Stage 2: `stage-2-backup-and-restore.md`;
- Stage 3: `stage-3-migration-and-preview.md`.

## 14. Browser acceptance contract

Use the existing external `playwright-cli`; do not add an unpinned frontend test dependency. If repository Playwright packages are later desired, update `TECH_STACK.md` and manifests first.

For each runtime-facing stage:

```bash
playwright-cli -s=stage-<n>-desktop open http://127.0.0.1:5173
playwright-cli -s=stage-<n>-desktop resize 1440 900
playwright-cli -s=stage-<n>-desktop tracing-start
# execute the stage scenario with snapshot/fill/click/drag/eval commands
playwright-cli -s=stage-<n>-desktop console error
playwright-cli -s=stage-<n>-desktop requests
playwright-cli -s=stage-<n>-desktop screenshot
playwright-cli -s=stage-<n>-desktop tracing-stop
playwright-cli -s=stage-<n>-desktop close

playwright-cli -s=stage-<n>-mobile open http://127.0.0.1:5173
playwright-cli -s=stage-<n>-mobile resize 390 844
# execute mobile scenario, capture console/requests/screenshot
playwright-cli -s=stage-<n>-mobile close
```

Pass requires:

- expected visible state and persisted server state;
- no unexpected console error;
- no unexpected failed network response;
- screenshot/trace path recorded;
- browser session closed;
- result tied to the integrated SHA.

## 15. Stage 1 — Reliability, navigation, and identity

### 15.1 Objective and implementation packets

Repair the false successful-create/generic-error behavior and make navigation/identity truthful without changing the content model.

Red acceptance must cover:

- successful Directory creation shows success, returned URL Path, and selected/opened node—not generic failure;
- server validation/conflict/auth errors become actionable messages;
- identity loading skeleton;
- invalid token clears to Anonymous Visitor;
- network failure shows retry and is not treated as anonymous;
- global bar keeps Recent, Directory button, one search input, and one identity entry;
- separate Search and permanent Admin links are absent;
- Reader menu exposes Logout;
- Author entry opens `/admin`;
- Reader `/admin` shows Author access required and Return to Recent;
- Anonymous `/admin` redirects to Login with safe return target;
- `/search` contains query/results/empty/error only;
- login/logout destinations and loop prevention.

Backend owner:

- preserve `RequireAdmin` and token semantics;
- update OpenAPI first only if shared error schemas change;
- make validation/conflict/auth error bodies precise;
- add tests before code.

Frontend owner:

- introduce one application current-user state;
- repair Admin create event/error flow without Stage 2 redesign;
- replace duplicate navigation entries;
- simplify Search page;
- implement role-aware login/logout/return behavior.

Primary touchpoints: `web/src/App.tsx`, `GlassNav.tsx`, `SearchPage.tsx`, `AuthPages.tsx`, `AdminPage.tsx`, `web/src/lib/api.ts`, `web/src/lib/auth.ts`, targeted frontend tests, auth middleware/handler tests.

### 15.2 Packet DAG

`.omx/plans/stages/stage-1-team-dag.json` splits backend Red/errors, frontend Red/identity-navigation/create, integrated acceptance, integrated security, and closeout. Every node contains exact paths, Red/Green command, expected result, and integration checkpoint.

### 15.3 Acceptance PASS

- top bar is minimal and identity-correct;
- Anonymous and Reader protection are distinct and truthful;
- Directory creation reports real success/error;
- one search input exists;
- public reading/search/comments/Likes/assets remain intact;
- desktop and mobile pass.

### 15.4 Security PASS

- missing/invalid/expired/tampered token cases;
- Reader cannot access Author API/data;
- network failure cannot silently downgrade auth state;
- return target cannot open-redirect or loop;
- no credentials/tokens in logs or Git.

## 16. Stage 2 — Chinese Author Workspace and protected Content Tree

Stage 2 replaces the current form-heavy Admin page with a desktop-first, Chinese **Author Workspace**. The protected route may remain `/admin`, but product UI must not present the surface as `Admin / Tree Manager`; it is the Author-facing workspace for managing the Content Tree, Files, assets, publication controls, and node settings.

Stage 2 directly addresses the failed user acceptance from 2026-06-13:

- newly created Directory/File nodes must appear immediately in every relevant Author Workspace tree/navigation surface, so success cannot be mistaken for failure;
- every subflow needs explicit in-app return controls and lightweight breadcrumbs, not browser-back dependence;
- generated Files must be selectable, editable, publishable, and unpublishable from the Author Workspace;
- Authors browsing public Directory/File pages need an Author-only `管理此目录` / `编辑文件` entry back into the workspace;
- the Author Workspace must be graphical, Chinese, operation-first, minimal, readable, and structurally clear.

### 16.1 Pre-stage controlled data gate

Execute Section 12 backup, checksum, disposable restore, inventory, and explicit-ID cleanup before schema migration, fixture cleanup, or destructive acceptance setup. Preserve `Smoke Notes / Local Smoke Renamed`; delete only confirmed accidental IDs already documented by Stage 1 evidence.

Create or refresh a dedicated Stage 2 acceptance fixture under a clearly named root such as `/stage-2-acceptance`. The fixture must cover:

- Directory and nested Directory cases;
- Draft File and Published File cases;
- Chinese and mixed Chinese/English URL Paths;
- same-parent ordering;
- cross-Directory move constraints;
- non-empty Directory and Published File deletion protection;
- Author-only public Directory/File manage/edit entry.

Document fixture creation and cleanup rules in `docs/verification/stage-2-acceptance.md`.

### 16.2 Backend/API packets

Update `docs/api/openapi.yaml` first for every shared API contract change. Implement protected Author Workspace APIs with clear repository/service/handler boundaries, keeping SQL in repositories and avoiding UI-specific backend hacks. Prioritize readability, extensibility, and a rigorous structure.

Required capabilities:

- protected Author Content Tree containing all Directories, Draft Files, Published Files, and Files with unpublished changes;
- node detail for Directory/File workspace loading, including path, parent, order, status, and relevant child metadata;
- context-aware create for Directory/File using selected parent Directory, with backend-authoritative Name-to-URL-Path generation;
- URL Path generation that preserves Chinese characters, normalizes Latin text to lowercase hyphenated segments, and appends numeric suffixes only for initial create conflicts;
- strict explicit URL Path edit handling without silent rewrite;
- same-parent mixed Directory/File reorder with transaction safety and lost-update protection where practical;
- graphical Directory Picker support for cross-Directory moves, including impact preview, cycle prevention, subtree path rewrite, and redirects for formerly public paths;
- protected deletion constraints with clear reasons for non-empty Directories and Published Files;
- publication state read/update sufficient for Stage 2 manual-save File workspace, without implementing Stage 3 autosave/version/publication snapshot semantics.

Do not implement Stage 3 Content Version history, Draft Preview, Draft/Published Asset split, or independent Published Content snapshots in Stage 2.

### 16.3 Frontend packets

Build a desktop-first Chinese Author Workspace. Mobile in Stage 2 is no-regression sanity only: phone width must not visibly break, must provide orientation or a basic Content Tree/exit path, and must avoid major overflow/overlap; complete mobile create/edit/move/delete flows are deferred.

#### 16.3.1 Workspace shell

- Desktop layout: fixed two-column workspace. Left: protected Content Tree. Right: current contextual workspace.
- Mobile layout: no-regression single-column fallback only for Stage 2.
- Visual direction: lightweight professional writing/management tool inside Glass Ricepaper; quiet, sparse, readable, and operation-first. Use stronger card/status treatment only for creation success, publication state, save/error state, and danger confirmations.
- UI copy: Author Workspace and Author-facing flows are Chinese. Public reading pages are not broadly redesigned, except Author-workflow touchpoints and necessary login/identity wording.
- Top workspace controls include context, public-view action where relevant, system/status actions, and Author logout without reviving the old `Admin / Tree Manager` hero.

#### 16.3.2 Content Tree

- Protected expand/collapse tree; no tree search in Stage 2.
- Show all Directories, Draft Files, Published Files, and Files with unpublished changes.
- Restore browser-local selection/expanded state when safe.
- Creation refreshes the tree, expands the parent Directory, selects the new node, and opens the correct right workspace.
- Public Directory/File Author entry expands ancestors and selects/opens the target node.
- Same-parent desktop drag sorting only; drag must never change parent Directory.
- Mobile sorting is deferred with mobile no-regression sanity only, unless explicitly implemented as safe up/down controls.

#### 16.3.3 Directory workspace

Selecting a Directory opens a Directory overview, not settings by default. The right workspace shows:

- Directory Name and current URL Path;
- clear `新建 Directory` and `新建 File` actions;
- child cards for current Directory contents;
- a Settings entry.

Creating a Directory uses only `名称`; creating a File uses `名称` and `格式` (`Markdown` / `HTML Document`). The UI shows a read-only final URL Path preview and never exposes Parent ID, Node ID, `slug`, or Sort order in the primary creation flow.

Creation success shows a lightweight Chinese toast, refreshes tree/navigation state, expands/selects/opens the new node, and displays the final path clearly.

#### 16.3.4 File workspace

Selecting a File opens a File workspace shell with the Stage 3-compatible shape, but Stage 2 keeps saving manual:

- header: File Name, status (`草稿`, `已发布`, `有未发布修改` where supported), URL Path, public-view action, and one primary publication action;
- sections/tabs: `内容`, `资源`, `设置`;
- `内容`: body editor, keywords, manual save, clear Chinese success/error messages;
- `资源`: upload/view/delete using existing asset model without Stage 3 Draft/Published Asset split;
- `设置`: Name, URL Path, move, delete constraints, and danger zone.

Publication control uses one primary action: `发布` for Draft, `发布更新` for saved unpublished changes, and status-only `已发布` when current. `撤回发布` is secondary/overflow/danger, not a sibling primary button.

#### 16.3.5 Settings, return, and public Author entry

- Settings are divided into `基础信息`, `位置`, and `危险操作`.
- Danger actions live at the bottom, are visually distinct, require Chinese second confirmation, and explain blocked Published File / non-empty Directory deletion.
- Cross-Directory movement uses a graphical Directory Picker and path/impact preview; it never requires Parent ID.
- All right-workspace subflows that replace the current workspace include explicit return buttons with destination-specific copy such as `返回当前目录`, `返回文件内容`, or `返回设置`.
- Lightweight breadcrumbs/path indicators aid orientation but are not the only return mechanism.
- Author-only public Directory/File entries show `管理此目录` or `编辑文件`; clicking enters the workspace with the target node selected.

### 16.4 Packet DAG update

Before Stage 2 implementation, revise the Stage 2 packet DAG and evidence checklist to match this replanned scope. The packet graph should preserve the fresh-Team-per-stage model and split work into at least:

1. coordinator preflight: PRD drag wording alignment, Stage 2 replanned scope application, fixture/data gate, and evidence skeleton;
2. backend contract/API Red tests and OpenAPI update;
3. backend protected Author tree/detail/create/reorder/move/delete implementation;
4. frontend Author Workspace shell/tree/create/directory workspace;
5. frontend File workspace/settings/public Author entry;
6. acceptance fixture and desktop browser workflow;
7. security review for protected tree, Draft leakage, destructive operations, redirects, and public Author entry;
8. independent architecture/code review and closeout.

Coordinator records checkpoints in `PROGRESS.md` and `docs/verification/stage-2-team-log.md`. Acceptance owns `docs/verification/stage-2-acceptance.md`; security owns `docs/verification/stage-2-security.md`; independent review owns `docs/verification/stage-2-code-review.md`.

### 16.5 Acceptance PASS

Stage 2 user acceptance is desktop-first and follows this primary path:

1. log in as Author and enter the Chinese Author Workspace;
2. create a Directory and File with minimal Chinese forms;
3. verify Content Tree refreshes immediately, expands the parent, selects/opens the new node, and shows a clear Chinese toast/path;
4. edit File content and keywords, manually save, and see clear Chinese feedback;
5. publish the File and verify public access;
6. from the public File, click Author-only `编辑文件` and return to the workspace with the File selected;
7. verify `撤回发布` is reachable as a secondary/danger action and hides the public File after confirmation;
8. verify Settings move/URL Path/delete-constrained cases show clear Chinese prompts and do not expose Parent ID, Node ID, or `slug`;
9. verify same-parent desktop drag sorting persists and never reparents;
10. verify public homepage, Recent cards, public Directory/File reading, comments/Likes, and Glass Ricepaper are not redesigned except required Author entry/regression repair;
11. verify mobile no-regression sanity only: phone width opens without major layout breakage and provides a basic orientation/exit path.

Automated and documented gates must additionally cover:

- Draft-only branches appear only in protected Author Content Tree;
- English/Chinese/mixed URL Path normalization and same-parent conflict suffix on create;
- explicit URL Path conflicts never silently rename;
- non-empty Directory and Published File deletion are protected;
- Reader and Anonymous Visitor cannot access Author Workspace APIs or Draft content;
- full-text search fallback remains preserved when semantic indexing is unavailable;
- iframe sandbox remains `sandbox="allow-scripts"` without `allow-same-origin`.

### 16.6 Security PASS

- Anonymous Visitor and Reader are denied protected Author Workspace APIs.
- Draft Files and draft-only branches do not leak through public tree, search, recent, assets, or public Author entry logic.
- Invalid parent, cycles, path traversal, reorder lost update, redirect loop/chain attacks, and destructive operation bypasses are rejected.
- Destructive actions show truthful impact and confirmation.
- Public Author-only edit/manage entry is hidden from Reader/Anonymous Visitor and does not expose Draft targets.
- Backup/restore and acceptance fixture evidence are recorded.

## 17. Stage 3 — Autosave, versions, publication snapshots, Draft Preview, Draft/Published Assets

Stage 3 additionally has a new final-stage MCP requirement from user replanning: build a separate Blog MCP Server process/package so external AI tools can interact with the Blog. The user wants a high-trust server-local stdio MCP Server with full Author permissions so trusted AI agents running on the server can autonomously create, edit, publish, unpublish, move, delete, modify URL Paths, manage assets, and read/search the Content Tree. The MCP Server should expose a complete Author tool set by Stage 3 closeout, grouped into read (`list_content_tree`, `get_file`, `search_files`), content (`create_directory`, `create_file`, `update_file_content`, `update_file_settings`), publish (`publish_file`, `unpublish_file`), tree (`move_node`, `reorder_children`, `delete_node`), assets (`upload_asset`, `delete_asset`, `list_assets`), and maintenance (`rebuild_search_index`, `export_backup`). It should reuse backend service/API-client capabilities instead of duplicating business logic or SQL, and include minimal engineering safeguards for presentation-quality architecture: explicit enablement configuration, operation audit logs, automatic backup/export before destructive batches where practical, authentication/deployment posture, and emergency disable behavior. Public HTTP/SSE MCP transport is out of initial scope and may be added later behind explicit auth and network-binding controls.


### 17.1 Migration/data model packet

Execute Section 12 Stage 3 backup/rehearsal. Add a new migration providing:

- Current Content Version;
- Previous Content Version;
- independent Published Content;
- monotonic revision for optimistic concurrency;
- Draft/Published Asset state.

Transactional mapping:

- existing Published File → Current + Published Content;
- existing Draft File → Current only;
- Previous empty;
- Asset state derived from prior File publication and actual Published Content references.

Do not enable new frontend behavior until disposable migration, second run, assertions, and restore proof pass.

### 17.2 Backend/API packets

Update OpenAPI first and implement:

- Current autosave with expected revision; stale write conflict;
- old Current becomes Previous only after successful changed save;
- identical content/Keywords/Render Format creates no version;
- restore atomically swaps Current/Previous;
- publish snapshots Current into Published Content;
- unpublish changes visibility and retains Published Content;
- Author-only Draft Preview and Draft Asset serving;
- publish promotes only Draft Assets referenced by Current content;
- deletion blocked for Published Asset referenced by Published Content;
- search indexes Published Content; semantic failure remains non-blocking.

### 17.3 Frontend packets

Implement:

- autosave 15 seconds after input stops and immediately on blur, node change, Publish, Logout, or leaving Admin;
- Editing, Saving, Saved, Save failed, Conflict states;
- failed required save blocks navigation/logout/publication and preserves text;
- conflict actions Reload latest / Copy my changes; no auto-merge;
- Current/Previous timestamps, compare, reversible restore swap;
- Publish / Publish changes / Published, Unpublished changes, confirmation with path/differences, Unpublish in overflow;
- responsive 55/45 split, Editor/Split/Preview, mobile Edit/Preview;
- exact sandboxed HTML preview;
- `/admin/preview/{file_id}` for Author, saved Current, Draft Assets, cross-tab refresh plus manual fallback;
- Draft/Published Asset presentation and publish summary.

### 17.4 Packet DAG

`.omx/plans/stages/stage-3-team-dag.json` splits backup/migration, OpenAPI, version/autosave, publication/preview, assets/search, editor autosave, version/publication UI, preview/assets UI, acceptance, security, and closeout.

### 17.5 Acceptance PASS

- migration fixtures for existing Draft/Published Files and Assets;
- controlled autosave timing and forced saves;
- no-op save produces no Previous;
- rotation and reversible restore;
- Published Content stable through Current edits/restore;
- unpublish/re-publish exact;
- stale revision never overwrites;
- save failure blocks unsafe transitions and preserves text;
- Draft Preview role matrix;
- Draft Asset isolation/reference promotion;
- Published Asset deletion guard;
- cross-tab refresh;
- desktop/mobile and sandbox regression;
- full API/browser/public regression.

### 17.6 Security PASS

- Anonymous/Reader denied Draft Preview and direct Draft Asset URLs;
- filename/path manipulation rejected;
- stale revision/retry races do not overwrite;
- save failure followed by navigation/logout/publish remains blocked;
- Published Content remains stable;
- Markdown XSS and iframe `sandbox="allow-scripts"` without `allow-same-origin` pass;
- public search indexes Published Content only;
- migration is transactional/re-runnable and restore is proven.

## 18. Expanded test and observability plan

### Unit

Identity transitions, API error classification, normalization/conflicts, reorder/cycle validation, revision compare-and-swap, no-op/rotation/restore, Asset guards, token/XSS/sandbox.

### Integration

OpenAPI routes/schemas, PostgreSQL tree/concurrent create/reorder/move/redirect/delete, migration and second-run fixture, optimistic concurrency, independent Published Content, Draft authorization, published search source and fallback.

### End-to-end

Anonymous/Reader/Author; graphical tree/create/order/move/settings; autosave failure/conflict/restore/publish/unpublish; Draft Preview/Assets; public reading/search/comments/Likes; desktop 1440×900 and mobile 390×844.

### Observability

Team/event snapshots, source/integration SHAs, exact versions, backup checksums and restore counts, sanitized API matrix, screenshot/trace paths, console/network result, security reproduction steps, and explicit external boundaries.

## 19. Deliberate pre-mortem

### Scenario 1 — coordinator becomes passive

Signal: progress lags Team state or claims completion without integrated evidence.

Mitigation: persistent control task, event cursor, milestone triggers, sole progress ownership, closes last.

### Scenario 2 — cleanup/migration corrupts accepted data

Signal: unrestorable backup, name-only deletion, partial migration, unexpected counts/references.

Mitigation: checksum, disposable restore, explicit UUID assertions, transactions, second migration run, stop-on-mismatch.

### Scenario 3 — Draft/public leak

Signal: Draft in public tree/search, Reader preview, unpromoted Asset public URL.

Mitigation: Published Content as public/search source, separate Author routes, adversarial security gate.

### Scenario 4 — detached integration tests stale code

Signal: worker says PASS but leader cherry-pick fails or integrated SHA was never retested.

Mitigation: source→integration ledger, clean verifier reset, verification reset after each integration, no progress advancement before integrated proof.

## 20. Stage closeout template

```text
Stage: <n>
Team: <team-name>
Integrated commit: <sha>
Team terminal counts: pending=0 in_progress=0 failed=0
Tested:
- <command> -> PASS (<observed result>)
Acceptance: <stage acceptance artifact> -> PASS
Security: <stage security artifact> -> PASS
Architect: <stage code review artifact> -> CLEAR
Code reviewer: <stage code review artifact> -> APPROVE
Not tested:
- <external boundary and reason, or none>
Rollback:
- Git checkpoint and database/uploads restore instruction
Next breakpoint:
- exact next command
```

## 21. Final native acceptance and deployment boundary

After Stage 3:

1. Create a clean acceptance Directory and Markdown/HTML Files.
2. Test identity, creation, ordering, moves, autosave, conflicts, versions, Draft Preview, publication, redirects, Assets, search, comments, and Likes.
3. Run desktop/mobile browser acceptance and console/network inspection.
4. Record evidence and update `PROGRESS.md`.
5. Obtain user acceptance.
6. Only then perform Docker/WSL Compose smoke and server deployment.

## 22. Stop conditions

Pause rather than broaden scope when:

- migration is not demonstrably lossless/reversible/re-runnable;
- public behavior regresses;
- acceptance/security/architect/code-review verdict is not PASS/PASS/CLEAR/APPROVE;
- a stage cannot end runnable;
- Team/task/worktree/integration state cannot be reconciled;
- ownership overlaps without leader reassignment;
- exact versions/contracts cannot be satisfied;
- a request changes the locked data/publication model;
- user acceptance finds a blocker.

Never weaken/delete tests, suppress errors, silently default, or change scope to pass a gate.

## 23. ADR

### Decision

Use a fresh five-worker Team per stage with coordinator, backend, frontend, acceptance, and security functional seats under automatic detached worktrees and leader-only integration. Require independent architect CLEAR and code-reviewer APPROVE.

### Drivers

Stage isolation, auth/data/publication risk, interruption recovery, and no self-approval.

### Alternatives considered

- one long-lived Team: rejected for stale state and weak rollback isolation;
- Ralph as primary: rejected because parallel monitoring/acceptance/security separation is required;
- native subagents only: rejected because durable Team task/mailbox/worktree recovery is required.

### Why chosen

It directly satisfies the required Agent roles, preserves the locked three stages, and produces auditable integration, acceptance, security, and recovery evidence.

### Consequences

More launch/integration overhead; stricter ownership and evidence discipline; clearer rollback and safety.

### Follow-ups

- apply this artifact as `docs/plans/SECOND_DEVELOPMENT.md` only after consensus handoff;
- optionally configure Matt skills through a separate user-confirmed setup;
- optionally wrap stages in an Ultragoal ledger.

## 24. Execution handoff and Goal-Mode suggestions

Planning ends after Architect and Critic approve this artifact. It must not start implementation itself.

Required execution engine: **Team** for every stage.

Recommended durable combination: **Ultragoal + Team**:

- Ultragoal owns leader goal/ledger/checkpoints.
- Team executes parallel functional seats and returns integrated evidence.
- Leader checkpoints only after terminal Team state and PASS/PASS/CLEAR/APPROVE.

Other paths:

- `$ultragoal`: default durable wrapper;
- `$team`: required stage execution;
- `$autoresearch-goal`: not appropriate;
- `$performance-goal`: only for a later measured optimization project;
- `$ralph`: explicit fallback only for a later narrow single-owner fix loop.

## 25. Planning changelog

- Preserved and inlined all locked product stages and acceptance boundaries.
- Added the required coordinator, development, acceptance, and security Team roles.
- Corrected for automatic detached worktrees and defined leader integration plus verifier reset.
- Added deterministic Team task/API templates, event cursor, and shutdown gates.
- Added exact out-of-Git backup/restore/cleanup/migration safety commands.
- Added requirement-test matrix and concrete browser acceptance contract.
- Required stage code-review evidence with architect CLEAR and code-reviewer APPROVE.
- Reconciled same-parent drag sorting versus prohibited drag reparenting.
- Expanded role roster, pre-mortem, ADR, skill routing, and Team + Ultragoal handoff.
