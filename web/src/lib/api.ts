import { z } from 'zod';
import { getToken } from './auth';
import type { BreadcrumbItem, CommentItem, CommentThread, ContentEntry, CurrentUser, DirectoryPayload, FileAsset, FilePayload, LikeState, ResolvePayload, SearchResult, AdminNodeDetail, AdminTreeNode, AdminTreeResponse, ContentFormat, EmbeddingState, NodeKind } from './types';

const apiBase = '/api';

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

async function requestJson<T>(path: string, schema: z.ZodType<T>, init: RequestInit = {}): Promise<T> {
  const response = await fetch(`${apiBase}${path}`, {
    ...init,
    headers: { Accept: 'application/json', ...init.headers },
  });

  if (!response.ok) {
    throw new ApiError(response.status, await readErrorMessage(response));
  }

  const json: unknown = await response.json();
  return schema.parse(json);
}

async function readErrorMessage(response: Response): Promise<string> {
  const fallback = `Request failed: ${response.status} ${response.statusText}`;

  try {
    const payload: unknown = await response.json();
    if (typeof payload !== 'object' || payload === null) return fallback;

    if ('message' in payload && typeof payload.message === 'string') {
      return payload.message;
    }
    if ('error' in payload && typeof payload.error === 'string') {
      return payload.error;
    }
    if ('error' in payload && typeof payload.error === 'object' && payload.error !== null
      && 'message' in payload.error && typeof payload.error.message === 'string') {
      return payload.error.message;
    }
  } catch {
    // Preserve the HTTP status fallback when an error body is empty or malformed.
  }

  return fallback;
}

const nullableStringSchema = z.string().nullable().optional();
const nullableNumberSchema = z.number().nullable().optional();

const publicUserSchema = z.object({
  id: z.string(),
  display_name: z.string(),
});

const currentUserSchema: z.ZodType<CurrentUser> = z.object({
  id: z.string(),
  email: z.string(),
  role: z.enum(['admin', 'reader']),
  display_name: z.string().nullable().optional(),
  provider: z.string(),
  created_at: z.string(),
});

const commentSchema: z.ZodType<CommentItem> = z.lazy(() => z.object({
  id: z.string(),
  file_node_id: z.string(),
  parent_id: z.string().nullable().optional(),
  reply_to_user_id: z.string().nullable().optional(),
  user: publicUserSchema,
  body: z.string(),
  created_at: z.string(),
  updated_at: z.string().optional(),
  deleted_at: z.string().nullable().optional(),
  deleted: z.boolean(),
  like_count: z.number(),
  viewer_has_liked: z.boolean().optional(),
  replies: z.array(commentSchema).default([]),
}));

const commentThreadSchema: z.ZodType<CommentThread> = z.object({
  file_id: z.string(),
  comments: z.array(commentSchema),
});

const likeStateSchema: z.ZodType<LikeState> = z.object({
  liked: z.boolean(),
  like_count: z.number(),
});

const fileAssetSchema: z.ZodType<FileAsset> = z.object({
  id: z.string(),
  file_node_id: z.string(),
  filename: z.string(),
  mime_type: z.string(),
  size_bytes: z.number(),
  storage_provider: z.string(),
  storage_key: z.string().optional(),
  public_url: z.string(),
  created_at: z.string(),
});

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
  assets: z.array(fileAssetSchema).optional(),
});

const adminFileContentSchema = z.object({
  node_id: z.string(),
  content_format: z.enum(['markdown', 'html_document']),
  keywords: z.array(z.string()).default([]),
  body_raw: z.string().default(''),
  body_html: z.string().nullable().optional(),
  search_text: z.string().default(''),
  status: z.enum(['draft', 'published']),
  published_at: nullableStringSchema,
  embedding_model: nullableStringSchema,
  embedding_status: z.enum(['pending', 'ready', 'failed']),
  embedding_error: nullableStringSchema,
  embedding_updated_at: nullableStringSchema,
});

const pathRedirectSchema = z.object({
  id: z.string(),
  old_path: z.string(),
  new_path: z.string(),
  node_id: z.string(),
  created_at: z.string(),
});


const adminTreeNodeSchema: z.ZodType<AdminTreeNode> = z.lazy(() => z.object({
  id: z.string(),
  parent_id: z.string().nullable().optional(),
  kind: z.enum(['directory', 'file']),
  name: z.string(),
  path: z.string(),
  status: z.enum(['draft', 'published']),
  children: z.array(adminTreeNodeSchema).default([]),
  content_format: z.enum(['markdown', 'html_document']).optional(),
}));

const adminTreeResponseSchema: z.ZodType<AdminTreeResponse> = z.object({
  roots: z.array(adminTreeNodeSchema),
});

const adminNodeDetailSchema: z.ZodType<AdminNodeDetail> = z.object({
  node: nodeSchema.extend({ slug: z.string(), sort_order: z.number(), parent_id: z.string().nullable().optional() }),
  content: adminFileContentSchema.nullable().optional(),
  assets: z.array(fileAssetSchema).default([]),
  redirects_created: z.array(pathRedirectSchema).default([]),
});

const embeddingStateSchema: z.ZodType<EmbeddingState> = z.object({
  file_id: z.string(),
  provider: z.literal('qwen'),
  model: z.literal('text-embedding-v4'),
  dimensions: z.literal(1024),
  status: z.enum(['pending', 'ready', 'failed']),
  error: nullableStringSchema,
  updated_at: nullableStringSchema,
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
      assets: [],
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
    viewer_has_liked: file.viewer_has_liked,
    assets: file.assets ?? [],
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

export function fetchCurrentUser(): Promise<CurrentUser> {
  return requestJson('/auth/me', currentUserSchema, authInit());
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
    like_count: entry.like_count,
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


export function fetchCommentThread(fileId: string): Promise<CommentThread> {
  return requestJson(`/files/${encodeURIComponent(fileId)}/comments`, commentThreadSchema, authInit());
}

export function createComment(fileId: string, body: string, parentId?: string | null, replyToUserId?: string | null): Promise<CommentItem> {
  return requestJson(`/files/${encodeURIComponent(fileId)}/comments`, commentSchema, jsonAuthInit('POST', {
    body,
    parent_id: parentId ?? null,
    reply_to_user_id: replyToUserId ?? null,
  }));
}

export async function deleteComment(commentId: string): Promise<void> {
  const response = await fetch(`${apiBase}/comments/${encodeURIComponent(commentId)}`, jsonAuthInit('DELETE'));
  if (!response.ok) {
    throw new Error(`Request failed: ${response.status} ${response.statusText}`);
  }
}

export function likeFile(fileId: string): Promise<LikeState> {
  return requestJson(`/files/${encodeURIComponent(fileId)}/like`, likeStateSchema, jsonAuthInit('PUT'));
}

export function unlikeFile(fileId: string): Promise<LikeState> {
  return requestJson(`/files/${encodeURIComponent(fileId)}/like`, likeStateSchema, jsonAuthInit('DELETE'));
}

export function likeComment(commentId: string): Promise<LikeState> {
  return requestJson(`/comments/${encodeURIComponent(commentId)}/like`, likeStateSchema, jsonAuthInit('PUT'));
}

export function unlikeComment(commentId: string): Promise<LikeState> {
  return requestJson(`/comments/${encodeURIComponent(commentId)}/like`, likeStateSchema, jsonAuthInit('DELETE'));
}

function authInit(init: RequestInit = {}): RequestInit {
  const token = getToken();
  return token ? {
    ...init,
    headers: { ...init.headers, Authorization: `Bearer ${token}` },
  } : init;
}

function jsonAuthInit(method: string, body?: unknown): RequestInit {
  return authInit({
    method,
    headers: { 'Content-Type': 'application/json' },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
}

export interface CreateAdminNodeInput {
  parent_id?: string | null;
  kind: NodeKind;
  name: string;
  slug?: string;
  sort_order?: number;
  content_format?: ContentFormat;
}

export interface UpdateAdminNodeInput {
  parent_id?: string | null;
  name?: string;
  slug?: string;
  sort_order?: number;
}

export interface UpsertFileContentInput {
  content_format: ContentFormat;
  body_raw: string;
  body_html?: string | null;
  keywords: string[];
}

export function fetchAdminTree(): Promise<AdminTreeResponse> {
  return requestJson('/admin/tree', adminTreeResponseSchema, authInit());
}

export function fetchAdminNode(nodeId: string): Promise<AdminNodeDetail> {
  return requestJson(`/admin/nodes/${encodeURIComponent(nodeId)}`, adminNodeDetailSchema, authInit());
}

export function createAdminNode(input: CreateAdminNodeInput): Promise<AdminNodeDetail> {
  return requestJson('/admin/nodes', adminNodeDetailSchema, jsonAuthInit('POST', input));
}

export function updateAdminNode(nodeId: string, input: UpdateAdminNodeInput): Promise<AdminNodeDetail> {
  return requestJson(`/admin/nodes/${encodeURIComponent(nodeId)}`, adminNodeDetailSchema, jsonAuthInit('PATCH', input));
}

export async function deleteAdminNode(nodeId: string): Promise<void> {
  const response = await fetch(`${apiBase}/admin/nodes/${encodeURIComponent(nodeId)}`, jsonAuthInit('DELETE'));
  if (!response.ok) {
    throw new Error(`Request failed: ${response.status} ${response.statusText}`);
  }
}

export function upsertFileContent(fileId: string, input: UpsertFileContentInput): Promise<AdminNodeDetail['content']> {
  return requestJson(`/admin/files/${encodeURIComponent(fileId)}/content`, adminFileContentSchema, jsonAuthInit('PUT', input));
}

export function publishFile(fileId: string): Promise<AdminNodeDetail['content']> {
  return requestJson(`/admin/files/${encodeURIComponent(fileId)}/publish`, adminFileContentSchema, jsonAuthInit('POST'));
}

export function unpublishFile(fileId: string): Promise<AdminNodeDetail['content']> {
  return requestJson(`/admin/files/${encodeURIComponent(fileId)}/unpublish`, adminFileContentSchema, jsonAuthInit('POST'));
}

export function refreshEmbedding(fileId: string): Promise<EmbeddingState> {
  return requestJson(`/admin/files/${encodeURIComponent(fileId)}/refresh-embedding`, embeddingStateSchema, jsonAuthInit('POST'));
}

export async function rebuildSearchIndex(): Promise<void> {
  const response = await fetch(`${apiBase}/admin/search-index/rebuild`, jsonAuthInit('POST'));
  if (!response.ok) {
    throw new Error(`Request failed: ${response.status} ${response.statusText}`);
  }
}


export function uploadAsset(fileId: string, file: File): Promise<FileAsset> {
  const form = new FormData();
  form.append('file', file);
  return requestJson(`/admin/files/${encodeURIComponent(fileId)}/assets`, fileAssetSchema, authInit({
    method: 'POST',
    body: form,
  }));
}

export async function deleteAsset(assetId: string): Promise<void> {
  const response = await fetch(`${apiBase}/admin/assets/${encodeURIComponent(assetId)}`, jsonAuthInit('DELETE'));
  if (!response.ok) {
    throw new Error(`Request failed: ${response.status} ${response.statusText}`);
  }
}
