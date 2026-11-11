package domain

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"reflect"

	// modernc.org/sqlite is the pure-Go SQLite driver used by Store.
	_ "modernc.org/sqlite"
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
	if o.OpportunityID == "" {
		o.OpportunityID = NewOpportunityID()
	}
	if err := o.Validate(); err != nil {
		return fmt.Errorf("validation: %w", err)
	}
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

	_, err = s.db.ExecContext(ctx, `
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

	err := s.db.QueryRowContext(ctx, "SELECT opportunity_id, company_id, title, summary, opportunity_type, status, municipality, province, autonomous_community, country_code, location_text, published_at, deadline_at, valid_from, valid_until, value_amount, value_currency, value_basis, fit_profile_version, fit_status, fit_evaluated_at, source_evidence_ids, first_seen_at, last_seen_at, duplicate_group_id, review_note, demand_evidence_type, demand_evidence_note FROM opportunities WHERE opportunity_id = ?", id).Scan(
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
	if o.Value != nil {
		if valCurr.Valid {
			o.Value.Currency = &valCurr.String
		}
		if valBasis.Valid {
			o.Value.Basis = &valBasis.String
		}
	}
	if fitProf.Valid {
		o.FitAssessment = &FitAssessment{ProfileVersion: fitProf.String, FitStatus: fitStatus.String}
		if fitEval.Valid {
			o.FitAssessment.EvaluatedAt = parseTime(fitEval.String)
		}
	}
	if err := json.Unmarshal([]byte(srcEvid.String), &o.SourceEvidenceIDs); err != nil {
		return nil, fmt.Errorf("unmarshal source_evidence_ids: %w", err)
	}
	o.FirstSeenAt = parseTime(firstSeen)
	o.LastSeenAt = parseTime(lastSeen)
	if dupGroup.Valid {
		o.DuplicateGroupID = ptrStrFromNull(dupGroup)
	}
	if reviewNote.Valid {
		o.ReviewNote = ptrStrFromNull(reviewNote)
	}
	if demType.Valid {
		o.DemandEvidenceType = DemandEvidenceType(demType.String)
	}
	if demNote.Valid {
		o.DemandEvidenceNote = ptrStrFromNull(demNote)
	}
	return &o, nil
}
// CreateSourceEvidence inserts a new source evidence record.
func (s *Store) CreateSourceEvidence(ctx context.Context, e *SourceEvidence) error {
	if e.EvidenceID == "" {
		e.EvidenceID = NewEvidenceID()
	}
	if e.ObservedAt.IsZero() {
		e.ObservedAt = time.Now().UTC()
	}
	facts, err := json.Marshal(e.Facts)
		if err != nil {
			return fmt.Errorf("marshal facts: %w", err)
		}

	_, err = s.db.ExecContext(ctx, `
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

	_, err = s.db.ExecContext(ctx, `
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
		boolToInt(&src.AuthenticationRequired), src.CredentialReference, src.RateLimits.MinimumIntervalSeconds,
		src.RateLimits.MaxParallelRequests, src.RateLimits.DailyRequestLimit, src.RateLimits.TimeoutSeconds,
		src.RateLimits.MaxResponseBytes, string(allowedFields), string(prohibitedFields),
		string(languages), src.ExpectedUpdateFrequency, timePtr(src.LastReviewedAt), src.Notes)
	return err
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
		FROM sources ORDER BY name
	`)
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
		if termsURL.Valid { s.TermsURL = ptrStrFromNull(termsURL) }
		if termsReviewed.Valid { s.TermsReviewedAt = parseTimePtr(termsReviewed) }
		if robotsReviewed.Valid { s.RobotsReviewedAt = parseTimePtr(robotsReviewed) }
		if approvalReason.Valid { s.ApprovalReason = ptrStrFromNull(approvalReason) }
		if adapterID.Valid { s.AdapterID = ptrStrFromNull(adapterID) }
		if adapterVersion.Valid { s.AdapterVersion = ptrStrFromNull(adapterVersion) }
		if credRef.Valid { s.CredentialReference = ptrStrFromNull(credRef) }
		if expectedFreq.Valid { s.ExpectedUpdateFrequency = ptrStrFromNull(expectedFreq) }
		if lastReviewed.Valid { s.LastReviewedAt = parseTimePtr(lastReviewed) }
		if notes.Valid { s.Notes = ptrStrFromNull(notes) }

		// Parse JSON arrays
		if allowedHosts.Valid {
			if err := json.Unmarshal([]byte(allowedHosts.String), &s.AllowedHosts); err != nil {
				return nil, fmt.Errorf("unmarshal allowed_hosts: %w", err)
			}
		}
		if allowedPaths.Valid {
			if err := json.Unmarshal([]byte(allowedPaths.String), &s.AllowedPathPrefixes); err != nil {
				return nil, fmt.Errorf("unmarshal allowed_path_prefixes: %w", err)
			}
		}
		if allowedFields.Valid {
			if err := json.Unmarshal([]byte(allowedFields.String), &s.AllowedFields); err != nil {
				return nil, fmt.Errorf("unmarshal allowed_fields: %w", err)
			}
		}
		if prohibitedFields.Valid {
			if err := json.Unmarshal([]byte(prohibitedFields.String), &s.ProhibitedFields); err != nil {
				return nil, fmt.Errorf("unmarshal prohibited_fields: %w", err)
			}
		}
		if languages.Valid {
			if err := json.Unmarshal([]byte(languages.String), &s.Languages); err != nil {
				return nil, fmt.Errorf("unmarshal languages: %w", err)
			}
		}

		out = append(out, &s)
	}
	return out, rows.Err()
}

func (s *Store) CreateResearchRun(ctx context.Context, r *ResearchRun) error {
	if r.RunID == "" {
		r.RunID = NewRunID()
	}
	if r.StartedAt.IsZero() {
		r.StartedAt = time.Now().UTC()
	}
	srcIDs, _ := json.Marshal(r.SourceIDs)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO research_runs (run_id, started_at, completed_at, status, source_ids,
			candidate_count, verified_count, duplicate_count, accepted_count, rejected_count, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.RunID, r.StartedAt.Format(time.RFC3339), timePtr(r.CompletedAt), r.Status, string(srcIDs),
		r.CandidateCount, r.VerifiedCount, r.DuplicateCount, r.AcceptedCount, r.RejectedCount, r.Notes)
	return err
}

// ListResearchRuns returns all research runs.
func (s *Store) ListResearchRuns(ctx context.Context) ([]*ResearchRun, error) {
	rows, err := s.db.QueryContext(ctx, `
			SELECT run_id, started_at, completed_at, status, source_ids,
				candidate_count, verified_count, duplicate_count, accepted_count, rejected_count, notes
			FROM research_runs ORDER BY started_at DESC
		`)
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
			if err := json.Unmarshal([]byte(srcIDs.String), &r.SourceIDs); err != nil {
				return nil, fmt.Errorf("unmarshal source_ids: %w", err)
			}
		}
		if notes.Valid {
			r.Notes = ptrStrFromNull(notes)
		}
		out = append(out, &r)
	}
	return out, rows.Err()
}

// validTransition checks whether a status transition is allowed for an entity type.
func validTransition(entityType, prevStatus, newStatus string) bool {
	switch entityType {
	case "opportunity":
		switch prevStatus {
		case "candidate":
			return newStatus == "verified" || newStatus == "needs_review" || newStatus == "rejected" || newStatus == "archived"
		case "verified":
			return newStatus == "needs_review" || newStatus == "qualified" || newStatus == "rejected" || newStatus == "expired" || newStatus == "archived"
		case "needs_review":
			return newStatus == "qualified" || newStatus == "rejected" || newStatus == "archived"
		case "qualified":
			return newStatus == "expired" || newStatus == "archived"
		default:
			return newStatus == "archived"
		}
	case "inbound_lead":
		switch prevStatus {
		case "new":
			return newStatus == "triage" || newStatus == "needs_information" || newStatus == "spam" || newStatus == "archived"
		case "triage":
			return newStatus == "needs_information" || newStatus == "qualified" || newStatus == "declined" || newStatus == "spam" || newStatus == "archived"
		case "needs_information":
			return newStatus == "triage" || newStatus == "qualified" || newStatus == "declined" || newStatus == "archived"
		case "qualified":
			return newStatus == "converted" || newStatus == "declined" || newStatus == "archived"
		case "converted", "declined", "spam":
			return newStatus == "archived"
		default:
			return newStatus == "archived"
		}
	case "company":
		return newStatus == "active" || newStatus == "archived"
	default:
		return false
	}
}

// CreateReviewDecision inserts a new review decision record.
func (s *Store) CreateReviewDecision(ctx context.Context, d *ReviewDecision) error {
        if d.DecisionID == "" {
                d.DecisionID = NewDecisionID()
        }
        if d.DecisionTime.IsZero() {
                d.DecisionTime = time.Now().UTC()
        }

        tx, err := s.db.BeginTx(ctx, nil)
        if err != nil {
                return fmt.Errorf("begin transaction: %w", err)
        }
        defer tx.Rollback()

        // Verify entity exists and read previous status.
        var prevStatus string
        switch d.EntityType {
        case "opportunity":
                err = tx.QueryRowContext(ctx, "SELECT status FROM opportunities WHERE opportunity_id = ?", d.EntityID).Scan(&prevStatus)
        case "inbound_lead":
                err = tx.QueryRowContext(ctx, "SELECT status FROM inbound_leads WHERE lead_id = ?", d.EntityID).Scan(&prevStatus)
        case "company":
                err = tx.QueryRowContext(ctx, "SELECT status FROM companies WHERE company_id = ?", d.EntityID).Scan(&prevStatus)
        default:
                return fmt.Errorf("unsupported entity type: %q", d.EntityType)
        }
        if err != nil {
                return fmt.Errorf("entity not found: %w", err)
        }
        d.PreviousStatus = prevStatus

        // Validate status transition.
        if !validTransition(d.EntityType, prevStatus, d.NewStatus) {
                return fmt.Errorf("invalid status transition for %s: %q -> %q", d.EntityType, prevStatus, d.NewStatus)
        }

        // Write the review decision.
        _, err = tx.ExecContext(ctx, `
                INSERT INTO review_decisions (decision_id, entity_type, entity_id, previous_status,
                        new_status, decision_time, decision_reason, profile_version, operator_ref)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
                d.DecisionID, d.EntityType, d.EntityID, d.PreviousStatus, d.NewStatus,
                d.DecisionTime.Format(time.RFC3339), d.DecisionReason, d.ProfileVersion, d.OperatorRef)
        if err != nil {
                return fmt.Errorf("insert review decision: %w", err)
        }

        // Change the entity status.
        switch d.EntityType {
        case "opportunity":
                _, err = tx.ExecContext(ctx, "UPDATE opportunities SET status = ? WHERE opportunity_id = ?", d.NewStatus, d.EntityID)
        case "inbound_lead":
                _, err = tx.ExecContext(ctx, "UPDATE inbound_leads SET status = ? WHERE lead_id = ?", d.NewStatus, d.EntityID)
        case "company":
                _, err = tx.ExecContext(ctx, "UPDATE companies SET status = ? WHERE company_id = ?", d.NewStatus, d.EntityID)
        }
        if err != nil {
                return fmt.Errorf("update entity status: %w", err)
        }

        if err := tx.Commit(); err != nil {
                return fmt.Errorf("commit transaction: %w", err)
        }
        return nil
}

// ListReviewDecisions returns all review decisions.
func (s *Store) ListReviewDecisions(ctx context.Context) ([]*ReviewDecision, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT decision_id, entity_type, entity_id, previous_status, new_status, decision_time, decision_reason, profile_version, operator_ref FROM review_decisions ORDER BY decision_time DESC")
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
		if opRef.Valid {
			d.OperatorRef = ptrStrFromNull(opRef)
		}
		out = append(out, &d)
	}
	return out, rows.Err()
}
// CreateInboundLead inserts a new inbound lead record.
func (s *Store) CreateInboundLead(ctx context.Context, l *InboundLead) error {
	if l.LeadID == "" {
		l.LeadID = NewLeadID()
	}
	// Apply backend defaults before validation so the frontend can submit
	// only request_summary and inbound_channel.
	if l.Status == "" {
		l.Status = InboundLeadStatusNew
	}
	if l.ReceivedAt.IsZero() {
		l.ReceivedAt = time.Now().UTC()
	}
	if err := l.Validate(); err != nil {
		return fmt.Errorf("validation: %w", err)
	}
	if l.RetentionState == "" {
		l.RetentionState = "active"
	}

	services, err := json.Marshal(l.RequestedServices)
	if err != nil {
		return fmt.Errorf("marshal requested services: %w", err)
	}
	missingInfoJSON, err := json.Marshal(l.MissingInformation)
	if err != nil {
		return fmt.Errorf("marshal missing information: %w", err)
	}
	var factsJSON []byte
	if l.AIAssessment != nil {
		factsJSON, err = json.Marshal(l.AIAssessment.MissingDetails)
		if err != nil {
			return fmt.Errorf("marshal ai missing details: %w", err)
		}
	}

	_, err = s.db.ExecContext(ctx, `
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
		l.Budget, l.BudgetCurrency, l.DesiredTiming,
		l.CompanyID, ptrStr(l.ContactInfo, "Name"), ptrStr(l.ContactInfo, "Email"),
		ptrStr(l.ContactInfo, "Phone"), ptrStr(l.ContactInfo, "Purpose"),
		ptrStr(l.AIAssessment, "ServiceCategory"), ptrStr(l.AIAssessment, "ProjectScope"),
		ptrStr(l.AIAssessment, "Location"), ptrF64Obj(l.AIAssessment, "Budget"),
		ptrStr(l.AIAssessment, "TimingUrgency"), string(factsJSON), ptrStr(l.AIAssessment, "CompanyMatch"),
		ptrStr(l.AIAssessment, "ResponseDraft"), timePtr(aiTime(l.AIAssessment)), ptrStr(l.AIAssessment, "ProfileVersion"),
		string(missingInfoJSON), l.DuplicateGroupID, l.Status, boolToInt(l.PrivacyNotice), l.RetentionState)
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
		// o.CompanyID is already string; ptrStrNil was a no-op identity cast
		o.PublishedAt = parseTimePtr(pubAt)
		o.DeadlineAt = parseTimePtr(deadAt)
		o.ValidFrom = parseTimePtr(validFrom)
		o.ValidUntil = parseTimePtr(validUntil)
		if valAmount.Valid || valCurr.Valid || valBasis.Valid {
			o.Value = &OpportunityValue{}
			if valAmount.Valid {
				v := valAmount.Float64
				o.Value.Amount = &v
			}
			if valCurr.Valid {
				o.Value.Currency = &valCurr.String
			}
			if valBasis.Valid {
				o.Value.Basis = &valBasis.String
			}
		}
		if fitProf.Valid {
			o.FitAssessment = &FitAssessment{ProfileVersion: fitProf.String, FitStatus: fitStatus.String}
			if fitEval.Valid {
				o.FitAssessment.EvaluatedAt = parseTime(fitEval.String)
			}
		}
		if err := json.Unmarshal([]byte(srcEvid.String), &o.SourceEvidenceIDs); err != nil {
			return nil, fmt.Errorf("unmarshal source_evidence_ids: %w", err)
		}
		o.FirstSeenAt = parseTime(firstSeen)
		o.LastSeenAt = parseTime(lastSeen)
		o.DuplicateGroupID = ptrStrFromNull(dupGroup)
		o.ReviewNote = ptrStrFromNull(reviewNote)
		o.DemandEvidenceType = DemandEvidenceType(demType.String)
		o.DemandEvidenceNote = ptrStrFromNull(demNote)
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
		var services, aiMissingDetails, missingInfo string
		var received string
		var aiEval sql.NullString
		var budget sql.NullFloat64
		var aiBudget sql.NullFloat64
		var burC sql.NullString

				var contactName, contactEmail, contactPhone, contactPurpose sql.NullString
		var aiServiceCat, aiProjectScope, aiLocationVal, aiTimingUrgency, aiCompanyMatch, aiResponseDraft, aiProfileVersion sql.NullString
		err := rows.Scan(
				&l.LeadID, &received, &l.InboundChannel, &l.ChannelMessageRef,
				&l.RequestSummary, &l.OriginalContent, &services,
				locScan(&l.ProjectLocation, "municipality"), locScan(&l.ProjectLocation, "province"),
				locScan(&l.ProjectLocation, "autonomous_community"), locScan(&l.ProjectLocation, "country_code"),
				locScan(&l.ProjectLocation, "location_text"),
				&budget, &burC, &l.DesiredTiming,
				&l.CompanyID, &contactName, &contactEmail, &contactPhone, &contactPurpose,
				&aiServiceCat, &aiProjectScope, &aiLocationVal, &aiBudget, &aiTimingUrgency, &aiMissingDetails,
				&aiCompanyMatch, &aiResponseDraft, &aiEval, &aiProfileVersion,
				&missingInfo, &l.DuplicateGroupID, &l.Status, &l.PrivacyNotice, &l.RetentionState,
			)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
				l.ReceivedAt = parseTime(received)
				if err := json.Unmarshal([]byte(services), &l.RequestedServices); err != nil {
					return nil, fmt.Errorf("unmarshal requested_services: %w", err)
				}
				l.Budget = ptrF64(budget)
				if burC.Valid {
					l.BudgetCurrency = &burC.String
				}

				// Rebuild nested structs only when data is present.
				if contactName.Valid || contactEmail.Valid || contactPhone.Valid || contactPurpose.Valid {
					l.ContactInfo = &ContactInfo{}
					if contactName.Valid {
						l.ContactInfo.Name = &contactName.String
					}
					if contactEmail.Valid {
						l.ContactInfo.Email = &contactEmail.String
					}
					if contactPhone.Valid {
						l.ContactInfo.Phone = &contactPhone.String
					}
					if contactPurpose.Valid {
						l.ContactInfo.Purpose = &contactPurpose.String
					}
				}
				if aiServiceCat.Valid || aiProjectScope.Valid || aiLocationVal.Valid || aiTimingUrgency.Valid || aiCompanyMatch.Valid || aiResponseDraft.Valid || aiProfileVersion.Valid || aiBudget.Valid || aiEval.Valid || aiMissingDetails != "" {
					l.AIAssessment = &InboundAIAssessment{}
					if aiServiceCat.Valid {
						l.AIAssessment.ServiceCategory = &aiServiceCat.String
					}
					if aiProjectScope.Valid {
						l.AIAssessment.ProjectScope = &aiProjectScope.String
					}
					if aiLocationVal.Valid {
						l.AIAssessment.Location = &aiLocationVal.String
					}
					if aiBudget.Valid {
						v := aiBudget.Float64
						l.AIAssessment.Budget = &v
					}
					if aiTimingUrgency.Valid {
						l.AIAssessment.TimingUrgency = &aiTimingUrgency.String
					}
					if aiCompanyMatch.Valid {
						l.AIAssessment.CompanyMatch = &aiCompanyMatch.String
					}
					if aiResponseDraft.Valid {
						l.AIAssessment.ResponseDraft = &aiResponseDraft.String
					}
					if aiProfileVersion.Valid {
						l.AIAssessment.ProfileVersion = aiProfileVersion.String
					}
					if aiEval.Valid {
						l.AIAssessment.EvaluatedAt = parseTime(aiEval.String)
					}
					if aiMissingDetails != "" {
						var missing []string
						if err := json.Unmarshal([]byte(aiMissingDetails), &missing); err != nil {
							return nil, fmt.Errorf("unmarshal ai missing details: %w", err)
						}
						l.AIAssessment.MissingDetails = missing
					}
				}
				if err := json.Unmarshal([]byte(missingInfo), &l.MissingInformation); err != nil {
					return nil, fmt.Errorf("unmarshal missing_information: %w", err)
				}
				// PrivacyNotice is already *bool, no conversion needed
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


// ptrF64Obj extracts a *float64 field from a struct by name using reflection.
// Returns nil if obj is nil, field is missing, or the field is not a *float64.
func ptrF64Obj(obj interface{}, field string) *float64 {
	if obj == nil {
		return nil
	}
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	f := v.FieldByName(field)
	if !f.IsValid() {
		return nil
	}
	if f.Kind() == reflect.Ptr && f.Type().Elem().Kind() == reflect.Float64 {
		if f.IsNil() {
			return nil
		}
		return f.Interface().(*float64)
	}
	return nil
}

// ptrBool converts an *int to an *bool.
func ptrBool(i *int) *bool {
	if i == nil {
		return nil
	}
	v := *i == 1
	return &v
}

// ptrStr extracts a *string field from a struct by name using reflection.
// Returns nil if obj is nil, field is missing, or the field is not a *string.
func ptrStr(obj interface{}, field string) *string {
	if obj == nil {
		return nil
	}
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	f := v.FieldByName(field)
	if !f.IsValid() {
		return nil
	}
	if f.Kind() == reflect.Ptr && f.Type().Elem().Kind() == reflect.String {
		if f.IsNil() {
			return nil
		}
		return f.Interface().(*string)
	}
	return nil
}

// ptrToIface extracts a *float64 field from a struct by name using reflection.
// Returns nil if obj is nil, field is missing, or not a *float64.
func ptrToIface(obj interface{}, field string) interface{} {
	if obj == nil {
		return nil
	}
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	f := v.FieldByName(field)
	if !f.IsValid() {
		return nil
	}
	if f.Kind() == reflect.Ptr && f.Type().Elem().Kind() == reflect.Float64 {
		if f.IsNil() {
			return nil
		}
		return f.Interface()
	}
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
c.Website = ptrStrFromNull(web)
c.Domain = ptrStrFromNull(domain)
if industryTags.Valid {
json.Unmarshal([]byte(industryTags.String), &c.IndustryTags)
}
c.Location = &Location{Municipality: ptrStrFromNull(muni), Province: ptrStrFromNull(prov), AutonomousCommunity: ptrStrFromNull(autoComm), CountryCode: ptrStrFromNull(country), LocationText: ptrStrFromNull(locText)}
if err := json.Unmarshal([]byte(srcEvid.String), &c.SourceEvidenceIDs); err != nil {
return nil, err
}
c.FirstSeenAt = parseTime(firstSeen)
c.LastSeenAt = parseTime(lastSeen)
c.Notes = ptrStrFromNull(notes)
out = append(out, &c)
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


// GetInboundLead returns a single inbound lead by ID.
func (s *Store) GetInboundLead(ctx context.Context, id string) (*InboundLead, error) {
	leads, err := s.ListInboundLeads(ctx, "", 1, 0)
	if err != nil {
		return nil, err
	}
	for _, l := range leads {
		if l.LeadID == id {
			return l, nil
		}
	}
	return nil, fmt.Errorf("inbound lead not found: %s", id)
}

// GetCompany returns a single company by ID.
func (s *Store) GetCompany(ctx context.Context, id string) (*Company, error) {
	companies, err := s.ListCompanies(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range companies {
		if c.CompanyID == id {
			return c, nil
		}
	}
	return nil, fmt.Errorf("company not found: %s", id)
}

// GetSource returns a single source by ID.
func (s *Store) GetSource(ctx context.Context, id string) (*Source, error) {
	sources, err := s.ListSources(ctx)
	if err != nil {
		return nil, err
	}
	for _, src := range sources {
		if src.SourceID == id {
			return src, nil
		}
	}
	return nil, fmt.Errorf("source not found: %s", id)
}

// GetResearchRun returns a single research run by ID.
func (s *Store) GetResearchRun(ctx context.Context, id string) (*ResearchRun, error) {
	runs, err := s.ListResearchRuns(ctx)
	if err != nil {
		return nil, err
	}
	for _, r := range runs {
		if r.RunID == id {
			return r, nil
		}
	}
	return nil, fmt.Errorf("research run not found: %s", id)
}

// GetReviewDecision returns a single review decision by ID.
func (s *Store) GetReviewDecision(ctx context.Context, id string) (*ReviewDecision, error) {
	decisions, err := s.ListReviewDecisions(ctx)
	if err != nil {
		return nil, err
	}
	for _, d := range decisions {
		if d.DecisionID == id {
			return d, nil
		}
	}
	return nil, fmt.Errorf("review decision not found: %s", id)
}

// ptrStrFromNull converts a sql.NullString to *string, returning nil if not valid.
func ptrStrFromNull(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}

















