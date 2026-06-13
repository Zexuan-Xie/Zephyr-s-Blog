import os from "node:os";
import path from "node:path";

function parseBool(value) {
  return value === "true" || value === "1" || value === "yes";
}

export function loadConfig(env = process.env) {
  return {
    enabled: parseBool(env.BLOG_MCP_ENABLED),
    killSwitch: parseBool(env.BLOG_MCP_KILL_SWITCH),
    auditLogPath:
      env.BLOG_MCP_AUDIT_LOG ?? path.join(os.homedir(), ".local", "share", "xlab-blog", "mcp-audit.jsonl"),
    apiBaseUrl: env.BLOG_API_BASE_URL ?? "http://127.0.0.1:8080",
    adminToken: env.BLOG_ADMIN_TOKEN,
  };
}

export function assertEnabled(config) {
  if (!config.enabled) {
    throw new Error("Blog MCP disabled: set BLOG_MCP_ENABLED=true to allow tool calls");
  }
  if (config.killSwitch) {
    throw new Error("Blog MCP kill switch active: BLOG_MCP_KILL_SWITCH=true");
  }
}
