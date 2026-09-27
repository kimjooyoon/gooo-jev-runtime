package jevcal

import "testing"

func generationTraceOutcomeClosureMetric(t *testing.T) DecisionRouteOutcomeMetric {
	t.Helper()
	observation := DecisionRouteReverseObservation{
		Status:               DecisionRouteReverseBound,
		SourceVersion:        "source-v1",
		ContractVersion:      "contract-v1",
		RouteStatus:          "approved",
		RouteRecordDigest:    "sha256:route",
		DecisionDigest:        "sha256:decision",
		ExpectedRouteStatus:  "approved",
		ObservedRouteStatus:  "approved",
		OutcomeDigest:        "sha256:outcome",
		TargetStage:          "decision_route_reverse_observation",
		Reason:               "route outcome is aligned",
		ObservationDigest:    "sha256:route-observation",
		IsReadOnly:           true,
		CanExecute:           false,
		CanAuthorize:         false,
	}
	metric := ObserveDecisionRouteOutcomeMetric(observation)
	if err := metric.Validate(); err != nil {
		t.Fatalf("route metric Validate() error = %v", err)
	}
	return metric
}

func generationTraceOutcomeClosureInput(t *testing.T) GenerationTraceOutcomeClosureInput {
	t.Helper()
	metric := generationTraceOutcomeClosureMetric(t)
	traceInput := generationTraceTestInput()
	traceInput.MetricDigest = metric.MetricDigest
	return GenerationTraceOutcomeClosureInput{
		Trace:        ObserveGenerationTrace(traceInput),
		RouteOutcome: metric,
	}
}

func TestObserveGenerationTraceOutcomeClosureBound(t *testing.T) {
	closure := ObserveGenerationTraceOutcomeClosure(generationTraceOutcomeClosureInput(t))
	if closure.Status != GenerationTraceOutcomeClosureBound ||
		closure.MissingStage != "" ||
		closure.TraceMetricDigest == "" ||
		closure.TraceMetricDigest != closure.RouteOutcomeMetricDigest {
		t.Fatalf("unexpected bound closure: %+v", closure)
	}
	if err := closure.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestObserveGenerationTraceOutcomeClosurePreservesMetricLinkGap(t *testing.T) {
	input := generationTraceOutcomeClosureInput(t)
	traceInput := generationTraceTestInput()
	traceInput.MetricDigest = "sha256:unlinked"
	input.Trace = ObserveGenerationTrace(traceInput)
	closure := ObserveGenerationTraceOutcomeClosure(input)

	if closure.Status != GenerationTraceOutcomeClosureUnknown ||
		closure.MissingStage != "metric_link" ||
		closure.TargetStage != "metric_link" {
		t.Fatalf("unexpected metric link closure: %+v", closure)
	}
	if err := closure.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestObserveGenerationTraceOutcomeClosurePreservesDeferredState(t *testing.T) {
	input := generationTraceOutcomeClosureInput(t)
	traceInput := generationTraceTestInput()
	traceInput.MetricDigest = input.RouteOutcome.MetricDigest
	traceInput.ProducerDeferred = true
	input.Trace = ObserveGenerationTrace(traceInput)
	closure := ObserveGenerationTraceOutcomeClosure(input)

	if closure.Status != GenerationTraceOutcomeClosureDeferred ||
		closure.MissingStage != "" ||
		closure.ClosureSignal != generationTraceOutcomeClosureDeferredSignal {
		t.Fatalf("unexpected deferred closure: %+v", closure)
	}
	if err := closure.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestObserveGenerationTraceOutcomeClosureRejectsTamperedEvidence(t *testing.T) {
	input := generationTraceOutcomeClosureInput(t)
	input.RouteOutcome.MetricDigest = "sha256:tampered"
	closure := ObserveGenerationTraceOutcomeClosure(input)

	if closure.Status != GenerationTraceOutcomeClosureUnknown ||
		closure.MissingStage != "decision_route_outcome_metric" {
		t.Fatalf("tampered route metric was not rejected: %+v", closure)
	}
	if err := closure.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestObserveGenerationTraceOutcomeClosureFailsClosedOnCapabilityBoundary(t *testing.T) {
	input := generationTraceOutcomeClosureInput(t)
	input.Trace.CanExecute = true
	closure := ObserveGenerationTraceOutcomeClosure(input)

	if closure.Status != GenerationTraceOutcomeClosureUnknown ||
		closure.MissingStage != "capability-boundary" ||
		closure.NonAuthorizing {
		t.Fatalf("capability boundary was not preserved: %+v", closure)
	}
}