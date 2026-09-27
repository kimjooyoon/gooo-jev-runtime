package jevcal

import "testing"

const typedDecisionRequestDigest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"

func TestObserveTypedDecisionSignalBoundCandidateNeverAuthorizes(t *testing.T) {
	receipt := ObserveTypedDecisionSignal(TypedDecisionSignalInput{
		SourceVersion:       "source-v1",
		ContractVersion:     "contract-v1",
		ModelRevision:       "jev-1.13",
		QuestionID:          "needs-review",
		RequestDigest:       typedDecisionRequestDigest,
		QuestionKind:        TypedDecisionQuestionNoul,
		SelectedValue:       "true",
		Probabilities:       map[string]float64{"true": 0.9, "false": 0.1},
		SelectedProbability: 0.9,
		Confidence:          0.95,
		AcceptanceThreshold: 0.8,
		NonAuthorizing:      true,
	})
	if receipt.Status != TypedDecisionSignalBound || !receipt.ThresholdMet || receipt.ReviewRequired {
		t.Fatalf("receipt = %#v", receipt)
	}
	if receipt.CanExecute || receipt.CanAuthorize || !receipt.NonExecuting || !receipt.NonAuthorizing {
		t.Fatalf("unsafe receipt = %#v", receipt)
	}
	if err := receipt.Validate(); err != nil {
		t.Fatalf("bound receipt invalid: %v", err)
	}
}

func TestObserveTypedDecisionSignalPreservesReviewWhenThresholdMisses(t *testing.T) {
	receipt := ObserveTypedDecisionSignal(TypedDecisionSignalInput{
		SourceVersion:       "source-v1",
		ContractVersion:     "contract-v1",
		ModelRevision:       "jev-1.13",
		QuestionID:          "route",
		RequestDigest:       typedDecisionRequestDigest,
		QuestionKind:        TypedDecisionQuestionChoice,
		SelectedValue:       "human",
		Probabilities:       map[string]float64{"human": 0.62, "automate": 0.38},
		SelectedProbability: 0.62,
		Confidence:          0.7,
		AcceptanceThreshold: 0.8,
		NonAuthorizing:      true,
	})
	if receipt.Status != TypedDecisionSignalBound || receipt.ThresholdMet || !receipt.ReviewRequired {
		t.Fatalf("receipt = %#v", receipt)
	}
	if err := receipt.Validate(); err != nil {
		t.Fatalf("review receipt invalid: %v", err)
	}
}

func TestObserveTypedDecisionSignalRejectsInvalidDistributionAndBoundary(t *testing.T) {
	receipt := ObserveTypedDecisionSignal(TypedDecisionSignalInput{
		SourceVersion:       "source-v1",
		ContractVersion:     "contract-v1",
		ModelRevision:       "jev-1.13",
		QuestionID:          "route",
		RequestDigest:       typedDecisionRequestDigest,
		QuestionKind:        TypedDecisionQuestionScore,
		SelectedValue:       "high",
		Probabilities:       map[string]float64{"high": 0.8, "low": 0.1},
		SelectedProbability: 0.8,
		Confidence:          0.8,
		AcceptanceThreshold: 0.8,
		NonAuthorizing:      true,
	})
	if receipt.Status != TypedDecisionSignalUnknown || receipt.FirstMismatch != "probability-normalization" {
		t.Fatalf("invalid distribution = %#v", receipt)
	}
	if err := receipt.Validate(); err != nil {
		t.Fatalf("unknown receipt invalid: %v", err)
	}

	boundary := ObserveTypedDecisionSignal(TypedDecisionSignalInput{
		SourceVersion:       "source-v1",
		ContractVersion:     "contract-v1",
		ModelRevision:       "jev-1.13",
		QuestionID:          "route",
		RequestDigest:       typedDecisionRequestDigest,
		QuestionKind:        TypedDecisionQuestionChoice,
		SelectedValue:       "human",
		Probabilities:       map[string]float64{"human": 1},
		SelectedProbability: 1,
		Confidence:          1,
		AcceptanceThreshold: 0.5,
		NonAuthorizing:      false,
	})
	if boundary.Status != TypedDecisionSignalUnknown || boundary.FirstMismatch != "authorization-boundary" {
		t.Fatalf("boundary = %#v", boundary)
	}
}

func TestObserveTypedDecisionSignalRejectsTamperedReceipt(t *testing.T) {
	receipt := ObserveTypedDecisionSignal(TypedDecisionSignalInput{
		SourceVersion:       "source-v1",
		ContractVersion:     "contract-v1",
		ModelRevision:       "jev-1.13",
		QuestionID:          "route",
		RequestDigest:       typedDecisionRequestDigest,
		QuestionKind:        TypedDecisionQuestionChoice,
		SelectedValue:       "human",
		Probabilities:       map[string]float64{"human": 0.8, "automate": 0.2},
		SelectedProbability: 0.8,
		Confidence:          0.9,
		AcceptanceThreshold: 0.8,
		NonAuthorizing:      true,
	})
	receipt.EvidenceDigest = "sha256:tampered"
	if err := receipt.Validate(); err == nil {
		t.Fatal("expected tampered receipt to fail validation")
	}
}

