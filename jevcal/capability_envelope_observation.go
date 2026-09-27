package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

type CapabilityEnvelopeObservationStatus string

const (
	CapabilityEnvelopeBound    CapabilityEnvelopeObservationStatus = "BOUND"
	CapabilityEnvelopeDeferred CapabilityEnvelopeObservationStatus = "DEFERRED"
	CapabilityEnvelopeUnknown  CapabilityEnvelopeObservationStatus = "UNKNOWN"
)

// CapabilityEnvelopeObservationInput links a plan, workload identity, and
// capability evidence without granting permission to use any capability.
type CapabilityEnvelopeObservationInput struct {
	SourceVersion            string
	ContractVersion          string
	WorkloadIdentityStatus   JEVWorkloadIdentityObservationStatus
	EvidencePrefixDigest     string
	PlanDigest               string
	NetworkAllowlistDigest   string
	RequestedCapabilities    []string
	ObservedCapabilities     []string
	ReverseObservationStatus string
	ReverseObservationDigest string
	ProducerDeferred         bool
}

// CapabilityEnvelopeObservation is a read-only capability boundary result. It
// is not an authorization decision and cannot execute a plan.
type CapabilityEnvelopeObservation struct {
	Status                 CapabilityEnvelopeObservationStatus
	SourceVersion          string
	ContractVersion        string
	WorkloadIdentityStatus JEVWorkloadIdentityObservationStatus
	EvidencePrefixDigest   string
	PlanDigest             string
	NetworkAllowlistDigest string
	RequestedCapabilities  []string
	ObservedCapabilities   []string
	MissingCapability      string
	TargetStage            string
	Reason                 string
	ObservationDigest      string
	IsReadOnly             bool
	CanExecute             bool
	CanAuthorize           bool
	Edits                  []string
	Command                string
}

// ObserveCapabilityEnvelope compares requested capabilities with observed
// evidence. Missing or unresolved evidence remains UNKNOWN or DEFERRED.
func ObserveCapabilityEnvelope(input CapabilityEnvelopeObservationInput) CapabilityEnvelopeObservation {
	observation := CapabilityEnvelopeObservation{
		Status:                 CapabilityEnvelopeUnknown,
		SourceVersion:          input.SourceVersion,
		ContractVersion:        input.ContractVersion,
		WorkloadIdentityStatus: input.WorkloadIdentityStatus,
		EvidencePrefixDigest:   input.EvidencePrefixDigest,
		PlanDigest:             input.PlanDigest,
		NetworkAllowlistDigest: input.NetworkAllowlistDigest,
		RequestedCapabilities:  append([]string(nil), input.RequestedCapabilities...),
		ObservedCapabilities:   append([]string(nil), input.ObservedCapabilities...),
		TargetStage:            "capability_envelope_observation",
		IsReadOnly:             true,
		CanExecute:             false,
		CanAuthorize:           false,
	}

	missingCapability := firstMissingCapability(input.RequestedCapabilities, input.ObservedCapabilities)
	switch {
	case input.SourceVersion == "" || input.ContractVersion == "":
		observation.TargetStage = "identity"
		observation.Reason = "source or contract identity is missing"
	case input.EvidencePrefixDigest == "":
		observation.TargetStage = "evidence_prefix"
		observation.Reason = "evidence prefix digest is missing"
	case input.PlanDigest == "":
		observation.TargetStage = "plan_digest"
		observation.Reason = "plan digest is missing"
	case input.NetworkAllowlistDigest == "":
		observation.TargetStage = "network_allowlist"
		observation.Reason = "network allowlist digest is missing"
	case input.ProducerDeferred:
		observation.Status = CapabilityEnvelopeDeferred
		observation.Reason = "capability envelope producer is deferred"
	case input.WorkloadIdentityStatus != JEVWorkloadIdentityBound:
		observation.TargetStage = "workload_identity"
		observation.Reason = "workload identity evidence is not BOUND"
	case input.ReverseObservationStatus == "DEFERRED":
		observation.Status = CapabilityEnvelopeDeferred
		observation.Reason = "capability envelope reverse observation is deferred"
	case input.ReverseObservationDigest == "":
		observation.TargetStage = "reverse_observation"
		observation.Reason = "reverse observation digest is missing"
	case missingCapability != "":
		observation.TargetStage = "capability_set"
		observation.MissingCapability = missingCapability
		observation.Reason = "requested capability is absent from observed capability evidence"
	case input.ReverseObservationStatus == "UNKNOWN":
		observation.TargetStage = "reverse_observation_status"
		observation.Reason = "capability envelope reverse observation remains unresolved"
	case input.ReverseObservationStatus == "BOUND":
		observation.Status = CapabilityEnvelopeBound
		observation.Reason = "identity, plan, allowlist, capability, and reverse evidence are aligned"
	default:
		observation.TargetStage = "reverse_observation_status"
		observation.Reason = "reverse observation status is not recognized"
	}

	observation.ObservationDigest = capabilityEnvelopeDigest(
		string(observation.Status),
		observation.SourceVersion,
		observation.ContractVersion,
		string(observation.WorkloadIdentityStatus),
		observation.EvidencePrefixDigest,
		observation.PlanDigest,
		observation.NetworkAllowlistDigest,
		strings.Join(canonicalCapabilities(observation.RequestedCapabilities), ","),
		strings.Join(canonicalCapabilities(observation.ObservedCapabilities), ","),
		observation.MissingCapability,
		input.ReverseObservationDigest,
		observation.TargetStage,
		observation.Reason,
	)
	return observation
}

func firstMissingCapability(requested, observed []string) string {
	observedSet := make(map[string]struct{}, len(observed))
	for _, capability := range observed {
		observedSet[capability] = struct{}{}
	}
	for _, capability := range requested {
		if _, ok := observedSet[capability]; !ok {
			return capability
		}
	}
	return ""
}

func canonicalCapabilities(capabilities []string) []string {
	canonical := append([]string(nil), capabilities...)
	sort.Strings(canonical)
	return canonical
}

func capabilityEnvelopeDigest(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}
