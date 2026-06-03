import type { DirectoryPayload } from '../lib/types';
import { Breadcrumb } from '../components/Breadcrumb';
import { ContentEntryCard } from '../components/ContentEntryCard';

interface DirectoryPageProps {
  directory: DirectoryPayload;
  isRoot?: boolean;
}

export function DirectoryPage({ directory, isRoot = false }: DirectoryPageProps) {
  return (
    <section className="page-stack">
      <Breadcrumb items={directory.breadcrumb} currentPath={directory.path} />
      {isRoot ? (
        <section className="glass hero-panel">
          <p className="eyebrow">Knowledge space</p>
          <h1>Warm technical notes in a Unix-like content tree.</h1>
          <p>
            Browse directories, read Markdown files, and open sandboxed HTML documents without leaving the
            Glass Ricepaper shell.
          </p>
        </section>
      ) : (
        <header className="section-heading">
          <p className="eyebrow">DIRECTORY</p>
          <h1>{directory.name}</h1>
          <p className="muted">{directory.path}</p>
        </header>
      )}
      {directory.children.length === 0 ? (
        <section className="glass empty-panel">No files yet / 暂无内容</section>
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
