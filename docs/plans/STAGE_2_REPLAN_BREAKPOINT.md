# Stage 2 Replan Breakpoint — Author Workspace Interview

Date: 2026-06-13 CST

This is the durable breakpoint for the Stage 2 replanning interview. The interview is not complete; resume by asking **Question 13** below, one question at a time, until there is at least 90% confidence in the user's intent.

## Context

Stage 1 engineering closeout passed internal verification, but user acceptance did not pass. The user reported four Author-facing blockers that must be corrected together with Stage 2:

1. Creating a Directory or File does not immediately refresh navigation / Content Tree surfaces, so creation looks like it failed until a manual refresh.
2. Author flows lack explicit in-app back/return controls and rely on Directory-name clicks or browser back.
3. Generated Files are not reachable/selectable enough in the Author surface, so edit/unpublish is hard. When an Author browses any public File or Directory, there must be an obvious way to enter the corresponding editing/management view.
4. The current Author-facing Admin UI is too complex and too implementation-shaped; Stage 2 needs a graphical Chinese Author Workspace with clearer prompts, hierarchy, typography, spacing, layout, and interaction flow.

## Terminology decision recorded

`docs/specs/CONTEXT.md` now defines **Author Workspace**:

- The Author-facing creation and management surface for the Content Tree, Files, assets, publication controls, and node settings.
- Product UI should describe it as the Author's workspace, not an admin console.
- Route may remain `/admin`; UI should not say `Admin / Tree Manager`.

## Confirmed decisions so far

### Q1 — Replace, do not patch, the current Admin page

Decision: **Yes.**

Stage 2 should replace the current form-heavy Admin page with a Chinese Author Workspace: left Content Tree, right contextual workspace, graphical creation/editing/publishing/moving/settings flows.

### Q2 — Workspace information architecture

Decision: **B. Desktop two-column; mobile single-column progressive flow.**

- Desktop: left complete Content Tree, right current workspace.
- Mobile: switch between Content Tree and workspace with explicit return controls.

### Q3 — Chinese UI scope

Decision: **C.**

- Stage 2 fully localizes Author Workspace and Author-facing flow text to Chinese.
- It also establishes a Chinese UI copy convention.
- Public reading pages are not broadly redesigned in Stage 2, except Author-workflow touchpoints such as edit/manage entry points and login wording where necessary.

### Q4 — Author public-page edit/manage entry

Decision: **B.**

- When an Author views a public Directory, show an Author-only action such as `管理此目录`.
- When an Author views a public File, show an Author-only action such as `编辑文件`.
- Clicking enters the Author Workspace and automatically selects/opens the corresponding Directory or File.

### Q5 — Content Tree behavior

Decision: **A. Expand/collapse tree, no tree search in Stage 2.**

- Protected Author Workspace Content Tree includes all Directories, Draft Files, Published Files, and Files with unpublished changes.
- New creation refreshes the tree, expands the parent Directory, and selects the new node.
- Public-page edit/manage entry expands ancestors and selects the target node.
- No Admin tree search in Stage 2.

### Q6 — Create form complexity

Decision: **A. Minimal create flow.**

- New Directory: only `名称`.
- New File: `名称` and `格式`.
- URL Path is generated automatically and shown as a read-only preview.
- URL Path editing is low-frequency and belongs in Settings.
- Parent ID, Node ID, and Sort order are not part of the primary create flow.

User clarification: URL Path will rarely be changed. Sorting and location should be controlled graphically, especially by drag where safe.

### Q7 — Drag behavior

Decision: **A. Same-parent drag sorting only.**

- Desktop supports dragging sibling Directory/File nodes only to change order within the same Directory.
- Drag never changes parent Directory.
- Cross-Directory move uses Settings → graphical Directory Picker.
- Mobile uses explicit move up/down controls.

### Q8 — Directory workspace default

Decision: **A. Directory overview + child cards + create entry.**

When selecting a Directory, the right workspace should show:

- Directory name and path;
- prominent new Directory / new File actions;
- child cards for current Directory contents;
- a Settings entry.

Left tree is for location; right workspace is for clear operation.

### Q9 — File workspace depth in Stage 2

Decision: **C. Build the full workspace shape, but keep saving manual.**

Stage 2 File workspace should establish:

- File header with name/status/path/view-public action/publication action;
- tabs or equivalent sections: Content, Assets, Settings;
- Content with body, keywords, and manual save;
- Assets upload/view/delete;
- Settings for Name, URL Path, move, delete.

Stage 2 does **not** implement autosave, Content Version history, Draft Preview, or Published Content snapshots. Those remain Stage 3.

### Q10 — Publish/unpublish control

Decision: **A. Single primary publication control; unpublish secondary.**

- Draft: status `草稿`, primary action `发布`.
- Published: status `已发布`; no redundant primary publish button.
- Published with saved unpublished changes: status `有未发布修改`, primary action `发布更新`.
- `撤回发布` is secondary / overflow / danger action.

### Q11 — Creation success feedback

Decision: **A. Lightweight toast + automatic node location.**

On successful create:

- refresh Author Workspace Content Tree;
- refresh relevant public navigation/drawer/cache surfaces as needed;
- expand parent Directory;
- select new node;
- open the appropriate Directory overview or File workspace;
- show a short Chinese toast, e.g. `已创建 Directory「研究」` or `已创建 File「第一篇笔记」`.

### Q12 — Return/back controls

Decision: **A + lightweight breadcrumbs.**

- Every right-workspace subflow that replaces the current workspace needs an explicit return button.
- Button labels must state the destination, e.g. `返回当前目录`, `返回文件内容`, `返回设置`, `返回 Content Tree` on mobile.
- A lightweight breadcrumb/path indicator helps orientation, but is not the only way back.
- Main flows must not depend on browser back.

### Q13 — Settings danger and advanced operation layering

Decision: **A. Ordinary settings and danger operations on the same Settings page, but clearly separated.**

Settings should have visible sections such as:

- 基础信息: Name and URL Path;
- 位置: Move to Directory Picker and sorting guidance;
- 危险操作: Delete and File-only unpublish/danger actions where appropriate.

Danger actions live at the bottom in a visually distinct warning zone, require Chinese second confirmation, and must block direct deletion of Published Files and non-empty Directories with clear explanations.

### Q14 — URL Path generation for Chinese / mixed names

Decision: **A. Preserve Chinese characters in generated URL Paths.**

- Chinese characters are kept directly in URL Path segments.
- Latin text is normalized to lowercase hyphenated segments.
- Mixed names preserve readable Chinese while normalizing Latin spacing/case.
- Initial creation may append a numeric suffix to resolve same-parent conflicts.
- Explicit URL Path edits remain strict and should not be silently rewritten.

### Q15 — Author Workspace visual direction within Glass Ricepaper

Decision: **A with restrained B for critical states; overall minimal and operation-first.**

The Author Workspace should feel like a lightweight professional writing/management tool: quiet, sparse, readable, and easy to understand. Use extra card/status treatment only where it improves comprehension or prevents mistakes, such as creation success, publication state, save/error state, and danger confirmations. Avoid decorative complexity; every visual element must help operation or understanding.

### Q16 — Mobile Stage 2 acceptance depth

Decision: **C. Stage 2 product acceptance is desktop-first; mobile Author Workspace can be deferred.**

The user chose desktop-only Stage 2 acceptance for the Author Workspace. This narrows Stage 2 product scope, but it must be reconciled with the repository's existing verification rule that runtime/auth/tree/publication changes also run desktop/mobile browser acceptance. Treat mobile in Stage 2 as no-regression/responsive sanity unless the user later expands scope.

### Q17 — Mobile no-regression boundary

Decision: **A. Mobile gets no-regression sanity only in Stage 2.**

Stage 2 does not need a complete mobile Author workflow. Mobile acceptance should verify the Author Workspace does not visibly break at phone width: it opens, gives a usable orientation or basic Content Tree, has an exit/return path, and avoids major overflow/overlap. New/create/edit/move/delete full mobile flows are deferred.

### Q18 — Backend/API scope for complete Author Content Tree

Decision: **A. Stage 2 may add and reshape protected Author Workspace APIs.**

Stage 2 should introduce or complete protected APIs for the Author Workspace, such as Author Content Tree, node detail, create, reorder, move/impact preview, and delete constraints. OpenAPI must be updated first. Backend code must keep SQL in repositories, preserve clear service boundaries, and prioritize readability, extensibility, and a rigorous overall structure rather than quick UI-specific hacks.

### Q19 — Acceptance fixture and test content policy

Decision: **A. Use a dedicated Stage 2 acceptance fixture.**

Stage 2 verification should create or maintain a clearly named acceptance root such as `/stage-2-acceptance` with controlled Directories and Files covering Draft/Published state, Chinese URL Paths, sorting, move constraints, delete constraints, and Author public-page edit/manage entry. Back up the local database before fixture cleanup or schema migration. Prefer stable fixture data for reproducible browser acceptance, with explicit cleanup rules documented in verification evidence.

### Q20 — Stage 2 user acceptance definition

Decision: **A. Desktop complete Author workflow acceptance.**

Stage 2 user acceptance should follow this primary desktop path: log in as Author, enter the Chinese Author Workspace, create a Directory/File, see the Content Tree update immediately and select/open the new node, edit and manually save a File, publish it, open the public File and use the Author-only `编辑文件` entry to return to the workspace, unpublish it, and verify Settings move/sort/delete-constrained scenarios have clear Chinese prompts. Mobile gets no-regression sanity only.

## Current confidence

Estimated shared understanding: **90%+**. The revised Stage 2 plan can now be drafted from Q1–Q20 unless a new unresolved branch appears.


## Additional Stage 3 branch raised during Stage 2 replanning

### Q21 — Final-stage MCP direction

Decision: **A. Build an MCP Server for the Blog project in the final stage.**

The user wants the final stage to complete an MCP Server for this Blog so external AI tools can interact with Blog capabilities. The user also expects later development to be delegable to AI running autonomously on the server, with an explicit decision still needed about how much authority the MCP Server should expose. This is a new Stage 3/final-stage scope branch and must be planned with a clear permission model, auditability, and safe defaults before implementation.

### Q22 — MCP permission model

Decision: **B. High-trust MCP Server with full Author permissions.**

The user wants the Blog MCP Server to have complete Author-level capability because this is a personal blog, source code will remain on GitHub, and catastrophic content damage can be cleared and rebuilt. The MCP Server may therefore expose create, edit, publish, unpublish, delete, move, URL Path modification, asset operations, and search/tree read operations, subject to final implementation details.

Even with full permissions, the final-stage plan should still define minimal operational safeguards for presentation-quality engineering: explicit enabling configuration, audit logs, backup/recovery guidance, and a kill switch.

### Q23 — Minimum safeguards for full-permission MCP

Decision: **A. Full Author permissions with minimal engineering safeguards.**

The MCP Server should retain complete Author authority, but it must include presentation-quality operational safeguards: explicit enablement configuration, operation audit logs, automatic backup/export before destructive batches where practical, and an emergency disable/kill switch. It does not need per-operation human confirmation for publish/delete/move once full-permission MCP is enabled.

### Q24 — MCP architecture shape

Decision: **A. Independent MCP Server process/package.**

The final-stage MCP work should add a separate MCP server entry/package, with its own tools, auth/enabling config, audit logging, and emergency disable path. It should reuse backend service/API-client capabilities and must not duplicate business logic or direct SQL in ad hoc MCP handlers. This supports presentation, answerability, server-side autonomous AI use, and clean shutdown/operation boundaries.

### Q25 — MCP first tool set

Decision: **A. Complete Author tool set, implemented in prioritized slices.**

The final-stage MCP Server should expose a complete Author tool set by Stage 3 closeout, grouped into read, content, publish, tree, assets, and maintenance tools. Implementation should still proceed in tested slices: read/search first, then content creation/editing, then publish/unpublish, then tree move/reorder/delete, then assets, then maintenance/backup/search-index operations.

Initial tool inventory target:

- read: `list_content_tree`, `get_file`, `search_files`;
- content: `create_directory`, `create_file`, `update_file_content`, `update_file_settings`;
- publish: `publish_file`, `unpublish_file`;
- tree: `move_node`, `reorder_children`, `delete_node`;
- assets: `upload_asset`, `delete_asset`, `list_assets`;
- maintenance: `rebuild_search_index`, `export_backup`.

### Q26 — MCP deployment/autonomous runtime boundary

Decision: **A. Start with server-local stdio MCP for trusted AI.**

The final-stage MCP Server should initially run as a server-local stdio MCP process for trusted AI agents operating on the same server. It should not expose a public network listener in the initial final-stage scope. HTTP/SSE transport can remain a future extension behind explicit auth and network-binding controls. This matches the user's goal of delegating later development/content work to AI on the server while keeping deployment and security boundaries simple.

### MCP branch confidence

Estimated shared understanding for the final-stage MCP branch: **90%+**.
