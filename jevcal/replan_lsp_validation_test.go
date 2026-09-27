package jevcal

import "testing"

func TestReplanProjectionValidatePreservesReadOnlyBoundary(t *testing.T) {
	projection := ProjectReplan(ReplanCandidate{
		Status:      ReplanProposed,
		TargetStage: "reverse_observation",
		PlanDigest:  "plan-v1",
	})
	if err := projection.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestReplanProjectionValidateRejectsCommandAndEdits(t *testing.T) {
	projection := ProjectReplan(ReplanCandidate{
		Status:     ReplanProposed,
		TargetStage: "reverse_observation",
		PlanDigest: "plan-v1",
	})
	projection.Command = "run"
	if err := projection.Validate(); err == nil {
		t.Fatal("replan LSP command was accepted")
	}
}

func TestReplanProjectionValidateRejectsMissingDigest(t *testing.T) {
	projection := ProjectReplan(ReplanCandidate{
		Status:      ReplanNoAction,
		PlanDigest:  "plan-v1",
	})
	projection.PlanDigest = ""
	if err := projection.Validate(); err == nil {
		t.Fatal("replan LSP projection without a plan digest was accepted")
	}
}

func TestReplanProjectionValidatePreservesAllStatuses(t *testing.T) {
	cases := []struct {
		name      string
		candidate ReplanCandidate
	}{
		{name: "no-action", candidate: ReplanCandidate{Status: ReplanNoAction, PlanDigest: "plan-stable"}},
		{name: "deferred", candidate: ReplanCandidate{Status: ReplanDeferred, PlanDigest: "plan-deferred"}},
		{name: "unknown", candidate: ReplanCandidate{Status: ReplanUnknown, PlanDigest: "plan-unknown"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if err := ProjectReplan(testCase.candidate).Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}