package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

// ClosureStatus is intentionally narrower than a task outcome: it describes
// only whether provenance evidence closes across the declared stages.
type ClosureStatus string

const (
	ClosureBound    ClosureStatus = "BOUND"
	ClosureUnknown  ClosureStatus = "UNKNOWN"
	ClosureDeferred ClosureStatus = "DEFERRED"
)

// ClosureEvidence is a provider-neutral receipt for a .gooo execution plan.
type ClosureEvidence struct {
	SourceVersion            string
	ContractVersion          string
	Declaration              string
	IR                       string
	Generated                string
	Action                   string
	ReverseObservation      string
	EvidencePrefix           string
	ActionStatus             ClosureStatus
	ReverseObservationStatus ClosureStatus
	ActionDigest             string
	ReverseDigest            string
	MissingStageIndex        int
}

// ClosureObservation is safe to expose to tooling and metrics. It is not an
// instruction, authorization, or proof that the action ran.
type ClosureObservation struct {
	Status             ClosureStatus
	DeclarationDigest  string
	IRDigest           string
	GeneratedDigest    string
	ActionDigest       string
	ReverseDigest      string
	EvidencePrefixDigest string
	MissingStage       string
	MissingStageIndex  int
	Reason             string
	ObservationDigest  string
}

// ObserveClosure closes a receipt only when all stage material, producer states,
// and chained digests agree exactly.
func ObserveClosure(e ClosureEvidence) ClosureObservation {
	observation := ClosureObservation{
		Status:              ClosureUnknown,
		DeclarationDigest:   digest(e.Declaration),
		IRDigest:            digest(e.IR),
		GeneratedDigest:     digest(e.Generated),
		ActionDigest:        e.ActionDigest,
		ReverseDigest:       e.ReverseDigest,
		EvidencePrefixDigest: digest(e.EvidencePrefix),
		MissingStageIndex:   e.MissingStageIndex,
	}
	if e.MissingStageIndex >= 0 {
		observation.MissingStage = "missing_stage_index"
		observation.Reason = "missing stage evidence"
		return finalizeClosureObservation(observation)
	}
	if missing := missingClosureStage(e); missing != "" {
		observation.MissingStage = missing
		observation.Reason = "missing closure evidence"
		return finalizeClosureObservation(observation)
	}
	if e.ActionStatus != ClosureBound || e.ReverseObservationStatus != ClosureBound {
		observation.Status = ClosureDeferred
		observation.Reason = "producer state is not BOUND"
		return finalizeClosureObservation(observation)
	}

	expectedAction := digest(e.SourceVersion, e.ContractVersion, e.Declaration, e.IR, e.Generated, e.Action, e.EvidencePrefix, strconv.Itoa(e.MissingStageIndex))
	if e.ActionDigest != expectedAction {
		observation.Reason = "action digest mismatch"
		return finalizeClosureObservation(observation)
	}
	if e.ReverseDigest != digest(e.ActionDigest, e.ReverseObservation) {
		observation.Reason = "reverse observation digest mismatch"
		return finalizeClosureObservation(observation)
	}

	observation.Status = ClosureBound
	observation.Reason = "closure evidence is digest-bound"
	return finalizeClosureObservation(observation)
}

func missingClosureStage(e ClosureEvidence) string {
	stages := []struct {
		name  string
		value string
	}{
		{"source_version", e.SourceVersion},
		{"contract_version", e.ContractVersion},
		{"declaration", e.Declaration},
		{"ir", e.IR},
		{"generated", e.Generated},
		{"action", e.Action},
		{"reverse_observation", e.ReverseObservation},
		{"evidence_prefix", e.EvidencePrefix},
		{"action_status", string(e.ActionStatus)},
		{"reverse_observation_status", string(e.ReverseObservationStatus)},
		{"action_digest", e.ActionDigest},
		{"reverse_digest", e.ReverseDigest},
	}
	for _, stage := range stages {
		if stage.value == "" {
			return stage.name
		}
	}
	return ""
}

func finalizeClosureObservation(observation ClosureObservation) ClosureObservation {
	observation.ObservationDigest = digest(
		string(observation.Status),
		observation.DeclarationDigest,
		observation.IRDigest,
		observation.GeneratedDigest,
		observation.ActionDigest,
		observation.ReverseDigest,
		observation.EvidencePrefixDigest,
		observation.MissingStage,
		strconv.Itoa(observation.MissingStageIndex),
		observation.Reason,
	)
	return observation
}

func digest(parts ...string) string {
	var encoded strings.Builder
	for _, part := range parts {
		encoded.WriteString(strconv.Itoa(len(part)))
		encoded.WriteByte(':')
		encoded.WriteString(part)
		encoded.WriteByte('|')
	}
	sum := sha256.Sum256([]byte(encoded.String()))
	return hex.EncodeToString(sum[:])
}