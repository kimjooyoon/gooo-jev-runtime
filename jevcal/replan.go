package jevcal

// ReplanStatus is the state of a generated next-step proposal. It is not an
// execution result.
type ReplanStatus string

const (
	ReplanProposed  ReplanStatus = "PROPOSED"
	ReplanNoAction  ReplanStatus = "NO_ACTION"
	ReplanUnknown   ReplanStatus = "UNKNOWN"
	ReplanDeferred  ReplanStatus = "DEFERRED"
)

// ReplanCandidate is a read-only proposal derived from two closure receipts.
type ReplanCandidate struct {
	Status                    ReplanStatus
	SourceVersion             string
	ContractVersion           string
	PreviousStatus            ClosureStatus
	CurrentStatus             ClosureStatus
	TargetStage               string
	MissingStageIndex         int
	PreviousObservationDigest string
	CurrentObservationDigest  string
	Reason                    string
	PlanDigest                string
	IsReadOnly                bool
	CanExecute                bool
	CanAuthorize              bool
}

// ProposeReplan creates a bounded proposal only when closure evidence regresses
// or remains unresolved. It never produces an executable command or authority.
func ProposeReplan(previous, current ClosureEvidence) ReplanCandidate {
	previousObservation := ObserveClosure(previous)
	currentObservation := ObserveClosure(current)
	candidate := ReplanCandidate{
		Status:                    ReplanUnknown,
		SourceVersion:             current.SourceVersion,
		ContractVersion:           current.ContractVersion,
		PreviousStatus:            previousObservation.Status,
		CurrentStatus:              currentObservation.Status,
		MissingStageIndex:         currentObservation.MissingStageIndex,
		PreviousObservationDigest: previousObservation.ObservationDigest,
		CurrentObservationDigest:  currentObservation.ObservationDigest,
		IsReadOnly:                true,
		CanExecute:                false,
		CanAuthorize:              false,
	}

	if previous.SourceVersion != current.SourceVersion || previous.ContractVersion != current.ContractVersion {
		candidate.Reason = "source or contract identity changed"
		return finalizeReplanCandidate(candidate)
	}
	if candidate.PreviousObservationDigest == "" || candidate.CurrentObservationDigest == "" {
		candidate.Reason = "replan basis evidence is missing"
		return finalizeReplanCandidate(candidate)
	}
	if previousObservation.Status == ClosureDeferred || currentObservation.Status == ClosureDeferred {
		candidate.Status = ReplanDeferred
		candidate.Reason = "closure producer state is deferred"
		return finalizeReplanCandidate(candidate)
	}
	if currentObservation.Status == ClosureBound && previousObservation.Status != ClosureBound {
		candidate.Status = ReplanNoAction
		candidate.Reason = "closure became BOUND"
		return finalizeReplanCandidate(candidate)
	}
	if previousObservation.Status == ClosureBound && currentObservation.Status == ClosureBound && previousObservation.ObservationDigest == currentObservation.ObservationDigest {
		candidate.Status = ReplanNoAction
		candidate.Reason = "BOUND closure observation is unchanged"
		return finalizeReplanCandidate(candidate)
	}
	if currentObservation.Status != ClosureBound {
		candidate.Status = ReplanProposed
		candidate.TargetStage = currentObservation.MissingStage
		if candidate.TargetStage == "" {
			candidate.TargetStage = "closure_digest"
		}
		candidate.Reason = "generate a bounded proposal for unresolved closure evidence"
		return finalizeReplanCandidate(candidate)
	}

	candidate.Reason = "bound observation changed without a comparable quality score"
	return finalizeReplanCandidate(candidate)
}

func finalizeReplanCandidate(candidate ReplanCandidate) ReplanCandidate {
	candidate.PlanDigest = digest(
		string(candidate.Status),
		candidate.SourceVersion,
		candidate.ContractVersion,
		string(candidate.PreviousStatus),
		string(candidate.CurrentStatus),
		candidate.TargetStage,
		candidate.PreviousObservationDigest,
		candidate.CurrentObservationDigest,
		candidate.Reason,
	)
	return candidate
}