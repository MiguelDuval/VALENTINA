package domain

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotConcurrentWAL(t *testing.T) {
	ctx := context.Background()
	s, err := NewStore(ctx, filepath.Join(t.TempDir(), "live.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err = s.db.Exec("PRAGMA journal_mode=WAL; CREATE TABLE snapshot_fixture (n INTEGER); CREATE TABLE snapshot_pair (n INTEGER)"); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	started := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			tx, e := s.db.Begin()
			if e != nil {
				done <- e
				return
			}
			_, e = tx.Exec("INSERT INTO snapshot_fixture VALUES (?); INSERT INTO snapshot_pair VALUES (?)", i, i)
			if e == nil {
				e = tx.Commit()
			} else {
				_ = tx.Rollback()
			}
			if e != nil {
				done <- e
				return
			}
			if i == 0 {
				close(started)
			}
		}
		done <- nil
	}()
	select {
	case <-started:
	case e := <-done:
		t.Fatal("writer failed", e)
	}
	snapshot, err := s.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer snapshot.Close()
	data, err := os.ReadFile(snapshot.File.Name())
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != snapshot.SHA256 || int64(len(data)) != snapshot.Size || snapshot.SchemaVersion < 14 {
		t.Fatal("invalid metadata")
	}
	db, err := sql.Open("sqlite", snapshot.File.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var integrity string
	if err = db.QueryRow("PRAGMA integrity_check").Scan(&integrity); err != nil || integrity != "ok" {
		t.Fatal("invalid snapshot", err)
	}
	var count int
	if err = db.QueryRow("SELECT count(*) FROM snapshot_fixture").Scan(&count); err != nil || count < 1 || count > 50 {
		t.Fatal("inconsistent snapshot", err)
	}
	var differences int
	if err = db.QueryRow("SELECT count(*) FROM (SELECT n FROM snapshot_fixture EXCEPT SELECT n FROM snapshot_pair)").Scan(&differences); err != nil || differences != 0 {
		t.Fatal("torn transaction", err)
	}
	if err = <-done; err != nil {
		t.Fatal(err)
	}
	db.Close()
	snapshot.Close()
	if _, err = os.Stat(filepath.Dir(snapshot.File.Name())); !os.IsNotExist(err) {
		t.Fatal("temporary snapshot survived cleanup")
	}
}

func TestSnapshotFailures(t *testing.T) {
	ctx := context.Background()
	s, err := NewStore(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if b, e := s.Snapshot(cancelled); e == nil {
		b.Close()
		t.Fatal("cancelled snapshot accepted")
	}
	s.Close()
	if b, e := s.Snapshot(ctx); e == nil {
		b.Close()
		t.Fatal("closed source accepted")
	}
	db, _ := sql.Open("sqlite", ":memory:")
	defer db.Close()
	if b, e := (&Store{db: db}).Snapshot(ctx); e == nil {
		b.Close()
		t.Fatal("empty source accepted")
	}
}

func TestSnapshotCorruptSourceAndCleanup(t *testing.T) {
	root := t.TempDir()
	t.Setenv("TMP", root)
	t.Setenv("TEMP", root)
	t.Setenv("TMPDIR", root)
	path := filepath.Join(root, "corrupt.sqlite")
	if err := os.WriteFile(path, []byte("not a SQLite database"), 0600); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if b, e := (&Store{db: db}).Snapshot(context.Background()); e == nil {
		b.Close()
		t.Fatal("corrupt database accepted")
	}
	// Force VACUUM to fail after private scratch creation, without corrupting
	// real data: SQLite query-only mode rejects VACUUM INTO.
	s, err := NewStore(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err = s.db.Exec("PRAGMA query_only=ON"); err != nil {
		t.Fatal(err)
	}
	if b, e := s.Snapshot(context.Background()); e == nil {
		b.Close()
		t.Fatal("invalid source accepted")
	}
	scratch, err := filepath.Glob(filepath.Join(root, "valentina-crm-backup-*"))
	if err != nil || len(scratch) != 0 {
		t.Fatal("failed snapshot leaked scratch")
	}
}
