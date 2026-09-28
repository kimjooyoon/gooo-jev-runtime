package query

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

// Status describes read-only capability discovery. It never authorizes or executes.
type Status string

const (
	StatusUnknown    Status = "UNKNOWN"
	StatusDeferred   Status = "DEFERRED"
	StatusBound      Status = "BOUND"
	StatusNeedsInput Status = "NEEDS_INPUT"
)

// Request is the deterministic boundary between natural-language collection and discovery.
// CandidateIDs and UnresolvedTerms must come from an external catalog/normalizer.
type Request struct {
	QueryText        string
	Terms            []string
	CandidateIDs     []string
	UnresolvedTerms  []string
	QueryDigest      string
	ProvenanceDigest string
}

// Result preserves ambiguity instead of treating an absent failure as success.
type Result struct {
	Status           Status
	NormalizedTerms  []string
	CandidateIDs     []string
	UnresolvedTerms  []string
	NextQuestion     string
	QueryDigest      string
	ProvenanceDigest string
}

// NormalizeTerms provides deterministic lexical normalization only. It does not infer intent.
func NormalizeTerms(terms []string) []string {
	seen := make(map[string]struct{}, len(terms))
	out := make([]string, 0, len(terms))
	for _, term := range terms {
		term = strings.ToLower(strings.TrimSpace(term))
		if term == "" {
			continue
		}
		if _, ok := seen[term]; ok {
			continue
		}
		seen[term] = struct{}{}
		out = append(out, term)
	}
	sort.Strings(out)
	return out
}

// DigestQuery binds the original text and normalized terms without executing anything.
func DigestQuery(queryText string, terms []string) string {
	normalized := NormalizeTerms(terms)
	var b strings.Builder
	b.WriteString(queryText)
	b.WriteByte(0)
	for _, term := range normalized {
		b.WriteString(term)
		b.WriteByte(0)
	}
	sum := sha256.Sum256([]byte(b.String()))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// Discover evaluates only the supplied discovery envelope and preserves unresolved boundaries.
func Discover(req Request) Result {
	result := Result{
		Status:           StatusUnknown,
		NormalizedTerms:  NormalizeTerms(req.Terms),
		CandidateIDs:     NormalizeTerms(req.CandidateIDs),
		UnresolvedTerms:  NormalizeTerms(req.UnresolvedTerms),
		QueryDigest:      req.QueryDigest,
		ProvenanceDigest: req.ProvenanceDigest,
	}
	if strings.TrimSpace(req.QueryText) == "" || req.QueryDigest == "" || req.ProvenanceDigest == "" {
		result.NextQuestion = "Provide the query and its source-bound evidence digests."
		return result
	}
	if len(result.UnresolvedTerms) > 0 {
		result.Status = StatusNeedsInput
		result.NextQuestion = "Clarify the unresolved terms before selecting a declared capability."
		return result
	}
	if len(result.CandidateIDs) == 0 {
		result.NextQuestion = "Which declared capability should be considered for this query?"
		return result
	}
	if len(result.CandidateIDs) == 1 {
		result.Status = StatusBound
		return result
	}
	result.Status = StatusDeferred
	result.NextQuestion = "Choose one declared capability from the deterministic shortlist."
	return result
}