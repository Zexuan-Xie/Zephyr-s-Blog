import assert from 'node:assert/strict';
import { mkdir, mkdtemp, readFile, symlink } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';

import { BlogBackendClient } from '../src/backendClient.ts';
import { loadConfig } from '../src/config.ts';
import { buildToolDefinitions } from '../src/tools.ts';

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


test('registers the full Stage 3 MCP tool surface', async () => {
  const config = loadConfig({ BLOG_MCP_ENABLED: 'false' });
  const client = new BlogBackendClient({ baseUrl: config.apiBaseUrl });
  const names = buildToolDefinitions(config, client).map((tool) => tool.name).sort();
  assert.deepEqual(names, [
    'create_directory',
    'create_file',
    'delete_asset',
    'delete_node',
    'export_backup',
    'get_file',
    'health_check',
    'list_assets',
    'list_content_tree',
    'move_node',
    'publish_file',
    'rebuild_search_index',
    'reorder_children',
    'search_files',
    'unpublish_file',
    'update_file_content',
    'update_file_settings',
    'upload_asset',
  ]);
});

test('destructive tools require explicit enablement before input validation or mutation', async () => {
  const tmp = await mkdtemp(path.join(os.tmpdir(), 'xlab-blog-mcp-test-'));
  const auditLogPath = path.join(tmp, 'audit.jsonl');
  const config = loadConfig({ BLOG_MCP_ENABLED: 'false', BLOG_MCP_AUDIT_LOG: auditLogPath });
  const client = new BlogBackendClient({ baseUrl: config.apiBaseUrl });
  const tool = buildToolDefinitions(config, client).find((item) => item.name === 'delete_node');
  const result = await tool.handler({ node_id: 'not-a-uuid', confirm: true });
  const audit = JSON.parse((await readFile(auditLogPath, 'utf8')).trim().split('\n').at(-1));
  assert.equal(result.isError, true);
  assert.match(result.content[0].text, /Blog MCP disabled/);
  assert.equal(audit.tool, 'delete_node');
  assert.equal(audit.result, 'refused');
  assert.equal(audit.destructive, true);
});

test('enabled content tool calls backend boundary with auth and audits ok', async () => {
  const tmp = await mkdtemp(path.join(os.tmpdir(), 'xlab-blog-mcp-test-'));
  const auditLogPath = path.join(tmp, 'audit.jsonl');
  const calls = [];
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async (url, init = {}) => {
    calls.push({ url: String(url), init });
    return new Response(JSON.stringify({ node: { id: 'created' } }), {
      status: 201,
      headers: { 'Content-Type': 'application/json' },
    });
  };
  try {
    const config = loadConfig({
      BLOG_MCP_ENABLED: 'true',
      BLOG_MCP_AUDIT_LOG: auditLogPath,
      BLOG_API_BASE_URL: 'http://backend.local/',
      BLOG_ADMIN_TOKEN: 'admin-secret',
    });
    const client = new BlogBackendClient({ baseUrl: config.apiBaseUrl, adminToken: config.adminToken });
    const tool = buildToolDefinitions(config, client).find((item) => item.name === 'create_file');
    const result = await tool.handler({ name: 'Draft', content_format: 'markdown' });
    const audit = JSON.parse((await readFile(auditLogPath, 'utf8')).trim().split('\n').at(-1));

    assert.equal(result.isError, undefined);
    assert.equal(calls.length, 1);
    assert.equal(calls[0].url, 'http://backend.local/api/admin/nodes');
    assert.equal(calls[0].init.method, 'POST');
    assert.equal(calls[0].init.headers.Authorization, 'Bearer admin-secret');
    assert.deepEqual(JSON.parse(calls[0].init.body), {
      parent_id: null,
      kind: 'file',
      name: 'Draft',
      content_format: 'markdown',
      sort_order: 0,
    });
    assert.equal(audit.tool, 'create_file');
    assert.equal(audit.result, 'ok');
    assert.equal(audit.destructive, true);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test('enabled destructive delete refuses without confirm before backend call and audits error', async () => {
  const tmp = await mkdtemp(path.join(os.tmpdir(), 'xlab-blog-mcp-test-'));
  const auditLogPath = path.join(tmp, 'audit.jsonl');
  const originalFetch = globalThis.fetch;
  let called = false;
  globalThis.fetch = async () => {
    called = true;
    return new Response('{}', { status: 200, headers: { 'Content-Type': 'application/json' } });
  };
  try {
    const config = loadConfig({ BLOG_MCP_ENABLED: 'true', BLOG_MCP_AUDIT_LOG: auditLogPath });
    const client = new BlogBackendClient({ baseUrl: config.apiBaseUrl });
    const tool = buildToolDefinitions(config, client).find((item) => item.name === 'delete_asset');
    const result = await tool.handler({ asset_id: '00000000-0000-4000-8000-000000000000', confirm: false });
    const audit = JSON.parse((await readFile(auditLogPath, 'utf8')).trim().split('\n').at(-1));

    assert.equal(result.isError, true);
    assert.equal(called, false);
    assert.match(result.content[0].text, /requires confirm=true/);
    assert.equal(audit.tool, 'delete_asset');
    assert.equal(audit.result, 'error');
    assert.equal(audit.destructive, true);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test('export_backup writes only under BLOG_MCP_BACKUP_DIR and audits ok', async () => {
  const tmp = await mkdtemp(path.join(os.tmpdir(), 'xlab-blog-mcp-test-'));
  const auditLogPath = path.join(tmp, 'audit.jsonl');
  const backupDir = path.join(tmp, 'backups');
  const calls = [];
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async (url, init = {}) => {
    calls.push({ url: String(url), init });
    if (String(url).endsWith('/api/admin/tree')) {
      return new Response(JSON.stringify({ nodes: [{ id: '00000000-0000-4000-8000-000000000001', kind: 'file', url_path: '/safe' }] }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      });
    }
    if (String(url).includes('/api/admin/files/00000000-0000-4000-8000-000000000001/content')) {
      return new Response(JSON.stringify({ current: { body_raw: 'draft' }, revision: 1 }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      });
    }
    return new Response(JSON.stringify({ error: 'unexpected url' }), { status: 500, headers: { 'Content-Type': 'application/json' } });
  };
  try {
    const config = loadConfig({ BLOG_MCP_ENABLED: 'true', BLOG_MCP_AUDIT_LOG: auditLogPath, BLOG_MCP_BACKUP_DIR: backupDir });
    const client = new BlogBackendClient({ baseUrl: config.apiBaseUrl });
    const tool = buildToolDefinitions(config, client).find((item) => item.name === 'export_backup');
    const result = await tool.handler({ label: 'before-delete' });
    const payload = JSON.parse(result.content[0].text);
    const audit = JSON.parse((await readFile(auditLogPath, 'utf8')).trim().split('\n').at(-1));
    const backupText = await readFile(payload.file_path, 'utf8');

    assert.equal(result.isError, undefined);
    assert.equal(path.relative(backupDir, payload.file_path).startsWith('..'), false);
    assert.match(payload.file_path, /before-delete/);
    assert.match(backupText, /"exported_at"/);
    assert.equal(calls.length, 2);
    assert.equal(audit.tool, 'export_backup');
    assert.equal(audit.result, 'ok');
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test('export_backup rejects traversal label before backend calls or writes', async () => {
  const tmp = await mkdtemp(path.join(os.tmpdir(), 'xlab-blog-mcp-test-'));
  const auditLogPath = path.join(tmp, 'audit.jsonl');
  const backupDir = path.join(tmp, 'backups');
  const originalFetch = globalThis.fetch;
  let called = false;
  globalThis.fetch = async () => {
    called = true;
    return new Response('{}', { status: 200, headers: { 'Content-Type': 'application/json' } });
  };
  try {
    const config = loadConfig({ BLOG_MCP_ENABLED: 'true', BLOG_MCP_AUDIT_LOG: auditLogPath, BLOG_MCP_BACKUP_DIR: backupDir });
    const client = new BlogBackendClient({ baseUrl: config.apiBaseUrl });
    const tool = buildToolDefinitions(config, client).find((item) => item.name === 'export_backup');
    const result = await tool.handler({ label: '../escape' });
    const audit = JSON.parse((await readFile(auditLogPath, 'utf8')).trim().split('\n').at(-1));

    assert.equal(result.isError, true);
    assert.equal(called, false);
    assert.match(result.content[0].text, /invalid path segment|stay inside BLOG_MCP_BACKUP_DIR/);
    assert.equal(audit.tool, 'export_backup');
    assert.equal(audit.result, 'error');
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test('export_backup rejects absolute label before backend calls or writes', async () => {
  const tmp = await mkdtemp(path.join(os.tmpdir(), 'xlab-blog-mcp-test-'));
  const auditLogPath = path.join(tmp, 'audit.jsonl');
  const backupDir = path.join(tmp, 'backups');
  const originalFetch = globalThis.fetch;
  let called = false;
  globalThis.fetch = async () => {
    called = true;
    return new Response('{}', { status: 200, headers: { 'Content-Type': 'application/json' } });
  };
  try {
    const config = loadConfig({ BLOG_MCP_ENABLED: 'true', BLOG_MCP_AUDIT_LOG: auditLogPath, BLOG_MCP_BACKUP_DIR: backupDir });
    const client = new BlogBackendClient({ baseUrl: config.apiBaseUrl });
    const tool = buildToolDefinitions(config, client).find((item) => item.name === 'export_backup');
    const result = await tool.handler({ label: path.join(os.tmpdir(), 'escape') });
    const audit = JSON.parse((await readFile(auditLogPath, 'utf8')).trim().split('\n').at(-1));

    assert.equal(result.isError, true);
    assert.equal(called, false);
    assert.match(result.content[0].text, /must be relative/);
    assert.equal(audit.tool, 'export_backup');
    assert.equal(audit.result, 'error');
  } finally {
    globalThis.fetch = originalFetch;
  }
});


test('export_backup rejects symlink labels that resolve outside backup root', async () => {
  const tmp = await mkdtemp(path.join(os.tmpdir(), 'xlab-blog-mcp-test-'));
  const auditLogPath = path.join(tmp, 'audit.jsonl');
  const backupDir = path.join(tmp, 'backups');
  const outsideDir = path.join(tmp, 'outside');
  await mkdir(backupDir, { recursive: true });
  await mkdir(outsideDir, { recursive: true });
  await symlink(outsideDir, path.join(backupDir, 'linked-outside'), 'dir');

  const originalFetch = globalThis.fetch;
  let called = false;
  globalThis.fetch = async () => {
    called = true;
    return new Response('{}', { status: 200, headers: { 'Content-Type': 'application/json' } });
  };
  try {
    const config = loadConfig({ BLOG_MCP_ENABLED: 'true', BLOG_MCP_AUDIT_LOG: auditLogPath, BLOG_MCP_BACKUP_DIR: backupDir });
    const client = new BlogBackendClient({ baseUrl: config.apiBaseUrl });
    const tool = buildToolDefinitions(config, client).find((item) => item.name === 'export_backup');
    const result = await tool.handler({ label: 'linked-outside' });
    const audit = JSON.parse((await readFile(auditLogPath, 'utf8')).trim().split('\n').at(-1));

    assert.equal(result.isError, true);
    assert.equal(called, false);
    assert.match(result.content[0].text, /escapes BLOG_MCP_BACKUP_DIR/);
    assert.equal(audit.tool, 'export_backup');
    assert.equal(audit.result, 'error');
  } finally {
    globalThis.fetch = originalFetch;
  }
});
