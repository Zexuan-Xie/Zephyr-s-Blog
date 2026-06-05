package assets

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	root string
}

func NewLocalStorage(root string) *LocalStorage {
	return &LocalStorage{root: root}
}

func (s *LocalStorage) Put(key string, reader io.Reader) error {
	path, err := s.pathForKey(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(file, reader)
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return closeErr
	}
	return os.Rename(tmp, path)
}

func (s *LocalStorage) Open(key string) (StoredObject, error) {
	path, err := s.pathForKey(key)
	if err != nil {
		return StoredObject{}, err
	}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return StoredObject{}, ErrAssetNotFound
		}
		return StoredObject{}, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return StoredObject{}, err
	}
	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(path)))
	return StoredObject{Reader: file, Size: info.Size(), ContentType: contentType}, nil
}

func (s *LocalStorage) Delete(key string) error {
	path, err := s.pathForKey(key)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *LocalStorage) pathForKey(key string) (string, error) {
	if s.root == "" || strings.HasPrefix(key, "/") || strings.Contains(key, "..") || strings.Contains(key, `\\`) {
		return "", ErrStorageKeyUnsafe
	}
	clean := filepath.Clean(filepath.FromSlash(key))
	if clean == "." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
		return "", ErrStorageKeyUnsafe
	}
	path := filepath.Join(s.root, clean)
	rootAbs, err := filepath.Abs(s.root)
	if err != nil {
		return "", err
	}
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if pathAbs != rootAbs && !strings.HasPrefix(pathAbs, rootAbs+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: %s", ErrStorageKeyUnsafe, key)
	}
	return pathAbs, nil
}
