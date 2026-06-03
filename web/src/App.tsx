import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Route, Routes } from 'react-router-dom';
import { DirectoryDrawer } from './components/DirectoryDrawer';
import { GlassNav } from './components/GlassNav';
import { fetchRootDirectory } from './lib/api';
import { AdminPage } from './pages/AdminPage';
import { AuthPage } from './pages/AuthPages';
import { ContentResolverPage } from './pages/ContentResolverPage';
import { RecentPage } from './pages/RecentPage';
import { RootPage } from './pages/RootPage';
import { SearchPage } from './pages/SearchPage';

export function App() {
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);
  const rootQuery = useQuery({
    queryKey: ['tree', 'root', 'drawer'],
    queryFn: fetchRootDirectory,
    staleTime: 60_000,
  });

  return (
    <div className="app-shell">
      <GlassNav onOpenDirectory={() => setIsDrawerOpen(true)} />
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
          <Route path="/login" element={<AuthPage mode="login" />} />
          <Route path="/register" element={<AuthPage mode="register" />} />
          <Route path="/admin" element={<AdminPage />} />
          <Route path="/*" element={<ContentResolverPage />} />
        </Routes>
      </main>
    </div>
  );
}
