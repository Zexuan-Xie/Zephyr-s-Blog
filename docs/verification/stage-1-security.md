# Stage 1 Security Verification

Status: pending

Stage: 1 — Reliability, navigation, and identity  
Security owner: worker-5 (independent; no feature implementation)  
Preparation source SHA: `f777238`  
Integrated SHA under test: pending leader handoff  
Verdict: pending

## Scope and invariants

This gate will test the leader-integrated Stage 1 SHA, not a developer worktree.
The required invariants are:

- Anonymous Visitors and Readers cannot call Author APIs or obtain Author-only data.
- Missing, malformed, expired, tampered, or deleted-user tokens do not authenticate.
- A current-user network failure is distinct from an invalid token and must not silently
  downgrade an authenticated user to Anonymous Visitor.
- Login and registration return targets remain same-origin application paths, cannot
  create redirect loops, and cannot send a Reader back into an Author-only loop.
- Public tree, resolve, recent, search, comments, and Asset responses disclose Published
  content only; Draft names, paths, bodies, keywords, snippets, and Assets remain private.
- Credentials and bearer tokens do not appear in application logs, browser-visible error
  text, committed files, screenshots, or captured request URLs.

## Threat matrix

| Boundary | Adversary / input | Abuse attempt | Required result | Evidence |
|---|---|---|---|---|
| `GET /api/auth/me` | Anonymous | Omit `Authorization` | `401`; no user data | API matrix |
| `GET /api/auth/me` | Token attacker | Empty bearer, wrong scheme, malformed JWT, bad signature, expired JWT | `401`; generic sanitized error | Go tests + API matrix |
| `GET /api/auth/me` | Stale identity | Validly signed token whose subject no longer exists | `401`; token cleared by client only for an authentication failure | Go test + browser |
| `/api/admin/**` | Anonymous | Call every Stage 1 Author route directly | `401`; no mutation or Draft response | API matrix |
| `/api/admin/**` | Reader | Replay Reader JWT against every Author route and alternate methods | `403`; Reader session remains valid | API matrix |
| `/api/admin/**` | Claim forger | Alter JWT role/email without re-signing; try unsupported signing method | `401`; database-loaded role remains authoritative | Go tests |
| Current-user state | Network failure | Offline, timeout, or `5xx` during `/auth/me` | Retry/error state; not Anonymous and no token destruction | browser |
| Current-user state | Invalid/expired token | `/auth/me` returns `401` | Token removed; identity becomes Anonymous without loop | browser |
| `/admin` | Reader | Navigate directly or use `return_to=/admin` after login | `Author access required`; no logout and no Login loop | browser |
| Login/register return target | Redirect attacker | Absolute URL, protocol-relative URL, encoded URL, backslash form, control characters | Reject or replace with safe local default | frontend test + browser |
| Login/register return target | Loop attacker | `/login`, `/register`, nested return target, or Reader-only `/admin` target | No redirect cycle; role-appropriate safe destination | frontend test + browser |
| Public tree/resolve/recent | Anonymous or Reader | Guess Draft path, node ID, parent path, or old redirect | `404`/empty; no Draft metadata or redirect | API/DB fixture matrix |
| Public search | Anonymous or Reader | Search unique Draft name, body, keyword, path, or embedding-only term | Zero Draft result and no Draft snippet | API/DB fixture matrix |
| Public Assets | Anonymous or Reader | Guess Draft Asset UUID and filename; vary filename encoding/path separators | `404`; no bytes, metadata, or storage path | API matrix |
| Public comments/Likes | Anonymous | Mutating request without a token | `401`; no state change | API regression |
| Public comments/Likes | Reader | Target Draft File UUID directly | Rejected; no existence oracle beyond normal not-found behavior | API/DB fixture matrix |
| Logging and errors | Any remote client | Send credentials/token-like strings in invalid requests | No secret echoed to response or server logs | sanitized log review |
| Browser storage/navigation | Script/content attacker | Render malicious Markdown/HTML while a token exists | Sanitized Markdown; HTML stays in `sandbox="allow-scripts"` without same-origin access | existing safety test + browser |

## Abuse cases

1. Issue Author and Reader tokens, then exercise the same Admin endpoint table with no
   token, Reader token, Author token, expired token, and signature-tampered token.
2. Change only the JWT role claim from Reader to Author and verify signature validation
   rejects it; separately verify the middleware loads the current database user instead
   of trusting the role claim for authorization.
3. Delete or disable the token subject after issuance and verify `/auth/me` and Admin
   routes reject the stale identity.
4. Force `/auth/me` offline and to `500`, then verify the UI presents retryable failure
   without clearing the token or presenting Anonymous state.
5. Submit `return_to` values for `https://attacker.example`, `//attacker.example`,
   encoded external URLs, `/login`, `/register`, and `/admin`; verify same-origin,
   non-looping, role-correct navigation.
6. Seed a Draft File with unique name, path, keyword, body marker, embedding marker, and
   Asset filename. Probe tree, resolve, recent, full-text search, semantic search, direct
   Asset URL, comments, and Likes for every marker/ID as Anonymous and Reader.
7. Confirm an Author can still use the intended Stage 1 Admin create flow, while the same
   request body with Anonymous/Reader credentials cannot create or disclose a node.
8. Inspect browser console/network capture and sanitized server logs after failed login,
   invalid JWT, Admin denial, and network failure for passwords, JWTs, storage keys, and
   Draft content.

## Baseline observations to retest after integration

These observations describe preparation SHA `f777238`; they are not the Stage 1 verdict.

- `web/src/lib/auth.ts` accepts `return_to` without validation, so an absolute URL is a
  candidate open-redirect finding until the integrated Stage 1 implementation proves a
  safe local-path policy.
- `web/src/pages/AdminPage.tsx` redirects every `/auth/me` error and every Reader to
  Login. This conflates network/auth/authorization states and can create a Reader Login
  loop. The Stage 1 contract requires separate retry, Anonymous, and
  `Author access required` states.
- Server Admin routes are grouped under `RequireAdmin`, and public tree/search/Asset SQL
  includes Published filters. The integrated gate must still prove these controls with
  real Draft fixtures and direct-ID probes rather than relying on source inspection.

## Planned verification

Run from a clean detached worktree reset to the leader-provided integration SHA:

```bash
git status --short
git checkout --detach <integration_sha>
(cd api && go test ./...)
(cd web && npm run build)
(cd web && npm run lint)
node --test web/tests/*.test.mjs
```

Then run the API role/token/Draft matrix against native services and execute desktop
`1440x900` plus mobile `390x844` browser scenarios with console, request, screenshot,
and trace capture. Record exact commands, observed status/body summaries, fixture IDs,
artifact paths, and sanitized log review here.

## Preparation evidence

- Threat matrix: ready.
- Independent abuse-case plan: ready.
- Final integrated-SHA API/browser execution: pending.
- Final security verdict: pending.

Tested:

- Source inspection of auth middleware/token validation, Admin route protection, public
  tree/search/Asset Published filters, frontend current-user handling, and return-target
  handling on preparation SHA `f777238`.

Not tested:

- Leader-integrated Stage 1 changes (integration SHA not yet provided).
- Live PostgreSQL Draft fixture matrix.
- Native API/browser/log-capture execution.
