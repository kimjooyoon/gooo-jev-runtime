package jevcal

import "testing"

func TestProjectOutcomeWindowDeltaObservationLSPPreservesPositiveDelta(t *testing.T) {
	observation := ObserveOutcomeWindowDelta(OutcomeWindowDeltaObservationInput{
		SourceVersion:        "source-v1",
		ContractVersion:      "contract-v1",
		EvidencePrefixDigest: "sha256:prefix-v1",
		ObservationDigest:    "sha256:observation-v1",
		ExpectedOutcomeCount: 4,
		ObservedOutcomeCount: 7,
	})
	projection := ProjectOutcomeWindowDeltaObservationLSP(observation)
	if projection.Status != OutcomeWindowDeltaLSPBound || projection.Code != "jev.outcome_window_delta.bound" || projection.Severity != "Information" || projection.Delta != 3 {
		t.Fatalf("unexpected positive delta projection: %+v", projection)
	}
	if !projection.IsReadOnly || projection.CanExecute || projection.CanAuthorize || len(projection.Edits) != 0 || projection.Command != "" {
		t.Fatalf("projection crossed authority boundary: %+v", projection)
	}
}

func TestProjectOutcomeWindowDeltaObservationLSPPreservesNegativeDelta(t *testing.T) {
	observation := ObserveOutcomeWindowDelta(OutcomeWindowDeltaObservationInput{
		SourceVersion:        "source-v1",
		ContractVersion:      "contract-v1",
		EvidencePrefixDigest: "sha256:prefix-v1",
		ObservationDigest:    "sha256:observation-v1",
		ExpectedOutcomeCount: 7,
		ObservedOutcomeCount: 4,
	})
	projection := ProjectOutcomeWindowDeltaObservationLSP(observation)
	if projection.Status != OutcomeWindowDeltaLSPBound || projection.Delta != -3 || projection.Title != "Signed outcome-window delta is recorded" {
		t.Fatalf("negative delta was misclassified: %+v", projection)
	}
}

func TestProjectOutcomeWindowDeltaObservationLSPPreservesMissingDigest(t *testing.T) {
	projection := ProjectOutcomeWindowDeltaObservationLSP(OutcomeWindowDeltaObservation{
		Status:      OutcomeWindowDeltaUnknown,
		TargetStage: "evidence_prefix",
		Reason:      "evidence prefix digest is missing",
		Delta:       -2,
	})
	if projection.Code != "jev.outcome_window_delta.unknown" || projection.TargetStage != "record_digest" || projection.Severity != "Error" || projection.Delta != -2 {
		t.Fatalf("missing record digest was not preserved: %+v", projection)
	}
}
