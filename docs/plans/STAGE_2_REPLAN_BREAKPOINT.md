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

## Next question to ask

### Q13 — Settings danger and advanced operation layering

Ask the user:

> 关于「设置」里的危险操作和高级操作，Stage 2 里你希望如何分层？
>
> A. 普通设置和危险操作同页，但明显分区：基础信息、位置、危险操作。危险操作在底部红色弱化区域，需要二次确认。
>
> B. 危险操作单独页面：设置只显示基础信息和位置，删除/撤回进入单独「危险操作」页。
>
> C. Stage 2 不做删除，只做编辑、发布、撤回。
>
> 我的推荐答案：A。理由：删除和撤回是 Author 管理内容需要的低频操作；当前系统已有删除能力；同页分区比单独页面简单，但必须用中文二次确认，并阻止直接删除 Published File 与非空 Directory。

Wait for the user's answer before continuing.

## Current confidence

Estimated shared understanding: about 70–75%. More questions are still needed before drafting the final Stage 2 plan.
