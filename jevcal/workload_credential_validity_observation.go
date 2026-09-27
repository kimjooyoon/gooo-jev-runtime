package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type WorkloadCredentialValidityStatus string

const (
	WorkloadCredentialValidityBound    WorkloadCredentialValidityStatus = "BOUND"
	WorkloadCredentialValidityDeferred WorkloadCredentialValidityStatus = "DEFERRED"
	WorkloadCredentialValidityUnknown  WorkloadCredentialValidityStatus = "UNKNOWN"

	workloadCredentialValidityBoundSignal = "jev-workload-credential-validity-bound"
	workloadCredentialValidityDeferredSignal = "jev-workload-credential-validity-deferred"
	workloadCredentialValidityUnknownSignal = "jev-workload-credential-validity-unknown"
)

// WorkloadCredentialValidityInput records time and revocation evidence for a
// SPIFFE-like credential without making an authentication decision.
type WorkloadCredentialValidityInput struct {
	SourceVersion       string
	ContractVersion     string
	WorkloadSpiffeID    string
	IssuerSpiffeID      string
	Audience            string
	CredentialDigest    string
	EvidencePrefixDigest string
	IssuedAtUnix        int64
	ExpiresAtUnix       int64
	ObservedAtUnix      int64
	RevocationStatus    string
	ProducerDeferred    bool
}

// WorkloadCredentialValidityObservation is a read-only temporal evidence
// observation. It never authenticates, authorizes, or executes a workload.
type WorkloadCredentialValidityObservation struct {
	Status               WorkloadCredentialValidityStatus
	SourceVersion        string
	ContractVersion      string
	WorkloadSpiffeID     string
	IssuerSpiffeID       string
	Audience             string
	CredentialDigest     string
	EvidencePrefixDigest string
	IssuedAtUnix         int64
	ExpiresAtUnix        int64
	ObservedAtUnix       int64
	RevocationStatus     string
	MissingStage         string
	TargetStage          string
	Reason               string
	ValiditySignal       string
	ObservationDigest    string
	IsReadOnly           bool
	CanExecute           bool
	CanAuthorize         bool
}

// ObserveWorkloadCredentialValidity preserves the first unresolved temporal
// or revocation stage and binds only complete evidence.
func ObserveWorkloadCredentialValidity(input WorkloadCredentialValidityInput) WorkloadCredentialValidityObservation {
	observation := WorkloadCredentialValidityObservation{
		Status:               WorkloadCredentialValidityUnknown,
		SourceVersion:        input.SourceVersion,
		ContractVersion:      input.ContractVersion,
		WorkloadSpiffeID:     input.WorkloadSpiffeID,
		IssuerSpiffeID:       input.IssuerSpiffeID,
		Audience:             input.Audience,
		CredentialDigest:     input.CredentialDigest,
		EvidencePrefixDigest: input.EvidencePrefixDigest,
		IssuedAtUnix:         input.IssuedAtUnix,
		ExpiresAtUnix:        input.ExpiresAtUnix,
		ObservedAtUnix:       input.ObservedAtUnix,
		RevocationStatus:     input.RevocationStatus,
		TargetStage:          "credential_identity",
		Reason:               "workload credential validity evidence is incomplete",
		ValiditySignal:       workloadCredentialValidityUnknownSignal,
		IsReadOnly:           true,
		CanExecute:           false,
		CanAuthorize:         false,
	}

	switch {
	case input.SourceVersion == "" || input.ContractVersion == "":
		observation.MissingStage = "credential_identity"
		observation.Reason = "source or contract identity is missing"
	case !validRuntimeSpiffeID(input.WorkloadSpiffeID):
		observation.MissingStage = "workload_identity"
		observation.TargetStage = "workload_identity"
		observation.Reason = "workload SPIFFE identity is malformed"
	case !validRuntimeSpiffeID(input.IssuerSpiffeID):
		observation.MissingStage = "issuer_identity"
		observation.TargetStage = "issuer_identity"
		observation.Reason = "issuer SPIFFE identity is malformed"
	case input.Audience == "":
		observation.MissingStage = "audience"
		observation.TargetStage = "audience"
		observation.Reason = "credential audience is missing"
	case input.CredentialDigest == "":
		observation.MissingStage = "credential_digest"
		observation.TargetStage = "credential_digest"
		observation.Reason = "credential digest is missing"
	case input.EvidencePrefixDigest == "":
		observation.MissingStage = "evidence_prefix_digest"
		observation.TargetStage = "evidence_prefix_digest"
		observation.Reason = "credential evidence prefix digest is missing"
	case input.IssuedAtUnix <= 0:
		observation.MissingStage = "issued_at"
		observation.TargetStage = "issued_at"
		observation.Reason = "credential issued-at time is missing"
	case input.ExpiresAtUnix <= input.IssuedAtUnix:
		observation.MissingStage = "expires_at"
		observation.TargetStage = "expires_at"
		observation.Reason = "credential expiry is not after issued-at time"
	case input.ObservedAtUnix <= 0:
		observation.MissingStage = "observed_at"
		observation.TargetStage = "observed_at"
		observation.Reason = "credential observation time is missing"
	case input.ProducerDeferred:
		observation.Status = WorkloadCredentialValidityDeferred
		observation.MissingStage = ""
		observation.TargetStage = "credential_validity_producer"
		observation.Reason = "credential validity producer is deferred"
		observation.ValiditySignal = workloadCredentialValidityDeferredSignal
	case input.RevocationStatus == "":
		observation.MissingStage = "revocation_status"
		observation.TargetStage = "revocation_status"
		observation.Reason = "credential revocation status is missing"
	case input.RevocationStatus == "UNKNOWN":
		observation.MissingStage = "revocation_status"
		observation.TargetStage = "revocation_status"
		observation.Reason = "credential revocation status is unresolved"
	case input.RevocationStatus == "REVOKED":
		observation.MissingStage = "revocation_status"
		observation.TargetStage = "revocation_status"
		observation.Reason = "credential is observed as revoked"
	case input.ObservedAtUnix >= input.ExpiresAtUnix:
		observation.MissingStage = "credential_expiry"
		observation.TargetStage = "credential_expiry"
		observation.Reason = "credential was observed after expiry"
	case input.RevocationStatus != "NOT_REVOKED":
		observation.MissingStage = "revocation_status"
		observation.TargetStage = "revocation_status"
		observation.Reason = "credential revocation status is not recognized"
	default:
		observation.Status = WorkloadCredentialValidityBound
		observation.MissingStage = ""
		observation.TargetStage = "credential_validity"
		observation.Reason = "issuer, audience, temporal, and revocation evidence are aligned"
		observation.ValiditySignal = workloadCredentialValidityBoundSignal
	}

	observation.ObservationDigest = workloadCredentialValidityDigest(observation)
	return observation
}

func (observation WorkloadCredentialValidityObservation) Validate() error {
	switch observation.Status {
	case WorkloadCredentialValidityBound, WorkloadCredentialValidityDeferred, WorkloadCredentialValidityUnknown:
	default:
		return fmt.Errorf("invalid workload credential validity status %q", observation.Status)
	}
	if observation.TargetStage == "" || observation.Reason == "" || observation.ObservationDigest == "" {
		return fmt.Errorf("workload credential validity identity is incomplete")
	}
	if !observation.IsReadOnly || observation.CanExecute || observation.CanAuthorize {
		return fmt.Errorf("workload credential validity crossed an execution or authorization boundary")
	}
	if observation.Status == WorkloadCredentialValidityUnknown {
		if observation.MissingStage == "" || observation.ValiditySignal != workloadCredentialValidityUnknownSignal {
			return fmt.Errorf("unknown workload credential validity must preserve its missing stage")
		}
	} else if observation.Status == WorkloadCredentialValidityDeferred {
		if observation.MissingStage != "" || observation.ValiditySignal != workloadCredentialValidityDeferredSignal {
			return fmt.Errorf("deferred workload credential validity is invalid")
		}
	} else {
		if observation.MissingStage != "" ||
			observation.ValiditySignal != workloadCredentialValidityBoundSignal ||
			!validRuntimeSpiffeID(observation.WorkloadSpiffeID) ||
			!validRuntimeSpiffeID(observation.IssuerSpiffeID) ||
			observation.Audience == "" ||
			observation.CredentialDigest == "" ||
			observation.EvidencePrefixDigest == "" ||
			observation.IssuedAtUnix <= 0 ||
			observation.ExpiresAtUnix <= observation.IssuedAtUnix ||
			observation.ObservedAtUnix < observation.IssuedAtUnix ||
			observation.ObservedAtUnix >= observation.ExpiresAtUnix ||
			observation.RevocationStatus != "NOT_REVOKED" {
			return fmt.Errorf("bound workload credential validity is incomplete")
		}
	}
	if observation.ObservationDigest != workloadCredentialValidityDigest(observation) {
		return fmt.Errorf("workload credential validity digest mismatch")
	}
	return nil
}

func workloadCredentialValidityDigest(observation WorkloadCredentialValidityObservation) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf(
		"%s|%s|%s|%s|%s|%s|%s|%d|%d|%d|%s|%s|%s|%s|%t|%t|%t",
		observation.Status,
		observation.SourceVersion,
		observation.ContractVersion,
		observation.WorkloadSpiffeID,
		observation.IssuerSpiffeID,
		observation.Audience,
		observation.CredentialDigest,
		observation.IssuedAtUnix,
		observation.ExpiresAtUnix,
		observation.ObservedAtUnix,
		observation.RevocationStatus,
		observation.EvidencePrefixDigest,
		observation.MissingStage,
		observation.TargetStage,
		observation.Reason,
		observation.ValiditySignal,
		observation.IsReadOnly,
		observation.CanExecute,
		observation.CanAuthorize,
	)))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}

// Keep strings linked to this package's security validation surface.
var _ = strings.HasPrefix