package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/gooo-jev-runtime/jevcal"
)

type windowJSON struct {
	Status                string  `json:"status"`
	ObservationCount      int     `json:"observation_count"`
	KnownObservationCount int     `json:"known_observation_count"`
	WithinToleranceCount  int     `json:"within_tolerance_count"`
	MeanAbsoluteError     float64 `json:"mean_absolute_error"`
	EvidenceCoverage      float64 `json:"evidence_coverage"`
	MinimumWindow         int     `json:"minimum_window"`
	FirstMismatch         string  `json:"first_mismatch"`
	MissingStage           string  `json:"missing_stage"`
	EvidencePrefixDigest  string  `json:"evidence_prefix_digest"`
	EvidenceDigest        string  `json:"evidence_digest"`
	IsReadOnly            bool    `json:"is_read_only"`
	CanExecute            bool    `json:"can_execute"`
	CanAuthorize          bool    `json:"can_authorize"`
}

func (input windowJSON) domain() jevcal.TypedDecisionCalibrationWindowObservation {
	return jevcal.TypedDecisionCalibrationWindowObservation{
		Status:                 input.Status,
		ObservationCount:       input.ObservationCount,
		KnownObservationCount:  input.KnownObservationCount,
		WithinToleranceCount:   input.WithinToleranceCount,
		MeanAbsoluteError:      input.MeanAbsoluteError,
		EvidenceCoverage:       input.EvidenceCoverage,
		MinimumWindow:          input.MinimumWindow,
		FirstMismatch:          input.FirstMismatch,
		MissingStage:           input.MissingStage,
		EvidencePrefixDigest:   input.EvidencePrefixDigest,
		EvidenceDigest:         input.EvidenceDigest,
		IsReadOnly:             input.IsReadOnly,
		CanExecute:             input.CanExecute,
		CanAuthorize:           input.CanAuthorize,
	}
}

type request struct {
	Previous   windowJSON `json:"previous"`
	Current    windowJSON `json:"current"`
	ComparedAt string     `json:"compared_at"`
}

type trendJSON struct {
	Status                 string  `json:"status"`
	PreviousEvidenceDigest string  `json:"previous_evidence_digest"`
	CurrentEvidenceDigest  string  `json:"current_evidence_digest"`
	PreviousMeanError      float64 `json:"previous_mean_error"`
	CurrentMeanError       float64 `json:"current_mean_error"`
	MeanErrorDelta         float64 `json:"mean_error_delta"`
	PreviousCoverage       float64 `json:"previous_coverage"`
	CurrentCoverage        float64 `json:"current_coverage"`
	CoverageDelta          float64 `json:"coverage_delta"`
	PreviousUnknownCount   int     `json:"previous_unknown_count"`
	CurrentUnknownCount    int     `json:"current_unknown_count"`
	Direction              string  `json:"direction"`
	ComparedAt             string  `json:"compared_at"`
	FirstMismatch          string  `json:"first_mismatch"`
	MissingStage            string  `json:"missing_stage"`
	EvidenceDigest         string  `json:"evidence_digest"`
	IsReadOnly             bool    `json:"is_read_only"`
	CanExecute             bool    `json:"can_execute"`
	CanAuthorize           bool    `json:"can_authorize"`
}

func renderTrend(observation jevcal.TypedDecisionCalibrationWindowTrendObservation) trendJSON {
	return trendJSON{
		Status:                 observation.Status,
		PreviousEvidenceDigest: observation.PreviousEvidenceDigest,
		CurrentEvidenceDigest:  observation.CurrentEvidenceDigest,
		PreviousMeanError:      observation.PreviousMeanError,
		CurrentMeanError:       observation.CurrentMeanError,
		MeanErrorDelta:         observation.MeanErrorDelta,
		PreviousCoverage:       observation.PreviousCoverage,
		CurrentCoverage:        observation.CurrentCoverage,
		CoverageDelta:          observation.CoverageDelta,
		PreviousUnknownCount:   observation.PreviousUnknownCount,
		CurrentUnknownCount:    observation.CurrentUnknownCount,
		Direction:              observation.Direction,
		ComparedAt:             observation.ComparedAt,
		FirstMismatch:          observation.FirstMismatch,
		MissingStage:           observation.MissingStage,
		EvidenceDigest:         observation.EvidenceDigest,
		IsReadOnly:             observation.IsReadOnly,
		CanExecute:             observation.CanExecute,
		CanAuthorize:           observation.CanAuthorize,
	}
}

func main() {
	var input request
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "read trend request: %v\n", err)
		os.Exit(64)
	}
	observation := jevcal.ObserveTypedDecisionCalibrationWindowTrend(
		jevcal.TypedDecisionCalibrationWindowTrendInput{
			Previous:   input.Previous.domain(),
			Current:    input.Current.domain(),
			ComparedAt: input.ComparedAt,
		},
	)
	if err := observation.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "invalid trend observation: %v\n", err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(struct {
		Trend trendJSON `json:"trend"`
	}{Trend: renderTrend(observation)}); err != nil {
		fmt.Fprintf(os.Stderr, "write trend observation: %v\n", err)
		os.Exit(1)
	}
}
