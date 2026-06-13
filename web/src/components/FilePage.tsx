import { FormEvent, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Download, Heart, MessageCircle, Reply, Trash2 } from 'lucide-react';
import { Link, useNavigate } from 'react-router-dom';
import {
  createComment,
  deleteComment,
  fetchCommentThread,
  likeComment,
  likeFile,
  unlikeComment,
  unlikeFile,
} from '../lib/api';
import { getToken } from '../lib/auth';
import { renderSafeMarkdown, sanitizeServerHtml } from '../lib/renderMarkdown';
import type { CommentItem, CurrentUser, FilePayload, LikeState } from '../lib/types';
import { Breadcrumb } from './Breadcrumb';

interface FilePageProps {
  file: FilePayload;
  currentUser: CurrentUser | null;
}

interface ReplyTarget {
  parentId: string;
  replyToUserId: string;
  displayName: string;
}

export function FilePage({ file, currentUser }: FilePageProps) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [fileLikeState, setFileLikeState] = useState<LikeState>({
    liked: Boolean(file.viewer_has_liked),
    like_count: file.like_count ?? 0,
  });
  const [replyTarget, setReplyTarget] = useState<ReplyTarget | null>(null);
  const [commentBody, setCommentBody] = useState('');
  const [formError, setFormError] = useState<string | null>(null);
  const token = getToken();
  const isAuthenticated = Boolean(token);
  const isAuthor = currentUser?.role === 'admin';
  const commentsQuery = useQuery({
    queryKey: ['comments', file.id],
    queryFn: () => fetchCommentThread(file.id),
    staleTime: 30_000,
  });

  const createCommentMutation = useMutation({
    mutationFn: () => createComment(file.id, commentBody, replyTarget?.parentId, replyTarget?.replyToUserId),
    onSuccess: async () => {
      setCommentBody('');
      setReplyTarget(null);
      setFormError(null);
      await queryClient.invalidateQueries({ queryKey: ['comments', file.id] });
    },
    onError: () => setFormError('Unable to post this comment. Please check the text and try again.'),
  });

  const deleteCommentMutation = useMutation({
    mutationFn: (commentId: string) => deleteComment(commentId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['comments', file.id] }),
  });

  const commentLikeMutation = useMutation({
    mutationFn: ({ commentId, liked }: { commentId: string; liked: boolean }) => (
      liked ? unlikeComment(commentId) : likeComment(commentId)
    ),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['comments', file.id] }),
  });

  const fileLikeMutation = useMutation({
    mutationFn: () => (fileLikeState.liked ? unlikeFile(file.id) : likeFile(file.id)),
    onSuccess: setFileLikeState,
  });

  const keywords = file.keywords?.slice(0, 3) ?? [];
  const markdownHtml = file.content_format === 'markdown'
    ? file.body_html
      ? sanitizeServerHtml(file.body_html)
      : renderSafeMarkdown(file.body_markdown ?? '')
    : '';

  function redirectToLogin() {
    navigate(`/login?return_to=${encodeURIComponent(file.path)}`);
  }

  function requireAuth(action: () => void) {
    if (!isAuthenticated) {
      redirectToLogin();
      return;
    }
    action();
  }

  function submitComment(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    requireAuth(() => {
      if (!commentBody.trim()) {
        setFormError('Comment body cannot be empty.');
        return;
      }
      createCommentMutation.mutate();
    });
  }

  function startReply(comment: CommentItem) {
    requireAuth(() => {
      setReplyTarget({
        parentId: comment.id,
        replyToUserId: comment.user.id,
        displayName: comment.user.display_name,
      });
    });
  }

  function toggleCommentLike(comment: CommentItem) {
    requireAuth(() => commentLikeMutation.mutate({ commentId: comment.id, liked: Boolean(comment.viewer_has_liked) }));
  }

  return (
    <article className="file-page">
      <Breadcrumb items={file.breadcrumb} currentPath={file.path} />
      <section className="file-heading">
        <div className="keyword-row">
          {keywords.map((keyword) => (
            <Link className="keyword-chip" key={keyword} to={`/search?q=${encodeURIComponent(keyword)}`}>
              {keyword}
            </Link>
          ))}
        </div>
        <h1>{file.name}</h1>
        <p className="muted">
          {file.path}
          {file.updated_at ? ` · updated ${new Date(file.updated_at).toLocaleDateString()}` : ''}
          {file.read_time_minutes ? ` · ${file.read_time_minutes} min read` : ''}
        </p>
      </section>

      {isAuthor ? (
        <div className="action-row" aria-label="Author file actions">
          <Link className="glass-button" to={`/admin?target=${encodeURIComponent(file.id)}`}>Edit</Link>
        </div>
      ) : null}

      {file.content_format === 'markdown' ? (
        <section className="glass file-reading-card" dangerouslySetInnerHTML={{ __html: markdownHtml }} />
      ) : (
        <section className="glass html-document-shell" aria-label={`${file.name} HTML document`}>
          <iframe
            title={`${file.name} document`}
            sandbox="allow-scripts"
            srcDoc={file.html_document ?? file.body_html ?? '<!doctype html><html><body><p>Empty HTML document.</p></body></html>'}
          />
        </section>
      )}


      {file.assets.length > 0 ? (
        <section className="glass asset-panel" aria-label="File assets">
          <p className="eyebrow">ASSETS</p>
          <h2>Files attached to this page</h2>
          <div className="asset-list">
            {file.assets.map((asset) => (
              <a className="asset-link" key={asset.id} href={asset.public_url} target="_blank" rel="noreferrer">
                <Download size={16} aria-hidden="true" />
                <span>{asset.filename}</span>
                <small>{asset.mime_type} · {formatBytes(asset.size_bytes)}</small>
              </a>
            ))}
          </div>
        </section>
      ) : null}

      <footer className="glass interaction-bar" aria-label="File interactions">
        <button
          className="glass-button"
          type="button"
          aria-pressed={fileLikeState.liked}
          onClick={() => requireAuth(() => fileLikeMutation.mutate())}
        >
          <Heart size={17} aria-hidden="true" />
          <span>{fileLikeState.liked ? 'Liked' : 'Like'} · {fileLikeState.like_count}</span>
        </button>
        <button className="glass-button" type="button" onClick={() => requireAuth(() => document.getElementById('comment-composer')?.focus())}>
          <MessageCircle size={17} aria-hidden="true" />
          <span>{isAuthenticated ? 'Comment' : 'Log in to comment'} · {file.comment_count ?? commentsQuery.data?.comments.length ?? 0}</span>
        </button>
      </footer>

      <section className="glass comments-panel" aria-label="Comments">
        <div className="comments-heading">
          <div>
            <p className="eyebrow">COMMENTS</p>
            <h2>Discussion</h2>
          </div>
          {!isAuthenticated ? <Link className="glass-button" to={`/login?return_to=${encodeURIComponent(file.path)}`}>Log in to comment</Link> : null}
        </div>

        {isAuthenticated ? (
          <form className="comment-form" onSubmit={submitComment}>
            {replyTarget ? (
              <p className="muted">
                Replying to {replyTarget.displayName}{' '}
                <button className="text-button" type="button" onClick={() => setReplyTarget(null)}>Cancel</button>
              </p>
            ) : null}
            <textarea
              id="comment-composer"
              name="body"
              maxLength={5000}
              placeholder="Write a comment…"
              value={commentBody}
              onChange={(event) => setCommentBody(event.target.value)}
            />
            {formError ? <p className="form-error">{formError}</p> : null}
            <button className="primary-button" type="submit" disabled={createCommentMutation.isPending}>
              {createCommentMutation.isPending ? 'Posting…' : replyTarget ? 'Post reply' : 'Post comment'}
            </button>
          </form>
        ) : (
          <p className="muted">Comments are public to read. Log in to join the conversation.</p>
        )}

        <div className="comment-thread">
          {commentsQuery.isLoading ? <p className="muted">Loading comments…</p> : null}
          {commentsQuery.isError ? <p className="form-error">Unable to load comments.</p> : null}
          {commentsQuery.data?.comments.length === 0 ? <p className="muted">No comments yet.</p> : null}
          {commentsQuery.data?.comments.map((comment) => (
            <CommentCard
              key={comment.id}
              comment={comment}
              canWrite={isAuthenticated}
              onReply={startReply}
              onDelete={(commentId) => requireAuth(() => deleteCommentMutation.mutate(commentId))}
              onToggleLike={toggleCommentLike}
            />
          ))}
        </div>
      </section>
    </article>
  );
}

interface CommentCardProps {
  comment: CommentItem;
  canWrite: boolean;
  onReply: (comment: CommentItem) => void;
  onDelete: (commentId: string) => void;
  onToggleLike: (comment: CommentItem) => void;
}

function CommentCard({ comment, canWrite, onReply, onDelete, onToggleLike }: CommentCardProps) {
  const deleted = comment.deleted;
  return (
    <article className={`comment-card${comment.parent_id ? ' is-reply' : ''}`}>
      <header className="comment-meta">
        <strong>{comment.user.display_name}</strong>
        <span>{new Date(comment.created_at).toLocaleDateString()}</span>
      </header>
      <p className={deleted ? 'muted deleted-comment' : undefined}>{deleted ? 'This comment has been deleted.' : comment.body}</p>
      {!deleted ? (
        <div className="comment-actions">
          <button className="glass-button" type="button" aria-pressed={Boolean(comment.viewer_has_liked)} onClick={() => onToggleLike(comment)}>
            <Heart size={15} aria-hidden="true" />
            <span>{comment.viewer_has_liked ? 'Liked' : 'Like'} · {comment.like_count}</span>
          </button>
          <button className="glass-button" type="button" onClick={() => onReply(comment)}>
            <Reply size={15} aria-hidden="true" />
            <span>{canWrite ? 'Reply' : 'Log in to reply'}</span>
          </button>
          {canWrite ? (
            <button className="glass-button" type="button" onClick={() => onDelete(comment.id)}>
              <Trash2 size={15} aria-hidden="true" />
              <span>Delete</span>
            </button>
          ) : null}
        </div>
      ) : null}
      {comment.replies.length > 0 ? (
        <div className="comment-replies">
          {comment.replies.map((reply) => (
            <CommentCard
              key={reply.id}
              comment={reply}
              canWrite={canWrite}
              onReply={onReply}
              onDelete={onDelete}
              onToggleLike={onToggleLike}
            />
          ))}
        </div>
      ) : null}
    </article>
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
