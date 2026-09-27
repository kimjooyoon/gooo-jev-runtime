package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

const (
	JEVWorkloadIdentityLSPInformation = "INFORMATION"
	JEVWorkloadIdentityLSPError       = "ERROR"

	jevWorkloadIdentityLSPBound    = "jev.security.workload_identity.bound"
	jevWorkloadIdentityLSPUnknown  = "jev.security.workload_identity.unknown"
	jevWorkloadIdentityLSPBoundary = "jev.security.workload_identity.capability_boundary"
	jevWorkloadIdentityLSPIntegrity = "jev.security.workload_identity.integrity"
)

// JEVWorkloadIdentityObservationLSPInput mirrors the runtime observation and
// its safety flags without authorizing any capability.
type JEVWorkloadIdentityObservationLSPInput struct {
	Status               JEVWorkloadIdentityObservationStatus
	SpiffeID             string
	Audience             string
	Capabilities         []string
	EvidencePrefixDigest string
	IsReadOnly           bool
	CanExecute           bool
	CanAuthorize         bool
}

// JEVWorkloadIdentityObservationLSP is a diagnostic-only projection. It
// validates identity shape and evidence linkage but never authenticates,
// authorizes, or executes a workload.
type JEVWorkloadIdentityObservationLSP struct {
	Status               string
	Publishable          bool
	Code                 string
	Message              string
	MissingStage         string
	SpiffeID             string
	Audience             string
	Capabilities         []string
	EvidencePrefixDigest string
	IdentityDigest       string
	ObservationDigest    string
	IsReadOnly           bool
	CanEdit              bool
	CanExecute           bool
	CanAuthorize         bool
}

// ProjectJEVWorkloadIdentityObservationLSP preserves an identity observation
// as an editor diagnostic while failing closed on malformed or unsafe input.
func ProjectJEVWorkloadIdentityObservationLSP(
	input JEVWorkloadIdentityObservationLSPInput,
) JEVWorkloadIdentityObservationLSP {
	projection := JEVWorkloadIdentityObservationLSP{
		Status:               JEVWorkloadIdentityLSPError,
		Publishable:          true,
		Code:                 jevWorkloadIdentityLSPUnknown,
		Message:              "workload identity evidence is missing or incomplete",
		MissingStage:         "workload_identity_evidence",
		SpiffeID:             input.SpiffeID,
		Audience:             input.Audience,
		Capabilities:         append([]string(nil), input.Capabilities...),
		EvidencePrefixDigest: input.EvidencePrefixDigest,
		IsReadOnly:           input.IsReadOnly,
		CanEdit:              false,
		CanExecute:           input.CanExecute,
		CanAuthorize:         input.CanAuthorize,
	}
	if !input.IsReadOnly || input.CanExecute || input.CanAuthorize {
		projection.Code = jevWorkloadIdentityLSPBoundary
		projection.Message = "workload identity observation crossed a capability boundary"
		projection.MissingStage = "capability_boundary"
		projection.ObservationDigest = workloadIdentityLSPDigest(projection)
		return projection
	}
	if input.Status != JEVWorkloadIdentityBound ||
		!validRuntimeSpiffeID(input.SpiffeID) ||
		input.Audience == "" ||
		input.EvidencePrefixDigest == "" {
		projection.ObservationDigest = workloadIdentityLSPDigest(projection)
		return projection
	}

	projection.Status = JEVWorkloadIdentityLSPInformation
	projection.Publishable = false
	projection.Code = jevWorkloadIdentityLSPBound
	projection.Message = "workload identity evidence is available for inspection"
	projection.MissingStage = ""
	projection.IdentityDigest = workloadIdentityIdentityDigest(
		input.SpiffeID,
		input.Audience,
		input.Capabilities,
		input.EvidencePrefixDigest,
	)
	projection.ObservationDigest = workloadIdentityLSPDigest(projection)
	return projection
}

func (projection JEVWorkloadIdentityObservationLSP) Validate() error {
	if projection.Status != JEVWorkloadIdentityLSPInformation &&
		projection.Status != JEVWorkloadIdentityLSPError {
		return fmt.Errorf("invalid workload identity LSP status %q", projection.Status)
	}
	if projection.CanEdit || projection.CanExecute || projection.CanAuthorize {
		return fmt.Errorf("workload identity LSP crossed a forbidden capability boundary")
	}
	if projection.Status == JEVWorkloadIdentityLSPInformation {
		if projection.Publishable ||
			projection.MissingStage != "" ||
			projection.Code != jevWorkloadIdentityLSPBound ||
			!projection.IsReadOnly ||
			!validRuntimeSpiffeID(projection.SpiffeID) ||
			projection.Audience == "" ||
			projection.EvidencePrefixDigest == "" ||
			projection.IdentityDigest == "" {
			return fmt.Errorf("bound workload identity LSP is incomplete")
		}
		expectedIdentity := workloadIdentityIdentityDigest(
			projection.SpiffeID,
			projection.Audience,
			projection.Capabilities,
			projection.EvidencePrefixDigest,
		)
		if projection.IdentityDigest != expectedIdentity {
			return fmt.Errorf("workload identity digest mismatch")
		}
	} else if projection.MissingStage == "" {
		return fmt.Errorf("unknown workload identity LSP must preserve its missing stage")
	}
	if projection.ObservationDigest != workloadIdentityLSPDigest(projection) {
		return fmt.Errorf("workload identity LSP observation digest mismatch")
	}
	return nil
}

func validRuntimeSpiffeID(id string) bool {
	if !strings.HasPrefix(id, "spiffe://") {
		return false
	}
	authorityAndPath := strings.TrimPrefix(id, "spiffe://")
	parts := strings.SplitN(authorityAndPath, "/", 2)
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

func workloadIdentityIdentityDigest(
	spiffeID string,
	audience string,
	capabilities []string,
	evidencePrefixDigest string,
) string {
	canonical := append([]string(nil), capabilities...)
	sort.Strings(canonical)
	sum := sha256.Sum256([]byte(strings.Join([]string{
		spiffeID,
		audience,
		strings.Join(canonical, ","),
		evidencePrefixDigest,
	}, "|")))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}

func workloadIdentityLSPDigest(projection JEVWorkloadIdentityObservationLSP) string {
	return digestWorkloadIdentityLSP([]string{
		projection.Status,
		fmt.Sprint(projection.Publishable),
		projection.Code,
		projection.Message,
		projection.MissingStage,
		projection.SpiffeID,
		projection.Audience,
		strings.Join(projection.Capabilities, ","),
		projection.EvidencePrefixDigest,
		projection.IdentityDigest,
		fmt.Sprint(projection.IsReadOnly),
		fmt.Sprint(projection.CanEdit),
		fmt.Sprint(projection.CanExecute),
		fmt.Sprint(projection.CanAuthorize),
	})
}

func digestWorkloadIdentityLSP(parts []string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}