# Stage 2 Gateway 1 — Backup, Restore Proof, and Fixture

Status: complete

Verdict: **PASS**

Team: `execute-approved-xlab-015f30a9`
Task: `7` / `s2-01-data-fixture`
Worker: `worker-4`
Reset/integration SHA: `d6b89497c702dc59fb188f61cb86a2b897384a21`

## Decision

Gateway 1 passes. The local database and uploads were backed up before creating
any Stage 2 fixture data, the database dump was restored into a disposable
database successfully, `/stage-2-acceptance` fixture IDs and paths are recorded,
and preserved baseline content was not modified.

No feature code was changed.

## Environment and services

Local services were started for fixture work:

```text
~/.local/share/xlab-blog/start-local.sh
curl -fsS http://127.0.0.1:8080/api/health  -> {"status":"ok","database":"ok"}
curl -fsS http://127.0.0.1:5173/ >/dev/null -> web-ok
```

Database URL used by the local stack:

```text
postgres://zephry_xzx@127.0.0.1:55432/xlab_blog?sslmode=disable
```

## Backup evidence

Backup directory:

```text
~/.local/share/xlab-blog/backups/stage-2-gateway1-20260613T134736+0800
```

Artifacts:

```text
xlab_blog.dump  PostgreSQL custom-format dump
uploads.tgz     uploads directory archive
SHA256SUMS.txt  checksums
```

Checksums:

```text
24b2047df91c5586f6cff5d45d53d8c6f88e57ad8d69c44c31431b1e1e41cabe  uploads.tgz
d2e1712d0316f6003867d856b167e1d1d33de6577c5dcc2705583383ee1bf86a  xlab_blog.dump
```

## Disposable restore proof

Restore proof: **PASS**.

The dump was restored into a disposable database, counted, then dropped:

```text
createdb xlab_blog_restore_stage2_20260613134917
pg_restore -d xlab_blog_restore_stage2_20260613134917 xlab_blog.dump
nodes=8
file_contents=5
file_assets=1
dropdb xlab_blog_restore_stage2_20260613134917
```

## Pre-fixture baseline snapshot

Before creating the fixture:

```text
before_nodes=8
before_stage2_nodes=0
before_non_stage2_paths=/research,/research/my-first-note,/smoke-notes,/smoke-notes/local-smoke-renamed,/smoke-notes/my-first-note,/test-lab,/test-lab/hello-stage-one,/test-lab/test-html
```

Because `before_stage2_nodes=0`, no fixture cleanup was needed or performed.

## Stage 2 fixture

Fixture root: `/stage-2-acceptance`

Recorded IDs and paths:

| Role | ID | Parent ID | Kind | Name | Slug | Path | Status |
|---|---|---|---|---|---|---|---|
| Fixture root | `77473f2e-6069-48ff-95a7-3d7173d090c4` | — | directory | Stage 2 Acceptance | `stage-2-acceptance` | `/stage-2-acceptance` | n/a |
| Draft branch | `f2f3fb74-33f0-4264-baf2-b26e5d06e83e` | `77473f2e-6069-48ff-95a7-3d7173d090c4` | directory | Draft Branch | `draft-branch` | `/stage-2-acceptance/draft-branch` | n/a |
| Draft file | `5b796f40-e15a-42fa-8832-9cfbd1dcd21e` | `f2f3fb74-33f0-4264-baf2-b26e5d06e83e` | file | Draft Fixture | `draft-fixture` | `/stage-2-acceptance/draft-branch/draft-fixture` | draft |
| Published file | `a260a9f7-ecf3-4c3d-a87e-cc96b44c73bc` | `77473f2e-6069-48ff-95a7-3d7173d090c4` | file | Published Fixture | `published-fixture` | `/stage-2-acceptance/published-fixture` | published |

Creation path output:

```text
77473f2e-6069-48ff-95a7-3d7173d090c4||directory|Stage 2 Acceptance|stage-2-acceptance|/stage-2-acceptance|10
f2f3fb74-33f0-4264-baf2-b26e5d06e83e|77473f2e-6069-48ff-95a7-3d7173d090c4|directory|Draft Branch|draft-branch|/stage-2-acceptance/draft-branch|10
5b796f40-e15a-42fa-8832-9cfbd1dcd21e|f2f3fb74-33f0-4264-baf2-b26e5d06e83e|file|Draft Fixture|draft-fixture|/stage-2-acceptance/draft-branch/draft-fixture|10
a260a9f7-ecf3-4c3d-a87e-cc96b44c73bc|77473f2e-6069-48ff-95a7-3d7173d090c4|file|Published Fixture|published-fixture|/stage-2-acceptance/published-fixture|20
```

Publication-state check:

```text
stage2_statuses=draft-fixture:draft,published-fixture:published
```

Public visibility smoke:

```text
GET /api/tree/resolve?path=/stage-2-acceptance/published-fixture
{
  "type": "file",
  "path": "/stage-2-acceptance/published-fixture",
  "status": "published"
}

GET /api/tree/resolve?path=/stage-2-acceptance/draft-branch/draft-fixture
HTTP 404
```

## Baseline preservation proof

After creating the fixture:

```text
after_nodes=12
after_stage2_nodes=4
after_non_stage2_paths=/research,/research/my-first-note,/smoke-notes,/smoke-notes/local-smoke-renamed,/smoke-notes/my-first-note,/test-lab,/test-lab/hello-stage-one,/test-lab/test-html
```

The `before_non_stage2_paths` and `after_non_stage2_paths` values match exactly.
The only new nodes are the four `/stage-2-acceptance` fixture nodes.

## Cleanup policy

- Do not delete or modify this fixture during backend/frontend development.
- Downstream Stage 2 acceptance may edit/publish/unpublish the fixture as part of
  the integrated desktop workflow, but must record any changes.
- Closeout cleanup, if requested, must use the explicit IDs recorded above and
  must take a fresh backup first.

## Verification

```text
PASS backup: pg_dump custom-format dump created before fixture work.
PASS uploads backup: uploads.tgz created before fixture work.
PASS checksums: SHA256SUMS.txt recorded for dump and uploads archive.
PASS disposable restore: pg_restore succeeded into xlab_blog_restore_stage2_20260613134917 and row counts were readable; disposable DB dropped.
PASS fixture: /stage-2-acceptance root, draft branch, draft file, and published file created with IDs/paths recorded.
PASS draft isolation smoke: draft fixture returns HTTP 404 through public resolver.
PASS published smoke: published fixture resolves publicly as published File.
PASS baseline preservation: non-stage2 path list unchanged before/after fixture creation.
PASS backend gates: go test ./..., go vet ./..., gofmt scan.
PASS frontend gates: node --test tests/*.test.mjs 19/19, npm run lint, npm run build.
```

## Not tested

- Stage 2 protected Author Workspace APIs and browser acceptance are downstream
  tasks after backend/frontend implementation and leader integration.
- Docker Compose/server deployment and external DashScope embeddings remain out of scope.
