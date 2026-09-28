package proposal

const (
	ContractModule = "jev_self_improvement_proposal"
	ContractInput  = "SelfImprovementProposal"
)

// GoooBinding is the explicit mapping from the runtime result to the .gooo contract.
// It carries values only; it does not generate, execute, or authorize a change.
type GoooBinding struct {
	Module               string
	Input                string
	DomainID             string
	SourceCapabilityIDs  []string
	MissingCapabilityIDs []string
	ObservedSignals      []string
	Proposal             string
	NextQuestion         string
	EvidenceDigests      []string
	SourceDigest         string
	ContractDigest       string
	Status               Status
}

func BindToGooo(result Result) GoooBinding {
	return GoooBinding{
		Module:               ContractModule,
		Input:                ContractInput,
		DomainID:             result.DomainID,
		MissingCapabilityIDs: append([]string(nil), result.MissingCapabilityIDs...),
		ObservedSignals:      append([]string(nil), result.ObservedSignals...),
		Proposal:             result.Proposal,
		NextQuestion:         result.NextQuestion,
		EvidenceDigests:      append([]string(nil), result.EvidenceDigests...),
		SourceDigest:         result.SourceDigest,
		ContractDigest:       result.ContractDigest,
		Status:               result.Status,
	}
}