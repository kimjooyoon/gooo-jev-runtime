package jevcal

import "testing"

func TestProjectGenerationTraceLSPProjectsBoundObservation(t *testing.T) {
	observation := ObserveGenerationTrace(generationTraceTestInput())
	projection := ProjectGenerationTraceLSP(observation)

	if projection.Status != GenerationTraceLSPBound ||
		projection.Code != generationTraceLSPComplete ||
		projection.Publishable ||
		projection.MissingStage != "" ||
		projection.MissingStageIndex != -1 ||
		projection.MetricDigest != observation.MetricDigest ||
		projection.TraceObservationDigest != observation.ObservationDigest {
		t.Fatalf("unexpected bound projection: %+v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestProjectGenerationTraceLSPPreservesUnknownStageIndex(t *testing.T) {
	input := generationTraceTestInput()
	input.GeneratedArtifactDigest = ""
	observation := ObserveGenerationTrace(input)
	projection := ProjectGenerationTraceLSP(observation)

	if projection.Status != GenerationTraceLSPUnknown ||
		!projection.Publishable ||
		projection.Code != generationTraceLSPUnknown ||
		projection.MissingStage != "generated_artifact_digest" ||
		projection.MissingStageIndex != 5 {
		t.Fatalf("unexpected unknown projection: %+v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestProjectGenerationTraceLSPProjectsDeferredObservation(t *testing.T) {
	input := generationTraceTestInput()
	input.ProducerDeferred = true
	projection := ProjectGenerationTraceLSP(ObserveGenerationTrace(input))

	if projection.Status != GenerationTraceLSPDeferred ||
		!projection.Publishable ||
		projection.Severity != "warning" ||
		projection.MissingStageIndex != 3 {
		t.Fatalf("unexpected deferred projection: %+v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestProjectGenerationTraceLSPRejectsTamperedObservation(t *testing.T) {
	observation := ObserveGenerationTrace(generationTraceTestInput())
	observation.MetricDigest = "sha256:tampered"
	projection := ProjectGenerationTraceLSP(observation)

	if projection.Status != GenerationTraceLSPUnknown ||
		projection.Code != generationTraceLSPIntegrity ||
		projection.MissingStage != "observation-integrity" ||
		projection.Publishable {
		t.Fatalf("tampered observation was projected: %+v", projection)
	}
}

func TestProjectGenerationTraceLSPFailsClosedOnCapabilityBoundary(t *testing.T) {
	observation := ObserveGenerationTrace(generationTraceTestInput())
	observation.CanExecute = true
	projection := ProjectGenerationTraceLSP(observation)

	if projection.Status != GenerationTraceLSPUnknown ||
		projection.Code != generationTraceLSPBoundary ||
		projection.NonAuthorizing ||
		projection.Publishable {
		t.Fatalf("capability boundary was not preserved: %+v", projection)
	}
}

func TestGenerationTraceMissingStageIndexUsesCanonicalOrder(t *testing.T) {
	cases := map[string]int{
		"source_identity":          0,
		"contract_identity":        1,
		"contract_digest":          2,
		"generation_trace_producer": 3,
		"ir_digest":                4,
		"generated_artifact_digest": 5,
		"reverse_observation_digest": 6,
		"metric_digest":             7,
		"evidence_prefix_digest":    8,
	}
	for stage, want := range cases {
		got, ok := GenerationTraceMissingStageIndex(stage)
		if !ok || got != want {
			t.Fatalf("stage %q index = (%d, %v), want (%d, true)", stage, got, ok, want)
		}
	}
}