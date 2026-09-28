package lifecycle

import "fmt"

type Stage string

const (
	Observed  Stage = "OBSERVED"
	Measured  Stage = "MEASURED"
	Proposed  Stage = "PROPOSED"
	Reviewed  Stage = "REVIEWED"
	Applied   Stage = "APPLIED"
	Reversed  Stage = "REVERSED"
	Unknown   Stage = "UNKNOWN"
)

type Evidence struct {
	ProposalDigest    string
	ReviewDigest      string
	ApplicationDigest string
	ReversalDigest    string
	ProvenanceDigest  string
}

type Transition struct {
	From     Stage
	To       Stage
	Evidence Evidence
}

// ValidateTransition validates an observed lifecycle record without performing a transition.
func ValidateTransition(transition Transition) error {
	if transition.Evidence.ProvenanceDigest == "" {
		return fmt.Errorf("provenance digest is required")
	}
	if transition.From == Unknown {
		return fmt.Errorf("unknown state cannot advance automatically")
	}
	switch transition.To {
	case Observed, Measured:
		return nil
	case Proposed:
		if transition.Evidence.ProposalDigest == "" {
			return fmt.Errorf("proposal digest is required")
		}
	case Reviewed:
		if transition.Evidence.ProposalDigest == "" || transition.Evidence.ReviewDigest == "" {
			return fmt.Errorf("proposal and review digests are required")
		}
	case Applied:
		if transition.Evidence.ProposalDigest == "" || transition.Evidence.ReviewDigest == "" || transition.Evidence.ApplicationDigest == "" {
			return fmt.Errorf("proposal, review, and application digests are required")
		}
	case Reversed:
		if transition.Evidence.ApplicationDigest == "" || transition.Evidence.ReversalDigest == "" {
			return fmt.Errorf("application and reversal digests are required")
		}
	case Unknown:
		return nil
	default:
		return fmt.Errorf("unknown target stage %q", transition.To)
	}
	return nil
}