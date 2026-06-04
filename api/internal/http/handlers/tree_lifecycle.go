package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"xlab-blog/api/internal/http/respond"
	"xlab-blog/api/internal/tree"
)

type TreeLifecycleService interface {
	UpsertFileContent(rctx context.Context, nodeID uuid.UUID, input tree.UpsertFileContentInput) (tree.FileContent, error)
	PublishFile(rctx context.Context, nodeID uuid.UUID) (tree.FileContent, error)
	UnpublishFile(rctx context.Context, nodeID uuid.UUID) (tree.FileContent, error)
	DeleteNode(rctx context.Context, nodeID uuid.UUID) error
}
