# BACKEND_STRUCTURE.md — 后端结构、数据库、认证、API、存储规则

> 版本：2026-06-03
> 目的：给后端实现 agent 一个无需猜测的 schema/API/edge-case 规格。
> API 单一事实来源：`docs/api/openapi.yaml`

## 1. Backend module layout

```txt
api/
  cmd/server/main.go
  internal/config/
  internal/db/
  internal/http/router.go
  internal/http/middleware/
  internal/http/handlers/
  internal/auth/
  internal/users/
  internal/tree/
  internal/content/
  internal/render/
  internal/search/
  internal/comments/
  internal/likes/
  internal/assets/
  internal/admin/
  migrations/
```

Required dependency direction:

```txt
handler -> service -> repository -> db
```

Rules:

- Handlers parse/validate HTTP and call services only。
- Services enforce business rules。
- Repositories contain SQL only。
- No SQL in handlers。
- Explicit errors; no swallowed errors。

## 2. Database schema

Use UUID primary keys. Prefer `gen_random_uuid()` from `pgcrypto`.

### 2.1 Extensions

```sql
create extension if not exists pgcrypto;
create extension if not exists vector;
```

### 2.2 users

```sql
create table users (
  id uuid primary key default gen_random_uuid(),
  email text not null unique,
  password_hash text not null,
  role text not null check (role in ('admin','reader')),
  display_name text,
  provider text not null default 'local',
  provider_id text,
  created_at timestamptz not null default now()
);

create unique index users_provider_provider_id_unique
  on users(provider, provider_id)
  where provider_id is not null;
```

Rules:

- `POST /auth/register` always creates `role='reader'`。
- `ADMIN_EMAIL` / `ADMIN_PASSWORD` seed creates or upgrades the admin account。
- `provider/provider_id` only preserve OAuth extension seam; no OAuth route in first release。

### 2.3 nodes

```sql
create table nodes (
  id uuid primary key default gen_random_uuid(),
  parent_id uuid references nodes(id) on delete restrict,
  kind text not null check (kind in ('directory','file')),
  name text not null,
  slug text not null,
  sort_order integer not null default 0,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique(parent_id, slug)
);

create index nodes_parent_sort_idx on nodes(parent_id, kind, sort_order, name);
```

Business constraints not expressible solely by SQL:

- root child slug cannot be reserved slug。
- `file` node must have one `file_contents` row。
- `directory` node must not have `file_contents`。
- moving node cannot create cycle。
- Directory with published descendants cannot be hard-deleted。

Reserved root slugs:

```txt
admin api auth login register recent search settings
```

### 2.4 file_contents

```sql
create table file_contents (
  node_id uuid primary key references nodes(id) on delete cascade,
  content_format text not null check (content_format in ('markdown','html_document')),
  keywords text[] not null default '{}',
  body_raw text not null,
  body_html text,
  search_text text not null default '',
  status text not null default 'draft' check (status in ('draft','published')),
  published_at timestamptz,
  embedding vector(1024),
  embedding_model text,
  embedding_status text not null default 'pending' check (embedding_status in ('pending','ready','failed')),
  embedding_error text,
  embedding_updated_at timestamptz
);

create index file_contents_status_idx on file_contents(status);
create index file_contents_keywords_gin_idx on file_contents using gin(keywords);
```

Full-text index implementation option:

```sql
alter table file_contents
  add column search_vector tsvector generated always as (
    setweight(to_tsvector('simple', coalesce(array_to_string(keywords, ' '), '')), 'A') ||
    setweight(to_tsvector('simple', coalesce(search_text, '')), 'B')
  ) stored;

create index file_contents_search_vector_idx on file_contents using gin(search_vector);
```

pgvector index after enough data exists:

```sql
create index file_contents_embedding_hnsw_idx
  on file_contents using hnsw (embedding vector_cosine_ops)
  where embedding_status = 'ready';
```

Rules:

- `published_at` set only on first transition draft→published。
- published File cannot directly switch `content_format`。
- File save must update `search_text` and set embedding pending/ready/failed.
- Qwen failure never rolls back content save。
- File autosave uses optimistic concurrency: the client submits the loaded content version and stale writes are rejected rather than silently overwriting newer content.

### 2.5 path_redirects

```sql
create table path_redirects (
  id uuid primary key default gen_random_uuid(),
  old_path text not null unique,
  new_path text not null,
  node_id uuid not null references nodes(id) on delete cascade,
  created_at timestamptz not null default now()
);
```

Rules:

- Create only for published File path changes。
- Directory slug/move changes create redirects for all published descendant Files。
- Do not create redirects for draft path changes。
- Do not create redirects for delete/unpublish。
- No redirect chains; update old records to final current path。

### 2.6 comments

```sql
create table comments (
  id uuid primary key default gen_random_uuid(),
  file_node_id uuid not null references nodes(id) on delete cascade,
  user_id uuid not null references users(id) on delete cascade,
  parent_id uuid references comments(id) on delete restrict,
  reply_to_user_id uuid references users(id) on delete set null,
  body text not null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz,
  deleted_by uuid references users(id) on delete set null
);

create index comments_file_parent_created_idx on comments(file_node_id, parent_id, created_at);
```

Rules:

- `file_node_id` must be Node kind `file`。
- parent comment must belong to same file。
- parent_id must point to top-level comment; reply-to-reply gets normalized。
- body plain text only。
- delete = set deleted_at/deleted_by; never physical delete for published comments path。

### 2.7 likes

```sql
create table likes (
  user_id uuid not null references users(id) on delete cascade,
  target_type text not null check (target_type in ('file','comment')),
  target_id uuid not null,
  created_at timestamptz not null default now(),
  primary key(user_id, target_type, target_id)
);

create index likes_target_idx on likes(target_type, target_id);
```

Rules:

- file target_id must be file node id。
- comment target_id must be comment id and not deleted for new likes。
- Like = idempotent upsert。
- Unlike = idempotent delete。

### 2.8 file_assets

```sql
create table file_assets (
  id uuid primary key default gen_random_uuid(),
  file_node_id uuid not null references nodes(id) on delete cascade,
  filename text not null,
  mime_type text not null,
  size_bytes bigint not null,
  storage_provider text not null default 'local',
  storage_key text not null,
  created_at timestamptz not null default now(),
  unique(file_node_id, filename)
);

create index file_assets_file_idx on file_assets(file_node_id);
```

Rules:

- file_node_id must be Node kind `file`。
- same filename replacement generates new asset or requires explicit detach then upload; no in-place overwrite。
- public URL uses asset id and filename。
- storage_key is provider-neutral; never expose local absolute path。

## 3. Authentication logic

### 3.1 Register

Input: email, password, display_name.

Steps:

1. Validate email format and password length >= 8。
2. If email exists, 409。
3. Hash password with bcrypt。
4. Insert user role reader, provider local。
5. Issue JWT。

### 3.2 Admin seed

On server startup:

1. Read `ADMIN_EMAIL`, `ADMIN_PASSWORD`。
2. If both set and user absent, create local admin。
3. If user exists, ensure role admin。
4. Never log password。

### 3.3 Login

1. Find user by email。
2. bcrypt compare。
3. Issue JWT with claims: `sub`, `role`, `email`, `exp`。
4. 401 on invalid credentials。

### 3.4 Middleware

- `RequireAuth`: validates JWT, loads user.
- `RequireAdmin`: requires role admin.
- Public endpoints must not require auth but may include viewer state if bearer token exists.

## 4. API endpoint contract summary

Detailed schema lives in `docs/api/openapi.yaml`.

### Public

```txt
POST   /api/auth/register
POST   /api/auth/login
GET    /api/auth/me
GET    /api/tree
GET    /api/tree/resolve?path=/...
GET    /api/tree/{node_id}/children
GET    /api/recent
GET    /api/search?q=...
GET    /api/files/{file_id}/comments
PUT    /api/files/{file_id}/like
DELETE /api/files/{file_id}/like
PUT    /api/comments/{comment_id}/like
DELETE /api/comments/{comment_id}/like
GET    /api/assets/{asset_id}/{filename}
```

Note: like endpoints require auth even though listed with public URL pattern.

### Protected reader/admin

```txt
POST   /api/files/{file_id}/comments
DELETE /api/comments/{comment_id}
```

### Admin

```txt
POST   /api/admin/nodes
GET    /api/admin/nodes/{node_id}
PATCH  /api/admin/nodes/{node_id}
DELETE /api/admin/nodes/{node_id}
PUT    /api/admin/files/{file_id}/content
POST   /api/admin/files/{file_id}/publish
POST   /api/admin/files/{file_id}/unpublish
POST   /api/admin/files/{file_id}/assets
DELETE /api/admin/assets/{asset_id}
POST   /api/admin/files/{file_id}/refresh-embedding
POST   /api/admin/search-index/rebuild
```

Author Workspace Content Tree requires a protected children-list/read model that includes all Directories and both Draft and Published Files. Public tree endpoints remain publication-filtered and must not serve as the Author Workspace tree source.

Node creation owns slug generation and conflict resolution on the server. It normalizes the submitted Name and selects the final same-parent unique slug transactionally; client path previews are advisory only.

Changing a Directory URL path rewrites its descendant paths and records redirects for formerly public paths in one transaction. Any destination conflict aborts the entire subtree change; Draft-only paths do not create public redirects.

Moving a Directory to another parent uses the same atomic subtree rewrite and redirect rules and rejects self/descendant destinations.

Redirects are retained, resolved directly to the current target without chains/loops, exposed read-only for system inspection, and cease resolving when their target node is deleted.

The content-version migration is lossless and transactional:

- Existing Published Files initialize Current and Published Content from the existing content; Previous is empty.
- Existing Draft Files initialize Current only; Previous and Published Content are empty.
- Existing Assets migrate to Draft/Published state from the File publication state and actual Published Content references.
- Back up the database before running the migration.

## 5. Tree path resolution

Algorithm:

1. Normalize input path: starts with `/`, collapse duplicate slashes, remove trailing slash except root。
2. If root `/`, return root directory。
3. Split slugs。
4. Walk from parent null using `(parent_id, slug)`。
5. If resolved Directory: return DirectoryPage with published children only for public endpoint。
6. If resolved File: require status published for public endpoint; return FilePage。
7. If not resolved: query `path_redirects.old_path`。
8. If redirect found and target node exists/published: return redirect new_path。
9. Else 404。

## 6. Publishing and deletion edge cases

### 6.1 Publish

- Only file nodes can publish。
- If current status draft and published_at null, set published_at=now。
- If assets exist, they become publicly accessible。
- File appears in `/recent`, search, path resolve。

### 6.2 Unpublish

- status → draft。
- Public path returns 404。
- Assets no longer public。
- Comments remain stored but not public because File is not public。
- No redirect created。

### 6.3 Delete

- Draft File can hard delete。
- Draft-only Directory subtree can hard delete。
- Published File cannot hard delete; must unpublish first。
- Directory containing any published File cannot hard delete。

## 7. Render rules

### 7.1 Markdown

Backend or frontend may render, but final displayed HTML must be sanitized.

Backend responsibilities:

- Produce `search_text` plain text。
- Optionally produce sanitized `body_html` cache。

Frontend responsibilities:

- Never render unsanitized Markdown HTML。

### 7.2 HTML Document

- Store raw full document in `body_raw`。
- Render using iframe only。
- Required iframe: `sandbox="allow-scripts"`。
- Prohibited: `allow-same-origin`。
- No direct injection into React DOM。
- Extract visible text for search.

## 8. Search implementation

### 8.1 Full-text

- Use `websearch_to_tsquery('simple', q)`。
- Rank with `ts_rank`。
- Snippet with `ts_headline` where applicable。

### 8.2 Semantic

Embedding input should include:

```txt
name\npath\nkeywords joined\nsearch_text
```

Provider:

- Qwen/DashScope `text-embedding-v4`
- dimensions 1024
- OpenAI-compatible request body must include `dimensions: 1024` and `encoding_format: "float"`。
- China API keys must use `https://dashscope.aliyuncs.com/compatible-mode/v1`; International API keys must use `https://dashscope-intl.aliyuncs.com/compatible-mode/v1`。

Failure handling:

- On Qwen error, set `embedding_status='failed'`, `embedding_error`。
- Content save still succeeds。
- Search still returns full-text results。

### 8.3 RRF

For each candidate from full-text and semantic topK:

```txt
score += 1 / (60 + rank)
```

Return match_sources based on which lists contained the File.

## 9. Asset storage rules

### 9.1 Local storage

Config:

```txt
ASSET_STORAGE=local
ASSET_UPLOAD_DIR=/app/uploads
ASSET_PUBLIC_BASE_URL=/api/assets
```

Storage key format:

```txt
files/{file_node_id}/{asset_id}-{safe_filename}
```

Never build filesystem path from raw filename without normalization.

### 9.2 MIME allowlist

Allowed:

```txt
image/png
image/jpeg
image/webp
image/gif
image/svg+xml
application/pdf
text/css
text/javascript
application/javascript
application/json
text/plain
text/csv
```

Size limits:

```txt
images: 5MB
PDF: 20MB
CSS/JS/JSON/text/CSV: 2MB
per-file total: 50MB
```

### 9.3 SVG checks

Reject SVG if it contains:

- `<script>`
- any `on*=` event attribute
- `javascript:` URL
- external href/src references
- `<foreignObject>`

### 9.4 Public asset response

For published File assets:

```txt
Cache-Control: public, max-age=31536000, immutable
```

PDF:

```txt
Content-Type: application/pdf
Content-Disposition: inline
```

Draft assets require admin auth and no public immutable cache.

## 10. Required tests

1. JWT validation and role guard.
2. Like idempotency with unique constraint.
3. Markdown XSS sanitization.
4. HTML iframe sandbox does not include `allow-same-origin`.
5. SVG asset rejection for malicious SVG.
6. Root reserved slug rejection.
7. Path redirect creation on published path change.
8. Qwen embedding failure does not fail content save.
9. Directory with published descendant cannot delete.
10. HTML visible text extraction excludes script/style.
