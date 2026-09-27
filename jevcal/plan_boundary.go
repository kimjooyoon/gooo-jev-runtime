package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type PlanBoundaryStatus string

const (
	PlanBoundaryProposed PlanBoundaryStatus = "PROPOSED"
	PlanBoundaryDeferred PlanBoundaryStatus = "DEFERRED"
	PlanBoundaryUnknown  PlanBoundaryStatus = "UNKNOWN"
)

// PlanBoundaryInput is the evidence required to form a proposal for plan
// application. It never contains an authorization decision.
type PlanBoundaryInput struct {
	SourceVersion     string
	ContractVersion   string
	RouteRecordDigest string
	PlanDigest        string
	RouteStatus       string
	ProducerDeferred  bool
}

// PlanBoundaryProjection is a read-only proposal boundary, not a plan runner.
type PlanBoundaryProjection struct {
	Status            PlanBoundaryStatus
	SourceVersion     string
	ContractVersion   string
	RouteRecordDigest string
	PlanDigest        string
	TargetStage       string
	Reason            string
	BoundaryDigest    string
	IsReadOnly        bool
	CanExecute        bool
	CanAuthorize      bool
	Edits             []string
	Command           string
}

func ProjectPlanBoundary(input PlanBoundaryInput) PlanBoundaryProjection {
	projection := PlanBoundaryProjection{
		Status:            PlanBoundaryUnknown,
		SourceVersion:     input.SourceVersion,
		ContractVersion:   input.ContractVersion,
		RouteRecordDigest: input.RouteRecordDigest,
		PlanDigest:        input.PlanDigest,
		TargetStage:       "plan_application",
		IsReadOnly:        true,
		CanExecute:        false,
		CanAuthorize:      false,
	}
	switch {
	case input.SourceVersion == "" || input.ContractVersion == "":
		projection.TargetStage = "identity"
		projection.Reason = "source or contract identity is missing"
	case input.RouteRecordDigest == "":
		projection.TargetStage = "route_record"
		projection.Reason = "route record evidence is missing"
	case input.PlanDigest == "":
		projection.TargetStage = "plan_digest"
		projection.Reason = "plan digest is missing"
	case input.ProducerDeferred:
		projection.Status = PlanBoundaryDeferred
		projection.Reason = "plan producer is deferred"
	case input.RouteStatus == "UNKNOWN":
		projection.Reason = "route evidence remains unresolved"
	case input.RouteStatus == "DEFERRED":
		projection.Status = PlanBoundaryDeferred
		projection.Reason = "route evidence is deferred"
	case input.RouteStatus == "ACCEPT_CANDIDATE" || input.RouteStatus == "REVIEW_CANDIDATE" || input.RouteStatus == "ABSTAIN_CANDIDATE":
		projection.Status = PlanBoundaryProposed
		projection.Reason = "form a proposal for explicit downstream review"
	default:
		projection.TargetStage = "route_status"
		projection.Reason = "route status is not recognized"
	}
	projection.BoundaryDigest = planBoundaryDigest(
		string(projection.Status),
		projection.SourceVersion,
		projection.ContractVersion,
		projection.RouteRecordDigest,
		projection.PlanDigest,
		projection.TargetStage,
		projection.Reason,
	)
	return projection
}

func planBoundaryDigest(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}