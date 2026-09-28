package proposal

import "testing"

func TestEnvelopeBindsNonExecutingBoundary(t *testing.T) {
	result := Build(Input{
		DomainID:             "gooo.provenance",
		MissingCapabilityIDs: []string{"gooo.provenance.compare"},
		ObservedSignals:      []string{SignalReviewMissing},
		EvidenceDigests:      []string{"sha256:observation"},
		SourceDigest:         "sha256:source",
		ContractDigest:       "sha256:contract",
	})
	envelope := NewEnvelope(result)
	if !envelope.NonExecuting || !envelope.NonAuthorizing {
		t.Fatal("proposal envelope crossed an execution or authorization boundary")
	}
	digest, err := EnvelopeDigest(result)
	if err != nil || len(digest) < len("sha256:") {
		t.Fatalf("envelope digest = %q, err = %v", digest, err)
	}
}