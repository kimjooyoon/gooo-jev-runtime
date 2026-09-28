package query

import "encoding/json"

type LSPResponse struct {
	QueryDigest           string   `json:"query_digest"`
	Status                Status   `json:"status"`
	CandidateCapabilityIDs []string `json:"candidate_capability_ids"`
	UnresolvedTerms       []string `json:"unresolved_terms"`
	NextQuestion          string   `json:"next_question"`
	ProvenanceDigest      string   `json:"provenance_digest"`
	ContractDigest        string   `json:"contract_digest,omitempty"`
}

// NewLSPResponse maps discovery into a read-only editor response.
func NewLSPResponse(result Result, contractDigest string) LSPResponse {
	return LSPResponse{
		QueryDigest:            result.QueryDigest,
		Status:                 result.Status,
		CandidateCapabilityIDs: append([]string(nil), result.CandidateIDs...),
		UnresolvedTerms:        append([]string(nil), result.UnresolvedTerms...),
		NextQuestion:           result.NextQuestion,
		ProvenanceDigest:       result.ProvenanceDigest,
		ContractDigest:         contractDigest,
	}
}

func MarshalLSPResponse(result Result, contractDigest string) ([]byte, error) {
	return json.Marshal(NewLSPResponse(result, contractDigest))
}