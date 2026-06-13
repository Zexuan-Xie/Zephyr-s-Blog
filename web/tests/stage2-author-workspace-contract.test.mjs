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
const glassNavSource = source('../src/components/GlassNav.tsx');

test('Stage 2 polish uses Aeolian brand and simple Author Workspace language', () => {
  assert.doesNotMatch(adminPageSource, />ADMIN</);
  assert.doesNotMatch(adminPageSource, /Tree Manager/);
  assert.doesNotMatch(adminPageSource, /Tree browser/);
  assert.doesNotMatch(adminPageSource, /Node id/);
  assert.doesNotMatch(adminPageSource, /Load selected node/);
  assert.match(glassNavSource, /Aeolian/);
  assert.doesNotMatch(glassNavSource, /xLab Blog/);
  assert.match(adminPageSource, /Author Workspace/);
  assert.match(adminPageSource, /Content Tree/);
  assert.match(adminPageSource, /Create, write, publish, and reorder/);
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
  assert.match(adminPageSource, /Directory/);
  assert.match(adminPageSource, /File/);
  assert.match(adminPageSource, /URL Path/);
});

test('Stage 2 empty Author Workspace can create the first root content item', () => {
  assert.match(adminPageSource, /EmptyRootWorkspace/);
  assert.match(adminPageSource, /Start here/);
  assert.match(adminPageSource, /Create your first item/);
  assert.match(adminPageSource, /Nothing here yet. Create your first item on the right./);
  assert.match(adminPageSource, /async function submitRootCreate/);
  assert.match(adminPageSource, /New root/);
  assert.match(adminPageSource, /rootCreateOpen/);
  assert.match(adminPageSource, /parent_id:\s*null/);
  assert.doesNotMatch(
    adminPageSource.slice(
      adminPageSource.indexOf('async function submitRootCreate'),
      adminPageSource.indexOf('async function deleteDirectory'),
    ),
    /slug:|sort_order:/,
  );
});

test('Stage 2 primary UI hides implementation identifiers while allowing Stage 3 publication UX', () => {
  for (const forbidden of [
    /Parent id/i,
    /Node id/i,
    /Sort order/i,
    /\bslug\b/i,
    /有未发布修改/,
    /发布更新/,
  ]) {
    assert.doesNotMatch(adminPageSource, forbidden);
  }

  assert.match(adminPageSource, /Draft/);
  assert.match(adminPageSource, /Live/);
  assert.match(adminPageSource, /Unpublish/);
  assert.match(adminPageSource, /Draft Preview/);
  assert.match(adminPageSource, /Autosave/);
});

test('Stage 2 workspace provides explicit return controls for subflows', () => {
  assert.match(adminPageSource, /Back to directory/);
  assert.match(adminPageSource, /Top/);
  assert.match(adminPageSource, /Cancel/);
});

test('Stage 2 public Author entries are role-gated and open admin with selected target', () => {
  const publicSources = [rootPageSource, directoryPageSource, contentResolverSource, filePageSource].join('\n');

  assert.match(publicSources, /currentUser\?\.role === ['"]admin['"]|isAuthor/);
  assert.match(publicSources, /Manage/);
  assert.match(publicSources, /Edit/);
  assert.match(publicSources, /\/admin\?[^`'"]*(target|node|select)=/);
  assert.doesNotMatch(publicSources, /currentUser\?\.role === ['"]reader['"][\s\S]{0,160}(Manage|Edit)/);
});

test('Stage 2 polish exposes tree-based same-parent drag', () => {
  assert.doesNotMatch(adminPageSource, /Drag cards/);
  assert.match(adminPageSource, /draggable/);
  assert.match(adminPageSource, /onDragStart/);
  assert.match(adminPageSource, /onDropNode=\{reorderWithinParent\}/);
  assert.match(adminPageSource, /GripVertical/);
  assert.match(adminPageSource, /author-tree-row/);
});

test('Stage 2 frontend types model the protected tree and Stage 3 publish statuses', () => {
  assert.match(typesSource, /AdminTreeNode/);
  assert.match(typesSource, /AdminTreeResponse/);
  assert.match(typesSource, /PublishStatus\s*=\s*"draft"\s*\|\s*"published"\s*\|\s*"unpublished_changes"/);
  assert.match(typesSource, /status:\s*PublishStatus/);
  assert.doesNotMatch(typesSource, /有未发布修改|modified/);
});

test('Stage 2 admin tree adapter accepts backend flat nodes contract and builds frontend roots', () => {
  assert.match(apiSource, /nodes:\s*z\.array\(flatAdminTreeNodeSchema\)/);
  assert.match(apiSource, /url_path:\s*z\.string\(\)/);
  assert.match(apiSource, /sort_order:\s*z\.number\(\)/);
  assert.match(apiSource, /buildAdminTreeRoots\(response\.nodes\)/);
  assert.match(apiSource, /path:\s*node\.url_path/);
  assert.match(apiSource, /parent\.children\.push\(treeNode\)/);
});

test('Stage 2 public resolver decodes browser pathname before API query encoding', () => {
  assert.match(apiSource, /function normalizeBrowserPath/);
  assert.match(apiSource, /decodeURIComponent\(path\)/);
  assert.match(apiSource, /encodeURIComponent\(normalizeBrowserPath\(path\)\)/);
});


test('Stage 2 drawer exposes expandable public tree', () => {
  const drawerSource = source('../src/components/DirectoryDrawer.tsx');
  assert.match(apiSource, /fetchDirectoryChildren/);
  assert.match(apiSource, /\/tree\/\$\{encodeURIComponent\(nodeId\)\}\/children/);
  assert.match(drawerSource, /DrawerTreeItem/);
  assert.match(drawerSource, /expandedIds/);
  assert.match(drawerSource, /fetchDirectoryChildren\(entry\.id\)/);
  assert.match(drawerSource, /ChevronRight/);
  assert.match(drawerSource, /to=\{entry\.path\}/);
});
