export interface BackendClientOptions {
  baseUrl: string;
  adminToken?: string;
}

/**
 * Thin backend API boundary for MCP tools.
 *
 * This package intentionally does not import database clients, repositories, or SQL.
 * Later Gateway 6 tool slices should add methods here that call the existing HTTP API
 * or a shared service client rather than duplicating business rules in MCP handlers.
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
}
