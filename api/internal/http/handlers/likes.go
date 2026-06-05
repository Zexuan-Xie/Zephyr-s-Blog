package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"xlab-blog/api/internal/http/middleware"
	"xlab-blog/api/internal/http/respond"
	"xlab-blog/api/internal/likes"
)

type LikeService interface {
	LikeFile(context.Context, uuid.UUID, uuid.UUID) (likes.State, error)
	UnlikeFile(context.Context, uuid.UUID, uuid.UUID) (likes.State, error)
	LikeComment(context.Context, uuid.UUID, uuid.UUID) (likes.State, error)
	UnlikeComment(context.Context, uuid.UUID, uuid.UUID) (likes.State, error)
}

type LikeHandler struct {
	service LikeService
}

func NewLikeHandler(service LikeService) *LikeHandler {
	return &LikeHandler{service: service}
}

func (h *LikeHandler) LikeFile(w http.ResponseWriter, r *http.Request) {
	h.handleFile(w, r, true)
}

func (h *LikeHandler) UnlikeFile(w http.ResponseWriter, r *http.Request) {
	h.handleFile(w, r, false)
}

func (h *LikeHandler) LikeComment(w http.ResponseWriter, r *http.Request) {
	h.handleComment(w, r, true)
}

func (h *LikeHandler) UnlikeComment(w http.ResponseWriter, r *http.Request) {
	h.handleComment(w, r, false)
}

func (h *LikeHandler) handleFile(w http.ResponseWriter, r *http.Request, liked bool) {
	fileID, ok := parseLikeID(w, r, "file_id")
	if !ok {
		return
	}
	user, ok := middleware.CurrentUser(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var state likes.State
	var err error
	if liked {
		state, err = h.service.LikeFile(r.Context(), user.ID, fileID)
	} else {
		state, err = h.service.UnlikeFile(r.Context(), user.ID, fileID)
	}
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, state)
}

func (h *LikeHandler) handleComment(w http.ResponseWriter, r *http.Request, liked bool) {
	commentID, ok := parseLikeID(w, r, "comment_id")
	if !ok {
		return
	}
	user, ok := middleware.CurrentUser(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var state likes.State
	var err error
	if liked {
		state, err = h.service.LikeComment(r.Context(), user.ID, commentID)
	} else {
		state, err = h.service.UnlikeComment(r.Context(), user.ID, commentID)
	}
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, state)
}

func parseLikeID(w http.ResponseWriter, r *http.Request, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, param))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid "+param)
		return uuid.Nil, false
	}
	return id, true
}

func (h *LikeHandler) respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, likes.ErrTargetNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, likes.ErrTargetDeleted):
		respond.Error(w, http.StatusConflict, err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "like request failed")
	}
}
