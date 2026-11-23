package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/scrypster/huginn/internal/valentina/domain"
	"github.com/scrypster/huginn/internal/valentina/sourceadapter"
)

func TestSourceRunEligibilityContract(t *testing.T) {
	s := testServer(t)
	approvedRSSSource(t, s)
	w := httptest.NewRecorder()
	s.handleListSources(w, httptest.NewRequest("GET", "/", nil))
	var response struct {
		Sources []struct {
			ManualRunAllowed bool   `json:"manual_run_allowed"`
			ManualRunReason  string `json:"manual_run_reason"`
		} `json:"sources"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Sources) != 1 || !response.Sources[0].ManualRunAllowed {
		t.Fatal("approved RSS must expose server eligibility")
	}
}

const syntheticRSS = `<rss><channel>
<item><title>Reforma de vivienda Valencia</title><link>https://example.com/obras/1</link>
<description>Se busca profesional para albañilería</description></item>
<item><title>Football results</title><link>https://example.com/football</link>
<description>Nothing related</description></item>
</channel></rss>`

func approvedRSSSource(t *testing.T, s *Server) *domain.Source {
	t.Helper()
	now := time.Now().UTC()
	reason := "Owner-reviewed official public RSS terms and path scope"
	src := &domain.Source{
		Name: "Public test feed", Publisher: "Example publisher",
		SourceType: "rss", BaseURL: "https://example.com/feed",
		AllowedHosts: []string{"example.com"}, AllowedPathPrefixes: []string{"/"},
		AllowedFields: []string{"title", "summary", "link"}, RetrievalMethod: "rss",
		ApprovalStatus: "approved", ApprovalReason: &reason, TermsReviewedAt: &now,
		RateLimits: domain.SourceRateLimits{MaxParallelRequests: 1, TimeoutSeconds: 10, MaxResponseBytes: 2048},
	}
	if err := s.store.CreateSource(context.Background(), src); err != nil {
		t.Fatal(err)
	}
	return src
}

// fixtureFeedClient has no transport: the complete adapter-to-SQLite test stays offline.
type fixtureFeedClient struct {
	body  string
	calls int
}

func (c *fixtureFeedClient) PolicyValidated() bool { return true }
func (c *fixtureFeedClient) Do(req *http.Request) (*http.Response, error) {
	c.calls++
	return &http.Response{StatusCode: http.StatusOK, Request: req,
		Header: http.Header{"Content-Type": {"application/rss+xml"}}, Body: io.NopCloser(strings.NewReader(c.body))}, nil
}

func TestManualRSSResearchPersistsCandidatesAndDeduplicates(t *testing.T) {
	s := testServer(t)
	src := approvedRSSSource(t, s)
	client := &fixtureFeedClient{body: syntheticRSS}
	adapter := sourceadapter.NewRSSAdapter("synthetic", client, time.Second)
	s.fetchSource = func(ctx context.Context, item *domain.Source) (*sourceadapter.FetchResult, error) {
		if item.SourceID != src.SourceID {
			t.Fatalf("wrong source")
		}
		return adapter.Fetch(ctx, item.SourceID, item.BaseURL)
	}
	execute := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/valentina/research-runs",
			strings.NewReader(`{"source_ids":["`+src.SourceID+`"]}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		s.handleCreateResearchRun(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected run created, got %d: %s", w.Code, w.Body.String())
		}
		return w
	}
	execute()
	opportunities, err := s.store.ListOpportunities(context.Background(), "", 10, 0)
	if err != nil || len(opportunities) != 1 {
		t.Fatalf("expected one renovation candidate, got %+v, %v", opportunities, err)
	}
	opp := opportunities[0]
	if opp.DemandEvidenceType != domain.DemandEvidenceUnknown || opp.Status != domain.OpportunityStatusCandidate ||
		len(opp.SourceEvidenceIDs) != 1 {
		t.Fatalf("feed must remain an unverified candidate: %+v", opp)
	}
	evidence, err := s.store.ListSourceEvidenceByIDs(context.Background(), opp.SourceEvidenceIDs)
	if err != nil || len(evidence) != 1 || evidence[0].RetrievalStatus != "feed_observed" ||
		evidence[0].CanonicalURL != "https://example.com/obras/1" || evidence[0].ObservedAt.IsZero() ||
		len(evidence[0].Facts) != 3 || evidence[0].Facts[0].EvidenceLocator == nil || *evidence[0].Facts[0].EvidenceLocator != src.BaseURL {
		t.Fatalf("source provenance lost: %+v %v", evidence, err)
	}
	client.body = strings.ReplaceAll(syntheticRSS, "https://example.com/obras/1", "https://EXAMPLE.com:443/obras/1?utm_source=repeat#top")
	execute()
	if client.calls != 2 {
		t.Fatalf("expected only two feed requests, got %d", client.calls)
	}
	opportunities, _ = s.store.ListOpportunities(context.Background(), "", 10, 0)
	runs, err := s.store.ListResearchRuns(context.Background())
	if err != nil || len(runs) != 2 || len(opportunities) != 1 {
		t.Fatalf("duplicate candidate persisted: runs=%d opportunities=%d err=%v", len(runs), len(opportunities), err)
	}
	if runs[0].DuplicateCount != 1 || runs[0].CompletedAt == nil || runs[0].Status != "completed" {
		t.Fatalf("run progress missing: %+v", runs[0])
	}
}

func TestMalformedFeedDoesNotPersistCandidate(t *testing.T) {
	s := testServer(t)
	src := approvedRSSSource(t, s)
	client := &fixtureFeedClient{body: `<rss><channel><item><title>Reforma</title><link>https://example.com/1</link></item>`}
	adapter := sourceadapter.NewRSSAdapter("synthetic", client, time.Second)
	s.fetchSource = func(ctx context.Context, item *domain.Source) (*sourceadapter.FetchResult, error) {
		return adapter.Fetch(ctx, item.SourceID, item.BaseURL)
	}
	w := httptest.NewRecorder()
	s.handleCreateResearchRun(w, httptest.NewRequest(http.MethodPost, "/api/v1/valentina/research-runs",
		strings.NewReader(`{"source_ids":["`+src.SourceID+`"]}`)))
	runs, err := s.store.ListResearchRuns(context.Background())
	if w.Code != http.StatusCreated || err != nil || len(runs) != 1 || runs[0].Status != "failed" {
		t.Fatal("malformed feed must produce a failed run")
	}
	opps, err := s.store.ListOpportunities(context.Background(), "", 10, 0)
	if err != nil || len(opps) != 0 {
		t.Fatal("malformed feed created candidates")
	}
}

func TestManualRSSFailureRecordedWithoutFakeOpportunity(t *testing.T) {
	s := testServer(t)
	src := approvedRSSSource(t, s)
	s.fetchSource = func(context.Context, *domain.Source) (*sourceadapter.FetchResult, error) {
		return nil, errors.New("synthetic fetch failure")
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/valentina/research-runs", strings.NewReader(`{"source_ids":["`+src.SourceID+`"]}`))
	w := httptest.NewRecorder()
	s.handleCreateResearchRun(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("failed attempt should have a persisted run, got %d %s", w.Code, w.Body.String())
	}
	runs, err := s.store.ListResearchRuns(context.Background())
	if err != nil || len(runs) != 1 || runs[0].Status != "failed" || runs[0].CompletedAt == nil {
		t.Fatalf("failure was not persisted correctly: %+v %v", runs, err)
	}
	opps, _ := s.store.ListOpportunities(context.Background(), "", 10, 0)
	if len(opps) != 0 {
		t.Fatal("failed fetch must not create opportunities")
	}
}

func TestApprovedSourceWithoutTermsIsBlockedBeforeFetch(t *testing.T) {
	s := testServer(t)
	src := &domain.Source{Name: "Unreviewed", Publisher: "Test", SourceType: "rss", BaseURL: "https://example.com/feed",
		AllowedHosts: []string{"example.com"}, AllowedPathPrefixes: []string{"/"},
		AllowedFields: []string{"title"}, RetrievalMethod: "rss", ApprovalStatus: "approved",
		RateLimits: domain.SourceRateLimits{MaxParallelRequests: 1, TimeoutSeconds: 10, MaxResponseBytes: 1024}}
	if err := s.store.CreateSource(context.Background(), src); err != nil {
		t.Fatal(err)
	}
	s.fetchSource = func(context.Context, *domain.Source) (*sourceadapter.FetchResult, error) {
		t.Fatal("unreviewed source fetched")
		return nil, nil
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/valentina/research-runs",
		strings.NewReader(`{"source_ids":["`+src.SourceID+`"]}`))
	w := httptest.NewRecorder()
	s.handleCreateResearchRun(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("unreviewed source: %d %s", w.Code, w.Body.String())
	}
	runs, _ := s.store.ListResearchRuns(context.Background())
	if len(runs) != 0 {
		t.Fatal("unreviewed source started research")
	}
}

func TestApprovedFeedRejectsLocalPrivateAddress(t *testing.T) {
	now := time.Now().UTC()
	reason := "explicitly reviewed"
	src := &domain.Source{SourceID: "private-test", Name: "Blocked", Publisher: "Blocked", SourceType: "rss",
		BaseURL: "https://127.0.0.1/feed", AllowedHosts: []string{"127.0.0.1"},
		AllowedPathPrefixes: []string{"/"}, RetrievalMethod: "rss",
		ApprovalStatus: "approved", TermsReviewedAt: &now, ApprovalReason: &reason}
	if _, err := fetchApprovedFeed(context.Background(), src); err == nil {
		t.Fatal("private loopback fetch unexpectedly permitted")
	}
}

// TestGate1OfflineSmoke traverses the real routes, adapters and on-disk SQLite.
// Only feed transport is substituted; no source or model network calls occur.
func TestGate1OfflineSmoke(t *testing.T) {
	const secret = "synthetic-secret-01234567890123456789"
	WithIntakeSecret(secret)
	defer WithIntakeSecret("")
	t.Setenv("VALENTINA_INTAKE_ALLOW_LEGACY_V1", "")
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "gate1.db")
	store, err := domain.NewStore(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { store.Close() }()
	s := NewServer(store)
	mux := http.NewServeMux()
	register := func() {
		mux = http.NewServeMux()
		for route, handler := range s.Routes() {
			mux.HandleFunc(route, handler)
		}
	}
	register()
	post := func(route, body string, signed bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", route, strings.NewReader(body))
		r.RemoteAddr = "192.0.2.80:12345"
		if signed {
			signV2(r, body, secret, fmt.Sprint(time.Now().Unix()))
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	body := `{"request_summary":"Synthetic repair inquiry","inbound_channel":"website","consent":true,"privacy_notice":true,"privacy_notice_version":"fixture-v1","idempotency_key":"gate1-smoke-001"}`
	w := post("/api/v1/valentina/intake", body, true)
	if w.Code != 201 {
		t.Fatalf("intake status=%d", w.Code)
	}
	var intake intakeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &intake); err != nil {
		t.Fatal(err)
	}
	w = post("/api/v1/valentina/review-decisions", `{"entity_type":"inbound_lead","entity_id":"`+intake.LeadID+`","new_status":"triage","decision_reason":"Synthetic human review"}`, false)
	if w.Code != 201 {
		t.Fatalf("triage: %d %s", w.Code, w.Body.String())
	}
	src := approvedRSSSource(t, s)
	client := &fixtureFeedClient{body: syntheticRSS}
	adapter := sourceadapter.NewRSSAdapter("fixture", client, time.Second)
	s.fetchSource = func(ctx context.Context, item *domain.Source) (*sourceadapter.FetchResult, error) {
		return adapter.Fetch(ctx, item.SourceID, item.BaseURL)
	}
	w = post("/api/v1/valentina/research-runs", `{"source_ids":["`+src.SourceID+`"]}`, false)
	if w.Code != 201 {
		t.Fatalf("research status=%d", w.Code)
	}
	opportunities, err := store.ListOpportunities(ctx, "", 10, 0)
	if err != nil || len(opportunities) != 1 {
		t.Fatalf("candidates=%d err=%v", len(opportunities), err)
	}
	opp := opportunities[0]
	if opp.Status != domain.OpportunityStatusCandidate || opp.DemandEvidenceType != domain.DemandEvidenceUnknown {
		t.Fatal("feed observation falsely verified")
	}
	w = post("/api/v1/valentina/review-decisions", `{"entity_type":"opportunity","entity_id":"`+opp.OpportunityID+`","new_status":"needs_review","decision_reason":"Original page still requires inspection"}`, false)
	if w.Code != 201 {
		t.Fatalf("review: %d %s", w.Code, w.Body.String())
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	store, err = domain.NewStore(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	s = NewServer(store)
	register()
	w = post("/api/v1/valentina/intake", body, true)
	var retry intakeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &retry); err != nil {
		t.Fatal(err)
	}
	if w.Code != 200 || retry.LeadID != intake.LeadID || !retry.Duplicate {
		t.Fatal("signed retry after reopen created another lead")
	}
	w = post("/api/v1/valentina/intake", strings.Replace(body, "repair inquiry", "changed inquiry", 1), true)
	if w.Code != 409 {
		t.Fatalf("changed body status=%d", w.Code)
	}
	lead, err := store.GetInboundLead(ctx, intake.LeadID)
	if err != nil || lead.Status != domain.InboundLeadStatusTriage || lead.PrivacyNoticeVersion == nil || *lead.PrivacyNoticeVersion != "fixture-v1" {
		t.Fatal("lead triage/notice lost after reopen")
	}
	persisted, err := store.GetOpportunity(ctx, opp.OpportunityID)
	if err != nil || persisted.Status != domain.OpportunityStatusNeedsReview {
		t.Fatal("cold review lost after reopen")
	}
	evidence, err := store.ListSourceEvidenceByIDs(ctx, persisted.SourceEvidenceIDs)
	if err != nil || len(evidence) != 1 || evidence[0].CanonicalURL != "https://example.com/obras/1" || evidence[0].ObservedAt.IsZero() || evidence[0].RetrievalStatus != "feed_observed" {
		t.Fatal("evidence provenance lost")
	}
	decisions, err := store.ListReviewDecisions(ctx)
	if err != nil || len(decisions) != 2 {
		t.Fatal("human decisions lost")
	}
	runs, err := store.ListResearchRuns(ctx)
	if err != nil || len(runs) != 1 || runs[0].CandidateCount != 1 || runs[0].VerifiedCount != 0 {
		t.Fatal("run counts lost or falsely verified")
	}
	// A fresh signature on malformed input must not reflect a submitted field.
	w = post("/api/v1/valentina/intake", `{"private-person@example.test":true}`, true)
	if w.Code != 400 || strings.Contains(w.Body.String(), "private-person") {
		t.Fatal("JSON error leaked submitted data")
	}
	// Stale signatures are rejected even if the idempotency key already exists.
	r := httptest.NewRequest("POST", "/api/v1/valentina/intake", strings.NewReader(body))
	r.RemoteAddr = "192.0.2.81:12345"
	signV2(r, body, secret, fmt.Sprint(time.Now().Add(-6*time.Minute).Unix()))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatal("stale replay accepted")
	}
}

func TestManualResearchRejectsIneligibleSources(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*domain.Source)
	}{
		{"pending", func(s *domain.Source) { s.ApprovalStatus = "pending_review" }},
		{"paused", func(s *domain.Source) { s.ApprovalStatus = "paused" }},
		{"no terms", func(s *domain.Source) { s.TermsReviewedAt = nil }},
		{"zero terms", func(s *domain.Source) { s.TermsReviewedAt = &time.Time{} }},
		{"no reason", func(s *domain.Source) { s.ApprovalReason = nil }},
		{"empty reason", func(s *domain.Source) { reason := " "; s.ApprovalReason = &reason }},
		{"authenticated", func(s *domain.Source) { s.AuthenticationRequired = true }},
		{"unsupported", func(s *domain.Source) { s.RetrievalMethod = "html" }},
		{"no scope", func(s *domain.Source) { s.AllowedPathPrefixes = nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := testServer(t)
			src := approvedRSSSource(t, s)
			// Recreate a separate source through the same persistence contract.
			src.SourceID = ""
			tc.change(src)
			if err := s.store.CreateSource(context.Background(), src); err != nil {
				t.Fatal(err)
			}
			s.fetchSource = func(context.Context, *domain.Source) (*sourceadapter.FetchResult, error) {
				t.Fatal("ineligible source fetched")
				return nil, nil
			}
			w := httptest.NewRecorder()
			s.handleCreateResearchRun(w, httptest.NewRequest("POST", "/", strings.NewReader(`{"source_ids":["`+src.SourceID+`"]}`)))
			if w.Code != 400 {
				t.Fatalf("ineligible POST status=%d", w.Code)
			}
			w = httptest.NewRecorder()
			s.handleListSources(w, httptest.NewRequest("GET", "/", nil))
			var payload struct {
				Sources []struct {
					ID      string `json:"source_id"`
					Allowed bool   `json:"manual_run_allowed"`
					Reason  string `json:"manual_run_reason"`
				} `json:"sources"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}
			for _, item := range payload.Sources {
				if item.ID == src.SourceID && (item.Allowed || item.Reason == "") {
					t.Fatal("list eligibility disagrees with POST")
				}
			}
		})
	}
}
