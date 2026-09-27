package jevcal

import "testing"

func TestObserveProvenanceStageCoverageBound(t *testing.T) {
	observation := ObserveProvenanceStageCoverage(ProvenanceStageCoverageInput{
		SourceVersion:            "src-v1",
		ContractVersion:          "contract-v1",
		ExpectedStages:           []string{"source", "ir", "generated", "reverse_observation"},
		ObservedStages:           []string{"source", "ir", "generated", "reverse_observation"},
		SourceDigest:             "sha256:source",
		IRDigest:                 "sha256:ir",
		GeneratedDigest:          "sha256:generated",
		ReverseObservationDigest: "sha256:reverse",
	})

	if observation.Status != ProvenanceStageCoverageBound {
		t.Fatalf("status = %q, want %q", observation.Status, ProvenanceStageCoverageBound)
	}
	if observation.CoverageNumerator != 4 || observation.CoverageDenominator != 4 || observation.CoveragePercent != 100 {
		t.Fatalf("coverage = %d/%d (%d%%), want 4/4 (100%%)", observation.CoverageNumerator, observation.CoverageDenominator, observation.CoveragePercent)
	}
	if observation.MissingStage != "" || !observation.ObservationalOnly || observation.ClaimsImprovement {
		t.Fatalf("unexpected bound observation: %+v", observation)
	}
}

func TestObserveProvenanceStageCoveragePreservesFirstMissingStage(t *testing.T) {
	observation := ObserveProvenanceStageCoverage(ProvenanceStageCoverageInput{
		SourceVersion:     "src-v1",
		ContractVersion:   "contract-v1",
		ExpectedStages:    []string{"source", "ir", "generated", "reverse_observation"},
		ObservedStages:    []string{"source", "ir"},
		SourceDigest:      "sha256:source",
		IRDigest:          "sha256:ir",
	})

	if observation.Status != ProvenanceStageCoverageUnknown {
		t.Fatalf("status = %q, want %q", observation.Status, ProvenanceStageCoverageUnknown)
	}
	if observation.MissingStage != "generated" || observation.TargetStage != "observed_stage" {
		t.Fatalf("missing stage = %q, target = %q", observation.MissingStage, observation.TargetStage)
	}
	if observation.CoverageNumerator != 2 || observation.CoveragePercent != 50 {
		t.Fatalf("coverage = %d/%d (%d%%), want 2/4 (50%%)", observation.CoverageNumerator, observation.CoverageDenominator, observation.CoveragePercent)
	}
}

func TestObserveProvenanceStageCoverageRequiresStageDigest(t *testing.T) {
	observation := ObserveProvenanceStageCoverage(ProvenanceStageCoverageInput{
		SourceVersion:     "src-v1",
		ContractVersion:   "contract-v1",
		ExpectedStages:    []string{"source", "ir"},
		ObservedStages:    []string{"source", "ir"},
		SourceDigest:      "sha256:source",
	})

	if observation.TargetStage != "stage_digest" || observation.MissingStage != "ir" {
		t.Fatalf("target = %q, missing = %q", observation.TargetStage, observation.MissingStage)
	}
	if observation.Status != ProvenanceStageCoverageUnknown {
		t.Fatalf("status = %q, want %q", observation.Status, ProvenanceStageCoverageUnknown)
	}
}

func TestObserveProvenanceStageCoveragePreservesDeferredState(t *testing.T) {
	observation := ObserveProvenanceStageCoverage(ProvenanceStageCoverageInput{
		SourceVersion:     "src-v1",
		ContractVersion:   "contract-v1",
		ExpectedStages:    []string{"source", "ir"},
		ObservedStages:    []string{"source", "ir"},
		SourceDigest:      "sha256:source",
		IRDigest:          "sha256:ir",
		ProducerDeferred:  true,
	})

	if observation.Status != ProvenanceStageCoverageDeferred {
		t.Fatalf("status = %q, want %q", observation.Status, ProvenanceStageCoverageDeferred)
	}
	if observation.ClaimsImprovement || !observation.ObservationalOnly || observation.CanExecute || observation.CanAuthorize {
		t.Fatalf("deferred observation crossed a forbidden boundary: %+v", observation)
	}
}

func TestObserveProvenanceStageCoverageDigestTracksEvidence(t *testing.T) {
	input := ProvenanceStageCoverageInput{
		SourceVersion:     "src-v1",
		ContractVersion:   "contract-v1",
		ExpectedStages:    []string{"source"},
		ObservedStages:    []string{"source"},
		SourceDigest:      "sha256:source",
	}
	first := ObserveProvenanceStageCoverage(input)
	input.SourceDigest = "sha256:tampered"
	second := ObserveProvenanceStageCoverage(input)

	if first.ObservationDigest == second.ObservationDigest {
		t.Fatal("observation digest did not change with source evidence")
	}
}
