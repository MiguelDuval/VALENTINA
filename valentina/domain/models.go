package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

// DemandEvidenceType classifies the nature of public demand evidence.
type DemandEvidenceType string

const (
	DemandEvidenceDirectRequest     DemandEvidenceType = "direct_request"
	DemandEvidencePublicTender      DemandEvidenceType = "public_tender"
	DemandEvidenceSubcontracting    DemandEvidenceType = "subcontracting_request"
	DemandEvidenceEarlyDemandSignal DemandEvidenceType = "early_demand_signal"
	DemandEvidenceUnknown           DemandEvidenceType = "unknown"
)

// OpportunityStatus is the cold-opportunity lifecycle state.
type OpportunityStatus string

const (
	OpportunityStatusCandidate   OpportunityStatus = "candidate"
	OpportunityStatusVerified    OpportunityStatus = "verified"
	OpportunityStatusNeedsReview OpportunityStatus = "needs_review"
	OpportunityStatusQualified   OpportunityStatus = "qualified"
	OpportunityStatusRejected    OpportunityStatus = "rejected"
	OpportunityStatusExpired     OpportunityStatus = "expired"
	OpportunityStatusArchived    OpportunityStatus = "archived"
)

// InboundLeadStatus is the hot-inbound lifecycle state.
type InboundLeadStatus string

const (
	InboundLeadStatusNew      InboundLeadStatus = "new"
	InboundLeadStatusTriage   InboundLeadStatus = "triage"
	InboundLeadStatusNeedsInfo InboundLeadStatus = "needs_information"
	InboundLeadStatusQualified InboundLeadStatus = "qualified"
	InboundLeadStatusConverted InboundLeadStatus = "converted"
	InboundLeadStatusDeclined InboundLeadStatus = "declined"
	InboundLeadStatusSpam     InboundLeadStatus = "spam"
	InboundLeadStatusArchived InboundLeadStatus = "archived"
)

// Location represents a geographic location.
type Location struct {
	Municipality         *string `json:"municipality,omitempty"`
	Province             *string `json:"province,omitempty"`
	AutonomousCommunity  *string `json:"autonomous_community,omitempty"`
	CountryCode          *string `json:"country_code,omitempty"`
	LocationText         *string `json:"location_text,omitempty"`
}

// OpportunityValue represents a published value/budget/rate.
type OpportunityValue struct {
	Amount   *float64 `json:"amount,omitempty"`
	Currency *string   `json:"currency,omitempty"`
	Basis    *string   `json:"basis,omitempty"`
}

// FitCriterion is a single criterion evaluation.
type FitCriterion struct {
	CriterionID  string   `json:"criterion_id"`
	Outcome      string   `json:"outcome"`
	Reason       string   `json:"reason"`
	EvidenceIDs  []string `json:"evidence_ids,omitempty"`
}

// FitAssessment records an owner-profile fit evaluation.
type FitAssessment struct {
	ProfileVersion     string        `json:"profile_version"`
	FitStatus          string        `json:"fit_status"`
	EvaluatedAt        time.Time     `json:"evaluated_at"`
	Criteria           []FitCriterion `json:"criteria"`
	MissingInformation []string      `json:"missing_information,omitempty"`
}

// EvidenceFact is a single extracted fact with provenance.
type EvidenceFact struct {
	FieldPath        string      `json:"field_path"`
	Value            interface{} `json:"value"`
	ExtractionMethod string      `json:"extraction_method"`
	EvidenceExcerpt  *string     `json:"evidence_excerpt,omitempty"`
	EvidenceLocator  *string     `json:"evidence_locator,omitempty"`
	Confidence       *float64    `json:"confidence,omitempty"`
}

// SourceRateLimits defines request constraints for a source.
type SourceRateLimits struct {
	MinimumIntervalSeconds int    `json:"minimum_interval_seconds"`
	MaxParallelRequests    int    `json:"max_parallel_requests"`
	DailyRequestLimit      *int   `json:"daily_request_limit,omitempty"`
	TimeoutSeconds         int    `json:"timeout_seconds"`
	MaxResponseBytes       int64  `json:"max_response_bytes"`
}

// ContactInfo is minimal contact information for an inbound customer.
type ContactInfo struct {
	Name    *string `json:"name,omitempty"`
	Email   *string `json:"email,omitempty"`
	Phone   *string `json:"phone,omitempty"`
	Purpose *string `json:"purpose,omitempty"`
}

// InboundAIAssessment is AI-assisted extraction/summary for an inbound lead.
type InboundAIAssessment struct {
	ServiceCategory *string   `json:"service_category,omitempty"`
	ProjectScope    *string   `json:"project_scope,omitempty"`
	Location        *string   `json:"location,omitempty"`
	Budget          *float64  `json:"budget,omitempty"`
	TimingUrgency   *string   `json:"timing_urgency,omitempty"`
	MissingDetails  []string  `json:"missing_details,omitempty"`
	CompanyMatch    *string   `json:"company_match,omitempty"`
	ResponseDraft   *string   `json:"response_draft,omitempty"`
	EvaluatedAt     time.Time `json:"evaluated_at"`
	ProfileVersion  string    `json:"profile_version,omitempty"`
}

// Company represents a public-facing organization or publisher.
type Company struct {
	CompanyID         string     `json:"company_id"`
	Name              string     `json:"name"`
	NormalizedName    string     `json:"normalized_name"`
	Website           *string    `json:"website,omitempty"`
	Domain            *string    `json:"domain,omitempty"`
	IndustryTags      []string   `json:"industry_tags,omitempty"`
	Location          *Location  `json:"location,omitempty"`
	SourceEvidenceIDs []string   `json:"source_evidence_ids"`
	FirstSeenAt       time.Time  `json:"first_seen_at"`
	LastSeenAt        time.Time  `json:"last_seen_at"`
	Notes             *string    `json:"notes,omitempty"`
}

// Opportunity represents a publicly discovered commercial opportunity.
type Opportunity struct {
	OpportunityID       string             `json:"opportunity_id"`
	CompanyID           *string            `json:"company_id,omitempty"`
	Title               string             `json:"title"`
	Summary             string             `json:"summary"`
	OpportunityType     string             `json:"opportunity_type"`
	Status              OpportunityStatus  `json:"status"`
	Geography           *Location          `json:"geography,omitempty"`
	PublishedAt         *time.Time         `json:"published_at,omitempty"`
	DeadlineAt          *time.Time         `json:"deadline_at,omitempty"`
	ValidFrom           *time.Time         `json:"valid_from,omitempty"`
	ValidUntil          *time.Time         `json:"valid_until,omitempty"`
	Value               *OpportunityValue  `json:"value,omitempty"`
	FitAssessment       *FitAssessment     `json:"fit_assessment,omitempty"`
	SourceEvidenceIDs   []string           `json:"source_evidence_ids"`
	FirstSeenAt         time.Time          `json:"first_seen_at"`
	LastSeenAt          time.Time          `json:"last_seen_at"`
	DuplicateGroupID    *string            `json:"duplicate_group_id,omitempty"`
	ReviewNote          *string            `json:"review_note,omitempty"`
	DemandEvidenceType  DemandEvidenceType `json:"demand_evidence_type"`
	DemandEvidenceNote  *string            `json:"demand_evidence_note,omitempty"`
}

// SourceEvidence records one observation of one page/document.
type SourceEvidence struct {
	EvidenceID       string         `json:"evidence_id"`
	SourceID         string         `json:"source_id"`
	CanonicalURL     string         `json:"canonical_url"`
	Publisher        string         `json:"publisher"`
	PageTitle        *string        `json:"page_title,omitempty"`
	ObservedAt       time.Time      `json:"observed_at"`
	PublishedAt      *time.Time     `json:"published_at,omitempty"`
	ContentHash      *string        `json:"content_hash,omitempty"`
	RetrievalStatus  string         `json:"retrieval_status"`
	ExtractorID      *string        `json:"extractor_id,omitempty"`
	ExtractorVersion *string        `json:"extractor_version,omitempty"`
	Facts            []EvidenceFact `json:"facts"`
}

// Source represents an approved public-source collection profile.
type Source struct {
	SourceID           string           `json:"source_id"`
	Name               string           `json:"name"`
	Publisher          string           `json:"publisher"`
	SourceType         string           `json:"source_type"`
	BaseURL            string           `json:"base_url"`
	AllowedHosts       []string         `json:"allowed_hosts"`
	AllowedPathPrefixes []string         `json:"allowed_path_prefixes,omitempty"`
	TermsURL           *string          `json:"terms_url,omitempty"`
	TermsReviewedAt    *time.Time       `json:"terms_reviewed_at,omitempty"`
	RobotsReviewedAt   *time.Time       `json:"robots_reviewed_at,omitempty"`
	ApprovalStatus     string           `json:"approval_status"`
	ApprovalReason     *string          `json:"approval_reason,omitempty"`
	RetrievalMethod    string           `json:"retrieval_method"`
	AdapterID          *string          `json:"adapter_id,omitempty"`
	AdapterVersion     *string          `json:"adapter_version,omitempty"`
	AuthenticationRequired bool          `json:"authentication_required"`
	CredentialReference *string         `json:"credential_reference,omitempty"`
	RateLimits         SourceRateLimits `json:"rate_limits"`
	AllowedFields      []string         `json:"allowed_fields"`
	ProhibitedFields   []string         `json:"prohibited_fields,omitempty"`
	Languages          []string         `json:"languages,omitempty"`
	ExpectedUpdateFrequency *string       `json:"expected_update_frequency,omitempty"`
	LastReviewedAt     *time.Time       `json:"last_reviewed_at,omitempty"`
	Notes              *string          `json:"notes,omitempty"`
}

// ResearchRun records a manual research action and its results.
type ResearchRun struct {
	RunID          string     `json:"run_id"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	Status         string     `json:"status"`
	SourceIDs      []string   `json:"source_ids,omitempty"`
	CandidateCount int        `json:"candidate_count"`
	VerifiedCount  int        `json:"verified_count"`
	DuplicateCount int        `json:"duplicate_count"`
	AcceptedCount  int        `json:"accepted_count"`
	RejectedCount  int        `json:"rejected_count"`
	Notes          *string    `json:"notes,omitempty"`
}

// ReviewDecision records a human qualification/rejection/conversion decision.
type ReviewDecision struct {
	DecisionID     string    `json:"decision_id"`
	EntityType     string    `json:"entity_type"`
	EntityID       string    `json:"entity_id"`
	PreviousStatus string    `json:"previous_status,omitempty"`
	NewStatus      string    `json:"new_status"`
	DecisionTime   time.Time `json:"decision_time"`
	DecisionReason string    `json:"decision_reason"`
	ProfileVersion string    `json:"profile_version,omitempty"`
	OperatorRef    *string   `json:"operator_ref,omitempty"`
}

// InboundLead represents a customer-initiated enquiry.
type InboundLead struct {
	LeadID            string              `json:"lead_id"`
	ReceivedAt        time.Time           `json:"received_at"`
	InboundChannel    string              `json:"inbound_channel"`
	ChannelMessageRef *string             `json:"channel_message_ref,omitempty"`
	RequestSummary    string              `json:"request_summary"`
	OriginalContent   *string             `json:"original_content,omitempty"`
	RequestedServices []string            `json:"requested_services,omitempty"`
	ProjectLocation   *Location           `json:"project_location,omitempty"`
	Budget            *float64            `json:"budget,omitempty"`
	BudgetCurrency    *string             `json:"budget_currency,omitempty"`
	DesiredTiming     *string             `json:"desired_timing,omitempty"`
	CompanyID         *string             `json:"company_id,omitempty"`
	ContactInfo       *ContactInfo        `json:"contact_info,omitempty"`
	AIAssessment      *InboundAIAssessment `json:"ai_assessment,omitempty"`
	MissingInformation []string            `json:"missing_information,omitempty"`
	DuplicateGroupID  *string             `json:"duplicate_group_id,omitempty"`
	Status            InboundLeadStatus   `json:"status"`
	DecisionHistory   []ReviewDecision    `json:"decision_history,omitempty"`
	PrivacyNotice     *bool               `json:"privacy_notice,omitempty"`
	RetentionState    string              `json:"retention_state"`
}

// NewOpportunityID generates a stable opportunity ID.
func NewOpportunityID() string { return uuid.New().String() }

// NewCompanyID generates a stable company ID.
func NewCompanyID() string { return uuid.New().String() }

// NewLeadID generates a stable inbound lead ID.
func NewLeadID() string { return uuid.New().String() }

// NewEvidenceID generates a stable evidence ID.
func NewEvidenceID() string { return ulid.Make().String() }

// NewRunID generates a stable research run ID.
func NewRunID() string { return ulid.Make().String() }

// NewSourceID generates a stable source ID.
func NewSourceID() string { return ulid.Make().String() }

// NewDecisionID generates a stable review decision ID.
func NewDecisionID() string { return uuid.New().String() }

// EnsureID assigns a generated ID when the caller left it empty. Must be called
// before Validate so that a missing ID never fails validation.
func (o *Opportunity) EnsureID() {
	if o.OpportunityID == "" {
		o.OpportunityID = NewOpportunityID()
	}
}

// EnsureID assigns a generated ID when the caller left it empty. Must be called
// before Validate so that a missing ID never fails validation.
func (l *InboundLead) EnsureID() {
	if l.LeadID == "" {
		l.LeadID = NewLeadID()
	}
}

// EnsureID assigns a generated ID when the caller left it empty.
func (c *Company) EnsureID() {
	if c.CompanyID == "" {
		c.CompanyID = NewCompanyID()
	}
}

// EnsureID assigns a generated ID when the caller left it empty.
func (s *Source) EnsureID() {
	if s.SourceID == "" {
		s.SourceID = NewSourceID()
	}

// NormalizeCompanyName produces a deterministic normalized company name.
func NormalizeCompanyName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// Validate checks that an Opportunity satisfies core business rules.
func (o *Opportunity) Validate() error {
	if o.OpportunityID == "" {
		return errors.New("opportunity_id is required")
	}
	if o.Title == "" {
		return errors.New("title is required")
	}
	if o.Summary == "" {
		return errors.New("summary is required")
	}
	if o.OpportunityType == "" {
		return errors.New("opportunity_type is required")
	}
	switch o.DemandEvidenceType {
	case DemandEvidenceDirectRequest, DemandEvidencePublicTender, DemandEvidenceSubcontracting, DemandEvidenceEarlyDemandSignal, DemandEvidenceUnknown:
	default:
		return fmt.Errorf("invalid demand_evidence_type: %q", o.DemandEvidenceType)
	}
	switch o.Status {
	case OpportunityStatusCandidate, OpportunityStatusVerified, OpportunityStatusNeedsReview, OpportunityStatusQualified, OpportunityStatusRejected, OpportunityStatusExpired, OpportunityStatusArchived:
	default:
		return fmt.Errorf("invalid status: %q", o.Status)
	}
	if len(o.SourceEvidenceIDs) == 0 {
		return errors.New("at least one source_evidence_id is required")
	}
	return nil
}

// Validate checks that an InboundLead satisfies core business rules.
func (l *InboundLead) Validate() error {
	if l.LeadID == "" {
		return errors.New("lead_id is required")
	}
	if l.RequestSummary == "" {
		return errors.New("request_summary is required")
	}
	if l.InboundChannel == "" {
		return errors.New("inbound_channel is required")
	}
	switch l.Status {
	case InboundLeadStatusNew, InboundLeadStatusTriage, InboundLeadStatusNeedsInfo, InboundLeadStatusQualified, InboundLeadStatusConverted, InboundLeadStatusDeclined, InboundLeadStatusSpam, InboundLeadStatusArchived:
	default:
		return fmt.Errorf("invalid status: %q", l.Status)
	}
	return nil
}

// CanonicalURL returns the opportunity's canonical URL for deduplication.
// Opportunities reference SourceEvidence records for their source URL,
// so we return an empty string here - dedup uses SourceEvidence canonical URLs.
func (o *Opportunity) CanonicalURL() string {
	return ""
}

// Publisher returns the publisher name from the first source evidence.
// Opportunities don't store publisher directly; it comes from SourceEvidence.
func (o *Opportunity) Publisher() string {
	return ""
}

// DateWindow returns the publication/deadline window string.
func (o *Opportunity) DateWindow() string {
	var parts []string
	if o.PublishedAt != nil {
		parts = append(parts, o.PublishedAt.Format("2006-01-02"))
	}
	if o.DeadlineAt != nil {
		parts = append(parts, o.DeadlineAt.Format("2006-01-02"))
	}
	return strings.Join(parts, "|")
}

// IsDuplicate checks whether two opportunities are likely duplicates based on
// deterministic matching: exact canonical URL, or exact title+publisher+date window.
func IsDuplicate(a, b *Opportunity) bool {
	if a.CanonicalURL() != "" && a.CanonicalURL() == b.CanonicalURL() {
		return true
	}
	if NormalizeCompanyName(a.Publisher()) == NormalizeCompanyName(b.Publisher()) &&
		a.Title == b.Title &&
		a.DateWindow() == b.DateWindow() &&
		a.DateWindow() != "" {
		return true
	}
	return false
}

// IsDuplicateLead checks whether two inbound leads are likely duplicates based
// on channel + message reference + timestamp window.
func IsDuplicateLead(a, b *InboundLead) bool {
	if a.ChannelMessageRef != nil && b.ChannelMessageRef != nil && *a.ChannelMessageRef == *b.ChannelMessageRef {
		return true
	}
	if a.ReceivedAt.Equal(b.ReceivedAt) &&
		a.InboundChannel == b.InboundChannel &&
		a.RequestSummary == b.RequestSummary {
		return true
	}
	return false
}

// ValidateDuplicateCandidate checks that an opportunity doesn't conflict with
// an existing duplicate group.
func ValidateDuplicateCandidate(newOpp *Opportunity, existing []*Opportunity) error {
	for _, existing := range existing {
		if IsDuplicate(newOpp, existing) {
			return fmt.Errorf("potential duplicate of opportunity %s (duplicate group: %s)",
				existing.OpportunityID, existing.DuplicateGroupID)
		}
	}
	return nil
}

// fmtDate formats a time.Time as YYYY-MM-DD, or empty string if zero.
func fmtDate(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

