package handler

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"time"

	"plat-detection-system/backend-go/internal/httpapi"
)

type DocsHandler struct {
	tmpl *template.Template
}

type docsViewModel struct {
	Title             string
	Brand             string
	Heading           string
	Year              int
	SampleSuccessJSON template.HTML
	SampleErrorJSON   template.HTML
}

func NewDocsHandler(templatePath string) (*DocsHandler, error) {
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, err
	}
	return &DocsHandler{tmpl: t}, nil
}

func (h *DocsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if h == nil || h.tmpl == nil {
		httpapi.WriteError(w, r.Context(), start, errors.New("docs template not configured"))
		return
	}

	vm := docsViewModel{
		Title:   "Plat Detection API · Docs",
		Brand:   "plat-detection-system",
		Heading: "Plat Detection API",
		Year:    time.Now().Year(),
		SampleSuccessJSON: template.HTML(mustJSON(sampleSuccess())),
		SampleErrorJSON:   template.HTML(mustJSON(sampleError())),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.Execute(w, vm); err != nil {
		httpapi.Logger(r.Context()).Error("render docs failed", "error", err.Error())
		httpapi.WriteError(w, r.Context(), start, err)
		return
	}
}

func sampleSuccess() any {
	return map[string]any{
		"status":             "success",
		"request_id":         "00000000-0000-4000-8000-000000000000",
		"processing_time_ms": 12345,
		"data": map[string]any{
			"plate": map[string]any{
				"raw":       "XX0000XXX",
				"formatted": "XX 0000 XXX",
				"confidence": 0.9629,
				"components": map[string]any{
					"prefix": "XX",
					"number": "0000",
					"suffix": "XXX",
				},
			},
			"vehicle_region": map[string]any{
				"daerah":         "Kota Contoh",
				"provinsi":       "Provinsi Contoh",
				"wilayah_samsat": "Samsat Contoh",
				"alamat_samsat":  "Jl. Contoh No. 1",
				"source":         "https://samsat.info",
			},
		},
		"meta": map[string]any{
			"model": map[string]any{
				"yolo":     "YOLOv8 custom",
				"ocr":      "TrOCR",
				"language": "id",
			},
		},
	}
}

func sampleError() any {
	return map[string]any{
		"status":             "error",
		"request_id":         "00000000-0000-4000-8000-000000000000",
		"processing_time_ms": 512,
		"error": map[string]any{
			"code":    "INFERENCE_FAILED",
			"message": "Plat tidak terdeteksi",
		},
	}
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
