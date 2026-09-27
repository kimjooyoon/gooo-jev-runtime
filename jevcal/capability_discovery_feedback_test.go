package jevcal

import "testing"

func TestObserveCapabilityDiscoveryFeedbackBindsKnownOutcome(t *testing.T) {
	observation := DiscoverCapabilities(CapabilityDiscoveryInput{
		SourceVersion:   "gooo-source-v1",
		ContractVersion: "jev-capability-discovery-v1",
		Query:           "What can gooo do?",
	})
	feedback := ObserveCapabilityDiscoveryFeedback(CapabilityDiscoveryFeedbackInput{
		Observation:      observation,
		OutcomeKnown:     true,
		OutcomeAccepted:  true,
		OutcomeDigest:    "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	})
	if feedback.Status != CapabilityDiscoveryFeedbackBound || feedback.FirstMismatch != "" {
		t.Fatalf("unexpected bound feedback: %+v", feedback)
	}
	if err := feedback.Validate(); err != nil {
		t.Fatalf("bound feedback should validate: %v", err)
	}
}

func TestObserveCapabilityDiscoveryFeedbackKeepsUnknownOutcome(t *testing.T) {
	observation := DiscoverCapabilities(CapabilityDiscoveryInput{
		SourceVersion:   "gooo-source-v1",
		ContractVersion: "jev-capability-discovery-v1",
		Query:           "What can gooo do?",
	})
	feedback := ObserveCapabilityDiscoveryFeedback(CapabilityDiscoveryFeedbackInput{Observation: observation})
	if feedback.Status != CapabilityDiscoveryFeedbackUnknown || feedback.FirstMismatch != "outcome" || feedback.MissingStage != "reverse_observation" {
		t.Fatalf("unexpected unknown feedback: %+v", feedback)
	}
	if err := feedback.Validate(); err != nil {
		t.Fatalf("unknown feedback should validate: %v", err)
	}
}

func TestObserveCapabilityDiscoveryFeedbackPreservesDeferredBoundary(t *testing.T) {
	observation := DiscoverCapabilities(CapabilityDiscoveryInput{
		SourceVersion:   "gooo-source-v1",
		ContractVersion: "jev-capability-discovery-v1",
		Query:           "Can gooo execute and authorize a network workload?",
	})
	feedback := ObserveCapabilityDiscoveryFeedback(CapabilityDiscoveryFeedbackInput{
		Observation:  observation,
		OutcomeKnown: true,
		OutcomeDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	})
	if feedback.Status != CapabilityDiscoveryFeedbackUnknown || feedback.FirstMismatch != "execution_or_authorization_boundary" {
		t.Fatalf("unexpected deferred feedback: %+v", feedback)
	}
	if err := feedback.Validate(); err != nil {
		t.Fatalf("deferred feedback should validate: %v", err)
	}
}

func TestObserveCapabilityDiscoveryFeedbackPreservesCapabilityBoundary(t *testing.T) {
	observation := DiscoverCapabilities(CapabilityDiscoveryInput{
		SourceVersion:   "gooo-source-v1",
		ContractVersion: "jev-capability-discovery-v1",
		Query:           "What can gooo do?",
	})
	observation.CanExecute = true
	feedback := ObserveCapabilityDiscoveryFeedback(CapabilityDiscoveryFeedbackInput{
		Observation:   observation,
		OutcomeKnown:  true,
		OutcomeDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	})
	if feedback.Status != CapabilityDiscoveryFeedbackUnknown || feedback.FirstMismatch != "capability-boundary" {
		t.Fatalf("unexpected capability boundary feedback: %+v", feedback)
	}
	if err := feedback.Validate(); err != nil {
		t.Fatalf("capability boundary feedback should validate: %v", err)
	}
}

func TestObserveCapabilityDiscoveryFeedbackRejectsTamperedDigest(t *testing.T) {
	observation := DiscoverCapabilities(CapabilityDiscoveryInput{
		SourceVersion:   "gooo-source-v1",
		ContractVersion: "jev-capability-discovery-v1",
		Query:           "What can gooo do?",
	})
	feedback := ObserveCapabilityDiscoveryFeedback(CapabilityDiscoveryFeedbackInput{
		Observation:  observation,
		OutcomeKnown: true,
		OutcomeDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	})
	feedback.Query = "tampered"
	if err := feedback.Validate(); err == nil {
		t.Fatal("tampered feedback should fail validation")
	}
}
