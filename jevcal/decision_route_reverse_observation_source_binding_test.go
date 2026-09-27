package jevcal

import "testing"

func TestBindDecisionRouteReverseObservationFromSourceBound(t *testing.T) {
	input := DecisionRouteReverseObservationSourceInput{
		SourceVersion:       "source-v1",
		ContractVersion:     "contract-v1",
		SourceText:          "module route\nstatus ACCEPT_CANDIDATE\n",
		RouteStatus:         "ACCEPT_CANDIDATE",
		RouteRecordDigest:   "sha256:route",
		DecisionDigest:      "sha256:decision",
		ExpectedRouteStatus: "ACCEPT_CANDIDATE",
		ObservedRouteStatus: "ACCEPT_CANDIDATE",
		OutcomeDigest:       "sha256:outcome",
		NonAuthorizing:      true,
	}

	binding := BindDecisionRouteReverseObservationFromSource(input)
	if binding.Status != DecisionRouteReverseBound {
		t.Fatalf("status = %s, want %s", binding.Status, DecisionRouteReverseBound)
	}
	if !binding.Validate() {
		t.Fatal("expected bound source binding to validate")
	}
	if binding.SourceDigest == "" || binding.ObservationDigest == "" || binding.BindingDigest == "" {
		t.Fatal("expected all provenance digests")
	}
}

func TestBindDecisionRouteReverseObservationFromSourcePreservesUnknownAndDeferred(t *testing.T) {
	base := DecisionRouteReverseObservationSourceInput{
		SourceVersion:       "source-v1",
		ContractVersion:     "contract-v1",
		SourceText:          "module route\n",
		RouteStatus:         "ACCEPT_CANDIDATE",
		RouteRecordDigest:   "sha256:route",
		DecisionDigest:      "sha256:decision",
		ExpectedRouteStatus: "ACCEPT_CANDIDATE",
		ObservedRouteStatus: "REVIEW_CANDIDATE",
		OutcomeDigest:       "sha256:outcome",
		NonAuthorizing:      true,
	}
	unknown := BindDecisionRouteReverseObservationFromSource(base)
	if unknown.Status != DecisionRouteReverseUnknown || unknown.FirstMismatch != "route_status_match" {
		t.Fatalf("unknown = %#v", unknown)
	}
	if !unknown.Validate() {
		t.Fatal("expected unknown source binding to validate")
	}

	base.ProducerDeferred = true
	deferred := BindDecisionRouteReverseObservationFromSource(base)
	if deferred.Status != DecisionRouteReverseDeferred || deferred.FirstMismatch != "producer-deferred" {
		t.Fatalf("deferred = %#v", deferred)
	}
	if !deferred.Validate() {
		t.Fatal("expected deferred source binding to validate")
	}
}

func TestBindDecisionRouteReverseObservationFromSourceRejectsMissingBoundaryAndSource(t *testing.T) {
	input := DecisionRouteReverseObservationSourceInput{
		SourceVersion:       "source-v1",
		ContractVersion:     "contract-v1",
		RouteStatus:         "ACCEPT_CANDIDATE",
		RouteRecordDigest:   "sha256:route",
		DecisionDigest:      "sha256:decision",
		ExpectedRouteStatus: "ACCEPT_CANDIDATE",
		ObservedRouteStatus: "ACCEPT_CANDIDATE",
		OutcomeDigest:       "sha256:outcome",
		NonAuthorizing:      false,
	}
	binding := BindDecisionRouteReverseObservationFromSource(input)
	if binding.Status != DecisionRouteReverseUnknown || binding.FirstMismatch != "source-text" {
		t.Fatalf("missing source = %#v", binding)
	}
	if !binding.Validate() {
		t.Fatal("expected missing source binding to validate structurally")
	}

	input.SourceText = "module route\n"
	binding = BindDecisionRouteReverseObservationFromSource(input)
	if binding.FirstMismatch != "authorization-boundary" {
		t.Fatalf("missing boundary = %#v", binding)
	}
	if !binding.Validate() {
		t.Fatal("expected boundary refusal to validate structurally")
	}
}

func TestDecisionRouteReverseObservationSourceBindingRejectsTampering(t *testing.T) {
	input := DecisionRouteReverseObservationSourceInput{
		SourceVersion:       "source-v1",
		ContractVersion:     "contract-v1",
		SourceText:          "module route\nstatus ACCEPT_CANDIDATE\n",
		RouteStatus:         "ACCEPT_CANDIDATE",
		RouteRecordDigest:   "sha256:route",
		DecisionDigest:      "sha256:decision",
		ExpectedRouteStatus: "ACCEPT_CANDIDATE",
		ObservedRouteStatus: "ACCEPT_CANDIDATE",
		OutcomeDigest:       "sha256:outcome",
		NonAuthorizing:      true,
	}
	binding := BindDecisionRouteReverseObservationFromSource(input)
	binding.ObservationDigest = "sha256:tampered"
	if binding.Validate() {
		t.Fatal("expected tampered observation digest to fail validation")
	}
}

