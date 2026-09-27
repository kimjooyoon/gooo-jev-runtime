package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// DecisionRouteRecord is a provenance record for a read-only route projection.
type DecisionRouteRecord struct {
	Status         string
	Code           string
	SourceVersion  string
	ContractVersion string
	TargetStage    string
	DecisionDigest string
	RecordDigest   string
	EvidenceKind   string
	Reason         string
	IsReadOnly     bool
	CanExecute     bool
	CanAuthorize   bool
}

// RecordDecisionRouteProjection records a route without granting authority.
func RecordDecisionRouteProjection(projection DecisionRouteLSPProjection, sourceVersion, contractVersion string) DecisionRouteRecord {
	record := DecisionRouteRecord{
		Status:          projection.Status,
		Code:            projection.Code,
		SourceVersion:   sourceVersion,
		ContractVersion: contractVersion,
		TargetStage:     projection.TargetStage,
		DecisionDigest:  projection.DecisionDigest,
		EvidenceKind:    "jev.calibration.decision_route",
		Reason:          projection.Title,
		IsReadOnly:      true,
		CanExecute:      false,
		CanAuthorize:    false,
	}
	if sourceVersion == "" || contractVersion == "" || projection.DecisionDigest == "" {
		record.Status = "UNKNOWN"
		record.Code = "jev.route.unknown"
		record.Reason = "route record identity or decision evidence is missing"
	}
	record.RecordDigest = decisionRouteRecordDigest(
		record.Status,
		record.Code,
		record.SourceVersion,
		record.ContractVersion,
		record.TargetStage,
		record.DecisionDigest,
		record.EvidenceKind,
		record.Reason,
	)
	return record
}

func decisionRouteRecordDigest(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}