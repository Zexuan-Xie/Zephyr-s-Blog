import { readdir } from "node:fs/promises";
import { spawnSync } from "node:child_process";

const files = (await readdir(new URL("../src/", import.meta.url)))
  .filter((name) => name.endsWith(".mjs"))
  .map((name) => `src/${name}`);

for (const file of files) {
  const result = spawnSync(process.execPath, ["--check", file], { stdio: "inherit" });
  if (result.status !== 0) process.exit(result.status ?? 1);
}
