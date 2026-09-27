package jevcal

import "testing"

func TestObservePlanOutcomeWindowBindsAlignedProvenance(t *testing.T) {
	observation := ObservePlanOutcomeWindow(PlanOutcomeWindowObservationInput{
		SourceVersion:                     "source-v1",
		ContractVersion:                   "contract-v1",
		PlanObservationDigest:             "sha256:plan-v1",
		ExpectedPlanObservationDigest:     "sha256:plan-v1",
		OutcomeWindowRecordDigest:          "sha256:window-v1",
		ExpectedOutcomeWindowRecordDigest: "sha256:window-v1",
	})
	if observation.Status != PlanOutcomeWindowBound || observation.TargetStage != "plan_outcome_window" || observation.ChainDigest == "" {
		t.Fatalf("unexpected plan-outcome observation: %+v", observation)
	}
	if !observation.IsReadOnly || observation.CanExecute || observation.CanAuthorize || len(observation.Edits) != 0 || observation.Command != "" {
		t.Fatalf("observation crossed authority boundary: %+v", observation)
	}
}

func TestObservePlanOutcomeWindowPreservesDeferredProducer(t *testing.T) {
	observation := ObservePlanOutcomeWindow(PlanOutcomeWindowObservationInput{
		SourceVersion:                     "source-v1",
		ContractVersion:                   "contract-v1",
		PlanObservationDigest:             "sha256:plan-v1",
		ExpectedPlanObservationDigest:     "sha256:plan-v1",
		OutcomeWindowRecordDigest:          "sha256:window-v1",
		ExpectedOutcomeWindowRecordDigest: "sha256:window-v1",
		ProducerDeferred:                   true,
	})
	if observation.Status != PlanOutcomeWindowDeferred || observation.Reason != "plan-outcome producer is deferred" {
		t.Fatalf("deferred producer was not preserved: %+v", observation)
	}
}

func TestObservePlanOutcomeWindowPreservesRecordMismatch(t *testing.T) {
	observation := ObservePlanOutcomeWindow(PlanOutcomeWindowObservationInput{
		SourceVersion:                     "source-v1",
		ContractVersion:                   "contract-v1",
		PlanObservationDigest:             "sha256:plan-v1",
		ExpectedPlanObservationDigest:     "sha256:plan-v1",
		OutcomeWindowRecordDigest:          "sha256:window-v1",
		ExpectedOutcomeWindowRecordDigest: "sha256:window-v2",
	})
	if observation.Status != PlanOutcomeWindowUnknown || observation.TargetStage != "outcome_window_record_digest_match" {
		t.Fatalf("record mismatch was not preserved: %+v", observation)
	}
}

func TestObservePlanOutcomeWindowPreservesFirstMissingStage(t *testing.T) {
	observation := ObservePlanOutcomeWindow(PlanOutcomeWindowObservationInput{
		SourceVersion:         "source-v1",
		ContractVersion:       "contract-v1",
		PlanObservationDigest: "sha256:plan-v1",
	})
	if observation.Status != PlanOutcomeWindowUnknown || observation.TargetStage != "expected_plan_observation_digest" {
		t.Fatalf("first missing stage was not preserved: %+v", observation)
	}
}
