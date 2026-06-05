package likes

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrTargetNotFound = errors.New("like target not found")
	ErrTargetDeleted  = errors.New("like target is deleted")
)

type TargetType string

const (
	TargetFile    TargetType = "file"
	TargetComment TargetType = "comment"
)

type Target struct {
	Type TargetType
	ID   uuid.UUID
}

type State struct {
	Liked     bool `json:"liked"`
	LikeCount int  `json:"like_count"`
}
