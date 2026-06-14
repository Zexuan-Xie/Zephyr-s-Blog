export interface BackendClientOptions {
  baseUrl: string;
  adminToken?: string;
}

interface RequestOptions {
  method?: string;
  body?: unknown;
  isMultipart?: boolean;
}

/**
 * Thin backend HTTP API boundary for MCP tools.
 *
 * MCP handlers stay orchestration-only: no database clients or repositories.
 * All blog state changes go through the same protected HTTP API used by Author UI.
 */
export class BlogBackendClient {
  readonly baseUrl: string;
  readonly adminToken?: string;

  constructor(options: BackendClientOptions) {
    this.baseUrl = options.baseUrl.replace(/\/+$/, "");
    this.adminToken = options.adminToken;
  }

  async health(): Promise<{ ok: boolean; base_url: string }> {
    return { ok: true, base_url: this.baseUrl };
  }

  async listContentTree(): Promise<any> {
    return this.#json("/api/admin/tree");
  }

  async getFile(fileId: string): Promise<any> {
    return this.#json(`/api/admin/files/${encodeURIComponent(fileId)}/content`);
  }

  async searchFiles(args: { query: string; limit?: number; offset?: number }): Promise<any> {
    const params = new URLSearchParams({ q: args.query });
    if (args.limit !== undefined) params.set("limit", String(args.limit));
    if (args.offset !== undefined) params.set("offset", String(args.offset));
    return this.#json(`/api/search?${params.toString()}`);
  }

  async createDirectory(args: { parentId?: string | null; name: string; slug?: string; sortOrder?: number }): Promise<any> {
    return this.#json("/api/admin/nodes", {
      method: "POST",
      body: {
        parent_id: args.parentId ?? null,
        kind: "directory",
        name: args.name,
        slug: args.slug,
        sort_order: args.sortOrder ?? 0,
      },
    });
  }

  async createFile(args: {
    parentId?: string | null;
    name: string;
    slug?: string;
    contentFormat?: "markdown" | "html_document";
    sortOrder?: number;
  }): Promise<any> {
    return this.#json("/api/admin/nodes", {
      method: "POST",
      body: {
        parent_id: args.parentId ?? null,
        kind: "file",
        name: args.name,
        slug: args.slug,
        content_format: args.contentFormat ?? "markdown",
        sort_order: args.sortOrder ?? 0,
      },
    });
  }

  async updateFileContent(fileId: string, input: Record<string, unknown>): Promise<any> {
    return this.#json(`/api/admin/files/${encodeURIComponent(fileId)}/content`, { method: "PUT", body: input });
  }

  async updateFileSettings(nodeId: string, input: Record<string, unknown>): Promise<any> {
    return this.#json(`/api/admin/nodes/${encodeURIComponent(nodeId)}`, { method: "PATCH", body: input });
  }

  async publishFile(fileId: string, expectedRevision: number): Promise<any> {
    return this.#json(`/api/admin/files/${encodeURIComponent(fileId)}/publish`, {
      method: "POST",
      body: { expected_revision: expectedRevision },
    });
  }

  async unpublishFile(fileId: string, expectedRevision: number): Promise<any> {
    return this.#json(`/api/admin/files/${encodeURIComponent(fileId)}/unpublish`, {
      method: "POST",
      body: { expected_revision: expectedRevision },
    });
  }

  async moveNode(nodeId: string, input: { newParentId?: string | null; expectedVersion: number }): Promise<any> {
    return this.#json(`/api/admin/nodes/${encodeURIComponent(nodeId)}/move`, {
      method: "POST",
      body: { new_parent_id: input.newParentId ?? null, expected_version: input.expectedVersion },
    });
  }

  async reorderChildren(parentId: string, input: { childIds: string[]; expectedVersion: number }): Promise<any> {
    return this.#json(`/api/admin/nodes/${encodeURIComponent(parentId)}/children/order`, {
      method: "PUT",
      body: { child_ids: input.childIds, expected_version: input.expectedVersion },
    });
  }

  async deleteNode(nodeId: string): Promise<any> {
    return this.#jsonOrEmpty(`/api/admin/nodes/${encodeURIComponent(nodeId)}`, { method: "DELETE" });
  }

  async uploadAsset(args: { fileId: string; filename: string; mimeType?: string; dataBase64: string }): Promise<any> {
    const data = Buffer.from(args.dataBase64, "base64");
    const form = new FormData();
    form.append("file", new Blob([data], { type: args.mimeType ?? "application/octet-stream" }), args.filename);
    return this.#json(`/api/admin/files/${encodeURIComponent(args.fileId)}/assets`, {
      method: "POST",
      body: form,
      isMultipart: true,
    });
  }

  async deleteAsset(assetId: string): Promise<any> {
    return this.#jsonOrEmpty(`/api/admin/assets/${encodeURIComponent(assetId)}`, { method: "DELETE" });
  }

  async listAssets(fileId: string): Promise<any> {
    return this.#json(`/api/admin/files/${encodeURIComponent(fileId)}/assets`);
  }

  async rebuildSearchIndex(): Promise<any> {
    return this.#jsonOrEmpty("/api/admin/search-index/rebuild", { method: "POST" });
  }

  async #json(pathname: string, options: RequestOptions = {}): Promise<any> {
    const response = await this.#fetch(pathname, options);
    return response.json();
  }

  async #jsonOrEmpty(pathname: string, options: RequestOptions = {}): Promise<any> {
    const response = await this.#fetch(pathname, options);
    if (response.status === 204) return { ok: true };
    const text = await response.text();
    return text ? JSON.parse(text) : { ok: true };
  }

  async #fetch(pathname: string, { method = "GET", body, isMultipart = false }: RequestOptions = {}): Promise<Response> {
    const headers: Record<string, string> = { Accept: "application/json" };
    if (this.adminToken) headers.Authorization = `Bearer ${this.adminToken}`;
    if (body !== undefined && !isMultipart) headers["Content-Type"] = "application/json";
    const response = await fetch(`${this.baseUrl}${pathname}`, {
      method,
      headers,
      body: body === undefined ? undefined : isMultipart ? (body as BodyInit) : JSON.stringify(body),
    });
    if (!response.ok) {
      throw new Error(await readError(response));
    }
    return response;
  }
}

async function readError(response: Response): Promise<string> {
  const fallback = `Backend request failed: ${response.status} ${response.statusText}`;
  try {
    const payload = await response.json();
    if (payload && typeof payload === "object") {
      const record = payload as Record<string, any>;
      if (typeof record.error === "string") return record.error;
      if (record.error && typeof record.error.message === "string") return record.error.message;
      if (typeof record.message === "string") return record.message;
    }
  } catch {
    // Preserve status fallback for empty or malformed error bodies.
  }
  return fallback;
}
