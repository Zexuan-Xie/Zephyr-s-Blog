import { randomUUID } from "node:crypto";
import { link, mkdir, realpath, rm, writeFile } from "node:fs/promises";
import path from "node:path";

export interface BackupBackendReader {
  listContentTree(): Promise<any>;
  getFile(fileId: string): Promise<any>;
}

export interface BackupExportResult {
  file_path: string;
  node_count?: number;
  file_count: number;
}

/**
 * Server-local backup/export workflow for MCP maintenance tools.
 *
 * This service owns local filesystem policy. BlogBackendClient remains a thin
 * HTTP boundary and supplies only Author API reads used to build the payload.
 */
export class BackupExportService {
  readonly backupDir: string;

  constructor(backupDir: string) {
    this.backupDir = path.resolve(backupDir);
  }

  async exportBackup(client: BackupBackendReader, label?: string): Promise<BackupExportResult> {
    const outputDir = await resolveBackupOutputDir(this.backupDir, label);

    const tree = await client.listContentTree();
    const nodes = Array.isArray(tree?.nodes) ? tree.nodes : [];
    const fileNodes = nodes.filter((node: any) => node.kind === "file");
    const files = [];
    for (const node of fileNodes) {
      files.push({ node_id: node.id, path: node.url_path ?? node.path, version_state: await client.getFile(node.id) });
    }

    const exportedAt = new Date().toISOString();
    const timestamp = exportedAt.replaceAll(":", "-");
    const filePath = path.join(outputDir, `aeolian-backup-${timestamp}.json`);
    const tempPath = path.join(outputDir, `.aeolian-backup-${timestamp}-${randomUUID()}.tmp`);
    const backup = { exported_at: exportedAt, tree, files };
    const body = JSON.stringify(backup, null, 2);

    await writeFile(tempPath, body, { encoding: "utf8", flag: "wx" });
    try {
      await link(tempPath, filePath);
    } finally {
      await rm(tempPath, { force: true });
    }

    return { file_path: filePath, node_count: nodes.length, file_count: files.length };
  }
}

async function resolveBackupOutputDir(backupDir: string, label?: string): Promise<string> {
  const backupRoot = path.resolve(backupDir);
  const requestedLabel = label?.trim() || ".";
  if (path.isAbsolute(requestedLabel)) {
    throw new Error("export_backup label must be relative to BLOG_MCP_BACKUP_DIR");
  }
  if (requestedLabel !== ".") {
    const rawSegments = requestedLabel.split(/[\\/]+/);
    if (rawSegments.some((segment) => !segment || segment === "." || segment === "..")) {
      throw new Error("export_backup label contains an invalid path segment");
    }
  }

  const outputDir = path.resolve(backupRoot, requestedLabel);
  const relative = path.relative(backupRoot, outputDir);
  if (relative === ".." || relative.startsWith(`..${path.sep}`) || path.isAbsolute(relative)) {
    throw new Error("export_backup label must stay inside BLOG_MCP_BACKUP_DIR");
  }

  await mkdir(backupRoot, { recursive: true });
  const realRoot = await realpath(backupRoot);
  const segments = relative === "" ? [] : relative.split(path.sep);
  let current = realRoot;
  for (const segment of segments) {
    if (!segment || segment === "." || segment === "..") {
      throw new Error("export_backup label contains an invalid path segment");
    }
    const candidate = path.join(current, segment);
    try {
      await mkdir(candidate);
    } catch (error) {
      if (!isNodeErrorCode(error, "EEXIST")) throw error;
    }
    const realCandidate = await realpath(candidate);
    const realRelative = path.relative(realRoot, realCandidate);
    if (realRelative === ".." || realRelative.startsWith(`..${path.sep}`) || path.isAbsolute(realRelative)) {
      throw new Error("export_backup resolved path escapes BLOG_MCP_BACKUP_DIR");
    }
    current = realCandidate;
  }
  return current;
}

function isNodeErrorCode(error: unknown, code: string): boolean {
  return typeof error === "object" && error !== null && "code" in error && (error as { code?: unknown }).code === code;
}
