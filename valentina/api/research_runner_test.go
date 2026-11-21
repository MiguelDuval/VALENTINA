package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/scrypster/huginn/internal/valentina/domain"
	"github.com/scrypster/huginn/internal/valentina/sourceadapter"
)

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
