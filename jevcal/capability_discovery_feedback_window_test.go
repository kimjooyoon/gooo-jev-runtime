package jevcal

import "testing"

func capabilityFeedbackForWindowTest(t *testing.T, accepted bool) CapabilityDiscoveryFeedbackObservation {
	t.Helper()
	observation := DiscoverCapabilities(CapabilityDiscoveryInput{
		SourceVersion:   "gooo-source-v1",
		ContractVersion: "jev-capability-discovery-v1",
		Query:           "What can gooo do?",
	})
	feedback := ObserveCapabilityDiscoveryFeedback(CapabilityDiscoveryFeedbackInput{
		Observation:    observation,
		OutcomeKnown:   true,
		OutcomeAccepted: accepted,
		OutcomeDigest:  "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	})
	if err := feedback.Validate(); err != nil {
		t.Fatalf("feedback fixture should validate: %v", err)
	}
	return feedback
}

func TestObserveCapabilityDiscoveryFeedbackWindowBindsCompleteSamples(t *testing.T) {
	window := ObserveCapabilityDiscoveryFeedbackWindow(CapabilityDiscoveryFeedbackWindowInput{
		Observations: []CapabilityDiscoveryFeedbackObservation{
			capabilityFeedbackForWindowTest(t, true),
			capabilityFeedbackForWindowTest(t, false),
		},
		MinimumWindow: 2,
	})
	if window.Status != CapabilityDiscoveryFeedbackWindowBound || window.AcceptedCount != 1 || window.RejectedCount != 1 {
		t.Fatalf("unexpected bound feedback window: %+v", window)
	}
	if err := window.Validate(); err != nil {
		t.Fatalf("bound feedback window should validate: %v", err)
	}
}

func TestObserveCapabilityDiscoveryFeedbackWindowKeepsMinimumUnknown(t *testing.T) {
	window := ObserveCapabilityDiscoveryFeedbackWindow(CapabilityDiscoveryFeedbackWindowInput{
		Observations:  []CapabilityDiscoveryFeedbackObservation{capabilityFeedbackForWindowTest(t, true)},
		MinimumWindow: 2,
	})
	if window.Status != CapabilityDiscoveryFeedbackWindowUnknown || window.FirstMismatch != "minimum-window" {
		t.Fatalf("unexpected incomplete feedback window: %+v", window)
	}
	if err := window.Validate(); err != nil {
		t.Fatalf("incomplete feedback window should validate: %v", err)
	}
}

func TestObserveCapabilityDiscoveryFeedbackWindowPreservesUnknownBoundary(t *testing.T) {
	observation := DiscoverCapabilities(CapabilityDiscoveryInput{
		SourceVersion:   "gooo-source-v1",
		ContractVersion: "jev-capability-discovery-v1",
		Query:           "Can gooo execute and authorize a network workload?",
	})
	feedback := ObserveCapabilityDiscoveryFeedback(CapabilityDiscoveryFeedbackInput{Observation: observation})
	window := ObserveCapabilityDiscoveryFeedbackWindow(CapabilityDiscoveryFeedbackWindowInput{
		Observations:  []CapabilityDiscoveryFeedbackObservation{feedback},
		MinimumWindow: 1,
	})
	if window.Status != CapabilityDiscoveryFeedbackWindowUnknown || window.MissingStage != "explicit_external_boundary" {
		t.Fatalf("unexpected deferred feedback window: %+v", window)
	}
	if err := window.Validate(); err != nil {
		t.Fatalf("deferred feedback window should validate: %v", err)
	}
}

func TestObserveCapabilityDiscoveryFeedbackWindowRejectsTamperedDigest(t *testing.T) {
	window := ObserveCapabilityDiscoveryFeedbackWindow(CapabilityDiscoveryFeedbackWindowInput{
		Observations:  []CapabilityDiscoveryFeedbackObservation{capabilityFeedbackForWindowTest(t, true)},
		MinimumWindow: 1,
	})
	window.AcceptedCount = 0
	if err := window.Validate(); err == nil {
		t.Fatal("tampered feedback window should fail validation")
	}
}
