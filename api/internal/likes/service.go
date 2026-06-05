package likes

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	FileTargetExists(ctx context.Context, fileID uuid.UUID) (bool, error)
	CommentTargetState(ctx context.Context, commentID uuid.UUID) (exists bool, deleted bool, err error)
	UpsertLike(ctx context.Context, userID uuid.UUID, target Target) error
	DeleteLike(ctx context.Context, userID uuid.UUID, target Target) error
	LikeState(ctx context.Context, userID uuid.UUID, target Target) (State, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) LikeFile(ctx context.Context, userID uuid.UUID, fileID uuid.UUID) (State, error) {
	target := Target{Type: TargetFile, ID: fileID}
	if err := s.ensureTarget(ctx, target, true); err != nil {
		return State{}, err
	}
	if err := s.repo.UpsertLike(ctx, userID, target); err != nil {
		return State{}, err
	}
	return s.repo.LikeState(ctx, userID, target)
}

func (s *Service) UnlikeFile(ctx context.Context, userID uuid.UUID, fileID uuid.UUID) (State, error) {
	target := Target{Type: TargetFile, ID: fileID}
	if err := s.ensureTarget(ctx, target, false); err != nil {
		return State{}, err
	}
	if err := s.repo.DeleteLike(ctx, userID, target); err != nil {
		return State{}, err
	}
	return s.repo.LikeState(ctx, userID, target)
}

func (s *Service) LikeComment(ctx context.Context, userID uuid.UUID, commentID uuid.UUID) (State, error) {
	target := Target{Type: TargetComment, ID: commentID}
	if err := s.ensureTarget(ctx, target, true); err != nil {
		return State{}, err
	}
	if err := s.repo.UpsertLike(ctx, userID, target); err != nil {
		return State{}, err
	}
	return s.repo.LikeState(ctx, userID, target)
}

func (s *Service) UnlikeComment(ctx context.Context, userID uuid.UUID, commentID uuid.UUID) (State, error) {
	target := Target{Type: TargetComment, ID: commentID}
	if err := s.ensureTarget(ctx, target, false); err != nil {
		return State{}, err
	}
	if err := s.repo.DeleteLike(ctx, userID, target); err != nil {
		return State{}, err
	}
	return s.repo.LikeState(ctx, userID, target)
}

func (s *Service) ensureTarget(ctx context.Context, target Target, forLike bool) error {
	switch target.Type {
	case TargetFile:
		exists, err := s.repo.FileTargetExists(ctx, target.ID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrTargetNotFound
		}
		return nil
	case TargetComment:
		exists, deleted, err := s.repo.CommentTargetState(ctx, target.ID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrTargetNotFound
		}
		if forLike && deleted {
			return ErrTargetDeleted
		}
		return nil
	default:
		return ErrTargetNotFound
	}
}
