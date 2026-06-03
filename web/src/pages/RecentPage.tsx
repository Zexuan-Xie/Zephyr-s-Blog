import { useQuery } from '@tanstack/react-query';
import { ContentEntryCard } from '../components/ContentEntryCard';
import { fetchRecentFiles } from '../lib/api';

export function RecentPage() {
  const recentQuery = useQuery({
    queryKey: ['recent'],
    queryFn: fetchRecentFiles,
  });

  if (recentQuery.isLoading) {
    return <section className="glass status-panel">Loading recent files…</section>;
  }

  if (recentQuery.isError || !recentQuery.data) {
    return <section className="glass status-panel error">Unable to load recent files.</section>;
  }

  return (
    <section className="page-stack">
      <header className="section-heading">
        <p className="eyebrow">RECENT</p>
        <h1>Recently updated files</h1>
        <p className="muted">Published Markdown and HTML Document files only.</p>
      </header>
      {recentQuery.data.length === 0 ? (
        <section className="glass empty-panel">No recent files</section>
      ) : (
        <div className="entry-grid">
          {recentQuery.data.map((entry) => (
            <ContentEntryCard entry={{ ...entry, kind: 'file' }} key={entry.id} />
          ))}
        </div>
      )}
    </section>
  );
}
