import { useEffect, useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Navigate, useSearchParams } from 'react-router-dom';
import {
  fetchAdminNode,
  fetchAdminTree,
  fetchCurrentUser,
} from '../lib/api';
import { getToken } from '../lib/auth';
import type { AdminNodeDetail, AdminTreeNode } from '../lib/types';

const selectionStorageKey = 'xlab-author-workspace:selected-node';
const expandedStorageKey = 'xlab-author-workspace:expanded-directories';

export function AdminPage({ onLogout }: { onLogout: () => void }) {
  const token = getToken();
  const [searchParams] = useSearchParams();
  const requestedTarget = searchParams.get('target') ?? searchParams.get('node') ?? searchParams.get('select') ?? '';
  const [selectedId, setSelectedId] = useState(() => requestedTarget || readStoredString(selectionStorageKey));
  const [expandedIds, setExpandedIds] = useState<Set<string>>(() => new Set(readStoredList(expandedStorageKey)));

  const viewerQuery = useQuery({
    queryKey: ['auth', 'me', 'admin'],
    queryFn: fetchCurrentUser,
    enabled: Boolean(token),
    retry: false,
  });
  const adminTreeQuery = useQuery({
    queryKey: ['admin', 'content-tree'],
    queryFn: fetchAdminTree,
    enabled: Boolean(token) && viewerQuery.data?.role === 'admin',
  });

  const flatTree = useMemo(() => flattenTree(adminTreeQuery.data?.roots ?? []), [adminTreeQuery.data]);
  const selectedNode = flatTree.find((node) => node.id === selectedId)
    ?? flatTree.find((node) => node.kind === 'directory')
    ?? flatTree[0]
    ?? null;
  const effectiveSelectedId = selectedNode?.id ?? selectedId;
  const visibleExpandedIds = selectedNode && adminTreeQuery.data
    ? expandAncestors(expandedIds, selectedNode.id, adminTreeQuery.data.roots)
    : expandedIds;
  const detailQuery = useQuery({
    queryKey: ['admin', 'node-detail', effectiveSelectedId],
    queryFn: () => fetchAdminNode(effectiveSelectedId),
    enabled: Boolean(effectiveSelectedId) && Boolean(selectedNode),
    retry: false,
  });

  useEffect(() => {
    if (selectedId) {
      window.localStorage.setItem(selectionStorageKey, selectedId);
    }
  }, [selectedId]);

  useEffect(() => {
    window.localStorage.setItem(expandedStorageKey, JSON.stringify([...expandedIds]));
  }, [expandedIds]);


  if (!token) {
    return <Navigate to="/login?return_to=%2Fadmin" replace />;
  }
  if (viewerQuery.isLoading) {
    return <section className="glass status-panel">正在确认作者权限…</section>;
  }
  if (viewerQuery.isError || viewerQuery.data?.role !== 'admin') {
    return <Navigate to="/login?return_to=%2Fadmin" replace />;
  }

  function selectNode(node: AdminTreeNode) {
    setSelectedId(node.id);
    if (node.kind === 'directory') {
      setExpandedIds((current) => new Set([...current, node.id]));
    }
  }

  function toggleDirectory(nodeId: string) {
    setExpandedIds((current) => {
      const next = new Set(current);
      if (next.has(nodeId)) {
        next.delete(nodeId);
      } else {
        next.add(nodeId);
      }
      return next;
    });
  }

  function logoutAuthor() {
    onLogout();
    window.location.assign('/');
  }

  return (
    <section className="page-stack admin-manager-page author-workspace-page">
      <section className="glass status-panel admin-hero author-workspace-hero">
        <p className="eyebrow">作者工作台</p>
        <h1>内容树</h1>
        <p>管理受保护的目录、草稿文件和已发布文件。URL Path 由系统展示，主要操作不暴露实现标识。</p>
        <div className="button-row">
          <button className="glass-button" type="button" onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}>返回内容树</button>
          <button className="glass-button" type="button" onClick={logoutAuthor}>退出登录</button>
        </div>
      </section>

      <section className="admin-grid author-workspace-grid">
        <aside className="glass admin-sidebar author-tree-panel" aria-label="内容树">
          <div className="panel-heading-row">
            <div>
              <p className="eyebrow">Content Tree</p>
              <h2>受保护内容树</h2>
            </div>
            <button className="glass-button" type="button" onClick={() => adminTreeQuery.refetch()}>刷新</button>
          </div>
          {adminTreeQuery.isLoading ? <p className="muted">正在加载目录、草稿和已发布文件…</p> : null}
          {adminTreeQuery.isError ? <p className="form-error">内容树加载失败。请刷新或重新登录。</p> : null}
          {adminTreeQuery.data && adminTreeQuery.data.roots.length === 0 ? <p className="muted">暂无内容。</p> : null}
          {adminTreeQuery.data ? (
            <TreeList
              nodes={adminTreeQuery.data.roots}
              expandedIds={visibleExpandedIds}
              selectedId={effectiveSelectedId}
              onSelect={selectNode}
              onToggle={toggleDirectory}
            />
          ) : null}
        </aside>

        <main className="admin-workspace author-detail-panel">
          {selectedNode ? (
            <WorkspaceDetail
              node={selectedNode}
              detail={detailQuery.data ?? null}
              isLoading={detailQuery.isLoading}
              isError={detailQuery.isError}
              onReturnToDirectory={() => {
                const parent = selectedNode.parent_id ? flatTree.find((node) => node.id === selectedNode.parent_id) : null;
                if (parent) selectNode(parent);
              }}
            />
          ) : (
            <section className="glass status-panel">请选择内容树中的目录或文件。</section>
          )}
        </main>
      </section>
    </section>
  );
}

function TreeList({ nodes, expandedIds, selectedId, onSelect, onToggle }: {
  nodes: AdminTreeNode[];
  expandedIds: Set<string>;
  selectedId: string;
  onSelect: (node: AdminTreeNode) => void;
  onToggle: (nodeId: string) => void;
}) {
  return (
    <div className="admin-tree-list author-tree-list">
      {nodes.map((node) => (
        <TreeNodeRow
          key={node.id}
          node={node}
          depth={0}
          expandedIds={expandedIds}
          selectedId={selectedId}
          onSelect={onSelect}
          onToggle={onToggle}
        />
      ))}
    </div>
  );
}

function TreeNodeRow({ node, depth, expandedIds, selectedId, onSelect, onToggle }: {
  node: AdminTreeNode;
  depth: number;
  expandedIds: Set<string>;
  selectedId: string;
  onSelect: (node: AdminTreeNode) => void;
  onToggle: (nodeId: string) => void;
}) {
  const hasChildren = node.children.length > 0;
  const expanded = expandedIds.has(node.id);
  const selected = selectedId === node.id;
  return (
    <div className="author-tree-node">
      <div className={`tree-row author-tree-row${selected ? ' is-selected' : ''}`} style={{ paddingLeft: `${0.65 + depth * 1.1}rem` }}>
        {node.kind === 'directory' ? (
          <button className="tree-toggle" type="button" aria-label={expanded ? '收起目录' : '展开目录'} onClick={() => onToggle(node.id)}>
            {hasChildren ? (expanded ? '▾' : '▸') : '•'}
          </button>
        ) : <span className="tree-toggle" aria-hidden="true">•</span>}
        <button className="tree-select-button" type="button" onClick={() => onSelect(node)}>
          <span>{node.kind === 'directory' ? '📁' : '📄'} {node.name}</span>
          <small>{node.path} · {node.status === 'published' ? '已发布' : '草稿'}</small>
        </button>
      </div>
      {node.kind === 'directory' && expanded && hasChildren ? (
        <div className="author-tree-children">
          {node.children.map((child) => (
            <TreeNodeRow
              key={child.id}
              node={child}
              depth={depth + 1}
              expandedIds={expandedIds}
              selectedId={selectedId}
              onSelect={onSelect}
              onToggle={onToggle}
            />
          ))}
        </div>
      ) : null}
    </div>
  );
}

function WorkspaceDetail({ node, detail, isLoading, isError, onReturnToDirectory }: {
  node: AdminTreeNode;
  detail: AdminNodeDetail | null;
  isLoading: boolean;
  isError: boolean;
  onReturnToDirectory: () => void;
}) {
  const children = node.children;
  return (
    <section className="glass admin-panel author-workspace-card">
      <div className="panel-heading-row">
        <div>
          <p className="eyebrow">{node.kind === 'directory' ? '目录概览' : '文件工作区'}</p>
          <h2>{node.name}</h2>
          <p className="path-text">URL Path：{node.path}</p>
        </div>
        <span className={`status-pill ${node.status}`}>{node.status === 'published' ? '已发布' : '草稿'}</span>
      </div>

      {node.kind === 'file' ? <button className="glass-button" type="button" onClick={onReturnToDirectory}>返回目录</button> : null}
      {isLoading ? <p className="muted">正在加载工作区详情…</p> : null}
      {isError ? <p className="form-error">工作区详情加载失败。可继续使用左侧内容树。</p> : null}

      {node.kind === 'directory' ? (
        <DirectoryOverview node={node} children={children} />
      ) : (
        <FileOverview node={node} detail={detail} />
      )}
    </section>
  );
}

function DirectoryOverview({ node, children }: { node: AdminTreeNode; children: AdminTreeNode[] }) {
  return (
    <div className="directory-overview">
      <p className="muted">此目录包含 {children.length} 个直接子项。新建与移动操作将在后续工作包中接入。</p>
      {children.length === 0 ? <p className="muted">此目录暂无子项。</p> : null}
      <div className="admin-child-card-grid">
        {children.map((child) => (
          <article className="admin-child-card" key={child.id}>
            <strong>{child.kind === 'directory' ? '📁' : '📄'} {child.name}</strong>
            <span>{child.path}</span>
            <small>{child.status === 'published' ? '已发布' : '草稿'}</small>
          </article>
        ))}
      </div>
      <p className="muted">当前目录：{node.path}</p>
    </div>
  );
}

function FileOverview({ node, detail }: { node: AdminTreeNode; detail: AdminNodeDetail | null }) {
  return (
    <div className="file-overview">
      <div className="admin-tabs" aria-label="文件工作区标签">
        <button className="primary-button" type="button">内容</button>
        <button className="glass-button" type="button">资源</button>
        <button className="glass-button" type="button">设置</button>
      </div>
      <p className="muted">文件状态：{node.status === 'published' ? '已发布' : '草稿'}</p>
      {detail?.content ? <p className="muted">格式：{detail.content.content_format === 'html_document' ? 'HTML Document' : 'Markdown'}</p> : null}
      <p className="muted">编辑、资源和设置操作将在后续工作包中接入；此处先提供稳定的作者工作区骨架。</p>
    </div>
  );
}

function flattenTree(nodes: AdminTreeNode[]): AdminTreeNode[] {
  return nodes.flatMap((node) => [node, ...flattenTree(node.children)]);
}

function expandAncestors(current: Set<string>, selectedId: string, roots: AdminTreeNode[]): Set<string> {
  const next = new Set(current);
  const path = findPathToNode(roots, selectedId);
  for (const node of path) {
    if (node.kind === 'directory') next.add(node.id);
  }
  return next;
}

function findPathToNode(nodes: AdminTreeNode[], selectedId: string, ancestors: AdminTreeNode[] = []): AdminTreeNode[] {
  for (const node of nodes) {
    const path = [...ancestors, node];
    if (node.id === selectedId) return path;
    const childPath = findPathToNode(node.children, selectedId, path);
    if (childPath.length > 0) return childPath;
  }
  return [];
}

function readStoredString(key: string) {
  if (typeof window === 'undefined') return '';
  return window.localStorage.getItem(key) ?? '';
}

function readStoredList(key: string) {
  if (typeof window === 'undefined') return [];
  try {
    const parsed: unknown = JSON.parse(window.localStorage.getItem(key) ?? '[]');
    return Array.isArray(parsed) ? parsed.filter((item): item is string => typeof item === 'string') : [];
  } catch {
    return [];
  }
}
