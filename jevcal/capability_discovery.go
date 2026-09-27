package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

type CapabilityDiscoveryStatus string

const (
	CapabilityDiscoveryBound    CapabilityDiscoveryStatus = "BOUND"
	CapabilityDiscoveryDeferred CapabilityDiscoveryStatus = "DEFERRED"
	CapabilityDiscoveryUnknown  CapabilityDiscoveryStatus = "UNKNOWN"
)

// CapabilityDiscoveryInput is a natural-language query over the read-only
// capability catalog. It never contains execution or authorization input.
type CapabilityDiscoveryInput struct {
	SourceVersion   string
	ContractVersion string
	Query           string
}

type CapabilityDiscoveryMatch struct {
	Key     string
	Summary string
	Stage   string
	Status  CapabilityDiscoveryStatus
}

// CapabilityDiscoveryObservation explains what this runtime can recognize
// without claiming that a recognized capability is executable or authorized.
type CapabilityDiscoveryObservation struct {
	Status             CapabilityDiscoveryStatus
	SourceVersion      string
	ContractVersion    string
	Query              string
	Matches            []CapabilityDiscoveryMatch
	Suggestions        []string
	FirstMismatch      string
	MissingStage       string
	TargetStage        string
	Reason             string
	DiscoveryDigest    string
	IsReadOnly         bool
	CanExecute         bool
	CanAuthorize       bool
}

type capabilityCatalogEntry struct {
	Key     string
	Summary string
	Stage   string
	Aliases []string
	Safe    bool
}

var capabilityCatalog = []capabilityCatalogEntry{
	{Key: "feedback-trend", Summary: "compare feedback and calibration windows with evidence lineage", Stage: "feedback_observation", Aliases: []string{"feedback", "trend", "calibration", "피드백", "추세"}, Safe: true},
	{Key: "generation", Summary: "produce and inspect generated artifacts with provenance", Stage: "generation_observation", Aliases: []string{"generate", "generation", "codegen", "code generation", "생성", "코드 생성"}, Safe: true},
	{Key: "provenance", Summary: "trace source, generation, and reverse-observation evidence", Stage: "provenance_observation", Aliases: []string{"provenance", "origin", "reverse observation", "기원", "역관찰"}, Safe: true},
	{Key: "security-boundary", Summary: "observe workload identity and network capability boundaries", Stage: "security_capability_boundary", Aliases: []string{"security", "spiffe", "workload identity", "network allowlist", "credential", "보안"}, Safe: false},
	{Key: "support-triage", Summary: "structure support-triage workflows and their next safe observations", Stage: "support_triage_observation", Aliases: []string{"support", "support triage", "triage", "workflow", "지원", "분류"}, Safe: true},
	{Key: "syntax-completion", Summary: "suggest syntax completions through a read-only language-service path", Stage: "syntax_completion", Aliases: []string{"syntax", "completion", "autocomplete", "lsp", "문법", "완성"}, Safe: true},
	{Key: "execution-boundary", Summary: "execution or authorization requires an explicit external boundary", Stage: "execution_or_authorization_boundary", Aliases: []string{"execute", "execution", "run", "authorize", "authorization", "permission", "실행", "권한"}, Safe: false},
}

// CapabilityDiscoveryCatalog returns a stable copy of the catalog exposed to
// users. Catalog entries describe observation surfaces, not permissions.
func CapabilityDiscoveryCatalog() []CapabilityDiscoveryMatch {
	matches := make([]CapabilityDiscoveryMatch, 0, len(capabilityCatalog))
	for _, entry := range capabilityCatalog {
		matches = append(matches, CapabilityDiscoveryMatch{
			Key:     entry.Key,
			Summary: entry.Summary,
			Stage:   entry.Stage,
			Status:  catalogStatus(entry),
		})
	}
	sort.Slice(matches, func(i, j int) bool { return matches[i].Key < matches[j].Key })
	return matches
}

func catalogStatus(entry capabilityCatalogEntry) CapabilityDiscoveryStatus {
	if entry.Safe {
		return CapabilityDiscoveryBound
	}
	return CapabilityDiscoveryDeferred
}

// DiscoverCapabilities resolves a natural-language query against the stable
// catalog. Unknown and deferred results preserve their first boundary.
func DiscoverCapabilities(input CapabilityDiscoveryInput) CapabilityDiscoveryObservation {
	observation := CapabilityDiscoveryObservation{
		Status:          CapabilityDiscoveryUnknown,
		SourceVersion:   input.SourceVersion,
		ContractVersion: input.ContractVersion,
		Query:           strings.TrimSpace(input.Query),
		TargetStage:     "capability_discovery",
		Reason:          "query did not resolve to a known capability",
		IsReadOnly:      true,
		CanExecute:      false,
		CanAuthorize:    false,
	}

	switch {
	case strings.TrimSpace(input.SourceVersion) == "" || strings.TrimSpace(input.ContractVersion) == "":
		observation.TargetStage = "identity"
		observation.FirstMismatch = "identity"
		observation.MissingStage = "source_or_contract_identity"
		observation.Reason = "source and contract identities are required for capability discovery"
	case observation.Query == "":
		observation.TargetStage = "query"
		observation.FirstMismatch = "query"
		observation.MissingStage = "capability_query"
		observation.Reason = "a natural-language capability query is required"
	default:
		observation.Matches = discoverMatches(observation.Query)
		if len(observation.Matches) == 0 {
			observation.FirstMismatch = "query"
			observation.MissingStage = "capability_catalog"
			observation.Reason = "the query does not match a capability in the read-only catalog"
		} else if hasDeferredMatch(observation.Matches) {
			observation.Status = CapabilityDiscoveryDeferred
			observation.FirstMismatch = "execution_or_authorization_boundary"
			observation.MissingStage = "explicit_external_boundary"
			observation.Reason = "the query reached a capability boundary that discovery cannot execute or authorize"
		} else {
			observation.Status = CapabilityDiscoveryBound
			observation.FirstMismatch = ""
			observation.MissingStage = ""
			observation.Reason = "the query matched read-only language capabilities"
		}
	}

	observation.Suggestions = capabilitySuggestions(observation.Query, observation.Matches)
	observation.DiscoveryDigest = capabilityDiscoveryDigest(observation)
	return observation
}

func discoverMatches(query string) []CapabilityDiscoveryMatch {
	normalized := strings.ToLower(strings.TrimSpace(query))
	matches := make([]CapabilityDiscoveryMatch, 0)
	if isOverviewQuery(normalized) {
		for _, entry := range capabilityCatalog {
			if entry.Safe {
				matches = append(matches, matchForEntry(entry))
			}
		}
	} else {
		for _, entry := range capabilityCatalog {
			for _, alias := range entry.Aliases {
				if strings.Contains(normalized, strings.ToLower(alias)) {
					matches = append(matches, matchForEntry(entry))
					break
				}
			}
		}
	}
	sort.Slice(matches, func(i, j int) bool { return matches[i].Key < matches[j].Key })
	return uniqueMatches(matches)
}

func matchForEntry(entry capabilityCatalogEntry) CapabilityDiscoveryMatch {
	return CapabilityDiscoveryMatch{Key: entry.Key, Summary: entry.Summary, Stage: entry.Stage, Status: catalogStatus(entry)}
}

func uniqueMatches(matches []CapabilityDiscoveryMatch) []CapabilityDiscoveryMatch {
	unique := make([]CapabilityDiscoveryMatch, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		if _, ok := seen[match.Key]; ok {
			continue
		}
		seen[match.Key] = struct{}{}
		unique = append(unique, match)
	}
	return unique
}

func hasDeferredMatch(matches []CapabilityDiscoveryMatch) bool {
	for _, match := range matches {
		if match.Status == CapabilityDiscoveryDeferred {
			return true
		}
	}
	return false
}

func isOverviewQuery(query string) bool {
	for _, phrase := range []string{"what can", "capabilities", "show examples", "discover", "help", "gooo", "무엇을", "가능", "할 수"} {
		if strings.Contains(query, phrase) {
			return true
		}
	}
	return false
}

func capabilitySuggestions(query string, matches []CapabilityDiscoveryMatch) []string {
	if len(matches) > 0 {
		suggestions := make([]string, 0, len(matches))
		for _, match := range matches {
			suggestions = append(suggestions, match.Key)
		}
		return suggestions
	}
	if strings.TrimSpace(query) == "" {
		return []string{"syntax-completion", "generation", "provenance"}
	}
	return []string{"syntax-completion", "generation", "provenance", "feedback-trend", "support-triage"}
}

func (observation CapabilityDiscoveryObservation) Validate() error {
	if observation.Status != CapabilityDiscoveryBound && observation.Status != CapabilityDiscoveryDeferred && observation.Status != CapabilityDiscoveryUnknown {
		return fmt.Errorf("invalid capability discovery status %q", observation.Status)
	}
	if strings.TrimSpace(observation.SourceVersion) == "" || strings.TrimSpace(observation.ContractVersion) == "" {
		return fmt.Errorf("capability discovery identity is missing")
	}
	if !observation.IsReadOnly || observation.CanExecute || observation.CanAuthorize {
		return fmt.Errorf("capability discovery crossed an execution or authorization boundary")
	}
	if observation.Status == CapabilityDiscoveryBound && (len(observation.Matches) == 0 || observation.FirstMismatch != "" || observation.MissingStage != "") {
		return fmt.Errorf("bound capability discovery is incomplete")
	}
	if observation.Status != CapabilityDiscoveryBound && (observation.FirstMismatch == "" || observation.MissingStage == "") {
		return fmt.Errorf("unresolved capability discovery lost its first boundary")
	}
	if observation.DiscoveryDigest != capabilityDiscoveryDigest(observation) {
		return fmt.Errorf("capability discovery digest mismatch")
	}
	return nil
}

func capabilityDiscoveryDigest(observation CapabilityDiscoveryObservation) string {
	parts := []string{
		"jev-capability-discovery",
		observation.SourceVersion,
		observation.ContractVersion,
		strings.ToLower(strings.TrimSpace(observation.Query)),
		string(observation.Status),
		observation.FirstMismatch,
		observation.MissingStage,
		observation.TargetStage,
		observation.Reason,
	}
	for _, match := range observation.Matches {
		parts = append(parts, match.Key, match.Summary, match.Stage, string(match.Status))
	}
	parts = append(parts, observation.Suggestions...)
	digest := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return "sha256:" + hex.EncodeToString(digest[:])
}
