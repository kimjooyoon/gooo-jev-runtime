package jevcal

import "testing"

func TestObserveJEVWorkloadIdentityBindsReadOnlyEvidence(t *testing.T) {
	capabilities := []string{"provenance.read", "lsp.observe"}
	observation := ObserveJEVWorkloadIdentity(
		"spiffe://example.test/ns/gooo/sa/jev",
		"jev-runtime",
		capabilities,
		"prefix-digest",
	)

	if observation.Status != JEVWorkloadIdentityBound {
		t.Fatalf("status = %q, want %q", observation.Status, JEVWorkloadIdentityBound)
	}
	if observation.SpiffeID != "spiffe://example.test/ns/gooo/sa/jev" {
		t.Fatalf("spiffe id = %q, want workload identity", observation.SpiffeID)
	}
	if len(observation.Capabilities) != 2 || observation.Capabilities[0] != "provenance.read" {
		t.Fatal("capability observations were not preserved")
	}
	if !observation.IsReadOnly || observation.CanExecute || observation.CanAuthorize {
		t.Fatal("identity observation must not execute or authorize")
	}
}

func TestObserveJEVWorkloadIdentityUnknownForMissingEvidence(t *testing.T) {
	observation := ObserveJEVWorkloadIdentity(
		"",
		"jev-runtime",
		nil,
		"prefix-digest",
	)

	if observation.Status != JEVWorkloadIdentityUnknown {
		t.Fatalf("status = %q, want %q", observation.Status, JEVWorkloadIdentityUnknown)
	}
}