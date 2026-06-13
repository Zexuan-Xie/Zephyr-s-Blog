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
  assert.match(adminPageSource, />内容<|aria-label="内容"/);
  assert.match(adminPageSource, />资源<|aria-label="资源"/);
  assert.match(adminPageSource, />设置<|aria-label="设置"/);
  assert.match(adminPageSource, /手动保存/);
  assert.match(adminPageSource, /内容已手动保存/);
  assert.match(adminPageSource, /发布/);
  assert.match(adminPageSource, /撤回发布/);
  assert.match(adminPageSource, /草稿/);
  assert.match(adminPageSource, /已发布/);
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
  assert.match(typesSource, /status:\s*'draft' \| 'published'/);
});

test('Task 14 directory settings include guarded delete and same-parent desktop drag wording', () => {
  assert.match(adminPageSource, /非空目录不能删除/);
  assert.match(adminPageSource, /deleteAdminNode\(node\.id\)/);
  assert.match(adminPageSource, /draggable/);
  assert.match(adminPageSource, /同级拖拽排序/);
  assert.match(adminPageSource, /reorderAdminChildren/);
  assert.doesNotMatch(adminPageSource, /有未发布修改|发布更新|Draft Preview|Autosave/);
});

test('Task 14 settings provide Directory Picker move preview and danger feedback', () => {
  assert.match(adminPageSource, /Directory Picker/);
  assert.match(adminPageSource, /previewAdminMove/);
  assert.match(adminPageSource, /moveAdminNode/);
  assert.match(adminPageSource, /移动预览已生成/);
  assert.match(adminPageSource, /确认移动/);
  assert.match(adminPageSource, /已发布文件不能直接删除，请先撤回发布/);
});
