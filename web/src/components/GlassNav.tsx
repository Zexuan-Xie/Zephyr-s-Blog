import { FolderTree, LogOut, Search, UserRound } from 'lucide-react';
import { FormEvent, useState } from 'react';
import { Link, NavLink, useLocation, useNavigate } from 'react-router-dom';
import { clearToken } from '../lib/auth';
import type { CurrentUser } from '../lib/types';

interface GlassNavProps {
  onOpenDirectory: () => void;
  currentUser: CurrentUser | null;
  identityStatus: 'loading' | 'error' | 'ready';
  retryIdentity: () => void;
  onLogout: () => void;
}

export function GlassNav({
  onOpenDirectory,
  currentUser,
  identityStatus,
  retryIdentity,
  onLogout,
}: GlassNavProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const [query, setQuery] = useState('');

  function submitSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmed = query.trim();
    if (trimmed) {
      navigate(`/search?q=${encodeURIComponent(trimmed)}`);
    } else {
      navigate('/search');
    }
  }

  function logout() {
    clearToken();
    onLogout();
    if (location.pathname.startsWith('/admin')) {
      navigate('/recent', { replace: true });
    }
  }

  return (
    <header className="site-header">
      <nav className="glass glass-nav" aria-label="Primary navigation">
        <Link className="brand" to="/" aria-label="Aeolian root">
          <span className="brand-dot" aria-hidden="true" />
          <span>Aeolian</span>
        </Link>
        <div className="nav-links">
          <NavLink to="/recent">Recent</NavLink>
        </div>
        <form className="nav-search" role="search" onSubmit={submitSearch}>
          <Search size={16} aria-hidden="true" />
          <input
            aria-label="Search files"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search"
          />
        </form>
        <button className="glass-button nav-icon" type="button" onClick={onOpenDirectory}>
          <FolderTree size={17} aria-hidden="true" />
          <span>Directory</span>
        </button>
        {identityStatus === 'loading' ? (
          <span className="glass-button nav-icon identity-loading" aria-label="Checking identity" />
        ) : null}
        {identityStatus === 'error' ? (
          <button className="glass-button nav-icon" type="button" onClick={retryIdentity}>
            <span>Retry</span>
          </button>
        ) : null}
        {identityStatus === 'ready' && !currentUser ? (
          <Link className="glass-button nav-icon" to="/login">
            <UserRound size={17} aria-hidden="true" />
            <span>Login</span>
          </Link>
        ) : null}
        {identityStatus === 'ready' && currentUser?.role === 'admin' ? (
          <Link className="glass-button nav-icon" to="/admin">
            <UserRound size={17} aria-hidden="true" />
            <span>Author</span>
          </Link>
        ) : null}
        {identityStatus === 'ready' && currentUser?.role === 'reader' ? (
          <details className="nav-icon">
            <summary className="glass-button">
              <UserRound size={17} aria-hidden="true" />
              <span>{currentUser.display_name || 'Reader'}</span>
            </summary>
            <button className="glass-button" type="button" onClick={logout}>
              <LogOut size={16} aria-hidden="true" />
              <span>Logout</span>
            </button>
          </details>
        ) : null}
      </nav>
    </header>
  );
}
