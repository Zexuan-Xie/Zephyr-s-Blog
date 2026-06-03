package tree

import (
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
	PublishStatusDraft     PublishStatus = "draft"
	PublishStatusPublished PublishStatus = "published"
)

type EmbeddingStatus string

const EmbeddingStatusPending EmbeddingStatus = "pending"

var (
	ErrNodeNotFound                     = errors.New("node not found")
	ErrFileContentNotFound              = errors.New("file content not found")
	ErrNodeIsNotFile                    = errors.New("node is not a file")
	ErrPublishedContentFormatChange     = errors.New("published file content_format cannot change")
	ErrPublishedFileDelete              = errors.New("published file cannot be hard-deleted")
	ErrDirectoryHasPublishedDescendants = errors.New("directory with published descendants cannot be hard-deleted")
	ErrInvalidPathChange                = errors.New("old and new paths must be non-empty")
)

type Node struct {
	ID        uuid.UUID `json:"id"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	Kind      NodeKind  `json:"kind"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Path      string    `json:"path"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FileContent struct {
	NodeID             uuid.UUID       `json:"node_id"`
	ContentFormat      ContentFormat  `json:"content_format"`
	Keywords           []string       `json:"keywords"`
	BodyRaw            string         `json:"body_raw"`
	BodyHTML           *string        `json:"body_html"`
	SearchText         string         `json:"search_text"`
	Status             PublishStatus  `json:"status"`
	PublishedAt        *time.Time     `json:"published_at"`
	EmbeddingStatus    EmbeddingStatus `json:"embedding_status"`
	EmbeddingModel     *string        `json:"embedding_model,omitempty"`
	EmbeddingError     *string        `json:"embedding_error,omitempty"`
	EmbeddingUpdatedAt *time.Time     `json:"embedding_updated_at,omitempty"`
}

type UpsertFileContentInput struct {
	ContentFormat ContentFormat `json:"content_format"`
	BodyRaw       string        `json:"body_raw"`
	BodyHTML      *string       `json:"body_html,omitempty"`
	Keywords      []string      `json:"keywords"`
	SearchText    string        `json:"search_text,omitempty"`
}

type PublishedFilePath struct {
	NodeID uuid.UUID
	Path   string
}
