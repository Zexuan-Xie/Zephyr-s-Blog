import { BookOpen, FolderTree, Search, UserRound } from 'lucide-react';
import { FormEvent, useState } from 'react';
import { Link, NavLink, useNavigate } from 'react-router-dom';

interface GlassNavProps {
  onOpenDirectory: () => void;
}

export function GlassNav({ onOpenDirectory }: GlassNavProps) {
  const navigate = useNavigate();
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

  return (
    <header className="site-header">
      <nav className="glass glass-nav" aria-label="Primary navigation">
        <Link className="brand" to="/" aria-label="xLab Blog root">
          <BookOpen size={18} aria-hidden="true" />
          <span>xLab Blog</span>
        </Link>
        <div className="nav-links">
          <NavLink to="/recent">Recent</NavLink>
          <NavLink to="/search">Search</NavLink>
          <NavLink to="/admin">Admin</NavLink>
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
          <span>Tree</span>
        </button>
        <Link className="glass-button nav-icon" to="/login">
          <UserRound size={17} aria-hidden="true" />
          <span>Login</span>
        </Link>
      </nav>
    </header>
  );
}
