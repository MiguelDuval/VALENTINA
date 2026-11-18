package api

import (
	"net/http"
	"time"
)

// handleCancelResearchRun cancels a running research run.
func (s *Server) handleCancelResearchRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "research run id is required")
		return
	}
	if err := s.store.UpdateResearchRun(r.Context(), id, "cancelled", 0, 0, 0, 0, 0, "cancelled by operator"); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled", "run_id": id})
}

// sourceUpdateRequest is the payload for source approve/pause operations.
type sourceUpdateRequest struct {
	ApprovalStatus  string  `json:"approval_status"`
	ApprovalReason  string  `json:"approval_reason,omitempty"`
	TermsReviewedAt *string `json:"terms_reviewed_at,omitempty"`
	Notes           string  `json:"notes,omitempty"`
}

// handleApproveSource approves a source for fetching.
func (s *Server) handleApproveSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	sourceID := r.PathValue("id")
	if sourceID == "" {
		writeError(w, http.StatusBadRequest, "source id is required")
		return
	}
	var req sourceUpdateRequest
	if err := withJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.ApprovalStatus != "approved" {
		writeError(w, http.StatusBadRequest, "approval_status must be 'approved'")
		return
	}
	var termsReviewedAt *time.Time
	if req.TermsReviewedAt != nil && *req.TermsReviewedAt != "" {
		t, err := time.Parse(time.RFC3339, *req.TermsReviewedAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid terms_reviewed_at timestamp")
			return
		}
		termsReviewedAt = &t
	}
	now := time.Now().UTC()
	if err := s.store.UpdateSource(r.Context(), sourceID, "approved", req.ApprovalReason, termsReviewedAt, &now, req.Notes); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "approved", "source_id": sourceID})
}

// handlePauseSource pauses a source, preventing further fetches.
func (s *Server) handlePauseSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	sourceID := r.PathValue("id")
	if sourceID == "" {
		writeError(w, http.StatusBadRequest, "source id is required")
		return
	}
	var req sourceUpdateRequest
	if err := withJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.ApprovalStatus != "paused" {
		writeError(w, http.StatusBadRequest, "approval_status must be 'paused'")
		return
	}
	now := time.Now().UTC()
	if err := s.store.UpdateSource(r.Context(), sourceID, "paused", req.ApprovalReason, nil, &now, req.Notes); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "paused", "source_id": sourceID})
}

// handleTestSource validates that a source passes VALENTINA policy approval.
// It checks that the source is registered and approved for fetching.
// It does NOT perform an actual network request or verify external connectivity.
// Network connectivity is only validated at fetch time through the source policy gateway.
func (s *Server) handleTestSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	sourceID := r.PathValue("id")
	if sourceID == "" {
		writeError(w, http.StatusBadRequest, "source id is required")
		return
	}
	src, err := s.store.GetSource(r.Context(), sourceID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if !src.IsApproved() {
		writeError(w, http.StatusBadRequest, "source not approved")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":    "ok",
		"source_id": sourceID,
		"message":   "policy approval valid",
	})
}

// queryPack represents a configured search query pack for a language/region.
type queryPack struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Language    string   `json:"language"`
	Region      string   `json:"region"`
	Queries     []string `json:"queries"`
	Description string   `json:"description,omitempty"`
}

// handleListQueryPacks returns the configured query packs.
func (s *Server) handleListQueryPacks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	packs := []queryPack{
		{
			ID:          "es-renovation",
			Name:        "Espa\u00f1ol — Renovaci\u00f3n",
			Language:    "es",
			Region:      "ES",
			Queries: []string{
				"reforma vivienda Valencia",
				"reparaci\u00f3n edificio Valencia",
				"obras interiores Valencia",
				"fontaner\u00eda electricidad Valencia",
				"pintura alba\u00f1iler\u00eda Valencia",
			},
			Description: "Spanish renovation and repair queries for Valencia",
		},
		{
			ID:          "ca-renovacio",
			Name:        "Valenci\u00e0 — Renovaci\u00f3",
			Language:    "ca",
			Region:      "ES-VLC",
			Queries: []string{
				"reforma habitatge Val\u00e8ncia",
				"reparaci\u00f3 edifici Val\u00e8ncia",
				"obres interiors Val\u00e8ncia",
				"fontaneria electricitat Val\u00e8ncia",
				"pintura alba\u00f1iler\u00eda Val\u00e8ncia",
			},
			Description: "Valencian renovation and repair queries for Val\u00e8ncia",
		},
		{
			ID:          "ru-remont",
			Name:        "\u0420\u0443\u0441\u0441\u043a\u0438\u0439 — \u0420\u0435\u043c\u043e\u043d\u0442",
			Language:    "ru",
			Region:      "RU",
			Queries: []string{
				"\u0440\u0435\u043c\u043e\u043d\u0442 \u043a\u0432\u0430\u0440\u0442\u0438\u0440\u044b \u0412\u0430\u043b\u0435\u043d\u0446\u0438\u044f",
				"\u0440\u0435\u043c\u043e\u043d\u0442 \u0437\u0434\u0430\u043d\u0438\u044f \u0412\u0430\u043b\u0435\u043d\u0446\u0438\u044f",
				"\u0432\u043d\u0443\u0442\u0440\u0435\u043d\u043d\u0438\u0435 \u0440\u0430\u0431\u043e\u0442\u044b \u0412\u0430\u043b\u0435\u043d\u0446\u0438\u044f",
				"\u0441\u0430\u043d\u0442\u0435\u0445\u043d\u0438\u043a\u0430 \u044d\u043b\u0435\u043a\u0442\u0440\u0438\u043a\u0430 \u0412\u0430\u043b\u0435\u043d\u0446\u0438\u044f",
				"\u043a\u0440\u0430\u0441\u043a\u0430 \u0448\u0442\u0443\u043a\u0430\u0442\u0443\u0440\u043a\u0430 \u0412\u0430\u043b\u0435\u043d\u0446\u0438\u044f",
			},
			Description: "Russian renovation queries for Valencia area",
		},
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"query_packs": packs, "count": len(packs)})
}

// geographyConfig represents the owner-configured geographic scope.
type geographyConfig struct {
	PrimaryMunicipality   string   `json:"primary_municipality"`
	PrimaryProvince       string   `json:"primary_province"`
	AutonomousCommunity   string   `json:"autonomous_community"`
	CountryCode           string   `json:"country_code"`
	PriorityBands         []string `json:"priority_bands,omitempty"`
	HardExclusions        []string `json:"hard_exclusions,omitempty"`
}

// handleGetGeography returns the owner-configured geographic scope.
func (s *Server) handleGetGeography(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	cfg := geographyConfig{
		PrimaryMunicipality:  "Valencia",
		PrimaryProvince:      "Valencia",
		AutonomousCommunity:  "Comunitat Valenciana",
		CountryCode:          "ES",
		PriorityBands: []string{
			"Valencia municipality",
			"Valencia province",
			"Comunitat Valenciana",
		},
		HardExclusions: []string{},
	}
	writeJSON(w, http.StatusOK, cfg)
}