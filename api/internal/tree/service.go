package tree

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type LifecycleRepository interface {
	GetNode(ctx context.Context, nodeID uuid.UUID) (Node, error)
	GetFileContent(ctx context.Context, nodeID uuid.UUID) (FileContent, error)
	UpsertFileContent(ctx context.Context, nodeID uuid.UUID, input UpsertFileContentInput) (FileContent, error)
	PublishFile(ctx context.Context, nodeID uuid.UUID) (FileContent, error)
	UnpublishFile(ctx context.Context, nodeID uuid.UUID) (FileContent, error)
	DeleteNode(ctx context.Context, nodeID uuid.UUID) error
	HasPublishedDescendantFiles(ctx context.Context, directoryID uuid.UUID) (bool, error)
	PublishedDescendantFilePaths(ctx context.Context, directoryID uuid.UUID) ([]PublishedFilePath, error)
	UpdateRedirectTargets(ctx context.Context, nodeID uuid.UUID, finalPath string) error
	UpsertPathRedirect(ctx context.Context, oldPath, newPath string, nodeID uuid.UUID) error
}

type LifecycleService struct {
	repo LifecycleRepository
}

func NewLifecycleService(repo LifecycleRepository) *LifecycleService {
	return &LifecycleService{repo: repo}
}

func (s *LifecycleService) UpsertFileContent(ctx context.Context, nodeID uuid.UUID, input UpsertFileContentInput) (FileContent, error) {
	node, err := s.repo.GetNode(ctx, nodeID)
	if err != nil {
		return FileContent{}, err
	}
	if node.Kind != NodeKindFile {
		return FileContent{}, ErrNodeIsNotFile
	}

	existing, err := s.repo.GetFileContent(ctx, nodeID)
	if err != nil && err != ErrFileContentNotFound {
		return FileContent{}, err
	}
	if err == nil && existing.Status == PublishStatusPublished && existing.ContentFormat != input.ContentFormat {
		return FileContent{}, ErrPublishedContentFormatChange
	}

	input.Keywords = normalizeKeywords(input.Keywords)
	if strings.TrimSpace(input.SearchText) == "" {
		input.SearchText = buildSearchText(node, input)
	}
	return s.repo.UpsertFileContent(ctx, nodeID, input)
}

func (s *LifecycleService) PublishFile(ctx context.Context, nodeID uuid.UUID) (FileContent, error) {
	node, err := s.repo.GetNode(ctx, nodeID)
	if err != nil {
		return FileContent{}, err
	}
	if node.Kind != NodeKindFile {
		return FileContent{}, ErrNodeIsNotFile
	}
	if _, err := s.repo.GetFileContent(ctx, nodeID); err != nil {
		return FileContent{}, err
	}
	return s.repo.PublishFile(ctx, nodeID)
}

func (s *LifecycleService) UnpublishFile(ctx context.Context, nodeID uuid.UUID) (FileContent, error) {
	node, err := s.repo.GetNode(ctx, nodeID)
	if err != nil {
		return FileContent{}, err
	}
	if node.Kind != NodeKindFile {
		return FileContent{}, ErrNodeIsNotFile
	}
	return s.repo.UnpublishFile(ctx, nodeID)
}

func (s *LifecycleService) DeleteNode(ctx context.Context, nodeID uuid.UUID) error {
	node, err := s.repo.GetNode(ctx, nodeID)
	if err != nil {
		return err
	}

	if node.Kind == NodeKindFile {
		content, err := s.repo.GetFileContent(ctx, nodeID)
		if err != nil && err != ErrFileContentNotFound {
			return err
		}
		if err == nil && content.Status == PublishStatusPublished {
			return ErrPublishedFileDelete
		}
		return s.repo.DeleteNode(ctx, nodeID)
	}

	hasPublished, err := s.repo.HasPublishedDescendantFiles(ctx, nodeID)
	if err != nil {
		return err
	}
	if hasPublished {
		return ErrDirectoryHasPublishedDescendants
	}
	return s.repo.DeleteNode(ctx, nodeID)
}

func (s *LifecycleService) RecordPathChange(ctx context.Context, nodeID uuid.UUID, oldPath string, newPath string) error {
	oldPath = normalizePath(oldPath)
	newPath = normalizePath(newPath)
	if oldPath == "/" || newPath == "/" || oldPath == newPath {
		return nil
	}

	node, err := s.repo.GetNode(ctx, nodeID)
	if err != nil {
		return err
	}

	if node.Kind == NodeKindFile {
		content, err := s.repo.GetFileContent(ctx, nodeID)
		if err != nil {
			if err == ErrFileContentNotFound {
				return nil
			}
			return err
		}
		if content.Status != PublishStatusPublished {
			return nil
		}
		if err := s.repo.UpdateRedirectTargets(ctx, nodeID, newPath); err != nil {
			return err
		}
		return s.repo.UpsertPathRedirect(ctx, oldPath, newPath, nodeID)
	}

	files, err := s.repo.PublishedDescendantFilePaths(ctx, nodeID)
	if err != nil {
		return err
	}
	for _, file := range files {
		finalPath := normalizePath(file.Path)
		oldFilePath := replacePathPrefix(finalPath, newPath, oldPath)
		if oldFilePath == finalPath {
			continue
		}
		if err := s.repo.UpdateRedirectTargets(ctx, file.NodeID, finalPath); err != nil {
			return err
		}
		if err := s.repo.UpsertPathRedirect(ctx, oldFilePath, finalPath, file.NodeID); err != nil {
			return err
		}
	}
	return nil
}

func normalizeKeywords(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, value)
	}
	return out
}

func buildSearchText(node Node, input UpsertFileContentInput) string {
	parts := []string{node.Name, node.Path, strings.Join(input.Keywords, " "), input.BodyRaw}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}
