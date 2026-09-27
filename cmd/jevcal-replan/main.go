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
	AcceptedCount         int     `json:"accepted_count"`
	RejectedCount         int     `json:"rejected_count"`
	EvidenceCoverage      float64 `json:"evidence_coverage"`
	MinimumWindow         int     `json:"minimum_window"`
	FirstMismatch         string  `json:"first_mismatch"`
	MissingStage          string  `json:"missing_stage"`
	EvidencePrefixDigest  string  `json:"evidence_prefix_digest"`
	EvidenceDigest        string  `json:"evidence_digest"`
	IsReadOnly            bool    `json:"is_read_only"`
	CanExecute            bool    `json:"can_execute"`
	CanAuthorize          bool    `json:"can_authorize"`
}

func (input windowJSON) domain() jevcal.CapabilityDiscoveryFeedbackWindowObservation {
	return jevcal.CapabilityDiscoveryFeedbackWindowObservation{
		Status:                jevcal.CapabilityDiscoveryFeedbackWindowStatus(input.Status),
		ObservationCount:      input.ObservationCount,
		KnownObservationCount: input.KnownObservationCount,
		AcceptedCount:         input.AcceptedCount,
		RejectedCount:         input.RejectedCount,
		EvidenceCoverage:      input.EvidenceCoverage,
		MinimumWindow:         input.MinimumWindow,
		FirstMismatch:         input.FirstMismatch,
		MissingStage:          input.MissingStage,
		EvidencePrefixDigest:  input.EvidencePrefixDigest,
		EvidenceDigest:        input.EvidenceDigest,
		IsReadOnly:            input.IsReadOnly,
		CanExecute:            input.CanExecute,
		CanAuthorize:          input.CanAuthorize,
	}
}

type request struct {
	Window windowJSON `json:"window"`
}

type replanJSON struct {
	Status           string `json:"status"`
	WindowDigest     string `json:"window_digest"`
	ObservationCount int    `json:"observation_count"`
	KnownCount       int    `json:"known_count"`
	AcceptedCount    int    `json:"accepted_count"`
	RejectedCount    int    `json:"rejected_count"`
	Disposition      string `json:"disposition"`
	NextOperation    string `json:"next_operation"`
	FirstMismatch    string `json:"first_mismatch"`
	MissingStage     string `json:"missing_stage"`
	EvidenceDigest   string `json:"evidence_digest"`
	IsReadOnly       bool   `json:"is_read_only"`
	CanExecute       bool   `json:"can_execute"`
	CanAuthorize     bool   `json:"can_authorize"`
}

func render(observation jevcal.CapabilityDiscoveryReplanObservation) replanJSON {
	return replanJSON{
		Status:           string(observation.Status),
		WindowDigest:     observation.WindowDigest,
		ObservationCount: observation.ObservationCount,
		KnownCount:       observation.KnownCount,
		AcceptedCount:    observation.AcceptedCount,
		RejectedCount:    observation.RejectedCount,
		Disposition:      string(observation.Disposition),
		NextOperation:    observation.NextOperation,
		FirstMismatch:    observation.FirstMismatch,
		MissingStage:     observation.MissingStage,
		EvidenceDigest:   observation.EvidenceDigest,
		IsReadOnly:       observation.IsReadOnly,
		CanExecute:       observation.CanExecute,
		CanAuthorize:     observation.CanAuthorize,
	}
}

func main() {
	var input request
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "read capability replan request: %v\n", err)
		os.Exit(64)
	}
	observation := jevcal.ProposeCapabilityDiscoveryReplan(input.Window.domain())
	if err := observation.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "invalid capability replan observation: %v\n", err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(struct {
		Replan replanJSON `json:"replan"`
	}{Replan: render(observation)}); err != nil {
		fmt.Fprintf(os.Stderr, "write capability replan observation: %v\n", err)
		os.Exit(1)
	}
}
