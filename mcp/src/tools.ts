import type { ZodRawShape } from "zod";
import type { BlogMcpConfig } from "./config.js";
import { assertEnabled } from "./config.js";
import { summarizeArgs, writeAudit } from "./audit.js";
import type { BlogBackendClient } from "./backendClient.js";

export interface ToolResultContent {
  type: "text";
  text: string;
}

export interface ToolResult {
  [key: string]: unknown;
  content: ToolResultContent[];
  isError?: boolean;
}

export interface ToolDefinition {
  name: string;
  title: string;
  description: string;
  inputSchema: ZodRawShape;
  destructive: boolean;
  handler: (args: Record<string, unknown>) => Promise<ToolResult>;
}

function textResult(payload: unknown): ToolResult {
  return { content: [{ type: "text", text: typeof payload === "string" ? payload : JSON.stringify(payload, null, 2) }] };
}

function errorResult(error: unknown): ToolResult {
  return { content: [{ type: "text", text: error instanceof Error ? error.message : String(error) }], isError: true };
}

export async function runGuardedTool<T extends Record<string, unknown>>(
  config: BlogMcpConfig,
  tool: string,
  destructive: boolean,
  args: T,
  operation: () => Promise<ToolResult>,
): Promise<ToolResult> {
  try {
    assertEnabled(config);
  } catch (error) {
    await writeAudit(config.auditLogPath, {
      tool,
      destructive,
      args_summary: summarizeArgs(args),
      result: "refused",
      message: error instanceof Error ? error.message : String(error),
    });
    return errorResult(error);
  }

  try {
    const result = await operation();
    await writeAudit(config.auditLogPath, {
      tool,
      destructive,
      args_summary: summarizeArgs(args),
      result: "ok",
    });
    return result;
  } catch (error) {
    await writeAudit(config.auditLogPath, {
      tool,
      destructive,
      args_summary: summarizeArgs(args),
      result: "error",
      message: error instanceof Error ? error.message : String(error),
    });
    return errorResult(error);
  }
}

export function buildToolDefinitions(config: BlogMcpConfig, client: BlogBackendClient): ToolDefinition[] {
  return [
    tool(
      "health_check",
      "Blog MCP health check",
      "Non-destructive tool proving enable/kill-switch guard, audit JSONL, and backend API-client boundary.",
      {},
      false,
      (args) => runGuardedTool(config, "health_check", false, args, async () => textResult(await client.health())),
    ),
    tool("list_content_tree", "List content tree", "Return the protected Author content tree.", {}, false,
      guarded(config, client, "list_content_tree", false, z.object({}), (api) => api.listContentTree())),
    tool("get_file", "Get file", "Return current/published version state for a file.", { file_id: uuidSchema }, false,
      guarded(config, client, "get_file", false, z.object({ file_id: uuidSchema }), (api, args) => api.getFile(args.file_id))),
    tool("search_files", "Search files", "Search public files through the backend search API.", { query: nonEmptyString, limit: z.number().int().nonnegative().optional(), offset: z.number().int().nonnegative().optional() }, false,
      guarded(config, client, "search_files", false, z.object({ query: nonEmptyString, limit: z.number().int().nonnegative().optional(), offset: z.number().int().nonnegative().optional() }), (api, args) => api.searchFiles(args))),

    tool("create_directory", "Create directory", "Create a directory through the protected backend API.", { parent_id: uuidSchema.nullable().optional(), name: nonEmptyString, slug: z.string().optional(), sort_order: z.number().int().optional() }, true,
      guarded(config, client, "create_directory", true, z.object({ parent_id: uuidSchema.nullable().optional(), name: nonEmptyString, slug: z.string().optional(), sort_order: z.number().int().optional() }), (api, args) => api.createDirectory({ parentId: args.parent_id ?? null, name: args.name, slug: args.slug, sortOrder: args.sort_order ?? 0 }))),
    tool("create_file", "Create file", "Create a file through the protected backend API.", { parent_id: uuidSchema.nullable().optional(), name: nonEmptyString, slug: z.string().optional(), content_format: contentFormatSchema.optional(), sort_order: z.number().int().optional() }, true,
      guarded(config, client, "create_file", true, z.object({ parent_id: uuidSchema.nullable().optional(), name: nonEmptyString, slug: z.string().optional(), content_format: contentFormatSchema.optional(), sort_order: z.number().int().optional() }), (api, args) => api.createFile({ parentId: args.parent_id ?? null, name: args.name, slug: args.slug, contentFormat: args.content_format ?? "markdown", sortOrder: args.sort_order ?? 0 }))),
    tool("update_file_content", "Update file content", "Update current draft content with revision protection.", { file_id: uuidSchema, expected_revision: positiveRevision, content_format: contentFormatSchema, body_raw: z.string(), body_html: z.string().nullable().optional(), keywords: z.array(z.string()).optional(), search_text: z.string().optional() }, true,
      guarded(config, client, "update_file_content", true, z.object({ file_id: uuidSchema, expected_revision: positiveRevision, content_format: contentFormatSchema, body_raw: z.string(), body_html: z.string().nullable().optional(), keywords: z.array(z.string()).optional(), search_text: z.string().optional() }), (api, args) => api.updateFileContent(args.file_id, { expected_revision: args.expected_revision, content_format: args.content_format, body_raw: args.body_raw, body_html: args.body_html, keywords: args.keywords ?? [], search_text: args.search_text ?? "" }))),
    tool("update_file_settings", "Update file settings", "Rename or move path settings for a node.", { node_id: uuidSchema, name: z.string().optional(), url_path: z.string().optional(), parent_id: uuidSchema.nullable().optional() }, true,
      guarded(config, client, "update_file_settings", true, z.object({ node_id: uuidSchema, name: z.string().optional(), url_path: z.string().optional(), parent_id: uuidSchema.nullable().optional() }), (api, args) => api.updateFileSettings(args.node_id, { name: args.name, url_path: args.url_path, parent_id: args.parent_id }))),

    tool("publish_file", "Publish file", "Publish current draft snapshot; expected_revision is required.", { file_id: uuidSchema, expected_revision: positiveRevision }, true,
      guarded(config, client, "publish_file", true, z.object({ file_id: uuidSchema, expected_revision: positiveRevision }), (api, args) => api.publishFile(args.file_id, args.expected_revision))),
    tool("unpublish_file", "Unpublish file", "Unpublish visible snapshot; expected_revision is required.", { file_id: uuidSchema, expected_revision: positiveRevision }, true,
      guarded(config, client, "unpublish_file", true, z.object({ file_id: uuidSchema, expected_revision: positiveRevision }), (api, args) => api.unpublishFile(args.file_id, args.expected_revision))),

    tool("move_node", "Move node", "Move a node to a new parent.", { node_id: uuidSchema, new_parent_id: uuidSchema.nullable().optional(), expected_version: positiveRevision }, true,
      guarded(config, client, "move_node", true, z.object({ node_id: uuidSchema, new_parent_id: uuidSchema.nullable().optional(), expected_version: positiveRevision }), (api, args) => api.moveNode(args.node_id, { newParentId: args.new_parent_id ?? null, expectedVersion: args.expected_version }))),
    tool("reorder_children", "Reorder children", "Set the order of a parent directory's children.", { parent_id: uuidSchema, child_ids: z.array(uuidSchema), expected_version: positiveRevision }, true,
      guarded(config, client, "reorder_children", true, z.object({ parent_id: uuidSchema, child_ids: z.array(uuidSchema), expected_version: positiveRevision }), (api, args) => api.reorderChildren(args.parent_id, { childIds: args.child_ids, expectedVersion: args.expected_version }))),
    tool("delete_node", "Delete node", "Delete an unpublished empty node. Requires confirm=true.", { node_id: uuidSchema, confirm: z.boolean() }, true,
      guarded(config, client, "delete_node", true, z.object({ node_id: uuidSchema, confirm: z.boolean() }), (api, args) => { requireConfirm(args, "delete_node requires confirm=true after backup/export if needed"); return api.deleteNode(args.node_id); })),

    tool("upload_asset", "Upload asset", "Upload a base64-encoded file asset to a draft file.", { file_id: uuidSchema, filename: nonEmptyString, mime_type: nonEmptyString.optional(), data_base64: nonEmptyString }, true,
      guarded(config, client, "upload_asset", true, z.object({ file_id: uuidSchema, filename: nonEmptyString, mime_type: nonEmptyString.optional(), data_base64: nonEmptyString }), (api, args) => api.uploadAsset({ fileId: args.file_id, filename: args.filename, mimeType: args.mime_type, dataBase64: args.data_base64 }))),
    tool("delete_asset", "Delete asset", "Delete a draft asset. Requires confirm=true.", { asset_id: uuidSchema, confirm: z.boolean() }, true,
      guarded(config, client, "delete_asset", true, z.object({ asset_id: uuidSchema, confirm: z.boolean() }), (api, args) => { requireConfirm(args, "delete_asset requires confirm=true after backup/export if needed"); return api.deleteAsset(args.asset_id); })),
    tool("list_assets", "List assets", "List draft and published asset state for a file.", { file_id: uuidSchema }, false,
      guarded(config, client, "list_assets", false, z.object({ file_id: uuidSchema }), (api, args) => api.listAssets(args.file_id))),

    tool("rebuild_search_index", "Rebuild search index", "Request backend search index rebuild. Requires confirm=true.", { confirm: z.boolean() }, true,
      guarded(config, client, "rebuild_search_index", true, z.object({ confirm: z.boolean() }), (api, args) => { requireConfirm(args, "rebuild_search_index requires confirm=true"); return api.rebuildSearchIndex(); })),
    tool("export_backup", "Export backup", "Export a local JSON backup of the Author content tree before destructive batches.", { output_dir: nonEmptyString }, false,
      guarded(config, client, "export_backup", false, z.object({ output_dir: nonEmptyString }), (api, args) => api.exportBackup({ outputDir: args.output_dir }))),
  ];
}
