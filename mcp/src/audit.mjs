import { mkdir, appendFile } from "node:fs/promises";
import path from "node:path";

export function summarizeArgs(args) {
  const summary = {};
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

export async function writeAudit(logPath, event) {
  await mkdir(path.dirname(logPath), { recursive: true });
  const line = { timestamp: new Date().toISOString(), ...event };
  await appendFile(logPath, `${JSON.stringify(line)}\n`, { encoding: "utf8" });
}
