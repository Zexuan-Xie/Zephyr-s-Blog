package tree

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNodeNotFound                     = errors.New("node not found")
	ErrFileContentNotFound              = errors.New("file content not found")
	ErrNodeIsNotFile                    = errors.New("node is not a file")
	ErrInvalidContentFormat             = errors.New("invalid content format")
	ErrPublishedContentFormatChange     = errors.New("published file content_format cannot change")
	ErrPublishedFileDelete              = errors.New("published file cannot be hard-deleted")
	ErrDirectoryHasPublishedDescendants = errors.New("directory with published descendants cannot be hard-deleted")
	ErrNonEmptyDirectoryDelete          = errors.New("non-empty directory cannot be hard-deleted")
	ErrDuplicatePath                    = errors.New("duplicate url path")
	ErrLostUpdate                       = errors.New("lost update")
	ErrInvalidPathChange                = errors.New("old and new paths must be non-empty")
)

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
