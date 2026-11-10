package domain

import (
	"crypto/sha256"
	"encoding/hex"
)

// DuplicateKey computes a deterministic deduplication key for an opportunity.
// It uses canonical URL + publisher + publication/deadline window.
// CanonicalURL, Publisher, DateWindow, IsDuplicate, IsDuplicateLead and
// ValidateDuplicateCandidate are defined in models.go and must not be duplicated here.
func DuplicateKey(o *Opportunity) string {
	h := sha256.New()
	h.Write([]byte(o.CanonicalURL()))
	h.Write([]byte(o.Publisher()))
	h.Write([]byte(o.DateWindow()))
	return hex.EncodeToString(h.Sum(nil))
}
