package jevcal

import "testing"

func TestDiscoverCapabilitiesOverviewIsBoundAndReadOnly(t *testing.T) {
	observation := DiscoverCapabilities(CapabilityDiscoveryInput{
		SourceVersion: "gooo-source-v1",
		ContractVersion: "jev-capability-discovery-v1",
		Query: "What can gooo do?",
	})
	if observation.Status != CapabilityDiscoveryBound || len(observation.Matches) != 5 {
		t.Fatalf("unexpected overview: %+v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("overview should validate: %v", err)
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
	if observation.Status != CapabilityDiscoveryUnknown || observation.FirstMismatch != "query" || observation.MissingStage != "capability_catalog" {
		t.Fatalf("unexpected unknown observation: %+v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("unknown observation should validate: %v", err)
	}
}

func TestDiscoverCapabilitiesRejectsTamperedDigest(t *testing.T) {
	observation := DiscoverCapabilities(CapabilityDiscoveryInput{
		SourceVersion: "gooo-source-v1",
		ContractVersion: "jev-capability-discovery-v1",
		Query: "show syntax completion and provenance",
	})
	observation.Matches[0].Summary = "tampered"
	if err := observation.Validate(); err == nil {
		t.Fatal("tampered capability discovery should fail validation")
	}
}
