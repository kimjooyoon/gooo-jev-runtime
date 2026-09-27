package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	ProvenanceStageCoverageLSPBound    = "BOUND"
	ProvenanceStageCoverageLSPDeferred = "DEFERRED"
	ProvenanceStageCoverageLSPUnknown  = "UNKNOWN"
	ProvenanceStageCoverageLSPError    = "ERROR"

	provenanceStageCoverageLSPComplete  = "provenance-stage-coverage-complete"
	provenanceStageCoverageLSPDeferredCode = "provenance-stage-coverage-deferred"
	provenanceStageCoverageLSPUnknownCode  = "provenance-stage-coverage-unknown"
	provenanceStageCoverageLSPIntegrity    = "observation-integrity"
	provenanceStageCoverageLSPBoundary     = "capability-boundary"
)

type ProvenanceStageCoverageLSPProjection struct {
	Status              string
	Publishable         bool
	Severity            string
	Code                string
	CoverageNumerator   int
	CoverageDenominator int
	CoveragePercent     int
	MissingStage        string
	MissingStageIndex   int
	TargetStage         string
	ObservationDigest   string
	EvidencePrefixDigest string
	ProjectionDigest    string
	ReadOnly            bool
	NonAuthorizing      bool
	CanExecute          bool
	CanAuthorize        bool
}

// ProvenanceStageCoverageMissingStageIndex maps observation targets to stable
// editor locations without treating a location as evidence of improvement.
func ProvenanceStageCoverageMissingStageIndex(stage string) (int, bool) {
	stages := []string{
		"identity",
		"expected_stages",
		"observed_stage",
		"stage_digest",
		"provenance_stage_producer",
	}
	for index, candidate := range stages {
		if candidate == stage {
			return index, true
		}
	}
	return -1, false
}

// ProjectProvenanceStageCoverageLSP exposes coverage evidence as a read-only
// diagnostic. It never claims that coverage means improvement or authorization.
func ProjectProvenanceStageCoverageLSP(
	input ProvenanceStageCoverage,
) ProvenanceStageCoverageLSPProjection {
	output := ProvenanceStageCoverageLSPProjection{
		Status:               ProvenanceStageCoverageLSPUnknown,
		Publishable:          false,
		Severity:             "error",
		Code:                 provenanceStageCoverageLSPUnknownCode,
		CoverageNumerator:    input.CoverageNumerator,
		CoverageDenominator:  input.CoverageDenominator,
		CoveragePercent:      input.CoveragePercent,
		MissingStage:         input.MissingStage,
		TargetStage:          input.TargetStage,
		ObservationDigest:    input.ObservationDigest,
		EvidencePrefixDigest: input.ObservationDigest,
		MissingStageIndex:    -1,
		ReadOnly:             true,
		NonAuthorizing:       true,
		CanExecute:           false,
		CanAuthorize:         false,
	}
	if output.TargetStage == "" {
		output.TargetStage = "provenance_stage_coverage"
	}
	if !input.ObservationalOnly || input.ClaimsImprovement || input.CanExecute || input.CanAuthorize {
		output.Status = ProvenanceStageCoverageLSPError
		output.Publishable = true
		output.Code = provenanceStageCoverageLSPBoundary
		output.MissingStage = "capability-boundary"
		output.TargetStage = "capability-boundary"
		output.MissingStageIndex = -1
		output.ProjectionDigest = provenanceStageCoverageLSPDigest(output)
		return output
	}

	switch input.Status {
	case ProvenanceStageCoverageBound:
		output.Status = ProvenanceStageCoverageLSPBound
		output.Severity = "info"
		output.Code = provenanceStageCoverageLSPComplete
		output.MissingStage = ""
		output.MissingStageIndex = -1
	case ProvenanceStageCoverageDeferred:
		output.Status = ProvenanceStageCoverageLSPDeferred
		output.Publishable = true
		output.Severity = "warning"
		output.Code = provenanceStageCoverageLSPDeferredCode
		output.MissingStage = "provenance_stage_producer"
		output.TargetStage = "provenance_stage_producer"
		output.MissingStageIndex, _ = ProvenanceStageCoverageMissingStageIndex(output.MissingStage)
	case ProvenanceStageCoverageUnknown:
		output.Publishable = true
		if output.MissingStage == "" {
			output.MissingStage = output.TargetStage
		}
		output.MissingStageIndex, _ = ProvenanceStageCoverageMissingStageIndex(output.TargetStage)
		if output.MissingStageIndex < 0 {
			output.Publishable = false
			output.Code = "diagnostic-location"
		}
	default:
		output.Status = ProvenanceStageCoverageLSPError
		output.Code = provenanceStageCoverageLSPIntegrity
		output.MissingStage = "observation-status"
		output.MissingStageIndex = -1
	}
	output.ProjectionDigest = provenanceStageCoverageLSPDigest(output)
	return output
}

func (projection ProvenanceStageCoverageLSPProjection) Validate() error {
	switch projection.Status {
	case ProvenanceStageCoverageLSPBound,
		ProvenanceStageCoverageLSPDeferred,
		ProvenanceStageCoverageLSPUnknown,
		ProvenanceStageCoverageLSPError:
	default:
		return fmt.Errorf("invalid provenance stage coverage LSP status %q", projection.Status)
	}
	if projection.MissingStageIndex < -1 {
		return fmt.Errorf("invalid provenance stage coverage missing stage index %d", projection.MissingStageIndex)
	}
	if !projection.ReadOnly || !projection.NonAuthorizing || projection.CanExecute || projection.CanAuthorize {
		return fmt.Errorf("provenance stage coverage LSP crossed a forbidden boundary")
	}
	switch projection.Status {
	case ProvenanceStageCoverageLSPBound:
		if projection.Publishable || projection.MissingStage != "" || projection.MissingStageIndex != -1 ||
			projection.Code != provenanceStageCoverageLSPComplete ||
			projection.CoverageDenominator <= 0 ||
			projection.CoverageNumerator != projection.CoverageDenominator ||
			projection.CoveragePercent != 100 || projection.ObservationDigest == "" {
			return fmt.Errorf("bound provenance stage coverage LSP is incomplete")
		}
	case ProvenanceStageCoverageLSPDeferred:
		if !projection.Publishable || projection.MissingStage != "provenance_stage_producer" ||
			projection.MissingStageIndex != 4 || projection.Code != provenanceStageCoverageLSPDeferredCode {
			return fmt.Errorf("deferred provenance stage coverage LSP is incomplete")
		}
	case ProvenanceStageCoverageLSPUnknown:
		if projection.MissingStage == "" || projection.MissingStageIndex < -1 ||
			(projection.Code != provenanceStageCoverageLSPUnknownCode && projection.Code != "diagnostic-location") {
			return fmt.Errorf("unknown provenance stage coverage LSP must preserve its unresolved stage")
		}
	case ProvenanceStageCoverageLSPError:
		if projection.MissingStage == "" ||
			(projection.Code != provenanceStageCoverageLSPIntegrity && projection.Code != provenanceStageCoverageLSPBoundary) {
			return fmt.Errorf("error provenance stage coverage LSP must preserve its failure stage")
		}
	}
	if projection.ProjectionDigest != provenanceStageCoverageLSPDigest(projection) {
		return fmt.Errorf("provenance stage coverage LSP projection digest mismatch")
	}
	return nil
}

func provenanceStageCoverageLSPDigest(projection ProvenanceStageCoverageLSPProjection) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		projection.Status,
		fmt.Sprint(projection.Publishable),
		projection.Severity,
		projection.Code,
		fmt.Sprint(projection.CoverageNumerator),
		fmt.Sprint(projection.CoverageDenominator),
		fmt.Sprint(projection.CoveragePercent),
		projection.MissingStage,
		fmt.Sprint(projection.MissingStageIndex),
		projection.TargetStage,
		projection.ObservationDigest,
		projection.EvidencePrefixDigest,
		fmt.Sprint(projection.ReadOnly),
		fmt.Sprint(projection.NonAuthorizing),
		fmt.Sprint(projection.CanExecute),
		fmt.Sprint(projection.CanAuthorize),
	}, "|")))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}