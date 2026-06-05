package search

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidQuery      = errors.New("invalid search query")
	ErrEmbeddingNotFound = errors.New("file embedding target not found")
)

type Repository interface {
	FullText(context.Context, string, int) ([]Candidate, error)
	Semantic(context.Context, string, int) ([]Candidate, error)
	EmbeddingInput(context.Context, uuid.UUID) (EmbeddingInput, error)
	PublishedEmbeddingInputs(context.Context) ([]EmbeddingInput, error)
	SetEmbeddingReady(context.Context, uuid.UUID, string, []float32) (EmbeddingState, error)
	SetEmbeddingFailed(context.Context, uuid.UUID, string, string) (EmbeddingState, error)
}

type Service struct {
	repo       Repository
	provider   EmbeddingProvider
	providerID string
	model      string
	dimensions int
	rrfK       int
}

func NewService(repo Repository, provider EmbeddingProvider, model string, dimensions int) *Service {
	return &Service{repo: repo, provider: provider, providerID: ProviderQwen, model: model, dimensions: dimensions, rrfK: DefaultRRFK}
}

func (s *Service) Search(ctx context.Context, rawQuery string, options Options) (Response, error) {
	query := strings.TrimSpace(rawQuery)
	if query == "" {
		return Response{}, ErrInvalidQuery
	}
	limit, offset := normalizeWindow(options)
	topK := limit + offset

	textCandidates, err := s.repo.FullText(ctx, query, topK)
	if err != nil {
		return Response{}, err
	}
	semanticCandidates := []Candidate{}
	if s.provider != nil {
		if embedding, err := s.provider.Embed(ctx, query); err == nil && len(embedding) == s.dimensions {
			semanticCandidates, err = s.repo.Semantic(ctx, vectorLiteral(embedding), topK)
			if err != nil {
				return Response{}, err
			}
		}
	}

	items := fuseCandidates(textCandidates, semanticCandidates, s.rrfK)
	if offset >= len(items) {
		items = []Result{}
	} else {
		end := offset + limit
		if end > len(items) {
			end = len(items)
		}
		items = items[offset:end]
	}
	return Response{Query: query, Items: items}, nil
}

func (s *Service) RefreshFileEmbedding(ctx context.Context, fileID uuid.UUID) (EmbeddingState, error) {
	input, err := s.repo.EmbeddingInput(ctx, fileID)
	if err != nil {
		return EmbeddingState{}, err
	}
	return s.refreshInput(ctx, input)
}

func (s *Service) Rebuild(ctx context.Context) (RebuildState, error) {
	inputs, err := s.repo.PublishedEmbeddingInputs(ctx)
	if err != nil {
		return RebuildState{}, err
	}
	for _, input := range inputs {
		if _, err := s.refreshInput(ctx, input); err != nil {
			return RebuildState{}, err
		}
	}
	return RebuildState{Status: "accepted"}, nil
}

func (s *Service) refreshInput(ctx context.Context, input EmbeddingInput) (EmbeddingState, error) {
	if s.provider == nil {
		return s.repo.SetEmbeddingFailed(ctx, input.FileID, s.model, "embedding provider is not configured")
	}
	embedding, err := s.provider.Embed(ctx, embeddingText(input))
	if err != nil {
		return s.repo.SetEmbeddingFailed(ctx, input.FileID, s.model, err.Error())
	}
	if len(embedding) != s.dimensions {
		state, updateErr := s.repo.SetEmbeddingFailed(ctx, input.FileID, s.model, fmt.Sprintf("embedding dimensions = %d, want %d", len(embedding), s.dimensions))
		if updateErr != nil {
			return EmbeddingState{}, updateErr
		}
		return state, nil
	}
	return s.repo.SetEmbeddingReady(ctx, input.FileID, s.model, embedding)
}

func normalizeWindow(options Options) (int, int) {
	limit := options.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	offset := options.Offset
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func embeddingText(input EmbeddingInput) string {
	return strings.TrimSpace(strings.Join([]string{input.Name, input.Path, strings.Join(input.Keywords, " "), input.SearchText}, "\n"))
}

func fuseCandidates(textCandidates []Candidate, semanticCandidates []Candidate, rrfK int) []Result {
	byID := map[uuid.UUID]*Result{}
	sources := map[uuid.UUID]map[string]struct{}{}
	add := func(candidates []Candidate, source string) {
		for index, candidate := range candidates {
			rank := candidate.Rank
			if rank <= 0 {
				rank = index + 1
			}
			result := byID[candidate.File.Node.ID]
			if result == nil {
				byID[candidate.File.Node.ID] = &Result{
					File:    candidate.File,
					Path:    candidate.Path,
					Snippet: candidate.Snippet,
				}
				result = byID[candidate.File.Node.ID]
				sources[candidate.File.Node.ID] = map[string]struct{}{}
			}
			result.Score += 1 / float64(rrfK+rank)
			if result.Snippet == "" || source == SourceText {
				result.Snippet = candidate.Snippet
			}
			sources[candidate.File.Node.ID][source] = struct{}{}
			if candidate.KeywordMatch {
				sources[candidate.File.Node.ID][SourceKeyword] = struct{}{}
			}
		}
	}
	add(textCandidates, SourceText)
	add(semanticCandidates, SourceSemantic)

	items := make([]Result, 0, len(byID))
	for id, result := range byID {
		result.MatchSources = orderedSources(sources[id])
		items = append(items, *result)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Score == items[j].Score {
			return items[i].Path < items[j].Path
		}
		return items[i].Score > items[j].Score
	})
	return items
}

func orderedSources(sourceSet map[string]struct{}) []string {
	order := []string{SourceText, SourceSemantic, SourceKeyword}
	out := make([]string, 0, len(sourceSet))
	for _, source := range order {
		if _, ok := sourceSet[source]; ok {
			out = append(out, source)
		}
	}
	return out
}

func vectorLiteral(values []float32) string {
	parts := make([]string, len(values))
	for index, value := range values {
		parts[index] = fmt.Sprintf("%g", value)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func EmbeddingTextForTest(input EmbeddingInput) string { return embeddingText(input) }
func VectorLiteralForTest(values []float32) string     { return vectorLiteral(values) }
