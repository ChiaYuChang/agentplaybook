package knowledge

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var (
	candidateRefPattern      = regexp.MustCompile(`^[a-f0-9]{40}$`)
	sha256PrefixPattern      = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
	asciiKebabPattern        = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	testPackagePattern       = regexp.MustCompile(`^e2e(/[a-zA-Z0-9_.-]+)*$`)
	evidenceURIPattern       = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://`)
	testedArtifactRefPattern = regexp.MustCompile(`^sha256:([a-f0-9]{64});candidate=([a-f0-9]{40});build_material=sha256:([a-f0-9]{64});source=(FRESH|VERIFIED_REUSE)$`)
)

// E2EBriefContext models the complete input specification dispatched by Planner to Verifier.
type E2EBriefContext struct {
	RunID                   string                       `json:"run_id"`
	RepoRef                 string                       `json:"repo_ref"`
	CandidateRef            string                       `json:"candidate_ref"`
	TestPackage             string                       `json:"test_package"`
	Mode                    string                       `json:"mode"`
	Command                 []string                     `json:"command"`
	BuildCommand            []string                     `json:"build_command"`
	TimeoutMS               int                          `json:"timeout_ms"`
	GoVersion               string                       `json:"go_version"`
	Platforms               []string                     `json:"platforms"`
	DependencyLockRef       string                       `json:"dependency_lock_ref"`
	Environment             []E2EEnvironmentInput        `json:"environment"`
	ExpectedOutputs         []string                     `json:"expected_outputs"`
	Scenarios               []E2EScenarioContract        `json:"scenarios"`
	AllowedSkipIDs          []string                     `json:"allowed_skip_ids"`
	Services                []E2EServiceRequirement      `json:"services"`
	CleanupObligations      []E2ECleanupObligation       `json:"cleanup_obligations"`
	CredentialPolicy        []E2ECredentialPolicyEntry   `json:"credential_policy"`
	CredentialAuthorization []E2ECredentialAuthorization `json:"credential_authorization"`
	PriorCleanupResolution  *E2ECleanupResolution        `json:"prior_cleanup_resolution"`
}

// E2EEnvironmentInput models an individual environment input for an E2E run.
type E2EEnvironmentInput struct {
	Name           string `json:"name"`
	Source         string `json:"source"`
	Value          string `json:"value"`
	ServiceID      string `json:"service_id"`
	CredentialName string `json:"credential_name"`
}

// E2EScenarioContract models a scenario execution contract within an E2E brief.
type E2EScenarioContract struct {
	ID             string   `json:"id"`
	Preconditions  string   `json:"preconditions"`
	Steps          []string `json:"steps"`
	ExpectedResult string   `json:"expected_result"`
	Verification   string   `json:"verification"`
}

// E2EServiceRequirement models an external service dependency.
type E2EServiceRequirement struct {
	ID               string `json:"id"`
	Kind             string `json:"kind"`
	ImageOrVersion   string `json:"image_or_version"`
	ReadyCondition   string `json:"ready_condition"`
	StartupTimeoutMS int    `json:"startup_timeout_ms"`
}

// E2ECleanupObligation models a cleanup obligation for a test resource or credential.
type E2ECleanupObligation struct {
	ResourceID string `json:"resource_id"`
	Kind       string `json:"kind"`
	Owner      string `json:"owner"`
	Action     string `json:"action"`
}

// E2ECredentialPolicyEntry models an ingress credential policy definition.
type E2ECredentialPolicyEntry struct {
	Name                  string `json:"name"`
	GenerationMode        string `json:"generation_mode"`
	GenerationInstruction string `json:"generation_instruction"`
}

// E2ECredentialAuthorization models human approval for a REAL_ONLY credential.
type E2ECredentialAuthorization struct {
	Name             string `json:"name"`
	Purpose          string `json:"purpose"`
	Lifetime         string `json:"lifetime"`
	AuthorizedBy     string `json:"authorized_by"`
	AuthorizedAt     string `json:"authorized_at"`
	AuthorizationRef string `json:"authorization_ref"`
}

// E2ECleanupResolution models a recorded human resolution of a prior unconfirmed cleanup.
type E2ECleanupResolution struct {
	PriorRunID        string   `json:"prior_run_id"`
	PriorCandidateRef string   `json:"prior_candidate_ref"`
	ResourceIDs       []string `json:"resource_ids"`
	Disposition       string   `json:"disposition"`
	Action            string   `json:"action"`
	HumanApprover     string   `json:"human_approver"`
	ResolvedAt        string   `json:"resolved_at"`
	EvidenceURI       string   `json:"evidence_uri"`
}

// E2EPriorCleanupContext models controller context regarding prior cleanup state.
type E2EPriorCleanupContext struct {
	Required                bool
	PriorRunID              string
	PriorCandidateRef       string
	UnresolvedResourceIDs   []string
	HumanResolutionVerified bool
}

// E2EBriefValidationContext models controller context provided when validating a brief.
type E2EBriefValidationContext struct {
	RunIDUnique                     bool
	CredentialAuthorizationVerified bool
	PriorCleanup                    E2EPriorCleanupContext
}

// E2EScenarioResult models an individual scenario result.
type E2EScenarioResult struct {
	ScenarioID string `json:"scenario_id"`
	Outcome    string `json:"outcome"`
}

// E2EScenarioCoverageContext models test coverage results across all brief scenarios.
type E2EScenarioCoverageContext struct {
	Results []E2EScenarioResult `json:"results"`
}

// E2EReportPayload models the verification report message artifact emitted by Verifier
// or finalized by controller on runner loss.
type E2EReportPayload struct {
	RunID                string              `json:"run_id"`
	RepoRef              string              `json:"repo_ref"`
	CandidateRef         string              `json:"candidate_ref"`
	TestPackage          string              `json:"test_package"`
	StageOutcome         string              `json:"stage_outcome"`
	BuildOutcome         string              `json:"build_outcome"`
	ExecutionOutcome     string              `json:"execution_outcome"`
	ScenariosSelected    int                 `json:"scenarios_selected"`
	ScenariosPassed      int                 `json:"scenarios_passed"`
	ScenariosFailed      int                 `json:"scenarios_failed"`
	ScenariosSkipped     int                 `json:"scenarios_skipped"`
	ScenariosNotRun      int                 `json:"scenarios_not_run"`
	ScenariosUnknown     int                 `json:"scenarios_unknown"`
	ScenarioResults      []E2EScenarioResult `json:"scenario_results"`
	ResultCompleteness   string              `json:"result_completeness"`
	TestedArtifactRef    string              `json:"tested_artifact_ref"`
	ExecutionMaterialRef string              `json:"execution_material_ref"`
	DurationMS           int                 `json:"duration_ms"`
	Diagnostic           string              `json:"diagnostic"`
	EvidenceURI          string              `json:"evidence_uri"`
	CleanupStatus        string              `json:"cleanup_status"`
	CleanupEvidenceURI   string              `json:"cleanup_evidence_uri"`
	SecretsAttestation   string              `json:"secrets_attestation"`
}

// E2EReportFinalizationContext models append-only controller-ledger finalization context.
type E2EReportFinalizationContext struct {
	State                           string // FIRST_FINALIZATION|DUPLICATE_IDENTICAL|CONFLICTING|LATE_AFTER_FINAL
	Authority                       string // incoming authority: VERIFIER|CONTROLLER_RUNNER_LOST
	ExistingTerminalReportAuthority string // VERIFIER|CONTROLLER_RUNNER_LOST
	ExistingTerminalReportDigest    string
	IncomingReportDigest            string
	TerminalReportImmutable         bool
}

// E2EReportValidationContext models controller facts supplied for report validation.
type E2EReportValidationContext struct {
	RunIDUnique                     bool
	CredentialAuthorizationVerified bool
	PriorCleanupResolved            bool
	PreflightOutcome                string // READY|AUTHORIZATION_DENIED
	BuildProcessState               string // NOT_STARTED|RUNNING|PASS|FAIL|ENV_ERROR|TIMEOUT|CANCELLED
	ExecutionProcessState           string // NOT_STARTED|RUNNING|PASS|FAIL|ENV_ERROR|TIMEOUT|CANCELLED
	ExecutionStarted                bool
	ServiceStartupAttempted         bool
	CredentialProvisioningState     string // NOT_STARTED|SUCCEEDED|FAILED|UNKNOWN
	ResourcesAllocated              bool
	AllocatedResourceIDs            []string
	UnresolvedResourceIDs           []string
	CleanupRequired                 bool
	CleanupStatus                   string // CONFIRMED|UNCONFIRMED
	CleanupRecordVerified           bool
	Finalization                    E2EReportFinalizationContext
}

// E2EExecutionMaterialContext captures provenance and material context for admission evaluation.
type E2EExecutionMaterialContext struct {
	RepoRef                   string
	CandidateTreeRef          string
	SourceTreeBeforeRef       string
	SourceTreeAfterRef        string
	CheckoutCleanBefore       bool
	CheckoutCleanAfter        bool
	InputProtectionMode       string
	ProtectionProofRef        string
	ProtectedIntervalComplete bool
	ProtectionProofVerified   bool
	TestPackagePath           string
	TestPackageTreeRef        string
	Command                   []string
	BuildCommand              []string
	GoVersion                 string
	Platforms                 []string
	DependencyLockRef         string
	ServiceRefs               []string
	FreshTestExecution        bool
	ArtifactSHA256            string
	ArtifactCandidateRef      string
	ArtifactBuildMaterialRef  string
	ArtifactSource            string
}

// CanonicalExecutionMaterial represents canonical UTF-8 JSON structure for execution material hash.
type CanonicalExecutionMaterial struct {
	Brief               E2EBriefContext `json:"brief"`
	CandidateTreeRef    string          `json:"candidate_tree_ref"`
	DependencyLockRef   string          `json:"dependency_lock_ref"`
	GoVersion           string          `json:"go_version"`
	InputProtectionMode string          `json:"input_protection_mode"`
	Platforms           []string        `json:"platforms"`
	ProtectionProofRef  string          `json:"protection_proof_ref"`
	RepoRef             string          `json:"repo_ref"`
	ServiceRefs         []string        `json:"service_refs"`
	TestPackagePath     string          `json:"test_package_path"`
	TestPackageTreeRef  string          `json:"test_package_tree_ref"`
}

// CanonicalBuildMaterial represents canonical UTF-8 JSON structure for build input digest.
type CanonicalBuildMaterial struct {
	BuildCommand      []string `json:"build_command"`
	CandidateRef      string   `json:"candidate_ref"`
	CandidateTreeRef  string   `json:"candidate_tree_ref"`
	DependencyLockRef string   `json:"dependency_lock_ref"`
	GoVersion         string   `json:"go_version"`
	Platforms         []string `json:"platforms"`
	RepoRef           string   `json:"repo_ref"`
}

// CanonicalJSON marshals v to JSON and re-encodes it with recursively lexically sorted object keys at all nested levels.
func CanonicalJSON(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var generic any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&generic); err != nil {
		return nil, err
	}
	return json.Marshal(generic)
}

// ComputeBuildMaterialDigest computes SHA-256 over canonical JSON of build material inputs.
func ComputeBuildMaterialDigest(repoRef, candidateRef, candidateTreeRef string, buildCommand []string, goVersion string, platforms []string, dependencyLockRef string) string {
	cbm := CanonicalBuildMaterial{
		BuildCommand:      buildCommand,
		CandidateRef:      candidateRef,
		CandidateTreeRef:  candidateTreeRef,
		DependencyLockRef: dependencyLockRef,
		GoVersion:         goVersion,
		Platforms:         platforms,
		RepoRef:           repoRef,
	}
	data, err := CanonicalJSON(cbm)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// ComputeExecutionMaterialRef computes sha256:<hex> over canonical UTF-8 JSON for execution material.
func ComputeExecutionMaterialRef(brief E2EBriefContext, material E2EExecutionMaterialContext) string {
	cem := CanonicalExecutionMaterial{
		Brief:               brief,
		CandidateTreeRef:    material.CandidateTreeRef,
		DependencyLockRef:   material.DependencyLockRef,
		GoVersion:           material.GoVersion,
		InputProtectionMode: material.InputProtectionMode,
		Platforms:           material.Platforms,
		ProtectionProofRef:  material.ProtectionProofRef,
		RepoRef:             material.RepoRef,
		ServiceRefs:         material.ServiceRefs,
		TestPackagePath:     material.TestPackagePath,
		TestPackageTreeRef:  material.TestPackageTreeRef,
	}
	data, err := CanonicalJSON(cem)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(h[:])
}

// FormatTestedArtifactRef formats a canonical tested_artifact_ref string.
func FormatTestedArtifactRef(artifactSHA256, candidateRef, buildMaterialDigest, source string) string {
	return fmt.Sprintf("sha256:%s;candidate=%s;build_material=sha256:%s;source=%s", artifactSHA256, candidateRef, buildMaterialDigest, source)
}

// ComputeCanonicalSecretsAttestation computes canonical SecretsAttestation based on credential policy and authorization.
func ComputeCanonicalSecretsAttestation(policy []E2ECredentialPolicyEntry, auth []E2ECredentialAuthorization) (string, error) {
	if len(policy) == 0 {
		return fmt.Sprintf("NOT_PROVISIONED sha256:%s", ComputePolicyDigest(policy)), nil
	}
	var realOnlyNames []string
	var allNames []string
	for _, entry := range policy {
		allNames = append(allNames, entry.Name)
		if entry.GenerationMode == "REAL_ONLY" {
			realOnlyNames = append(realOnlyNames, entry.Name)
		}
	}
	sort.Strings(allNames)
	sort.Strings(realOnlyNames)

	if len(realOnlyNames) == 0 {
		var hashInput []byte
		if len(allNames) > 0 {
			hashInput = []byte(strings.Join(allNames, "\n") + "\n")
		}
		h := sha256.Sum256(hashInput)
		return fmt.Sprintf("ATTESTED:THROWAWAY sha256:%s", hex.EncodeToString(h[:])), nil
	}

	authNames := make([]string, len(auth))
	for i, a := range auth {
		authNames[i] = a.Name
	}
	sort.Strings(authNames)
	if len(authNames) != len(realOnlyNames) {
		return "", errors.New("mismatch between real-only policy entries and authorization count")
	}
	for i := range realOnlyNames {
		if authNames[i] != realOnlyNames[i] {
			return "", fmt.Errorf("authorization name %q does not match policy %q", authNames[i], realOnlyNames[i])
		}
	}

	hashInput := []byte(strings.Join(realOnlyNames, "\n") + "\n")
	h := sha256.Sum256(hashInput)
	return fmt.Sprintf("ATTESTED:MIXED:%s sha256:%s", strings.Join(realOnlyNames, ","), hex.EncodeToString(h[:])), nil
}

func isValidAuthority(auth string) bool {
	return auth == "VERIFIER" || auth == "CONTROLLER_RUNNER_LOST"
}

var goTestValueFlags = map[string]bool{
	"o":                    true,
	"p":                    true,
	"cpu":                  true,
	"tags":                 true,
	"ldflags":              true,
	"gcflags":              true,
	"asmflags":             true,
	"gccgoflags":           true,
	"pkgdir":               true,
	"toolexec":             true,
	"installsuffix":        true,
	"workdir":              true,
	"bench":                true,
	"benchtime":            true,
	"blockprofile":         true,
	"blockprofilerate":     true,
	"count":                true,
	"covermode":            true,
	"coverpkg":             true,
	"coverprofile":         true,
	"cpuprofile":           true,
	"exec":                 true,
	"fuzz":                 true,
	"fuzzcachedir":         true,
	"fuzzminimizetime":     true,
	"fuzztime":             true,
	"list":                 true,
	"memprofile":           true,
	"memprofilerate":       true,
	"mutexprofile":         true,
	"mutexprofilefraction": true,
	"outputdir":            true,
	"parallel":             true,
	"run":                  true,
	"shuffle":              true,
	"skip":                 true,
	"timeout":              true,
	"trace":                true,
	"vet":                  true,
}

var goTestBoolFlags = map[string]bool{
	"a":          true,
	"i":          true,
	"n":          true,
	"v":          true,
	"x":          true,
	"race":       true,
	"msan":       true,
	"asan":       true,
	"cover":      true,
	"work":       true,
	"trimpath":   true,
	"linkshared": true,
	"benchmem":   true,
	"c":          true,
	"failfast":   true,
	"fullpath":   true,
	"gocheck":    true,
	"panicnil":   true,
	"short":      true,
}

func isIllegalControlOrLineChar(r rune) bool {
	return unicode.IsControl(r) || r == '\u2028' || r == '\u2029'
}

func validateGoTestCommand(cmd []string) error {
	if len(cmd) < 3 {
		return fmt.Errorf("formal test command has too few arguments (%d); must be a valid Go test invocation", len(cmd))
	}
	bin := filepath.Base(cmd[0])
	if bin != "go" && bin != "go.exe" {
		return fmt.Errorf("test command executable must be 'go', got %q", cmd[0])
	}
	if cmd[1] != "test" {
		return fmt.Errorf("test command subcommand must be 'test', got %q", cmd[1])
	}

	// Flags for 'go test' only apply before -args
	flagsEnd := len(cmd)
	for i := 2; i < len(cmd); i++ {
		if cmd[i] == "-args" || cmd[i] == "--args" {
			flagsEnd = i
			break
		}
	}

	var countSeen bool
	var effectiveCount string
	for i := 2; i < flagsEnd; i++ {
		arg := cmd[i]
		if arg == "--" {
			// In Go / unix flag parsing, '--' terminates flags.
			break
		}
		if !strings.HasPrefix(arg, "-") {
			// Package path or source file
			continue
		}

		flagBody := strings.TrimPrefix(strings.TrimPrefix(arg, "--"), "-")
		if flagBody == "" {
			return errors.New("formal test command contains invalid empty flag")
		}

		var flagName string
		var flagVal string
		var hasVal bool
		if idx := strings.Index(flagBody, "="); idx != -1 {
			flagName = flagBody[:idx]
			flagVal = flagBody[idx+1:]
			hasVal = true
		} else {
			flagName = flagBody
		}

		cleanName := strings.TrimPrefix(flagName, "test.")

		switch {
		case cleanName == "count":
			if hasVal {
				countSeen = true
				effectiveCount = flagVal
			} else {
				if i+1 < flagsEnd {
					countSeen = true
					effectiveCount = cmd[i+1]
					i++
				} else {
					return errors.New("flag -count requires an argument")
				}
			}
		case goTestValueFlags[cleanName]:
			if !hasVal {
				if i+1 < flagsEnd {
					i++
				} else {
					return fmt.Errorf("flag %q requires an argument", arg)
				}
			}
		case goTestBoolFlags[cleanName]:
			// recognized bool flag, does not consume next token
		default:
			// Unknown flag
			if !hasVal {
				return fmt.Errorf("ambiguous or unrecognized flag %q without '=' in formal test command", arg)
			}
		}
	}

	if !countSeen {
		return errors.New("formal test command must include distinct -count=1 flag before -args to prevent cached test verdicts")
	}
	if effectiveCount != "1" {
		return fmt.Errorf("effective -count flag value must be '1', got %q", effectiveCount)
	}

	// Forwarded arguments after -args / --args are passed directly to the test binary.
	// Model native test-binary flag parsing: stop interpreting flag-shaped tokens after
	// the first positional argument or '--' terminator, and consume the value of known
	// value-taking native flags (e.g., -test.outputdir), even if that value begins with '-'.
	for i := flagsEnd + 1; i < len(cmd); i++ {
		arg := cmd[i]
		if arg == "--" {
			// '--' terminates flag parsing for the test binary
			break
		}
		if len(arg) < 2 || arg[0] != '-' {
			// Flag parsing stops just before the first non-flag argument in Go flag.FlagSet
			break
		}
		flagBody := strings.TrimPrefix(strings.TrimPrefix(arg, "--"), "-")
		if flagBody == "" {
			break
		}

		var flagName string
		var flagVal string
		var hasVal bool
		if idx := strings.Index(flagBody, "="); idx != -1 {
			flagName = flagBody[:idx]
			flagVal = flagBody[idx+1:]
			hasVal = true
		} else {
			flagName = flagBody
		}

		cleanName := strings.TrimPrefix(flagName, "test.")
		if cleanName == "count" {
			var val string
			if hasVal {
				val = flagVal
			} else {
				if i+1 < len(cmd) {
					val = cmd[i+1]
					i++
				} else {
					return errors.New("forwarded test flag -count requires an argument")
				}
			}
			if val != "1" {
				return fmt.Errorf("forwarded test argument %q overrides effective test count to %q, disabling repeat protection", arg, val)
			}
		} else if goTestValueFlags[cleanName] {
			if !hasVal {
				if i+1 < len(cmd) {
					// In Go flag.FlagSet, value-taking flags consume the next argument,
					// even if that value begins with '-'.
					i++
				} else {
					return fmt.Errorf("forwarded test flag %q requires an argument", arg)
				}
			}
		}
	}

	return nil
}

// ComputePolicyDigest computes SHA-256 hex digest over sorted lines of declared credential policy entries.
func ComputePolicyDigest(policy []E2ECredentialPolicyEntry) string {
	if len(policy) == 0 {
		h := sha256.Sum256([]byte{})
		return hex.EncodeToString(h[:])
	}
	lines := make([]string, len(policy))
	for i, p := range policy {
		lines[i] = fmt.Sprintf("%s:%s", p.Name, p.GenerationMode)
	}
	sort.Strings(lines)
	hashInput := []byte(strings.Join(lines, "\n") + "\n")
	h := sha256.Sum256(hashInput)
	return hex.EncodeToString(h[:])
}

// FormatE2EReport renders a deterministic structured string representation of E2EReportPayload
// containing all 23 required report fields in canonical catalog order losslessly.
func FormatE2EReport(r E2EReportPayload) string {
	var sb strings.Builder
	sb.WriteString("run_id: ")
	sb.WriteString(r.RunID)
	sb.WriteString("\nrepo_ref: ")
	sb.WriteString(r.RepoRef)
	sb.WriteString("\ncandidate_ref: ")
	sb.WriteString(r.CandidateRef)
	sb.WriteString("\ntest_package: ")
	sb.WriteString(r.TestPackage)
	sb.WriteString("\nstage_outcome: ")
	sb.WriteString(r.StageOutcome)
	sb.WriteString("\nbuild_outcome: ")
	sb.WriteString(r.BuildOutcome)
	sb.WriteString("\nexecution_outcome: ")
	sb.WriteString(r.ExecutionOutcome)
	// strings.Builder writes cannot fail, discarding return values explicitly.
	_, _ = fmt.Fprintf(&sb, "\nscenarios_selected: %d", r.ScenariosSelected)
	_, _ = fmt.Fprintf(&sb, "\nscenarios_passed: %d", r.ScenariosPassed)
	_, _ = fmt.Fprintf(&sb, "\nscenarios_failed: %d", r.ScenariosFailed)
	_, _ = fmt.Fprintf(&sb, "\nscenarios_skipped: %d", r.ScenariosSkipped)
	_, _ = fmt.Fprintf(&sb, "\nscenarios_not_run: %d", r.ScenariosNotRun)
	_, _ = fmt.Fprintf(&sb, "\nscenarios_unknown: %d", r.ScenariosUnknown)
	sb.WriteString("\nscenario_results:")
	for _, sr := range r.ScenarioResults {
		_, _ = fmt.Fprintf(&sb, "\n  - %s: %s", sr.ScenarioID, sr.Outcome)
	}
	sb.WriteString("\nresult_completeness: ")
	sb.WriteString(r.ResultCompleteness)
	sb.WriteString("\ntested_artifact_ref: ")
	sb.WriteString(r.TestedArtifactRef)
	sb.WriteString("\nexecution_material_ref: ")
	sb.WriteString(r.ExecutionMaterialRef)
	_, _ = fmt.Fprintf(&sb, "\nduration_ms: %d", r.DurationMS)
	sb.WriteString("\ndiagnostic: ")
	sb.WriteString(r.Diagnostic)
	sb.WriteString("\nevidence_uri: ")
	sb.WriteString(r.EvidenceURI)
	sb.WriteString("\ncleanup_status: ")
	sb.WriteString(r.CleanupStatus)
	sb.WriteString("\ncleanup_evidence_uri: ")
	sb.WriteString(r.CleanupEvidenceURI)
	sb.WriteString("\nsecrets_attestation: ")
	sb.WriteString(r.SecretsAttestation)
	return sb.String()
}

// ComputeReportDigest computes canonical SHA-256 hex digest over exact UTF-8 bytes returned by FormatE2EReport.
func ComputeReportDigest(report E2EReportPayload) string {
	formatted := FormatE2EReport(report)
	h := sha256.Sum256([]byte(formatted))
	return hex.EncodeToString(h[:])
}

// ValidateE2EBrief validates that an E2EBriefContext is complete, coherent, and adheres to all execution contracts.
func ValidateE2EBrief(brief E2EBriefContext, state E2EBriefValidationContext) error {
	if strings.TrimSpace(brief.RunID) == "" {
		return errors.New("E2E_BRIEF_INVALID: run_id cannot be empty")
	}
	if !utf8.ValidString(brief.RunID) {
		return errors.New("E2E_BRIEF_INVALID: field run_id contains invalid UTF-8")
	}
	for _, r := range brief.RunID {
		if isIllegalControlOrLineChar(r) {
			return errors.New("E2E_BRIEF_INVALID: field run_id contains illegal control character")
		}
	}
	if !state.RunIDUnique {
		return fmt.Errorf("E2E_BRIEF_INVALID: run_id %q is not unique in ledger", brief.RunID)
	}
	if strings.TrimSpace(brief.RepoRef) == "" {
		return errors.New("E2E_BRIEF_INVALID: repo_ref cannot be empty")
	}
	if !utf8.ValidString(brief.RepoRef) {
		return errors.New("E2E_BRIEF_INVALID: field repo_ref contains invalid UTF-8")
	}
	for _, r := range brief.RepoRef {
		if isIllegalControlOrLineChar(r) {
			return errors.New("E2E_BRIEF_INVALID: field repo_ref contains illegal control character")
		}
	}
	if strings.TrimSpace(brief.CandidateRef) == "" {
		return errors.New("E2E_BRIEF_INVALID: candidate_ref cannot be empty")
	}
	if !candidateRefPattern.MatchString(brief.CandidateRef) {
		return fmt.Errorf("E2E_BRIEF_INVALID: candidate_ref %q is invalid; must be an immutable 40-character lowercase hex commit SHA", brief.CandidateRef)
	}
	if strings.TrimSpace(brief.TestPackage) == "" {
		return errors.New("E2E_BRIEF_INVALID: test_package cannot be empty")
	}
	if !utf8.ValidString(brief.TestPackage) {
		return errors.New("E2E_BRIEF_INVALID: field test_package contains invalid UTF-8")
	}
	for _, r := range brief.TestPackage {
		if isIllegalControlOrLineChar(r) {
			return errors.New("E2E_BRIEF_INVALID: field test_package contains illegal control character")
		}
	}
	if strings.Contains(brief.TestPackage, "..") || !testPackagePattern.MatchString(brief.TestPackage) {
		return fmt.Errorf("E2E_BRIEF_INVALID: test_package %q is invalid; must be a safe repository-relative path rooted at e2e/", brief.TestPackage)
	}
	if brief.Mode != "build-and-run" && brief.Mode != "run-only" {
		return fmt.Errorf("E2E_BRIEF_INVALID: mode %q must be 'build-and-run' or 'run-only'", brief.Mode)
	}
	if len(brief.Command) == 0 {
		return errors.New("E2E_BRIEF_INVALID: command cannot be empty")
	}
	for _, arg := range brief.Command {
		if strings.TrimSpace(arg) == "" {
			return errors.New("E2E_BRIEF_INVALID: command arguments cannot be empty strings")
		}
	}
	if err := validateGoTestCommand(brief.Command); err != nil {
		return fmt.Errorf("E2E_BRIEF_INVALID: %w", err)
	}
	if brief.Mode == "build-and-run" {
		if len(brief.BuildCommand) == 0 {
			return errors.New("E2E_BRIEF_INVALID: build_command cannot be empty in build-and-run mode")
		}
		for _, arg := range brief.BuildCommand {
			if strings.TrimSpace(arg) == "" {
				return errors.New("E2E_BRIEF_INVALID: build_command arguments cannot be empty strings")
			}
		}
	}
	if brief.TimeoutMS <= 0 {
		return errors.New("E2E_BRIEF_INVALID: timeout_ms must be greater than 0")
	}
	if strings.TrimSpace(brief.GoVersion) == "" {
		return errors.New("E2E_BRIEF_INVALID: go_version cannot be empty")
	}
	if len(brief.Platforms) == 0 {
		return errors.New("E2E_BRIEF_INVALID: platforms cannot be empty")
	}
	for _, p := range brief.Platforms {
		if strings.TrimSpace(p) == "" {
			return errors.New("E2E_BRIEF_INVALID: platform cannot be empty string")
		}
	}
	if strings.TrimSpace(brief.DependencyLockRef) == "" {
		return errors.New("E2E_BRIEF_INVALID: dependency_lock_ref cannot be empty")
	}
	if brief.DependencyLockRef != "NONE" && !sha256PrefixPattern.MatchString(brief.DependencyLockRef) {
		return fmt.Errorf("E2E_BRIEF_INVALID: dependency_lock_ref %q is invalid; must be 'NONE' or sha256:<64 lowercase hex>", brief.DependencyLockRef)
	}

	// ExpectedOutputs validation
	if brief.ExpectedOutputs == nil {
		return errors.New("E2E_BRIEF_INVALID: expected_outputs is required (use explicit empty list if none)")
	}
	for _, outPath := range brief.ExpectedOutputs {
		trimmed := strings.TrimSpace(outPath)
		if trimmed == "" {
			return errors.New("E2E_BRIEF_INVALID: expected_outputs path cannot be empty")
		}
		if strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, "\\") || filepath.IsAbs(trimmed) || strings.HasPrefix(trimmed, "~") {
			return fmt.Errorf("E2E_BRIEF_INVALID: expected_outputs path %q cannot be absolute; must be run-scoped relative path", outPath)
		}
		if strings.Contains(trimmed, "..") {
			return fmt.Errorf("E2E_BRIEF_INVALID: expected_outputs path %q cannot contain directory traversal '..'", outPath)
		}
	}

	// Environment validation
	envNames := make(map[string]bool)
	for _, env := range brief.Environment {
		if strings.TrimSpace(env.Name) == "" {
			return errors.New("E2E_BRIEF_INVALID: environment input name cannot be empty")
		}
		if envNames[env.Name] {
			return fmt.Errorf("E2E_BRIEF_INVALID: duplicate environment input name %q", env.Name)
		}
		envNames[env.Name] = true
		switch env.Source {
		case "LITERAL":
			if strings.TrimSpace(env.Value) == "" || env.ServiceID != "" || env.CredentialName != "" {
				return fmt.Errorf("E2E_BRIEF_INVALID: LITERAL environment %q must provide only Value", env.Name)
			}
		case "SERVICE":
			if strings.TrimSpace(env.ServiceID) == "" || env.Value != "" || env.CredentialName != "" {
				return fmt.Errorf("E2E_BRIEF_INVALID: SERVICE environment %q must provide only ServiceID", env.Name)
			}
		case "CREDENTIAL":
			if strings.TrimSpace(env.CredentialName) == "" || env.Value != "" || env.ServiceID != "" {
				return fmt.Errorf("E2E_BRIEF_INVALID: CREDENTIAL environment %q must provide only CredentialName", env.Name)
			}
		default:
			return fmt.Errorf("E2E_BRIEF_INVALID: invalid environment source %q for %q", env.Source, env.Name)
		}
	}

	// Scenarios validation
	if len(brief.Scenarios) == 0 {
		return errors.New("E2E_BRIEF_INVALID: scenarios cannot be empty")
	}
	scenarioMap := make(map[string]bool)
	for _, sc := range brief.Scenarios {
		if strings.TrimSpace(sc.ID) == "" {
			return errors.New("E2E_BRIEF_INVALID: scenario ID cannot be empty")
		}
		if !utf8.ValidString(sc.ID) {
			return fmt.Errorf("E2E_BRIEF_INVALID: scenario_id %q contains invalid UTF-8", sc.ID)
		}
		for _, r := range sc.ID {
			if isIllegalControlOrLineChar(r) {
				return fmt.Errorf("E2E_BRIEF_INVALID: scenario_id %q contains illegal control character", sc.ID)
			}
		}
		if scenarioMap[sc.ID] {
			return fmt.Errorf("E2E_BRIEF_INVALID: duplicate scenario ID %q", sc.ID)
		}
		scenarioMap[sc.ID] = true
		if len(sc.Steps) == 0 {
			return fmt.Errorf("E2E_BRIEF_INVALID: scenario %q must define at least one step", sc.ID)
		}
		for i, step := range sc.Steps {
			if strings.TrimSpace(step) == "" {
				return fmt.Errorf("E2E_BRIEF_INVALID: scenario %q step %d cannot be empty", sc.ID, i+1)
			}
		}
		if strings.TrimSpace(sc.ExpectedResult) == "" {
			return fmt.Errorf("E2E_BRIEF_INVALID: scenario %q expected_result cannot be empty", sc.ID)
		}
		if strings.TrimSpace(sc.Verification) == "" {
			return fmt.Errorf("E2E_BRIEF_INVALID: scenario %q verification cannot be empty", sc.ID)
		}
	}

	// AllowedSkipIDs validation
	skipMap := make(map[string]bool)
	for _, skipID := range brief.AllowedSkipIDs {
		if !scenarioMap[skipID] {
			return fmt.Errorf("E2E_BRIEF_INVALID: allowed skip ID %q is not declared in scenarios", skipID)
		}
		if skipMap[skipID] {
			return fmt.Errorf("E2E_BRIEF_INVALID: duplicate allowed skip ID %q", skipID)
		}
		skipMap[skipID] = true
	}

	// Services validation
	if brief.Services == nil {
		return errors.New("E2E_BRIEF_INVALID: services is required (use explicit empty list [] if none)")
	}
	serviceMap := make(map[string]bool)
	for _, svc := range brief.Services {
		if strings.TrimSpace(svc.ID) == "" {
			return errors.New("E2E_BRIEF_INVALID: service ID cannot be empty")
		}
		if serviceMap[svc.ID] {
			return fmt.Errorf("E2E_BRIEF_INVALID: duplicate service ID %q", svc.ID)
		}
		serviceMap[svc.ID] = true
		switch svc.Kind {
		case "TESTCONTAINER", "LOCAL_PROCESS", "EXTERNAL_ENDPOINT":
		default:
			return fmt.Errorf("E2E_BRIEF_INVALID: invalid service kind %q for %q", svc.Kind, svc.ID)
		}
		if strings.TrimSpace(svc.ImageOrVersion) == "" {
			return fmt.Errorf("E2E_BRIEF_INVALID: service %q image_or_version cannot be empty", svc.ID)
		}
		if strings.TrimSpace(svc.ReadyCondition) == "" {
			return fmt.Errorf("E2E_BRIEF_INVALID: service %q ready_condition cannot be empty", svc.ID)
		}
		if svc.StartupTimeoutMS <= 0 {
			return fmt.Errorf("E2E_BRIEF_INVALID: service %q startup_timeout_ms must be > 0", svc.ID)
		}
	}

	// Cross-check environment service IDs
	for _, env := range brief.Environment {
		if env.Source == "SERVICE" && !serviceMap[env.ServiceID] {
			return fmt.Errorf("E2E_BRIEF_INVALID: environment %q references undeclared service %q", env.Name, env.ServiceID)
		}
	}

	// Cleanup obligations validation
	if brief.CleanupObligations == nil {
		return errors.New("E2E_BRIEF_INVALID: cleanup_obligations is required (use explicit empty list if none)")
	}
	serviceObligations := make(map[string]int)
	for _, obl := range brief.CleanupObligations {
		if strings.TrimSpace(obl.ResourceID) == "" {
			return errors.New("E2E_BRIEF_INVALID: cleanup obligation resource_id cannot be empty")
		}
		switch obl.Kind {
		case "SERVICE", "TEST_DATA", "PROCESS", "NAMESPACE", "CREDENTIAL", "OTHER":
		default:
			return fmt.Errorf("E2E_BRIEF_INVALID: invalid cleanup obligation kind %q for %q", obl.Kind, obl.ResourceID)
		}
		if strings.TrimSpace(obl.Owner) == "" {
			return fmt.Errorf("E2E_BRIEF_INVALID: cleanup obligation %q owner cannot be empty", obl.ResourceID)
		}
		if strings.TrimSpace(obl.Action) == "" {
			return fmt.Errorf("E2E_BRIEF_INVALID: cleanup obligation %q action cannot be empty", obl.ResourceID)
		}
		if obl.Kind == "SERVICE" {
			serviceObligations[obl.ResourceID]++
		}
	}
	for svcID := range serviceMap {
		count := serviceObligations[svcID]
		if count != 1 {
			return fmt.Errorf("E2E_BRIEF_INVALID: declared service %q has %d SERVICE cleanup obligations (exactly 1 required)", svcID, count)
		}
	}

	// CredentialPolicy validation
	if brief.CredentialPolicy == nil {
		return errors.New("E2E_BRIEF_INVALID: credential_policy is required (use explicit empty list if none)")
	}
	credPolicyMap := make(map[string]E2ECredentialPolicyEntry)
	var realOnlyCount int
	for _, cp := range brief.CredentialPolicy {
		if strings.TrimSpace(cp.Name) == "" {
			return errors.New("E2E_BRIEF_INVALID: credential policy name cannot be empty")
		}
		if !asciiKebabPattern.MatchString(cp.Name) {
			return fmt.Errorf("E2E_BRIEF_INVALID: credential policy name %q must be ASCII-kebab", cp.Name)
		}
		if _, exists := credPolicyMap[cp.Name]; exists {
			return fmt.Errorf("E2E_BRIEF_INVALID: duplicate credential policy name %q", cp.Name)
		}
		credPolicyMap[cp.Name] = cp
		switch cp.GenerationMode {
		case "THROWAWAY":
			if strings.TrimSpace(cp.GenerationInstruction) == "" {
				return fmt.Errorf("E2E_BRIEF_INVALID: THROWAWAY credential %q must provide generation_instruction", cp.Name)
			}
		case "REAL_ONLY":
			realOnlyCount++
		default:
			return fmt.Errorf("E2E_BRIEF_INVALID: invalid credential generation mode %q for %q", cp.GenerationMode, cp.Name)
		}
	}

	// Cross-check environment credentials
	for _, env := range brief.Environment {
		if env.Source == "CREDENTIAL" {
			if _, exists := credPolicyMap[env.CredentialName]; !exists {
				return fmt.Errorf("E2E_BRIEF_INVALID: environment %q references undeclared credential %q", env.Name, env.CredentialName)
			}
		}
	}

	// CredentialAuthorization validation
	if realOnlyCount == 0 {
		if len(brief.CredentialAuthorization) > 0 {
			return errors.New("E2E_BRIEF_INVALID: credential_authorization must be empty when no REAL_ONLY credentials exist")
		}
	} else {
		if len(brief.CredentialAuthorization) != realOnlyCount {
			return fmt.Errorf("E2E_BRIEF_INVALID: credential_authorization count %d does not match REAL_ONLY credentials count %d", len(brief.CredentialAuthorization), realOnlyCount)
		}
		if !state.CredentialAuthorizationVerified {
			return errors.New("E2E_BRIEF_INVALID: credential authorization not verified by controller/planner")
		}
		for i := 0; i < len(brief.CredentialAuthorization)-1; i++ {
			if brief.CredentialAuthorization[i].Name >= brief.CredentialAuthorization[i+1].Name {
				return errors.New("E2E_BRIEF_INVALID: credential_authorization entries must be uniquely sorted by name")
			}
		}
		for _, ca := range brief.CredentialAuthorization {
			cp, exists := credPolicyMap[ca.Name]
			if !exists || cp.GenerationMode != "REAL_ONLY" {
				return fmt.Errorf("E2E_BRIEF_INVALID: authorized credential %q is not a declared REAL_ONLY credential", ca.Name)
			}
			if strings.TrimSpace(ca.Purpose) == "" {
				return fmt.Errorf("E2E_BRIEF_INVALID: credential authorization for %q purpose cannot be empty", ca.Name)
			}
			if ca.Lifetime != "SINGLE_RUN" {
				return fmt.Errorf("E2E_BRIEF_INVALID: credential authorization for %q lifetime must be SINGLE_RUN, got %q", ca.Name, ca.Lifetime)
			}
			if strings.TrimSpace(ca.AuthorizedBy) == "" {
				return fmt.Errorf("E2E_BRIEF_INVALID: credential authorization for %q authorized_by cannot be empty", ca.Name)
			}
			if strings.TrimSpace(ca.AuthorizedAt) == "" {
				return fmt.Errorf("E2E_BRIEF_INVALID: credential authorization for %q authorized_at cannot be empty", ca.Name)
			}
			_, err := time.Parse(time.RFC3339, ca.AuthorizedAt)
			if err != nil || !strings.HasSuffix(ca.AuthorizedAt, "Z") {
				return fmt.Errorf("E2E_BRIEF_INVALID: credential authorization for %q authorized_at must be valid UTC RFC3339, got %q", ca.Name, ca.AuthorizedAt)
			}
			if strings.TrimSpace(ca.AuthorizationRef) == "" {
				return fmt.Errorf("E2E_BRIEF_INVALID: credential authorization for %q authorization_ref cannot be empty", ca.Name)
			}
		}
	}

	// PriorCleanupResolution validation
	if state.PriorCleanup.Required {
		if brief.PriorCleanupResolution == nil {
			return errors.New("E2E_BRIEF_INVALID: prior cleanup resolution is required but missing from brief")
		}
		res := brief.PriorCleanupResolution
		if res.PriorRunID != state.PriorCleanup.PriorRunID {
			return fmt.Errorf("E2E_BRIEF_INVALID: prior cleanup run_id mismatch: expected %q, got %q", state.PriorCleanup.PriorRunID, res.PriorRunID)
		}
		if res.PriorCandidateRef != state.PriorCleanup.PriorCandidateRef {
			return fmt.Errorf("E2E_BRIEF_INVALID: prior cleanup candidate_ref mismatch: expected %q, got %q", state.PriorCleanup.PriorCandidateRef, res.PriorCandidateRef)
		}
		if len(res.ResourceIDs) != len(state.PriorCleanup.UnresolvedResourceIDs) {
			return fmt.Errorf("E2E_BRIEF_INVALID: prior cleanup resource IDs count mismatch: expected %d, got %d", len(state.PriorCleanup.UnresolvedResourceIDs), len(res.ResourceIDs))
		}
		for i := 0; i < len(res.ResourceIDs)-1; i++ {
			if res.ResourceIDs[i] >= res.ResourceIDs[i+1] {
				return errors.New("E2E_BRIEF_INVALID: prior cleanup resolution resource_ids must be strictly sorted and unique")
			}
		}
		for i, expectedID := range state.PriorCleanup.UnresolvedResourceIDs {
			if res.ResourceIDs[i] != expectedID {
				return fmt.Errorf("E2E_BRIEF_INVALID: prior cleanup resource ID mismatch at index %d: expected %q, got %q", i, expectedID, res.ResourceIDs[i])
			}
		}
		if res.Disposition != "REMOVED" && res.Disposition != "ISOLATED" {
			return fmt.Errorf("E2E_BRIEF_INVALID: prior cleanup disposition must be 'REMOVED' or 'ISOLATED', got %q", res.Disposition)
		}
		if strings.TrimSpace(res.Action) == "" {
			return errors.New("E2E_BRIEF_INVALID: prior cleanup resolution action cannot be empty")
		}
		if strings.TrimSpace(res.HumanApprover) == "" {
			return errors.New("E2E_BRIEF_INVALID: prior cleanup resolution human_approver cannot be empty")
		}
		if strings.TrimSpace(res.ResolvedAt) == "" {
			return errors.New("E2E_BRIEF_INVALID: prior cleanup resolution resolved_at cannot be empty")
		}
		_, err := time.Parse(time.RFC3339, res.ResolvedAt)
		if err != nil || !strings.HasSuffix(res.ResolvedAt, "Z") {
			return fmt.Errorf("E2E_BRIEF_INVALID: prior cleanup resolution resolved_at must be valid UTC RFC3339, got %q", res.ResolvedAt)
		}
		if strings.TrimSpace(res.EvidenceURI) == "" || !evidenceURIPattern.MatchString(res.EvidenceURI) {
			return fmt.Errorf("E2E_BRIEF_INVALID: prior cleanup resolution evidence_uri must be valid URI, got %q", res.EvidenceURI)
		}
		if !state.PriorCleanup.HumanResolutionVerified {
			return errors.New("E2E_BRIEF_INVALID: prior cleanup resolution not verified by controller/planner")
		}
	} else if brief.PriorCleanupResolution != nil {
		return errors.New("E2E_BRIEF_INVALID: prior cleanup resolution provided but no prior cleanup was required")
	}

	return nil
}

// ValidateE2EReport validates that an E2EReportPayload adheres to all correlation, count reconciliation,
// embedded scenario results, cleanup, provenance, runner-loss mapping, and finalization contracts.
func ValidateE2EReport(brief E2EBriefContext, report E2EReportPayload, state E2EReportValidationContext) error {
	if brief.CredentialPolicy == nil {
		return errors.New("E2E_REPORT_INVALID: credential_policy is required (use explicit empty list [] if none)")
	}
	if brief.Services == nil {
		return errors.New("E2E_REPORT_INVALID: services is required (use explicit empty list [] if none)")
	}

	// 1. Controller ledger finalization context
	switch state.Finalization.State {
	case "FIRST_FINALIZATION", "DUPLICATE_IDENTICAL":
	case "CONFLICTING":
		return errors.New("E2E_REPORT_INVALID: conflicting finalization report rejected")
	case "LATE_AFTER_FINAL":
		return errors.New("E2E_REPORT_INVALID: late report received after terminal finalization rejected")
	default:
		return fmt.Errorf("E2E_REPORT_INVALID: invalid finalization state %q", state.Finalization.State)
	}

	if !state.Finalization.TerminalReportImmutable {
		return errors.New("E2E_REPORT_INVALID: terminal report must be marked immutable")
	}

	if !isValidAuthority(state.Finalization.Authority) {
		return fmt.Errorf("E2E_REPORT_INVALID: invalid finalization authority %q (must be VERIFIER or CONTROLLER_RUNNER_LOST)", state.Finalization.Authority)
	}

	expectedDigest := ComputeReportDigest(report)
	incomingClean := strings.TrimPrefix(state.Finalization.IncomingReportDigest, "sha256:")
	if incomingClean != expectedDigest {
		return fmt.Errorf("E2E_REPORT_INVALID: incoming report digest mismatch: expected %s, got %s", expectedDigest, state.Finalization.IncomingReportDigest)
	}

	switch state.Finalization.State {
	case "FIRST_FINALIZATION":
		if state.Finalization.ExistingTerminalReportDigest != "" {
			return errors.New("E2E_REPORT_INVALID: FIRST_FINALIZATION requires empty ExistingTerminalReportDigest")
		}
		if state.Finalization.ExistingTerminalReportAuthority != "" {
			return errors.New("E2E_REPORT_INVALID: FIRST_FINALIZATION requires empty ExistingTerminalReportAuthority")
		}
		if report.StageOutcome == "RUNNER_LOST" {
			if state.Finalization.Authority != "CONTROLLER_RUNNER_LOST" {
				return fmt.Errorf("E2E_REPORT_INVALID: RUNNER_LOST report requires authority CONTROLLER_RUNNER_LOST, got %s", state.Finalization.Authority)
			}
		} else {
			if state.Finalization.Authority != "VERIFIER" {
				return fmt.Errorf("E2E_REPORT_INVALID: non-RUNNER_LOST report requires authority VERIFIER, got %s", state.Finalization.Authority)
			}
		}
	case "DUPLICATE_IDENTICAL":
		if state.Finalization.ExistingTerminalReportDigest == "" {
			return errors.New("E2E_REPORT_INVALID: DUPLICATE_IDENTICAL requires non-empty ExistingTerminalReportDigest")
		}
		existingClean := strings.TrimPrefix(state.Finalization.ExistingTerminalReportDigest, "sha256:")
		if incomingClean != existingClean {
			return errors.New("E2E_REPORT_INVALID: DUPLICATE_IDENTICAL digest does not match existing terminal digest")
		}
		if !isValidAuthority(state.Finalization.ExistingTerminalReportAuthority) {
			return fmt.Errorf("E2E_REPORT_INVALID: invalid existing terminal report authority %q (must be VERIFIER or CONTROLLER_RUNNER_LOST)", state.Finalization.ExistingTerminalReportAuthority)
		}
		if state.Finalization.Authority != state.Finalization.ExistingTerminalReportAuthority {
			return fmt.Errorf("E2E_REPORT_INVALID: DUPLICATE_IDENTICAL authority %q does not match existing terminal authority %q", state.Finalization.Authority, state.Finalization.ExistingTerminalReportAuthority)
		}
	}

	// 2. Controller ledger and prior cleanup consistency
	if !state.RunIDUnique {
		return errors.New("E2E_REPORT_INVALID: run_id is not unique in ledger")
	}
	if !state.PriorCleanupResolved {
		return errors.New("E2E_REPORT_INVALID: prior cleanup is not resolved")
	}

	// 3. Correlation with brief
	if report.RunID != brief.RunID {
		return fmt.Errorf("E2E_REPORT_INVALID: run_id mismatch: report %q != brief %q", report.RunID, brief.RunID)
	}
	if report.RepoRef != brief.RepoRef {
		return fmt.Errorf("E2E_REPORT_INVALID: repo_ref mismatch: report %q != brief %q", report.RepoRef, brief.RepoRef)
	}
	if report.CandidateRef != brief.CandidateRef {
		return fmt.Errorf("E2E_REPORT_INVALID: candidate_ref mismatch: report %q != brief %q", report.CandidateRef, brief.CandidateRef)
	}
	if !candidateRefPattern.MatchString(report.CandidateRef) {
		return fmt.Errorf("E2E_REPORT_INVALID: candidate_ref %q is invalid; must be an immutable 40-character lowercase hex commit SHA", report.CandidateRef)
	}
	if report.TestPackage != brief.TestPackage {
		return fmt.Errorf("E2E_REPORT_INVALID: test_package mismatch: report %q != brief %q", report.TestPackage, brief.TestPackage)
	}

	// 4. Diagnostic check
	if strings.TrimSpace(report.Diagnostic) == "" {
		return errors.New("E2E_REPORT_INVALID: diagnostic cannot be empty")
	}
	if len([]byte(report.Diagnostic)) > 160 {
		return fmt.Errorf("E2E_REPORT_INVALID: diagnostic length %d bytes exceeds 160-byte limit", len([]byte(report.Diagnostic)))
	}

	// 5. String fields UTF-8 and control/line-separator character validation in deterministic catalog order
	type reportStringField struct {
		name string
		val  string
	}
	reportFields := []reportStringField{
		{"run_id", report.RunID},
		{"repo_ref", report.RepoRef},
		{"candidate_ref", report.CandidateRef},
		{"test_package", report.TestPackage},
		{"stage_outcome", report.StageOutcome},
		{"build_outcome", report.BuildOutcome},
		{"execution_outcome", report.ExecutionOutcome},
		{"result_completeness", report.ResultCompleteness},
		{"tested_artifact_ref", report.TestedArtifactRef},
		{"execution_material_ref", report.ExecutionMaterialRef},
		{"diagnostic", report.Diagnostic},
		{"evidence_uri", report.EvidenceURI},
		{"cleanup_status", report.CleanupStatus},
		{"cleanup_evidence_uri", report.CleanupEvidenceURI},
		{"secrets_attestation", report.SecretsAttestation},
	}
	for _, f := range reportFields {
		if !utf8.ValidString(f.val) {
			return fmt.Errorf("E2E_REPORT_INVALID: field %s contains invalid UTF-8", f.name)
		}
		for _, r := range f.val {
			if isIllegalControlOrLineChar(r) {
				return fmt.Errorf("E2E_REPORT_INVALID: field %s contains illegal control character", f.name)
			}
		}
	}
	for _, sr := range report.ScenarioResults {
		if !utf8.ValidString(sr.ScenarioID) {
			return fmt.Errorf("E2E_REPORT_INVALID: scenario_id %q contains invalid UTF-8", sr.ScenarioID)
		}
		for _, r := range sr.ScenarioID {
			if isIllegalControlOrLineChar(r) {
				return fmt.Errorf("E2E_REPORT_INVALID: scenario_id %q contains illegal control character", sr.ScenarioID)
			}
		}
		if !utf8.ValidString(sr.Outcome) {
			return fmt.Errorf("E2E_REPORT_INVALID: scenario outcome %q contains invalid UTF-8", sr.Outcome)
		}
		for _, r := range sr.Outcome {
			if isIllegalControlOrLineChar(r) {
				return fmt.Errorf("E2E_REPORT_INVALID: scenario outcome %q contains illegal control character", sr.Outcome)
			}
		}
	}

	// Check if REAL_ONLY credentials exist in brief
	var hasRealOnly bool
	for _, cp := range brief.CredentialPolicy {
		if cp.GenerationMode == "REAL_ONLY" {
			hasRealOnly = true
			break
		}
	}
	if hasRealOnly && !state.CredentialAuthorizationVerified {
		return errors.New("E2E_REPORT_INVALID: credential authorization must be verified when REAL_ONLY credentials are declared")
	}

	// 6. Lifecycle preflight outcome check
	if state.PreflightOutcome != "READY" && state.PreflightOutcome != "AUTHORIZATION_DENIED" {
		return fmt.Errorf("E2E_REPORT_INVALID: invalid preflight_outcome %q (must be READY or AUTHORIZATION_DENIED)", state.PreflightOutcome)
	}

	// PASS outcomes require READY preflight and ExecutionStarted
	if report.StageOutcome == "PASS" || report.ExecutionOutcome == "PASS" {
		if state.PreflightOutcome != "READY" {
			return fmt.Errorf("E2E_REPORT_INVALID: outcome PASS requires preflight_outcome READY, got %s", state.PreflightOutcome)
		}
		if !state.ExecutionStarted {
			return errors.New("E2E_REPORT_INVALID: outcome PASS requires execution_started to be true")
		}
	}
	if report.BuildOutcome == "PASS" && state.PreflightOutcome != "READY" {
		return fmt.Errorf("E2E_REPORT_INVALID: outcome PASS requires preflight_outcome READY, got %s", state.PreflightOutcome)
	}

	// 7. Process states validation
	switch state.BuildProcessState {
	case "NOT_STARTED":
		if report.BuildOutcome != "NOT_RUN" {
			return fmt.Errorf("E2E_REPORT_INVALID: build_process_state NOT_STARTED requires build_outcome NOT_RUN, got %s", report.BuildOutcome)
		}
	case "RUNNING":
		if report.BuildOutcome != "UNKNOWN" {
			return fmt.Errorf("E2E_REPORT_INVALID: build_process_state RUNNING requires build_outcome UNKNOWN, got %s", report.BuildOutcome)
		}
	case "PASS", "FAIL", "ENV_ERROR", "TIMEOUT", "CANCELLED":
		if report.BuildOutcome != state.BuildProcessState {
			return fmt.Errorf("E2E_REPORT_INVALID: build_outcome %s does not match durable build_process_state %s", report.BuildOutcome, state.BuildProcessState)
		}
	default:
		return fmt.Errorf("E2E_REPORT_INVALID: invalid build_process_state %q", state.BuildProcessState)
	}

	switch state.ExecutionProcessState {
	case "NOT_STARTED":
		if state.PreflightOutcome == "AUTHORIZATION_DENIED" {
			if report.StageOutcome == "RUNNER_LOST" {
				if report.ExecutionOutcome != "NOT_RUN" {
					return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED RUNNER_LOST execution_process_state NOT_STARTED requires execution_outcome NOT_RUN, got %s", report.ExecutionOutcome)
				}
			} else {
				if report.ExecutionOutcome != "ENV_ERROR" {
					return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED execution_process_state NOT_STARTED requires execution_outcome ENV_ERROR, got %s", report.ExecutionOutcome)
				}
			}
		} else {
			if report.ExecutionOutcome != "NOT_RUN" {
				return fmt.Errorf("E2E_REPORT_INVALID: execution_process_state NOT_STARTED requires execution_outcome NOT_RUN, got %s", report.ExecutionOutcome)
			}
		}
	case "RUNNING":
		if report.ExecutionOutcome != "UNKNOWN" {
			return fmt.Errorf("E2E_REPORT_INVALID: execution_process_state RUNNING requires execution_outcome UNKNOWN, got %s", report.ExecutionOutcome)
		}
	case "PASS", "FAIL", "ENV_ERROR", "TIMEOUT", "CANCELLED":
		if report.ExecutionOutcome != state.ExecutionProcessState {
			return fmt.Errorf("E2E_REPORT_INVALID: execution_outcome %s does not match durable execution_process_state %s", report.ExecutionOutcome, state.ExecutionProcessState)
		}
	default:
		return fmt.Errorf("E2E_REPORT_INVALID: invalid execution_process_state %q", state.ExecutionProcessState)
	}

	// CredentialProvisioningState validation
	switch state.CredentialProvisioningState {
	case "NOT_STARTED", "SUCCEEDED", "FAILED", "UNKNOWN":
	default:
		return fmt.Errorf("E2E_REPORT_INVALID: invalid credential_provisioning_state %q", state.CredentialProvisioningState)
	}
	if len(brief.CredentialPolicy) == 0 {
		if state.CredentialProvisioningState != "NOT_STARTED" {
			return fmt.Errorf("E2E_REPORT_INVALID: empty credential policy requires credential provisioning state NOT_STARTED, got %s", state.CredentialProvisioningState)
		}
	}

	// Execution start/attempt itself requires independent bounded cleanup record,
	// even with no services, credentials, or allocated IDs.
	if state.ExecutionStarted || state.ExecutionProcessState != "NOT_STARTED" ||
		state.ServiceStartupAttempted || state.CredentialProvisioningState != "NOT_STARTED" ||
		state.ResourcesAllocated || len(state.AllocatedResourceIDs) > 0 {
		if !state.CleanupRequired {
			return errors.New("E2E_REPORT_INVALID: contradictory lifecycle state: execution/startup/provisioning/allocation requires cleanup_required to be true")
		}
	}

	// ResourcesAllocated consistency
	if state.ResourcesAllocated && len(state.AllocatedResourceIDs) == 0 {
		return errors.New("E2E_REPORT_INVALID: resources_allocated is true but allocated_resource_ids is empty")
	}
	if !state.ResourcesAllocated && len(state.AllocatedResourceIDs) > 0 {
		return errors.New("E2E_REPORT_INVALID: resources_allocated is false but allocated_resource_ids is non-empty")
	}

	// Preflight outcome check for AUTHORIZATION_DENIED
	if state.PreflightOutcome == "AUTHORIZATION_DENIED" {
		if !state.CredentialAuthorizationVerified {
			return errors.New("E2E_REPORT_INVALID: AUTHORIZATION_DENIED preflight requires verified controller credential authorization context")
		}
		if len(brief.CredentialAuthorization) == 0 {
			return errors.New("E2E_REPORT_INVALID: AUTHORIZATION_DENIED preflight requires explicit per-name human REAL_ONLY authorization records in brief")
		}
		var realOnlyNames []string
		for _, cp := range brief.CredentialPolicy {
			if cp.GenerationMode == "REAL_ONLY" {
				realOnlyNames = append(realOnlyNames, cp.Name)
			}
		}
		if len(realOnlyNames) == 0 {
			return errors.New("E2E_REPORT_INVALID: AUTHORIZATION_DENIED preflight requires declared REAL_ONLY credentials in credential policy")
		}
		sort.Strings(realOnlyNames)
		authNames := make([]string, len(brief.CredentialAuthorization))
		for i, ca := range brief.CredentialAuthorization {
			authNames[i] = ca.Name
		}
		sort.Strings(authNames)
		if len(authNames) != len(realOnlyNames) {
			return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED authorization records count %d does not match REAL_ONLY credentials count %d", len(authNames), len(realOnlyNames))
		}
		for i := range realOnlyNames {
			if authNames[i] != realOnlyNames[i] {
				return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED authorization name %q does not match declared REAL_ONLY credential %q", authNames[i], realOnlyNames[i])
			}
		}
		for _, ca := range brief.CredentialAuthorization {
			if strings.TrimSpace(ca.Name) == "" || strings.TrimSpace(ca.Purpose) == "" || ca.Lifetime != "SINGLE_RUN" ||
				strings.TrimSpace(ca.AuthorizedBy) == "" || strings.TrimSpace(ca.AuthorizedAt) == "" || strings.TrimSpace(ca.AuthorizationRef) == "" {
				return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED requires well-formed authorization record for %q", ca.Name)
			}
			if _, err := time.Parse(time.RFC3339, ca.AuthorizedAt); err != nil || !strings.HasSuffix(ca.AuthorizedAt, "Z") {
				return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED authorization for %q has invalid UTC RFC3339 timestamp %q", ca.Name, ca.AuthorizedAt)
			}
		}
		if state.ExecutionStarted || state.ServiceStartupAttempted || state.CredentialProvisioningState != "NOT_STARTED" ||
			state.ResourcesAllocated || state.CleanupRequired ||
			len(state.AllocatedResourceIDs) > 0 || len(state.UnresolvedResourceIDs) > 0 {
			return errors.New("E2E_REPORT_INVALID: AUTHORIZATION_DENIED preflight cannot attempt execution, service startup, credential provisioning, or allocate resources")
		}
		if state.BuildProcessState != "NOT_STARTED" || state.ExecutionProcessState != "NOT_STARTED" {
			return errors.New("E2E_REPORT_INVALID: AUTHORIZATION_DENIED preflight requires both build_process_state and execution_process_state to be NOT_STARTED")
		}
		if state.CleanupStatus != "CONFIRMED" {
			return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED requires controller cleanup_status CONFIRMED, got %s", state.CleanupStatus)
		}
		if report.StageOutcome == "RUNNER_LOST" {
			if report.BuildOutcome != "NOT_RUN" {
				return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED RUNNER_LOST requires build_outcome NOT_RUN, got %s", report.BuildOutcome)
			}
			if report.ExecutionOutcome != "NOT_RUN" {
				return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED RUNNER_LOST requires execution_outcome NOT_RUN, got %s", report.ExecutionOutcome)
			}
		} else {
			if report.StageOutcome != "ENV_BLOCKED" {
				return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED requires stage_outcome ENV_BLOCKED or RUNNER_LOST, got %s", report.StageOutcome)
			}
			if report.BuildOutcome != "NOT_RUN" {
				return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED requires build_outcome NOT_RUN, got %s", report.BuildOutcome)
			}
			if report.ExecutionOutcome != "ENV_ERROR" {
				return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED requires execution_outcome ENV_ERROR, got %s", report.ExecutionOutcome)
			}
		}
		if report.ResultCompleteness != "NONE" {
			return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED requires result_completeness NONE, got %s", report.ResultCompleteness)
		}
		if report.TestedArtifactRef != "NONE" {
			return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED requires tested_artifact_ref NONE, got %s", report.TestedArtifactRef)
		}
		if report.ExecutionMaterialRef != "UNAVAILABLE" {
			return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED requires execution_material_ref UNAVAILABLE, got %s", report.ExecutionMaterialRef)
		}
		if report.EvidenceURI != "UNAVAILABLE" {
			return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED requires evidence_uri UNAVAILABLE, got %s", report.EvidenceURI)
		}
		if report.CleanupStatus != "CONFIRMED" {
			return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED requires cleanup_status CONFIRMED, got %s", report.CleanupStatus)
		}
		if report.CleanupEvidenceURI != "NOT_REQUIRED" {
			return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED requires cleanup_evidence_uri NOT_REQUIRED, got %s", report.CleanupEvidenceURI)
		}
		if report.SecretsAttestation != "ENV_BLOCKED:AUTHORIZATION_DENIED" {
			return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED requires secrets_attestation ENV_BLOCKED:AUTHORIZATION_DENIED, got %s", report.SecretsAttestation)
		}
		if report.ScenariosPassed != 0 || report.ScenariosFailed != 0 || report.ScenariosSkipped != 0 || report.ScenariosUnknown != 0 || report.ScenariosNotRun != report.ScenariosSelected {
			return fmt.Errorf("E2E_REPORT_INVALID: AUTHORIZATION_DENIED requires all scenarios NOT_RUN, got passed=%d, failed=%d, skipped=%d, unknown=%d, not_run=%d", report.ScenariosPassed, report.ScenariosFailed, report.ScenariosSkipped, report.ScenariosUnknown, report.ScenariosNotRun)
		}
	}

	// 8. Allowed outcome enums
	switch report.StageOutcome {
	case "PASS", "BUILD_FAILURE", "PRODUCT_FAILURE", "ENV_BLOCKED", "CLARIFICATION_REQUIRED", "TIMEOUT", "CANCELLED", "RUNNER_LOST", "INVALID_EVIDENCE":
	default:
		return fmt.Errorf("E2E_REPORT_INVALID: invalid stage_outcome %q", report.StageOutcome)
	}

	switch report.BuildOutcome {
	case "PASS", "FAIL", "NOT_RUN", "UNKNOWN", "ENV_ERROR", "TIMEOUT", "CANCELLED":
	default:
		return fmt.Errorf("E2E_REPORT_INVALID: invalid build_outcome %q", report.BuildOutcome)
	}

	switch report.ExecutionOutcome {
	case "PASS", "FAIL", "NOT_RUN", "UNKNOWN", "ENV_ERROR", "TIMEOUT", "CANCELLED", "INVALID_EVIDENCE":
	default:
		return fmt.Errorf("E2E_REPORT_INVALID: invalid execution_outcome %q", report.ExecutionOutcome)
	}

	switch report.ResultCompleteness {
	case "COMPLETE", "PARTIAL", "NONE":
	default:
		return fmt.Errorf("E2E_REPORT_INVALID: invalid result_completeness %q", report.ResultCompleteness)
	}

	// 9. Embedded Scenario Results and count reconciliation
	if len(report.ScenarioResults) != len(brief.Scenarios) {
		return fmt.Errorf("E2E_REPORT_INVALID: scenario results count %d does not match brief scenarios count %d", len(report.ScenarioResults), len(brief.Scenarios))
	}
	briefScenarioMap := make(map[string]bool)
	for _, sc := range brief.Scenarios {
		briefScenarioMap[sc.ID] = true
	}
	var passedCount, failedCount, skippedCount, notRunCount, unknownCount int
	seenResults := make(map[string]bool)
	for _, res := range report.ScenarioResults {
		if !briefScenarioMap[res.ScenarioID] {
			return fmt.Errorf("E2E_REPORT_INVALID: result for undeclared scenario %q", res.ScenarioID)
		}
		if seenResults[res.ScenarioID] {
			return fmt.Errorf("E2E_REPORT_INVALID: duplicate scenario result for %q", res.ScenarioID)
		}
		seenResults[res.ScenarioID] = true
		switch res.Outcome {
		case "PASS":
			passedCount++
		case "FAIL":
			failedCount++
		case "SKIP":
			skippedCount++
		case "NOT_RUN":
			notRunCount++
		case "UNKNOWN":
			unknownCount++
		default:
			return fmt.Errorf("E2E_REPORT_INVALID: invalid scenario outcome %q for %q", res.Outcome, res.ScenarioID)
		}
	}
	if report.ScenariosSelected <= 0 {
		return errors.New("E2E_REPORT_INVALID: scenarios_selected must be greater than 0")
	}
	if report.ScenariosSelected != len(brief.Scenarios) {
		return fmt.Errorf("E2E_REPORT_INVALID: scenarios_selected %d does not match brief scenarios count %d", report.ScenariosSelected, len(brief.Scenarios))
	}
	if report.ScenariosPassed != passedCount {
		return fmt.Errorf("E2E_REPORT_INVALID: scenarios_passed count %d does not match results count %d", report.ScenariosPassed, passedCount)
	}
	if report.ScenariosFailed != failedCount {
		return fmt.Errorf("E2E_REPORT_INVALID: scenarios_failed count %d does not match results count %d", report.ScenariosFailed, failedCount)
	}
	if report.ScenariosSkipped != skippedCount {
		return fmt.Errorf("E2E_REPORT_INVALID: scenarios_skipped count %d does not match results count %d", report.ScenariosSkipped, skippedCount)
	}
	if report.ScenariosNotRun != notRunCount {
		return fmt.Errorf("E2E_REPORT_INVALID: scenarios_not_run count %d does not match results count %d", report.ScenariosNotRun, notRunCount)
	}
	if report.ScenariosUnknown != unknownCount {
		return fmt.Errorf("E2E_REPORT_INVALID: scenarios_unknown count %d does not match results count %d", report.ScenariosUnknown, unknownCount)
	}
	if report.ScenariosSelected != passedCount+failedCount+skippedCount+notRunCount+unknownCount {
		return fmt.Errorf("E2E_REPORT_INVALID: scenario count reconciliation failed: selected (%d) != passed (%d) + failed (%d) + skipped (%d) + not_run (%d) + unknown (%d)",
			report.ScenariosSelected, report.ScenariosPassed, report.ScenariosFailed, report.ScenariosSkipped, report.ScenariosNotRun, report.ScenariosUnknown)
	}

	// ResultCompleteness derivation
	switch {
	case unknownCount > 0:
		if report.ResultCompleteness != "PARTIAL" {
			return fmt.Errorf("E2E_REPORT_INVALID: presence of unknown scenario results requires result_completeness PARTIAL, got %s", report.ResultCompleteness)
		}
	case notRunCount == len(brief.Scenarios) && !state.ExecutionStarted:
		if report.ResultCompleteness != "NONE" {
			return fmt.Errorf("E2E_REPORT_INVALID: all scenarios not run without execution start requires result_completeness NONE, got %s", report.ResultCompleteness)
		}
	default:
		if report.ResultCompleteness != "COMPLETE" {
			return fmt.Errorf("E2E_REPORT_INVALID: all scenario outcomes known requires result_completeness COMPLETE, got %s", report.ResultCompleteness)
		}
	}

	// Allowed-skip policy
	allowedSkipSet := make(map[string]bool)
	for _, id := range brief.AllowedSkipIDs {
		allowedSkipSet[id] = true
	}
	for _, res := range report.ScenarioResults {
		if res.Outcome == "SKIP" && !allowedSkipSet[res.ScenarioID] && report.StageOutcome == "PASS" {
			return fmt.Errorf("E2E_REPORT_INVALID: unauthorized skipped scenario %q prohibits stage outcome PASS", res.ScenarioID)
		}
	}
	if (unknownCount > 0 || notRunCount > 0) && report.StageOutcome == "PASS" {
		return errors.New("E2E_REPORT_INVALID: unknown or not-run scenarios prohibit stage outcome PASS")
	}

	// 10. Cleanup verification
	switch report.CleanupStatus {
	case "CONFIRMED", "UNCONFIRMED":
	default:
		return fmt.Errorf("E2E_REPORT_INVALID: invalid cleanup_status %q", report.CleanupStatus)
	}
	switch state.CleanupStatus {
	case "CONFIRMED", "UNCONFIRMED":
	default:
		return fmt.Errorf("E2E_REPORT_INVALID: invalid controller cleanup_status %q", state.CleanupStatus)
	}
	if report.CleanupStatus != state.CleanupStatus {
		return fmt.Errorf("E2E_REPORT_INVALID: report cleanup_status %q does not match controller cleanup_status %q", report.CleanupStatus, state.CleanupStatus)
	}

	if !state.CleanupRequired {
		if state.ExecutionStarted || state.ExecutionProcessState != "NOT_STARTED" ||
			state.ServiceStartupAttempted || state.CredentialProvisioningState != "NOT_STARTED" ||
			state.ResourcesAllocated || len(state.AllocatedResourceIDs) > 0 || len(state.UnresolvedResourceIDs) > 0 {
			return errors.New("E2E_REPORT_INVALID: contradictory lifecycle state: cleanup not required but execution/startup/provisioning/allocation occurred")
		}
		if report.CleanupStatus != "CONFIRMED" {
			return fmt.Errorf("E2E_REPORT_INVALID: cleanup not required requires cleanup_status CONFIRMED, got %s", report.CleanupStatus)
		}
		if report.CleanupEvidenceURI != "NOT_REQUIRED" {
			return fmt.Errorf("E2E_REPORT_INVALID: cleanup not required requires cleanup_evidence_uri NOT_REQUIRED, got %s", report.CleanupEvidenceURI)
		}
		if state.CleanupStatus != "CONFIRMED" {
			return fmt.Errorf("E2E_REPORT_INVALID: cleanup not required requires controller cleanup_status CONFIRMED, got %s", state.CleanupStatus)
		}
	} else {
		if report.CleanupEvidenceURI == "NOT_REQUIRED" {
			return errors.New("E2E_REPORT_INVALID: cleanup required prohibits cleanup_evidence_uri NOT_REQUIRED")
		}
		if !evidenceURIPattern.MatchString(report.CleanupEvidenceURI) {
			return fmt.Errorf("E2E_REPORT_INVALID: cleanup required requires valid cleanup_evidence_uri, got %s", report.CleanupEvidenceURI)
		}
		if !state.CleanupRecordVerified {
			return errors.New("E2E_REPORT_INVALID: cleanup record must be verified by controller")
		}
		if state.CleanupStatus == "CONFIRMED" && len(state.UnresolvedResourceIDs) != 0 {
			return fmt.Errorf("E2E_REPORT_INVALID: cleanup_status CONFIRMED requires empty unresolved resource IDs, got %v", state.UnresolvedResourceIDs)
		}
		if state.CleanupStatus == "UNCONFIRMED" && len(state.UnresolvedResourceIDs) == 0 {
			return errors.New("E2E_REPORT_INVALID: cleanup_status UNCONFIRMED requires non-empty unresolved resource IDs")
		}
	}

	// 11. EvidenceURI validation
	if report.EvidenceURI == "UNAVAILABLE" {
		if report.StageOutcome == "PASS" {
			return errors.New("E2E_REPORT_INVALID: evidence_uri UNAVAILABLE is prohibited when stage_outcome is PASS")
		}
		if report.StageOutcome == "PRODUCT_FAILURE" {
			return errors.New("E2E_REPORT_INVALID: evidence_uri UNAVAILABLE is prohibited when stage_outcome is PRODUCT_FAILURE")
		}
		if report.ExecutionOutcome != "NOT_RUN" && report.ExecutionOutcome != "ENV_ERROR" {
			return fmt.Errorf("E2E_REPORT_INVALID: evidence_uri UNAVAILABLE requires execution_outcome NOT_RUN or ENV_ERROR, got %s", report.ExecutionOutcome)
		}
		if report.ResultCompleteness != "NONE" {
			return fmt.Errorf("E2E_REPORT_INVALID: evidence_uri UNAVAILABLE requires result_completeness NONE, got %s", report.ResultCompleteness)
		}
		if report.ScenariosPassed != 0 || report.ScenariosFailed != 0 {
			return fmt.Errorf("E2E_REPORT_INVALID: evidence_uri UNAVAILABLE requires scenarios_passed == 0 and scenarios_failed == 0, got passed=%d, failed=%d", report.ScenariosPassed, report.ScenariosFailed)
		}
	} else if !evidenceURIPattern.MatchString(report.EvidenceURI) {
		return fmt.Errorf("E2E_REPORT_INVALID: report evidence_uri %q is invalid; must be 'UNAVAILABLE' or a valid URI", report.EvidenceURI)
	}

	// 12. ExecutionMaterialRef validation
	if report.ExecutionMaterialRef == "UNAVAILABLE" {
		if report.StageOutcome == "PASS" {
			return errors.New("E2E_REPORT_INVALID: execution_material_ref UNAVAILABLE is prohibited when stage_outcome is PASS")
		}
	} else if !sha256PrefixPattern.MatchString(report.ExecutionMaterialRef) {
		return fmt.Errorf("E2E_REPORT_INVALID: execution_material_ref %q is invalid; must be 'UNAVAILABLE' or sha256:<64 lowercase hex>", report.ExecutionMaterialRef)
	}

	// 13. TestedArtifactRef validation
	if report.TestedArtifactRef == "NONE" {
		if report.StageOutcome == "PASS" || report.StageOutcome == "PRODUCT_FAILURE" {
			return fmt.Errorf("E2E_REPORT_INVALID: tested_artifact_ref cannot be NONE for stage outcome %s", report.StageOutcome)
		}
		if report.BuildOutcome == "PASS" && brief.Mode == "build-and-run" {
			return errors.New("E2E_REPORT_INVALID: tested_artifact_ref cannot be NONE when build outcome is PASS in build-and-run mode")
		}
	} else {
		match := testedArtifactRefPattern.FindStringSubmatch(report.TestedArtifactRef)
		if match == nil {
			return fmt.Errorf("E2E_REPORT_INVALID: tested_artifact_ref %q is invalid; must match canonical pattern sha256:<hex>;candidate=<40-hex>;build_material=sha256:<hex>;source=(FRESH|VERIFIED_REUSE)", report.TestedArtifactRef)
		}
		if match[2] != brief.CandidateRef {
			return fmt.Errorf("E2E_REPORT_INVALID: tested_artifact_ref candidate %q does not match brief candidate %q", match[2], brief.CandidateRef)
		}
	}

	// 14. SecretsAttestation validation mapped to CredentialProvisioningState
	if len(brief.CredentialPolicy) == 0 {
		expectedPolicyDigest := ComputePolicyDigest(brief.CredentialPolicy)
		expectedAttest := fmt.Sprintf("NOT_PROVISIONED sha256:%s", expectedPolicyDigest)
		if report.SecretsAttestation != expectedAttest {
			return fmt.Errorf("E2E_REPORT_INVALID: empty credential policy requires secrets_attestation %q, got %q", expectedAttest, report.SecretsAttestation)
		}
		if state.CredentialProvisioningState != "NOT_STARTED" {
			return fmt.Errorf("E2E_REPORT_INVALID: empty credential policy requires credential provisioning state NOT_STARTED, got %s", state.CredentialProvisioningState)
		}
	} else {
		switch state.CredentialProvisioningState {
		case "NOT_STARTED":
			switch {
			case state.PreflightOutcome == "AUTHORIZATION_DENIED":
				if report.SecretsAttestation != "ENV_BLOCKED:AUTHORIZATION_DENIED" {
					return fmt.Errorf("E2E_REPORT_INVALID: preflight AUTHORIZATION_DENIED requires secrets_attestation ENV_BLOCKED:AUTHORIZATION_DENIED, got %s", report.SecretsAttestation)
				}
			case report.SecretsAttestation == "ENV_BLOCKED:GENERATOR_UNAVAILABLE" && (report.StageOutcome == "ENV_BLOCKED" || report.StageOutcome == "RUNNER_LOST"):
				// known durable preflight/runner-loss generator unavailable block preserved
			default:
				expectedPolicyDigest := ComputePolicyDigest(brief.CredentialPolicy)
				expectedAttest := fmt.Sprintf("NOT_PROVISIONED sha256:%s", expectedPolicyDigest)
				if report.SecretsAttestation != expectedAttest {
					return fmt.Errorf("E2E_REPORT_INVALID: credential provisioning NOT_STARTED requires secrets_attestation %q, got %q", expectedAttest, report.SecretsAttestation)
				}
			}
		case "SUCCEEDED":
			canonical, err := ComputeCanonicalSecretsAttestation(brief.CredentialPolicy, brief.CredentialAuthorization)
			if err != nil {
				return fmt.Errorf("E2E_REPORT_INVALID: failed to compute canonical secrets attestation: %w", err)
			}
			if report.SecretsAttestation != canonical {
				return fmt.Errorf("E2E_REPORT_INVALID: secrets_attestation %q does not match canonical attestation %q", report.SecretsAttestation, canonical)
			}
		case "FAILED":
			if report.SecretsAttestation != "ENV_BLOCKED:PROVISIONING_FAILED" {
				return fmt.Errorf("E2E_REPORT_INVALID: failed credential provisioning requires secrets_attestation ENV_BLOCKED:PROVISIONING_FAILED, got %s", report.SecretsAttestation)
			}
			if report.StageOutcome != "ENV_BLOCKED" && report.StageOutcome != "RUNNER_LOST" {
				return fmt.Errorf("E2E_REPORT_INVALID: failed credential provisioning requires stage_outcome ENV_BLOCKED or RUNNER_LOST, got %s", report.StageOutcome)
			}
		case "UNKNOWN":
			if report.StageOutcome != "RUNNER_LOST" {
				return errors.New("E2E_REPORT_INVALID: credential provisioning state UNKNOWN is only permitted for stage_outcome RUNNER_LOST")
			}
			if report.SecretsAttestation != "UNVERIFIED:RUNNER_LOST" {
				return fmt.Errorf("E2E_REPORT_INVALID: credential provisioning state UNKNOWN requires secrets_attestation UNVERIFIED:RUNNER_LOST, got %s", report.SecretsAttestation)
			}
		}
	}

	// 15. Fail-closed cross-field stage outcome mappings
	switch report.StageOutcome {
	case "PASS":
		switch brief.Mode {
		case "build-and-run":
			if report.BuildOutcome == "NOT_RUN" {
				return errors.New("E2E_REPORT_INVALID: stage outcome PASS in build-and-run mode cannot have build outcome NOT_RUN")
			}
			if report.BuildOutcome != "PASS" {
				return fmt.Errorf("E2E_REPORT_INVALID: stage outcome PASS in build-and-run mode requires build outcome PASS, got %s", report.BuildOutcome)
			}
		case "run-only":
			if report.BuildOutcome != "NOT_RUN" {
				return fmt.Errorf("E2E_REPORT_INVALID: stage outcome PASS in run-only mode requires build outcome NOT_RUN, got %s", report.BuildOutcome)
			}
		}
		if report.ExecutionOutcome != "PASS" {
			return fmt.Errorf("E2E_REPORT_INVALID: stage outcome PASS requires execution outcome PASS, got %s", report.ExecutionOutcome)
		}
		if report.ResultCompleteness != "COMPLETE" {
			return fmt.Errorf("E2E_REPORT_INVALID: stage outcome PASS requires result completeness COMPLETE, got %s", report.ResultCompleteness)
		}
		if report.ScenariosFailed != 0 {
			return fmt.Errorf("E2E_REPORT_INVALID: stage outcome PASS requires scenarios_failed == 0, got %d", report.ScenariosFailed)
		}
		if report.ScenariosNotRun != 0 {
			return fmt.Errorf("E2E_REPORT_INVALID: stage outcome PASS requires scenarios_not_run == 0, got %d", report.ScenariosNotRun)
		}
		if report.ScenariosUnknown != 0 {
			return fmt.Errorf("E2E_REPORT_INVALID: stage outcome PASS requires scenarios_unknown == 0, got %d", report.ScenariosUnknown)
		}
		if report.ScenariosPassed <= 0 {
			return fmt.Errorf("E2E_REPORT_INVALID: stage outcome PASS requires scenarios_passed > 0, got %d", report.ScenariosPassed)
		}
		if report.EvidenceURI == "UNAVAILABLE" {
			return errors.New("E2E_REPORT_INVALID: stage outcome PASS cannot have evidence_uri UNAVAILABLE")
		}
		if report.TestedArtifactRef == "NONE" {
			return errors.New("E2E_REPORT_INVALID: stage outcome PASS cannot have tested_artifact_ref NONE")
		}
		if report.CleanupStatus != "CONFIRMED" {
			return fmt.Errorf("E2E_REPORT_INVALID: stage outcome PASS requires cleanup_status CONFIRMED, got %s", report.CleanupStatus)
		}

	case "BUILD_FAILURE":
		if report.BuildOutcome == "ENV_ERROR" {
			return errors.New("E2E_REPORT_INVALID: build outcome ENV_ERROR must map to stage outcome ENV_BLOCKED, not BUILD_FAILURE")
		}
		if report.BuildOutcome == "TIMEOUT" {
			return errors.New("E2E_REPORT_INVALID: build outcome TIMEOUT must map to stage outcome TIMEOUT, not BUILD_FAILURE")
		}
		if report.BuildOutcome != "FAIL" {
			return fmt.Errorf("E2E_REPORT_INVALID: stage outcome BUILD_FAILURE strictly requires build outcome FAIL, got %s", report.BuildOutcome)
		}
		if brief.Mode != "build-and-run" {
			return errors.New("E2E_REPORT_INVALID: stage outcome BUILD_FAILURE is only valid in build-and-run mode")
		}
		if report.ExecutionOutcome != "NOT_RUN" {
			return fmt.Errorf("E2E_REPORT_INVALID: stage outcome BUILD_FAILURE requires execution outcome NOT_RUN, got %s", report.ExecutionOutcome)
		}
		if report.TestedArtifactRef != "NONE" {
			return fmt.Errorf("E2E_REPORT_INVALID: stage outcome BUILD_FAILURE requires tested_artifact_ref NONE, got %s", report.TestedArtifactRef)
		}

	case "PRODUCT_FAILURE":
		if report.ExecutionOutcome != "FAIL" {
			return fmt.Errorf("E2E_REPORT_INVALID: stage outcome PRODUCT_FAILURE requires execution outcome FAIL, got %s", report.ExecutionOutcome)
		}
		if report.ScenariosFailed <= 0 {
			return fmt.Errorf("E2E_REPORT_INVALID: stage outcome PRODUCT_FAILURE requires scenarios_failed > 0, got %d", report.ScenariosFailed)
		}
		if brief.Mode == "build-and-run" && report.BuildOutcome != "PASS" {
			return fmt.Errorf("E2E_REPORT_INVALID: stage outcome PRODUCT_FAILURE in build-and-run mode requires build outcome PASS, got %s", report.BuildOutcome)
		}
		if brief.Mode == "run-only" && report.BuildOutcome != "NOT_RUN" {
			return fmt.Errorf("E2E_REPORT_INVALID: stage outcome PRODUCT_FAILURE in run-only mode requires build outcome NOT_RUN, got %s", report.BuildOutcome)
		}

	case "ENV_BLOCKED":
		if report.BuildOutcome != "ENV_ERROR" && report.ExecutionOutcome != "ENV_ERROR" {
			return errors.New("E2E_REPORT_INVALID: stage outcome ENV_BLOCKED requires build outcome or execution outcome to be ENV_ERROR")
		}
		if report.BuildOutcome == "PASS" && report.ExecutionOutcome == "PASS" {
			return errors.New("E2E_REPORT_INVALID: stage outcome ENV_BLOCKED cannot have both build outcome PASS and execution outcome PASS")
		}

	case "CLARIFICATION_REQUIRED":
		if report.BuildOutcome != "ENV_ERROR" && report.BuildOutcome != "NOT_RUN" {
			return fmt.Errorf("E2E_REPORT_INVALID: stage outcome CLARIFICATION_REQUIRED requires build outcome ENV_ERROR or NOT_RUN, got %s", report.BuildOutcome)
		}
		if report.ExecutionOutcome != "NOT_RUN" {
			return fmt.Errorf("E2E_REPORT_INVALID: stage outcome CLARIFICATION_REQUIRED requires execution outcome NOT_RUN, got %s", report.ExecutionOutcome)
		}
		if report.ResultCompleteness != "NONE" {
			return fmt.Errorf("E2E_REPORT_INVALID: stage outcome CLARIFICATION_REQUIRED requires result completeness NONE, got %s", report.ResultCompleteness)
		}

	case "TIMEOUT":
		if report.BuildOutcome != "TIMEOUT" && report.ExecutionOutcome != "TIMEOUT" {
			return errors.New("E2E_REPORT_INVALID: stage outcome TIMEOUT requires build outcome or execution outcome to be TIMEOUT")
		}
		if report.ExecutionOutcome == "PASS" {
			return errors.New("E2E_REPORT_INVALID: stage outcome TIMEOUT cannot have execution outcome PASS")
		}

	case "CANCELLED":
		if report.BuildOutcome != "CANCELLED" && report.ExecutionOutcome != "CANCELLED" {
			return errors.New("E2E_REPORT_INVALID: stage outcome CANCELLED requires build outcome or execution outcome to be CANCELLED")
		}
		if report.ExecutionOutcome == "PASS" {
			return errors.New("E2E_REPORT_INVALID: stage outcome CANCELLED cannot have execution outcome PASS")
		}

	case "RUNNER_LOST":

	case "INVALID_EVIDENCE":
		if report.ExecutionOutcome != "INVALID_EVIDENCE" {
			return fmt.Errorf("E2E_REPORT_INVALID: stage outcome INVALID_EVIDENCE requires execution outcome INVALID_EVIDENCE, got %s", report.ExecutionOutcome)
		}
	}

	return nil
}

// EvaluateE2EAdmission strictly evaluates whether an E2E run qualifies for admission to final code review.
func EvaluateE2EAdmission(brief E2EBriefContext, report E2EReportPayload, material E2EExecutionMaterialContext, state E2EReportValidationContext) error {
	// 1. Report must be structurally and contextually valid
	if brief.CredentialPolicy == nil {
		return errors.New("E2E_ADMISSION_DENIED: credential_policy is required (use explicit empty list [] if none)")
	}
	if brief.Services == nil {
		return errors.New("E2E_ADMISSION_DENIED: services is required (use explicit empty list [] if none)")
	}
	if state.PreflightOutcome == "AUTHORIZATION_DENIED" {
		return errors.New("E2E_ADMISSION_DENIED: preflight AUTHORIZATION_DENIED cannot be admitted")
	}
	if err := ValidateE2EReport(brief, report, state); err != nil {
		return fmt.Errorf("E2E_ADMISSION_DENIED: report validation failed: %w", err)
	}

	// 2. Admission requires PASS
	if report.StageOutcome != "PASS" {
		return fmt.Errorf("E2E_ADMISSION_DENIED: stage outcome is %s (PASS required)", report.StageOutcome)
	}
	if report.ResultCompleteness != "COMPLETE" {
		return fmt.Errorf("E2E_ADMISSION_DENIED: result completeness is %s (COMPLETE required)", report.ResultCompleteness)
	}
	if report.ScenariosFailed > 0 || report.ScenariosNotRun > 0 || report.ScenariosUnknown > 0 {
		return fmt.Errorf("E2E_ADMISSION_DENIED: failed scenarios (%d), not-run scenarios (%d), or unknown scenarios (%d) exist", report.ScenariosFailed, report.ScenariosNotRun, report.ScenariosUnknown)
	}
	if report.CleanupStatus != "CONFIRMED" || state.CleanupStatus != "CONFIRMED" {
		return errors.New("E2E_ADMISSION_DENIED: cleanup status is not CONFIRMED")
	}
	if len(state.UnresolvedResourceIDs) > 0 {
		return errors.New("E2E_ADMISSION_DENIED: unresolved cleanup resources exist")
	}

	// Admission requires successful credential provisioning for non-empty policy
	if len(brief.CredentialPolicy) > 0 {
		if state.CredentialProvisioningState != "SUCCEEDED" {
			return fmt.Errorf("E2E_ADMISSION_DENIED: credential provisioning state is %s (SUCCEEDED required for non-empty credential policy)", state.CredentialProvisioningState)
		}
	}

	// 3. Execution material checks
	if material.RepoRef != brief.RepoRef {
		return fmt.Errorf("E2E_ADMISSION_DENIED: repo_ref mismatch in execution material: expected %q, got %q", brief.RepoRef, material.RepoRef)
	}
	if material.InputProtectionMode != "READ_ONLY_MATERIALIZATION" && material.InputProtectionMode != "EXCLUSIVE_USE_LEASE" {
		return fmt.Errorf("E2E_ADMISSION_DENIED: invalid input protection mode %q", material.InputProtectionMode)
	}
	if !material.ProtectionProofVerified {
		return errors.New("E2E_ADMISSION_DENIED: protection proof not verified")
	}
	if !material.ProtectedIntervalComplete {
		return errors.New("E2E_ADMISSION_DENIED: protected interval not complete")
	}
	if strings.TrimSpace(material.ProtectionProofRef) == "" {
		return errors.New("E2E_ADMISSION_DENIED: empty protection proof ref")
	}
	if !sha256PrefixPattern.MatchString(material.CandidateTreeRef) {
		return fmt.Errorf("E2E_ADMISSION_DENIED: invalid candidate_tree_ref %q", material.CandidateTreeRef)
	}
	if material.SourceTreeBeforeRef != material.CandidateTreeRef {
		return fmt.Errorf("E2E_ADMISSION_DENIED: source tree before ref %q does not match candidate tree ref %q", material.SourceTreeBeforeRef, material.CandidateTreeRef)
	}
	if material.SourceTreeAfterRef != material.CandidateTreeRef {
		return fmt.Errorf("E2E_ADMISSION_DENIED: source tree after ref %q does not match candidate tree ref %q", material.SourceTreeAfterRef, material.CandidateTreeRef)
	}
	if !material.CheckoutCleanBefore || !material.CheckoutCleanAfter {
		return errors.New("E2E_ADMISSION_DENIED: checkout was not clean before and after execution")
	}
	if material.TestPackagePath != brief.TestPackage {
		return fmt.Errorf("E2E_ADMISSION_DENIED: test package path %q does not match brief %q", material.TestPackagePath, brief.TestPackage)
	}
	if !sha256PrefixPattern.MatchString(material.TestPackageTreeRef) {
		return fmt.Errorf("E2E_ADMISSION_DENIED: invalid test_package_tree_ref %q", material.TestPackageTreeRef)
	}
	if !material.FreshTestExecution {
		return errors.New("E2E_ADMISSION_DENIED: fresh test execution is required; cached test verdicts prohibited")
	}

	// Bind execution-material services and versions exactly to approved brief entries
	if len(material.ServiceRefs) != len(brief.Services) {
		return fmt.Errorf("E2E_ADMISSION_DENIED: service_refs count %d does not match brief services count %d", len(material.ServiceRefs), len(brief.Services))
	}
	for i, svc := range brief.Services {
		expectedRef1 := fmt.Sprintf("%s=%s", svc.ID, svc.ImageOrVersion)
		expectedRef2 := fmt.Sprintf("%s:%s", svc.ID, svc.ImageOrVersion)
		actualRef := material.ServiceRefs[i]
		if actualRef != expectedRef1 && actualRef != expectedRef2 {
			return fmt.Errorf("E2E_ADMISSION_DENIED: service ref at index %d (%q) does not match brief service %s with image/version %s", i, actualRef, svc.ID, svc.ImageOrVersion)
		}
	}

	// Validate genuine approved Go test invocation with distinct -count=1
	if err := validateGoTestCommand(material.Command); err != nil {
		return fmt.Errorf("E2E_ADMISSION_DENIED: %w", err)
	}

	if !slicesEqual(material.Command, brief.Command) {
		return errors.New("E2E_ADMISSION_DENIED: command mismatch between material and brief")
	}
	if !slicesEqual(material.BuildCommand, brief.BuildCommand) {
		return errors.New("E2E_ADMISSION_DENIED: build_command mismatch between material and brief")
	}
	if material.GoVersion != brief.GoVersion {
		return fmt.Errorf("E2E_ADMISSION_DENIED: go_version mismatch: expected %q, got %q", brief.GoVersion, material.GoVersion)
	}
	if !slicesEqual(material.Platforms, brief.Platforms) {
		return errors.New("E2E_ADMISSION_DENIED: platforms mismatch between material and brief")
	}
	if material.DependencyLockRef != brief.DependencyLockRef {
		return fmt.Errorf("E2E_ADMISSION_DENIED: dependency_lock_ref mismatch: expected %q, got %q", brief.DependencyLockRef, material.DependencyLockRef)
	}

	// 4. Tested Artifact provenance checks
	buildMaterialDigest := ComputeBuildMaterialDigest(material.RepoRef, brief.CandidateRef, material.CandidateTreeRef, material.BuildCommand, material.GoVersion, material.Platforms, material.DependencyLockRef)
	if material.ArtifactCandidateRef != brief.CandidateRef {
		return fmt.Errorf("E2E_ADMISSION_DENIED: artifact candidate_ref %q does not match brief %q", material.ArtifactCandidateRef, brief.CandidateRef)
	}
	if material.ArtifactBuildMaterialRef != "sha256:"+buildMaterialDigest {
		return fmt.Errorf("E2E_ADMISSION_DENIED: artifact build material ref %q does not match expected sha256:%s", material.ArtifactBuildMaterialRef, buildMaterialDigest)
	}
	if material.ArtifactSource != "FRESH" && material.ArtifactSource != "VERIFIED_REUSE" {
		return fmt.Errorf("E2E_ADMISSION_DENIED: invalid artifact source %q", material.ArtifactSource)
	}
	expectedArtifactRef := FormatTestedArtifactRef(material.ArtifactSHA256, brief.CandidateRef, buildMaterialDigest, material.ArtifactSource)
	if report.TestedArtifactRef != expectedArtifactRef {
		return fmt.Errorf("E2E_ADMISSION_DENIED: tested_artifact_ref mismatch: expected %q, got %q", expectedArtifactRef, report.TestedArtifactRef)
	}

	// 5. ExecutionMaterialRef equality
	expectedExecutionMaterialRef := ComputeExecutionMaterialRef(brief, material)
	if report.ExecutionMaterialRef != expectedExecutionMaterialRef {
		return fmt.Errorf("E2E_ADMISSION_DENIED: execution_material_ref mismatch: expected %q, got %q", expectedExecutionMaterialRef, report.ExecutionMaterialRef)
	}

	return nil
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
