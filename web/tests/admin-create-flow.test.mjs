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

test('Stage 2 create UI is simple, minimal, and hides implementation controls', () => {
  assert.match(adminPageSource, /Directory/);
  assert.match(adminPageSource, /File/);
  assert.match(adminPageSource, /Create/);
  assert.match(adminPageSource, /URL Path preview/);
  assert.match(adminPageSource, /readOnly/);
  assert.doesNotMatch(adminPageSource, />Slug</);
  assert.doesNotMatch(adminPageSource, /Check slug|root slugs|slug uniqueness/i);
  assert.doesNotMatch(adminPageSource, /Parent id/i);
  assert.doesNotMatch(adminPageSource, /Sort order/i);
});

test('Stage 2 create failures use simple actionable feedback', () => {
  assert.match(adminPageSource, /formatAdminCreateError\(error\)/);
  assert.match(adminPageSource, /Sign in again/);
  assert.match(adminPageSource, /Author access is required/);
  assert.match(adminPageSource, /Target directory was not found/);
  assert.match(adminPageSource, /URL Path already exists/);
  assert.match(adminPageSource, /Create failed/);
});
