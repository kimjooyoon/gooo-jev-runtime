package jevcal

import "testing"

func TestProjectWorkloadCredentialValidityLSPBound(t *testing.T) {
	observation := ObserveWorkloadCredentialValidity(workloadCredentialValidityTestInput())
	projection := ProjectWorkloadCredentialValidityLSP(observation)

	if projection.Status != WorkloadCredentialValidityLSPInformation ||
		projection.Code != workloadCredentialValidityLSPBound ||
		projection.Publishable ||
		projection.MissingStage != "" {
		t.Fatalf("unexpected bound credential validity LSP: %+v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestProjectWorkloadCredentialValidityLSPPreservesExpiryStage(t *testing.T) {
	input := workloadCredentialValidityTestInput()
	input.ObservedAtUnix = input.ExpiresAtUnix
	projection := ProjectWorkloadCredentialValidityLSP(ObserveWorkloadCredentialValidity(input))

	if projection.Status != WorkloadCredentialValidityLSPError ||
		projection.Code != workloadCredentialValidityLSPUnknown ||
		projection.MissingStage != "credential_expiry" ||
		!projection.Publishable {
		t.Fatalf("unexpected expiry projection: %+v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestProjectWorkloadCredentialValidityLSPPreservesDeferredState(t *testing.T) {
	input := workloadCredentialValidityTestInput()
	input.ProducerDeferred = true
	projection := ProjectWorkloadCredentialValidityLSP(ObserveWorkloadCredentialValidity(input))

	if projection.Status != WorkloadCredentialValidityLSPWarning ||
		projection.Code != workloadCredentialValidityLSPDeferred ||
		projection.MissingStage != "credential_validity_producer" {
		t.Fatalf("unexpected deferred projection: %+v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestProjectWorkloadCredentialValidityLSPRejectsTamperedObservation(t *testing.T) {
	observation := ObserveWorkloadCredentialValidity(workloadCredentialValidityTestInput())
	observation.CredentialDigest = "sha256:tampered"
	projection := ProjectWorkloadCredentialValidityLSP(observation)

	if projection.Status != WorkloadCredentialValidityLSPError ||
		projection.Code != workloadCredentialValidityLSPIntegrity ||
		projection.MissingStage != "observation_integrity" ||
		!projection.Publishable {
		t.Fatalf("tampered credential observation was projected: %+v", projection)
	}
}

func TestProjectWorkloadCredentialValidityLSPFailsClosedOnCapabilityBoundary(t *testing.T) {
	observation := ObserveWorkloadCredentialValidity(workloadCredentialValidityTestInput())
	observation.CanAuthorize = true
	projection := ProjectWorkloadCredentialValidityLSP(observation)

	if projection.Status != WorkloadCredentialValidityLSPError ||
		projection.Code != workloadCredentialValidityLSPBoundary ||
		projection.MissingStage != "capability_boundary" ||
		projection.CanAuthorize == false {
		t.Fatalf("capability boundary was not preserved: %+v", projection)
	}
}

func TestProjectWorkloadCredentialValidityLSPRejectsTamperedProjectionDigest(t *testing.T) {
	projection := ProjectWorkloadCredentialValidityLSP(
		ObserveWorkloadCredentialValidity(workloadCredentialValidityTestInput()),
	)
	projection.ProjectionDigest = "sha256:tampered"
	if err := projection.Validate(); err == nil {
		t.Fatal("Validate() accepted a tampered projection digest")
	}
}