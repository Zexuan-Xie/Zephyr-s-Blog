import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { resolveContentPath } from '../lib/api';
import { FilePage } from '../components/FilePage';
import { DirectoryPage } from './DirectoryPage';
import type { CurrentUser } from '../lib/types';

export function ContentResolverPage({ currentUser }: { currentUser: CurrentUser | null }) {
  const location = useLocation();
  const navigate = useNavigate();
  const path = location.pathname;
  const resolveQuery = useQuery({
    queryKey: ['tree', 'resolve', path],
    queryFn: () => resolveContentPath(path),
    retry: false,
  });

  useEffect(() => {
    if (resolveQuery.data?.type === 'redirect') {
      navigate(resolveQuery.data.new_path, { replace: true });
    }
  }, [navigate, resolveQuery.data]);

  if (resolveQuery.isLoading) {
    return <section className="glass status-panel">Resolving {path}…</section>;
  }

  if (resolveQuery.isError || !resolveQuery.data) {
    return (
      <section className="glass status-panel error">
        <p className="eyebrow">404</p>
        <h1>Path not found</h1>
        <p>{path} does not map to a published directory or file.</p>
        <div className="action-row">
          <Link className="primary-button" to="/">Return root</Link>
          <Link className="glass-button" to={`/search?q=${encodeURIComponent(path)}`}>Search this path</Link>
        </div>
      </section>
    );
  }

  if (resolveQuery.data.type === 'redirect') {
    return <section className="glass status-panel">Redirecting…</section>;
  }

  if ('content_format' in resolveQuery.data) {
    return <FilePage file={resolveQuery.data} currentUser={currentUser} />;
  }

  return <DirectoryPage directory={resolveQuery.data} currentUser={currentUser} />;
}
