package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/scrypster/huginn/internal/valentina/internal/domain"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	store, err := domain.NewStore(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return NewServer(store)
}

func TestHandleHealth(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest("GET", "/api/v1/valentina/health", nil)
	w := httptest.NewRecorder()
	s.handleHealth(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("Expected status ok, got %q", body["status"])
	}
}

func TestHandleCreateAndGetOpportunity(t *testing.T) {
	s := testServer(t)
	body := `{"title":"Test Opp","summary":"A test","opportunity_type":"direct_request","status":"candidate","demand_evidence_type":"unknown","source_evidence_ids":["ev1"]}`
	req := httptest.NewRequest("POST", "/api/v1/valentina/opportunities", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleCreateOpportunity(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var created domain.Opportunity
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	if created.OpportunityID == "" {
		t.Fatal("OpportunityID should be generated")
	}
	req = httptest.NewRequest("GET", "/api/v1/valentina/opportunities/"+created.OpportunityID, nil)
	w = httptest.NewRecorder()
	s.handleGetOpportunity(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
	var got domain.Opportunity
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	if got.Title != "Test Opp" {
		t.Fatalf("Title mismatch: %q", got.Title)
	}
}

func TestHandleGetOpportunityMissingID(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest("GET", "/api/v1/valentina/opportunities/", nil)
	w := httptest.NewRecorder()
	s.handleGetOpportunity(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400, got %d", w.Code)
	}
}

func TestHandleGetOpportunityNotFound(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest("GET", "/api/v1/valentina/opportunities/nonexistent", nil)
	w := httptest.NewRecorder()
	s.handleGetOpportunity(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestHandleCreateAndGetLead(t *testing.T) {
	s := testServer(t)
	body := `{"request_summary":"Test lead","inbound_channel":"web"}`
	req := httptest.NewRequest("POST", "/api/v1/valentina/leads", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleCreateLead(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var created domain.InboundLead
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	if created.LeadID == "" {
		t.Fatal("LeadID should be generated")
	}
	req = httptest.NewRequest("GET", "/api/v1/valentina/leads/"+created.LeadID, nil)
	w = httptest.NewRecorder()
	s.handleGetLead(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
	var got domain.InboundLead
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	if got.RequestSummary != "Test lead" {
		t.Fatalf("RequestSummary mismatch: %q", got.RequestSummary)
	}
}

func TestHandleCreateLeadMalformedJSON(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest("POST", "/api/v1/valentina/leads", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleCreateLead(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for malformed JSON, got %d", w.Code)
	}
}

func TestHandleCreateLeadSecondJSONRejected(t *testing.T) {
	s := testServer(t)
	body := `{"request_summary":"A","inbound_channel":"web"}{"request_summary":"B"}`
	req := httptest.NewRequest("POST", "/api/v1/valentina/leads", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleCreateLead(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for second JSON object, got %d", w.Code)
	}
}

func TestHandleListCompanies(t *testing.T) {
	s := testServer(t)
	s.store.CreateCompany(context.Background(), &domain.Company{
		Name:          "Test Co",
		NormalizedName: "test co",
	})
	req := httptest.NewRequest("GET", "/api/v1/valentina/companies", nil)
	w := httptest.NewRecorder()
	s.handleListCompanies(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
	var body struct {
		Companies []*domain.Company `json:"companies"`
		Count     int               `json:"count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	if body.Count < 1 {
		t.Fatal("Expected at least 1 company")
	}
}

func TestHandleListSources(t *testing.T) {
	s := testServer(t)
	s.store.CreateSource(context.Background(), &domain.Source{
		Name:          "Test Source",
		Publisher:     "Pub",
		SourceType:    "public",
		BaseURL:       "https://example.com",
		AllowedHosts:  []string{"example.com"},
		ApprovalStatus: "pending_review",
		RetrievalMethod: "manual",
		RateLimits: domain.SourceRateLimits{
			MinimumIntervalSeconds: 60,
			MaxParallelRequests: 1,
			TimeoutSeconds: 30,
			MaxResponseBytes: 1024,
		},
	})
	req := httptest.NewRequest("GET", "/api/v1/valentina/sources", nil)
	w := httptest.NewRecorder()
	s.handleListSources(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
	var body struct {
		Sources []*domain.Source `json:"sources"`
		Count   int              `json:"count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	if body.Count < 1 {
		t.Fatal("Expected at least 1 source")
	}
}

func TestHandleListResearchRuns(t *testing.T) {
	s := testServer(t)
	s.store.CreateResearchRun(context.Background(), &domain.ResearchRun{
		StartedAt:   time.Now().UTC(),
		Status:      "completed",
		SourceIDs:   []string{"s1"},
	})
	req := httptest.NewRequest("GET", "/api/v1/valentina/research-runs", nil)
	w := httptest.NewRecorder()
	s.handleListResearchRuns(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
}

func TestHandleListReviewDecisions(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest("GET", "/api/v1/valentina/review-decisions", nil)
	w := httptest.NewRecorder()
	s.handleListReviewDecisions(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
}

func TestHandleCreateReviewDecision(t *testing.T) {
	s := testServer(t)
	s.store.CreateOpportunity(context.Background(), &domain.Opportunity{
		Title:               "Opp",
		Summary:             "S",
		OpportunityType:     "direct_request",
		Status:              domain.OpportunityStatusNeedsReview,
		DemandEvidenceType:  domain.DemandEvidenceUnknown,
		SourceEvidenceIDs:   []string{"ev"},
	})
	body := `{"entity_type":"opportunity","new_status":"qualified","decision_reason":"Good"}`
	req := httptest.NewRequest("POST", "/api/v1/valentina/review-decisions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleCreateReviewDecision(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRoutesReturnsAllEndpoints(t *testing.T) {
	s := testServer(t)
	routes := s.Routes()
	expected := []string{
		"GET /api/v1/valentina/opportunities",
		"POST /api/v1/valentina/opportunities",
		"GET /api/v1/valentina/opportunities/{id}",
		"GET /api/v1/valentina/leads",
		"POST /api/v1/valentina/leads",
		"GET /api/v1/valentina/leads/{id}",
		"GET /api/v1/valentina/companies",
		"GET /api/v1/valentina/sources",
		"GET /api/v1/valentina/research-runs",
		"GET /api/v1/valentina/review-decisions",
		"GET /api/v1/valentina/health",
	}
	for _, e := range expected {
		if _, ok := routes[e]; !ok {
			t.Fatalf("Missing route: %s", e)
		}
	}
}
