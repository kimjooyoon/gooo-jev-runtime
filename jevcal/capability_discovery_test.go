package jevcal

import "testing"

func TestDiscoverCapabilitiesOverviewIsBoundAndReadOnly(t *testing.T) {
	observation := DiscoverCapabilities(CapabilityDiscoveryInput{
		SourceVersion: "gooo-source-v1",
		ContractVersion: "jev-capability-discovery-v1",
		Query: "What can gooo do?",
	})
	if observation.Status != CapabilityDiscoveryBound || len(observation.Matches) != 5 || len(observation.SuggestedQueries) == 0 {
		t.Fatalf("unexpected overview: %+v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("overview should validate: %v", err)
	}
	if !hasCapabilityExample(observation.SuggestedQueries, "How do I generate a canonical .gooo declaration?") {
		t.Fatalf("overview omitted an actionable example query: %+v", observation.SuggestedQueries)
	}
	if !hasCapabilityOperation(observation.Matches, "generation", "write_generated_declaration") {
		t.Fatalf("overview omitted an actionable next operation: %+v", observation.Matches)
	}
}

func TestDiscoverCapabilitiesPreservesDeferredBoundary(t *testing.T) {
	observation := DiscoverCapabilities(CapabilityDiscoveryInput{
		SourceVersion: "gooo-source-v1",
		ContractVersion: "jev-capability-discovery-v1",
		Query: "Can gooo execute and authorize a network workload?",
	})
	if observation.Status != CapabilityDiscoveryDeferred || observation.FirstMismatch != "execution_or_authorization_boundary" {
		t.Fatalf("unexpected deferred observation: %+v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("deferred observation should validate: %v", err)
	}
}

func TestDiscoverCapabilitiesUnknownRetainsCatalogBoundary(t *testing.T) {
	observation := DiscoverCapabilities(CapabilityDiscoveryInput{
		SourceVersion: "gooo-source-v1",
		ContractVersion: "jev-capability-discovery-v1",
		Query: "quantum breakfast compiler",
	})
	if observation.Status != CapabilityDiscoveryUnknown || observation.FirstMismatch != "query" || observation.MissingStage != "capability_catalog" || len(observation.SuggestedQueries) == 0 {
		t.Fatalf("unexpected unknown observation: %+v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("unknown observation should validate: %v", err)
	}
}

func hasCapabilityOperation(matches []CapabilityDiscoveryMatch, key, expected string) bool {
	for _, match := range matches {
		if match.Key == key && match.NextOperation == expected {
			return true
		}
	}
	return false
}

func hasCapabilityExample(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func TestDiscoverCapabilitiesRejectsTamperedDigest(t *testing.T) {
	observation := DiscoverCapabilities(CapabilityDiscoveryInput{
		SourceVersion: "gooo-source-v1",
		ContractVersion: "jev-capability-discovery-v1",
		Query: "show syntax completion and provenance",
	})
	observation.Matches[0].ExampleQuery = "tampered"
	if err := observation.Validate(); err == nil {
		t.Fatal("tampered capability discovery should fail validation")
	}
}
