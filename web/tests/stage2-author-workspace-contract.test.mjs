import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

function source(path) {
  return readFileSync(new URL(path, import.meta.url), 'utf8');
}

const adminPageSource = source('../src/pages/AdminPage.tsx');
const apiSource = source('../src/lib/api.ts');
const typesSource = source('../src/lib/types.ts');
const rootPageSource = source('../src/pages/RootPage.tsx');
const directoryPageSource = source('../src/pages/DirectoryPage.tsx');
const contentResolverSource = source('../src/pages/ContentResolverPage.tsx');
const filePageSource = source('../src/components/FilePage.tsx');

test('Stage 2 replaces old Admin chrome with Chinese Author Workspace and protected Content Tree language', () => {
  assert.doesNotMatch(adminPageSource, />ADMIN</);
  assert.doesNotMatch(adminPageSource, /Tree Manager/);
  assert.doesNotMatch(adminPageSource, /Tree browser/);
  assert.doesNotMatch(adminPageSource, /Node id/);
  assert.doesNotMatch(adminPageSource, /Load selected node/);
  assert.match(adminPageSource, /作者工作台|Author Workspace/);
  assert.match(adminPageSource, /内容树|Content Tree/);
  assert.match(adminPageSource, /目录概览/);
});

test('Stage 2 minimal create flow is slugless and opens the returned node after tree refresh', () => {
  const createStart = adminPageSource.indexOf('async function submitCreate');
  const logoutStart = adminPageSource.indexOf('function logoutAuthor', createStart);
  const createFlow = adminPageSource.slice(createStart, logoutStart);

  assert.match(apiSource, /fetchAdminTree/);
  assert.match(apiSource, /requestJson\(['"]\/admin\/tree['"]/);
  assert.match(apiSource, /createAdminNode\(input: CreateAdminNodeInput\): Promise<AdminNodeDetail>/);
  assert.doesNotMatch(apiSource, /CreateAdminNodeInput[\s\S]*slug:/);
  assert.doesNotMatch(createFlow, /slug:/);
  assert.doesNotMatch(createFlow, /sort_order:/);
  assert.match(createFlow, /setSelectedId\(created\.node\.id\)/);
  assert.match(adminPageSource, /queryFn: \(\) => fetchAdminNode\(effectiveSelectedId\)/);
  assert.match(adminPageSource, /detail=\{detailQuery\.data \?\? null\}/);
  assert.match(createFlow, /adminTreeQuery\.refetch\(\)/);
  assert.match(createFlow, /created\.node\.path|created\.node\.url_path/);
  assert.match(adminPageSource, /新建目录/);
  assert.match(adminPageSource, /新建文件/);
  assert.match(adminPageSource, /URL Path/);
});

test('Stage 2 primary UI hides implementation identifiers and Stage 3 publication concepts', () => {
  for (const forbidden of [
    /Parent id/i,
    /Node id/i,
    /Sort order/i,
    /\bslug\b/i,
    /有未发布修改/,
    /发布更新/,
    /Draft Preview/i,
    /Autosave/i,
  ]) {
    assert.doesNotMatch(adminPageSource, forbidden);
  }

  assert.match(adminPageSource, /草稿/);
  assert.match(adminPageSource, /已发布/);
  assert.match(adminPageSource, /撤回发布/);
});

test('Stage 2 workspace provides explicit return controls for subflows', () => {
  assert.match(adminPageSource, /返回目录/);
  assert.match(adminPageSource, /返回内容树/);
  assert.match(adminPageSource, /取消/);
});

test('Stage 2 public Author entries are role-gated and open admin with selected target', () => {
  const publicSources = [rootPageSource, directoryPageSource, contentResolverSource, filePageSource].join('\n');

  assert.match(publicSources, /currentUser\?\.role === ['"]admin['"]|isAuthor/);
  assert.match(publicSources, /管理此目录/);
  assert.match(publicSources, /编辑文件/);
  assert.match(publicSources, /\/admin\?[^`'"]*(target|node|select)=/);
  assert.doesNotMatch(publicSources, /currentUser\?\.role === ['"]reader['"][\s\S]{0,160}(管理此目录|编辑文件)/);
});

test('Stage 2 frontend types model the protected tree and draft/published-only statuses', () => {
  assert.match(typesSource, /AdminTreeNode/);
  assert.match(typesSource, /AdminTreeResponse/);
  assert.match(typesSource, /status:\s*'draft' \| 'published'/);
  assert.doesNotMatch(typesSource, /有未发布修改|unpublished_changes|modified/);
});
