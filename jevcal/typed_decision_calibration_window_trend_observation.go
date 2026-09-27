package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
)

const (
	TypedDecisionCalibrationWindowTrendBound  = "BOUND"
	TypedDecisionCalibrationWindowTrendUnknown = "UNKNOWN"
	TypedDecisionCalibrationWindowTrendLowerError = "lower-error"
	TypedDecisionCalibrationWindowTrendHigherError = "higher-error"
	TypedDecisionCalibrationWindowTrendFlat = "flat"
)

type TypedDecisionCalibrationWindowTrendInput struct {
	Previous   TypedDecisionCalibrationWindowObservation
	Current    TypedDecisionCalibrationWindowObservation
	ComparedAt string
}

// TypedDecisionCalibrationWindowTrendObservation compares measurements only.
// It does not claim that a lower error is an authorized improvement.
type TypedDecisionCalibrationWindowTrendObservation struct {
	Status                 string
	PreviousEvidenceDigest string
	CurrentEvidenceDigest  string
	PreviousMeanError      float64
	CurrentMeanError       float64
	MeanErrorDelta         float64
	PreviousCoverage       float64
	CurrentCoverage        float64
	CoverageDelta          float64
	PreviousUnknownCount   int
	CurrentUnknownCount    int
	Direction              string
	ComparedAt             string
	FirstMismatch          string
	MissingStage           string
	EvidenceDigest         string
	IsReadOnly             bool
	CanExecute             bool
	CanAuthorize           bool
}

// ObserveTypedDecisionCalibrationWindowTrend compares two windows when
// evidence lineage is stable and otherwise preserves the first mismatch.
func ObserveTypedDecisionCalibrationWindowTrend(
	input TypedDecisionCalibrationWindowTrendInput,
) TypedDecisionCalibrationWindowTrendObservation {
	observation := TypedDecisionCalibrationWindowTrendObservation{
		Status:                 TypedDecisionCalibrationWindowTrendUnknown,
		PreviousEvidenceDigest: input.Previous.EvidenceDigest,
		CurrentEvidenceDigest:  input.Current.EvidenceDigest,
		PreviousMeanError:      input.Previous.MeanAbsoluteError,
		CurrentMeanError:       input.Current.MeanAbsoluteError,
		PreviousCoverage:       input.Previous.EvidenceCoverage,
		CurrentCoverage:        input.Current.EvidenceCoverage,
		PreviousUnknownCount:   input.Previous.ObservationCount - input.Previous.KnownObservationCount,
		CurrentUnknownCount:    input.Current.ObservationCount - input.Current.KnownObservationCount,
		Direction:              TypedDecisionCalibrationWindowTrendFlat,
		ComparedAt:             input.ComparedAt,
		FirstMismatch:          "windows",
		MissingStage:           "calibration-window-trend",
		IsReadOnly:             true,
		CanExecute:             false,
		CanAuthorize:           false,
	}
	if err := input.Previous.Validate(); err != nil {
		observation.FirstMismatch = "previous-window-integrity"
		return finalizeTypedDecisionCalibrationWindowTrend(observation)
	}
	if err := input.Current.Validate(); err != nil {
		observation.FirstMismatch = "current-window-integrity"
		return finalizeTypedDecisionCalibrationWindowTrend(observation)
	}
	if strings.TrimSpace(input.ComparedAt) == "" {
		observation.FirstMismatch = "comparison-time"
		return finalizeTypedDecisionCalibrationWindowTrend(observation)
	}
	if input.Previous.EvidencePrefixDigest != input.Current.EvidencePrefixDigest {
		observation.FirstMismatch = "evidence-prefix"
		return finalizeTypedDecisionCalibrationWindowTrend(observation)
	}
	observation.FirstMismatch = ""
	observation.MissingStage = ""
	observation.MeanErrorDelta = observation.CurrentMeanError - observation.PreviousMeanError
	observation.CoverageDelta = observation.CurrentCoverage - observation.PreviousCoverage
	switch {
	case observation.MeanErrorDelta < 0:
		observation.Direction = TypedDecisionCalibrationWindowTrendLowerError
	case observation.MeanErrorDelta > 0:
		observation.Direction = TypedDecisionCalibrationWindowTrendHigherError
	default:
		observation.Direction = TypedDecisionCalibrationWindowTrendFlat
	}
	observation.Status = TypedDecisionCalibrationWindowTrendBound
	return finalizeTypedDecisionCalibrationWindowTrend(observation)
}

func finalizeTypedDecisionCalibrationWindowTrend(
	observation TypedDecisionCalibrationWindowTrendObservation,
) TypedDecisionCalibrationWindowTrendObservation {
	observation.EvidenceDigest = typedDecisionCalibrationWindowTrendEvidenceDigest(observation)
	return observation
}

func (observation TypedDecisionCalibrationWindowTrendObservation) Validate() error {
	if observation.Status != TypedDecisionCalibrationWindowTrendBound &&
		observation.Status != TypedDecisionCalibrationWindowTrendUnknown {
		return fmt.Errorf("invalid typed decision calibration window trend status %q", observation.Status)
	}
	if !calibrationWindowDigestValid(observation.PreviousEvidenceDigest) ||
		!calibrationWindowDigestValid(observation.CurrentEvidenceDigest) ||
		!calibrationWindowDigestValid(observation.EvidenceDigest) ||
		strings.TrimSpace(observation.ComparedAt) == "" ||
		observation.PreviousUnknownCount < 0 ||
		observation.CurrentUnknownCount < 0 ||
		math.IsNaN(observation.PreviousMeanError) ||
		math.IsInf(observation.PreviousMeanError, 0) ||
		math.IsNaN(observation.CurrentMeanError) ||
		math.IsInf(observation.CurrentMeanError, 0) ||
		math.IsNaN(observation.MeanErrorDelta) ||
		math.IsInf(observation.MeanErrorDelta, 0) ||
		math.IsNaN(observation.PreviousCoverage) ||
		math.IsInf(observation.PreviousCoverage, 0) ||
		math.IsNaN(observation.CurrentCoverage) ||
		math.IsInf(observation.CurrentCoverage, 0) ||
		math.IsNaN(observation.CoverageDelta) ||
		math.IsInf(observation.CoverageDelta, 0) ||
		observation.PreviousCoverage < 0 || observation.PreviousCoverage > 1 ||
		observation.CurrentCoverage < 0 || observation.CurrentCoverage > 1 ||
		observation.MeanErrorDelta != observation.CurrentMeanError-observation.PreviousMeanError ||
		observation.CoverageDelta != observation.CurrentCoverage-observation.PreviousCoverage {
		return fmt.Errorf("typed decision calibration window trend arithmetic is invalid")
	}
	if !observation.IsReadOnly || observation.CanExecute || observation.CanAuthorize {
		return fmt.Errorf("typed decision calibration window trend crossed a capability boundary")
	}
	switch observation.Direction {
	case TypedDecisionCalibrationWindowTrendLowerError:
		if observation.MeanErrorDelta >= 0 {
			return fmt.Errorf("lower-error trend has no negative delta")
		}
	case TypedDecisionCalibrationWindowTrendHigherError:
		if observation.MeanErrorDelta <= 0 {
			return fmt.Errorf("higher-error trend has no positive delta")
		}
	case TypedDecisionCalibrationWindowTrendFlat:
		if observation.MeanErrorDelta != 0 {
			return fmt.Errorf("flat trend has a non-zero delta")
		}
	default:
		return fmt.Errorf("invalid typed decision calibration window trend direction")
	}
	if observation.Status == TypedDecisionCalibrationWindowTrendBound &&
		(observation.FirstMismatch != "" || observation.MissingStage != "") {
		return fmt.Errorf("bound trend retains a mismatch")
	}
	if observation.Status == TypedDecisionCalibrationWindowTrendUnknown &&
		(observation.FirstMismatch == "" || observation.MissingStage == "") {
		return fmt.Errorf("unknown trend lost its first mismatch")
	}
	if observation.EvidenceDigest != typedDecisionCalibrationWindowTrendEvidenceDigest(observation) {
		return fmt.Errorf("typed decision calibration window trend evidence digest mismatch")
	}
	return nil
}

func typedDecisionCalibrationWindowTrendEvidenceDigest(
	observation TypedDecisionCalibrationWindowTrendObservation,
) string {
	payload := fmt.Sprintf(
		"jev-typed-decision-calibration-window-trend|%s|%s|%s|%.9f|%.9f|%.9f|%.9f|%.9f|%.9f|%d|%d|%s|%s|%s|%s|%t|%t|%t",
		observation.Status,
		observation.PreviousEvidenceDigest,
		observation.CurrentEvidenceDigest,
		observation.PreviousMeanError,
		observation.CurrentMeanError,
		observation.MeanErrorDelta,
		observation.PreviousCoverage,
		observation.CurrentCoverage,
		observation.CoverageDelta,
		observation.PreviousUnknownCount,
		observation.CurrentUnknownCount,
		observation.Direction,
		observation.ComparedAt,
		observation.FirstMismatch,
		observation.MissingStage,
		observation.IsReadOnly,
		observation.CanExecute,
		observation.CanAuthorize,
	)
	digest := sha256.Sum256([]byte(payload))
	return "sha256:" + hex.EncodeToString(digest[:])
}