package sourcepolicy

import (
	"errors"
	"strings"
	"sync"
	"time"
)

// ApprovalStatus describes the review/approval state of a source profile.
type ApprovalStatus string

const (
	// ApprovalPending means the source has not been reviewed and must never be fetched.
	ApprovalPending ApprovalStatus = "pending_review"
	// ApprovalApproved means the source has been reviewed and may be fetched.
	ApprovalApproved ApprovalStatus = "approved"
	// ApprovalPaused means the source was approved but is temporarily suspended.
	ApprovalPaused ApprovalStatus = "paused"
	// ApprovalBlocked means the source is explicitly prohibited.
	ApprovalBlocked ApprovalStatus = "blocked"
)

// SourceType describes the retrieval method for a source.
type SourceType string

const (
	SourceTypeOfficialAPI SourceType = "official_api"
	SourceTypeRSS         SourceType = "rss"
	SourceTypeAtom        SourceType = "atom"
	SourceTypeSitemap     SourceType = "sitemap"
	SourceTypeHTML        SourceType = "html"
	SourceTypeDataset     SourceType = "dataset"
	SourceTypeDocument    SourceType = "document"
)

// SourceProfile is the registry entry for one approved source.
// It defines what may be fetched, how often, and under what conditions.
type SourceProfile struct {
	SourceID                string         `json:"source_id"`
	Name                    string         `json:"name"`
	Publisher               string         `json:"publisher"`
	SourceType              SourceType     `json:"source_type"`
	BaseURL                 string         `json:"base_url"`
	AllowedHosts            []string       `json:"allowed_hosts"`
	AllowedPathPrefixes     []string       `json:"allowed_path_prefixes"`
	AllowedSchemes          []string       `json:"allowed_schemes"`
	TermsURL                string         `json:"terms_url,omitempty"`
	TermsReviewedAt         *time.Time     `json:"terms_reviewed_at,omitempty"`
	RobotsReviewedAt        *time.Time     `json:"robots_reviewed_at,omitempty"`
	ApprovalStatus          ApprovalStatus `json:"approval_status"`
	ApprovalReason          string         `json:"approval_reason,omitempty"`
	RetrievalMethod         string         `json:"retrieval_method"`
	AdapterID               string         `json:"adapter_id,omitempty"`
	AdapterVersion          string         `json:"adapter_version,omitempty"`
	AuthenticationRequired  bool           `json:"authentication_required"`
	CredentialReference     string         `json:"credential_reference,omitempty"`
	MinimumIntervalSeconds  int            `json:"minimum_interval_seconds"`
	MaxParallelRequests     int            `json:"max_parallel_requests"`
	DailyRequestLimit       *int           `json:"daily_request_limit,omitempty"`
	TimeoutSeconds          int            `json:"timeout_seconds"`
	MaxResponseBytes        int            `json:"max_response_bytes"`
	AllowedFields           []string       `json:"allowed_fields"`
	ProhibitedFields        []string       `json:"prohibited_fields,omitempty"`
	Languages               []string       `json:"languages,omitempty"`
	ExpectedUpdateFrequency string         `json:"expected_update_frequency,omitempty"`
	LastReviewedAt          *time.Time     `json:"last_reviewed_at,omitempty"`
	Notes                   string         `json:"notes,omitempty"`
}

// IsApproved returns true only when the source is approved (not pending, paused, or blocked).
func (p *SourceProfile) IsApproved() bool {
	return p.ApprovalStatus == ApprovalApproved
}

// CanFetch requires explicit approval plus a recorded terms review and reason.
// Approval status alone is insufficient to authorize an external request.
func (p *SourceProfile) CanFetch() bool {
	return p != nil && p.ApprovalStatus == ApprovalApproved && p.TermsReviewedAt != nil && !p.TermsReviewedAt.IsZero() && strings.TrimSpace(p.ApprovalReason) != ""
}

// Registry holds the in-memory source profile registry.
// All access is guarded by a mutex.
type Registry struct {
	mu      sync.RWMutex
	sources map[string]*SourceProfile
}

// NewRegistry creates an empty source registry.
func NewRegistry() *Registry {
	return &Registry{
		sources: make(map[string]*SourceProfile),
	}
}

// Register inserts or updates a source profile. Returns an error if the profile is invalid.
func (r *Registry) Register(p *SourceProfile) error {
	if p == nil {
		return errors.New("source profile is required")
	}
	p = cloneProfile(p)
	if p.SourceID == "" {
		return errors.New("source_id is required")
	}
	if p.ApprovalStatus == "" {
		p.ApprovalStatus = ApprovalPending
	}
	if p.TimeoutSeconds <= 0 {
		p.TimeoutSeconds = 10
	}
	if p.MaxResponseBytes <= 0 {
		p.MaxResponseBytes = 512 * 1024
	}
	if p.MinimumIntervalSeconds < 0 {
		p.MinimumIntervalSeconds = 0
	}
	if p.MaxParallelRequests <= 0 {
		p.MaxParallelRequests = 1
	}
	if len(p.AllowedSchemes) == 0 {
		p.AllowedSchemes = []string{"https"}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sources[p.SourceID] = p
	return nil
}

// cloneProfile returns a deep copy of p so callers cannot mutate registry state.
func cloneProfile(p *SourceProfile) *SourceProfile {
	if p == nil {
		return nil
	}
	cp := *p
	cp.AllowedHosts = append([]string(nil), p.AllowedHosts...)
	cp.AllowedPathPrefixes = append([]string(nil), p.AllowedPathPrefixes...)
	cp.AllowedSchemes = append([]string(nil), p.AllowedSchemes...)
	cp.AllowedFields = append([]string(nil), p.AllowedFields...)
	cp.ProhibitedFields = append([]string(nil), p.ProhibitedFields...)
	cp.Languages = append([]string(nil), p.Languages...)
	if p.TermsReviewedAt != nil {
		t := *p.TermsReviewedAt
		cp.TermsReviewedAt = &t
	}
	if p.RobotsReviewedAt != nil {
		t := *p.RobotsReviewedAt
		cp.RobotsReviewedAt = &t
	}
	if p.LastReviewedAt != nil {
		t := *p.LastReviewedAt
		cp.LastReviewedAt = &t
	}
	if p.DailyRequestLimit != nil {
		l := *p.DailyRequestLimit
		cp.DailyRequestLimit = &l
	}
	return &cp
}

// Get returns a deep copy of the source profile by ID, or nil if not found.
// Callers cannot mutate registry state through the returned pointer.
func (r *Registry) Get(sourceID string) *SourceProfile {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return cloneProfile(r.sources[sourceID])
}

// List returns deep copies of all registered source profiles.
// Callers cannot mutate registry state through the returned slice.
func (r *Registry) List() []*SourceProfile {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*SourceProfile, 0, len(r.sources))
	for _, p := range r.sources {
		out = append(out, cloneProfile(p))
	}
	return out
}

// Remove deletes a source profile by ID.
func (r *Registry) Remove(sourceID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sources, sourceID)
}

// Count returns the number of registered sources.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.sources)
}

// ApprovedCount returns the number of approved sources.
func (r *Registry) ApprovedCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := 0
	for _, p := range r.sources {
		if p.CanFetch() {
			n++
		}
	}
	return n
}

// FetchableCount returns the number of sources that can currently be fetched.
func (r *Registry) FetchableCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := 0
	for _, p := range r.sources {
		if p.CanFetch() {
			n++
		}
	}
	return n
}
