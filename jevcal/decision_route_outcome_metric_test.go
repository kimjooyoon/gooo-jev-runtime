package jevcal

import "testing"

func TestObserveDecisionRouteOutcomeMetricBound(t *testing.T) {
	observation := ObserveDecisionRouteReverseEvidence(DecisionRouteReverseObservationInput{
		SourceVersion:       "source-v1",
		ContractVersion:     "contract-v1",
		RouteStatus:         "REVIEW_CANDIDATE",
		RouteRecordDigest:   "sha256:route-record",
		DecisionDigest:       "sha256:decision",
		ExpectedRouteStatus: "REVIEW_CANDIDATE",
		ObservedRouteStatus: "REVIEW_CANDIDATE",
		OutcomeDigest:       "sha256:outcome",
	})
	metric := ObserveDecisionRouteOutcomeMetric(observation)
	if metric.Status != DecisionRouteOutcomeMetricBound ||
		metric.MetricSignal != decisionRouteOutcomeMetricBoundSignal {
		t.Fatalf("expected bound metric, got %#v", metric)
	}
	if err := metric.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestObserveDecisionRouteOutcomeMetricPreservesDeferred(t *testing.T) {
	observation := ObserveDecisionRouteReverseEvidence(DecisionRouteReverseObservationInput{
		SourceVersion:       "source-v1",
		ContractVersion:     "contract-v1",
		RouteStatus:         "DEFERRED",
		RouteRecordDigest:   "sha256:route-record",
		DecisionDigest:      "sha256:decision",
		ProducerDeferred:    true,
	})
	metric := ObserveDecisionRouteOutcomeMetric(observation)
	if metric.Status != DecisionRouteOutcomeMetricDeferred ||
		metric.MetricSignal != decisionRouteOutcomeMetricDeferredSignal {
		t.Fatalf("expected deferred metric, got %#v", metric)
	}
	if err := metric.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestObserveDecisionRouteOutcomeMetricPreservesUnknown(t *testing.T) {
	observation := ObserveDecisionRouteReverseEvidence(DecisionRouteReverseObservationInput{
		SourceVersion:  "source-v1",
		ContractVersion: "contract-v1",
	})
	metric := ObserveDecisionRouteOutcomeMetric(observation)
	if metric.Status != DecisionRouteOutcomeMetricUnknown ||
		metric.TargetStage != "route_record" ||
		metric.MetricSignal != decisionRouteOutcomeMetricUnknownSignal {
		t.Fatalf("expected unknown metric, got %#v", metric)
	}
	if err := metric.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestObserveDecisionRouteOutcomeMetricRejectsUnsafeObservation(t *testing.T) {
	metric := ObserveDecisionRouteOutcomeMetric(DecisionRouteReverseObservation{
		Status:            DecisionRouteReverseBound,
		ObservationDigest: "sha256:observation",
		IsReadOnly:        false,
		CanExecute:        true,
	})
	if metric.Status != DecisionRouteOutcomeMetricUnknown ||
		metric.TargetStage != "capability_boundary" ||
		metric.CanExecute ||
		metric.CanAuthorize ||
		!metric.IsReadOnly {
		t.Fatalf("expected safe unknown metric, got %#v", metric)
	}
	if err := metric.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestDecisionRouteOutcomeMetricRejectsTampering(t *testing.T) {
	metric := DecisionRouteOutcomeMetric{
		Status:            DecisionRouteOutcomeMetricUnknown,
		TargetStage:       "input",
		Reason:            "unknown",
		ObservationDigest: "sha256:observation",
		MetricSignal:      decisionRouteOutcomeMetricUnknownSignal,
		MetricDigest:      "sha256:tampered",
		IsReadOnly:        true,
	}
	if err := metric.Validate(); err == nil {
		t.Fatal("expected tampered metric to be rejected")
	}
}
