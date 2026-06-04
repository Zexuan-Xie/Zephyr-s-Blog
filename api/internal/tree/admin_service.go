package tree

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidNodeInput   = errors.New("invalid node input")
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
	SortOrder     int           `json:"sort_order"`
	ContentFormat ContentFormat `json:"content_format,omitempty"`
}

type UpdateNodeInput struct {
	ParentID    *uuid.UUID `json:"parent_id"`
	ParentIDSet bool       `json:"-"`
	Name        *string    `json:"name"`
	Slug        *string    `json:"slug"`
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

type PathChangeRecorder interface {
	RecordPathChange(ctx context.Context, nodeID uuid.UUID, oldPath, newPath string) error
}

type AdminService struct {
	repo      AdminRepository
	redirects PathChangeRecorder
}

func NewAdminService(repo AdminRepository, redirects PathChangeRecorder) *AdminService {
	return &AdminService{repo: repo, redirects: redirects}
}

func (s *AdminService) GetNode(ctx context.Context, nodeID uuid.UUID) (AdminNodeDetail, error) {
	return s.repo.GetAdminNode(ctx, nodeID)
}

func (s *AdminService) CreateNode(ctx context.Context, input CreateNodeInput) (AdminNodeDetail, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.TrimSpace(input.Slug)
	if err := validateCreateNodeInput(input); err != nil {
		return AdminNodeDetail{}, err
	}
	return s.repo.CreateNode(ctx, input)
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
	if err := validateUpdateNodeInput(current.Node, input); err != nil {
		return AdminNodeDetail{}, err
	}

	updated, err := s.repo.UpdateNode(ctx, nodeID, input)
	if err != nil {
		return AdminNodeDetail{}, err
	}
	if s.redirects != nil && current.Node.Path != updated.Node.Path {
		if err := s.redirects.RecordPathChange(ctx, nodeID, current.Node.Path, updated.Node.Path); err != nil {
			return AdminNodeDetail{}, err
		}
	}
	return updated, nil
}

func validateCreateNodeInput(input CreateNodeInput) error {
	if !validNodeKind(input.Kind) || !validNodeNameAndSlug(input.Name, input.Slug) {
		return ErrInvalidNodeInput
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
	if !validNodeNameAndSlug(name, slug) {
		return ErrInvalidNodeInput
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

func validNodeNameAndSlug(name, slug string) bool {
	return strings.TrimSpace(name) != "" &&
		strings.TrimSpace(slug) != "" &&
		slug != "." &&
		slug != ".." &&
		!strings.Contains(slug, "/")
}

func isReservedRootSlug(slug string) bool {
	_, ok := reservedRootSlugs[strings.ToLower(strings.TrimSpace(slug))]
	return ok
}
