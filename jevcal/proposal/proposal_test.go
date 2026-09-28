package proposal

import "testing"

func TestBuildCreatesReviewOnlyProposal(t *testing.T) {
	result := Build(Input{
		DomainID:             "gooo.provenance",
		MissingCapabilityIDs: []string{"gooo.provenance.compare"},
		ObservedSignals:      []string{SignalReviewMissing},
		EvidenceDigests:      []string{"sha256:observation"},
		SourceDigest:         "sha256:source",
		ContractDigest:       "sha256:contract",
	})
	if result.Status != StatusProposed {
		t.Fatalf("status = %q, want %q", result.Status, StatusProposed)
	}
	if result.Proposal == "" || result.NextQuestion == "" {
		t.Fatal("expected proposal and read-only next question")
	}
}

func TestBuildRejectsUnknownSignal(t *testing.T) {
	result := Build(Input{
		DomainID:        "gooo.provenance",
		ObservedSignals: []string{"EXECUTE_NOW"},
		EvidenceDigests: []string{"sha256:observation"},
		SourceDigest:    "sha256:source",
		ContractDigest:  "sha256:contract",
	})
	if result.Status != StatusRejected {
		t.Fatalf("status = %q, want %q", result.Status, StatusRejected)
	}
}