package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"xlab-blog/api/internal/http/respond"
	"xlab-blog/api/internal/tree"
)

type TreeLifecycleService interface {
	UpsertFileContent(context.Context, uuid.UUID, tree.UpsertFileContentInput) (tree.FileContent, error)
	PublishFile(context.Context, uuid.UUID) (tree.FileContent, error)
	UnpublishFile(context.Context, uuid.UUID) (tree.FileContent, error)
	DeleteNode(context.Context, uuid.UUID) error
}

type TreeLifecycleHandler struct {
	service TreeLifecycleService
}

func NewTreeLifecycleHandler(service TreeLifecycleService) *TreeLifecycleHandler {
	return &TreeLifecycleHandler{service: service}
}

func (h *TreeLifecycleHandler) UpsertFileContent(w http.ResponseWriter, r *http.Request) {
	fileID, ok := parseTreeID(w, r, "file_id")
	if !ok {
		return
	}
	var input tree.UpsertFileContentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	content, err := h.service.UpsertFileContent(r.Context(), fileID, input)
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, content)
}

func (h *TreeLifecycleHandler) PublishFile(w http.ResponseWriter, r *http.Request) {
	fileID, ok := parseTreeID(w, r, "file_id")
	if !ok {
		return
	}
	content, err := h.service.PublishFile(r.Context(), fileID)
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, content)
}

func (h *TreeLifecycleHandler) UnpublishFile(w http.ResponseWriter, r *http.Request) {
	fileID, ok := parseTreeID(w, r, "file_id")
	if !ok {
		return
	}
	content, err := h.service.UnpublishFile(r.Context(), fileID)
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, content)
}

func (h *TreeLifecycleHandler) DeleteNode(w http.ResponseWriter, r *http.Request) {
	nodeID, ok := parseTreeID(w, r, "node_id")
	if !ok {
		return
	}
	if err := h.service.DeleteNode(r.Context(), nodeID); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseTreeID(w http.ResponseWriter, r *http.Request, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, param))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid "+param)
		return uuid.Nil, false
	}
	return id, true
}

func (h *TreeLifecycleHandler) respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, tree.ErrNodeNotFound), errors.Is(err, tree.ErrFileContentNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, tree.ErrNodeIsNotFile), errors.Is(err, tree.ErrInvalidContentFormat):
		respond.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, tree.ErrNonEmptyDirectoryDelete):
		respond.JSON(w, http.StatusConflict, respond.ErrorResponse{Error: err.Error(), Details: map[string]any{"reason": "non_empty_directory"}})
	case errors.Is(err, tree.ErrPublishedContentFormatChange),
		errors.Is(err, tree.ErrPublishedFileDelete),
		errors.Is(err, tree.ErrDirectoryHasPublishedDescendants):
		respond.Error(w, http.StatusConflict, err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "tree lifecycle request failed")
	}
}
