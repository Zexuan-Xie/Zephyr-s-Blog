package handlers

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"xlab-blog/api/internal/assets"
)

func TestAssetHandlerUploadServeAndDelete(t *testing.T) {
	fileID := uuid.New()
	assetID := uuid.New()
	service := &fakeAssetService{asset: assets.FileAsset{ID: assetID, FileID: fileID, Filename: "demo.txt", MIMEType: "text/plain", SizeBytes: 4, PublicURL: "/api/assets/" + assetID.String() + "/demo.txt"}, object: assets.StoredObject{Reader: io.NopCloser(strings.NewReader("demo")), Size: 4, ContentType: "text/plain"}}
	handler := NewAssetHandler(service)
	router := chi.NewRouter()
	router.Post("/admin/files/{file_id}/assets", handler.Upload)
	router.Get("/assets/{asset_id}/{filename}", handler.ServePublished)
	router.Delete("/admin/assets/{asset_id}", handler.Delete)

	body, contentType := multipartBody(t, "file", "demo.txt", "text/plain", "demo")
	request := httptest.NewRequest(http.MethodPost, "/admin/files/"+fileID.String()+"/assets", body)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("upload status = %d, want %d; body=%s", response.Code, http.StatusCreated, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"public_url"`) {
		t.Fatalf("upload body = %s, want public_url", response.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "/assets/"+assetID.String()+"/demo.txt", nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("serve status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("Cache-Control = %q", got)
	}
	if response.Body.String() != "demo" {
		t.Fatalf("body = %q, want demo", response.Body.String())
	}

	request = httptest.NewRequest(http.MethodDelete, "/admin/assets/"+assetID.String(), nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestAssetHandlerMapsUnsafeSVG(t *testing.T) {
	fileID := uuid.New()
	handler := NewAssetHandler(&fakeAssetService{err: assets.ErrUnsafeSVG})
	router := chi.NewRouter()
	router.Post("/admin/files/{file_id}/assets", handler.Upload)
	body, contentType := multipartBody(t, "file", "bad.svg", "image/svg+xml", `<svg><script /></svg>`)
	request := httptest.NewRequest(http.MethodPost, "/admin/files/"+fileID.String()+"/assets", body)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusBadRequest, response.Body.String())
	}
}

func multipartBody(t *testing.T, field, filename, contentType, value string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="`+field+`"; filename="`+filename+`"`)
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte(value)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return body, writer.FormDataContentType()
}

type fakeAssetService struct {
	asset  assets.FileAsset
	object assets.StoredObject
	err    error
}

func (f *fakeAssetService) Upload(context.Context, uuid.UUID, assets.Upload) (assets.FileAsset, error) {
	return f.asset, f.err
}

func (f *fakeAssetService) OpenPublished(context.Context, uuid.UUID, string) (assets.FileAsset, assets.StoredObject, error) {
	return f.asset, f.object, f.err
}

func (f *fakeAssetService) Delete(context.Context, uuid.UUID) error {
	return f.err
}
