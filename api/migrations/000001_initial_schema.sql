begin;

create extension if not exists pgcrypto;
create extension if not exists vector;

create table if not exists users (
  id uuid primary key default gen_random_uuid(),
  email text not null unique,
  password_hash text not null,
  role text not null check (role in ('admin','reader')),
  display_name text,
  provider text not null default 'local',
  provider_id text,
  created_at timestamptz not null default now()
);

create unique index if not exists users_provider_provider_id_unique
  on users(provider, provider_id)
  where provider_id is not null;

create table if not exists nodes (
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

create index if not exists nodes_parent_sort_idx on nodes(parent_id, kind, sort_order, name);
create unique index if not exists nodes_root_slug_unique_idx on nodes(slug) where parent_id is null;

create table if not exists file_contents (
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

create index if not exists file_contents_status_idx on file_contents(status);
create index if not exists file_contents_keywords_gin_idx on file_contents using gin(keywords);
create index if not exists file_contents_search_vector_idx on file_contents using gin(search_vector);

create table if not exists path_redirects (
  id uuid primary key default gen_random_uuid(),
  old_path text not null unique,
  new_path text not null,
  node_id uuid not null references nodes(id) on delete cascade,
  created_at timestamptz not null default now()
);

create table if not exists comments (
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

create index if not exists comments_file_parent_created_idx on comments(file_node_id, parent_id, created_at);

create table if not exists likes (
  user_id uuid not null references users(id) on delete cascade,
  target_type text not null check (target_type in ('file','comment')),
  target_id uuid not null,
  created_at timestamptz not null default now(),
  primary key(user_id, target_type, target_id)
);

create index if not exists likes_target_idx on likes(target_type, target_id);

create table if not exists file_assets (
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

create index if not exists file_assets_file_idx on file_assets(file_node_id);

commit;
