package jevcal

type OutcomeWindowDeltaObservationLSPStatus string

const (
	OutcomeWindowDeltaLSPBound    OutcomeWindowDeltaObservationLSPStatus = "BOUND"
	OutcomeWindowDeltaLSPDeferred OutcomeWindowDeltaObservationLSPStatus = "DEFERRED"
	OutcomeWindowDeltaLSPUnknown  OutcomeWindowDeltaObservationLSPStatus = "UNKNOWN"
)

// OutcomeWindowDeltaObservationLSPProjection is an inspection-only view of
// a signed metric. It cannot edit, execute, or authorize.
type OutcomeWindowDeltaObservationLSPProjection struct {
	Status       OutcomeWindowDeltaObservationLSPStatus
	Code         string
	Severity     string
	Title        string
	TargetStage  string
	Delta        int
	RecordDigest string
	Reason       string
	IsReadOnly   bool
	CanExecute   bool
	CanAuthorize bool
	Edits        []string
	Command      string
}

func ProjectOutcomeWindowDeltaObservationLSP(
	observation OutcomeWindowDeltaObservation,
) OutcomeWindowDeltaObservationLSPProjection {
	projection := OutcomeWindowDeltaObservationLSPProjection{
		Status:       OutcomeWindowDeltaLSPUnknown,
		Code:         "jev.outcome_window_delta.unknown",
		Severity:     "Error",
		Title:        "Signed outcome-window delta is unresolved",
		TargetStage:  observation.TargetStage,
		Delta:        observation.Delta,
		RecordDigest: observation.RecordDigest,
		Reason:       observation.Reason,
		IsReadOnly:   true,
		CanExecute:   false,
		CanAuthorize: false,
	}

	if observation.RecordDigest == "" {
		projection.TargetStage = "record_digest"
		return projection
	}

	switch observation.Status {
	case OutcomeWindowDeltaBound:
		projection.Status = OutcomeWindowDeltaLSPBound
		projection.Code = "jev.outcome_window_delta.bound"
		projection.Severity = "Information"
		projection.Title = "Signed outcome-window delta is recorded"
	case OutcomeWindowDeltaDeferred:
		projection.Status = OutcomeWindowDeltaLSPDeferred
		projection.Code = "jev.outcome_window_delta.deferred"
		projection.Severity = "Hint"
		projection.Title = "Signed outcome-window delta is deferred"
	}
	return projection
}
