begin;

alter table file_contents add column if not exists revision integer not null default 1;
alter table file_contents add column if not exists last_saved_at timestamptz not null default now();
alter table file_contents drop constraint if exists file_contents_status_check;
alter table file_contents add constraint file_contents_status_check check (status in ('draft','published','unpublished_changes'));

create table if not exists file_content_previous_versions (
  node_id uuid primary key references nodes(id) on delete cascade,
  revision integer not null,
  content_format text not null check (content_format in ('markdown','html_document')),
  keywords text[] not null default '{}',
  body_raw text not null,
  body_html text,
  search_text text not null default '',
  status text not null default 'draft' check (status in ('draft','published','unpublished_changes')),
  last_saved_at timestamptz not null,
  created_at timestamptz not null default now()
);

create table if not exists published_file_contents (
  node_id uuid primary key references nodes(id) on delete cascade,
  source_revision integer not null,
  content_format text not null check (content_format in ('markdown','html_document')),
  keywords text[] not null default '{}',
  body_raw text not null,
  body_html text,
  search_text text not null default '',
  published_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  visible boolean not null default true,
  embedding vector(1024),
  embedding_model text,
  embedding_status text not null default 'pending' check (embedding_status in ('pending','ready','failed')),
  embedding_error text,
  embedding_updated_at timestamptz,
  search_vector tsvector not null default ''::tsvector
);

create index if not exists published_file_contents_visible_idx on published_file_contents(visible, updated_at desc);
create index if not exists published_file_contents_keywords_gin_idx on published_file_contents using gin(keywords);
create index if not exists published_file_contents_search_vector_idx on published_file_contents using gin(search_vector);

create or replace function update_published_file_contents_search_vector()
returns trigger
language plpgsql
as $$
begin
  new.search_vector :=
    setweight(to_tsvector('simple', coalesce(array_to_string(new.keywords, ' '), '')), 'A') ||
    setweight(to_tsvector('simple', coalesce(new.search_text, '')), 'B');
  return new;
end;
$$;

drop trigger if exists published_file_contents_search_vector_trigger on published_file_contents;
create trigger published_file_contents_search_vector_trigger
before insert or update of keywords, search_text on published_file_contents
for each row execute function update_published_file_contents_search_vector();

alter table file_assets add column if not exists state text not null default 'draft';
alter table file_assets add column if not exists published_asset_id uuid;
alter table file_assets drop constraint if exists file_assets_state_check;
alter table file_assets add constraint file_assets_state_check check (state in ('draft','published','draft_and_published'));
create index if not exists file_assets_file_state_idx on file_assets(file_node_id, state);

create table if not exists published_file_assets (
  published_asset_id uuid primary key default gen_random_uuid(),
  asset_id uuid not null references file_assets(id) on delete cascade,
  file_node_id uuid not null references nodes(id) on delete cascade,
  filename text not null,
  mime_type text not null,
  size_bytes bigint not null,
  storage_provider text not null default 'local',
  storage_key text not null,
  published_at timestamptz not null default now(),
  unique(file_node_id, filename)
);

insert into published_file_contents (
  node_id, source_revision, content_format, keywords, body_raw, body_html, search_text,
  published_at, updated_at, visible, embedding, embedding_model, embedding_status, embedding_error, embedding_updated_at
)
select node_id, revision, content_format, keywords, body_raw, body_html, search_text,
  coalesce(published_at, now()), now(), true, embedding, embedding_model, embedding_status, embedding_error, embedding_updated_at
from file_contents
where status = 'published'
on conflict (node_id) do nothing;

update file_assets a
set state = 'published'
where exists (select 1 from published_file_contents p where p.node_id = a.file_node_id and p.visible);

commit;
