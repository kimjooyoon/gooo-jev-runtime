package proposal

import (
	"sort"
	"strings"
)

type Status string

const (
	StatusProposed  Status = "PROPOSED"
	StatusNeedsInput Status = "NEEDS_INPUT"
	StatusUnknown   Status = "UNKNOWN"
	StatusRejected  Status = "REJECTED"
)

const (
	SignalCollectMore        = "COLLECT_MORE_OBSERVATIONS"
	SignalReviewMissing      = "REVIEW_MISSING_CAPABILITIES"
	SignalReviewInvestment   = "REVIEW_INVESTMENT"
	SignalNoActionableSignal = "NO_ACTIONABLE_SIGNAL"
)

// Input is a measurement result already bound to its source and contract identity.
type Input struct {
	DomainID             string
	MissingCapabilityIDs []string
	ObservedSignals      []string
	EvidenceDigests      []string
	SourceDigest         string
	ContractDigest       string
}

// Result is a review proposal, never an execution plan or permission grant.
type Result struct {
	Status               Status
	DomainID             string
	MissingCapabilityIDs []string
	ObservedSignals      []string
	Proposal             string
	NextQuestion         string
	EvidenceDigests      []string
	SourceDigest         string
	ContractDigest       string
}

// Build converts bounded measurement signals into a review-only proposal.
func Build(input Input) Result {
	result := Result{
		Status:               StatusUnknown,
		DomainID:             strings.TrimSpace(input.DomainID),
		MissingCapabilityIDs: normalize(input.MissingCapabilityIDs),
		ObservedSignals:      normalize(input.ObservedSignals),
		EvidenceDigests:      normalize(input.EvidenceDigests),
		SourceDigest:         strings.TrimSpace(input.SourceDigest),
		ContractDigest:       strings.TrimSpace(input.ContractDigest),
	}
	if result.DomainID == "" || result.SourceDigest == "" || result.ContractDigest == "" || len(result.EvidenceDigests) == 0 {
		result.NextQuestion = "Provide the domain, source and contract digests, and at least one observation digest."
		return result
	}
	if len(result.ObservedSignals) == 0 {
		result.Status = StatusNeedsInput
		result.NextQuestion = "Which bounded measurement signal should guide the review?"
		return result
	}
	for _, signal := range result.ObservedSignals {
		switch signal {
		case SignalCollectMore:
			result.Status = StatusProposed
			result.Proposal = "Collect another bounded observation before changing implementation scope."
			return result
		case SignalReviewMissing:
			if len(result.MissingCapabilityIDs) == 0 {
				result.Status = StatusNeedsInput
				result.NextQuestion = "Which expected capability remains missing from the declared domain boundary?"
				return result
			}
			result.Status = StatusProposed
			result.Proposal = "Review the explicitly missing capabilities before selecting implementation work."
			result.NextQuestion = "Which missing capability should be analyzed first?"
			return result
		case SignalReviewInvestment:
			result.Status = StatusProposed
			result.Proposal = "Review the measured domain boundary and decide whether additional investment is justified."
			return result
		case SignalNoActionableSignal:
			result.Status = StatusNeedsInput
			result.NextQuestion = "What new observation would make the domain decision actionable?"
			return result
		default:
			result.Status = StatusRejected
			result.NextQuestion = "Provide a signal from the declared measurement contract."
			return result
		}
	}
	return result
}

func normalize(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}