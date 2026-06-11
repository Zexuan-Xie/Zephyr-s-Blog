import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

function source(path) {
  return readFileSync(new URL(path, import.meta.url), 'utf8');
}

const appSource = source('../src/App.tsx');
const navSource = source('../src/components/GlassNav.tsx');
const authPageSource = source('../src/pages/AuthPages.tsx');
const searchPageSource = source('../src/pages/SearchPage.tsx');
const apiSource = source('../src/lib/api.ts');
const authSource = source('../src/lib/auth.ts');

test('App owns the single current-user state and distinguishes invalid credentials from outages', () => {
  assert.match(appSource, /fetchCurrentUser/);
  assert.match(appSource, /queryKey:\s*\[['"]auth['"],\s*['"]current-user['"]\]/);
  assert.match(appSource, /clearToken\(\)/);
  assert.match(appSource, /\b401\b/);
  assert.match(appSource, /retryIdentity|refetch/);
  assert.match(apiSource, /class ApiError extends Error/);
  assert.match(apiSource, /response\.status/);
});

test('the global bar has one search field, Directory, Recent, and one truthful identity entry', () => {
  assert.equal(navSource.match(/role="search"/g)?.length, 1);
  assert.equal(navSource.match(/<input\b/g)?.length, 1);
  assert.match(navSource, />Recent</);
  assert.match(navSource, />Directory</);
  assert.doesNotMatch(navSource, /<NavLink to="\/search">/);
  assert.doesNotMatch(navSource, /<NavLink to="\/admin">/);
  assert.match(navSource, /identity-loading|Loading identity|Checking identity/);
  assert.match(navSource, /Retry/);
  assert.match(navSource, /currentUser\?\.role === ['"]admin['"]/);
  assert.match(navSource, /to="\/admin"/);
  assert.match(navSource, /Logout/);
});

test('Admin routing keeps Anonymous and Reader outcomes distinct', () => {
  assert.match(appSource, /\/login\?return_to=%2Fadmin/);
  assert.match(appSource, /Author access required/);
  assert.match(appSource, /Return to Recent/);
  assert.match(appSource, /currentUser\?\.role === ['"]admin['"]/);
});

test('Search is a results-only page because the global bar owns the sole search input', () => {
  assert.doesNotMatch(searchPageSource, /<form\b/);
  assert.doesNotMatch(searchPageSource, /<input\b/);
  assert.match(searchPageSource, /useSearchParams/);
  assert.match(searchPageSource, /searchFiles\(query\)/);
  assert.match(searchPageSource, /Searching…/);
  assert.match(searchPageSource, /Search failed/);
  assert.match(searchPageSource, /No published files matched/);
});

test('login and logout use safe, loop-free destinations', () => {
  assert.match(authSource, /isSafeReturnTo|sanitizeReturnTo/);
  assert.match(authSource, /startsWith\(['"]\/['"]\)/);
  assert.match(authSource, /startsWith\(['"]\/\/['"]\)/);
  assert.match(authSource, /['"]\/login['"]/);
  assert.match(authPageSource, /getReturnTo\(['"]\/recent['"]\)/);
  assert.match(navSource, /clearToken\(\)/);
  assert.match(navSource, /navigate\(['"]\/recent['"]/);
});
