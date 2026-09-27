package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

type OutcomeWindowDeltaObservationStatus string

const (
	OutcomeWindowDeltaBound    OutcomeWindowDeltaObservationStatus = "BOUND"
	OutcomeWindowDeltaDeferred OutcomeWindowDeltaObservationStatus = "DEFERRED"
	OutcomeWindowDeltaUnknown  OutcomeWindowDeltaObservationStatus = "UNKNOWN"
)

// OutcomeWindowDeltaObservationInput records a signed metric change without
// claiming that an increase is success or a decrease is failure.
type OutcomeWindowDeltaObservationInput struct {
	SourceVersion        string
	ContractVersion      string
	EvidencePrefixDigest string
	ObservationDigest    string
	ExpectedOutcomeCount int
	ObservedOutcomeCount int
	ProducerDeferred     bool
}

// OutcomeWindowDeltaObservation is a read-only signed metric observation.
type OutcomeWindowDeltaObservation struct {
	Status               OutcomeWindowDeltaObservationStatus
	SourceVersion        string
	ContractVersion      string
	EvidencePrefixDigest string
	ObservationDigest    string
	ExpectedOutcomeCount int
	ObservedOutcomeCount int
	Delta                int
	TargetStage          string
	Reason               string
	RecordDigest         string
	IsReadOnly           bool
	CanExecute           bool
	CanAuthorize         bool
	Edits                []string
	Command              string
}

func ObserveOutcomeWindowDelta(
	input OutcomeWindowDeltaObservationInput,
) OutcomeWindowDeltaObservation {
	observation := OutcomeWindowDeltaObservation{
		Status:               OutcomeWindowDeltaUnknown,
		SourceVersion:        input.SourceVersion,
		ContractVersion:      input.ContractVersion,
		EvidencePrefixDigest: input.EvidencePrefixDigest,
		ObservationDigest:    input.ObservationDigest,
		ExpectedOutcomeCount: input.ExpectedOutcomeCount,
		ObservedOutcomeCount: input.ObservedOutcomeCount,
		Delta:                input.ObservedOutcomeCount - input.ExpectedOutcomeCount,
		TargetStage:          "outcome_window_delta",
		IsReadOnly:           true,
		CanExecute:           false,
		CanAuthorize:         false,
	}

	switch {
	case input.SourceVersion == "" || input.ContractVersion == "":
		observation.TargetStage = "identity"
		observation.Reason = "source or contract identity is missing"
	case input.EvidencePrefixDigest == "":
		observation.TargetStage = "evidence_prefix"
		observation.Reason = "evidence prefix digest is missing"
	case input.ObservationDigest == "":
		observation.TargetStage = "observation"
		observation.Reason = "outcome observation digest is missing"
	case input.ProducerDeferred:
		observation.Status = OutcomeWindowDeltaDeferred
		observation.Reason = "outcome window delta producer is deferred"
	case input.ExpectedOutcomeCount < 0 || input.ObservedOutcomeCount < 0:
		observation.TargetStage = "outcome_count"
		observation.Reason = "outcome counts must be non-negative"
	default:
		observation.Status = OutcomeWindowDeltaBound
		observation.Reason = "signed outcome-count delta is recorded without an improvement claim"
	}

	observation.RecordDigest = outcomeWindowDeltaObservationDigest(
		string(observation.Status),
		observation.SourceVersion,
		observation.ContractVersion,
		observation.EvidencePrefixDigest,
		observation.ObservationDigest,
		strconv.Itoa(observation.ExpectedOutcomeCount),
		strconv.Itoa(observation.ObservedOutcomeCount),
		strconv.Itoa(observation.Delta),
		observation.TargetStage,
		observation.Reason,
	)
	return observation
}

func outcomeWindowDeltaObservationDigest(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}
