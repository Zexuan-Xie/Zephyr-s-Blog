import { X } from 'lucide-react';
import { Link } from 'react-router-dom';
import type { ContentEntry } from '../lib/types';

interface DirectoryDrawerProps {
  entries: ContentEntry[];
  isOpen: boolean;
  onClose: () => void;
}

export function DirectoryDrawer({ entries, isOpen, onClose }: DirectoryDrawerProps) {
  return (
    <div className={isOpen ? 'drawer-layer open' : 'drawer-layer'} aria-hidden={!isOpen}>
      <button className="drawer-scrim" type="button" onClick={onClose} aria-label="Close directory drawer" />
      <aside className="glass directory-drawer" aria-label="Directory tree">
        <div className="drawer-header">
          <div>
            <p className="eyebrow">Content Tree</p>
            <h2>Directories & files</h2>
          </div>
          <button className="glass-button icon-only" type="button" onClick={onClose} aria-label="Close">
            <X size={18} aria-hidden="true" />
          </button>
        </div>
        {entries.length === 0 ? (
          <p className="muted">Tree will appear after the API returns root entries.</p>
        ) : (
          <ul className="tree-list">
            {entries.map((entry) => (
              <li key={entry.id}>
                <Link to={entry.path} onClick={onClose}>
                  <span className="entry-kind">{entry.kind}</span>
                  <span>{entry.path}</span>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </aside>
    </div>
  );
}
