# Blog 实现计划（基于 2026-06-02 决策收束）

> 输入来源：`PRD.md`、`BLOG_FLOW.md`、`TECH_STACK.md`、`BACKEND_STRUCTURE.md`、`blog-design-decisions-2026-06-02.md`、`DESIGN.md`、`docs/api/openapi.yaml`、`CONTEXT.md`。
> 目标：按当前已确认的新模型实现个人全栈 blog：Unix-like 内容树、玻璃态 UI、Qwen hybrid search、per-file assets、评论/点赞、admin Tree Manager、Docker/Caddy 单服务器部署。

---

## 1. 成功标准

1. 线上可访问：Caddy HTTPS + React SPA + Go API + Postgres/pgvector + uploads volume。
2. 代码可 review：Go handler → service → repository 分层；OpenAPI 契约优先；错误显式；SQL 收敛在 repository。
3. 技术选型可辩护：
   - Go + chi + pgx + 手写 SQL。
   - Unix-like `Node` 内容树替代旧扁平内容模型。
   - Search = Postgres full-text + Qwen/DashScope `text-embedding-v4` + pgvector + RRF。
   - Full HTML document 通过 sandboxed iframe 隔离。
   - per-file assets 本地 volume，但用 `AssetStorage` 抽象保留迁移路线。
4. 高风险测试通过：JWT、点赞幂等、Markdown XSS、HTML iframe sandbox、SVG asset 安全检测。

---

## 2. 核心数据模型

### 2.1 Users/Auth

```sql
users(
  id uuid pk,
  email text unique not null,
  password_hash text not null,
  role text check(role in ('admin','reader')) not null,
  display_name text,
  provider text not null default 'local',
  provider_id text null,
  created_at timestamptz not null default now()
)
```

规则：

- 公开注册永远创建 `reader`。
- 唯一 admin 通过 `ADMIN_EMAIL` / `ADMIN_PASSWORD` seed。
- GitHub OAuth 只保留字段接缝，本期不暴露 OAuth 路由。

### 2.2 Content tree

```sql
nodes(
  id uuid pk,
  parent_id uuid null references nodes(id),
  kind text check(kind in ('directory','file')) not null,
  name text not null,
  slug text not null,
  sort_order int not null default 0,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique(parent_id, slug)
)
```

约束：

- root 子节点 slug 禁用：`admin/api/auth/login/register/recent/search/settings`。
- Directory/File 的公开 URL 由 slug path 组成。
- name 可中文/英文/混写；slug 建议英文数字短横线。
- Directory 页面排序：Directory first → `sort_order asc` → `name asc`。

### 2.3 File content

```sql
file_contents(
  node_id uuid pk references nodes(id),
  content_format text check(content_format in ('markdown','html_document')) not null,
  keywords text[] not null default '{}',
  body_raw text not null,
  body_html text null,
  search_text text not null default '',
  status text check(status in ('draft','published')) not null default 'draft',
  published_at timestamptz null,
  embedding vector(1024) null,
  embedding_model text null,
  embedding_status text check(embedding_status in ('pending','ready','failed')) not null default 'pending',
  embedding_error text null,
  embedding_updated_at timestamptz null
)
```

规则：

- `markdown`：渲染/净化后缓存到 `body_html`。
- `html_document`：完整 HTML 原文存 `body_raw`，通过 iframe sandbox 渲染，不注入主 DOM。
- `search_text`：Markdown 提取纯文本；HTML 只提取可见文本，排除 `script/style/meta/link/noscript/hidden`。
- `keywords` 用于检索增强；公开 File 页面最多展示前 3 个 chips。
- published 后不能直接切换 `content_format`；需先 unpublish。

### 2.4 Path redirects

```sql
path_redirects(
  id uuid pk,
  old_path text unique not null,
  new_path text not null,
  node_id uuid not null references nodes(id),
  created_at timestamptz not null default now()
)
```

规则：

- 改 slug / 移动 published File 或含 published File 的 Directory 时创建 redirect。
- 不保留 redirect 链；后续路径变化要更新到最终 current path。
- 删除/撤回不创建 redirect。
- 本期用 API resolve + 前端 `replace` 软跳转，不做 HTTP 301。

### 2.5 Comments / likes

```sql
comments(
  id uuid pk,
  file_node_id uuid not null references nodes(id),
  user_id uuid not null references users(id),
  parent_id uuid null references comments(id),
  reply_to_user_id uuid null references users(id),
  body text not null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz null,
  deleted_by uuid null references users(id)
)

likes(
  user_id uuid not null references users(id),
  target_type text check(target_type in ('file','comment')) not null,
  target_id uuid not null,
  created_at timestamptz not null default now(),
  primary key(user_id, target_type, target_id)
)
```

规则：

- 评论/点赞只挂 File，不挂 Directory。
- 评论两层：顶层 comment + replies；reply-to-reply 归一到顶层 parent，并设置 `reply_to_user_id`。
- 删除评论为软删除，保留楼层和子回复。
- 匿名可读评论/点赞数，互动需登录。

### 2.6 File assets

```sql
file_assets(
  id uuid pk,
  file_node_id uuid not null references nodes(id),
  filename text not null,
  mime_type text not null,
  size_bytes bigint not null,
  storage_provider text not null default 'local',
  storage_key text not null,
  created_at timestamptz not null default now(),
  unique(file_node_id, filename)
)
```

规则：

- 本期 `LocalAssetStorage` + Docker named volume；DB 存 provider-neutral `storage_key`。
- published assets immutable URL：`/api/assets/{asset_id}/{filename}`。
- 强缓存：`Cache-Control: public, max-age=31536000, immutable`。
- 替换资产生成新 asset，不原地覆盖。
- MIME allowlist：PNG/JPEG/WebP/GIF/SVG/PDF/CSS/JS/JSON/text/CSV。
- SVG 必须安全检测；PDF 点击 inline 打开但不默认嵌入页面。

---

## 3. API 契约

单一事实来源：`docs/api/openapi.yaml`（OpenAPI 3.2.0）。

实现前先对齐这些路径：

- Auth：`/auth/register`、`/auth/login`、`/auth/me`
- Public tree：`/tree/resolve`、`/tree`、`/tree/{node_id}/children`、`/recent`
- Search：`/search?q=`
- Comments：`/files/{file_id}/comments`、`/comments/{comment_id}`
- Likes：`/files/{file_id}/like`、`/comments/{comment_id}/like`
- Assets：`/assets/{asset_id}/{filename}`、`/admin/files/{file_id}/assets`
- Admin tree：`/admin/nodes`、`/admin/nodes/{node_id}`、`/admin/files/{file_id}/content`、publish/unpublish
- Admin search：`/admin/files/{file_id}/refresh-embedding`、`/admin/search-index/rebuild`

---

## 4. Go 后端结构

建议目录：

```txt
api/
  cmd/server/main.go
  internal/config/
  internal/db/
  internal/http/
    router.go
    middleware/
    handlers/
  internal/auth/
  internal/tree/
  internal/content/
  internal/render/
  internal/search/
  internal/comments/
  internal/likes/
  internal/assets/
  internal/users/
  migrations/
```

分层：

```txt
handler -> service -> repository
```

关键接口：

```go
type EmbeddingProvider interface {
  EmbedText(ctx context.Context, text string) ([]float32, error)
  ModelName() string
  Dimensions() int
}

type SearchService interface {
  Search(ctx context.Context, query string, limit, offset int) ([]SearchResult, error)
  RefreshFileIndex(ctx context.Context, fileID uuid.UUID) error
  Rebuild(ctx context.Context) error
}

type AssetStorage interface {
  Save(ctx context.Context, key string, r io.Reader) error
  Open(ctx context.Context, key string) (io.ReadCloser, error)
  Delete(ctx context.Context, key string) error
}
```

Qwen 配置：

```txt
EMBEDDING_PROVIDER=qwen
DASHSCOPE_API_KEY=...
EMBEDDING_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1
EMBEDDING_MODEL=text-embedding-v4
EMBEDDING_DIMENSIONS=1024
```

OpenAI-compatible request body must send:

```json
{
  "model": "text-embedding-v4",
  "input": "name\npath\nkeywords joined\nsearch_text",
  "dimensions": 1024,
  "encoding_format": "float"
}
```

Endpoint rule：China API keys use `https://dashscope.aliyuncs.com/compatible-mode/v1`; International API keys use `https://dashscope-intl.aliyuncs.com/compatible-mode/v1`。

---

## 5. React 前端结构

建议路由：

```txt
/                       root directory
/recent                 logged-in default when no return target
/search?q=...
/login
/register
/admin                  Tree Manager
/*                      public content path -> /api/tree/resolve?path=...
```

组件：

```txt
AppShell
GlassNav
DirectoryDrawer
Breadcrumb
ContentEntryCard
DirectoryPage
FilePage
MarkdownRenderer
HtmlDocumentFrame
CommentThread
LikeButton
RecentPage
SearchPage
AdminTreeManager
AssetManager
```

视觉约束：

- 遵守 `DESIGN.md` 的 glass-ricepaper token。
- Directory/File 入口统一 `content-entry-card`。
- 卡片主 label：Directory = `DIRECTORY`，File = `FILE`。
- Markdown/HTML 只是渲染方式，不作为公开筛选维度。
- Directory drawer 抽屉式覆盖，不推开内容。
- File 页面保留 nav / breadcrumb / like / comments。
- HTML document iframe：`sandbox="allow-scripts"`，不得 `allow-same-origin`；固定高度。

---

## 6. 搜索实现顺序

1. 建表：`search_text`、`embedding vector(1024)`、GIN/HNSW 索引。
2. 实现文本提取：Markdown → plain text；HTML document → visible text only。
3. 实现 full-text query：`websearch_to_tsquery`、`ts_rank`、`ts_headline`。
4. 实现 Qwen `EmbeddingProvider`。
5. 实现 semantic topK with pgvector。
6. 实现 RRF：`score = Σ 1 / (k + rank_i)`，默认 `k=60`。
7. 写入降级：保存 File 不被 Qwen 失败阻塞。
8. Admin refresh/rebuild。

搜索结果必须显示：`FILE`、name、完整 path、snippet、match source badge。

---

## 7. Assets 实现顺序

1. `file_assets` migration。
2. `LocalAssetStorage`，写入 `/app/uploads/{storage_key}`。
3. MIME/size allowlist。
4. SVG 安全检测。
5. Admin upload endpoint。
6. Public immutable asset endpoint。
7. Markdown/HTML editor 插入 asset URL。
8. Draft asset admin-only 访问。

限制：

```txt
image max 5MB
PDF max 20MB
CSS/JS/JSON/text/csv max 2MB
per-file total max 50MB
```

---

## 8. 安全与测试清单

必须自动测试：

1. JWT：过期、签名错误、role guard。
2. Like 幂等：重复 like、重复 unlike、并发唯一约束。
3. Markdown XSS：`<script>`、事件属性、javascript URL 不进入 `body_html`。
4. HTML iframe：渲染组件必须有 `sandbox="allow-scripts"` 且不含 `allow-same-origin`。
5. SVG asset：拒绝 `<script>`、`on*`、`javascript:`、外链、`foreignObject`。

建议补充测试：

- root reserved slug 拒绝。
- path_redirects 创建与 resolve。
- Directory 删除规则：含 published File 不可删。
- HTML visible text extraction 不索引 CSS/JS。
- Qwen embedding 失败不阻塞 File 保存。

---

## 9. Docker/Caddy 部署

服务：

```txt
web  Caddy: React static + /api reverse proxy + HTTPS
api  Go chi server
db   pgvector/pgvector Postgres
```

Volumes：

```txt
postgres_data
uploads
```

关键环境变量：

```txt
DATABASE_URL=...
JWT_SECRET=...
ADMIN_EMAIL=...
ADMIN_PASSWORD=...
ASSET_STORAGE=local
ASSET_UPLOAD_DIR=/app/uploads
ASSET_PUBLIC_BASE_URL=/api/assets
EMBEDDING_PROVIDER=qwen
DASHSCOPE_API_KEY=...
EMBEDDING_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1
EMBEDDING_MODEL=text-embedding-v4
EMBEDDING_DIMENSIONS=1024
```

OpenAI-compatible embedding calls must include `dimensions: 1024` and `encoding_format: "float"`。

---

## 10. 推荐实现阶段

### Phase 1 — 地基

- Go server / React app scaffold
- DB migrations
- Auth + admin seed
- OpenAPI path/schema 对齐

### Phase 2 — 内容树 MVP

- Node CRUD
- public resolve/root/directory pages
- Admin Tree Manager basic
- draft/published/unpublish/delete rules
- path_redirects + SPA soft redirect

### Phase 3 — File rendering

- Markdown renderer + sanitizer
- HTML document iframe
- breadcrumb / directory drawer / content cards
- `/recent`

### Phase 4 — comments/likes

- two-level comment thread
- soft delete
- idempotent likes
- anonymous read / logged-in write

### Phase 5 — assets

- per-file upload
- immutable asset endpoint
- MIME/size/SVG/PDF rules
- editor insertion

### Phase 6 — hybrid search

- search_text extraction
- full-text search
- Qwen embeddings
- pgvector semantic retrieval
- RRF
- refresh/rebuild endpoints

### Phase 7 — deployment + review

- Docker Compose + Caddy
- server deploy smoke test
- code review pass
- fix review findings
- final manual verification

---

## 11. Known non-goals for first release

- Real HTTP 301 redirects.
- OAuth implementation.
- Drag-and-drop Tree Manager.
- Object storage implementation.
- LLM query expansion / reranking.
- Auto-resizing HTML iframe.
- Immersive full-screen HTML mode.
- Directory comments/likes.
- Translation groups or content-language switching.
