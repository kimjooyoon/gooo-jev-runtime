package jevcal

import "fmt"

// Validate enforces the .gooo replan LSP boundary after projection. A
// diagnostic may describe a proposal, but it cannot carry executable edits,
// commands, or authority.
func (projection ReplanProjection) Validate() error {
	if !projection.IsReadOnly || projection.CanExecute || projection.CanAuthorize {
		return fmt.Errorf("replan LSP projection crossed a forbidden capability boundary")
	}
	if projection.Command != "" || len(projection.Edits) != 0 {
		return fmt.Errorf("replan LSP projection contains executable edits or a command")
	}
	if projection.PlanDigest == "" {
		return fmt.Errorf("replan LSP projection is missing its plan digest")
	}
	switch projection.Status {
	case ReplanProposed:
		if projection.Code != "gooo.replan.proposed" || !projection.IsActionable || projection.TargetStage == "" {
			return fmt.Errorf("proposed replan LSP projection is incomplete")
		}
	case ReplanNoAction:
		if projection.Code != "gooo.replan.no_action" || projection.IsActionable || projection.Kind != "info" {
			return fmt.Errorf("no-action replan LSP projection is incomplete")
		}
	case ReplanDeferred:
		if projection.Code != "gooo.replan.deferred" || !projection.IsActionable {
			return fmt.Errorf("deferred replan LSP projection is incomplete")
		}
	case ReplanUnknown:
		if projection.Code != "gooo.replan.unknown" || !projection.IsActionable {
			return fmt.Errorf("unknown replan LSP projection is incomplete")
		}
	default:
		return fmt.Errorf("invalid replan LSP status %q", projection.Status)
	}
	return nil
}