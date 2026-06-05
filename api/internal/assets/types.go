package assets

import (
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
)

const (
	StorageProviderLocal = "local"
	MaxImageSize         = 5 * 1024 * 1024
	MaxPDFSize           = 20 * 1024 * 1024
	MaxTextSize          = 2 * 1024 * 1024
	MaxPerFileTotalSize  = 50 * 1024 * 1024
)

var (
	ErrFileNotFound       = errors.New("file not found")
	ErrAssetNotFound      = errors.New("asset not found")
	ErrInvalidFilename    = errors.New("invalid asset filename")
	ErrMIMETypeNotAllowed = errors.New("asset MIME type not allowed")
	ErrAssetTooLarge      = errors.New("asset too large")
	ErrPerFileLimit       = errors.New("per-file asset limit exceeded")
	ErrUnsafeSVG          = errors.New("unsafe SVG asset")
	ErrDuplicateAssetName = errors.New("asset filename already exists")
	ErrStorageKeyUnsafe   = errors.New("asset storage key is unsafe")
)

type FileAsset struct {
	ID              uuid.UUID `json:"id"`
	FileID          uuid.UUID `json:"file_node_id"`
	Filename        string    `json:"filename"`
	MIMEType        string    `json:"mime_type"`
	SizeBytes       int64     `json:"size_bytes"`
	StorageProvider string    `json:"storage_provider"`
	StorageKey      string    `json:"storage_key"`
	PublicURL       string    `json:"public_url"`
	CreatedAt       time.Time `json:"created_at"`
}

type Upload struct {
	Filename string
	MIMEType string
	Size     int64
	Reader   io.Reader
}

type StoredObject struct {
	Reader      io.ReadCloser
	Size        int64
	ContentType string
}

type Storage interface {
	Put(key string, reader io.Reader) error
	Open(key string) (StoredObject, error)
	Delete(key string) error
}
