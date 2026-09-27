package jevcal

type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
)

type Diagnostic struct {
	Code                 string
	Severity             Severity
	Message              string
	EvidencePrefixDigest string
}

func ProjectDecision(decision Decision) Diagnostic {
	switch decision.Status {
	case StatusConverged:
		return Diagnostic{
			Code:                 "JEV_CONVERGED",
			Severity:             SeverityInfo,
			Message:              "calibration score improved on comparable evidence",
			EvidencePrefixDigest: decision.EvidencePrefixDigest,
		}
	case StatusRegressed:
		return Diagnostic{
			Code:                 "JEV_REGRESSED",
			Severity:             SeverityWarning,
			Message:              "calibration score regressed on comparable evidence",
			EvidencePrefixDigest: decision.EvidencePrefixDigest,
		}
	case StatusDeferred:
		return Diagnostic{
			Code:                 "JEV_DEFERRED",
			Severity:             SeverityInfo,
			Message:              "calibration comparison is deferred until evidence is comparable",
			EvidencePrefixDigest: decision.EvidencePrefixDigest,
		}
	default:
		return Diagnostic{
			Code:                 "JEV_UNKNOWN",
			Severity:             SeverityInfo,
			Message:              "calibration evidence is incomplete or invalid",
			EvidencePrefixDigest: decision.EvidencePrefixDigest,
		}
	}
}
