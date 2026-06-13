package tree

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidNodeInput   = errors.New("invalid node input")
	ErrNodeNameRequired   = fmt.Errorf("%w: node name is required", ErrInvalidNodeInput)
	ErrNodeSlugRequired   = fmt.Errorf("%w: node slug is required", ErrInvalidNodeInput)
	ErrInvalidNodeKind    = fmt.Errorf("%w: node kind must be directory or file", ErrInvalidNodeInput)
	ErrInvalidNodeSlug    = fmt.Errorf("%w: node slug must not be '.', '..', or contain '/'", ErrInvalidNodeInput)
	ErrReservedRootSlug   = errors.New("reserved root slug")
	ErrDuplicateSlug      = errors.New("duplicate slug under parent")
	ErrParentNotDirectory = errors.New("parent node is not a directory")
	ErrNodeCycle          = errors.New("node move would create a cycle")
)

var reservedRootSlugs = map[string]struct{}{
	"admin":    {},
	"api":      {},
	"auth":     {},
	"login":    {},
	"register": {},
	"recent":   {},
	"search":   {},
	"settings": {},
}

type CreateNodeInput struct {
	ParentID      *uuid.UUID    `json:"parent_id"`
	Kind          NodeKind      `json:"kind"`
	Name          string        `json:"name"`
	Slug          string        `json:"slug"`
	SlugSet       bool          `json:"-"`
	SortOrder     int           `json:"sort_order"`
	ContentFormat ContentFormat `json:"content_format,omitempty"`
}

func (input *CreateNodeInput) UnmarshalJSON(data []byte) error {
	type createNodeAlias CreateNodeInput
	var decoded createNodeAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	*input = CreateNodeInput(decoded)
	_, input.SlugSet = fields["slug"]
	return nil
}

type UpdateNodeInput struct {
	ParentID    *uuid.UUID `json:"parent_id"`
	ParentIDSet bool       `json:"-"`
	Name        *string    `json:"name"`
	Slug        *string    `json:"slug"`
	URLPath     *string    `json:"url_path"`
	SortOrder   *int       `json:"sort_order"`
}

func (input *UpdateNodeInput) UnmarshalJSON(data []byte) error {
	type updateNodeAlias UpdateNodeInput
	var decoded updateNodeAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	*input = UpdateNodeInput(decoded)
	_, input.ParentIDSet = fields["parent_id"]
	return nil
}

type AdminNodeDetail struct {
	Node             Node           `json:"node"`
	Content          *FileContent   `json:"content,omitempty"`
	Assets           []FileAsset    `json:"assets"`
	RedirectsCreated []PathRedirect `json:"redirects_created"`
}

type PathRedirect struct {
	ID        uuid.UUID `json:"id"`
	OldPath   string    `json:"old_path"`
	NewPath   string    `json:"new_path"`
	NodeID    uuid.UUID `json:"node_id"`
	CreatedAt time.Time `json:"created_at"`
}

type AdminRepository interface {
	GetAdminNode(ctx context.Context, nodeID uuid.UUID) (AdminNodeDetail, error)
	CreateNode(ctx context.Context, input CreateNodeInput) (AdminNodeDetail, error)
	UpdateNode(ctx context.Context, nodeID uuid.UUID, input UpdateNodeInput) (AdminNodeDetail, error)
}

type AdminTreeRepository interface {
	AdminTree(ctx context.Context) (AdminTreeResponse, error)
}

type ReorderChildrenRepository interface {
	ReorderChildren(ctx context.Context, parentID uuid.UUID, input ReorderChildrenInput) (ReorderChildrenResult, error)
}

type MoveNodeRepository interface {
	PreviewMove(ctx context.Context, nodeID uuid.UUID, input MoveNodeInput) (MovePreview, error)
	MoveNode(ctx context.Context, nodeID uuid.UUID, input MoveNodeInput) (AdminNodeDetail, error)
}

type DeleteNodeRepository interface {
	DeleteNode(ctx context.Context, nodeID uuid.UUID) error
}

type PathChangeRecorder interface {
	RecordPathChange(ctx context.Context, nodeID uuid.UUID, oldPath, newPath string) error
}

type atomicPathChangeRepository interface {
	recordsPathChangesAtomically()
}

type AdminService struct {
	repo      AdminRepository
	redirects PathChangeRecorder
}

func NewAdminService(repo AdminRepository, redirects PathChangeRecorder) *AdminService {
	return &AdminService{repo: repo, redirects: redirects}
}

func (s *AdminService) AdminTree(ctx context.Context) (AdminTreeResponse, error) {
	if repo, ok := s.repo.(AdminTreeRepository); ok {
		return repo.AdminTree(ctx)
	}
	return AdminTreeResponse{Nodes: []AdminTreeNode{}}, nil
}

func (s *AdminService) GetNode(ctx context.Context, nodeID uuid.UUID) (AdminNodeDetail, error) {
	return s.repo.GetAdminNode(ctx, nodeID)
}

func (s *AdminService) CreateNode(ctx context.Context, input CreateNodeInput) (AdminNodeDetail, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.TrimSpace(input.Slug)
	autoSlug := input.Slug == "" && !input.SlugSet
	if autoSlug {
		input.Slug = generateURLSegment(input.Name)
	}
	if err := validateCreateNodeInput(input); err != nil {
		return AdminNodeDetail{}, err
	}
	created, err := s.repo.CreateNode(ctx, input)
	if err == nil || !autoSlug || !errors.Is(err, ErrDuplicateSlug) {
		return created, err
	}
	baseSlug := input.Slug
	for suffix := 2; suffix <= 99; suffix++ {
		input.Slug = fmt.Sprintf("%s-%d", baseSlug, suffix)
		created, err = s.repo.CreateNode(ctx, input)
		if err == nil || !errors.Is(err, ErrDuplicateSlug) {
			return created, err
		}
	}
	return AdminNodeDetail{}, ErrDuplicateSlug
}

func (s *AdminService) UpdateNode(ctx context.Context, nodeID uuid.UUID, input UpdateNodeInput) (AdminNodeDetail, error) {
	current, err := s.repo.GetAdminNode(ctx, nodeID)
	if err != nil {
		return AdminNodeDetail{}, err
	}
	if input.ParentID != nil {
		input.ParentIDSet = true
	}

	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		input.Name = &trimmed
	}
	if input.Slug != nil {
		trimmed := strings.TrimSpace(*input.Slug)
		input.Slug = &trimmed
	}
	if input.URLPath != nil {
		trimmed := strings.Trim(strings.TrimSpace(*input.URLPath), "/")
		input.URLPath = &trimmed
		input.Slug = &trimmed
	}
	if err := validateUpdateNodeInput(current.Node, input); err != nil {
		return AdminNodeDetail{}, err
	}

	updated, err := s.repo.UpdateNode(ctx, nodeID, input)
	if err != nil {
		return AdminNodeDetail{}, err
	}
	_, redirectsRecordedAtomically := s.repo.(atomicPathChangeRepository)
	if !redirectsRecordedAtomically && s.redirects != nil && current.Node.Path != updated.Node.Path {
		if err := s.redirects.RecordPathChange(ctx, nodeID, current.Node.Path, updated.Node.Path); err != nil {
			return AdminNodeDetail{}, err
		}
	}
	return updated, nil
}

func validateCreateNodeInput(input CreateNodeInput) error {
	if err := validateNodeNameAndSlug(input.Name, input.Slug); err != nil {
		return err
	}
	if !validNodeKind(input.Kind) {
		return ErrInvalidNodeKind
	}
	if input.ParentID == nil && isReservedRootSlug(input.Slug) {
		return ErrReservedRootSlug
	}
	if input.Kind == NodeKindFile && !validContentFormat(input.ContentFormat) {
		return ErrInvalidContentFormat
	}
	return nil
}

func validateUpdateNodeInput(current Node, input UpdateNodeInput) error {
	name := current.Name
	if input.Name != nil {
		name = *input.Name
	}
	slug := current.Slug
	if input.Slug != nil {
		slug = *input.Slug
	}
	if err := validateNodeNameAndSlug(name, slug); err != nil {
		return err
	}
	parentID := current.ParentID
	if input.ParentIDSet {
		parentID = input.ParentID
	}
	if parentID == nil && isReservedRootSlug(slug) {
		return ErrReservedRootSlug
	}
	return nil
}

func validNodeKind(kind NodeKind) bool {
	return kind == NodeKindDirectory || kind == NodeKindFile
}

func validContentFormat(format ContentFormat) bool {
	return format == ContentFormatMarkdown || format == ContentFormatHTMLDocument
}

func validateNodeNameAndSlug(name, slug string) error {
	if strings.TrimSpace(name) == "" {
		return ErrNodeNameRequired
	}
	if strings.TrimSpace(slug) == "" {
		return ErrNodeSlugRequired
	}
	if slug == "." || slug == ".." || strings.Contains(slug, "/") {
		return ErrInvalidNodeSlug
	}
	return nil
}

func isReservedRootSlug(slug string) bool {
	_, ok := reservedRootSlugs[strings.ToLower(strings.TrimSpace(slug))]
	return ok
}

func (s *AdminService) ReorderChildren(ctx context.Context, parentID uuid.UUID, input ReorderChildrenInput) (ReorderChildrenResult, error) {
	if repo, ok := s.repo.(ReorderChildrenRepository); ok {
		return repo.ReorderChildren(ctx, parentID, input)
	}
	return ReorderChildrenResult{ParentID: parentID, ChildIDs: append([]uuid.UUID(nil), input.ChildIDs...), Version: input.ExpectedVersion + 1}, nil
}

func (s *AdminService) PreviewMove(ctx context.Context, nodeID uuid.UUID, input MoveNodeInput) (MovePreview, error) {
	if repo, ok := s.repo.(MoveNodeRepository); ok {
		return repo.PreviewMove(ctx, nodeID, input)
	}
	node, err := s.repo.GetAdminNode(ctx, nodeID)
	if err != nil {
		return MovePreview{}, err
	}
	destination := node.Node.Path
	return MovePreview{NodeID: nodeID, DestinationPath: destination, AffectedPaths: []string{destination}, Redirects: []PathRedirectPreview{{OldPath: destination, NewPath: destination, NodeID: nodeID}}, BlockedReasons: []string{}}, nil
}

func (s *AdminService) MoveNode(ctx context.Context, nodeID uuid.UUID, input MoveNodeInput) (AdminNodeDetail, error) {
	if repo, ok := s.repo.(MoveNodeRepository); ok {
		return repo.MoveNode(ctx, nodeID, input)
	}
	update := UpdateNodeInput{ParentID: input.NewParentID, ParentIDSet: true}
	return s.UpdateNode(ctx, nodeID, update)
}

func generateURLSegment(name string) string {
	fields := strings.Fields(strings.TrimSpace(name))
	if len(fields) == 0 {
		return ""
	}
	return strings.Join(fields, "-")
}

func (s *AdminService) DeleteNode(ctx context.Context, nodeID uuid.UUID) error {
	if repo, ok := s.repo.(DeleteNodeRepository); ok {
		return repo.DeleteNode(ctx, nodeID)
	}
	return ErrNodeNotFound
}
