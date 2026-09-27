package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type ProvenanceStageCoverageStatus string

const (
	ProvenanceStageCoverageBound    ProvenanceStageCoverageStatus = "BOUND"
	ProvenanceStageCoverageDeferred ProvenanceStageCoverageStatus = "DEFERRED"
	ProvenanceStageCoverageUnknown  ProvenanceStageCoverageStatus = "UNKNOWN"
)

// ProvenanceStageCoverageInput records the evidence stages available for
// observation. It never asserts that a change improved the language.
type ProvenanceStageCoverageInput struct {
	SourceVersion            string
	ContractVersion          string
	ExpectedStages           []string
	ObservedStages           []string
	SourceDigest             string
	IRDigest                 string
	GeneratedDigest          string
	ReverseObservationDigest string
	ProducerDeferred         bool
}

// ProvenanceStageCoverage is an evidence coverage metric, not an improvement
// metric. It does not execute, authorize, or classify a change as beneficial.
type ProvenanceStageCoverage struct {
	Status                    ProvenanceStageCoverageStatus
	SourceVersion             string
	ContractVersion           string
	ExpectedStages            []string
	ObservedStages            []string
	CoverageNumerator         int
	CoverageDenominator       int
	CoveragePercent            int
	MissingStage              string
	TargetStage               string
	Reason                    string
	ObservationDigest         string
	ObservationalOnly         bool
	ClaimsImprovement         bool
	CanExecute                bool
	CanAuthorize              bool
}

// ObserveProvenanceStageCoverage counts only expected stages with an observed
// stage and its corresponding digest. Missing evidence remains UNKNOWN.
func ObserveProvenanceStageCoverage(input ProvenanceStageCoverageInput) ProvenanceStageCoverage {
	observation := ProvenanceStageCoverage{
		Status:                    ProvenanceStageCoverageUnknown,
		SourceVersion:             input.SourceVersion,
		ContractVersion:           input.ContractVersion,
		ExpectedStages:            append([]string(nil), input.ExpectedStages...),
		ObservedStages:            append([]string(nil), input.ObservedStages...),
		CoverageDenominator:       uniqueExpectedStageCount(input.ExpectedStages),
		TargetStage:               "provenance_stage_coverage",
		ObservationalOnly:         true,
		ClaimsImprovement:         false,
		CanExecute:                false,
		CanAuthorize:              false,
	}

	observedSet := stageSet(input.ObservedStages)
	for _, stage := range uniqueStages(input.ExpectedStages) {
		if _, ok := observedSet[stage]; ok && stageDigest(input, stage) != "" {
			observation.CoverageNumerator++
		}
	}
	if observation.CoverageDenominator > 0 {
		observation.CoveragePercent = observation.CoverageNumerator * 100 / observation.CoverageDenominator
	}

	switch {
	case input.SourceVersion == "" || input.ContractVersion == "":
		observation.TargetStage = "identity"
		observation.Reason = "source or contract identity is missing"
	case input.ProducerDeferred:
		observation.Status = ProvenanceStageCoverageDeferred
		observation.Reason = "provenance stage producer is deferred"
	case observation.CoverageDenominator == 0:
		observation.TargetStage = "expected_stages"
		observation.Reason = "expected provenance stages are missing"
	default:
		for _, stage := range uniqueStages(input.ExpectedStages) {
			if _, ok := observedSet[stage]; !ok {
				observation.TargetStage = "observed_stage"
				observation.MissingStage = stage
				observation.Reason = "expected provenance stage was not observed"
				break
			}
			if stageDigest(input, stage) == "" {
				observation.TargetStage = "stage_digest"
				observation.MissingStage = stage
				observation.Reason = "observed provenance stage has no digest"
				break
			}
		}
		if observation.MissingStage == "" {
			observation.Status = ProvenanceStageCoverageBound
			observation.Reason = "all expected provenance stages are observed and digested"
		}
	}

	observation.ObservationDigest = provenanceStageCoverageDigest(
		string(observation.Status),
		observation.SourceVersion,
		observation.ContractVersion,
		strings.Join(observation.ExpectedStages, ","),
		strings.Join(observation.ObservedStages, ","),
		fmt.Sprintf("%d", observation.CoverageNumerator),
		fmt.Sprintf("%d", observation.CoverageDenominator),
		fmt.Sprintf("%d", observation.CoveragePercent),
		observation.MissingStage,
		observation.TargetStage,
		observation.Reason,
		input.SourceDigest,
		input.IRDigest,
		input.GeneratedDigest,
		input.ReverseObservationDigest,
	)
	return observation
}

func uniqueExpectedStageCount(stages []string) int {
	return len(uniqueStages(stages))
}

func uniqueStages(stages []string) []string {
	seen := make(map[string]struct{}, len(stages))
	unique := make([]string, 0, len(stages))
	for _, stage := range stages {
		if _, ok := seen[stage]; ok {
			continue
		}
		seen[stage] = struct{}{}
		unique = append(unique, stage)
	}
	return unique
}

func stageSet(stages []string) map[string]struct{} {
	set := make(map[string]struct{}, len(stages))
	for _, stage := range stages {
		set[stage] = struct{}{}
	}
	return set
}

func stageDigest(input ProvenanceStageCoverageInput, stage string) string {
	switch stage {
	case "source":
		return input.SourceDigest
	case "ir":
		return input.IRDigest
	case "generated":
		return input.GeneratedDigest
	case "reverse_observation":
		return input.ReverseObservationDigest
	default:
		return ""
	}
}

func provenanceStageCoverageDigest(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}
