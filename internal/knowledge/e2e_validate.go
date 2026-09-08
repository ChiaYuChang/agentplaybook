package knowledge

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	sha256Pattern       = regexp.MustCompile(`^sha256:[a-fA-F0-9]{64}\s+\S+.*$`)
	candidateRefPattern = regexp.MustCompile(`^[a-fA-F0-9]{40}$`)
	testPackagePattern  = regexp.MustCompile(`^e2e/[a-zA-Z0-9_-]+(/[a-zA-Z0-9_.-]+)*$`)
	evidenceURIPattern  = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://`)
)

// E2EBriefContext models the input specification dispatched by Planner to Verifier.
type E2EBriefContext struct {
	RunID        string
	CandidateRef string
	TestPackage  string
	Mode         string
	TimeoutMS    int
}

// E2EScenarioCoverageContext models local sandbox manifest scenario evidence.
type E2EScenarioCoverageContext struct {
	SelectedScenarioIDs []string
	SkippedScenarioIDs  []string
	AllowedSkipIDs      []string
}

// E2EReportPayload models the compact verification report message artifact emitted by Verifier to Planner.
type E2EReportPayload struct {
	RunID              string
	CandidateRef       string
	StageOutcome       string
	BuildOutcome       string
	ExecutionOutcome   string
	ScenariosSelected  int
	ScenariosPassed    int
	ScenariosFailed    int
	ScenariosSkipped   int
	ResultCompleteness string
	TestedArtifactRef  string
	DurationMS         int
	Diagnostic         string
	EvidenceURI        string
}

// FormatE2EReport renders a compact, structured string representation of E2EReportPayload
// suitable for deterministic token budget estimation via EstimateTokenCount.
func FormatE2EReport(r E2EReportPayload) string {
	var sb strings.Builder
	sb.WriteString("run_id: ")
	sb.WriteString(r.RunID)
	sb.WriteString("\ncandidate_ref: ")
	sb.WriteString(r.CandidateRef)
	sb.WriteString("\nstage_outcome: ")
	sb.WriteString(r.StageOutcome)
	sb.WriteString("\nbuild_outcome: ")
	sb.WriteString(r.BuildOutcome)
	sb.WriteString("\nexecution_outcome: ")
	sb.WriteString(r.ExecutionOutcome)
	sb.WriteString(fmt.Sprintf("\nscenarios_selected: %d", r.ScenariosSelected))
	sb.WriteString(fmt.Sprintf("\nscenarios_passed: %d", r.ScenariosPassed))
	sb.WriteString(fmt.Sprintf("\nscenarios_failed: %d", r.ScenariosFailed))
	sb.WriteString(fmt.Sprintf("\nscenarios_skipped: %d", r.ScenariosSkipped))
	sb.WriteString("\nresult_completeness: ")
	sb.WriteString(r.ResultCompleteness)
	sb.WriteString("\ntested_artifact_ref: ")
	sb.WriteString(r.TestedArtifactRef)
	sb.WriteString(fmt.Sprintf("\nduration_ms: %d", r.DurationMS))
	sb.WriteString("\ndiagnostic: ")
	sb.WriteString(r.Diagnostic)
	sb.WriteString("\nevidence_uri: ")
	sb.WriteString(r.EvidenceURI)
	return sb.String()
}

// ValidateE2EReport validates that an E2EReportPayload adheres to all correlation,
// count reconciliation, stage outcome mapping, tested artifact provenance, and token budget contracts.
func ValidateE2EReport(brief E2EBriefContext, report E2EReportPayload, coverage E2EScenarioCoverageContext) error {
	// 1. Validate mandatory fields on brief
	if strings.TrimSpace(brief.RunID) == "" {
		return errors.New("brief run_id cannot be empty")
	}
	if strings.TrimSpace(brief.CandidateRef) == "" {
		return errors.New("brief candidate_ref cannot be empty")
	}
	if !candidateRefPattern.MatchString(brief.CandidateRef) {
		return fmt.Errorf("brief candidate_ref %q is invalid; must be an immutable 40-character hex commit SHA", brief.CandidateRef)
	}
	if strings.TrimSpace(brief.TestPackage) == "" {
		return errors.New("brief test_package cannot be empty")
	}
	if strings.Contains(brief.TestPackage, "..") || !testPackagePattern.MatchString(brief.TestPackage) {
		return fmt.Errorf("brief test_package %q is invalid; must be a repository-relative path under e2e/", brief.TestPackage)
	}
	if brief.Mode != "build-and-run" && brief.Mode != "run-only" {
		return fmt.Errorf("brief mode %q must be 'build-and-run' or 'run-only'", brief.Mode)
	}
	if brief.TimeoutMS <= 0 {
		return errors.New("brief timeout_ms must be greater than 0")
	}

	// 2. Validate mandatory fields on report
	if strings.TrimSpace(report.RunID) == "" {
		return errors.New("report run_id cannot be empty")
	}
	if strings.TrimSpace(report.CandidateRef) == "" {
		return errors.New("report candidate_ref cannot be empty")
	}
	if !candidateRefPattern.MatchString(report.CandidateRef) {
		return fmt.Errorf("report candidate_ref %q is invalid; must be an immutable 40-character hex commit SHA", report.CandidateRef)
	}
	if strings.TrimSpace(report.StageOutcome) == "" {
		return errors.New("report stage_outcome cannot be empty")
	}
	if strings.TrimSpace(report.BuildOutcome) == "" {
		return errors.New("report build_outcome cannot be empty")
	}
	if strings.TrimSpace(report.ExecutionOutcome) == "" {
		return errors.New("report execution_outcome cannot be empty")
	}
	if strings.TrimSpace(report.ResultCompleteness) == "" {
		return errors.New("report result_completeness cannot be empty")
	}
	if strings.TrimSpace(report.TestedArtifactRef) == "" {
		return errors.New("report tested_artifact_ref cannot be empty")
	}
	if strings.TrimSpace(report.Diagnostic) == "" {
		return errors.New("report diagnostic cannot be empty")
	}
	if strings.TrimSpace(report.EvidenceURI) == "" {
		return errors.New("report evidence_uri cannot be empty")
	}
	if report.EvidenceURI == "UNAVAILABLE" {
		if report.StageOutcome == "PASS" {
			return errors.New("evidence_uri UNAVAILABLE is prohibited when stage_outcome is PASS")
		}
		if report.StageOutcome == "PRODUCT_FAILURE" {
			return errors.New("evidence_uri UNAVAILABLE is prohibited when stage_outcome is PRODUCT_FAILURE")
		}
		if report.ExecutionOutcome != "NOT_RUN" && report.ExecutionOutcome != "ENV_ERROR" {
			return fmt.Errorf("evidence_uri UNAVAILABLE requires execution_outcome NOT_RUN or ENV_ERROR, got %s", report.ExecutionOutcome)
		}
		if report.ResultCompleteness != "NONE" {
			return fmt.Errorf("evidence_uri UNAVAILABLE requires result_completeness NONE, got %s", report.ResultCompleteness)
		}
		if report.ScenariosPassed != 0 || report.ScenariosFailed != 0 {
			return fmt.Errorf("evidence_uri UNAVAILABLE requires scenarios_passed == 0 and scenarios_failed == 0, got passed=%d, failed=%d", report.ScenariosPassed, report.ScenariosFailed)
		}
	} else if !evidenceURIPattern.MatchString(report.EvidenceURI) {
		return fmt.Errorf("report evidence_uri %q is invalid; must be 'UNAVAILABLE' or a valid URI", report.EvidenceURI)
	}

	// 3. Correlation invariants
	if report.RunID != brief.RunID {
		return fmt.Errorf("run_id mismatch: report has %q, brief has %q", report.RunID, brief.RunID)
	}
	if report.CandidateRef != brief.CandidateRef {
		return fmt.Errorf("candidate_ref mismatch: report has %q, brief has %q", report.CandidateRef, brief.CandidateRef)
	}

	// 4. Validate allowed enums
	switch report.StageOutcome {
	case "PASS", "BUILD_FAILURE", "PRODUCT_FAILURE", "ENV_BLOCKED", "CLARIFICATION_REQUIRED", "TIMEOUT", "CANCELLED", "INVALID_EVIDENCE":
	default:
		return fmt.Errorf("invalid stage_outcome: %q", report.StageOutcome)
	}

	switch report.BuildOutcome {
	case "PASS", "FAIL", "NOT_RUN", "ENV_ERROR", "TIMEOUT", "CANCELLED":
	default:
		return fmt.Errorf("invalid build_outcome: %q", report.BuildOutcome)
	}

	switch report.ExecutionOutcome {
	case "PASS", "FAIL", "NOT_RUN", "ENV_ERROR", "TIMEOUT", "CANCELLED", "INVALID_EVIDENCE":
	default:
		return fmt.Errorf("invalid execution_outcome: %q", report.ExecutionOutcome)
	}

	switch report.ResultCompleteness {
	case "COMPLETE", "PARTIAL", "NONE":
	default:
		return fmt.Errorf("invalid result_completeness: %q", report.ResultCompleteness)
	}

	// 5. Positive selection and scenario count reconciliation
	if report.ScenariosSelected <= 0 {
		return errors.New("scenarios_selected must be greater than 0")
	}
	if report.ScenariosSelected != len(coverage.SelectedScenarioIDs) {
		return fmt.Errorf("scenarios_selected count %d does not match selected scenario IDs count %d", report.ScenariosSelected, len(coverage.SelectedScenarioIDs))
	}

	selectedSet := make(map[string]bool, len(coverage.SelectedScenarioIDs))
	for _, id := range coverage.SelectedScenarioIDs {
		if strings.TrimSpace(id) == "" {
			return errors.New("selected scenario ID cannot be empty")
		}
		if selectedSet[id] {
			return fmt.Errorf("duplicate selected scenario ID: %q", id)
		}
		selectedSet[id] = true
	}

	skippedSet := make(map[string]bool, len(coverage.SkippedScenarioIDs))
	for _, id := range coverage.SkippedScenarioIDs {
		if strings.TrimSpace(id) == "" {
			return errors.New("skipped scenario ID cannot be empty")
		}
		if skippedSet[id] {
			return fmt.Errorf("duplicate skipped scenario ID: %q", id)
		}
		if !selectedSet[id] {
			return fmt.Errorf("skipped scenario ID %q is not in selected scenario IDs", id)
		}
		skippedSet[id] = true
	}

	if len(coverage.SkippedScenarioIDs) != report.ScenariosSkipped {
		return fmt.Errorf("scenarios_skipped count %d does not match skipped scenario IDs count %d", report.ScenariosSkipped, len(coverage.SkippedScenarioIDs))
	}

	if report.ScenariosPassed < 0 || report.ScenariosFailed < 0 || report.ScenariosSkipped < 0 {
		return errors.New("scenario counts cannot be negative")
	}
	if report.ScenariosSelected != report.ScenariosPassed+report.ScenariosFailed+report.ScenariosSkipped {
		return fmt.Errorf("scenario count reconciliation failed: selected (%d) != passed (%d) + failed (%d) + skipped (%d)",
			report.ScenariosSelected, report.ScenariosPassed, report.ScenariosFailed, report.ScenariosSkipped)
	}

	// Allowed-skip policy
	allowedSkipSet := make(map[string]bool, len(coverage.AllowedSkipIDs))
	for _, id := range coverage.AllowedSkipIDs {
		allowedSkipSet[id] = true
	}
	for _, id := range coverage.SkippedScenarioIDs {
		if !allowedSkipSet[id] && report.StageOutcome == "PASS" {
			return fmt.Errorf("unauthorized skipped scenario %q prohibits stage outcome PASS", id)
		}
	}

	// 6. tested_artifact_ref validation
	if report.TestedArtifactRef == "NONE" {
		if report.StageOutcome == "PASS" || report.StageOutcome == "PRODUCT_FAILURE" {
			return fmt.Errorf("tested_artifact_ref cannot be NONE for stage outcome %s", report.StageOutcome)
		}
		if report.BuildOutcome == "PASS" && brief.Mode == "build-and-run" {
			return errors.New("tested_artifact_ref cannot be NONE when build outcome is PASS")
		}
	} else {
		if report.TestedArtifactRef == "N/A" || !sha256Pattern.MatchString(report.TestedArtifactRef) {
			return fmt.Errorf("tested_artifact_ref %q is invalid; must carry immutable sha256:<hex> digest and provenance descriptor", report.TestedArtifactRef)
		}
	}

	// 7. Fail-closed cross-field stage outcome mappings
	switch report.StageOutcome {
	case "PASS":
		if brief.Mode == "build-and-run" {
			if report.BuildOutcome == "NOT_RUN" {
				return errors.New("stage outcome PASS in build-and-run mode cannot have build outcome NOT_RUN")
			}
			if report.BuildOutcome != "PASS" {
				return fmt.Errorf("stage outcome PASS in build-and-run mode requires build outcome PASS, got %s", report.BuildOutcome)
			}
		} else if brief.Mode == "run-only" {
			if report.BuildOutcome != "NOT_RUN" {
				return fmt.Errorf("stage outcome PASS in run-only mode requires build outcome NOT_RUN, got %s", report.BuildOutcome)
			}
		}
		if report.ExecutionOutcome != "PASS" {
			return fmt.Errorf("stage outcome PASS requires execution outcome PASS, got %s", report.ExecutionOutcome)
		}
		if report.ResultCompleteness != "COMPLETE" {
			return fmt.Errorf("stage outcome PASS requires result completeness COMPLETE, got %s", report.ResultCompleteness)
		}
		if report.ScenariosFailed != 0 {
			return fmt.Errorf("stage outcome PASS requires scenarios_failed == 0, got %d", report.ScenariosFailed)
		}
		if report.ScenariosPassed <= 0 {
			return fmt.Errorf("stage outcome PASS requires scenarios_passed > 0, got %d", report.ScenariosPassed)
		}
		if report.EvidenceURI == "UNAVAILABLE" {
			return errors.New("stage outcome PASS cannot have evidence_uri UNAVAILABLE")
		}
		if report.TestedArtifactRef == "NONE" {
			return errors.New("stage outcome PASS cannot have tested_artifact_ref NONE")
		}

	case "BUILD_FAILURE":
		if report.BuildOutcome == "ENV_ERROR" {
			return errors.New("build outcome ENV_ERROR must map to stage outcome ENV_BLOCKED, not BUILD_FAILURE")
		}
		if report.BuildOutcome == "TIMEOUT" {
			return errors.New("build outcome TIMEOUT must map to stage outcome TIMEOUT, not BUILD_FAILURE")
		}
		if report.BuildOutcome != "FAIL" {
			return fmt.Errorf("stage outcome BUILD_FAILURE strictly requires build outcome FAIL, got %s", report.BuildOutcome)
		}
		if brief.Mode != "build-and-run" {
			return errors.New("stage outcome BUILD_FAILURE is only valid in build-and-run mode")
		}
		if report.ExecutionOutcome != "NOT_RUN" {
			return fmt.Errorf("stage outcome BUILD_FAILURE requires execution outcome NOT_RUN, got %s", report.ExecutionOutcome)
		}
		if report.TestedArtifactRef != "NONE" {
			return fmt.Errorf("stage outcome BUILD_FAILURE requires tested_artifact_ref NONE, got %s", report.TestedArtifactRef)
		}

	case "PRODUCT_FAILURE":
		if report.ExecutionOutcome != "FAIL" {
			return fmt.Errorf("stage outcome PRODUCT_FAILURE requires execution outcome FAIL, got %s", report.ExecutionOutcome)
		}
		if report.ScenariosFailed <= 0 {
			return fmt.Errorf("stage outcome PRODUCT_FAILURE requires scenarios_failed > 0, got %d", report.ScenariosFailed)
		}
		if brief.Mode == "build-and-run" && report.BuildOutcome != "PASS" {
			return fmt.Errorf("stage outcome PRODUCT_FAILURE in build-and-run mode requires build outcome PASS, got %s", report.BuildOutcome)
		}
		if brief.Mode == "run-only" && report.BuildOutcome != "NOT_RUN" {
			return fmt.Errorf("stage outcome PRODUCT_FAILURE in run-only mode requires build outcome NOT_RUN, got %s", report.BuildOutcome)
		}

	case "ENV_BLOCKED":
		if report.BuildOutcome != "ENV_ERROR" && report.ExecutionOutcome != "ENV_ERROR" {
			return errors.New("stage outcome ENV_BLOCKED requires build outcome or execution outcome to be ENV_ERROR")
		}
		if report.BuildOutcome == "PASS" && report.ExecutionOutcome == "PASS" {
			return errors.New("stage outcome ENV_BLOCKED cannot have both build outcome PASS and execution outcome PASS")
		}

	case "CLARIFICATION_REQUIRED":
		if report.BuildOutcome != "ENV_ERROR" && report.BuildOutcome != "NOT_RUN" {
			return fmt.Errorf("stage outcome CLARIFICATION_REQUIRED requires build outcome ENV_ERROR or NOT_RUN, got %s", report.BuildOutcome)
		}
		if report.ExecutionOutcome != "NOT_RUN" {
			return fmt.Errorf("stage outcome CLARIFICATION_REQUIRED requires execution outcome NOT_RUN, got %s", report.ExecutionOutcome)
		}
		if report.ResultCompleteness != "NONE" {
			return fmt.Errorf("stage outcome CLARIFICATION_REQUIRED requires result completeness NONE, got %s", report.ResultCompleteness)
		}

	case "TIMEOUT":
		if report.BuildOutcome != "TIMEOUT" && report.ExecutionOutcome != "TIMEOUT" {
			return errors.New("stage outcome TIMEOUT requires build outcome or execution outcome to be TIMEOUT")
		}
		if report.ExecutionOutcome == "PASS" {
			return errors.New("stage outcome TIMEOUT cannot have execution outcome PASS")
		}
		if report.ResultCompleteness != "PARTIAL" && report.ResultCompleteness != "NONE" {
			return fmt.Errorf("stage outcome TIMEOUT requires result completeness PARTIAL or NONE, got %s", report.ResultCompleteness)
		}

	case "CANCELLED":
		if report.BuildOutcome != "CANCELLED" && report.ExecutionOutcome != "CANCELLED" {
			return errors.New("stage outcome CANCELLED requires build outcome or execution outcome to be CANCELLED")
		}
		if report.ExecutionOutcome == "PASS" {
			return errors.New("stage outcome CANCELLED cannot have execution outcome PASS")
		}
		if report.ResultCompleteness != "PARTIAL" && report.ResultCompleteness != "NONE" {
			return fmt.Errorf("stage outcome CANCELLED requires result completeness PARTIAL or NONE, got %s", report.ResultCompleteness)
		}

	case "INVALID_EVIDENCE":
		if report.ExecutionOutcome != "INVALID_EVIDENCE" {
			return fmt.Errorf("stage outcome INVALID_EVIDENCE requires execution outcome INVALID_EVIDENCE, got %s", report.ExecutionOutcome)
		}
	}

	// 8. Deterministic token budget limit (<150 tokens)
	formatted := FormatE2EReport(report)
	tokens := EstimateTokenCount(formatted)
	if tokens >= 150 {
		return fmt.Errorf("e2e report token budget exceeded: %d tokens (must be <150 tokens)", tokens)
	}

	return nil
}
