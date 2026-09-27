package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	DecisionRouteOutcomeMetricBound    = "BOUND"
	DecisionRouteOutcomeMetricDeferred = "DEFERRED"
	DecisionRouteOutcomeMetricUnknown  = "UNKNOWN"

	decisionRouteOutcomeMetricBoundSignal    = "jev-decision-route-outcome-bound"
	decisionRouteOutcomeMetricDeferredSignal = "jev-decision-route-outcome-deferred"
	decisionRouteOutcomeMetricUnknownSignal  = "jev-decision-route-outcome-unknown"
)

// DecisionRouteOutcomeMetric normalizes one reverse observation for downstream
// reporting without judging outcome quality or granting authority.
type DecisionRouteOutcomeMetric struct {
	Status               string
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
	MetricSignal         string
	MetricDigest         string
	IsReadOnly           bool
	CanExecute           bool
	CanAuthorize         bool
}

// ObserveDecisionRouteOutcomeMetric preserves the reverse observation state.
// It is a metric of evidence availability, not a correctness or improvement
// score.
func ObserveDecisionRouteOutcomeMetric(
	observation DecisionRouteReverseObservation,
) DecisionRouteOutcomeMetric {
	metric := DecisionRouteOutcomeMetric{
		Status:              DecisionRouteOutcomeMetricUnknown,
		SourceVersion:       observation.SourceVersion,
		ContractVersion:     observation.ContractVersion,
		RouteStatus:         observation.RouteStatus,
		RouteRecordDigest:   observation.RouteRecordDigest,
		DecisionDigest:      observation.DecisionDigest,
		ExpectedRouteStatus: observation.ExpectedRouteStatus,
		ObservedRouteStatus: observation.ObservedRouteStatus,
		OutcomeDigest:       observation.OutcomeDigest,
		TargetStage:         observation.TargetStage,
		Reason:              observation.Reason,
		ObservationDigest:   observation.ObservationDigest,
		MetricSignal:        decisionRouteOutcomeMetricUnknownSignal,
		IsReadOnly:          true,
		CanExecute:          false,
		CanAuthorize:        false,
	}
	if metric.TargetStage == "" {
		metric.TargetStage = "decision_route_outcome_metric"
	}
	if metric.Reason == "" {
		metric.Reason = "reverse observation remains UNKNOWN"
	}
	if !observation.IsReadOnly || observation.CanExecute || observation.CanAuthorize {
		metric.TargetStage = "capability_boundary"
		metric.Reason = "reverse observation crossed a non-authorizing boundary"
		return finalizeDecisionRouteOutcomeMetric(metric)
	}
	if metric.ObservationDigest == "" {
		metric.TargetStage = "observation_digest"
		metric.Reason = "reverse observation digest is missing"
		return finalizeDecisionRouteOutcomeMetric(metric)
	}

	switch observation.Status {
	case DecisionRouteReverseBound:
		if observation.SourceVersion == "" ||
			observation.ContractVersion == "" ||
			observation.RouteRecordDigest == "" ||
			observation.DecisionDigest == "" ||
			observation.OutcomeDigest == "" ||
			observation.ExpectedRouteStatus == "" ||
			observation.ExpectedRouteStatus != observation.ObservedRouteStatus {
			metric.TargetStage = "bound_evidence"
			metric.Reason = "bound reverse observation is missing aligned route or outcome evidence"
			return finalizeDecisionRouteOutcomeMetric(metric)
		}
		metric.Status = DecisionRouteOutcomeMetricBound
		metric.MetricSignal = decisionRouteOutcomeMetricBoundSignal
	case DecisionRouteReverseDeferred:
		metric.Status = DecisionRouteOutcomeMetricDeferred
		metric.MetricSignal = decisionRouteOutcomeMetricDeferredSignal
	case DecisionRouteReverseUnknown:
		metric.Status = DecisionRouteOutcomeMetricUnknown
		metric.MetricSignal = decisionRouteOutcomeMetricUnknownSignal
	default:
		metric.TargetStage = "reverse_observation_status"
		metric.Reason = "reverse observation status is not recognized"
	}
	return finalizeDecisionRouteOutcomeMetric(metric)
}

func (metric DecisionRouteOutcomeMetric) Validate() error {
	if metric.Status != DecisionRouteOutcomeMetricBound &&
		metric.Status != DecisionRouteOutcomeMetricDeferred &&
		metric.Status != DecisionRouteOutcomeMetricUnknown {
		return fmt.Errorf("invalid decision route outcome metric status %q", metric.Status)
	}
	if metric.TargetStage == "" || metric.Reason == "" || metric.ObservationDigest == "" {
		return fmt.Errorf("decision route outcome metric evidence identity is incomplete")
	}
	if !metric.IsReadOnly || metric.CanExecute || metric.CanAuthorize {
		return fmt.Errorf("decision route outcome metric crossed an execution or authorization boundary")
	}
	switch metric.Status {
	case DecisionRouteOutcomeMetricBound:
		if metric.MetricSignal != decisionRouteOutcomeMetricBoundSignal ||
			metric.SourceVersion == "" ||
			metric.ContractVersion == "" ||
			metric.RouteRecordDigest == "" ||
			metric.DecisionDigest == "" ||
			metric.OutcomeDigest == "" ||
			metric.ExpectedRouteStatus == "" ||
			metric.ExpectedRouteStatus != metric.ObservedRouteStatus {
			return fmt.Errorf("bound decision route outcome metric is incomplete")
		}
	case DecisionRouteOutcomeMetricDeferred:
		if metric.MetricSignal != decisionRouteOutcomeMetricDeferredSignal {
			return fmt.Errorf("deferred decision route outcome metric signal is invalid")
		}
	case DecisionRouteOutcomeMetricUnknown:
		if metric.MetricSignal != decisionRouteOutcomeMetricUnknownSignal {
			return fmt.Errorf("unknown decision route outcome metric signal is invalid")
		}
	}
	if metric.MetricDigest == "" {
		return fmt.Errorf("decision route outcome metric digest is required")
	}
	if metric.MetricDigest != decisionRouteOutcomeMetricDigest(metric) {
		return fmt.Errorf("decision route outcome metric digest mismatch")
	}
	return nil
}

func finalizeDecisionRouteOutcomeMetric(
	metric DecisionRouteOutcomeMetric,
) DecisionRouteOutcomeMetric {
	metric.MetricDigest = decisionRouteOutcomeMetricDigest(metric)
	return metric
}

func decisionRouteOutcomeMetricDigest(metric DecisionRouteOutcomeMetric) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		metric.Status,
		metric.SourceVersion,
		metric.ContractVersion,
		metric.RouteStatus,
		metric.RouteRecordDigest,
		metric.DecisionDigest,
		metric.ExpectedRouteStatus,
		metric.ObservedRouteStatus,
		metric.OutcomeDigest,
		metric.TargetStage,
		metric.Reason,
		metric.ObservationDigest,
		metric.MetricSignal,
		strings.ToLower(fmt.Sprint(metric.IsReadOnly)),
		strings.ToLower(fmt.Sprint(metric.CanExecute)),
		strings.ToLower(fmt.Sprint(metric.CanAuthorize)),
	}, "|")))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}
