package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

// DecisionRouteReverseObservationSourceInput binds exact source text to the
// typed reverse-observation inputs without authorizing or executing a route.
type DecisionRouteReverseObservationSourceInput struct {
	SourceVersion       string
	ContractVersion     string
	SourceText          string
	RouteStatus         string
	RouteRecordDigest   string
	DecisionDigest      string
	ExpectedRouteStatus string
	ObservedRouteStatus string
	OutcomeDigest       string
	ProducerDeferred    bool
	NonAuthorizing      bool
}

// DecisionRouteReverseObservationSourceBinding is a read-only source binding.
// It records the first unresolved stage and never judges the route outcome.
type DecisionRouteReverseObservationSourceBinding struct {
	Status           DecisionRouteReverseObservationStatus
	SourceVersion    string
	ContractVersion  string
	SourceDigest     string
	ObservationDigest string
	BindingDigest    string
	FirstMismatch    string
	TargetStage      string
	NonExecuting     bool
	NonAuthorizing   bool
	IsReadOnly       bool
}

// BindDecisionRouteReverseObservationFromSource binds exact source text and
// typed reverse-observation evidence while preserving UNKNOWN and DEFERRED.
func BindDecisionRouteReverseObservationFromSource(input DecisionRouteReverseObservationSourceInput) DecisionRouteReverseObservationSourceBinding {
	binding := DecisionRouteReverseObservationSourceBinding{
		Status:          DecisionRouteReverseUnknown,
		SourceVersion:   input.SourceVersion,
		ContractVersion: input.ContractVersion,
		SourceDigest:    decisionRouteReverseSourceDigest(input),
		TargetStage:     "decision_route_reverse_observation",
		NonExecuting:    true,
		NonAuthorizing:  true,
		IsReadOnly:      true,
	}

	switch {
	case strings.TrimSpace(input.SourceText) == "":
		binding.FirstMismatch = "source-text"
		binding.TargetStage = "source-text"
	case !input.NonAuthorizing:
		binding.FirstMismatch = "authorization-boundary"
		binding.TargetStage = "authorization-boundary"
	default:
		observation := ObserveDecisionRouteReverseEvidence(DecisionRouteReverseObservationInput{
			SourceVersion:       input.SourceVersion,
			ContractVersion:     input.ContractVersion,
			RouteStatus:         input.RouteStatus,
			RouteRecordDigest:   input.RouteRecordDigest,
			DecisionDigest:      input.DecisionDigest,
			ExpectedRouteStatus: input.ExpectedRouteStatus,
			ObservedRouteStatus: input.ObservedRouteStatus,
			OutcomeDigest:       input.OutcomeDigest,
			ProducerDeferred:    input.ProducerDeferred,
		})
		binding.Status = observation.Status
		binding.ObservationDigest = observation.ObservationDigest
		binding.TargetStage = observation.TargetStage
		switch observation.Status {
		case DecisionRouteReverseDeferred:
			binding.FirstMismatch = "producer-deferred"
		case DecisionRouteReverseUnknown:
			binding.FirstMismatch = observation.TargetStage
		}
	}

	binding.BindingDigest = decisionRouteReverseBindingDigest(binding)
	return binding
}

// Validate checks structural integrity of the source-binding projection. It
// does not infer correctness from the observed outcome.
func (binding DecisionRouteReverseObservationSourceBinding) Validate() bool {
	if !binding.NonExecuting || !binding.NonAuthorizing || !binding.IsReadOnly {
		return false
	}
	if binding.SourceDigest == "" || binding.BindingDigest == "" {
		return false
	}
	switch binding.Status {
	case DecisionRouteReverseBound:
		if binding.FirstMismatch != "" || binding.ObservationDigest == "" {
			return false
		}
	case DecisionRouteReverseDeferred, DecisionRouteReverseUnknown:
		if binding.FirstMismatch == "" || binding.ObservationDigest == "" {
			return false
		}
	default:
		return false
	}
	return binding.BindingDigest == decisionRouteReverseBindingDigest(binding)
}

func decisionRouteReverseSourceDigest(input DecisionRouteReverseObservationSourceInput) string {
	return decisionRouteReverseLengthDigest(
		input.SourceVersion,
		input.ContractVersion,
		input.SourceText,
		input.RouteStatus,
		input.RouteRecordDigest,
		input.DecisionDigest,
		input.ExpectedRouteStatus,
		input.ObservedRouteStatus,
		input.OutcomeDigest,
		strconv.FormatBool(input.ProducerDeferred),
		strconv.FormatBool(input.NonAuthorizing),
	)
}

func decisionRouteReverseBindingDigest(binding DecisionRouteReverseObservationSourceBinding) string {
	return decisionRouteReverseLengthDigest(
		string(binding.Status),
		binding.SourceVersion,
		binding.ContractVersion,
		binding.SourceDigest,
		binding.ObservationDigest,
		binding.FirstMismatch,
		binding.TargetStage,
		strconv.FormatBool(binding.NonExecuting),
		strconv.FormatBool(binding.NonAuthorizing),
		strconv.FormatBool(binding.IsReadOnly),
	)
}

func decisionRouteReverseLengthDigest(parts ...string) string {
	var encoded strings.Builder
	for _, part := range parts {
		encoded.WriteString(strconv.Itoa(len(part)))
		encoded.WriteByte(':')
		encoded.WriteString(part)
		encoded.WriteByte(';')
	}
	sum := sha256.Sum256([]byte(encoded.String()))
	return "sha256:" + hex.EncodeToString(sum[:])
}

