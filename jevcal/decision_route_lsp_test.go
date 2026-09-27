package jevcal

import "testing"

func TestProjectDecisionRouteLSPPreservesInspectionOnlyBoundary(t *testing.T) {
	projection := ProjectDecisionRouteLSP("ACCEPT_CANDIDATE", "reverse_observation", "sha256:route-v1")
	if projection.Code != "jev.route.accept_candidate" || projection.Severity != "Information" {
		t.Fatalf("unexpected acceptance projection: %+v", projection)
	}
	if !projection.Actionable || !projection.ReadOnly || projection.CanExecute || projection.CanAuthorize || len(projection.Edits) != 0 || projection.Command != "" {
		t.Fatalf("projection crossed an authority boundary: %+v", projection)
	}
}

func TestProjectDecisionRouteLSPMapsReviewAndAbstain(t *testing.T) {
	tests := []struct {
		status string
		code   string
		level  string
	}{
		{status: "REVIEW_CANDIDATE", code: "jev.route.review_candidate", level: "Warning"},
		{status: "ABSTAIN_CANDIDATE", code: "jev.route.abstain_candidate", level: "Information"},
		{status: "DEFERRED", code: "jev.route.deferred", level: "Hint"},
	}
	for _, test := range tests {
		projection := ProjectDecisionRouteLSP(test.status, "calibration", "sha256:route-v1")
		if projection.Code != test.code || projection.Severity != test.level {
			t.Fatalf("status %q produced %+v", test.status, projection)
		}
	}
}

func TestProjectDecisionRouteLSPRejectsMissingOrUnknownEvidence(t *testing.T) {
	missing := ProjectDecisionRouteLSP("ACCEPT_CANDIDATE", "stage", "")
	if missing.Status != "UNKNOWN" || missing.TargetStage != "decision_digest" {
		t.Fatalf("missing digest was not preserved: %+v", missing)
	}
	unknown := ProjectDecisionRouteLSP("NOT_A_ROUTE", "stage", "sha256:route-v1")
	if unknown.Status != "UNKNOWN" || unknown.TargetStage != "route_status" || unknown.Code != "jev.route.unknown" {
		t.Fatalf("unknown status was not preserved: %+v", unknown)
	}
}