package query

import (
	"testing"

	"github.com/kimjooyoon/gooo-jev-runtime/jevcal"
)

func TestRenderCapabilityDiscoveryFollowUpPreservesEvidence(t *testing.T) {
	observation := jevcal.CapabilityDiscoveryObservation{
		Status:          jevcal.CapabilityDiscoveryBound,
		Matches:         []jevcal.CapabilityDiscoveryMatch{{Key: "generation"}, {Key: "provenance"}},
		QueryDigest:     "sha256:query",
		ProvenanceDigest: "sha256:provenance",
		DiscoveryDigest: "sha256:discovery",
	}
	response := RenderCapabilityDiscoveryFollowUp(observation)
	if response.Status != "BOUND" || response.QueryDigest != observation.QueryDigest || response.ProvenanceDigest != observation.ProvenanceDigest || response.DiscoveryDigest != observation.DiscoveryDigest {
		t.Fatalf("follow-up projection lost evidence: %+v", response)
	}
	if err := response.Validate(); err != nil {
		t.Fatalf("follow-up projection should validate: %v", err)
	}
}
