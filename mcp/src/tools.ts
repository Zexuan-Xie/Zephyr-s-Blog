import { z } from "zod";
import type { BlogMcpConfig } from "./config.js";
import { assertEnabled } from "./config.js";
import { summarizeArgs, writeAudit } from "./audit.js";
import type { BlogBackendClient } from "./backendClient.js";

export interface ToolResultContent {
  type: "text";
  text: string;
}

export interface ToolResult {
  content: ToolResultContent[];
  isError?: boolean;
}

export interface ToolDefinition {
  name: string;
  title: string;
  description: string;
  inputSchema: Record<string, z.ZodTypeAny>;
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
    {
      name: "health_check",
      title: "Blog MCP health check",
      description:
        "Non-destructive skeleton tool proving enable/kill-switch guard, audit JSONL, and backend API-client boundary.",
      inputSchema: {},
      destructive: false,
      handler: (args) =>
        runGuardedTool(config, "health_check", false, args, async () => textResult(await client.health())),
    },
  ];
}
