package jevcal

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

type Status string

const (
	StatusConverged Status = "CONVERGED"
	StatusRegressed Status = "REGRESSED"
	StatusDeferred  Status = "DEFERRED"
	StatusUnknown   Status = "UNKNOWN"
)

type Choice struct {
	ID          string
	Probability float64
}

type Outcome struct {
	ChoiceID string
	Observed bool
}

type Evidence struct {
	SourceDigest      string
	ObservationDigest string
	PrefixDigest      string
	Complete          bool
}

type Window struct {
	Choices  []Choice
	Outcomes []Outcome
	Evidence Evidence
	Score    float64
}

type Decision struct {
	Status               Status
	Reason               string
	Comparable           bool
	ChoiceSetDigest      string
	CurrentScore         float64
	BaselineScore        float64
	EvidencePrefixDigest string
}

func ChoiceSetDigest(choices []Choice) string {
	var b strings.Builder
	for _, c := range choices {
		b.WriteString(c.ID)
		b.WriteByte(0)
		b.WriteString(strconv.FormatFloat(c.Probability, 'g', 17, 64))
		b.WriteByte(0)
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func EvidenceDigest(e Evidence) string {
	if e.SourceDigest == "" && e.ObservationDigest == "" && e.PrefixDigest == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(e.SourceDigest + ":" + e.ObservationDigest + ":" + e.PrefixDigest))
	return hex.EncodeToString(sum[:])
}