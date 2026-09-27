package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

type OutcomeWindowObservationStatus string

const (
	OutcomeWindowBound    OutcomeWindowObservationStatus = "BOUND"
	OutcomeWindowDeferred OutcomeWindowObservationStatus = "DEFERRED"
	OutcomeWindowUnknown  OutcomeWindowObservationStatus = "UNKNOWN"
)

// OutcomeWindowObservationInput contains the identity and metric evidence
// needed to compare a calibration outcome window without changing it.
type OutcomeWindowObservationInput struct {
	SourceVersion          string
	ContractVersion        string
	ChoiceSetDigest        string
	ExpectedChoiceSetDigest string
	EvidencePrefixDigest   string
	ObservationDigest      string
	OutcomeCount           int
	ProducerDeferred       bool
}

// OutcomeWindowObservation is a read-only metric observation, not a score
// update, execution decision, or authorization result.
type OutcomeWindowObservation struct {
	Status                 OutcomeWindowObservationStatus
	SourceVersion          string
	ContractVersion        string
	ChoiceSetDigest        string
	EvidencePrefixDigest   string
	ObservationDigest      string
	OutcomeCount           int
	TargetStage            string
	Reason                 string
	RecordDigest            string
	IsReadOnly             bool
	CanExecute             bool
	CanAuthorize           bool
	Edits                  []string
	Command                string
}

func ObserveOutcomeWindow(input OutcomeWindowObservationInput) OutcomeWindowObservation {
	observation := OutcomeWindowObservation{
		Status:               OutcomeWindowUnknown,
		SourceVersion:        input.SourceVersion,
		ContractVersion:      input.ContractVersion,
		ChoiceSetDigest:      input.ChoiceSetDigest,
		EvidencePrefixDigest: input.EvidencePrefixDigest,
		ObservationDigest:    input.ObservationDigest,
		OutcomeCount:         input.OutcomeCount,
		TargetStage:          "outcome_window",
		IsReadOnly:           true,
		CanExecute:           false,
		CanAuthorize:         false,
	}

	switch {
	case input.SourceVersion == "" || input.ContractVersion == "":
		observation.TargetStage = "identity"
		observation.Reason = "source or contract identity is missing"
	case input.ChoiceSetDigest == "":
		observation.TargetStage = "choice_set"
		observation.Reason = "choice-set digest is missing"
	case input.ExpectedChoiceSetDigest == "":
		observation.TargetStage = "expected_choice_set"
		observation.Reason = "expected choice-set digest is missing"
	case input.EvidencePrefixDigest == "":
		observation.TargetStage = "evidence_prefix"
		observation.Reason = "evidence prefix digest is missing"
	case input.ObservationDigest == "":
		observation.TargetStage = "observation"
		observation.Reason = "outcome observation digest is missing"
	case input.ProducerDeferred:
		observation.Status = OutcomeWindowDeferred
		observation.Reason = "outcome window producer is deferred"
	case input.ChoiceSetDigest != input.ExpectedChoiceSetDigest:
		observation.TargetStage = "choice_set_match"
		observation.Reason = "choice-set digest does not match expected digest"
	case input.OutcomeCount <= 0:
		observation.TargetStage = "outcome_count"
		observation.Reason = "outcome window has no positive observations"
	default:
		observation.Status = OutcomeWindowBound
		observation.Reason = "choice-set and evidence-prefix metrics are aligned"
	}

	observation.RecordDigest = outcomeWindowObservationDigest(
		string(observation.Status),
		observation.SourceVersion,
		observation.ContractVersion,
		observation.ChoiceSetDigest,
		input.ExpectedChoiceSetDigest,
		observation.EvidencePrefixDigest,
		observation.ObservationDigest,
		strconv.Itoa(observation.OutcomeCount),
		observation.TargetStage,
		observation.Reason,
	)
	return observation
}

func outcomeWindowObservationDigest(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}