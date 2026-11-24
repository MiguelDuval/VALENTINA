package domain

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrResearchBusy   = errors.New("research source busy")
	ErrResearchQuota  = errors.New("research quota unavailable")
	ErrResearchPolicy = errors.New("research policy invalid")
)

// ValidateResearchLimits rejects invalid configuration rather than silently
// replacing it with permissive defaults. A nil daily cap explicitly means no
// daily cap; zero interval allows consecutive (never overlapping) manual runs.
func ValidateResearchLimits(l SourceRateLimits) error {
	if l.MinimumIntervalSeconds < 0 || int64(l.MinimumIntervalSeconds) > int64((1<<63-1)/time.Second) ||
		l.MaxParallelRequests < 1 || l.TimeoutSeconds < 1 || int64(l.TimeoutSeconds) > int64((1<<63-1)/time.Second) ||
		l.MaxResponseBytes < 1 || (l.DailyRequestLimit != nil && *l.DailyRequestLimit < 1) {
		return ErrResearchPolicy
	}
	return nil
}

// AdmitResearchRun reserves every source in one transaction. Store owns one
// SQLite connection, so concurrent admissions cannot read the same budget.
// All attempted runs count, regardless of their eventual outcome. No retry or
// network operation occurs inside the transaction.
func (s *Store) AdmitResearchRun(ctx context.Context, ids []string, now time.Time) (*ResearchRun, error) {
	if len(ids) == 0 || now.IsZero() {
		return nil, ErrResearchPolicy
	}
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if strings.TrimSpace(id) == "" || seen[id] {
			return nil, ErrResearchPolicy
		}
		seen[id] = true
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	now = now.UTC()
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for _, id := range ids {
		var limits SourceRateLimits
		var approval string
		var reason, reviewed sql.NullString
		err = tx.QueryRowContext(ctx, `SELECT minimum_interval_seconds, max_parallel_requests,
			daily_request_limit, timeout_seconds, max_response_bytes, approval_status,
			approval_reason, terms_reviewed_at FROM sources WHERE source_id = ?`, id).Scan(
			&limits.MinimumIntervalSeconds, &limits.MaxParallelRequests, &limits.DailyRequestLimit,
			&limits.TimeoutSeconds, &limits.MaxResponseBytes, &approval, &reason, &reviewed)
		if err != nil {
			return nil, ErrResearchPolicy
		}
		if ValidateResearchLimits(limits) != nil || approval != "approved" ||
			strings.TrimSpace(reason.String) == "" || parseTime(reviewed.String).IsZero() {
			return nil, ErrResearchPolicy
		}
		rows, err := tx.QueryContext(ctx, `SELECT started_at, status FROM research_runs
			WHERE EXISTS (SELECT 1 FROM json_each(research_runs.source_ids) WHERE value = ?)`, id)
		if err != nil {
			return nil, err
		}
		count := 0
		var last time.Time
		busy := false
		for rows.Next() {
			var started, status string
			if err = rows.Scan(&started, &status); err != nil {
				rows.Close()
				return nil, err
			}
			at, parseErr := time.Parse(time.RFC3339Nano, started)
			if parseErr != nil {
				rows.Close()
				return nil, ErrResearchPolicy
			}
			if status == "running" {
				busy = true
			}
			if !at.Before(day) && at.Before(day.AddDate(0, 0, 1)) {
				count++
			}
			if at.After(last) {
				last = at
			}
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return nil, err
		}
		if busy {
			return nil, ErrResearchBusy
		}
		if (limits.DailyRequestLimit != nil && count >= *limits.DailyRequestLimit) ||
			(!last.IsZero() && now.Sub(last) < time.Duration(limits.MinimumIntervalSeconds)*time.Second) {
			return nil, ErrResearchQuota
		}
	}
	run := &ResearchRun{RunID: NewRunID(), StartedAt: now, Status: "running", SourceIDs: append([]string(nil), ids...)}
	encoded, err := json.Marshal(ids)
	if err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO research_runs
		(run_id, started_at, status, source_ids, candidate_count, verified_count, duplicate_count, accepted_count, rejected_count)
		VALUES (?, ?, 'running', ?, 0, 0, 0, 0, 0)`, run.RunID, now.Format(time.RFC3339Nano), string(encoded))
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return run, nil
}

// Called only by NewStore, before the single owning process serves requests.
// Reconstructing API handlers never performs recovery over active executions.
func (s *Store) recoverResearchRuns(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `UPDATE research_runs SET status = 'failed', completed_at = ?,
		notes = CASE WHEN notes IS NULL OR notes = '' THEN '' ELSE notes || '; ' END ||
		'Interrupted by application shutdown; manual retry required'
		WHERE status = 'running'`, time.Now().UTC().Format(time.RFC3339Nano))
	return err
}
