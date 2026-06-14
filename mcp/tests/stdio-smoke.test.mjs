import assert from 'node:assert/strict';
import { mkdir, mkdtemp, readFile, symlink } from 'node:fs/promises';
import { createServer } from 'node:http';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';

import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js';


async function withBackendServer() {
  const calls = [];
  const server = createServer((req, res) => {
    const chunks = [];
    req.on('data', (chunk) => chunks.push(chunk));
    req.on('end', () => {
      const body = Buffer.concat(chunks).toString('utf8');
      calls.push({ method: req.method, url: req.url, body, authorization: req.headers.authorization });
      res.setHeader('Content-Type', 'application/json');
      if (req.url === '/api/admin/tree') {
        res.statusCode = 200;
        res.end(JSON.stringify({ nodes: [{ id: '00000000-0000-4000-8000-000000000001', kind: 'file', url_path: '/safe' }] }));
        return;
      }
      if (req.url === '/api/admin/files/00000000-0000-4000-8000-000000000001/content') {
        res.statusCode = req.method === 'PUT' ? 200 : 200;
        res.end(JSON.stringify({ current: { body_raw: 'draft' }, revision: 2 }));
        return;
      }
      if (req.url?.startsWith('/api/search?')) {
        res.statusCode = 200;
        res.end(JSON.stringify({ results: [] }));
        return;
      }
      if (req.url === '/api/admin/nodes' && req.method === 'POST') {
        res.statusCode = 201;
        res.end(JSON.stringify({ node: { id: '00000000-0000-4000-8000-000000000002' } }));
        return;
      }
      if (req.url === '/api/admin/nodes/00000000-0000-4000-8000-000000000001') {
        res.statusCode = req.method === 'DELETE' ? 204 : 200;
        res.end(req.method === 'DELETE' ? '' : JSON.stringify({ node: { id: '00000000-0000-4000-8000-000000000001' } }));
        return;
      }
      if (req.url === '/api/admin/files/00000000-0000-4000-8000-000000000001/publish' || req.url === '/api/admin/files/00000000-0000-4000-8000-000000000001/unpublish') {
        res.statusCode = 200;
        res.end(JSON.stringify({ current: { revision: 3 } }));
        return;
      }
      if (req.url === '/api/admin/nodes/00000000-0000-4000-8000-000000000001/move' || req.url === '/api/admin/nodes/00000000-0000-4000-8000-000000000001/children/order') {
        res.statusCode = 200;
        res.end(JSON.stringify({ ok: true }));
        return;
      }
      if (req.url === '/api/admin/files/00000000-0000-4000-8000-000000000001/assets') {
        res.statusCode = req.method === 'POST' ? 201 : 200;
        res.end(JSON.stringify({ assets: [], asset: { id: '00000000-0000-4000-8000-000000000003' } }));
        return;
      }
      if (req.url === '/api/admin/assets/00000000-0000-4000-8000-000000000003') {
        res.statusCode = 204;
        res.end('');
        return;
      }
      if (req.url === '/api/admin/search-index/rebuild') {
        res.statusCode = 204;
        res.end('');
        return;
      }
      res.statusCode = 404;
      res.end(JSON.stringify({ error: `unexpected ${req.method} ${req.url}` }));
    });
  });
  await new Promise((resolve, reject) => {
    server.on('error', reject);
    server.listen(0, '127.0.0.1', resolve);
  });
  const address = server.address();
  assert.ok(address && typeof address === 'object');
  return {
    calls,
    baseUrl: `http://127.0.0.1:${address.port}`,
    close: () => new Promise((resolve, reject) => server.close((error) => error ? reject(error) : resolve())),
  };
}

async function withClient(env, fn) {
  const client = new Client({ name: 'aeolian-mcp-smoke', version: '0.1.0' }, { capabilities: {} });
  const transport = new StdioClientTransport({
    command: process.execPath,
    args: ['--import', 'tsx', 'src/server.ts'],
    cwd: path.resolve(new URL('..', import.meta.url).pathname),
    env: { ...process.env, ...env },
    stderr: 'pipe',
  });
  await client.connect(transport);
  try {
    return await fn(client, transport);
  } finally {
    await client.close();
  }
}

test('stdio smoke: disabled server lists tools and refuses before backend calls', async () => {
  const tmp = await mkdtemp(path.join(os.tmpdir(), 'xlab-blog-mcp-stdio-'));
  const auditLogPath = path.join(tmp, 'audit.jsonl');
  await withClient({ BLOG_MCP_ENABLED: 'false', BLOG_MCP_AUDIT_LOG: auditLogPath }, async (client) => {
    const tools = await client.listTools();
    assert.ok(tools.tools.some((tool) => tool.name === 'health_check'));
    const result = await client.callTool({ name: 'health_check', arguments: { token: 'secret-token' } });
    assert.equal(result.isError, true);
    assert.match(result.content[0].text, /Blog MCP disabled/);
  });
  const audit = JSON.parse((await readFile(auditLogPath, 'utf8')).trim().split('\n').at(-1));
  assert.equal(audit.tool, 'health_check');
  assert.equal(audit.result, 'refused');
});

test('stdio smoke: kill switch refuses an enabled call', async () => {
  const tmp = await mkdtemp(path.join(os.tmpdir(), 'xlab-blog-mcp-stdio-'));
  const auditLogPath = path.join(tmp, 'audit.jsonl');
  await withClient({ BLOG_MCP_ENABLED: 'true', BLOG_MCP_KILL_SWITCH: 'true', BLOG_MCP_AUDIT_LOG: auditLogPath }, async (client) => {
    const result = await client.callTool({ name: 'health_check', arguments: {} });
    assert.equal(result.isError, true);
    assert.match(result.content[0].text, /kill switch active/);
  });
  const audit = JSON.parse((await readFile(auditLogPath, 'utf8')).trim().split('\n').at(-1));
  assert.equal(audit.result, 'refused');
});

test('stdio smoke: export_backup safe and rejected labels are audited', async () => {
  const tmp = await mkdtemp(path.join(os.tmpdir(), 'xlab-blog-mcp-stdio-'));
  const auditLogPath = path.join(tmp, 'audit.jsonl');
  const backupDir = path.join(tmp, 'backups');
  const outsideDir = path.join(tmp, 'outside');
  await mkdir(backupDir, { recursive: true });
  await mkdir(outsideDir, { recursive: true });
  await symlink(outsideDir, path.join(backupDir, 'linked-outside'), 'dir');

  const backend = await withBackendServer();
  try {
    await withClient({ BLOG_MCP_ENABLED: 'true', BLOG_MCP_AUDIT_LOG: auditLogPath, BLOG_MCP_BACKUP_DIR: backupDir, BLOG_API_BASE_URL: backend.baseUrl }, async (client) => {
    const safeResult = await client.callTool({ name: 'export_backup', arguments: { label: 'acceptance' } });
    assert.equal(safeResult.isError, undefined);
    const payload = JSON.parse(safeResult.content[0].text);
    assert.equal(path.relative(backupDir, payload.file_path).startsWith('..'), false);
    await readFile(payload.file_path, 'utf8');

    const traversalResult = await client.callTool({ name: 'export_backup', arguments: { label: '../escape' } });
    assert.equal(traversalResult.isError, true);
    assert.match(traversalResult.content[0].text, /invalid path segment|stay inside BLOG_MCP_BACKUP_DIR/);

    const absoluteResult = await client.callTool({ name: 'export_backup', arguments: { label: path.join(os.tmpdir(), 'escape') } });
    assert.equal(absoluteResult.isError, true);
    assert.match(absoluteResult.content[0].text, /must be relative/);

    const symlinkResult = await client.callTool({ name: 'export_backup', arguments: { label: 'linked-outside' } });
    assert.equal(symlinkResult.isError, true);
    assert.match(symlinkResult.content[0].text, /escapes BLOG_MCP_BACKUP_DIR/);
    });

    assert.equal(backend.calls.length, 2);
  } finally {
    await backend.close();
  }

  const auditLines = (await readFile(auditLogPath, 'utf8')).trim().split('\n').map((line) => JSON.parse(line));
  assert.deepEqual(auditLines.map((line) => [line.tool, line.result]), [
    ['export_backup', 'ok'],
    ['export_backup', 'error'],
    ['export_backup', 'error'],
    ['export_backup', 'error'],
  ]);
});


test('stdio smoke: all required tools operate through the backend boundary and audit ok', async () => {
  const tmp = await mkdtemp(path.join(os.tmpdir(), 'xlab-blog-mcp-stdio-'));
  const auditLogPath = path.join(tmp, 'audit.jsonl');
  const backupDir = path.join(tmp, 'backups');
  const backend = await withBackendServer();
  const id = '00000000-0000-4000-8000-000000000001';
  const assetId = '00000000-0000-4000-8000-000000000003';
  try {
    await withClient({
      BLOG_MCP_ENABLED: 'true',
      BLOG_MCP_AUDIT_LOG: auditLogPath,
      BLOG_MCP_BACKUP_DIR: backupDir,
      BLOG_API_BASE_URL: backend.baseUrl,
      BLOG_ADMIN_TOKEN: 'admin-secret',
    }, async (client) => {
      const calls = [
        ['list_content_tree', {}],
        ['get_file', { file_id: id }],
        ['search_files', { query: 'safe', limit: 5 }],
        ['create_directory', { name: 'Dir' }],
        ['create_file', { name: 'File', content_format: 'markdown' }],
        ['update_file_content', { file_id: id, expected_revision: 1, content_format: 'markdown', body_raw: 'draft' }],
        ['update_file_settings', { node_id: id, name: 'Renamed' }],
        ['publish_file', { file_id: id, expected_revision: 2 }],
        ['unpublish_file', { file_id: id, expected_revision: 3 }],
        ['move_node', { node_id: id, expected_version: 1 }],
        ['reorder_children', { parent_id: id, child_ids: [id], expected_version: 1 }],
        ['delete_node', { node_id: id, confirm: true }],
        ['upload_asset', { file_id: id, filename: 'note.txt', mime_type: 'text/plain', data_base64: Buffer.from('asset').toString('base64') }],
        ['list_assets', { file_id: id }],
        ['delete_asset', { asset_id: assetId, confirm: true }],
        ['rebuild_search_index', { confirm: true }],
      ];
      for (const [name, args] of calls) {
        const result = await client.callTool({ name, arguments: args });
        assert.equal(result.isError, undefined, `${name} should succeed: ${result.content?.[0]?.text}`);
      }
    });

    assert.ok(backend.calls.length >= 16);
    assert.ok(backend.calls.every((call) => call.authorization === 'Bearer admin-secret'));
  } finally {
    await backend.close();
  }

  const auditLines = (await readFile(auditLogPath, 'utf8')).trim().split('\n').map((line) => JSON.parse(line));
  const okTools = new Set(auditLines.filter((line) => line.result === 'ok').map((line) => line.tool));
  for (const tool of ['list_content_tree', 'get_file', 'search_files', 'create_directory', 'create_file', 'update_file_content', 'update_file_settings', 'publish_file', 'unpublish_file', 'move_node', 'reorder_children', 'delete_node', 'upload_asset', 'list_assets', 'delete_asset', 'rebuild_search_index']) {
    assert.ok(okTools.has(tool), `${tool} should have an ok audit entry`);
  }
});
