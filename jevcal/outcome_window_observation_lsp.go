package jevcal

// OutcomeWindowObservationLSPProjection is an inspection-only view of
// outcome-window evidence. It never edits, executes, scores, or authorizes.
type OutcomeWindowObservationLSPProjection struct {
	Status               OutcomeWindowObservationStatus
	Code                 string
	Severity             string
	Title                string
	TargetStage          string
	RecordDigest         string
	Reason               string
	IsReadOnly           bool
	CanExecute           bool
	CanAuthorize         bool
	Edits                []string
	Command              string
}

func ProjectOutcomeWindowObservationLSP(
	observation OutcomeWindowObservation,
) OutcomeWindowObservationLSPProjection {
	projection := OutcomeWindowObservationLSPProjection{
		Status:        observation.Status,
		Code:          "jev.outcome_window.unknown",
		Severity:      "Error",
		Title:         "Outcome-window evidence is unresolved",
		TargetStage:   observation.TargetStage,
		RecordDigest:  observation.RecordDigest,
		Reason:        observation.Reason,
		IsReadOnly:    true,
		CanExecute:    false,
		CanAuthorize:  false,
	}

	if observation.RecordDigest == "" {
		projection.TargetStage = "record_digest"
		return projection
	}

	switch observation.Status {
	case OutcomeWindowBound:
		projection.Code = "jev.outcome_window.bound"
		projection.Severity = "Information"
		projection.Title = "Outcome-window evidence is bound"
	case OutcomeWindowDeferred:
		projection.Code = "jev.outcome_window.deferred"
		projection.Severity = "Hint"
		projection.Title = "Outcome-window evidence is deferred"
	}
	return projection
}