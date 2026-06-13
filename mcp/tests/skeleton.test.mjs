import assert from 'node:assert/strict';
import { mkdtemp, readFile } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';

import { BlogBackendClient } from '../src/backendClient.js';
import { loadConfig } from '../src/config.js';
import { buildToolDefinitions } from '../src/tools.js';

async function runHealth(env) {
  const tmp = await mkdtemp(path.join(os.tmpdir(), 'xlab-blog-mcp-test-'));
  const auditLogPath = path.join(tmp, 'audit.jsonl');
  const config = loadConfig({ ...env, BLOG_MCP_AUDIT_LOG: auditLogPath });
  const client = new BlogBackendClient({ baseUrl: config.apiBaseUrl, adminToken: config.adminToken });
  const tool = buildToolDefinitions(config, client).find((item) => item.name === 'health_check');
  assert.ok(tool, 'health_check tool should be registered');
  const result = await tool.handler({ token: 'secret-token', note: 'hello' });
  const auditLines = (await readFile(auditLogPath, 'utf8')).trim().split('\n');
  return { result, audit: JSON.parse(auditLines.at(-1)) };
}

test('health_check refuses when BLOG_MCP_ENABLED is not true and writes audit JSONL', async () => {
  const { result, audit } = await runHealth({ BLOG_MCP_ENABLED: 'false' });
  assert.equal(result.isError, true);
  assert.match(result.content[0].text, /Blog MCP disabled/);
  assert.equal(audit.tool, 'health_check');
  assert.equal(audit.result, 'refused');
  assert.equal(audit.destructive, false);
  assert.equal(audit.args_summary.token, '[redacted]');
});

test('health_check refuses when per-call kill switch is active', async () => {
  const { result, audit } = await runHealth({ BLOG_MCP_ENABLED: 'true', BLOG_MCP_KILL_SWITCH: 'true' });
  assert.equal(result.isError, true);
  assert.match(result.content[0].text, /kill switch active/);
  assert.equal(audit.result, 'refused');
});

test('health_check succeeds when explicitly enabled and records ok audit event', async () => {
  const { result, audit } = await runHealth({ BLOG_MCP_ENABLED: 'true', BLOG_API_BASE_URL: 'http://127.0.0.1:8080/' });
  assert.equal(result.isError, undefined);
  assert.match(result.content[0].text, /127\.0\.0\.1:8080/);
  assert.equal(audit.result, 'ok');
  assert.equal(audit.tool, 'health_check');
});
