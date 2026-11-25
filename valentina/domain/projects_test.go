package domain

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestEstimateIntegerArithmetic(t *testing.T) {
	lines := []EstimateLine{{Kind: "labor", Name: "Synthetic labor", QuantityMilli: 1500, UnitCents: 101, VATBPS: 2100}, {Kind: "material", Name: "Synthetic material", QuantityMilli: 1000, UnitCents: 100, VATBPS: 1000}}
	totals, err := CalculateEstimate(lines)
	if err != nil || totals.NetCents != 252 || totals.VATCents != 42 || totals.TotalCents != 294 {
		t.Fatalf("totals: %+v %v", totals, err)
	}
	for _, line := range []EstimateLine{
		{Kind: "labor", Name: "x", QuantityMilli: 0, UnitCents: 1},
		{Kind: "labor", Name: "x", QuantityMilli: -1, UnitCents: 1},
		{Kind: "labor", Name: "x", QuantityMilli: 1, UnitCents: -1},
		{Kind: "labor", Name: "x", QuantityMilli: math.MaxInt64, UnitCents: math.MaxInt64},
		{Kind: "labor", Name: "x", QuantityMilli: 1000, UnitCents: 1, VATBPS: 10001},
	} {
		if _, err := CalculateEstimate([]EstimateLine{line}); err == nil {
			t.Fatalf("accepted invalid line: %+v", line)
		}
	}
	if _, err := CalculateEstimate(nil); err == nil {
		t.Fatal("accepted empty estimate")
	}
	if got, err := CalculateEstimate([]EstimateLine{{Kind: "labor", Name: "free", QuantityMilli: 1000}}); err != nil || got.TotalCents != 0 {
		t.Fatal(got, err)
	}
}

func TestProjectMigrationFromHistoricSchemas(t *testing.T) {
	for _, baseline := range []int{11, 12, 13} {
		t.Run(string(rune('A'+baseline)), func(t *testing.T) {
			ctx := context.Background()
			path := filepath.Join(t.TempDir(), "migration.db")
			db, err := sql.Open("sqlite", path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = db.Exec("CREATE TABLE schema_migrations(version INTEGER PRIMARY KEY,applied_at TEXT NOT NULL)"); err != nil {
				t.Fatal(err)
			}
			for _, m := range migrations {
				if m.version > baseline {
					break
				}
				if _, err = db.Exec(m.stmt); err != nil {
					t.Fatal(err)
				}
				if _, err = db.Exec("INSERT INTO schema_migrations VALUES(?,?)", m.version, time.Now().UTC().Format(time.RFC3339)); err != nil {
					t.Fatal(err)
				}
			}
			if baseline >= 12 {
				if _, err = db.Exec("INSERT INTO schema_migrations VALUES(12,'synthetic')"); err != nil {
					t.Fatal(err)
				}
			}
			// Existing v13 must already include the reconciliation column.
			if baseline == 13 {
				if _, err = db.Exec("ALTER TABLE research_runs ADD COLUMN updated_at TEXT"); err != nil {
					t.Fatal(err)
				}
			}
			if _, err = db.Exec(`INSERT INTO review_decisions(decision_id,entity_type,entity_id,new_status,decision_time,decision_reason) VALUES('synthetic-old','inbound_lead','old','qualified','2026-01-01','Preserve')`); err != nil {
				t.Fatal(err)
			}
			db.Close()
			s, err := NewStore(ctx, path)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			p, err := s.CreateProject(ctx, ProjectInput{Title: "Migrated synthetic"})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = s.GetProject(ctx, p.ProjectID); err != nil {
				t.Fatal(err)
			}
			var version, count int
			if err = s.db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version); err != nil || version != 14 {
				t.Fatal(version, err)
			}
			if err = s.db.QueryRow("SELECT COUNT(*) FROM review_decisions WHERE decision_id='synthetic-old'").Scan(&count); err != nil || count != 1 {
				t.Fatal(count, err)
			}
			if err = s.migrate(ctx); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestProjectValidationAndDocumentScope(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	defer s.Close()
	for _, in := range []ProjectInput{{}, {Title: "x", StartDate: "2026-02-30"}, {Title: "x", StartDate: "2026-12-01", EndDate: "2026-01-01"}, {Title: "x", CompanyID: "missing"}} {
		if _, err := s.CreateProject(ctx, in); err == nil {
			t.Fatalf("accepted %+v", in)
		}
	}
	p, err := s.CreateProject(ctx, ProjectInput{Title: "One"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := s.CreateProject(ctx, ProjectInput{Title: "Two"})
	if err != nil {
		t.Fatal(err)
	}
	e, err := s.SaveEstimate(ctx, other.ProjectID, EstimateInput{Lines: []EstimateLine{{Kind: "material", Name: "Synthetic", QuantityMilli: 1, UnitCents: 500}}})
	if err != nil {
		t.Fatal(err)
	}
	if e.NetCents != 1 {
		t.Fatal("half-cent must round up")
	}
	if _, err = s.CreateProjectDocument(ctx, p.ProjectID, DocumentInput{Type: "estimate", Name: "Wrong project", EstimateID: e.EstimateID}); err == nil {
		t.Fatal("cross-project association")
	}
	if _, err = s.ApproveEstimate(ctx, p.ProjectID, e.EstimateID); err == nil {
		t.Fatal("cross-project approval")
	}
	for _, name := range []string{"../x", "C:\\x", "https://example.test/x", ""} {
		if _, err = s.CreateProjectDocument(ctx, p.ProjectID, DocumentInput{Type: "note", Name: name}); err == nil {
			t.Fatal("path accepted")
		}
	}
	list, err := s.ListProjects(ctx, 0)
	if err != nil || len(list) != 2 {
		t.Fatal(list, err)
	}
	for _, status := range []string{"active", "on_hold", "active", "completed"} {
		p, err = s.SetProjectStatus(ctx, p.ProjectID, status, p.Version)
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err = s.SetProjectStatus(ctx, p.ProjectID, "active", p.Version); err == nil {
		t.Fatal("terminal project reopened")
	}
	if _, err = s.SetProjectStatus(ctx, other.ProjectID, "cancelled", other.Version); err != nil {
		t.Fatal(err)
	}
}

func TestVATGroupedRoundingAndEstimateConcurrency(t *testing.T) {
	lines := []EstimateLine{{Kind: "labor", Name: "A", QuantityMilli: 1000, UnitCents: 1, VATBPS: 2500}, {Kind: "material", Name: "B", QuantityMilli: 1000, UnitCents: 1, VATBPS: 2500}}
	totals, err := CalculateEstimate(lines)
	if err != nil || totals.VATCents != 1 || totals.TotalCents != 3 {
		t.Fatal(totals, err)
	}
	ctx := context.Background()
	s := testStore(t)
	defer s.Close()
	p, err := s.CreateProject(ctx, ProjectInput{Title: "Synthetic"})
	if err != nil {
		t.Fatal(err)
	}
	results := make(chan error, 6)
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.SaveEstimate(ctx, p.ProjectID, EstimateInput{Lines: lines})
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else if !errors.Is(err, ErrProjectConflict) {
			t.Fatal(err)
		}
	}
	if success != 1 {
		t.Fatal(success)
	}
	lines = []EstimateLine{{Kind: "labor", Name: "A", QuantityMilli: 1000, UnitCents: maxMoney}, {Kind: "labor", Name: "B", QuantityMilli: 1000, UnitCents: 1}}
	if _, err = CalculateEstimate(lines); err == nil {
		t.Fatal("sum overflow accepted")
	}
}

func TestQualificationRequiresHumanDecision(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	defer s.Close()
	for _, status := range []InboundLeadStatus{InboundLeadStatusQualified, InboundLeadStatusSpam} {
		lead := &InboundLead{RequestSummary: "Synthetic", InboundChannel: "manual", Status: status}
		if err := s.CreateInboundLead(ctx, lead); err != nil {
			t.Fatal(err)
		}
		if _, err := s.CreateProject(ctx, ProjectInput{Title: "Synthetic", SourceLeadID: lead.LeadID}); err == nil {
			t.Fatal("lead without human qualification converted")
		}
	}
}

func TestConnectedProjectHistoryAndReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "synthetic.db")
	s, err := NewStore(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { s.Close() }()
	p, err := s.CreateProject(ctx, ProjectInput{Title: "Synthetic project", ServiceScope: "Operator scope"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.SetProjectStatus(ctx, p.ProjectID, "completed", p.Version); err == nil {
		t.Fatal("invalid transition accepted")
	}
	p, err = s.SetProjectStatus(ctx, p.ProjectID, "active", p.Version)
	if err != nil {
		t.Fatal(err)
	}
	p, err = s.UpdateProject(ctx, p.ProjectID, ProjectInput{Title: "Edited", ServiceScope: "Scope"}, p.Version)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.UpdateProject(ctx, p.ProjectID, ProjectInput{Title: "Stale"}, 1); !errors.Is(err, ErrProjectConflict) {
		t.Fatal(err)
	}
	input := EstimateInput{Lines: []EstimateLine{{Kind: "labor", Name: "Synthetic", QuantityMilli: 1000, UnitCents: 100, VATBPS: 2100}}}
	e, err := s.SaveEstimate(ctx, p.ProjectID, input)
	if err != nil {
		t.Fatal(err)
	}
	if e.Status != "draft" || e.Revision != 1 {
		t.Fatal(e)
	}
	if _, err = s.ApproveEstimate(ctx, p.ProjectID, e.EstimateID); err != nil {
		t.Fatal(err)
	}
	input.BaseRevision = 1
	input.Lines[0].UnitCents = 200
	newer, err := s.SaveEstimate(ctx, p.ProjectID, input)
	if err != nil || newer.Revision != 2 || newer.Status != "draft" {
		t.Fatal(newer, err)
	}
	if _, err = s.SaveEstimate(ctx, p.ProjectID, input); !errors.Is(err, ErrProjectConflict) {
		t.Fatal(err)
	}
	doc, err := s.CreateProjectDocument(ctx, p.ProjectID, DocumentInput{Type: "estimate", Name: "Revision one", EstimateID: e.EstimateID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.db.Exec("UPDATE project_estimates SET snapshot='{}' WHERE estimate_id=?", e.EstimateID); err == nil {
		t.Fatal("mutable snapshot")
	}
	s.Close()
	s, err = NewStore(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	history, err := s.ListEstimates(ctx, p.ProjectID)
	if err != nil || len(history) != 2 || history[1].TotalCents != 121 || history[1].Status != "approved" {
		t.Fatal(history, err)
	}
	docs, err := s.ListProjectDocuments(ctx, p.ProjectID)
	if err != nil || len(docs) != 1 || docs[0].DocumentID != doc.DocumentID {
		t.Fatal(docs, err)
	}
	got, err := s.GetProject(ctx, p.ProjectID)
	if err != nil || got.Title != "Edited" || got.Status != "active" {
		t.Fatal(got, err)
	}
}

func TestHumanLeadConversionConcurrent(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	defer s.Close()
	lead := &InboundLead{RequestSummary: "Synthetic enquiry", InboundChannel: "manual"}
	if err := s.CreateInboundLead(ctx, lead); err != nil {
		t.Fatal(err)
	}
	input := ProjectInput{Title: "Synthetic", SourceLeadID: lead.LeadID}
	if _, err := s.CreateProject(ctx, input); err == nil {
		t.Fatal("unqualified conversion")
	}
	for _, status := range []string{"triage", "qualified"} {
		if err := s.CreateReviewDecision(ctx, &ReviewDecision{EntityType: "inbound_lead", EntityID: lead.LeadID, NewStatus: status, DecisionReason: "Human synthetic review"}); err != nil {
			t.Fatal(err)
		}
	}
	var wg sync.WaitGroup
	ids := make(chan string, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p, err := s.CreateProject(ctx, input)
			if err != nil {
				t.Error(err)
				return
			}
			ids <- p.ProjectID
		}()
	}
	wg.Wait()
	close(ids)
	id := ""
	for got := range ids {
		if id != "" && got != id {
			t.Fatal("duplicate project")
		}
		id = got
	}
	input.Title = "Different"
	if _, err := s.CreateProject(ctx, input); !errors.Is(err, ErrProjectConflict) {
		t.Fatal(err)
	}
	original, err := s.GetInboundLead(ctx, lead.LeadID)
	if err != nil || original.Status != InboundLeadStatusQualified {
		t.Fatal(original, err)
	}
	decisions, err := s.ListReviewDecisionsForEntity(ctx, "inbound_lead", lead.LeadID)
	if err != nil || len(decisions) != 2 {
		t.Fatal(decisions, err)
	}
}
