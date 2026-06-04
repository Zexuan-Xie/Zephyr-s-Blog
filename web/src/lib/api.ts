import { z } from 'zod';
import type { BreadcrumbItem, ContentEntry, DirectoryPayload, FilePayload, ResolvePayload, SearchResult } from './types';

const apiBase = '/api';

async function requestJson<T>(path: string, schema: z.ZodType<T>): Promise<T> {
  const response = await fetch(`${apiBase}${path}`, {
    headers: { Accept: 'application/json' },
  });

  if (!response.ok) {
    throw new Error(`Request failed: ${response.status} ${response.statusText}`);
  }

  const json: unknown = await response.json();
  return schema.parse(json);
}

const nullableStringSchema = z.string().nullable().optional();
const nullableNumberSchema = z.number().nullable().optional();

const breadcrumbSchema = z.object({
  name: z.string(),
  path: z.string(),
});

const nodeSchema = z.object({
  id: z.string(),
  kind: z.enum(['directory', 'file']),
  name: z.string(),
  slug: z.string().optional(),
  path: z.string(),
  sort_order: z.number().optional(),
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
});

const legacyContentEntrySchema = z.object({
  id: z.string(),
  kind: z.enum(['directory', 'file']),
  name: z.string(),
  path: z.string(),
  sort_order: z.number().optional(),
  updated_at: z.string().optional(),
  child_directory_count: z.number().optional(),
  child_file_count: z.number().optional(),
  comment_count: z.number().optional(),
  content_format: z.enum(['markdown', 'html_document']).optional(),
  read_time_minutes: z.number().optional(),
  keywords: z.array(z.string()).optional(),
});

const directoryEntrySchema = z.object({
  node: nodeSchema.extend({ kind: z.literal('directory') }),
  child_directory_count: z.number(),
  child_file_count: z.number(),
});

const fileEntrySchema = z.object({
  node: nodeSchema.extend({ kind: z.literal('file') }),
  content_format: z.enum(['markdown', 'html_document']),
  status: z.string().optional(),
  keywords: z.array(z.string()).optional(),
  published_at: nullableStringSchema,
  like_count: z.number().optional(),
  comment_count: z.number().optional(),
  reading_time_minutes: nullableNumberSchema,
});

const entrySchema = z.union([legacyContentEntrySchema, directoryEntrySchema, fileEntrySchema]).transform(toContentEntry);

const legacyDirectorySchema = z.object({
  type: z.literal('directory').optional(),
  id: z.string(),
  name: z.string(),
  path: z.string(),
  breadcrumb: z.array(breadcrumbSchema).optional(),
  children: z.array(legacyContentEntrySchema),
});

const openApiDirectorySchema = z.object({
  node: nodeSchema.extend({ kind: z.literal('directory') }).nullable().optional(),
  path: z.string().optional(),
  entries: z.array(entrySchema),
});

const directorySchema: z.ZodType<DirectoryPayload> = z.union([legacyDirectorySchema, openApiDirectorySchema]).transform((directory) => {
  if ('children' in directory) {
    return {
      ...directory,
      type: 'directory' as const,
      breadcrumb: directory.breadcrumb ?? buildBreadcrumb(directory.path),
    };
  }

  const path = normalizeAbsolutePath(directory.path ?? directory.node?.path ?? '/');

  return {
    type: 'directory' as const,
    id: directory.node?.id ?? 'root',
    name: directory.node?.name ?? 'Root',
    path,
    breadcrumb: buildBreadcrumb(path),
    children: directory.entries,
  };
});

const legacyFileSchema = z.object({
  type: z.literal('file').optional(),
  id: z.string(),
  name: z.string(),
  path: z.string(),
  breadcrumb: z.array(breadcrumbSchema).optional(),
  content_format: z.enum(['markdown', 'html_document']),
  body_markdown: z.string().optional(),
  body_html: z.string().nullable().optional(),
  html_document: z.string().optional(),
  keywords: z.array(z.string()).optional(),
  updated_at: z.string().optional(),
  published_at: nullableStringSchema,
  read_time_minutes: nullableNumberSchema,
  like_count: z.number().optional(),
  comment_count: z.number().optional(),
});

const fileContentSchema = z.object({
  node_id: z.string(),
  content_format: z.enum(['markdown', 'html_document']),
  keywords: z.array(z.string()).optional(),
  body_raw: z.string().optional(),
  body_html: z.string().nullable().optional(),
  status: z.string().optional(),
  published_at: nullableStringSchema,
  reading_time_minutes: nullableNumberSchema,
});

const openApiFileSchema = z.object({
  node: nodeSchema.extend({ kind: z.literal('file') }),
  content: fileContentSchema,
  keywords_public: z.array(z.string()).optional(),
  like_count: z.number().optional(),
  viewer_has_liked: z.boolean().optional(),
  comment_count: z.number().optional(),
  assets: z.array(z.unknown()).optional(),
});

const fileSchema: z.ZodType<FilePayload> = z.union([legacyFileSchema, openApiFileSchema]).transform((file) => {
  if ('content_format' in file) {
    const path = normalizeAbsolutePath(file.path);
    return {
      ...file,
      type: 'file' as const,
      path,
      breadcrumb: file.breadcrumb ?? buildBreadcrumb(path),
      body_html: file.body_html ?? undefined,
      published_at: file.published_at ?? undefined,
      read_time_minutes: file.read_time_minutes ?? undefined,
    };
  }

  const path = normalizeAbsolutePath(file.node.path);
  const isHtmlDocument = file.content.content_format === 'html_document';
  const rawBody = file.content.body_raw ?? '';

  return {
    type: 'file' as const,
    id: file.node.id,
    name: file.node.name,
    path,
    breadcrumb: buildBreadcrumb(path),
    content_format: file.content.content_format,
    body_markdown: isHtmlDocument ? undefined : rawBody,
    body_html: file.content.body_html ?? undefined,
    html_document: isHtmlDocument ? rawBody : undefined,
    keywords: file.keywords_public ?? file.content.keywords ?? [],
    updated_at: file.node.updated_at,
    published_at: file.content.published_at ?? undefined,
    read_time_minutes: file.content.reading_time_minutes ?? undefined,
    like_count: file.like_count,
    comment_count: file.comment_count,
  };
});

const redirectSchema = z.object({
  type: z.literal('redirect'),
  new_path: z.string().transform(normalizeAbsolutePath),
});

const openApiResolveSchema = z.union([
  z.object({ type: z.literal('directory'), directory: directorySchema }),
  z.object({ type: z.literal('file'), file: fileSchema }),
  redirectSchema,
]);

const resolveSchema: z.ZodType<ResolvePayload> = z.union([openApiResolveSchema, fileSchema, directorySchema]).transform((payload) => {
  if ('directory' in payload) {
    return payload.directory;
  }

  if ('file' in payload) {
    return payload.file;
  }

  return payload;
});

const recentSchema = z.object({
  items: z.array(z.union([legacyContentEntrySchema, fileEntrySchema]).transform(toContentEntry)),
});

const legacySearchResultSchema = z.object({
  id: z.string(),
  name: z.string(),
  path: z.string(),
  snippet: z.string(),
  sources: z.array(z.enum(['text', 'semantic', 'keyword'])),
});

const openApiSearchResultSchema = z.object({
  file: fileEntrySchema,
  path: z.string(),
  snippet: z.string(),
  score: z.number().optional(),
  match_sources: z.array(z.enum(['text', 'semantic', 'keyword'])),
}).transform((result): SearchResult => {
  const entry = toContentEntry(result.file);
  return {
    id: entry.id,
    name: entry.name,
    path: normalizeAbsolutePath(result.path || entry.path),
    snippet: result.snippet,
    sources: result.match_sources,
  };
});

const searchResultSchema: z.ZodType<SearchResult> = z.union([legacySearchResultSchema, openApiSearchResultSchema]);

const searchSchema = z.object({
  items: z.array(searchResultSchema),
});

export function fetchRootDirectory(): Promise<DirectoryPayload> {
  return requestJson('/tree', directorySchema);
}

export function resolveContentPath(path: string): Promise<ResolvePayload> {
  return requestJson(`/tree/resolve?path=${encodeURIComponent(normalizeAbsolutePath(path))}`, resolveSchema);
}

export async function fetchRecentFiles(): Promise<ContentEntry[]> {
  const payload = await requestJson('/recent?limit=24&offset=0', recentSchema);
  return payload.items;
}

export async function searchFiles(query: string): Promise<SearchResult[]> {
  const payload = await requestJson(`/search?q=${encodeURIComponent(query)}`, searchSchema);
  return payload.items;
}

function toContentEntry(entry: z.infer<typeof legacyContentEntrySchema> | z.infer<typeof directoryEntrySchema> | z.infer<typeof fileEntrySchema>): ContentEntry {
  if ('id' in entry) {
    return {
      ...entry,
      path: normalizeAbsolutePath(entry.path),
    };
  }

  if ('child_directory_count' in entry) {
    return {
      id: entry.node.id,
      kind: 'directory',
      name: entry.node.name,
      path: normalizeAbsolutePath(entry.node.path),
      sort_order: entry.node.sort_order,
      updated_at: entry.node.updated_at,
      child_directory_count: entry.child_directory_count,
      child_file_count: entry.child_file_count,
    };
  }

  return {
    id: entry.node.id,
    kind: 'file',
    name: entry.node.name,
    path: normalizeAbsolutePath(entry.node.path),
    sort_order: entry.node.sort_order,
    updated_at: entry.node.updated_at,
    content_format: entry.content_format,
    read_time_minutes: entry.reading_time_minutes ?? undefined,
    keywords: entry.keywords ?? [],
    comment_count: entry.comment_count,
  };
}

function normalizeAbsolutePath(path: string): string {
  const trimmed = path.trim();

  if (!trimmed || trimmed === '/') {
    return '/';
  }

  return `/${trimmed.replace(/^\/+|\/+$/g, '')}`;
}

function buildBreadcrumb(path: string): BreadcrumbItem[] {
  const normalized = normalizeAbsolutePath(path);
  const crumbs: BreadcrumbItem[] = [{ name: 'Root', path: '/' }];

  if (normalized === '/') {
    return crumbs;
  }

  const segments = normalized.split('/').filter(Boolean);
  let currentPath = '';

  for (const segment of segments) {
    currentPath = `${currentPath}/${segment}`;
    crumbs.push({
      name: decodeURIComponent(segment).replaceAll('-', ' '),
      path: currentPath,
    });
  }

  return crumbs;
}
