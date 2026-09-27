package jevcal

// JEVSelfImprovementCandidateObservationStatus identifies whether the
// source-to-reverse-observation evidence chain is complete.
type JEVSelfImprovementCandidateObservationStatus string

const (
	JEVSelfImprovementCandidateBound   JEVSelfImprovementCandidateObservationStatus = "BOUND"
	JEVSelfImprovementCandidateUnknown JEVSelfImprovementCandidateObservationStatus = "UNKNOWN"
)

// JEVSelfImprovementCandidateObservation preserves every evidence boundary
// without claiming that a candidate improves the language.
type JEVSelfImprovementCandidateObservation struct {
	Status                    JEVSelfImprovementCandidateObservationStatus
	SourceDigest              string
	IRDigest                  string
	GeneratedDigest           string
	ReverseObservationDigest  string
	ExpectedDelta             float64
	ObservedDelta             float64
	MissingStageIndex         int
	IsReadOnly                bool
	CanExecute                bool
	CanAuthorize              bool
}

// ObserveJEVSelfImprovementCandidate binds source, IR, generated output, and
// reverse observation in order. Missing evidence remains UNKNOWN.
func ObserveJEVSelfImprovementCandidate(
	sourceDigest string,
	irDigest string,
	generatedDigest string,
	reverseObservationDigest string,
	expectedDelta float64,
	observedDelta float64,
) JEVSelfImprovementCandidateObservation {
	observation := JEVSelfImprovementCandidateObservation{
		Status:                   JEVSelfImprovementCandidateUnknown,
		SourceDigest:             sourceDigest,
		IRDigest:                 irDigest,
		GeneratedDigest:          generatedDigest,
		ReverseObservationDigest: reverseObservationDigest,
		ExpectedDelta:            expectedDelta,
		ObservedDelta:            observedDelta,
		MissingStageIndex:        -1,
		IsReadOnly:               true,
		CanExecute:               false,
		CanAuthorize:              false,
	}

	stages := [...]string{
		sourceDigest,
		irDigest,
		generatedDigest,
		reverseObservationDigest,
	}
	for index, digest := range stages {
		if digest == "" {
			observation.MissingStageIndex = index
			return observation
		}
	}

	observation.Status = JEVSelfImprovementCandidateBound
	return observation
}