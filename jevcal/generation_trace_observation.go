package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	GenerationTraceBound    = "BOUND"
	GenerationTraceDeferred = "DEFERRED"
	GenerationTraceUnknown  = "UNKNOWN"

	GenerationTraceContractName = "jev_generation_trace_observation"

	generationTraceBoundSignal    = "jev-generation-trace-bound"
	generationTraceDeferredSignal = "jev-generation-trace-deferred"
	generationTraceUnknownSignal  = "jev-generation-trace-unknown"
)

// GenerationTraceInput links a declaration, its IR, generated artifact,
// reverse observation, and metric evidence without granting execution or
// authorization.
type GenerationTraceInput struct {
	SourceVersion            string
	ContractVersion          string
	ContractName             string
	ContractDigest            string
	IRDigest                 string
	GeneratedArtifactDigest  string
	ReverseObservationDigest string
	MetricDigest             string
	EvidencePrefixDigest     string
	ProducerDeferred         bool
}

// GenerationTraceObservation is an evidence-availability observation, not an
// improvement or correctness judgment.
type GenerationTraceObservation struct {
	Status                    string
	SourceVersion             string
	ContractVersion           string
	ContractName              string
	ContractDigest            string
	IRDigest                  string
	GeneratedArtifactDigest   string
	ReverseObservationDigest  string
	MetricDigest              string
	EvidencePrefixDigest      string
	MissingStage              string
	TargetStage               string
	Reason                    string
	TraceSignal               string
	ObservationDigest         string
	ObservationalOnly         bool
	ClaimsImprovement         bool
	CanExecute                bool
	CanAuthorize              bool
}

// ObserveGenerationTrace preserves the first missing stage and binds only
// complete, aligned evidence.
func ObserveGenerationTrace(input GenerationTraceInput) GenerationTraceObservation {
	observation := GenerationTraceObservation{
		Status:                      GenerationTraceUnknown,
		SourceVersion:               input.SourceVersion,
		ContractVersion:             input.ContractVersion,
		ContractName:                input.ContractName,
		ContractDigest:              input.ContractDigest,
		IRDigest:                    input.IRDigest,
		GeneratedArtifactDigest:     input.GeneratedArtifactDigest,
		ReverseObservationDigest:    input.ReverseObservationDigest,
		MetricDigest:                input.MetricDigest,
		EvidencePrefixDigest:        input.EvidencePrefixDigest,
		TargetStage:                 "generation_trace_evidence",
		TraceSignal:                 generationTraceUnknownSignal,
		ObservationalOnly:           true,
		ClaimsImprovement:           false,
		CanExecute:                  false,
		CanAuthorize:                false,
	}

	switch {
	case input.SourceVersion == "":
		observation.TargetStage = "source_identity"
		observation.MissingStage = "source_identity"
		observation.Reason = "source identity is missing"
	case input.ContractVersion == "":
		observation.TargetStage = "contract_identity"
		observation.MissingStage = "contract_identity"
		observation.Reason = "contract version is missing"
	case input.ContractName != GenerationTraceContractName:
		observation.TargetStage = "contract_identity"
		observation.MissingStage = "contract_identity"
		observation.Reason = "generation trace contract identity is not recognized"
	case input.ContractDigest == "":
		observation.TargetStage = "contract_digest"
		observation.MissingStage = "contract_digest"
		observation.Reason = "contract digest is missing"
	case input.ProducerDeferred:
		observation.Status = GenerationTraceDeferred
		observation.TraceSignal = generationTraceDeferredSignal
		observation.TargetStage = "generation_trace_producer"
		observation.Reason = "generation trace producer is deferred"
	case input.IRDigest == "":
		observation.TargetStage = "ir_digest"
		observation.MissingStage = "ir_digest"
		observation.Reason = "IR digest is missing"
	case input.GeneratedArtifactDigest == "":
		observation.TargetStage = "generated_artifact_digest"
		observation.MissingStage = "generated_artifact_digest"
		observation.Reason = "generated artifact digest is missing"
	case input.ReverseObservationDigest == "":
		observation.TargetStage = "reverse_observation_digest"
		observation.MissingStage = "reverse_observation_digest"
		observation.Reason = "reverse observation digest is missing"
	case input.MetricDigest == "":
		observation.TargetStage = "metric_digest"
		observation.MissingStage = "metric_digest"
		observation.Reason = "metric digest is missing"
	case input.EvidencePrefixDigest == "":
		observation.TargetStage = "evidence_prefix_digest"
		observation.MissingStage = "evidence_prefix_digest"
		observation.Reason = "evidence prefix digest is missing"
	default:
		observation.Status = GenerationTraceBound
		observation.MissingStage = ""
		observation.TargetStage = "generation_trace_evidence"
		observation.Reason = "declaration, IR, generation, reverse observation, and metric evidence are linked"
		observation.TraceSignal = generationTraceBoundSignal
	}

	observation.ObservationDigest = generationTraceObservationDigest(observation)
	return observation
}

func (observation GenerationTraceObservation) Validate() error {
	switch observation.Status {
	case GenerationTraceBound, GenerationTraceDeferred, GenerationTraceUnknown:
	default:
		return fmt.Errorf("invalid generation trace status %q", observation.Status)
	}
	if observation.TargetStage == "" || observation.Reason == "" || observation.ObservationDigest == "" {
		return fmt.Errorf("generation trace evidence identity is incomplete")
	}
	if !observation.ObservationalOnly || observation.ClaimsImprovement || observation.CanExecute || observation.CanAuthorize {
		return fmt.Errorf("generation trace crossed an execution, authorization, or improvement boundary")
	}

	switch observation.Status {
	case GenerationTraceBound:
		if observation.TraceSignal != generationTraceBoundSignal ||
			observation.MissingStage != "" ||
			observation.SourceVersion == "" ||
			observation.ContractVersion == "" ||
			observation.ContractName != GenerationTraceContractName ||
			observation.ContractDigest == "" ||
			observation.IRDigest == "" ||
			observation.GeneratedArtifactDigest == "" ||
			observation.ReverseObservationDigest == "" ||
			observation.MetricDigest == "" ||
			observation.EvidencePrefixDigest == "" {
			return fmt.Errorf("bound generation trace is incomplete")
		}
	case GenerationTraceDeferred:
		if observation.TraceSignal != generationTraceDeferredSignal || observation.MissingStage != "" {
			return fmt.Errorf("deferred generation trace signal is invalid")
		}
	case GenerationTraceUnknown:
		if observation.TraceSignal != generationTraceUnknownSignal || observation.MissingStage == "" {
			return fmt.Errorf("unknown generation trace must preserve its missing stage")
		}
	}

	if observation.ObservationDigest != generationTraceObservationDigest(observation) {
		return fmt.Errorf("generation trace observation digest mismatch")
	}
	return nil
}

func generationTraceObservationDigest(observation GenerationTraceObservation) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		observation.Status,
		observation.SourceVersion,
		observation.ContractVersion,
		observation.ContractName,
		observation.ContractDigest,
		observation.IRDigest,
		observation.GeneratedArtifactDigest,
		observation.ReverseObservationDigest,
		observation.MetricDigest,
		observation.EvidencePrefixDigest,
		observation.MissingStage,
		observation.TargetStage,
		observation.Reason,
		observation.TraceSignal,
		fmt.Sprint(observation.ObservationalOnly),
		fmt.Sprint(observation.ClaimsImprovement),
		fmt.Sprint(observation.CanExecute),
		fmt.Sprint(observation.CanAuthorize),
	}, "|")))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}