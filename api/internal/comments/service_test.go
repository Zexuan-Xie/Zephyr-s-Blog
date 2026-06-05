package comments

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"xlab-blog/api/internal/users"
)

func TestServiceCreateNormalizesReplyToReply(t *testing.T) {
	fileID := uuid.New()
	topLevelID := uuid.New()
	replyID := uuid.New()
	replyAuthorID := uuid.New()
	actorID := uuid.New()
	repo := &fakeRepository{
		published: true,
		comments: map[uuid.UUID]Comment{
			replyID: {
				ID:         replyID,
				FileNodeID: fileID,
				ParentID:   &topLevelID,
				User:       PublicUser{ID: replyAuthorID, DisplayName: "Reply Author"},
			},
		},
		inserted: Comment{ID: uuid.New(), FileNodeID: fileID, User: PublicUser{ID: actorID, DisplayName: "Actor"}},
	}

	comment, err := NewService(repo).Create(context.Background(), fileID, actorID, CreateInput{Body: "  hello reply  ", ParentID: &replyID})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if comment.ID == uuid.Nil {
		t.Fatal("Create() returned empty comment")
	}
	if repo.lastInput.Body != "hello reply" {
		t.Fatalf("body = %q, want trimmed", repo.lastInput.Body)
	}
	if repo.lastInput.ParentID == nil || *repo.lastInput.ParentID != topLevelID {
		t.Fatalf("parent id = %v, want top-level %s", repo.lastInput.ParentID, topLevelID)
	}
	if repo.lastInput.ReplyToUserID == nil || *repo.lastInput.ReplyToUserID != replyAuthorID {
		t.Fatalf("reply_to_user_id = %v, want reply author %s", repo.lastInput.ReplyToUserID, replyAuthorID)
	}
}

func TestServiceCreateRejectsInvalidBodyAndMissingFile(t *testing.T) {
	fileID := uuid.New()
	actorID := uuid.New()
	service := NewService(&fakeRepository{published: true})
	if _, err := service.Create(context.Background(), fileID, actorID, CreateInput{Body: " \n\t "}); !errors.Is(err, ErrInvalidCommentBody) {
		t.Fatalf("empty body error = %v, want ErrInvalidCommentBody", err)
	}

	service = NewService(&fakeRepository{published: false})
	if _, err := service.Create(context.Background(), fileID, actorID, CreateInput{Body: "hello"}); !errors.Is(err, ErrFileNotFound) {
		t.Fatalf("missing file error = %v, want ErrFileNotFound", err)
	}
}

func TestServiceCreateRejectsParentFromDifferentFile(t *testing.T) {
	fileID := uuid.New()
	otherFileID := uuid.New()
	parentID := uuid.New()
	repo := &fakeRepository{
		published: true,
		comments: map[uuid.UUID]Comment{
			parentID: {ID: parentID, FileNodeID: otherFileID, User: PublicUser{ID: uuid.New(), DisplayName: "Other"}},
		},
	}
	if _, err := NewService(repo).Create(context.Background(), fileID, uuid.New(), CreateInput{Body: "hello", ParentID: &parentID}); !errors.Is(err, ErrParentMismatch) {
		t.Fatalf("Create() error = %v, want ErrParentMismatch", err)
	}
}

func TestServiceDeleteAllowsOwnerAndAdminOnly(t *testing.T) {
	ownerID := uuid.New()
	commentID := uuid.New()
	repo := &fakeRepository{comments: map[uuid.UUID]Comment{
		commentID: {ID: commentID, User: PublicUser{ID: ownerID, DisplayName: "Owner"}},
	}}
	service := NewService(repo)

	if _, err := service.Delete(context.Background(), commentID, users.User{ID: uuid.New(), Role: users.RoleReader}); !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("non-owner error = %v, want ErrPermissionDenied", err)
	}
	if _, err := service.Delete(context.Background(), commentID, users.User{ID: ownerID, Role: users.RoleReader}); err != nil {
		t.Fatalf("owner Delete() error = %v", err)
	}
	if repo.deletedBy != ownerID {
		t.Fatalf("deletedBy = %s, want owner %s", repo.deletedBy, ownerID)
	}
	adminID := uuid.New()
	if _, err := service.Delete(context.Background(), commentID, users.User{ID: adminID, Role: users.RoleAdmin}); err != nil {
		t.Fatalf("admin Delete() error = %v", err)
	}
}

type fakeRepository struct {
	published bool
	comments  map[uuid.UUID]Comment
	thread    Thread
	inserted  Comment
	lastInput CreateInput
	deletedBy uuid.UUID
}

func (f *fakeRepository) PublishedFileExists(context.Context, uuid.UUID) (bool, error) {
	return f.published, nil
}

func (f *fakeRepository) ListThread(context.Context, uuid.UUID, *uuid.UUID) (Thread, error) {
	return f.thread, nil
}

func (f *fakeRepository) FindComment(_ context.Context, commentID uuid.UUID) (Comment, error) {
	comment, ok := f.comments[commentID]
	if !ok {
		return Comment{}, ErrCommentNotFound
	}
	return comment, nil
}

func (f *fakeRepository) InsertComment(_ context.Context, fileID uuid.UUID, userID uuid.UUID, input CreateInput) (Comment, error) {
	f.lastInput = input
	if f.inserted.ID == uuid.Nil {
		f.inserted = Comment{ID: uuid.New(), FileNodeID: fileID, User: PublicUser{ID: userID, DisplayName: "Reader"}, Body: input.Body}
	}
	return f.inserted, nil
}

func (f *fakeRepository) SoftDeleteComment(_ context.Context, commentID uuid.UUID, deletedBy uuid.UUID) (Comment, error) {
	f.deletedBy = deletedBy
	comment, ok := f.comments[commentID]
	if !ok {
		return Comment{}, ErrCommentNotFound
	}
	comment.Deleted = true
	return comment, nil
}
