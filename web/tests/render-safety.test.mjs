import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const filePageSource = readFileSync(new URL('../src/components/FilePage.tsx', import.meta.url), 'utf8');
const renderMarkdownSource = readFileSync(new URL('../src/lib/renderMarkdown.ts', import.meta.url), 'utf8');
const prohibitedSandboxToken = ['allow', 'same-origin'].join('-');

test('HTML Documents stay isolated behind the exact iframe sandbox', () => {
  assert.match(filePageSource, /<iframe[\s\S]*?sandbox="allow-scripts"[\s\S]*?srcDoc=/);
  assert.doesNotMatch(filePageSource, new RegExp(prohibitedSandboxToken));
  assert.equal(filePageSource.match(/\bsandbox=/g)?.length, 1);
  assert.equal(filePageSource.match(/\bsrcDoc=/g)?.length, 1);
});

test('only the Markdown branch injects sanitized HTML into the main DOM', () => {
  const markdownBranch = filePageSource.indexOf("file.content_format === 'markdown'");
  const htmlDocumentBranch = filePageSource.indexOf('<iframe');
  const htmlInjection = filePageSource.indexOf('dangerouslySetInnerHTML');

  assert.ok(markdownBranch >= 0);
  assert.ok(htmlInjection > markdownBranch);
  assert.ok(htmlDocumentBranch > htmlInjection);
  assert.equal(filePageSource.match(/dangerouslySetInnerHTML/g)?.length, 1);
  assert.match(filePageSource, /sanitizeServerHtml\(file\.body_html\)/);
  assert.match(filePageSource, /renderSafeMarkdown\(file\.body_markdown \?\? ''\)/);
});

test('both Markdown render paths use the hardened DOMPurify config', () => {
  assert.match(renderMarkdownSource, /marked\.parse\(markdown,/);
  assert.equal(
    renderMarkdownSource.match(/DOMPurify\.sanitize\([^,]+, markdownSanitizeConfig\)/g)?.length,
    2,
  );

  for (const forbiddenTag of ['script', 'style', 'iframe', 'object', 'embed', 'form']) {
    assert.match(renderMarkdownSource, new RegExp(`['"]${forbiddenTag}['"]`));
  }
});

test('Packet F reader interactions use API endpoints and login return target', () => {
  const apiSource = readFileSync(new URL('../src/lib/api.ts', import.meta.url), 'utf8');

  assert.match(filePageSource, /fetchCommentThread\(file\.id\)/);
  assert.match(filePageSource, /createComment\(file\.id, commentBody, replyTarget\?\.parentId, replyTarget\?\.replyToUserId\)/);
  assert.match(filePageSource, /navigate\(`\/login\?return_to=\$\{encodeURIComponent\(file\.path\)\}`\)/);
  assert.match(apiSource, /\/files\/\$\{encodeURIComponent\(fileId\)\}\/comments/);
  assert.match(apiSource, /\/files\/\$\{encodeURIComponent\(fileId\)\}\/like/);
  assert.match(apiSource, /\/comments\/\$\{encodeURIComponent\(commentId\)\}\/like/);
  assert.match(apiSource, /Authorization: `Bearer \$\{token\}`/);
});

test('Packet G assets keep public URLs and admin upload/delete API helpers for later workspace packets', () => {
  const apiSource = readFileSync(new URL('../src/lib/api.ts', import.meta.url), 'utf8');

  assert.match(filePageSource, /file\.assets\.map/);
  assert.match(filePageSource, /href=\{asset\.public_url\}/);
  assert.match(apiSource, /\/admin\/files\/\$\{encodeURIComponent\(fileId\)\}\/assets/);
  assert.match(apiSource, /\/admin\/assets\/\$\{encodeURIComponent\(assetId\)\}/);
  assert.match(apiSource, /FormData\(\)/);
});


test('Packet H search page calls search API and renders source badges', () => {
  const apiSource = readFileSync(new URL('../src/lib/api.ts', import.meta.url), 'utf8');
  const searchPageSource = readFileSync(new URL('../src/pages/SearchPage.tsx', import.meta.url), 'utf8');

  assert.match(apiSource, /\/search\?q=\$\{encodeURIComponent\(query\)\}/);
  assert.match(apiSource, /match_sources/);
  assert.match(apiSource, /semantic/);
  assert.match(searchPageSource, /searchFiles\(query\)/);
  assert.match(searchPageSource, /result\.path/);
  assert.match(searchPageSource, /result\.snippet/);
  assert.match(searchPageSource, /result\.sources\.map/);
});


test('Stage 2 Author Workspace shell wires protected tree and keeps HTML sandbox isolated to readers', () => {
  const apiSource = readFileSync(new URL('../src/lib/api.ts', import.meta.url), 'utf8');
  const adminPageSource = readFileSync(new URL('../src/pages/AdminPage.tsx', import.meta.url), 'utf8');

  assert.match(apiSource, /fetchAdminTree/);
  assert.match(apiSource, /\/admin\/tree/);
  assert.match(apiSource, /fetchAdminNode/);
  assert.match(apiSource, /fetchCurrentUser/);
  assert.match(adminPageSource, /Author Workspace/);
  assert.match(adminPageSource, /Content Tree/);
  assert.match(adminPageSource, /Navigate to="\/login\?return_to=%2Fadmin"/);
  assert.match(adminPageSource, /Draft/);
  assert.match(adminPageSource, /Live/);
  assert.doesNotMatch(adminPageSource, /sandbox="allow-scripts"/);
});
