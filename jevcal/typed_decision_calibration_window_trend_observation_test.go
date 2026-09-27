package jevcal

import (
	"strings"
	"testing"
)

func testTypedDecisionCalibrationWindowTrendWindow(mean, coverage float64, prefix string) TypedDecisionCalibrationWindowObservation {
	observation := TypedDecisionCalibrationWindowObservation{
		Status:                TypedDecisionCalibrationWindowBound,
		ObservationCount:     2,
		KnownObservationCount: 2,
		WithinToleranceCount: 2,
		MeanAbsoluteError:    mean,
		EvidenceCoverage:     coverage,
		MinimumWindow:        1,
		EvidencePrefixDigest: prefix,
		IsReadOnly:            true,
		CanExecute:            false,
		CanAuthorize:          false,
	}
	observation.EvidenceDigest = typedDecisionCalibrationWindowEvidenceDigest(observation)
	return observation
}

func TestObserveTypedDecisionCalibrationWindowTrendBindsStableLineage(t *testing.T) {
	prefix := typedDecisionCalibrationWindowEvidencePrefixDigest([]string{"sha256:" + strings.Repeat("a", 64)})
	observation := ObserveTypedDecisionCalibrationWindowTrend(TypedDecisionCalibrationWindowTrendInput{
		Previous:   testTypedDecisionCalibrationWindowTrendWindow(0.4, 1, prefix),
		Current:    testTypedDecisionCalibrationWindowTrendWindow(0.2, 1, prefix),
		ComparedAt: "2026-09-28T00:00:00Z",
	})
	if err := observation.Validate(); err != nil {
		t.Fatalf("validate trend: %v", err)
	}
	if observation.Status != TypedDecisionCalibrationWindowTrendBound ||
		observation.Direction != TypedDecisionCalibrationWindowTrendLowerError ||
		observation.MeanErrorDelta != -0.2 {
		t.Fatalf("unexpected trend: %#v", observation)
	}
}

func TestObserveTypedDecisionCalibrationWindowTrendLowersResolutionOnPrefixChange(t *testing.T) {
	previousPrefix := typedDecisionCalibrationWindowEvidencePrefixDigest([]string{"sha256:" + strings.Repeat("a", 64)})
	currentPrefix := typedDecisionCalibrationWindowEvidencePrefixDigest([]string{"sha256:" + strings.Repeat("b", 64)})
	observation := ObserveTypedDecisionCalibrationWindowTrend(TypedDecisionCalibrationWindowTrendInput{
		Previous:   testTypedDecisionCalibrationWindowTrendWindow(0.4, 1, previousPrefix),
		Current:    testTypedDecisionCalibrationWindowTrendWindow(0.2, 1, currentPrefix),
		ComparedAt: "2026-09-28T00:00:00Z",
	})
	if err := observation.Validate(); err != nil {
		t.Fatalf("validate unknown trend: %v", err)
	}
	if observation.Status != TypedDecisionCalibrationWindowTrendUnknown ||
		observation.FirstMismatch != "evidence-prefix" {
		t.Fatalf("expected unknown prefix transition: %#v", observation)
	}
}

