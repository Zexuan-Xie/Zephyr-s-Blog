package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"xlab-blog/api/internal/http/respond"
	"xlab-blog/api/internal/tree"
)

type TreeHandler struct {
	service *tree.Service
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
