package jevcal

// PlanOutcomeWindowObservationLSPProjection is an inspection-only view of
// linked plan and outcome-window provenance. It never edits, executes, or authorizes.
type PlanOutcomeWindowObservationLSPProjection struct {
	Status                    PlanOutcomeWindowObservationStatus
	Code                      string
	Severity                  string
	Title                     string
	TargetStage               string
	ChainDigest               string
	Reason                    string
	IsReadOnly                bool
	CanExecute                bool
	CanAuthorize              bool
	Edits                     []string
	Command                   string
}

func ProjectPlanOutcomeWindowObservationLSP(
	observation PlanOutcomeWindowObservation,
) PlanOutcomeWindowObservationLSPProjection {
	projection := PlanOutcomeWindowObservationLSPProjection{
		Status:       PlanOutcomeWindowUnknown,
		Code:         "jev.plan_outcome_window.unknown",
		Severity:     "Error",
		Title:        "Plan and outcome-window provenance is unresolved",
		TargetStage:  observation.TargetStage,
		ChainDigest:  observation.ChainDigest,
		Reason:       observation.Reason,
		IsReadOnly:   true,
		CanExecute:   false,
		CanAuthorize: false,
	}

	if observation.ChainDigest == "" {
		projection.TargetStage = "chain_digest"
		return projection
	}

	switch observation.Status {
	case PlanOutcomeWindowBound:
		projection.Status = PlanOutcomeWindowBound
		projection.Code = "jev.plan_outcome_window.bound"
		projection.Severity = "Information"
		projection.Title = "Plan and outcome-window provenance is bound"
	case PlanOutcomeWindowDeferred:
		projection.Status = PlanOutcomeWindowDeferred
		projection.Code = "jev.plan_outcome_window.deferred"
		projection.Severity = "Hint"
		projection.Title = "Plan and outcome-window provenance is deferred"
	}
	return projection
}
