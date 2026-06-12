import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

function source(path) {
  return readFileSync(new URL(path, import.meta.url), 'utf8');
}

const adminPageSource = source('../src/pages/AdminPage.tsx');
const apiSource = source('../src/lib/api.ts');

test('successful Directory creation consumes the returned detail before UI cleanup', () => {
  const createStart = adminPageSource.indexOf('async function submitCreate');
  const nextHandler = adminPageSource.indexOf('async function submitNodeUpdate', createStart);
  const createFlow = adminPageSource.slice(createStart, nextHandler);

  assert.match(apiSource, /createAdminNode\(input: CreateAdminNodeInput\): Promise<AdminNodeDetail>/);
  assert.match(apiSource, /requestJson\(['"]\/admin\/nodes['"], adminNodeDetailSchema/);
  assert.match(createFlow, /const createForm = event\.currentTarget;/);
  assert.match(createFlow, /let created: AdminNodeDetail;/);
  assert.match(createFlow, /created = await createAdminNode\(input\);/);
  assert.match(createFlow, /setDetail\(created\);/);
  assert.match(createFlow, /setSelectedId\(created\.node\.id\);/);
  assert.match(createFlow, /rootQuery\.refetch\(\)/);
  assert.match(createFlow, /created\.node\.path/);
  assert.match(createFlow, /createForm\.reset\(\);/);

  const catchIndex = createFlow.indexOf('catch');
  assert.ok(catchIndex >= 0, 'create flow should handle API errors');
  assert.ok(createFlow.indexOf('setDetail(created)') > catchIndex, 'successful state updates must not be inside the API failure catch');
});

test('create failures are actionable and Author-facing without implementation language', () => {
  assert.match(adminPageSource, /formatAdminCreateError\(error\)/);
  assert.match(adminPageSource, /Your session expired\. Log in again\./);
  assert.match(adminPageSource, /This URL path is already in use\./);
  assert.match(adminPageSource, /The destination Directory no longer exists\./);
  assert.match(adminPageSource, /Check the connection and try again\./);
  assert.doesNotMatch(adminPageSource, />Slug</);
  assert.doesNotMatch(adminPageSource, /Check slug|root slugs|slug uniqueness/i);
});

test('URL Path conflicts take precedence over parent wording in the API message', () => {
  const formatterStart = adminPageSource.indexOf('function formatAdminCreateError');
  const formatterEnd = adminPageSource.indexOf('function formatBytes', formatterStart);
  const formatter = adminPageSource.slice(formatterStart, formatterEnd);
  const conflictIndex = formatter.indexOf('error.status === 409');
  const missingParentIndex = formatter.indexOf('/parent|destination/i');

  assert.ok(conflictIndex >= 0, 'create errors should classify HTTP 409 conflicts');
  assert.ok(missingParentIndex >= 0, 'create errors should classify a missing destination');
  assert.ok(
    conflictIndex < missingParentIndex,
    'HTTP 409 must win before matching backend conflict text that mentions the selected parent',
  );
});
