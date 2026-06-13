/**
 * Thin backend API boundary for MCP tools.
 *
 * This package intentionally does not import database clients, repositories, or SQL.
 * Later Gateway 6 tool slices should add methods here that call the existing HTTP API
 * or a shared service client rather than duplicating business rules in MCP handlers.
 */
export class BlogBackendClient {
  constructor(options) {
    this.baseUrl = options.baseUrl.replace(/\/+$/, "");
    this.adminToken = options.adminToken;
  }

  async health() {
    return { ok: true, base_url: this.baseUrl };
  }
}
