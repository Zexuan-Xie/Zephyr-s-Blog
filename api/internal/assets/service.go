package assets

import (
	"bytes"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Repository interface {
	FileTargetExists(ctx context.Context, fileID uuid.UUID) (bool, error)
	FileAssetTotalBytes(ctx context.Context, fileID uuid.UUID) (int64, error)
	CreateAsset(ctx context.Context, asset FileAsset) (FileAsset, error)
	FindPublishedAsset(ctx context.Context, assetID uuid.UUID, filename string) (FileAsset, error)
	FindDraftAsset(ctx context.Context, assetID uuid.UUID, filename string) (FileAsset, error)
	ListAssetState(ctx context.Context, fileID uuid.UUID) ([]FileAsset, []FileAsset, error)
	PromoteDraftAssets(ctx context.Context, fileID uuid.UUID) ([]FileAsset, error)
	DeleteAsset(ctx context.Context, assetID uuid.UUID) (FileAsset, error)
}

type Service struct {
	repo    Repository
	storage Storage
}

func NewService(repo Repository, storage Storage) *Service {
	return &Service{repo: repo, storage: storage}
}

func (s *Service) Upload(ctx context.Context, fileID uuid.UUID, upload Upload) (FileAsset, error) {
	filename, err := SanitizeFilename(upload.Filename)
	if err != nil {
		return FileAsset{}, err
	}
	data, mimeType, err := ReadUpload(Upload{Filename: filename, MIMEType: upload.MIMEType, Size: upload.Size, Reader: upload.Reader})
	if err != nil {
		return FileAsset{}, err
	}
	exists, err := s.repo.FileTargetExists(ctx, fileID)
	if err != nil {
		return FileAsset{}, err
	}
	if !exists {
		return FileAsset{}, ErrFileNotFound
	}
	total, err := s.repo.FileAssetTotalBytes(ctx, fileID)
	if err != nil {
		return FileAsset{}, err
	}
	if total+int64(len(data)) > MaxPerFileTotalSize {
		return FileAsset{}, ErrPerFileLimit
	}

	assetID := uuid.New()
	asset := FileAsset{
		ID:              assetID,
		FileID:          fileID,
		Filename:        filename,
		MIMEType:        mimeType,
		SizeBytes:       int64(len(data)),
		StorageProvider: StorageProviderLocal,
		StorageKey:      fmt.Sprintf("files/%s/%s-%s", fileID, assetID, filename),
		State:           "draft",
	}
	if err := s.storage.Put(asset.StorageKey, bytes.NewReader(data)); err != nil {
		return FileAsset{}, err
	}
	created, err := s.repo.CreateAsset(ctx, asset)
	if err != nil {
		_ = s.storage.Delete(asset.StorageKey)
		return FileAsset{}, err
	}
	return created, nil
}

func (s *Service) OpenPublished(ctx context.Context, assetID uuid.UUID, filename string) (FileAsset, StoredObject, error) {
	asset, err := s.repo.FindPublishedAsset(ctx, assetID, filename)
	if err != nil {
		return FileAsset{}, StoredObject{}, err
	}
	object, err := s.storage.Open(asset.StorageKey)
	if err != nil {
		return FileAsset{}, StoredObject{}, err
	}
	if object.ContentType == "" {
		object.ContentType = asset.MIMEType
	}
	return asset, object, nil
}

func (s *Service) Delete(ctx context.Context, assetID uuid.UUID) error {
	asset, err := s.repo.DeleteAsset(ctx, assetID)
	if err != nil {
		return err
	}
	return s.storage.Delete(asset.StorageKey)
}
