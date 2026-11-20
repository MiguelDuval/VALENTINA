package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewSourceCannotBypassHumanApproval(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/valentina/sources",
		strings.NewReader(`{"name":"Unsafe auto-approved source","publisher":"External","source_type":"rss","base_url":"https://example.com/feed","approval_status":"approved","retrieval_method":"rss"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleCreateSource(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("approved source created without review: %d %s", w.Code, w.Body.String())
	}
	sources, err := s.store.ListSources(context.Background())
	if err != nil || len(sources) != 0 {
		t.Fatalf("approved source persisted: %+v %v", sources, err)
	}
}
