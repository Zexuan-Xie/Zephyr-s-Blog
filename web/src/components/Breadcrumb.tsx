import { Link } from 'react-router-dom';
import type { BreadcrumbItem } from '../lib/types';

interface BreadcrumbProps {
  items?: BreadcrumbItem[];
  currentPath: string;
}

export function Breadcrumb({ items, currentPath }: BreadcrumbProps) {
  const crumbs = items?.length ? items : [{ name: 'Root', path: '/' }];

  return (
    <nav className="breadcrumb" aria-label="Breadcrumb">
      {crumbs.map((item, index) => (
        <span key={`${item.path}-${index}`}>
          <Link to={item.path}>{item.name || 'Root'}</Link>
          {index < crumbs.length - 1 ? <span className="separator">/</span> : null}
        </span>
      ))}
      {!items?.length && currentPath !== '/' ? <span className="separator">{currentPath}</span> : null}
    </nav>
  );
}
