package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/scrypster/huginn/internal/valentina/internal/domain"
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

// withJSON parses a JSON request body into the given target.
func withJSON(r *http.Request, target interface{}) error {
	if r.Body == nil {
		return errors.New("request body is empty")
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

// writeJSON sends a JSON response.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Warn("failed to encode JSON response", "err", err)
	}
}

// writeError sends a JSON error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
// Routes returns a map of HTTP method+path to handler for VALENTINA API routes.
func (s *Server) Routes() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"GET /api/v1/valentina/opportunities":            s.handleListOpportunities,
		"POST /api/v1/valentina/opportunities":           s.handleCreateOpportunity,
		"GET /api/v1/valentina/opportunities/{id}":       s.handleGetOpportunity,
		"GET /api/v1/valentina/leads":                    s.handleListLeads,
		"POST /api/v1/valentina/leads":                   s.handleCreateLead,
		"GET /api/v1/valentina/leads/{id}":               s.handleGetLead,
		"GET /api/v1/valentina/companies":                s.handleListCompanies,
		"POST /api/v1/valentina/companies":               s.handleCreateCompany,
		"GET /api/v1/valentina/sources":                  s.handleListSources,
		"POST /api/v1/valentina/sources":                 s.handleCreateSource,
		"GET /api/v1/valentina/research-runs":            s.handleListResearchRuns,
		"POST /api/v1/valentina/research-runs":           s.handleCreateResearchRun,
		"GET /api/v1/valentina/review-decisions":         s.handleListReviewDecisions,
		"POST /api/v1/valentina/review-decisions":        s.handleCreateReviewDecision,
		"GET /api/v1/valentina/health":                   s.handleHealth,
	}
}

// handleHealth returns the health of the VALENTINA domain service.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "valentina-domain"})
}
// handleListOpportunities lists opportunities with optional status filter.
func (s *Server) handleListOpportunities(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit := parseInt(r.URL.Query().Get("limit"), 50)
	offset := parseInt(r.URL.Query().Get("offset"), 0)
	opps, err := s.store.ListOpportunities(r.Context(), status, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"opportunities": opps, "count": len(opps)})
}

// handleCreateOpportunity creates a new opportunity.
func (s *Server) handleCreateOpportunity(w http.ResponseWriter, r *http.Request) {
	var body domain.Opportunity
	if err := withJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.store.CreateOpportunity(r.Context(), &body); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, body)
}

// handleGetOpportunity returns a single opportunity by ID.
func (s *Server) handleGetOpportunity(w http.ResponseWriter, r *http.Request) {
	id := pathValue(r, "id")
	if id == "" {`n`t\twriteError(w, http.StatusBadRequest, "missing opportunity id")`n`t\treturn`n\t}`n\topp, err := s.store.GetOpportunity(r.Context(), id)`n\tif err != nil {`n`t\twriteError(w, http.StatusNotFound, err.Error())`n`t\treturn`n\t}`n\twriteJSON(w, http.StatusOK, opp)`n}
	writeError(w, http.StatusNotImplemented, "not implemented")
}

// handleListLeads lists inbound leads with optional status filter.
func (s *Server) handleListLeads(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit := parseInt(r.URL.Query().Get("limit"), 50)
	offset := parseInt(r.URL.Query().Get("offset"), 0)
	leads, err := s.store.ListInboundLeads(r.Context(), status, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
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
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, body)
}

// handleGetLead returns a single lead by ID.
func (s *Server) handleGetLead(w http.ResponseWriter, r *http.Request) {
	id := pathValue(r, "id")
	if id == "" {`n`t\twriteError(w, http.StatusBadRequest, "missing opportunity id")`n`t\treturn`n\t}`n\topp, err := s.store.GetOpportunity(r.Context(), id)`n\tif err != nil {`n`t\twriteError(w, http.StatusNotFound, err.Error())`n`t\treturn`n\t}`n\twriteJSON(w, http.StatusOK, opp)`n}
	writeError(w, http.StatusNotImplemented, "not implemented")
}

// handleListCompanies lists companies.
func (s *Server) handleListCompanies(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"companies": []interface{}{}, "count": 0})
}

// handleCreateCompany creates a new company.
func (s *Server) handleCreateCompany(w http.ResponseWriter, r *http.Request) {
	var body domain.Company
	if err := withJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.store.CreateCompany(r.Context(), &body); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, body)
}

// handleListSources lists sources.
func (s *Server) handleListSources(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"sources": []interface{}{}, "count": 0})
}

// handleCreateSource creates a new source.
func (s *Server) handleCreateSource(w http.ResponseWriter, r *http.Request) {
	var body domain.Source
	if err := withJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.store.CreateSource(r.Context(), &body); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, body)
}

// handleListResearchRuns lists research runs.
func (s *Server) handleListResearchRuns(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"runs": []interface{}{}, "count": 0})
}

// handleCreateResearchRun creates a new research run.
func (s *Server) handleCreateResearchRun(w http.ResponseWriter, r *http.Request) {
	var body domain.ResearchRun
	if err := withJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.store.CreateResearchRun(r.Context(), &body); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, body)
}

// handleListReviewDecisions lists review decisions.
func (s *Server) handleListReviewDecisions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"decisions": []interface{}{}, "count": 0})
}

// handleCreateReviewDecision creates a new review decision.
func (s *Server) handleCreateReviewDecision(w http.ResponseWriter, r *http.Request) {
	var body domain.ReviewDecision
	if err := withJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.store.CreateReviewDecision(r.Context(), &body); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
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

// pathValue extracts a URL path value by key (simple implementation).
func pathValue(r *http.Request, key string) string {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	for i, p := range parts {
		if p == key && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

var _ = fmt.Sprintf



