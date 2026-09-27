package jevcal

import (
	"math"
	"testing"
)

func TestValidateChoicesRejectsInvalidProbabilitySet(t *testing.T) {
	if ValidateChoices([]Choice{{ID: "a", Probability: 0.7}, {ID: "b", Probability: 0.2}}) {
		t.Fatal("expected probabilities that do not sum to one to be rejected")
	}
	if ValidateChoices([]Choice{{ID: "a", Probability: 0.5}, {ID: "a", Probability: 0.5}}) {
		t.Fatal("expected duplicate choice IDs to be rejected")
	}
}

func TestBrierScoreUsesObservedOutcome(t *testing.T) {
	choices := []Choice{{ID: "a", Probability: 0.75}, {ID: "b", Probability: 0.25}}
	score, ok := BrierScore(choices, []Outcome{{ChoiceID: "a", Observed: true}})
	if !ok {
		t.Fatal("expected a valid score")
	}
	expected := (math.Pow(0.75-1, 2) + math.Pow(0.25, 2)) / 2
	if math.Abs(score-expected) > 1e-12 {
		t.Fatalf("score = %v, want %v", score, expected)
	}
}

func TestChoiceSetDigestPreservesOrder(t *testing.T) {
	first := []Choice{{ID: "a", Probability: 0.75}, {ID: "b", Probability: 0.25}}
	second := []Choice{{ID: "b", Probability: 0.25}, {ID: "a", Probability: 0.75}}
	if ChoiceSetDigest(first) == ChoiceSetDigest(second) {
		t.Fatal("expected ordered choice sets to have different digests")
	}
}