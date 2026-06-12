import { useQuery } from '@tanstack/react-query';
import { Link, useSearchParams } from 'react-router-dom';
import { searchFiles } from '../lib/api';

export function SearchPage() {
  const [params] = useSearchParams();
  const query = params.get('q')?.trim() ?? '';
  const searchQuery = useQuery({
    queryKey: ['search', query],
    queryFn: () => searchFiles(query),
    enabled: query.length > 0,
  });

  return (
    <section className="page-stack">
      <section className="glass search-panel">
        <p className="eyebrow">SEARCH</p>
        <h1>{query ? `Results for “${query}”` : 'Find published files'}</h1>
        <p className="muted">Use the search field in the global bar to search published content.</p>
      </section>

      {!query ? <section className="glass empty-panel">Enter a query to search text, semantic, and keyword matches.</section> : null}
      {searchQuery.isLoading ? <section className="glass status-panel">Searching…</section> : null}
      {searchQuery.isError ? <section className="glass status-panel error">Search failed. Full-text fallback is expected server-side when embeddings are unavailable.</section> : null}
      {query && searchQuery.isSuccess && searchQuery.data.length === 0
        ? <section className="glass empty-panel">No published files matched “{query}”.</section>
        : null}
      {searchQuery.data?.map((result) => (
        <Link className="glass search-result" to={result.path} key={result.id}>
          <span className="eyebrow">FILE</span>
          <h2>{result.name}</h2>
          <p className="path-text">{result.path}</p>
          <p>{result.snippet}</p>
          <div className="keyword-row">
            {result.sources.map((source) => <span className="keyword-chip" key={source}>{source}</span>)}
          </div>
        </Link>
      ))}
    </section>
  );
}
