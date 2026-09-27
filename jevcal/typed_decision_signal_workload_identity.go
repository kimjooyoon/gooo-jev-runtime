package jevcal

import (
	"fmt"
	"strings"
)

const (
	TypedDecisionSignalIdentityBound  = "BOUND"
	TypedDecisionSignalIdentityUnknown = "UNKNOWN"

	TypedDecisionSignalSVIDX509 = "x509"
	TypedDecisionSignalSVIDJWT  = "jwt"
	TypedDecisionSignalSVIDWIT  = "wit"
)

type TypedDecisionSignalWorkloadIdentityInput struct {
	SourceVersion       string
	ContractVersion     string
	Receipt             TypedDecisionSignalReceipt
	TrustDomain         string
	SPIFFEID            string
	SVIDFormat          string
	CredentialDigest    string
	ValidUntilUnix      int64
	ObservedAtUnix      int64
	NonAuthorizing      bool
}

// TypedDecisionSignalWorkloadIdentityBinding records workload identity
// evidence next to a JEV receipt without issuing or authorizing an identity.
type TypedDecisionSignalWorkloadIdentityBinding struct {
	Status              string
	SourceVersion       string
	ContractVersion     string
	TrustDomain         string
	SPIFFEID            string
	SVIDFormat          string
	CredentialDigest    string
	ValidUntilUnix      int64
	ObservedAtUnix      int64
	SignalEvidenceDigest string
	IdentityDigest      string
	BindingDigest       string
	FirstMismatch       string
	NonExecuting        bool
	NonAuthorizing      bool
	CanExecute          bool
	CanAuthorize        bool
}

// BindTypedDecisionSignalWorkloadIdentity validates identity evidence beside a
// typed JEV signal. It never issues an SVID or turns identity into authority.
func BindTypedDecisionSignalWorkloadIdentity(
	input TypedDecisionSignalWorkloadIdentityInput,
) TypedDecisionSignalWorkloadIdentityBinding {
	binding := TypedDecisionSignalWorkloadIdentityBinding{
		Status:               TypedDecisionSignalIdentityUnknown,
		SourceVersion:        input.SourceVersion,
		ContractVersion:      input.ContractVersion,
		TrustDomain:          input.TrustDomain,
		SPIFFEID:             input.SPIFFEID,
		SVIDFormat:           input.SVIDFormat,
		CredentialDigest:     input.CredentialDigest,
		ValidUntilUnix:       input.ValidUntilUnix,
		ObservedAtUnix:       input.ObservedAtUnix,
		SignalEvidenceDigest: input.Receipt.EvidenceDigest,
		NonExecuting:         true,
		NonAuthorizing:       true,
		CanExecute:           false,
		CanAuthorize:         false,
	}

	switch {
	case !input.NonAuthorizing:
		binding.FirstMismatch = "authorization-boundary"
	case !input.Receipt.NonAuthorizing || input.Receipt.CanExecute || input.Receipt.CanAuthorize:
		binding.FirstMismatch = "signal-boundary"
	case input.Receipt.Validate() != nil:
		binding.FirstMismatch = "signal-integrity"
	case input.Receipt.Status != TypedDecisionSignalBound:
		binding.FirstMismatch = "signal-status"
	case strings.TrimSpace(input.SourceVersion) == "" || strings.TrimSpace(input.ContractVersion) == "":
		binding.FirstMismatch = "identity"
	case strings.TrimSpace(input.TrustDomain) == "":
		binding.FirstMismatch = "trust-domain"
	case !validTypedDecisionSignalSPIFFEID(input.SPIFFEID, input.TrustDomain):
		binding.FirstMismatch = "spiffe-id"
	case !validTypedDecisionSignalSVIDFormat(input.SVIDFormat):
		binding.FirstMismatch = "svid-format"
	case !typedDecisionSignalDigestValid(input.CredentialDigest):
		binding.FirstMismatch = "credential-digest"
	case input.ObservedAtUnix <= 0:
		binding.FirstMismatch = "observed-at"
	case input.ValidUntilUnix <= input.ObservedAtUnix:
		binding.FirstMismatch = "credential-expiry"
	default:
		binding.Status = TypedDecisionSignalIdentityBound
		binding.IdentityDigest = typedDecisionHash(fmt.Sprintf(
			"jev-workload-identity|%s|%s|%s|%s|%d|%d|%s",
			input.TrustDomain,
			input.SPIFFEID,
			input.SVIDFormat,
			input.CredentialDigest,
			input.ValidUntilUnix,
			input.ObservedAtUnix,
			input.Receipt.EvidenceDigest,
		))
	}

	binding.BindingDigest = typedDecisionSignalWorkloadIdentityDigest(binding)
	return binding
}

// Validate checks the identity binding's integrity and authority boundary.
// It does not prove that an external authority issued the credential.
func (binding TypedDecisionSignalWorkloadIdentityBinding) Validate() error {
	if binding.Status != TypedDecisionSignalIdentityBound &&
		binding.Status != TypedDecisionSignalIdentityUnknown {
		return fmt.Errorf("invalid workload identity binding status %q", binding.Status)
	}
	if !binding.NonExecuting || !binding.NonAuthorizing ||
		binding.CanExecute || binding.CanAuthorize {
		return fmt.Errorf("workload identity binding crossed an authority boundary")
	}
	if binding.BindingDigest != typedDecisionSignalWorkloadIdentityDigest(binding) {
		return fmt.Errorf("workload identity binding digest mismatch")
	}
	if binding.Status == TypedDecisionSignalIdentityUnknown {
		if binding.FirstMismatch == "" {
			return fmt.Errorf("unknown workload identity binding lost its first mismatch")
		}
		return nil
	}
	if binding.FirstMismatch != "" ||
		!validTypedDecisionSignalSPIFFEID(binding.SPIFFEID, binding.TrustDomain) ||
		!validTypedDecisionSignalSVIDFormat(binding.SVIDFormat) ||
		!typedDecisionSignalDigestValid(binding.CredentialDigest) ||
		!typedDecisionSignalDigestValid(binding.SignalEvidenceDigest) ||
		binding.ObservedAtUnix <= 0 ||
		binding.ValidUntilUnix <= binding.ObservedAtUnix ||
		binding.IdentityDigest == "" {
		return fmt.Errorf("bound workload identity binding is incomplete")
	}
	return nil
}

func validTypedDecisionSignalSPIFFEID(id, trustDomain string) bool {
	prefix := "spiffe://" + trustDomain + "/"
	return strings.HasPrefix(id, prefix) && len(id) > len(prefix)
}

func validTypedDecisionSignalSVIDFormat(format string) bool {
	switch format {
	case TypedDecisionSignalSVIDX509, TypedDecisionSignalSVIDJWT, TypedDecisionSignalSVIDWIT:
		return true
	default:
		return false
	}
}

func typedDecisionSignalWorkloadIdentityDigest(
	binding TypedDecisionSignalWorkloadIdentityBinding,
) string {
	return typedDecisionHash(fmt.Sprintf(
		"jev-workload-identity-binding|%s|%s|%s|%s|%s|%s|%d|%d|%s|%s|%s|%s|%t|%t|%t|%t",
		binding.Status,
		binding.SourceVersion,
		binding.ContractVersion,
		binding.TrustDomain,
		binding.SPIFFEID,
		binding.SVIDFormat,
		binding.ValidUntilUnix,
		binding.ObservedAtUnix,
		binding.CredentialDigest,
		binding.SignalEvidenceDigest,
		binding.IdentityDigest,
		binding.FirstMismatch,
		binding.NonExecuting,
		binding.NonAuthorizing,
		binding.CanExecute,
		binding.CanAuthorize,
	))
}

