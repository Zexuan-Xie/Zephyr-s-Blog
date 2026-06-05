package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"xlab-blog/api/internal/http/respond"
	"xlab-blog/api/internal/search"
)

type SearchService interface {
	Search(context.Context, string, search.Options) (search.Response, error)
	RefreshFileEmbedding(context.Context, uuid.UUID) (search.EmbeddingState, error)
	Rebuild(context.Context) (search.RebuildState, error)
}

type SearchHandler struct {
	service SearchService
}

func NewSearchHandler(service SearchService) *SearchHandler {
	return &SearchHandler{service: service}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	options, err := parseSearchOptions(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := h.service.Search(r.Context(), r.URL.Query().Get("q"), options)
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, response)
}

func (h *SearchHandler) RefreshEmbedding(w http.ResponseWriter, r *http.Request) {
	fileID, ok := parseTreeID(w, r, "file_id")
	if !ok {
		return
	}
	state, err := h.service.RefreshFileEmbedding(r.Context(), fileID)
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusAccepted, state)
}

func (h *SearchHandler) Rebuild(w http.ResponseWriter, r *http.Request) {
	state, err := h.service.Rebuild(r.Context())
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusAccepted, state)
}

func parseSearchOptions(r *http.Request) (search.Options, error) {
	var options search.Options
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit < 0 {
			return options, errors.New("invalid limit")
		}
		options.Limit = limit
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		offset, err := strconv.Atoi(raw)
		if err != nil || offset < 0 {
			return options, errors.New("invalid offset")
		}
		options.Offset = offset
	}
	return options, nil
}

func (h *SearchHandler) respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, search.ErrInvalidQuery):
		respond.Error(w, http.StatusBadRequest, "invalid search query")
	case errors.Is(err, search.ErrEmbeddingNotFound):
		respond.Error(w, http.StatusNotFound, "file embedding target not found")
	default:
		respond.Error(w, http.StatusInternalServerError, "search request failed")
	}
}
