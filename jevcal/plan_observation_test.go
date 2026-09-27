package jevcal

import "testing"

func TestObservePlanApplicationBindsAlignedEvidence(t *testing.T) {
	observation := ObservePlanApplication(PlanApplicationObservationInput{
		SourceVersion:            "source-v1",
		ContractVersion:          "contract-v1",
		ProposalStatus:           PlanBoundaryProposed,
		ProposalPlanDigest:       "sha256:plan-v1",
		ExpectedPlanDigest:       "sha256:plan-v1",
		ReverseObservationStatus: "BOUND",
		ReverseObservationDigest: "sha256:reverse-v1",
	})
	if observation.Status != PlanObservationBound || observation.TargetStage != "plan_application_observation" || observation.ObservationDigest == "" {
		t.Fatalf("unexpected plan observation: %+v", observation)
	}
	if !observation.IsReadOnly || observation.CanExecute || observation.CanAuthorize || len(observation.Edits) != 0 || observation.Command != "" {
		t.Fatalf("plan observation crossed authority boundary: %+v", observation)
	}
}

func TestObservePlanApplicationPreservesDeferredEvidence(t *testing.T) {
	observation := ObservePlanApplication(PlanApplicationObservationInput{
		SourceVersion:            "source-v1",
		ContractVersion:          "contract-v1",
		ProposalStatus:           PlanBoundaryProposed,
		ProposalPlanDigest:       "sha256:plan-v1",
		ExpectedPlanDigest:       "sha256:plan-v1",
		ReverseObservationStatus: "DEFERRED",
		ReverseObservationDigest: "sha256:reverse-v1",
	})
	if observation.Status != PlanObservationDeferred || observation.Reason != "reverse observation is deferred" {
		t.Fatalf("deferred reverse observation was not preserved: %+v", observation)
	}
}

func TestObservePlanApplicationRejectsDigestMismatch(t *testing.T) {
	observation := ObservePlanApplication(PlanApplicationObservationInput{
		SourceVersion:            "source-v1",
		ContractVersion:          "contract-v1",
		ProposalStatus:           PlanBoundaryProposed,
		ProposalPlanDigest:       "sha256:plan-v1",
		ExpectedPlanDigest:       "sha256:plan-v2",
		ReverseObservationStatus: "BOUND",
		ReverseObservationDigest: "sha256:reverse-v1",
	})
	if observation.Status != PlanObservationUnknown || observation.TargetStage != "plan_digest_match" {
		t.Fatalf("digest mismatch was not preserved: %+v", observation)
	}
}

func TestObservePlanApplicationKeepsMissingStage(t *testing.T) {
	observation := ObservePlanApplication(PlanApplicationObservationInput{
		SourceVersion:       "source-v1",
		ContractVersion:     "contract-v1",
		ProposalStatus:      PlanBoundaryProposed,
		ExpectedPlanDigest:  "sha256:plan-v1",
	})
	if observation.Status != PlanObservationUnknown || observation.TargetStage != "plan_digest" || observation.ObservationDigest == "" {
		t.Fatalf("missing plan evidence was not preserved: %+v", observation)
	}
}