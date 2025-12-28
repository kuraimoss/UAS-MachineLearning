package handler

import (
	"encoding/json"
	"net/http"
)

func OpenAPIHandler(w http.ResponseWriter, r *http.Request) {
	spec := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":   "Plat Detection API",
			"version": "1.0.0",
		},
		"paths": map[string]any{
			"/health": map[string]any{
				"get": map[string]any{
					"summary": "Health check",
					"responses": map[string]any{
						"200": map[string]any{
							"description": "OK",
						},
					},
				},
			},
			"/detect": map[string]any{
				"post": map[string]any{
					"summary":     "Detect plate + vehicle region",
					"requestBody": map[string]any{"required": true},
					"responses": map[string]any{
						"200": map[string]any{"description": "Success"},
						"4XX": map[string]any{"description": "Client error"},
						"5XX": map[string]any{"description": "Server error"},
					},
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(spec)
}

