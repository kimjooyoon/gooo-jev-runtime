package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strings"
)

const (
	TypedDecisionQuestionChoice = "choice"
	TypedDecisionQuestionScore  = "score"
	TypedDecisionQuestionNoul   = "noul"

	TypedDecisionSignalBound  DecisionRouteReverseObservationStatus = "BOUND"
	TypedDecisionSignalUnknown DecisionRouteReverseObservationStatus = "UNKNOWN"
)

type TypedDecisionSignalInput struct {
	SourceVersion       string
	ContractVersion     string
	ModelRevision       string
	QuestionID          string
	RequestDigest       string
	QuestionKind        string
	SelectedValue       string
	Probabilities       map[string]float64
	SelectedProbability float64
	Confidence          float64
	AcceptanceThreshold float64
	NonAuthorizing      bool
}

// TypedDecisionSignalReceipt records a structured JEV signal without granting
// the signal authority to execute or authorize a route.
type TypedDecisionSignalReceipt struct {
	Status              DecisionRouteReverseObservationStatus
	SourceVersion       string
	ContractVersion     string
	ModelRevision       string
	QuestionID          string
	RequestDigest       string
	QuestionKind        string
	SelectedValue       string
	ProbabilityDigest   string
	ProbabilityCount    int
	SelectedProbability float64
	Confidence          float64
	AcceptanceThreshold float64
	ThresholdMet        bool
	ReviewRequired      bool
	FirstMismatch       string
	EvidenceDigest      string
	NonExecuting        bool
	NonAuthorizing      bool
	CanExecute          bool
	CanAuthorize        bool
}

// ObserveTypedDecisionSignal validates a typed JEV signal and preserves its
// uncertainty. A threshold is a candidate gate, never an execution grant.
func ObserveTypedDecisionSignal(input TypedDecisionSignalInput) TypedDecisionSignalReceipt {
	receipt := TypedDecisionSignalReceipt{
		Status:              TypedDecisionSignalUnknown,
		SourceVersion:       input.SourceVersion,
		ContractVersion:     input.ContractVersion,
		ModelRevision:       input.ModelRevision,
		QuestionID:          input.QuestionID,
		RequestDigest:       input.RequestDigest,
		QuestionKind:        input.QuestionKind,
		SelectedValue:       input.SelectedValue,
		ProbabilityDigest:   typedDecisionProbabilityDigest(input.Probabilities),
		ProbabilityCount:    len(input.Probabilities),
		SelectedProbability: input.SelectedProbability,
		Confidence:          input.Confidence,
		AcceptanceThreshold: input.AcceptanceThreshold,
		ReviewRequired:      true,
		NonExecuting:        true,
		NonAuthorizing:      true,
		CanExecute:          false,
		CanAuthorize:        false,
	}

	switch {
	case !input.NonAuthorizing:
		receipt.FirstMismatch = "authorization-boundary"
	case strings.TrimSpace(input.SourceVersion) == "" || strings.TrimSpace(input.ContractVersion) == "":
		receipt.FirstMismatch = "identity"
	case strings.TrimSpace(input.ModelRevision) == "":
		receipt.FirstMismatch = "model-revision"
	case strings.TrimSpace(input.QuestionID) == "":
		receipt.FirstMismatch = "question-id"
	case !typedDecisionSignalDigestValid(input.RequestDigest):
		receipt.FirstMismatch = "request-digest"
	case !validTypedDecisionQuestionKind(input.QuestionKind):
		receipt.FirstMismatch = "question-kind"
	case strings.TrimSpace(input.SelectedValue) == "":
		receipt.FirstMismatch = "selected-value"
	case len(input.Probabilities) == 0:
		receipt.FirstMismatch = "probabilities"
	case math.IsNaN(input.SelectedProbability) || math.IsInf(input.SelectedProbability, 0) ||
		input.SelectedProbability < 0 || input.SelectedProbability > 1:
		receipt.FirstMismatch = "selected-probability"
	case math.IsNaN(input.Confidence) || math.IsInf(input.Confidence, 0) ||
		input.Confidence < 0 || input.Confidence > 1:
		receipt.FirstMismatch = "confidence"
	case math.IsNaN(input.AcceptanceThreshold) || math.IsInf(input.AcceptanceThreshold, 0) ||
		input.AcceptanceThreshold < 0 || input.AcceptanceThreshold > 1:
		receipt.FirstMismatch = "acceptance-threshold"
	default:
		total := 0.0
		selectedFound := false
		selectedActual := 0.0
		keys := make([]string, 0, len(input.Probabilities))
		for key := range input.Probabilities {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			probability := input.Probabilities[key]
			if math.IsNaN(probability) || math.IsInf(probability, 0) ||
				probability < 0 || probability > 1 {
				receipt.FirstMismatch = "probability-range"
				break
			}
			total += probability
			if key == input.SelectedValue {
				selectedFound = true
				selectedActual = probability
			}
		}
		switch {
		case receipt.FirstMismatch != "":
		case math.Abs(total-1) > 0.001:
			receipt.FirstMismatch = "probability-normalization"
		case !selectedFound:
			receipt.FirstMismatch = "selected-value"
		case math.Abs(selectedActual-input.SelectedProbability) > 0.000000001:
			receipt.FirstMismatch = "selected-probability"
		default:
			receipt.Status = TypedDecisionSignalBound
			receipt.ThresholdMet = input.SelectedProbability >= input.AcceptanceThreshold
			receipt.ReviewRequired = !receipt.ThresholdMet
		}
	}

	receipt.EvidenceDigest = typedDecisionSignalEvidenceDigest(receipt)
	return receipt
}

// Validate checks the receipt's own integrity and the non-authorizing boundary.
// It does not turn threshold satisfaction into permission.
func (receipt TypedDecisionSignalReceipt) Validate() error {
	if receipt.Status != TypedDecisionSignalBound && receipt.Status != TypedDecisionSignalUnknown {
		return fmt.Errorf("invalid typed decision signal status %q", receipt.Status)
	}
	if !receipt.NonExecuting || !receipt.NonAuthorizing || receipt.CanExecute || receipt.CanAuthorize {
		return fmt.Errorf("typed decision signal crossed an execution or authorization boundary")
	}
	if receipt.EvidenceDigest != typedDecisionSignalEvidenceDigest(receipt) {
		return fmt.Errorf("typed decision signal evidence digest mismatch")
	}
	if receipt.Status == TypedDecisionSignalUnknown {
		if receipt.FirstMismatch == "" {
			return fmt.Errorf("unknown typed decision signal lost its first mismatch")
		}
		return nil
	}
	if !typedDecisionSignalDigestValid(receipt.RequestDigest) ||
		!typedDecisionSignalDigestValid(receipt.ProbabilityDigest) ||
		receipt.FirstMismatch != "" ||
		receipt.ProbabilityCount == 0 ||
		receipt.ThresholdMet == receipt.ReviewRequired ||
		math.IsNaN(receipt.SelectedProbability) ||
		math.IsInf(receipt.SelectedProbability, 0) ||
		math.IsNaN(receipt.Confidence) ||
		math.IsInf(receipt.Confidence, 0) ||
		math.IsNaN(receipt.AcceptanceThreshold) ||
		math.IsInf(receipt.AcceptanceThreshold, 0) ||
		receipt.SelectedProbability < 0 || receipt.SelectedProbability > 1 ||
		receipt.Confidence < 0 || receipt.Confidence > 1 ||
		receipt.AcceptanceThreshold < 0 || receipt.AcceptanceThreshold > 1 {
		return fmt.Errorf("bound typed decision signal is incomplete")
	}
	return nil
}

func validTypedDecisionQuestionKind(kind string) bool {
	switch kind {
	case TypedDecisionQuestionChoice, TypedDecisionQuestionScore, TypedDecisionQuestionNoul:
		return true
	default:
		return false
	}
}

func typedDecisionSignalDigestValid(value string) bool {
	return strings.HasPrefix(value, "sha256:") && len(value) == len("sha256:")+64
}

func typedDecisionProbabilityDigest(probabilities map[string]float64) string {
	keys := make([]string, 0, len(probabilities))
	for key := range probabilities {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%0.9f", key, probabilities[key]))
	}
	return typedDecisionHash(strings.Join(parts, "|"))
}

func typedDecisionSignalEvidenceDigest(receipt TypedDecisionSignalReceipt) string {
	return typedDecisionHash(fmt.Sprintf(
		"jev-typed-decision-signal|%s|%s|%s|%s|%s|%s|%s|%s|%d|%0.9f|%0.9f|%0.9f|%t|%t|%s|%t|%t|%t|%t",
		receipt.Status,
		receipt.SourceVersion,
		receipt.ContractVersion,
		receipt.ModelRevision,
		receipt.QuestionID,
		receipt.RequestDigest,
		receipt.QuestionKind,
		receipt.SelectedValue,
		receipt.ProbabilityCount,
		receipt.SelectedProbability,
		receipt.Confidence,
		receipt.AcceptanceThreshold,
		receipt.ThresholdMet,
		receipt.ReviewRequired,
		receipt.FirstMismatch,
		receipt.NonExecuting,
		receipt.NonAuthorizing,
		receipt.CanExecute,
		receipt.CanAuthorize,
	))
}

func typedDecisionHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}

