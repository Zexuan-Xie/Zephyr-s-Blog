import { Link } from 'react-router-dom';
import type { CurrentUser, DirectoryPayload } from '../lib/types';
import { Breadcrumb } from '../components/Breadcrumb';
import { ContentEntryCard } from '../components/ContentEntryCard';

interface DirectoryPageProps {
  directory: DirectoryPayload;
  isRoot?: boolean;
  currentUser: CurrentUser | null;
}

export function DirectoryPage({ directory, isRoot = false, currentUser }: DirectoryPageProps) {
  const isAuthor = currentUser?.role === 'admin';

  return (
    <section className="page-stack">
      <Breadcrumb items={directory.breadcrumb} currentPath={directory.path} />
      {isRoot ? (
        <section className="glass hero-panel">
          <p className="eyebrow">Aeolian</p>
          <h1 className="hero-title">Notes the wind leaves behind.</h1>
          <p>
            Stray currents of thought, caught and written down.
          </p>
        </section>
      ) : (
        <header className="section-heading">
          <p className="eyebrow">DIRECTORY</p>
          <h1>{directory.name}</h1>
          <p className="muted">{directory.path}</p>
        </header>
      )}
      {isAuthor ? (
        <div className="action-row" aria-label="Author directory actions">
          <Link className="glass-button" to={`/admin?target=${encodeURIComponent(directory.id)}`}>Manage</Link>
        </div>
      ) : null}
      {directory.children.length === 0 ? (
        <section className="glass empty-panel">No files yet.</section>
      ) : (
        <div className="entry-grid">
          {directory.children.map((entry) => (
            <ContentEntryCard entry={entry} key={entry.id} />
          ))}
        </div>
      )}
    </section>
  );
}
