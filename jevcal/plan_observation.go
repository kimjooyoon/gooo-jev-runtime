package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type PlanObservationStatus string

const (
	PlanObservationBound    PlanObservationStatus = "BOUND"
	PlanObservationDeferred PlanObservationStatus = "DEFERRED"
	PlanObservationUnknown  PlanObservationStatus = "UNKNOWN"
)

// PlanApplicationObservationInput links a plan proposal to reverse-observed
// evidence without granting permission to apply the plan.
type PlanApplicationObservationInput struct {
	SourceVersion            string
	ContractVersion          string
	ProposalStatus           PlanBoundaryStatus
	ProposalPlanDigest       string
	ExpectedPlanDigest       string
	ReverseObservationStatus string
	ReverseObservationDigest string
	ProducerDeferred         bool
}

// PlanApplicationObservation is a read-only provenance result. It is not an
// execution result and cannot authorize a plan.
type PlanApplicationObservation struct {
	Status                   PlanObservationStatus
	SourceVersion            string
	ContractVersion          string
	PlanDigest               string
	ReverseObservationDigest string
	TargetStage              string
	Reason                   string
	ObservationDigest        string
	IsReadOnly               bool
	CanExecute               bool
	CanAuthorize             bool
	Edits                    []string
	Command                  string
}

func ObservePlanApplication(input PlanApplicationObservationInput) PlanApplicationObservation {
	observation := PlanApplicationObservation{
		Status:                   PlanObservationUnknown,
		SourceVersion:            input.SourceVersion,
		ContractVersion:          input.ContractVersion,
		PlanDigest:               input.ProposalPlanDigest,
		ReverseObservationDigest: input.ReverseObservationDigest,
		TargetStage:              "plan_application_observation",
		IsReadOnly:               true,
		CanExecute:               false,
		CanAuthorize:             false,
	}

	switch {
	case input.SourceVersion == "" || input.ContractVersion == "":
		observation.TargetStage = "identity"
		observation.Reason = "source or contract identity is missing"
	case input.ProposalPlanDigest == "":
		observation.TargetStage = "plan_digest"
		observation.Reason = "plan proposal digest is missing"
	case input.ExpectedPlanDigest == "":
		observation.TargetStage = "expected_plan_digest"
		observation.Reason = "expected plan digest is missing"
	case input.ProducerDeferred:
		observation.Status = PlanObservationDeferred
		observation.Reason = "plan observation producer is deferred"
	case input.ProposalStatus == PlanBoundaryDeferred:
		observation.Status = PlanObservationDeferred
		observation.Reason = "plan proposal is deferred"
	case input.ReverseObservationStatus == "DEFERRED":
		observation.Status = PlanObservationDeferred
		observation.Reason = "reverse observation is deferred"
	case input.ReverseObservationDigest == "":
		observation.TargetStage = "reverse_observation"
		observation.Reason = "reverse observation digest is missing"
	case input.ProposalPlanDigest != input.ExpectedPlanDigest:
		observation.TargetStage = "plan_digest_match"
		observation.Reason = "plan proposal digest does not match expected digest"
	case input.ProposalStatus != PlanBoundaryProposed:
		observation.TargetStage = "proposal_status"
		observation.Reason = "plan proposal status is not PROPOSED"
	case input.ReverseObservationStatus == "UNKNOWN":
		observation.TargetStage = "reverse_observation_status"
		observation.Reason = "reverse observation remains unresolved"
	case input.ReverseObservationStatus == "BOUND":
		observation.Status = PlanObservationBound
		observation.Reason = "plan proposal and reverse observation are aligned"
	default:
		observation.TargetStage = "reverse_observation_status"
		observation.Reason = "reverse observation status is not recognized"
	}

	observation.ObservationDigest = planObservationDigest(
		string(observation.Status),
		observation.SourceVersion,
		observation.ContractVersion,
		observation.PlanDigest,
		input.ExpectedPlanDigest,
		observation.ReverseObservationDigest,
		observation.TargetStage,
		observation.Reason,
	)
	return observation
}

func planObservationDigest(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}