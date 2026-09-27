package jevcal

import (
	"strings"
	"testing"
)

func calibrationWindowReceipt(probability float64, outcome bool, marker byte) TypedDecisionCalibrationObservation {
	return ObserveTypedDecisionCalibration(TypedDecisionCalibrationInput{
		Receipt: ObserveTypedDecisionSignal(TypedDecisionSignalInput{
			SourceVersion:       "gooo-jev-runtime",
			ContractVersion:     "jev-typed-decision/v1",
			ModelRevision:       "jev-latest",
			QuestionID:          "route",
			RequestDigest:       "sha256:" + strings.Repeat("1", 64),
			QuestionKind:        TypedDecisionQuestionChoice,
			SelectedValue:       "accept",
			Probabilities:       map[string]float64{"accept": probability, "review": 1 - probability},
			SelectedProbability: probability,
			Confidence:          0.9,
			AcceptanceThreshold: 0.7,
			NonAuthorizing:      true,
		}),
		OutcomeKnown:    true,
		ObservedOutcome: outcome,
		OutcomeDigest:   "sha256:" + strings.Repeat(string(marker), 64),
		WindowSize:      1,
		Tolerance:       0.2,
	})
}

func TestObserveTypedDecisionCalibrationWindowBound(t *testing.T) {
	observation := ObserveTypedDecisionCalibrationWindow(TypedDecisionCalibrationWindowInput{
		Observations: []TypedDecisionCalibrationObservation{
			calibrationWindowReceipt(0.8, false, '2'),
			calibrationWindowReceipt(0.9, true, '3'),
		},
		MinimumWindow: 2,
	})
	if observation.Status != TypedDecisionCalibrationWindowBound ||
		observation.KnownObservationCount != 2 ||
		observation.WithinToleranceCount != 1 ||
		observation.MeanAbsoluteError != 0.45 ||
		observation.EvidenceCoverage != 1 ||
		!calibrationWindowDigestValid(observation.EvidencePrefixDigest) {
		t.Fatalf("observation = %#v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestObserveTypedDecisionCalibrationWindowPreservesInsufficientWindow(t *testing.T) {
	observation := ObserveTypedDecisionCalibrationWindow(TypedDecisionCalibrationWindowInput{
		Observations:  []TypedDecisionCalibrationObservation{calibrationWindowReceipt(0.8, false, '2')},
		MinimumWindow: 2,
	})
	if observation.Status != TypedDecisionCalibrationWindowUnknown ||
		observation.FirstMismatch != "minimum-window" ||
		observation.MissingStage != "calibration-window" ||
		observation.EvidenceCoverage != 1 ||
		!calibrationWindowDigestValid(observation.EvidencePrefixDigest) {
		t.Fatalf("observation = %#v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestObserveTypedDecisionCalibrationWindowRejectsTampering(t *testing.T) {
	observation := ObserveTypedDecisionCalibrationWindow(TypedDecisionCalibrationWindowInput{
		Observations:  []TypedDecisionCalibrationObservation{calibrationWindowReceipt(0.8, false, '2')},
		MinimumWindow: 1,
	})
	observation.EvidenceDigest = "sha256:" + strings.Repeat("f", 64)
	if err := observation.Validate(); err == nil {
		t.Fatal("expected tampered window evidence to fail validation")
	}
}

func TestObserveTypedDecisionCalibrationWindowRejectsTamperedPrefixDigest(t *testing.T) {
	observation := ObserveTypedDecisionCalibrationWindow(TypedDecisionCalibrationWindowInput{
		Observations:  []TypedDecisionCalibrationObservation{calibrationWindowReceipt(0.8, false, '2')},
		MinimumWindow: 1,
	})
	observation.EvidencePrefixDigest = "sha256:" + strings.Repeat("f", 64)
	if err := observation.Validate(); err == nil {
		t.Fatal("expected tampered evidence prefix digest to fail validation")
	}
}
