package jevcal

import "testing"

func TestProjectPlanObservationLSPProjectsBoundEvidence(t *testing.T) {
	observation := ObservePlanApplication(PlanApplicationObservationInput{
		SourceVersion:            "source-v1",
		ContractVersion:          "contract-v1",
		ProposalStatus:           PlanBoundaryProposed,
		ProposalPlanDigest:       "sha256:plan-v1",
		ExpectedPlanDigest:       "sha256:plan-v1",
		ReverseObservationStatus: "BOUND",
		ReverseObservationDigest: "sha256:reverse-v1",
	})
	projection := ProjectPlanObservationLSP(observation)
	if projection.Status != PlanObservationBound || projection.Code != "jev.plan_application.bound" || projection.Severity != "Information" {
		t.Fatalf("unexpected bound projection: %+v", projection)
	}
	if !projection.IsReadOnly || projection.CanExecute || projection.CanAuthorize || len(projection.Edits) != 0 || projection.Command != "" {
		t.Fatalf("projection crossed authority boundary: %+v", projection)
	}
}

func TestProjectPlanObservationLSPProjectsDeferredEvidence(t *testing.T) {
	observation := ObservePlanApplication(PlanApplicationObservationInput{
		SourceVersion:            "source-v1",
		ContractVersion:          "contract-v1",
		ProposalStatus:           PlanBoundaryProposed,
		ProposalPlanDigest:       "sha256:plan-v1",
		ExpectedPlanDigest:       "sha256:plan-v1",
		ReverseObservationStatus: "DEFERRED",
		ReverseObservationDigest: "sha256:reverse-v1",
	})
	projection := ProjectPlanObservationLSP(observation)
	if projection.Status != PlanObservationDeferred || projection.Code != "jev.plan_application.deferred" || projection.Severity != "Hint" {
		t.Fatalf("unexpected deferred projection: %+v", projection)
	}
}

func TestProjectPlanObservationLSPPreservesMissingDigest(t *testing.T) {
	projection := ProjectPlanObservationLSP(PlanApplicationObservation{
		Status:       PlanObservationUnknown,
		TargetStage:  "reverse_observation",
		Reason:       "reverse observation digest is missing",
		IsReadOnly:   true,
	})
	if projection.Code != "jev.plan_application.unknown" || projection.TargetStage != "observation_digest" || projection.Severity != "Error" {
		t.Fatalf("missing observation digest was not preserved: %+v", projection)
	}
}