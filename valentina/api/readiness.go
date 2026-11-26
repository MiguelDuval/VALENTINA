package api

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

//go:embed LICENSE.huginn.txt
var huginnLicense string

func (s *Server) handleLicenses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]string{"license": huginnLicense, "sha256": fmt.Sprintf("%x", sha256.Sum256([]byte(huginnLicense))), "upstream_commit": "f01c1d84c08b1de972d8cc268da25602598adff8"})
}

func (s *Server) handleCRMBackup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if !s.backupMu.TryLock() {
		writeError(w, http.StatusConflict, "CRM backup already in progress")
		return
	}
	defer s.backupMu.Unlock()
	if r.URL.RawQuery != "" {
		writeError(w, http.StatusBadRequest, "backup does not accept parameters")
		return
	}
	// Bound slow clients as well as SQLite work; the host bearer wrapper remains
	// mandatory. Never log the snapshot, filenames, records or underlying errors.
	controller := http.NewResponseController(w)
	_ = controller.SetWriteDeadline(time.Now().Add(60 * time.Second))
	defer controller.SetWriteDeadline(time.Time{})
	snapshot, err := s.store.Snapshot(r.Context())
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "CRM backup unavailable")
		return
	}
	defer snapshot.Close()
	w.Header().Set("Content-Type", "application/vnd.sqlite3")
	w.Header().Set("Content-Disposition", `attachment; filename="VALENTINA-CRM.sqlite"`)
	w.Header().Set("Content-Length", strconv.FormatInt(snapshot.Size, 10))
	w.Header().Set("X-Backup-SHA256", snapshot.SHA256)
	w.Header().Set("X-Backup-Schema-Version", strconv.Itoa(snapshot.SchemaVersion))
	w.Header().Set("X-Backup-Scope", "valentina-crm-only")
	_, _ = io.CopyN(w, snapshot.File, snapshot.Size)
}
