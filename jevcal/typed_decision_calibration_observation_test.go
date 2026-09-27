package jevcal

import (
	"strings"
	"testing"
)

func calibrationReceipt() TypedDecisionSignalReceipt {
	return ObserveTypedDecisionSignal(TypedDecisionSignalInput{
		SourceVersion:       "gooo-jev-runtime",
		ContractVersion:     "jev-typed-decision/v1",
		ModelRevision:       "jev-latest",
		QuestionID:          "route",
		RequestDigest:       "sha256:" + strings.Repeat("1", 64),
		QuestionKind:        TypedDecisionQuestionChoice,
		SelectedValue:       "accept",
		Probabilities:       map[string]float64{"accept": 0.8, "review": 0.2},
		SelectedProbability: 0.8,
		Confidence:          0.9,
		ConfidenceMethod:    TypedDecisionConfidenceMethodCalibrated,
		AcceptanceThreshold: 0.7,
		NonAuthorizing:      true,
	})
}

func TestObserveTypedDecisionCalibrationBound(t *testing.T) {
	observation := ObserveTypedDecisionCalibration(TypedDecisionCalibrationInput{
		Receipt:         calibrationReceipt(),
		OutcomeKnown:    true,
		ObservedOutcome: false,
		OutcomeDigest:   "sha256:" + strings.Repeat("2", 64),
		WindowSize:      10,
		Tolerance:       0.2,
	})
	if observation.Status != TypedDecisionCalibrationBound ||
		observation.AbsoluteError != 0.8 || observation.WithinTolerance {
		t.Fatalf("observation = %#v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestObserveTypedDecisionCalibrationPreservesUnknownOutcome(t *testing.T) {
	observation := ObserveTypedDecisionCalibration(TypedDecisionCalibrationInput{
		Receipt:   calibrationReceipt(),
		WindowSize: 10,
		Tolerance: 0.2,
	})
	if observation.Status != TypedDecisionCalibrationUnknown ||
		observation.FirstMismatch != "outcome" ||
		observation.MissingStage != "reverse_observation" {
		t.Fatalf("observation = %#v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestObserveTypedDecisionCalibrationRejectsCapabilityBoundary(t *testing.T) {
	receipt := calibrationReceipt()
	receipt.CanExecute = true
	observation := ObserveTypedDecisionCalibration(TypedDecisionCalibrationInput{
		Receipt:       receipt,
		OutcomeKnown:  true,
		OutcomeDigest: "sha256:" + strings.Repeat("2", 64),
		WindowSize:    1,
		Tolerance:     0.1,
	})
	if observation.Status != TypedDecisionCalibrationUnknown || observation.FirstMismatch != "capability-boundary" {
		t.Fatalf("observation = %#v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
