package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"xlab-blog/api/internal/assets"
	"xlab-blog/api/internal/http/respond"
)

const maxMultipartMemory = 8 << 20

type AssetService interface {
	Upload(context.Context, uuid.UUID, assets.Upload) (assets.FileAsset, error)
	OpenPublished(context.Context, uuid.UUID, string) (assets.FileAsset, assets.StoredObject, error)
	Delete(context.Context, uuid.UUID) error
}

type AssetHandler struct {
	service AssetService
}

func NewAssetHandler(service AssetService) *AssetHandler {
	return &AssetHandler{service: service}
}

func (h *AssetHandler) Upload(w http.ResponseWriter, r *http.Request) {
	fileID, ok := parseAssetID(w, r, "file_id")
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(maxMultipartMemory); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	asset, err := h.service.Upload(r.Context(), fileID, assets.Upload{
		Filename: header.Filename,
		MIMEType: header.Header.Get("Content-Type"),
		Size:     header.Size,
		Reader:   file,
	})
	if err != nil {
		h.respondError(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, asset)
}

func (h *AssetHandler) ServePublished(w http.ResponseWriter, r *http.Request) {
	assetID, ok := parseAssetID(w, r, "asset_id")
	if !ok {
		return
	}
	filename := strings.TrimSpace(chi.URLParam(r, "filename"))
	asset, object, err := h.service.OpenPublished(r.Context(), assetID, filename)
	if err != nil {
		h.respondError(w, err)
		return
	}
	defer object.Reader.Close()

	contentType := object.ContentType
	if contentType == "" {
		contentType = asset.MIMEType
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	if asset.MIMEType == "application/pdf" {
		w.Header().Set("Content-Disposition", "inline")
	}
	w.Header().Set("Content-Length", int64String(object.Size))
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, object.Reader)
}

func (h *AssetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	assetID, ok := parseAssetID(w, r, "asset_id")
	if !ok {
		return
	}
	if err := h.service.Delete(r.Context(), assetID); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseAssetID(w http.ResponseWriter, r *http.Request, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, param))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid "+param)
		return uuid.Nil, false
	}
	return id, true
}

func (h *AssetHandler) respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, assets.ErrFileNotFound), errors.Is(err, assets.ErrAssetNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, assets.ErrInvalidFilename), errors.Is(err, assets.ErrMIMETypeNotAllowed), errors.Is(err, assets.ErrAssetTooLarge), errors.Is(err, assets.ErrPerFileLimit), errors.Is(err, assets.ErrUnsafeSVG), errors.Is(err, assets.ErrDuplicateAssetName):
		respond.Error(w, http.StatusBadRequest, err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "asset request failed")
	}
}

func int64String(value int64) string {
	return strconvFormatInt(value)
}

// isolated for tests and to keep handler logic explicit without leaking local paths.
var strconvFormatInt = func(value int64) string {
	return strconv.FormatInt(value, 10)
}
