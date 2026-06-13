import { ChevronRight, FileText, Folder, X } from 'lucide-react';
import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchDirectoryChildren } from '../lib/api';
import type { AdminTreeNode, ContentEntry } from '../lib/types';

type DrawerEntry = ContentEntry | AdminTreeNode;

interface DirectoryDrawerProps {
  adminEntries?: AdminTreeNode[];
  entries: ContentEntry[];
  isOpen: boolean;
  onClose: () => void;
}

export function DirectoryDrawer({ adminEntries, entries, isOpen, onClose }: DirectoryDrawerProps) {
  const [expandedIds, setExpandedIds] = useState<Set<string>>(() => new Set());

  function toggleDirectory(entryId: string) {
    setExpandedIds((current) => {
      const next = new Set(current);
      if (next.has(entryId)) {
        next.delete(entryId);
      } else {
        next.add(entryId);
      }
      return next;
    });
  }

  const visibleEntries = adminEntries ?? entries;

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
        {visibleEntries.length === 0 ? (
          <p className="muted">No files yet.</p>
        ) : (
          <ul className="tree-list drawer-tree-list">
            {visibleEntries.map((entry) => (
              <DrawerTreeItem
                key={entry.id}
                entry={entry}
                depth={0}
                expandedIds={expandedIds}
                onClose={onClose}
                onToggle={toggleDirectory}
              />
            ))}
          </ul>
        )}
      </aside>
    </div>
  );
}

function DrawerTreeItem({
  entry,
  depth,
  expandedIds,
  onClose,
  onToggle,
}: {
  entry: DrawerEntry;
  depth: number;
  expandedIds: Set<string>;
  onClose: () => void;
  onToggle: (entryId: string) => void;
}) {
  const isDirectory = entry.kind === 'directory';
  const expanded = expandedIds.has(entry.id);
  const adminChildren = 'children' in entry ? entry.children : undefined;
  const publicChildDirectoryCount = 'child_directory_count' in entry
    ? (entry.child_directory_count ?? 0)
    : 0;
  const publicChildFileCount = 'child_file_count' in entry
    ? (entry.child_file_count ?? 0)
    : 0;
  const hasKnownChildren =
    isDirectory &&
    ((adminChildren?.length ?? 0) > 0 ||
      publicChildDirectoryCount > 0 ||
      publicChildFileCount > 0);
  const childQuery = useQuery({
    queryKey: ['tree', 'children', entry.id],
    queryFn: () => fetchDirectoryChildren(entry.id),
    enabled: expanded && isDirectory && !adminChildren,
    staleTime: 30_000,
  });
  const children: DrawerEntry[] = adminChildren ?? childQuery.data?.children ?? [];

  return (
    <li className="drawer-tree-item">
      <div className="drawer-tree-row" style={{ paddingLeft: `${depth * 1.05}rem` }}>
        {isDirectory ? (
          <button
            className="drawer-tree-toggle"
            type="button"
            aria-label={expanded ? `Collapse ${entry.name}` : `Expand ${entry.name}`}
            onClick={() => onToggle(entry.id)}
          >
            <ChevronRight
              className={expanded ? 'is-expanded' : ''}
              size={15}
              aria-hidden="true"
            />
          </button>
        ) : (
          <span className="drawer-tree-toggle" aria-hidden="true" />
        )}
        <Link className="drawer-tree-link" to={entry.path} onClick={onClose}>
          <span className="drawer-tree-icon" aria-hidden="true">
            {isDirectory ? <Folder size={15} /> : <FileText size={15} />}
          </span>
          <span className="drawer-tree-main">
            <span>{entry.name}</span>
            <small>{entry.path}</small>
          </span>
        </Link>
      </div>
      {expanded && isDirectory ? (
        <div className="drawer-tree-children">
          {childQuery.isLoading ? <p className="muted drawer-tree-note">Loading…</p> : null}
          {childQuery.isError ? <p className="form-error drawer-tree-note">Failed to load.</p> : null}
          {!childQuery.isLoading && !childQuery.isError && children.length === 0 ? (
            <p className="muted drawer-tree-note">
              {hasKnownChildren ? 'No visible files here yet.' : 'Empty.'}
            </p>
          ) : null}
          {children.length > 0 ? (
            <ul className="tree-list drawer-tree-list nested">
              {children.map((child) => (
                <DrawerTreeItem
                  key={child.id}
                  entry={child}
                  depth={depth + 1}
                  expandedIds={expandedIds}
                  onClose={onClose}
                  onToggle={onToggle}
                />
              ))}
            </ul>
          ) : null}
        </div>
      ) : null}
    </li>
  );
}
