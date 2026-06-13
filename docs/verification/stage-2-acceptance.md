# Stage 2 Acceptance

Status: Gateway 1 fixture ready; integrated acceptance pending

Verdict: **PENDING** for Gateway 6 integrated desktop/mobile acceptance.

Current evidence baseline: Gateway 1 fixture is ready and recorded in
`docs/verification/stage-2-backup-and-fixture.md`.

## Fixture for downstream acceptance

Use the recorded fixture root and IDs:

```text
/stage-2-acceptance
root:          77473f2e-6069-48ff-95a7-3d7173d090c4
draft branch:  f2f3fb74-33f0-4264-baf2-b26e5d06e83e
draft file:    5b796f40-e15a-42fa-8832-9cfbd1dcd21e
published file:a260a9f7-ecf3-4c3d-a87e-cc96b44c73bc
```

Public smoke at Gateway 1:

- `/stage-2-acceptance/published-fixture` resolves as a published File.
- `/stage-2-acceptance/draft-branch/draft-fixture` returns HTTP 404 publicly.

## Integrated acceptance checklist (pending)

To be run only after backend/frontend Stage 2 work is integrated and worker-4 is
reset to the leader-provided integrated SHA:

1. Author login → Chinese Author Workspace.
2. Create Directory/File using minimal forms.
3. Tree refreshes, expands parent, selects/opens new node, shows Chinese toast/path.
4. Edit File, manual save, publish.
5. Public File opens; `编辑文件` returns to workspace with File selected.
6. Public Directory shows Author-only `管理此目录` and returns to selected Directory.
7. Anonymous Visitor and Reader do not see `编辑文件` or `管理此目录`.
8. `撤回发布` hides public File after confirmation.
9. Settings move/URL Path/delete-constrained scenarios show clear Chinese prompts,
   including non-empty Directory deletion blocked even for draft-only children.
10. Same-parent drag reorder persists and never reparents.
11. Mobile no-regression sanity at `390x844`.
12. Public homepage/Recent/public reading/comments/Likes not redesigned.

## Required future evidence

- Integrated SHA.
- Backend and frontend gate outputs.
- Native API smoke output.
- Desktop screenshots/traces/console/network checks under `docs/verification/stage-2-browser-<date>/`.
- Mobile screenshot/console/network sanity evidence.
- PASS/FAIL verdict with any repair tasks, if needed.
