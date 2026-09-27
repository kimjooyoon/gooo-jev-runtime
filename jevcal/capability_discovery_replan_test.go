package jevcal

import "testing"

func TestProposeCapabilityDiscoveryReplanSelectsNextIteration(t *testing.T) {
	window := ObserveCapabilityDiscoveryFeedbackWindow(CapabilityDiscoveryFeedbackWindowInput{
		Observations: []CapabilityDiscoveryFeedbackObservation{
			capabilityFeedbackForWindowTest(t, true),
			capabilityFeedbackForWindowTest(t, false),
		},
		MinimumWindow: 2,
	})
	proposal := ProposeCapabilityDiscoveryReplan(window)
	if proposal.Status != CapabilityDiscoveryReplanBound || proposal.Disposition != CapabilityDiscoveryReplanNextIteration || proposal.NextOperation != "select_next_capability_iteration" {
		t.Fatalf("unexpected next-iteration proposal: %+v", proposal)
	}
	if err := proposal.Validate(); err != nil {
		t.Fatalf("next-iteration proposal should validate: %v", err)
	}
}

func TestProposeCapabilityDiscoveryReplanHoldsIncompleteEvidence(t *testing.T) {
	window := ObserveCapabilityDiscoveryFeedbackWindow(CapabilityDiscoveryFeedbackWindowInput{
		Observations: []CapabilityDiscoveryFeedbackObservation{capabilityFeedbackForWindowTest(t, true)},
		MinimumWindow: 2,
	})
	proposal := ProposeCapabilityDiscoveryReplan(window)
	if proposal.Status != CapabilityDiscoveryReplanUnknown || proposal.Disposition != CapabilityDiscoveryReplanHoldEvidence || proposal.FirstMismatch != "minimum-window" {
		t.Fatalf("unexpected evidence hold: %+v", proposal)
	}
	if err := proposal.Validate(); err != nil {
		t.Fatalf("evidence hold should validate: %v", err)
	}
}

func TestProposeCapabilityDiscoveryReplanReviewsWithoutAcceptedOutcome(t *testing.T) {
	window := ObserveCapabilityDiscoveryFeedbackWindow(CapabilityDiscoveryFeedbackWindowInput{
		Observations: []CapabilityDiscoveryFeedbackObservation{capabilityFeedbackForWindowTest(t, false)},
		MinimumWindow: 1,
	})
	proposal := ProposeCapabilityDiscoveryReplan(window)
	if proposal.Status != CapabilityDiscoveryReplanBound || proposal.Disposition != CapabilityDiscoveryReplanReviewOutcome || proposal.NextOperation != "review_capability_use" {
		t.Fatalf("unexpected review proposal: %+v", proposal)
	}
	if err := proposal.Validate(); err != nil {
		t.Fatalf("review proposal should validate: %v", err)
	}
}

func TestProposeCapabilityDiscoveryReplanPreservesInvalidWindowAsUnknown(t *testing.T) {
	window := CapabilityDiscoveryFeedbackWindowObservation{
		Status:          CapabilityDiscoveryFeedbackWindowBound,
		EvidenceDigest:  "not-a-digest",
		ObservationCount: 1,
		KnownObservationCount: 1,
		AcceptedCount:   1,
		MinimumWindow:   1,
		EvidenceCoverage: 1,
		IsReadOnly:      true,
	}
	proposal := ProposeCapabilityDiscoveryReplan(window)
	if proposal.Status != CapabilityDiscoveryReplanUnknown || proposal.FirstMismatch != "window-integrity" {
		t.Fatalf("unexpected invalid-window proposal: %+v", proposal)
	}
	if err := proposal.Validate(); err != nil {
		t.Fatalf("invalid-window UNKNOWN should validate: %v", err)
	}
}

func TestProposeCapabilityDiscoveryReplanRejectsTampering(t *testing.T) {
	window := ObserveCapabilityDiscoveryFeedbackWindow(CapabilityDiscoveryFeedbackWindowInput{
		Observations: []CapabilityDiscoveryFeedbackObservation{capabilityFeedbackForWindowTest(t, true)},
		MinimumWindow: 1,
	})
	proposal := ProposeCapabilityDiscoveryReplan(window)
	proposal.NextOperation = "mutate_repository"
	if err := proposal.Validate(); err == nil {
		t.Fatal("tampered replan should fail validation")
	}
}
