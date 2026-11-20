package domain

import (
	"testing"
	"time"
)

func TestSourceCanFetchRequiresReviewedTermsAndReason(t *testing.T) {
	reason := "Public RSS scope and publisher terms were checked"
	reviewed := time.Now().UTC()
	s := &Source{ApprovalStatus: "approved"}
	if s.CanFetch() {
		t.Fatal("approved status alone must not permit fetching")
	}
	s.TermsReviewedAt = &reviewed
	if s.CanFetch() {
		t.Fatal("missing approval reason must block fetching")
	}
	s.ApprovalReason = &reason
	if !s.CanFetch() {
		t.Fatal("reviewed source unexpectedly rejected")
	}
	s.ApprovalStatus = "paused"
	if s.CanFetch() {
		t.Fatal("paused source unexpectedly fetchable")
	}
}
