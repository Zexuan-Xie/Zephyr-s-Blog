import { FormEvent, useState } from 'react';
import { deleteAsset, uploadAsset } from '../lib/api';
import type { FileAsset } from '../lib/types';

export function AdminPage() {
  const [fileId, setFileId] = useState('');
  const [assets, setAssets] = useState<FileAsset[]>([]);
  const [status, setStatus] = useState<string | null>(null);

  async function submitUpload(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    const file = data.get('file');
    if (!(file instanceof File) || !fileId.trim()) {
      setStatus('Choose a file and provide a File node id.');
      return;
    }
    try {
      const uploaded = await uploadAsset(fileId.trim(), file);
      setAssets((current) => [uploaded, ...current.filter((asset) => asset.id !== uploaded.id)]);
      setStatus(`Uploaded ${uploaded.filename}.`);
      event.currentTarget.reset();
    } catch {
      setStatus('Upload failed. Check admin login, MIME type, file size, or SVG safety.');
    }
  }

  async function removeAsset(assetId: string) {
    try {
      await deleteAsset(assetId);
      setAssets((current) => current.filter((asset) => asset.id !== assetId));
      setStatus('Asset deleted.');
    } catch {
      setStatus('Delete failed. Check admin login and asset id.');
    }
  }

  return (
    <section className="glass status-panel admin-assets-panel">
      <p className="eyebrow">ADMIN</p>
      <h1>Asset Manager foundation</h1>
      <p>
        Packet G provides per-file asset upload and deletion while the full Tree Manager remains reserved for Packet I.
        Use a File node id from the backend/admin API, then upload an allowed asset.
      </p>
      <form className="auth-form asset-upload-form" onSubmit={submitUpload}>
        <input value={fileId} onChange={(event) => setFileId(event.target.value)} placeholder="File node id" required />
        <input name="file" type="file" required />
        <button className="primary-button" type="submit">Upload asset</button>
      </form>
      {status ? <p className="muted">{status}</p> : null}
      <div className="asset-list admin-asset-list">
        {assets.map((asset) => (
          <div className="asset-link" key={asset.id}>
            <span>{asset.filename}</span>
            <small>{asset.mime_type} · {formatBytes(asset.size_bytes)}</small>
            <button className="glass-button" type="button" onClick={() => removeAsset(asset.id)}>Delete</button>
          </div>
        ))}
      </div>
    </section>
  );
}

function formatBytes(bytes: number) {
  if (bytes < 1024) {
    return `${bytes} B`;
  }
  if (bytes < 1024 * 1024) {
    return `${Math.round(bytes / 102.4) / 10} KB`;
  }
  return `${Math.round(bytes / 1024 / 102.4) / 10} MB`;
}
