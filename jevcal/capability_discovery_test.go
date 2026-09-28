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

func TestDiscoverCapabilitiesReturnsNaturalQueryEvidence(t *testing.T) {
	observation := DiscoverCapabilities(CapabilityDiscoveryInput{
		SourceVersion: "gooo-source-v1",
		ContractVersion: "jev-capability-discovery-v1",
		Query: "What can gooo do with provenance?",
	})
	if len(observation.NormalizedTerms) == 0 || len(observation.MatchedTerms) == 0 || observation.QueryDigest == "" || observation.ProvenanceDigest == "" {
		t.Fatalf("natural query evidence is incomplete: %+v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("natural query evidence should validate: %v", err)
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

func TestDiscoverCapabilitiesBindsDeclarationObservation(t *testing.T) {
	observation := DiscoverCapabilitiesForDeclaration(CapabilityDiscoveryInput{
		SourceVersion:   "gooo-source-v1",
		ContractVersion: "jev-capability-discovery-v1",
		Query:           "What can gooo do with this declaration?",
	}, "entity Usage\noperation observe\n")
	if !observation.DeclarationBound || observation.DeclarationSourceDigest == "" {
		t.Fatalf("declaration source was not bound: %+v", observation)
	}
	if len(observation.DeclarationObservedSignals) != 2 || observation.DeclarationObservedSignals[0] != "entity" || observation.DeclarationObservedSignals[1] != "operation" {
		t.Fatalf("unexpected declaration signals: %+v", observation.DeclarationObservedSignals)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("declaration-bound observation should validate: %v", err)
	}
}

func TestDiscoverCapabilitiesRejectsTamperedDeclarationBinding(t *testing.T) {
	observation := DiscoverCapabilitiesForDeclaration(CapabilityDiscoveryInput{
		SourceVersion:   "gooo-source-v1",
		ContractVersion: "jev-capability-discovery-v1",
		Query:           "show provenance",
	}, "entity Usage")
	observation.DeclarationSourceDigest = "sha256:tampered"
	if err := observation.Validate(); err == nil {
		t.Fatal("tampered declaration binding should fail validation")
	}
}

func TestDiscoverCapabilitiesSuggestedExamplesAreDiscoverable(t *testing.T) {
	for _, entry := range capabilityCatalog {
		if !entry.Safe {
			continue
		}
		observation := DiscoverCapabilities(CapabilityDiscoveryInput{
			SourceVersion:   "gooo-source-v1",
			ContractVersion: "jev-capability-discovery-v1",
			Query:           entry.ExampleQuery,
		})
		if observation.Status != CapabilityDiscoveryBound || len(observation.Matches) == 0 {
			t.Fatalf("catalog example is not discoverable: key=%q query=%q observation=%+v", entry.Key, entry.ExampleQuery, observation)
		}
		if err := observation.Validate(); err != nil {
			t.Fatalf("catalog example should validate: key=%q error=%v", entry.Key, err)
		}
	}
}
