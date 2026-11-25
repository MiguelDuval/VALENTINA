package domain

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrProjectConflict = errors.New("project or revision changed; reload before retrying")

// projectSchema is append-only migration 14. Financial snapshots cannot be changed,
// even through an accidental future UPDATE; approval is a separate record.
const projectSchema = `
CREATE TABLE projects (
 project_id TEXT PRIMARY KEY, source_lead_id TEXT UNIQUE REFERENCES inbound_leads(lead_id),
 creation_hash TEXT NOT NULL, company_id TEXT REFERENCES companies(company_id),
 title TEXT NOT NULL, description TEXT NOT NULL, service_scope TEXT NOT NULL,
 location_text TEXT NOT NULL, start_date TEXT NOT NULL, end_date TEXT NOT NULL,
 status TEXT NOT NULL CHECK(status IN ('planned','active','on_hold','completed','cancelled')),
 version INTEGER NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL
);
CREATE TABLE project_estimates (
 estimate_id TEXT PRIMARY KEY, project_id TEXT NOT NULL REFERENCES projects(project_id),
 revision INTEGER NOT NULL CHECK(revision>0), snapshot TEXT NOT NULL,
 created_at TEXT NOT NULL, UNIQUE(project_id,revision), UNIQUE(project_id,estimate_id)
);
CREATE TABLE estimate_approvals (
 estimate_id TEXT PRIMARY KEY REFERENCES project_estimates(estimate_id), approved_at TEXT NOT NULL
);
CREATE TABLE project_documents (
 document_id TEXT PRIMARY KEY, project_id TEXT NOT NULL REFERENCES projects(project_id),
 type TEXT NOT NULL, name TEXT NOT NULL, annotation TEXT NOT NULL,
 estimate_id TEXT, created_at TEXT NOT NULL,
 FOREIGN KEY(project_id,estimate_id) REFERENCES project_estimates(project_id,estimate_id)
);
CREATE TRIGGER estimate_snapshot_immutable BEFORE UPDATE ON project_estimates BEGIN SELECT RAISE(ABORT,'immutable estimate'); END;
CREATE TRIGGER estimate_snapshot_no_delete BEFORE DELETE ON project_estimates BEGIN SELECT RAISE(ABORT,'immutable estimate'); END;
CREATE INDEX idx_project_documents_project ON project_documents(project_id);
`

type ProjectInput struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	ServiceScope string `json:"service_scope"`
	CompanyID    string `json:"company_id"`
	SourceLeadID string `json:"source_lead_id"`
	LocationText string `json:"location_text"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
}
type Project struct {
	ProjectInput
	ProjectID string `json:"project_id"`
	Status    string `json:"status"`
	Version   int64  `json:"version"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
type EstimateLine struct {
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	QuantityMilli int64  `json:"quantity_milli"`
	UnitCents     int64  `json:"unit_cents"`
	VATBPS        int64  `json:"vat_bps"`
}
type EstimateTotals struct {
	NetCents   int64 `json:"net_cents"`
	VATCents   int64 `json:"vat_cents"`
	TotalCents int64 `json:"total_cents"`
}
type EstimateInput struct {
	BaseRevision int            `json:"base_revision"`
	Lines        []EstimateLine `json:"lines"`
}
type Estimate struct {
	EstimateTotals
	EstimateID string         `json:"estimate_id"`
	ProjectID  string         `json:"project_id"`
	Revision   int            `json:"revision"`
	Currency   string         `json:"currency"`
	Lines      []EstimateLine `json:"lines"`
	Status     string         `json:"status"`
	CreatedAt  string         `json:"created_at"`
	ApprovedAt string         `json:"approved_at,omitempty"`
}
type DocumentInput struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	Annotation string `json:"annotation"`
	EstimateID string `json:"estimate_id"`
}
type ProjectDocument struct {
	DocumentInput
	DocumentID string `json:"document_id"`
	ProjectID  string `json:"project_id"`
	CreatedAt  string `json:"created_at"`
}

func projectValidation(message string) error { return errors.New("validation: " + message) }
func validText(value string, max int, required bool) bool {
	return len(value) <= max && (!required || strings.TrimSpace(value) != "") && !strings.ContainsRune(value, 0)
}
func validateProject(p ProjectInput) error {
	if !validText(p.Title, 240, true) || !validText(p.Description, 10000, false) || !validText(p.ServiceScope, 4000, false) || !validText(p.LocationText, 500, false) || !validText(p.CompanyID, 100, false) || !validText(p.SourceLeadID, 100, false) {
		return projectValidation("invalid project fields")
	}
	for _, d := range []string{p.StartDate, p.EndDate} {
		if d != "" {
			if _, err := time.Parse("2006-01-02", d); err != nil {
				return projectValidation("dates must be YYYY-MM-DD")
			}
		}
	}
	if p.StartDate != "" && p.EndDate != "" && p.StartDate > p.EndDate {
		return projectValidation("end date precedes start date")
	}
	return nil
}

// Money is never converted through float64. Round each line net half up, group
// the rounded net amounts by operator-entered VAT rate, then round VAT per group.
// Bound results to JS's exact integer range as they cross the browser JSON API.
const maxMoney int64 = 9007199254740991

func CalculateEstimate(lines []EstimateLine) (EstimateTotals, error) {
	var result EstimateTotals
	if len(lines) == 0 || len(lines) > 200 {
		return result, projectValidation("estimate requires 1-200 lines")
	}
	groups := map[int64]*big.Int{}
	net := new(big.Int)
	round := func(n *big.Int, d int64) *big.Int {
		return new(big.Int).Quo(new(big.Int).Add(n, big.NewInt(d/2)), big.NewInt(d))
	}
	for _, l := range lines {
		if (l.Kind != "labor" && l.Kind != "material") || !validText(l.Name, 240, true) || l.QuantityMilli <= 0 || l.QuantityMilli > maxMoney || l.UnitCents < 0 || l.UnitCents > maxMoney || l.VATBPS < 0 || l.VATBPS > 10000 {
			return result, projectValidation("invalid estimate line")
		}
		amount := round(new(big.Int).Mul(big.NewInt(l.QuantityMilli), big.NewInt(l.UnitCents)), 1000)
		net.Add(net, amount)
		if groups[l.VATBPS] == nil {
			groups[l.VATBPS] = new(big.Int)
		}
		groups[l.VATBPS].Add(groups[l.VATBPS], amount)
	}
	vat := new(big.Int)
	for rate, amount := range groups {
		vat.Add(vat, round(new(big.Int).Mul(amount, big.NewInt(rate)), 10000))
	}
	total := new(big.Int).Add(net, vat)
	if total.Cmp(big.NewInt(maxMoney)) > 0 {
		return result, projectValidation("estimate amount exceeds exact integer limit")
	}
	return EstimateTotals{net.Int64(), vat.Int64(), total.Int64()}, nil
}

const projectColumns = `project_id,COALESCE(source_lead_id,''),COALESCE(company_id,''),title,description,service_scope,location_text,start_date,end_date,status,version,created_at,updated_at`

type scanner interface{ Scan(...any) error }

func scanProject(row scanner) (*Project, error) {
	p := new(Project)
	err := row.Scan(&p.ProjectID, &p.SourceLeadID, &p.CompanyID, &p.Title, &p.Description, &p.ServiceScope, &p.LocationText, &p.StartDate, &p.EndDate, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}
func nullableID(s string) any {
	if s == "" {
		return nil
	}
	return s
}
func (s *Store) GetProject(ctx context.Context, id string) (*Project, error) {
	return scanProject(s.db.QueryRowContext(ctx, "SELECT "+projectColumns+" FROM projects WHERE project_id=?", id))
}

func (s *Store) ProjectForLead(ctx context.Context, id string) (*Project, error) {
	p, err := scanProject(s.db.QueryRowContext(ctx, "SELECT "+projectColumns+" FROM projects WHERE source_lead_id=?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return p, err
}
func (s *Store) ListProjects(ctx context.Context, offset int) ([]*Project, error) {
	if offset < 0 {
		return nil, projectValidation("invalid offset")
	}
	rows, err := s.db.QueryContext(ctx, "SELECT "+projectColumns+" FROM projects ORDER BY created_at DESC,project_id LIMIT 100 OFFSET ?", offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []*Project{}
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}
func checkCompany(ctx context.Context, tx *sql.Tx, id string) error {
	if id == "" {
		return nil
	}
	var exists int
	if err := tx.QueryRowContext(ctx, "SELECT 1 FROM companies WHERE company_id=?", id).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return projectValidation("unknown company")
		}
		return err
	}
	return nil
}
func (s *Store) CreateProject(ctx context.Context, in ProjectInput) (*Project, error) {
	if err := validateProject(in); err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(in)
	hash := sha256.Sum256(raw)
	fingerprint := hex.EncodeToString(hash[:])
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if in.SourceLeadID != "" {
		var id, previous string
		err = tx.QueryRowContext(ctx, "SELECT project_id,creation_hash FROM projects WHERE source_lead_id=?", in.SourceLeadID).Scan(&id, &previous)
		if err == nil {
			if previous != fingerprint {
				return nil, ErrProjectConflict
			}
			return scanProject(tx.QueryRowContext(ctx, "SELECT "+projectColumns+" FROM projects WHERE project_id=?", id))
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		var qualified int
		err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM inbound_leads l WHERE l.lead_id=? AND l.status='qualified' AND l.retention_state='active' AND EXISTS (SELECT 1 FROM review_decisions d WHERE d.entity_type='inbound_lead' AND d.entity_id=l.lead_id AND d.new_status='qualified')`, in.SourceLeadID).Scan(&qualified)
		if err != nil {
			return nil, err
		}
		if qualified != 1 {
			return nil, projectValidation("source lead must be human-qualified and active")
		}
	}
	if err = checkCompany(ctx, tx, in.CompanyID); err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	p := &Project{ProjectInput: in, ProjectID: uuid.NewString(), Status: "planned", Version: 1, CreatedAt: now, UpdatedAt: now}
	_, err = tx.ExecContext(ctx, `INSERT INTO projects(project_id,source_lead_id,creation_hash,company_id,title,description,service_scope,location_text,start_date,end_date,status,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, p.ProjectID, nullableID(in.SourceLeadID), fingerprint, nullableID(in.CompanyID), in.Title, in.Description, in.ServiceScope, in.LocationText, in.StartDate, in.EndDate, p.Status, p.Version, now, now)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return p, nil
}
func (s *Store) UpdateProject(ctx context.Context, id string, in ProjectInput, version int64) (*Project, error) {
	if err := validateProject(in); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	p, err := scanProject(tx.QueryRowContext(ctx, "SELECT "+projectColumns+" FROM projects WHERE project_id=?", id))
	if err != nil {
		return nil, err
	}
	if version != p.Version || in.SourceLeadID != p.SourceLeadID {
		return nil, ErrProjectConflict
	}
	if err = checkCompany(ctx, tx, in.CompanyID); err != nil {
		return nil, err
	}
	p.ProjectInput = in
	p.Version++
	p.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	_, err = tx.ExecContext(ctx, `UPDATE projects SET company_id=?,title=?,description=?,service_scope=?,location_text=?,start_date=?,end_date=?,version=?,updated_at=? WHERE project_id=?`, nullableID(in.CompanyID), in.Title, in.Description, in.ServiceScope, in.LocationText, in.StartDate, in.EndDate, p.Version, p.UpdatedAt, id)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return p, nil
}
func (s *Store) SetProjectStatus(ctx context.Context, id, status string, version int64) (*Project, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	p, err := scanProject(tx.QueryRowContext(ctx, "SELECT "+projectColumns+" FROM projects WHERE project_id=?", id))
	if err != nil {
		return nil, err
	}
	if p.Version != version {
		return nil, ErrProjectConflict
	}
	allowed := map[string]map[string]bool{
		"planned": {"active": true, "cancelled": true},
		"active":  {"on_hold": true, "completed": true, "cancelled": true},
		"on_hold": {"active": true, "cancelled": true},
	}
	if !allowed[p.Status][status] {
		return nil, projectValidation("invalid status transition")
	}
	p.Status = status
	p.Version++
	p.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	_, err = tx.ExecContext(ctx, "UPDATE projects SET status=?,version=?,updated_at=? WHERE project_id=?", status, p.Version, p.UpdatedAt, id)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return p, nil
}
func (s *Store) SaveEstimate(ctx context.Context, projectID string, in EstimateInput) (*Estimate, error) {
	totals, err := CalculateEstimate(in.Lines)
	if err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var id string
	if err = tx.QueryRowContext(ctx, "SELECT project_id FROM projects WHERE project_id=?", projectID).Scan(&id); err != nil {
		return nil, err
	}
	var latest int
	if err = tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(revision),0) FROM project_estimates WHERE project_id=?", projectID).Scan(&latest); err != nil {
		return nil, err
	}
	if in.BaseRevision != latest {
		return nil, ErrProjectConflict
	}
	e := &Estimate{EstimateTotals: totals, EstimateID: uuid.NewString(), ProjectID: projectID, Revision: latest + 1, Currency: "EUR", Lines: append([]EstimateLine(nil), in.Lines...), Status: "draft", CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	raw, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO project_estimates(estimate_id,project_id,revision,snapshot,created_at) VALUES(?,?,?,?,?)", e.EstimateID, projectID, e.Revision, string(raw), e.CreatedAt)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return e, nil
}
func (s *Store) ListEstimates(ctx context.Context, projectID string) ([]*Estimate, error) {
	if _, err := s.GetProject(ctx, projectID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT e.snapshot,COALESCE(a.approved_at,'') FROM project_estimates e LEFT JOIN estimate_approvals a ON a.estimate_id=e.estimate_id WHERE e.project_id=? ORDER BY e.revision DESC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []*Estimate{}
	for rows.Next() {
		var raw, approved string
		if err = rows.Scan(&raw, &approved); err != nil {
			return nil, err
		}
		e := new(Estimate)
		if err = json.Unmarshal([]byte(raw), e); err != nil {
			return nil, err
		}
		if approved != "" {
			e.Status = "approved"
			e.ApprovedAt = approved
		}
		list = append(list, e)
	}
	return list, rows.Err()
}
func (s *Store) ApproveEstimate(ctx context.Context, projectID, id string) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var found string
	if err = tx.QueryRowContext(ctx, "SELECT estimate_id FROM project_estimates WHERE project_id=? AND estimate_id=?", projectID, id).Scan(&found); err != nil {
		return "", err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err = tx.ExecContext(ctx, "INSERT INTO estimate_approvals(estimate_id,approved_at) VALUES(?,?) ON CONFLICT(estimate_id) DO NOTHING", id, now); err != nil {
		return "", err
	}
	if err = tx.QueryRowContext(ctx, "SELECT approved_at FROM estimate_approvals WHERE estimate_id=?", id).Scan(&now); err != nil {
		return "", err
	}
	return now, tx.Commit()
}
func (s *Store) CreateProjectDocument(ctx context.Context, projectID string, in DocumentInput) (*ProjectDocument, error) {
	if !validText(in.Name, 240, true) || strings.ContainsAny(in.Name, "/\\:") || !validText(in.Annotation, 4000, false) || !validText(in.EstimateID, 100, false) {
		return nil, projectValidation("invalid document metadata; names cannot be paths")
	}
	switch in.Type {
	case "estimate", "contract", "plan", "note", "other":
	default:
		return nil, projectValidation("invalid document type")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var exists string
	if err = tx.QueryRowContext(ctx, "SELECT project_id FROM projects WHERE project_id=?", projectID).Scan(&exists); err != nil {
		return nil, err
	}
	if in.EstimateID != "" {
		if err = tx.QueryRowContext(ctx, "SELECT estimate_id FROM project_estimates WHERE project_id=? AND estimate_id=?", projectID, in.EstimateID).Scan(&exists); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, projectValidation("estimate does not belong to project")
			}
			return nil, err
		}
	}
	d := &ProjectDocument{DocumentInput: in, DocumentID: uuid.NewString(), ProjectID: projectID, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	_, err = tx.ExecContext(ctx, "INSERT INTO project_documents(document_id,project_id,type,name,annotation,estimate_id,created_at) VALUES(?,?,?,?,?,?,?)", d.DocumentID, projectID, in.Type, in.Name, in.Annotation, nullableID(in.EstimateID), d.CreatedAt)
	if err != nil {
		return nil, err
	}
	return d, tx.Commit()
}
func (s *Store) ListProjectDocuments(ctx context.Context, projectID string) ([]*ProjectDocument, error) {
	if _, err := s.GetProject(ctx, projectID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, "SELECT document_id,type,name,annotation,COALESCE(estimate_id,''),created_at FROM project_documents WHERE project_id=? ORDER BY created_at DESC,document_id", projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []*ProjectDocument{}
	for rows.Next() {
		d := &ProjectDocument{ProjectID: projectID}
		if err = rows.Scan(&d.DocumentID, &d.Type, &d.Name, &d.Annotation, &d.EstimateID, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("read document: %w", err)
		}
		list = append(list, d)
	}
	return list, rows.Err()
}
