package jevcal

import "testing"

func TestProjectDecisionPreservesUnknownEvidence(t *testing.T) {
	decision := Decision{
		Status:               StatusUnknown,
		EvidencePrefixDigest: "prefix-digest",
	}
	diagnostic := ProjectDecision(decision)
	if diagnostic.Code != "JEV_UNKNOWN" {
		t.Fatalf("code = %q, want JEV_UNKNOWN", diagnostic.Code)
	}
	if diagnostic.EvidencePrefixDigest != decision.EvidencePrefixDigest {
		t.Fatalf("prefix digest = %q, want %q", diagnostic.EvidencePrefixDigest, decision.EvidencePrefixDigest)
	}
}

func TestProjectDecisionMapsRegressionToWarning(t *testing.T) {
	diagnostic := ProjectDecision(Decision{Status: StatusRegressed})
	if diagnostic.Code != "JEV_REGRESSED" || diagnostic.Severity != SeverityWarning {
		t.Fatalf("diagnostic = %#v, want regression warning", diagnostic)
	}
}