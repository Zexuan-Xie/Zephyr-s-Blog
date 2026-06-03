import { FileText, Folder } from 'lucide-react';
import { Link } from 'react-router-dom';
import type { ContentEntry } from '../lib/types';

interface ContentEntryCardProps {
  entry: ContentEntry;
}

export function ContentEntryCard({ entry }: ContentEntryCardProps) {
  const isDirectory = entry.kind === 'directory';
  const Icon = isDirectory ? Folder : FileText;

  return (
    <Link className="glass content-entry-card" to={entry.path}>
      <div className="card-label-row">
        <span className="eyebrow">{isDirectory ? 'DIRECTORY' : 'FILE'}</span>
        <Icon size={18} aria-hidden="true" />
      </div>
      <h3>{entry.name}</h3>
      <p className="path-text">{entry.path}</p>
      <p className="muted">
        {isDirectory
          ? `${entry.child_directory_count ?? 0} dirs · ${entry.child_file_count ?? 0} files`
          : `${entry.content_format ?? 'markdown'}${entry.read_time_minutes ? ` · ${entry.read_time_minutes} min` : ''}`}
      </p>
    </Link>
  );
}
