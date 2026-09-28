package query

import "testing"

func TestDiscoverPreservesAmbiguity(t *testing.T) {
	result := Discover(Request{
		QueryText:        "compare declared history",
		Terms:            []string{"history", "declared", "history"},
		CandidateIDs:     []string{"gooo.provenance.compare", "gooo.declaration.diff"},
		QueryDigest:      "sha256:query",
		ProvenanceDigest: "sha256:provenance",
	})
	if result.Status != StatusDeferred {
		t.Fatalf("status = %q, want %q", result.Status, StatusDeferred)
	}
	if result.NextQuestion == "" {
		t.Fatal("expected a read-only next question")
	}
}

func TestDiscoverRejectsMissingEvidence(t *testing.T) {
	result := Discover(Request{QueryText: "find capability"})
	if result.Status != StatusUnknown {
		t.Fatalf("status = %q, want %q", result.Status, StatusUnknown)
	}
}