package proposal

import "testing"

func TestBindToGoooPreservesContractIdentity(t *testing.T) {
	result := Build(Input{
		DomainID:             "gooo.provenance",
		MissingCapabilityIDs: []string{"gooo.provenance.compare"},
		ObservedSignals:      []string{SignalReviewMissing},
		EvidenceDigests:      []string{"sha256:observation"},
		SourceDigest:         "sha256:source",
		ContractDigest:       "sha256:contract",
	})
	binding := BindToGooo(result)
	if binding.Module != ContractModule || binding.Input != ContractInput {
		t.Fatalf("contract identity = %q/%q, want %q/%q", binding.Module, binding.Input, ContractModule, ContractInput)
	}
	if len(binding.MissingCapabilityIDs) != 1 || binding.Status != StatusProposed {
		t.Fatalf("binding lost proposal fields: %+v", binding)
	}
}