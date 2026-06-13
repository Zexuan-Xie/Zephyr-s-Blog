import {
  FormEvent,
  useEffect,
  useMemo,
  useState,
} from "react";
import {
  FileText,
  Folder,
  GripVertical,
  Plus,
  Upload,
} from "lucide-react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Navigate, useSearchParams } from "react-router-dom";
import {
  ApiError,
  createAdminNode,
  deleteAdminNode,
  deleteAsset,
  fetchAdminNode,
  fetchAdminTree,
  fetchCurrentUser,
  moveAdminNode,
  previewAdminMove,
  publishFile,
  reorderAdminChildren,
  type CreateAdminNodeInput,
  unpublishFile,
  updateAdminNode,
  uploadAsset,
  upsertFileContent,
} from "../lib/api";
import { getToken } from "../lib/auth";
import type {
  AdminNodeDetail,
  AdminTreeNode,
  ContentFormat,
  MovePreviewResponse,
  NodeKind,
} from "../lib/types";

const selectionStorageKey = "xlab-author-workspace:selected-node";
const expandedStorageKey = "xlab-author-workspace:expanded-directories";

type FileWorkspaceTab = "content" | "assets" | "settings";

export function AdminPage({ onLogout }: { onLogout: () => void }) {
  const token = getToken();
  const [searchParams] = useSearchParams();
  const requestedTarget =
    searchParams.get("target") ??
    searchParams.get("node") ??
    searchParams.get("select") ??
    "";
  const [selectedId, setSelectedId] = useState(
    () => requestedTarget || readStoredString(selectionStorageKey),
  );
  const [expandedIds, setExpandedIds] = useState<Set<string>>(
    () => new Set(readStoredList(expandedStorageKey)),
  );
  const [statusMessage, setStatusMessage] = useState<string | null>(null);
  const [rootCreateOpen, setRootCreateOpen] = useState(false);
  const queryClient = useQueryClient();

  const viewerQuery = useQuery({
    queryKey: ["auth", "me", "admin"],
    queryFn: fetchCurrentUser,
    enabled: Boolean(token),
    retry: false,
  });
  const adminTreeQuery = useQuery({
    queryKey: ["admin", "content-tree"],
    queryFn: fetchAdminTree,
    enabled: Boolean(token) && viewerQuery.data?.role === "admin",
  });

  const flatTree = useMemo(
    () => flattenTree(adminTreeQuery.data?.roots ?? []),
    [adminTreeQuery.data],
  );
  const directoryOptions = flatTree.filter((node) => node.kind === "directory");
  const selectedNode =
    flatTree.find((node) => node.id === selectedId) ??
    flatTree.find((node) => node.kind === "directory") ??
    flatTree[0] ??
    null;
  const effectiveSelectedId = selectedNode?.id ?? selectedId;
  const visibleExpandedIds =
    selectedNode && adminTreeQuery.data
      ? expandAncestors(expandedIds, selectedNode.id, adminTreeQuery.data.roots)
      : expandedIds;
  const detailQuery = useQuery({
    queryKey: ["admin", "node-detail", effectiveSelectedId],
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
    window.localStorage.setItem(
      expandedStorageKey,
      JSON.stringify([...expandedIds]),
    );
  }, [expandedIds]);

  if (!token) {
    return <Navigate to="/login?return_to=%2Fadmin" replace />;
  }
  if (viewerQuery.isLoading) {
    return <section className="glass status-panel">Checking author access…</section>;
  }
  if (viewerQuery.isError || viewerQuery.data?.role !== "admin") {
    return <Navigate to="/login?return_to=%2Fadmin" replace />;
  }

  function selectNode(node: AdminTreeNode) {
    setSelectedId(node.id);
    if (node.kind === "directory") {
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

  async function refreshPublicDirectoryCache() {
    await queryClient.invalidateQueries({ queryKey: ["tree"] });
    await queryClient.invalidateQueries({ queryKey: ["admin", "content-tree"] });
  }

  async function refreshWorkspace() {
    await adminTreeQuery.refetch();
    await detailQuery.refetch();
    await refreshPublicDirectoryCache();
  }

  async function submitCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedNode || selectedNode.kind !== "directory") {
      setStatusMessage("Pick a directory first.");
      return;
    }

    const createForm = event.currentTarget;
    const form = new FormData(createForm);
    const kind = stringValue(form, "kind") as NodeKind;
    const input: CreateAdminNodeInput = {
      parent_id: selectedNode.id,
      kind,
      name: stringValue(form, "name"),
      content_format:
        kind === "file"
          ? (stringValue(form, "content_format") as ContentFormat)
          : undefined,
    };

    let created: AdminNodeDetail;
    try {
      created = await createAdminNode(input);
    } catch (error) {
      setStatusMessage(formatAdminCreateError(error));
      return;
    }

    setSelectedId(created.node.id);
    setExpandedIds(
      (current) => new Set([...current, selectedNode.id, created.node.id]),
    );
    setStatusMessage(
      `${created.node.kind === "directory" ? "Directory" : "File"} created: ${created.node.path}`,
    );
    createForm.reset();
    await adminTreeQuery.refetch();
    await refreshPublicDirectoryCache();
  }

  async function submitRootCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const createForm = event.currentTarget;
    const form = new FormData(createForm);
    const kind = stringValue(form, "kind") as NodeKind;
    const input: CreateAdminNodeInput = {
      parent_id: null,
      kind,
      name: stringValue(form, "name"),
      content_format:
        kind === "file"
          ? (stringValue(form, "content_format") as ContentFormat)
          : undefined,
    };

    let created: AdminNodeDetail;
    try {
      created = await createAdminNode(input);
    } catch (error) {
      setStatusMessage(formatAdminCreateError(error));
      return;
    }

    setSelectedId(created.node.id);
    if (created.node.kind === "directory") {
      setExpandedIds((current) => new Set([...current, created.node.id]));
    }
    setStatusMessage(
      `${created.node.kind === "directory" ? "Directory" : "File"} created: ${created.node.path}`,
    );
    setRootCreateOpen(false);
    createForm.reset();
    await adminTreeQuery.refetch();
    await refreshPublicDirectoryCache();
  }

  async function deleteDirectory(node: AdminTreeNode) {
    if (node.children.length > 0) {
      setStatusMessage("This directory is not empty. Move or remove its items first.");
      return;
    }

    try {
      await deleteAdminNode(node.id);
      setStatusMessage(`Directory deleted: ${node.path}`);
      if (node.parent_id) setSelectedId(node.parent_id);
      await refreshWorkspace();
    } catch (error) {
      setStatusMessage(
        formatAdminActionError(error, "Delete failed. Check the item state and try again."),
      );
    }
  }

  async function reorderChildren(parent: AdminTreeNode, childIds: string[]) {
    try {
      await reorderAdminChildren(parent.id, {
        child_ids: childIds,
        expected_version: 0,
      });
      setStatusMessage("Order saved.");
      await refreshWorkspace();
    } catch (error) {
      setStatusMessage(
        formatAdminActionError(error, "Order save failed. Refresh and try again."),
      );
    }
  }

  function logoutAuthor() {
    onLogout();
    window.location.assign("/");
  }

  return (
    <section className="page-stack admin-manager-page author-workspace-page">
      <section className="glass status-panel admin-hero author-workspace-hero">
        <p className="eyebrow">Author Workspace</p>
        <h1>Tree</h1>
        <p>
          Create, write, publish, and reorder your content with only the controls you need.
        </p>
        <div className="button-row">
          <button
            className="glass-button"
            type="button"
            onClick={() => window.scrollTo({ top: 0, behavior: "smooth" })}
          >
            Top
          </button>
          <button className="glass-button" type="button" onClick={logoutAuthor}>
            Sign out
          </button>
        </div>
      </section>

      <section className="admin-grid author-workspace-grid">
        <aside
          className="glass admin-sidebar author-tree-panel"
          aria-label="Tree"
        >
          <div className="panel-heading-row">
            <div>
              <p className="eyebrow">Content Tree</p>
              <h2>Content Tree</h2>
            </div>
            <div className="button-row compact-actions">
              <button
                className="glass-button"
                type="button"
                onClick={() => setRootCreateOpen((open) => !open)}
              >
                New root
              </button>
              <button
                className="glass-button"
                type="button"
                onClick={() => adminTreeQuery.refetch()}
              >
                Refresh
              </button>
            </div>
          </div>
          {rootCreateOpen ? (
            <RootCreatePanel
              onCreate={submitRootCreate}
              onCancel={() => setRootCreateOpen(false)}
            />
          ) : null}
          {adminTreeQuery.isLoading ? (
            <p className="muted">Loading your tree…</p>
          ) : null}
          {adminTreeQuery.isError ? (
            <p className="form-error">Tree failed to load. Refresh or sign in again.</p>
          ) : null}
          {adminTreeQuery.data && adminTreeQuery.data.roots.length === 0 ? (
            <p className="muted">Nothing here yet. Create your first item on the right.</p>
          ) : null}
          {adminTreeQuery.data ? (
            <TreeList
              nodes={adminTreeQuery.data.roots}
              expandedIds={visibleExpandedIds}
              selectedId={effectiveSelectedId}
              onSelect={selectNode}
              onToggle={toggleDirectory}
              onReorderChildren={reorderChildren}
            />
          ) : null}
        </aside>

        <main className="admin-workspace author-detail-panel">
          {selectedNode ? (
            <WorkspaceDetail
              node={selectedNode}
              detail={detailQuery.data ?? null}
              directoryOptions={directoryOptions}
              isLoading={detailQuery.isLoading}
              isError={detailQuery.isError}
              statusMessage={statusMessage}
              onCreate={submitCreate}
              onCancelCreate={() => setStatusMessage(null)}
              onDeleteDirectory={deleteDirectory}
              onFeedback={setStatusMessage}
              onRefresh={refreshWorkspace}
              onReturnToDirectory={() => {
                const parent = selectedNode.parent_id
                  ? flatTree.find((node) => node.id === selectedNode.parent_id)
                  : null;
                if (parent) selectNode(parent);
              }}
            />
          ) : (
            <EmptyRootWorkspace
              statusMessage={statusMessage}
              onCreate={submitRootCreate}
              onCancelCreate={() => setStatusMessage(null)}
            />
          )}
        </main>
      </section>
    </section>
  );
}

function EmptyRootWorkspace({
  statusMessage,
  onCreate,
  onCancelCreate,
}: {
  statusMessage: string | null;
  onCreate: (event: FormEvent<HTMLFormElement>) => void;
  onCancelCreate: () => void;
}) {
  return (
    <section className="glass admin-panel author-workspace-card">
      <div className="panel-heading-row">
        <div>
          <p className="eyebrow">Start here</p>
          <h2>Create your first item</h2>
          <p className="path-text">
            Your tree is empty. Start with a directory, or create a file at the root.
          </p>
        </div>
      </div>
      {statusMessage ? <p className="muted">{statusMessage}</p> : null}
      <RootCreatePanel onCreate={onCreate} onCancel={onCancelCreate} />
    </section>
  );
}

function RootCreatePanel({
  onCreate,
  onCancel,
}: {
  onCreate: (event: FormEvent<HTMLFormElement>) => void;
  onCancel: () => void;
}) {
  const [kind, setKind] = useState<NodeKind>("directory");
  const previewName = kind === "directory" ? "New directory" : "New file";

  return (
    <section className="nested-create-panel root-create-panel" aria-label="Create root item">
      <h3>{kind === "directory" ? "New root directory" : "New root file"}</h3>
      <form className="admin-form root-create-form" onSubmit={onCreate}>
        <input type="hidden" name="kind" value={kind} />
        <label>
          Type
          <select
            value={kind}
            onChange={(event) => setKind(event.target.value as NodeKind)}
          >
            <option value="directory">Directory</option>
            <option value="file">File</option>
          </select>
        </label>
        <label>
          Name
          <input name="name" required placeholder={previewName} />
        </label>
        {kind === "file" ? (
          <label>
            Format
            <select name="content_format" defaultValue="markdown">
              <option value="markdown">Markdown</option>
              <option value="html_document">HTML</option>
            </select>
          </label>
        ) : null}
        <label>
          URL Path preview
          <input readOnly value={`/${previewName}`} />
        </label>
        <div className="button-row">
          <button className="primary-button" type="submit">
            Create
          </button>
          <button className="glass-button" type="button" onClick={onCancel}>
            Cancel
          </button>
        </div>
      </form>
    </section>
  );
}

function TreeList({
  nodes,
  expandedIds,
  selectedId,
  onSelect,
  onToggle,
  onReorderChildren,
}: {
  nodes: AdminTreeNode[];
  expandedIds: Set<string>;
  selectedId: string;
  onSelect: (node: AdminTreeNode) => void;
  onToggle: (nodeId: string) => void;
  onReorderChildren: (parent: AdminTreeNode, childIds: string[]) => void;
}) {
  const [draggedNode, setDraggedNode] = useState<{
    id: string;
    parentId: string | null;
  } | null>(null);
  const [dragOverNodeId, setDragOverNodeId] = useState<string | null>(null);

  function reorderWithinParent(
    parent: AdminTreeNode | null,
    siblings: AdminTreeNode[],
    targetId: string,
  ) {
    if (!parent || !draggedNode || draggedNode.parentId !== parent.id) return;
    if (draggedNode.id === targetId) return;
    const childIds = siblings.map((child) => child.id);
    const from = childIds.indexOf(draggedNode.id);
    const to = childIds.indexOf(targetId);
    if (from < 0 || to < 0) return;
    const [moved] = childIds.splice(from, 1);
    childIds.splice(to, 0, moved);
    onReorderChildren(parent, childIds);
  }

  return (
    <div className="admin-tree-list author-tree-list" aria-label="Draggable content tree">
      {nodes.map((node) => (
        <TreeNodeRow
          key={node.id}
          node={node}
          parent={null}
          siblings={nodes}
          depth={0}
          expandedIds={expandedIds}
          selectedId={selectedId}
          draggedNode={draggedNode}
          dragOverNodeId={dragOverNodeId}
          onSelect={onSelect}
          onToggle={onToggle}
          onDragStartNode={setDraggedNode}
          onDragOverNode={setDragOverNodeId}
          onDropNode={reorderWithinParent}
          onDragEnd={() => {
            setDraggedNode(null);
            setDragOverNodeId(null);
          }}
        />
      ))}
    </div>
  );
}

function TreeNodeRow({
  node,
  parent,
  siblings,
  depth,
  expandedIds,
  selectedId,
  draggedNode,
  dragOverNodeId,
  onSelect,
  onToggle,
  onDragStartNode,
  onDragOverNode,
  onDropNode,
  onDragEnd,
}: {
  node: AdminTreeNode;
  parent: AdminTreeNode | null;
  siblings: AdminTreeNode[];
  depth: number;
  expandedIds: Set<string>;
  selectedId: string;
  draggedNode: { id: string; parentId: string | null } | null;
  dragOverNodeId: string | null;
  onSelect: (node: AdminTreeNode) => void;
  onToggle: (nodeId: string) => void;
  onDragStartNode: (node: { id: string; parentId: string | null }) => void;
  onDragOverNode: (nodeId: string | null) => void;
  onDropNode: (
    parent: AdminTreeNode | null,
    siblings: AdminTreeNode[],
    targetId: string,
  ) => void;
  onDragEnd: () => void;
}) {
  const hasChildren = node.children.length > 0;
  const expanded = expandedIds.has(node.id);
  const selected = selectedId === node.id;
  const canDrag = Boolean(parent);
  const canDropHere = Boolean(
    parent && draggedNode && draggedNode.parentId === parent.id,
  );
  const depthClass = `tree-depth-${Math.min(depth, 3)}`;
  return (
    <div className="author-tree-node">
      <div
        className={`tree-row author-tree-row ${depthClass}${selected ? " is-selected" : ""}${draggedNode?.id === node.id ? " is-dragging" : ""}${dragOverNodeId === node.id && canDropHere ? " is-drop-target" : ""}`}
        draggable={canDrag}
        onDragStart={(event) => {
          if (!canDrag) return;
          event.dataTransfer.effectAllowed = "move";
          event.dataTransfer.setData("text/plain", node.id);
          onDragStartNode({ id: node.id, parentId: parent?.id ?? null });
        }}
        onDragEnter={() => {
          if (canDropHere) onDragOverNode(node.id);
        }}
        onDragOver={(event) => {
          if (!canDropHere) return;
          event.preventDefault();
          event.dataTransfer.dropEffect = "move";
          onDragOverNode(node.id);
        }}
        onDragLeave={() => onDragOverNode(null)}
        onDrop={(event) => {
          event.preventDefault();
          onDropNode(parent, siblings, node.id);
          onDragEnd();
        }}
        onDragEnd={onDragEnd}
        style={{ paddingLeft: `${0.55 + depth * 0.95}rem` }}
      >
        <span className="tree-drag-handle" aria-hidden="true">
          {canDrag ? <GripVertical size={14} /> : null}
        </span>
        {node.kind === "directory" ? (
          <button
            className="tree-toggle"
            type="button"
            aria-label={expanded ? "Collapse directory" : "Expand directory"}
            onClick={() => onToggle(node.id)}
          >
            {hasChildren ? (expanded ? "▾" : "▸") : "•"}
          </button>
        ) : (
          <span className="tree-toggle" aria-hidden="true">
            •
          </span>
        )}
        <button
          className="tree-select-button"
          type="button"
          onClick={() => onSelect(node)}
        >
          <span>
            {node.kind === "directory" ? "Directory" : "File"} {node.name}
          </span>
          <small>
            {node.path} · {node.status === "published" ? "Live" : "Draft"}
          </small>
        </button>
      </div>
      {node.kind === "directory" && expanded && hasChildren ? (
        <div className="author-tree-children">
          {node.children.map((child) => (
            <TreeNodeRow
              key={child.id}
              node={child}
              parent={node}
              siblings={node.children}
              depth={depth + 1}
              expandedIds={expandedIds}
              selectedId={selectedId}
              draggedNode={draggedNode}
              dragOverNodeId={dragOverNodeId}
              onSelect={onSelect}
              onToggle={onToggle}
              onDragStartNode={onDragStartNode}
              onDragOverNode={onDragOverNode}
              onDropNode={onDropNode}
              onDragEnd={onDragEnd}
            />
          ))}
        </div>
      ) : null}
    </div>
  );
}

function WorkspaceDetail({
  node,
  detail,
  directoryOptions,
  isLoading,
  isError,
  statusMessage,
  onCreate,
  onCancelCreate,
  onDeleteDirectory,
  onFeedback,
  onRefresh,
  onReturnToDirectory,
}: {
  node: AdminTreeNode;
  detail: AdminNodeDetail | null;
  directoryOptions: AdminTreeNode[];
  isLoading: boolean;
  isError: boolean;
  statusMessage: string | null;
  onCreate: (event: FormEvent<HTMLFormElement>) => void;
  onCancelCreate: () => void;
  onDeleteDirectory: (node: AdminTreeNode) => void;
  onFeedback: (message: string | null) => void;
  onRefresh: () => Promise<void>;
  onReturnToDirectory: () => void;
}) {
  const children = node.children;
  return (
    <section className="glass admin-panel author-workspace-card">
      <div className="panel-heading-row">
        <div>
          <p className="eyebrow">
            {node.kind === "directory" ? "Directory" : "File"}
          </p>
          <h2>{node.name}</h2>
          <p className="path-text">URL Path：{node.path}</p>
        </div>
        <span className={`status-pill ${node.status}`}>
          {node.status === "published" ? "Live" : "Draft"}
        </span>
      </div>

      {node.kind === "file" ? (
        <button
          className="glass-button"
          type="button"
          onClick={onReturnToDirectory}
        >
          Back to directory
        </button>
      ) : null}
      {statusMessage ? <p className="muted">{statusMessage}</p> : null}
      {isLoading ? <p className="muted">Loading details…</p> : null}
      {isError ? (
        <p className="form-error">Details failed to load. You can still use the tree.</p>
      ) : null}

      {node.kind === "directory" ? (
        <DirectoryOverview
          node={node}
          children={children}
          onCreate={onCreate}
          onCancelCreate={onCancelCreate}
          onDeleteDirectory={onDeleteDirectory}
        />
      ) : (
        <FileOverview
          node={node}
          detail={detail}
          directoryOptions={directoryOptions}
          onFeedback={onFeedback}
          onRefresh={onRefresh}
        />
      )}
    </section>
  );
}

function DirectoryOverview({
  node,
  children,
  onCreate,
  onCancelCreate,
  onDeleteDirectory,
}: {
  node: AdminTreeNode;
  children: AdminTreeNode[];
  onCreate: (event: FormEvent<HTMLFormElement>) => void;
  onCancelCreate: () => void;
  onDeleteDirectory: (node: AdminTreeNode) => void;
}) {
  const [kind, setKind] = useState<NodeKind>("directory");
  const previewName = kind === "directory" ? "New directory" : "New file";
  const previewPath = `${node.path.replace(/\/$/, "")}/${previewName}`.replace(
    /^$/,
    "/",
  );

  return (
    <div className="directory-overview">
      <div className="workspace-action-strip" aria-label="Create content">
        <div>
          <p className="eyebrow">Create</p>
          <h3>{kind === "directory" ? "New directory" : "New file"}</h3>
        </div>
        <div className="segmented-control" aria-label="Choose item type">
          <button
            className={kind === "directory" ? "is-active" : ""}
            type="button"
            onClick={() => setKind("directory")}
            aria-pressed={kind === "directory"}
          >
            <Folder size={16} aria-hidden="true" />
            Directory
          </button>
          <button
            className={kind === "file" ? "is-active" : ""}
            type="button"
            onClick={() => setKind("file")}
            aria-pressed={kind === "file"}
          >
            <FileText size={16} aria-hidden="true" />
            File
          </button>
        </div>
      </div>
      <section className="nested-create-panel compact-create-panel" aria-label="Create item">
        <h3>{kind === "directory" ? "Directory" : "File"}</h3>
        <form className="admin-form" onSubmit={onCreate}>
          <input type="hidden" name="kind" value={kind} />
          <label>
            Name
            <input name="name" required placeholder={previewName} />
          </label>
          {kind === "file" ? (
            <label>
              Format
              <select name="content_format" defaultValue="markdown">
                <option value="markdown">Markdown</option>
                <option value="html_document">HTML</option>
              </select>
            </label>
          ) : null}
          <label>
            URL Path preview
            <input readOnly value={previewPath} />
          </label>
          <div className="button-row">
            <button className="primary-button" type="submit">
              <Plus size={16} aria-hidden="true" />
              Create
            </button>
            <button
              className="glass-button"
              type="button"
              onClick={onCancelCreate}
            >
              Cancel
            </button>
          </div>
        </form>
      </section>
      <section className="danger-zone" aria-label="Directory danger actions">
        <h3>Danger</h3>
        <p className="muted">
          Delete is blocked while items remain inside.
        </p>
        <button
          className="glass-button danger-button"
          type="button"
          onClick={() => onDeleteDirectory(node)}
        >
          {children.length > 0 ? "Not empty" : "Delete directory"}
        </button>
      </section>
      <p className="muted">Current path: {node.path}</p>
    </div>
  );
}

function FileOverview({
  node,
  detail,
  directoryOptions,
  onFeedback,
  onRefresh,
}: {
  node: AdminTreeNode;
  detail: AdminNodeDetail | null;
  directoryOptions: AdminTreeNode[];
  onFeedback: (message: string | null) => void;
  onRefresh: () => Promise<void>;
}) {
  const [activeTab, setActiveTab] = useState<FileWorkspaceTab>("content");
  const [movePreview, setMovePreview] = useState<MovePreviewResponse | null>(
    null,
  );
  const [moveDestinationId, setMoveDestinationId] = useState<string | null>(null);
  const contentFormat =
    detail?.content?.content_format ?? node.content_format ?? "markdown";
  const bodyRaw = detail?.content?.body_raw ?? "";
  const keywords = detail?.content?.keywords.join(", ") ?? "";
  const availableDestinations = directoryOptions.filter(
    (directory) => directory.id !== node.id,
  );

  async function submitContent(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    try {
      await upsertFileContent(node.id, {
        content_format: stringValue(form, "content_format") as ContentFormat,
        body_raw: stringValue(form, "body_raw"),
        keywords: stringValue(form, "keywords")
          .split(",")
          .map((item) => item.trim())
          .filter(Boolean),
      });
      onFeedback("Saved.");
      await onRefresh();
    } catch (error) {
      onFeedback(formatAdminActionError(error, "Save failed. Check the content and try again."));
    }
  }

  async function togglePublish(nextStatus: "draft" | "published") {
    try {
      if (nextStatus === "published") {
        await publishFile(node.id);
        onFeedback("Published.");
      } else {
        await unpublishFile(node.id);
        onFeedback("Unpublished. This file is now a draft.");
      }
      await onRefresh();
    } catch (error) {
      onFeedback(
        formatAdminActionError(error, "Publish state update failed. Try again."),
      );
    }
  }

  async function submitAsset(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const file = form.get("asset");
    if (!(file instanceof File)) {
      onFeedback("Choose a file to upload.");
      return;
    }
    try {
      await uploadAsset(node.id, file);
      onFeedback(`Asset uploaded: ${file.name}`);
      event.currentTarget.reset();
      await onRefresh();
    } catch (error) {
      onFeedback(
        formatAdminActionError(error, "Upload failed. Check the file and try again."),
      );
    }
  }

  async function removeAsset(assetId: string) {
    try {
      await deleteAsset(assetId);
      onFeedback("Asset deleted.");
      await onRefresh();
    } catch (error) {
      onFeedback(formatAdminActionError(error, "Delete asset failed. Try again."));
    }
  }

  async function submitSettings(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    try {
      await updateAdminNode(node.id, {
        name: stringValue(form, "name"),
        url_path: stringValue(form, "url_path"),
      });
      onFeedback("Settings saved.");
      await onRefresh();
    } catch (error) {
      onFeedback(
        formatAdminActionError(error, "Settings save failed. Check the URL Path."),
      );
    }
  }

  async function submitMovePreview(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const newParentId = stringValue(form, "new_parent_id");
    try {
      const preview = await previewAdminMove(node.id, {
        new_parent_id: newParentId || null,
        expected_version: 0,
      });
      setMovePreview(preview);
      setMoveDestinationId(newParentId || null);
      onFeedback(`Move preview: ${preview.destination_path}`);
    } catch (error) {
      onFeedback(formatAdminActionError(error, "Move preview failed. Pick another directory."));
    }
  }

  async function commitMove() {
    try {
      await moveAdminNode(node.id, {
        new_parent_id: moveDestinationId,
        expected_version: 0,
      });
      onFeedback("Moved.");
      setMovePreview(null);
      setMoveDestinationId(null);
      await onRefresh();
    } catch (error) {
      onFeedback(formatAdminActionError(error, "Move failed. Preview again."));
    }
  }

  async function deleteFile() {
    try {
      await deleteAdminNode(node.id);
      onFeedback(`File deleted: ${node.path}`);
      await onRefresh();
    } catch (error) {
      onFeedback(
        formatAdminActionError(error, "Live files cannot be deleted. Unpublish first."),
      );
    }
  }

  return (
    <div className="file-overview">
      <div className="admin-tabs" aria-label="File tabs">
        <button
          className={
            activeTab === "content" ? "primary-button" : "glass-button"
          }
          type="button"
          onClick={() => setActiveTab("content")}
        >
          Write
        </button>
        <button
          className={activeTab === "assets" ? "primary-button" : "glass-button"}
          type="button"
          onClick={() => setActiveTab("assets")}
        >
          Assets
        </button>
        <button
          className={
            activeTab === "settings" ? "primary-button" : "glass-button"
          }
          type="button"
          onClick={() => setActiveTab("settings")}
        >
          Settings
        </button>
      </div>
      <p className="muted">
        Status: {node.status === "published" ? "Live" : "Draft"}
      </p>

      {activeTab === "content" ? (
        <section className="workspace-tab-panel" aria-label="Write">
          <div className="button-row">
            {node.status === "draft" ? (
              <button
                className="primary-button"
                type="button"
                onClick={() => togglePublish("published")}
              >
                Publish
              </button>
            ) : null}
            {node.status === "published" ? (
              <button
                className="glass-button"
                type="button"
                onClick={() => togglePublish("draft")}
              >
                Unpublish
              </button>
            ) : null}
          </div>
          <form className="admin-form" onSubmit={submitContent}>
            <label>
              Format
              <select name="content_format" defaultValue={contentFormat}>
                <option value="markdown">Markdown</option>
                <option value="html_document">HTML</option>
              </select>
            </label>
            <label>
              Keywords
              <input
                name="keywords"
                defaultValue={keywords}
                placeholder="comma separated"
              />
            </label>
            <label>
              Body
              <textarea
                name="body_raw"
                defaultValue={bodyRaw}
                rows={12}
                placeholder="Write your draft here"
              />
            </label>
            <div className="button-row">
              <button className="primary-button" type="submit">
                Save
              </button>
            </div>
          </form>
        </section>
      ) : null}

      {activeTab === "assets" ? (
        <section className="workspace-tab-panel" aria-label="Assets">
          <form className="admin-form" onSubmit={submitAsset}>
            <label>
              Asset
              <input name="asset" type="file" />
            </label>
            <div className="button-row">
              <button className="primary-button" type="submit">
                <Upload size={16} aria-hidden="true" />
                Upload
              </button>
            </div>
          </form>
          {detail?.assets.length ? (
            <div className="admin-asset-list">
              {detail.assets.map((asset) => (
                <article className="asset-link" key={asset.id}>
                  <span>{asset.filename}</span>
                  <small>
                    {asset.mime_type} · {formatBytes(asset.size_bytes)}
                  </small>
                  <a
                    className="glass-button"
                    href={asset.public_url}
                    target="_blank"
                    rel="noreferrer"
                  >
                    Open
                  </a>
                  <button
                    className="glass-button danger-button"
                    type="button"
                    onClick={() => removeAsset(asset.id)}
                  >
                    Delete
                  </button>
                </article>
              ))}
            </div>
          ) : (
            <p className="muted">No assets yet.</p>
          )}
        </section>
      ) : null}

      {activeTab === "settings" ? (
        <section className="workspace-tab-panel" aria-label="Settings">
          <form className="admin-form" onSubmit={submitSettings}>
            <label>
              Name
              <input name="name" defaultValue={node.name} required />
            </label>
            <label>
              URL Path
              <input name="url_path" defaultValue={node.path} />
            </label>
            <div className="button-row">
              <button className="primary-button" type="submit">
                Save settings
              </button>
            </div>
          </form>

          <section className="nested-create-panel" aria-label="Move">
            <h3>Move</h3>
            <form className="admin-form" onSubmit={submitMovePreview}>
              <label>
                Directory Picker
                <select
                  name="new_parent_id"
                  defaultValue={node.parent_id ?? ""}
                >
                  <option value="">Root</option>
                  {availableDestinations.map((directory) => (
                    <option value={directory.id} key={directory.id}>
                      {directory.path}
                    </option>
                  ))}
                </select>
              </label>
              <div className="button-row">
                <button className="glass-button" type="submit">
                  Preview move
                </button>
              </div>
            </form>
            {movePreview ? (
              <div className="move-preview-panel">
                <p>New path: {movePreview.destination_path}</p>
                <p>Affected paths: {movePreview.affected_paths.length || 0}</p>
                {movePreview.redirects.length > 0 ? (
                  <p>Creates {movePreview.redirects.length} public redirects.</p>
                ) : null}
                {movePreview.blocked_reasons.length > 0 ? (
                  <p className="form-error">
                    Blocked: {movePreview.blocked_reasons.join("，")}
                  </p>
                ) : null}
                <button
                  className="primary-button"
                  type="button"
                  disabled={movePreview.blocked_reasons.length > 0}
                  onClick={commitMove}
                >
                  Move here
                </button>
              </div>
            ) : null}
          </section>

          <section className="danger-zone" aria-label="Danger">
            <h3>Danger</h3>
            <p className="muted">
              Live files are protected. Unpublish before deleting.
            </p>
            <button
              className="glass-button danger-button"
              type="button"
              onClick={deleteFile}
            >
              Delete file
            </button>
          </section>
        </section>
      ) : null}
    </div>
  );
}

function flattenTree(nodes: AdminTreeNode[]): AdminTreeNode[] {
  return nodes.flatMap((node) => [node, ...flattenTree(node.children)]);
}

function expandAncestors(
  current: Set<string>,
  selectedId: string,
  roots: AdminTreeNode[],
): Set<string> {
  const next = new Set(current);
  const path = findPathToNode(roots, selectedId);
  for (const node of path) {
    if (node.kind === "directory") next.add(node.id);
  }
  return next;
}

function findPathToNode(
  nodes: AdminTreeNode[],
  selectedId: string,
  ancestors: AdminTreeNode[] = [],
): AdminTreeNode[] {
  for (const node of nodes) {
    const path = [...ancestors, node];
    if (node.id === selectedId) return path;
    const childPath = findPathToNode(node.children, selectedId, path);
    if (childPath.length > 0) return childPath;
  }
  return [];
}

function stringValue(form: FormData, key: string) {
  return String(form.get(key) ?? "").trim();
}

function formatAdminCreateError(error: unknown) {
  if (error instanceof ApiError) {
    if (error.status === 401) return "Sign in again.";
    if (error.status === 403) return "Author access is required.";
    if (error.status === 404) return "Target directory was not found. Refresh the tree.";
    if (error.status === 409) return "URL Path already exists. Try another name.";
    if (/name is required/i.test(error.message)) return "Name is required.";
  }
  return "Create failed. Check the network and try again.";
}

function formatAdminActionError(error: unknown, fallback: string) {
  if (error instanceof ApiError) {
    if (error.status === 401) return "Sign in again.";
    if (error.status === 403) return "Author access is required.";
    if (error.status === 404) return "Target item was not found. Refresh the tree.";
    if (error.status === 409) return error.message || "This action is blocked.";
  }
  return fallback;
}

function formatBytes(bytes: number) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 102.4) / 10} KB`;
  return `${Math.round(bytes / 1024 / 102.4) / 10} MB`;
}

function readStoredString(key: string) {
  if (typeof window === "undefined") return "";
  return window.localStorage.getItem(key) ?? "";
}

function readStoredList(key: string) {
  if (typeof window === "undefined") return [];
  try {
    const parsed: unknown = JSON.parse(
      window.localStorage.getItem(key) ?? "[]",
    );
    return Array.isArray(parsed)
      ? parsed.filter((item): item is string => typeof item === "string")
      : [];
  } catch {
    return [];
  }
}
