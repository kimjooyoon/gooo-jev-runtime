package jevcal

import "testing"

func TestObserveJEVSelfImprovementCandidateBindsCompleteChain(t *testing.T) {
	observation := ObserveJEVSelfImprovementCandidate(
		"source-digest",
		"ir-digest",
		"generated-digest",
		"reverse-digest",
		0.25,
		-0.10,
	)

	if observation.Status != JEVSelfImprovementCandidateBound {
		t.Fatalf("status = %q, want %q", observation.Status, JEVSelfImprovementCandidateBound)
	}
	if observation.MissingStageIndex != -1 {
		t.Fatalf("missing stage = %d, want -1", observation.MissingStageIndex)
	}
	if observation.ExpectedDelta != 0.25 || observation.ObservedDelta != -0.10 {
		t.Fatal("signed candidate deltas were not preserved")
	}
	if !observation.IsReadOnly || observation.CanExecute || observation.CanAuthorize {
		t.Fatal("candidate observation must remain read-only and non-authorizing")
	}
}

func TestObserveJEVSelfImprovementCandidateKeepsFirstMissingStage(t *testing.T) {
	observation := ObserveJEVSelfImprovementCandidate(
		"source-digest",
		"ir-digest",
		"",
		"reverse-digest",
		0.25,
		0.10,
	)

	if observation.Status != JEVSelfImprovementCandidateUnknown {
		t.Fatalf("status = %q, want %q", observation.Status, JEVSelfImprovementCandidateUnknown)
	}
	if observation.MissingStageIndex != 2 {
		t.Fatalf("missing stage = %d, want 2", observation.MissingStageIndex)
	}
}

func TestObserveJEVSelfImprovementCandidateMissingSourceWins(t *testing.T) {
	observation := ObserveJEVSelfImprovementCandidate(
		"",
		"ir-digest",
		"generated-digest",
		"reverse-digest",
		0,
		0,
	)

	if observation.MissingStageIndex != 0 {
		t.Fatalf("missing stage = %d, want 0", observation.MissingStageIndex)
	}
}