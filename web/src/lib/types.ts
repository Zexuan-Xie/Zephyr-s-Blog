export type NodeKind = "directory" | "file";
export type ContentFormat = "markdown" | "html_document";

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
  like_count?: number;
  comment_count?: number;
  content_format?: ContentFormat;
  read_time_minutes?: number;
  keywords?: string[];
}

export interface DirectoryPayload {
  type?: "directory";
  id: string;
  name: string;
  path: string;
  breadcrumb?: BreadcrumbItem[];
  children: ContentEntry[];
}

export interface FilePayload {
  type?: "file";
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
  viewer_has_liked?: boolean;
  assets: FileAsset[];
}

export interface FileAsset {
  id: string;
  file_node_id: string;
  filename: string;
  mime_type: string;
  size_bytes: number;
  storage_provider: string;
  storage_key?: string;
  public_url: string;
  created_at: string;
}

export interface PublicUser {
  id: string;
  display_name: string;
}

export interface CommentItem {
  id: string;
  file_node_id: string;
  parent_id?: string | null;
  reply_to_user_id?: string | null;
  user: PublicUser;
  body: string;
  created_at: string;
  updated_at?: string;
  deleted_at?: string | null;
  deleted: boolean;
  like_count: number;
  viewer_has_liked?: boolean;
  replies: CommentItem[];
}

export interface CommentThread {
  file_id: string;
  comments: CommentItem[];
}

export interface LikeState {
  liked: boolean;
  like_count: number;
}

export interface RedirectPayload {
  type: "redirect";
  new_path: string;
}

export type ResolvePayload = DirectoryPayload | FilePayload | RedirectPayload;

export interface SearchResult {
  id: string;
  name: string;
  path: string;
  snippet: string;
  sources: Array<"text" | "semantic" | "keyword">;
}

export interface AdminTreeNode {
  id: string;
  parent_id?: string | null;
  kind: NodeKind;
  name: string;
  path: string;
  status: "draft" | "published";
  children: AdminTreeNode[];
  content_format?: ContentFormat;
}

export interface AdminTreeResponse {
  roots: AdminTreeNode[];
}

export interface AdminNodeDetail {
  node: {
    id: string;
    parent_id?: string | null;
    kind: NodeKind;
    name: string;
    slug: string;
    path: string;
    sort_order: number;
    created_at?: string;
    updated_at?: string;
  };
  content?: {
    node_id: string;
    content_format: ContentFormat;
    keywords: string[];
    body_raw: string;
    body_html?: string | null;
    search_text: string;
    status: "draft" | "published";
    published_at?: string | null;
    embedding_model?: string | null;
    embedding_status: "pending" | "ready" | "failed";
    embedding_error?: string | null;
    embedding_updated_at?: string | null;
  } | null;
  assets: FileAsset[];
  redirects_created: Array<{
    id: string;
    old_path: string;
    new_path: string;
    node_id: string;
    created_at: string;
  }>;
}

export interface ReorderChildrenInput {
  child_ids: string[];
  expected_version: number;
}

export interface ReorderChildrenResponse {
  parent_id: string;
  child_ids: string[];
  version: number;
}

export interface MoveNodeInput {
  new_parent_id?: string | null;
  expected_version: number;
}

export interface MovePreviewResponse {
  node_id: string;
  destination_path: string;
  affected_paths: string[];
  redirects: Array<{ old_path: string; new_path: string }>;
  blocked_reasons: string[];
}

export interface EmbeddingState {
  file_id: string;
  provider: "qwen";
  model: "text-embedding-v4";
  dimensions: 1024;
  status: "pending" | "ready" | "failed";
  error?: string | null;
  updated_at?: string | null;
}

export interface CurrentUser {
  id: string;
  email: string;
  role: "admin" | "reader";
  display_name?: string | null;
  provider: string;
  created_at: string;
}
