package jevcal

import "testing"

const typedDecisionCredentialDigest = "sha256:1111111111111111111111111111111111111111111111111111111111111111"

func boundTypedDecisionSignalReceipt() TypedDecisionSignalReceipt {
	return ObserveTypedDecisionSignal(TypedDecisionSignalInput{
		SourceVersion:       "source-v1",
		ContractVersion:     "contract-v1",
		ModelRevision:       "jev-1.13",
		QuestionID:          "needs-review",
		RequestDigest:       typedDecisionRequestDigest,
		QuestionKind:        TypedDecisionQuestionNoul,
		SelectedValue:       "true",
		Probabilities:       map[string]float64{"true": 0.9, "false": 0.1},
		SelectedProbability: 0.9,
		Confidence:          0.95,
		AcceptanceThreshold: 0.8,
		NonAuthorizing:      true,
	})
}

func TestBindTypedDecisionSignalWorkloadIdentityBound(t *testing.T) {
	binding := BindTypedDecisionSignalWorkloadIdentity(TypedDecisionSignalWorkloadIdentityInput{
		SourceVersion:        "source-v1",
		ContractVersion:      "contract-v1",
		Receipt:              boundTypedDecisionSignalReceipt(),
		TrustDomain:          "example.org",
		SPIFFEID:             "spiffe://example.org/ns/default/sa/gooo",
		SVIDFormat:           TypedDecisionSignalSVIDX509,
		CredentialDigest:     typedDecisionCredentialDigest,
		ValidUntilUnix:       200,
		ObservedAtUnix:       100,
		NonAuthorizing:       true,
	})
	if binding.Status != TypedDecisionSignalIdentityBound {
		t.Fatalf("status = %s, want %s", binding.Status, TypedDecisionSignalIdentityBound)
	}
	if binding.CanExecute || binding.CanAuthorize {
		t.Fatalf("unsafe binding = %#v", binding)
	}
	if err := binding.Validate(); err != nil {
		t.Fatalf("bound identity binding invalid: %v", err)
	}
}

func TestBindTypedDecisionSignalWorkloadIdentityRejectsExpiredOrMalformedEvidence(t *testing.T) {
	input := TypedDecisionSignalWorkloadIdentityInput{
		SourceVersion:        "source-v1",
		ContractVersion:      "contract-v1",
		Receipt:              boundTypedDecisionSignalReceipt(),
		TrustDomain:          "example.org",
		SPIFFEID:             "spiffe://example.org/ns/default/sa/gooo",
		SVIDFormat:           TypedDecisionSignalSVIDJWT,
		CredentialDigest:     typedDecisionCredentialDigest,
		ValidUntilUnix:       100,
		ObservedAtUnix:       100,
		NonAuthorizing:       true,
	}
	expired := BindTypedDecisionSignalWorkloadIdentity(input)
	if expired.Status != TypedDecisionSignalIdentityUnknown || expired.FirstMismatch != "credential-expiry" {
		t.Fatalf("expired = %#v", expired)
	}
	if err := expired.Validate(); err != nil {
		t.Fatalf("expired binding invalid: %v", err)
	}

	input.SPIFFEID = "https://example.org/workload"
	input.ValidUntilUnix = 200
	malformed := BindTypedDecisionSignalWorkloadIdentity(input)
	if malformed.Status != TypedDecisionSignalIdentityUnknown || malformed.FirstMismatch != "spiffe-id" {
		t.Fatalf("malformed = %#v", malformed)
	}
}

func TestBindTypedDecisionSignalWorkloadIdentityRejectsTampering(t *testing.T) {
	binding := BindTypedDecisionSignalWorkloadIdentity(TypedDecisionSignalWorkloadIdentityInput{
		SourceVersion:        "source-v1",
		ContractVersion:      "contract-v1",
		Receipt:              boundTypedDecisionSignalReceipt(),
		TrustDomain:          "example.org",
		SPIFFEID:             "spiffe://example.org/ns/default/sa/gooo",
		SVIDFormat:           TypedDecisionSignalSVIDWIT,
		CredentialDigest:     typedDecisionCredentialDigest,
		ValidUntilUnix:       200,
		ObservedAtUnix:       100,
		NonAuthorizing:       true,
	})
	binding.IdentityDigest = "sha256:tampered"
	if err := binding.Validate(); err == nil {
		t.Fatal("expected tampered identity digest to fail validation")
	}
}

