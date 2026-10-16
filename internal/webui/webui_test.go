package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebUI_IndexAndFallback(t *testing.T) {
	h := Handler()

	// 1. Root / should return index.html
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for root, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "<!doctype html>") && !strings.Contains(rec.Body.String(), "<html") {
		t.Fatalf("expected HTML body, got: %s", rec.Body.String())
	}

	// 2. SPA Route (e.g. /acme/my-repo) should fallback to index.html with 200
	reqSPA := httptest.NewRequest(http.MethodGet, "/acme/my-repo", nil)
	recSPA := httptest.NewRecorder()
	h.ServeHTTP(recSPA, reqSPA)

	if recSPA.Code != http.StatusOK {
		t.Fatalf("expected 200 for SPA route, got %d", recSPA.Code)
	}
	if !strings.Contains(recSPA.Body.String(), "<!doctype html>") && !strings.Contains(recSPA.Body.String(), "<html") {
		t.Fatalf("expected HTML fallback for SPA route, got: %s", recSPA.Body.String())
	}
}
