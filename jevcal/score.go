package jevcal

import "math"

func ValidateChoices(choices []Choice) bool {
	if len(choices) == 0 {
		return false
	}
	seen := make(map[string]struct{}, len(choices))
	total := 0.0
	for _, choice := range choices {
		if choice.ID == "" || choice.Probability < 0 || choice.Probability > 1 {
			return false
		}
		if _, exists := seen[choice.ID]; exists {
			return false
		}
		seen[choice.ID] = struct{}{}
		total += choice.Probability
	}
	return math.Abs(total-1) <= 1e-9
}

func BrierScore(choices []Choice, outcomes []Outcome) (float64, bool) {
	if !ValidateChoices(choices) || len(outcomes) == 0 {
		return 0, false
	}
	known := make(map[string]struct{}, len(choices))
	for _, choice := range choices {
		known[choice.ID] = struct{}{}
	}
	total := 0.0
	for _, outcome := range outcomes {
		if _, exists := known[outcome.ChoiceID]; !exists {
			return 0, false
		}
		for _, choice := range choices {
			target := 0.0
			if choice.ID == outcome.ChoiceID && outcome.Observed {
				target = 1
			}
			delta := choice.Probability - target
			total += delta * delta
		}
	}
	return total / float64(len(outcomes)*len(choices)), true
}
