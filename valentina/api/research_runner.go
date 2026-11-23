package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/scrypster/huginn/internal/valentina/domain"
	"github.com/scrypster/huginn/internal/valentina/sourceadapter"
	"github.com/scrypster/huginn/internal/valentina/sourcepolicy"
)

// Manual research v1 intentionally supports only RSS/Atom feeds. Other adapter
// types cannot be marked executable until structured extraction is validated.
// This is one bounded, synchronous operator-triggered pass, not a scheduler.
func sourceRunnable(src *domain.Source) bool {
	return sourceRunReason(src) == ""
}

// Shared by the list response and POST authorization; UI flags are never trusted.
func sourceRunReason(src *domain.Source) string {
	if src == nil || !src.CanFetch() {
		return "Источник ожидает утверждения, заблокирован или приостановлен"
	}
	if src.TermsReviewedAt == nil || src.TermsReviewedAt.IsZero() {
		return "Условия источника не проверены"
	}
	if src.ApprovalReason == nil || strings.TrimSpace(*src.ApprovalReason) == "" {
		return "Не указано основание утверждения"
	}
	if src.AuthenticationRequired {
		return "Источники с авторизацией пока не поддерживаются"
	}
	if len(src.AllowedHosts) == 0 || len(src.AllowedPathPrefixes) == 0 {
		return "Не задана разрешённая область источника"
	}
	method := strings.ToLower(strings.TrimSpace(src.RetrievalMethod))
	if method != "rss" && method != "atom" {
		return "Ручной запуск поддерживает только RSS/Atom"
	}
	u, err := url.Parse(src.BaseURL)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil {
		return "Нужен HTTPS-адрес ленты без учётных данных"
	}
	return ""
}

// fetchApprovedFeed creates a narrow policy gateway bound to the approved
// database source. It never uses http.DefaultClient or generic fetch_url.
func fetchApprovedFeed(ctx context.Context, src *domain.Source) (*sourceadapter.FetchResult, error) {
	if !sourceRunnable(src) {
		return nil, errors.New("source has not passed manual feed approval")
	}
	parsed, err := url.Parse(src.BaseURL)
	if err != nil || parsed.Scheme != "https" || parsed.User != nil {
		return nil, errors.New("source requires an HTTPS feed URL")
	}
	profile := &sourcepolicy.SourceProfile{
		SourceID:               src.SourceID,
		Name:                   src.Name,
		Publisher:              src.Publisher,
		BaseURL:                src.BaseURL,
		SourceType:             sourcepolicy.SourceType(src.RetrievalMethod),
		ApprovalStatus:         sourcepolicy.ApprovalApproved,
		ApprovalReason:         *src.ApprovalReason,
		TermsReviewedAt:        src.TermsReviewedAt,
		AllowedHosts:           src.AllowedHosts,
		AllowedPathPrefixes:    src.AllowedPathPrefixes,
		AllowedSchemes:         []string{"https"},
		MinimumIntervalSeconds: src.RateLimits.MinimumIntervalSeconds,
		MaxParallelRequests:    src.RateLimits.MaxParallelRequests,
		DailyRequestLimit:      src.RateLimits.DailyRequestLimit,
		TimeoutSeconds:         src.RateLimits.TimeoutSeconds,
		MaxResponseBytes:       int(src.RateLimits.MaxResponseBytes),
	}
	registry := sourcepolicy.NewRegistry()
	if err := registry.Register(profile); err != nil {
		return nil, errors.New("source policy configuration invalid")
	}
	gateway := sourcepolicy.NewPolicyGateway(registry)
	if decision := gateway.Evaluate(src.SourceID, src.BaseURL); !decision.Allowed {
		return nil, errors.New("source URL rejected by policy gateway")
	}
	client, err := sourcepolicy.NewValidatedHTTPClient(gateway, src.SourceID, profile)
	if err != nil {
		return nil, errors.New("source HTTP policy initialization failed")
	}
	timeout := time.Duration(profile.TimeoutSeconds) * time.Second
	adapter := sourceadapter.NewRSSAdapter(src.Name, client, timeout)
	result, err := adapter.Fetch(ctx, src.SourceID, src.BaseURL)
	if err != nil {
		return nil, errors.New("approved source feed fetch failed")
	}
	return result, nil
}

// checkResearchBudget uses durable run history to enforce a conservative per-
// source run budget even when a new policy HTTP client is built each time.
// Every manual run makes at most one request per source.
func (s *Server) checkResearchBudget(ctx context.Context, sources []*domain.Source) error {
	previous, err := s.store.ListResearchRuns(ctx)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, src := range sources {
		countToday := 0
		var last time.Time
		for _, r := range previous {
			used := false
			for _, id := range r.SourceIDs {
				if id == src.SourceID {
					used = true
					break
				}
			}
			if !used {
				continue
			}
			if r.StartedAt.UTC().Format("2006-01-02") == now.Format("2006-01-02") {
				countToday++
			}
			if r.StartedAt.After(last) {
				last = r.StartedAt
			}
		}
		if src.RateLimits.DailyRequestLimit != nil &&
			countToday >= *src.RateLimits.DailyRequestLimit {
			return fmt.Errorf("daily research budget exhausted for source %s", src.SourceID)
		}
		if src.RateLimits.MinimumIntervalSeconds > 0 &&
			!last.IsZero() && now.Sub(last) < time.Duration(src.RateLimits.MinimumIntervalSeconds)*time.Second {
			return fmt.Errorf("minimum research interval not reached for source %s", src.SourceID)
		}
	}
	return nil
}

// executeManualResearch stores only source-backed, unverified candidates.
// A feed entry is not proof that the linked article was retrieved or checked.
func (s *Server) executeManualResearch(ctx context.Context, run *domain.ResearchRun, sources []*domain.Source) (string, error) {
	candidateCount, duplicateCount, successCount := 0, 0, 0
	failures := make([]string, 0)
	for _, src := range sources {
		if ctx.Err() != nil {
			break
		}
		fetch := s.fetchSource
		if fetch == nil {
			fetch = fetchApprovedFeed
		}
		fetched, err := fetch(ctx, src)
		if err != nil || fetched == nil || fetched.SourceID != src.SourceID ||
			fetched.StatusCode < 200 || fetched.StatusCode >= 300 {
			failures = append(failures, src.SourceID+": fetch failed")
			continue
		}
		feedURL := fetched.CanonicalURL
		if feedURL == "" {
			feedURL = src.BaseURL
		}
		entries, err := sourceadapter.ExtractFeedCandidates(fetched.RawFeed, feedURL)
		if err != nil {
			failures = append(failures, src.SourceID+": feed parse failed")
			continue
		}
		successCount++
		for _, entry := range entries {
			if ctx.Err() != nil {
				break
			}
			if !sourceadapter.RelevantRenovationCandidate(entry) {
				continue
			}
			candidateCount++
			// External article links are observations only, not permitted fetch targets.
			// Their provenance remains the original approved feed URL.
			duplicate, err := s.store.HasSourceEvidence(ctx, src.SourceID, entry.URL)
			if err != nil {
				failures = append(failures, src.SourceID+": storage lookup failed")
				break
			}
			if duplicate {
				duplicateCount++
				continue
			}
			title, summary := entry.Title, entry.Summary
			if summary == "" {
				summary = title
			}
			digest := sha256.Sum256([]byte(title + "\n" + summary))
			hash := hex.EncodeToString(digest[:])
			extractor := "rss_atom_feed_v1"
			locator := src.BaseURL
			note := "Discovered in approved RSS/Atom feed only; linked article has not been verified"
			evidence := &domain.SourceEvidence{
				SourceID:        src.SourceID,
				CanonicalURL:    entry.URL,
				Publisher:       src.Publisher,
				PageTitle:       &title,
				ObservedAt:      fetched.FetchedAt,
				PublishedAt:     entry.PublishedAt,
				ContentHash:     &hash,
				RetrievalStatus: "feed_observed",
				ExtractorID:     &extractor,
				Facts: []domain.EvidenceFact{
					{FieldPath: "title", Value: title, ExtractionMethod: "rss_atom_xml", EvidenceLocator: &locator},
					{FieldPath: "summary", Value: summary, ExtractionMethod: "rss_atom_xml", EvidenceLocator: &locator},
					{FieldPath: "linked_url", Value: entry.URL, ExtractionMethod: "rss_atom_xml", EvidenceLocator: &locator},
				},
			}
			if err := s.store.CreateSourceEvidence(ctx, evidence); err != nil {
				failures = append(failures, src.SourceID+": evidence persistence failed")
				break
			}
			opportunity := &domain.Opportunity{
				Title: title, Summary: summary, OpportunityType: "unverified_feed_notice",
				Status:             domain.OpportunityStatusCandidate,
				PublishedAt:        entry.PublishedAt,
				DemandEvidenceType: domain.DemandEvidenceUnknown,
				DemandEvidenceNote: &note,
				SourceEvidenceIDs:  []string{evidence.EvidenceID},
			}
			if err := s.store.CreateOpportunity(ctx, opportunity); err != nil {
				failures = append(failures, src.SourceID+": candidate persistence failed")
				break
			}
		}
	}
	run.CandidateCount = candidateCount
	run.DuplicateCount = duplicateCount
	run.VerifiedCount = 0 // a feed observation is not a verified public demand request
	run.AcceptedCount = 0 // human approval is always separate
	run.RejectedCount = 0
	note := fmt.Sprintf("RSS/Atom manual pass: %d candidate items, %d previously observed; %d/%d feeds fetched", candidateCount, duplicateCount, successCount, len(sources))
	if len(failures) > 0 {
		note += "; issues: " + strings.Join(failures, "; ")
	}
	if ctx.Err() != nil {
		return note, errors.New("research request cancelled")
	}
	if successCount == 0 {
		return note, errors.New("no sources fetched successfully")
	}
	return note, nil
}
