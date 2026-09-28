package lifecycle

import "testing"

func TestValidateTransitionRequiresExplicitReviewForApplied(t *testing.T) {
	err := ValidateTransition(Transition{
		From: Measured,
		To: Applied,
		Evidence: Evidence{
			ProposalDigest:   "sha256:proposal",
			ProvenanceDigest: "sha256:provenance",
		},
	})
	if err == nil {
		t.Fatal("expected missing review and application evidence")
	}
}

func TestValidateTransitionAcceptsReversalEvidence(t *testing.T) {
	err := ValidateTransition(Transition{
		From: Applied,
		To: Reversed,
		Evidence: Evidence{
			ApplicationDigest: "sha256:application",
			ReversalDigest:    "sha256:reversal",
			ProvenanceDigest:  "sha256:provenance",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}