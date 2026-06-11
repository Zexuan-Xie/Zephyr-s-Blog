# Stage 1 Acceptance

Status: pending

Verdict: pending

Stage: Reliability, navigation, and identity

Owner: Worker 4, independent acceptance seat

Prepared against: `f7772381459fefe4435455cdef31f5b03bdf09e9`

Integrated SHA under test: pending leader handoff

This matrix prepares independent acceptance only. A PASS may be recorded only after the leader supplies the integrated Stage 1 SHA, this worktree is clean and reset to that SHA, and all static, API, database, desktop, mobile, console, and network checks below have fresh evidence.

## Preconditions

1. Receive the leader-recorded Stage 1 integration SHA and confirm every required backend/frontend source commit has an integration SHA in the team log.
2. Preserve this evidence file, then reset a clean detached worktree to the integration SHA:

   ```bash
   git status --short
   git checkout --detach "$STAGE_INTEGRATION_SHA"
   test "$(git rev-parse HEAD)" = "$STAGE_INTEGRATION_SHA"
   ```

3. Start the native PostgreSQL, API, and Vite services; do not use Docker before native user acceptance:

   ```bash
   ~/.local/share/xlab-blog/start-local.sh
   curl -fsS http://127.0.0.1:8080/api/health
   curl -fsS http://127.0.0.1:5173/ >/dev/null
   ```

4. Use dedicated disposable acceptance records with recorded IDs. Do not delete or mutate unrelated baseline content.
5. Record exact Node.js, npm, Go, PostgreSQL, and `playwright-cli` versions.

## Scenarios

| ID | Actor / viewport | Scenario | Required observations |
|---|---|---|---|
| S1-01 | Anonymous Visitor, desktop and mobile | Open the public site with no stored token while current-user resolution is pending. | The identity control reserves its final space with a quiet skeleton; Login does not flash before resolution. |
| S1-02 | Anonymous Visitor, desktop and mobile | Resolve current user with no token, then inspect the global bar. | Recent, Directory button, exactly one search input, and exactly one identity entry are present. A separate Search link and permanent Admin link are absent. |
| S1-03 | Invalid-token session, desktop | Seed an invalid/expired token and load a public route. | The invalid token is cleared and the truthful final identity is Anonymous Visitor; the page does not loop or expose Author controls. |
| S1-04 | Network-failure session, desktop and mobile | Route or disable the current-user request while a stored session exists. | A retryable network error is shown; the state is not silently downgraded to Anonymous Visitor and Login is not presented as an authentication decision. |
| S1-05 | Reader, desktop and mobile | Log in with no return target and open the identity menu. | Login lands on `/recent`; the menu identifies the Reader and exposes Logout; Logout reaches the documented public destination without a redirect loop. |
| S1-06 | Reader, desktop and mobile | Navigate directly to `/admin`. | The Reader remains logged in, sees `Author access required`, and can use `Return to Recent`; no Author data or controls render. |
| S1-07 | Anonymous Visitor, desktop and mobile | Navigate directly to `/admin`. | The browser reaches Login with a relative, encoded return target for `/admin`; successful Author login returns once to `/admin` without a loop. |
| S1-08 | Anonymous Visitor, desktop | Attempt unsafe or looping return targets during Login. | External, scheme-relative, login/register, and recursive targets are rejected or replaced by the safe default; no open redirect or loop occurs. |
| S1-09 | Author, desktop and mobile | Resolve current user, use the single identity entry, and open the Author destination. | The identity entry opens `/admin`; the Admin workspace renders without a permanent Admin link in the public navigation. |
| S1-10 | Anonymous Visitor, desktop and mobile | Search from the global bar, including results, no-results, and API-error cases. | `/search` contains query/results/empty/error content only, has no second search input, and preserves actionable error text. |
| S1-11 | Author, desktop | Create a valid Directory from the existing Stage 1 Admin flow. | One success notice appears, the final returned URL Path is shown, the tree refreshes, and the returned node is selected/opened. No generic failure appears after success. Persisted server state matches the visible node and URL Path. |
| S1-12 | Author, desktop and mobile | Exercise validation, reserved root URL Path, conflict, expired-session, missing-parent, and offline create failures. | Each failure is truthful and actionable near the relevant action/field; raw generic or success-then-failure messaging is absent. The offline case offers retry and does not create a node. |
| S1-13 | Anonymous Visitor / Reader / Author, desktop and mobile | Regress public root/Directory/File navigation, Recent, published search, redirects, Assets, comments, and Likes using the preserved baseline fixture. | Existing public behavior remains intact; Draft data stays absent from public routes/search; Reader interactions still work; no unrelated Glass Ricepaper redesign is introduced. |
| S1-14 | Desktop `1440x900` and mobile `390x844` | Complete the applicable scenarios with trace, screenshot, console, and request capture. | Expected responsive state is visible; there are no unexpected console errors or failed network responses; evidence paths and the tested integration SHA are recorded; sessions close cleanly. |

## Static and API gates

Run from the clean integrated SHA:

```bash
conda run -n blogenv bash -lc \
  'cd api && CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./... && CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go vet ./... && test -z "$(gofmt -l .)"'

conda run -n blogenv bash -lc \
  'cd web && node --test tests/*.test.mjs && npm run lint && npm run build'
```

The frontend build is the TypeScript gate because `npm run build` executes `tsc --noEmit` before Vite.

Targeted Stage 1 evidence must identify tests for:

- precise validation/conflict/auth API errors while preserving `RequireAdmin`;
- current-user loading, invalid-token, and network-error state classification;
- role-aware navigation and safe Login return targets;
- exactly one global search input and no duplicate Search/Admin navigation;
- successful create using the returned node without entering a generic failure branch;
- actionable create failure classification.

## Browser commands

Use the existing external `playwright-cli`; do not add a repository Playwright dependency.

Desktop evidence:

```bash
playwright-cli -s=stage-1-desktop open http://127.0.0.1:5173
playwright-cli -s=stage-1-desktop resize 1440 900
playwright-cli -s=stage-1-desktop tracing-start
# Execute S1-01 through S1-13 as applicable with snapshot/fill/click/route/eval.
playwright-cli -s=stage-1-desktop console error
playwright-cli -s=stage-1-desktop requests
playwright-cli -s=stage-1-desktop screenshot
playwright-cli -s=stage-1-desktop tracing-stop
playwright-cli -s=stage-1-desktop close
```

Mobile evidence:

```bash
playwright-cli -s=stage-1-mobile open http://127.0.0.1:5173
playwright-cli -s=stage-1-mobile resize 390 844
playwright-cli -s=stage-1-mobile tracing-start
# Repeat role/navigation/search/create/public-regression checks that have mobile UI.
playwright-cli -s=stage-1-mobile console error
playwright-cli -s=stage-1-mobile requests
playwright-cli -s=stage-1-mobile screenshot
playwright-cli -s=stage-1-mobile tracing-stop
playwright-cli -s=stage-1-mobile close
```

Use `route` or `network-state-set offline` for controlled network failures, then restore the route/network state before continuing. Inspect every expected non-2xx request separately so an intentional validation/auth response is not confused with an unexpected transport or server failure.

## Evidence record

Complete this section during integrated acceptance:

- Integration SHA:
- Worktree reset command/output:
- Service health output:
- Version output:
- Backend tests/vet/gofmt:
- Frontend tests/lint/typecheck/build:
- Targeted Stage 1 tests:
- Desktop trace:
- Desktop screenshot(s):
- Desktop console result:
- Desktop request result:
- Mobile trace:
- Mobile screenshot(s):
- Mobile console result:
- Mobile request result:
- Created acceptance record IDs and cleanup result:
- Tested:
- Not tested:
- Findings:

## Pass criteria

Verdict may change to `PASS` only when:

1. S1-01 through S1-14 pass against the recorded integrated SHA.
2. Expected visible state agrees with persisted API/database state.
3. Backend tests, vet, and formatting pass.
4. Frontend tests, lint, TypeScript check, and production build pass.
5. No unexpected browser console error or failed network response remains.
6. Desktop and mobile screenshot/trace paths are durable and recorded.
7. Public reading, search, comments, Likes, Assets, redirects, and Draft isolation have no regression.
8. Acceptance data cleanup uses only recorded explicit IDs and preserves unrelated fixtures.

If a product behavior fails, complete the gate with `Verdict: FAIL`, record exact reproduction and evidence, and request a new development fix plus acceptance-retest task. Do not implement the product fix in this acceptance worktree.

## Preparation result

Tested: acceptance matrix structure, locked Stage 1 requirements, exact desktop/mobile dimensions, static gate commands, and browser evidence contract.

Not tested: Stage 1 product behavior, integrated runtime, persistent state, screenshots/traces, console/network results, and final verdict; these require the leader-supplied integrated SHA.
