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
