package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// ValidateOpportunityJSON validates raw JSON against the opportunity schema
// and returns a parsed Opportunity if valid.
func ValidateOpportunityJSON(data []byte) (*Opportunity, error) {
	var opp Opportunity
	if err := json.Unmarshal(data, &opp); err != nil {
		return nil, fmt.Errorf("unmarshal opportunity: %w", err)
	}
	if err := opp.Validate(); err != nil {
		return nil, fmt.Errorf("validate opportunity: %w", err)
	}
	return &opp, nil
}

// ValidateLeadJSON validates raw JSON against the inbound lead schema
// and returns a parsed InboundLead if valid.
func ValidateLeadJSON(data []byte) (*InboundLead, error) {
	var lead InboundLead
	if err := json.Unmarshal(data, &lead); err != nil {
		return nil, fmt.Errorf("unmarshal inbound lead: %w", err)
	}
	if err := lead.Validate(); err != nil {
		return nil, fmt.Errorf("validate inbound lead: %w", err)
	}
	return &lead, nil
}

// ValidateSourceJSON validates raw JSON against the source schema
// and returns a parsed Source if valid.
func ValidateSourceJSON(data []byte) (*Source, error) {
	var src Source
	if err := json.Unmarshal(data, &src); err != nil {
		return nil, fmt.Errorf("unmarshal source: %w", err)
	}
	if src.SourceID == "" {
		return nil, errors.New("source_id is required")
	}
	if src.ApprovalStatus == "" {
		src.ApprovalStatus = "pending_review"
	}
	return &src, nil
}

// ValidateEvidenceJSON validates raw JSON against the source evidence schema
// and returns a parsed SourceEvidence if valid.
func ValidateEvidenceJSON(data []byte) (*SourceEvidence, error) {
	var ev SourceEvidence
	if err := json.Unmarshal(data, &ev); err != nil {
		return nil, fmt.Errorf("unmarshal source evidence: %w", err)
	}
	if ev.EvidenceID == "" {
		return nil, errors.New("evidence_id is required")
	}
	if ev.SourceID == "" {
		return nil, errors.New("source_id is required")
	}
	if ev.CanonicalURL == "" {
		return nil, errors.New("canonical_url is required")
	}
	if ev.Publisher == "" {
		return nil, errors.New("publisher is required")
	}
	return &ev, nil
}

// LoadSchema loads a JSON schema file from the schemas directory.
func LoadSchema(name string) ([]byte, error) {
	path := fmt.Sprintf("schemas/%s.schema.json", name)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read schema %s: %w", name, err)
	}
	return data, nil
}

// SchemaDir returns the path to the schemas directory.
func SchemaDir() string {
	return "schemas"
}
