package domain

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"modernc.org/sqlite"
)

// migration describes a single schema migration with version and SQL.
type migration struct {
	version int
	stmt    string
}

// migrations are applied in order. Each migration is idempotent.
var migrations = []migration{
	{version: 1, stmt: `
		CREATE TABLE IF NOT EXISTS companies (
			company_id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			normalized_name TEXT NOT NULL,
			website TEXT,
			domain TEXT,
			industry_tags TEXT,
			municipality TEXT,
			province TEXT,
			autonomous_community TEXT,
			country_code TEXT,
			location_text TEXT,
			source_evidence_ids TEXT NOT NULL,
			first_seen_at TEXT NOT NULL,
			last_seen_at TEXT NOT NULL,
			notes TEXT,
			UNIQUE(normalized_name)
		);`},
	{version: 2, stmt: `
		CREATE TABLE IF NOT EXISTS opportunities (
			opportunity_id TEXT PRIMARY KEY,
			company_id TEXT,
			title TEXT NOT NULL,
			summary TEXT NOT NULL,
			opportunity_type TEXT NOT NULL,
			status TEXT NOT NULL,
			municipality TEXT,
			province TEXT,
			autonomous_community TEXT,
			country_code TEXT,
			location_text TEXT,
			published_at TEXT,
			deadline_at TEXT,
			valid_from TEXT,
			valid_until TEXT,
			value_amount REAL,
			value_currency TEXT,
			value_basis TEXT,
			fit_profile_version TEXT,
			fit_status TEXT,
			fit_evaluated_at TEXT,
			source_evidence_ids TEXT NOT NULL,
			first_seen_at TEXT NOT NULL,
			last_seen_at TEXT NOT NULL,
			duplicate_group_id TEXT,
			review_note TEXT,
			demand_evidence_type TEXT NOT NULL,
			demand_evidence_note TEXT,
			FOREIGN KEY (company_id) REFERENCES companies(company_id)
		);`},
	{version: 3, stmt: `
		CREATE TABLE IF NOT EXISTS source_evidence (
			evidence_id TEXT PRIMARY KEY,
			source_id TEXT NOT NULL,
			canonical_url TEXT NOT NULL,
			publisher TEXT NOT NULL,
			page_title TEXT,
			observed_at TEXT NOT NULL,
			published_at TEXT,
			content_hash TEXT,
			retrieval_status TEXT NOT NULL,
			extractor_id TEXT,
			extractor_version TEXT,
			facts TEXT NOT NULL
		);`},
	{version: 4, stmt: `
		CREATE TABLE IF NOT EXISTS sources (
			source_id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			publisher TEXT NOT NULL,
			source_type TEXT NOT NULL,
			base_url TEXT NOT NULL,
			allowed_hosts TEXT NOT NULL,
			allowed_path_prefixes TEXT,
			terms_url TEXT,
			terms_reviewed_at TEXT,
			robots_reviewed_at TEXT,
			approval_status TEXT NOT NULL,
			approval_reason TEXT,
			retrieval_method TEXT NOT NULL,
			adapter_id TEXT,
			adapter_version TEXT,
			authentication_required INTEGER NOT NULL DEFAULT 0,
			credential_reference TEXT,
			minimum_interval_seconds INTEGER NOT NULL,
			max_parallel_requests INTEGER NOT NULL,
			daily_request_limit INTEGER,
			timeout_seconds INTEGER NOT NULL,
			max_response_bytes INTEGER NOT NULL,
			allowed_fields TEXT NOT NULL,
			prohibited_fields TEXT,
			languages TEXT,
			expected_update_frequency TEXT,
			last_reviewed_at TEXT,
			notes TEXT
		);`},
	{version: 5, stmt: `
		CREATE TABLE IF NOT EXISTS research_runs (
			run_id TEXT PRIMARY KEY,
			started_at TEXT NOT NULL,
			completed_at TEXT,
			status TEXT NOT NULL,
			source_ids TEXT,
			candidate_count INTEGER NOT NULL DEFAULT 0,
			verified_count INTEGER NOT NULL DEFAULT 0,
			duplicate_count INTEGER NOT NULL DEFAULT 0,
			accepted_count INTEGER NOT NULL DEFAULT 0,
			rejected_count INTEGER NOT NULL DEFAULT 0,
			notes TEXT
		);`},
	{version: 6, stmt: `
		CREATE TABLE IF NOT EXISTS review_decisions (
			decision_id TEXT PRIMARY KEY,
			entity_type TEXT NOT NULL,
			entity_id TEXT NOT NULL,
			previous_status TEXT,
			new_status TEXT NOT NULL,
			decision_time TEXT NOT NULL,
			decision_reason TEXT NOT NULL,
			profile_version TEXT,
			operator_ref TEXT
		);`},
	{version: 7, stmt: `
		CREATE TABLE IF NOT EXISTS inbound_leads (
			lead_id TEXT PRIMARY KEY,
			received_at TEXT NOT NULL,
			inbound_channel TEXT NOT NULL,
			channel_message_ref TEXT,
			request_summary TEXT NOT NULL,
			original_content TEXT,
			requested_services TEXT,
			municipality TEXT,
			province TEXT,
			autonomous_community TEXT,
			country_code TEXT,
			location_text TEXT,
			budget REAL,
			budget_currency TEXT,
			desired_timing TEXT,
			company_id TEXT,
			contact_name TEXT,
			contact_email TEXT,
			contact_phone TEXT,
			contact_purpose TEXT,
			ai_service_category TEXT,
			ai_project_scope TEXT,
			ai_location TEXT,
			ai_budget REAL,
			ai_timing_urgency TEXT,
			ai_missing_details TEXT,
			ai_company_match TEXT,
			ai_response_draft TEXT,
			ai_evaluated_at TEXT,
			ai_profile_version TEXT,
			missing_information TEXT,
			duplicate_group_id TEXT,
			status TEXT NOT NULL,
			privacy_notice INTEGER,
			retention_state TEXT NOT NULL DEFAULT 'active',
			FOREIGN KEY (company_id) REFERENCES companies(company_id)
		);`},
	{version: 8, stmt: `
		CREATE INDEX IF NOT EXISTS idx_opportunities_status ON opportunities(status);
		CREATE INDEX IF NOT EXISTS idx_opportunities_company ON opportunities(company_id);
		CREATE INDEX IF NOT EXISTS idx_opportunities_demand_type ON opportunities(demand_evidence_type);
		CREATE INDEX IF NOT EXISTS idx_opportunities_first_seen ON opportunities(first_seen_at);
		CREATE INDEX IF NOT EXISTS idx_inbound_leads_status ON inbound_leads(status);
		CREATE INDEX IF NOT EXISTS idx_inbound_leads_received ON inbound_leads(received_at);
		CREATE INDEX IF NOT EXISTS idx_inbound_leads_company ON inbound_leads(company_id);
		CREATE INDEX IF NOT EXISTS idx_source_evidence_source ON source_evidence(source_id);
		CREATE INDEX IF NOT EXISTS idx_review_decisions_entity ON review_decisions(entity_type, entity_id);
		CREATE INDEX IF NOT EXISTS idx_research_runs_started ON research_runs(started_at);`},
}

// companiesTable is the CREATE TABLE statement for the companies table.
const companiesTable = `
CREATE TABLE IF NOT EXISTS companies (
	company_id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	normalized_name TEXT NOT NULL,
	website TEXT,
	domain TEXT,
	industry_tags TEXT,
	municipality TEXT,
	province TEXT,
	autonomous_community TEXT,
	country_code TEXT,
	location_text TEXT,
	source_evidence_ids TEXT NOT NULL,
	first_seen_at TEXT NOT NULL,
	last_seen_at TEXT NOT NULL,
	notes TEXT,
	UNIQUE(normalized_name)
);`

// Store is the SQLite-backed persistence layer for VALENTINA domain records.
type Store struct {
	db *sql.DB
}

// NewStore opens (creating if needed) the SQLite database at the given path.
func NewStore(ctx context.Context, dbPath string) (*Store, error) {
	if dbPath == "" {
		return nil, errors.New("database path is required")
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// migrate applies schema migrations in order. Each migration is idempotent.
func (s *Store) migrate(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Create schema_migrations table to track applied migrations
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL
		);`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	// Get already applied migrations
	applied := make(map[int]bool)
	rows, err := tx.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("query schema_migrations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return fmt.Errorf("scan migration version: %w", err)
		}
		applied[v] = true
	}

	// Apply pending migrations
	for _, m := range migrations {
		if applied[m.version] {
			continue
		}
		if _, err := tx.Exec(m.stmt); err != nil {
			return fmt.Errorf("apply migration %d: %w", m.version, err)
		}
		if _, err := tx.Exec("INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)", m.version, time.Now().UTC().Format(time.RFC3339)); err != nil {
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migrations: %w", err)
	}
	return nil
}

// CreateCompany inserts a new company record.
func (s *Store) CreateCompany(ctx context.Context, c *Company) error {
	if c.CompanyID == "" {
		c.CompanyID = NewCompanyID()
	}
	c.NormalizedName = NormalizeCompanyName(c.Name)
	now := time.Now().UTC()
	if c.FirstSeenAt.IsZero() {
		c.FirstSeenAt = now
	}
	c.LastSeenAt = now

	industryTags, err := json.Marshal(c.IndustryTags)
	if err != nil {
		return fmt.Errorf("marshal industry_tags: %w", err)
	}
	sourceEvid, err := json.Marshal(c.SourceEvidenceIDs)
	if err != nil {
		return fmt.Errorf("marshal source_evidence_ids: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO companies (company_id, name, normalized_name, website, domain, industry_tags,
			municipality, province, autonomous_community, country_code, location_text,
			source_evidence_ids, first_seen_at, last_seen_at, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.CompanyID, c.Name, c.NormalizedName, c.Website, c.Domain, string(industryTags),
		locStr(c.Location, "municipality"), locStr(c.Location, "province"),
		locStr(c.Location, "autonomous_community"), locStr(c.Location, "country_code"), locStr(c.Location, "location_text"),
		string(sourceEvid), c.FirstSeenAt.Format(time.RFC3339), c.LastSeenAt.Format(time.RFC3339), c.Notes)
	return err
}

// sourceEvidenceTable is the CREATE TABLE statement for source evidence.
// canonical_url is NOT unique - multiple observations of the same URL are allowed.
const sourceEvidenceTable = `
CREATE TABLE IF NOT EXISTS source_evidence (
	evidence_id TEXT PRIMARY KEY,
	source_id TEXT NOT NULL,
	canonical_url TEXT NOT NULL,
	publisher TEXT NOT NULL,
	page_title TEXT,
	observed_at TEXT NOT NULL,
	published_at TEXT,
	content_hash TEXT,
	retrieval_status TEXT NOT NULL,
	extractor_id TEXT,
	extractor_version TEXT,
	facts TEXT NOT NULL
);`

// sourcesTable is the CREATE TABLE statement for sources.
const sourcesTable = `
CREATE TABLE IF NOT EXISTS sources (
	source_id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	publisher TEXT NOT NULL,
	source_type TEXT NOT NULL,
	base_url TEXT NOT NULL,
	allowed_hosts TEXT NOT NULL,
	allowed_path_prefixes TEXT,
	terms_url TEXT,
	terms_reviewed_at TEXT,
	robots_reviewed_at TEXT,
	approval_status TEXT NOT NULL,
	approval_reason TEXT,
	retrieval_method TEXT NOT NULL,
	adapter_id TEXT,
	adapter_version TEXT,
	authentication_required INTEGER NOT NULL DEFAULT 0,
	credential_reference TEXT,
	minimum_interval_seconds INTEGER NOT NULL,
	max_parallel_requests INTEGER NOT NULL,
	daily_request_limit INTEGER,
	timeout_seconds INTEGER NOT NULL,
	max_response_bytes INTEGER NOT NULL,
	allowed_fields TEXT NOT NULL,
	prohibited_fields TEXT,
	languages TEXT,
	expected_update_frequency TEXT,
	last_reviewed_at TEXT,
	notes TEXT
);`

// researchRunsTable is the CREATE TABLE statement for research runs.
const researchRunsTable = `
CREATE TABLE IF NOT EXISTS research_runs (
	run_id TEXT PRIMARY KEY,
	started_at TEXT NOT NULL,
	completed_at TEXT,
	status TEXT NOT NULL,
	source_ids TEXT,
	candidate_count INTEGER NOT NULL DEFAULT 0,
	verified_count INTEGER NOT NULL DEFAULT 0,
	duplicate_count INTEGER NOT NULL DEFAULT 0,
	accepted_count INTEGER NOT NULL DEFAULT 0,
	rejected_count INTEGER NOT NULL DEFAULT 0,
	notes TEXT
);`

// reviewDecisionsTable is the CREATE TABLE statement for review decisions.
const reviewDecisionsTable = `
CREATE TABLE IF NOT EXISTS review_decisions (
	decision_id TEXT PRIMARY KEY,
	entity_type TEXT NOT NULL,
	entity_id TEXT NOT NULL,
	previous_status TEXT,
	new_status TEXT NOT NULL,
	decision_time TEXT NOT NULL,
	decision_reason TEXT NOT NULL,
	profile_version TEXT,
	operator_ref TEXT
);`

// inboundLeadsTable is the CREATE TABLE statement for inbound leads.
const inboundLeadsTable = `
CREATE TABLE IF NOT EXISTS inbound_leads (
	lead_id TEXT PRIMARY KEY,
	received_at TEXT NOT NULL,
	inbound_channel TEXT NOT NULL,
	channel_message_ref TEXT,
	request_summary TEXT NOT NULL,
	original_content TEXT,
	requested_services TEXT,
	municipality TEXT,
	province TEXT,
	autonomous_community TEXT,
	country_code TEXT,
	location_text TEXT,
	budget REAL,
	budget_currency TEXT,
	desired_timing TEXT,
	company_id TEXT,
	contact_name TEXT,
	contact_email TEXT,
	contact_phone TEXT,
	contact_purpose TEXT,
	ai_service_category TEXT,
	ai_project_scope TEXT,
	ai_location TEXT,
	ai_budget REAL,
	ai_timing_urgency TEXT,
	ai_missing_details TEXT,
	ai_company_match TEXT,
	ai_response_draft TEXT,
	ai_evaluated_at TEXT,
	ai_profile_version TEXT,
	missing_information TEXT,
	duplicate_group_id TEXT,
	status TEXT NOT NULL,
	privacy_notice INTEGER,
	retention_state TEXT NOT NULL DEFAULT 'active',
	FOREIGN KEY (company_id) REFERENCES companies(company_id)
);`

// indexesTable creates indexes for common query patterns.
const indexesTable = `
CREATE INDEX IF NOT EXISTS idx_opportunities_status ON opportunities(status);
CREATE INDEX IF NOT EXISTS idx_opportunities_company ON opportunities(company_id);
CREATE INDEX IF NOT EXISTS idx_opportunities_demand_type ON opportunities(demand_evidence_type);
CREATE INDEX IF NOT EXISTS idx_opportunities_first_seen ON opportunities(first_seen_at);
CREATE INDEX IF NOT EXISTS idx_inbound_leads_status ON inbound_leads(status);
CREATE INDEX IF NOT EXISTS idx_inbound_leads_received ON inbound_leads(received_at);
CREATE INDEX IF NOT EXISTS idx_inbound_leads_company ON inbound_leads(company_id);
CREATE INDEX IF NOT EXISTS idx_source_evidence_source ON source_evidence(source_id);
CREATE INDEX IF NOT EXISTS idx_review_decisions_entity ON review_decisions(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_research_runs_started ON research_runs(started_at);
`
// locStr extracts a string field from a Location pointer by field name.
func locStr(l *Location, field string) *string {
	if l == nil {
		return nil
	}
	switch field {
	case "municipality":
		return l.Municipality
	case "province":
		return l.Province
	case "autonomous_community":
		return l.AutonomousCommunity
	case "country_code":
		return l.CountryCode
	case "location_text":
		return l.LocationText
	}
	return nil
}

// timePtr formats a time.Time pointer as RFC3339, or nil if zero/nil.
func timePtr(t *time.Time) interface{} {
	if t == nil || t.IsZero() {
		return nil
	}
	return t.Format(time.RFC3339)
}

// boolToInt converts a bool pointer to an int pointer (1/0), or nil.
func boolToInt(b *bool) *int {
	if b == nil {
		return nil
	}
	if *b {
		v := 1
		return &v
	}
	v := 0
	return &v
}

// CreateOpportunity inserts a new opportunity record.
func (s *Store) CreateOpportunity(ctx context.Context, o *Opportunity) error {
	if o.OpportunityID == "" {\r\n\t\to.OpportunityID = NewOpportunityID()\r\n\t}\r\n\tif err := o.Validate(); err != nil {\r\n\t\treturn fmt.Errorf("validation: %w", err)\r\n\t}
	now := time.Now().UTC()
	if o.FirstSeenAt.IsZero() {
		o.FirstSeenAt = now
	}
	o.LastSeenAt = now

	srcEvid, _ := json.Marshal(o.SourceEvidenceIDs)
	var fitProf, fitStatus, fitEval string
	if o.FitAssessment != nil {
		fitProf = o.FitAssessment.ProfileVersion
		fitStatus = o.FitAssessment.FitStatus
		if !o.FitAssessment.EvaluatedAt.IsZero() {
			fitEval = o.FitAssessment.EvaluatedAt.Format(time.RFC3339)
		}
	}
	var valAmount interface{}
	var valCurr, valBasis *string
	if o.Value != nil {
		if o.Value.Amount != nil {
			valAmount = *o.Value.Amount
		}
		valCurr = o.Value.Currency
		valBasis = o.Value.Basis
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO opportunities (opportunity_id, company_id, title, summary, opportunity_type, status,
			municipality, province, autonomous_community, country_code, location_text,
			published_at, deadline_at, valid_from, valid_until,
			value_amount, value_currency, value_basis,
			fit_profile_version, fit_status, fit_evaluated_at,
			source_evidence_ids, first_seen_at, last_seen_at,
			review_note, demand_evidence_type, demand_evidence_note)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		o.OpportunityID, o.CompanyID, o.Title, o.Summary, o.OpportunityType, o.Status,
		locStr(o.Geography, "municipality"), locStr(o.Geography, "province"),
		locStr(o.Geography, "autonomous_community"), locStr(o.Geography, "country_code"), locStr(o.Geography, "location_text"),
		timePtr(o.PublishedAt), timePtr(o.DeadlineAt), timePtr(o.ValidFrom), timePtr(o.ValidUntil),
		valAmount, valCurr, valBasis,
		fitProf, fitStatus, fitEval,
		string(srcEvid), o.FirstSeenAt.Format(time.RFC3339), o.LastSeenAt.Format(time.RFC3339),
		o.ReviewNote, o.DemandEvidenceType, o.DemandEvidenceNote)
	return err
}
// GetOpportunity returns a single opportunity by ID.`nfunc (s *Store) GetOpportunity(ctx context.Context, id string) (*Opportunity, error) {`n`tvar o Opportunity`n`tvar fitProf, fitStatus, fitEval sql.NullString`n`tvar valAmount sql.NullFloat64`n`tvar valCurr, valBasis sql.NullString`n`tvar pubAt, deadAt, vf, vu sql.NullString`n`tvar firstSeen, lastSeen string`n`tvar srcEvid string`n`tvar dupGroup, reviewNote, demType, demNote sql.NullString`n`terr := s.db.QueryRowContext(ctx, "SELECT opportunity_id, company_id, title, summary, opportunity_type, status,`n`t\t\t\t\t\tmunicipality, province, autonomous_community, country_code, location_text,`n`t\t\t\t\tpublished_at, deadline_at, valid_from, valid_until,`n`t\t\t\t\tvalue_amount, value_currency, value_basis,`n`t\t\t\t\tfit_profile_version, fit_status, fit_evaluated_at,`n`t\t\t\t\tsource_evidence_ids, first_seen_at, last_seen_at,`n`t\t\t\t\tduplicate_group_id, review_note,`n`t\t\t\t\tdemand_evidence_type, demand_evidence_note`n`t\t\t\t\tFROM opportunities WHERE opportunity_id = ?", o.OpportunityID).Scan(`n`t\t\t&o.OpportunityID, &o.CompanyID, &o.Title, &o.Summary, &o.OpportunityType, &o.Status,`n`t\t\tlocScan(&o.Geography, "municipality"), locScan(&o.Geography, "province"),`n`t\t\tlocScan(&o.Geography, "autonomous_community"), locScan(&o.Geography, "country_code"), locScan(&o.Geography, "location_text"),`n`t\t\ttimeScan(&pubAt), timeScan(&deadAt), timeScan(&vf), timeScan(&vu),`n`t\t\t&valAmount, &valCurr, &valBasis,`n`t\t\t&fitProf, &fitStatus, &fitEval,`n`t\t\t&srcEvid, &firstSeen, &lastSeen,`n`t\t\t&dupGroup, &reviewNote,`n`t\t\t&demType, &demNote,`n`t\t)`n`t`tif err != nil {`n`t\t\treturn nil, err`n`t\t}`n`t\to.PublishedAt = parseTimePtr(pubAt)`n`t\to.DeadlineAt = parseTimePtr(deadAt)`n`t\to.ValidFrom = parseTimePtr(vf)`n`t\to.ValidUntil = parseTimePtr(vu)`n`t\tif valAmount.Valid {`n`t\t\tv := valAmount.Float64`n`t\t\to.Value = &OpportunityValue{Amount: &v}`n`t\t}`n`t\to.Value.Currency = valCurr`n`t\to.Value.Basis = valBasis`n`t\tif fitProf.Valid {`n`t\t\to.FitAssessment = &FitAssessment{ProfileVersion: fitProf.String, FitStatus: fitStatus.String}`n`t\t\tif fitEval.Valid {`n`t\t\t\to.FitAssessment.EvaluatedAt = parseTime(fitEval.String)`n`t\t\t}`n`t\t}`n`t\tif err := json.Unmarshal([]byte(srcEvid), &o.SourceEvidenceIDs); err != nil {`n`t\t\treturn nil, fmt.Errorf("unmarshal source_evidence_ids: %w", err)`n`t\t}`n`t\to.FirstSeenAt = parseTime(firstSeen)`n`t\to.LastSeenAt = parseTime(lastSeen)`n`t\to.DuplicateGroupID = dupGroup`n`t\to.ReviewNote = reviewNote`n`t\to.DemandEvidenceType = DemandEvidenceType(demType.String)`n`t\to.DemandEvidenceNote = demNote`n`t\treturn &o, nil`n`t}`n`n// CreateSourceEvidence inserts a new source evidence record.
func (s *Store) CreateSourceEvidence(ctx context.Context, e *SourceEvidence) error {
	if e.EvidenceID == "" {
		e.EvidenceID = NewEvidenceID()
	}
	if e.ObservedAt.IsZero() {
		e.ObservedAt = time.Now().UTC()
	}
	facts, _ := json.Marshal(e.Facts)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO source_evidence (evidence_id, source_id, canonical_url, publisher, page_title,
			observed_at, published_at, content_hash, retrieval_status, extractor_id, extractor_version, facts)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.EvidenceID, e.SourceID, e.CanonicalURL, e.Publisher, e.PageTitle,
		e.ObservedAt.Format(time.RFC3339), timePtr(e.PublishedAt), e.ContentHash, e.RetrievalStatus,
		e.ExtractorID, e.ExtractorVersion, string(facts))
	return err
}

// CreateSource inserts a new source record.
func (s *Store) CreateSource(ctx context.Context, src *Source) error {
	if src.SourceID == "" {
		return errors.New("source_id is required")
	}
	if src.ApprovalStatus == "" {
		src.ApprovalStatus = "pending_review"
	}
	allowedHosts, _ := json.Marshal(src.AllowedHosts)
	allowedPaths, _ := json.Marshal(src.AllowedPathPrefixes)
	allowedFields, _ := json.Marshal(src.AllowedFields)
	prohibitedFields, _ := json.Marshal(src.ProhibitedFields)
	languages, _ := json.Marshal(src.Languages)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sources (source_id, name, publisher, source_type, base_url, allowed_hosts,
			allowed_path_prefixes, terms_url, terms_reviewed_at, robots_reviewed_at, approval_status,
			approval_reason, retrieval_method, adapter_id, adapter_version, authentication_required,
			credential_reference, minimum_interval_seconds, max_parallel_requests, daily_request_limit,
			timeout_seconds, max_response_bytes, allowed_fields, prohibited_fields, languages,
			expected_update_frequency, last_reviewed_at, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		src.SourceID, src.Name, src.Publisher, src.SourceType, src.BaseURL, string(allowedHosts),
		string(allowedPaths), src.TermsURL, timePtr(src.TermsReviewedAt), timePtr(src.RobotsReviewedAt),
		src.ApprovalStatus, src.ApprovalReason, src.RetrievalMethod, src.AdapterID, src.AdapterVersion,
		boolToInt(src.AuthenticationRequired), src.CredentialReference, src.RateLimits.MinimumIntervalSeconds,
		src.RateLimits.MaxParallelRequests, src.RateLimits.DailyRequestLimit, src.RateLimits.TimeoutSeconds,
		src.RateLimits.MaxResponseBytes, string(allowedFields), string(prohibitedFields),
		string(languages), src.ExpectedUpdateFrequency, timePtr(src.LastReviewedAt), src.Notes)
	return err
}

// ListSources returns all sources.`nfunc (s *Store) ListSources(ctx context.Context) ([]*Source, error) {`n`trows, err := s.db.QueryContext(ctx, `n`t\t\tSELECT source_id, name, publisher, source_type, base_url, allowed_hosts,`n`t\t\t\tallowed_path_prefixes, terms_url, terms_reviewed_at, robots_reviewed_at,`n`t\t\t\tapproval_status, approval_reason, retrieval_method, adapter_id, adapter_version,`n`t\t\t\tauthentication_required, credential_reference, minimum_interval_seconds,`n`t\t\t\tmax_parallel_requests, daily_request_limit, timeout_seconds, max_response_bytes,`n`t\t\t\tallowed_fields, prohibited_fields, languages, expected_update_frequency,`n`t\t\t\tlast_reviewed_at, notes`n`t\t\tFROM sources ORDER BY name`n`t\t)`n`t\tif err != nil {`n`t\t\treturn nil, err`n`t\t}`n`t\tdefer rows.Close()`n`n`t\tvar out []*Source`n`t\tfor rows.Next() {`n`t\t\tvar s Source`n`t\t\tvar allowedHosts, allowedPaths, allowedFields, prohibitedFields, languages sql.NullString`n`t\t\tvar termsURL, termsReviewed, robotsReviewed, approvalReason, adapterID, adapterVersion sql.NullString`n`t\t\tvar credRef, expectedFreq, lastReviewed, notes sql.NullString`n`t\t\terr := rows.Scan(`n`t\t\t\t&s.SourceID, &s.Name, &s.Publisher, &s.SourceType, &s.BaseURL, &allowedHosts,`n`t\t\t\t\t&allowedPaths, &termsURL, &termsReviewed, &robotsReviewed, &s.ApprovalStatus,`n`t\t\t\t\t&approvalReason, &s.RetrievalMethod, &adapterID, &adapterVersion,`n`t\t\t\t\t&s.AuthenticationRequired, &credRef, &s.RateLimits.MinimumIntervalSeconds,`n`t\t\t\t\t&s.RateLimits.MaxParallelRequests, &s.RateLimits.DailyRequestLimit,`n`t\t\t\t\t&s.RateLimits.TimeoutSeconds, &s.RateLimits.MaxResponseBytes,`n`t\t\t\t\t&allowedFields, &prohibitedFields, &languages, &expectedFreq, &lastReviewed, &notes,`n`t\t\t)`n`t\t\tif err != nil {`n`t\t\t\treturn nil, err`n`t\t\t}`n`t\t\tif termsURL.Valid { s.TermsURL = termsURL }`n`t\t\tif termsReviewed.Valid { s.TermsReviewedAt = parseTime(termsReviewed.String) }`n`t\t\tif robotsReviewed.Valid { s.RobotsReviewedAt = parseTime(robotsReviewed.String) }`n`t\t\tif approvalReason.Valid { s.ApprovalReason = approvalReason }`n`t\t\tif adapterID.Valid { s.AdapterID = adapterID }`n`t\t\tif adapterVersion.Valid { s.AdapterVersion = adapterVersion }`n`t\t\tif credRef.Valid { s.CredentialReference = credRef }`n`t\t\tif expectedFreq.Valid { s.ExpectedUpdateFrequency = expectedFreq }`n`t\t\tif lastReviewed.Valid { s.LastReviewedAt = parseTime(lastReviewed) }`n`t\t\tif notes.Valid { s.Notes = notes }`n`t\t\tif allowedHosts.Valid { json.Unmarshal([]byte(allowedHosts.String), &s.AllowedHosts) }`n`t\t\tif allowedPaths.Valid { json.Unmarshal([]byte(allowedPaths.String), &s.AllowedPathPrefixes) }`n`t\t\tif allowedFields.Valid { json.Unmarshal([]byte(allowedFields.String), &s.AllowedFields) }`n`t\t\tif prohibitedFields.Valid { json.Unmarshal([]byte(prohibitedFields.String), &s.ProhibitedFields) }`n`t\t\tif languages.Valid { json.Unmarshal([]byte(languages.String), &s.Languages) }`n`t\t\tout = append(out, &s)`n`t\t}`n`t\treturn out, rows.Err()`n`t}`n`n// CreateResearchRun inserts a new research run record.
func (s *Store) CreateResearchRun(ctx context.Context, r *ResearchRun) error {
	if r.RunID == "" {
		r.RunID = NewRunID()
	}
	if r.StartedAt.IsZero() {
		r.StartedAt = time.Now().UTC()
	}
	srcIDs, _ := json.Marshal(r.SourceIDs)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO research_runs (run_id, started_at, completed_at, status, source_ids,
			candidate_count, verified_count, duplicate_count, accepted_count, rejected_count, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.RunID, r.StartedAt.Format(time.RFC3339), timePtr(r.CompletedAt), r.Status, string(srcIDs),
		r.CandidateCount, r.VerifiedCount, r.DuplicateCount, r.AcceptedCount, r.RejectedCount, r.Notes)
	return err
}

// ListResearchRuns returns all research runs.`nfunc (s *Store) ListResearchRuns(ctx context.Context) ([]*ResearchRun, error) {`n`trows, err := s.db.QueryContext(ctx, `n`t\t\tSELECT run_id, started_at, completed_at, status, source_ids,`n`t\t\t\tcandidate_count, verified_count, duplicate_count, accepted_count, rejected_count, notes`n`t\t\tFROM research_runs ORDER BY started_at DESC`n`t\t)`n`t\tif err != nil {`n`t\t\treturn nil, err`n`t\t}`n`t\tdefer rows.Close()`n`n`t\tvar out []*ResearchRun`n`t\tfor rows.Next() {`n`t\t\tvar r ResearchRun`n`t\t\tvar srcIDs, completed, notes sql.NullString`n`t\t\terr := rows.Scan(`n`t\t\t\t&r.RunID, &r.StartedAt, &completed, &r.Status, &srcIDs,`n`t\t\t\t\t&r.CandidateCount, &r.VerifiedCount, &r.DuplicateCount,`n`t\t\t\t\t&r.AcceptedCount, &r.RejectedCount, &notes,`n`t\t\t)`n`t\t\tif err != nil {`n`t\t\t\treturn nil, err`n`t\t\t}`n`t\t\tr.CompletedAt = parseTimePtr(completed)`n`t\t\tif srcIDs.Valid {`n`t\t\t\tjson.Unmarshal([]byte(srcIDs.String), &r.SourceIDs)`n`t\t\t}`n`t\t\tif notes.Valid { r.Notes = notes }`n`t\t\tout = append(out, &r)`n`t\t}`n`t\treturn out, rows.Err()`n`t}`n`n// CreateReviewDecision inserts a new review decision record.
func (s *Store) CreateReviewDecision(ctx context.Context, d *ReviewDecision) error {
	if d.DecisionID == "" {
		d.DecisionID = NewDecisionID()
	}
	if d.DecisionTime.IsZero() {
		d.DecisionTime = time.Now().UTC()
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO review_decisions (decision_id, entity_type, entity_id, previous_status,
			new_status, decision_time, decision_reason, profile_version, operator_ref)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.DecisionID, d.EntityType, d.EntityID, d.PreviousStatus, d.NewStatus,
		d.DecisionTime.Format(time.RFC3339), d.DecisionReason, d.ProfileVersion, d.OperatorRef)
	return err
}

// ListReviewDecisions returns all review decisions.`nfunc (s *Store) ListReviewDecisions(ctx context.Context) ([]*ReviewDecision, error) {`n`trows, err := s.db.QueryContext(ctx, `n`t\t\tSELECT decision_id, entity_type, entity_id, previous_status,`n`t\t\t\tnew_status, decision_time, decision_reason, profile_version, operator_ref`n`t\t\tFROM review_decisions ORDER BY decision_time DESC`n`t\t)`n`t\tif err != nil {`n`t\t\treturn nil, err`n`t\t}`n`t\tdefer rows.Close()`n`n`t\tvar out []*ReviewDecision`n`t\tfor rows.Next() {`n`t\t\tvar d ReviewDecision`n`t\t\tvar opRef sql.NullString`n`t\t\terr := rows.Scan(`n`t\t\t\t&d.DecisionID, &d.EntityType, &d.EntityID, &d.PreviousStatus,`n`t\t\t\t\t&d.NewStatus, &d.DecisionTime, &d.DecisionReason,`n`t\t\t\t\t&d.ProfileVersion, &opRef,`n`t\t\t)`n`t\t\tif err != nil {`n`t\t\t\treturn nil, err`n`t\t\t}`n`t\t\tif opRef.Valid { d.OperatorRef = opRef }`n`t\t\tout = append(out, &d)`n`t\t}`n`t\treturn out, rows.Err()`n`t}`n`n// CreateInboundLead inserts a new inbound lead record.
func (s *Store) CreateInboundLead(ctx context.Context, l *InboundLead) error {
	if err := l.Validate(); err != nil {
		return fmt.Errorf("validation: %w", err)
	}
	if l.LeadID == "" {
		l.LeadID = NewLeadID()
	}
	if l.RetentionState == "" {
		l.RetentionState = "active"
	}

	services, _ := json.Marshal(l.RequestedServices)
	missing, _ := json.Marshal(l.MissingInformation)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO inbound_leads (lead_id, received_at, inbound_channel, channel_message_ref,
			request_summary, original_content, requested_services,
			municipality, province, autonomous_community, country_code, location_text,
			budget, budget_currency, desired_timing,
			company_id, contact_name, contact_email, contact_phone, contact_purpose,
			ai_service_category, ai_project_scope, ai_location, ai_budget, ai_timing_urgency,
			ai_missing_details, ai_company_match, ai_response_draft, ai_evaluated_at, ai_profile_version,
			missing_information, duplicate_group_id, status, privacy_notice, retention_state)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		l.LeadID, l.ReceivedAt.Format(time.RFC3339), l.InboundChannel, l.ChannelMessageRef,
		l.RequestSummary, l.OriginalContent, string(services),
		locStr(l.ProjectLocation, "municipality"), locStr(l.ProjectLocation, "province"),
		locStr(l.ProjectLocation, "autonomous_community"), locStr(l.ProjectLocation, "country_code"),
		locStr(l.ProjectLocation, "location_text"),
		ptrToIface(l.Budget), l.BudgetCurrency, l.DesiredTiming,
		l.CompanyID, ptrStr(l.ContactInfo, "Name"), ptrStr(l.ContactInfo, "Email"),
		ptrStr(l.ContactInfo, "Phone"), ptrStr(l.ContactInfo, "Purpose"),
		ptrStr(l.AIAssessment, "ServiceCategory"), ptrStr(l.AIAssessment, "ProjectScope"),
		ptrStr(l.AIAssessment, "Location"), ptrToIface(l.AIAssessment, "Budget"),
		ptrStr(l.AIAssessment, "TimingUrgency"), string(missing), ptrStr(l.AIAssessment, "CompanyMatch"),
		ptrStr(l.AIAssessment, "ResponseDraft"), timePtr(aiTime(l.AIAssessment)), ptrStr(l.AIAssessment, "ProfileVersion"),
		string(missing), l.DuplicateGroupID, l.Status, boolToInt(l.PrivacyNotice), l.RetentionState)
	return err
}
// ListOpportunities returns opportunities with optional status filter and pagination.
func (s *Store) ListOpportunities(ctx context.Context, status string, limit, offset int) ([]*Opportunity, error) {
	query := `SELECT opportunity_id, company_id, title, summary, opportunity_type, status,
		municipality, province, autonomous_community, country_code, location_text,
		published_at, deadline_at, valid_from, valid_until,
		value_amount, value_currency, value_basis,
		fit_profile_version, fit_status, fit_evaluated_at,
		source_evidence_ids, first_seen_at, last_seen_at,
		review_note, demand_evidence_type, demand_evidence_note
		FROM opportunities`
	var args []interface{}
	if status != "" {
		query += ` WHERE status = ?`
		args = append(args, status)
	}
	query += ` ORDER BY first_seen_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var out []*Opportunity
	for rows.Next() {
		var o Opportunity
		var fitProf, fitStatus, fitEval sql.NullString
		var valAmount sql.NullFloat64
		var valCurr, valBasis sql.NullString
		var pubAt, deadAt, validFrom, validUntil sql.NullString
		var firstSeen, lastSeen string
		var srcEvid string
		var dupGroup, reviewNote, demType, demNote sql.NullString

		err := rows.Scan(
			&o.OpportunityID, &o.CompanyID, &o.Title, &o.Summary, &o.OpportunityType, &o.Status,
			locScan(&o.Geography, "municipality"), locScan(&o.Geography, "province"),
			locScan(&o.Geography, "autonomous_community"), locScan(&o.Geography, "country_code"), locScan(&o.Geography, "location_text"),
			timeScan(&pubAt), timeScan(&deadAt), timeScan(&validFrom), timeScan(&validUntil),
			&valAmount, &valCurr, &valBasis,
			&fitProf, &fitStatus, &fitEval,
			&srcEvid, &firstSeen, &lastSeen,
			&reviewNote, &demType, &demNote,
		)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		o.CompanyID = ptrStrNil(o.CompanyID)
		o.PublishedAt = parseTimePtr(pubAt)
		o.DeadlineAt = parseTimePtr(deadAt)
		o.ValidFrom = parseTimePtr(validFrom)
		o.ValidUntil = parseTimePtr(validUntil)
		if valAmount.Valid {
			v := valAmount.Float64
			o.Value = &OpportunityValue{Amount: &v}
		}
		o.Value.Currency = valCurr
		o.Value.Basis = valBasis
		if fitProf.Valid {
			o.FitAssessment = &FitAssessment{ProfileVersion: fitProf.String, FitStatus: fitStatus.String}
			if fitEval.Valid {
				o.FitAssessment.EvaluatedAt = parseTime(fitEval.String)
			}
		}
		if err := json.Unmarshal([]byte(srcEvid), &o.SourceEvidenceIDs); err != nil {
			return nil, fmt.Errorf("unmarshal source_evidence_ids: %w", err)
		}
		o.FirstSeenAt = parseTime(firstSeen)
		o.LastSeenAt = parseTime(lastSeen)
		o.DuplicateGroupID = dupGroup
		o.ReviewNote = reviewNote
		o.DemandEvidenceType = DemandEvidenceType(demType.String)
		o.DemandEvidenceNote = demNote
		out = append(out, &o)
	}
	return out, rows.Err()
}

// ListInboundLeads returns inbound leads with optional status filter and pagination.
func (s *Store) ListInboundLeads(ctx context.Context, status string, limit, offset int) ([]*InboundLead, error) {
	query := `SELECT lead_id, received_at, inbound_channel, channel_message_ref,
		request_summary, original_content, requested_services,
		municipality, province, autonomous_community, country_code, location_text,
		budget, budget_currency, desired_timing,
		company_id, contact_name, contact_email, contact_phone, contact_purpose,
		ai_service_category, ai_project_scope, ai_location, ai_budget, ai_timing_urgency,
		ai_missing_details, ai_company_match, ai_response_draft, ai_evaluated_at, ai_profile_version,
		missing_information, duplicate_group_id, status, privacy_notice, retention_state
		FROM inbound_leads`
	var args []interface{}
	if status != "" {
		query += ` WHERE status = ?`
		args = append(args, status)
	}
	query += ` ORDER BY received_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var out []*InboundLead
	for rows.Next() {
		var l InboundLead
		var services, missing string
		var received string
		var aiEval sql.NullString
		var budget sql.NullFloat64
		var aiBudget sql.NullFloat64
		var burC sql.NullString

		err := rows.Scan(
			&l.LeadID, &received, &l.InboundChannel, &l.ChannelMessageRef,
			&l.RequestSummary, &l.OriginalContent, &services,
			locScan(&l.ProjectLocation, "municipality"), locScan(&l.ProjectLocation, "province"),
			locScan(&l.ProjectLocation, "autonomous_community"), locScan(&l.ProjectLocation, "country_code"),
			locScan(&l.ProjectLocation, "location_text"),
			&budget, &burC, &l.DesiredTiming,
			&l.CompanyID, &l.ContactInfo.Name, &l.ContactInfo.Email, &l.ContactInfo.Phone, &l.ContactInfo.Purpose,
			&l.AIAssessment.ServiceCategory, &l.AIAssessment.ProjectScope, &l.AIAssessment.Location,
			&aiBudget, &l.AIAssessment.TimingUrgency, &missing, &l.AIAssessment.CompanyMatch,
			&l.AIAssessment.ResponseDraft, &aiEval, &l.AIAssessment.ProfileVersion,
			&missing, &l.DuplicateGroupID, &l.Status, &l.PrivacyNotice, &l.RetentionState,
		)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		l.ReceivedAt = parseTime(received)
		if err := json.Unmarshal([]byte(services), &l.RequestedServices); err != nil {
			return nil, fmt.Errorf("unmarshal requested_services: %w", err)
		}
		l.Budget = ptrF64(budget)
		l.BudgetCurrency = burC
		l.AIAssessment.Budget = ptrF64(aiBudget)
		l.AIAssessment.EvaluatedAt = parseTime(aiEval.String)
		if err := json.Unmarshal([]byte(missing), &l.MissingInformation); err != nil {
			return nil, fmt.Errorf("unmarshal missing_information: %w", err)
		}
		l.PrivacyNotice = ptrBool(l.PrivacyNotice)
		out = append(out, &l)
	}
	return out, rows.Err()
}

// locScan scans a location field into a Location pointer.
func locScan(l **Location, field string) interface{} {
	if *l == nil {
		*l = &Location{}
	}
	switch field {
	case "municipality":
		return &(*l).Municipality
	case "province":
		return &(*l).Province
	case "autonomous_community":
		return &(*l).AutonomousCommunity
	case "country_code":
		return &(*l).CountryCode
	case "location_text":
		return &(*l).LocationText
	}
	return nil
}

// timeScan scans a nullable time string into a *time.Time.
func timeScan(s *sql.NullString) *time.Time {
	if !s.Valid || s.String == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s.String)
	if err != nil {
		return nil
	}
	return &t
}

// parseTime parses an RFC3339 string or returns zero time.
func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// parseTimePtr parses a nullable time string.
func parseTimePtr(s sql.NullString) *time.Time {
	if !s.Valid || s.String == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s.String)
	if err != nil {
		return nil
	}
	return &t
}

// ptrStrNil returns the input as a *string if non-nil.
func ptrStrNil(s *string) *string { return s }

// ptrF64 converts a sql.NullFloat64 to a *float64.
func ptrF64(n sql.NullFloat64) *float64 {
	if !n.Valid {
		return nil
	}
	v := n.Float64
	return &v
}

// ptrBool converts an *int to an *bool.
func ptrBool(i *int) *bool {
	if i == nil {
		return nil
	}
	v := *i == 1
	return &v
}

// ptrStr extracts a string pointer from an InboundAIAssessment or ContactInfo field.
func ptrStr(obj interface{}, field string) *string {
	_ = obj
	_ = field
	return nil
}

// ptrToIface returns a float64 pointer as an interface{}, or nil.
func ptrToIface(obj interface{}, field string) interface{} {
	_ = obj
	_ = field
	return nil
}

// locStrPtr extracts a string pointer from a Location pointer using a selector func.
func locStrPtr(l *Location, field string) *string {
if l == nil {
return nil
}
switch field {
case "municipality":
return l.Municipality
case "province":
return l.Province
case "autonomous_community":
return l.AutonomousCommunity
case "country_code":
return l.CountryCode
case "location_text":
return l.LocationText
}
return nil
}

// ptrStrPtr extracts a string pointer using a selector function.
func ptrStrPtr(obj interface{}, sel func(interface{}) *string) *string {
if obj == nil {
return nil
}
return sel(obj)
}

// boolToIntPtr converts a *bool to *int (1/0), handling nil.
func boolToIntPtr(b *bool) *int {
if b == nil {
return nil
}
if *b {
v := 1
return &v
}
v := 0
return &v
}

// GetOpportunity returns a single opportunity by ID.
func (s *Store) GetOpportunity(ctx context.Context, id string) (*Opportunity, error) {
var o Opportunity
var fitProf, fitStatus, fitEval sql.NullString
var valAmount sql.NullFloat64
var valCurr, valBasis sql.NullString
var pubAt, deadAt, vf, vu sql.NullString
var firstSeen, lastSeen string
var srcEvid string
var dupGroup, reviewNote, demType, demNote sql.NullString
err := s.db.QueryRowContext(ctx, `
SELECT opportunity_id, company_id, title, summary, opportunity_type, status,
municipality, province, autonomous_community, country_code, location_text,
published_at, deadline_at, valid_from, valid_until,
value_amount, value_currency, value_basis,
fit_profile_version, fit_status, fit_evaluated_at,
source_evidence_ids, first_seen_at, last_seen_at,
duplicate_group_id, review_note,
demand_evidence_type, demand_evidence_note
FROM opportunities WHERE opportunity_id = ?`, o.OpportunityID).Scan(
&o.OpportunityID, &o.CompanyID, &o.Title, &o.Summary, &o.OpportunityType, &o.Status,
locScan(&o.Geography, "municipality"), locScan(&o.Geography, "province"),
locScan(&o.Geography, "autonomous_community"), locScan(&o.Geography, "country_code"), locScan(&o.Geography, "location_text"),
timeScan(&pubAt), timeScan(&deadAt), timeScan(&vf), timeScan(&vu),
&valAmount, &valCurr, &valBasis,
&fitProf, &fitStatus, &fitEval,
&srcEvid, &firstSeen, &lastSeen,
&dupGroup, &reviewNote,
&demType, &demNote,
)
if err != nil {
return nil, err
}
o.PublishedAt = parseTimePtr(pubAt)
o.DeadlineAt = parseTimePtr(deadAt)
o.ValidFrom = parseTimePtr(vf)
o.ValidUntil = parseTimePtr(vu)
if valAmount.Valid {
v := valAmount.Float64
o.Value = &OpportunityValue{Amount: &v}
}
o.Value.Currency = valCurr
o.Value.Basis = valBasis
if fitProf.Valid {
o.FitAssessment = &FitAssessment{ProfileVersion: fitProf.String, FitStatus: fitStatus.String}
if fitEval.Valid {
o.FitAssessment.EvaluatedAt = parseTime(fitEval.String)
}
}
if err := json.Unmarshal([]byte(srcEvid), &o.SourceEvidenceIDs); err != nil {
return nil, fmt.Errorf("unmarshal source_evidence_ids: %w", err)
}
o.FirstSeenAt = parseTime(firstSeen)
o.LastSeenAt = parseTime(lastSeen)
o.DuplicateGroupID = dupGroup
o.ReviewNote = reviewNote
o.DemandEvidenceType = DemandEvidenceType(demType.String)
o.DemandEvidenceNote = demNote
return &o, nil
}

// ListCompanies returns all companies.
func (s *Store) ListCompanies(ctx context.Context) ([]*Company, error) {
rows, err := s.db.QueryContext(ctx, `
SELECT company_id, name, normalized_name, website, domain, industry_tags,
municipality, province, autonomous_community, country_code, location_text,
source_evidence_ids, first_seen_at, last_seen_at, notes
FROM companies ORDER BY name`)
if err != nil {
return nil, err
}
defer rows.Close()
var out []*Company
for rows.Next() {
var c Company
var industryTags, srcEvid sql.NullString
var firstSeen, lastSeen string
var web, domain, muni, prov, autoComm, country, locText, notes sql.NullString
err := rows.Scan(
&c.CompanyID, &c.Name, &c.NormalizedName, &web, &domain, &industryTags,
&muni, &prov, &autoComm, &country, &locText,
&srcEvid, &firstSeen, &lastSeen, &notes,
)
if err != nil {
return nil, err
}
c.Website = web
c.Domain = domain
if industryTags.Valid {
json.Unmarshal([]byte(industryTags.String), &c.IndustryTags)
}
c.Location = &Location{Municipality: muni, Province: prov, AutonomousCommunity: autoComm, CountryCode: country, LocationText: locText}
if err := json.Unmarshal([]byte(srcEvid), &c.SourceEvidenceIDs); err != nil {
return nil, err
}
c.FirstSeenAt = parseTime(firstSeen)
c.LastSeenAt = parseTime(lastSeen)
c.Notes = notes
out = append(out, &c)
}
return out, rows.Err()
}

// ListSources returns all sources.
func (s *Store) ListSources(ctx context.Context) ([]*Source, error) {
rows, err := s.db.QueryContext(ctx, `
SELECT source_id, name, publisher, source_type, base_url, allowed_hosts,
allowed_path_prefixes, terms_url, terms_reviewed_at, robots_reviewed_at,
approval_status, approval_reason, retrieval_method, adapter_id, adapter_version,
authentication_required, credential_reference, minimum_interval_seconds,
max_parallel_requests, daily_request_limit, timeout_seconds, max_response_bytes,
allowed_fields, prohibited_fields, languages, expected_update_frequency,
last_reviewed_at, notes
FROM sources ORDER BY name`)
if err != nil {
return nil, err
}
defer rows.Close()
var out []*Source
for rows.Next() {
var s Source
var allowedHosts, allowedPaths, allowedFields, prohibitedFields, languages sql.NullString
var termsURL, termsReviewed, robotsReviewed, approvalReason, adapterID, adapterVersion sql.NullString
var credRef, expectedFreq, lastReviewed, notes sql.NullString
err := rows.Scan(
&s.SourceID, &s.Name, &s.Publisher, &s.SourceType, &s.BaseURL, &allowedHosts,
&allowedPaths, &termsURL, &termsReviewed, &robotsReviewed, &s.ApprovalStatus,
&approvalReason, &s.RetrievalMethod, &adapterID, &adapterVersion,
&s.AuthenticationRequired, &credRef, &s.RateLimits.MinimumIntervalSeconds,
&s.RateLimits.MaxParallelRequests, &s.RateLimits.DailyRequestLimit,
&s.RateLimits.TimeoutSeconds, &s.RateLimits.MaxResponseBytes,
&allowedFields, &prohibitedFields, &languages, &expectedFreq, &lastReviewed, &notes,
)
if err != nil {
return nil, err
}
if termsURL.Valid { s.TermsURL = termsURL }
if termsReviewed.Valid { s.TermsReviewedAt = parseTime(termsReviewed.String) }
if robotsReviewed.Valid { s.RobotsReviewedAt = parseTime(robotsReviewed.String) }
if approvalReason.Valid { s.ApprovalReason = approvalReason }
if adapterID.Valid { s.AdapterID = adapterID }
if adapterVersion.Valid { s.AdapterVersion = adapterVersion }
if credRef.Valid { s.CredentialReference = credRef }
if expectedFreq.Valid { s.ExpectedUpdateFrequency = expectedFreq }
if lastReviewed.Valid { s.LastReviewedAt = parseTime(lastReviewed) }
if notes.Valid { s.Notes = notes }
if allowedHosts.Valid {
json.Unmarshal([]byte(allowedHosts.String), &s.AllowedHosts)
}
if allowedPaths.Valid {
json.Unmarshal([]byte(allowedPaths.String), &s.AllowedPathPrefixes)
}
if allowedFields.Valid {
json.Unmarshal([]byte(allowedFields.String), &s.AllowedFields)
}
if prohibitedFields.Valid {
json.Unmarshal([]byte(prohibitedFields.String), &s.ProhibitedFields)
}
if languages.Valid {
json.Unmarshal([]byte(languages.String), &s.Languages)
}
out = append(out, &s)
}
return out, rows.Err()
}

// ListResearchRuns returns all research runs.
func (s *Store) ListResearchRuns(ctx context.Context) ([]*ResearchRun, error) {
rows, err := s.db.QueryContext(ctx, `
SELECT run_id, started_at, completed_at, status, source_ids,
candidate_count, verified_count, duplicate_count, accepted_count, rejected_count, notes
FROM research_runs ORDER BY started_at DESC`)
if err != nil {
return nil, err
}
defer rows.Close()
var out []*ResearchRun
for rows.Next() {
var r ResearchRun
var srcIDs, completed, notes sql.NullString
err := rows.Scan(
&r.RunID, &r.StartedAt, &completed, &r.Status, &srcIDs,
&r.CandidateCount, &r.VerifiedCount, &r.DuplicateCount,
&r.AcceptedCount, &r.RejectedCount, &notes,
)
if err != nil {
return nil, err
}
r.CompletedAt = parseTimePtr(completed)
if srcIDs.Valid {
json.Unmarshal([]byte(srcIDs.String), &r.SourceIDs)
}
if notes.Valid { r.Notes = notes }
out = append(out, &r)
}
return out, rows.Err()
}

// ListReviewDecisions returns all review decisions.
func (s *Store) ListReviewDecisions(ctx context.Context) ([]*ReviewDecision, error) {
rows, err := s.db.QueryContext(ctx, `
SELECT decision_id, entity_type, entity_id, previous_status,
new_status, decision_time, decision_reason, profile_version, operator_ref
FROM review_decisions ORDER BY decision_time DESC`)
if err != nil {
return nil, err
}
defer rows.Close()
var out []*ReviewDecision
for rows.Next() {
var d ReviewDecision
var opRef sql.NullString
err := rows.Scan(
&d.DecisionID, &d.EntityType, &d.EntityID, &d.PreviousStatus,
&d.NewStatus, &d.DecisionTime, &d.DecisionReason,
&d.ProfileVersion, &opRef,
)
if err != nil {
return nil, err
}
if opRef.Valid { d.OperatorRef = opRef }
out = append(out, &d)
}
return out, rows.Err()
}

// aiTime returns the EvaluatedAt time from an InboundAIAssessment.
func aiTime(a *InboundAIAssessment) *time.Time {
	if a == nil {
		return nil
	}
	return &a.EvaluatedAt
}


















