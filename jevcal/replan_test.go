package jevcal

import "testing"

func replanEvidence() ClosureEvidence {
	evidence := boundClosureEvidence()
	return evidence
}

func TestProposeReplanDoesNotProposeAfterClosureBinds(t *testing.T) {
	previous := replanEvidence()
	current := replanEvidence()
	candidate := ProposeReplan(previous, current)
	if candidate.Status != ReplanNoAction {
		t.Fatalf("status = %q, want NO_ACTION", candidate.Status)
	}
	if candidate.PlanDigest == "" || !candidate.IsReadOnly || candidate.CanExecute || candidate.CanAuthorize {
		t.Fatalf("candidate crossed authority boundary or lacks digest: %+v", candidate)
	}
}

func TestProposeReplanTargetsMissingStage(t *testing.T) {
	previous := replanEvidence()
	current := replanEvidence()
	current.MissingStageIndex = 2
	candidate := ProposeReplan(previous, current)
	if candidate.Status != ReplanProposed || candidate.TargetStage != "missing_stage_index" || candidate.MissingStageIndex != 2 {
		t.Fatalf("unexpected candidate: %+v", candidate)
	}
}

func TestProposeReplanDefersUnboundProducer(t *testing.T) {
	previous := replanEvidence()
	current := replanEvidence()
	current.ActionStatus = ClosureDeferred
	candidate := ProposeReplan(previous, current)
	if candidate.Status != ReplanDeferred {
		t.Fatalf("status = %q, want DEFERRED", candidate.Status)
	}
}

func TestProposeReplanPreservesIdentityChangeAsUnknown(t *testing.T) {
	previous := replanEvidence()
	current := replanEvidence()
	current.SourceVersion = "source-v2"
	candidate := ProposeReplan(previous, current)
	if candidate.Status != ReplanUnknown || candidate.Reason != "source or contract identity changed" {
		t.Fatalf("identity change was not rejected: %+v", candidate)
	}
}