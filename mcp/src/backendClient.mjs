import { createWriteStream } from "node:fs";
import { mkdir } from "node:fs/promises";
import path from "node:path";

/**
 * Thin backend HTTP API boundary for MCP tools.
 *
 * MCP handlers must stay orchestration-only: no database clients or repositories.
 * All blog state changes go through the same protected HTTP API used by Author UI.
 */
export class BlogBackendClient {
  constructor(options) {
    this.baseUrl = options.baseUrl.replace(/\/+$/, "");
    this.adminToken = options.adminToken;
  }

  async health() {
    return { ok: true, base_url: this.baseUrl };
  }

  async listContentTree() {
    return this.#json("/api/admin/tree");
  }

  async getFile(fileId) {
    return this.#json(`/api/admin/files/${encodeURIComponent(fileId)}/content`);
  }

  async searchFiles({ query, limit, offset }) {
    const params = new URLSearchParams({ q: query });
    if (limit !== undefined) params.set("limit", String(limit));
    if (offset !== undefined) params.set("offset", String(offset));
    return this.#json(`/api/search?${params.toString()}`);
  }

  async createDirectory({ parentId = null, name, slug, sortOrder = 0 }) {
    return this.#json("/api/admin/nodes", {
      method: "POST",
      body: { parent_id: parentId, kind: "directory", name, slug, sort_order: sortOrder },
    });
  }

  async createFile({ parentId = null, name, slug, contentFormat = "markdown", sortOrder = 0 }) {
    return this.#json("/api/admin/nodes", {
      method: "POST",
      body: { parent_id: parentId, kind: "file", name, slug, content_format: contentFormat, sort_order: sortOrder },
    });
  }

  async updateFileContent(fileId, input) {
    return this.#json(`/api/admin/files/${encodeURIComponent(fileId)}/content`, {
      method: "PUT",
      body: input,
    });
  }

  async updateFileSettings(nodeId, input) {
    return this.#json(`/api/admin/nodes/${encodeURIComponent(nodeId)}`, {
      method: "PATCH",
      body: input,
    });
  }

  async publishFile(fileId, expectedRevision) {
    return this.#json(`/api/admin/files/${encodeURIComponent(fileId)}/publish`, {
      method: "POST",
      body: { expected_revision: expectedRevision },
    });
  }

  async unpublishFile(fileId, expectedRevision) {
    return this.#json(`/api/admin/files/${encodeURIComponent(fileId)}/unpublish`, {
      method: "POST",
      body: { expected_revision: expectedRevision },
    });
  }

  async moveNode(nodeId, input) {
    return this.#json(`/api/admin/nodes/${encodeURIComponent(nodeId)}/move`, {
      method: "POST",
      body: { new_parent_id: input.newParentId ?? null, expected_version: input.expectedVersion },
    });
  }

  async reorderChildren(parentId, input) {
    return this.#json(`/api/admin/nodes/${encodeURIComponent(parentId)}/children/order`, {
      method: "PUT",
      body: { child_ids: input.childIds, expected_version: input.expectedVersion },
    });
  }

  async deleteNode(nodeId) {
    return this.#jsonOrEmpty(`/api/admin/nodes/${encodeURIComponent(nodeId)}`, { method: "DELETE" });
  }

  async uploadAsset({ fileId, filename, mimeType = "application/octet-stream", dataBase64 }) {
    const data = Buffer.from(dataBase64, "base64");
    const form = new FormData();
    form.append("file", new Blob([data], { type: mimeType }), filename);
    return this.#json(`/api/admin/files/${encodeURIComponent(fileId)}/assets`, {
      method: "POST",
      body: form,
      isMultipart: true,
    });
  }

  async deleteAsset(assetId) {
    return this.#jsonOrEmpty(`/api/admin/assets/${encodeURIComponent(assetId)}`, { method: "DELETE" });
  }

  async listAssets(fileId) {
    return this.#json(`/api/admin/files/${encodeURIComponent(fileId)}/assets`);
  }

  async rebuildSearchIndex() {
    return this.#jsonOrEmpty("/api/admin/search-index/rebuild", { method: "POST" });
  }

  async exportBackup({ outputDir }) {
    const tree = await this.listContentTree();
    const fileNodes = Array.isArray(tree.nodes) ? tree.nodes.filter((node) => node.kind === "file") : [];
    const files = [];
    for (const node of fileNodes) {
      files.push({ node_id: node.id, path: node.url_path ?? node.path, version_state: await this.getFile(node.id) });
    }

    await mkdir(outputDir, { recursive: true });
    const timestamp = new Date().toISOString().replaceAll(":", "-");
    const filePath = path.join(outputDir, `aeolian-backup-${timestamp}.json`);
    const stream = createWriteStream(filePath, { flags: "wx", encoding: "utf8" });
    const backup = { exported_at: new Date().toISOString(), tree, files };
    await new Promise((resolve, reject) => {
      stream.on("error", reject);
      stream.on("finish", resolve);
      stream.end(JSON.stringify(backup, null, 2));
    });
    return { file_path: filePath, node_count: Array.isArray(tree.nodes) ? tree.nodes.length : undefined, file_count: files.length };
  }

  async #json(pathname, options = {}) {
    const response = await this.#fetch(pathname, options);
    return response.json();
  }

  async #jsonOrEmpty(pathname, options = {}) {
    const response = await this.#fetch(pathname, options);
    if (response.status === 204) return { ok: true };
    const text = await response.text();
    return text ? JSON.parse(text) : { ok: true };
  }

  async #fetch(pathname, { method = "GET", body, isMultipart = false } = {}) {
    const headers = { Accept: "application/json" };
    if (this.adminToken) headers.Authorization = `Bearer ${this.adminToken}`;
    if (body !== undefined && !isMultipart) headers["Content-Type"] = "application/json";
    const response = await fetch(`${this.baseUrl}${pathname}`, {
      method,
      headers,
      body: body === undefined ? undefined : isMultipart ? body : JSON.stringify(body),
    });
    if (!response.ok) {
      throw new Error(await readError(response));
    }
    return response;
  }
}

async function readError(response) {
  const fallback = `Backend request failed: ${response.status} ${response.statusText}`;
  try {
    const payload = await response.json();
    if (payload && typeof payload === "object") {
      if (typeof payload.error === "string") return payload.error;
      if (payload.error && typeof payload.error.message === "string") return payload.error.message;
      if (typeof payload.message === "string") return payload.message;
    }
  } catch {
    // keep fallback
  }
  return fallback;
}
