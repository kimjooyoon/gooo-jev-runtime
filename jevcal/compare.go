package jevcal

func Compare(current, baseline Window) Decision {
	currentDigest := ChoiceSetDigest(current.Choices)
	if currentDigest != ChoiceSetDigest(baseline.Choices) {
		return Decision{
			Status:               StatusDeferred,
			Reason:               "choice-set-or-order-changed",
			ChoiceSetDigest:      currentDigest,
			CurrentScore:         current.Score,
			BaselineScore:        baseline.Score,
			EvidencePrefixDigest: current.Evidence.PrefixDigest,
		}
	}
	if !current.Evidence.Complete || !baseline.Evidence.Complete {
		return Decision{
			Status:               StatusUnknown,
			Reason:               "incomplete-evidence",
			ChoiceSetDigest:      currentDigest,
			CurrentScore:         current.Score,
			BaselineScore:        baseline.Score,
			EvidencePrefixDigest: current.Evidence.PrefixDigest,
		}
	}
	if EvidenceDigest(current.Evidence) == "" || EvidenceDigest(baseline.Evidence) == "" {
		return Decision{
			Status:               StatusUnknown,
			Reason:               "missing-evidence-digest",
			ChoiceSetDigest:      currentDigest,
			CurrentScore:         current.Score,
			BaselineScore:        baseline.Score,
			EvidencePrefixDigest: current.Evidence.PrefixDigest,
		}
	}
	if current.Score < baseline.Score {
		return Decision{
			Status:               StatusConverged,
			Reason:               "score-improved",
			Comparable:           true,
			ChoiceSetDigest:      currentDigest,
			CurrentScore:         current.Score,
			BaselineScore:        baseline.Score,
			EvidencePrefixDigest: current.Evidence.PrefixDigest,
		}
	}
	if current.Score > baseline.Score {
		return Decision{
			Status:               StatusRegressed,
			Reason:               "score-regressed",
			Comparable:           true,
			ChoiceSetDigest:      currentDigest,
			CurrentScore:         current.Score,
			BaselineScore:        baseline.Score,
			EvidencePrefixDigest: current.Evidence.PrefixDigest,
		}
	}
	return Decision{
		Status:               StatusDeferred,
		Reason:               "score-flat",
		Comparable:           true,
		ChoiceSetDigest:      currentDigest,
		CurrentScore:         current.Score,
		BaselineScore:        baseline.Score,
		EvidencePrefixDigest: current.Evidence.PrefixDigest,
	}
}