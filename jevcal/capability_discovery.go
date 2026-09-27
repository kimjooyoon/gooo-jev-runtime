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
	Summary      string
	ExampleQuery   string
	NextOperation string
	Stage         string
	Status        CapabilityDiscoveryStatus
}

// CapabilityDiscoveryObservation explains what this runtime can recognize
// without claiming that a recognized capability is executable or authorized.
type CapabilityDiscoveryObservation struct {
	Status          CapabilityDiscoveryStatus
	SourceVersion   string
	ContractVersion string
	Query           string
	DeclarationSourceDigest    string
	DeclarationObservedSignals []string
	DeclarationBound            bool
	Matches         []CapabilityDiscoveryMatch
	Suggestions      []string
	SuggestedQueries []string
	FirstMismatch   string
	MissingStage    string
	TargetStage     string
	Reason          string
	DiscoveryDigest string
	IsReadOnly      bool
	CanExecute      bool
	CanAuthorize    bool
}

type capabilityCatalogEntry struct {
	Key     string
	Summary      string
	ExampleQuery   string
	NextOperation string
	Stage         string
	Aliases       []string
	Safe          bool
}

var capabilityCatalog = []capabilityCatalogEntry{
	{Key: "feedback-trend", NextOperation: "compare_feedback_window", ExampleQuery: "How has gooo feedback changed over time?", Summary: "compare feedback and calibration windows with evidence lineage", Stage: "feedback_observation", Aliases: []string{"feedback", "trend", "calibration", "피드백", "추세"}, Safe: true},
	{Key: "generation", NextOperation: "write_generated_declaration", ExampleQuery: "How do I generate a canonical .gooo declaration?", Summary: "produce and inspect generated artifacts with provenance", Stage: "generation_observation", Aliases: []string{"generate", "generation", "codegen", "code generation", "생성", "코드 생성"}, Safe: true},
	{Key: "provenance", NextOperation: "inspect_provenance_chain", ExampleQuery: "Where did this .gooo declaration come from?", Summary: "trace source, generation, and reverse-observation evidence", Stage: "provenance_observation", Aliases: []string{"provenance", "origin", "reverse observation", "기원", "역관찰"}, Safe: true},
	{Key: "security-boundary", NextOperation: "bind_external_security_evidence", ExampleQuery: "What external security boundary is required?", Summary: "observe workload identity and network capability boundaries", Stage: "security_capability_boundary", Aliases: []string{"security", "spiffe", "workload identity", "network allowlist", "credential", "보안"}, Safe: false},
	{Key: "support-triage", NextOperation: "inspect_support_route", ExampleQuery: "How should I triage this gooo support request?", Summary: "structure support-triage workflows and their next safe observations", Stage: "support_triage_observation", Aliases: []string{"support", "support triage", "triage", "workflow", "지원", "분류"}, Safe: true},
	{Key: "syntax-completion", NextOperation: "edit_declaration", ExampleQuery: "How do I complete a .gooo declaration?", Summary: "suggest syntax completions through a read-only language-service path", Stage: "syntax_completion", Aliases: []string{"syntax", "completion", "autocomplete", "lsp", "문법", "완성"}, Safe: true},
	{Key: "execution-boundary", NextOperation: "provide_explicit_external_boundary", ExampleQuery: "What explicit boundary is needed before execution?", Summary: "execution or authorization requires an explicit external boundary", Stage: "execution_or_authorization_boundary", Aliases: []string{"execute", "execution", "run", "authorize", "authorization", "permission", "실행", "권한"}, Safe: false},
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
	observation.SuggestedQueries = capabilityExampleSuggestions()
	observation.DiscoveryDigest = capabilityDiscoveryDigest(observation)
	return observation
}

// DiscoverCapabilitiesForDeclaration binds capability discovery to the
// supplied declaration bytes without treating the declaration as executable
// or authoritative. The observed signals are intentionally structural.
func DiscoverCapabilitiesForDeclaration(input CapabilityDiscoveryInput, declaration string) CapabilityDiscoveryObservation {
	observation := DiscoverCapabilities(input)
	raw := strings.TrimSpace(declaration)
	if raw != "" {
		observation.DeclarationBound = true
		observation.DeclarationSourceDigest = capabilityDeclarationDigest(raw)
		observation.DeclarationObservedSignals = capabilityDeclarationSignals(raw)
	}
	observation.DiscoveryDigest = capabilityDiscoveryDigest(observation)
	return observation
}

func capabilityDeclarationSignals(declaration string) []string {
	signals := make([]string, 0)
	for _, line := range strings.Split(declaration, "\n") {
		normalized := strings.ToLower(strings.TrimSpace(line))
		for _, candidate := range []struct {
			prefix string
			id     string
		}{
			{prefix: "entity ", id: "entity"},
			{prefix: "operation ", id: "operation"},
			{prefix: "observe ", id: "observe"},
			{prefix: "transform ", id: "transform"},
			{prefix: "contract ", id: "contract"},
			{prefix: "policy ", id: "policy"},
			{prefix: "workflow ", id: "workflow"},
		} {
			if !strings.HasPrefix(normalized, candidate.prefix) || containsString(signals, candidate.id) {
				continue
			}
			signals = append(signals, candidate.id)
		}
	}
	sort.Strings(signals)
	return signals
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func capabilityDeclarationDigest(declaration string) string {
	digest := sha256.Sum256([]byte("jev-capability-declaration|" + declaration))
	return "sha256:" + hex.EncodeToString(digest[:])
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
	return CapabilityDiscoveryMatch{Key: entry.Key, Summary: entry.Summary, ExampleQuery: entry.ExampleQuery, NextOperation: entry.NextOperation, Stage: entry.Stage, Status: catalogStatus(entry)}
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
	for _, phrase := range []string{"what can", "capabilities", "show examples", "discover", "help", "무엇을", "가능", "할 수"} {
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

func capabilityExampleSuggestions() []string {
	suggestions := make([]string, 0)
	for _, entry := range capabilityCatalog {
		if entry.Safe && strings.TrimSpace(entry.ExampleQuery) != "" {
			suggestions = append(suggestions, entry.ExampleQuery)
		}
	}
	sort.Strings(suggestions)
	return suggestions
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
	if len(observation.SuggestedQueries) == 0 {
		return fmt.Errorf("capability discovery has no natural-language follow-up suggestions")
	}
	if observation.DeclarationBound != (strings.TrimSpace(observation.DeclarationSourceDigest) != "") {
		return fmt.Errorf("capability declaration binding is incomplete")
	}
	if observation.DeclarationSourceDigest != "" && !capabilityFeedbackDigestValid(observation.DeclarationSourceDigest) {
		return fmt.Errorf("capability declaration source digest is invalid")
	}
	for index, signal := range observation.DeclarationObservedSignals {
		if strings.TrimSpace(signal) == "" || (index > 0 && observation.DeclarationObservedSignals[index-1] >= signal) {
			return fmt.Errorf("capability declaration signals are not sorted and unique")
		}
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
		parts = append(parts, match.Key, match.Summary, match.ExampleQuery, match.NextOperation, match.Stage, string(match.Status))
	}
	parts = append(parts, observation.Suggestions...)
	parts = append(parts, observation.SuggestedQueries...)
	if observation.DeclarationBound {
		parts = append(parts, "declaration", observation.DeclarationSourceDigest)
		parts = append(parts, observation.DeclarationObservedSignals...)
	}
	digest := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return "sha256:" + hex.EncodeToString(digest[:])
}
