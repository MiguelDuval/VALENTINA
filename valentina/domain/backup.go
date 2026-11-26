package domain

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const MaxSnapshotBytes int64 = 128 << 20

// Snapshot owns a private temporary SQLite snapshot. Call Close after streaming.
type Snapshot struct {
	File          *os.File
	Size          int64
	SHA256        string
	SchemaVersion int
	directory     string
}

func (b *Snapshot) Close() { _ = b.File.Close(); _ = os.RemoveAll(b.directory) }

// Snapshot serializes with the owning Store's connection, including WAL writes.
// No caller-supplied filename is accepted. It never opens/migrates the live file
// through a second Store, nor includes Huginn settings or credential storage.
func (s *Store) Snapshot(parent context.Context) (_ *Snapshot, err error) {
	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	var pages, pageSize int64
	if err = conn.QueryRowContext(ctx, "PRAGMA page_count").Scan(&pages); err != nil {
		return nil, err
	}
	if err = conn.QueryRowContext(ctx, "PRAGMA page_size").Scan(&pageSize); err != nil {
		return nil, err
	}
	if pages <= 0 || pageSize <= 0 || pages > MaxSnapshotBytes/pageSize {
		return nil, fmt.Errorf("snapshot size limit")
	}
	var version int
	if err = conn.QueryRowContext(ctx, "SELECT max(version) FROM schema_migrations").Scan(&version); err != nil {
		return nil, err
	}
	if version != migrations[len(migrations)-1].version {
		return nil, fmt.Errorf("unsupported snapshot schema")
	}
	var sourceIntegrity string
	if err = conn.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&sourceIntegrity); err != nil {
		return nil, err
	}
	if sourceIntegrity != "ok" {
		return nil, fmt.Errorf("source integrity check failed")
	}
	directory, err := os.MkdirTemp("", "valentina-crm-backup-")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(directory)
		}
	}()
	path := filepath.Join(directory, "crm.sqlite")
	// SQLite VACUUM INTO creates a transactionally consistent standalone image.
	if _, err = conn.ExecContext(ctx, "VACUUM main INTO ?", path); err != nil {
		return nil, err
	}
	check, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	defer check.Close()
	var integrity string
	if err = check.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&integrity); err != nil {
		return nil, err
	}
	if integrity != "ok" {
		return nil, fmt.Errorf("snapshot integrity check failed")
	}
	if err = check.QueryRowContext(ctx, "SELECT max(version) FROM schema_migrations").Scan(&version); err != nil {
		return nil, err
	}
	if err = check.Close(); err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = file.Close()
		}
	}()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() <= 0 || info.Size() > MaxSnapshotBytes {
		return nil, fmt.Errorf("snapshot size limit")
	}
	hash := sha256.New()
	if _, err = io.Copy(hash, file); err != nil {
		return nil, err
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	return &Snapshot{File: file, Size: info.Size(), SHA256: fmt.Sprintf("%x", hash.Sum(nil)), SchemaVersion: version, directory: directory}, nil
}
