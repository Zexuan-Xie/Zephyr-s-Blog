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
| `/admin` | admin | Author Workspace | admin APIs |
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
- Author 的 `Logout` 位于 Author Workspace 内，不在顶部增加第二层入口。
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

## 11. Author Workspace flow

Stage 2 boundary note: Stage 2 implements the Chinese Author Workspace shell, protected complete Content Tree, minimal create, manual-save File workspace, Settings, same-parent desktop drag sorting, graphical Directory Picker, and Author-only public manage/edit entries. Draft Preview, autosave, Content Versions, `Unpublished changes`, `Publish changes`, Draft/Published Assets, System status, and Rebuild search are Stage 3 or later unless the active stage plan explicitly says otherwise.

### 11.1 Layout

- Tree browser + selected node editor。
- 移动端可退化为 list → edit page。
- Author Workspace 不显示旧的 `ADMIN / Tree Manager` 介绍卡，进入后直接显示工作区。
- Stage 2 顶部/工作区使用中文：`作者工作区`、`内容树`、`访问路径`、`查看公开页面`、`退出登录`。
- Stage 2 不显示 `Admin / Tree Manager`、`Content`、`View site`、`Rebuild search`、`System status` 等英文主操作。
- 未选中节点时，右侧只显示 `请选择目录或文件`、一句简短说明和 `新建目录` / `新建文件`。
- 空状态不显示统计仪表盘或使用教程；其创建入口与树顶部入口相同。
- `查看公开页面` 始终在新标签页打开，不离开当前 Author Workspace。
- 选中 Published File 时打开其公开 File；选中公开 Directory 时打开该 Directory。
- Stage 2 不提供 Draft Preview；选中 Draft File 时不显示公开查看入口。
- 未选中节点时打开网站首页。
- Content Tree 顶部提供唯一的 `＋ New` 创建入口。
- 点击新建后先显示两个大型图形类型卡片：
  - `目录` — 整理文件
  - `文件` — 创建内容
- 新建流程替换右侧工作区，不使用 Modal；左侧内容树保持可见。
- 选择类型后进入简短创建面板，并以可读 path 显示 parent context，例如 `Create in /research/notes`。
- 创建期间切换左侧选中的 Directory 会同步更新 parent context。
- 切换 parent 时保留已输入的 Name 和 File Format，更新路径预览，并轻量提示 `Creation location updated`。
- 切换 parent 不弹确认框，因为尚未产生持久化数据或输入丢失。
- `取消` 返回此前的节点工作区。
- 移动端进入全宽创建视图，返回后回到 Content Tree。
- 创建主流程不显示 Parent ID。
- 离开 New workspace 时，仅当 Name 已填写且尚未创建才提示 `Discard this new item?`，操作为 `Keep editing` / `Discard`。
- Name 为空时，即使已选择 Directory/File 类型，也可直接退出。
- 单击节点时选中并打开；Directory 使用独立箭头展开或折叠。
- Author Workspace Content Tree 显示所有 Directory、Draft File 与 Published File，包括只包含 Draft 的 Directory。
- Stage 2 使用一次受保护完整树加载；Directory 工作区 child cards 可复用该树或调用 detail API。
- File 显示轻量状态：`草稿` 或 `已发布`。
- Stage 2 不显示 `Unpublished changes` / `Changes` / `Save failed` 树状态；这些属于 Stage 3 autosave/snapshot。
- 状态具有文字或 ARIA 描述，不只依赖颜色。
- 折叠 Directory 在 Stage 2 不汇总 `有未发布修改` 或保存失败状态；不显示后代状态数量。
- Author Workspace Content Tree 使用独立的受保护管理 API，不能用过滤 Draft 的 Public Tree 代替。
- 重新进入 `/admin` 时恢复本浏览器上次选中的节点及 Directory 展开状态。
- 上次节点已删除时，选择最近仍存在的 parent Directory；首次进入只显示根级，不展开全部。
- Author Workspace tree navigation state 不跨设备同步。
- Author Workspace Content Tree 不提供独立节点搜索或 `Find in tree`。
- 全站公开搜索只检索 Published File；Draft 不进入公开搜索。
- Author 仅通过 Author Workspace Content Tree 查看和定位 Draft。
- 仅在节点选中时显示 `···` 操作菜单，避免常驻操作按钮造成视觉噪音。
- Directory 菜单：`New inside`、`Advanced settings`。
- File 菜单：`Open editor`、`Advanced settings`。
- `设置` 是当前节点的右侧工作区视图：通过选中节点的菜单打开，不使用弹窗或独立路由。
- 保存或取消后返回该节点原来的 Directory/File 工作区。
- `设置` 包含 `基础信息`、`位置`、`危险操作`；可包含默认折叠的只读技术细节，但主 UI 不显示 Node ID、Parent ID、sort_order 或 `slug`。
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
- 拖拽松开后立即保存排序，不增加 `保存顺序` 按钮。
- 保存期间显示轻量 `正在保存…` 状态，成功后提示 `顺序已更新`。
- 保存失败时恢复操作前顺序并提示 `未能更新顺序`。
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
- Stage 2 仅为 formerly public Published File 旧路径创建 redirect；Directory 自身不作为公开 redirect 目标，除非其 Published File 后代受影响。
- 只包含 Draft 且从未公开可见的 Directory/路径不创建公开 redirect。
- Redirect 默认永久保留并自动压平到当前路径，不允许形成 chain 或 loop。
- Redirect 首版不提供 Author Workspace 主流程管理入口；调试/维护查看不属于 Stage 2 主 UI。
- 目标节点删除后，其历史 Redirect 失效并返回 404。

### 11.3 Create File

1. 在 Content Tree 中选择 parent Directory；创建流程不要求输入或复制 Node ID。
2. 通过图形化创建入口选择 `File`。
3. 简短创建面板只包含 `Name` 与图形化 `Format` 选择：`Markdown` 或 `HTML Document`。
4. 系统根据 Name 自动生成 URL path 的最后一段，sort order 默认放在末尾，新 File 默认 draft。
5. URL path 与 sort order 仅在创建后的 `Advanced settings` 中编辑。
6. 成功后显示明确成功反馈、刷新 Content Tree、自动选中新 File 并进入 editor。

Directory/File 创建失败时只显示后端实际失败原因，不得把成功后的前端状态异常误报为创建失败。

Author Workspace 写操作错误使用面向 Author 的可操作文案，并尽量显示在对应字段或操作附近：

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

Stage 2:

- File workspace 顶部提供 `内容`、`资源`、`设置` 三个标签，默认进入 `内容`。
- `内容` 包含正文、Keywords、格式只读/受限显示、手动保存、发布/撤回发布。
- `资源` 使用现有 per-file asset model：上传、文件列表、复制链接与删除；不区分 Draft Asset / Published Asset。
- `设置` 打开该节点的设置视图。
- Markdown：Markdown editor + article preview。
- HTML Document：full HTML editor + sandbox iframe preview。
- File 内容与 Keywords 使用手动保存；保存成功/失败显示明确中文反馈。
- File editor 标题栏右上角只显示 Stage 2 发布状态/操作：
  - 草稿：`发布`
  - 已发布：只读 `已发布`
  - `撤回发布` 位于次级/危险区域，不作为常驻主按钮。
- Stage 2 不显示 `有未发布修改`、`发布更新`、`Version history`、autosave 状态或 Draft Preview。
- Embedding 失败但 full-text search 可用时不阻断保存或发布。

Stage 3 adds autosave, Current/Previous Content Versions, independent Published Content, `有未发布修改`, `发布更新`, Draft Preview, Draft/Published Assets, and version/asset publication summaries.

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
- Stage 2 所有非空 Directory 禁止删除，包括 draft-only subtree；不提供递归删除。
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
