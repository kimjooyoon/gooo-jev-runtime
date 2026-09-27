package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type CapabilityDiscoveryReplanStatus string

type CapabilityDiscoveryReplanDisposition string

const (
	CapabilityDiscoveryReplanBound  CapabilityDiscoveryReplanStatus = "BOUND"
	CapabilityDiscoveryReplanUnknown CapabilityDiscoveryReplanStatus = "UNKNOWN"

	CapabilityDiscoveryReplanNextIteration CapabilityDiscoveryReplanDisposition = "PROPOSE_NEXT_ITERATION"
	CapabilityDiscoveryReplanHoldEvidence   CapabilityDiscoveryReplanDisposition = "HOLD_FOR_EVIDENCE"
	CapabilityDiscoveryReplanReviewOutcome  CapabilityDiscoveryReplanDisposition = "REVIEW_NO_ACCEPTED_OUTCOME"
)

// CapabilityDiscoveryReplanObservation is a read-only proposal derived from
// a feedback window. It never applies a change or grants authority.
type CapabilityDiscoveryReplanObservation struct {
	Status              CapabilityDiscoveryReplanStatus
	WindowDigest        string
	ObservationCount    int
	KnownCount          int
	AcceptedCount       int
	RejectedCount       int
	Disposition         CapabilityDiscoveryReplanDisposition
	NextOperation       string
	FirstMismatch       string
	MissingStage        string
	EvidenceDigest      string
	IsReadOnly          bool
	CanExecute          bool
	CanAuthorize        bool
}

// ProposeCapabilityDiscoveryReplan produces a bounded next-iteration proposal.
// It retains UNKNOWN when the feedback window is incomplete or invalid.
func ProposeCapabilityDiscoveryReplan(window CapabilityDiscoveryFeedbackWindowObservation) CapabilityDiscoveryReplanObservation {
	observation := CapabilityDiscoveryReplanObservation{
		Status:           CapabilityDiscoveryReplanUnknown,
		WindowDigest:     window.EvidenceDigest,
		ObservationCount: window.ObservationCount,
		KnownCount:       window.KnownObservationCount,
		AcceptedCount:    window.AcceptedCount,
		RejectedCount:    window.RejectedCount,
		Disposition:      CapabilityDiscoveryReplanHoldEvidence,
		NextOperation:    "complete_missing_feedback",
		FirstMismatch:    "window",
		MissingStage:     "capability-feedback-window",
		IsReadOnly:       true,
		CanExecute:       false,
		CanAuthorize:     false,
	}
	if err := window.Validate(); err != nil {
		observation.FirstMismatch = "window-integrity"
		observation.MissingStage = "capability-feedback-window"
		return finalizeCapabilityDiscoveryReplan(observation)
	}
	if window.Status != CapabilityDiscoveryFeedbackWindowBound {
		observation.FirstMismatch = window.FirstMismatch
		if observation.FirstMismatch == "" {
			observation.FirstMismatch = "window"
		}
		observation.MissingStage = window.MissingStage
		if observation.MissingStage == "" {
			observation.MissingStage = "capability-feedback-window"
		}
		return finalizeCapabilityDiscoveryReplan(observation)
	}

	observation.Status = CapabilityDiscoveryReplanBound
	observation.FirstMismatch = ""
	observation.MissingStage = ""
	if window.AcceptedCount > 0 {
		observation.Disposition = CapabilityDiscoveryReplanNextIteration
		observation.NextOperation = "select_next_capability_iteration"
	} else {
		observation.Disposition = CapabilityDiscoveryReplanReviewOutcome
		observation.NextOperation = "review_capability_use"
	}
	return finalizeCapabilityDiscoveryReplan(observation)
}

func finalizeCapabilityDiscoveryReplan(observation CapabilityDiscoveryReplanObservation) CapabilityDiscoveryReplanObservation {
	observation.EvidenceDigest = capabilityDiscoveryReplanDigest(observation)
	return observation
}

func (observation CapabilityDiscoveryReplanObservation) Validate() error {
	if observation.Status != CapabilityDiscoveryReplanBound && observation.Status != CapabilityDiscoveryReplanUnknown {
		return fmt.Errorf("invalid capability discovery replan status %q", observation.Status)
	}
	if observation.Disposition != CapabilityDiscoveryReplanNextIteration &&
		observation.Disposition != CapabilityDiscoveryReplanHoldEvidence &&
		observation.Disposition != CapabilityDiscoveryReplanReviewOutcome {
		return fmt.Errorf("invalid capability discovery replan disposition %q", observation.Disposition)
	}
	if !observation.IsReadOnly || observation.CanExecute || observation.CanAuthorize {
		return fmt.Errorf("capability discovery replan crossed an execution or authorization boundary")
	}
	if !capabilityFeedbackWindowDigestValid(observation.EvidenceDigest) {
		return fmt.Errorf("capability discovery replan evidence digest is invalid")
	}
	if observation.ObservationCount < 0 || observation.KnownCount < 0 ||
		observation.KnownCount > observation.ObservationCount || observation.AcceptedCount < 0 ||
		observation.RejectedCount < 0 || observation.AcceptedCount+observation.RejectedCount != observation.KnownCount {
		return fmt.Errorf("capability discovery replan counts are invalid")
	}
	if observation.Status == CapabilityDiscoveryReplanUnknown {
		if observation.Disposition != CapabilityDiscoveryReplanHoldEvidence ||
			observation.FirstMismatch == "" || observation.MissingStage == "" || observation.NextOperation == "" {
			return fmt.Errorf("unknown capability discovery replan lost its boundary")
		}
		return nil
	}
	if !capabilityFeedbackWindowDigestValid(observation.WindowDigest) {
		return fmt.Errorf("bound capability discovery replan window digest is invalid")
	}
	if observation.FirstMismatch != "" || observation.MissingStage != "" || observation.NextOperation == "" {
		return fmt.Errorf("bound capability discovery replan is incomplete")
	}
	if observation.AcceptedCount > 0 && observation.Disposition != CapabilityDiscoveryReplanNextIteration {
		return fmt.Errorf("accepted feedback must propose the next iteration")
	}
	if observation.AcceptedCount == 0 && observation.Disposition != CapabilityDiscoveryReplanReviewOutcome {
		return fmt.Errorf("feedback without accepted outcomes must remain review-only")
	}
	if observation.EvidenceDigest != capabilityDiscoveryReplanDigest(observation) {
		return fmt.Errorf("capability discovery replan digest mismatch")
	}
	return nil
}

func capabilityDiscoveryReplanDigest(observation CapabilityDiscoveryReplanObservation) string {
	payload := strings.Join([]string{
		"jev-capability-discovery-replan",
		string(observation.Status),
		observation.WindowDigest,
		fmt.Sprintf("%d", observation.ObservationCount),
		fmt.Sprintf("%d", observation.KnownCount),
		fmt.Sprintf("%d", observation.AcceptedCount),
		fmt.Sprintf("%d", observation.RejectedCount),
		string(observation.Disposition),
		observation.NextOperation,
		observation.FirstMismatch,
		observation.MissingStage,
		fmt.Sprintf("%t", observation.IsReadOnly),
		fmt.Sprintf("%t", observation.CanExecute),
		fmt.Sprintf("%t", observation.CanAuthorize),
	}, "|")
	digest := sha256.Sum256([]byte(payload))
	return "sha256:" + hex.EncodeToString(digest[:])
}
