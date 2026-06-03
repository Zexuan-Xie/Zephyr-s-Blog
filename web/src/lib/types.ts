export type NodeKind = 'directory' | 'file';
export type ContentFormat = 'markdown' | 'html_document';

export interface BreadcrumbItem {
  name: string;
  path: string;
}

export interface ContentEntry {
  id: string;
  kind: NodeKind;
  name: string;
  path: string;
  sort_order?: number;
  updated_at?: string;
  child_directory_count?: number;
  child_file_count?: number;
  comment_count?: number;
  content_format?: ContentFormat;
  read_time_minutes?: number;
  keywords?: string[];
}

export interface DirectoryPayload {
  type?: 'directory';
  id: string;
  name: string;
  path: string;
  breadcrumb?: BreadcrumbItem[];
  children: ContentEntry[];
}

export interface FilePayload {
  type?: 'file';
  id: string;
  name: string;
  path: string;
  breadcrumb?: BreadcrumbItem[];
  content_format: ContentFormat;
  body_markdown?: string;
  body_html?: string;
  html_document?: string;
  keywords?: string[];
  updated_at?: string;
  published_at?: string;
  read_time_minutes?: number;
  like_count?: number;
  comment_count?: number;
}

export interface RedirectPayload {
  type: 'redirect';
  new_path: string;
}

export type ResolvePayload = DirectoryPayload | FilePayload | RedirectPayload;

export interface SearchResult {
  id: string;
  name: string;
  path: string;
  snippet: string;
  sources: Array<'text' | 'semantic' | 'keyword'>;
}
