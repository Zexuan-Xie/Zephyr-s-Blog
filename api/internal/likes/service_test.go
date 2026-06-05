package likes

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestServiceLikeUnlikeFileIsIdempotent(t *testing.T) {
	userID := uuid.New()
	fileID := uuid.New()
	repo := &fakeRepository{files: map[uuid.UUID]bool{fileID: true}}
	service := NewService(repo)

	state, err := service.LikeFile(context.Background(), userID, fileID)
	if err != nil {
		t.Fatalf("LikeFile() error = %v", err)
	}
	if !state.Liked || state.LikeCount != 1 {
		t.Fatalf("LikeFile() state = %+v, want liked count 1", state)
	}
	state, err = service.LikeFile(context.Background(), userID, fileID)
	if err != nil {
		t.Fatalf("second LikeFile() error = %v", err)
	}
	if !state.Liked || state.LikeCount != 1 {
		t.Fatalf("second LikeFile() state = %+v, want stable liked count 1", state)
	}
	state, err = service.UnlikeFile(context.Background(), userID, fileID)
	if err != nil {
		t.Fatalf("UnlikeFile() error = %v", err)
	}
	if state.Liked || state.LikeCount != 0 {
		t.Fatalf("UnlikeFile() state = %+v, want unliked count 0", state)
	}
	state, err = service.UnlikeFile(context.Background(), userID, fileID)
	if err != nil {
		t.Fatalf("second UnlikeFile() error = %v", err)
	}
	if state.Liked || state.LikeCount != 0 {
		t.Fatalf("second UnlikeFile() state = %+v, want stable unliked count 0", state)
	}
}

func TestServiceLikeCommentRejectsMissingAndDeletedTargets(t *testing.T) {
	userID := uuid.New()
	missingID := uuid.New()
	deletedID := uuid.New()
	repo := &fakeRepository{comments: map[uuid.UUID]bool{deletedID: true}}
	service := NewService(repo)

	if _, err := service.LikeComment(context.Background(), userID, missingID); !errors.Is(err, ErrTargetNotFound) {
		t.Fatalf("missing comment error = %v, want ErrTargetNotFound", err)
	}
	if _, err := service.LikeComment(context.Background(), userID, deletedID); !errors.Is(err, ErrTargetDeleted) {
		t.Fatalf("deleted comment error = %v, want ErrTargetDeleted", err)
	}
	if _, err := service.UnlikeComment(context.Background(), userID, deletedID); err != nil {
		t.Fatalf("UnlikeComment(deleted) error = %v, want nil idempotent unlike", err)
	}
}

func TestServiceRejectsMissingFileTarget(t *testing.T) {
	if _, err := NewService(&fakeRepository{}).LikeFile(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, ErrTargetNotFound) {
		t.Fatalf("LikeFile missing error = %v, want ErrTargetNotFound", err)
	}
}

type fakeRepository struct {
	files    map[uuid.UUID]bool
	comments map[uuid.UUID]bool
	likes    map[likeKey]struct{}
}

func (f *fakeRepository) FileTargetExists(_ context.Context, fileID uuid.UUID) (bool, error) {
	return f.files[fileID], nil
}

func (f *fakeRepository) CommentTargetState(_ context.Context, commentID uuid.UUID) (bool, bool, error) {
	deleted, ok := f.comments[commentID]
	return ok, deleted, nil
}

func (f *fakeRepository) UpsertLike(_ context.Context, userID uuid.UUID, target Target) error {
	if f.likes == nil {
		f.likes = map[likeKey]struct{}{}
	}
	f.likes[likeKey{userID: userID, targetType: target.Type, targetID: target.ID}] = struct{}{}
	return nil
}

func (f *fakeRepository) DeleteLike(_ context.Context, userID uuid.UUID, target Target) error {
	delete(f.likes, likeKey{userID: userID, targetType: target.Type, targetID: target.ID})
	return nil
}

func (f *fakeRepository) LikeState(_ context.Context, userID uuid.UUID, target Target) (State, error) {
	state := State{}
	for key := range f.likes {
		if key.targetType == target.Type && key.targetID == target.ID {
			state.LikeCount++
			if key.userID == userID {
				state.Liked = true
			}
		}
	}
	return state, nil
}

type likeKey struct {
	userID     uuid.UUID
	targetType TargetType
	targetID   uuid.UUID
}
