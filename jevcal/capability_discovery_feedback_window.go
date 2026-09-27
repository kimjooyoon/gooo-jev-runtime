package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
)

type CapabilityDiscoveryFeedbackWindowStatus string

const (
	CapabilityDiscoveryFeedbackWindowBound   CapabilityDiscoveryFeedbackWindowStatus = "BOUND"
	CapabilityDiscoveryFeedbackWindowUnknown CapabilityDiscoveryFeedbackWindowStatus = "UNKNOWN"
)

// CapabilityDiscoveryFeedbackWindowInput aggregates already recorded feedback
// observations without executing or re-evaluating any capability.
type CapabilityDiscoveryFeedbackWindowInput struct {
	Observations  []CapabilityDiscoveryFeedbackObservation
	MinimumWindow int
}

// CapabilityDiscoveryFeedbackWindowObservation summarizes feedback evidence.
// Counts measure observed outcomes, not language quality or improvement.
type CapabilityDiscoveryFeedbackWindowObservation struct {
	Status                CapabilityDiscoveryFeedbackWindowStatus
	ObservationCount      int
	KnownObservationCount int
	AcceptedCount         int
	RejectedCount         int
	EvidenceCoverage      float64
	MinimumWindow         int
	FirstMismatch         string
	MissingStage          string
	EvidencePrefixDigest  string
	EvidenceDigest        string
	IsReadOnly            bool
	CanExecute            bool
	CanAuthorize          bool
}

// ObserveCapabilityDiscoveryFeedbackWindow preserves the first unavailable
// feedback stage and only binds a complete, sufficiently large window.
func ObserveCapabilityDiscoveryFeedbackWindow(input CapabilityDiscoveryFeedbackWindowInput) CapabilityDiscoveryFeedbackWindowObservation {
	observation := CapabilityDiscoveryFeedbackWindowObservation{
		Status:               CapabilityDiscoveryFeedbackWindowUnknown,
		ObservationCount:     len(input.Observations),
		MinimumWindow:        input.MinimumWindow,
		FirstMismatch:        "observations",
		MissingStage:         "capability-feedback-window",
		EvidencePrefixDigest: capabilityFeedbackWindowPrefixDigest(nil),
		IsReadOnly:           true,
		CanExecute:           false,
		CanAuthorize:         false,
	}
	if len(input.Observations) == 0 {
		return finalizeCapabilityDiscoveryFeedbackWindow(observation)
	}
	if input.MinimumWindow <= 0 {
		observation.FirstMismatch = "minimum-window"
		return finalizeCapabilityDiscoveryFeedbackWindow(observation)
	}

	prefix := make([]string, 0, len(input.Observations))
	for index, item := range input.Observations {
		if err := item.Validate(); err != nil {
			observation.FirstMismatch = fmt.Sprintf("observation[%d]-integrity", index)
			observation.MissingStage = "capability_discovery_feedback"
			observation.EvidenceCoverage = feedbackWindowCoverage(observation)
			observation.EvidencePrefixDigest = capabilityFeedbackWindowPrefixDigest(prefix)
			return finalizeCapabilityDiscoveryFeedbackWindow(observation)
		}
		prefix = append(prefix, item.EvidenceDigest)
		if item.Status != CapabilityDiscoveryFeedbackBound {
			observation.FirstMismatch = fmt.Sprintf("observation[%d]", index)
			observation.MissingStage = item.MissingStage
			if observation.MissingStage == "" {
				observation.MissingStage = "capability_discovery_feedback"
			}
			observation.EvidenceCoverage = feedbackWindowCoverage(observation)
			observation.EvidencePrefixDigest = capabilityFeedbackWindowPrefixDigest(prefix)
			return finalizeCapabilityDiscoveryFeedbackWindow(observation)
		}
		observation.KnownObservationCount++
		if item.OutcomeAccepted {
			observation.AcceptedCount++
		} else {
			observation.RejectedCount++
		}
	}
	observation.EvidenceCoverage = feedbackWindowCoverage(observation)
	observation.EvidencePrefixDigest = capabilityFeedbackWindowPrefixDigest(prefix)
	if len(input.Observations) < input.MinimumWindow {
		observation.FirstMismatch = "minimum-window"
		return finalizeCapabilityDiscoveryFeedbackWindow(observation)
	}
	observation.Status = CapabilityDiscoveryFeedbackWindowBound
	observation.FirstMismatch = ""
	observation.MissingStage = ""
	return finalizeCapabilityDiscoveryFeedbackWindow(observation)
}

func feedbackWindowCoverage(observation CapabilityDiscoveryFeedbackWindowObservation) float64 {
	if observation.ObservationCount == 0 {
		return 0
	}
	return float64(observation.KnownObservationCount) / float64(observation.ObservationCount)
}

func finalizeCapabilityDiscoveryFeedbackWindow(observation CapabilityDiscoveryFeedbackWindowObservation) CapabilityDiscoveryFeedbackWindowObservation {
	observation.EvidenceDigest = capabilityDiscoveryFeedbackWindowDigest(observation)
	return observation
}

func (observation CapabilityDiscoveryFeedbackWindowObservation) Validate() error {
	if observation.Status != CapabilityDiscoveryFeedbackWindowBound && observation.Status != CapabilityDiscoveryFeedbackWindowUnknown {
		return fmt.Errorf("invalid capability discovery feedback window status %q", observation.Status)
	}
	if observation.ObservationCount < 0 || observation.KnownObservationCount < 0 ||
		observation.KnownObservationCount > observation.ObservationCount ||
		observation.AcceptedCount < 0 || observation.RejectedCount < 0 ||
		observation.AcceptedCount+observation.RejectedCount != observation.KnownObservationCount ||
		observation.MinimumWindow <= 0 || math.IsNaN(observation.EvidenceCoverage) ||
		math.IsInf(observation.EvidenceCoverage, 0) || observation.EvidenceCoverage < 0 ||
		observation.EvidenceCoverage > 1 {
		return fmt.Errorf("capability discovery feedback window arithmetic is invalid")
	}
	if !observation.IsReadOnly || observation.CanExecute || observation.CanAuthorize {
		return fmt.Errorf("capability discovery feedback window crossed an execution or authorization boundary")
	}
	if !capabilityFeedbackWindowDigestValid(observation.EvidencePrefixDigest) ||
		observation.EvidenceDigest != capabilityDiscoveryFeedbackWindowDigest(observation) {
		return fmt.Errorf("capability discovery feedback window digest is invalid")
	}
	if observation.Status == CapabilityDiscoveryFeedbackWindowBound {
		if observation.ObservationCount < observation.MinimumWindow ||
			observation.KnownObservationCount != observation.ObservationCount ||
			observation.EvidenceCoverage != 1 || observation.FirstMismatch != "" || observation.MissingStage != "" {
			return fmt.Errorf("bound capability discovery feedback window is incomplete")
		}
	}
	if observation.Status == CapabilityDiscoveryFeedbackWindowUnknown &&
		(observation.FirstMismatch == "" || observation.MissingStage == "") {
		return fmt.Errorf("unknown capability discovery feedback window lost its first mismatch")
	}
	return nil
}

func capabilityDiscoveryFeedbackWindowDigest(observation CapabilityDiscoveryFeedbackWindowObservation) string {
	payload := fmt.Sprintf(
		"jev-capability-discovery-feedback-window|%s|%d|%d|%d|%d|%.9f|%d|%s|%s|%s|%t|%t|%t",
		observation.Status,
		observation.ObservationCount,
		observation.KnownObservationCount,
		observation.AcceptedCount,
		observation.RejectedCount,
		observation.EvidenceCoverage,
		observation.MinimumWindow,
		observation.FirstMismatch,
		observation.MissingStage,
		observation.EvidencePrefixDigest,
		observation.IsReadOnly,
		observation.CanExecute,
		observation.CanAuthorize,
	)
	digest := sha256.Sum256([]byte(payload))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func capabilityFeedbackWindowPrefixDigest(evidenceDigests []string) string {
	payload := "jev-capability-discovery-feedback-window-prefix|" + strings.Join(evidenceDigests, "|")
	digest := sha256.Sum256([]byte(payload))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func capabilityFeedbackWindowDigestValid(value string) bool {
	const prefix = "sha256:"
	if !strings.HasPrefix(value, prefix) || len(value) != len(prefix)+64 {
		return false
	}
	_, err := hex.DecodeString(value[len(prefix):])
	return err == nil
}
