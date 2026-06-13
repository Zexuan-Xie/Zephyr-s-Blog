package tree

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type NodeKind string

const (
	NodeKindDirectory NodeKind = "directory"
	NodeKindFile      NodeKind = "file"
)

type ContentFormat string

const (
	ContentFormatMarkdown     ContentFormat = "markdown"
	ContentFormatHTMLDocument ContentFormat = "html_document"
)

type PublishStatus string

const (
	PublishStatusDraft              PublishStatus = "draft"
	PublishStatusPublished          PublishStatus = "published"
	PublishStatusUnpublishedChanges PublishStatus = "unpublished_changes"
)

type EmbeddingStatus string

const (
	EmbeddingStatusPending EmbeddingStatus = "pending"
	EmbeddingStatusReady   EmbeddingStatus = "ready"
	EmbeddingStatusFailed  EmbeddingStatus = "failed"
)

type Node struct {
	ID        uuid.UUID  `json:"id"`
	ParentID  *uuid.UUID `json:"parent_id"`
	Kind      NodeKind   `json:"kind"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	Path      string     `json:"path"`
	SortOrder int        `json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type DirectoryEntry struct {
	Node                Node `json:"node"`
	ChildDirectoryCount int  `json:"child_directory_count"`
	ChildFileCount      int  `json:"child_file_count"`
}

type FileEntry struct {
	Node               Node          `json:"node"`
	ContentFormat      ContentFormat `json:"content_format"`
	Status             PublishStatus `json:"status"`
	Keywords           []string      `json:"keywords"`
	PublishedAt        *time.Time    `json:"published_at"`
	LikeCount          int           `json:"like_count"`
	CommentCount       int           `json:"comment_count"`
	ReadingTimeMinutes *int          `json:"reading_time_minutes"`
}

type FileEntryList struct {
	Items []FileEntry `json:"items"`
}

// AdminTreeNode is the flat protected Author Workspace tree item.
type AdminTreeNode struct {
	ID            uuid.UUID      `json:"id"`
	ParentID      *uuid.UUID     `json:"parent_id"`
	Kind          NodeKind       `json:"kind"`
	Name          string         `json:"name"`
	URLPath       string         `json:"url_path"`
	SortOrder     int            `json:"sort_order"`
	Status        PublishStatus  `json:"status"`
	ContentFormat *ContentFormat `json:"content_format,omitempty"`
}

type AdminTreeResponse struct {
	Nodes []AdminTreeNode `json:"nodes"`
}

type ReorderChildrenInput struct {
	ChildIDs        []uuid.UUID `json:"child_ids"`
	ExpectedVersion int         `json:"expected_version"`
}

type ReorderChildrenResult struct {
	ParentID uuid.UUID   `json:"parent_id"`
	ChildIDs []uuid.UUID `json:"child_ids"`
	Version  int         `json:"version"`
}

type MoveNodeInput struct {
	NewParentID     *uuid.UUID `json:"new_parent_id"`
	ExpectedVersion int        `json:"expected_version"`
}

type PathRedirectPreview struct {
	OldPath string    `json:"old_path"`
	NewPath string    `json:"new_path"`
	NodeID  uuid.UUID `json:"node_id"`
}

type MovePreview struct {
	NodeID          uuid.UUID             `json:"node_id"`
	DestinationPath string                `json:"destination_path"`
	AffectedPaths   []string              `json:"affected_paths"`
	Redirects       []PathRedirectPreview `json:"redirects"`
	BlockedReasons  []string              `json:"blocked_reasons"`
}

type DirectoryPage struct {
	Node    *Node  `json:"node"`
	Path    string `json:"path,omitempty"`
	Entries []any  `json:"entries"`
}

type FileContent struct {
	NodeID             uuid.UUID       `json:"node_id"`
	Revision           int             `json:"revision"`
	ContentFormat      ContentFormat   `json:"content_format"`
	Keywords           []string        `json:"keywords"`
	BodyRaw            string          `json:"body_raw"`
	BodyHTML           *string         `json:"body_html"`
	SearchText         string          `json:"search_text"`
	Status             PublishStatus   `json:"status"`
	PublishedAt        *time.Time      `json:"published_at"`
	LastSavedAt        time.Time       `json:"last_saved_at"`
	EmbeddingModel     *string         `json:"embedding_model"`
	EmbeddingStatus    EmbeddingStatus `json:"embedding_status"`
	EmbeddingError     *string         `json:"embedding_error"`
	EmbeddingUpdatedAt *time.Time      `json:"embedding_updated_at"`
}

type FileAsset struct {
	ID               uuid.UUID  `json:"id"`
	FileID           uuid.UUID  `json:"file_node_id"`
	Filename         string     `json:"filename"`
	MIMEType         string     `json:"mime_type"`
	SizeBytes        int64      `json:"size_bytes"`
	StorageProvider  string     `json:"-"`
	StorageKey       string     `json:"-"`
	PublicURL        string     `json:"public_url,omitempty"`
	State            string     `json:"state"`
	PublishedAssetID *uuid.UUID `json:"published_asset_id"`
	CreatedAt        time.Time  `json:"created_at"`
}

type PublishedContent struct {
	NodeID         uuid.UUID     `json:"node_id"`
	SourceRevision int           `json:"source_revision"`
	ContentFormat  ContentFormat `json:"content_format"`
	Keywords       []string      `json:"keywords"`
	BodyRaw        string        `json:"body_raw"`
	BodyHTML       *string       `json:"body_html"`
	SearchText     string        `json:"search_text"`
	PublishedAt    time.Time     `json:"published_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	Visible        bool          `json:"visible"`
}

type PublicFileContent struct {
	NodeID        uuid.UUID     `json:"node_id"`
	ContentFormat ContentFormat `json:"content_format"`
	Keywords      []string      `json:"keywords"`
	BodyRaw       string        `json:"body_raw"`
	BodyHTML      *string       `json:"body_html"`
	SearchText    string        `json:"search_text"`
	PublishedAt   time.Time     `json:"published_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type FileVersionState struct {
	Current               FileContent       `json:"current"`
	Previous              *FileContent      `json:"previous"`
	Published             *PublishedContent `json:"published"`
	HasUnpublishedChanges bool              `json:"has_unpublished_changes"`
	DraftAssets           []FileAsset       `json:"draft_assets"`
	PublishedAssets       []FileAsset       `json:"published_assets"`
}

type PublishResult struct {
	Current        FileContent      `json:"current"`
	Published      PublishedContent `json:"published"`
	PromotedAssets []FileAsset      `json:"promoted_assets"`
}

type FilePage struct {
	Node           Node              `json:"node"`
	Content        PublicFileContent `json:"content"`
	KeywordsPublic []string          `json:"keywords_public"`
	LikeCount      int               `json:"like_count"`
	ViewerHasLiked bool              `json:"viewer_has_liked"`
	CommentCount   int               `json:"comment_count"`
	Assets         []FileAsset       `json:"assets"`
}

type ResolveType string

const (
	ResolveTypeDirectory ResolveType = "directory"
	ResolveTypeFile      ResolveType = "file"
	ResolveTypeRedirect  ResolveType = "redirect"
)

type ResolveResponse struct {
	Type      ResolveType    `json:"type"`
	Directory *DirectoryPage `json:"directory,omitempty"`
	File      *FilePage      `json:"file,omitempty"`
	NewPath   string         `json:"new_path,omitempty"`
}

var (
	ErrNotFound    = errors.New("tree node not found")
	ErrInvalidPath = errors.New("invalid tree path")
)

type Repository interface {
	DirectoryPage(ctx context.Context, parentID *uuid.UUID) (DirectoryPage, error)
	FilePage(ctx context.Context, node Node) (FilePage, error)
	FindNodeByParentAndSlug(ctx context.Context, parentID *uuid.UUID, slug string) (Node, error)
	RedirectPath(ctx context.Context, oldPath string) (string, error)
}
