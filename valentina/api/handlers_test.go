package api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/scrypster/huginn/internal/valentina/domain"
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
	req.SetPathValue("id", created.OpportunityID)
	w = httptest.NewRecorder()
	s.handleGetOpportunity(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
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
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()
	s.handleGetOpportunity(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestHandleListOpportunitiesFiltersAndPaginates(t *testing.T) {
	s := testServer(t)
	for _, item := range []*domain.Opportunity{
		{Title: "Candidate one", Summary: "A", OpportunityType: "direct_request", Status: domain.OpportunityStatusCandidate, DemandEvidenceType: domain.DemandEvidenceUnknown, SourceEvidenceIDs: []string{"ev1"}},
		{Title: "Verified", Summary: "B", OpportunityType: "direct_request", Status: domain.OpportunityStatusVerified, DemandEvidenceType: domain.DemandEvidenceUnknown, SourceEvidenceIDs: []string{"ev2"}},
		{Title: "Candidate two", Summary: "C", OpportunityType: "direct_request", Status: domain.OpportunityStatusCandidate, DemandEvidenceType: domain.DemandEvidenceUnknown, SourceEvidenceIDs: []string{"ev3"}},
	} {
		if err := s.store.CreateOpportunity(context.Background(), item); err != nil {
			t.Fatalf("CreateOpportunity: %v", err)
		}
	}
	req := httptest.NewRequest("GET", "/api/v1/valentina/opportunities?status=candidate&limit=1&offset=0", nil)
	w := httptest.NewRecorder()
	s.handleListOpportunities(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var response struct {
		Count         int                  `json:"count"`
		HasMore       bool                 `json:"has_more"`
		Opportunities []domain.Opportunity `json:"opportunities"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Count != 1 || !response.HasMore || len(response.Opportunities) != 1 || response.Opportunities[0].Status != domain.OpportunityStatusCandidate {
		t.Fatalf("status filter/pagination mismatch: %+v", response)
	}
}

func TestHandleCreateLeadDefaults(t *testing.T) {
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
	if created.Status != domain.InboundLeadStatusNew {
		t.Fatalf("Status should default to %q, got %q", domain.InboundLeadStatusNew, created.Status)
	}
	if created.ReceivedAt.IsZero() {
		t.Fatal("ReceivedAt should be set")
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
	req.SetPathValue("id", created.LeadID)
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

func TestHandleCreateLeadRejectsOversizedBody(t *testing.T) {
	s := testServer(t)
	body := "{\"request_summary\":\"A\",\"inbound_channel\":\"web\"}" + strings.Repeat(" ", maxRequestBody)
	req := httptest.NewRequest("POST", "/api/v1/valentina/leads", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleCreateLead(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for oversized request body, got %d", w.Code)
	}
}

func TestHandleListCompanies(t *testing.T) {
	s := testServer(t)
	s.store.CreateCompany(context.Background(), &domain.Company{
		Name:           "Test Co",
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
	if err := s.store.CreateSource(context.Background(), &domain.Source{
		Name:            "Test Source",
		Publisher:       "Pub",
		SourceType:      "public",
		BaseURL:         "https://example.com",
		AllowedHosts:    []string{"example.com"},
		ApprovalStatus:  "pending_review",
		RetrievalMethod: "manual",
		RateLimits: domain.SourceRateLimits{
			MinimumIntervalSeconds: 60,
			MaxParallelRequests:    1,
			TimeoutSeconds:         30,
			MaxResponseBytes:       1024,
		},
	}); err != nil {
		t.Fatalf("CreateSource: %v", err)
	}
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
	if err := s.store.CreateResearchRun(context.Background(), &domain.ResearchRun{
		StartedAt: time.Now().UTC(),
		Status:    "completed",
		SourceIDs: []string{"s1"},
	}); err != nil {
		t.Fatalf("CreateResearchRun: %v", err)
	}
	req := httptest.NewRequest("GET", "/api/v1/valentina/research-runs", nil)
	w := httptest.NewRecorder()
	s.handleListResearchRuns(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
}

func TestHandleListReviewDecisions(t *testing.T) {
	s := testServer(t)
	opportunity := &domain.Opportunity{
		Title: "Review history", Summary: "Test", OpportunityType: "direct_request",
		Status: domain.OpportunityStatusNeedsReview, DemandEvidenceType: domain.DemandEvidenceUnknown,
		SourceEvidenceIDs: []string{"ev"},
	}
	if err := s.store.CreateOpportunity(context.Background(), opportunity); err != nil {
		t.Fatalf("CreateOpportunity: %v", err)
	}
	if err := s.store.CreateReviewDecision(context.Background(), &domain.ReviewDecision{
		EntityType: "opportunity", EntityID: opportunity.OpportunityID,
		NewStatus: string(domain.OpportunityStatusQualified), DecisionReason: "Approved",
	}); err != nil {
		t.Fatalf("CreateReviewDecision: %v", err)
	}
	req := httptest.NewRequest("GET", "/api/v1/valentina/review-decisions?entity_type=opportunity&entity_id="+opportunity.OpportunityID, nil)
	w := httptest.NewRecorder()
	s.handleListReviewDecisions(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var response struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil || response.Count != 1 {
		t.Fatalf("Expected one scoped decision, got %+v err=%v", response, err)
	}
}

func TestHandleListReviewDecisionsRequiresEntity(t *testing.T) {
	s := testServer(t)
	for _, path := range []string{
		"/api/v1/valentina/review-decisions",
		"/api/v1/valentina/review-decisions?entity_type=opportunity",
		"/api/v1/valentina/review-decisions?entity_id=abc",
		"/api/v1/valentina/review-decisions?entity_type=invalid&entity_id=abc",
	} {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		s.handleListReviewDecisions(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d", path, w.Code)
		}
	}
}

func TestHandleListSourceEvidence(t *testing.T) {
	s := testServer(t)
	source := &domain.Source{
		Name: "Evidence source", Publisher: "Example", SourceType: "public_tender_portal",
		BaseURL: "https://example.test", AllowedHosts: []string{"example.test"},
		AllowedFields: []string{"title"}, RetrievalMethod: "manual",
		RateLimits: domain.SourceRateLimits{MinimumIntervalSeconds: 1, MaxParallelRequests: 1, TimeoutSeconds: 10, MaxResponseBytes: 1 << 20},
	}
	if err := s.store.CreateSource(context.Background(), source); err != nil {
		t.Fatalf("CreateSource: %v", err)
	}
	evidence := &domain.SourceEvidence{
		SourceID: source.SourceID, CanonicalURL: "https://example.test/notice", Publisher: "Example",
		RetrievalStatus: "success", Facts: []domain.EvidenceFact{{FieldPath: "title", Value: "Notice"}},
	}
	if err := s.store.CreateSourceEvidence(context.Background(), evidence); err != nil {
		t.Fatalf("CreateSourceEvidence: %v", err)
	}
	req := httptest.NewRequest("GET", "/api/v1/valentina/source-evidence?id="+evidence.EvidenceID+"&id=missing", nil)
	w := httptest.NewRecorder()
	s.handleListSourceEvidence(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var response struct {
		Count    int                     `json:"count"`
		Evidence []domain.SourceEvidence `json:"evidence"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil || response.Count != 1 || len(response.Evidence) != 1 || response.Evidence[0].EvidenceID != evidence.EvidenceID {
		t.Fatalf("unexpected evidence response: %+v err=%v", response, err)
	}
}

func TestHandleListSourceEvidenceRejectsTooManyIDs(t *testing.T) {
	s := testServer(t)
	values := make([]string, 101)
	for i := range values {
		values[i] = "id"
	}
	req := httptest.NewRequest("GET", "/api/v1/valentina/source-evidence?id="+strings.Join(values, "&id="), nil)
	w := httptest.NewRecorder()
	s.handleListSourceEvidence(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for too many IDs, got %d", w.Code)
	}
}

func TestHandleCreateReviewDecision(t *testing.T) {
	s := testServer(t)
	opportunity := &domain.Opportunity{
		Title:              "Opp",
		Summary:            "S",
		OpportunityType:    "direct_request",
		Status:             domain.OpportunityStatusNeedsReview,
		DemandEvidenceType: domain.DemandEvidenceUnknown,
		SourceEvidenceIDs:  []string{"ev"},
	}
	if err := s.store.CreateOpportunity(context.Background(), opportunity); err != nil {
		t.Fatalf("CreateOpportunity: %v", err)
	}
	body := `{"entity_type":"opportunity","entity_id":"` + opportunity.OpportunityID + `","new_status":"qualified","decision_reason":"Good"}`
	req := httptest.NewRequest("POST", "/api/v1/valentina/review-decisions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleCreateReviewDecision(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleCreateReviewDecisionRejectsBlankReason(t *testing.T) {
	s := testServer(t)
	opportunity := &domain.Opportunity{
		Title: "Reason required", Summary: "Test", OpportunityType: "direct_request",
		Status: domain.OpportunityStatusNeedsReview, DemandEvidenceType: domain.DemandEvidenceUnknown,
		SourceEvidenceIDs: []string{"ev"},
	}
	if err := s.store.CreateOpportunity(context.Background(), opportunity); err != nil {
		t.Fatalf("CreateOpportunity: %v", err)
	}
	body := `{"entity_type":"opportunity","entity_id":"` + opportunity.OpportunityID + `","new_status":"qualified","decision_reason":"  "}`
	req := httptest.NewRequest("POST", "/api/v1/valentina/review-decisions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleCreateReviewDecision(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for blank reason, got %d: %s", w.Code, w.Body.String())
	}
	got, err := s.store.GetOpportunity(context.Background(), opportunity.OpportunityID)
	if err != nil {
		t.Fatalf("GetOpportunity: %v", err)
	}
	if got.Status != domain.OpportunityStatusNeedsReview {
		t.Fatalf("status changed despite rejected decision: %q", got.Status)
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
		"POST /api/v1/valentina/intake",
		"GET /api/v1/valentina/companies",
		"GET /api/v1/valentina/sources",
		"GET /api/v1/valentina/research-runs",
		"GET /api/v1/valentina/source-evidence",
		"GET /api/v1/valentina/review-decisions",
		"GET /api/v1/valentina/health",
	}
	for _, e := range expected {
		if _, ok := routes[e]; !ok {
			t.Fatalf("Missing route: %s", e)
		}
	}
}

func TestHandleIntakeLeadSuccess(t *testing.T) {
	WithIntakeSecret("test-secret")
	defer func() { WithIntakeSecret("") }()

	s := testServer(t)
	body := `{"request_summary":"Need bathroom renovation","inbound_channel":"website","consent":true,"privacy_notice":true}`
	req := httptest.NewRequest("POST", "/api/v1/valentina/intake", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write([]byte(body))
	req.Header.Set("X-Valentina-Signature", hex.EncodeToString(mac.Sum(nil)))
	w := httptest.NewRecorder()
	s.handleIntakeLead(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	if resp["lead_id"] == "" {
		t.Fatal("Expected lead_id in response")
	}
}

func TestHandleIntakeLeadMissingConsent(t *testing.T) {
	WithIntakeSecret("test-secret")
	defer func() { WithIntakeSecret("") }()

	s := testServer(t)
	body := `{"request_summary":"Need bathroom renovation","inbound_channel":"website","consent":false,"privacy_notice":true}`
	req := httptest.NewRequest("POST", "/api/v1/valentina/intake", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write([]byte(body))
	req.Header.Set("X-Valentina-Signature", hex.EncodeToString(mac.Sum(nil)))
	w := httptest.NewRecorder()
	s.handleIntakeLead(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400, got %d", w.Code)
	}
}

func TestHandleIntakeLeadMissingSignature(t *testing.T) {
	WithIntakeSecret("test-secret")
	defer func() { WithIntakeSecret("") }()

	s := testServer(t)
	body := `{"request_summary":"Need bathroom renovation","inbound_channel":"website","consent":true,"privacy_notice":true}`
	req := httptest.NewRequest("POST", "/api/v1/valentina/intake", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleIntakeLead(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401, got %d", w.Code)
	}
}

func TestHandleIntakeLeadInvalidSignature(t *testing.T) {
	WithIntakeSecret("test-secret")
	defer func() { WithIntakeSecret("") }()

	s := testServer(t)
	body := `{"request_summary":"Need bathroom renovation","inbound_channel":"website","consent":true,"privacy_notice":true}`
	req := httptest.NewRequest("POST", "/api/v1/valentina/intake", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Valentina-Signature", "invalid-signature")
	w := httptest.NewRecorder()
	s.handleIntakeLead(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401, got %d", w.Code)
	}
}

func TestHandleIntakeLeadHoneypot(t *testing.T) {
	WithIntakeSecret("test-secret")
	defer func() { WithIntakeSecret("") }()

	s := testServer(t)
	body := `{"request_summary":"Need renovation","inbound_channel":"website","consent":true,"privacy_notice":true,"website":"http://evil.com"}`
	req := httptest.NewRequest("POST", "/api/v1/valentina/intake", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write([]byte(body))
	req.Header.Set("X-Valentina-Signature", hex.EncodeToString(mac.Sum(nil)))
	w := httptest.NewRecorder()
	s.handleIntakeLead(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for honeypot, got %d", w.Code)
	}
}
