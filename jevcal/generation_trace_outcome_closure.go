package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	GenerationTraceOutcomeClosureBound    = "BOUND"
	GenerationTraceOutcomeClosureDeferred = "DEFERRED"
	GenerationTraceOutcomeClosureUnknown  = "UNKNOWN"

	GenerationTraceOutcomeClosureMetricName = "jev-generation-trace-route-outcome-closure"
	generationTraceOutcomeClosureBoundSignal = "jev-generation-trace-route-outcome-bound"
	generationTraceOutcomeClosureDeferredSignal = "jev-generation-trace-route-outcome-deferred"
	generationTraceOutcomeClosureUnknownSignal = "jev-generation-trace-route-outcome-unknown"
)

// GenerationTraceOutcomeClosureInput links generation evidence to the
// decision-route outcome metric without granting execution or authority.
type GenerationTraceOutcomeClosureInput struct {
	Trace        GenerationTraceObservation
	RouteOutcome DecisionRouteOutcomeMetric
}

// GenerationTraceOutcomeClosure is a provenance closure, not a quality,
// correctness, or improvement score.
type GenerationTraceOutcomeClosure struct {
	Status                    string
	MissingStage              string
	MetricName                string
	SourceVersion             string
	ContractVersion           string
	TraceObservationDigest    string
	TraceMetricDigest         string
	RouteOutcomeMetricDigest  string
	RouteRecordDigest         string
	DecisionDigest            string
	OutcomeDigest             string
	TargetStage               string
	Reason                    string
	ClosureSignal             string
	ClosureDigest             string
	ReadOnly                  bool
	ClaimsImprovement         bool
	CanExecute                bool
	CanAuthorize              bool
	NonAuthorizing            bool
}

// ObserveGenerationTraceOutcomeClosure binds the generation trace and route
// outcome only when both evidence chains are valid and metric-linked.
func ObserveGenerationTraceOutcomeClosure(input GenerationTraceOutcomeClosureInput) GenerationTraceOutcomeClosure {
	output := GenerationTraceOutcomeClosure{
		Status:            GenerationTraceOutcomeClosureUnknown,
		MissingStage:      "generation_trace_observation",
		MetricName:        GenerationTraceOutcomeClosureMetricName,
		TargetStage:       "generation_trace_outcome_closure",
		Reason:            "generation trace and route outcome are not yet bound",
		ClosureSignal:     generationTraceOutcomeClosureUnknownSignal,
		ReadOnly:          true,
		ClaimsImprovement: false,
		CanExecute:        false,
		CanAuthorize:      false,
		NonAuthorizing:    true,
	}

	if !input.Trace.ObservationalOnly ||
		input.Trace.ClaimsImprovement ||
		input.Trace.CanExecute ||
		input.Trace.CanAuthorize ||
		!input.RouteOutcome.IsReadOnly ||
		input.RouteOutcome.CanExecute ||
		input.RouteOutcome.CanAuthorize {
		output.MissingStage = "capability-boundary"
		output.TargetStage = "capability-boundary"
		output.Reason = "generation trace or route outcome crossed a capability boundary"
		output.NonAuthorizing = false
		output.ClosureDigest = generationTraceOutcomeClosureDigest(output)
		return output
	}
	if err := input.Trace.Validate(); err != nil {
		output.MissingStage = "generation_trace_observation"
		output.TargetStage = "generation_trace_observation"
		output.Reason = "generation trace observation is invalid"
		output.ClosureDigest = generationTraceOutcomeClosureDigest(output)
		return output
	}
	if err := input.RouteOutcome.Validate(); err != nil {
		output.MissingStage = "decision_route_outcome_metric"
		output.TargetStage = "decision_route_outcome_metric"
		output.Reason = "decision route outcome metric is invalid"
		output.ClosureDigest = generationTraceOutcomeClosureDigest(output)
		return output
	}

	output.SourceVersion = input.Trace.SourceVersion
	output.ContractVersion = input.Trace.ContractVersion
	output.TraceObservationDigest = input.Trace.ObservationDigest
	output.TraceMetricDigest = input.Trace.MetricDigest
	output.RouteOutcomeMetricDigest = input.RouteOutcome.MetricDigest
	output.RouteRecordDigest = input.RouteOutcome.RouteRecordDigest
	output.DecisionDigest = input.RouteOutcome.DecisionDigest
	output.OutcomeDigest = input.RouteOutcome.OutcomeDigest

	if input.Trace.Status == GenerationTraceUnknown ||
		input.RouteOutcome.Status == DecisionRouteOutcomeMetricUnknown {
		if input.Trace.Status == GenerationTraceUnknown {
			output.MissingStage = input.Trace.MissingStage
			output.TargetStage = input.Trace.TargetStage
			output.Reason = input.Trace.Reason
		} else {
			output.MissingStage = "decision_route_outcome_metric"
			output.TargetStage = input.RouteOutcome.TargetStage
			output.Reason = input.RouteOutcome.Reason
		}
		output.ClosureDigest = generationTraceOutcomeClosureDigest(output)
		return output
	}
	if input.Trace.Status == GenerationTraceDeferred ||
		input.RouteOutcome.Status == DecisionRouteOutcomeMetricDeferred {
		output.Status = GenerationTraceOutcomeClosureDeferred
		output.MissingStage = ""
		output.TargetStage = "generation_trace_outcome_closure"
		output.Reason = "generation trace or route outcome producer is deferred"
		output.ClosureSignal = generationTraceOutcomeClosureDeferredSignal
		output.ClosureDigest = generationTraceOutcomeClosureDigest(output)
		return output
	}
	if input.Trace.MetricDigest == "" ||
		input.Trace.MetricDigest != input.RouteOutcome.MetricDigest {
		output.MissingStage = "metric_link"
		output.TargetStage = "metric_link"
		output.Reason = "generation trace metric is not linked to route outcome metric"
		output.ClosureDigest = generationTraceOutcomeClosureDigest(output)
		return output
	}

	output.Status = GenerationTraceOutcomeClosureBound
	output.MissingStage = ""
	output.TargetStage = "generation_trace_outcome_closure"
	output.Reason = "generation trace and route outcome metric are linked"
	output.ClosureSignal = generationTraceOutcomeClosureBoundSignal
	output.ClosureDigest = generationTraceOutcomeClosureDigest(output)
	return output
}

func (closure GenerationTraceOutcomeClosure) Validate() error {
	switch closure.Status {
	case GenerationTraceOutcomeClosureBound, GenerationTraceOutcomeClosureDeferred, GenerationTraceOutcomeClosureUnknown:
	default:
		return fmt.Errorf("invalid generation trace outcome closure status %q", closure.Status)
	}
	if closure.MetricName != GenerationTraceOutcomeClosureMetricName ||
		closure.TargetStage == "" ||
		closure.Reason == "" ||
		closure.ClosureDigest == "" {
		return fmt.Errorf("generation trace outcome closure identity is incomplete")
	}
	if !closure.ReadOnly || closure.ClaimsImprovement || closure.CanExecute || closure.CanAuthorize {
		return fmt.Errorf("generation trace outcome closure crossed a forbidden boundary")
	}
	if closure.Status == GenerationTraceOutcomeClosureUnknown {
		if closure.MissingStage == "" || closure.ClosureSignal != generationTraceOutcomeClosureUnknownSignal {
			return fmt.Errorf("unknown generation trace outcome closure must preserve its missing stage")
		}
	} else if closure.Status == GenerationTraceOutcomeClosureDeferred {
		if closure.MissingStage != "" || closure.ClosureSignal != generationTraceOutcomeClosureDeferredSignal {
			return fmt.Errorf("deferred generation trace outcome closure is invalid")
		}
	} else {
		if closure.MissingStage != "" ||
			closure.ClosureSignal != generationTraceOutcomeClosureBoundSignal ||
			closure.SourceVersion == "" ||
			closure.ContractVersion == "" ||
			closure.TraceObservationDigest == "" ||
			closure.TraceMetricDigest == "" ||
			closure.TraceMetricDigest != closure.RouteOutcomeMetricDigest ||
			closure.RouteRecordDigest == "" ||
			closure.DecisionDigest == "" ||
			closure.OutcomeDigest == "" {
			return fmt.Errorf("bound generation trace outcome closure is incomplete")
		}
	}
	if closure.ClosureDigest != generationTraceOutcomeClosureDigest(closure) {
		return fmt.Errorf("generation trace outcome closure digest mismatch")
	}
	return nil
}

func generationTraceOutcomeClosureDigest(closure GenerationTraceOutcomeClosure) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		closure.Status,
		closure.MissingStage,
		closure.MetricName,
		closure.SourceVersion,
		closure.ContractVersion,
		closure.TraceObservationDigest,
		closure.TraceMetricDigest,
		closure.RouteOutcomeMetricDigest,
		closure.RouteRecordDigest,
		closure.DecisionDigest,
		closure.OutcomeDigest,
		closure.TargetStage,
		closure.Reason,
		closure.ClosureSignal,
		fmt.Sprint(closure.ReadOnly),
		fmt.Sprint(closure.ClaimsImprovement),
		fmt.Sprint(closure.CanExecute),
		fmt.Sprint(closure.CanAuthorize),
		fmt.Sprint(closure.NonAuthorizing),
	}, "|")))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}