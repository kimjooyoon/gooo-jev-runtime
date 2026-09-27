package jevcal

import (
	"math"
	"strings"
)

// JEVConfidenceGateObservationStatus identifies the safe routing boundary for a
// calibrated decision without claiming that the decision is correct.
type JEVConfidenceGateObservationStatus string

const (
	JEVConfidenceGateAuto    JEVConfidenceGateObservationStatus = "AUTO"
	JEVConfidenceGateReview  JEVConfidenceGateObservationStatus = "REVIEW"
	JEVConfidenceGateUnknown JEVConfidenceGateObservationStatus = "UNKNOWN"
)

// JEVConfidenceGateObservation is a read-only record of a confidence gate.
// Missing evidence never becomes an automatic decision.
type JEVConfidenceGateObservation struct {
	Status               JEVConfidenceGateObservationStatus
	Confidence           float64
	Threshold            float64
	ConfidenceMethod     string
	EvidencePrefixDigest string
	SourceVersion        string
	MissingStageIndex    int
	IsReadOnly           bool
	CanExecute           bool
	CanAuthorize         bool
}

// ObserveJEVConfidenceGate maps a bounded confidence value to an explicit
// routing boundary while preserving the evidence required to audit it.
func ObserveJEVConfidenceGate(
	confidence float64,
	threshold float64,
	confidenceMethod string,
	evidencePrefixDigest string,
	sourceVersion string,
	missingStageIndex int,
) JEVConfidenceGateObservation {
	confidenceMethod = strings.TrimSpace(confidenceMethod)
	observation := JEVConfidenceGateObservation{
		Status:               JEVConfidenceGateUnknown,
		Confidence:           confidence,
		Threshold:            threshold,
		ConfidenceMethod:     confidenceMethod,
		EvidencePrefixDigest: evidencePrefixDigest,
		SourceVersion:        sourceVersion,
		MissingStageIndex:    missingStageIndex,
		IsReadOnly:           true,
		CanExecute:           false,
		CanAuthorize:         false,
	}

	if math.IsNaN(confidence) || math.IsInf(confidence, 0) ||
		math.IsNaN(threshold) || math.IsInf(threshold, 0) ||
		confidence < 0 || confidence > 1 ||
		threshold < 0 || threshold > 1 ||
		!validTypedDecisionConfidenceMethod(confidenceMethod) ||
		confidenceMethod == TypedDecisionConfidenceMethodUnspecified ||
		evidencePrefixDigest == "" || sourceVersion == "" {
		return observation
	}

	if confidence >= threshold {
		observation.Status = JEVConfidenceGateAuto
		return observation
	}

	observation.Status = JEVConfidenceGateReview
	return observation
}
