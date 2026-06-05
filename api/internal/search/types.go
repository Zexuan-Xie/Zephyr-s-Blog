package search

import (
	"time"

	"github.com/google/uuid"

	"xlab-blog/api/internal/tree"
)

const (
	DefaultLimit = 20
	MaxLimit     = 50
	DefaultRRFK  = 60

	SourceText     = "text"
	SourceSemantic = "semantic"
	SourceKeyword  = "keyword"
)

type Result struct {
	File         tree.FileEntry `json:"file"`
	Path         string         `json:"path"`
	Snippet      string         `json:"snippet"`
	Score        float64        `json:"score"`
	MatchSources []string       `json:"match_sources"`
}

type Response struct {
	Query string   `json:"query"`
	Items []Result `json:"items"`
}

type Candidate struct {
	File         tree.FileEntry
	Path         string
	Snippet      string
	Rank         int
	Score        float64
	Source       string
	KeywordMatch bool
}

type Options struct {
	Limit  int
	Offset int
}

type EmbeddingInput struct {
	FileID     uuid.UUID
	Name       string
	Path       string
	Keywords   []string
	SearchText string
}

type EmbeddingState struct {
	FileID     uuid.UUID            `json:"file_id"`
	Provider   string               `json:"provider"`
	Model      string               `json:"model"`
	Dimensions int                  `json:"dimensions"`
	Status     tree.EmbeddingStatus `json:"status"`
	Error      *string              `json:"error"`
	UpdatedAt  *time.Time           `json:"updated_at"`
}

type RebuildState struct {
	Status string `json:"status"`
}
