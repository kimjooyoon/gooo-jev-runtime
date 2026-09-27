package jevcal

import "testing"

func TestObserveDecisionRouteReverseEvidenceBound(t *testing.T) {
	observation := ObserveDecisionRouteReverseEvidence(DecisionRouteReverseObservationInput{
		SourceVersion:        "src-v1",
		ContractVersion:      "contract-v1",
		RouteStatus:          "REVIEW_CANDIDATE",
		RouteRecordDigest:    "sha256:record",
		DecisionDigest:       "sha256:decision",
		ExpectedRouteStatus:  "REVIEW_CANDIDATE",
		ObservedRouteStatus:  "REVIEW_CANDIDATE",
		OutcomeDigest:        "sha256:outcome",
	})

	if observation.Status != DecisionRouteReverseBound {
		t.Fatalf("status = %q, want %q", observation.Status, DecisionRouteReverseBound)
	}
	if !observation.ReviewRequired {
		t.Fatal("review candidate must preserve review requirement")
	}
	if !observation.IsReadOnly || observation.CanExecute || observation.CanAuthorize {
		t.Fatal("reverse observation crossed an authority boundary")
	}
	if observation.ObservationDigest == "" {
		t.Fatal("observation digest is empty")
	}
}

func TestObserveDecisionRouteReverseEvidenceRejectsRouteMismatch(t *testing.T) {
	observation := ObserveDecisionRouteReverseEvidence(DecisionRouteReverseObservationInput{
		SourceVersion:        "src-v1",
		ContractVersion:      "contract-v1",
		RouteStatus:          "ACCEPT_CANDIDATE",
		RouteRecordDigest:    "sha256:record",
		DecisionDigest:       "sha256:decision",
		ExpectedRouteStatus:  "ACCEPT_CANDIDATE",
		ObservedRouteStatus:  "REVIEW_CANDIDATE",
		OutcomeDigest:        "sha256:outcome",
	})

	if observation.Status != DecisionRouteReverseUnknown {
		t.Fatalf("status = %q, want %q", observation.Status, DecisionRouteReverseUnknown)
	}
	if observation.TargetStage != "route_status_match" {
		t.Fatalf("target stage = %q, want route_status_match", observation.TargetStage)
	}
}

func TestObserveDecisionRouteReverseEvidencePreservesDeferred(t *testing.T) {
	observation := ObserveDecisionRouteReverseEvidence(DecisionRouteReverseObservationInput{
		SourceVersion:        "src-v1",
		ContractVersion:      "contract-v1",
		RouteRecordDigest:    "sha256:record",
		DecisionDigest:       "sha256:decision",
		ExpectedRouteStatus:  "ACCEPT_CANDIDATE",
		ObservedRouteStatus:  "ACCEPT_CANDIDATE",
		OutcomeDigest:        "sha256:outcome",
		ProducerDeferred:     true,
	})

	if observation.Status != DecisionRouteReverseDeferred {
		t.Fatalf("status = %q, want %q", observation.Status, DecisionRouteReverseDeferred)
	}
}

func TestObserveDecisionRouteReverseEvidenceRequiresOutcome(t *testing.T) {
	observation := ObserveDecisionRouteReverseEvidence(DecisionRouteReverseObservationInput{
		SourceVersion:        "src-v1",
		ContractVersion:      "contract-v1",
		RouteRecordDigest:    "sha256:record",
		DecisionDigest:       "sha256:decision",
		ExpectedRouteStatus:  "ABSTAIN_CANDIDATE",
		ObservedRouteStatus:  "ABSTAIN_CANDIDATE",
	})

	if observation.TargetStage != "outcome" || observation.Status != DecisionRouteReverseUnknown {
		t.Fatalf("unexpected missing outcome observation: %+v", observation)
	}
}

func TestObserveDecisionRouteReverseEvidenceDigestTracksOutcome(t *testing.T) {
	input := DecisionRouteReverseObservationInput{
		SourceVersion:        "src-v1",
		ContractVersion:      "contract-v1",
		RouteStatus:          "DEFERRED",
		RouteRecordDigest:    "sha256:record",
		DecisionDigest:       "sha256:decision",
		ExpectedRouteStatus:  "DEFERRED",
		ObservedRouteStatus:  "DEFERRED",
		OutcomeDigest:        "sha256:outcome",
	}
	first := ObserveDecisionRouteReverseEvidence(input)
	input.OutcomeDigest = "sha256:tampered"
	second := ObserveDecisionRouteReverseEvidence(input)

	if first.ObservationDigest == second.ObservationDigest {
		t.Fatal("observation digest did not change with outcome evidence")
	}
}
