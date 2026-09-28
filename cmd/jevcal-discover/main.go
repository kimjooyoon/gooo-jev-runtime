package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/gooo-jev-runtime/jevcal"
)

type request struct {
	SourceVersion   string `json:"source_version"`
	ContractVersion string `json:"contract_version"`
	Query           string `json:"query"`
	Declaration     string `json:"declaration,omitempty"`
}

type matchJSON struct {
	Key           string `json:"key"`
	Summary       string `json:"summary"`
	ExampleQuery  string `json:"example_query"`
	NextOperation string `json:"next_operation"`
	Stage         string `json:"stage"`
	Status        string `json:"status"`
}

type discoveryJSON struct {
	Status                     string      `json:"status"`
	SourceVersion              string      `json:"source_version"`
	ContractVersion            string      `json:"contract_version"`
	Query                      string      `json:"query"`
	NormalizedTerms              []string    `json:"normalized_terms"`
	MatchedTerms                 []string    `json:"matched_terms"`
	QueryDigest                  string      `json:"query_digest"`
	ProvenanceDigest             string      `json:"provenance_digest"`
	Matches                    []matchJSON `json:"matches"`
	Suggestions                []string    `json:"suggestions"`
	SuggestedQueries           []string    `json:"suggested_queries"`
	FirstMismatch              string      `json:"first_mismatch"`
	MissingStage               string      `json:"missing_stage"`
	TargetStage                string      `json:"target_stage"`
	Reason                     string      `json:"reason"`
	DiscoveryDigest            string      `json:"discovery_digest"`
	DeclarationSourceDigest    string      `json:"declaration_source_digest,omitempty"`
	DeclarationObservedSignals []string    `json:"declaration_observed_signals,omitempty"`
	DeclarationBound           bool        `json:"declaration_bound"`
	IsReadOnly                 bool        `json:"is_read_only"`
	CanExecute                 bool        `json:"can_execute"`
	CanAuthorize               bool        `json:"can_authorize"`
}

func render(observation jevcal.CapabilityDiscoveryObservation) discoveryJSON {
	matches := make([]matchJSON, 0, len(observation.Matches))
	for _, match := range observation.Matches {
		matches = append(matches, matchJSON{
			Key:           match.Key,
			Summary:       match.Summary,
			ExampleQuery:  match.ExampleQuery,
			NextOperation: match.NextOperation,
			Stage:         match.Stage,
			Status:        string(match.Status),
		})
	}
	return discoveryJSON{
		Status:                     string(observation.Status),
		SourceVersion:              observation.SourceVersion,
		ContractVersion:            observation.ContractVersion,
		Query:                      observation.Query,
		NormalizedTerms: observation.NormalizedTerms,
		MatchedTerms: observation.MatchedTerms,
		QueryDigest: observation.QueryDigest,
		ProvenanceDigest: observation.ProvenanceDigest,
		Matches:                    matches,
		Suggestions:                observation.Suggestions,
		SuggestedQueries:           observation.SuggestedQueries,
		FirstMismatch:              observation.FirstMismatch,
		MissingStage:               observation.MissingStage,
		TargetStage:                observation.TargetStage,
		Reason:                     observation.Reason,
		DiscoveryDigest:            observation.DiscoveryDigest,
		DeclarationSourceDigest:    observation.DeclarationSourceDigest,
		DeclarationObservedSignals: observation.DeclarationObservedSignals,
		DeclarationBound:           observation.DeclarationBound,
		IsReadOnly:                 observation.IsReadOnly,
		CanExecute:                 observation.CanExecute,
		CanAuthorize:               observation.CanAuthorize,
	}
}

func main() {
	var (
		queryFlag           string
		sourceVersionFlag   string
		contractVersionFlag string
		declarationFileFlag string
	)
	flag.StringVar(&queryFlag, "query", "", "natural-language capability query")
	flag.StringVar(&sourceVersionFlag, "source-version", "", "source identity for direct query mode")
	flag.StringVar(&contractVersionFlag, "contract-version", "", "contract identity for direct query mode")
	flag.StringVar(&declarationFileFlag, "declaration-file", "", "optional .gooo declaration file for direct query mode")
	flag.Parse()

	var input request
	directQuery := queryFlag != "" || sourceVersionFlag != "" || contractVersionFlag != "" || declarationFileFlag != ""
	if directQuery {
		input.SourceVersion = sourceVersionFlag
		input.ContractVersion = contractVersionFlag
		input.Query = queryFlag
		if declarationFileFlag != "" {
			declaration, err := os.ReadFile(declarationFileFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "read declaration file: %v\n", err)
				os.Exit(64)
			}
			input.Declaration = string(declaration)
		}
	} else if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "read capability discovery request: %v\n", err)
		os.Exit(64)
	}

	discoveryInput := jevcal.CapabilityDiscoveryInput{
		SourceVersion:   input.SourceVersion,
		ContractVersion: input.ContractVersion,
		Query:           input.Query,
	}
	var observation jevcal.CapabilityDiscoveryObservation
	if input.Declaration != "" {
		observation = jevcal.DiscoverCapabilitiesForDeclaration(discoveryInput, input.Declaration)
	} else {
		observation = jevcal.DiscoverCapabilities(discoveryInput)
	}
	if err := observation.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "invalid capability discovery observation: %v\n", err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(struct {
		Discovery discoveryJSON `json:"discovery"`
	}{Discovery: render(observation)}); err != nil {
		fmt.Fprintf(os.Stderr, "write capability discovery observation: %v\n", err)
		os.Exit(1)
	}
}
