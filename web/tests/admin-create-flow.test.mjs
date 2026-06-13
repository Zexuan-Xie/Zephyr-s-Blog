import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

function source(path) {
  return readFileSync(new URL(path, import.meta.url), 'utf8');
}

const adminPageSource = source('../src/pages/AdminPage.tsx');
const apiSource = source('../src/lib/api.ts');

test('Stage 2 minimal create is slugless and opens the returned node after tree refresh', () => {
  const createStart = adminPageSource.indexOf('async function submitCreate');
  const logoutStart = adminPageSource.indexOf('function logoutAuthor', createStart);
  const createFlow = adminPageSource.slice(createStart, logoutStart);

  assert.match(apiSource, /createAdminNode\(input: CreateAdminNodeInput\): Promise<AdminNodeDetail>/);
  assert.doesNotMatch(apiSource, /CreateAdminNodeInput[\s\S]*slug:/);
  assert.match(createFlow, /const createForm = event\.currentTarget/);
  assert.match(createFlow, /let created: AdminNodeDetail/);
  assert.match(createFlow, /created = await createAdminNode\(input\)/);
  assert.match(createFlow, /setSelectedId\(created\.node\.id\)/);
  assert.match(createFlow, /setExpandedIds/);
  assert.match(createFlow, /await adminTreeQuery\.refetch\(\)/);
  assert.match(createFlow, /created\.node\.path/);
  assert.doesNotMatch(createFlow, /slug:/);
  assert.doesNotMatch(createFlow, /sort_order:/);
});

test('Stage 2 create UI is Chinese, minimal, and hides implementation controls', () => {
  assert.match(adminPageSource, /新建目录/);
  assert.match(adminPageSource, /新建文件/);
  assert.match(adminPageSource, /创建并打开/);
  assert.match(adminPageSource, /URL Path preview/);
  assert.match(adminPageSource, /readOnly/);
  assert.doesNotMatch(adminPageSource, />Slug</);
  assert.doesNotMatch(adminPageSource, /Check slug|root slugs|slug uniqueness/i);
  assert.doesNotMatch(adminPageSource, /Parent id/i);
  assert.doesNotMatch(adminPageSource, /Sort order/i);
});

test('Stage 2 create failures use Chinese actionable feedback', () => {
  assert.match(adminPageSource, /formatAdminCreateError\(error\)/);
  assert.match(adminPageSource, /登录已过期，请重新登录/);
  assert.match(adminPageSource, /需要作者权限才能创建内容/);
  assert.match(adminPageSource, /目标目录不存在/);
  assert.match(adminPageSource, /URL Path 已存在/);
  assert.match(adminPageSource, /创建失败，请检查网络后重试/);
});
