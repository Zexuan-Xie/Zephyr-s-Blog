import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

function source(path) {
  return readFileSync(new URL(path, import.meta.url), 'utf8');
}

const adminPageSource = source('../src/pages/AdminPage.tsx');
const stage2ContractSource = source('./stage2-author-workspace-contract.test.mjs');
const apiSource = source('../src/lib/api.ts');

test('Stage 2 shell delegates minimal create Red coverage to the Gateway 4 contract', () => {
  assert.match(stage2ContractSource, /Stage 2 minimal create flow is slugless/);
  assert.ok(stage2ContractSource.includes('createAdminNode\\(input: CreateAdminNodeInput\\): Promise<AdminNodeDetail>'));
  assert.match(apiSource, /createAdminNode\(input: CreateAdminNodeInput\): Promise<AdminNodeDetail>/);
  assert.match(adminPageSource, /新建与移动操作将在后续工作包中接入/);
});

test('Stage 2 shell removes old implementation-language create controls from primary UI', () => {
  assert.doesNotMatch(adminPageSource, />Slug</);
  assert.doesNotMatch(adminPageSource, /Check slug|root slugs|slug uniqueness/i);
  assert.doesNotMatch(adminPageSource, /Parent id/i);
  assert.doesNotMatch(adminPageSource, /Sort order/i);
  assert.match(adminPageSource, /URL Path/);
});

test('Stage 2 create contract remains Red until the directory create packet implements it', () => {
  assert.ok(stage2ContractSource.includes('setSelectedId\\(created\\.node\\.id\\)'));
  assert.ok(stage2ContractSource.includes('adminTreeQuery\\.refetch\\(\\)'));
  assert.doesNotMatch(adminPageSource, /async function submitCreate/);
});
