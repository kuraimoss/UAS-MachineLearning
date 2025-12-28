package router

import (
	"net/http"

	"plat-detection-system/backend-go/internal/config"
	"plat-detection-system/backend-go/internal/handler"
	"plat-detection-system/backend-go/internal/httpapi"
	"plat-detection-system/backend-go/internal/scraper"
	"plat-detection-system/backend-go/internal/service"
)

func New(cfg config.Config) http.Handler {
	mux := http.NewServeMux()

	yolo := service.NewYOLOService(cfg)
	samsat := service.NewSamsatService(scraper.NewSamsatScraper(cfg))
	detectSvc := service.NewDetectService(cfg, yolo, samsat)
	h := handler.New(cfg, detectSvc)
	docs, _ := handler.NewDocsHandler("web/templates/docs.html")

	staticFS := http.FileServer(http.Dir("web/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticFS))

	mux.HandleFunc("GET /healthz", h.Healthz)
	mux.HandleFunc("GET /health", h.Healthz)
	mux.HandleFunc("POST /detect", h.Detect)
	if docs != nil {
		mux.Handle("GET /", docs)
		mux.Handle("GET /docs", docs)
	}
	mux.HandleFunc("GET /openapi.json", handler.OpenAPIHandler)

	var handler http.Handler = mux
	handler = httpapi.RecoverMiddleware(handler)
	handler = httpapi.LoggingMiddleware(handler)
	handler = httpapi.RequestIDMiddleware(handler)
	return handler
}
