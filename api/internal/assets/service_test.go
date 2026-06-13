package assets

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestValidateSVGRejectsMaliciousContent(t *testing.T) {
	fixtures := []string{
		`<svg><script>alert(1)</script></svg>`,
		`<svg onload="alert(1)"></svg>`,
		`<svg><a href="javascript:alert(1)">x</a></svg>`,
		`<svg><image href="https://example.com/a.png" /></svg>`,
		`<svg><foreignObject><p>x</p></foreignObject></svg>`,
	}
	for _, fixture := range fixtures {
		if err := ValidateSVG([]byte(fixture)); !errors.Is(err, ErrUnsafeSVG) {
			t.Fatalf("ValidateSVG(%q) = %v, want ErrUnsafeSVG", fixture, err)
		}
	}
}

func TestServiceUploadValidatesAndStoresProviderNeutralKey(t *testing.T) {
	fileID := uuid.New()
	repo := &fakeRepository{files: map[uuid.UUID]bool{fileID: true}}
	storage := &fakeStorage{objects: map[string][]byte{}}
	service := NewService(repo, storage)

	asset, err := service.Upload(context.Background(), fileID, Upload{
		Filename: "../hello world.svg",
		MIMEType: "image/svg+xml",
		Reader:   strings.NewReader(`<svg xmlns="http://www.w3.org/2000/svg"><circle cx="1" cy="1" r="1" /></svg>`),
	})
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if asset.Filename != "hello-world.svg" {
		t.Fatalf("filename = %q, want sanitized", asset.Filename)
	}
	if !strings.HasPrefix(asset.StorageKey, "files/"+fileID.String()+"/") || strings.HasPrefix(asset.StorageKey, "/") {
		t.Fatalf("storage key = %q, want provider-neutral relative key", asset.StorageKey)
	}
	if asset.PublicURL != "" {
		t.Fatalf("draft public_url = %q, want empty until publish", asset.PublicURL)
	}
	if _, ok := storage.objects[asset.StorageKey]; !ok {
		t.Fatalf("stored object missing at key %q", asset.StorageKey)
	}
}

func TestServiceUploadRejectsMissingFileAndPerFileLimit(t *testing.T) {
	fileID := uuid.New()
	service := NewService(&fakeRepository{}, &fakeStorage{objects: map[string][]byte{}})
	_, err := service.Upload(context.Background(), fileID, Upload{Filename: "a.txt", MIMEType: "text/plain", Reader: strings.NewReader("hello")})
	if !errors.Is(err, ErrFileNotFound) {
		t.Fatalf("missing file error = %v, want ErrFileNotFound", err)
	}

	service = NewService(&fakeRepository{files: map[uuid.UUID]bool{fileID: true}, total: MaxPerFileTotalSize}, &fakeStorage{objects: map[string][]byte{}})
	_, err = service.Upload(context.Background(), fileID, Upload{Filename: "a.txt", MIMEType: "text/plain", Reader: strings.NewReader("hello")})
	if !errors.Is(err, ErrPerFileLimit) {
		t.Fatalf("per-file limit error = %v, want ErrPerFileLimit", err)
	}
}

func TestLocalStorageRejectsUnsafeKeys(t *testing.T) {
	storage := NewLocalStorage(t.TempDir())
	for _, key := range []string{"/absolute", "../escape", "files/../../escape", `files\\escape`} {
		if err := storage.Put(key, strings.NewReader("x")); !errors.Is(err, ErrStorageKeyUnsafe) {
			t.Fatalf("Put(%q) error = %v, want ErrStorageKeyUnsafe", key, err)
		}
	}
}

type fakeRepository struct {
	files   map[uuid.UUID]bool
	total   int64
	assets  map[uuid.UUID]FileAsset
	created FileAsset
}

func (f *fakeRepository) FileTargetExists(_ context.Context, fileID uuid.UUID) (bool, error) {
	return f.files[fileID], nil
}

func (f *fakeRepository) FileAssetTotalBytes(context.Context, uuid.UUID) (int64, error) {
	return f.total, nil
}

func (f *fakeRepository) CreateAsset(_ context.Context, asset FileAsset) (FileAsset, error) {
	f.created = asset
	if f.assets == nil {
		f.assets = map[uuid.UUID]FileAsset{}
	}
	f.assets[asset.ID] = asset
	return asset, nil
}

func (f *fakeRepository) FindPublishedAsset(_ context.Context, assetID uuid.UUID, filename string) (FileAsset, error) {
	asset, ok := f.assets[assetID]
	if !ok || asset.Filename != filename {
		return FileAsset{}, ErrAssetNotFound
	}
	return asset, nil
}

func (f *fakeRepository) FindDraftAsset(_ context.Context, assetID uuid.UUID, filename string) (FileAsset, error) {
	return f.FindPublishedAsset(context.Background(), assetID, filename)
}

func (f *fakeRepository) ListAssetState(_ context.Context, fileID uuid.UUID) ([]FileAsset, []FileAsset, error) {
	assets := []FileAsset{}
	for _, asset := range f.assets {
		if asset.FileID == fileID {
			assets = append(assets, asset)
		}
	}
	return assets, assets, nil
}

func (f *fakeRepository) PromoteDraftAssets(_ context.Context, fileID uuid.UUID) ([]FileAsset, error) {
	_, published, err := f.ListAssetState(context.Background(), fileID)
	return published, err
}

func (f *fakeRepository) DeleteAsset(_ context.Context, assetID uuid.UUID) (FileAsset, error) {
	asset, ok := f.assets[assetID]
	if !ok {
		return FileAsset{}, ErrAssetNotFound
	}
	delete(f.assets, assetID)
	return asset, nil
}

type fakeStorage struct {
	objects map[string][]byte
}

func (f *fakeStorage) Put(key string, reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	f.objects[key] = data
	return nil
}

func (f *fakeStorage) Open(key string) (StoredObject, error) {
	data, ok := f.objects[key]
	if !ok {
		return StoredObject{}, ErrAssetNotFound
	}
	return StoredObject{Reader: io.NopCloser(bytes.NewReader(data)), Size: int64(len(data)), ContentType: "text/plain"}, nil
}

func (f *fakeStorage) Delete(key string) error {
	delete(f.objects, key)
	return nil
}
