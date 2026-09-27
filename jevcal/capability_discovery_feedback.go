package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

type CapabilityDiscoveryFeedbackStatus string

const (
	CapabilityDiscoveryFeedbackBound   CapabilityDiscoveryFeedbackStatus = "BOUND"
	CapabilityDiscoveryFeedbackUnknown CapabilityDiscoveryFeedbackStatus = "UNKNOWN"
)

// CapabilityDiscoveryFeedbackInput records whether a discovery result became
// useful evidence. It never executes the suggested operation or grants authority.
type CapabilityDiscoveryFeedbackInput struct {
	Observation   CapabilityDiscoveryObservation
	OutcomeKnown  bool
	OutcomeAccepted bool
	OutcomeDigest string
}

// CapabilityDiscoveryFeedbackObservation binds discovery to a later outcome.
// Missing outcome evidence remains UNKNOWN rather than becoming improvement.
type CapabilityDiscoveryFeedbackObservation struct {
	Status              CapabilityDiscoveryFeedbackStatus
	SourceVersion       string
	ContractVersion     string
	Query               string
	DiscoveryDigest     string
	MatchedCapabilities []string
	OutcomeKnown        bool
	OutcomeAccepted     bool
	OutcomeDigest       string
	FirstMismatch       string
	MissingStage        string
	EvidenceDigest      string
	IsReadOnly          bool
	CanExecute          bool
	CanAuthorize        bool
}

// ObserveCapabilityDiscoveryFeedback records a bounded feedback observation.
// It does not claim that a capability is correct, executable, or authorized.
func ObserveCapabilityDiscoveryFeedback(input CapabilityDiscoveryFeedbackInput) CapabilityDiscoveryFeedbackObservation {
	observation := CapabilityDiscoveryFeedbackObservation{
		Status:              CapabilityDiscoveryFeedbackUnknown,
		SourceVersion:       input.Observation.SourceVersion,
		ContractVersion:     input.Observation.ContractVersion,
		Query:               input.Observation.Query,
		DiscoveryDigest:     input.Observation.DiscoveryDigest,
		MatchedCapabilities: capabilityFeedbackKeys(input.Observation.Matches),
		OutcomeKnown:        input.OutcomeKnown,
		OutcomeAccepted:     input.OutcomeAccepted,
		OutcomeDigest:       input.OutcomeDigest,
		FirstMismatch:       "discovery",
		MissingStage:        "capability_discovery",
		IsReadOnly:          true,
		CanExecute:          false,
		CanAuthorize:        false,
	}

	switch {
	case input.Observation.Validate() != nil:
		observation.FirstMismatch = "discovery-integrity"
		observation.MissingStage = "capability_discovery"
	case !input.Observation.IsReadOnly || input.Observation.CanExecute || input.Observation.CanAuthorize:
		observation.FirstMismatch = "capability-boundary"
		observation.MissingStage = "capability-boundary"
	case input.Observation.Status != CapabilityDiscoveryBound:
		observation.FirstMismatch = input.Observation.FirstMismatch
		if observation.FirstMismatch == "" {
			observation.FirstMismatch = "capability_discovery"
		}
		observation.MissingStage = input.Observation.MissingStage
		if observation.MissingStage == "" {
			observation.MissingStage = "capability_discovery"
		}
	case !input.OutcomeKnown:
		observation.FirstMismatch = "outcome"
		observation.MissingStage = "reverse_observation"
	case !capabilityFeedbackDigestValid(input.OutcomeDigest):
		observation.FirstMismatch = "outcome-digest"
		observation.MissingStage = "reverse_observation"
	default:
		observation.Status = CapabilityDiscoveryFeedbackBound
		observation.FirstMismatch = ""
		observation.MissingStage = ""
	}

	observation.EvidenceDigest = capabilityDiscoveryFeedbackDigest(observation)
	return observation
}

func capabilityFeedbackKeys(matches []CapabilityDiscoveryMatch) []string {
	keys := make([]string, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		if strings.TrimSpace(match.Key) == "" {
			continue
		}
		if _, ok := seen[match.Key]; ok {
			continue
		}
		seen[match.Key] = struct{}{}
		keys = append(keys, match.Key)
	}
	sort.Strings(keys)
	return keys
}

func (observation CapabilityDiscoveryFeedbackObservation) Validate() error {
	if observation.Status != CapabilityDiscoveryFeedbackBound && observation.Status != CapabilityDiscoveryFeedbackUnknown {
		return fmt.Errorf("invalid capability discovery feedback status %q", observation.Status)
	}
	if !observation.IsReadOnly || observation.CanExecute || observation.CanAuthorize {
		return fmt.Errorf("capability discovery feedback crossed an execution or authorization boundary")
	}
	if observation.EvidenceDigest != capabilityDiscoveryFeedbackDigest(observation) {
		return fmt.Errorf("capability discovery feedback digest mismatch")
	}
	seen := make(map[string]struct{}, len(observation.MatchedCapabilities))
	for _, key := range observation.MatchedCapabilities {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("capability discovery feedback contains an empty capability")
		}
		if _, ok := seen[key]; ok {
			return fmt.Errorf("capability discovery feedback capability %q is duplicated", key)
		}
		seen[key] = struct{}{}
	}
	if observation.Status == CapabilityDiscoveryFeedbackUnknown {
		if observation.FirstMismatch == "" || observation.MissingStage == "" {
			return fmt.Errorf("unknown capability discovery feedback lost its first boundary")
		}
		return nil
	}
	if strings.TrimSpace(observation.SourceVersion) == "" ||
		strings.TrimSpace(observation.ContractVersion) == "" ||
		strings.TrimSpace(observation.Query) == "" ||
		!capabilityFeedbackDigestValid(observation.DiscoveryDigest) ||
		!observation.OutcomeKnown ||
		!capabilityFeedbackDigestValid(observation.OutcomeDigest) ||
		observation.FirstMismatch != "" || observation.MissingStage != "" {
		return fmt.Errorf("bound capability discovery feedback is incomplete")
	}
	return nil
}

func capabilityFeedbackDigestValid(value string) bool {
	const prefix = "sha256:"
	if !strings.HasPrefix(value, prefix) || len(value) != len(prefix)+64 {
		return false
	}
	_, err := hex.DecodeString(value[len(prefix):])
	return err == nil
}

func capabilityDiscoveryFeedbackDigest(observation CapabilityDiscoveryFeedbackObservation) string {
	parts := []string{
		"jev-capability-discovery-feedback",
		string(observation.Status),
		observation.SourceVersion,
		observation.ContractVersion,
		observation.Query,
		observation.DiscoveryDigest,
		fmt.Sprintf("%t", observation.OutcomeKnown),
		fmt.Sprintf("%t", observation.OutcomeAccepted),
		observation.OutcomeDigest,
		observation.FirstMismatch,
		observation.MissingStage,
		fmt.Sprintf("%t", observation.IsReadOnly),
		fmt.Sprintf("%t", observation.CanExecute),
		fmt.Sprintf("%t", observation.CanAuthorize),
	}
	parts = append(parts, observation.MatchedCapabilities...)
	digest := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return "sha256:" + hex.EncodeToString(digest[:])
}
