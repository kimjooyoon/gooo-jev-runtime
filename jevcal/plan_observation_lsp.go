package jevcal

// PlanObservationLSPProjection is an inspection-only view of plan application
// provenance. It never contains an edit, command, execution, or authorization.
type PlanObservationLSPProjection struct {
	Status                   PlanObservationStatus
	Code                     string
	Severity                 string
	Title                    string
	TargetStage              string
	ObservationDigest        string
	Reason                   string
	IsReadOnly               bool
	CanExecute               bool
	CanAuthorize             bool
	Edits                    []string
	Command                  string
}

func ProjectPlanObservationLSP(
	observation PlanApplicationObservation,
) PlanObservationLSPProjection {
	projection := PlanObservationLSPProjection{
		Status:                   observation.Status,
		Code:                     "jev.plan_application.unknown",
		Severity:                 "Error",
		Title:                    "Plan application evidence is unresolved",
		TargetStage:              observation.TargetStage,
		ObservationDigest:        observation.ObservationDigest,
		Reason:                   observation.Reason,
		IsReadOnly:               true,
		CanExecute:               false,
		CanAuthorize:             false,
	}

	if observation.ObservationDigest == "" {
		projection.TargetStage = "observation_digest"
		return projection
	}

	switch observation.Status {
	case PlanObservationBound:
		projection.Code = "jev.plan_application.bound"
		projection.Severity = "Information"
		projection.Title = "Plan application evidence is bound"
	case PlanObservationDeferred:
		projection.Code = "jev.plan_application.deferred"
		projection.Severity = "Hint"
		projection.Title = "Plan application evidence is deferred"
	}
	return projection
}