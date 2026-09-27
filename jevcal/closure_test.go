package jevcal

import "testing"

func boundClosureEvidence() ClosureEvidence {
	evidence := ClosureEvidence{
		SourceVersion:            "source-v1",
		ContractVersion:          "contract-v1",
		Declaration:              "decl:runtime",
		IR:                       "ir:runtime",
		Generated:                "generated:runtime",
		Action:                  "action:runtime",
		ReverseObservation:      "reverse:runtime",
		EvidencePrefix:           "prefix:runtime",
		ActionStatus:             ClosureBound,
		ReverseObservationStatus: ClosureBound,
		MissingStageIndex:        -1,
	}
	evidence.ActionDigest = digest(evidence.SourceVersion, evidence.ContractVersion, evidence.Declaration, evidence.IR, evidence.Generated, evidence.Action, evidence.EvidencePrefix, "-1")
	evidence.ReverseDigest = digest(evidence.ActionDigest, evidence.ReverseObservation)
	return evidence
}

func TestObserveClosureBindsRuntimeEvidence(t *testing.T) {
	observation := ObserveClosure(boundClosureEvidence())
	if observation.Status != ClosureBound {
		t.Fatalf("status = %q, want BOUND", observation.Status)
	}
	if observation.ObservationDigest == "" {
		t.Fatal("observation digest is empty")
	}
}

func TestObserveClosurePreservesMissingStage(t *testing.T) {
	evidence := boundClosureEvidence()
	evidence.MissingStageIndex = 2
	observation := ObserveClosure(evidence)
	if observation.Status != ClosureUnknown || observation.MissingStage != "missing_stage_index" || observation.MissingStageIndex != 2 {
		t.Fatalf("missing stage was not preserved: %+v", observation)
	}
}

func TestObserveClosureRejectsTamperedReverseObservation(t *testing.T) {
	evidence := boundClosureEvidence()
	evidence.ReverseObservation = "reverse:tampered"
	observation := ObserveClosure(evidence)
	if observation.Status != ClosureUnknown || observation.Reason != "reverse observation digest mismatch" {
		t.Fatalf("tamper was not rejected: %+v", observation)
	}
}

func TestObserveClosureDefersUnboundProducer(t *testing.T) {
	evidence := boundClosureEvidence()
	evidence.ActionStatus = ClosureUnknown
	observation := ObserveClosure(evidence)
	if observation.Status != ClosureDeferred {
		t.Fatalf("status = %q, want DEFERRED", observation.Status)
	}
}