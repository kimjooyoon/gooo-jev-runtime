package jevcal

// JEVWorkloadIdentityObservationStatus identifies whether a workload identity
// observation contains enough evidence to be inspected safely.
type JEVWorkloadIdentityObservationStatus string

const (
	JEVWorkloadIdentityBound   JEVWorkloadIdentityObservationStatus = "BOUND"
	JEVWorkloadIdentityUnknown JEVWorkloadIdentityObservationStatus = "UNKNOWN"
)

// JEVWorkloadIdentityObservation is a read-only record for a SPIFFE-like
// identity boundary. It is not an authorization decision.
type JEVWorkloadIdentityObservation struct {
	Status               JEVWorkloadIdentityObservationStatus
	SpiffeID             string
	Audience             string
	Capabilities         []string
	EvidencePrefixDigest string
	IsReadOnly           bool
	CanExecute           bool
	CanAuthorize         bool
}

// ObserveJEVWorkloadIdentity records identity evidence without granting any
// capability. Missing identity evidence remains UNKNOWN.
func ObserveJEVWorkloadIdentity(
	spiffeID string,
	audience string,
	capabilities []string,
	evidencePrefixDigest string,
) JEVWorkloadIdentityObservation {
	observation := JEVWorkloadIdentityObservation{
		Status:               JEVWorkloadIdentityUnknown,
		SpiffeID:             spiffeID,
		Audience:             audience,
		Capabilities:         append([]string(nil), capabilities...),
		EvidencePrefixDigest: evidencePrefixDigest,
		IsReadOnly:           true,
		CanExecute:           false,
		CanAuthorize:         false,
	}

	if spiffeID == "" || audience == "" || evidencePrefixDigest == "" {
		return observation
	}

	observation.Status = JEVWorkloadIdentityBound
	return observation
}