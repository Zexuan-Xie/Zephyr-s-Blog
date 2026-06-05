package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"xlab-blog/api/internal/comments"
	"xlab-blog/api/internal/http/middleware"
	"xlab-blog/api/internal/http/respond"
	"xlab-blog/api/internal/users"
)

type CommentService interface {
	Thread(context.Context, uuid.UUID, *uuid.UUID) (comments.Thread, error)
	Create(context.Context, uuid.UUID, uuid.UUID, comments.CreateInput) (comments.Comment, error)
	Delete(context.Context, uuid.UUID, users.User) (comments.Comment, error)
}

type CommentHandler struct {
	service CommentService
}

func NewCommentHandler(service CommentService) *CommentHandler {
	return &CommentHandler{service: service}
}

func (h *CommentHandler) Thread(w http.ResponseWriter, r *http.Request) {
	fileID, ok := parseCommentID(w, r, "file_id")
	if !ok {
		return
	}
	var viewerID *uuid.UUID
	if user, ok := middleware.CurrentUser(r.Context()); ok {
		viewerID = &user.ID
	}
	thread, err := h.service.Thread(r.Context(), fileID, viewerID)
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, thread)
}

func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	fileID, ok := parseCommentID(w, r, "file_id")
	if !ok {
		return
	}
	user, ok := middleware.CurrentUser(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var input comments.CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	comment, err := h.service.Create(r.Context(), fileID, user.ID, input)
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, comment)
}

func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	commentID, ok := parseCommentID(w, r, "comment_id")
	if !ok {
		return
	}
	user, ok := middleware.CurrentUser(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if _, err := h.service.Delete(r.Context(), commentID, user); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseCommentID(w http.ResponseWriter, r *http.Request, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, param))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid "+param)
		return uuid.Nil, false
	}
	return id, true
}

func (h *CommentHandler) respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, comments.ErrFileNotFound), errors.Is(err, comments.ErrCommentNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, comments.ErrInvalidCommentBody), errors.Is(err, comments.ErrParentMismatch):
		respond.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, comments.ErrPermissionDenied):
		respond.Error(w, http.StatusForbidden, err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "comment request failed")
	}
}
