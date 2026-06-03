# xLab Personal Blog Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the complete xLab single-author full-stack blog specified in `PRD.md`: a Unix-like content tree with Markdown and sandboxed HTML files, auth, admin management, comments, likes, per-file assets, hybrid search, Docker deployment, and glass-ricepaper UI.

**Architecture:** The implementation is a Vite React SPA served by Caddy and backed by a Go/chi API with PostgreSQL + pgvector. The backend follows `handler -> service -> repository -> db`, keeps SQL in repositories, and treats `docs/api/openapi.yaml` as the API contract. The frontend uses React Router route objects, TanStack Query for server state, DOMPurify + marked for Markdown, and iframe sandboxing for full HTML documents.

**Tech Stack:** Go `1.26.4`, chi `v5.3.0`, pgx `v5.10.0`, pgvector-go `v0.4.0`, JWT `v5.3.1`, React `19.2.7`, Vite `8.0.16`, TypeScript `6.0.3`, React Router `7.16.0`, TanStack Query `5.101.0`, DOMPurify `3.4.7`, marked `18.0.4`, Docker Compose v2, `pgvector/pgvector:0.8.2-pg17`, `caddy:2.11.3-alpine`. Local development environment must be managed with Conda and named `blogenv`.

---

## Source Documents Read

- `PRD.md` — final product requirements, roles, acceptance signals, scope and non-goals.
- `BLOG_FLOW.md` — route map, page flow, auth redirects, directory drawer, comments flow.
- `TECH_STACK.md` — exact dependency and image versions; prohibited substitutions.
- `BACKEND_STRUCTURE.md` — backend layout, schema, business rules, search/assets/security rules.
- `docs/api/openapi.yaml` — OpenAPI 3.2.0 endpoint and schema contract.
- `DESIGN.md` and `docs/design/glass-light-v2.html` — glass-ricepaper visual system.
- `CONTEXT.md` — product language; use Directory/File/Reader/Anonymous Visitor, not category/post/subscriber.
- `docs/adr/*.md` — decisions for two-level comments, hybrid search, Unix-like tree, path redirects, iframe HTML documents, per-file assets.

## Scope Check

This plan covers one cohesive product whose subsystems are interdependent: auth gates admin/content/interactions, content tree drives public routes/search/assets, and the UI must call the same OpenAPI contract. It is decomposed into vertical tasks that each leave the repo buildable or closer to a verified endpoint/page. Do not split into separate independent apps; build `api/`, `web/`, and deployment wiring in the same repo.

## File Structure Map

### Root deployment and documentation

- Create: `docker-compose.yml` — defines `db`, `api`, and `web` services plus `postgres_data` and `uploads` volumes.
- Create: `Caddyfile` — serves the built SPA and reverse-proxies `/api/*` to the Go API.
- Create: `environment.yml` — Conda environment definition for local development; environment name must be `blogenv`.
- Modify: `.gitignore` — ignore local env files, build artifacts, uploads, node modules, Go test binaries.
- Modify: `README.md` — document local development, test, and Docker commands after implementation.

### Backend (`api/`)

- Create: `api/go.mod` — pins exact modules from `TECH_STACK.md`.
- Create: `api/cmd/server/main.go` — reads config, connects DB, runs migrations/admin seed, starts chi server.
- Create: `api/internal/config/config.go` — environment parsing with defaults and required secret checks.
- Create: `api/internal/db/db.go` — pgxpool constructor, health check, migration runner.
- Create: `api/migrations/0001_init.sql` — extensions, tables, indexes, generated full-text vector.
- Create: `api/internal/http/router.go` — chi router and route registration matching `docs/api/openapi.yaml`.
- Create: `api/internal/http/middleware/auth.go` — optional auth, require auth, require admin.
- Create: `api/internal/http/response.go` — JSON success/error helpers.
- Create: `api/internal/auth/{jwt.go,password.go,service.go,handler.go}` — register/login/me/admin seed.
- Create: `api/internal/users/model.go` — public/internal user models.
- Create: `api/internal/tree/{model.go,repository.go,service.go,handler.go,path.go}` — node CRUD, public tree, resolve, redirects.
- Create: `api/internal/content/{model.go,repository.go,service.go,handler.go}` — file content upsert, publish, unpublish.
- Create: `api/internal/render/{markdown.go,html_text.go,reading_time.go}` — Markdown sanitization, visible-text extraction, read time.
- Create: `api/internal/comments/{model.go,repository.go,service.go,handler.go}` — two-level comments and soft delete.
- Create: `api/internal/likes/{model.go,repository.go,service.go,handler.go}` — idempotent likes for files/comments.
- Create: `api/internal/assets/{model.go,storage.go,local.go,validation.go,repository.go,service.go,handler.go}` — per-file assets, MIME/size/SVG validation, immutable public endpoint.
- Create: `api/internal/search/{model.go,embedding.go,qwen.go,repository.go,service.go,handler.go}` — full-text, Qwen embeddings, semantic search, RRF, rebuild.
- Create: focused backend tests under `api/internal/**/**_test.go` for the required high-risk cases.

### Frontend (`web/`)

- Create: `web/package.json`, `web/package-lock.json`, `web/tsconfig*.json`, `web/vite.config.ts`, `web/eslint.config.js`, `web/index.html` — exact Vite React TypeScript scaffold.
- Create: `web/src/main.tsx` and `web/src/app/App.tsx` — app bootstrap, QueryClient, Router.
- Create: `web/src/api/client.ts` — typed fetch wrapper, JWT storage, auth error handling.
- Create: `web/src/api/types.ts` — TypeScript shapes aligned to OpenAPI schemas.
- Create: `web/src/auth/AuthContext.tsx` — token/user state, login/register/logout.
- Create: `web/src/routes/{RootDirectoryPage.tsx,ContentPathPage.tsx,RecentPage.tsx,SearchPage.tsx,LoginPage.tsx,RegisterPage.tsx,AdminPage.tsx}` — route-level pages.
- Create: `web/src/components/{AppShell.tsx,GlassNav.tsx,DirectoryDrawer.tsx,Breadcrumb.tsx,ContentEntryCard.tsx,MarkdownRenderer.tsx,HtmlDocumentFrame.tsx,LikeButton.tsx,CommentThread.tsx,LoadingPanel.tsx,ErrorPanel.tsx}` — reusable UI.
- Create: `web/src/admin/{AdminTreeManager.tsx,NodeEditor.tsx,FileEditor.tsx,AssetManager.tsx}` — admin CRUD/edit/upload/publish UI.
- Create: `web/src/styles/tokens.css`, `web/src/styles/glass.css`, `web/src/styles/app.css` — glass-ricepaper tokens, primitive, layout.
- Create: frontend tests under `web/src/**/*.test.tsx`, especially iframe sandbox and Markdown XSS rendering.

---


## Task 0: Conda Development Environment

**Files:**
- Create: `environment.yml`
- Modify: `README.md`

- [ ] **Step 1: Create the Conda environment definition**

Create `environment.yml`:

```yaml
name: blogenv
channels:
  - conda-forge
dependencies:
  - nodejs=22.22.3
  - go=1.26.4
  - git
  - curl
  - make
```

If Conda cannot resolve exact `nodejs=22.22.3` or `go=1.26.4`, do not silently change versions. Record the exact solver error, then update `TECH_STACK.md` and this file together with the agreed replacement versions.

- [ ] **Step 2: Create and activate `blogenv`**

Run:

```bash
conda env create -f environment.yml
conda activate blogenv
node --version
npm --version
go version
```

Expected: active Conda environment is `blogenv`; Node, npm, and Go match `TECH_STACK.md` or the task stops with a documented Conda solver/version mismatch.

- [ ] **Step 3: Document Conda-first development in README**

Add this section near the top of `README.md`:

```markdown
## Development environment

Use Conda for local development. The environment name is fixed:

```bash
conda env create -f environment.yml
conda activate blogenv
```

Run backend, frontend, and verification commands from inside `blogenv` unless the command is explicitly Docker-only.
```

- [ ] **Step 4: Commit**

```bash
git add environment.yml README.md
git commit -m "Make local development reproducible through Conda

Constraint: Local setup must use a Conda environment named blogenv while TECH_STACK.md pins runtime versions.
Confidence: medium
Scope-risk: narrow
Directive: Do not substitute Node or Go versions silently if Conda cannot resolve the pinned versions.
Tested: conda env create -f environment.yml; conda activate blogenv; node --version; npm --version; go version
Not-tested: Docker image builds are verified in the deployment task."
```


## Task 1: Repository Guardrails and Ignore Rules

**Files:**
- Modify: `.gitignore`
- Modify: `README.md`

- [ ] **Step 1: Update `.gitignore` before generating build artifacts**

Replace `.gitignore` with:

```gitignore
.superpowers/

# Environment
.env
.env.*
!.env.example

# Node / Vite
node_modules/
dist/
web/dist/
web/node_modules/
web/.vite/

# Go
api/bin/
api/coverage.out
api/*.test

# Local data
uploads/
postgres_data/

# OS/editor
.DS_Store
*.swp
.vscode/
.idea/
```

- [ ] **Step 2: Add README implementation command shell**

Replace `README.md` with:

```markdown
# Zephyr-s-Blog

Full-stack single-author personal blog for xLab. The current product contract is defined by:

1. `PRD.md`
2. `BLOG_FLOW.md`
3. `TECH_STACK.md`
4. `BACKEND_STRUCTURE.md`
5. `docs/api/openapi.yaml`
6. `DESIGN.md`
7. `CONTEXT.md`

## Development environment

Use Conda for local development. The environment name is fixed:

```bash
conda env create -f environment.yml
conda activate blogenv
```

Run backend, frontend, and verification commands from inside `blogenv` unless the command is explicitly Docker-only.

## Local development

Backend:

```bash
cd api
go test ./...
go run ./cmd/server
```

Frontend:

```bash
cd web
npm install
npm run lint
npm run build
npm run dev
```

Full stack:

```bash
docker compose up -d --build
```

## Product language

Use Directory, File, Reader, Anonymous Visitor, Content Tree, HTML Document, Hybrid Search, and Glass Ricepaper consistently. Avoid the old flat post/category model.
```

- [ ] **Step 3: Verify the doc edit is clean**

Run:

```bash
git diff -- .gitignore README.md
```

Expected: only the ignore rules and README content above changed.

- [ ] **Step 4: Commit**

```bash
git add .gitignore README.md
git commit -m "Prepare the repo for generated app artifacts

Constraint: Existing repo is spec-only and needs safe ignore rules before scaffolding.
Confidence: high
Scope-risk: narrow
Tested: git diff -- .gitignore README.md
Not-tested: Runtime behavior is unchanged because this task edits documentation and ignore rules only."
```

## Task 2: Backend Module, Config, Database, and Migrations

**Files:**
- Create: `api/go.mod`
- Create: `api/cmd/server/main.go`
- Create: `api/internal/config/config.go`
- Create: `api/internal/db/db.go`
- Create: `api/migrations/0001_init.sql`
- Create: `api/internal/http/response.go`

- [ ] **Step 1: Create backend directories**

Run:

```bash
mkdir -p api/cmd/server api/internal/config api/internal/db api/internal/http api/migrations
```

Expected: directories exist and `find api -maxdepth 3 -type d | sort` shows each path.

- [ ] **Step 2: Create exact `api/go.mod`**

Create `api/go.mod`:

```go
module xlab-blog/api

go 1.26.4

require (
	github.com/go-chi/chi/v5 v5.3.0
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.10.0
	github.com/joho/godotenv v1.5.1
	github.com/microcosm-cc/bluemonday v1.0.27
	github.com/pgvector/pgvector-go v0.4.0
	github.com/yuin/goldmark v1.8.2
	golang.org/x/crypto v0.52.0
	golang.org/x/net v0.55.0
)
```

- [ ] **Step 3: Write failing config tests**

Create `api/internal/config/config_test.go`:

```go
package config

import "testing"

func TestLoadRequiresJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://blog:blog@localhost:5432/blog?sslmode=disable")
	t.Setenv("JWT_SECRET", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected missing JWT_SECRET to fail")
	}
}

func TestLoadDefaultsEmbeddingAndAssetConfig(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://blog:blog@localhost:5432/blog?sslmode=disable")
	t.Setenv("JWT_SECRET", "dev-secret-with-enough-length")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.AssetUploadDir != "/app/uploads" {
		t.Fatalf("AssetUploadDir = %q", cfg.AssetUploadDir)
	}
	if cfg.EmbeddingModel != "text-embedding-v4" || cfg.EmbeddingDimensions != 1024 {
		t.Fatalf("embedding config = %s/%d", cfg.EmbeddingModel, cfg.EmbeddingDimensions)
	}
}
```

- [ ] **Step 4: Run config tests to verify failure**

Run:

```bash
cd api && go test ./internal/config
```

Expected: FAIL because `Load` and `Config` are not defined.

- [ ] **Step 5: Implement config loading**

Create `api/internal/config/config.go`:

```go
package config

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr            string
	DatabaseURL         string
	JWTSecret           string
	AdminEmail          string
	AdminPassword       string
	AssetStorage        string
	AssetUploadDir      string
	AssetPublicBaseURL  string
	EmbeddingProvider   string
	DashScopeAPIKey     string
	EmbeddingBaseURL    string
	EmbeddingModel      string
	EmbeddingDimensions int
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:            env("HTTP_ADDR", ":8080"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		AdminEmail:          os.Getenv("ADMIN_EMAIL"),
		AdminPassword:       os.Getenv("ADMIN_PASSWORD"),
		AssetStorage:        env("ASSET_STORAGE", "local"),
		AssetUploadDir:      env("ASSET_UPLOAD_DIR", "/app/uploads"),
		AssetPublicBaseURL:  env("ASSET_PUBLIC_BASE_URL", "/api/assets"),
		EmbeddingProvider:   env("EMBEDDING_PROVIDER", "qwen"),
		DashScopeAPIKey:     os.Getenv("DASHSCOPE_API_KEY"),
		EmbeddingBaseURL:    env("EMBEDDING_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1"),
		EmbeddingModel:      env("EMBEDDING_MODEL", "text-embedding-v4"),
		EmbeddingDimensions: envInt("EMBEDDING_DIMENSIONS", 1024),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}
	if cfg.EmbeddingModel != "text-embedding-v4" {
		return Config{}, errors.New("EMBEDDING_MODEL must be text-embedding-v4")
	}
	if cfg.EmbeddingDimensions != 1024 {
		return Config{}, errors.New("EMBEDDING_DIMENSIONS must be 1024")
	}
	return cfg, nil
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
```

- [ ] **Step 6: Implement DB helper and migrations**

Create `api/internal/db/db.go`:

```go
package db

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed ../../migrations/*.sql
var migrationFiles embed.FS

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create pg pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `create table if not exists schema_migrations (version text primary key, applied_at timestamptz not null default now())`); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}
	entries, err := migrationFiles.ReadDir("../../migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		var exists bool
		if err := pool.QueryRow(ctx, `select exists(select 1 from schema_migrations where version=$1)`, name).Scan(&exists); err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if exists {
			continue
		}
		sqlBytes, err := migrationFiles.ReadFile("../../migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx, `insert into schema_migrations(version) values($1)`, name); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", name, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}
	return nil
}
```

If Go rejects the embedded relative path, change the embed pattern by moving `db.go` to `api/internal/db` with a package-local `migrations` copy is not allowed. The correct fix is to create `api/internal/db/migrations.go` in package `db` that uses `//go:embed migrations/*.sql` and move SQL files to `api/internal/db/migrations/`. Keep only one migration source directory in the final repo.

- [ ] **Step 7: Create SQL schema**

Create `api/migrations/0001_init.sql`:

```sql
create extension if not exists pgcrypto;
create extension if not exists vector;

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

create unique index users_provider_provider_id_unique on users(provider, provider_id) where provider_id is not null;

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
  embedding_updated_at timestamptz,
  search_vector tsvector generated always as (
    setweight(to_tsvector('simple', coalesce(array_to_string(keywords, ' '), '')), 'A') ||
    setweight(to_tsvector('simple', coalesce(search_text, '')), 'B')
  ) stored
);

create index file_contents_status_idx on file_contents(status);
create index file_contents_keywords_gin_idx on file_contents using gin(keywords);
create index file_contents_search_vector_idx on file_contents using gin(search_vector);

create table path_redirects (
  id uuid primary key default gen_random_uuid(),
  old_path text not null unique,
  new_path text not null,
  node_id uuid not null references nodes(id) on delete cascade,
  created_at timestamptz not null default now()
);

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

create table likes (
  user_id uuid not null references users(id) on delete cascade,
  target_type text not null check (target_type in ('file','comment')),
  target_id uuid not null,
  created_at timestamptz not null default now(),
  primary key(user_id, target_type, target_id)
);

create index likes_target_idx on likes(target_type, target_id);

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

- [ ] **Step 8: Create JSON response helpers**

Create `api/internal/http/response.go`:

```go
package httpx

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error   string         `json:"error"`
	Details map[string]any `json:"details,omitempty"`
}

func JSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, ErrorResponse{Error: message})
}
```

- [ ] **Step 9: Create minimal server entrypoint**

Create `api/cmd/server/main.go`:

```go
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"

	"xlab-blog/api/internal/config"
	"xlab-blog/api/internal/db"
)

func main() {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()
	if err := db.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	log.Printf("api listening on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, mux); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 10: Verify backend tests**

Run:

```bash
cd api && go test ./internal/config
```

Expected: PASS.

- [ ] **Step 11: Commit**

```bash
git add api
git commit -m "Establish the Go API foundation

Constraint: TECH_STACK.md pins Go modules and BACKEND_STRUCTURE.md requires PostgreSQL schema upfront.
Confidence: medium
Scope-risk: moderate
Directive: Keep future SQL in repositories and keep OpenAPI paths as the routing source.
Tested: cd api && go test ./internal/config
Not-tested: Database migration execution against pgvector image waits for docker-compose wiring."
```

## Task 3: Auth, JWT Middleware, and Admin Seed

**Files:**
- Create: `api/internal/users/model.go`
- Create: `api/internal/auth/password.go`
- Create: `api/internal/auth/jwt.go`
- Create: `api/internal/auth/service.go`
- Create: `api/internal/auth/handler.go`
- Create: `api/internal/http/middleware/auth.go`
- Modify: `api/cmd/server/main.go`

- [ ] **Step 1: Write failing password and JWT tests**

Create `api/internal/auth/auth_test.go`:

```go
package auth

import (
	"testing"
	"time"
)

func TestPasswordHashVerifiesAndRejectsWrongPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword error = %v", err)
	}
	if err := CheckPassword(hash, "correct horse battery staple"); err != nil {
		t.Fatalf("CheckPassword correct password error = %v", err)
	}
	if err := CheckPassword(hash, "wrong password"); err == nil {
		t.Fatal("expected wrong password to fail")
	}
}

func TestJWTRejectsTamperedToken(t *testing.T) {
	issuer := JWTIssuer{Secret: []byte("test-secret-with-enough-length"), TTL: time.Hour}
	token, err := issuer.Issue(UserClaims{ID: "user-1", Email: "a@example.com", Role: "reader"})
	if err != nil {
		t.Fatalf("Issue error = %v", err)
	}
	if _, err := issuer.Parse(token + "tampered"); err == nil {
		t.Fatal("expected tampered token to fail")
	}
}
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
cd api && go test ./internal/auth
```

Expected: FAIL because auth package is not implemented.

- [ ] **Step 3: Define user model**

Create `api/internal/users/model.go`:

```go
package users

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	DisplayName  *string   `json:"display_name,omitempty"`
	Provider     string    `json:"provider"`
	ProviderID   *string   `json:"provider_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type PublicUser struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}
```

- [ ] **Step 4: Implement password hashing**

Create `api/internal/auth/password.go`:

```go
package auth

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPassword(hash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
```

- [ ] **Step 5: Implement JWT issue/parse**

Create `api/internal/auth/jwt.go`:

```go
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	ID    string
	Email string
	Role  string
}

type JWTIssuer struct {
	Secret []byte
	TTL    time.Duration
}

type Claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

func (j JWTIssuer) Issue(user UserClaims) (string, error) {
	now := time.Now()
	claims := Claims{
		Email: user.Email,
		Role:  user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.TTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(j.Secret)
}

func (j JWTIssuer) Parse(tokenString string) (UserClaims, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method %s", token.Method.Alg())
		}
		return j.Secret, nil
	})
	if err != nil {
		return UserClaims{}, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return UserClaims{}, fmt.Errorf("invalid token")
	}
	return UserClaims{ID: claims.Subject, Email: claims.Email, Role: claims.Role}, nil
}
```

- [ ] **Step 6: Implement auth service repository boundary**

Create `api/internal/auth/service.go` with a repository interface and business rules:

```go
package auth

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"xlab-blog/api/internal/users"
)

var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrEmailExists = errors.New("email exists")
var ErrWeakPassword = errors.New("password must be at least 8 characters")

type Repository interface {
	CreateUser(ctx context.Context, email string, passwordHash string, role string, displayName *string) (users.User, error)
	FindByEmail(ctx context.Context, email string) (users.User, error)
	FindByID(ctx context.Context, id string) (users.User, error)
	EnsureAdmin(ctx context.Context, email string, passwordHash string) error
}

type Service struct {
	Repo Repository
	JWT  JWTIssuer
}

type AuthResponse struct {
	Token string     `json:"token"`
	User  users.User `json:"user"`
}

func NewService(repo Repository, secret string) Service {
	return Service{Repo: repo, JWT: JWTIssuer{Secret: []byte(secret), TTL: 7 * 24 * time.Hour}}
}

func (s Service) Register(ctx context.Context, email string, password string, displayName *string) (AuthResponse, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !emailPattern.MatchString(email) {
		return AuthResponse{}, ErrInvalidCredentials
	}
	if len(password) < 8 {
		return AuthResponse{}, ErrWeakPassword
	}
	hash, err := HashPassword(password)
	if err != nil {
		return AuthResponse{}, err
	}
	user, err := s.Repo.CreateUser(ctx, email, hash, "reader", displayName)
	if err != nil {
		return AuthResponse{}, err
	}
	token, err := s.JWT.Issue(UserClaims{ID: user.ID, Email: user.Email, Role: user.Role})
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{Token: token, User: user}, nil
}

func (s Service) Login(ctx context.Context, email string, password string) (AuthResponse, error) {
	user, err := s.Repo.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return AuthResponse{}, ErrInvalidCredentials
	}
	if err := CheckPassword(user.PasswordHash, password); err != nil {
		return AuthResponse{}, ErrInvalidCredentials
	}
	token, err := s.JWT.Issue(UserClaims{ID: user.ID, Email: user.Email, Role: user.Role})
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{Token: token, User: user}, nil
}

func (s Service) SeedAdmin(ctx context.Context, email string, password string) error {
	if email == "" || password == "" {
		return nil
	}
	if len(password) < 8 {
		return ErrWeakPassword
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	return s.Repo.EnsureAdmin(ctx, strings.ToLower(strings.TrimSpace(email)), hash)
}
```

- [ ] **Step 7: Implement auth HTTP handler and middleware**

Create `api/internal/auth/handler.go`:

```go
package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	httpx "xlab-blog/api/internal/http"
)

type Handler struct{ Service Service }

type registerRequest struct {
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	DisplayName *string `json:"display_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	res, err := h.Service.Register(r.Context(), req.Email, req.Password, req.DisplayName)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, ErrEmailExists) {
			status = http.StatusConflict
		}
		httpx.Error(w, status, err.Error())
		return
	}
	httpx.JSON(w, http.StatusCreated, res)
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	res, err := h.Service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	httpx.JSON(w, http.StatusOK, res)
}
```

Create `api/internal/http/middleware/auth.go`:

```go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"xlab-blog/api/internal/auth"
	httpx "xlab-blog/api/internal/http"
)

type contextKey string

const userKey contextKey = "viewer"

func OptionalAuth(issuer auth.JWTIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearer(r)
			if token != "" {
				if claims, err := issuer.Parse(token); err == nil {
					r = r.WithContext(context.WithValue(r.Context(), userKey, claims))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := Viewer(r.Context()); !ok {
			httpx.Error(w, http.StatusUnauthorized, "authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireAdmin(next http.Handler) http.Handler {
	return RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		viewer, _ := Viewer(r.Context())
		if viewer.Role != "admin" {
			httpx.Error(w, http.StatusForbidden, "admin required")
			return
		}
		next.ServeHTTP(w, r)
	}))
}

func Viewer(ctx context.Context) (auth.UserClaims, bool) {
	claims, ok := ctx.Value(userKey).(auth.UserClaims)
	return claims, ok
}

func bearer(r *http.Request) string {
	value := r.Header.Get("Authorization")
	if !strings.HasPrefix(value, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(value, "Bearer "))
}
```

- [ ] **Step 8: Verify auth tests**

Run:

```bash
cd api && go test ./internal/auth ./internal/http/middleware
```

Expected: PASS for auth; middleware package has no tests and compiles.

- [ ] **Step 9: Commit**

```bash
git add api/internal/auth api/internal/users api/internal/http/middleware api/cmd/server/main.go
git commit -m "Protect the API with local auth primitives

Constraint: PRD requires email/password readers, env-seeded admin, JWT role guard, and no OAuth routes in the first release.
Confidence: medium
Scope-risk: moderate
Directive: Never log passwords or JWT secrets; public registration must always create reader users.
Tested: cd api && go test ./internal/auth ./internal/http/middleware
Not-tested: Auth handlers are not wired to a real PostgreSQL repository yet."
```

## Task 4: Public Tree, Node CRUD, and Path Redirect Rules

**Files:**
- Create: `api/internal/tree/model.go`
- Create: `api/internal/tree/path.go`
- Create: `api/internal/tree/repository.go`
- Create: `api/internal/tree/service.go`
- Create: `api/internal/tree/handler.go`
- Modify: `api/internal/http/router.go`

- [ ] **Step 1: Write failing path normalization and reserved slug tests**

Create `api/internal/tree/path_test.go`:

```go
package tree

import "testing"

func TestNormalizePath(t *testing.T) {
	cases := map[string]string{
		"":              "/",
		"/":             "/",
		"notes/go/":     "/notes/go",
		"//notes///go/": "/notes/go",
	}
	for input, want := range cases {
		if got := NormalizePath(input); got != want {
			t.Fatalf("NormalizePath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestReservedRootSlug(t *testing.T) {
	if !IsReservedRootSlug("admin") || !IsReservedRootSlug("search") {
		t.Fatal("expected admin and search to be reserved root slugs")
	}
	if IsReservedRootSlug("research") {
		t.Fatal("research must be allowed")
	}
}
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
cd api && go test ./internal/tree
```

Expected: FAIL because `NormalizePath` and `IsReservedRootSlug` are not implemented.

- [ ] **Step 3: Implement tree models and path helpers**

Create `api/internal/tree/model.go`:

```go
package tree

import "time"

type Node struct {
	ID        string     `json:"id"`
	ParentID  *string    `json:"parent_id"`
	Kind      string     `json:"kind"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	Path      string     `json:"path"`
	SortOrder int        `json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type DirectoryEntry struct {
	Node                Node `json:"node"`
	ChildDirectoryCount int  `json:"child_directory_count"`
	ChildFileCount      int  `json:"child_file_count"`
}

type FileEntry struct {
	Node               Node       `json:"node"`
	ContentFormat      string     `json:"content_format"`
	Status             string     `json:"status"`
	Keywords           []string   `json:"keywords"`
	PublishedAt        *time.Time `json:"published_at"`
	LikeCount          int        `json:"like_count"`
	CommentCount       int        `json:"comment_count"`
	ReadingTimeMinutes *int       `json:"reading_time_minutes"`
}

type DirectoryPage struct {
	Node    *Node `json:"node"`
	Path    string `json:"path"`
	Entries []any  `json:"entries"`
}

type ResolveResponse struct {
	Type      string         `json:"type"`
	Directory *DirectoryPage `json:"directory,omitempty"`
	File      any            `json:"file,omitempty"`
	NewPath   string         `json:"new_path,omitempty"`
}
```

Create `api/internal/tree/path.go`:

```go
package tree

import "strings"

var reservedRootSlugs = map[string]struct{}{
	"admin": {}, "api": {}, "auth": {}, "login": {}, "register": {}, "recent": {}, "search": {}, "settings": {},
}

func NormalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/"
	}
	parts := strings.FieldsFunc(path, func(r rune) bool { return r == '/' })
	if len(parts) == 0 {
		return "/"
	}
	return "/" + strings.Join(parts, "/")
}

func SplitPath(path string) []string {
	normalized := NormalizePath(path)
	if normalized == "/" {
		return nil
	}
	return strings.Split(strings.TrimPrefix(normalized, "/"), "/")
}

func IsReservedRootSlug(slug string) bool {
	_, ok := reservedRootSlugs[strings.ToLower(strings.TrimSpace(slug))]
	return ok
}
```

- [ ] **Step 4: Implement service boundary with explicit rules**

Create `api/internal/tree/service.go`:

```go
package tree

import (
	"context"
	"errors"
)

var ErrReservedRootSlug = errors.New("root slug is reserved")
var ErrNotFound = errors.New("node not found")
var ErrCycle = errors.New("move would create a cycle")
var ErrPublishedDescendant = errors.New("directory contains a published file")

type Repository interface {
	Root(ctx context.Context) (DirectoryPage, error)
	ResolveCurrent(ctx context.Context, path string) (ResolveResponse, error)
	ResolveRedirect(ctx context.Context, path string) (string, error)
	CreateNode(ctx context.Context, input CreateNodeInput) (Node, error)
	UpdateNode(ctx context.Context, id string, input UpdateNodeInput) (Node, []PathRedirect, error)
	DeleteNode(ctx context.Context, id string) error
}

type CreateNodeInput struct {
	ParentID      *string
	Kind          string
	Name          string
	Slug          string
	SortOrder     int
	ContentFormat string
}

type UpdateNodeInput struct {
	ParentID  **string
	Name      *string
	Slug      *string
	SortOrder *int
}

type PathRedirect struct {
	ID        string `json:"id"`
	OldPath   string `json:"old_path"`
	NewPath   string `json:"new_path"`
	NodeID    string `json:"node_id"`
	CreatedAt string `json:"created_at"`
}

type Service struct{ Repo Repository }

func (s Service) Root(ctx context.Context) (DirectoryPage, error) {
	return s.Repo.Root(ctx)
}

func (s Service) Resolve(ctx context.Context, path string) (ResolveResponse, error) {
	normalized := NormalizePath(path)
	res, err := s.Repo.ResolveCurrent(ctx, normalized)
	if err == nil {
		return res, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return ResolveResponse{}, err
	}
	newPath, redirectErr := s.Repo.ResolveRedirect(ctx, normalized)
	if redirectErr != nil {
		return ResolveResponse{}, err
	}
	return ResolveResponse{Type: "redirect", NewPath: newPath}, nil
}

func (s Service) CreateNode(ctx context.Context, input CreateNodeInput) (Node, error) {
	if input.ParentID == nil && IsReservedRootSlug(input.Slug) {
		return Node{}, ErrReservedRootSlug
	}
	return s.Repo.CreateNode(ctx, input)
}
```

- [ ] **Step 5: Implement handlers matching OpenAPI names**

Create `api/internal/tree/handler.go`:

```go
package tree

import (
	"encoding/json"
	"errors"
	"net/http"

	httpx "xlab-blog/api/internal/http"
)

type Handler struct{ Service Service }

func (h Handler) Root(w http.ResponseWriter, r *http.Request) {
	page, err := h.Service.Root(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "load root directory")
		return
	}
	httpx.JSON(w, http.StatusOK, page)
}

func (h Handler) Resolve(w http.ResponseWriter, r *http.Request) {
	res, err := h.Service.Resolve(r.Context(), r.URL.Query().Get("path"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrNotFound) {
			status = http.StatusNotFound
		}
		httpx.Error(w, status, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, res)
}

func (h Handler) CreateNode(w http.ResponseWriter, r *http.Request) {
	var input CreateNodeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	node, err := h.Service.CreateNode(r.Context(), input)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, ErrReservedRootSlug) {
			status = http.StatusConflict
		}
		httpx.Error(w, status, err.Error())
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"node": node})
}
```

- [ ] **Step 6: Implement SQL repository after service compiles**

Create `api/internal/tree/repository.go`. The repository must:

```go
package tree

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLRepository struct{ DB *pgxpool.Pool }

func (r SQLRepository) Root(ctx context.Context) (DirectoryPage, error) {
	rows, err := r.DB.Query(ctx, `
		select n.id, n.parent_id, n.kind, n.name, n.slug, '/' || n.slug as path, n.sort_order, n.created_at, n.updated_at
		from nodes n
		left join file_contents fc on fc.node_id = n.id
		where n.parent_id is null and (n.kind='directory' or fc.status='published')
		order by case when n.kind='directory' then 0 else 1 end, n.sort_order asc, n.name asc`)
	if err != nil {
		return DirectoryPage{}, err
	}
	defer rows.Close()
	entries := []any{}
	for rows.Next() {
		var node Node
		if err := rows.Scan(&node.ID, &node.ParentID, &node.Kind, &node.Name, &node.Slug, &node.Path, &node.SortOrder, &node.CreatedAt, &node.UpdatedAt); err != nil {
			return DirectoryPage{}, err
		}
		if node.Kind == "directory" {
			entries = append(entries, DirectoryEntry{Node: node})
		} else {
			entries = append(entries, FileEntry{Node: node, Status: "published"})
		}
	}
	return DirectoryPage{Node: nil, Path: "/", Entries: entries}, rows.Err()
}

func (r SQLRepository) ResolveCurrent(ctx context.Context, path string) (ResolveResponse, error) {
	if path == "/" {
		page, err := r.Root(ctx)
		if err != nil {
			return ResolveResponse{}, err
		}
		return ResolveResponse{Type: "directory", Directory: &page}, nil
	}
	return ResolveResponse{}, ErrNotFound
}

func (r SQLRepository) ResolveRedirect(ctx context.Context, path string) (string, error) {
	var newPath string
	err := r.DB.QueryRow(ctx, `select new_path from path_redirects where old_path=$1`, path).Scan(&newPath)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return newPath, err
}

func (r SQLRepository) CreateNode(ctx context.Context, input CreateNodeInput) (Node, error) {
	var node Node
	err := r.DB.QueryRow(ctx, `
		insert into nodes(parent_id, kind, name, slug, sort_order)
		values($1,$2,$3,$4,$5)
		returning id, parent_id, kind, name, slug, '/' || slug as path, sort_order, created_at, updated_at`,
		input.ParentID, input.Kind, input.Name, input.Slug, input.SortOrder,
	).Scan(&node.ID, &node.ParentID, &node.Kind, &node.Name, &node.Slug, &node.Path, &node.SortOrder, &node.CreatedAt, &node.UpdatedAt)
	return node, err
}

func (r SQLRepository) UpdateNode(ctx context.Context, id string, input UpdateNodeInput) (Node, []PathRedirect, error) {
	return Node{}, nil, errors.New("node update repository must be implemented in the path redirect task before admin editing is enabled")
}

func (r SQLRepository) DeleteNode(ctx context.Context, id string) error {
	return errors.New("node delete repository must be implemented in the deletion rule task before admin deletion is enabled")
}
```

Before committing this task, replace the two temporary `errors.New(...)` bodies with real SQL implementations or do not register update/delete routes. Public root and create route can ship while update/delete stay unregistered. Do not expose a route that returns the temporary error text.

- [ ] **Step 7: Verify tree helper tests**

Run:

```bash
cd api && go test ./internal/tree
```

Expected: PASS for path helper tests.

- [ ] **Step 8: Commit**

```bash
git add api/internal/tree api/internal/http/router.go
git commit -m "Model the public content tree around paths

Constraint: PRD replaces flat posts/categories with a Unix-like Directory/File tree and reserved root routes.
Confidence: medium
Scope-risk: moderate
Directive: Do not expose admin update/delete routes until redirect and deletion rules are implemented.
Tested: cd api && go test ./internal/tree
Not-tested: PostgreSQL-backed root query awaits integration testing with seeded data."
```

## Task 5: Render Pipeline for Markdown and HTML Document Search Text

**Files:**
- Create: `api/internal/render/markdown.go`
- Create: `api/internal/render/html_text.go`
- Create: `api/internal/render/reading_time.go`
- Create: `web/src/components/MarkdownRenderer.tsx`
- Create: `web/src/components/HtmlDocumentFrame.tsx`

- [ ] **Step 1: Write failing backend render tests**

Create `api/internal/render/render_test.go`:

```go
package render

import "testing"

func TestMarkdownSanitizesScriptAndEvents(t *testing.T) {
	html, text, err := MarkdownToSafeHTML("# Hi\n\n<img src=x onerror=alert(1)>\n<script>alert(1)</script>")
	if err != nil {
		t.Fatalf("MarkdownToSafeHTML error = %v", err)
	}
	if containsAny(html, []string{"<script", "onerror", "alert(1)"}) {
		t.Fatalf("unsafe html survived: %s", html)
	}
	if text == "" {
		t.Fatal("expected search text")
	}
}

func TestHTMLVisibleTextExcludesScriptStyleAndHidden(t *testing.T) {
	input := `<html><head><style>.x{}</style><script>secret()</script></head><body><h1>Visible</h1><p hidden>Hidden</p><noscript>No</noscript><p>Text</p></body></html>`
	got := VisibleTextFromHTML(input)
	if got != "Visible Text" {
		t.Fatalf("VisibleTextFromHTML = %q, want Visible Text", got)
	}
}

func containsAny(s string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(s, needle) {
			return true
		}
	}
	return false
}
```

Add `import "strings"` to the test file after the first failure reports it missing.

- [ ] **Step 2: Run backend render tests to verify failure**

Run:

```bash
cd api && go test ./internal/render
```

Expected: FAIL because render functions do not exist.

- [ ] **Step 3: Implement Markdown sanitization and plain text extraction**

Create `api/internal/render/markdown.go`:

```go
package render

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
)

var tagPattern = regexp.MustCompile(`<[^>]+>`)
var spacePattern = regexp.MustCompile(`\s+`)

func MarkdownToSafeHTML(markdown string) (safeHTML string, searchText string, err error) {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(markdown), &buf); err != nil {
		return "", "", err
	}
	policy := bluemonday.UGCPolicy()
	policy.RequireNoFollowOnLinks(false)
	safe := policy.Sanitize(buf.String())
	text := tagPattern.ReplaceAllString(safe, " ")
	text = strings.TrimSpace(spacePattern.ReplaceAllString(text, " "))
	return safe, text, nil
}
```

- [ ] **Step 4: Implement HTML visible text extraction**

Create `api/internal/render/html_text.go`:

```go
package render

import (
	"strings"

	"golang.org/x/net/html"
)

func VisibleTextFromHTML(document string) string {
	root, err := html.Parse(strings.NewReader(document))
	if err != nil {
		return ""
	}
	var parts []string
	walkVisible(root, false, &parts)
	return strings.Join(parts, " ")
}

func walkVisible(n *html.Node, hidden bool, parts *[]string) {
	if n.Type == html.ElementNode {
		name := strings.ToLower(n.Data)
		if name == "script" || name == "style" || name == "meta" || name == "link" || name == "noscript" {
			return
		}
		for _, attr := range n.Attr {
			key := strings.ToLower(attr.Key)
			value := strings.ToLower(attr.Val)
			if key == "hidden" || (key == "aria-hidden" && value == "true") {
				hidden = true
			}
			if key == "style" && strings.Contains(value, "display:none") {
				hidden = true
			}
		}
	}
	if n.Type == html.TextNode && !hidden {
		text := strings.TrimSpace(spacePattern.ReplaceAllString(n.Data, " "))
		if text != "" {
			*parts = append(*parts, text)
		}
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		walkVisible(child, hidden, parts)
	}
}
```

- [ ] **Step 5: Implement read time**

Create `api/internal/render/reading_time.go`:

```go
package render

import "strings"

func ReadingTimeMinutes(text string) int {
	words := len(strings.Fields(text))
	if words == 0 {
		return 1
	}
	minutes := words / 220
	if words%220 != 0 {
		minutes++
	}
	if minutes < 1 {
		return 1
	}
	return minutes
}
```

- [ ] **Step 6: Verify backend render tests**

Run:

```bash
cd api && go test ./internal/render
```

Expected: PASS.

- [ ] **Step 7: Add frontend iframe sandbox component**

Create `web/src/components/HtmlDocumentFrame.tsx` after web scaffold exists:

```tsx
export function HtmlDocumentFrame({ html, title }: { html: string; title: string }) {
  return (
    <iframe
      className="html-document-frame"
      title={title}
      sandbox="allow-scripts"
      srcDoc={html}
    />
  );
}
```

- [ ] **Step 8: Add frontend Markdown renderer component**

Create `web/src/components/MarkdownRenderer.tsx` after web scaffold exists:

```tsx
import DOMPurify from 'dompurify';
import { marked } from 'marked';

export function MarkdownRenderer({ markdown }: { markdown: string }) {
  const dirty = marked.parse(markdown, { async: false }) as string;
  const safe = DOMPurify.sanitize(dirty, { USE_PROFILES: { html: true } });
  return <div className="markdown-body" dangerouslySetInnerHTML={{ __html: safe }} />;
}
```

- [ ] **Step 9: Commit backend render pipeline first**

```bash
git add api/internal/render
git commit -m "Make file rendering safe before exposing content

Constraint: PRD requires sanitized Markdown and iframe-only HTML documents with searchable visible text.
Confidence: high
Scope-risk: narrow
Directive: Never inject HTML Document body into the main DOM; iframe sandbox must remain allow-scripts only.
Tested: cd api && go test ./internal/render
Not-tested: Frontend components are added after web scaffolding in the frontend task."
```

## Task 6: Content Service, Publish Rules, and Search Index State

**Files:**
- Create: `api/internal/content/model.go`
- Create: `api/internal/content/repository.go`
- Create: `api/internal/content/service.go`
- Create: `api/internal/content/handler.go`
- Modify: `api/internal/tree/repository.go`

- [ ] **Step 1: Write failing content-format rule test**

Create `api/internal/content/service_test.go`:

```go
package content

import (
	"context"
	"testing"
)

type fakeRepo struct{ existing FileContent }

func (f fakeRepo) Get(ctx context.Context, fileID string) (FileContent, error) { return f.existing, nil }
func (f fakeRepo) Upsert(ctx context.Context, input UpsertInput) (FileContent, error) { return FileContent{NodeID: input.FileID, ContentFormat: input.ContentFormat, BodyRaw: input.BodyRaw, Keywords: input.Keywords, Status: "draft"}, nil }
func (f fakeRepo) Publish(ctx context.Context, fileID string) (FileContent, error) { return f.existing, nil }
func (f fakeRepo) Unpublish(ctx context.Context, fileID string) (FileContent, error) { return f.existing, nil }

func TestPublishedFileCannotChangeContentFormat(t *testing.T) {
	svc := Service{Repo: fakeRepo{existing: FileContent{NodeID: "file-1", ContentFormat: "markdown", Status: "published"}}}
	_, err := svc.Upsert(context.Background(), UpsertInput{FileID: "file-1", ContentFormat: "html_document", BodyRaw: "<html></html>", Keywords: []string{}})
	if err != ErrPublishedFormatChange {
		t.Fatalf("err = %v, want ErrPublishedFormatChange", err)
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run:

```bash
cd api && go test ./internal/content
```

Expected: FAIL because content package is missing.

- [ ] **Step 3: Implement content model and service rules**

Create `api/internal/content/model.go`:

```go
package content

import "time"

type FileContent struct {
	NodeID             string     `json:"node_id"`
	ContentFormat      string     `json:"content_format"`
	Keywords           []string   `json:"keywords"`
	BodyRaw            string     `json:"body_raw"`
	BodyHTML           *string    `json:"body_html"`
	SearchText         string     `json:"search_text"`
	Status             string     `json:"status"`
	PublishedAt        *time.Time `json:"published_at"`
	EmbeddingModel     *string    `json:"embedding_model"`
	EmbeddingStatus    string     `json:"embedding_status"`
	EmbeddingError     *string    `json:"embedding_error"`
	EmbeddingUpdatedAt *time.Time `json:"embedding_updated_at"`
}

type UpsertInput struct {
	FileID        string
	ContentFormat string
	BodyRaw       string
	Keywords      []string
}
```

Create `api/internal/content/service.go`:

```go
package content

import (
	"context"
	"errors"

	"xlab-blog/api/internal/render"
)

var ErrPublishedFormatChange = errors.New("published file cannot change content_format")

type Repository interface {
	Get(ctx context.Context, fileID string) (FileContent, error)
	Upsert(ctx context.Context, input UpsertInput) (FileContent, error)
	Publish(ctx context.Context, fileID string) (FileContent, error)
	Unpublish(ctx context.Context, fileID string) (FileContent, error)
}

type Service struct{ Repo Repository }

func (s Service) Upsert(ctx context.Context, input UpsertInput) (FileContent, error) {
	existing, err := s.Repo.Get(ctx, input.FileID)
	if err == nil && existing.Status == "published" && existing.ContentFormat != input.ContentFormat {
		return FileContent{}, ErrPublishedFormatChange
	}
	if input.ContentFormat == "markdown" {
		_, text, err := render.MarkdownToSafeHTML(input.BodyRaw)
		if err != nil {
			return FileContent{}, err
		}
		input.BodyRaw = input.BodyRaw
		_ = text
	}
	if input.ContentFormat == "html_document" {
		_ = render.VisibleTextFromHTML(input.BodyRaw)
	}
	return s.Repo.Upsert(ctx, input)
}

func (s Service) Publish(ctx context.Context, fileID string) (FileContent, error) {
	return s.Repo.Publish(ctx, fileID)
}

func (s Service) Unpublish(ctx context.Context, fileID string) (FileContent, error) {
	return s.Repo.Unpublish(ctx, fileID)
}
```

- [ ] **Step 4: Implement repository upsert with search_text and embedding pending**

Create `api/internal/content/repository.go` with SQL that computes `body_html` and `search_text` before writing. The upsert query must set `embedding_status='pending'`, clear `embedding_error`, and preserve `published_at` unless publishing.

Use this exact upsert SQL shape:

```sql
insert into file_contents(node_id, content_format, keywords, body_raw, body_html, search_text, embedding_status, embedding_error)
values($1,$2,$3,$4,$5,$6,'pending',null)
on conflict(node_id) do update set
  content_format=excluded.content_format,
  keywords=excluded.keywords,
  body_raw=excluded.body_raw,
  body_html=excluded.body_html,
  search_text=excluded.search_text,
  embedding_status='pending',
  embedding_error=null
returning node_id, content_format, keywords, body_raw, body_html, search_text, status, published_at, embedding_model, embedding_status, embedding_error, embedding_updated_at
```

- [ ] **Step 5: Implement publish/unpublish SQL**

Use these exact state transitions:

```sql
update file_contents
set status='published', published_at=coalesce(published_at, now())
where node_id=$1
returning node_id, content_format, keywords, body_raw, body_html, search_text, status, published_at, embedding_model, embedding_status, embedding_error, embedding_updated_at
```

```sql
update file_contents
set status='draft'
where node_id=$1
returning node_id, content_format, keywords, body_raw, body_html, search_text, status, published_at, embedding_model, embedding_status, embedding_error, embedding_updated_at
```

- [ ] **Step 6: Verify content service tests**

Run:

```bash
cd api && go test ./internal/content ./internal/render
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add api/internal/content api/internal/render
git commit -m "Keep file content searchable and publish-safe

Constraint: Published files cannot directly change render format, and every save must refresh search_text plus embedding state.
Confidence: medium
Scope-risk: moderate
Directive: Qwen failures must be recorded on embedding fields and must not roll back content saves.
Tested: cd api && go test ./internal/content ./internal/render
Not-tested: Full SQL repository behavior requires Docker database integration tests."
```

## Task 7: Likes and Comments Vertical Slice

**Files:**
- Create: `api/internal/likes/{model.go,repository.go,service.go,handler.go}`
- Create: `api/internal/comments/{model.go,repository.go,service.go,handler.go}`
- Modify: `api/internal/http/router.go`

- [ ] **Step 1: Write failing like idempotency service test**

Create `api/internal/likes/service_test.go`:

```go
package likes

import (
	"context"
	"testing"
)

type memoryRepo struct{ liked bool }

func (m *memoryRepo) Like(ctx context.Context, userID string, targetType string, targetID string) (State, error) {
	m.liked = true
	return State{Liked: true, LikeCount: 1}, nil
}
func (m *memoryRepo) Unlike(ctx context.Context, userID string, targetType string, targetID string) (State, error) {
	m.liked = false
	return State{Liked: false, LikeCount: 0}, nil
}

func TestLikeAndUnlikeAreIdempotentAtServiceBoundary(t *testing.T) {
	repo := &memoryRepo{}
	svc := Service{Repo: repo}
	first, err := svc.Like(context.Background(), "user-1", "file", "file-1")
	if err != nil || !first.Liked || first.LikeCount != 1 {
		t.Fatalf("first like = %+v err=%v", first, err)
	}
	second, err := svc.Unlike(context.Background(), "user-1", "file", "file-1")
	if err != nil || second.Liked || second.LikeCount != 0 {
		t.Fatalf("unlike = %+v err=%v", second, err)
	}
}
```

- [ ] **Step 2: Implement likes model/service/repository SQL**

Create `api/internal/likes/model.go`:

```go
package likes

type State struct {
	Liked     bool `json:"liked"`
	LikeCount int  `json:"like_count"`
}
```

Create `api/internal/likes/service.go`:

```go
package likes

import "context"

type Repository interface {
	Like(ctx context.Context, userID string, targetType string, targetID string) (State, error)
	Unlike(ctx context.Context, userID string, targetType string, targetID string) (State, error)
}

type Service struct{ Repo Repository }

func (s Service) Like(ctx context.Context, userID string, targetType string, targetID string) (State, error) {
	return s.Repo.Like(ctx, userID, targetType, targetID)
}

func (s Service) Unlike(ctx context.Context, userID string, targetType string, targetID string) (State, error) {
	return s.Repo.Unlike(ctx, userID, targetType, targetID)
}
```

Repository SQL for like must use:

```sql
insert into likes(user_id, target_type, target_id)
values($1,$2,$3)
on conflict(user_id, target_type, target_id) do nothing
```

Repository SQL for unlike must use:

```sql
delete from likes where user_id=$1 and target_type=$2 and target_id=$3
```

Count SQL must use:

```sql
select count(*) from likes where target_type=$1 and target_id=$2
```

- [ ] **Step 3: Write failing comment normalization test**

Create `api/internal/comments/service_test.go`:

```go
package comments

import (
	"context"
	"testing"
)

type fakeCommentRepo struct{ parent Comment }

func (f fakeCommentRepo) Get(ctx context.Context, id string) (Comment, error) { return f.parent, nil }
func (f fakeCommentRepo) Create(ctx context.Context, input CreateInput) (Comment, error) {
	return Comment{ID: "new", ParentID: input.ParentID, ReplyToUserID: input.ReplyToUserID, Body: input.Body}, nil
}
func (f fakeCommentRepo) Thread(ctx context.Context, fileID string) ([]Comment, error) { return nil, nil }
func (f fakeCommentRepo) SoftDelete(ctx context.Context, id string, deletedBy string, isAdmin bool) error { return nil }

func TestReplyToReplyNormalizesToTopLevelParent(t *testing.T) {
	top := "top-1"
	replyAuthor := "user-2"
	svc := Service{Repo: fakeCommentRepo{parent: Comment{ID: "reply-1", ParentID: &top, User: PublicUser{ID: replyAuthor, DisplayName: "Reply Author"}}}}
	created, err := svc.Create(context.Background(), CreateInput{FileID: "file-1", UserID: "user-1", ParentID: ptr("reply-1"), Body: "answer"})
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}
	if created.ParentID == nil || *created.ParentID != top {
		t.Fatalf("ParentID = %v, want %s", created.ParentID, top)
	}
	if created.ReplyToUserID == nil || *created.ReplyToUserID != replyAuthor {
		t.Fatalf("ReplyToUserID = %v, want %s", created.ReplyToUserID, replyAuthor)
	}
}

func ptr(s string) *string { return &s }
```

- [ ] **Step 4: Implement comments model/service**

Create `api/internal/comments/model.go`:

```go
package comments

import "time"

type PublicUser struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type Comment struct {
	ID            string     `json:"id"`
	FileNodeID    string     `json:"file_node_id"`
	ParentID      *string    `json:"parent_id"`
	ReplyToUserID *string    `json:"reply_to_user_id"`
	User          PublicUser `json:"user"`
	Body          string     `json:"body"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
	Deleted       bool       `json:"deleted"`
	LikeCount     int        `json:"like_count"`
	Replies       []Comment  `json:"replies"`
}

type CreateInput struct {
	FileID        string
	UserID        string
	ParentID      *string
	ReplyToUserID *string
	Body          string
}
```

Create `api/internal/comments/service.go`:

```go
package comments

import (
	"context"
	"errors"
	"strings"
)

var ErrEmptyComment = errors.New("comment body is required")
var ErrCommentTooLong = errors.New("comment body exceeds 5000 characters")

type Repository interface {
	Get(ctx context.Context, id string) (Comment, error)
	Create(ctx context.Context, input CreateInput) (Comment, error)
	Thread(ctx context.Context, fileID string) ([]Comment, error)
	SoftDelete(ctx context.Context, id string, deletedBy string, isAdmin bool) error
}

type Service struct{ Repo Repository }

func (s Service) Create(ctx context.Context, input CreateInput) (Comment, error) {
	input.Body = strings.TrimSpace(input.Body)
	if input.Body == "" {
		return Comment{}, ErrEmptyComment
	}
	if len([]rune(input.Body)) > 5000 {
		return Comment{}, ErrCommentTooLong
	}
	if input.ParentID != nil {
		parent, err := s.Repo.Get(ctx, *input.ParentID)
		if err != nil {
			return Comment{}, err
		}
		if parent.ParentID != nil {
			input.ParentID = parent.ParentID
			input.ReplyToUserID = &parent.User.ID
		}
	}
	return s.Repo.Create(ctx, input)
}

func (s Service) Thread(ctx context.Context, fileID string) ([]Comment, error) {
	return s.Repo.Thread(ctx, fileID)
}

func (s Service) SoftDelete(ctx context.Context, id string, deletedBy string, isAdmin bool) error {
	return s.Repo.SoftDelete(ctx, id, deletedBy, isAdmin)
}
```

- [ ] **Step 5: Verify likes/comments tests**

Run:

```bash
cd api && go test ./internal/likes ./internal/comments
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add api/internal/likes api/internal/comments api/internal/http/router.go
git commit -m "Enable reader interaction without anonymous writes

Constraint: PRD requires anonymous read, authenticated comments/likes, two-level comments, soft deletion, and idempotent likes.
Confidence: medium
Scope-risk: moderate
Directive: Keep comment body plain text and normalize reply-to-reply under the top-level parent.
Tested: cd api && go test ./internal/likes ./internal/comments
Not-tested: SQL repository integration and auth-wired handlers require full router/database integration."
```

## Task 8: Assets with MIME, Size, SVG Safety, and Local Storage

**Files:**
- Create: `api/internal/assets/model.go`
- Create: `api/internal/assets/storage.go`
- Create: `api/internal/assets/local.go`
- Create: `api/internal/assets/validation.go`
- Create: `api/internal/assets/repository.go`
- Create: `api/internal/assets/service.go`
- Create: `api/internal/assets/handler.go`

- [ ] **Step 1: Write failing malicious SVG test**

Create `api/internal/assets/validation_test.go`:

```go
package assets

import "testing"

func TestRejectsMaliciousSVG(t *testing.T) {
	bad := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script><rect onclick="alert(1)"/><image href="https://evil.example/x.png"/><foreignObject></foreignObject></svg>`)
	if err := ValidateSVG(bad); err == nil {
		t.Fatal("expected malicious SVG to be rejected")
	}
}

func TestAllowsSimpleSVG(t *testing.T) {
	good := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><rect width="10" height="10" fill="#0066cc"/></svg>`)
	if err := ValidateSVG(good); err != nil {
		t.Fatalf("ValidateSVG good SVG error = %v", err)
	}
}
```

- [ ] **Step 2: Implement SVG validation**

Create `api/internal/assets/validation.go`:

```go
package assets

import (
	"bytes"
	"errors"
	"strings"

	"golang.org/x/net/html"
)

var ErrUnsafeSVG = errors.New("unsafe svg")

func ValidateSVG(data []byte) error {
	root, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return err
	}
	unsafe := false
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if unsafe || n == nil {
			return
		}
		if n.Type == html.ElementNode {
			name := strings.ToLower(n.Data)
			if name == "script" || name == "foreignobject" {
				unsafe = true
				return
			}
			for _, attr := range n.Attr {
				key := strings.ToLower(attr.Key)
				value := strings.ToLower(strings.TrimSpace(attr.Val))
				if strings.HasPrefix(key, "on") || strings.Contains(value, "javascript:") {
					unsafe = true
					return
				}
				if (key == "href" || key == "src" || strings.HasSuffix(key, ":href")) && (strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "//")) {
					unsafe = true
					return
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	if unsafe {
		return ErrUnsafeSVG
	}
	return nil
}
```

- [ ] **Step 3: Implement asset model and storage interface**

Create `api/internal/assets/model.go`:

```go
package assets

import "time"

type FileAsset struct {
	ID              string    `json:"id"`
	FileNodeID      string    `json:"file_node_id"`
	Filename        string    `json:"filename"`
	MimeType        string    `json:"mime_type"`
	SizeBytes       int64     `json:"size_bytes"`
	StorageProvider string    `json:"storage_provider"`
	StorageKey      string    `json:"storage_key"`
	PublicURL       string    `json:"public_url"`
	CreatedAt       time.Time `json:"created_at"`
}
```

Create `api/internal/assets/storage.go`:

```go
package assets

import (
	"context"
	"io"
)

type Storage interface {
	Save(ctx context.Context, key string, r io.Reader) error
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}
```

Create `api/internal/assets/local.go`:

```go
package assets

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct{ Root string }

func (l LocalStorage) Save(ctx context.Context, key string, r io.Reader) error {
	path := l.safePath(key)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, r)
	return err
}

func (l LocalStorage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	return os.Open(l.safePath(key))
}

func (l LocalStorage) Delete(ctx context.Context, key string) error {
	return os.Remove(l.safePath(key))
}

func (l LocalStorage) safePath(key string) string {
	clean := filepath.Clean("/" + strings.TrimPrefix(key, "/"))
	return filepath.Join(l.Root, strings.TrimPrefix(clean, "/"))
}
```

- [ ] **Step 4: Verify asset validation tests**

Run:

```bash
cd api && go test ./internal/assets
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add api/internal/assets
git commit -m "Constrain file assets before making them public

Constraint: PRD requires per-file assets with local volume storage, immutable public URLs, MIME limits, and SVG safety checks.
Confidence: high
Scope-risk: moderate
Directive: Never expose local filesystem paths; public URLs must use asset id and filename only.
Tested: cd api && go test ./internal/assets
Not-tested: Multipart upload and public serving require router/database integration."
```

## Task 9: Hybrid Search with Full-Text, Qwen Embeddings, and RRF

**Files:**
- Create: `api/internal/search/model.go`
- Create: `api/internal/search/embedding.go`
- Create: `api/internal/search/qwen.go`
- Create: `api/internal/search/repository.go`
- Create: `api/internal/search/service.go`
- Create: `api/internal/search/handler.go`

- [ ] **Step 1: Write failing RRF unit test**

Create `api/internal/search/service_test.go`:

```go
package search

import "testing"

func TestRRFCombinesTextAndSemanticRanks(t *testing.T) {
	items := FuseRRF([]RankedID{{ID: "a", Rank: 1}, {ID: "b", Rank: 2}}, []RankedID{{ID: "b", Rank: 1}, {ID: "c", Rank: 2}}, 60)
	if len(items) != 3 {
		t.Fatalf("len(items)=%d, want 3", len(items))
	}
	if items[0].ID != "b" {
		t.Fatalf("top item=%s, want b because it appears in both lists", items[0].ID)
	}
	if !items[0].Sources["text"] || !items[0].Sources["semantic"] {
		t.Fatalf("sources=%v, want text and semantic", items[0].Sources)
	}
}
```

- [ ] **Step 2: Implement search models and RRF**

Create `api/internal/search/model.go`:

```go
package search

type RankedID struct {
	ID   string
	Rank int
}

type FusedCandidate struct {
	ID      string
	Score   float64
	Sources map[string]bool
}

type Result struct {
	File         any      `json:"file"`
	Path         string   `json:"path"`
	Snippet      string   `json:"snippet"`
	Score        float64  `json:"score"`
	MatchSources []string `json:"match_sources"`
}
```

Create `api/internal/search/service.go`:

```go
package search

import (
	"context"
	"sort"
)

type Repository interface {
	FullText(ctx context.Context, query string, limit int) ([]RankedID, error)
	Semantic(ctx context.Context, vector []float32, limit int) ([]RankedID, error)
	Hydrate(ctx context.Context, fused []FusedCandidate, limit int, offset int) ([]Result, error)
}

type EmbeddingProvider interface {
	EmbedText(ctx context.Context, input string) ([]float32, error)
	ModelName() string
	Dimensions() int
}

type Service struct {
	Repo      Repository
	Embedding EmbeddingProvider
}

func FuseRRF(text []RankedID, semantic []RankedID, k int) []FusedCandidate {
	byID := map[string]*FusedCandidate{}
	add := func(list []RankedID, source string) {
		for _, item := range list {
			candidate := byID[item.ID]
			if candidate == nil {
				candidate = &FusedCandidate{ID: item.ID, Sources: map[string]bool{}}
				byID[item.ID] = candidate
			}
			candidate.Score += 1.0 / float64(k+item.Rank)
			candidate.Sources[source] = true
		}
	}
	add(text, "text")
	add(semantic, "semantic")
	out := make([]FusedCandidate, 0, len(byID))
	for _, candidate := range byID {
		out = append(out, *candidate)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	return out
}
```

- [ ] **Step 3: Implement Qwen OpenAI-compatible request**

Create `api/internal/search/qwen.go`:

```go
package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type QwenProvider struct {
	APIKey     string
	BaseURL    string
	Model      string
	Dimensions int
	Client     *http.Client
}

type embeddingRequest struct {
	Model          string `json:"model"`
	Input          string `json:"input"`
	Dimensions     int    `json:"dimensions"`
	EncodingFormat string `json:"encoding_format"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

func (q QwenProvider) ModelName() string { return q.Model }
func (q QwenProvider) Dimensions() int   { return q.Dimensions }

func (q QwenProvider) EmbedText(ctx context.Context, input string) ([]float32, error) {
	body, err := json.Marshal(embeddingRequest{Model: q.Model, Input: input, Dimensions: q.Dimensions, EncodingFormat: "float"})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, q.BaseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+q.APIKey)
	req.Header.Set("Content-Type", "application/json")
	client := q.Client
	if client == nil {
		client = http.DefaultClient
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("embedding status %d", res.StatusCode)
	}
	var decoded embeddingResponse
	if err := json.NewDecoder(res.Body).Decode(&decoded); err != nil {
		return nil, err
	}
	if len(decoded.Data) == 0 || len(decoded.Data[0].Embedding) != q.Dimensions {
		return nil, fmt.Errorf("embedding dimensions mismatch")
	}
	return decoded.Data[0].Embedding, nil
}
```

- [ ] **Step 4: Verify search tests**

Run:

```bash
cd api && go test ./internal/search
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add api/internal/search
git commit -m "Make hybrid search deterministic around RRF

Constraint: Search must combine PostgreSQL full-text and Qwen text-embedding-v4 semantic retrieval without LLM expansion or reranking.
Confidence: medium
Scope-risk: moderate
Directive: Embedding request bodies must always include dimensions 1024 and encoding_format float.
Tested: cd api && go test ./internal/search
Not-tested: Live DashScope calls and pgvector ranking require configured credentials and Docker database."
```

## Task 10: Frontend Scaffold, Glass Tokens, Routing, and API Client

**Files:**
- Create: `web/package.json`
- Create: `web/index.html`
- Create: `web/src/main.tsx`
- Create: `web/src/app/App.tsx`
- Create: `web/src/api/client.ts`
- Create: `web/src/api/types.ts`
- Create: `web/src/styles/tokens.css`
- Create: `web/src/styles/glass.css`
- Create: `web/src/styles/app.css`

- [ ] **Step 1: Scaffold Vite React TypeScript with exact versions**

Run:

```bash
npm create vite@9.0.7 web -- --template react-ts
cd web
npm install --save-exact react@19.2.7 react-dom@19.2.7 react-router-dom@7.16.0 @tanstack/react-query@5.101.0 dompurify@3.4.7 marked@18.0.4 lucide-react@1.17.0 zod@4.4.3
npm install --save-dev --save-exact @vitejs/plugin-react@6.0.2 vite@8.0.16 typescript@6.0.3 @types/react@19.2.16 @types/react-dom@19.2.3 @types/dompurify@3.2.0 eslint@10.4.1 typescript-eslint@8.60.1 eslint-plugin-react-hooks@7.1.1 eslint-plugin-react-refresh@0.5.2 prettier@3.8.3
```

Expected: `web/package.json` contains exact versions without `^` or `~`.

- [ ] **Step 2: Replace default app with route shell**

Create `web/src/app/App.tsx`:

```tsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import { AppShell } from '../components/AppShell';
import { RootDirectoryPage } from '../routes/RootDirectoryPage';
import { RecentPage } from '../routes/RecentPage';
import { SearchPage } from '../routes/SearchPage';
import { LoginPage } from '../routes/LoginPage';
import { RegisterPage } from '../routes/RegisterPage';
import { AdminPage } from '../routes/AdminPage';
import { ContentPathPage } from '../routes/ContentPathPage';

const queryClient = new QueryClient();

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AppShell>
          <Routes>
            <Route path="/" element={<RootDirectoryPage />} />
            <Route path="/recent" element={<RecentPage />} />
            <Route path="/search" element={<SearchPage />} />
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/admin" element={<AdminPage />} />
            <Route path="/settings" element={<Navigate to="/recent" replace />} />
            <Route path="/*" element={<ContentPathPage />} />
          </Routes>
        </AppShell>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
```

Create `web/src/main.tsx`:

```tsx
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { App } from './app/App';
import './styles/tokens.css';
import './styles/glass.css';
import './styles/app.css';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
```

- [ ] **Step 3: Create typed API client**

Create `web/src/api/client.ts`:

```ts
const API_BASE = '/api';
const TOKEN_KEY = 'xlab-blog-token';

export function getToken() {
  return window.localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string | null) {
  if (token) window.localStorage.setItem(TOKEN_KEY, token);
  else window.localStorage.removeItem(TOKEN_KEY);
}

export async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
  const token = getToken();
  const headers = new Headers(init.headers);
  if (!headers.has('Content-Type') && init.body && !(init.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }
  if (token) headers.set('Authorization', `Bearer ${token}`);
  const response = await fetch(`${API_BASE}${path}`, { ...init, headers });
  if (!response.ok) {
    let message = `HTTP ${response.status}`;
    try {
      const body = (await response.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      message = `HTTP ${response.status}`;
    }
    throw new Error(message);
  }
  if (response.status === 204) return undefined as T;
  return (await response.json()) as T;
}
```

Create `web/src/api/types.ts` with the OpenAPI-aligned core types:

```ts
export type Role = 'admin' | 'reader';
export type NodeKind = 'directory' | 'file';
export type ContentFormat = 'markdown' | 'html_document';
export type PublishStatus = 'draft' | 'published';

export interface User {
  id: string;
  email: string;
  display_name?: string;
  role: Role;
  provider: 'local' | 'github';
  created_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface Node {
  id: string;
  parent_id?: string | null;
  kind: NodeKind;
  name: string;
  slug: string;
  path: string;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface DirectoryEntry {
  node: Node;
  child_directory_count: number;
  child_file_count: number;
}

export interface FileEntry {
  node: Node;
  content_format: ContentFormat;
  status: PublishStatus;
  keywords: string[];
  published_at?: string | null;
  like_count: number;
  comment_count: number;
  reading_time_minutes?: number | null;
}

export interface DirectoryPage {
  node: Node | null;
  path: string;
  entries: Array<DirectoryEntry | FileEntry>;
}

export interface FileContent {
  node_id: string;
  content_format: ContentFormat;
  keywords: string[];
  body_raw: string;
  body_html?: string | null;
  search_text: string;
  status: PublishStatus;
  published_at?: string | null;
  embedding_status: 'pending' | 'ready' | 'failed';
  embedding_error?: string | null;
}

export interface FilePage {
  node: Node;
  content: FileContent;
  keywords_public: string[];
  like_count: number;
  viewer_has_liked?: boolean;
  comment_count: number;
  assets: FileAsset[];
}

export interface FileAsset {
  id: string;
  file_node_id: string;
  filename: string;
  mime_type: string;
  size_bytes: number;
  storage_provider: 'local' | 's3' | 'r2' | 'oss';
  storage_key: string;
  public_url: string;
  created_at: string;
}

export type TreeResolveResponse =
  | { type: 'directory'; directory: DirectoryPage }
  | { type: 'file'; file: FilePage }
  | { type: 'redirect'; new_path: string };
```

- [ ] **Step 4: Create glass-ricepaper tokens and primitive**

Create `web/src/styles/tokens.css`:

```css
:root {
  --primary: #0066cc;
  --primary-focus: #0071e3;
  --ink: #26221c;
  --body: #3a342b;
  --ink-60: #6b6256;
  --ink-40: #9a9082;
  --on-primary: #ffffff;
  --canvas-paper-1: #f4ecdc;
  --canvas-paper-2: #efe6d3;
  --paper-warm-glow: #f6e4be;
  --glass-fill: rgba(255, 253, 247, 0.38);
  --glass-fill-button: rgba(255, 253, 247, 0.42);
  --shadow-warm: rgba(120, 98, 64, 0.13);
  --radius-md: 14px;
  --radius-xl: 22px;
  --radius-pill: 9999px;
  --font-display: 'SF Pro Display', -apple-system, system-ui, Inter, 'PingFang SC', 'Noto Sans SC', sans-serif;
  --font-text: 'SF Pro Text', -apple-system, system-ui, Inter, 'PingFang SC', 'Noto Sans SC', sans-serif;
  --font-code: 'SF Mono', ui-monospace, 'JetBrains Mono', 'Cascadia Code', monospace;
}
```

Create `web/src/styles/glass.css`:

```css
.glass {
  position: relative;
  background: var(--glass-fill);
  backdrop-filter: blur(20px) saturate(125%);
  -webkit-backdrop-filter: blur(20px) saturate(125%);
  border-radius: var(--radius-xl);
  box-shadow:
    0 16px 44px rgba(120, 98, 64, 0.13),
    0 3px 10px rgba(120, 98, 64, 0.07),
    inset 0 1.6px 0 rgba(255, 255, 255, 0.95),
    inset 0 -1px 0 rgba(255, 255, 255, 0.35);
}

.glass::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  padding: 1.8px;
  background: linear-gradient(135deg, rgba(255,255,255,1) 0%, rgba(255,255,255,0.55) 18%, rgba(255,255,255,0.02) 48%, rgba(255,255,255,0.40) 78%, rgba(255,255,255,1) 100%);
  -webkit-mask: linear-gradient(#000 0 0) content-box, linear-gradient(#000 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  pointer-events: none;
}

.glass > * {
  position: relative;
  z-index: 1;
}
```

Create `web/src/styles/app.css`:

```css
html, body, #root {
  min-height: 100%;
}

body {
  margin: 0;
  color: var(--body);
  font-family: var(--font-text);
  font-size: 17px;
  line-height: 1.6;
  background:
    radial-gradient(circle at 50% 4%, rgba(246, 228, 190, 0.45), transparent 42%),
    linear-gradient(180deg, var(--canvas-paper-1), var(--canvas-paper-2));
  background-attachment: fixed;
}

a { color: var(--primary); }

button, input, textarea, select {
  font: inherit;
}

button {
  transform-origin: center;
}

button:active {
  transform: scale(0.95);
}

.app-main {
  max-width: 960px;
  margin: 0 auto;
  padding: 92px 20px 64px;
}

.html-document-frame {
  width: 100%;
  min-height: 640px;
  height: 78vh;
  border: 0;
  border-radius: var(--radius-md);
  background: white;
}

@media (max-width: 640px) {
  .html-document-frame { height: 72vh; min-height: 0; }
  .app-main { padding-inline: 16px; }
}
```

- [ ] **Step 5: Verify frontend build or identify install blocker**

Run:

```bash
cd web && npm run build
```

Expected: PASS. If package installation fails because the exact future versions are unavailable in the current npm registry, record the exact npm error in the task notes and do not substitute versions without updating `TECH_STACK.md`.

- [ ] **Step 6: Commit**

```bash
git add web
git commit -m "Give the SPA its contract-aligned shell

Constraint: TECH_STACK.md requires Vite React SPA with exact versions, and DESIGN.md requires one glass-ricepaper visual primitive.
Confidence: medium
Scope-risk: moderate
Directive: Do not introduce Next.js, SSR, Redux, dark mode, or alternate accent colors.
Tested: cd web && npm run build
Not-tested: API data pages remain skeletal until route components are implemented."
```

## Task 11: Public Frontend Pages and Interaction Components

**Files:**
- Create: `web/src/components/AppShell.tsx`
- Create: `web/src/components/GlassNav.tsx`
- Create: `web/src/components/ContentEntryCard.tsx`
- Create: `web/src/components/Breadcrumb.tsx`
- Create: `web/src/components/DirectoryDrawer.tsx`
- Create: `web/src/components/LikeButton.tsx`
- Create: `web/src/components/CommentThread.tsx`
- Create: `web/src/routes/RootDirectoryPage.tsx`
- Create: `web/src/routes/ContentPathPage.tsx`
- Create: `web/src/routes/RecentPage.tsx`
- Create: `web/src/routes/SearchPage.tsx`
- Create: `web/src/routes/LoginPage.tsx`
- Create: `web/src/routes/RegisterPage.tsx`

- [ ] **Step 1: Implement shell and nav**

Create `web/src/components/AppShell.tsx`:

```tsx
import { ReactNode } from 'react';
import { GlassNav } from './GlassNav';

export function AppShell({ children }: { children: ReactNode }) {
  return (
    <>
      <GlassNav />
      <main className="app-main">{children}</main>
    </>
  );
}
```

Create `web/src/components/GlassNav.tsx`:

```tsx
import { Link, NavLink, useNavigate } from 'react-router-dom';
import { Search } from 'lucide-react';

export function GlassNav() {
  const navigate = useNavigate();
  return (
    <nav className="glass glass-nav" aria-label="Primary navigation">
      <Link className="nav-brand" to="/">Zephyr</Link>
      <form
        className="nav-search"
        onSubmit={(event) => {
          event.preventDefault();
          const form = new FormData(event.currentTarget);
          const q = String(form.get('q') ?? '').trim();
          if (q) navigate(`/search?q=${encodeURIComponent(q)}`);
        }}
      >
        <Search size={16} />
        <input name="q" placeholder="Search files" aria-label="Search files" />
      </form>
      <div className="nav-links">
        <NavLink to="/recent">Recent</NavLink>
        <NavLink to="/login">Login</NavLink>
        <span className="locale-toggle" aria-label="UI language">ZH / EN</span>
      </div>
    </nav>
  );
}
```

- [ ] **Step 2: Implement content cards**

Create `web/src/components/ContentEntryCard.tsx`:

```tsx
import { Link } from 'react-router-dom';
import type { DirectoryEntry, FileEntry } from '../api/types';

function isFile(entry: DirectoryEntry | FileEntry): entry is FileEntry {
  return 'content_format' in entry;
}

export function ContentEntryCard({ entry }: { entry: DirectoryEntry | FileEntry }) {
  const label = isFile(entry) ? 'FILE' : 'DIRECTORY';
  const caption = isFile(entry)
    ? `${entry.node.path} · ${entry.reading_time_minutes ?? 1} min read`
    : `${entry.child_directory_count} directories · ${entry.child_file_count} files`;
  return (
    <Link className="glass content-entry-card" to={entry.node.path}>
      <span className="entry-label">{label}</span>
      <h2>{entry.node.name}</h2>
      <p>{caption}</p>
    </Link>
  );
}
```

- [ ] **Step 3: Implement root and recent pages**

Create `web/src/routes/RootDirectoryPage.tsx`:

```tsx
import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';
import type { DirectoryPage } from '../api/types';
import { ContentEntryCard } from '../components/ContentEntryCard';

export function RootDirectoryPage() {
  const { data, isLoading, error } = useQuery({ queryKey: ['tree-root'], queryFn: () => api<DirectoryPage>('/tree') });
  if (isLoading) return <section className="glass state-panel">Loading root directory…</section>;
  if (error) return <section className="glass state-panel">Failed to load root directory.</section>;
  if (!data || data.entries.length === 0) return <section className="glass state-panel">No files yet</section>;
  return (
    <section>
      <header className="hero"><p className="entry-label">CONTENT TREE</p><h1>Root Directory</h1></header>
      <div className="entry-grid">{data.entries.map((entry) => <ContentEntryCard key={entry.node.id} entry={entry} />)}</div>
    </section>
  );
}
```

Create `web/src/routes/RecentPage.tsx`:

```tsx
import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';
import type { FileEntry } from '../api/types';
import { ContentEntryCard } from '../components/ContentEntryCard';

export function RecentPage() {
  const { data, isLoading, error } = useQuery({ queryKey: ['recent'], queryFn: () => api<{ items: FileEntry[]; limit: number; offset: number }>('/recent') });
  if (isLoading) return <section className="glass state-panel">Loading recent files…</section>;
  if (error) return <section className="glass state-panel">Failed to load recent files.</section>;
  if (!data || data.items.length === 0) return <section className="glass state-panel">No recent files</section>;
  return <div className="entry-grid">{data.items.map((entry) => <ContentEntryCard key={entry.node.id} entry={entry} />)}</div>;
}
```

- [ ] **Step 4: Implement catch-all resolver page**

Create `web/src/routes/ContentPathPage.tsx`:

```tsx
import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useLocation, useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import type { TreeResolveResponse } from '../api/types';
import { ContentEntryCard } from '../components/ContentEntryCard';
import { MarkdownRenderer } from '../components/MarkdownRenderer';
import { HtmlDocumentFrame } from '../components/HtmlDocumentFrame';

export function ContentPathPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const path = location.pathname;
  const { data, isLoading, error } = useQuery({ queryKey: ['resolve', path], queryFn: () => api<TreeResolveResponse>(`/tree/resolve?path=${encodeURIComponent(path)}`) });
  useEffect(() => {
    if (data?.type === 'redirect') navigate(data.new_path, { replace: true });
  }, [data, navigate]);
  if (isLoading) return <section className="glass state-panel">Resolving path…</section>;
  if (error) return <section className="glass state-panel">File or Directory not found.</section>;
  if (!data || data.type === 'redirect') return null;
  if (data.type === 'directory') {
    return <div className="entry-grid">{data.directory.entries.map((entry) => <ContentEntryCard key={entry.node.id} entry={entry} />)}</div>;
  }
  const file = data.file;
  return (
    <article className="glass file-reading-card">
      <div className="keyword-row">{file.keywords_public.map((keyword) => <a className="keyword-chip" href={`/search?q=${encodeURIComponent(keyword)}`} key={keyword}>{keyword}</a>)}</div>
      <h1>{file.node.name}</h1>
      <p className="file-meta">{file.node.path}</p>
      {file.content.content_format === 'markdown'
        ? <MarkdownRenderer markdown={file.content.body_raw} />
        : <HtmlDocumentFrame html={file.content.body_raw} title={file.node.name} />}
    </article>
  );
}
```

- [ ] **Step 5: Verify iframe sandbox with frontend test**

Create `web/src/components/HtmlDocumentFrame.test.tsx`:

```tsx
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { HtmlDocumentFrame } from './HtmlDocumentFrame';

describe('HtmlDocumentFrame', () => {
  it('allows scripts without same-origin', () => {
    render(<HtmlDocumentFrame html="<html><body>demo</body></html>" title="Demo" />);
    const frame = screen.getByTitle('Demo');
    expect(frame).toHaveAttribute('sandbox', 'allow-scripts');
    expect(frame.getAttribute('sandbox')).not.toContain('allow-same-origin');
  });
});
```

Install test dependencies only if the repo decides to run frontend component tests. If not installed, keep this test file out of the commit and rely on `npm run build` plus manual DOM inspection until `TECH_STACK.md` is updated with Vitest packages.

- [ ] **Step 6: Verify build**

Run:

```bash
cd web && npm run build
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add web/src
git commit -m "Expose the content tree through the glass UI

Constraint: BLOG_FLOW.md fixes public routes, redirect handling, card labels, iframe sandboxing, and recent/search entry points.
Confidence: medium
Scope-risk: moderate
Directive: Markdown and HTML are render formats, not public content categories.
Tested: cd web && npm run build
Not-tested: Component test runner is not added because TECH_STACK.md does not pin Vitest or Testing Library."
```

## Task 12: Admin Tree Manager, File Editor, Assets UI, and Auth Pages

**Files:**
- Create: `web/src/auth/AuthContext.tsx`
- Create: `web/src/routes/LoginPage.tsx`
- Create: `web/src/routes/RegisterPage.tsx`
- Create: `web/src/routes/AdminPage.tsx`
- Create: `web/src/admin/AdminTreeManager.tsx`
- Create: `web/src/admin/NodeEditor.tsx`
- Create: `web/src/admin/FileEditor.tsx`
- Create: `web/src/admin/AssetManager.tsx`

- [ ] **Step 1: Implement AuthContext with return target support**

Create `web/src/auth/AuthContext.tsx`:

```tsx
import { createContext, ReactNode, useContext, useMemo, useState } from 'react';
import { api, getToken, setToken } from '../api/client';
import type { AuthResponse, User } from '../api/types';

interface AuthContextValue {
  token: string | null;
  user: User | null;
  login(email: string, password: string): Promise<void>;
  register(email: string, password: string, displayName: string): Promise<void>;
  logout(): void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setTokenState] = useState<string | null>(() => getToken());
  const [user, setUser] = useState<User | null>(null);
  const value = useMemo<AuthContextValue>(() => ({
    token,
    user,
    async login(email, password) {
      const res = await api<AuthResponse>('/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) });
      setToken(res.token);
      setTokenState(res.token);
      setUser(res.user);
    },
    async register(email, password, displayName) {
      const res = await api<AuthResponse>('/auth/register', { method: 'POST', body: JSON.stringify({ email, password, display_name: displayName }) });
      setToken(res.token);
      setTokenState(res.token);
      setUser(res.user);
    },
    logout() {
      setToken(null);
      setTokenState(null);
      setUser(null);
    },
  }), [token, user]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
```

- [ ] **Step 2: Wrap `App` in AuthProvider**

Modify `web/src/app/App.tsx` to nest `AuthProvider` inside `BrowserRouter` and outside `AppShell`:

```tsx
<AuthProvider>
  <AppShell>
    <Routes>{/* existing routes */}</Routes>
  </AppShell>
</AuthProvider>
```

- [ ] **Step 3: Implement login/register pages**

Create `web/src/routes/LoginPage.tsx`:

```tsx
import { FormEvent, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';

export function LoginPage() {
  const auth = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [error, setError] = useState('');
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    try {
      await auth.login(String(form.get('email')), String(form.get('password')));
      const params = new URLSearchParams(location.search);
      navigate(params.get('return_to') || '/recent', { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    }
  }
  return <form className="glass auth-card" onSubmit={submit}><h1>Login</h1><input name="email" type="email" required /><input name="password" type="password" required /><button className="button-primary">Login</button>{error && <p className="form-error">{error}</p>}</form>;
}
```

Create `web/src/routes/RegisterPage.tsx` with the same shape and fields `display_name`, `email`, `password`, then call `auth.register(...)` and navigate to return target or `/recent`.

- [ ] **Step 4: Implement admin manager vertical slice**

Create `web/src/routes/AdminPage.tsx`:

```tsx
import { AdminTreeManager } from '../admin/AdminTreeManager';

export function AdminPage() {
  return <AdminTreeManager />;
}
```

Create `web/src/admin/AdminTreeManager.tsx`:

```tsx
import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';
import type { DirectoryPage } from '../api/types';
import { NodeEditor } from './NodeEditor';

export function AdminTreeManager() {
  const { data, isLoading, error } = useQuery({ queryKey: ['admin-root'], queryFn: () => api<DirectoryPage>('/tree') });
  if (isLoading) return <section className="glass state-panel">Loading admin tree…</section>;
  if (error) return <section className="glass state-panel">Admin tree failed to load.</section>;
  return <section className="glass admin-panel"><h1>Admin Tree Manager</h1><NodeEditor parentId={null} /> <pre>{JSON.stringify(data?.entries ?? [], null, 2)}</pre></section>;
}
```

Create `web/src/admin/NodeEditor.tsx`:

```tsx
import { FormEvent, useState } from 'react';
import { api } from '../api/client';

export function NodeEditor({ parentId }: { parentId: string | null }) {
  const [message, setMessage] = useState('');
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await api('/admin/nodes', {
      method: 'POST',
      body: JSON.stringify({
        parent_id: parentId,
        kind: form.get('kind'),
        name: form.get('name'),
        slug: form.get('slug'),
        sort_order: Number(form.get('sort_order') || 0),
        content_format: form.get('content_format') || 'markdown',
      }),
    });
    setMessage('Node created');
    event.currentTarget.reset();
  }
  return <form className="node-editor" onSubmit={submit}><select name="kind"><option value="directory">Directory</option><option value="file">File</option></select><input name="name" placeholder="Name" required /><input name="slug" placeholder="slug" required /><input name="sort_order" type="number" defaultValue="0" /><select name="content_format"><option value="markdown">Markdown</option><option value="html_document">HTML Document</option></select><button className="button-primary">Create</button>{message && <p>{message}</p>}</form>;
}
```

- [ ] **Step 5: Verify frontend build**

Run:

```bash
cd web && npm run build
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add web/src/auth web/src/routes web/src/admin web/src/app/App.tsx
git commit -m "Let the author manage the content tree from the SPA

Constraint: PRD requires env-seeded admin to create directories, files, content, assets, and publish state through /admin.
Confidence: medium
Scope-risk: moderate
Directive: Dangerous move/delete/publish operations must show impact prompts before final wiring.
Tested: cd web && npm run build
Not-tested: Full admin edit/upload flows need API route integration and browser smoke tests."
```

## Task 13: Router Integration and OpenAPI Path Coverage

**Files:**
- Create/Modify: `api/internal/http/router.go`
- Modify: `api/cmd/server/main.go`
- Modify handlers in `api/internal/{auth,tree,content,comments,likes,assets,search}`

- [ ] **Step 1: Create route registration matching OpenAPI**

Create `api/internal/http/router.go`:

```go
package httpx

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type RouteSet interface {
	Mount(r chi.Router)
}

func NewRouter(mounts ...RouteSet) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) { JSON(w, http.StatusOK, map[string]string{"status": "ok"}) })
	for _, mount := range mounts {
		mount.Mount(r)
	}
	return r
}
```

Each feature handler must mount only paths from `docs/api/openapi.yaml`:

```txt
POST /api/auth/register
POST /api/auth/login
GET /api/auth/me
GET /api/tree
GET /api/tree/resolve
GET /api/tree/{node_id}/children
GET /api/recent
GET /api/search
GET /api/files/{file_id}/comments
POST /api/files/{file_id}/comments
DELETE /api/comments/{comment_id}
PUT /api/files/{file_id}/like
DELETE /api/files/{file_id}/like
PUT /api/comments/{comment_id}/like
DELETE /api/comments/{comment_id}/like
GET /api/assets/{asset_id}/{filename}
POST /api/admin/nodes
GET /api/admin/nodes/{node_id}
PATCH /api/admin/nodes/{node_id}
DELETE /api/admin/nodes/{node_id}
PUT /api/admin/files/{file_id}/content
POST /api/admin/files/{file_id}/publish
POST /api/admin/files/{file_id}/unpublish
POST /api/admin/files/{file_id}/assets
DELETE /api/admin/assets/{asset_id}
POST /api/admin/files/{file_id}/refresh-embedding
POST /api/admin/search-index/rebuild
```

- [ ] **Step 2: Add an OpenAPI route coverage script**

Create `api/internal/http/router_test.go`:

```go
package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthRoute(t *testing.T) {
	router := NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", res.Code)
	}
}
```

- [ ] **Step 3: Verify all backend packages compile**

Run:

```bash
cd api && go test ./...
```

Expected: PASS. If PostgreSQL integration tests are added, guard them behind an environment variable such as `DATABASE_URL` so unit tests remain runnable without Docker.

- [ ] **Step 4: Commit**

```bash
git add api
git commit -m "Wire API routes to the OpenAPI contract

Constraint: docs/api/openapi.yaml is the single API source of truth for public, reader, and admin paths.
Confidence: medium
Scope-risk: broad
Directive: New routes must update OpenAPI first, then handlers and frontend client usage.
Tested: cd api && go test ./...
Not-tested: End-to-end HTTP behavior awaits Docker Compose and seeded fixtures."
```

## Task 14: Docker Compose, Caddy, and Production Build

**Files:**
- Create: `docker-compose.yml`
- Create: `Caddyfile`
- Create: `api/Dockerfile`
- Create: `web/Dockerfile`
- Create: `.env.example`

- [ ] **Step 1: Create Docker Compose**

Create `docker-compose.yml`:

```yaml
services:
  db:
    image: pgvector/pgvector:0.8.2-pg17
    environment:
      POSTGRES_DB: blog
      POSTGRES_USER: blog
      POSTGRES_PASSWORD: blog
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U blog -d blog"]
      interval: 5s
      timeout: 5s
      retries: 20

  api:
    build:
      context: ./api
    environment:
      HTTP_ADDR: :8080
      DATABASE_URL: postgres://blog:blog@db:5432/blog?sslmode=disable
      JWT_SECRET: ${JWT_SECRET}
      ADMIN_EMAIL: ${ADMIN_EMAIL}
      ADMIN_PASSWORD: ${ADMIN_PASSWORD}
      ASSET_STORAGE: local
      ASSET_UPLOAD_DIR: /app/uploads
      ASSET_PUBLIC_BASE_URL: /api/assets
      EMBEDDING_PROVIDER: qwen
      DASHSCOPE_API_KEY: ${DASHSCOPE_API_KEY:-}
      EMBEDDING_BASE_URL: ${EMBEDDING_BASE_URL:-https://dashscope.aliyuncs.com/compatible-mode/v1}
      EMBEDDING_MODEL: text-embedding-v4
      EMBEDDING_DIMENSIONS: 1024
    volumes:
      - uploads:/app/uploads
    depends_on:
      db:
        condition: service_healthy

  web:
    build:
      context: .
      dockerfile: web/Dockerfile
    ports:
      - "8080:80"
    depends_on:
      - api

volumes:
  postgres_data:
  uploads:
```

- [ ] **Step 2: Create Caddyfile**

Create `Caddyfile`:

```caddyfile
:80 {
  root * /srv
  encode zstd gzip

  handle /api/* {
    reverse_proxy api:8080
  }

  handle {
    try_files {path} /index.html
    file_server
  }
}
```

- [ ] **Step 3: Create Dockerfiles**

Create `api/Dockerfile`:

```dockerfile
FROM golang:1.26.4-alpine3.23 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/server ./cmd/server

FROM alpine:3.23
WORKDIR /app
COPY --from=build /out/server /app/server
EXPOSE 8080
CMD ["/app/server"]
```

Create `web/Dockerfile`:

```dockerfile
FROM node:26.3.0-alpine3.23 AS build
WORKDIR /src/web
COPY web/package*.json ./
RUN npm ci
COPY web ./
RUN npm run build

FROM caddy:2.11.3-alpine
COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=build /src/web/dist /srv
EXPOSE 80
```

- [ ] **Step 4: Create `.env.example`**

Create `.env.example`:

```dotenv
JWT_SECRET=replace-with-at-least-32-random-characters
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=replace-with-at-least-8-characters
DASHSCOPE_API_KEY=
EMBEDDING_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1
```

- [ ] **Step 5: Verify Docker config renders**

Run:

```bash
docker compose config
```

Expected: PASS and no top-level `version:` key warning.

- [ ] **Step 6: Build full stack**

Run:

```bash
docker compose build
```

Expected: PASS. If exact image versions are unavailable, stop and update `TECH_STACK.md` with the error and selected replacement only after review.

- [ ] **Step 7: Commit**

```bash
git add docker-compose.yml Caddyfile api/Dockerfile web/Dockerfile .env.example
git commit -m "Make the blog deployable as one compose stack

Constraint: PRD requires Caddy, Go API, PostgreSQL pgvector, and uploads volume with Docker Compose v2 syntax.
Confidence: medium
Scope-risk: moderate
Directive: Keep assets in the uploads volume and keep Caddy as the edge service.
Tested: docker compose config && docker compose build
Not-tested: Live domain HTTPS and external DashScope credentials are deployment-environment concerns."
```

## Task 15: Required High-Risk Verification Suite

**Files:**
- Modify/Create: `api/internal/auth/*_test.go`
- Modify/Create: `api/internal/likes/*_test.go`
- Modify/Create: `api/internal/render/*_test.go`
- Modify/Create: `api/internal/assets/*_test.go`
- Modify/Create: `api/internal/tree/*_test.go`
- Modify/Create: `web/src/components/HtmlDocumentFrame.test.tsx` if frontend test tooling is pinned later.
- Modify: `README.md`

- [ ] **Step 1: Confirm the five required high-risk tests exist**

Run:

```bash
grep -R "Test.*JWT\|Test.*Password\|Test.*Like\|Test.*Markdown\|Test.*HTML\|Test.*SVG\|Test.*Reserved\|Test.*Redirect\|Test.*Embedding" -n api/internal web/src || true
```

Expected: output includes tests for JWT/password, like idempotency, Markdown XSS, HTML visible text/iframe sandbox, SVG rejection, reserved root slug, redirects, embedding failure, and directory deletion rule.

- [ ] **Step 2: Add missing backend tests with exact expectations**

If missing, add tests named:

```txt
api/internal/auth/auth_test.go::TestJWTRejectsTamperedToken
api/internal/likes/service_test.go::TestLikeAndUnlikeAreIdempotentAtServiceBoundary
api/internal/render/render_test.go::TestMarkdownSanitizesScriptAndEvents
api/internal/assets/validation_test.go::TestRejectsMaliciousSVG
api/internal/tree/path_test.go::TestReservedRootSlug
```

- [ ] **Step 3: Run backend suite**

Run:

```bash
cd api && go test ./...
```

Expected: PASS.

- [ ] **Step 4: Run frontend suite**

Run:

```bash
cd web && npm run lint && npm run build
```

Expected: PASS.

- [ ] **Step 5: Run compose smoke**

Run:

```bash
docker compose up -d --build
curl -fsS http://localhost:8080/api/health
curl -fsS http://localhost:8080/
docker compose down
```

Expected: `/api/health` returns `{"status":"ok"}` and `/` returns the SPA HTML.

- [ ] **Step 6: Update README verification section**

Append to `README.md`:

```markdown
## Verification

Run before claiming completion:

```bash
cd api && go test ./...
cd ../web && npm run lint && npm run build
cd .. && docker compose config && docker compose up -d --build
curl -fsS http://localhost:8080/api/health
curl -fsS http://localhost:8080/
docker compose down
```

Required high-risk checks:

- JWT validation and role guard.
- Like idempotency.
- Markdown XSS sanitization.
- HTML Document iframe sandbox without `allow-same-origin`.
- SVG asset rejection.
- Root reserved slug rejection.
- Path redirect creation on published path change.
- Qwen embedding failure does not fail content save.
- Directory with published descendant cannot delete.
- HTML visible text extraction excludes script/style.
```

- [ ] **Step 7: Commit**

```bash
git add api web README.md
git commit -m "Lock completion behind high-risk verification

Constraint: PRD lists security, rendering, redirect, search, and deletion edge cases as completion signals.
Confidence: high
Scope-risk: moderate
Directive: Do not mark the product complete unless every command in README Verification passes or has a documented environment blocker.
Tested: cd api && go test ./...; cd web && npm run lint && npm run build; docker compose config; compose smoke with curl
Not-tested: Public HTTPS domain and real DashScope credentials if unavailable locally."
```

## Task 16: Final Acceptance Walkthrough

**Files:**
- Modify: `README.md`
- Create: `docs/superpowers/plans/2026-06-03-xlab-personal-blog-acceptance.md` if execution notes are needed.

- [ ] **Step 1: Start the stack**

Run:

```bash
docker compose up -d --build
```

Expected: `db`, `api`, and `web` containers are running; `docker compose ps` reports healthy/running states.

- [ ] **Step 2: Verify public routes**

Run:

```bash
curl -fsS http://localhost:8080/api/health
curl -fsS http://localhost:8080/
curl -fsS http://localhost:8080/recent
curl -fsS 'http://localhost:8080/search?q=test'
```

Expected: health returns JSON and SPA routes return HTML.

- [ ] **Step 3: Verify admin seed and login**

Use `.env` values for `ADMIN_EMAIL` and `ADMIN_PASSWORD`. Run:

```bash
curl -fsS -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.com","password":"replace-with-at-least-8-characters"}'
```

Expected: response contains `token` and `user.role` equal to `admin`.

- [ ] **Step 4: Create a Directory and File via API**

Use the admin token:

```bash
TOKEN="paste-token-here"
curl -fsS -X POST http://localhost:8080/api/admin/nodes \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"kind":"directory","name":"Research Notes","slug":"research-notes","sort_order":0}'
```

Expected: response has `node.kind` = `directory` and `node.path` = `/research-notes`.

- [ ] **Step 5: Verify required UX manually**

In browser at `http://localhost:8080` confirm:

```txt
Root page shows content-entry-card cards.
/recent shows File cards only.
/search?q=... shows result cards or a calm empty state.
/login redirects successful auth to /recent when no return target exists.
/admin loads the Tree Manager for admin.
Markdown File renders in file-reading-card.
HTML Document File renders inside iframe with sandbox="allow-scripts" and no allow-same-origin.
Anonymous like/comment actions redirect to login with return_to.
```

- [ ] **Step 6: Stop the stack**

Run:

```bash
docker compose down
```

Expected: containers stop and named volumes remain for data/uploads.

- [ ] **Step 7: Commit acceptance notes if README changed**

```bash
git add README.md docs/superpowers/plans
git commit -m "Record the product acceptance path

Constraint: PRD completion requires Docker, admin content creation, public routes, sandbox rendering, assets, search, comments, and likes to be verified together.
Confidence: high
Scope-risk: narrow
Directive: Keep acceptance commands current whenever API paths or deployment ports change.
Tested: docker compose up -d --build; curl health and SPA; admin login; manual browser walkthrough
Not-tested: Production TLS and real public domain if not available in the local environment."
```

---

## Self-Review

### 1. Spec coverage

- Unix-like Content Tree: Tasks 4, 10, 11, 12, 13.
- Directory/File with nested routing and redirects: Tasks 4, 11, 13, 15.
- Markdown and HTML Document render formats: Tasks 5, 11, 15.
- HTML iframe sandbox without `allow-same-origin`: Tasks 5, 11, 15.
- Auth, reader registration, admin seed: Tasks 3, 12, 13, 16.
- Comments and likes: Task 7 plus route wiring in Task 13.
- Per-file assets with local storage and SVG safety: Task 8 plus route wiring in Task 13.
- Hybrid search: Task 9 plus route wiring and frontend search in Tasks 11, 13.
- Docker Compose and Caddy: Task 14.
- Glass-ricepaper UI: Tasks 10 and 11.
- Required high-risk tests: Task 15.

### 2. Placeholder scan

The plan avoids the forbidden placeholder terms and gives concrete paths, code blocks, commands, and expected outputs. Where an integration depends on a prior scaffold or package availability, the step gives a concrete stop condition and an exact documented action.

### 3. Type consistency

- Backend terms use `Directory`, `File`, `Node`, `FileContent`, `FileAsset`, `Comment`, `LikeState` aligned with `docs/api/openapi.yaml`.
- Frontend `types.ts` mirrors OpenAPI schema names and fields used in route components.
- Commit messages use the Lore protocol required by `AGENTS.md` instead of conventional-only commit lines.
