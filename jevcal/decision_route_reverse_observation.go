package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type DecisionRouteReverseObservationStatus string

const (
	DecisionRouteReverseBound    DecisionRouteReverseObservationStatus = "BOUND"
	DecisionRouteReverseDeferred DecisionRouteReverseObservationStatus = "DEFERRED"
	DecisionRouteReverseUnknown  DecisionRouteReverseObservationStatus = "UNKNOWN"
)

// DecisionRouteReverseObservationInput links a route record to observed route
// and outcome evidence without claiming the decision was correct.
type DecisionRouteReverseObservationInput struct {
	SourceVersion       string
	ContractVersion     string
	RouteStatus         string
	RouteRecordDigest   string
	DecisionDigest      string
	ExpectedRouteStatus string
	ObservedRouteStatus string
	OutcomeDigest       string
	ProducerDeferred    bool
}

// DecisionRouteReverseObservation is a read-only reverse observation. It is
// not an outcome judgment and cannot authorize or execute a route.
type DecisionRouteReverseObservation struct {
	Status               DecisionRouteReverseObservationStatus
	SourceVersion        string
	ContractVersion      string
	RouteStatus          string
	RouteRecordDigest    string
	DecisionDigest       string
	ExpectedRouteStatus  string
	ObservedRouteStatus  string
	OutcomeDigest        string
	TargetStage          string
	Reason               string
	ObservationDigest    string
	ReviewRequired       bool
	IsReadOnly           bool
	CanExecute           bool
	CanAuthorize         bool
}

// ObserveDecisionRouteReverseEvidence compares the expected and observed route
// while preserving explicit uncertainty and outcome provenance.
func ObserveDecisionRouteReverseEvidence(input DecisionRouteReverseObservationInput) DecisionRouteReverseObservation {
	observation := DecisionRouteReverseObservation{
		Status:              DecisionRouteReverseUnknown,
		SourceVersion:       input.SourceVersion,
		ContractVersion:     input.ContractVersion,
		RouteStatus:         input.RouteStatus,
		RouteRecordDigest:   input.RouteRecordDigest,
		DecisionDigest:      input.DecisionDigest,
		ExpectedRouteStatus: input.ExpectedRouteStatus,
		ObservedRouteStatus: input.ObservedRouteStatus,
		OutcomeDigest:       input.OutcomeDigest,
		TargetStage:         "decision_route_reverse_observation",
		ReviewRequired:      input.RouteStatus == "REVIEW_CANDIDATE",
		IsReadOnly:          true,
		CanExecute:          false,
		CanAuthorize:        false,
	}

	switch {
	case input.SourceVersion == "" || input.ContractVersion == "":
		observation.TargetStage = "identity"
		observation.Reason = "source or contract identity is missing"
	case input.RouteRecordDigest == "":
		observation.TargetStage = "route_record"
		observation.Reason = "route record digest is missing"
	case input.DecisionDigest == "":
		observation.TargetStage = "decision_digest"
		observation.Reason = "decision digest is missing"
	case input.ProducerDeferred:
		observation.Status = DecisionRouteReverseDeferred
		observation.Reason = "decision route reverse observation producer is deferred"
	case input.ExpectedRouteStatus == "":
		observation.TargetStage = "expected_route_status"
		observation.Reason = "expected route status is missing"
	case input.ObservedRouteStatus == "":
		observation.TargetStage = "observed_route_status"
		observation.Reason = "observed route status is missing"
	case input.OutcomeDigest == "":
		observation.TargetStage = "outcome"
		observation.Reason = "outcome evidence digest is missing"
	case input.ExpectedRouteStatus != input.ObservedRouteStatus:
		observation.TargetStage = "route_status_match"
		observation.Reason = "observed route status does not match expected route status"
	case !validDecisionRouteStatus(input.ExpectedRouteStatus):
		observation.TargetStage = "route_status"
		observation.Reason = "route status is not recognized"
	default:
		observation.Status = DecisionRouteReverseBound
		observation.Reason = "route and outcome evidence are aligned"
	}

	observation.ObservationDigest = decisionRouteReverseDigest(
		string(observation.Status),
		observation.SourceVersion,
		observation.ContractVersion,
		observation.RouteStatus,
		observation.RouteRecordDigest,
		observation.DecisionDigest,
		observation.ExpectedRouteStatus,
		observation.ObservedRouteStatus,
		observation.OutcomeDigest,
		observation.TargetStage,
		observation.Reason,
	)
	return observation
}

func validDecisionRouteStatus(status string) bool {
	switch status {
	case "ACCEPT_CANDIDATE", "REVIEW_CANDIDATE", "ABSTAIN_CANDIDATE", "DEFERRED":
		return true
	default:
		return false
	}
}

func decisionRouteReverseDigest(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}
