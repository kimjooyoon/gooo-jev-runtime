package jevcal

import "testing"

func TestObserveJEVConfidenceGateAutoPreservesEvidence(t *testing.T) {
	observation := ObserveJEVConfidenceGate(0.96, 0.90, TypedDecisionConfidenceMethodCalibrated, "prefix-digest", "jev-1.13.0", -1)

	if observation.Status != JEVConfidenceGateAuto {
		t.Fatalf("status = %q, want %q", observation.Status, JEVConfidenceGateAuto)
	}
	if observation.EvidencePrefixDigest != "prefix-digest" {
		t.Fatalf("digest = %q, want prefix-digest", observation.EvidencePrefixDigest)
	}
	if observation.ConfidenceMethod != TypedDecisionConfidenceMethodCalibrated {
		t.Fatalf("method = %q, want %q", observation.ConfidenceMethod, TypedDecisionConfidenceMethodCalibrated)
	}
	if observation.SourceVersion != "jev-1.13.0" {
		t.Fatalf("version = %q, want jev-1.13.0", observation.SourceVersion)
	}
	if !observation.IsReadOnly || observation.CanExecute || observation.CanAuthorize {
		t.Fatal("confidence gate must remain read-only and non-authorizing")
	}
}

func TestObserveJEVConfidenceGateReviewKeepsNegativeBoundary(t *testing.T) {
	observation := ObserveJEVConfidenceGate(0.72, 0.90, TypedDecisionConfidenceMethodCalibrated, "prefix-digest", "jev-1.13.0", -1)

	if observation.Status != JEVConfidenceGateReview {
		t.Fatalf("status = %q, want %q", observation.Status, JEVConfidenceGateReview)
	}
}

func TestObserveJEVConfidenceGateUnknownForMissingEvidence(t *testing.T) {
	observation := ObserveJEVConfidenceGate(0.99, 0.90, TypedDecisionConfidenceMethodCalibrated, "", "jev-1.13.0", 4)

	if observation.Status != JEVConfidenceGateUnknown {
		t.Fatalf("status = %q, want %q", observation.Status, JEVConfidenceGateUnknown)
	}
	if observation.MissingStageIndex != 4 {
		t.Fatalf("missing stage = %d, want 4", observation.MissingStageIndex)
	}
}

func TestObserveJEVConfidenceGateUnknownForOutOfRangeConfidence(t *testing.T) {
	observation := ObserveJEVConfidenceGate(1.1, 0.90, TypedDecisionConfidenceMethodCalibrated, "prefix-digest", "jev-1.13.0", -1)

	if observation.Status != JEVConfidenceGateUnknown {
		t.Fatalf("status = %q, want %q", observation.Status, JEVConfidenceGateUnknown)
	}
}

func TestObserveJEVConfidenceGateUnknownForUnspecifiedMethod(t *testing.T) {
	observation := ObserveJEVConfidenceGate(0.99, 0.90, TypedDecisionConfidenceMethodUnspecified, "prefix-digest", "jev-1.13.0", -1)

	if observation.Status != JEVConfidenceGateUnknown {
		t.Fatalf("status = %q, want %q", observation.Status, JEVConfidenceGateUnknown)
	}
}
