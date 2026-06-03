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

Root reserved slugs：`admin`、`api`、`auth`、`login`、`register`、`recent`、`search`、`settings`。

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

- 无 query：显示 search input + 最近/提示，不自动搜索。

### 7.2 搜索提交

1. 用户输入 query。
2. URL 更新 `/search?q=...`。
3. 调用 `GET /api/search?q=...`。
4. 显示 loading skeleton。
5. 展示结果。

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
5. 若 URL/state 有 return target，跳回原目标；否则 `/recent`。

错误：

- 401 → inline error，不跳转。

### 8.3 Protected interaction redirect

匿名点击 like/comment/reply：

1. 保存 intended action context。
2. navigate `/login?return_to=current_path`。
3. 登录成功后返回 current_path。
4. UI 聚焦评论框或重试 like intent；若无法重试，显示提示。

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

### 11.2 Create Directory

1. 选择 parent。
2. 输入 name。
3. slug 自动建议，可编辑。
4. `POST /api/admin/nodes` kind=directory。
5. root reserved slug 命中时报错。

### 11.3 Create File

1. 选择 parent。
2. 输入 name/slug。
3. 选择 content_format：markdown 或 html_document。
4. 新 File 默认 draft。
5. 进入 editor。

### 11.4 Edit File

- Markdown：Markdown editor + article preview。
- HTML Document：full HTML editor + sandbox iframe preview。
- keywords editor。
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

### 11.6 Move / Rename / Slug change

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
- unauthorized admin。
- root reserved slug conflict。
- duplicate slug under same parent。
- asset MIME rejected。
- SVG unsafe rejected。
- Qwen embedding failed but File saved。
- search no results。
- empty directory。
- no comments yet。
