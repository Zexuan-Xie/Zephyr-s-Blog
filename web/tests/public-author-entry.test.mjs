import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

function source(path) {
  return readFileSync(new URL(path, import.meta.url), 'utf8');
}

const appSource = source('../src/App.tsx');
const rootPageSource = source('../src/pages/RootPage.tsx');
const directoryPageSource = source('../src/pages/DirectoryPage.tsx');
const resolverSource = source('../src/pages/ContentResolverPage.tsx');
const filePageSource = source('../src/components/FilePage.tsx');

test('Task 15 threads confirmed identity into public directory and file surfaces', () => {
  assert.match(appSource, /<RootPage currentUser=\{currentUser\}/);
  assert.match(appSource, /<ContentResolverPage currentUser=\{currentUser\}/);
  assert.match(rootPageSource, /currentUser: CurrentUser \| null/);
  assert.match(resolverSource, /currentUser: CurrentUser \| null/);
  assert.match(resolverSource, /<FilePage file=\{resolveQuery\.data\} currentUser=\{currentUser\}/);
  assert.match(resolverSource, /<DirectoryPage directory=\{resolveQuery\.data\} currentUser=\{currentUser\}/);
});

test('Task 15 directory manage entry is Author-only and targets Admin selection', () => {
  assert.match(directoryPageSource, /const isAuthor = currentUser\?\.role === ['"]admin['"]/);
  assert.match(directoryPageSource, /isAuthor \?/);
  assert.match(directoryPageSource, /管理此目录/);
  assert.match(directoryPageSource, /\/admin\?target=\$\{encodeURIComponent\(directory\.id\)\}/);
  assert.doesNotMatch(directoryPageSource, /currentUser\?\.role === ['"]reader['"][\s\S]{0,160}管理此目录/);
});

test('Task 15 file edit entry is Author-only and targets Admin selection', () => {
  assert.match(filePageSource, /const isAuthor = currentUser\?\.role === ['"]admin['"]/);
  assert.match(filePageSource, /isAuthor \?/);
  assert.match(filePageSource, /编辑文件/);
  assert.match(filePageSource, /\/admin\?target=\$\{encodeURIComponent\(file\.id\)\}/);
  assert.doesNotMatch(filePageSource, /currentUser\?\.role === ['"]reader['"][\s\S]{0,160}编辑文件/);
});
