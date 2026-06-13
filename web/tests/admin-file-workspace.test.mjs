import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

function source(path) {
  return readFileSync(new URL(path, import.meta.url), 'utf8');
}

const adminPageSource = source('../src/pages/AdminPage.tsx');
const apiSource = source('../src/lib/api.ts');
const typesSource = source('../src/lib/types.ts');

test('Task 14 file workspace exposes manual save, assets, settings, and publish controls', () => {
  assert.match(adminPageSource, /type FileWorkspaceTab = ["']content["'] \| ["']assets["'] \| ["']settings["']/);
  assert.match(adminPageSource, />Write<|aria-label="Write"/);
  assert.match(adminPageSource, />Assets<|aria-label="Assets"/);
  assert.match(adminPageSource, />Settings<|aria-label="Settings"/);
  assert.match(adminPageSource, /Save/);
  assert.match(adminPageSource, /Saved/);
  assert.match(adminPageSource, /Publish/);
  assert.match(adminPageSource, /Unpublish/);
  assert.match(adminPageSource, /Draft/);
  assert.match(adminPageSource, /Live/);
});

test('Task 14 API helpers bind Stage 2 workspace endpoints without widening create input', () => {
  assert.doesNotMatch(apiSource, /CreateAdminNodeInput[\s\S]*slug\?:/);
  assert.doesNotMatch(apiSource, /CreateAdminNodeInput[\s\S]*sort_order\?:/);
  assert.match(apiSource, /createAdminNode\(input: CreateAdminNodeInput\): Promise<AdminNodeDetail>/);
  assert.match(apiSource, /upsertFileContent/);
  assert.match(apiSource, /publishFile/);
  assert.match(apiSource, /unpublishFile/);
  assert.match(apiSource, /uploadAsset/);
  assert.match(apiSource, /deleteAsset/);
  assert.match(apiSource, /reorderAdminChildren/);
  assert.match(apiSource, /previewAdminMove/);
  assert.match(apiSource, /moveAdminNode/);
  assert.match(typesSource, /MovePreviewResponse/);
  assert.match(typesSource, /PublishStatus\s*=\s*"draft"\s*\|\s*"published"\s*\|\s*"unpublished_changes"/);
  assert.match(typesSource, /status:\s*PublishStatus/);
});

test('Task 14 directory settings include guarded delete and tree-based same-parent drag', () => {
  assert.match(adminPageSource, /Not empty|Delete is blocked/);
  assert.match(adminPageSource, /deleteAdminNode\(node\.id\)/);
  assert.match(adminPageSource, /draggable/);
  assert.doesNotMatch(adminPageSource, /Drag cards/);
  assert.match(adminPageSource, /reorderAdminChildren/);
  assert.doesNotMatch(adminPageSource, /有未发布修改|发布更新/);
  assert.match(adminPageSource, /Draft Preview|Autosave/);
});

test('Task 14 settings provide Directory Picker move preview and danger feedback', () => {
  assert.match(adminPageSource, /Directory Picker/);
  assert.match(adminPageSource, /previewAdminMove/);
  assert.match(adminPageSource, /moveAdminNode/);
  assert.match(adminPageSource, /Move preview/);
  assert.match(adminPageSource, /Move here/);
  assert.match(adminPageSource, /Live files cannot be deleted|Live files are protected/);
});
