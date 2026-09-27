package jevcal

import "testing"

func generationTraceTestInput() GenerationTraceInput {
	return GenerationTraceInput{
		SourceVersion:            "source-v1",
		ContractVersion:          "contract-v1",
		ContractName:             GenerationTraceContractName,
		ContractDigest:            "sha256:contract",
		IRDigest:                 "sha256:ir",
		GeneratedArtifactDigest:  "sha256:generated",
		ReverseObservationDigest: "sha256:reverse",
		MetricDigest:             "sha256:metric",
		EvidencePrefixDigest:     "sha256:prefix",
	}
}

func TestObserveGenerationTraceBound(t *testing.T) {
	observation := ObserveGenerationTrace(generationTraceTestInput())
	if observation.Status != GenerationTraceBound {
		t.Fatalf("status = %q, want %q", observation.Status, GenerationTraceBound)
	}
	if observation.MissingStage != "" || observation.TraceSignal != generationTraceBoundSignal {
		t.Fatalf("unexpected bound observation: %+v", observation)
	}
	if observation.ClaimsImprovement || observation.CanExecute || observation.CanAuthorize {
		t.Fatalf("bound observation crossed a forbidden boundary: %+v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestObserveGenerationTracePreservesFirstMissingStage(t *testing.T) {
	input := generationTraceTestInput()
	input.GeneratedArtifactDigest = ""
	input.ReverseObservationDigest = ""
	observation := ObserveGenerationTrace(input)

	if observation.Status != GenerationTraceUnknown {
		t.Fatalf("status = %q, want %q", observation.Status, GenerationTraceUnknown)
	}
	if observation.MissingStage != "generated_artifact_digest" || observation.TargetStage != "generated_artifact_digest" {
		t.Fatalf("missing stage = %q, target = %q", observation.MissingStage, observation.TargetStage)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestObserveGenerationTracePreservesDeferredState(t *testing.T) {
	input := generationTraceTestInput()
	input.ProducerDeferred = true
	observation := ObserveGenerationTrace(input)

	if observation.Status != GenerationTraceDeferred {
		t.Fatalf("status = %q, want %q", observation.Status, GenerationTraceDeferred)
	}
	if observation.MissingStage != "" || observation.ClaimsImprovement || observation.CanExecute || observation.CanAuthorize {
		t.Fatalf("deferred observation crossed a forbidden boundary: %+v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestObserveGenerationTraceDigestTracksEvidence(t *testing.T) {
	first := ObserveGenerationTrace(generationTraceTestInput())
	input := generationTraceTestInput()
	input.MetricDigest = "sha256:tampered"
	second := ObserveGenerationTrace(input)

	if first.ObservationDigest == second.ObservationDigest {
		t.Fatal("observation digest did not change with metric evidence")
	}
}

func TestGenerationTraceRejectsTamperedObservation(t *testing.T) {
	observation := ObserveGenerationTrace(generationTraceTestInput())
	observation.MetricDigest = "sha256:tampered"
	if err := observation.Validate(); err == nil {
		t.Fatal("Validate() accepted a tampered observation")
	}
}