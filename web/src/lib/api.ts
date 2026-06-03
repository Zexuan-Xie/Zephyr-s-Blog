import { z } from 'zod';
import type { ContentEntry, DirectoryPayload, FilePayload, ResolvePayload, SearchResult } from './types';

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

const breadcrumbSchema = z.object({
  name: z.string(),
  path: z.string(),
});

const contentEntrySchema = z.object({
  id: z.string(),
  kind: z.enum(['directory', 'file']),
  name: z.string(),
  path: z.string(),
  sort_order: z.number().optional(),
  updated_at: z.string().optional(),
  child_directory_count: z.number().optional(),
  child_file_count: z.number().optional(),
  content_format: z.enum(['markdown', 'html_document']).optional(),
  read_time_minutes: z.number().optional(),
  keywords: z.array(z.string()).optional(),
});

const directorySchema: z.ZodType<DirectoryPayload> = z.object({
  type: z.literal('directory').optional(),
  id: z.string(),
  name: z.string(),
  path: z.string(),
  breadcrumb: z.array(breadcrumbSchema).optional(),
  children: z.array(contentEntrySchema),
});

const fileSchema: z.ZodType<FilePayload> = z.object({
  type: z.literal('file').optional(),
  id: z.string(),
  name: z.string(),
  path: z.string(),
  breadcrumb: z.array(breadcrumbSchema).optional(),
  content_format: z.enum(['markdown', 'html_document']),
  body_markdown: z.string().optional(),
  body_html: z.string().optional(),
  html_document: z.string().optional(),
  keywords: z.array(z.string()).optional(),
  updated_at: z.string().optional(),
  published_at: z.string().optional(),
  read_time_minutes: z.number().optional(),
  like_count: z.number().optional(),
});

const redirectSchema = z.object({
  type: z.literal('redirect'),
  new_path: z.string(),
});

const resolveSchema: z.ZodType<ResolvePayload> = z.union([redirectSchema, fileSchema, directorySchema]);

const recentSchema = z.object({
  items: z.array(contentEntrySchema),
});

const searchResultSchema: z.ZodType<SearchResult> = z.object({
  id: z.string(),
  name: z.string(),
  path: z.string(),
  snippet: z.string(),
  sources: z.array(z.enum(['text', 'semantic', 'keyword'])),
});

const searchSchema = z.object({
  items: z.array(searchResultSchema),
});

export function fetchRootDirectory(): Promise<DirectoryPayload> {
  return requestJson('/tree', directorySchema);
}

export function resolveContentPath(path: string): Promise<ResolvePayload> {
  return requestJson(`/tree/resolve?path=${encodeURIComponent(path)}`, resolveSchema);
}

export async function fetchRecentFiles(): Promise<ContentEntry[]> {
  const payload = await requestJson('/recent?limit=24&offset=0', recentSchema);
  return payload.items;
}

export async function searchFiles(query: string): Promise<SearchResult[]> {
  const payload = await requestJson(`/search?q=${encodeURIComponent(query)}`, searchSchema);
  return payload.items;
}
