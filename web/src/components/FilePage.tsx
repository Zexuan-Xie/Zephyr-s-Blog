import { Heart, MessageCircle } from 'lucide-react';
import { Link } from 'react-router-dom';
import { renderSafeMarkdown, sanitizeServerHtml } from '../lib/renderMarkdown';
import type { FilePayload } from '../lib/types';
import { Breadcrumb } from './Breadcrumb';

interface FilePageProps {
  file: FilePayload;
}

export function FilePage({ file }: FilePageProps) {
  const keywords = file.keywords?.slice(0, 3) ?? [];
  const markdownHtml = file.content_format === 'markdown'
    ? file.body_html
      ? sanitizeServerHtml(file.body_html)
      : renderSafeMarkdown(file.body_markdown ?? '')
    : '';

  return (
    <article className="file-page">
      <Breadcrumb items={file.breadcrumb} currentPath={file.path} />
      <section className="file-heading">
        <div className="keyword-row">
          {keywords.map((keyword) => (
            <Link className="keyword-chip" key={keyword} to={`/search?q=${encodeURIComponent(keyword)}`}>
              {keyword}
            </Link>
          ))}
        </div>
        <h1>{file.name}</h1>
        <p className="muted">
          {file.path}
          {file.updated_at ? ` · updated ${new Date(file.updated_at).toLocaleDateString()}` : ''}
          {file.read_time_minutes ? ` · ${file.read_time_minutes} min read` : ''}
        </p>
      </section>

      {file.content_format === 'markdown' ? (
        <section className="glass file-reading-card" dangerouslySetInnerHTML={{ __html: markdownHtml }} />
      ) : (
        <section className="glass html-document-shell" aria-label={`${file.name} HTML document`}>
          <iframe
            title={`${file.name} document`}
            sandbox="allow-scripts"
            srcDoc={file.html_document ?? file.body_html ?? '<!doctype html><html><body><p>Empty HTML document.</p></body></html>'}
          />
        </section>
      )}

      <footer className="glass interaction-bar" aria-label="File interactions">
        <button className="glass-button" type="button">
          <Heart size={17} aria-hidden="true" />
          <span>{file.like_count ?? 0}</span>
        </button>
        <button className="glass-button" type="button">
          <MessageCircle size={17} aria-hidden="true" />
          <span>Log in to comment</span>
        </button>
      </footer>
    </article>
  );
}
