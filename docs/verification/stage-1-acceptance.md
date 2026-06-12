# Stage 1 Acceptance

Status: complete

Verdict: **PASS**

Stage: Reliability, navigation, and identity

Integrated product SHA: `d6c7d092ae41b2e37bbf1c89ea30cff4ec551ef6`

Evidence directory: `docs/verification/stage-1-browser-20260612/`

## Decision

The native Stage 1 candidate passes required backend/frontend gates, PostgreSQL/API
smoke, desktop/mobile browser acceptance, role boundaries, navigation requirements,
Directory creation/error handling, safe return targets, public regressions, and test
data cleanup.

This is engineering acceptance only. Stage 1 is not user-accepted until the user
completes real use and explicitly accepts it.

## Environment

```text
Node.js 22.22.3
npm 10.9.8
Go 1.26.4
PostgreSQL 17.10
playwright-cli 0.1.11
```

The native API and Vite service were restored in persistent tmux session
`xlab-stack`.

Evidence:

- `stage-1-browser-20260612/environment-final.log`
- `stage-1-browser-20260612/service-recovery-final.log`

## Required static gates

Backend:

```text
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./...  PASS
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go vet ./...            PASS
test -z "$(gofmt -l .)"                                                PASS
```

Frontend:

```text
node --test tests/*.test.mjs  PASS (17 assertions across 3 files)
npm run lint                  PASS
npm run build                 PASS
```

Evidence:

- `stage-1-backend-gates-20260612.log`
- `stage-1-frontend-gates-20260612.log`
- `stage-1-browser-20260612/backend-gates-final.log`
- `stage-1-browser-20260612/frontend-gates-final.log`

## Native PostgreSQL/API smoke

Result:

```text
SMOKE_OK checks=21
```

Covered:

1. health/database readiness;
2. Author login;
3. Reader registration;
4. Reader 403 on Author creation;
5. Directory creation and returned URL Path;
6. persisted Directory detail;
7. duplicate URL Path conflict;
8. reserved root URL Path;
9. child File creation;
10. content save;
11. publish;
12. Anonymous Visitor resolution;
13. full-text fallback search;
14. Reader comment;
15. File Like;
16. idempotent repeated Like;
17. Asset upload;
18. immutable Published Asset read;
19. public comment thread;
20. unpublish isolation;
21. search exclusion after unpublish.

All disposable IDs were recorded and removed.

Evidence:

- `stage-1-native-api-smoke-20260612.log`

## Browser acceptance

### Anonymous Visitor

- Public root renders on desktop `1440x900` and mobile `390x844`.
- The global navigation has one search input, Recent, Directory, and one identity entry.
- There is no duplicate Search link or permanent Admin link.
- `/admin` reaches `/login?return_to=%2Fadmin`.
- Invalid-token evidence clears the token.
- Network-failure evidence preserves the session and offers Retry.

Evidence:

- `final-anonymous-desktop-home.png`
- `final-anonymous-mobile-home.png`
- `final-anonymous-admin-redirect.png`
- `invalid-token-cleared.png`
- `identity-network-failure.png`

### Reader

- Reader remains signed in when opening `/admin`.
- `Author access required` and `Return to Recent` render.
- No Author workspace or data renders.

Evidence:

- `reader-desktop-admin-denied.png`
- `reader-mobile-admin-denied.png`
- API smoke Reader 403 result.

### Author

- Public identity location displays `Author`.
- `/admin` renders the existing Stage 1 Tree Manager.
- Successful Directory creation shows the returned final URL Path, refreshes the tree,
  and opens/selects the returned Directory.
- Duplicate URL Path displays `This URL path is already in use`.
- Reserved URL Path displays `This URL path is reserved`.
- No success-then-generic-failure regression occurs.

Evidence:

- `final-author-admin-clean.png`
- `final-author-create-success.png`
- `final-author-create-conflict.png`
- `final-author-create-reserved.png`

### Search and public regressions

- Search is driven by the single global input.
- `/search` renders results without a second input.
- Published File reading, Recent, search fallback, Assets, comments, and Likes passed
  the 21-step API smoke.
- Draft/unpublished File resolution and search exclusion passed.
- No Glass Ricepaper redesign was introduced.

Evidence:

- `anonymous-desktop-search.png`
- `stage-1-native-api-smoke-20260612.log`

### Safe return targets

- Absolute external and backslash-form targets default to `/recent`.
- Final Chromium backslash test produced no browser error.

Evidence:

- `backslash-return-fixed.png`
- frontend return-target regression tests.

## Cleanup and recovery

Before removing older acceptance fixtures, the local database was backed up to:

```text
~/.local/share/xlab-blog/backups/pre-stage1-acceptance-cleanup-20260612T180837.dump
```

Cleanup used only explicitly recorded IDs. Verification after cleanup:

```text
remaining acceptance nodes = 0
remaining acceptance Reader = 0
final UI-created Directory remaining = 0
```

The persistent local stack remains available at:

```text
http://127.0.0.1:5173
```

## Remaining boundary

- Real DashScope embedding generation was not tested because no API key is configured.
  Full-text search fallback and missing-provider behavior remain the supported local path.
- Docker Compose/server deployment remains deferred until after user acceptance and
  production secret/upload hardening.
