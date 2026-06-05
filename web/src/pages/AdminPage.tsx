import { FormEvent, useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  createAdminNode,
  deleteAdminNode,
  deleteAsset,
  fetchAdminNode,
  fetchRootDirectory,
  publishFile,
  rebuildSearchIndex,
  refreshEmbedding,
  unpublishFile,
  updateAdminNode,
  uploadAsset,
  upsertFileContent,
  type CreateAdminNodeInput,
} from '../lib/api';
import type { AdminNodeDetail, ContentEntry, ContentFormat, FileAsset, NodeKind } from '../lib/types';

export function AdminPage() {
  const rootQuery = useQuery({ queryKey: ['admin-root-tree'], queryFn: fetchRootDirectory });
  const [selectedId, setSelectedId] = useState('');
  const [detail, setDetail] = useState<AdminNodeDetail | null>(null);
  const [status, setStatus] = useState<string | null>(null);

  const selectedFileId = detail?.node.kind === 'file' ? detail.node.id : selectedId.trim();

  async function loadNode(nodeId = selectedId.trim()) {
    if (!nodeId) {
      setStatus('Enter or select a node id first.');
      return;
    }
    try {
      const loaded = await fetchAdminNode(nodeId);
      setDetail(loaded);
      setSelectedId(loaded.node.id);
      setStatus(`Loaded ${loaded.node.path}.`);
    } catch {
      setStatus('Load failed. Check admin login and node id.');
    }
  }

  async function submitCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const parent = stringValue(form, 'parent_id');
    const kind = stringValue(form, 'kind') as NodeKind;
    const input: CreateAdminNodeInput = {
      parent_id: parent || null,
      kind,
      name: stringValue(form, 'name'),
      slug: stringValue(form, 'slug'),
      sort_order: numberValue(form, 'sort_order'),
      content_format: kind === 'file' ? stringValue(form, 'content_format') as ContentFormat : undefined,
    };
    try {
      const created = await createAdminNode(input);
      setDetail(created);
      setSelectedId(created.node.id);
      setStatus(`Created ${created.node.kind} at ${created.node.path}.`);
      event.currentTarget.reset();
    } catch {
      setStatus('Create failed. Check slug uniqueness, reserved root slugs, parent id, and admin login.');
    }
  }

  async function submitNodeUpdate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!detail) return;
    const form = new FormData(event.currentTarget);
    const nextParent = stringValue(form, 'parent_id');
    const nextSlug = stringValue(form, 'slug');
    const movingPublished = detail.content?.status === 'published' && (nextParent !== (detail.node.parent_id ?? '') || nextSlug !== detail.node.slug);
    if (movingPublished && !window.confirm('This published path change can create redirects. Continue?')) {
      return;
    }
    try {
      const updated = await updateAdminNode(detail.node.id, {
        parent_id: nextParent || null,
        name: stringValue(form, 'name'),
        slug: nextSlug,
        sort_order: numberValue(form, 'sort_order'),
      });
      setDetail(updated);
      setStatus(`Updated ${updated.node.path}.`);
    } catch {
      setStatus('Update failed. Check slug uniqueness, reserved root slugs, and move constraints.');
    }
  }

  async function submitContent(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!detail || detail.node.kind !== 'file') return;
    const form = new FormData(event.currentTarget);
    const contentFormat = stringValue(form, 'content_format') as ContentFormat;
    if (detail.content?.status === 'published' && contentFormat !== detail.content.content_format) {
      setStatus('Published files cannot directly change content_format. Unpublish or create a new File.');
      return;
    }
    try {
      const content = await upsertFileContent(detail.node.id, {
        content_format: contentFormat,
        body_raw: stringValue(form, 'body_raw'),
        body_html: contentFormat === 'html_document' ? stringValue(form, 'body_raw') : null,
        keywords: splitKeywords(stringValue(form, 'keywords')),
      });
      setDetail({ ...detail, content });
      setStatus('Content saved. Embedding status reset for refresh/fallback.');
    } catch {
      setStatus('Save failed. Check content format and admin login.');
    }
  }

  async function changePublishState(action: 'publish' | 'unpublish') {
    if (!detail || detail.node.kind !== 'file') return;
    if (action === 'unpublish' && !window.confirm('Unpublish hides the public file and its assets. Continue?')) return;
    try {
      const content = action === 'publish' ? await publishFile(detail.node.id) : await unpublishFile(detail.node.id);
      setDetail({ ...detail, content });
      setStatus(action === 'publish' ? 'File published.' : 'File unpublished.');
    } catch {
      setStatus(`${action} failed. Ensure file content exists and admin login is valid.`);
    }
  }

  async function removeSelectedNode() {
    if (!detail) return;
    if (!window.confirm(`Delete ${detail.node.path}? Published files/directories with published descendants are protected.`)) return;
    try {
      await deleteAdminNode(detail.node.id);
      setDetail(null);
      setStatus('Node deleted.');
    } catch {
      setStatus('Delete failed. Published content is protected or login expired.');
    }
  }

  async function submitUpload(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    const file = data.get('file');
    if (!(file instanceof File) || !selectedFileId) {
      setStatus('Choose a file and select or enter a File node id.');
      return;
    }
    try {
      const uploaded = await uploadAsset(selectedFileId, file);
      setDetail((current) => current ? { ...current, assets: [uploaded, ...current.assets.filter((asset) => asset.id !== uploaded.id)] } : current);
      setStatus(`Uploaded ${uploaded.filename}.`);
      event.currentTarget.reset();
    } catch {
      setStatus('Upload failed. Check admin login, MIME type, file size, or SVG safety.');
    }
  }

  async function removeAsset(assetId: string) {
    try {
      await deleteAsset(assetId);
      setDetail((current) => current ? { ...current, assets: current.assets.filter((asset) => asset.id !== assetId) } : current);
      setStatus('Asset deleted.');
    } catch {
      setStatus('Delete failed. Check admin login and asset id.');
    }
  }

  async function refreshSelectedEmbedding() {
    if (!detail || detail.node.kind !== 'file') return;
    try {
      const state = await refreshEmbedding(detail.node.id);
      setStatus(state.status === 'failed' ? `Embedding failed and full-text fallback remains active: ${state.error ?? 'unknown error'}` : 'Embedding refreshed.');
    } catch {
      setStatus('Embedding refresh failed.');
    }
  }

  async function rebuildSearch() {
    if (!window.confirm('Rebuild search_text/embeddings for all published files?')) return;
    try {
      await rebuildSearchIndex();
      setStatus('Search rebuild accepted.');
    } catch {
      setStatus('Search rebuild failed.');
    }
  }

  return (
    <section className="page-stack admin-manager-page">
      <section className="glass status-panel admin-hero">
        <p className="eyebrow">ADMIN</p>
        <h1>Tree Manager</h1>
        <p>Create, edit, move, publish, unpublish, upload assets, and refresh hybrid-search embeddings from one Packet I workspace.</p>
        {status ? <p className="muted">{status}</p> : null}
      </section>

      <section className="admin-grid">
        <aside className="glass admin-sidebar">
          <h2>Tree browser</h2>
          {rootQuery.isLoading ? <p className="muted">Loading root…</p> : null}
          {rootQuery.data ? <TreeList entries={rootQuery.data.children} onSelect={(id) => { setSelectedId(id); void loadNode(id); }} /> : null}
          <form className="auth-form compact-form" onSubmit={(event) => { event.preventDefault(); void loadNode(); }}>
            <label>
              Node id
              <input value={selectedId} onChange={(event) => setSelectedId(event.target.value)} placeholder="Directory or File node id" />
            </label>
            <button className="primary-button" type="submit">Load selected node</button>
          </form>
        </aside>

        <main className="admin-workspace">
          <CreateNodePanel parentId={detail?.node.kind === 'directory' ? detail.node.id : detail?.node.parent_id ?? ''} onCreate={submitCreate} />
          {detail ? (
            <>
              <NodeEditor detail={detail} onSubmit={submitNodeUpdate} onDelete={removeSelectedNode} />
              {detail.node.kind === 'file' ? (
                <>
                  <ContentEditor key={`${detail.node.id}:${detail.content?.content_format ?? 'markdown'}`} detail={detail} onSubmit={submitContent} onPublish={() => changePublishState('publish')} onUnpublish={() => changePublishState('unpublish')} onRefreshEmbedding={refreshSelectedEmbedding} onRebuildSearch={rebuildSearch} />
                  <AssetPanel assets={detail.assets} onUpload={submitUpload} onDelete={removeAsset} />
                </>
              ) : null}
            </>
          ) : <section className="glass status-panel">Select or load a node to edit.</section>}
        </main>
      </section>
    </section>
  );
}

function TreeList({ entries, onSelect }: { entries: ContentEntry[]; onSelect: (id: string) => void }) {
  if (entries.length === 0) {
    return <p className="muted">No visible root entries yet. Draft files can still be loaded by id.</p>;
  }
  return (
    <div className="admin-tree-list">
      {entries.map((entry) => (
        <button className="tree-row" key={entry.id} type="button" onClick={() => onSelect(entry.id)}>
          <span>{entry.kind === 'directory' ? '📁' : '📄'} {entry.name}</span>
          <small>{entry.path}</small>
        </button>
      ))}
    </div>
  );
}

function CreateNodePanel({ parentId, onCreate }: { parentId?: string | null; onCreate: (event: FormEvent<HTMLFormElement>) => void }) {
  const [kind, setKind] = useState<NodeKind>('directory');
  return (
    <section className="glass admin-panel">
      <h2>Create Directory / File</h2>
      <form className="admin-form" onSubmit={onCreate}>
        <label>Parent id<input name="parent_id" defaultValue={parentId ?? ''} placeholder="blank for root" /></label>
        <label>Name<input name="name" required placeholder="Research Notes" /></label>
        <label>Slug<input name="slug" required placeholder="research-notes" /></label>
        <label>Sort order<input name="sort_order" type="number" defaultValue="0" /></label>
        <label>Kind<select name="kind" value={kind} onChange={(event) => setKind(event.target.value as NodeKind)}><option value="directory">Directory</option><option value="file">File</option></select></label>
        {kind === 'file' ? <label>Content format<select name="content_format" defaultValue="markdown"><option value="markdown">Markdown</option><option value="html_document">HTML Document</option></select></label> : null}
        <button className="primary-button" type="submit">Create</button>
      </form>
    </section>
  );
}

function NodeEditor({ detail, onSubmit, onDelete }: { detail: AdminNodeDetail; onSubmit: (event: FormEvent<HTMLFormElement>) => void; onDelete: () => void }) {
  return (
    <section className="glass admin-panel">
      <h2>Edit node</h2>
      <p className="path-text">{detail.node.kind} · {detail.node.path}</p>
      <form className="admin-form" onSubmit={onSubmit}>
        <label>Parent id<input name="parent_id" defaultValue={detail.node.parent_id ?? ''} placeholder="blank for root" /></label>
        <label>Name<input name="name" defaultValue={detail.node.name} required /></label>
        <label>Slug<input name="slug" defaultValue={detail.node.slug} required /></label>
        <label>Sort order<input name="sort_order" type="number" defaultValue={detail.node.sort_order} /></label>
        <div className="button-row"><button className="primary-button" type="submit">Save node</button><button className="glass-button danger-button" type="button" onClick={onDelete}>Delete node</button></div>
      </form>
    </section>
  );
}

function ContentEditor({ detail, onSubmit, onPublish, onUnpublish, onRefreshEmbedding, onRebuildSearch }: { detail: AdminNodeDetail; onSubmit: (event: FormEvent<HTMLFormElement>) => void; onPublish: () => void; onUnpublish: () => void; onRefreshEmbedding: () => void; onRebuildSearch: () => void }) {
  const content = detail.content;
  const initialFormat = content?.content_format ?? 'markdown';
  const [format, setFormat] = useState<ContentFormat>(initialFormat);
  const keywords = useMemo(() => content?.keywords.join(', ') ?? '', [content?.keywords]);
  return (
    <section className="glass admin-panel">
      <h2>File editor</h2>
      <p className="muted">Status: {content?.status ?? 'draft'} · Embedding: {content?.embedding_status ?? 'pending'}{content?.embedding_error ? ` · ${content.embedding_error}` : ''}</p>
      <form className="admin-form" onSubmit={onSubmit}>
        <label>Content format<select name="content_format" value={format} onChange={(event) => setFormat(event.target.value as ContentFormat)} disabled={content?.status === 'published'}><option value="markdown">Markdown</option><option value="html_document">HTML Document</option></select></label>
        <label>Keywords<input name="keywords" defaultValue={keywords} placeholder="go, search, notes" /></label>
        <label>{format === 'html_document' ? 'Full HTML document' : 'Markdown body'}<textarea name="body_raw" defaultValue={content?.body_raw ?? ''} rows={14} placeholder={format === 'html_document' ? '<!doctype html>…' : '# Markdown'} /></label>
        <div className="button-row"><button className="primary-button" type="submit">Save content</button><button className="glass-button" type="button" onClick={onPublish}>Publish</button><button className="glass-button" type="button" onClick={onUnpublish}>Unpublish</button><button className="glass-button" type="button" onClick={onRefreshEmbedding}>Refresh embedding</button><button className="glass-button" type="button" onClick={onRebuildSearch}>Rebuild search</button></div>
      </form>
      {format === 'html_document' ? <iframe className="admin-preview-frame" title="HTML preview" sandbox="allow-scripts" srcDoc={content?.body_raw ?? ''} /> : <pre className="markdown-preview">{content?.body_raw ?? 'Markdown preview appears here after save.'}</pre>}
    </section>
  );
}

function AssetPanel({ assets, onUpload, onDelete }: { assets: FileAsset[]; onUpload: (event: FormEvent<HTMLFormElement>) => void; onDelete: (assetId: string) => void }) {
  return (
    <section className="glass admin-panel admin-assets-panel">
      <h2>Assets</h2>
      <form className="auth-form asset-upload-form" onSubmit={onUpload}>
        <input name="file" type="file" required />
        <button className="primary-button" type="submit">Upload asset</button>
      </form>
      <div className="asset-list admin-asset-list">
        {assets.map((asset) => (
          <div className="asset-link" key={asset.id}>
            <span>{asset.filename}</span>
            <small>{asset.mime_type} · {formatBytes(asset.size_bytes)}</small>
            <a className="glass-button" href={asset.public_url}>Open</a>
            <button className="glass-button" type="button" onClick={() => onDelete(asset.id)}>Delete</button>
          </div>
        ))}
      </div>
    </section>
  );
}

function stringValue(form: FormData, key: string) {
  return String(form.get(key) ?? '').trim();
}

function numberValue(form: FormData, key: string) {
  const raw = stringValue(form, key);
  return raw ? Number.parseInt(raw, 10) || 0 : 0;
}

function splitKeywords(value: string) {
  return value.split(',').map((keyword) => keyword.trim()).filter(Boolean);
}

function formatBytes(bytes: number) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 102.4) / 10} KB`;
  return `${Math.round(bytes / 1024 / 102.4) / 10} MB`;
}
