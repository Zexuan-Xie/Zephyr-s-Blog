package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"xlab-blog/api/internal/http/respond"
	"xlab-blog/api/internal/render"
	"xlab-blog/api/internal/tree"
)

type TreeLifecycleService interface {
	GetFileVersionState(context.Context, uuid.UUID) (tree.FileVersionState, error)
	UpsertFileContent(context.Context, uuid.UUID, tree.UpsertFileContentInput) (tree.FileContent, error)
	RestorePreviousContent(context.Context, uuid.UUID, int) (tree.FileVersionState, error)
	PublishCurrentSnapshot(context.Context, uuid.UUID, int) (tree.PublishResult, error)
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

func (h *TreeLifecycleHandler) GetFileContent(w http.ResponseWriter, r *http.Request) {
	fileID, ok := parseTreeID(w, r, "file_id")
	if !ok {
		return
	}
	state, err := h.service.GetFileVersionState(r.Context(), fileID)
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, state)
}

func (h *TreeLifecycleHandler) RestorePreviousContent(w http.ResponseWriter, r *http.Request) {
	fileID, ok := parseTreeID(w, r, "file_id")
	if !ok {
		return
	}
	var input struct {
		ExpectedRevision int `json:"expected_revision"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	state, err := h.service.RestorePreviousContent(r.Context(), fileID, input.ExpectedRevision)
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, state)
}

func (h *TreeLifecycleHandler) PublishSummary(w http.ResponseWriter, r *http.Request) {
	fileID, ok := parseTreeID(w, r, "file_id")
	if !ok {
		return
	}
	state, err := h.service.GetFileVersionState(r.Context(), fileID)
	if err != nil {
		h.respondError(w, err)
		return
	}
	var publishedRevision *int
	willUpdate := true
	if state.Published != nil {
		publishedRevision = &state.Published.SourceRevision
		willUpdate = state.Published.SourceRevision != state.Current.Revision || !state.Published.Visible
	}
	respond.JSON(w, http.StatusOK, map[string]any{
		"file_id":                   fileID,
		"current_revision":          state.Current.Revision,
		"published_source_revision": publishedRevision,
		"will_update_content":       willUpdate,
		"draft_assets":              state.DraftAssets,
		"published_assets":          state.PublishedAssets,
		"asset_changes":             []any{},
	})
}

func (h *TreeLifecycleHandler) DraftPreview(w http.ResponseWriter, r *http.Request) {
	fileID, ok := parseTreeID(w, r, "file_id")
	if !ok {
		return
	}
	state, err := h.service.GetFileVersionState(r.Context(), fileID)
	if err != nil {
		h.respondError(w, err)
		return
	}
	html := state.Current.BodyRaw
	if state.Current.ContentFormat == tree.ContentFormatMarkdown {
		if state.Current.BodyHTML != nil && strings.TrimSpace(*state.Current.BodyHTML) != "" {
			html = *state.Current.BodyHTML
		} else if rendered, _, err := render.MarkdownToSafeHTML(state.Current.BodyRaw); err == nil {
			html = rendered
		}
	}
	respond.JSON(w, http.StatusOK, map[string]any{
		"current":        state.Current,
		"html":           html,
		"assets":         state.DraftAssets,
		"iframe_sandbox": "allow-scripts",
	})
}

func (h *TreeLifecycleHandler) FileAssetState(w http.ResponseWriter, r *http.Request) {
	fileID, ok := parseTreeID(w, r, "file_id")
	if !ok {
		return
	}
	state, err := h.service.GetFileVersionState(r.Context(), fileID)
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"draft_assets": state.DraftAssets, "published_assets": state.PublishedAssets})
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
	var input struct {
		ExpectedRevision int `json:"expected_revision"`
	}
	_ = json.NewDecoder(r.Body).Decode(&input)
	if input.ExpectedRevision > 0 {
		result, err := h.service.PublishCurrentSnapshot(r.Context(), fileID, input.ExpectedRevision)
		if err != nil {
			h.respondError(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, result)
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
	case errors.Is(err, tree.ErrLostUpdate):
		respond.JSON(w, http.StatusConflict, respond.ErrorResponse{Error: "revision conflict", Details: map[string]any{"reason": "revision_conflict"}})
	case errors.Is(err, tree.ErrPublishedContentFormatChange),
		errors.Is(err, tree.ErrPublishedFileDelete),
		errors.Is(err, tree.ErrDirectoryHasPublishedDescendants):
		respond.Error(w, http.StatusConflict, err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "tree lifecycle request failed")
	}
}
