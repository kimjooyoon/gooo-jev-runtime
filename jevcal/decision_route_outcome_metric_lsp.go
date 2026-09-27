package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// DecisionRouteOutcomeMetricLSPProjection is an inspection-only view of a
// route outcome metric. It never executes, authorizes, or judges quality.
type DecisionRouteOutcomeMetricLSPProjection struct {
	Status            string
	Code              string
	Severity          string
	Title             string
	TargetStage       string
	Reason            string
	MetricSignal      string
	MetricDigest      string
	ObservationDigest string
	EvidenceDigest    string
	IsReadOnly        bool
	CanExecute        bool
	CanAuthorize      bool
}

// ProjectDecisionRouteOutcomeMetricLSP preserves BOUND, DEFERRED, and UNKNOWN
// states while exposing only evidence-linked editor information.
func ProjectDecisionRouteOutcomeMetricLSP(
	metric DecisionRouteOutcomeMetric,
) DecisionRouteOutcomeMetricLSPProjection {
	projection := DecisionRouteOutcomeMetricLSPProjection{
		Status:       DecisionRouteOutcomeMetricUnknown,
		Code:         "jev.route_outcome.unknown",
		Severity:     "Error",
		Title:        "Decision route outcome evidence is incomplete",
		TargetStage:  metric.TargetStage,
		Reason:       metric.Reason,
		MetricSignal: decisionRouteOutcomeMetricUnknownSignal,
		IsReadOnly:   true,
		CanExecute:   false,
		CanAuthorize: false,
	}
	if projection.TargetStage == "" {
		projection.TargetStage = "outcome_metric"
	}
	if projection.Reason == "" {
		projection.Reason = "decision route outcome metric is UNKNOWN"
	}
	if err := metric.Validate(); err != nil {
		projection.TargetStage = "metric_evidence"
		projection.Reason = "decision route outcome metric is invalid"
		return finalizeDecisionRouteOutcomeMetricLSP(projection)
	}

	projection.Status = metric.Status
	projection.TargetStage = metric.TargetStage
	projection.Reason = metric.Reason
	projection.MetricSignal = metric.MetricSignal
	projection.MetricDigest = metric.MetricDigest
	projection.ObservationDigest = metric.ObservationDigest
	switch metric.Status {
	case DecisionRouteOutcomeMetricBound:
		projection.Code = "jev.route_outcome.bound"
		projection.Severity = "Information"
		projection.Title = "Route and outcome evidence are aligned; no quality judgment was made"
	case DecisionRouteOutcomeMetricDeferred:
		projection.Code = "jev.route_outcome.deferred"
		projection.Severity = "Hint"
		projection.Title = "Route outcome producer is deferred"
	case DecisionRouteOutcomeMetricUnknown:
		projection.Code = "jev.route_outcome.unknown"
		projection.Severity = "Warning"
		projection.Title = "Route outcome evidence remains UNKNOWN"
	}
	return finalizeDecisionRouteOutcomeMetricLSP(projection)
}

func (projection DecisionRouteOutcomeMetricLSPProjection) Validate() error {
	if projection.Status != DecisionRouteOutcomeMetricBound &&
		projection.Status != DecisionRouteOutcomeMetricDeferred &&
		projection.Status != DecisionRouteOutcomeMetricUnknown {
		return fmt.Errorf("invalid route outcome LSP status %q", projection.Status)
	}
	if projection.TargetStage == "" || projection.Reason == "" {
		return fmt.Errorf("route outcome LSP target and reason are required")
	}
	if projection.MetricSignal != decisionRouteOutcomeMetricBoundSignal &&
		projection.MetricSignal != decisionRouteOutcomeMetricDeferredSignal &&
		projection.MetricSignal != decisionRouteOutcomeMetricUnknownSignal {
		return fmt.Errorf("invalid route outcome LSP metric signal %q", projection.MetricSignal)
	}
	if !projection.IsReadOnly || projection.CanExecute || projection.CanAuthorize {
		return fmt.Errorf("route outcome LSP projection crossed an execution or authorization boundary")
	}
	if projection.Status != DecisionRouteOutcomeMetricUnknown &&
		(projection.MetricDigest == "" || projection.ObservationDigest == "") {
		return fmt.Errorf("bound or deferred route outcome LSP projection lacks evidence")
	}
	if projection.EvidenceDigest == "" {
		return fmt.Errorf("route outcome LSP evidence digest is required")
	}
	if projection.EvidenceDigest != decisionRouteOutcomeMetricLSPDigest(projection) {
		return fmt.Errorf("route outcome LSP evidence digest mismatch")
	}
	return nil
}

func finalizeDecisionRouteOutcomeMetricLSP(
	projection DecisionRouteOutcomeMetricLSPProjection,
) DecisionRouteOutcomeMetricLSPProjection {
	projection.EvidenceDigest = decisionRouteOutcomeMetricLSPDigest(projection)
	return projection
}

func decisionRouteOutcomeMetricLSPDigest(
	projection DecisionRouteOutcomeMetricLSPProjection,
) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		projection.Status,
		projection.Code,
		projection.Severity,
		projection.Title,
		projection.TargetStage,
		projection.Reason,
		projection.MetricSignal,
		projection.MetricDigest,
		projection.ObservationDigest,
		fmt.Sprint(projection.IsReadOnly),
		fmt.Sprint(projection.CanExecute),
		fmt.Sprint(projection.CanAuthorize),
	}, "|")))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}
