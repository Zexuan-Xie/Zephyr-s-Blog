package assets

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
)

var allowedMIMEs = map[string]int64{
	"image/png":              MaxImageSize,
	"image/jpeg":             MaxImageSize,
	"image/webp":             MaxImageSize,
	"image/gif":              MaxImageSize,
	"image/svg+xml":          MaxImageSize,
	"application/pdf":        MaxPDFSize,
	"text/css":               MaxTextSize,
	"text/javascript":        MaxTextSize,
	"application/javascript": MaxTextSize,
	"application/json":       MaxTextSize,
	"text/plain":             MaxTextSize,
	"text/csv":               MaxTextSize,
}

var safeFilenameRe = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func SanitizeFilename(filename string) (string, error) {
	filename = filepath.Base(strings.TrimSpace(filename))
	filename = strings.Trim(filename, ". ")
	filename = safeFilenameRe.ReplaceAllString(filename, "-")
	filename = strings.Trim(filename, "-")
	if filename == "" || filename == "." || filename == ".." || strings.Contains(filename, "/") || strings.Contains(filename, `\\`) {
		return "", ErrInvalidFilename
	}
	return filename, nil
}

func NormalizeMIME(filename string, provided string, sample []byte) (string, error) {
	mimeType := strings.ToLower(strings.TrimSpace(strings.Split(provided, ";")[0]))
	if mimeType == "" || mimeType == "application/octet-stream" {
		if extType := mime.TypeByExtension(strings.ToLower(filepath.Ext(filename))); extType != "" {
			mimeType = strings.ToLower(strings.Split(extType, ";")[0])
		} else {
			mimeType = strings.ToLower(http.DetectContentType(sample))
		}
	}
	if mimeType == "text/javascript" {
		return mimeType, nil
	}
	if _, ok := allowedMIMEs[mimeType]; !ok && mimeType == "application/x-javascript" {
		mimeType = "application/javascript"
	}
	if _, ok := allowedMIMEs[mimeType]; !ok {
		return "", ErrMIMETypeNotAllowed
	}
	return mimeType, nil
}

func ValidatePayload(filename, mimeType string, size int64, data []byte) error {
	if limit, ok := allowedMIMEs[mimeType]; !ok {
		return ErrMIMETypeNotAllowed
	} else if size > limit {
		return ErrAssetTooLarge
	}
	if mimeType == "image/svg+xml" {
		return ValidateSVG(data)
	}
	return nil
}

func ValidateSVG(data []byte) error {
	lower := strings.ToLower(string(data))
	checks := []string{"<script", "javascript:", "<foreignobject"}
	for _, check := range checks {
		if strings.Contains(lower, check) {
			return ErrUnsafeSVG
		}
	}
	if regexp.MustCompile(`\son[a-z0-9_-]+\s*=`).FindStringIndex(lower) != nil {
		return ErrUnsafeSVG
	}
	if regexp.MustCompile(`\s(?:href|src|xlink:href)\s*=\s*['"]\s*(?:https?:)?//`).FindStringIndex(lower) != nil {
		return ErrUnsafeSVG
	}
	return nil
}

func ReadUpload(upload Upload) ([]byte, string, error) {
	filename, err := SanitizeFilename(upload.Filename)
	if err != nil {
		return nil, "", err
	}
	limit := int64(MaxPDFSize + 1)
	if upload.Size > 0 && upload.Size < limit {
		limit = upload.Size + 1
	}
	data, err := io.ReadAll(io.LimitReader(upload.Reader, limit))
	if err != nil {
		return nil, "", err
	}
	if upload.Size > 0 && upload.Size != int64(len(data)) {
		// Multipart size can be unknown or approximate; trust the actual read bytes.
	}
	mimeType, err := NormalizeMIME(filename, upload.MIMEType, data[:min(len(data), 512)])
	if err != nil {
		return nil, "", err
	}
	if err := ValidatePayload(filename, mimeType, int64(len(data)), data); err != nil {
		return nil, "", err
	}
	return data, mimeType, nil
}

func readerForBytes(data []byte) io.Reader {
	return bytes.NewReader(data)
}
