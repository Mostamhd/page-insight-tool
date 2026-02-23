package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/moustafa/home24/internal/analyzer"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

type analyzeRequest struct {
	URL string `json:"url"`
}

type errorResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func Analyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req analyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	parsed, err := url.ParseRequestURI(req.URL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		writeError(w, http.StatusBadRequest, "url must be an absolute URL with http or https scheme")
		return
	}

	body, statusCode, finalURL, err := fetchURL(r.Context(), req.URL)
	if err != nil {
		writeError(w, http.StatusBadGateway, fmt.Sprintf("failed to fetch URL: %v", err))
		return
	}

	if statusCode >= 400 {
		writeError(w, statusCode, fmt.Sprintf("upstream returned status %d", statusCode))
		return
	}

	result, err := analyzer.Analyze(r.Context(), body, finalURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("analysis failed: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func fetchURL(ctx context.Context, rawURL string) (body []byte, statusCode int, finalURL string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, 0, "", fmt.Errorf("creating request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, "", fmt.Errorf("fetching URL: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, 0, "", fmt.Errorf("reading response: %w", err)
	}

	return b, resp.StatusCode, resp.Request.URL.String(), nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("error encoding response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{
		StatusCode: status,
		Message:    message,
	})
}
