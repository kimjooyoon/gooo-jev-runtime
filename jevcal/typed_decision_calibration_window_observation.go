package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
)

// TypedDecisionCalibrationWindowBound and Unknown describe evidence
// availability, not model correctness or improvement.
const (
	TypedDecisionCalibrationWindowBound  = "BOUND"
	TypedDecisionCalibrationWindowUnknown = "UNKNOWN"
)

// TypedDecisionCalibrationWindowInput aggregates already observed calibration
// receipts without executing or re-evaluating any decision.
type TypedDecisionCalibrationWindowInput struct {
	Observations  []TypedDecisionCalibrationObservation
	MinimumWindow int
}

// TypedDecisionCalibrationWindowObservation is a read-only sample-window
// summary. Its mean error is a measurement, not a quality claim.
type TypedDecisionCalibrationWindowObservation struct {
	Status                 string
	ObservationCount       int
	KnownObservationCount  int
	WithinToleranceCount   int
	MeanAbsoluteError      float64
	EvidenceCoverage       float64
	MinimumWindow          int
	FirstMismatch           string
	MissingStage            string
	EvidencePrefixDigest    string
	EvidenceDigest          string
	IsReadOnly              bool
	CanExecute              bool
	CanAuthorize            bool
}

// ObserveTypedDecisionCalibrationWindow summarizes a bounded observation
// window and preserves the first unavailable stage.
func ObserveTypedDecisionCalibrationWindow(
	input TypedDecisionCalibrationWindowInput,
) TypedDecisionCalibrationWindowObservation {
	observation := TypedDecisionCalibrationWindowObservation{
		Status:                TypedDecisionCalibrationWindowUnknown,
		ObservationCount:      len(input.Observations),
		MinimumWindow:         input.MinimumWindow,
		FirstMismatch:         "observations",
		MissingStage:          "calibration-window",
		EvidencePrefixDigest:  typedDecisionCalibrationWindowEvidencePrefixDigest(nil),
		IsReadOnly:            true,
		CanExecute:             false,
		CanAuthorize:           false,
	}
	if len(input.Observations) == 0 {
		observation.FirstMismatch = "observations"
		return finalizeTypedDecisionCalibrationWindow(observation)
	}
	if input.MinimumWindow <= 0 {
		observation.FirstMismatch = "minimum-window"
		return finalizeTypedDecisionCalibrationWindow(observation)
	}

	var errorSum float64
	evidencePrefix := make([]string, 0, len(input.Observations))
	for index, item := range input.Observations {
		if err := item.Validate(); err != nil {
			observation.FirstMismatch = fmt.Sprintf("observation[%d]-integrity", index)
			observation.MissingStage = "typed_decision_calibration"
			observation.EvidenceCoverage = float64(observation.KnownObservationCount) / float64(observation.ObservationCount)
			observation.EvidencePrefixDigest = typedDecisionCalibrationWindowEvidencePrefixDigest(evidencePrefix)
			return finalizeTypedDecisionCalibrationWindow(observation)
		}
		evidencePrefix = append(evidencePrefix, item.EvidenceDigest)
		if item.Status != TypedDecisionCalibrationBound {
			observation.FirstMismatch = fmt.Sprintf("observation[%d]", index)
			observation.MissingStage = item.MissingStage
			if observation.MissingStage == "" {
				observation.MissingStage = "typed_decision_calibration"
			}
			observation.EvidenceCoverage = float64(observation.KnownObservationCount) / float64(observation.ObservationCount)
			observation.EvidencePrefixDigest = typedDecisionCalibrationWindowEvidencePrefixDigest(evidencePrefix)
			return finalizeTypedDecisionCalibrationWindow(observation)
		}
		observation.KnownObservationCount++
		if item.WithinTolerance {
			observation.WithinToleranceCount++
		}
		errorSum += item.AbsoluteError
	}
	observation.EvidenceCoverage = float64(observation.KnownObservationCount) / float64(observation.ObservationCount)
	if len(input.Observations) < input.MinimumWindow {
		observation.FirstMismatch = "minimum-window"
		observation.MissingStage = "calibration-window"
		observation.EvidencePrefixDigest = typedDecisionCalibrationWindowEvidencePrefixDigest(evidencePrefix)
		return finalizeTypedDecisionCalibrationWindow(observation)
	}
	observation.Status = TypedDecisionCalibrationWindowBound
	observation.FirstMismatch = ""
	observation.MissingStage = ""
	observation.EvidencePrefixDigest = typedDecisionCalibrationWindowEvidencePrefixDigest(evidencePrefix)
	observation.MeanAbsoluteError = errorSum / float64(observation.KnownObservationCount)
	return finalizeTypedDecisionCalibrationWindow(observation)
}

func finalizeTypedDecisionCalibrationWindow(
	observation TypedDecisionCalibrationWindowObservation,
) TypedDecisionCalibrationWindowObservation {
	observation.EvidenceDigest = typedDecisionCalibrationWindowEvidenceDigest(observation)
	return observation
}

// Validate checks the window's evidence arithmetic and capability boundary.
// It does not interpret the summary as correctness or improvement.
func (observation TypedDecisionCalibrationWindowObservation) Validate() error {
	if observation.Status != TypedDecisionCalibrationWindowBound &&
		observation.Status != TypedDecisionCalibrationWindowUnknown {
		return fmt.Errorf("invalid typed decision calibration window status %q", observation.Status)
	}
	if observation.ObservationCount < 0 ||
		observation.KnownObservationCount < 0 ||
		observation.KnownObservationCount > observation.ObservationCount ||
		observation.WithinToleranceCount < 0 ||
		observation.WithinToleranceCount > observation.KnownObservationCount ||
		observation.MinimumWindow <= 0 ||
		math.IsNaN(observation.MeanAbsoluteError) ||
		math.IsInf(observation.MeanAbsoluteError, 0) ||
		observation.MeanAbsoluteError < 0 ||
		observation.MeanAbsoluteError > 1 ||
		math.IsNaN(observation.EvidenceCoverage) ||
		math.IsInf(observation.EvidenceCoverage, 0) ||
		observation.EvidenceCoverage < 0 ||
		observation.EvidenceCoverage > 1 {
		return fmt.Errorf("typed decision calibration window arithmetic is invalid")
	}
	if !observation.IsReadOnly || observation.CanExecute || observation.CanAuthorize {
		return fmt.Errorf("typed decision calibration window crossed a capability boundary")
	}
	if !calibrationWindowDigestValid(observation.EvidencePrefixDigest) {
		return fmt.Errorf("typed decision calibration window evidence prefix digest is invalid")
	}
	if observation.EvidenceDigest != typedDecisionCalibrationWindowEvidenceDigest(observation) {
		return fmt.Errorf("typed decision calibration window evidence digest mismatch")
	}
	if observation.Status == TypedDecisionCalibrationWindowBound {
		if observation.ObservationCount < observation.MinimumWindow ||
			observation.KnownObservationCount != observation.ObservationCount ||
			observation.FirstMismatch != "" ||
			observation.MissingStage != "" ||
			observation.EvidenceCoverage != 1 {
			return fmt.Errorf("bound typed decision calibration window is incomplete")
		}
	}
	if observation.Status == TypedDecisionCalibrationWindowUnknown &&
		(observation.FirstMismatch == "" || observation.MissingStage == "") {
		return fmt.Errorf("unknown typed decision calibration window lost its first mismatch")
	}
	return nil
}

func typedDecisionCalibrationWindowEvidenceDigest(
	observation TypedDecisionCalibrationWindowObservation,
) string {
	payload := fmt.Sprintf(
		"jev-typed-decision-calibration-window|%s|%d|%d|%d|%.9f|%.9f|%d|%s|%s|%s|%t|%t|%t",
		observation.Status,
		observation.ObservationCount,
		observation.KnownObservationCount,
		observation.WithinToleranceCount,
		observation.MeanAbsoluteError,
		observation.EvidenceCoverage,
		observation.MinimumWindow,
		observation.FirstMismatch,
		observation.MissingStage,
		observation.EvidencePrefixDigest,
		observation.IsReadOnly,
		observation.CanExecute,
		observation.CanAuthorize,
	)
	digest := sha256.Sum256([]byte(payload))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func typedDecisionCalibrationWindowEvidencePrefixDigest(evidenceDigests []string) string {
	payload := "jev-typed-decision-calibration-window-prefix|" + strings.Join(evidenceDigests, "|")
	digest := sha256.Sum256([]byte(payload))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func calibrationWindowDigestValid(value string) bool {
	const prefix = "sha256:"
	if !strings.HasPrefix(value, prefix) || len(value) != len(prefix)+64 {
		return false
	}
	_, err := hex.DecodeString(value[len(prefix):])
	return err == nil
}
