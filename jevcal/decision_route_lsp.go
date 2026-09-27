package jevcal

// DecisionRouteLSPProjection is an inspection-only LSP-shaped view of a
// calibrated decision route. It never carries edits, commands, execution, or
// authorization.
type DecisionRouteLSPProjection struct {
	Status        string
	Code          string
	Severity      string
	Title         string
	Kind          string
	TargetStage   string
	DecisionDigest string
	Actionable    bool
	ReadOnly      bool
	CanExecute    bool
	CanAuthorize  bool
	Edits         []string
	Command       string
}

func ProjectDecisionRouteLSP(status, targetStage, decisionDigest string) DecisionRouteLSPProjection {
	projection := DecisionRouteLSPProjection{
		Status:         status,
		Code:           "jev.route.unknown",
		Severity:       "Error",
		Title:          "Decision route evidence is incomplete",
		Kind:           "diagnostic",
		TargetStage:    targetStage,
		DecisionDigest: decisionDigest,
		ReadOnly:       true,
		CanExecute:     false,
		CanAuthorize:   false,
	}
	if decisionDigest == "" {
		projection.Status = "UNKNOWN"
		projection.TargetStage = "decision_digest"
		return projection
	}

	switch status {
	case "ACCEPT_CANDIDATE":
		projection.Code = "jev.route.accept_candidate"
		projection.Severity = "Information"
		projection.Title = "Acceptance candidate requires explicit authorization"
		projection.Actionable = true
	case "REVIEW_CANDIDATE":
		projection.Code = "jev.route.review_candidate"
		projection.Severity = "Warning"
		projection.Title = "Review candidate requires human inspection"
		projection.Actionable = true
	case "ABSTAIN_CANDIDATE":
		projection.Code = "jev.route.abstain_candidate"
		projection.Severity = "Information"
		projection.Title = "Evidence recommends abstention"
		projection.Actionable = true
	case "DEFERRED":
		projection.Code = "jev.route.deferred"
		projection.Severity = "Hint"
		projection.Title = "Decision producer is deferred"
	default:
		projection.Status = "UNKNOWN"
		projection.TargetStage = "route_status"
	}
	return projection
}