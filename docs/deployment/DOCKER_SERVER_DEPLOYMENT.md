# Docker Server Deployment — Aeolian Blog

Target first deployment:

- Server OS: Ubuntu Server 24.04 LTS
- Public IP: `154.37.222.233`
- Domain: none yet
- Runtime: Docker + Docker Compose
- Public URL for first release: `http://154.37.222.233:8080`

This deployment runs four containers:

```text
caddy  -> public HTTP entry, static SPA, /api reverse proxy
web    -> builds/copies React/Vite dist into shared volume
api    -> Go API, runs SQL migrations at startup
db     -> PostgreSQL 17 + pgvector
```

MCP is intentionally not exposed as a public service. The Blog MCP Server remains server-local stdio only; see `mcp/README.md` if a trusted local AI agent needs it later.

## 1. Server prerequisites

On the server:

```bash
uname -a
cat /etc/os-release
docker --version
docker compose version
```

If Docker is missing, install Docker Engine using Docker's official Ubuntu instructions, then verify the commands above.

Optional but recommended for a 2-core / 4GB server: add swap if the provider image has none.

```bash
free -h
```

If swap is `0B`, create a 2GB swap file:

```bash
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

## 2. Get the code

Use GitHub if available:

```bash
git clone <YOUR_REPO_URL> aeolian-blog
cd aeolian-blog
```

Or upload/copy the repository directory to the server and enter it.

## 3. Create production environment file

```bash
cp .env.production.example .env
nano .env
```

Required changes:

```text
POSTGRES_PASSWORD=strong random value
DATABASE_URL=postgres://xlab_blog:<same password>@db:5432/xlab_blog?sslmode=disable
JWT_SECRET=strong random value
ADMIN_EMAIL=your Author email
ADMIN_PASSWORD=strong Author password
PUBLIC_SITE_URL=http://154.37.222.233:8080
PUBLIC_HTTP_PORT=8080
```

Generate random secrets on the server:

```bash
openssl rand -base64 32
openssl rand -base64 48
```

Never commit `.env`.

## 4. Start the stack

```bash
docker compose up -d --build
```

Check status:

```bash
docker compose ps
docker compose logs --tail=100 api
docker compose logs --tail=100 caddy
```

Expected status:

```text
blog-db-1      healthy
blog-api-1     healthy
blog-web-1     healthy
blog-caddy-1   running, 0.0.0.0:8080->80/tcp
```

## 5. Smoke test

From the server:

```bash
curl -fsS http://127.0.0.1:8080/api/health
curl -fsS http://127.0.0.1:8080/ >/dev/null
```

From your browser:

```text
http://154.37.222.233:8080
```

Acceptance path:

1. Open the public URL.
2. Login as Author using `ADMIN_EMAIL` / `ADMIN_PASSWORD` from `.env`.
3. Enter Author Workspace.
4. Create a Directory and File.
5. Edit content and wait for autosave.
6. Publish the File and open it publicly.
7. Restart containers and confirm data remains.

Restart persistence check:

```bash
docker compose restart
curl -fsS http://127.0.0.1:8080/api/health
```

## 6. Firewall / provider security group

For the no-domain first deployment, open inbound TCP:

```text
8080
```

If you later switch to a domain and HTTPS, open:

```text
80
443
```

and change `PUBLIC_HTTP_PORT` / Caddy config accordingly.

## 7. Update deployment

```bash
git pull
# edit .env only if new variables were added
docker compose up -d --build
```

Check:

```bash
docker compose ps
curl -fsS http://127.0.0.1:8080/api/health
```

## 8. Backup

Create a backup directory outside Git:

```bash
mkdir -p ~/aeolian-backups
```

Database backup:

```bash
docker compose exec -T db pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB" > ~/aeolian-backups/xlab_blog-$(date +%Y%m%d-%H%M%S).sql
```

Uploads backup:

```bash
docker run --rm -v blog_uploads:/data -v "$HOME/aeolian-backups:/backup" alpine tar -czf /backup/uploads-$(date +%Y%m%d-%H%M%S).tgz -C /data .
```

## 9. Restore outline

Stop the stack first:

```bash
docker compose down
```

Restore should be done carefully on a fresh volume or after a verified backup. Do not run destructive restore commands on the only copy of production data without a current backup.

## 10. Stop or remove

Stop containers but keep data volumes:

```bash
docker compose down
```

Danger: remove containers and all named volumes, including database and uploads:

```bash
docker compose down -v
```

Do not use `-v` unless you intentionally want to delete the deployment data.

## 11. Troubleshooting

### `POSTGRES_PASSWORD is required` or `JWT_SECRET is required`

Create and edit `.env` from `.env.production.example`.

### API unhealthy

```bash
docker compose logs --tail=200 api
docker compose logs --tail=200 db
```

Common causes:

- `DATABASE_URL` password does not match `POSTGRES_PASSWORD`.
- Database volume was initialized with an old password. If this is a throwaway first run, use `docker compose down -v` and recreate. Do not do this after real content exists.
- Migrations failed.

### Public site not reachable

Check provider firewall/security group and server firewall:

```bash
docker compose ps
sudo ufw status
```

For first deployment, TCP `8080` must be open.

### Search semantic embedding unavailable

If `DASHSCOPE_API_KEY` is empty or embedding calls fail, lexical/full-text search fallback still works by design.
