package httpapi

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"plat-detection-system/backend-go/internal/config"
	"plat-detection-system/backend-go/internal/errx"
)

func ParseImageUpload(w http.ResponseWriter, r *http.Request, cfg config.Config) (path string, cleanup func(), err error) {
	ct := r.Header.Get("Content-Type")
	if !strings.Contains(ct, "multipart/form-data") {
		return "", nil, errx.New("INVALID_CONTENT_TYPE", "content-type harus multipart/form-data", http.StatusBadRequest)
	}

	r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxUploadBytes)
	if err := r.ParseMultipartForm(cfg.MaxUploadBytes); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "http: request body too large") {
			return "", nil, errx.Wrap(err, "UPLOAD_TOO_LARGE", "file terlalu besar", http.StatusRequestEntityTooLarge)
		}
		return "", nil, errx.Wrap(err, "INVALID_MULTIPART", "multipart form tidak valid", http.StatusBadRequest)
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		file, header, err = r.FormFile("file")
	}
	if err != nil {
		return "", nil, errx.New("MISSING_FILE", "field file tidak ditemukan (key: image)", http.StatusBadRequest)
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".jpg"
	}
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		return "", nil, errx.New("UNSUPPORTED_MEDIA_TYPE", "format file tidak didukung (jpg/jpeg/png/webp)", http.StatusUnsupportedMediaType)
	}

	tmp, err := os.CreateTemp("", "plat-*"+ext)
	if err != nil {
		return "", nil, errx.Wrap(err, "TEMPFILE_FAILED", "gagal membuat file sementara", http.StatusInternalServerError)
	}
	tmpPath := tmp.Name()
	defer tmp.Close()

	n, err := io.Copy(tmp, file)
	if err != nil {
		_ = os.Remove(tmpPath)
		return "", nil, errx.Wrap(err, "UPLOAD_WRITE_FAILED", "gagal menyimpan file", http.StatusInternalServerError)
	}
	if n == 0 {
		_ = os.Remove(tmpPath)
		return "", nil, errx.New("EMPTY_FILE", "file kosong", http.StatusBadRequest)
	}

	return tmpPath, func() { _ = os.Remove(tmpPath) }, nil
}
