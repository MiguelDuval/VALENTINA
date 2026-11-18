package sourcepolicy

import (
	"testing"
	"time"
)

func TestRegistryApprovedAndFetchableCounts(t *testing.T) {
	r := NewRegistry()
	if r.Count() != 0 {
		t.Fatalf("expected 0 sources, got %d", r.Count())
	}

	p1 := &SourceProfile{
		SourceID:            "src-1",
		Name:                "Test Source 1",
		Publisher:           "Test Publisher",
		ApprovalStatus:      ApprovalApproved,
		AllowedHosts:        []string{"example.com"},
		AllowedPathPrefixes: []string{"/"},
		TimeoutSeconds:      10,
		MaxResponseBytes:    1024,
	}
	if err := r.Register(p1); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if r.Count() != 1 {
		t.Fatalf("expected 1 source, got %d", r.Count())
	}
	if r.ApprovedCount() != 1 {
		t.Fatalf("expected 1 approved source, got %d", r.ApprovedCount())
	}
	if r.FetchableCount() != 1 {
		t.Fatalf("expected 1 fetchable source, got %d", r.FetchableCount())
	}

	// Add a pending source.
	p2 := &SourceProfile{
		SourceID:            "src-2",
		Name:                "Test Source 2",
		Publisher:           "Test Publisher",
		ApprovalStatus:      ApprovalPending,
		AllowedHosts:        []string{"example.org"},
		AllowedPathPrefixes: []string{"/"},
		TimeoutSeconds:      10,
		MaxResponseBytes:    1024,
	}
	if err := r.Register(p2); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if r.Count() != 2 {
		t.Fatalf("expected 2 sources, got %d", r.Count())
	}
	if r.ApprovedCount() != 1 {
		t.Fatalf("expected 1 approved source, got %d", r.ApprovedCount())
	}
	if r.FetchableCount() != 1 {
		t.Fatalf("expected 1 fetchable source, got %d", r.FetchableCount())
	}

	// Get should return the right profile.
	got := r.Get("src-1")
	if got == nil || got.SourceID != "src-1" {
		t.Fatalf("Get(src-1) returned wrong profile: %+v", got)
	}
	if r.Get("nonexistent") != nil {
		t.Fatal("Get(nonexistent) should return nil")
	}

	// Remove should work.
	r.Remove("src-2")
	if r.Count() != 1 {
		t.Fatalf("expected 1 source after remove, got %d", r.Count())
	}
}

func TestSourceProfileDefaults(t *testing.T) {
	p := &SourceProfile{
		SourceID:       "src-1",
		ApprovalStatus: ApprovalPending,
	}
	// Simulate defaults applied by Register.
	if p.ApprovalStatus == "" {
		p.ApprovalStatus = ApprovalPending
	}
	if p.TimeoutSeconds <= 0 {
		p.TimeoutSeconds = 10
	}
	if p.MaxResponseBytes <= 0 {
		p.MaxResponseBytes = 512 * 1024
	}

	if p.IsApproved() {
		t.Fatal("pending source should not be approved")
	}
	if p.CanFetch() {
		t.Fatal("pending source should not be fetchable")
	}

	p.ApprovalStatus = ApprovalApproved
	if !p.IsApproved() {
		t.Fatal("approved source should be approved")
	}
	if !p.CanFetch() {
		t.Fatal("approved source should be fetchable")
	}

	p.ApprovalStatus = ApprovalPaused
	if p.CanFetch() {
		t.Fatal("paused source should not be fetchable")
	}

	p.ApprovalStatus = ApprovalBlocked
	if p.CanFetch() {
		t.Fatal("blocked source should not be fetchable")
	}
}

func TestSourceProfileDefaultsWithRegistry(t *testing.T) {
	r := NewRegistry()
	p := &SourceProfile{
		SourceID:            "src-default",
		Name:                "Default Test",
		Publisher:           "Pub",
		AllowedHosts:        []string{"example.com"},
		AllowedPathPrefixes: []string{"/"},
		// ApprovalStatus intentionally empty -> should default to pending_review
		// TimeoutSeconds intentionally 0 -> should default to 10
		// MaxResponseBytes intentionally 0 -> should default to 512KB
	}
	if err := r.Register(p); err != nil {
		t.Fatalf("Register: %v", err)
	}
	got := r.Get("src-default")
	if got.ApprovalStatus != ApprovalPending {
		t.Fatalf("expected pending_review default, got %s", got.ApprovalStatus)
	}
	if got.TimeoutSeconds != 10 {
		t.Fatalf("expected default timeout 10, got %d", got.TimeoutSeconds)
	}
	if got.MaxResponseBytes != 512*1024 {
		t.Fatalf("expected default max response 512KB, got %d", got.MaxResponseBytes)
	}
	if got.CanFetch() {
		t.Fatal("default pending source should not be fetchable")
	}
}

func TestRegistryValidation(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(&SourceProfile{}); err == nil {
		t.Fatal("expected error for empty source_id")
	}
}

func TestApprovalStatusConstants(t *testing.T) {
	if ApprovalPending != "pending_review" {
		t.Fatal("unexpected ApprovalPending value")
	}
	if ApprovalApproved != "approved" {
		t.Fatal("unexpected ApprovalApproved value")
	}
	if ApprovalPaused != "paused" {
		t.Fatal("unexpected ApprovalPaused value")
	}
	if ApprovalBlocked != "blocked" {
		t.Fatal("unexpected ApprovalBlocked value")
	}
}

func TestTimeFields(t *testing.T) {
	now := time.Now()
	p := &SourceProfile{
		SourceID:        "src-time",
		TermsReviewedAt: &now,
		LastReviewedAt:  &now,
	}
	r := NewRegistry()
	if err := r.Register(p); err != nil {
		t.Fatalf("Register: %v", err)
	}
	got := r.Get("src-time")
	if got.TermsReviewedAt == nil || got.LastReviewedAt == nil {
		t.Fatal("time fields should be preserved")
	}
}

func TestRegistryGetAndListReturnDeepCopies(t *testing.T) {
	registry := makeTestRegistry()
	got := registry.Get("allowed-src")
	if got == nil {
		t.Fatal("expected registered source")
	}
	got.AllowedHosts[0] = "attacker.invalid"
	got.AllowedPathPrefixes[0] = "/private/"
	listed := registry.List()
	for _, profile := range listed {
		if profile.SourceID == "allowed-src" {
			profile.AllowedHosts[0] = "another.invalid"
		}
	}
	fresh := registry.Get("allowed-src")
	if fresh.AllowedHosts[0] != "example.com" || fresh.AllowedPathPrefixes[0] != "/api/" {
		t.Fatalf("caller mutation leaked into registry: hosts=%v paths=%v", fresh.AllowedHosts, fresh.AllowedPathPrefixes)
	}
}

func TestRegistryRegisterClonesCallerProfile(t *testing.T) {
	registry := NewRegistry()
	input := &SourceProfile{SourceID: "clone-input", AllowedHosts: []string{"example.com"}, AllowedPathPrefixes: []string{"/api/"}}
	if err := registry.Register(input); err != nil {
		t.Fatalf("Register: %v", err)
	}
	input.AllowedHosts[0] = "attacker.invalid"
	input.AllowedPathPrefixes[0] = "/private/"
	stored := registry.Get("clone-input")
	if stored.AllowedHosts[0] != "example.com" || stored.AllowedPathPrefixes[0] != "/api/" {
		t.Fatalf("caller mutation leaked into registry: %+v", stored)
	}
}
