package tree

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidNodeInput  = errors.New("invalid node input")
	ErrReservedRootSlug  = errors.New("reserved root slug")
	ErrDuplicateSlug     = errors.New("duplicate slug under parent")
	ErrParentNotDirectory = errors.New("parent node is not a directory")
	ErrNodeCycle         = errors.New("node move would create a cycle")
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
	ParentID     *uuid.UUID    `json:"parent_id"`
	Kind         NodeKind      `json:"kind"`
	Name         string        `json:"name"`
	Slug         string        `json:"slug"`
	SortOrder    int           `json:"sort_order"`
	ContentFormat ContentFormat `json:"content_format,omitempty"`
}

type AdminRepository interface {
	CreateNode(ctx context.Context, input CreateNodeInput) (Node, error)
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

func (s *AdminService) CreateNode(ctx context.Context, input CreateNodeInput) (AdminNodeDetail, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.TrimSpace(input.Slug)
	if input.ParentID == nil && isReservedRootSlug(input.Slug) {
		return AdminNodeDetail{}, ErrReservedRootSlug
	}
	node, err := s.repo.CreateNode(ctx, input)
	if err != nil {
		return AdminNodeDetail{}, err
	}
	return AdminNodeDetail{Node: node}, nil
}

func isReservedRootSlug(slug string) bool {
	_, ok := reservedRootSlugs[strings.ToLower(strings.TrimSpace(slug))]
	return ok
}

type AdminNodeDetail struct {
	Node Node `json:"node"`
}
