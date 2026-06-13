import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

function source(path) {
  return readFileSync(new URL(path, import.meta.url), 'utf8');
}

function assertAny(sourceText, patterns, message) {
  assert.ok(
    patterns.some((pattern) => pattern.test(sourceText)),
    message ?? `Expected one of ${patterns.map(String).join(', ')}`,
  );
}

const adminPageSource = source('../src/pages/AdminPage.tsx');
const apiSource = source('../src/lib/api.ts');
const typesSource = source('../src/lib/types.ts');

test('Stage 3 frontend models Current/Previous versions, optimistic revision saves, and restore', () => {
  assert.match(typesSource, /revision:\s*number/, 'file content types must expose a numeric revision');
  assert.match(typesSource, /previous/i, 'frontend types must model the Previous content slot');
  assert.match(typesSource, /last_saved_at|saved_at/i, 'frontend types must expose save timestamps');

  assert.match(apiSource, /expected_revision/, 'save requests must send expected_revision');
  assert.match(apiSource, /revision/, 'API parser must accept revision from the backend');
  assertAny(
    apiSource,
    [/fetch(Current|File).*Previous/i, /fetch.*Versions/i, /get.*Previous/i],
    'API helpers must fetch Current and Previous content versions',
  );
  assertAny(
    apiSource,
    [/restore.*Previous/i, /restore.*Version/i],
    'API helpers must expose reversible Previous restore',
  );

  assert.match(adminPageSource, /Current/i, 'workspace must label the Current version');
  assert.match(adminPageSource, /Previous/i, 'workspace must label the Previous version');
  assert.match(adminPageSource, /Restore/i, 'workspace must expose restore controls');
  assert.match(adminPageSource, /Compare/i, 'workspace must expose Current/Previous compare');
});

test('Stage 3 Author Workspace autosaves with required-save triggers and guarded failure behavior', () => {
  for (const state of [
    'Editing',
    'Saving',
    'Saved',
    'Save failed',
    'Conflict',
    'Unpublished changes',
  ]) {
    assert.match(adminPageSource, new RegExp(state), `missing autosave state: ${state}`);
  }

  assertAny(
    adminPageSource,
    [/useAutosaveFile/, /setTimeout\([^,]+,\s*15_?000\)/, /15000/],
    'workspace must debounce autosave for 15 seconds after input stops',
  );
  assert.match(adminPageSource, /onBlur/, 'blur must trigger an immediate save');
  assert.match(adminPageSource, /beforeunload|useUnsavedNavigationGuard|blocker/i, 'unsafe leave/navigation must be guarded');
  assert.match(adminPageSource, /logout/i, 'logout must participate in required-save guarding');
  assert.match(adminPageSource, /publish/i, 'publish must require a successful save first');
  assert.match(adminPageSource, /Save failed/i, 'failed saves must be visible');
  assert.match(adminPageSource, /preserve|draft text|typed text|local draft/i, 'failed saves must preserve typed text');
});

test('Stage 3 frontend handles stale revision conflicts without auto-merging', () => {
  assert.match(apiSource, /409|Conflict/i, 'API layer must preserve 409 conflict errors');
  assert.match(adminPageSource, /Conflict/i, 'workspace must show a conflict state');
  assert.match(adminPageSource, /Reload latest/i, 'conflict UI must offer Reload latest');
  assert.match(adminPageSource, /Copy my changes/i, 'conflict UI must offer Copy my changes');
  assert.doesNotMatch(
    adminPageSource,
    /auto[- ]?merge|merge\s+automatically/i,
    'conflict UI must not promise automatic merging',
  );
});

test('Stage 3 publish controls distinguish stable Published Content snapshots from draft edits', () => {
  assertAny(
    apiSource,
    [/published.*content/i, /publish.*snapshot/i, /source_revision/i],
    'API helpers/types must model independent Published Content snapshots',
  );
  assert.match(adminPageSource, /Publish changes/i, 'published files with draft edits must show Publish changes');
  assert.match(adminPageSource, /Published/i, 'workspace must show a Published state');
  assert.match(adminPageSource, /Unpublished changes/i, 'workspace must show unpublished draft edits');
  assert.match(adminPageSource, /Publish summary|will become public/i, 'publish flow must summarize content/asset changes');
  assert.match(adminPageSource, /Unpublish/i, 'unpublish must remain available as a danger/overflow action');
});

test('Stage 3 Draft Preview is Author-only and separate from public Published Content', () => {
  assertAny(
    apiSource,
    [/fetchDraftPreview/i, /\/admin\/preview/, /draftPreview/i],
    'API layer must expose an Author-only draft preview endpoint',
  );
  assert.match(adminPageSource, /Draft Preview/i, 'workspace must expose Draft Preview');
  assert.match(adminPageSource, /PreviewSplit|editor.*preview|preview.*split/i, 'desktop editor/preview split must be represented');
  assert.match(adminPageSource, /Author-only|Author only|Requires author/i, 'Draft Preview must communicate author-only access');
  assert.doesNotMatch(
    adminPageSource,
    /allow-same-origin/,
    'HTML document preview must preserve iframe sandbox without allow-same-origin',
  );
});

test('Stage 3 asset UX separates Draft Assets from Published Assets and safe promotion/delete semantics', () => {
  assert.match(typesSource, /draft.*assets|published.*assets|asset.*state/i, 'types must model draft/published asset state');
  assert.match(apiSource, /draft.*assets|published.*assets|asset.*state/i, 'API layer must parse draft/published assets');
  assert.match(adminPageSource, /Draft assets/i, 'workspace must label Draft assets');
  assert.match(adminPageSource, /Published assets/i, 'workspace must label Published assets');
  assert.match(adminPageSource, /not public until Publish|will become public/i, 'asset upload copy must say draft uploads are not public yet');
  assert.match(adminPageSource, /Published snapshot|published content before the next Publish/i, 'deleting draft state must not break current public assets');
});
