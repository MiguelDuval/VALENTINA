package api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/scrypster/huginn/internal/valentina/domain"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadinessRoutes(t *testing.T) {
	store, err := domain.NewStore(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	server := NewServer(store)
	r := httptest.NewRecorder()
	server.Routes()["GET /api/v1/valentina/licenses"](r, httptest.NewRequest("GET", "/api/v1/valentina/licenses", nil))
	var license map[string]string
	if err = json.Unmarshal(r.Body.Bytes(), &license); err != nil || license["license"] != huginnLicense || len(huginnLicense) < 1000 {
		t.Fatal("license payload mismatch", err)
	}
	r = httptest.NewRecorder()
	server.Routes()["GET /api/v1/valentina/backup"](r, httptest.NewRequest("GET", "/api/v1/valentina/backup", nil))
	if r.Code != 200 || !strings.HasPrefix(r.Body.String(), "SQLite format 3") {
		t.Fatal("missing SQLite backup", r.Code)
	}
	if r.Header().Get("X-Backup-SHA256") != fmt.Sprintf("%x", sha256.Sum256(r.Body.Bytes())) {
		t.Fatal("hash differs")
	}
	for k, v := range map[string]string{"Cache-Control": "no-store", "X-Content-Type-Options": "nosniff", "X-Backup-Scope": "valentina-crm-only", "X-Backup-Schema-Version": "14", "Content-Disposition": `attachment; filename="VALENTINA-CRM.sqlite"`} {
		if r.Header().Get(k) != v {
			t.Fatal("header mismatch", k)
		}
	}
	r = httptest.NewRecorder()
	server.handleCRMBackup(r, httptest.NewRequest("GET", "/api/v1/valentina/backup?path=arbitrary", nil))
	if r.Code != 400 {
		t.Fatal("accepted request path")
	}
	store.Close()
	r = httptest.NewRecorder()
	server.handleCRMBackup(r, httptest.NewRequest("GET", "/api/v1/valentina/backup", nil))
	if r.Code != 503 || strings.Contains(r.Body.String(), "sql") {
		t.Fatal("unsafe failure")
	}
}
