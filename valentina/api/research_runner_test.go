package api

import (
	"context"
	"errors"
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

func TestManualRSSResearchPersistsCandidatesAndDeduplicates(t *testing.T) {
	s := testServer(t)
	src := approvedRSSSource(t, s)
	s.fetchSource = func(ctx context.Context, item *domain.Source) (*sourceadapter.FetchResult, error) {
		if item.SourceID != src.SourceID {
			t.Fatalf("wrong source")
		}
		return &sourceadapter.FetchResult{SourceID: item.SourceID, URL: item.BaseURL, StatusCode: 200,
			FetchedAt: time.Now().UTC(), RawFeed: []byte(syntheticRSS)}, nil
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
		evidence[0].CanonicalURL != "https://example.com/obras/1" {
		t.Fatalf("source provenance lost: %+v %v", evidence, err)
	}
	execute()
	opportunities, _ = s.store.ListOpportunities(context.Background(), "", 10, 0)
	runs, err := s.store.ListResearchRuns(context.Background())
	if err != nil || len(runs) != 2 || len(opportunities) != 1 {
		t.Fatalf("duplicate candidate persisted: runs=%d opportunities=%d err=%v", len(runs), len(opportunities), err)
	}
	if runs[0].DuplicateCount != 1 || runs[0].CompletedAt == nil || runs[0].Status != "completed" {
		t.Fatalf("run progress missing: %+v", runs[0])
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
