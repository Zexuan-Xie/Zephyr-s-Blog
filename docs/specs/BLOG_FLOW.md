# BLOG_FLOW.md — 页面、路由与用户流

> 版本：2026-06-03
> 目的：固定用户如何在站点中移动，防止实现 agent 猜测交互。

## 1. 路由总览

| 路由 | 访问 | 页面 | 数据来源 |
|---|---|---|---|
| `/` | public | Root Directory Page | `GET /api/tree` |
| `/recent` | public | Recent Files Page | `GET /api/recent` |
| `/search?q=` | public | Search Page | `GET /api/search?q=` |
| `/login` | public | Login Page | `POST /api/auth/login` |
| `/register` | public | Register Page | `POST /api/auth/register` |
| `/admin` | admin | Admin Tree Manager | admin APIs |
| `/*` | public | Content Path Resolver | `GET /api/tree/resolve?path=...` |

Root reserved URL path segments：`admin`、`api`、`auth`、`login`、`register`、`recent`、`search`、`settings`。

## 2. 全局 Shell

所有 public 页面共享：

- glass nav
- search entry
- directory drawer trigger
- ZH/EN UI chrome toggle
- warm rice-paper background

File 页面额外显示：

- breadcrumb
- like button/count
- comment thread

Directory 页面显示：

- breadcrumb
- content-entry-card grid
- 不显示 like/comment

## 3. 首页 `/` flow

### 3.1 初始加载

1. 用户访问 `/`。
2. 前端调用 `GET /api/tree`。
3. API 返回 synthetic root directory page。
4. 页面展示可选 glass hero + root children cards。

### 3.2 成功状态

- Directory card：label `DIRECTORY`，显示 name、child directory/file counts、path。
- File card：label `FILE`，显示 name、path、updated/read meta。
- 卡片全部使用 `content-entry-card`。

### 3.3 错误状态

- API 5xx：显示 glass error panel 和 retry。
- 空 root：显示 empty state：`No files yet` / `暂无内容`。

## 4. Content path `/*` flow

### 4.1 当前路径解析

1. 用户访问 `/notes/go/auth`。
2. React catch-all route 调用：`GET /api/tree/resolve?path=/notes/go/auth`。
3. API 按当前 tree path 解析。
4. 若成功，返回 `directory` 或 `file`。
5. 若当前 path 不存在，查 `path_redirects.old_path`。
6. 若找到 redirect，返回 `{ type: "redirect", new_path }`。
7. 前端 `navigate(new_path, { replace: true })`。
8. 若都找不到，显示 404。

### 4.2 Directory result

显示：

- breadcrumb，每层可点击。
- child cards：Directory first，File second；同组 sort_order/name。
- Directory drawer 可呼出。
- 无评论/点赞。

### 4.3 File result — Markdown

显示：

1. breadcrumb。
2. `file-reading-card`。
3. keywords chips 最多 3 个；点击进入 `/search?q=keyword`。
4. title/name。
5. meta：path、published/updated time、read time。
6. sanitized Markdown HTML。
7. like/comment bar。
8. comment thread。

### 4.4 File result — HTML Document

显示：

1. breadcrumb。
2. title/name + up to 3 keywords。
3. iframe container inside blog shell。
4. iframe attributes exactly include `sandbox="allow-scripts"` and must not include `allow-same-origin`。
5. desktop height: `min-height: 640px; height: 78vh`。
6. mobile height: `height: 72vh`。
7. like/comment bar and comment thread below/around shell。

### 4.5 404 状态

- 显示 glass 404 panel。
- 提供返回 root、打开 directory drawer、搜索。

## 5. Directory Drawer flow

### 5.1 打开

触发：nav 中目录按钮。

行为：

- 左侧 overlay drawer。
- desktop width ~320px。
- mobile width `min(88vw, 360px)`。
- warm lightweight scrim。
- 不推开主内容。

### 5.2 浏览

- 显示完整 directory tree。
- 当前 path 用 Action Blue。
- 节点缩进显示层级。
- 点击 Directory/File 后 navigate 对应 path 并关闭 drawer。

### 5.3 关闭

关闭方式：

- 点击 scrim。
- Esc。
- close pill。
- 导航成功后自动关闭。

## 6. `/recent` flow

### 6.1 访问

- public 可访问。
- 登录成功且无 return target 时默认跳转。

### 6.2 数据

调用 `GET /api/recent?limit=&offset=`。

排序：published File 按 updated/published desc。

### 6.3 UI

- 显示 File cards only。
- 每张卡必须显示完整 path。
- 不显示 Directory。
- 无 format filter；Markdown/HTML 都是 File。

### 6.4 错误/空状态

- 空：`No recent files`。
- API error：glass error panel + retry。

## 7. `/search` flow

### 7.1 初始状态

- 顶部 nav search 是全站唯一搜索输入。
- 无 query：显示简洁提示，不自动搜索，不重复显示第二个 search input。

### 7.2 搜索提交

1. 用户在顶部 nav search 输入 query。
2. URL 更新 `/search?q=...`。
3. 调用 `GET /api/search?q=...`。
4. 显示 loading skeleton。
5. 页面显示当前 query 和结果，不重复显示 search input。

### 7.3 搜索结果 card

每条结果：

- label `FILE`。
- File name。
- full current path。
- snippet。
- match source badges：`text`、`semantic`、`keyword`。
- 点击 current path。

### 7.4 错误状态

- Qwen embedding unavailable 不应导致整体搜索失败；显示 full-text results。
- API 5xx 显示 retry。

## 8. Auth flow

### 8.1 Register

1. 用户访问 `/register`。
2. 输入 email/password/display_name。
3. `POST /api/auth/register`。
4. 成功后获得 token 和 reader user。
5. 跳转 return target 或 `/recent`。

错误：

- email exists → inline error。
- weak password → inline error。

### 8.2 Login

1. 用户访问 `/login`。
2. 输入 email/password。
3. `POST /api/auth/login`。
4. 成功保存 JWT。
5. 若 URL/state 有 return target，优先跳回原目标。
6. 主动从顶部 Login 且无 return target 时，Reader 进入 `/recent`，Author 进入 `/admin`。
7. Reader 的 return target 若为 `/admin` 或 Draft Preview，则显示 `Author access required`，不循环跳转 Login。

错误：

- 401 → inline error，不跳转。

### 8.3 Protected interaction redirect

匿名点击 like/comment/reply：

1. 保存 intended action context。
2. navigate `/login?return_to=current_path`。
3. 登录成功后返回 current_path。
4. UI 聚焦评论框或重试 like intent；若无法重试，显示提示。

### 8.4 Reader visits admin

已登录 Reader 直接访问 `/admin`：

1. 保持 Reader 的登录状态，不清除 token，也不跳转 Login。
2. 显示简洁权限页：`Author access required`。
3. 提供 `Return to Recent`。
4. 所有 admin API 仍由后端返回 `403 Forbidden`。

### 8.5 Single identity entry

顶部只显示一个随身份变化的入口：

- Anonymous Visitor：`Login`，点击进入 `/login`。
- Anonymous Visitor 不显示任何 `Admin` 或 `Author` 入口；`/admin` 的直接访问仍跳转 Login。
- Reader：显示 Reader 的 `display_name`；点击打开仅包含 `Logout` 的极简菜单。
- Author：显示 `Author`；点击直接进入 `/admin`。
- Author 的 `Logout` 位于 Admin Tree Manager 内，不在顶部增加第二层入口。
- 桌面与移动端保持相同语义。
- 身份尚未确认时显示固定尺寸的无文字骨架占位，不先显示 Login 再闪烁为 Reader/Author。
- Token 无效时清除本地身份并显示 Login。
- 身份检查因网络失败时显示可重试状态，不直接误判为 Anonymous Visitor。
- Reader Logout 后留在当前公开页面，并立即恢复 Anonymous Visitor 的交互权限。
- Author 从 Admin Logout 前先立即保存待处理内容；失败时阻止退出并提供 `Try again` / `Discard and logout`。
- Author Logout 成功后跳转网站首页。
- Draft Preview 在 Logout 后立即失去访问权限并跳转 Login。

## 9. Comment flow

### 9.1 Anonymous

- 可读 comments。
- comment input 显示 `Log in to comment`。
- like/comment actions redirect login。

### 9.2 Reader/Admin create top-level comment

1. 输入纯文本。
2. `POST /api/files/{file_id}/comments` with body。
3. 成功插入 thread。
4. 清空输入。

错误：

- 401 → login redirect。
- empty/too long → inline validation。

### 9.3 Reply

1. 用户点击 reply。
2. 输入框显示 reply target。
3. 提交 `parent_id` 和可选 `reply_to_user_id`。
4. 若 reply-to-reply，后端归一到 top-level parent。

### 9.4 Delete

- 用户可删除自己的评论；admin 可删除任何评论。
- 删除为软删除。
- UI 显示 `该评论已删除` / `This comment has been deleted`。
- 子回复保留。

## 10. Like flow

### 10.1 Like File

- `PUT /api/files/{file_id}/like`。
- 重复调用仍 liked=true。
- `DELETE` 取消，重复调用仍 liked=false。

### 10.2 Like Comment

- 同 File，但 endpoint 为 `/api/comments/{comment_id}/like`。
- deleted comment 不允许新增 like；UI 不展示 deleted comment like count。

## 11. Admin Tree Manager flow

### 11.1 Layout

- Tree browser + selected node editor。
- 移动端可退化为 list → edit page。
- Admin 不显示大型 `ADMIN / Tree Manager` 介绍卡，进入后直接显示工作区。
- Admin 顶部工具栏左侧显示 `Content` 与当前节点 path。
- 右侧显示 `View site` 与页面级 `···` 菜单。
- 页面级菜单包含 `Rebuild search`、`System status`、`Logout`。
- 未选中节点时，右侧只显示 `Select a Directory or File`、一句简短说明和 `＋ New`。
- 空状态不显示统计仪表盘或使用教程；其 `＋ New` 与树顶部入口相同。
- `View site` 始终在新标签页打开，不离开当前 Admin 编辑状态。
- 选中 Published File 时打开其公开 File；选中公开 Directory 时打开该 Directory。
- 选中 Draft File 时文案为 `Preview draft`，打开仅 Author 可访问的预览。
- 未选中节点时打开网站首页。
- Draft Preview 固定使用 `/admin/preview/{file_id}`，不复用公开 URL path。
- Anonymous Visitor 访问 Draft Preview 时跳转 Login；Reader 显示 `Author access required`。
- 页面显示克制但明确的 `Draft preview` 标识，渲染 Current 自动保存内容。
- Draft assets 继续要求 Author 权限；不生成可分享的临时公开链接。
- 已打开的 Draft Preview 只在编辑器成功自动保存后更新，不逐字符同步。
- 保存期间继续显示上一份已保存内容；保存失败时 Preview 不变化。
- 同源标签页使用跨标签通信触发更新，不通过持续轮询；Preview 提供手动 `Refresh` 兜底。
- Content Tree 顶部提供唯一的 `＋ New` 创建入口。
- 点击 `＋ New` 后先显示两个大型图形类型卡片：
  - `Directory` — Organize files
  - `File` — Create content
- `＋ New` 流程替换右侧工作区，不使用 Modal；左侧 Content Tree 保持可见。
- 选择类型后进入简短创建面板，并以可读 path 显示 parent context，例如 `Create in /research/notes`。
- 创建期间切换左侧选中的 Directory 会同步更新 parent context。
- 切换 parent 时保留已输入的 Name 和 File Format，更新路径预览，并轻量提示 `Creation location updated`。
- 切换 parent 不弹确认框，因为尚未产生持久化数据或输入丢失。
- `Cancel` 返回此前的节点工作区。
- 移动端进入全宽创建视图，返回后回到 Content Tree。
- 创建主流程不显示 Parent ID。
- 离开 New workspace 时，仅当 Name 已填写且尚未创建才提示 `Discard this new item?`，操作为 `Keep editing` / `Discard`。
- Name 为空时，即使已选择 Directory/File 类型，也可直接退出。
- 单击节点时选中并打开；Directory 使用独立箭头展开或折叠。
- Admin Content Tree 显示所有 Directory、Draft File 与 Published File，包括只包含 Draft 的 Directory。
- File 显示轻量状态：`Draft`、`Published` 或 `Unpublished changes`。
- Draft：灰色空心圆 + `Draft`。
- Published：绿色小圆点，不重复显示状态文字。
- Unpublished changes：琥珀色小圆点 + `Changes`。
- Save failed：红色小圆点 + `Save failed`。
- 状态具有文字或 ARIA 描述，不只依赖颜色。
- 折叠 Directory 只汇总需要处理的后代状态：存在 Save failed 时显示红点，否则存在 Unpublished changes 时显示琥珀点。
- 仅包含 Draft/Published 时不显示 Directory 汇总点；不显示后代状态数量。
- 子节点按需加载，不一次性加载完整大树。
- Admin Content Tree 使用独立的受保护管理 API，不能用过滤 Draft 的 Public Tree 代替。
- 重新进入 `/admin` 时恢复本浏览器上次选中的节点及 Directory 展开状态。
- 上次节点已删除时，选择最近仍存在的 parent Directory；首次进入只显示根级，不展开全部。
- Admin tree navigation state 不跨设备同步。
- Admin Content Tree 不提供独立节点搜索或 `Find in tree`。
- 全站公开搜索只检索 Published File；Draft 不进入公开搜索。
- Author 仅通过 Admin Content Tree 查看和定位 Draft。
- 仅在节点选中时显示 `···` 操作菜单，避免常驻操作按钮造成视觉噪音。
- Directory 菜单：`New inside`、`Advanced settings`。
- File 菜单：`Open editor`、`Advanced settings`。
- `Advanced settings` 是当前节点的右侧工作区视图：通过选中节点的 `···` 菜单打开，不使用弹窗或独立路由。
- 保存或取消后返回该节点原来的 Directory/File 工作区。
- `Advanced settings` 包含 Name/Rename、URL path、Move to、Sort position、Delete，以及默认折叠的只读 `Technical details`。
- File 内容、Keywords、Publish/Unpublish 不进入 `Advanced settings`。
- Delete 位于 `Advanced settings` 最底部独立的 Danger zone。
- Draft File 或空 Directory 点击 Delete 后展开内联确认，并要求再次点击 `Delete permanently`。
- Published File 禁止直接删除并提示先 Unpublish。
- 非空 Directory 禁止删除，显示其包含的 Directory/File 数量，不提供递归删除。
- 删除确认不要求输入节点名称。
- Danger zone 在删除前显示会失效的历史 Redirect 数量和 404 影响，并可展开查看旧路径列表。
- Redirect 影响不增加第三次确认；空 Directory 只统计其自身 Redirect。
- 桌面端允许同一 parent 下的 Directory/File 通过拖拽调整 sort order。
- 同级 Directory 与 File 共用一个混合排序序列，不按类型强制分组。
- 拖拽排序不改变 parent，不支持把节点拖入另一个 Directory。
- 跨 Directory 移动仍通过 `Advanced settings` 完成。
- 移动端使用 `Move up` / `Move down`，不要求触屏拖拽。
- 拖拽松开或移动端排序操作后立即自动保存，不增加 `Save order` 按钮。
- 保存期间显示轻量 `Saving…` 状态，成功后提示 `Order updated`。
- 保存失败时恢复操作前顺序并提示 `Could not update order`。
- `Advanced settings → Move to` 使用可展开的图形化 Directory Picker，不显示 Parent ID。
- 当前 parent 高亮；选择目标 Directory 后显示新路径预览，点击 `Move` 后才执行。
- Directory 不能移动到自身或后代。
- Published File 移动前明确提示旧路径将创建 redirect。

### 11.2 Create Directory

1. 在 Content Tree 中选择 parent Directory；创建流程不要求输入或复制 Node ID。
2. 通过图形化创建入口选择 `Directory`。
3. 简短创建面板只包含 `Name`。
4. 系统根据 Name 自动生成 URL path 的最后一段，sort order 默认放在末尾。
5. 如果生成的 URL path 在同一 parent 下冲突，系统自动建议带数字的可用路径，例如 `research-2`，而不是直接让创建失败。
6. URL path 与 sort order 仅在创建后的 `Advanced settings` 中编辑。
7. `POST /api/admin/nodes` kind=directory。
8. 成功后显示明确成功反馈、刷新 Content Tree、自动选中并进入新 Directory。
9. 创建界面清空，但保留合理的 parent context。
10. root reserved URL path 命中时显示具体错误。

产品 UI 统一使用 `URL path`，不向 Author 显示 `Slug` 一词。内部自动 slug 规则同时适用于 Directory 与 File：

- 中文字符直接保留。
- 英文转为小写。
- 空格与连续分隔符规范为单个 `-`。
- 中英混合示例：`Go 并发笔记` → `go-并发笔记`。
- 不做拼音音译。
- URL 编码由浏览器和路由层处理。
- Author 可在创建后的 `Advanced settings` 修改 URL path 的最后一段。

创建面板实时显示只读最终路径预览：

- 示例：`Will be created at /research/research-notes`。
- 同一 parent 下发生 URL path 冲突时，预览直接显示自动编号后的可用路径，例如 `/research/research-notes-2`。
- 路径预览不是额外输入字段。
- 前端路径预览是即时预期值，后端在创建事务中重新执行相同内部路径规范并决定最终唯一 URL path。
- 并发创建导致预览过期时，后端可选择下一个编号路径并成功创建；Toast 显示最终 URL path。
- 前端不得依赖“先查询可用性再创建”保证唯一性。
- 自动数字后缀仅用于首次创建。
- Author 在 Advanced settings 明确修改 URL path 时，后端严格采用输入；冲突则阻止保存并提示 `This URL path is already in use.`。
- 系统可建议可用 URL path，但只有 Author 选择建议或自行修改后才保存，不静默改写。
- 节点创建后，Name 与 URL path 解耦。
- Rename 只修改显示名称，不改变 URL path，也不创建 redirect。
- URL 变化必须由 Author 在 Advanced settings 单独修改 URL path。
- 修改 Directory URL path 会自动重写所有后代路径。
- 保存前显示受影响数量和若干旧路径 → 新路径示例。
- 只为修改前公开可访问的旧路径创建 redirect；Draft 路径不创建公开 redirect。
- 子树路径重写与 redirect 持久化必须原子完成；任一目标冲突则全部回滚。
- Directory 跨 parent 移动采用完全相同的子树路径、影响预览、公开 redirect、Draft 排除和原子回滚规则。
- Directory 不能移动到自身或其后代。
- 曾公开可访问的 Directory 旧路径与其 Published File 后代旧路径都创建 redirect。
- 只包含 Draft 且从未公开可见的 Directory/路径不创建公开 redirect。
- Redirect 默认永久保留并自动压平到当前路径，不允许形成 chain 或 loop。
- Redirect 仅在 `System status → Redirects` 只读查看，首版不提供手动删除。
- 目标节点删除后，其历史 Redirect 失效并返回 404。

### 11.3 Create File

1. 在 Content Tree 中选择 parent Directory；创建流程不要求输入或复制 Node ID。
2. 通过图形化创建入口选择 `File`。
3. 简短创建面板只包含 `Name` 与图形化 `Format` 选择：`Markdown` 或 `HTML Document`。
4. 系统根据 Name 自动生成 URL path 的最后一段，sort order 默认放在末尾，新 File 默认 draft。
5. URL path 与 sort order 仅在创建后的 `Advanced settings` 中编辑。
6. 成功后显示明确成功反馈、刷新 Content Tree、自动选中新 File 并进入 editor。

Directory/File 创建失败时只显示后端实际失败原因，不得把成功后的前端状态异常误报为创建失败。

Admin 写操作错误使用面向 Author 的可操作文案，并尽量显示在对应字段或操作附近：

- URL path 冲突：`This path is already in use. Try research-2.`
- reserved root URL path：`“admin” is reserved. Choose another name.`
- session 失效：`Your session expired. Log in again.`
- parent 不存在：`The destination Directory no longer exists.`
- 网络错误：`Could not save changes. Check the connection and try again.`

服务器原始错误与 HTTP status 仅放在默认折叠的 `Technical details` 中。

成功反馈保持轻量：

- 使用约 3 秒自动消失的 Toast，例如 `Directory created`。
- 同时通过刷新 Content Tree、选中新节点和切换右侧工作区表达结果。
- 页面顶部不长期保留成功文字。
- 连续操作只保留最新一条 Toast。

### 11.4 Edit File

- File workspace 顶部提供 `Content`、`Assets`、`Settings` 三个标签，默认进入 `Content`。
- `Content` 包含正文、Keywords、Preview、Content Version 与 Publish。
- `Assets` 包含上传、文件列表、复制链接与删除。
- `Settings` 打开该节点的 Advanced settings。
- Assets 不长期堆叠在正文编辑器下方。
- Published File 新上传的 Asset 先保持 Draft，仅供 Current 内容和 Draft Preview 使用。
- `Publish changes` 后 Asset 才可公开；已有 Published Assets 继续可访问。
- 从 Current 内容移除引用不会自动删除 Asset。
- 删除 Published Asset 前必须确认 Published Content 已不再引用它，否则阻止删除并提示先发布解除引用的内容。
- `Publish changes` 只公开 Current 内容实际引用的 Draft Assets。
- Markdown 从正文 Asset URL 识别引用；HTML Document 从 HTML/CSS Asset URL 识别引用。
- 未引用的 Draft Assets 保持私有；发布确认显示将公开与仍保持私有的 Asset 数量。

现有数据迁移：

- Published File：Current 与 Published Content 均从现有内容初始化，Previous 为空。
- Draft File：只初始化 Current，Previous 与 Published Content 为空。
- Existing Assets 根据 File 发布状态和 Published Content 实际引用迁移为 Draft/Published。
- 迁移前备份数据库，迁移在事务中完成，失败时不留下部分状态。
- Markdown：Markdown editor + article preview。
- HTML Document：full HTML editor + sandbox iframe preview。
- keywords editor。
- File 内容与 Keywords 自动保存，不提供主要的手动 Save 按钮。
- 每个 File 只保留两个内容状态：Current version 与 Previous version。
- File editor 提供恢复 Previous version 的能力。
- 停止输入 15 秒后自动保存。
- 编辑器失焦、切换节点、Publish 或离开 Admin 前立即保存。
- 保存期间显示 `Saving…`，成功后显示 `Saved`。
- 只有一次完整保存成功后，原 Current 才成为 Previous。
- 内容与 Keywords 没有实际变化时不创建新版本。
- 恢复 Previous version 时原子交换 Current 与 Previous，因此恢复操作可再次撤销。
- 恢复前显示版本时间与内容差异摘要，并要求一次确认。
- 内容版本快照包含正文、Keywords 与 Render Format。
- 内容版本不包含 Name、URL Path、Parent Directory、Sort position、Publish 状态、Assets、Comments 或 Likes。
- 编辑 Published File 时，自动保存只更新 Author 的 Draft 内容，不立即改变公开页面。
- 公开页面继续显示最近一次 Published 内容；编辑器显示 `Unpublished changes`。
- 首次公开使用 `Publish`；已有 Published 内容的更新使用 `Publish changes`。
- `Publish` / `Publish changes` 前立即完成待处理自动保存，然后才更新公开快照。
- Restore 只交换 Author 的 Current/Previous Content Version，不自动改变 Published Content。
- Published File 恢复后通常显示 `Unpublished changes`，由 Author 检查后执行 `Publish changes`。
- 如果恢复后的 Current 与 Published Content 完全相同，则不显示 `Unpublished changes`。
- Unpublish 只改变公开可见性，不删除最后一次 Published Content 快照。
- 重新 Publish 默认发布当前 Draft；若当前 Draft 与保留的 Published Content 不同，先显示差异提示。
- Published Content 独立于 Author 的 Current/Previous 两版编辑历史。
- 每个 File 最多保留三份内容状态：Current、Previous、Published Content。
- “只保留最新版和上一版”仅约束 Author 编辑历史，不允许历史轮换删除公开快照。
- 切换节点、离开 Admin、Publish 前的立即保存如果失败，则阻止当前导航或操作。
- 保留编辑器内容并显示 `Could not save your changes`，提供 `Try again` 与 `Discard changes`。
- 不允许保存失败后静默离开；浏览器刷新或关闭使用系统离开确认。
- File editor 加载时记录内容版本号，自动保存时携带该版本号。
- 服务器版本已变化时拒绝覆盖并提示 `This File was updated elsewhere.`。
- 冲突操作仅提供 `Reload latest` 与 `Copy my changes`；不自动合并 Markdown 或 HTML。
- File editor 标题栏右侧固定显示保存/发布状态：`Saved`、`Editing…`、`Saving…`、`Unpublished changes`、`Save failed`、`Conflict`。
- 同一位置提供 `Version history`，展开后显示 Current/Previous 保存时间、`Compare` 与 `Restore previous`。
- 写作期间的正常保存状态不使用频繁 Toast。
- 桌面 File editor 默认使用可调宽度双栏：Editor 约 55%，Preview 约 45%。
- 桌面提供 `Editor only`、`Split`、`Preview only` 三种视图，中间分隔条可拖动。
- 移动端使用 `Edit` / `Preview` 切换，默认 `Edit`，不并排。
- Markdown 与 HTML Document 共用该布局；HTML Preview 始终使用 sandbox iframe。
- File editor 标题栏右上角只保留一个发布主状态/操作：
  - Draft：`Publish`
  - 已发布且有修改：`Publish changes`
  - 已发布且无修改：只读 `Published`
- 发布确认面板显示最终公开路径、Current 与 Published Content 的差异摘要，以及 `Cancel` / `Publish`。
- Unpublish 位于发布区域旁的 `···` 次级菜单中，不作为常驻主按钮。
- File editor 主界面默认不显示 embedding pending/ready 等技术状态。
- Embedding 失败但 full-text search 可用时不阻断保存或发布。
- File editor `··· → Search status` 显示 Full-text 与 Semantic search 状态，并可 `Retry semantic indexing`。
- 全站 `Rebuild search` 位于 Admin 页面级工具菜单，不放入每个 File 的主操作区。
- assets panel。
- save 后刷新 search_text；embedding refresh 可异步/降级。

### 11.5 Publish / Unpublish

Publish：

- 设置 status=published。
- 首次发布设置 published_at。
- assets 公开。
- search/recent 可见。

Unpublish：

- status=draft。
- public path 404。
- assets 不公开。
- 不创建 redirect。

### 11.6 Move / Rename / URL path change

- draft path change：不创建 redirect。
- published File path change：创建 redirect。
- Directory path change：为子树中所有 published File 创建 redirects。
- UI 必须显示影响提示。

### 11.7 Delete

- draft File 可删除。
- draft-only Directory/subtree 可删除。
- published File 不可硬删，先 unpublish。
- 含 published File 的 Directory 不可删。

## 12. Asset flow

### 12.1 Upload

1. admin 在 File editor 选择 upload。
2. 前端检查 size/mime 初筛。
3. `POST /api/admin/files/{file_id}/assets` multipart。
4. 后端 MIME/size/SVG 检测。
5. 返回 immutable `public_url`。
6. editor 可插入 URL。

### 12.2 Public open

- Published File assets: `GET /api/assets/{asset_id}/{filename}`。
- Header：`Cache-Control: public, max-age=31536000, immutable`。
- PDF：`Content-Disposition: inline`，点击打开，不自动嵌入。

### 12.3 Replace

- 不原地覆盖。
- 新上传生成新 asset_id/public_url。
- admin 手动更新 content 引用。

## 13. Error screen checklist

- 404 content path。
- 500 API error。
- Reader 访问 admin 时显示 `Author access required`，不误导为未登录。
- root reserved URL path conflict。
- duplicate URL path under same parent。
- asset MIME rejected。
- SVG unsafe rejected。
- Qwen embedding failed but File saved。
- search no results。
- empty directory。
- no comments yet。
