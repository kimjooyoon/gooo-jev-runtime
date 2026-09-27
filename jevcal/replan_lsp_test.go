package jevcal

import "testing"

func TestProjectReplanProposedIsReadOnly(t *testing.T) {
	projection := ProjectReplan(ReplanCandidate{
		Status:      ReplanProposed,
		TargetStage: "reverse_observation",
		PlanDigest:  "plan-v1",
	})
	if projection.Code != "gooo.replan.proposed" || !projection.IsActionable || projection.TargetStage != "reverse_observation" {
		t.Fatalf("unexpected proposal projection: %+v", projection)
	}
	if !projection.IsReadOnly || projection.CanExecute || projection.CanAuthorize || projection.Command != "" || len(projection.Edits) != 0 {
		t.Fatalf("projection crossed authority boundary: %+v", projection)
	}
}

func TestProjectReplanNoActionIsInformational(t *testing.T) {
	projection := ProjectReplan(ReplanCandidate{Status: ReplanNoAction, PlanDigest: "plan-stable"})
	if projection.Code != "gooo.replan.no_action" || projection.IsActionable || projection.Kind != "info" {
		t.Fatalf("unexpected no-action projection: %+v", projection)
	}
}

func TestProjectReplanPreservesDeferredAndUnknown(t *testing.T) {
	deferred := ProjectReplan(ReplanCandidate{Status: ReplanDeferred, PlanDigest: "plan-deferred"})
	unknown := ProjectReplan(ReplanCandidate{Status: ReplanUnknown, PlanDigest: "plan-unknown"})
	if deferred.Code != "gooo.replan.deferred" || unknown.Code != "gooo.replan.unknown" {
		t.Fatalf("statuses were not preserved: deferred=%+v unknown=%+v", deferred, unknown)
	}
	if !deferred.IsReadOnly || !unknown.IsReadOnly || deferred.CanExecute || unknown.CanAuthorize {
		t.Fatalf("projection crossed authority boundary: deferred=%+v unknown=%+v", deferred, unknown)
	}
}