# PRD.md — xLab Personal Blog 产品需求文档

> 版本：2026-06-03
> 状态：实现用最终 PRD
> 读者：新开的 Codex/Claude Code session、实现 agent、review agent
> 配套文档：`docs/specs/BLOG_FLOW.md`、`docs/specs/TECH_STACK.md`、`docs/specs/BACKEND_STRUCTURE.md`、`docs/api/openapi.yaml`、`docs/specs/DESIGN.md`、`docs/specs/CONTEXT.md`

## 1. 产品目标

构建一个单作者全栈个人 blog / knowledge space，用于 xLab 考核与长期技术内容沉淀。它不是传统 flat article blog，而是一个 **Unix-like content tree**：Directory 可嵌套 Directory 和 File，File 可以是 Markdown reading page 或 sandboxed full HTML document。

产品要同时证明：

1. 线上可访问和可部署。
2. 前后端解耦全栈能力。
3. Go 后端、Postgres/pgvector、OpenAPI-first、assets、安全隔离、hybrid search 等工程判断力。
4. glass-ricepaper 视觉打磨。

## 2. 成功标准

### 2.1 必须满足

- 用户可访问线上站点，匿名可浏览公开内容树、File、评论、点赞数和搜索结果。
- 用户可 email/password 注册 reader，登录后默认进入 `/recent`。
- admin 账号由环境变量 seed 创建，admin 可管理内容树、File、assets、发布状态、embedding 刷新。
- 首页 `/` 是 root directory 页面，展示 root 下一级 Directory/File 卡片。
- Directory 支持任意层级嵌套。
- File 支持两种 render format：`markdown` 与 `html_document`。
- Markdown 经净化后渲染；HTML Document 通过 iframe sandbox 渲染，允许 JS，但不得 `allow-same-origin`。
- File 可上传 per-file assets；assets 本期存在本地 Docker volume，但通过 `AssetStorage` 抽象保留迁移能力。
- 搜索使用 full-text + Qwen/DashScope `text-embedding-v4` semantic retrieval + RRF。
- 评论为两层楼中楼；评论软删除；点赞幂等。
- API 以 `docs/api/openapi.yaml` 为契约。
- UI 符合 `docs/specs/DESIGN.md` 的 glass-ricepaper 视觉系统。

### 2.2 完成的验收信号

新 session 读完本 PRD 后，如果以下都成立，可判定产品完成：

1. Docker Compose 启动后可访问 web/api/db，Caddy 能服务 SPA 并反代 `/api`。
2. admin seed 后可登录 `/admin`，创建 Directory 和 File。
3. root directory、nested directory、File path 均可通过 SPA 路由访问。
4. 改名/移动 published File 或包含 published File 的 Directory 会创建 `path_redirects`，旧 path 经 `/api/tree/resolve` 返回 redirect。
5. Markdown File 可渲染正文、keywords chips、点赞、评论。
6. HTML Document File 在主站 shell 中通过 fixed-height sandbox iframe 渲染，JS 可运行但不能访问主站 origin/token。
7. assets 上传、公开访问、强缓存、MIME/size/SVG/PDF 规则生效。
8. `/search` 返回包含 path/snippet/source badge 的结果；Qwen embedding 失败时仍可全文搜索。
9. 匿名用户能读公开内容和评论；互动动作跳登录；登录后回到原目标或 `/recent`。
10. 五项高风险测试通过：JWT、点赞幂等、Markdown XSS、HTML iframe sandbox、SVG asset 安全检测。

## 3. 用户角色

### 3.1 Anonymous Visitor

未登录访客。

能做：

- 浏览 root directory、Directory、published File。
- 使用搜索。
- 查看 `/recent`。
- 查看评论线程和点赞数。
- 打开 published assets。

不能做：

- 评论、回复、点赞/取消点赞。
- 访问 admin。
- 查看 draft File 或 draft assets。

### 3.2 Reader

普通注册用户。

能做：

- Anonymous Visitor 的所有能力。
- 评论、回复、点赞/取消点赞 File 或 Comment。
- 删除自己的评论（软删除）。

不能做：

- 创建、编辑、发布内容。
- 上传 assets。
- 管理内容树。
- 访问 `/admin` 时不退出登录；页面明确提示需要 Author 权限并提供返回 `/recent`。
- Reader 访问 `/admin/preview/{file_id}` 时同样显示 Author 权限提示，不能查看 Draft。

### 3.3 Admin / Author

唯一作者，由部署环境 seed 创建。

能做：

- Reader 的所有能力。
- 创建/编辑/move/delete Directory 和 File。
- 发布/撤回 File。
- 上传/删除 per-file assets。
- 修改 slug/name/sort_order/keywords/content_format（受 published 约束）。
- 刷新 embedding / rebuild search index。
- 删除任何评论（软删除）。

## 4. 范围内功能

### 4.1 Content Tree

- Root directory 是 `/`。
- Directory 可嵌套 Directory/File，无固定层级上限。
- Directory/File 共享 Node 基础属性：`id`、`parent_id`、`kind`、`name`、`slug`、`sort_order`、timestamps。
- `name` 可中文、英文或混写。
- `slug` 是 URL 片段；同一 parent 下唯一。
- root-level reserved slugs：`admin`、`api`、`auth`、`login`、`register`、`recent`、`search`、`settings`。
- Directory 页面排序：Directory first → `sort_order asc` → `name asc`。

### 4.2 File

- File 统一作为内容单位；公开卡片主 label 为 `FILE`。
- `content_format` 只影响渲染方式，不作为公开分类或筛选维度。
- Markdown File：渲染到 `file-reading-card`。
- HTML Document File：渲染到 sandboxed iframe，保留主站 nav/breadcrumb/like/comment shell。
- File 有 `keywords text[]`，用于检索增强；公开最多显示前三个 chips。
- File 有 `draft/published` 状态。
- published File 不能直接切换 `content_format`；需先 unpublish。

### 4.3 Path Redirects

- published File 或 published 子树路径变更时创建 `path_redirects`。
- 本期不做 HTTP 301；SPA 通过 `/api/tree/resolve?path=...` 获取 redirect 并 `replace` 跳转。
- 不对删除/撤回创建 redirect。
- 不保留 redirect 链；更新为最终 current path。

### 4.4 Recent

- `/recent` 显示所有 published File。
- 只展示 File，不展示 Directory。
- 每张卡片显示完整 path。
- 已登录用户在无 return target 时登录后跳转 `/recent`。
- 首页仍是 root directory，不被 `/recent` 取代。

### 4.5 Search

- 搜索对象：published File。
- 搜索范围：File name + keywords + extracted search_text + path。
- 搜索只返回 Published File；Draft 即使对 Author 可见，也只在 Admin Content Tree 中定位，不进入搜索。
- Markdown search_text：从 Markdown 原文提取纯文本。
- HTML search_text：只提取可见文本，排除 CSS/JS/meta/link/noscript/hidden。
- Retrieval：Postgres full-text + Qwen/DashScope embedding + pgvector semantic search。
- Fusion：RRF，默认 k=60。
- 搜索结果显示：`FILE`、name、完整 path、snippet、match sources（text/semantic/keyword）。
- 不做 LLM query expansion / rerank。

### 4.6 Comments / Likes

- 只挂 File，不挂 Directory。
- 匿名可读评论/点赞数。
- reader/admin 可评论、回复、点赞。
- 评论两层：top-level comments + replies。
- reply-to-reply 归一到 top-level parent，并设置 `reply_to_user_id`。
- 评论内容纯文本，不支持 Markdown/HTML。
- 删除评论为软删除，保留楼层和子回复。
- 点赞幂等，唯一约束 `(user_id, target_type, target_id)`。

### 4.7 Assets

- per-file assets，只归属某个 File。
- 本期本地 Docker volume 存储；DB 存 provider-neutral storage_key。
- AssetStorage interface 保留 S3/R2/OSS 迁移路线。
- Published assets immutable URL + 强缓存。
- MIME allowlist：PNG/JPEG/WebP/GIF/SVG/PDF/CSS/JS/JSON/text/CSV。
- SVG 必须安全检测。
- PDF 可点击 inline 打开，但不默认嵌入页面。
- HTML Document 可引用同 File assets 中的 JS/CSS，但仍运行在 sandbox iframe。

### 4.8 Admin Tree Manager

- 文件管理器式后台。
- 不做拖拽；移动节点用 parent directory select。
- 排序用 sort_order 数字或上/下按钮。
- 危险操作（移动 published 子树、改 slug、删除）必须提示影响。
- 可上传 assets、刷新 embedding、发布/撤回 File。

## 5. 明确不在范围内

- 旧 flat article/category 信息架构。
- 内容翻译组、内容语言版本切换、自动翻译。
- OAuth 完整实现；只保留 provider 字段接缝。
- SSR / Next.js / Server Components。
- HTTP 301 redirect。
- Object storage 实现；只保留接口。
- 跨 Directory 的 drag-and-drop reparenting；首版只支持同一 parent 内的桌面拖拽排序，移动端使用上移/下移。
- 同级 Directory/File 使用统一混合顺序，不强制 Directory 优先。
- LLM query expansion、rerank、cross-language query translation。
- HTML iframe auto-resize。
- Immersive full-screen HTML mode。
- Directory comments/likes。
- 匿名评论/匿名点赞。
- Comment Markdown/HTML。
- PDF 自动内嵌到文章。
- Dark mode。

## 6. 用户故事

### 6.1 Anonymous browsing

作为匿名访客，我打开 `/`，看到 root directory 的悬浮玻璃卡片入口。我点击 Directory 卡片进入下一层，breadcrumb 能带我回到任意上级。

验收：未登录状态下可以浏览 published Directory/File，但不能看到 draft。

### 6.2 Reader login and recent

作为 reader，我注册并登录。登录后如果没有原目标，进入 `/recent` 查看最新发布/更新的 File。

验收：登录成功后有 JWT；无 return target 时跳 `/recent`；有 return target 时回原 File。

### 6.3 File interaction

作为 reader，我打开 File，能看到 keywords chips、点赞数、评论线程。我点击点赞后计数变化；我可以评论或回复。

验收：重复点赞/取消点赞幂等；评论回复限制两层。

### 6.4 Admin content management

作为 admin，我在 `/admin` 创建 Directory 和 File，编辑内容，上传 assets，发布 File。移动 published File 后旧 path 会软跳到新 path。

验收：Tree Manager 中可完成 CRUD/move/publish/unpublish/asset upload/embedding refresh；published path change 创建 redirect。

### 6.5 HTML Document demo

作为 admin，我创建 HTML Document File，上传 JS/CSS/assets，并发布。访客打开后在 blog shell 中看到 iframe demo。

验收：iframe sandbox 只包含 `allow-scripts`，无 `allow-same-origin`。

### 6.6 Search

作为访客，我搜索关键词，结果能同时命中精确词和语义相关内容，且显示完整 path 与 snippet。

验收：Qwen key 不可用时仍能返回 full-text 结果；embedding ready 后结果带 semantic source。

## 7. 非功能需求

- 安全：JWT role guard、XSS 防护、sandbox、asset MIME/size/SVG 检测。
- 可部署：Docker Compose + Caddy + pgvector + uploads volume。
- 可迁移：assets 使用 provider-neutral key；OpenAPI 契约稳定。
- 可 review：小接口、显式错误、手写 SQL 在 repository 层。
- 可维护：旧方案文档已归档，不得用旧模型实现。

## 8. 新 session 开始时必须读取

1. `docs/specs/PRD.md`
2. `docs/specs/BLOG_FLOW.md`
3. `docs/specs/TECH_STACK.md`
4. `docs/specs/BACKEND_STRUCTURE.md`
5. `docs/api/openapi.yaml`
6. `docs/specs/DESIGN.md`
7. `docs/specs/CONTEXT.md`
