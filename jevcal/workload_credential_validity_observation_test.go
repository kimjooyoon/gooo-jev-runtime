package jevcal

import "testing"

func workloadCredentialValidityTestInput() WorkloadCredentialValidityInput {
	return WorkloadCredentialValidityInput{
		SourceVersion:        "source-v1",
		ContractVersion:      "contract-v1",
		WorkloadSpiffeID:     "spiffe://example.test/ns/gooo/sa/jev",
		IssuerSpiffeID:       "spiffe://example.test/trust/issuer",
		Audience:             "jev-runtime",
		CredentialDigest:     "sha256:credential",
		EvidencePrefixDigest: "sha256:prefix",
		IssuedAtUnix:         90,
		ExpiresAtUnix:        200,
		ObservedAtUnix:       100,
		RevocationStatus:     "NOT_REVOKED",
	}
}

func TestObserveWorkloadCredentialValidityBindsCompleteEvidence(t *testing.T) {
	observation := ObserveWorkloadCredentialValidity(workloadCredentialValidityTestInput())
	if observation.Status != WorkloadCredentialValidityBound ||
		observation.MissingStage != "" ||
		observation.ValiditySignal != workloadCredentialValidityBoundSignal {
		t.Fatalf("unexpected bound credential validity: %+v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestObserveWorkloadCredentialValidityPreservesExpiryStage(t *testing.T) {
	input := workloadCredentialValidityTestInput()
	input.ObservedAtUnix = 200
	observation := ObserveWorkloadCredentialValidity(input)

	if observation.Status != WorkloadCredentialValidityUnknown ||
		observation.MissingStage != "credential_expiry" ||
		observation.TargetStage != "credential_expiry" {
		t.Fatalf("unexpected expiry observation: %+v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestObserveWorkloadCredentialValidityPreservesRevocationStage(t *testing.T) {
	input := workloadCredentialValidityTestInput()
	input.RevocationStatus = "REVOKED"
	observation := ObserveWorkloadCredentialValidity(input)

	if observation.Status != WorkloadCredentialValidityUnknown ||
		observation.MissingStage != "revocation_status" {
		t.Fatalf("unexpected revocation observation: %+v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestObserveWorkloadCredentialValidityPreservesDeferredProducer(t *testing.T) {
	input := workloadCredentialValidityTestInput()
	input.ProducerDeferred = true
	observation := ObserveWorkloadCredentialValidity(input)

	if observation.Status != WorkloadCredentialValidityDeferred ||
		observation.MissingStage != "" {
		t.Fatalf("unexpected deferred observation: %+v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestObserveWorkloadCredentialValidityRejectsMalformedIdentity(t *testing.T) {
	input := workloadCredentialValidityTestInput()
	input.IssuerSpiffeID = "https://untrusted.example/issuer"
	observation := ObserveWorkloadCredentialValidity(input)

	if observation.Status != WorkloadCredentialValidityUnknown ||
		observation.MissingStage != "issuer_identity" {
		t.Fatalf("unexpected malformed issuer observation: %+v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestObserveWorkloadCredentialValidityRejectsTamperedDigest(t *testing.T) {
	observation := ObserveWorkloadCredentialValidity(workloadCredentialValidityTestInput())
	observation.CredentialDigest = "sha256:tampered"
	if err := observation.Validate(); err == nil {
		t.Fatal("Validate() accepted a tampered credential observation")
	}
}