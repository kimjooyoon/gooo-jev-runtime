package jevcal

// ReplanProjection is the LSP-safe view of a runtime replan candidate. It
// contains no edit or command and cannot authorize execution.
type ReplanProjection struct {
	Status        ReplanStatus
	Code          string
	Title         string
	Kind          string
	TargetStage   string
	PlanDigest    string
	IsActionable  bool
	IsReadOnly    bool
	CanExecute    bool
	CanAuthorize  bool
	Edits         []string
	Command       string
}

// ProjectReplan projects bounded replan evidence into a diagnostic surface
// without converting a proposal into an executable action.
func ProjectReplan(candidate ReplanCandidate) ReplanProjection {
	projection := ReplanProjection{
		Status:       candidate.Status,
		Code:         "gooo.replan.unknown",
		Title:        "Inspect bounded replan evidence",
		Kind:         "quickfix",
		TargetStage:  candidate.TargetStage,
		PlanDigest:   candidate.PlanDigest,
		IsActionable: true,
		IsReadOnly:   true,
		CanExecute:   false,
		CanAuthorize: false,
	}

	switch candidate.Status {
	case ReplanProposed:
		projection.Code = "gooo.replan.proposed"
		projection.Title = "Inspect bounded replan proposal"
	case ReplanNoAction:
		projection.Code = "gooo.replan.no_action"
		projection.Title = "No replan is required"
		projection.Kind = "info"
		projection.IsActionable = false
	case ReplanDeferred:
		projection.Code = "gooo.replan.deferred"
		projection.Title = "Replan producer is deferred"
	case ReplanUnknown:
		projection.Code = "gooo.replan.unknown"
	}
	return projection
}