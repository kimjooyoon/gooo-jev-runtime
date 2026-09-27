package jevcal

import "testing"

func TestProjectPlanBoundaryProducesProposalWithoutAuthority(t *testing.T) {
	projection := ProjectPlanBoundary(PlanBoundaryInput{
		SourceVersion:     "source-v1",
		ContractVersion:   "contract-v1",
		RouteRecordDigest: "sha256:record-v1",
		PlanDigest:        "sha256:plan-v1",
		RouteStatus:       "REVIEW_CANDIDATE",
	})
	if projection.Status != PlanBoundaryProposed || projection.TargetStage != "plan_application" || projection.BoundaryDigest == "" {
		t.Fatalf("unexpected plan boundary: %+v", projection)
	}
	if !projection.IsReadOnly || projection.CanExecute || projection.CanAuthorize || len(projection.Edits) != 0 || projection.Command != "" {
		t.Fatalf("plan boundary crossed authority: %+v", projection)
	}
}

func TestProjectPlanBoundaryPreservesDeferredProducer(t *testing.T) {
	projection := ProjectPlanBoundary(PlanBoundaryInput{
		SourceVersion:     "source-v1",
		ContractVersion:   "contract-v1",
		RouteRecordDigest: "sha256:record-v1",
		PlanDigest:        "sha256:plan-v1",
		RouteStatus:       "ACCEPT_CANDIDATE",
		ProducerDeferred:  true,
	})
	if projection.Status != PlanBoundaryDeferred || projection.Reason != "plan producer is deferred" {
		t.Fatalf("deferred producer was not preserved: %+v", projection)
	}
}

func TestProjectPlanBoundaryRejectsMissingEvidence(t *testing.T) {
	projection := ProjectPlanBoundary(PlanBoundaryInput{
		SourceVersion: "source-v1",
		ContractVersion: "contract-v1",
		RouteStatus: "ACCEPT_CANDIDATE",
	})
	if projection.Status != PlanBoundaryUnknown || projection.TargetStage != "route_record" || projection.BoundaryDigest == "" {
		t.Fatalf("missing route evidence was not preserved: %+v", projection)
	}
}

func TestProjectPlanBoundaryPreservesUnknownRoute(t *testing.T) {
	projection := ProjectPlanBoundary(PlanBoundaryInput{
		SourceVersion:     "source-v1",
		ContractVersion:   "contract-v1",
		RouteRecordDigest: "sha256:record-v1",
		PlanDigest:        "sha256:plan-v1",
		RouteStatus:       "UNKNOWN",
	})
	if projection.Status != PlanBoundaryUnknown || projection.Reason != "route evidence remains unresolved" {
		t.Fatalf("unknown route was not preserved: %+v", projection)
	}
}