import { FormEvent, useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
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
    return <section className="glass status-panel">正在确认作者权限…</section>;
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

  async function refreshWorkspace() {
    await adminTreeQuery.refetch();
    await detailQuery.refetch();
  }

  async function submitCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedNode || selectedNode.kind !== "directory") {
      setStatusMessage("请先选择一个目录。");
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
      `${created.node.kind === "directory" ? "目录" : "文件"}已创建：${created.node.path}`,
    );
    createForm.reset();
    await adminTreeQuery.refetch();
  }

  async function deleteDirectory(node: AdminTreeNode) {
    if (node.children.length > 0) {
      setStatusMessage("非空目录不能删除，请先移动或删除子项。");
      return;
    }

    try {
      await deleteAdminNode(node.id);
      setStatusMessage(`目录已删除：${node.path}`);
      if (node.parent_id) setSelectedId(node.parent_id);
      await refreshWorkspace();
    } catch (error) {
      setStatusMessage(
        formatAdminActionError(error, "删除目录失败，请检查状态后重试。"),
      );
    }
  }

  async function reorderChildren(parent: AdminTreeNode, childIds: string[]) {
    try {
      await reorderAdminChildren(parent.id, {
        child_ids: childIds,
        expected_version: 0,
      });
      setStatusMessage("同级排序已保存。");
      await refreshWorkspace();
    } catch (error) {
      setStatusMessage(
        formatAdminActionError(error, "排序保存失败，请刷新内容树后重试。"),
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
        <p className="eyebrow">作者工作台</p>
        <h1>内容树</h1>
        <p>
          管理受保护的目录、草稿文件和已发布文件。URL Path
          由系统展示，主要操作不暴露实现标识。
        </p>
        <div className="button-row">
          <button
            className="glass-button"
            type="button"
            onClick={() => window.scrollTo({ top: 0, behavior: "smooth" })}
          >
            返回内容树
          </button>
          <button className="glass-button" type="button" onClick={logoutAuthor}>
            退出登录
          </button>
        </div>
      </section>

      <section className="admin-grid author-workspace-grid">
        <aside
          className="glass admin-sidebar author-tree-panel"
          aria-label="内容树"
        >
          <div className="panel-heading-row">
            <div>
              <p className="eyebrow">Content Tree</p>
              <h2>受保护内容树</h2>
            </div>
            <button
              className="glass-button"
              type="button"
              onClick={() => adminTreeQuery.refetch()}
            >
              刷新
            </button>
          </div>
          {adminTreeQuery.isLoading ? (
            <p className="muted">正在加载目录、草稿和已发布文件…</p>
          ) : null}
          {adminTreeQuery.isError ? (
            <p className="form-error">内容树加载失败。请刷新或重新登录。</p>
          ) : null}
          {adminTreeQuery.data && adminTreeQuery.data.roots.length === 0 ? (
            <p className="muted">暂无内容。</p>
          ) : null}
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
              directoryOptions={directoryOptions}
              isLoading={detailQuery.isLoading}
              isError={detailQuery.isError}
              statusMessage={statusMessage}
              onCreate={submitCreate}
              onCancelCreate={() => setStatusMessage(null)}
              onDeleteDirectory={deleteDirectory}
              onReorderChildren={reorderChildren}
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
            <section className="glass status-panel">
              请选择内容树中的目录或文件。
            </section>
          )}
        </main>
      </section>
    </section>
  );
}

function TreeList({
  nodes,
  expandedIds,
  selectedId,
  onSelect,
  onToggle,
}: {
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

function TreeNodeRow({
  node,
  depth,
  expandedIds,
  selectedId,
  onSelect,
  onToggle,
}: {
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
      <div
        className={`tree-row author-tree-row${selected ? " is-selected" : ""}`}
        style={{ paddingLeft: `${0.65 + depth * 1.1}rem` }}
      >
        {node.kind === "directory" ? (
          <button
            className="tree-toggle"
            type="button"
            aria-label={expanded ? "收起目录" : "展开目录"}
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
            {node.kind === "directory" ? "目录" : "文件"} {node.name}
          </span>
          <small>
            {node.path} · {node.status === "published" ? "已发布" : "草稿"}
          </small>
        </button>
      </div>
      {node.kind === "directory" && expanded && hasChildren ? (
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
  onReorderChildren,
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
  onReorderChildren: (parent: AdminTreeNode, childIds: string[]) => void;
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
            {node.kind === "directory" ? "目录概览" : "文件工作区"}
          </p>
          <h2>{node.name}</h2>
          <p className="path-text">URL Path：{node.path}</p>
        </div>
        <span className={`status-pill ${node.status}`}>
          {node.status === "published" ? "已发布" : "草稿"}
        </span>
      </div>

      {node.kind === "file" ? (
        <button
          className="glass-button"
          type="button"
          onClick={onReturnToDirectory}
        >
          返回目录
        </button>
      ) : null}
      {statusMessage ? <p className="muted">{statusMessage}</p> : null}
      {isLoading ? <p className="muted">正在加载工作区详情…</p> : null}
      {isError ? (
        <p className="form-error">工作区详情加载失败。可继续使用左侧内容树。</p>
      ) : null}

      {node.kind === "directory" ? (
        <DirectoryOverview
          node={node}
          children={children}
          onCreate={onCreate}
          onCancelCreate={onCancelCreate}
          onDeleteDirectory={onDeleteDirectory}
          onReorderChildren={onReorderChildren}
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
  onReorderChildren,
}: {
  node: AdminTreeNode;
  children: AdminTreeNode[];
  onCreate: (event: FormEvent<HTMLFormElement>) => void;
  onCancelCreate: () => void;
  onDeleteDirectory: (node: AdminTreeNode) => void;
  onReorderChildren: (parent: AdminTreeNode, childIds: string[]) => void;
}) {
  const [kind, setKind] = useState<NodeKind>("directory");
  const [draggedChildId, setDraggedChildId] = useState<string | null>(null);
  const previewName = kind === "directory" ? "新目录" : "新文件";
  const previewPath = `${node.path.replace(/\/$/, "")}/${previewName}`.replace(
    /^$/,
    "/",
  );

  function dropOnChild(targetId: string) {
    if (!draggedChildId || draggedChildId === targetId) return;
    const childIds = children.map((child) => child.id);
    const from = childIds.indexOf(draggedChildId);
    const to = childIds.indexOf(targetId);
    if (from < 0 || to < 0) return;
    const [moved] = childIds.splice(from, 1);
    childIds.splice(to, 0, moved);
    onReorderChildren(node, childIds);
    setDraggedChildId(null);
  }

  return (
    <div className="directory-overview">
      <p className="muted">
        此目录包含 {children.length} 个直接子项。可在当前目录中新建目录或文件。
      </p>
      <section className="nested-create-panel" aria-label="新建目录或文件">
        <h3>{kind === "directory" ? "新建目录" : "新建文件"}</h3>
        <form className="admin-form" onSubmit={onCreate}>
          <input type="hidden" name="kind" value={kind} />
          <label>
            类型
            <select
              value={kind}
              onChange={(event) => setKind(event.target.value as NodeKind)}
            >
              <option value="directory">新建目录</option>
              <option value="file">新建文件</option>
            </select>
          </label>
          <label>
            名称
            <input name="name" required placeholder={previewName} />
          </label>
          {kind === "file" ? (
            <label>
              格式
              <select name="content_format" defaultValue="markdown">
                <option value="markdown">Markdown</option>
                <option value="html_document">HTML Document</option>
              </select>
            </label>
          ) : null}
          <label>
            URL Path preview
            <input readOnly value={previewPath} />
          </label>
          <div className="button-row">
            <button className="primary-button" type="submit">
              创建并打开
            </button>
            <button
              className="glass-button"
              type="button"
              onClick={onCancelCreate}
            >
              取消
            </button>
          </div>
        </form>
      </section>
      <section className="danger-zone" aria-label="目录危险操作">
        <h3>危险操作</h3>
        <p className="muted">
          非空目录不能删除。请先移动或删除所有子项后再删除目录。
        </p>
        <button
          className="glass-button danger-button"
          type="button"
          onClick={() => onDeleteDirectory(node)}
        >
          {children.length > 0 ? "非空目录不能删除" : "删除空目录"}
        </button>
      </section>
      {children.length === 0 ? <p className="muted">此目录暂无子项。</p> : null}
      {children.length > 1 ? (
        <p className="muted">
          桌面端同级拖拽排序：只能调整当前目录内的子项顺序，不会移动到其他目录。
        </p>
      ) : null}
      <div className="admin-child-card-grid">
        {children.map((child) => (
          <article
            className="admin-child-card"
            draggable
            key={child.id}
            onDragStart={() => setDraggedChildId(child.id)}
            onDragOver={(event) => event.preventDefault()}
            onDrop={() => dropOnChild(child.id)}
          >
            <strong>
              {child.kind === "directory" ? "目录" : "文件"} {child.name}
            </strong>
            <span>{child.path}</span>
            <small>{child.status === "published" ? "已发布" : "草稿"}</small>
          </article>
        ))}
      </div>
      <p className="muted">当前目录：{node.path}</p>
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
      onFeedback("内容已手动保存。");
      await onRefresh();
    } catch (error) {
      onFeedback(formatAdminActionError(error, "内容保存失败，请检查后重试。"));
    }
  }

  async function togglePublish(nextStatus: "draft" | "published") {
    try {
      if (nextStatus === "published") {
        await publishFile(node.id);
        onFeedback("文件已发布。");
      } else {
        await unpublishFile(node.id);
        onFeedback("文件已撤回发布，当前为草稿。");
      }
      await onRefresh();
    } catch (error) {
      onFeedback(
        formatAdminActionError(error, "发布状态更新失败，请稍后重试。"),
      );
    }
  }

  async function submitAsset(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const file = form.get("asset");
    if (!(file instanceof File)) {
      onFeedback("请选择要上传的资源。");
      return;
    }
    try {
      await uploadAsset(node.id, file);
      onFeedback(`资源已上传：${file.name}`);
      event.currentTarget.reset();
      await onRefresh();
    } catch (error) {
      onFeedback(
        formatAdminActionError(error, "资源上传失败，请检查文件后重试。"),
      );
    }
  }

  async function removeAsset(assetId: string) {
    try {
      await deleteAsset(assetId);
      onFeedback("资源已删除。");
      await onRefresh();
    } catch (error) {
      onFeedback(formatAdminActionError(error, "资源删除失败，请稍后重试。"));
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
      onFeedback("基础信息已保存。");
      await onRefresh();
    } catch (error) {
      onFeedback(
        formatAdminActionError(error, "基础信息保存失败，请检查 URL Path。"),
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
      onFeedback(`移动预览已生成：${preview.destination_path}`);
    } catch (error) {
      onFeedback(formatAdminActionError(error, "移动预览失败，请换一个目录。"));
    }
  }

  async function commitMove() {
    try {
      await moveAdminNode(node.id, {
        new_parent_id: moveDestinationId,
        expected_version: 0,
      });
      onFeedback("位置已移动。");
      setMovePreview(null);
      setMoveDestinationId(null);
      await onRefresh();
    } catch (error) {
      onFeedback(formatAdminActionError(error, "移动失败，请重新生成预览。"));
    }
  }

  async function deleteFile() {
    try {
      await deleteAdminNode(node.id);
      onFeedback(`文件已删除：${node.path}`);
      await onRefresh();
    } catch (error) {
      onFeedback(
        formatAdminActionError(error, "已发布文件不能直接删除，请先撤回发布。"),
      );
    }
  }

  return (
    <div className="file-overview">
      <div className="admin-tabs" aria-label="文件工作区标签">
        <button
          className={
            activeTab === "content" ? "primary-button" : "glass-button"
          }
          type="button"
          onClick={() => setActiveTab("content")}
        >
          内容
        </button>
        <button
          className={activeTab === "assets" ? "primary-button" : "glass-button"}
          type="button"
          onClick={() => setActiveTab("assets")}
        >
          资源
        </button>
        <button
          className={
            activeTab === "settings" ? "primary-button" : "glass-button"
          }
          type="button"
          onClick={() => setActiveTab("settings")}
        >
          设置
        </button>
      </div>
      <p className="muted">
        文件状态：{node.status === "published" ? "已发布" : "草稿"}
      </p>

      {activeTab === "content" ? (
        <section className="workspace-tab-panel" aria-label="内容">
          <div className="button-row">
            {node.status === "draft" ? (
              <button
                className="primary-button"
                type="button"
                onClick={() => togglePublish("published")}
              >
                发布
              </button>
            ) : null}
            {node.status === "published" ? (
              <button
                className="glass-button"
                type="button"
                onClick={() => togglePublish("draft")}
              >
                撤回发布
              </button>
            ) : null}
          </div>
          <form className="admin-form" onSubmit={submitContent}>
            <label>
              格式
              <select name="content_format" defaultValue={contentFormat}>
                <option value="markdown">Markdown</option>
                <option value="html_document">HTML Document</option>
              </select>
            </label>
            <label>
              关键词
              <input
                name="keywords"
                defaultValue={keywords}
                placeholder="用逗号分隔"
              />
            </label>
            <label>
              正文
              <textarea
                name="body_raw"
                defaultValue={bodyRaw}
                rows={12}
                placeholder="在这里手动编辑草稿内容"
              />
            </label>
            <div className="button-row">
              <button className="primary-button" type="submit">
                手动保存
              </button>
            </div>
          </form>
        </section>
      ) : null}

      {activeTab === "assets" ? (
        <section className="workspace-tab-panel" aria-label="资源">
          <form className="admin-form" onSubmit={submitAsset}>
            <label>
              上传资源
              <input name="asset" type="file" />
            </label>
            <div className="button-row">
              <button className="primary-button" type="submit">
                上传资源
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
                    打开
                  </a>
                  <button
                    className="glass-button danger-button"
                    type="button"
                    onClick={() => removeAsset(asset.id)}
                  >
                    删除
                  </button>
                </article>
              ))}
            </div>
          ) : (
            <p className="muted">暂无资源。</p>
          )}
        </section>
      ) : null}

      {activeTab === "settings" ? (
        <section className="workspace-tab-panel" aria-label="设置">
          <form className="admin-form" onSubmit={submitSettings}>
            <label>
              名称
              <input name="name" defaultValue={node.name} required />
            </label>
            <label>
              URL Path
              <input name="url_path" defaultValue={node.path} />
            </label>
            <div className="button-row">
              <button className="primary-button" type="submit">
                保存基础信息
              </button>
            </div>
          </form>

          <section className="nested-create-panel" aria-label="位置">
            <h3>位置</h3>
            <form className="admin-form" onSubmit={submitMovePreview}>
              <label>
                Directory Picker
                <select
                  name="new_parent_id"
                  defaultValue={node.parent_id ?? ""}
                >
                  <option value="">根目录</option>
                  {availableDestinations.map((directory) => (
                    <option value={directory.id} key={directory.id}>
                      {directory.path}
                    </option>
                  ))}
                </select>
              </label>
              <div className="button-row">
                <button className="glass-button" type="submit">
                  预览移动
                </button>
              </div>
            </form>
            {movePreview ? (
              <div className="move-preview-panel">
                <p>目标路径：{movePreview.destination_path}</p>
                <p>影响路径：{movePreview.affected_paths.length || 0} 个</p>
                {movePreview.redirects.length > 0 ? (
                  <p>将创建 {movePreview.redirects.length} 条公开文件跳转。</p>
                ) : null}
                {movePreview.blocked_reasons.length > 0 ? (
                  <p className="form-error">
                    阻止原因：{movePreview.blocked_reasons.join("，")}
                  </p>
                ) : null}
                <button
                  className="primary-button"
                  type="button"
                  disabled={movePreview.blocked_reasons.length > 0}
                  onClick={commitMove}
                >
                  确认移动
                </button>
              </div>
            ) : null}
          </section>

          <section className="danger-zone" aria-label="危险操作">
            <h3>危险操作</h3>
            <p className="muted">
              已发布文件会被后端阻止删除。请先使用撤回发布。
            </p>
            <button
              className="glass-button danger-button"
              type="button"
              onClick={deleteFile}
            >
              删除文件
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
    if (error.status === 401) return "登录已过期，请重新登录。";
    if (error.status === 403) return "需要作者权限才能创建内容。";
    if (error.status === 404) return "目标目录不存在，请刷新内容树后重试。";
    if (error.status === 409) return "URL Path 已存在，请换一个名称。";
    if (/name is required/i.test(error.message)) return "请输入名称。";
  }
  return "创建失败，请检查网络后重试。";
}

function formatAdminActionError(error: unknown, fallback: string) {
  if (error instanceof ApiError) {
    if (error.status === 401) return "登录已过期，请重新登录。";
    if (error.status === 403) return "需要作者权限才能执行此操作。";
    if (error.status === 404) return "目标内容不存在，请刷新内容树。";
    if (error.status === 409) return error.message || "当前状态不允许此操作。";
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
