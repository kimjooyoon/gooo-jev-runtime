package jevcal

import "testing"

func provenanceStageCoverageLSPTestObservation() ProvenanceStageCoverage {
	return ObserveProvenanceStageCoverage(ProvenanceStageCoverageInput{
		SourceVersion:            "gooo-1",
		ContractVersion:          "jev-1",
		ExpectedStages:           []string{"source", "ir", "generated", "reverse_observation"},
		ObservedStages:           []string{"source", "ir", "generated", "reverse_observation"},
		SourceDigest:             "sha256:source",
		IRDigest:                 "sha256:ir",
		GeneratedDigest:          "sha256:generated",
		ReverseObservationDigest: "sha256:reverse",
	})
}

func TestProjectProvenanceStageCoverageLSPBound(t *testing.T) {
	projection := ProjectProvenanceStageCoverageLSP(provenanceStageCoverageLSPTestObservation())
	if projection.Status != ProvenanceStageCoverageLSPBound ||
		projection.Code != provenanceStageCoverageLSPComplete || projection.MissingStage != "" {
		t.Fatalf("bound provenance coverage was not preserved: %+v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestProjectProvenanceStageCoverageLSPPreservesDeferred(t *testing.T) {
	observation := provenanceStageCoverageLSPTestObservation()
	observation.Status = ProvenanceStageCoverageDeferred
	observation.ProducerDeferred = true
	projection := ProjectProvenanceStageCoverageLSP(observation)
	if projection.Status != ProvenanceStageCoverageLSPDeferred ||
		projection.MissingStage != "provenance_stage_producer" || !projection.Publishable {
		t.Fatalf("deferred provenance coverage was not preserved: %+v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestProjectProvenanceStageCoverageLSPPreservesUnknownStage(t *testing.T) {
	observation := ObserveProvenanceStageCoverage(ProvenanceStageCoverageInput{
		SourceVersion:   "gooo-1",
		ContractVersion: "jev-1",
		ExpectedStages:  []string{"source", "ir"},
		ObservedStages:  []string{"source"},
		SourceDigest:    "sha256:source",
	})
	projection := ProjectProvenanceStageCoverageLSP(observation)
	if projection.Status != ProvenanceStageCoverageLSPUnknown ||
		projection.MissingStage != "ir" || projection.MissingStageIndex != 2 {
		t.Fatalf("unknown provenance stage was not preserved: %+v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestProjectProvenanceStageCoverageLSPFailsClosedOnBoundary(t *testing.T) {
	observation := provenanceStageCoverageLSPTestObservation()
	observation.ClaimsImprovement = true
	projection := ProjectProvenanceStageCoverageLSP(observation)
	if projection.Status != ProvenanceStageCoverageLSPError ||
		projection.Code != provenanceStageCoverageLSPBoundary || projection.MissingStage != "capability-boundary" {
		t.Fatalf("unsafe provenance coverage was not bounded: %+v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestProvenanceStageCoverageLSPRejectsProjectionTampering(t *testing.T) {
	projection := ProjectProvenanceStageCoverageLSP(provenanceStageCoverageLSPTestObservation())
	projection.CoveragePercent = 99
	if err := projection.Validate(); err == nil {
		t.Fatal("tampered provenance coverage projection was accepted")
	}
}