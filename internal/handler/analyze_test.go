package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/moustafa/home24/internal/analyzer"
)

func TestAnalyze_Success(t *testing.T) {
	htmlBody := `<!DOCTYPE html><html><head><title>Test Page</title></head><body><h1>Hello</h1><h2>World</h2></body></html>`
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(htmlBody))
	}))
	defer upstream.Close()

	body, _ := json.Marshal(analyzeRequest{URL: upstream.URL})
	req := httptest.NewRequest(http.MethodPost, "/api/analyze", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	Analyze(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var resp analyzer.AnalyzeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.HTMLVersion != "HTML5" {
		t.Errorf("HTMLVersion = %q, want %q", resp.HTMLVersion, "HTML5")
	}
	if resp.Title != "Test Page" {
		t.Errorf("Title = %q, want %q", resp.Title, "Test Page")
	}
	if resp.Headings["h1"] != 1 {
		t.Errorf("Headings[h1] = %d, want 1", resp.Headings["h1"])
	}
	if resp.Headings["h2"] != 1 {
		t.Errorf("Headings[h2] = %d, want 1", resp.Headings["h2"])
	}
}

func TestAnalyze_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/analyze", strings.NewReader("not json"))
	rec := httptest.NewRecorder()

	Analyze(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestAnalyze_InvalidURL(t *testing.T) {
	body, _ := json.Marshal(analyzeRequest{URL: "not-a-url"})
	req := httptest.NewRequest(http.MethodPost, "/api/analyze", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	Analyze(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestAnalyze_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/analyze", nil)
	rec := httptest.NewRecorder()

	Analyze(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestAnalyze_FetchError(t *testing.T) {
	body, _ := json.Marshal(analyzeRequest{URL: "http://localhost:1"})
	req := httptest.NewRequest(http.MethodPost, "/api/analyze", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	Analyze(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", rec.Code)
	}
}
