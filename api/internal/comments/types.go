package comments

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const MaxBodyLength = 5000

var (
	ErrFileNotFound       = errors.New("file not found")
	ErrCommentNotFound    = errors.New("comment not found")
	ErrInvalidCommentBody = errors.New("invalid comment body")
	ErrParentMismatch     = errors.New("parent comment does not belong to file")
	ErrPermissionDenied   = errors.New("comment delete forbidden")
)

type PublicUser struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
}

type Comment struct {
	ID             uuid.UUID  `json:"id"`
	FileNodeID     uuid.UUID  `json:"file_node_id"`
	ParentID       *uuid.UUID `json:"parent_id"`
	ReplyToUserID  *uuid.UUID `json:"reply_to_user_id"`
	User           PublicUser `json:"user"`
	Body           string     `json:"body"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at"`
	Deleted        bool       `json:"deleted"`
	LikeCount      int        `json:"like_count"`
	ViewerHasLiked bool       `json:"viewer_has_liked"`
	Replies        []Comment  `json:"replies"`
}

type Thread struct {
	FileID   uuid.UUID `json:"file_id"`
	Comments []Comment `json:"comments"`
}

type CreateInput struct {
	Body          string     `json:"body"`
	ParentID      *uuid.UUID `json:"parent_id,omitempty"`
	ReplyToUserID *uuid.UUID `json:"reply_to_user_id,omitempty"`
}
