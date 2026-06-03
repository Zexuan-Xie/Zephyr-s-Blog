import { useQuery } from '@tanstack/react-query';
import { fetchRootDirectory } from '../lib/api';
import { DirectoryPage } from './DirectoryPage';

export function RootPage() {
  const rootQuery = useQuery({
    queryKey: ['tree', 'root'],
    queryFn: fetchRootDirectory,
  });

  if (rootQuery.isLoading) {
    return <section className="glass status-panel">Loading root directory…</section>;
  }

  if (rootQuery.isError) {
    return <section className="glass status-panel error">Unable to load root directory.</section>;
  }

  return <DirectoryPage directory={rootQuery.data} isRoot />;
}
