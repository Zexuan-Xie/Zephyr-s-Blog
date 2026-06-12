import { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Link, Navigate, Route, Routes } from 'react-router-dom';
import { DirectoryDrawer } from './components/DirectoryDrawer';
import { GlassNav } from './components/GlassNav';
import { ApiError, fetchCurrentUser, fetchRootDirectory } from './lib/api';
import { clearToken, getToken } from './lib/auth';
import { AdminPage } from './pages/AdminPage';
import { AuthPage } from './pages/AuthPages';
import { ContentResolverPage } from './pages/ContentResolverPage';
import { RecentPage } from './pages/RecentPage';
import { RootPage } from './pages/RootPage';
import { SearchPage } from './pages/SearchPage';

export function App() {
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);
  const [token, setIdentityToken] = useState(getToken);
  const queryClient = useQueryClient();
  const rootQuery = useQuery({
    queryKey: ['tree', 'root', 'drawer'],
    queryFn: fetchRootDirectory,
    staleTime: 60_000,
  });
  const identityQuery = useQuery({
    queryKey: ['auth', 'current-user'],
    queryFn: async () => {
      try {
        return await fetchCurrentUser();
      } catch (error) {
        if (error instanceof ApiError && error.status === 401) {
          clearToken();
          setIdentityToken(null);
        }
        throw error;
      }
    },
    enabled: Boolean(token),
    retry: false,
  });
  const currentUser = token ? identityQuery.data ?? null : null;
  const identityStatus = !token || identityQuery.isSuccess
    ? 'ready'
    : identityQuery.isError
      ? 'error'
      : 'loading';

  function resetIdentity() {
    queryClient.removeQueries({ queryKey: ['auth', 'current-user'] });
    setIdentityToken(getToken());
  }

  return (
    <div className="app-shell">
      <GlassNav
        currentUser={currentUser}
        identityStatus={identityStatus}
        onLogout={resetIdentity}
        onOpenDirectory={() => setIsDrawerOpen(true)}
        retryIdentity={() => void identityQuery.refetch()}
      />
      <DirectoryDrawer
        entries={rootQuery.data?.children ?? []}
        isOpen={isDrawerOpen}
        onClose={() => setIsDrawerOpen(false)}
      />
      <main className="main-content">
        <Routes>
          <Route path="/" element={<RootPage />} />
          <Route path="/recent" element={<RecentPage />} />
          <Route path="/search" element={<SearchPage />} />
          <Route path="/login" element={<AuthPage mode="login" onAuthenticated={resetIdentity} />} />
          <Route path="/register" element={<AuthPage mode="register" onAuthenticated={resetIdentity} />} />
          <Route
            path="/admin"
            element={
              identityStatus === 'loading'
                ? <section className="glass status-panel">Checking identity…</section>
                : identityStatus === 'error'
                  ? (
                    <section className="glass status-panel error">
                      <p>Identity check failed. Your session has not been changed.</p>
                      <button className="primary-button" type="button" onClick={() => void identityQuery.refetch()}>
                        Retry
                      </button>
                    </section>
                  )
                  : !currentUser
                    ? <Navigate to="/login?return_to=%2Fadmin" replace />
                    : currentUser?.role === 'admin'
                      ? <AdminPage onLogout={clearIdentity} />
                      : (
                        <section className="glass status-panel">
                          <p className="eyebrow">AUTHOR</p>
                          <h1>Author access required</h1>
                          <p>This Reader account stays signed in, but it cannot manage blog content.</p>
                          <Link className="primary-button" to="/recent">Return to Recent</Link>
                        </section>
                      )
            }
          />
          <Route path="/*" element={<ContentResolverPage />} />
        </Routes>
      </main>
    </div>
  );
}
