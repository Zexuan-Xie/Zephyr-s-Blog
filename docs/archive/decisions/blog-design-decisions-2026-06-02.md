# Blog 设计决策（brainstorming 增量）

> 日期：2026-06-02 · 来源：与 Claude 的 brainstorming 会话
> 定位：本文件记录 2026-06-02 的最终设计决策；旧 `blog-design-spec.md` 与 `blog-learning-roadmap.md` 已归档，不再作为实现依据。
> 配套：`blog-implementation-plan-2026-06-02.md`（实现计划）、`docs/api/openapi.yaml`（API 契约）、`DESIGN.md`（视觉单一事实来源）、`docs/design/glass-light-v2.html`（视觉原型）。

---

## 1. 验收定位（新增 —— spec 未明确）

- 这是 **xLab 考核作品，带筛选属性**：助教按条打分，但与其他提交**横向竞争**。
- 评分重心（三条）：**① 线上能访问（一票否决）· ② 代码经得起 review · ③ 技术选型讲得出道理**。
- 推论：**优化目标是"工程判断力 + 打磨"，不是功能数量**。spec 里"反直觉但有理由"的选择（chi 而非 gin、手写 SQL 而非 ORM、前后端解耦看见 HTTP 边界）本身就是 review 弹药。
- 时间：硬截止约 **2026-06-13**。

## 2. 交付策略：完整度 + 设计打磨

- 当前实现范围以 `blog-implementation-plan-2026-06-02.md` 为准，不再沿用旧 flat blog spec。
- 全部核心能力做到"都能跑" + 配 `DESIGN.md` 的玻璃设计打磨。
- 扩展点继续作为**干净接缝**保留：`SearchService` / `EmbeddingProvider`、pgvector、`AssetStorage`、与渲染解耦的 JSON API（可喂 MCP）、未来 `SearchLLMProvider` / `RerankProvider`。

## 3. 后端：All-in Go

- 后端固定为 Go + chi + pgx + 手写 SQL。
- 不设置后端技术栈回退线；遇到 auth 或部署问题时集中修复 Go 实现。

## 4. 部署链路（覆盖 spec §4.4 / §6）

- **新租独立服务器**（可报销），不与旧服务同机。
- 开发模式：**repo 内本地开发 → Docker 打包 → 发服务器**。链路：`git push` → 服务器 `git pull && docker compose up -d --build`。
- **边缘栈 = Caddy**。Caddy 单二进制、自动 Let's Encrypt（域名 A 记录指过去即自动签发/续期）、一个 `Caddyfile` 同时托管 React 静态产物 + 反代 `/api` + HTTPS。三 service 结构（web/api/db）不变。
- **域名**：待购置；购得后 A 记录 → 服务器 IP。
- **homepage 集成**：GitHub homepage（纯静态，跑不了 Go/PG）上放一个**指向 blog 域名的链接**，而非同处托管。

## 5. 视觉：glass-ricepaper（见 DESIGN.md）

- 浅色 + 暖宣纸底 + **统一磨砂玻璃表面语言**：强折射圆角边 + 哑光颗粒 + 透明度高 + 单一暖米黄柔光。单一 Action Blue (#0066cc) 强调色，暖墨色 #26221c 文字。
- 经 v2 原型批准（`docs/design/glass-light-v2.html`）。所有 token 已固化进 `DESIGN.md`，作为约束 AI 生成 UI 的单一事实来源。
- **暗色模式明确不做**（light-only）。

## 6. 内容结构：Unix-like 内容树

> 最终模型是 `Node` 内容树 + `FileContent`。中英结合指内容空间允许中文、英文或混写输入，并能正常渲染两种语言；不做翻译组或内容语言版本拆分。

### 6.1 内容空间：Directory / File 树

- 整体内容结构采用 **Unix-like tree**：`Directory` 可以包含 `Directory` 与 `File`，允许任意层级嵌套。
- 这替代传统 blog 的“分区 + 文章列表”作为主内容组织方式；目录就是组织单位，不再单独建分类模型。
- `Directory` 与 `File` 名称都允许中文或英文。
- 树结构必须支持灵活增减目录和文件，适合把个人 blog 扩展为长期知识库/技术笔记空间。
- 公开导航以目录树为核心：目录页展示子目录与文件，文件页渲染内容。

推荐领域模型：

```
Node:
  id
  parent_id nullable FK→Node.id   -- root 下为 null；directory/file 共用树节点
  kind enum('directory','file')
  name                            -- 可中文/英文，用于展示
  slug                            -- URL 片段，后续问题决定是否从 name 生成/是否允许中文
  sort_order
  created_at, updated_at

FileContent:
  node_id PK/FK→Node.id
  content_format enum('markdown','html_document')
  keywords text[] default '{}'      -- admin 维护的检索增强关键词
  body_raw
  body_html                       -- markdown 净化缓存；html_document 可为空或存安全处理后缓存
  search_text                      -- 后端提取后的可索引文本
  status enum('draft','published')
  published_at nullable
  embedding vector(1024) nullable
  embedding_model nullable
  embedding_status enum('pending','ready','failed')
  embedding_error text nullable
  embedding_updated_at nullable
```

### 6.2 语言边界：支持中英文内容，而非翻译组

- 内容可以是中文、英文，或中英混写；系统职责是支持两种语言输入、检索与正常渲染。
- 不要求每个文件有中英两个版本。
- 不再把评论/点赞按“语言版本”拆分；评论与点赞挂在具体 `File` 上。
- UI chrome 仍可保留中英文界面字典，但这只是界面语言，不再驱动内容实体拆分。

### 6.3 渲染格式：Markdown + 完整 HTML 界面

- `File` 支持两种格式：`markdown` 与 `html_document`。
- Markdown：作为普通文章渲染在 `article.glass` 阅读容器内；渲染为 HTML 后必须净化，保留原文 `body_raw` 与净化缓存 `body_html`。
- HTML Document：允许 author/admin 输入完整 HTML 界面（可包含 `<!doctype html>`、`<html>`、`<head>`、`<body>`、内联 CSS/布局），用于做独立 demo、可视化或复杂页面。
- 完整 HTML **不得直接注入主 React DOM**；必须通过 sandboxed iframe 渲染，避免接管主站导航、读取主站 token、污染全局 CSS。
- HTML Document 允许 JavaScript，但 iframe sandbox 只能使用 `sandbox="allow-scripts"`；**不得添加 `allow-same-origin`**。
- HTML Document 不得访问主站 JWT/localStorage/cookie，不得调用 admin API；外链脚本 CDN 本期默认禁用，优先自包含 HTML/CSS/JS。
- HTML Document 的 `body_raw` 存原始完整文档；搜索索引从 HTML 中抽取可见文本写入 `search_text`；不把完整 HTML 作为主 DOM 净化片段注入。
- HTML Document 仍在 blog 主站外壳内展示：保留 nav、面包屑/路径、点赞与评论区，iframe 作为正文内容区域；本期不做 immersive 独立展示模式。
- HTML iframe 本期使用固定视口高度：桌面 `min-height: 640px; height: 78vh`，移动端约 `height: 72vh`；iframe 内部自行滚动。
- 本期不做 `postMessage` 自动高度；`[future-optimization]` 可增加 iframe auto-resize。
- 评论内容仍为纯文本，不跟随 File 的 Markdown/HTML 能力升级，避免评论区成为第二个富文本安全面。

### 6.4 搜索适配

- 搜索对象是 published `File`。
- 搜索范围：`File.name` + `FileContent.keywords` + `FileContent.search_text` + path。
- Markdown 的 `search_text` 从 Markdown 原文提取纯文本；HTML Document 的 `search_text` 只提取可见文本，不索引 CSS/JS 源码。
- HTML Document 搜索提取需排除 `script`、`style`、`meta`、`link`、`noscript`、hidden 内容。
- hybrid retrieval 仍保留：Postgres full-text search + Qwen/DashScope embedding semantic retrieval + RRF。
- Directory 可参与路径/导航过滤，但默认不作为正文搜索结果，除非未来做“搜索目录”。
- 搜索结果必须显示完整 current path，避免不同目录下同名 File 混淆。
- 搜索结果卡片主 label 统一为 `FILE`，显示 File `name`、完整 path、snippet、可选命中来源 badge（`text` / `semantic` / `keyword` / `text+semantic`）；Markdown/HTML 只作为弱 meta 或渲染方式，不作为筛选维度。
- 点击搜索结果跳转 current path；如果命中旧 path redirect，应先解析到 current path 再展示。

### 6.5 API 方向

- 公开 API 围绕内容树：`/tree/resolve`、root directory、directory children、file content、recent、search。
- 公开 URL 直接使用树路径（如 `/notes/go/auth`），由 `GET /api/tree/resolve?path=...` 解析。
- 详细契约见 `docs/api/openapi.yaml`。


### 6.6 URL 身份：name 展示，slug 走路径

- `name` 是展示名，允许中文、英文、空格与标点。
- `slug` 是 URL 片段，建议小写英文、数字与 `-`，不直接等同于 `name`。
- Directory 与 File 都使用 slug 组成公开路径，例如 `/research-notes/go-auth-practice`。
- 同一父目录下 slug 唯一：`unique(parent_id, slug)`；不同目录下可复用同名 slug。
- 创建时可根据 name 自动生成 slug 建议，但 admin 可手动编辑。
- Directory 与 File 的 slug 都允许修改，但对 published 可见路径的变更必须自动记录 `path_redirects`，保证旧链接可跳转到新路径。
- 移动 published File 或 published 子树也视为路径变更，必须记录旧路径到新路径的 redirect。
- draft 节点路径变化不创建 redirect，因为 draft 没有公开 URL。
- 取舍理由：展示名保留中英自由，URL 保持稳定、可读、易分享；同时通过 `path_redirects` 支持 Unix-like 内容树的长期重构能力。


### 6.7 Path redirects：支持内容树重构

- 本期实现 `path_redirects`，用于 Directory/File 改 slug 或移动位置后保留旧公开链接。
- 表结构建议：

```
PathRedirect:
  id
  old_path              -- unique, 以 / 开头的完整旧路径
  new_path              -- 当前目标路径
  node_id FK→Node.id    -- 被重定向到的当前节点
  created_at
```

- 触发时机：published File 的 slug 改变；published File 移动目录；包含 published File 的 Directory 改 slug 或移动。
- 子树变更：Directory 路径变化时，为其下所有 published File 的旧路径批量创建 redirect。
- 解析规则：公开路径先按当前内容树解析；解析失败再查 `path_redirects.old_path`。
- 链处理：不保留 redirect 链；如果旧 redirect 的 `new_path` 又变化，应更新为最终 current path。
- 约束：`old_path` unique；不得创建指向 draft 或不存在节点的 redirect；删除节点不自动创建 redirect。
- `[future-optimization]` 可增加 redirect 管理后台、访问统计、手动 redirect、gone/410 策略。
- 取舍理由：路径可重构是 Unix-like 内容树的核心体验；redirect 增加一些实现复杂度，但能保护外部分享、搜索结果和 GitHub homepage 链接。


### 6.8 路径解析 API：SPA 软跳转

- 本期不做服务端 301；保持 Caddy 托管 React 静态产物 + `/api` 反代的简单部署形态。
- 前端进入任意内容路径后调用 `GET /api/tree/resolve?path=/old/path`。
- API 返回三类结果：

```json
{ "type": "directory", "node": { } }
{ "type": "file", "node": { }, "content": { } }
{ "type": "redirect", "new_path": "/new/path" }
```

- 如果返回 `redirect`，前端执行 `navigate(new_path, { replace: true })`，避免浏览器历史保留旧路径。
- 只有当前树解析失败时才查 `path_redirects.old_path`；当前路径优先，避免 redirect 抢占新内容。
- `[future-optimization]` 如未来需要 SEO/HTTP 语义，再改为 Go resolver 返回真实 301，或让 Caddy 将公开内容路径反代到 Go。
- 取舍理由：SPA 软跳转足以保护用户体验，同时不破坏当前 Caddy 静态托管的低复杂度部署。


### 6.9 内容树交互：悬浮卡片 + 可呼出目录侧边栏

- 每个公开内容树页面以悬浮玻璃态卡片展示当前目录下的下一层入口。
- Directory 入口与 File 入口都使用同一 `content-entry-card` 玻璃卡片，保持 `DESIGN.md` 的单一磨砂玻璃表面语言。
- Directory 卡片用于进入下一层目录；File 卡片用于打开内容文件。
- 当前路径通过面包屑/路径栏展示，帮助用户理解 Unix-like tree 的层级位置；breadcrumb 每层都可点击，支持快速返回 root 或任一上级 Directory。
- 侧边栏可被呼出，用于浏览完整 directory tree；侧边栏本身也是 glass surface，不引入第二套视觉语言。
- 侧边栏采用**抽屉式覆盖**，不推开主内容，避免破坏约 760px 阅读宽度。
- 桌面宽度约 `320px`；移动端宽度 `min(88vw, 360px)`。
- 打开时使用温暖、轻量遮罩；点击遮罩、Esc、关闭按钮都可收起。
- 当前路径在目录树中用 Action Blue 标记，其余节点使用 `{colors.ink-60}` / `{colors.ink-40}` 层级。
- 本期不做永久固定左栏作为默认布局，目录树默认按需呼出。


### 6.10 Content entry card 信息结构

- Directory 与 File 入口使用统一 `content-entry-card`，通过 label 与 caption 区分类型。
- Directory 卡片：label=`DIRECTORY`；title=`name`；caption=子目录数 + 文件数 + path；点击进入目录页。
- File 卡片：主 label 统一为 `FILE`；title=`name`；caption=更新时间 / 预计阅读时间或弱 meta / path；点击后按 `content_format` 渲染为 Markdown 阅读面板或 HTML iframe。
- Markdown/HTML 不作为公开信息架构分类，也不提供 format filter；二者都被用户理解为 File，只是展示方式不同。
- 图标可选，但只能使用单色 warm ink / Action Blue，不引入多彩文件类型图标。
- 卡片 hover/press 仍遵循玻璃体系：轻微 lift 或 Action Blue 文本强调，按压 `scale(0.95)` 只用于按钮，不强制用于整卡。


### 6.11 首页：Root directory

- 首页 `/` 就是 content tree 的 root directory 页面。
- Root directory 不需要真实 slug；其子节点构成站点一级入口。
- 首页展示 root 下的 Directory/File `content-entry-card` grid，可选在顶部放一块 glass hero/intro。
- 不把“最新文章 feed”作为首页主结构，避免回到传统 blog 信息架构。
- 顶部保留 nav、搜索入口、目录侧边栏按钮。
- 取舍理由：首页直接体现 Unix-like tree 隐喻，让用户从根目录开始浏览内容空间。


### 6.12 Recent 页面与登录后默认落点

- 保留辅助页面 `/recent`，展示所有 published File，按最近更新/发布排序。
- `/recent` 只展示 File，不展示 Directory；每张 `content-entry-card` 必须显示其所在 path，帮助用户回到树结构。
- 首页 `/` 仍是 root directory，不被 latest feed 取代。
- 匿名访客默认进入 `/`；用户登录成功后默认跳转 `/recent`。
- 若用户因访问受保护动作被重定向到登录页，登录成功后优先回到原目标；没有原目标时才去 `/recent`。
- nav 可提供 `Recent` 链接，但树状浏览仍是一等入口。
- 取舍理由：`/recent` 满足 blog 的“看更新”体验，同时不破坏 Unix-like tree 作为主信息架构。


### 6.13 匿名阅读与互动权限

- 匿名访客可浏览 published File、目录树、搜索结果、评论线程与点赞数。
- 匿名访客不能评论、回复、点赞或取消点赞。
- 匿名点击评论/回复/点赞动作时跳转登录；登录成功后回到原 File 路径并恢复用户意图上下文。
- reader/admin 可评论、回复、点赞/取消点赞。
- 取舍理由：公开讨论内容增强 blog 的展示价值；写互动绑定登录用户，权限边界清晰。


### 6.14 Admin Tree Manager

- admin 后台采用文件管理器式 Tree Manager，与公开 Unix-like content tree 保持同一心智模型。
- 核心操作：创建 Directory、创建 Markdown File、创建 HTML Document File、编辑 name、编辑 slug、移动节点、删除 draft 节点、发布/撤回 File、刷新 embedding、查看路径变化产生的 redirect。
- 本期不做拖拽移动；移动节点通过选择目标 parent directory 完成。
- 排序使用 `sort_order`，本期可用数字输入或上/下按钮，不做复杂拖拽排序。
- UI 结构：tree browser + selected node editor；移动端可退化为列表 + 编辑页。
- 危险操作（改 slug、移动 published 子树、删除）必须显示影响提示，尤其是会创建 `path_redirects` 的路径变化。
- 取舍理由：后台与公开结构共享同一树模型，减少概念转换；去掉拖拽以降低实现风险。


### 6.15 删除与撤回规则

- 本期不允许硬删除 published File 或包含 published File 的 Directory。
- published File 只能先 `unpublish` 撤回为 draft；撤回后公开路径返回 404。
- draft File 可硬删除。
- Directory 若包含任何 published File（包括深层子树）则不可删除；只包含 draft 子树时可硬删除。
- `path_redirects` 只处理改 slug / 移动造成的路径变化，不处理删除或撤回。
- 删除操作不创建 redirect；本期不做 tombstone 或 410 页面。
- 取舍理由：防止误删公开内容和评论，避免 redirect/tombstone 逻辑膨胀。


### 6.16 UI locale：只切 chrome，不切内容

- 保留 ZH/EN UI chrome 切换，用于导航、按钮、表单、评论占位、卡片 label 等界面文案。
- UI locale 不改变 Directory/File 内容；内容文件本身可中文、英文或混写。
- 实现：`LocaleContext` + `messages/{zh,en}`。
- 不做内容语言回退、内容版本切换或自动翻译。
- 取舍理由：保留双语界面友好性，但避免回到内容语言版本拆分。


### 6.17 Root reserved slugs

- 内容树公开路径会与系统路由共享 root，因此 root 子节点 slug 必须避开系统保留词。
- Root-level reserved slugs：`admin`、`api`、`auth`、`login`、`register`、`recent`、`search`、`settings`。
- 保留词只限制 root 层；子目录下仍可使用同名 slug，例如 `/notes/search`。
- admin 创建或修改 root 子节点 slug 时，若命中 reserved list，后端必须拒绝。
- 后续新增系统页时必须同步更新 reserved list。
- 取舍理由：避免内容树路径与 React Router / Caddy / API 系统路径冲突。


### 6.18 评论/点赞目标：只挂 File

- 评论线程与点赞只挂 File，不挂 Directory。
- Directory 页面是导航容器，不显示评论区或点赞按钮。
- Markdown File 与 HTML Document File 都显示点赞/评论区。
- `Comment.file_node_id` 指向 `Node.id`，并由 service/约束保证目标 `Node.kind='file'`。
- `Like.target_type in ('file','comment')`；file like 的 `target_id` 指向 File node id。
- 取舍理由：File 是内容单位，Directory 是组织容器；互动挂 File 语义更清晰。


### 6.19 Keywords：检索增强而非摘要字段

- 不设置 `summary` 字段；卡片摘要与搜索 snippet 从 `search_text` 或命中片段生成。
- File 增加 `keywords text[] default '{}'`，由 admin 维护，用于增强检索和发现。
- keywords 可中文或英文，适合放术语、别名、缩写、相关概念，例如 `JWT`、`鉴权`、`middleware`。
- keywords 参与全文检索与 embedding 输入，但不作为正文展示的主摘要。
- keywords 在 File 页面公开展示为小型 chips，最多显示前 3 个；完整 keywords 仍用于检索。
- 点击公开 keyword chip 进入 `/search?q=keyword`。
- Directory 卡片不展示 keywords；搜索结果可选展示命中的 keyword。
- HTML Document 如果可见文本较少，应通过 keywords 补足检索入口。
- 取舍理由：keywords 比 summary 更适合增强 hybrid search，且避免 admin 为每个文件额外写摘要；公开展示前三个帮助读者理解主题但不制造视觉噪声。


### 6.20 Directory 页面排序

- Directory 页面默认排序：Directory 在前，File 在后；同组内按 `sort_order asc`，再按 `name asc`。
- admin 可编辑 `sort_order`；本期可用数字输入或上/下按钮。
- `/recent` 例外：只展示 File，按最近更新/发布排序。
- `/search` 例外：按 RRF score 排序。
- 取舍理由：符合文件管理器直觉，同时保留手动排序能力。


### 6.21 Admin 内容格式选择

- 公开侧统一展示为 File；admin 创建 File 时必须选择 `content_format = markdown | html_document`。
- 不同格式使用不同编辑器/预览：Markdown 使用 Markdown editor + article preview；HTML Document 使用 full HTML editor + sandbox iframe preview。
- draft 阶段允许切换 `content_format`。
- published 后不允许直接切换 `content_format`；若确需切换，先 unpublish 为 draft，修改格式并重新检查后再发布。
- 格式切换必须刷新 `search_text`、Markdown `body_html` 缓存和 embedding 状态。
- 取舍理由：公开信息架构不暴露类型分裂，但后台必须承认两种渲染/安全路径不同。


### 6.22 File assets：每个 File 的附件系统

- 本期实现 per-file assets，支持 Markdown 与 HTML Document 引用图片、CSS、JS、数据文件等附件。
- 模型建议：

```
FileAsset:
  id
  file_node_id FK→Node.id
  filename              -- 原始/展示文件名，同一 File 下唯一
  mime_type
  size_bytes
  storage_provider      -- local 本期；未来 s3/r2/oss
  storage_key           -- provider-neutral key，如 files/{file_id}/{asset_id}-{filename}
  created_at
```

- 本期存储：本地 Docker named volume，例如 `uploads:/app/uploads`。
- 存储抽象：实现 `AssetStorage` interface（`Save/Open/Delete`），本期为 `LocalAssetStorage`，未来可替换 S3/R2/OSS。
- 配置：`ASSET_STORAGE=local`、`ASSET_UPLOAD_DIR=/app/uploads`、`ASSET_PUBLIC_BASE_URL=/api/assets`。
- 公开访问路径：建议统一 `GET /api/assets/{asset_id}` 或 `/api/files/{file_id}/assets/{filename}`；实现时必须通过 DB 查 `storage_key`，不得直接拼接用户路径，避免路径穿越。
- Markdown 可用相对引用或插入后的资产 URL；HTML Document 可引用同 File 下 assets。
- 上传只允许 admin。
- 安全限制：限制单文件大小、总大小、MIME allowlist；禁止可执行服务端文件；文件名规范化；同名替换不原地覆盖，而是生成新 asset。
- MIME allowlist：`image/png`、`image/jpeg`、`image/webp`、`image/gif`、`image/svg+xml`、`application/pdf`、`text/css`、`text/javascript`、`application/javascript`、`application/json`、`text/plain`、`text/csv`。
- SVG 是高风险图片格式：允许上传但必须净化/检测；禁止 `<script>`、`on*` 事件属性、`javascript:` URL、外链引用、`foreignObject`。检测不过拒绝上传。
- 大小限制建议：图片 5MB；PDF 20MB；CSS/JS/JSON/text/csv 2MB；单个 File assets 总量 50MB。
- PDF 允许作为 asset 点击打开，`Content-Type: application/pdf`，`Content-Disposition: inline`；但不默认自动嵌入 Markdown article 或 HTML shell。
- 公开缓存：published assets 使用不可变 URL，例如 `/api/assets/{asset_id}/{filename}`；响应 `Cache-Control: public, max-age=31536000, immutable`。
- 替换资产：生成新 `asset_id` / `storage_key` / URL，并更新 File 内容引用；旧 asset 可保留或由 admin 清理。
- draft assets 不公开强缓存，仅 admin token 可访问。
- HTML Document 的外链脚本 CDN 本期默认禁用，但允许引用同 File assets 中的 JS；仍只在 sandbox iframe 内执行。
- 删除/撤回 File 时 assets 不公开；draft assets 仅 admin 可访问。
- `[future-optimization]` S3/R2/OSS `AssetStorage`、图片压缩/缩略图、asset 引用清理、版本化与缓存指纹。
- 取舍理由：完整 HTML 界面和长期知识库需要资产管理；per-file assets 将资产作用域限制在 File 内，避免全局媒体库复杂度。本地 volume 符合单服务器部署，但 provider-neutral key 与存储接口保留迁移路线。

## 7. 质量护栏：B 档（中）

工作流是**全量 vibe coding → 事后通读**，故质量靠**生成约束 + 生成后审查闸门**前置保证：

1. **版本锁**：严格遵 spec §7（chi v5 / golang-jwt v5 / pgx v5 / bcrypt / Vite / react-router v6+ / DOMPurify / Compose v2）。
2. **统一分层**：handler → service → repository，小接口、显式 error、不吞错；SQL 手写但收敛在 repository 层。
3. **测试只打五个高风险点**：**JWT 校验 · 点赞幂等（唯一约束 + ON CONFLICT upsert/delete）· Markdown XSS 净化 · HTML Document iframe sandbox 不含 allow-same-origin · SVG asset 安全检测**。其余手动验证。
4. **审查闸门**：每个大模块生成后，先跑 `code-reviewer` 审查并修完，再交给作者通读。

> 方法论本身（spec + DESIGN.md 约束生成 + 自动 review 把关）是 review 时的加分叙事：证明非无脑 vibe，而有工程治理。

## 8. 对 `blog-design-spec.md` 的覆盖清单（便于核对）

| spec 位置 | 原内容 | 本文件覆盖为 |
|---|---|---|
| 旧后端回退线 | 切换到 TS 后端 | **删除**，All-in Go |
| 旧边缘服务 | 旧静态反代方案 | **Caddy** |
| 旧内容模型 | flat 文章 + 分类 | **Unix-like `Node` 内容树 + `FileContent`**，Directory/File 可嵌套 |
| 旧部署风险 | 与其他服务同机 | **作废**，使用新独立服务器 |
| 全局 | 验收定位模糊 | **xLab 筛选考核**，评分=可访问/可review/选型可辩护 |
| 全局 | 未涉中英文内容 | **内容可中文/英文/混写**，不做翻译组或语言版本拆分 |
| 全局 | 未涉视觉 | **`DESIGN.md`（glass-ricepaper）** |

## 9. 实现入口文档

- 产品需求：`PRD.md`。
- 页面/用户流：`BLOG_FLOW.md`。
- 精确技术栈：`TECH_STACK.md`。
- 后端结构：`BACKEND_STRUCTURE.md`。
- 阶段实现计划：`blog-implementation-plan-2026-06-02.md`。
- API 契约骨架：`docs/api/openapi.yaml`。
- 视觉单一事实来源：`DESIGN.md`。
- 领域术语：`CONTEXT.md`。
