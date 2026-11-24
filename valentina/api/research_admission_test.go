package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/scrypster/huginn/internal/valentina/domain"
	"github.com/scrypster/huginn/internal/valentina/sourceadapter"
)

func researchHandler(s *Server) http.Handler {
	mux := http.NewServeMux()
	for pattern, handler := range s.Routes() {
		mux.HandleFunc(pattern, handler)
	}
	return mux
}

func researchPost(s *Server, ctx context.Context, ids ...string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]any{"source_ids": ids})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/valentina/research-runs", strings.NewReader(string(body))).WithContext(ctx)
	w := httptest.NewRecorder()
	researchHandler(s).ServeHTTP(w, r)
	return w
}

func TestResearchConcurrentAdmissionAndCancellation(t *testing.T) {
	s := testServer(t)
	src := approvedRSSSource(t, s)
	entered, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	s.fetchSource = func(ctx context.Context, src *domain.Source) (*sourceadapter.FetchResult, error) {
		if calls.Add(1) == 1 {
			close(entered)
		}
		select {
		case <-release:
		case <-ctx.Done():
		}
		return nil, errors.New("synthetic failure")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() { done <- researchPost(s, ctx, src.SourceID) }()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("fetch did not start")
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := researchPost(s, context.Background(), src.SourceID)
			if w.Code != http.StatusConflict || strings.Contains(w.Body.String(), src.SourceID) {
				t.Errorf("unsafe admission: %d %s", w.Code, w.Body)
			}
		}()
	}
	wg.Wait()
	// A second API facade on the same Store must not recover an active run.
	_ = NewServer(s.store)
	runs, err := s.store.ListResearchRuns(context.Background())
	if err != nil || len(runs) != 1 || runs[0].Status != "running" {
		t.Fatalf("active run changed: %v %+v", err, runs)
	}
	cancel()
	select {
	case w := <-done:
		if w.Code != http.StatusCreated {
			t.Fatal(w.Code, w.Body)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("cancellation not persisted")
	}
	close(release)
	runs, err = s.store.ListResearchRuns(context.Background())
	if err != nil || len(runs) != 1 || runs[0].Status != "failed" || runs[0].CompletedAt == nil || !strings.Contains(*runs[0].Notes, "interrupted") || calls.Load() != 1 {
		t.Fatalf("bad terminal run: %v %+v calls=%d", err, runs, calls.Load())
	}
	w := httptest.NewRecorder()
	researchHandler(s).ServeHTTP(w, httptest.NewRequest("POST", "/api/v1/valentina/research-runs/"+runs[0].RunID+"/cancel", nil))
	if w.Code != http.StatusConflict {
		t.Fatalf("terminal cancel: %d", w.Code)
	}
}

func TestResearchOperatorCancelKeepsReservationUntilStopped(t *testing.T) {
	s := testServer(t)
	src := approvedRSSSource(t, s)
	entered, stopping, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	s.fetchSource = func(ctx context.Context, _ *domain.Source) (*sourceadapter.FetchResult, error) {
		close(entered)
		<-ctx.Done()
		close(stopping)
		<-release
		return nil, ctx.Err()
	}
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() { done <- researchPost(s, context.Background(), src.SourceID) }()
	<-entered
	runs, _ := s.store.ListResearchRuns(context.Background())
	w := httptest.NewRecorder()
	researchHandler(s).ServeHTTP(w, httptest.NewRequest("POST", "/api/v1/valentina/research-runs/"+runs[0].RunID+"/cancel", nil))
	if w.Code != http.StatusAccepted || !strings.Contains(w.Body.String(), "cancellation_requested") {
		t.Fatal(w.Code, w.Body)
	}
	<-stopping
	if w := researchPost(s, context.Background(), src.SourceID); w.Code != http.StatusConflict {
		t.Fatal("source released before fetch stopped")
	}
	close(release)
	if w := <-done; w.Code != http.StatusCreated {
		t.Fatal(w.Code, w.Body)
	}
}

func TestResearchAtomicSelectionAndDuplicateIDs(t *testing.T) {
	s := testServer(t)
	a, b := approvedRSSSource(t, s), approvedRSSSource(t, s)
	ctx := context.Background()
	active, err := s.store.AdmitResearchRun(ctx, []string{b.SourceID}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, ids := range [][]string{{a.SourceID, b.SourceID}, {b.SourceID, a.SourceID}} {
		if w := researchPost(s, ctx, ids...); w.Code != http.StatusConflict {
			t.Fatal(w.Code, w.Body)
		}
	}
	if w := researchPost(s, ctx, a.SourceID, a.SourceID); w.Code != http.StatusBadRequest {
		t.Fatal(w.Code, w.Body)
	}
	runs, _ := s.store.ListResearchRuns(ctx)
	if len(runs) != 1 || runs[0].RunID != active.RunID {
		t.Fatal("partial reservation persisted")
	}
	if _, err := s.store.AdmitResearchRun(ctx, []string{a.SourceID}, time.Now()); err != nil {
		t.Fatal("unrelated source consumed", err)
	}
}

func TestResearchRecoveryQuotaAndUTCMidnight(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "research.db")
	store, err := domain.NewStore(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	s := NewServer(store)
	src := approvedRSSSource(t, s)
	// Recreate a separate fixture with explicit persisted limits.
	src.SourceID = ""
	limit := 1
	src.RateLimits.DailyRequestLimit = &limit
	src.RateLimits.MinimumIntervalSeconds = 60
	if err := store.CreateSource(ctx, src); err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 11, 24, 23, 59, 30, 0, time.UTC)
	run, err := store.AdmitResearchRun(ctx, []string{src.SourceID}, at)
	if err != nil {
		t.Fatal(err)
	}
	evidence := &domain.SourceEvidence{SourceID: src.SourceID, CanonicalURL: "https://example.com/synthetic", Publisher: "Synthetic", ObservedAt: at, RetrievalStatus: "feed_observed"}
	if err := store.CreateSourceEvidence(ctx, evidence); err != nil {
		t.Fatal(err)
	}
	historic := &domain.ResearchRun{Status: "running", SourceIDs: []string{"synthetic-historic"}, CandidateCount: 4, DuplicateCount: 2, StartedAt: at}
	if err := store.CreateResearchRun(ctx, historic); err != nil {
		t.Fatal(err)
	}
	finished := &domain.ResearchRun{Status: "completed", SourceIDs: []string{"synthetic-finished"}, CandidateCount: 7, StartedAt: at, CompletedAt: &at}
	if err := store.CreateResearchRun(ctx, finished); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	store, err = domain.NewStore(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	recovered, err := store.GetResearchRun(ctx, run.RunID)
	if err != nil || recovered.Status != "failed" || recovered.CompletedAt == nil || recovered.StartedAt != at {
		t.Fatalf("recovery: %+v %v", recovered, err)
	}
	history, err := store.GetResearchRun(ctx, historic.RunID)
	if err != nil || history.Status != "failed" || history.CandidateCount != 4 || history.DuplicateCount != 2 {
		t.Fatal("recovery lost counts", err, history)
	}
	terminal, err := store.GetResearchRun(ctx, finished.RunID)
	if err != nil || terminal.Status != "completed" || terminal.CandidateCount != 7 || !terminal.CompletedAt.Equal(at) {
		t.Fatal("recovery mutated terminal run", err, terminal)
	}
	saved, err := store.ListSourceEvidenceByIDs(ctx, []string{evidence.EvidenceID})
	if err != nil || len(saved) != 1 {
		t.Fatal("recovery lost evidence", err)
	}
	for _, now := range []time.Time{at.Add(10 * time.Second), at.Add(30 * time.Second)} {
		if _, err := store.AdmitResearchRun(ctx, []string{src.SourceID}, now); !errors.Is(err, domain.ErrResearchQuota) {
			t.Fatal("quota/interval lost across reopen", err)
		}
	}
	// 00:00:30 UTC is still the previous local day in this offset.
	next := at.Add(time.Minute).In(time.FixedZone("synthetic west", -5*3600))
	if _, err := store.AdmitResearchRun(ctx, []string{src.SourceID}, next); err != nil {
		t.Fatal("UTC reset/exact interval", err)
	}
}

func TestResearchInvalidLimitsFailClosed(t *testing.T) {
	for _, change := range []func(*domain.SourceRateLimits){
		func(l *domain.SourceRateLimits) { l.MinimumIntervalSeconds = -1 },
		func(l *domain.SourceRateLimits) { l.MaxParallelRequests = 0 },
		func(l *domain.SourceRateLimits) { n := 0; l.DailyRequestLimit = &n },
		func(l *domain.SourceRateLimits) { l.TimeoutSeconds = 0 },
		func(l *domain.SourceRateLimits) { l.MaxResponseBytes = 0 },
	} {
		s := testServer(t)
		src := approvedRSSSource(t, s)
		src.SourceID = ""
		change(&src.RateLimits)
		if err := s.store.CreateSource(context.Background(), src); err != nil {
			t.Fatal(err)
		}
		if _, err := s.store.AdmitResearchRun(context.Background(), []string{src.SourceID}, time.Now()); !errors.Is(err, domain.ErrResearchPolicy) {
			t.Fatal("invalid policy admitted", err)
		}
		if sourceRunnable(src) {
			t.Fatal("invalid policy shown as runnable")
		}
	}
}

func TestResearchQuotaHTTPAndAtomicRollback(t *testing.T) {
	s := testServer(t)
	s.researchNow = func() time.Time { return time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC) }
	a, b := approvedRSSSource(t, s), approvedRSSSource(t, s)
	b.SourceID = ""
	cap := 1
	b.RateLimits.DailyRequestLimit = &cap
	if err := s.store.CreateSource(context.Background(), b); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	s.fetchSource = func(context.Context, *domain.Source) (*sourceadapter.FetchResult, error) {
		calls.Add(1)
		return nil, errors.New("synthetic failure")
	}
	if w := researchPost(s, context.Background(), b.SourceID); w.Code != http.StatusCreated {
		t.Fatal(w.Code, w.Body)
	}
	for _, ids := range [][]string{{a.SourceID, b.SourceID}, {b.SourceID, a.SourceID}} {
		w := researchPost(s, context.Background(), ids...)
		if w.Code != http.StatusTooManyRequests || strings.Contains(w.Body.String(), b.SourceID) {
			t.Fatal(w.Code, w.Body)
		}
	}
	runs, err := s.store.ListResearchRuns(context.Background())
	if err != nil || len(runs) != 1 || calls.Load() != 1 {
		t.Fatal("failed admission consumed quota or fetched", err, len(runs), calls.Load())
	}
	if w := researchPost(s, context.Background(), a.SourceID); w.Code != http.StatusCreated {
		t.Fatal(w.Code, w.Body)
	}
}

func TestResearchStoreConcurrentAdmission(t *testing.T) {
	s := testServer(t)
	src := approvedRSSSource(t, s)
	var accepted, blocked atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := s.store.AdmitResearchRun(context.Background(), []string{src.SourceID}, time.Now())
			if err == nil {
				accepted.Add(1)
			} else if errors.Is(err, domain.ErrResearchBusy) {
				blocked.Add(1)
			} else {
				t.Error(err)
			}
		}()
	}
	close(start)
	wg.Wait()
	if accepted.Load() != 1 || blocked.Load() != 11 {
		t.Fatal("transaction admitted overlapping runs", accepted.Load(), blocked.Load())
	}
}
