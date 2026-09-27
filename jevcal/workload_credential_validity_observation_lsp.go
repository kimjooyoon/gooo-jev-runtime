package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	WorkloadCredentialValidityLSPInformation = "INFORMATION"
	WorkloadCredentialValidityLSPWarning     = "WARNING"
	WorkloadCredentialValidityLSPError       = "ERROR"

	workloadCredentialValidityLSPBound    = "jev.security.workload_credential_validity.bound"
	workloadCredentialValidityLSPDeferred = "jev.security.workload_credential_validity.deferred"
	workloadCredentialValidityLSPUnknown  = "jev.security.workload_credential_validity.unknown"
	workloadCredentialValidityLSPIntegrity = "jev.security.workload_credential_validity.integrity"
	workloadCredentialValidityLSPBoundary = "jev.security.workload_credential_validity.capability_boundary"
)

// WorkloadCredentialValidityObservationLSP is a diagnostic-only projection of
// temporal and revocation evidence. It never authenticates or authorizes.
type WorkloadCredentialValidityObservationLSP struct {
	Status               string
	Publishable          bool
	Code                 string
	Message              string
	MissingStage         string
	WorkloadSpiffeID     string
	IssuerSpiffeID       string
	Audience             string
	CredentialDigest     string
	EvidencePrefixDigest string
	IssuedAtUnix         int64
	ExpiresAtUnix        int64
	ObservedAtUnix       int64
	RevocationStatus     string
	ObservationDigest    string
	ProjectionDigest     string
	IsReadOnly           bool
	CanEdit              bool
	CanExecute           bool
	CanAuthorize         bool
}

// ProjectWorkloadCredentialValidityLSP preserves the observation status and
// temporal stage while keeping the projection read-only and non-authorizing.
func ProjectWorkloadCredentialValidityLSP(
	input WorkloadCredentialValidityObservation,
) WorkloadCredentialValidityObservationLSP {
	projection := WorkloadCredentialValidityObservationLSP{
		Status:               WorkloadCredentialValidityLSPError,
		Publishable:          true,
		Code:                 workloadCredentialValidityLSPUnknown,
		Message:              "workload credential validity evidence is missing or unresolved",
		MissingStage:         input.MissingStage,
		WorkloadSpiffeID:     input.WorkloadSpiffeID,
		IssuerSpiffeID:       input.IssuerSpiffeID,
		Audience:             input.Audience,
		CredentialDigest:     input.CredentialDigest,
		EvidencePrefixDigest: input.EvidencePrefixDigest,
		IssuedAtUnix:         input.IssuedAtUnix,
		ExpiresAtUnix:        input.ExpiresAtUnix,
		ObservedAtUnix:       input.ObservedAtUnix,
		RevocationStatus:     input.RevocationStatus,
		ObservationDigest:    input.ObservationDigest,
		IsReadOnly:           input.IsReadOnly,
		CanEdit:              false,
		CanExecute:           input.CanExecute,
		CanAuthorize:         input.CanAuthorize,
	}
	if projection.MissingStage == "" {
		projection.MissingStage = "credential_validity_observation"
	}
	if !input.IsReadOnly || input.CanExecute || input.CanAuthorize {
		projection.Code = workloadCredentialValidityLSPBoundary
		projection.Message = "workload credential validity crossed a capability boundary"
		projection.MissingStage = "capability_boundary"
		projection.ProjectionDigest = workloadCredentialValidityLSPDigest(projection)
		return projection
	}
	if err := input.Validate(); err != nil {
		projection.Code = workloadCredentialValidityLSPIntegrity
		projection.Message = "workload credential validity observation failed integrity validation"
		projection.MissingStage = "observation_integrity"
		projection.ProjectionDigest = workloadCredentialValidityLSPDigest(projection)
		return projection
	}

	switch input.Status {
	case WorkloadCredentialValidityBound:
		projection.Status = WorkloadCredentialValidityLSPInformation
		projection.Publishable = false
		projection.Code = workloadCredentialValidityLSPBound
		projection.Message = "workload credential validity evidence is available for inspection"
		projection.MissingStage = ""
	case WorkloadCredentialValidityDeferred:
		projection.Status = WorkloadCredentialValidityLSPWarning
		projection.Code = workloadCredentialValidityLSPDeferred
		projection.Message = "workload credential validity evidence is deferred"
		projection.MissingStage = "credential_validity_producer"
	case WorkloadCredentialValidityUnknown:
		projection.Code = workloadCredentialValidityLSPUnknown
	default:
		projection.Code = workloadCredentialValidityLSPIntegrity
		projection.Message = "workload credential validity status is unrecognized"
		projection.MissingStage = "observation_status"
	}
	projection.ProjectionDigest = workloadCredentialValidityLSPDigest(projection)
	return projection
}

func (projection WorkloadCredentialValidityObservationLSP) Validate() error {
	switch projection.Status {
	case WorkloadCredentialValidityLSPInformation,
		WorkloadCredentialValidityLSPWarning,
		WorkloadCredentialValidityLSPError:
	default:
		return fmt.Errorf("invalid workload credential validity LSP status %q", projection.Status)
	}
	if projection.CanEdit || projection.CanExecute || projection.CanAuthorize {
		return fmt.Errorf("workload credential validity LSP crossed a forbidden boundary")
	}
	if projection.Status == WorkloadCredentialValidityLSPInformation {
		if projection.Publishable ||
			projection.MissingStage != "" ||
			projection.Code != workloadCredentialValidityLSPBound ||
			!projection.IsReadOnly ||
			projection.ObservationDigest == "" ||
			projection.CredentialDigest == "" ||
			projection.EvidencePrefixDigest == "" ||
			projection.ObservedAtUnix < projection.IssuedAtUnix ||
			projection.ObservedAtUnix >= projection.ExpiresAtUnix ||
			projection.RevocationStatus != "NOT_REVOKED" {
			return fmt.Errorf("bound workload credential validity LSP is incomplete")
		}
	}
	if projection.Status == WorkloadCredentialValidityLSPWarning {
		if !projection.Publishable ||
			projection.Code != workloadCredentialValidityLSPDeferred ||
			projection.MissingStage != "credential_validity_producer" {
			return fmt.Errorf("deferred workload credential validity LSP is incomplete")
		}
	}
	if projection.Status == WorkloadCredentialValidityLSPError &&
		projection.MissingStage == "" {
		return fmt.Errorf("unknown workload credential validity LSP must preserve its missing stage")
	}
	if projection.ProjectionDigest != workloadCredentialValidityLSPDigest(projection) {
		return fmt.Errorf("workload credential validity LSP projection digest mismatch")
	}
	return nil
}

func workloadCredentialValidityLSPDigest(
	projection WorkloadCredentialValidityObservationLSP,
) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		projection.Status,
		fmt.Sprint(projection.Publishable),
		projection.Code,
		projection.Message,
		projection.MissingStage,
		projection.WorkloadSpiffeID,
		projection.IssuerSpiffeID,
		projection.Audience,
		projection.CredentialDigest,
		projection.EvidencePrefixDigest,
		fmt.Sprint(projection.IssuedAtUnix),
		fmt.Sprint(projection.ExpiresAtUnix),
		fmt.Sprint(projection.ObservedAtUnix),
		projection.RevocationStatus,
		projection.ObservationDigest,
		fmt.Sprint(projection.IsReadOnly),
		fmt.Sprint(projection.CanEdit),
		fmt.Sprint(projection.CanExecute),
		fmt.Sprint(projection.CanAuthorize),
	}, "|")))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}