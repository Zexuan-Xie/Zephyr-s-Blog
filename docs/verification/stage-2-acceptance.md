# Stage 2 Acceptance

Status: complete

Verdict: **FAIL**

Integrated SHA tested: `fe97dcd44dc74ca2c0eff36923e9a68882e577f3`
Team: `execute-approved-xlab-015f30a9`
Task: `16` / `s2-11-acceptance`
Worker: `worker-4`
Evidence directory: `docs/verification/stage-2-browser-20260613/`

## Decision

Gateway 6 integrated acceptance fails before browser approval because the required
backend gate does not compile at the integrated SHA. Per the Stage 2 gate rules,
worker-4 did not patch feature code in the acceptance lane and did not claim a
browser PASS on top of a failing backend gate.

Blocking failure:

```text
api/internal/tree/lifecycle_service_test.go:213:35: method fakeLifecycleRepository.HasChildNodes already declared at api/internal/tree/lifecycle_service_test.go:200:35
```

This duplicate test-helper method causes both `go test ./...` and `go vet ./...`
to fail for `xlab-blog/api/internal/tree`.

## Fixture baseline

Gateway 1 fixture evidence remains recorded in
`docs/verification/stage-2-backup-and-fixture.md`.

Use the recorded fixture root and IDs for the retest after repair:

```text
/stage-2-acceptance
root:           77473f2e-6069-48ff-95a7-3d7173d090c4
draft branch:   f2f3fb74-33f0-4264-baf2-b26e5d06e83e
draft file:     5b796f40-e15a-42fa-8832-9cfbd1dcd21e
published file: a260a9f7-ecf3-4c3d-a87e-cc96b44c73bc
```

Gateway 1 public smoke:

- `/stage-2-acceptance/published-fixture` resolves as a published File.
- `/stage-2-acceptance/draft-branch/draft-fixture` returns HTTP 404 publicly.

## Verification results

### Backend gate — FAIL

Command:

```bash
cd api
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./...
```

Evidence: `stage-2-browser-20260613/backend-go-test.log`

Result:

```text
FAIL xlab-blog/api/internal/tree [build failed]
```

Root cause from output:

```text
internal/tree/lifecycle_service_test.go:213:35: method fakeLifecycleRepository.HasChildNodes already declared at internal/tree/lifecycle_service_test.go:200:35
```

Command:

```bash
cd api
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go vet ./...
```

Evidence: `stage-2-browser-20260613/backend-go-vet.log`

Result:

```text
vet: internal/tree/lifecycle_service_test.go:213:35: method fakeLifecycleRepository.HasChildNodes already declared at internal/tree/lifecycle_service_test.go:200:35
```

Command:

```bash
cd api
test -z "$(gofmt -l .)"
```

Evidence: `stage-2-browser-20260613/backend-gofmt.log`

Result: **PASS**; no gofmt output.

### Frontend gate — PASS

Command:

```bash
cd web
node --test tests/*.test.mjs
```

Evidence: `stage-2-browser-20260613/frontend-node-test.log`

Result: **PASS**, 32/32 tests.

Command:

```bash
cd web
npm run lint
```

Evidence: `stage-2-browser-20260613/frontend-lint.log`

Result: **PASS**.

Command:

```bash
cd web
npm run build
```

Evidence: `stage-2-browser-20260613/frontend-build.log`

Result: **PASS**; `tsc --noEmit` and Vite build completed.

### Runtime/browser acceptance — NOT RUN

Desktop/mobile browser acceptance was intentionally not run to a PASS/FAIL UI
verdict after the backend compile gate failed. Running browser checks against a
candidate that cannot pass the mandatory backend gate would produce misleading
acceptance evidence.

Services were reachable before gate execution:

```text
curl -fsS http://127.0.0.1:8080/api/health -> {"status":"ok","database":"ok"}
curl -fsS http://127.0.0.1:5173/ >/dev/null -> web-ok
```

## Retest checklist after repair

After the backend duplicate-method issue is repaired and integrated, worker-4 or
a follow-up acceptance task should reset to the new leader-integrated SHA and run:

1. Full backend gate: `go test ./...`, `go vet ./...`, gofmt scan.
2. Full frontend gate: `node --test tests/*.test.mjs`, `npm run lint`, `npm run build`.
3. Native API smoke against `/stage-2-acceptance`.
4. Desktop browser acceptance at `1440x900` with screenshots, console, and network checks.
5. Mobile no-regression sanity at `390x844` with screenshot, console, and network checks.

## Not tested

- Author primary desktop workflow.
- Public Author entries in a real browser.
- Unpublish hiding public File in a real browser.
- Move/delete/reorder prompts in a real browser.
- Mobile no-regression sanity.

These remain blocked by the backend compile failure above.
