# TECH_STACK.md — 精确技术栈与版本锁

> 版本：2026-06-03
> 原则：实现必须使用本文件列出的确切版本。不要自行替换框架、包或 major version。若安装失败，先报告原因并更新本文件，而不是静默换依赖。

## 1. Runtime / Tooling

| 用途 | 版本 |
|---|---|
| Node.js 本地开发 | `22.22.3` |
| npm | `10.9.8` |
| create-vite | `9.0.7` |
| Go | `1.26.4` |
| Docker Compose | Compose v2 CLI：`docker compose` |
| OpenAPI | `3.2.0` |

说明：当前环境未安装 Go CLI，但实现目标版本锁为 Go `1.26.4`，Docker build 使用 `golang:1.26.4-alpine3.23`。

## 2. Frontend

本项目是 Vite React SPA，不使用 Next.js / SSR。

### 2.1 npm dependencies

```json
{
  "@vitejs/plugin-react": "6.0.2",
  "vite": "8.0.16",
  "react": "19.2.7",
  "react-dom": "19.2.7",
  "typescript": "6.0.3",
  "react-router-dom": "7.16.0",
  "@tanstack/react-query": "5.101.0",
  "dompurify": "3.4.7",
  "marked": "18.0.4",
  "lucide-react": "1.17.0",
  "zod": "4.4.3"
}
```

### 2.2 npm devDependencies

```json
{
  "@types/react": "19.2.16",
  "@types/react-dom": "19.2.3",
  "@types/dompurify": "3.2.0",
  "eslint": "10.4.1",
  "typescript-eslint": "8.60.1",
  "eslint-plugin-react-hooks": "7.1.1",
  "eslint-plugin-react-refresh": "0.5.2",
  "prettier": "3.8.3"
}
```

### 2.3 Frontend constraints

- React function components + hooks only。
- React Router v7 route objects / `<Routes>` API；禁止 v5 `<Switch>`。
- No Redux。
- Request state may use TanStack Query `5.101.0`。
- Markdown rendering：`marked@18.0.4` → `dompurify@3.4.7`。
- HTML Document rendering：iframe only, `sandbox="allow-scripts"` without `allow-same-origin`。
- UI must follow `docs/specs/DESIGN.md` tokens; no dark mode。


### 2.4 package.json template

`web/package.json` dependencies must use exact versions, not ranges. Example:

```json
{
  "dependencies": {
    "@tanstack/react-query": "5.101.0",
    "dompurify": "3.4.7",
    "lucide-react": "1.17.0",
    "marked": "18.0.4",
    "react": "19.2.7",
    "react-dom": "19.2.7",
    "react-router-dom": "7.16.0",
    "zod": "4.4.3"
  },
  "devDependencies": {
    "@types/dompurify": "3.2.0",
    "@types/react": "19.2.16",
    "@types/react-dom": "19.2.3",
    "@vitejs/plugin-react": "6.0.2",
    "eslint": "10.4.1",
    "eslint-plugin-react-hooks": "7.1.1",
    "eslint-plugin-react-refresh": "0.5.2",
    "prettier": "3.8.3",
    "typescript": "6.0.3",
    "typescript-eslint": "8.60.1",
    "vite": "8.0.16"
  }
}
```

## 3. Backend Go modules

`go.mod` must pin:

```txt
go 1.26.4

require (
  github.com/go-chi/chi/v5 v5.3.0
  github.com/golang-jwt/jwt/v5 v5.3.1
  github.com/jackc/pgx/v5 v5.10.0
  github.com/google/uuid v1.6.0
  github.com/pgvector/pgvector-go v0.4.0
  github.com/microcosm-cc/bluemonday v1.0.27
  github.com/yuin/goldmark v1.8.2
  github.com/joho/godotenv v1.5.1
  golang.org/x/crypto v0.52.0
  golang.org/x/net v0.55.0
)
```

### 3.1 Backend package constraints

- Router: `github.com/go-chi/chi/v5@v5.3.0`。
- JWT: `github.com/golang-jwt/jwt/v5@v5.3.1`；禁止 `dgrijalva/jwt-go`。
- Password hash: `golang.org/x/crypto/bcrypt` from `v0.52.0`。
- DB: `github.com/jackc/pgx/v5@v5.10.0`。
- UUID: `github.com/google/uuid@v1.6.0`。
- pgvector Go helper: `github.com/pgvector/pgvector-go@v0.4.0`。
- Markdown: `github.com/yuin/goldmark@v1.8.2`。
- Sanitizer: `github.com/microcosm-cc/bluemonday@v1.0.27`。
- HTML visible text extraction may use `golang.org/x/net/html@v0.55.0`。

## 4. Database / Images

| 用途 | 镜像 / 版本 |
|---|---|
| Postgres + pgvector | `pgvector/pgvector:0.8.2-pg17` |
| Caddy | `caddy:2.11.3-alpine` |
| Go builder | `golang:1.26.4-alpine3.23` |
| Node builder | `node:26.3.0-alpine3.23` |

Database extensions:

```sql
create extension if not exists vector;
create extension if not exists pgcrypto;
```

Notes:

- Use Docker Compose v2 syntax; no top-level `version:` key。
- Postgres data volume: `postgres_data`。
- Asset volume: `uploads` mounted at `/app/uploads`。

## 5. External APIs

### 5.1 Qwen / DashScope embeddings

Provider: Qwen/DashScope OpenAI-compatible embeddings.

Config:

```txt
EMBEDDING_PROVIDER=qwen
DASHSCOPE_API_KEY=...
EMBEDDING_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1
EMBEDDING_MODEL=text-embedding-v4
EMBEDDING_DIMENSIONS=1024
```

International endpoint if needed:

```txt
https://dashscope-intl.aliyuncs.com/compatible-mode/v1
```

Request body:

```json
{
  "model": "text-embedding-v4",
  "input": "name\npath\nkeywords joined\nsearch_text",
  "dimensions": 1024,
  "encoding_format": "float"
}
```

Constraints:

- Use `text-embedding-v4` exactly。
- Store embeddings as `vector(1024)`。
- Always send `dimensions: 1024` and `encoding_format: "float"` for OpenAI-compatible embedding calls。
- China API keys must use `https://dashscope.aliyuncs.com/compatible-mode/v1`; International API keys must use `https://dashscope-intl.aliyuncs.com/compatible-mode/v1`。
- Qwen failures must not block File save。
- No DeepSeek embedding in first release。
- No LLM query expansion/rerank in first release。

## 6. Security libraries / policies

- JWT secret from `JWT_SECRET` env; never hardcode。
- bcrypt password hash; never log password。
- DOMPurify frontend and bluemonday backend sanitize Markdown-rendered HTML path as appropriate。
- HTML Document is isolated by iframe sandbox and not sanitized into main DOM。
- SVG asset detection must reject script/event/javascript/external/foreignObject.

## 7. Commands

Frontend:

```bash
npm create vite@9.0.7 web -- --template react-ts
cd web
npm install --save-exact react@19.2.7 react-dom@19.2.7 react-router-dom@7.16.0 @tanstack/react-query@5.101.0 dompurify@3.4.7 marked@18.0.4 lucide-react@1.17.0 zod@4.4.3
npm install --save-dev --save-exact @vitejs/plugin-react@6.0.2 vite@8.0.16 typescript@6.0.3 @types/react@19.2.16 @types/react-dom@19.2.3 @types/dompurify@3.2.0 eslint@10.4.1 typescript-eslint@8.60.1 eslint-plugin-react-hooks@7.1.1 eslint-plugin-react-refresh@0.5.2 prettier@3.8.3
npm run dev
npm run build
npm run lint
```

Backend:

```bash
go mod init xlab-blog/api
go test ./...
go run ./cmd/server
```

Deploy:

```bash
docker compose up -d --build
```

## 8. Prohibited substitutions

Do not use:

- Next.js / Remix / SSR framework。
- Redux。
- CRA / create-react-app。
- gin / echo / gorilla/mux。
- GORM or ORM。
- `dgrijalva/jwt-go`。
- Docker Compose v1 `docker-compose`。
- nginx as edge service。
- Object storage implementation in first release unless this doc is updated。
