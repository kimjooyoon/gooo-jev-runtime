package jevcal

import "testing"

func TestProjectDecisionRouteLSPValidateCandidates(t *testing.T) {
	for _, status := range []string{"ACCEPT_CANDIDATE", "REVIEW_CANDIDATE", "ABSTAIN_CANDIDATE"} {
		t.Run(status, func(t *testing.T) {
			if err := ProjectDecisionRouteLSP(status, "decision_route", "decision-v1").Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func TestProjectDecisionRouteLSPValidateDeferredAndUnknown(t *testing.T) {
	if err := ProjectDecisionRouteLSP("DEFERRED", "decision_route", "decision-v1").Validate(); err != nil {
		t.Fatalf("deferred Validate() error = %v", err)
	}
	if err := ProjectDecisionRouteLSP("UNKNOWN", "", "decision-v1").Validate(); err != nil {
		t.Fatalf("unknown Validate() error = %v", err)
	}
}

func TestDecisionRouteLSPValidateRejectsCommandAndEdits(t *testing.T) {
	projection := ProjectDecisionRouteLSP("REVIEW_CANDIDATE", "decision_route", "decision-v1")
	projection.Command = "execute"
	if err := projection.Validate(); err == nil {
		t.Fatal("decision route command was accepted")
	}
}

func TestDecisionRouteLSPValidateRejectsMissingDigest(t *testing.T) {
	projection := ProjectDecisionRouteLSP("REVIEW_CANDIDATE", "decision_route", "")
	if err := projection.Validate(); err == nil {
		t.Fatal("decision route without a digest was accepted")
	}
}