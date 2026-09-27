package jevcal

import "testing"

func TestRecordDecisionRouteProjectionPreservesEvidenceAndBoundary(t *testing.T) {
	projection := ProjectDecisionRouteLSP("REVIEW_CANDIDATE", "calibration", "sha256:route-v1")
	record := RecordDecisionRouteProjection(projection, "source-v1", "contract-v1")
	if record.Status != "REVIEW_CANDIDATE" || record.Code != "jev.route.review_candidate" || record.DecisionDigest != "sha256:route-v1" {
		t.Fatalf("unexpected route record: %+v", record)
	}
	if record.RecordDigest == "" || record.EvidenceKind == "" || !record.IsReadOnly || record.CanExecute || record.CanAuthorize {
		t.Fatalf("route record crossed authority boundary or lacks provenance: %+v", record)
	}
}

func TestRecordDecisionRouteProjectionPreservesUnknownIdentity(t *testing.T) {
	projection := ProjectDecisionRouteLSP("ACCEPT_CANDIDATE", "stage", "sha256:route-v1")
	record := RecordDecisionRouteProjection(projection, "", "contract-v1")
	if record.Status != "UNKNOWN" || record.Code != "jev.route.unknown" || record.RecordDigest == "" {
		t.Fatalf("missing source identity was not preserved: %+v", record)
	}
}

func TestRecordDecisionRouteProjectionRejectsMissingDecisionEvidence(t *testing.T) {
	projection := ProjectDecisionRouteLSP("ACCEPT_CANDIDATE", "stage", "")
	record := RecordDecisionRouteProjection(projection, "source-v1", "contract-v1")
	if record.Status != "UNKNOWN" || record.DecisionDigest != "" || record.RecordDigest == "" {
		t.Fatalf("missing decision evidence was not preserved: %+v", record)
	}
}