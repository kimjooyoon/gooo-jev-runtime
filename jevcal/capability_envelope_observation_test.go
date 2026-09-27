package jevcal

import "testing"

func TestObserveCapabilityEnvelopeBound(t *testing.T) {
	observation := ObserveCapabilityEnvelope(CapabilityEnvelopeObservationInput{
		SourceVersion:            "src-v1",
		ContractVersion:          "contract-v1",
		WorkloadIdentityStatus:   JEVWorkloadIdentityBound,
		EvidencePrefixDigest:     "sha256:prefix",
		PlanDigest:               "sha256:plan",
		NetworkAllowlistDigest:   "sha256:allowlist",
		RequestedCapabilities:    []string{"fs.read", "net.read"},
		ObservedCapabilities:     []string{"net.read", "fs.read"},
		ReverseObservationStatus: "BOUND",
		ReverseObservationDigest: "sha256:reverse",
	})

	if observation.Status != CapabilityEnvelopeBound {
		t.Fatalf("status = %q, want %q", observation.Status, CapabilityEnvelopeBound)
	}
	if !observation.IsReadOnly || observation.CanExecute || observation.CanAuthorize {
		t.Fatalf("unexpected execution or authorization capability: %+v", observation)
	}
	if observation.ObservationDigest == "" {
		t.Fatal("observation digest is empty")
	}
}

func TestObserveCapabilityEnvelopeStopsAtMissingCapability(t *testing.T) {
	observation := ObserveCapabilityEnvelope(CapabilityEnvelopeObservationInput{
		SourceVersion:            "src-v1",
		ContractVersion:          "contract-v1",
		WorkloadIdentityStatus:   JEVWorkloadIdentityBound,
		EvidencePrefixDigest:     "sha256:prefix",
		PlanDigest:               "sha256:plan",
		NetworkAllowlistDigest:   "sha256:allowlist",
		RequestedCapabilities:    []string{"fs.read", "net.write"},
		ObservedCapabilities:     []string{"fs.read"},
		ReverseObservationStatus: "BOUND",
		ReverseObservationDigest: "sha256:reverse",
	})

	if observation.Status != CapabilityEnvelopeUnknown {
		t.Fatalf("status = %q, want %q", observation.Status, CapabilityEnvelopeUnknown)
	}
	if observation.TargetStage != "capability_set" {
		t.Fatalf("target stage = %q, want capability_set", observation.TargetStage)
	}
	if observation.MissingCapability != "net.write" {
		t.Fatalf("missing capability = %q, want net.write", observation.MissingCapability)
	}
}

func TestObserveCapabilityEnvelopePreservesDeferredProducer(t *testing.T) {
	observation := ObserveCapabilityEnvelope(CapabilityEnvelopeObservationInput{
		SourceVersion:          "src-v1",
		ContractVersion:        "contract-v1",
		WorkloadIdentityStatus: JEVWorkloadIdentityBound,
		EvidencePrefixDigest:   "sha256:prefix",
		PlanDigest:             "sha256:plan",
		NetworkAllowlistDigest: "sha256:allowlist",
		ProducerDeferred:       true,
	})

	if observation.Status != CapabilityEnvelopeDeferred {
		t.Fatalf("status = %q, want %q", observation.Status, CapabilityEnvelopeDeferred)
	}
	if observation.TargetStage != "capability_envelope_observation" {
		t.Fatalf("target stage = %q, want capability_envelope_observation", observation.TargetStage)
	}
}

func TestObserveCapabilityEnvelopeRequiresIdentityAndReverseEvidence(t *testing.T) {
	missingIdentity := ObserveCapabilityEnvelope(CapabilityEnvelopeObservationInput{
		SourceVersion:          "src-v1",
		ContractVersion:        "contract-v1",
		EvidencePrefixDigest:   "sha256:prefix",
		PlanDigest:             "sha256:plan",
		NetworkAllowlistDigest: "sha256:allowlist",
	})
	if missingIdentity.TargetStage != "workload_identity" {
		t.Fatalf("target stage = %q, want workload_identity", missingIdentity.TargetStage)
	}

	missingReverse := ObserveCapabilityEnvelope(CapabilityEnvelopeObservationInput{
		SourceVersion:            "src-v1",
		ContractVersion:          "contract-v1",
		WorkloadIdentityStatus:   JEVWorkloadIdentityBound,
		EvidencePrefixDigest:     "sha256:prefix",
		PlanDigest:               "sha256:plan",
		NetworkAllowlistDigest:   "sha256:allowlist",
		ReverseObservationStatus: "BOUND",
	})
	if missingReverse.TargetStage != "reverse_observation" {
		t.Fatalf("target stage = %q, want reverse_observation", missingReverse.TargetStage)
	}
}

func TestObserveCapabilityEnvelopeDigestTracksCapabilityEvidence(t *testing.T) {
	input := CapabilityEnvelopeObservationInput{
		SourceVersion:            "src-v1",
		ContractVersion:          "contract-v1",
		WorkloadIdentityStatus:   JEVWorkloadIdentityBound,
		EvidencePrefixDigest:     "sha256:prefix",
		PlanDigest:               "sha256:plan",
		NetworkAllowlistDigest:   "sha256:allowlist",
		RequestedCapabilities:    []string{"fs.read"},
		ObservedCapabilities:     []string{"fs.read"},
		ReverseObservationStatus: "BOUND",
		ReverseObservationDigest: "sha256:reverse",
	}
	first := ObserveCapabilityEnvelope(input)
	input.ObservedCapabilities = []string{"fs.write"}
	second := ObserveCapabilityEnvelope(input)

	if first.ObservationDigest == second.ObservationDigest {
		t.Fatal("observation digest did not change with capability evidence")
	}
}
