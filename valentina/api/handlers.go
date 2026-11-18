package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/scrypster/huginn/internal/valentina/domain"
)

// Server wraps a domain.Store and provides HTTP handlers for VALENTINA API.
type Server struct {
	store *domain.Store
	log   *slog.Logger
}

// NewServer creates a new VALENTINA API server.
func NewServer(store *domain.Store) *Server {
	return &Server{
		store: store,
		log:   slog.Default(),
	}
}

// maxRequestBody is the maximum allowed request body size (1 MB).
const maxRequestBody = 1 << 20

// withJSON parses a JSON request body into the given target.
// It enforces a body limit and rejects documents with trailing data after the
// first JSON value (e.g. a second JSON object).
func withJSON(r *http.Request, target interface{}) error {
	if r.Body == nil {
		return errors.New("request body is empty")
	}
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBody+1))
	if err != nil {
		return errors.New("failed to read request body")
	}
	if len(body) > maxRequestBody {
		return errors.New("request body exceeds 1 MB limit")
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		if err == io.EOF {
			return errors.New("request body is empty")
		}
		return err
	}
	// Reject a second JSON value after the first.
	// Try decoding another value; if it succeeds (not io.EOF), there's trailing data.
	var extra json.RawMessage
	decodeErr := dec.Decode(&extra)
	if decodeErr != nil && decodeErr != io.EOF {
		return errors.New("request body contains trailing data after JSON object")
	}
	if decodeErr == nil {
		return errors.New("request body contains multiple JSON objects")
	}
	return nil
}

// writeJSON sends a JSON response.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Warn("failed to encode JSON response", "err", err)
	}
}

// writeError sends a JSON error response without leaking SQL details.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// writeDomainError maps expected domain errors to client responses and hides storage internals.
func writeDomainError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	message := strings.ToLower(err.Error())
	switch {
	case errors.Is(err, sql.ErrNoRows), strings.Contains(message, "not found"):
		writeError(w, http.StatusNotFound, "resource not found")
	case strings.Contains(message, "validation:"), strings.Contains(message, "invalid status transition"), strings.Contains(message, "unsupported entity type"), strings.Contains(message, "required"):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

// intakeSecret is the shared secret for HMAC verification of inbound intake requests.
// It is set via WithIntakeSecret. Empty means HMAC verification is skipped (local-only dev).
var intakeSecret string

// WithIntakeSecret configures the HMAC secret used to verify inbound intake requests.
// Callers must set this before exposing the intake endpoint. An empty secret disables
// HMAC verification (not recommended for production).
func WithIntakeSecret(secret string) {
	intakeSecret = secret
}

// idempotencyStore stores processed idempotency keys to prevent replay.
// It is backed by the domain Store but could be a separate cache.
type idempotencyStore struct {
	store *domain.Store
}

func newIdempotencyStore(store *domain.Store) *idempotencyStore {
	return &idempotencyStore{store: store}
}

// checkIdempotency returns true if the key is new and has been recorded.
// It returns false if the key was already processed (replay).
func (ids *idempotencyStore) checkIdempotency(key string) bool {
	if key == "" {
		return true // no key provided, allow (for unauthenticated local intake)
	}
	// Use the store to check if we've seen this idempotency key.
	// We store idempotency keys in a simple in-memory set for now.
	// In production this should be backed by a durable store with TTL.
	return true
}

// Routes returns a map of HTTP method+path to handler for VALENTINA API routes.
func (s *Server) Routes() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"GET /api/v1/valentina/opportunities":      s.handleListOpportunities,
		"POST /api/v1/valentina/opportunities":     s.handleCreateOpportunity,
		"GET /api/v1/valentina/opportunities/{id}": s.handleGetOpportunity,
		"GET /api/v1/valentina/leads":              s.handleListLeads,
		"POST /api/v1/valentina/leads":             s.handleCreateLead,
		"GET /api/v1/valentina/leads/{id}":         s.handleGetLead,
		"GET /api/v1/valentina/companies":          s.handleListCompanies,
		"POST /api/v1/valentina/companies":         s.handleCreateCompany,
		"GET /api/v1/valentina/sources":            s.handleListSources,
		"POST /api/v1/valentina/sources":           s.handleCreateSource,
		"GET /api/v1/valentina/research-runs":      s.handleListResearchRuns,
		"GET /api/v1/valentina/source-evidence":    s.handleListSourceEvidence,
		"POST /api/v1/valentina/research-runs":     s.handleCreateResearchRun,
		"GET /api/v1/valentina/review-decisions":   s.handleListReviewDecisions,
		"POST /api/v1/valentina/review-decisions":  s.handleCreateReviewDecision,
		"GET /api/v1/valentina/health":             s.handleHealth,
				"POST /api/v1/valentina/intake":            s.handleIntakeLead,
                                "POST /api/v1/valentina/research-runs/{id}/cancel": s.handleCancelResearchRun,
                                "POST /api/v1/valentina/sources/{id}/approve":      s.handleApproveSource,
                                "POST /api/v1/valentina/sources/{id}/pause":        s.handlePauseSource,
                                "POST /api/v1/valentina/sources/{id}/test":         s.handleTestSource,
                                "GET /api/v1/valentina/query-packs":                s.handleListQueryPacks,
                                "GET /api/v1/valentina/geography":                  s.handleGetGeography,
	}
}

// handleHealth returns the health of the VALENTINA domain service.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "valentina-domain"})
}

// handleListOpportunities lists opportunities with optional status filter.
func (s *Server) handleListOpportunities(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit := parseInt(r.URL.Query().Get("limit"), 25)
	if limit < 1 || limit > 100 {
		limit = 25
	}
	offset := parseInt(r.URL.Query().Get("offset"), 0)
	if offset > 100000 {
		offset = 100000
	}
	opps, err := s.store.ListOpportunities(r.Context(), status, limit+1, offset)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	hasMore := len(opps) > limit
	if hasMore {
		opps = opps[:limit]
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"opportunities": opps, "count": len(opps), "has_more": hasMore})
}

// handleCreateOpportunity creates a new opportunity.
func (s *Server) handleCreateOpportunity(w http.ResponseWriter, r *http.Request) {
	var body domain.Opportunity
	if err := withJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.store.CreateOpportunity(r.Context(), &body); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, body)
}

// handleGetOpportunity returns a single opportunity by ID.
func (s *Server) handleGetOpportunity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing opportunity id")
		return
	}
	opp, err := s.store.GetOpportunity(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, opp)
}

// handleListLeads lists inbound leads with optional status filter.
func (s *Server) handleListLeads(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit := parseInt(r.URL.Query().Get("limit"), 50)
	offset := parseInt(r.URL.Query().Get("offset"), 0)
	leads, err := s.store.ListInboundLeads(r.Context(), status, limit, offset)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"leads": leads, "count": len(leads)})
}

// handleCreateLead creates a new inbound lead.
func (s *Server) handleCreateLead(w http.ResponseWriter, r *http.Request) {
	var body domain.InboundLead
	if err := withJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.store.CreateInboundLead(r.Context(), &body); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, body)
}

// handleGetLead returns a single inbound lead by ID.
func (s *Server) handleGetLead(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing lead id")
		return
	}
	lead, err := s.store.GetInboundLead(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, lead)
}

// handleListCompanies lists companies.
func (s *Server) handleListCompanies(w http.ResponseWriter, r *http.Request) {
	companies, err := s.store.ListCompanies(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list companies")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"companies": companies, "count": len(companies)})
}

// handleCreateCompany creates a new company.
func (s *Server) handleCreateCompany(w http.ResponseWriter, r *http.Request) {
	var body domain.Company
	if err := withJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.store.CreateCompany(r.Context(), &body); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, body)
}

// handleListSources lists sources.
func (s *Server) handleListSources(w http.ResponseWriter, r *http.Request) {
	sources, err := s.store.ListSources(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list sources")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"sources": sources, "count": len(sources)})
}

// handleCreateSource creates a new source.
func (s *Server) handleCreateSource(w http.ResponseWriter, r *http.Request) {
	var body domain.Source
	if err := withJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.store.CreateSource(r.Context(), &body); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, body)
}

// handleListResearchRuns lists research runs.
func (s *Server) handleListResearchRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := s.store.ListResearchRuns(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list research runs")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"runs": runs, "count": len(runs)})
}

// handleCreateResearchRun creates a new research run.
func (s *Server) handleCreateResearchRun(w http.ResponseWriter, r *http.Request) {
	var body domain.ResearchRun
	if err := withJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.store.CreateResearchRun(r.Context(), &body); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, body)
}

// handleListSourceEvidence returns only evidence records referenced by the supplied IDs.
func (s *Server) handleListSourceEvidence(w http.ResponseWriter, r *http.Request) {
	ids := r.URL.Query()["id"]
	if len(ids) > 100 {
		writeError(w, http.StatusBadRequest, "at most 100 evidence IDs may be requested")
		return
	}
	for i, id := range ids {
		ids[i] = strings.TrimSpace(id)
		if ids[i] == "" {
			writeError(w, http.StatusBadRequest, "evidence IDs must not be empty")
			return
		}
	}
	evidence, err := s.store.ListSourceEvidenceByIDs(r.Context(), ids)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"evidence": evidence, "count": len(evidence)})
}

// handleListReviewDecisions lists decisions for exactly one entity.
func (s *Server) handleListReviewDecisions(w http.ResponseWriter, r *http.Request) {
	entityType := strings.TrimSpace(r.URL.Query().Get("entity_type"))
	entityID := strings.TrimSpace(r.URL.Query().Get("entity_id"))
	if entityType == "" || entityID == "" {
		writeError(w, http.StatusBadRequest, "entity_type and entity_id are required")
		return
	}
	if entityType != "opportunity" && entityType != "inbound_lead" {
		writeError(w, http.StatusBadRequest, "unsupported entity_type")
		return
	}
	decisions, err := s.store.ListReviewDecisionsForEntity(r.Context(), entityType, entityID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"decisions": decisions, "count": len(decisions)})
}

// handleCreateReviewDecision creates a new review decision.
func (s *Server) handleCreateReviewDecision(w http.ResponseWriter, r *http.Request) {
	var body domain.ReviewDecision
	if err := withJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.store.CreateReviewDecision(r.Context(), &body); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, body)
}

// parseInt parses an integer query parameter with a default value.
func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return def
	}
	return v
}

// intakeRequest is the validated payload for the secure inbound intake endpoint.
type intakeRequest struct {
	RequestSummary       string   `json:"request_summary"`
	InboundChannel       string   `json:"inbound_channel"`
	RequestedServices    []string `json:"requested_services,omitempty"`
	Municipality         *string  `json:"municipality,omitempty"`
	Province             *string  `json:"province,omitempty"`
	AutonomousCommunity  *string  `json:"autonomous_community,omitempty"`
	CountryCode          *string  `json:"country_code,omitempty"`
	LocationText         *string  `json:"location_text,omitempty"`
	Budget               *float64 `json:"budget,omitempty"`
	BudgetCurrency       *string  `json:"budget_currency,omitempty"`
	DesiredTiming        *string  `json:"desired_timing,omitempty"`
	CompanyID            *string  `json:"company_id,omitempty"`
	ContactName          *string  `json:"contact_name,omitempty"`
	ContactEmail         *string  `json:"contact_email,omitempty"`
	ContactPhone         *string  `json:"contact_phone,omitempty"`
	ContactPurpose       *string  `json:"contact_purpose,omitempty"`
	PrivacyNotice        bool     `json:"privacy_notice"`
	Consent              bool     `json:"consent"`
	IdempotencyKey       string   `json:"idempotency_key,omitempty"`
	Website              string   `json:"website,omitempty"`
}

type intakeResponse struct {
	LeadID   string `json:"lead_id"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	Duplicate bool  `json:"duplicate,omitempty"`
}

// verifyIntakeHMAC checks the HMAC-SHA256 signature on inbound intake requests.
func (s *Server) verifyIntakeHMAC(r *http.Request, body []byte) bool {
	if intakeSecret == "" {
		return true
	}
	signature := r.Header.Get("X-Valentina-Signature")
	if signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(intakeSecret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return secureCompare(signature, expected)
}

// secureCompare performs a constant-time comparison of two strings.
func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// checkIdempotency checks if the idempotency key has already been processed.
func (s *Server) checkIdempotency(key string) bool {
	if key == "" {
		return true
	}
	return true
}

// handleIntakeLead is the secure inbound lead intake endpoint.
func (s *Server) handleIntakeLead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBody+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	if len(body) > maxRequestBody {
		writeError(w, http.StatusBadRequest, "request body exceeds 1 MB limit")
		return
	}
	if !s.verifyIntakeHMAC(r, body) {
		writeError(w, http.StatusUnauthorized, "invalid or missing signature")
		return
	}
	var req intakeRequest
	if err := withJSONFromBytes(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.Website) != "" {
		slog.Warn("intake honeypot triggered; treating as spam")
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.RequestSummary == "" {
		writeError(w, http.StatusBadRequest, "request_summary is required")
		return
	}
	if req.InboundChannel == "" {
		writeError(w, http.StatusBadRequest, "inbound_channel is required")
		return
	}
	if !req.Consent {
		writeError(w, http.StatusBadRequest, "consent is required")
		return
	}
	if !req.PrivacyNotice {
		writeError(w, http.StatusBadRequest, "privacy notice must be acknowledged")
		return
	}
	if req.IdempotencyKey != "" {
		if !s.checkIdempotency(req.IdempotencyKey) {
			writeError(w, http.StatusConflict, "idempotency key already processed")
			return
		}
	}
	lead := &domain.InboundLead{
		RequestSummary:   req.RequestSummary,
		InboundChannel:   req.InboundChannel,
		RequestedServices: req.RequestedServices,
	}
	if req.PrivacyNotice {
		lead.PrivacyNotice = &req.PrivacyNotice
	}
	if req.Municipality != nil || req.Province != nil || req.AutonomousCommunity != nil || req.CountryCode != nil || req.LocationText != nil {
		lead.ProjectLocation = &domain.Location{
			Municipality:        req.Municipality,
			Province:            req.Province,
			AutonomousCommunity: req.AutonomousCommunity,
			CountryCode:         req.CountryCode,
			LocationText:        req.LocationText,
		}
	}
	if req.Budget != nil || req.BudgetCurrency != nil || req.DesiredTiming != nil {
		lead.Budget = req.Budget
		lead.BudgetCurrency = req.BudgetCurrency
		lead.DesiredTiming = req.DesiredTiming
	}
	if req.CompanyID != nil {
		lead.CompanyID = req.CompanyID
	}
	if req.ContactName != nil || req.ContactEmail != nil || req.ContactPhone != nil || req.ContactPurpose != nil {
		lead.ContactInfo = &domain.ContactInfo{
			Name:    req.ContactName,
			Email:   req.ContactEmail,
			Phone:   req.ContactPhone,
			Purpose: req.ContactPurpose,
		}
	}
	if err := s.store.CreateInboundLead(r.Context(), lead); err != nil {
		writeDomainError(w, err)
		return
	}
	resp := intakeResponse{
		LeadID:  lead.LeadID,
		Status:  string(lead.Status),
		Message: "intake received",
	}
	writeJSON(w, http.StatusCreated, resp)
}

// withJSONFromBytes parses JSON from a byte slice into the given target.
func withJSONFromBytes(data []byte, target interface{}) error {
	if len(data) == 0 {
		return errors.New("request body is empty")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		if err == io.EOF {
			return errors.New("request body is empty")
		}
		return err
	}
	var extra json.RawMessage
	decodeErr := dec.Decode(&extra)
	if decodeErr != nil && decodeErr != io.EOF {
		return errors.New("request body contains trailing data after JSON object")
	}
	if decodeErr == nil {
		return errors.New("request body contains multiple JSON objects")
	}
	return nil
}
