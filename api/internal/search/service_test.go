package search

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"xlab-blog/api/internal/tree"
)

func TestSearchFusesFullTextAndSemanticWithRRF(t *testing.T) {
	fileA := searchTestFile("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "/notes/a", "Alpha")
	fileB := searchTestFile("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", "/notes/b", "Beta")
	repo := &fakeRepository{
		text:     []Candidate{{File: fileA, Path: fileA.Node.Path, Snippet: "<mark>alpha</mark>", KeywordMatch: true}, {File: fileB, Path: fileB.Node.Path, Snippet: "beta"}},
		semantic: []Candidate{{File: fileB, Path: fileB.Node.Path, Snippet: "semantic beta"}},
	}
	service := NewService(repo, fakeProvider{embedding: make([]float32, 1024)}, "text-embedding-v4", 1024)

	response, err := service.Search(context.Background(), " alpha ", Options{Limit: 10})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if response.Query != "alpha" {
		t.Fatalf("query = %q", response.Query)
	}
	if len(response.Items) != 2 {
		t.Fatalf("items len = %d", len(response.Items))
	}
	if response.Items[0].Path != "/notes/b" {
		t.Fatalf("first result path = %q, want fused semantic/text result /notes/b", response.Items[0].Path)
	}
	if got := response.Items[1].MatchSources; len(got) != 2 || got[0] != SourceText || got[1] != SourceKeyword {
		t.Fatalf("sources = %#v, want text+keyword", got)
	}
	if repo.semanticVector == "" {
		t.Fatal("semantic repository was not called with vector literal")
	}
}

func TestSearchFallsBackToFullTextWhenEmbeddingFails(t *testing.T) {
	file := searchTestFile("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "/notes/a", "Alpha")
	repo := &fakeRepository{text: []Candidate{{File: file, Path: file.Node.Path, Snippet: "alpha"}}}
	service := NewService(repo, fakeProvider{err: errors.New("qwen unavailable")}, "text-embedding-v4", 1024)

	response, err := service.Search(context.Background(), "alpha", Options{})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(response.Items) != 1 || response.Items[0].Path != "/notes/a" {
		t.Fatalf("items = %#v", response.Items)
	}
	if repo.semanticVector != "" {
		t.Fatal("semantic search should not run when embedding fails")
	}
}

func TestRefreshEmbeddingRecordsProviderFailureWithoutHardFail(t *testing.T) {
	fileID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	repo := &fakeRepository{inputs: map[uuid.UUID]EmbeddingInput{fileID: {FileID: fileID, Name: "A", Path: "/a", Keywords: []string{"go"}, SearchText: "body"}}}
	service := NewService(repo, fakeProvider{err: errors.New("network down")}, "text-embedding-v4", 1024)

	state, err := service.RefreshFileEmbedding(context.Background(), fileID)
	if err != nil {
		t.Fatalf("RefreshFileEmbedding() error = %v", err)
	}
	if state.Status != tree.EmbeddingStatusFailed || state.Error == nil || *state.Error != "network down" {
		t.Fatalf("state = %#v", state)
	}
	if repo.lastFailedText == "" {
		t.Fatal("failure was not persisted")
	}
}

func TestRefreshEmbeddingSendsSpecInputAndStoresReady(t *testing.T) {
	fileID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	input := EmbeddingInput{FileID: fileID, Name: "A", Path: "/a", Keywords: []string{"go", "search"}, SearchText: "body"}
	repo := &fakeRepository{inputs: map[uuid.UUID]EmbeddingInput{fileID: input}}
	provider := fakeProvider{embedding: make([]float32, 1024)}
	service := NewService(repo, provider, "text-embedding-v4", 1024)

	state, err := service.RefreshFileEmbedding(context.Background(), fileID)
	if err != nil {
		t.Fatalf("RefreshFileEmbedding() error = %v", err)
	}
	if state.Status != tree.EmbeddingStatusReady {
		t.Fatalf("status = %s", state.Status)
	}
	if repo.lastReadyText != "A\n/a\ngo search\nbody" {
		t.Fatalf("embedding input = %q", repo.lastReadyText)
	}
}

func searchTestFile(id string, path string, name string) tree.FileEntry {
	published := time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC)
	reading := 1
	return tree.FileEntry{
		Node:               tree.Node{ID: uuid.MustParse(id), Kind: tree.NodeKindFile, Name: name, Path: path, Slug: name},
		ContentFormat:      tree.ContentFormatMarkdown,
		Status:             tree.PublishStatusPublished,
		Keywords:           []string{"go"},
		PublishedAt:        &published,
		ReadingTimeMinutes: &reading,
	}
}

type fakeProvider struct {
	embedding []float32
	err       error
}

func (f fakeProvider) Embed(context.Context, string) ([]float32, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.embedding, nil
}

type fakeRepository struct {
	text           []Candidate
	semantic       []Candidate
	semanticVector string
	inputs         map[uuid.UUID]EmbeddingInput
	lastReadyText  string
	lastFailedText string
}

func (f *fakeRepository) FullText(context.Context, string, int) ([]Candidate, error) {
	return f.text, nil
}
func (f *fakeRepository) Semantic(_ context.Context, vector string, _ int) ([]Candidate, error) {
	f.semanticVector = vector
	return f.semantic, nil
}
func (f *fakeRepository) EmbeddingInput(_ context.Context, fileID uuid.UUID) (EmbeddingInput, error) {
	input, ok := f.inputs[fileID]
	if !ok {
		return EmbeddingInput{}, ErrEmbeddingNotFound
	}
	return input, nil
}
func (f *fakeRepository) PublishedEmbeddingInputs(context.Context) ([]EmbeddingInput, error) {
	inputs := make([]EmbeddingInput, 0, len(f.inputs))
	for _, input := range f.inputs {
		inputs = append(inputs, input)
	}
	return inputs, nil
}
func (f *fakeRepository) SetEmbeddingReady(_ context.Context, fileID uuid.UUID, model string, embedding []float32) (EmbeddingState, error) {
	f.lastReadyText = embeddingText(f.inputs[fileID])
	return EmbeddingState{FileID: fileID, Provider: ProviderQwen, Model: model, Dimensions: 1024, Status: tree.EmbeddingStatusReady}, nil
}
func (f *fakeRepository) SetEmbeddingFailed(_ context.Context, fileID uuid.UUID, model string, message string) (EmbeddingState, error) {
	f.lastFailedText = message
	return EmbeddingState{FileID: fileID, Provider: ProviderQwen, Model: model, Dimensions: 1024, Status: tree.EmbeddingStatusFailed, Error: &message}, nil
}
