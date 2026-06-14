import os from "node:os";
import path from "node:path";

export interface BlogMcpConfig {
  enabled: boolean;
  killSwitch: boolean;
  auditLogPath: string;
  apiBaseUrl: string;
  adminToken?: string;
  backupDir: string;
}

function parseBool(value: string | undefined): boolean {
  return value === "true" || value === "1" || value === "yes";
}

export function loadConfig(env: NodeJS.ProcessEnv = process.env): BlogMcpConfig {
  return {
    enabled: parseBool(env.BLOG_MCP_ENABLED),
    killSwitch: parseBool(env.BLOG_MCP_KILL_SWITCH),
    auditLogPath:
      env.BLOG_MCP_AUDIT_LOG ?? path.join(os.homedir(), ".local", "share", "xlab-blog", "mcp-audit.jsonl"),
    apiBaseUrl: env.BLOG_API_BASE_URL ?? "http://127.0.0.1:8080",
    adminToken: env.BLOG_ADMIN_TOKEN,
    backupDir: path.resolve(
      env.BLOG_MCP_BACKUP_DIR ?? path.join(os.homedir(), ".local", "share", "xlab-blog", "mcp-backups"),
    ),
  };
}

export function assertEnabled(config: Pick<BlogMcpConfig, "enabled" | "killSwitch">): void {
  if (!config.enabled) {
    throw new Error("Blog MCP disabled: set BLOG_MCP_ENABLED=true to allow tool calls");
  }
  if (config.killSwitch) {
    throw new Error("Blog MCP kill switch active: BLOG_MCP_KILL_SWITCH=true");
  }
}
