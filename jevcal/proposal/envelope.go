package proposal

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const EnvelopeSchema = "jev.self-improvement-proposal.v1"

type Envelope struct {
	Schema               string   `json:"schema"`
	Status               Status   `json:"status"`
	DomainID             string   `json:"domain_id"`
	MissingCapabilityIDs []string `json:"missing_capability_ids"`
	ObservedSignals      []string `json:"observed_signals"`
	Proposal             string   `json:"proposal"`
	NextQuestion         string   `json:"next_question"`
	EvidenceDigests      []string `json:"evidence_digests"`
	SourceDigest         string   `json:"source_digest"`
	ContractDigest       string   `json:"contract_digest"`
	NonExecuting         bool     `json:"non_executing"`
	NonAuthorizing       bool     `json:"non_authorizing"`
}

// NewEnvelope preserves the proposal boundary for storage or later review.
func NewEnvelope(result Result) Envelope {
	return Envelope{
		Schema:               EnvelopeSchema,
		Status:               result.Status,
		DomainID:             result.DomainID,
		MissingCapabilityIDs: append([]string(nil), result.MissingCapabilityIDs...),
		ObservedSignals:      append([]string(nil), result.ObservedSignals...),
		Proposal:             result.Proposal,
		NextQuestion:         result.NextQuestion,
		EvidenceDigests:      append([]string(nil), result.EvidenceDigests...),
		SourceDigest:         result.SourceDigest,
		ContractDigest:       result.ContractDigest,
		NonExecuting:         true,
		NonAuthorizing:       true,
	}
}

func MarshalEnvelope(result Result) ([]byte, error) {
	return json.Marshal(NewEnvelope(result))
}

func EnvelopeDigest(result Result) (string, error) {
	payload, err := MarshalEnvelope(result)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}