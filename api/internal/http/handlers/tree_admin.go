package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"xlab-blog/api/internal/http/respond"
	"xlab-blog/api/internal/tree"
)

type AdminNodeService interface {
	GetNode(context.Context, uuid.UUID) (tree.AdminNodeDetail, error)
	CreateNode(context.Context, tree.CreateNodeInput) (tree.AdminNodeDetail, error)
	UpdateNode(context.Context, uuid.UUID, tree.UpdateNodeInput) (tree.AdminNodeDetail, error)
}

type AdminNodeHandler struct {
	service AdminNodeService
}

func NewAdminNodeHandler(service AdminNodeService) *AdminNodeHandler {
	return &AdminNodeHandler{service: service}
}

func (h *AdminNodeHandler) GetNode(w http.ResponseWriter, r *http.Request) {
	nodeID, ok := parseTreeID(w, r, "node_id")
	if !ok {
		return
	}
	detail, err := h.service.GetNode(r.Context(), nodeID)
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, detail)
}

func (h *AdminNodeHandler) CreateNode(w http.ResponseWriter, r *http.Request) {
	var input tree.CreateNodeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	detail, err := h.service.CreateNode(r.Context(), input)
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, detail)
}

func (h *AdminNodeHandler) UpdateNode(w http.ResponseWriter, r *http.Request) {
	nodeID, ok := parseTreeID(w, r, "node_id")
	if !ok {
		return
	}
	var input tree.UpdateNodeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	detail, err := h.service.UpdateNode(r.Context(), nodeID, input)
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, detail)
}

func (h *AdminNodeHandler) respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, tree.ErrNodeNotFound), errors.Is(err, tree.ErrFileContentNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, tree.ErrDuplicateSlug):
		respond.Error(w, http.StatusConflict, "a node with this slug already exists under the selected parent")
	case errors.Is(err, tree.ErrReservedRootSlug):
		respond.Error(w, http.StatusConflict, err.Error())
	case errors.Is(err, tree.ErrNodeNameRequired):
		respond.Error(w, http.StatusBadRequest, "node name is required")
	case errors.Is(err, tree.ErrNodeSlugRequired):
		respond.Error(w, http.StatusBadRequest, "node slug is required")
	case errors.Is(err, tree.ErrInvalidNodeKind):
		respond.Error(w, http.StatusBadRequest, "node kind must be directory or file")
	case errors.Is(err, tree.ErrInvalidNodeSlug):
		respond.Error(w, http.StatusBadRequest, "node slug must not be '.', '..', or contain '/'")
	case errors.Is(err, tree.ErrInvalidNodeInput),
		errors.Is(err, tree.ErrInvalidContentFormat),
		errors.Is(err, tree.ErrParentNotDirectory),
		errors.Is(err, tree.ErrNodeCycle):
		respond.Error(w, http.StatusBadRequest, err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "admin tree request failed")
	}
}
