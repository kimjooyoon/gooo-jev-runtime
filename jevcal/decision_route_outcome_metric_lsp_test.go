package jevcal

import "testing"

func routeOutcomeMetricForLSP(
	t *testing.T,
	input DecisionRouteReverseObservationInput,
) DecisionRouteOutcomeMetric {
	t.Helper()
	return ObserveDecisionRouteOutcomeMetric(
		ObserveDecisionRouteReverseEvidence(input),
	)
}

func TestProjectDecisionRouteOutcomeMetricLSPBound(t *testing.T) {
	metric := routeOutcomeMetricForLSP(t, DecisionRouteReverseObservationInput{
		SourceVersion:       "source-v1",
		ContractVersion:     "contract-v1",
		RouteStatus:         "REVIEW_CANDIDATE",
		RouteRecordDigest:   "sha256:route-record",
		DecisionDigest:       "sha256:decision",
		ExpectedRouteStatus: "REVIEW_CANDIDATE",
		ObservedRouteStatus: "REVIEW_CANDIDATE",
		OutcomeDigest:       "sha256:outcome",
	})
	projection := ProjectDecisionRouteOutcomeMetricLSP(metric)
	if projection.Status != DecisionRouteOutcomeMetricBound ||
		projection.Code != "jev.route_outcome.bound" ||
		projection.Severity != "Information" {
		t.Fatalf("unexpected bound projection: %#v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestProjectDecisionRouteOutcomeMetricLSPPreservesDeferred(t *testing.T) {
	metric := routeOutcomeMetricForLSP(t, DecisionRouteReverseObservationInput{
		SourceVersion:     "source-v1",
		ContractVersion:   "contract-v1",
		RouteStatus:       "DEFERRED",
		RouteRecordDigest: "sha256:route-record",
		DecisionDigest:    "sha256:decision",
		ProducerDeferred:  true,
	})
	projection := ProjectDecisionRouteOutcomeMetricLSP(metric)
	if projection.Status != DecisionRouteOutcomeMetricDeferred ||
		projection.Code != "jev.route_outcome.deferred" {
		t.Fatalf("unexpected deferred projection: %#v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestProjectDecisionRouteOutcomeMetricLSPPreservesUnknown(t *testing.T) {
	metric := routeOutcomeMetricForLSP(t, DecisionRouteReverseObservationInput{
		SourceVersion:  "source-v1",
		ContractVersion: "contract-v1",
	})
	projection := ProjectDecisionRouteOutcomeMetricLSP(metric)
	if projection.Status != DecisionRouteOutcomeMetricUnknown ||
		projection.Code != "jev.route_outcome.unknown" ||
		projection.TargetStage != "route_record" {
		t.Fatalf("unexpected unknown projection: %#v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestProjectDecisionRouteOutcomeMetricLSPRejectsTampering(t *testing.T) {
	metric := DecisionRouteOutcomeMetric{
		Status:            DecisionRouteOutcomeMetricBound,
		TargetStage:       "bound_evidence",
		Reason:            "tampered",
		ObservationDigest: "sha256:observation",
		MetricSignal:      decisionRouteOutcomeMetricBoundSignal,
		MetricDigest:      "sha256:tampered",
		IsReadOnly:        true,
	}
	projection := ProjectDecisionRouteOutcomeMetricLSP(metric)
	if projection.Status != DecisionRouteOutcomeMetricUnknown ||
		projection.TargetStage != "metric_evidence" {
		t.Fatalf("expected unknown tampered projection: %#v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatal(err)
	}
}
