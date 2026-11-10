package domain

import (
	"context"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

func TestCreateAndGetOpportunity(t *testing.T) {
	s := testStore(t)
	o := &Opportunity{
		Title:               "Test Opportunity",
		Summary:             "A test summary",
		OpportunityType:     "direct_request",
		Status:              OpportunityStatusCandidate,
		Geography:           &Location{Municipality: strPtr("Valencia")},
		DemandEvidenceType:  DemandEvidenceDirectRequest,
		SourceEvidenceIDs:   []string{"ev1"},
	}
	if err := s.CreateOpportunity(context.Background(), o); err != nil {
		t.Fatalf("CreateOpportunity: %v", err)
	}
	if o.OpportunityID == "" {
		t.Fatal("OpportunityID should be generated")
	}
	got, err := s.GetOpportunity(context.Background(), o.OpportunityID)
	if err != nil {
		t.Fatalf("GetOpportunity: %v", err)
	}
	if got.Title != o.Title {
		t.Fatalf("Title mismatch: got %q want %q", got.Title, o.Title)
	}
}

func TestGetOpportunityMissingID(t *testing.T) {
	s := testStore(t)
	_, err := s.GetOpportunity(context.Background(), "")
	if err == nil {
		t.Fatal("GetOpportunity with empty ID should return error")
	}
}

func TestGetOpportunityNotFound(t *testing.T) {
	s := testStore(t)
	_, err := s.GetOpportunity(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("GetOpportunity with nonexistent ID should return error")
	}
}

func TestListOpportunities(t *testing.T) {
	s := testStore(t)
	for i := 0; i < 3; i++ {
		o := &Opportunity{
			Title:               "Opp",
			Summary:             "Summary",
			OpportunityType:     "direct_request",
			Status:              OpportunityStatusCandidate,
			DemandEvidenceType:  DemandEvidenceUnknown,
			SourceEvidenceIDs:   []string{"ev"},
		}
		if err := s.CreateOpportunity(context.Background(), o); err != nil {
			t.Fatalf("CreateOpportunity: %v", err)
		}
	}
	list, err := s.ListOpportunities(context.Background(), "", 10, 0)
	if err != nil {
		t.Fatalf("ListOpportunities: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("Expected 3 opportunities, got %d", len(list))
	}
}

func TestCreateInboundLeadIDBeforeValidate(t *testing.T) {
	s := testStore(t)
	l := &InboundLead{
		RequestSummary: "Test",
		InboundChannel: "web",
	}
	if err := s.CreateInboundLead(context.Background(), l); err != nil {
		t.Fatalf("CreateInboundLead should auto-generate ID: %v", err)
	}
	if l.LeadID == "" {
		t.Fatal("LeadID should be auto-generated")
	}
}

func TestCreateInboundLeadDefaults(t *testing.T) {
	s := testStore(t)
	l := &InboundLead{
		RequestSummary:   "Test lead",
		InboundChannel:   "web",
		ContactInfo:      nil,
		AIAssessment:     nil,
	}
	if err := s.CreateInboundLead(context.Background(), l); err != nil {
		t.Fatalf("CreateInboundLead should apply defaults and validate: %v", err)
	}
	if l.LeadID == "" {
		t.Fatal("LeadID should be auto-generated")
	}
	if l.Status != InboundLeadStatusNew {
		t.Fatalf("Status should default to %q, got %q", InboundLeadStatusNew, l.Status)
	}
	if l.ReceivedAt.IsZero() {
		t.Fatal("ReceivedAt should be set to current time")
	}
	got, err := s.GetInboundLead(context.Background(), l.LeadID)
	if err != nil {
		t.Fatalf("GetInboundLead: %v", err)
	}
	if got.Status != InboundLeadStatusNew {
		t.Fatalf("Persisted status should be %q, got %q", InboundLeadStatusNew, got.Status)
	}
	if !got.ReceivedAt.After(l.ReceivedAt.Add(-time.Second)) && !got.ReceivedAt.Before(l.ReceivedAt.Add(time.Second)) {
		t.Fatalf("Persisted ReceivedAt should match, got %v expected ~%v", got.ReceivedAt, l.ReceivedAt)
	}
}

func TestListInboundLeadsNilNested(t *testing.T) {
	s := testStore(t)
	l := &InboundLead{
		RequestSummary:   "Test",
		InboundChannel:   "web",
		ContactInfo:      nil,
		AIAssessment:     nil,
	}
	if err := s.CreateInboundLead(context.Background(), l); err != nil {
		t.Fatalf("CreateInboundLead: %v", err)
	}
	list, err := s.ListInboundLeads(context.Background(), "", 10, 0)
	if err != nil {
		t.Fatalf("ListInboundLeads: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("Expected 1 lead, got %d", len(list))
	}
	if list[0].ContactInfo != nil {
		t.Fatal("ContactInfo should be nil when not stored")
	}
	if list[0].AIAssessment != nil {
		t.Fatal("AIAssessment should be nil when not stored")
	}
}

func TestCreateAndGetCompany(t *testing.T) {
	s := testStore(t)
	c := &Company{
		Name:          "Test Co",
		NormalizedName: "test co",
	}
	if err := s.CreateCompany(context.Background(), c); err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	if c.CompanyID == "" {
		t.Fatal("CompanyID should be generated")
	}
	got, err := s.GetCompany(context.Background(), c.CompanyID)
	if err != nil {
		t.Fatalf("GetCompany: %v", err)
	}
	if got.Name != c.Name {
		t.Fatal("Name mismatch")
	}
}

func TestListCompanies(t *testing.T) {
	s := testStore(t)
	for i := 0; i < 2; i++ {
		c := &Company{
			Name:          "Co",
			NormalizedName: "co",
		}
		if err := s.CreateCompany(context.Background(), c); err != nil {
			t.Fatalf("CreateCompany: %v", err)
		}
	}
	list, err := s.ListCompanies(context.Background())
	if err != nil {
		t.Fatalf("ListCompanies: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("Expected 2 companies, got %d", len(list))
	}
}

func TestCreateAndListSources(t *testing.T) {
	s := testStore(t)
	src := &Source{
		Name:             "Test Source",
		Publisher:        "Pub",
		SourceType:       "public",
		BaseURL:          "https://example.com",
		AllowedHosts:     []string{"example.com"},
		AllowedFields:    []string{"title"},
		ApprovalStatus:   "pending_review",
		RetrievalMethod:  "manual",
		RateLimits:       SourceRateLimits{MinimumIntervalSeconds: 60, MaxParallelRequests: 1, TimeoutSeconds: 30, MaxResponseBytes: 1024},
	}
	if err := s.CreateSource(context.Background(), src); err != nil {
		t.Fatalf("CreateSource: %v", err)
	}
	if src.SourceID == "" {
		t.Fatal("SourceID should be generated")
	}
	list, err := s.ListSources(context.Background())
	if err != nil {
		t.Fatalf("ListSources: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("Expected 1 source, got %d", len(list))
	}
	if list[0].ApprovalStatus != "pending_review" {
		t.Fatal("ApprovalStatus should be pending_review")
	}
}

func TestCreateAndListResearchRuns(t *testing.T) {
	s := testStore(t)
	r := &ResearchRun{
		StartedAt:      time.Now().UTC(),
		Status:         "completed",
		SourceIDs:      []string{"s1"},
		CandidateCount: 5,
		VerifiedCount:  3,
	}
	if err := s.CreateResearchRun(context.Background(), r); err != nil {
		t.Fatalf("CreateResearchRun: %v", err)
	}
	list, err := s.ListResearchRuns(context.Background())
	if err != nil {
		t.Fatalf("ListResearchRuns: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("Expected 1 research run, got %d", len(list))
	}
	if list[0].Status != "completed" {
		t.Fatal("Status mismatch")
	}
}

func TestCreateReviewDecisionTransactional(t *testing.T) {
	s := testStore(t)
	o := &Opportunity{
		Title:               "Opp",
		Summary:             "S",
		OpportunityType:     "direct_request",
		Status:              OpportunityStatusNeedsReview,
		DemandEvidenceType:  DemandEvidenceUnknown,
		SourceEvidenceIDs:   []string{"ev"},
	}
	if err := s.CreateOpportunity(context.Background(), o); err != nil {
		t.Fatalf("CreateOpportunity: %v", err)
	}
	d := &ReviewDecision{
		EntityType:      "opportunity",
		EntityID:        o.OpportunityID,
		NewStatus:       OpportunityStatusQualified,
		DecisionReason:  "Looks good",
	}
	if err := s.CreateReviewDecision(context.Background(), d); err != nil {
		t.Fatalf("CreateReviewDecision: %v", err)
	}
	got, err := s.GetOpportunity(context.Background(), o.OpportunityID)
	if err != nil {
		t.Fatalf("GetOpportunity: %v", err)
	}
	if got.Status != OpportunityStatusQualified {
		t.Fatalf("Status should be qualified, got %q", got.Status)
	}
	if d.PreviousStatus != string(OpportunityStatusNeedsReview) {
		t.Fatalf("PreviousStatus should be needs_review, got %q", d.PreviousStatus)
	}
}

func TestCreateReviewDecisionInvalidTransition(t *testing.T) {
	s := testStore(t)
	o := &Opportunity{
		Title:               "Opp",
		Summary:             "S",
		OpportunityType:     "direct_request",
		Status:              OpportunityStatusArchived,
		DemandEvidenceType:  DemandEvidenceUnknown,
		SourceEvidenceIDs:   []string{"ev"},
	}
	if err := s.CreateOpportunity(context.Background(), o); err != nil {
		t.Fatalf("CreateOpportunity: %v", err)
	}
	d := &ReviewDecision{
		EntityType:    "opportunity",
		EntityID:      o.OpportunityID,
		NewStatus:     OpportunityStatusQualified,
		DecisionReason: "Bad transition",
	}
	err := s.CreateReviewDecision(context.Background(), d)
	if err == nil {
		t.Fatal("CreateReviewDecision should reject invalid status transition")
	}
}

func TestCreateReviewDecisionEntityNotFound(t *testing.T) {
	s := testStore(t)
	d := &ReviewDecision{
		EntityType:   "opportunity",
		EntityID:     "nonexistent",
		NewStatus:    OpportunityStatusQualified,
		DecisionReason: "test",
	}
	err := s.CreateReviewDecision(context.Background(), d)
	if err == nil {
		t.Fatal("CreateReviewDecision should fail for nonexistent entity")
	}
}

func TestMigrationIdempotent(t *testing.T) {
	s := testStore(t)
	if err := s.migrate(context.Background()); err != nil {
		t.Fatalf("First migrate: %v", err)
	}
	if err := s.migrate(context.Background()); err != nil {
		t.Fatalf("Second migrate should be idempotent: %v", err)
	}
}
