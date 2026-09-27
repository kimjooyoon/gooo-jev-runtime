package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	GenerationTraceLSPBound    = "BOUND"
	GenerationTraceLSPDeferred = "DEFERRED"
	GenerationTraceLSPUnknown  = "UNKNOWN"

	generationTraceLSPComplete = "generation-trace-complete"
	generationTraceLSPDeferred = "generation-trace-deferred"
	generationTraceLSPUnknown  = "generation-trace"
	generationTraceLSPIntegrity = "observation-integrity"
	generationTraceLSPBoundary = "capability-boundary"
)

// GenerationTraceLSPProjection exposes generation trace evidence to an editor
// without executing, authorizing, or judging an improvement.
type GenerationTraceLSPProjection struct {
	Status                 string
	Publishable            bool
	Severity               string
	Code                   string
	MissingStage           string
	MissingStageIndex      int
	TraceSignal            string
	MetricDigest           string
	EvidencePrefixDigest   string
	TraceObservationDigest string
	ObservationDigest      string
	ReadOnly               bool
	NonAuthorizing         bool
	CanExecute             bool
	CanAuthorize           bool
}

// GenerationTraceMissingStageIndex derives the canonical stage index instead
// of trusting a caller-provided editor location.
func GenerationTraceMissingStageIndex(stage string) (int, bool) {
	stages := []string{
		"source_identity",
		"contract_identity",
		"contract_digest",
		"generation_trace_producer",
		"ir_digest",
		"generated_artifact_digest",
		"reverse_observation_digest",
		"metric_digest",
		"evidence_prefix_digest",
	}
	for index, candidate := range stages {
		if stage == candidate {
			return index, true
		}
	}
	return -1, false
}

// ProjectGenerationTraceLSP converts a validated observation into a
// read-only diagnostic or clear projection while preserving UNKNOWN.
func ProjectGenerationTraceLSP(input GenerationTraceObservation) GenerationTraceLSPProjection {
	output := GenerationTraceLSPProjection{
		Status:               GenerationTraceLSPUnknown,
		Publishable:          false,
		Severity:             "error",
		Code:                 generationTraceLSPUnknown,
		MissingStage:         input.MissingStage,
		MissingStageIndex:    -1,
		TraceSignal:          input.TraceSignal,
		EvidencePrefixDigest: input.EvidencePrefixDigest,
		ReadOnly:             true,
		NonAuthorizing:       true,
		CanExecute:           false,
		CanAuthorize:         false,
	}
	if output.MissingStage == "" {
		output.MissingStage = "generation_trace_observation"
	}
	if !input.ObservationalOnly || input.ClaimsImprovement || input.CanExecute || input.CanAuthorize {
		output.NonAuthorizing = false
		output.Code = generationTraceLSPBoundary
		output.MissingStage = "capability-boundary"
		output.ObservationDigest = generationTraceLSPDigest(output)
		return output
	}
	if err := input.Validate(); err != nil {
		output.Code = generationTraceLSPIntegrity
		output.MissingStage = "observation-integrity"
		output.ObservationDigest = generationTraceLSPDigest(output)
		return output
	}

	output.TraceObservationDigest = input.ObservationDigest
	switch input.Status {
	case GenerationTraceBound:
		output.Status = GenerationTraceLSPBound
		output.Publishable = false
		output.Severity = "info"
		output.Code = generationTraceLSPComplete
		output.MissingStage = ""
		output.MissingStageIndex = -1
		output.MetricDigest = input.MetricDigest
	case GenerationTraceDeferred:
		output.Status = GenerationTraceLSPDeferred
		output.Publishable = true
		output.Severity = "warning"
		output.Code = generationTraceLSPDeferred
		output.MissingStage = "generation_trace_producer"
		output.MissingStageIndex, _ = GenerationTraceMissingStageIndex(output.MissingStage)
	case GenerationTraceUnknown:
		output.Publishable = true
		output.MissingStageIndex, _ = GenerationTraceMissingStageIndex(output.MissingStage)
		if output.MissingStageIndex < 0 {
			output.Publishable = false
			output.Code = "diagnostic-location"
		}
	default:
		output.Code = generationTraceLSPIntegrity
		output.MissingStage = "observation-status"
	}
	output.ObservationDigest = generationTraceLSPDigest(output)
	return output
}

func (projection GenerationTraceLSPProjection) Validate() error {
	switch projection.Status {
	case GenerationTraceLSPBound, GenerationTraceLSPDeferred, GenerationTraceLSPUnknown:
	default:
		return fmt.Errorf("invalid generation trace LSP status %q", projection.Status)
	}
	if projection.MissingStageIndex < -1 {
		return fmt.Errorf("invalid missing stage index %d", projection.MissingStageIndex)
	}
	if !projection.ReadOnly || !projection.NonAuthorizing || projection.CanExecute || projection.CanAuthorize {
		return fmt.Errorf("generation trace LSP crossed a forbidden boundary")
	}
	if projection.Status == GenerationTraceLSPBound {
		if projection.Publishable ||
			projection.MissingStage != "" ||
			projection.MissingStageIndex != -1 ||
			projection.Code != generationTraceLSPComplete ||
			projection.MetricDigest == "" ||
			projection.EvidencePrefixDigest == "" ||
			projection.TraceObservationDigest == "" {
			return fmt.Errorf("bound generation trace LSP is incomplete")
		}
	}
	if projection.Status == GenerationTraceLSPDeferred {
		if !projection.Publishable ||
			projection.MissingStage != "generation_trace_producer" ||
			projection.MissingStageIndex != 3 ||
			projection.Code != generationTraceLSPDeferred {
			return fmt.Errorf("deferred generation trace LSP is incomplete")
		}
	}
	if projection.Status == GenerationTraceLSPUnknown {
		if projection.MissingStage == "" || projection.MissingStageIndex < 0 {
			return fmt.Errorf("unknown generation trace LSP must preserve a canonical missing stage")
		}
	}
	if projection.ObservationDigest != generationTraceLSPDigest(projection) {
		return fmt.Errorf("generation trace LSP observation digest mismatch")
	}
	return nil
}

func generationTraceLSPDigest(projection GenerationTraceLSPProjection) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		projection.Status,
		fmt.Sprint(projection.Publishable),
		projection.Severity,
		projection.Code,
		projection.MissingStage,
		fmt.Sprint(projection.MissingStageIndex),
		projection.TraceSignal,
		projection.MetricDigest,
		projection.EvidencePrefixDigest,
		projection.TraceObservationDigest,
		fmt.Sprint(projection.ReadOnly),
		fmt.Sprint(projection.NonAuthorizing),
		fmt.Sprint(projection.CanExecute),
		fmt.Sprint(projection.CanAuthorize),
	}, "|")))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}