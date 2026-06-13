package search

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"xlab-blog/api/internal/tree"
)

const searchNodePathsCTE = `
	with recursive node_paths as (
		select id, parent_id, kind, name, slug, sort_order, created_at, updated_at,
			('/' || slug)::text as path
		from nodes
		where parent_id is null
		union all
		select n.id, n.parent_id, n.kind, n.name, n.slug, n.sort_order, n.created_at, n.updated_at,
			(np.path || '/' || n.slug)::text as path
		from nodes n
		join node_paths np on np.id = n.parent_id
	)`

type SQLRepository struct {
	pool *pgxpool.Pool
}

func NewSQLRepository(pool *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{pool: pool}
}

func (r *SQLRepository) FullText(ctx context.Context, query string, limit int) ([]Candidate, error) {
	const sqlQuery = searchNodePathsCTE + `,
	query_terms as (select websearch_to_tsquery('simple', $1) as q)
	select p.id, p.parent_id, p.kind, p.name, p.slug, p.path, p.sort_order, p.created_at, p.updated_at,
		pfc.content_format, 'published'::text as status, coalesce(pfc.keywords, '{}'::text[]) as keywords, pfc.published_at,
		coalesce((select count(*) from likes l where l.target_type = 'file' and l.target_id = p.id), 0) as like_count,
		coalesce((select count(*) from comments c where c.file_node_id = p.id and c.deleted_at is null), 0) as comment_count,
		pfc.search_text,
		ts_headline('simple', concat_ws(' ', p.name, p.path, array_to_string(pfc.keywords, ' '), pfc.search_text), query_terms.q,
			'StartSel=<mark>, StopSel=</mark>, MaxWords=28, MinWords=8') as snippet,
		ts_rank((setweight(to_tsvector('simple', coalesce(p.name, '')), 'A') ||
			setweight(to_tsvector('simple', coalesce(p.path, '')), 'A') || pfc.search_vector), query_terms.q) as rank_score,
		exists(select 1 from unnest(pfc.keywords) kw where lower(kw) like '%' || lower($1) || '%') as keyword_match
	from node_paths p
	join published_file_contents pfc on pfc.node_id = p.id and pfc.visible
	cross join query_terms
	where p.kind = 'file'
		and ((setweight(to_tsvector('simple', coalesce(p.name, '')), 'A') ||
			setweight(to_tsvector('simple', coalesce(p.path, '')), 'A') || pfc.search_vector) @@ query_terms.q)
	order by rank_score desc, fc.published_at desc nulls last, p.name
	limit $2`
	rows, err := r.pool.Query(ctx, sqlQuery, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCandidates(rows, SourceText)
}

func (r *SQLRepository) Semantic(ctx context.Context, vector string, limit int) ([]Candidate, error) {
	const sqlQuery = searchNodePathsCTE + `
	select p.id, p.parent_id, p.kind, p.name, p.slug, p.path, p.sort_order, p.created_at, p.updated_at,
		pfc.content_format, 'published'::text as status, coalesce(pfc.keywords, '{}'::text[]) as keywords, pfc.published_at,
		coalesce((select count(*) from likes l where l.target_type = 'file' and l.target_id = p.id), 0) as like_count,
		coalesce((select count(*) from comments c where c.file_node_id = p.id and c.deleted_at is null), 0) as comment_count,
		pfc.search_text,
		left(pfc.search_text, 240) as snippet,
		(1 - (pfc.embedding <=> $1::vector))::float8 as rank_score,
		false as keyword_match
	from node_paths p
	join published_file_contents pfc on pfc.node_id = p.id and pfc.visible
	where p.kind = 'file'
		and pfc.embedding_status = 'ready'
		and pfc.embedding is not null
	order by pfc.embedding <=> $1::vector asc, fc.published_at desc nulls last, p.name
	limit $2`
	rows, err := r.pool.Query(ctx, sqlQuery, vector, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCandidates(rows, SourceSemantic)
}

func (r *SQLRepository) EmbeddingInput(ctx context.Context, fileID uuid.UUID) (EmbeddingInput, error) {
	const query = searchNodePathsCTE + `
	select p.id, p.name, p.path, coalesce(fc.keywords, '{}'::text[]) as keywords, pfc.search_text
	from node_paths p
	join published_file_contents pfc on pfc.node_id = p.id and pfc.visible
	where p.id = $1 and p.kind = 'file'`
	input, err := scanEmbeddingInput(r.pool.QueryRow(ctx, query, fileID))
	if errors.Is(err, pgx.ErrNoRows) {
		return EmbeddingInput{}, ErrEmbeddingNotFound
	}
	return input, err
}

func (r *SQLRepository) PublishedEmbeddingInputs(ctx context.Context) ([]EmbeddingInput, error) {
	const query = searchNodePathsCTE + `
	select p.id, p.name, p.path, coalesce(fc.keywords, '{}'::text[]) as keywords, pfc.search_text
	from node_paths p
	join published_file_contents pfc on pfc.node_id = p.id and pfc.visible
	where p.kind = 'file' and fc.status = 'published'
	order by p.path`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	inputs := make([]EmbeddingInput, 0)
	for rows.Next() {
		input, err := scanEmbeddingInput(rows)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, input)
	}
	return inputs, rows.Err()
}

func (r *SQLRepository) SetEmbeddingReady(ctx context.Context, fileID uuid.UUID, model string, embedding []float32) (EmbeddingState, error) {
	const query = `
	update published_file_contents
	set embedding = $2::vector,
		embedding_model = $3,
		embedding_status = 'ready',
		embedding_error = null,
		embedding_updated_at = now()
	where node_id = $1
	returning node_id, embedding_status, embedding_error, embedding_updated_at`
	state, err := scanEmbeddingState(r.pool.QueryRow(ctx, query, fileID, vectorLiteral(embedding), model), model)
	if errors.Is(err, pgx.ErrNoRows) {
		return EmbeddingState{}, ErrEmbeddingNotFound
	}
	return state, err
}

func (r *SQLRepository) SetEmbeddingFailed(ctx context.Context, fileID uuid.UUID, model string, message string) (EmbeddingState, error) {
	const query = `
	update published_file_contents
	set embedding = null,
		embedding_model = $2,
		embedding_status = 'failed',
		embedding_error = $3,
		embedding_updated_at = now()
	where node_id = $1
	returning node_id, embedding_status, embedding_error, embedding_updated_at`
	state, err := scanEmbeddingState(r.pool.QueryRow(ctx, query, fileID, model, message), model)
	if errors.Is(err, pgx.ErrNoRows) {
		return EmbeddingState{}, ErrEmbeddingNotFound
	}
	return state, err
}

type rowScanner interface {
	Scan(dest ...any) error
}

type rowIterator interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanCandidates(rows rowIterator, source string) ([]Candidate, error) {
	candidates := make([]Candidate, 0)
	rank := 1
	for rows.Next() {
		candidate, err := scanCandidate(rows, source, rank)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
		rank++
	}
	return candidates, rows.Err()
}

func scanCandidate(row rowScanner, source string, rank int) (Candidate, error) {
	var node tree.Node
	var parentID uuid.NullUUID
	var kind string
	var contentFormat string
	var status string
	var keywords []string
	var publishedAt sql.NullTime
	var likeCount int
	var commentCount int
	var searchText string
	var snippet sql.NullString
	var score float64
	var keywordMatch bool
	if err := row.Scan(
		&node.ID,
		&parentID,
		&kind,
		&node.Name,
		&node.Slug,
		&node.Path,
		&node.SortOrder,
		&node.CreatedAt,
		&node.UpdatedAt,
		&contentFormat,
		&status,
		&keywords,
		&publishedAt,
		&likeCount,
		&commentCount,
		&searchText,
		&snippet,
		&score,
		&keywordMatch,
	); err != nil {
		return Candidate{}, err
	}
	if parentID.Valid {
		node.ParentID = &parentID.UUID
	}
	node.Kind = tree.NodeKind(kind)
	published := (*time.Time)(nil)
	if publishedAt.Valid {
		published = &publishedAt.Time
	}
	readingTime := readingTimeMinutes(searchText)
	return Candidate{
		File: tree.FileEntry{
			Node:               node,
			ContentFormat:      tree.ContentFormat(contentFormat),
			Status:             tree.PublishStatus(status),
			Keywords:           keywords,
			PublishedAt:        published,
			LikeCount:          likeCount,
			CommentCount:       commentCount,
			ReadingTimeMinutes: &readingTime,
		},
		Path:         node.Path,
		Snippet:      fallbackSnippet(snippet.String, searchText),
		Rank:         rank,
		Score:        score,
		Source:       source,
		KeywordMatch: keywordMatch,
	}, nil
}

func scanEmbeddingInput(row rowScanner) (EmbeddingInput, error) {
	var input EmbeddingInput
	if err := row.Scan(&input.FileID, &input.Name, &input.Path, &input.Keywords, &input.SearchText); err != nil {
		return EmbeddingInput{}, err
	}
	return input, nil
}

func scanEmbeddingState(row rowScanner, model string) (EmbeddingState, error) {
	var state EmbeddingState
	var status string
	var embeddingError sql.NullString
	var updatedAt sql.NullTime
	if err := row.Scan(&state.FileID, &status, &embeddingError, &updatedAt); err != nil {
		return EmbeddingState{}, err
	}
	state.Provider = ProviderQwen
	state.Model = model
	state.Dimensions = 1024
	state.Status = tree.EmbeddingStatus(status)
	if embeddingError.Valid {
		state.Error = &embeddingError.String
	}
	if updatedAt.Valid {
		state.UpdatedAt = &updatedAt.Time
	}
	return state, nil
}

func fallbackSnippet(snippet string, searchText string) string {
	snippet = strings.TrimSpace(snippet)
	if snippet != "" {
		return snippet
	}
	text := strings.TrimSpace(searchText)
	if len(text) <= 240 {
		return text
	}
	return text[:240]
}

func readingTimeMinutes(text string) int {
	words := len(strings.Fields(text))
	if words == 0 {
		return 1
	}
	minutes := words / 200
	if words%200 != 0 {
		minutes++
	}
	if minutes < 1 {
		return 1
	}
	return minutes
}
