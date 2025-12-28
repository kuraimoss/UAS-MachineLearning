package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr string

	MaxUploadBytes int64

	PythonBin    string
	PythonScript string
	PythonTimeout time.Duration

	ScrapeTimeout time.Duration

	MinPlateConfidence float64

	SamsatFirestoreAPIKey string
	SamsatPageURL         string
	SamsatFirestoreBaseURL string
}

func Load() Config {
	return Config{
		Addr:               envOr("ADDR", ":8080"),
		MaxUploadBytes:     int64(envInt("MAX_UPLOAD_MB", 15)) * 1024 * 1024,
		PythonBin:          envOr("PYTHON_BIN", "python"),
		PythonScript:       os.Getenv("YOLO_PY_SCRIPT"),
		PythonTimeout:      time.Duration(envInt("YOLO_TIMEOUT_SECONDS", 120)) * time.Second,
		ScrapeTimeout:      time.Duration(envInt("SCRAPE_TIMEOUT_SECONDS", 15)) * time.Second,
		MinPlateConfidence: envFloat("MIN_PLATE_CONFIDENCE", 0.0),
		SamsatFirestoreAPIKey: envOr("SAMSAT_FIRESTORE_API_KEY", "AIzaSyCFQJzJ8y_Ka6M29gESe8Mbh-0zLe2UIWo"),
		SamsatPageURL:          envOr("SAMSAT_PAGE_URL", "https://samsat.info/cek-lokasi-daerah-plat-nomor-kendaraan-online"),
		SamsatFirestoreBaseURL: envOr("SAMSAT_FIRESTORE_BASE_URL", "https://firestore.googleapis.com/v1/projects/informasisamsat/databases/(default)/documents"),
	}
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	if n, err := strconv.Atoi(v); err == nil {
		return n
	}
	return fallback
}

func envFloat(key string, fallback float64) float64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		return f
	}
	return fallback
}
