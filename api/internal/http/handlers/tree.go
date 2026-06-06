package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"xlab-blog/api/internal/http/respond"
	"xlab-blog/api/internal/tree"
)

type TreeHandler struct {
	service *tree.Service
}

type RecentTreeService interface {
	Recent(context.Context, int, int) (tree.FileEntryList, error)
}

func NewTreeHandler(service *tree.Service) *TreeHandler {
	return &TreeHandler{service: service}
}

func (h *TreeHandler) Root(w http.ResponseWriter, r *http.Request) {
	page, err := h.service.Root(r.Context())
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, page)
}

func (h *TreeHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Resolve(r.Context(), r.URL.Query().Get("path"))
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, result)
}

func (h *TreeHandler) Children(w http.ResponseWriter, r *http.Request) {
	nodeID, err := uuid.Parse(chi.URLParam(r, "node_id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid node_id")
		return
	}
	page, err := h.service.Children(r.Context(), nodeID)
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, page)
}

func RecentFiles(service RecentTreeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, err := optionalNonNegativeInt(r, "limit")
		if err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid limit")
			return
		}
		offset, err := optionalNonNegativeInt(r, "offset")
		if err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid offset")
			return
		}
		items, err := service.Recent(r.Context(), limit, offset)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "recent files request failed")
			return
		}
		respond.JSON(w, http.StatusOK, items)
	}
}

func optionalNonNegativeInt(r *http.Request, key string) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, errors.New("invalid pagination")
	}
	return value, nil
}

func (h *TreeHandler) respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, tree.ErrInvalidPath):
		respond.Error(w, http.StatusBadRequest, "invalid path")
	case errors.Is(err, tree.ErrNotFound):
		respond.Error(w, http.StatusNotFound, "tree node not found")
	default:
		respond.Error(w, http.StatusInternalServerError, "tree request failed")
	}
}
