package query

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/gooo-jev-runtime/jevcal"
)

// CapabilityDiscoveryFollowUpResponse is the read-only LSP projection of a
// capability discovery boundary. It keeps source evidence and never becomes
// an execution or authorization request.
type CapabilityDiscoveryFollowUpResponse struct {
	Status                   string   `json:"status"`
	Questions                []string `json:"questions"`
	FirstBoundary            string   `json:"first_boundary"`
	IsReadOnly               bool     `json:"is_read_only"`
	QueryDigest              string   `json:"query_digest"`
	ProvenanceDigest         string   `json:"provenance_digest"`
	DeclarationSourceDigest  string   `json:"declaration_source_digest,omitempty"`
	DiscoveryDigest          string   `json:"discovery_digest"`
}

func RenderCapabilityDiscoveryFollowUp(observation jevcal.CapabilityDiscoveryObservation) CapabilityDiscoveryFollowUpResponse {
	followUp := jevcal.SuggestCapabilityDiscoveryFollowUps(observation)
	return CapabilityDiscoveryFollowUpResponse{
		Status:                  string(followUp.Status),
		Questions:               followUp.Questions,
		FirstBoundary:           followUp.FirstBoundary,
		IsReadOnly:              followUp.IsReadOnly,
		QueryDigest:             observation.QueryDigest,
		ProvenanceDigest:        observation.ProvenanceDigest,
		DeclarationSourceDigest: observation.DeclarationSourceDigest,
		DiscoveryDigest:         observation.DiscoveryDigest,
	}
}

func (response CapabilityDiscoveryFollowUpResponse) Validate() error {
	if response.Status == "" {
		return fmt.Errorf("capability discovery follow-up status is missing")
	}
	if !response.IsReadOnly {
		return fmt.Errorf("capability discovery follow-up crossed the read-only boundary")
	}
	if strings.TrimSpace(response.FirstBoundary) == "" {
		return fmt.Errorf("capability discovery follow-up first boundary is missing")
	}
	if len(response.Questions) == 0 {
		return fmt.Errorf("capability discovery follow-up has no questions")
	}
	if strings.TrimSpace(response.QueryDigest) == "" || strings.TrimSpace(response.ProvenanceDigest) == "" || strings.TrimSpace(response.DiscoveryDigest) == "" {
		return fmt.Errorf("capability discovery follow-up lost evidence digests")
	}
	return nil
}
