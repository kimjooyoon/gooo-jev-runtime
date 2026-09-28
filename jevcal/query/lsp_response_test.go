package query

import "testing"

func TestNewLSPResponsePreservesUnresolvedTerms(t *testing.T) {
	result := Result{
		QueryDigest:      "sha256:query",
		Status:           StatusNeedsInput,
		CandidateIDs:     []string{"gooo.declaration.diff"},
		UnresolvedTerms:  []string{"safe"},
		NextQuestion:     "Clarify safe.",
		ProvenanceDigest: "sha256:catalog",
	}
	response := NewLSPResponse(result, "sha256:lsp-contract")
	if response.Status != StatusNeedsInput || len(response.UnresolvedTerms) != 1 || response.NextQuestion == "" {
		t.Fatalf("response lost discovery boundary: %+v", response)
	}
	if response.ContractDigest == "" {
		t.Fatal("missing contract digest")
	}
}