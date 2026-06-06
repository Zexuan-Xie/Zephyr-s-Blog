# Technical Stack

Last synchronized with manifests: 2026-06-06.

Use exact versions. Update this file and the relevant manifest together; do not silently substitute packages or major versions.

## Runtime and local tooling

| Component | Version |
|---|---:|
| Node.js | `22.22.3` |
| npm | `10.9.8` |
| Go | `1.26.4` |
| PostgreSQL local | `17.10` |
| pgvector local | `0.8.1` |
| OpenAPI | `3.2.0` |

The Conda environment is `blogenv` and is declared in `environment.yml`. Install exact npm separately after creating the environment:

```bash
conda env create -f environment.yml
conda run -n blogenv npm install -g npm@10.9.8
```

## Frontend

React/Vite SPA; no SSR framework.

Runtime dependencies:

| Package | Version |
|---|---:|
| `react` / `react-dom` | `19.2.7` |
| `react-router-dom` | `7.16.0` |
| `@tanstack/react-query` | `5.101.0` |
| `dompurify` | `3.4.7` |
| `marked` | `18.0.4` |
| `lucide-react` | `1.17.0` |
| `zod` | `4.4.3` |

Development dependencies:

| Package | Version |
|---|---:|
| `typescript` | `6.0.3` |
| `vite` | `8.0.16` |
| `@vitejs/plugin-react` | `6.0.2` |
| `@types/react` | `19.2.16` |
| `@types/react-dom` | `19.2.3` |
| `eslint` | `10.4.1` |
| `typescript-eslint` | `8.60.1` |
| `eslint-plugin-react-hooks` | `7.1.1` |
| `eslint-plugin-react-refresh` | `0.5.2` |

Notes:

- DOMPurify ships its own types; do not add `@types/dompurify`.
- No Prettier workflow is configured; ESLint and TypeScript are the active frontend gates.
- Markdown: Marked → DOMPurify.
- Full HTML Documents: iframe only with `sandbox="allow-scripts"`, never `allow-same-origin`.
- No Redux or dark mode.

## Backend

Direct dependencies are the source of truth in `api/go.mod`:

| Package | Version / purpose |
|---|---|
| `github.com/go-chi/chi/v5` | `v5.3.0`, router |
| `github.com/golang-jwt/jwt/v5` | `v5.3.1`, JWT |
| `github.com/google/uuid` | `v1.6.0`, identifiers |
| `github.com/jackc/pgx/v5` | `v5.10.0`, PostgreSQL |
| `github.com/joho/godotenv` | `v1.5.1`, local env loading |
| `github.com/microcosm-cc/bluemonday` | `v1.0.27`, HTML sanitization |
| `github.com/yuin/goldmark` | `v1.8.2`, Markdown |
| `golang.org/x/crypto` | `v0.52.0`, bcrypt |
| `golang.org/x/net` | `v0.55.0`, HTML parsing |

Vector queries use PostgreSQL/pgvector SQL directly; no Go pgvector helper dependency is required.

## Database and containers

| Purpose | Image/version |
|---|---|
| PostgreSQL + pgvector | `pgvector/pgvector:0.8.2-pg17` |
| Caddy | `caddy:2.11.3-alpine` |
| API builder | `golang:1.26.4-alpine` |
| Web builder | `node:22.22.3-alpine` |
| Runtime base | `alpine:3.22` |

Required extensions:

```sql
create extension if not exists vector;
create extension if not exists pgcrypto;
```

Use Docker Compose v2 syntax with no top-level `version` key.

## Embeddings

Provider: Qwen/DashScope OpenAI-compatible embeddings.

```text
EMBEDDING_PROVIDER=qwen
DASHSCOPE_API_KEY=...
EMBEDDING_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1
EMBEDDING_MODEL=text-embedding-v4
EMBEDDING_DIMENSIONS=1024
```

Always request `dimensions: 1024` and `encoding_format: "float"`. Semantic failure must not block saving or full-text search.

## Security invariants

- JWT secret comes from environment configuration.
- Passwords use bcrypt and are never logged.
- SQL uses pgx parameters.
- Markdown HTML is sanitized.
- Full HTML Documents remain isolated from the main DOM.
- SVG validation rejects scripts, event handlers, JavaScript URLs, external references, and `foreignObject`.

## Commands

```bash
# Frontend
cd web
npm ci
npm run lint
npm run build
node --test tests/render-safety.test.mjs

# Backend
cd api
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go test -count=1 ./...
CGO_ENABLED=0 GOCACHE=/tmp/xlab-blog-go-cache go vet ./...
test -z "$(gofmt -l .)"

# Deployment once Docker is available
docker compose config
docker compose up -d --build
```

## Prohibited substitutions

Do not introduce Next.js/Remix/SSR, Redux, CRA, Gin/Echo/Gorilla mux, an ORM, `dgrijalva/jwt-go`, Docker Compose v1, nginx, or object storage without an explicit spec decision.
