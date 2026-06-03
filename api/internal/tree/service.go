package tree

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Root(ctx context.Context) (DirectoryPage, error) {
	return s.repo.DirectoryPage(ctx, nil)
}

func (s *Service) Children(ctx context.Context, nodeID uuid.UUID) (DirectoryPage, error) {
	return s.repo.DirectoryPage(ctx, &nodeID)
}

func (s *Service) Resolve(ctx context.Context, rawPath string) (ResolveResponse, error) {
	path, err := NormalizePath(rawPath)
	if err != nil {
		return ResolveResponse{}, err
	}
	if path == "/" {
		page, err := s.repo.DirectoryPage(ctx, nil)
		if err != nil {
			return ResolveResponse{}, err
		}
		return ResolveResponse{Type: ResolveTypeDirectory, Directory: &page}, nil
	}

	slugs := strings.Split(strings.TrimPrefix(path, "/"), "/")
	var parentID *uuid.UUID
	var node Node
	for _, slug := range slugs {
		node, err = s.repo.FindNodeByParentAndSlug(ctx, parentID, slug)
		if err != nil {
			return s.redirectOrNotFound(ctx, path)
		}
		currentID := node.ID
		parentID = &currentID
	}

	switch node.Kind {
	case NodeKindDirectory:
		page, err := s.repo.DirectoryPage(ctx, &node.ID)
		if err != nil {
			return ResolveResponse{}, err
		}
		return ResolveResponse{Type: ResolveTypeDirectory, Directory: &page}, nil
	case NodeKindFile:
		page, err := s.repo.FilePage(ctx, node)
		if err != nil {
			if err == ErrNotFound {
				return s.redirectOrNotFound(ctx, path)
			}
			return ResolveResponse{}, err
		}
		return ResolveResponse{Type: ResolveTypeFile, File: &page}, nil
	default:
		return ResolveResponse{}, ErrNotFound
	}
}

func (s *Service) redirectOrNotFound(ctx context.Context, path string) (ResolveResponse, error) {
	newPath, err := s.repo.RedirectPath(ctx, path)
	if err == nil {
		return ResolveResponse{Type: ResolveTypeRedirect, NewPath: newPath}, nil
	}
	return ResolveResponse{}, ErrNotFound
}

func NormalizePath(rawPath string) (string, error) {
	path := strings.TrimSpace(rawPath)
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		return "", ErrInvalidPath
	}

	parts := strings.FieldsFunc(path, func(r rune) bool { return r == '/' })
	if len(parts) == 0 {
		return "/", nil
	}
	for _, part := range parts {
		if part == "." || part == ".." {
			return "", ErrInvalidPath
		}
	}
	return "/" + strings.Join(parts, "/"), nil
}

func PublicKeywords(keywords []string) []string {
	limit := 3
	if len(keywords) < limit {
		limit = len(keywords)
	}
	out := make([]string, limit)
	copy(out, keywords[:limit])
	return out
}
