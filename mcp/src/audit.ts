import { mkdir, appendFile } from "node:fs/promises";
import path from "node:path";

export interface AuditEvent {
  timestamp: string;
  tool: string;
  destructive: boolean;
  args_summary: Record<string, unknown>;
  result: "started" | "ok" | "error" | "refused";
  message?: string;
}

export function summarizeArgs(args: Record<string, unknown> | undefined): Record<string, unknown> {
  const summary: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(args ?? {})) {
    if (/token|password|secret|key/i.test(key)) {
      summary[key] = "[redacted]";
    } else if (typeof value === "string" && value.length > 160) {
      summary[key] = `${value.slice(0, 157)}...`;
    } else {
      summary[key] = value;
    }
  }
  return summary;
}

export async function writeAudit(logPath: string, event: Omit<AuditEvent, "timestamp">): Promise<void> {
  await mkdir(path.dirname(logPath), { recursive: true });
  const line: AuditEvent = { timestamp: new Date().toISOString(), ...event };
  await appendFile(logPath, `${JSON.stringify(line)}\n`, { encoding: "utf8" });
}
