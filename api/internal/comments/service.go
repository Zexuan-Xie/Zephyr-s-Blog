package comments

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"xlab-blog/api/internal/users"
)

type Repository interface {
	PublishedFileExists(ctx context.Context, fileID uuid.UUID) (bool, error)
	ListThread(ctx context.Context, fileID uuid.UUID, viewerID *uuid.UUID) (Thread, error)
	FindComment(ctx context.Context, commentID uuid.UUID) (Comment, error)
	InsertComment(ctx context.Context, fileID uuid.UUID, userID uuid.UUID, input CreateInput) (Comment, error)
	SoftDeleteComment(ctx context.Context, commentID uuid.UUID, deletedBy uuid.UUID) (Comment, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Thread(ctx context.Context, fileID uuid.UUID, viewerID *uuid.UUID) (Thread, error) {
	exists, err := s.repo.PublishedFileExists(ctx, fileID)
	if err != nil {
		return Thread{}, err
	}
	if !exists {
		return Thread{}, ErrFileNotFound
	}
	return s.repo.ListThread(ctx, fileID, viewerID)
}

func (s *Service) Create(ctx context.Context, fileID uuid.UUID, userID uuid.UUID, input CreateInput) (Comment, error) {
	body, err := normalizeBody(input.Body)
	if err != nil {
		return Comment{}, err
	}
	exists, err := s.repo.PublishedFileExists(ctx, fileID)
	if err != nil {
		return Comment{}, err
	}
	if !exists {
		return Comment{}, ErrFileNotFound
	}
	input.Body = body
	if input.ParentID != nil {
		parent, err := s.repo.FindComment(ctx, *input.ParentID)
		if err != nil {
			return Comment{}, err
		}
		if parent.FileNodeID != fileID {
			return Comment{}, ErrParentMismatch
		}
		if parent.ParentID != nil {
			input.ParentID = parent.ParentID
			if input.ReplyToUserID == nil {
				replyToUserID := parent.User.ID
				input.ReplyToUserID = &replyToUserID
			}
		}
	}
	return s.repo.InsertComment(ctx, fileID, userID, input)
}

func (s *Service) Delete(ctx context.Context, commentID uuid.UUID, actor users.User) (Comment, error) {
	comment, err := s.repo.FindComment(ctx, commentID)
	if err != nil {
		return Comment{}, err
	}
	if actor.Role != users.RoleAdmin && comment.User.ID != actor.ID {
		return Comment{}, ErrPermissionDenied
	}
	return s.repo.SoftDeleteComment(ctx, commentID, actor.ID)
}

func normalizeBody(body string) (string, error) {
	body = strings.TrimSpace(body)
	if body == "" || len(body) > MaxBodyLength || !utf8.ValidString(body) {
		return "", ErrInvalidCommentBody
	}
	for _, r := range body {
		if r == 0 || (r < 0x20 && r != '\n' && r != '\r' && r != '\t') {
			return "", ErrInvalidCommentBody
		}
	}
	return body, nil
}
