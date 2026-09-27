package jevcal

import "testing"

func TestProjectJEVWorkloadIdentityObservationLSPBound(t *testing.T) {
	input := JEVWorkloadIdentityObservationLSPInput{
		Status:               JEVWorkloadIdentityBound,
		SpiffeID:             "spiffe://example.test/ns/gooo/sa/jev",
		Audience:             "jev-runtime",
		Capabilities:         []string{"provenance.read", "lsp.observe"},
		EvidencePrefixDigest: "sha256:prefix",
		IsReadOnly:           true,
	}
	projection := ProjectJEVWorkloadIdentityObservationLSP(input)

	if projection.Status != JEVWorkloadIdentityLSPInformation ||
		projection.Code != jevWorkloadIdentityLSPBound ||
		projection.Publishable ||
		projection.MissingStage != "" ||
		projection.IdentityDigest == "" {
		t.Fatalf("unexpected bound identity projection: %+v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestProjectJEVWorkloadIdentityObservationLSPPreservesUnknown(t *testing.T) {
	projection := ProjectJEVWorkloadIdentityObservationLSP(JEVWorkloadIdentityObservationLSPInput{
		Status:     JEVWorkloadIdentityUnknown,
		IsReadOnly: true,
	})

	if projection.Status != JEVWorkloadIdentityLSPError ||
		projection.Code != jevWorkloadIdentityLSPUnknown ||
		projection.MissingStage != "workload_identity_evidence" ||
		!projection.Publishable {
		t.Fatalf("unexpected unknown identity projection: %+v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func (p JEVWorkloadIdentityObservationLSP) NonAuthorizing() bool { return p.IsReadOnly && !p.CanExecute && !p.CanAuthorize }

func TestProjectJEVWorkloadIdentityObservationLSPRejectsMalformedSpiffeID(t *testing.T) {
	projection := ProjectJEVWorkloadIdentityObservationLSP(JEVWorkloadIdentityObservationLSPInput{
		Status:               JEVWorkloadIdentityBound,
		SpiffeID:             "https://untrusted.example/jev",
		Audience:             "jev-runtime",
		EvidencePrefixDigest: "sha256:prefix",
		IsReadOnly:           true,
	})

	if projection.Status != JEVWorkloadIdentityLSPError ||
		projection.Code != jevWorkloadIdentityLSPUnknown ||
		projection.MissingStage != "workload_identity_evidence" {
		t.Fatalf("malformed SPIFFE identity was accepted: %+v", projection)
	}
}

func TestProjectJEVWorkloadIdentityObservationLSPFailsClosedOnCapability(t *testing.T) {
	projection := ProjectJEVWorkloadIdentityObservationLSP(JEVWorkloadIdentityObservationLSPInput{
		Status:               JEVWorkloadIdentityBound,
		SpiffeID:             "spiffe://example.test/ns/gooo/sa/jev",
		Audience:             "jev-runtime",
		EvidencePrefixDigest: "sha256:prefix",
		IsReadOnly:           false,
		CanAuthorize:         true,
	})

	if projection.Status != JEVWorkloadIdentityLSPError ||
		projection.Code != jevWorkloadIdentityLSPBoundary ||
		projection.MissingStage != "capability_boundary" {
		t.Fatalf("capability boundary was not preserved: %+v", projection)
	}
	if projection.NonAuthorizing() {
		t.Fatal("unsafe projection must expose the unsafe boundary")
	}
}

func TestProjectJEVWorkloadIdentityObservationLSPRejectsTamperedDigest(t *testing.T) {
	projection := ProjectJEVWorkloadIdentityObservationLSP(JEVWorkloadIdentityObservationLSPInput{
		Status:               JEVWorkloadIdentityBound,
		SpiffeID:             "spiffe://example.test/ns/gooo/sa/jev",
		Audience:             "jev-runtime",
		Capabilities:         []string{"lsp.observe", "provenance.read"},
		EvidencePrefixDigest: "sha256:prefix",
		IsReadOnly:           true,
	})
	projection.IdentityDigest = "sha256:tampered"
	if err := projection.Validate(); err == nil {
		t.Fatal("Validate() accepted a tampered identity digest")
	}
}

func TestWorkloadIdentityDigestCanonicalizesCapabilities(t *testing.T) {
	first := workloadIdentityIdentityDigest(
		"spiffe://example.test/ns/gooo/sa/jev",
		"jev-runtime",
		[]string{"lsp.observe", "provenance.read"},
		"sha256:prefix",
	)
	second := workloadIdentityIdentityDigest(
		"spiffe://example.test/ns/gooo/sa/jev",
		"jev-runtime",
		[]string{"provenance.read", "lsp.observe"},
		"sha256:prefix",
	)
	if first != second {
		t.Fatalf("capability order changed identity digest: %q != %q", first, second)
	}
}