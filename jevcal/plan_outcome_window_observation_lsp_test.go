package jevcal

import "testing"

func TestProjectPlanOutcomeWindowObservationLSPProjectsBound(t *testing.T) {
	observation := ObservePlanOutcomeWindow(PlanOutcomeWindowObservationInput{
		SourceVersion:                     "source-v1",
		ContractVersion:                   "contract-v1",
		PlanObservationDigest:             "sha256:plan-v1",
		ExpectedPlanObservationDigest:     "sha256:plan-v1",
		OutcomeWindowRecordDigest:          "sha256:window-v1",
		ExpectedOutcomeWindowRecordDigest: "sha256:window-v1",
	})
	projection := ProjectPlanOutcomeWindowObservationLSP(observation)
	if projection.Status != PlanOutcomeWindowBound || projection.Code != "jev.plan_outcome_window.bound" || projection.Severity != "Information" {
		t.Fatalf("unexpected bound projection: %+v", projection)
	}
	if !projection.IsReadOnly || projection.CanExecute || projection.CanAuthorize || len(projection.Edits) != 0 || projection.Command != "" {
		t.Fatalf("projection crossed authority boundary: %+v", projection)
	}
}

func TestProjectPlanOutcomeWindowObservationLSPProjectsDeferred(t *testing.T) {
	observation := ObservePlanOutcomeWindow(PlanOutcomeWindowObservationInput{
		SourceVersion:                     "source-v1",
		ContractVersion:                   "contract-v1",
		PlanObservationDigest:             "sha256:plan-v1",
		ExpectedPlanObservationDigest:     "sha256:plan-v1",
		OutcomeWindowRecordDigest:          "sha256:window-v1",
		ExpectedOutcomeWindowRecordDigest: "sha256:window-v1",
		ProducerDeferred:                   true,
	})
	projection := ProjectPlanOutcomeWindowObservationLSP(observation)
	if projection.Status != PlanOutcomeWindowDeferred || projection.Code != "jev.plan_outcome_window.deferred" || projection.Severity != "Hint" {
		t.Fatalf("unexpected deferred projection: %+v", projection)
	}
}

func TestProjectPlanOutcomeWindowObservationLSPPreservesMissingDigest(t *testing.T) {
	projection := ProjectPlanOutcomeWindowObservationLSP(PlanOutcomeWindowObservation{
		Status:      PlanOutcomeWindowUnknown,
		TargetStage: "outcome_window_record_digest",
		Reason:      "outcome-window record digest is missing",
	})
	if projection.Code != "jev.plan_outcome_window.unknown" || projection.TargetStage != "chain_digest" || projection.Severity != "Error" {
		t.Fatalf("missing chain digest was not preserved: %+v", projection)
	}
}
