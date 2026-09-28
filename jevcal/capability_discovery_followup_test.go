package jevcal

import "testing"

func TestSuggestCapabilityDiscoveryFollowUpsKeepsBoundDiscoveryReadOnly(t *testing.T) {
	followUp := SuggestCapabilityDiscoveryFollowUps(CapabilityDiscoveryObservation{
		Status: CapabilityDiscoveryBound,
		Matches: []CapabilityDiscoveryMatch{
			{Key: "generation"},
			{Key: "provenance"},
		},
	})
	if len(followUp.Questions) != 2 || followUp.FirstBoundary != "" || !followUp.IsReadOnly {
		t.Fatalf("unexpected bound follow-up: %+v", followUp)
	}
	if err := followUp.Validate(); err != nil {
		t.Fatalf("bound follow-up should validate: %v", err)
	}
}

func TestSuggestCapabilityDiscoveryFollowUpsPreservesDeferredBoundary(t *testing.T) {
	followUp := SuggestCapabilityDiscoveryFollowUps(CapabilityDiscoveryObservation{
		Status:       CapabilityDiscoveryDeferred,
		MissingStage: "explicit_external_boundary",
	})
	if followUp.FirstBoundary != "explicit_external_boundary" || len(followUp.Questions) != 1 {
		t.Fatalf("unexpected deferred follow-up: %+v", followUp)
	}
	if err := followUp.Validate(); err != nil {
		t.Fatalf("deferred follow-up should validate: %v", err)
	}
}

func TestSuggestCapabilityDiscoveryFollowUpsAsksForIdentityBeforeUnknownQuery(t *testing.T) {
	followUp := SuggestCapabilityDiscoveryFollowUps(CapabilityDiscoveryObservation{
		Status: CapabilityDiscoveryUnknown,
		Query:  "show capabilities",
	})
	if followUp.FirstBoundary != "capability_discovery" || len(followUp.Questions) != 1 {
		t.Fatalf("unexpected unknown follow-up: %+v", followUp)
	}
	if err := followUp.Validate(); err != nil {
		t.Fatalf("unknown follow-up should validate: %v", err)
	}
}
