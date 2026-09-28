package jevcal

import (
	"fmt"
	"strings"
)

// CapabilityDiscoveryFollowUp keeps natural discovery conversational without
// crossing into execution, authorization, or mutation.
type CapabilityDiscoveryFollowUp struct {
	Status         CapabilityDiscoveryStatus
	Questions      []string
	FirstBoundary  string
	IsReadOnly     bool
}

// SuggestCapabilityDiscoveryFollowUps converts an observation boundary into
// deterministic, read-only questions. It never selects a capability or grants
// permission on behalf of the caller.
func SuggestCapabilityDiscoveryFollowUps(observation CapabilityDiscoveryObservation) CapabilityDiscoveryFollowUp {
	followUp := CapabilityDiscoveryFollowUp{
		Status:        observation.Status,
		FirstBoundary: observation.MissingStage,
		IsReadOnly:    true,
	}

	switch observation.Status {
	case CapabilityDiscoveryBound:
		if observation.DeclarationBound {
			followUp.Questions = append(followUp.Questions, "Which declared operation or observation should be inspected next?")
		} else {
			followUp.Questions = append(followUp.Questions, "Which discovered capability should be inspected next?")
		}
		if len(observation.Matches) > 1 {
			followUp.Questions = append(followUp.Questions, "Which one capability should be selected for read-only inspection?")
		}
	case CapabilityDiscoveryDeferred:
		followUp.Questions = append(followUp.Questions, "Which explicit external boundary is available before any execution or authorization?")
		if followUp.FirstBoundary == "" {
			followUp.FirstBoundary = "explicit_external_boundary"
		}
	case CapabilityDiscoveryUnknown:
		switch {
		case strings.TrimSpace(observation.SourceVersion) == "" || strings.TrimSpace(observation.ContractVersion) == "":
			followUp.Questions = append(followUp.Questions, "Which source and contract identities should bind this discovery request?")
		case strings.TrimSpace(observation.Query) == "":
			followUp.Questions = append(followUp.Questions, "What would you like to know about the declared capabilities?")
		default:
			followUp.Questions = append(followUp.Questions, "Which declaration, catalog term, or example query should clarify this capability request?")
		}
		if followUp.FirstBoundary == "" {
			followUp.FirstBoundary = "capability_discovery"
		}
	default:
		followUp.Questions = append(followUp.Questions, "Which read-only capability discovery boundary should be clarified?")
		if followUp.FirstBoundary == "" {
			followUp.FirstBoundary = "capability_discovery"
		}
	}
	return followUp
}

func (followUp CapabilityDiscoveryFollowUp) Validate() error {
	if followUp.Status != CapabilityDiscoveryBound && followUp.Status != CapabilityDiscoveryDeferred && followUp.Status != CapabilityDiscoveryUnknown {
		return fmt.Errorf("invalid capability discovery follow-up status %q", followUp.Status)
	}
	if !followUp.IsReadOnly {
		return fmt.Errorf("capability discovery follow-up is not read-only")
	}
	if strings.TrimSpace(followUp.FirstBoundary) == "" {
		return fmt.Errorf("capability discovery follow-up boundary is missing")
	}
	if len(followUp.Questions) == 0 {
		return fmt.Errorf("capability discovery follow-up has no question")
	}
	for _, question := range followUp.Questions {
		if strings.TrimSpace(question) == "" {
			return fmt.Errorf("capability discovery follow-up contains an empty question")
		}
	}
	return nil
}
