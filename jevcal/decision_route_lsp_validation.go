package jevcal

import "fmt"

// Validate closes the decision-route LSP boundary. A route is an inspection
// signal only; this method never promotes it to an executable action.
func (projection DecisionRouteLSPProjection) Validate() error {
	if !projection.ReadOnly || projection.CanExecute || projection.CanAuthorize {
		return fmt.Errorf("decision route LSP crossed a forbidden capability boundary")
	}
	if projection.Command != "" || len(projection.Edits) != 0 {
		return fmt.Errorf("decision route LSP contains executable edits or a command")
	}
	if projection.DecisionDigest == "" {
		return fmt.Errorf("decision route LSP is missing its decision digest")
	}
	switch projection.Status {
	case "ACCEPT_CANDIDATE":
		if projection.Code != "jev.route.accept_candidate" || projection.Severity != "Information" || !projection.Actionable {
			return fmt.Errorf("accept candidate decision route LSP is incomplete")
		}
	case "REVIEW_CANDIDATE":
		if projection.Code != "jev.route.review_candidate" || projection.Severity != "Warning" || !projection.Actionable {
			return fmt.Errorf("review candidate decision route LSP is incomplete")
		}
	case "ABSTAIN_CANDIDATE":
		if projection.Code != "jev.route.abstain_candidate" || projection.Severity != "Information" || !projection.Actionable {
			return fmt.Errorf("abstain candidate decision route LSP is incomplete")
		}
	case "DEFERRED":
		if projection.Code != "jev.route.deferred" || projection.Severity != "Hint" || projection.Actionable {
			return fmt.Errorf("deferred decision route LSP is incomplete")
		}
	case "UNKNOWN":
		if projection.Actionable || projection.TargetStage == "" || projection.Code != "jev.route.unknown" {
			return fmt.Errorf("unknown decision route LSP must preserve its unresolved stage")
		}
	default:
		return fmt.Errorf("invalid decision route LSP status %q", projection.Status)
	}
	return nil
}