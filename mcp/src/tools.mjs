import { assertEnabled } from "./config.mjs";
import { summarizeArgs, writeAudit } from "./audit.mjs";

function textResult(payload) {
  return { content: [{ type: "text", text: typeof payload === "string" ? payload : JSON.stringify(payload, null, 2) }] };
}

function errorResult(error) {
  return { content: [{ type: "text", text: error instanceof Error ? error.message : String(error) }], isError: true };
}

export async function runGuardedTool(config, tool, destructive, args, operation) {
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

export function buildToolDefinitions(config, client) {
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
