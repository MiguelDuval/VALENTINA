package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// DuplicateKey computes a deterministic deduplication key for an opportunity.
// It uses canonical URL + publisher + publication/deadline window.
func DuplicateKey(o *Opportunity) string {
	h := sha256.New()
	h.Write([]byte(o.CanonicalURL()))
	h.Write([]byte(o.Publisher()))
	h.Write([]byte(o.DateWindow()))
	return hex.EncodeToString(h.Sum(nil))
}

// CanonicalURL returns the opportunity's canonical URL for dedup purposes.
func (o *Opportunity) CanonicalURL() string {
	// Opportunities don't store a direct URL; evidence does.
	// This is a placeholder that subclasses/implementations can override.
	return ""
}

// Publisher returns the publisher name.
func (o *Opportunity) Publisher() string {
	return ""
}

// DateWindow returns the publication/deadline window string.
func (o *Opportunity) DateWindow() string {
	var parts []string
	if o.PublishedAt != nil {
		parts = append(parts, o.PublishedAt.Format("2006-01-02"))
	}
	if o.DeadlineAt != nil {
		parts = append(parts, o.DeadlineAt.Format("2006-01-02"))
	}
	return strings.Join(parts, "|")
}

// IsDuplicate checks whether two opportunities are likely duplicates based on
// deterministic matching: exact canonical URL, or exact title+publisher+date window.
func IsDuplicate(a, b *Opportunity) bool {
	if a.CanonicalURL() != "" && a.CanonicalURL() == b.CanonicalURL() {
		return true
	}
	if NormalizeCompanyName(a.Publisher()) == NormalizeCompanyName(b.Publisher()) &&
		a.Title == b.Title &&
		a.DateWindow() == b.DateWindow() &&
		a.DateWindow() != "" {
		return true
	}
	return false
}

// IsDuplicateLead checks whether two inbound leads are likely duplicates based
// on channel + message reference + timestamp window.
func IsDuplicateLead(a, b *InboundLead) bool {
	if a.ChannelMessageRef != nil && b.ChannelMessageRef != nil && *a.ChannelMessageRef == *b.ChannelMessageRef {
		return true
	}
	if a.ReceivedAt.Equal(b.ReceivedAt) &&
		a.InboundChannel == b.InboundChannel &&
		a.RequestSummary == b.RequestSummary {
		return true
	}
	return false
}

// fmtDate formats a time.Time as YYYY-MM-DD, or empty string if zero.
func fmtDate(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

// ValidateDuplicateCandidate checks that an opportunity doesn't conflict with
// an existing duplicate group.
func ValidateDuplicateCandidate(newOpp *Opportunity, existing []*Opportunity) error {
	for _, existing := range existing {
		if IsDuplicate(newOpp, existing) {
			return fmt.Errorf("potential duplicate of opportunity %s (duplicate group: %s)",
				existing.OpportunityID, existing.DuplicateGroupID)
		}
	}
	return nil
}
