package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/scrypster/huginn/internal/valentina/domain"
)

func projectError(w http.ResponseWriter, err error) {
	if errors.Is(err, domain.ErrProjectConflict) {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeDomainError(w, err)
}

func (s *Server) handleLeadProject(w http.ResponseWriter, r *http.Request) {
	if _, err := s.store.GetInboundLead(r.Context(), r.PathValue("id")); err != nil {
		projectError(w, err)
		return
	}
	p, err := s.store.ProjectForLead(r.Context(), r.PathValue("id"))
	if err != nil {
		projectError(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{"project": p})
}
func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		offset := 0
		var err error
		if raw := r.URL.Query().Get("offset"); raw != "" {
			offset, err = strconv.Atoi(raw)
		}
		if err != nil || offset < 0 {
			writeError(w, 400, "invalid offset")
			return
		}
		list, err := s.store.ListProjects(r.Context(), offset)
		if err != nil {
			projectError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"projects": list})
		return
	}
	var in domain.ProjectInput
	if err := withJSON(r, &in); err != nil {
		writeError(w, 400, "invalid project JSON")
		return
	}
	p, err := s.store.CreateProject(r.Context(), in)
	if err != nil {
		projectError(w, err)
		return
	}
	writeJSON(w, 201, p)
}
func (s *Server) handleProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if r.Method == http.MethodGet {
		p, err := s.store.GetProject(r.Context(), id)
		if err != nil {
			projectError(w, err)
			return
		}
		writeJSON(w, 200, p)
		return
	}
	var in struct {
		domain.ProjectInput
		Version int64 `json:"version"`
	}
	if err := withJSON(r, &in); err != nil {
		writeError(w, 400, "invalid project JSON")
		return
	}
	p, err := s.store.UpdateProject(r.Context(), id, in.ProjectInput, in.Version)
	if err != nil {
		projectError(w, err)
		return
	}
	writeJSON(w, 200, p)
}
func (s *Server) handleProjectStatus(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Status  string `json:"status"`
		Version int64  `json:"version"`
	}
	if err := withJSON(r, &in); err != nil {
		writeError(w, 400, "invalid status JSON")
		return
	}
	p, err := s.store.SetProjectStatus(r.Context(), r.PathValue("id"), in.Status, in.Version)
	if err != nil {
		projectError(w, err)
		return
	}
	writeJSON(w, 200, p)
}
func (s *Server) handleProjectEstimates(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if r.Method == http.MethodGet {
		list, err := s.store.ListEstimates(r.Context(), id)
		if err != nil {
			projectError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"estimates": list})
		return
	}
	var in domain.EstimateInput
	if err := withJSON(r, &in); err != nil {
		writeError(w, 400, "invalid estimate JSON")
		return
	}
	e, err := s.store.SaveEstimate(r.Context(), id, in)
	if err != nil {
		projectError(w, err)
		return
	}
	writeJSON(w, 201, e)
}
func (s *Server) handleEstimatePreview(w http.ResponseWriter, r *http.Request) {
	if _, err := s.store.GetProject(r.Context(), r.PathValue("id")); err != nil {
		projectError(w, err)
		return
	}
	var in domain.EstimateInput
	if err := withJSON(r, &in); err != nil {
		writeError(w, 400, "invalid estimate JSON")
		return
	}
	totals, err := domain.CalculateEstimate(in.Lines)
	if err != nil {
		projectError(w, err)
		return
	}
	writeJSON(w, 200, totals)
}
func (s *Server) handleEstimateApproval(w http.ResponseWriter, r *http.Request) {
	var in struct{}
	if err := withJSON(r, &in); err != nil {
		writeError(w, 400, "invalid approval JSON")
		return
	}
	stamp, err := s.store.ApproveEstimate(r.Context(), r.PathValue("id"), r.PathValue("estimate"))
	if err != nil {
		projectError(w, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "approved", "approved_at": stamp})
}
func (s *Server) handleProjectDocuments(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if r.Method == http.MethodGet {
		list, err := s.store.ListProjectDocuments(r.Context(), id)
		if err != nil {
			projectError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"documents": list, "storage": "metadata_only"})
		return
	}
	var in domain.DocumentInput
	if err := withJSON(r, &in); err != nil {
		writeError(w, 400, "invalid document JSON")
		return
	}
	d, err := s.store.CreateProjectDocument(r.Context(), id, in)
	if err != nil {
		projectError(w, err)
		return
	}
	writeJSON(w, 201, d)
}

// Routes are consumed by the existing Huginn api(handler) bearer wrapper.
// No operations endpoints are added to the public/HMAC route exceptions.
func (s *Server) projectRoutes() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"GET /api/v1/valentina/leads/{id}/project":                          s.handleLeadProject,
		"GET /api/v1/valentina/projects":                                    s.handleProjects,
		"POST /api/v1/valentina/projects":                                   s.handleProjects,
		"GET /api/v1/valentina/projects/{id}":                               s.handleProject,
		"PUT /api/v1/valentina/projects/{id}":                               s.handleProject,
		"POST /api/v1/valentina/projects/{id}/status":                       s.handleProjectStatus,
		"GET /api/v1/valentina/projects/{id}/estimates":                     s.handleProjectEstimates,
		"POST /api/v1/valentina/projects/{id}/estimates":                    s.handleProjectEstimates,
		"POST /api/v1/valentina/projects/{id}/estimates/preview":            s.handleEstimatePreview,
		"POST /api/v1/valentina/projects/{id}/estimates/{estimate}/approve": s.handleEstimateApproval,
		"GET /api/v1/valentina/projects/{id}/documents":                     s.handleProjectDocuments,
		"POST /api/v1/valentina/projects/{id}/documents":                    s.handleProjectDocuments,
	}
}
