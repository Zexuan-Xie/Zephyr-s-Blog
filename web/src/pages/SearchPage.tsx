import { FormEvent, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { searchFiles } from '../lib/api';

export function SearchPage() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const query = params.get('q')?.trim() ?? '';
  const [draft, setDraft] = useState(query);
  const searchQuery = useQuery({
    queryKey: ['search', query],
    queryFn: () => searchFiles(query),
    enabled: query.length > 0,
  });

  function submitSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmed = draft.trim();
    navigate(trimmed ? `/search?q=${encodeURIComponent(trimmed)}` : '/search');
  }

  return (
    <section className="page-stack">
      <section className="glass search-panel">
        <p className="eyebrow">SEARCH</p>
        <h1>Find published files</h1>
        <form className="large-search" onSubmit={submitSearch}>
          <input value={draft} onChange={(event) => setDraft(event.target.value)} placeholder="Search by path, keyword, or text" />
          <button className="primary-button" type="submit">Search</button>
        </form>
      </section>

      {!query ? <section className="glass empty-panel">Enter a query to search text, semantic, and keyword matches.</section> : null}
      {searchQuery.isLoading ? <section className="glass status-panel">Searching…</section> : null}
      {searchQuery.isError ? <section className="glass status-panel error">Search failed. Full-text fallback is expected server-side when embeddings are unavailable.</section> : null}
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
