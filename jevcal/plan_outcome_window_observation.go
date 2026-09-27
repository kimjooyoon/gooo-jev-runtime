package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type PlanOutcomeWindowObservationStatus string

const (
	PlanOutcomeWindowBound    PlanOutcomeWindowObservationStatus = "BOUND"
	PlanOutcomeWindowDeferred PlanOutcomeWindowObservationStatus = "DEFERRED"
	PlanOutcomeWindowUnknown  PlanOutcomeWindowObservationStatus = "UNKNOWN"
)

// PlanOutcomeWindowObservationInput links plan provenance to an outcome
// window without mutating metrics or granting authority.
type PlanOutcomeWindowObservationInput struct {
	SourceVersion                   string
	ContractVersion                 string
	PlanObservationDigest           string
	ExpectedPlanObservationDigest   string
	OutcomeWindowRecordDigest       string
	ExpectedOutcomeWindowRecordDigest string
	ProducerDeferred                bool
}

// PlanOutcomeWindowObservation is a read-only provenance-chain result.
type PlanOutcomeWindowObservation struct {
	Status                    PlanOutcomeWindowObservationStatus
	SourceVersion             string
	ContractVersion           string
	PlanObservationDigest     string
	OutcomeWindowRecordDigest string
	TargetStage               string
	Reason                    string
	ChainDigest               string
	IsReadOnly                bool
	CanExecute                bool
	CanAuthorize              bool
	Edits                     []string
	Command                   string
}

func ObservePlanOutcomeWindow(
	input PlanOutcomeWindowObservationInput,
) PlanOutcomeWindowObservation {
	observation := PlanOutcomeWindowObservation{
		Status:                    PlanOutcomeWindowUnknown,
		SourceVersion:             input.SourceVersion,
		ContractVersion:           input.ContractVersion,
		PlanObservationDigest:     input.PlanObservationDigest,
		OutcomeWindowRecordDigest: input.OutcomeWindowRecordDigest,
		TargetStage:               "plan_outcome_window",
		IsReadOnly:                true,
		CanExecute:                false,
		CanAuthorize:              false,
	}

	switch {
	case input.SourceVersion == "" || input.ContractVersion == "":
		observation.TargetStage = "identity"
		observation.Reason = "source or contract identity is missing"
	case input.PlanObservationDigest == "":
		observation.TargetStage = "plan_observation_digest"
		observation.Reason = "plan observation digest is missing"
	case input.ExpectedPlanObservationDigest == "":
		observation.TargetStage = "expected_plan_observation_digest"
		observation.Reason = "expected plan observation digest is missing"
	case input.OutcomeWindowRecordDigest == "":
		observation.TargetStage = "outcome_window_record_digest"
		observation.Reason = "outcome-window record digest is missing"
	case input.ExpectedOutcomeWindowRecordDigest == "":
		observation.TargetStage = "expected_outcome_window_record_digest"
		observation.Reason = "expected outcome-window record digest is missing"
	case input.ProducerDeferred:
		observation.Status = PlanOutcomeWindowDeferred
		observation.Reason = "plan-outcome producer is deferred"
	case input.PlanObservationDigest != input.ExpectedPlanObservationDigest:
		observation.TargetStage = "plan_observation_digest_match"
		observation.Reason = "plan observation digest does not match expected digest"
	case input.OutcomeWindowRecordDigest != input.ExpectedOutcomeWindowRecordDigest:
		observation.TargetStage = "outcome_window_record_digest_match"
		observation.Reason = "outcome-window record digest does not match expected digest"
	default:
		observation.Status = PlanOutcomeWindowBound
		observation.Reason = "plan observation and outcome-window record are aligned"
	}

	observation.ChainDigest = planOutcomeWindowObservationDigest(
		string(observation.Status),
		observation.SourceVersion,
		observation.ContractVersion,
		observation.PlanObservationDigest,
		input.ExpectedPlanObservationDigest,
		observation.OutcomeWindowRecordDigest,
		input.ExpectedOutcomeWindowRecordDigest,
		observation.TargetStage,
		observation.Reason,
	)
	return observation
}

func planOutcomeWindowObservationDigest(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}
