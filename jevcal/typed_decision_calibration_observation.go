package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
)

const (
	TypedDecisionCalibrationBound   = "BOUND"
	TypedDecisionCalibrationUnknown = "UNKNOWN"
)

// TypedDecisionCalibrationInput joins one typed decision signal to an observed
// outcome. It measures calibration evidence and never grants authority.
type TypedDecisionCalibrationInput struct {
	Receipt         TypedDecisionSignalReceipt
	OutcomeKnown    bool
	ObservedOutcome bool
	OutcomeDigest   string
	WindowSize      int
	Tolerance       float64
}

// TypedDecisionCalibrationObservation is a read-only comparison between a
// selected probability and a later observed outcome.
type TypedDecisionCalibrationObservation struct {
	Status               string
	SourceVersion        string
	ContractVersion      string
	ModelRevision        string
	QuestionID           string
	SignalEvidenceDigest string
	SelectedProbability  float64
	Confidence           float64
	OutcomeKnown         bool
	ObservedOutcome      bool
	OutcomeDigest        string
	WindowSize           int
	Tolerance            float64
	AbsoluteError        float64
	WithinTolerance      bool
	FirstMismatch        string
	MissingStage         string
	EvidenceDigest       string
	IsReadOnly           bool
	CanExecute           bool
	CanAuthorize         bool
}

// ObserveTypedDecisionCalibration records a bounded calibration observation.
// It does not claim that the model is correct or that a route is safe.
func ObserveTypedDecisionCalibration(
	input TypedDecisionCalibrationInput,
) TypedDecisionCalibrationObservation {
	observation := TypedDecisionCalibrationObservation{
		Status:               TypedDecisionCalibrationUnknown,
		SourceVersion:        input.Receipt.SourceVersion,
		ContractVersion:      input.Receipt.ContractVersion,
		ModelRevision:        input.Receipt.ModelRevision,
		QuestionID:           input.Receipt.QuestionID,
		SignalEvidenceDigest: input.Receipt.EvidenceDigest,
		SelectedProbability:  input.Receipt.SelectedProbability,
		Confidence:           input.Receipt.Confidence,
		OutcomeKnown:         input.OutcomeKnown,
		ObservedOutcome:      input.ObservedOutcome,
		OutcomeDigest:        input.OutcomeDigest,
		WindowSize:           input.WindowSize,
		Tolerance:            input.Tolerance,
		FirstMismatch:        "receipt",
		MissingStage:         "typed_decision_signal",
		IsReadOnly:           true,
		CanExecute:           false,
		CanAuthorize:         false,
	}

	switch {
	case !input.Receipt.NonExecuting || !input.Receipt.NonAuthorizing ||
		input.Receipt.CanExecute || input.Receipt.CanAuthorize:
		observation.FirstMismatch = "capability-boundary"
		observation.MissingStage = "capability-boundary"
	case input.Receipt.Validate() != nil:
		observation.FirstMismatch = "receipt-integrity"
		observation.MissingStage = "typed_decision_signal"
	case input.Receipt.Status != TypedDecisionSignalBound:
		observation.FirstMismatch = input.Receipt.FirstMismatch
		if observation.FirstMismatch == "" {
			observation.FirstMismatch = "typed_decision_signal"
		}
		observation.MissingStage = "typed_decision_signal"
	case !input.OutcomeKnown:
		observation.FirstMismatch = "outcome"
		observation.MissingStage = "reverse_observation"
	case !calibrationDigestValid(input.OutcomeDigest):
		observation.FirstMismatch = "outcome-digest"
		observation.MissingStage = "reverse_observation"
	case input.WindowSize <= 0:
		observation.FirstMismatch = "sample-window"
		observation.MissingStage = "calibration-window"
	case !finiteCalibrationUnit(input.Tolerance):
		observation.FirstMismatch = "tolerance"
		observation.MissingStage = "calibration-window"
	default:
		target := 0.0
		if input.ObservedOutcome {
			target = 1
		}
		observation.AbsoluteError = math.Abs(input.Receipt.SelectedProbability - target)
		observation.WithinTolerance = observation.AbsoluteError <= input.Tolerance
		observation.Status = TypedDecisionCalibrationBound
		observation.FirstMismatch = ""
		observation.MissingStage = ""
	}

	observation.EvidenceDigest = typedDecisionCalibrationEvidenceDigest(observation)
	return observation
}

// Validate verifies the observation chain but does not turn a passing
// calibration window into an improvement or authorization decision.
func (observation TypedDecisionCalibrationObservation) Validate() error {
	if observation.Status != TypedDecisionCalibrationBound && observation.Status != TypedDecisionCalibrationUnknown {
		return fmt.Errorf("invalid typed decision calibration status %q", observation.Status)
	}
	if !observation.IsReadOnly || observation.CanExecute || observation.CanAuthorize {
		return fmt.Errorf("typed decision calibration crossed an execution or authorization boundary")
	}
	if observation.EvidenceDigest != typedDecisionCalibrationEvidenceDigest(observation) {
		return fmt.Errorf("typed decision calibration evidence digest mismatch")
	}
	if observation.Status == TypedDecisionCalibrationUnknown {
		if observation.FirstMismatch == "" || observation.MissingStage == "" {
			return fmt.Errorf("unknown typed decision calibration lost its first mismatch")
		}
		return nil
	}
	if !calibrationDigestValid(observation.SignalEvidenceDigest) ||
		!calibrationDigestValid(observation.OutcomeDigest) ||
		observation.SourceVersion == "" ||
		observation.ContractVersion == "" ||
		observation.ModelRevision == "" ||
		observation.QuestionID == "" ||
		!observation.OutcomeKnown ||
		observation.WindowSize <= 0 ||
		!finiteCalibrationUnit(observation.SelectedProbability) ||
		!finiteCalibrationUnit(observation.Confidence) ||
		!finiteCalibrationUnit(observation.Tolerance) ||
		observation.FirstMismatch != "" ||
		observation.MissingStage != "" {
		return fmt.Errorf("bound typed decision calibration is incomplete")
	}
	target := 0.0
	if observation.ObservedOutcome {
		target = 1
	}
	if math.Abs(math.Abs(observation.SelectedProbability-target)-observation.AbsoluteError) > 1e-9 ||
		observation.WithinTolerance != (observation.AbsoluteError <= observation.Tolerance) {
		return fmt.Errorf("typed decision calibration error does not match its evidence")
	}
	return nil
}

func finiteCalibrationUnit(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1
}

func calibrationDigestValid(value string) bool {
	const prefix = "sha256:"
	if !strings.HasPrefix(value, prefix) || len(value) != len(prefix)+64 {
		return false
	}
	_, err := hex.DecodeString(value[len(prefix):])
	return err == nil
}

func typedDecisionCalibrationEvidenceDigest(
	observation TypedDecisionCalibrationObservation,
) string {
	payload := fmt.Sprintf(
		"jev-typed-decision-calibration|%s|%s|%s|%s|%s|%0.9f|%0.9f|%t|%t|%s|%d|%0.9f|%0.9f|%t|%s|%s|%t|%t|%t",
		observation.Status,
		observation.SourceVersion,
		observation.ContractVersion,
		observation.ModelRevision,
		observation.QuestionID,
		observation.SelectedProbability,
		observation.Confidence,
		observation.OutcomeKnown,
		observation.ObservedOutcome,
		observation.OutcomeDigest,
		observation.WindowSize,
		observation.Tolerance,
		observation.AbsoluteError,
		observation.WithinTolerance,
		observation.FirstMismatch,
		observation.MissingStage,
		observation.IsReadOnly,
		observation.CanExecute,
		observation.CanAuthorize,
	)
	digest := sha256.Sum256([]byte(payload))
	return "sha256:" + hex.EncodeToString(digest[:])
}