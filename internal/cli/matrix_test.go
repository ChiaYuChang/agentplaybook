package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/ChiaYuChang/agentplaybook/internal/cli"
	"github.com/ChiaYuChang/agentplaybook/internal/knowledge"
)

func TestCLI_GoldenDiscoveryMatrix(t *testing.T) {
	t.Parallel()

	discoveryCommands := [][]string{
		{},
		{"role"},
		{"flow"},
		{"artifact"},
		{"rule"},
	}

	for _, cmd := range discoveryCommands {
		cmdName := strings.Join(cmd, " ")
		if cmdName == "" {
			cmdName = "<root>"
		}
		t.Run("discovery_"+cmdName, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := cli.Execute(cmd, &stdout, &stderr, "v0.1.0")
			if err != nil {
				t.Fatalf("expected discovery command %q to succeed with exit 0, got err: %v", cmdName, err)
			}
			if stdout.Len() == 0 {
				t.Errorf("expected non-empty output for discovery command %q", cmdName)
			}
			if stderr.Len() != 0 {
				t.Errorf("expected empty stderr for discovery command %q, got: %s", cmdName, stderr.String())
			}
		})
	}
}

func TestCLI_GoldenJSONMatrix(t *testing.T) {
	t.Parallel()

	jsonCommands := [][]string{
		{"--language"},
		{"role", "planner"},
		{"role", "builder"},
		{"role", "reviewer"},
		{"role", "scout"},
		{"role", "builder", "--responsibility"},
		{"role", "builder", "--boundary"},
		{"role", "builder", "--communication"},
		{"role", "scout", "--responsibility"},
		{"role", "scout", "--boundary"},
		{"role", "scout", "--communication"},
		{"role", "planner", "--communication"},
		{"flow", "init"},
		{"flow", "init", "--step", "1"},
		{"flow", "init", "--step", "2"},
		{"flow", "plan"},
		{"flow", "build"},
		{"flow", "review"},
		{"flow", "commit"},
		{"artifact", "agents-md"},
		{"artifact", "build-plan"},
		{"artifact", "review-plan"},
		{"artifact", "review-findings"},
		{"artifact", "scout-survey"},
		{"rule", "list"},
		{"rule", "explain", "anti-cheating"},
		{"rule", "explain", "atomic-change-units"},
		{"rule", "explain", "mandatory-alignment"},
		{"rule", "explain", "tdd-reproduction"},
		{"rule", "explain", "repo-context-storage"},
		{"rule", "explain", "agents-md-single-writer"},
		{"rule", "explain", "commit-authority-separation"},
		{"rule", "explain", "ephemeral-communication-buffers"},
		{"rule", "explain", "event-driven-transport-coordination"},
		{"rule", "explain", "interface-stability-contract-testing"},
		{"rule", "explain", "scout-recon-read-only"},
	}

	for _, cmd := range jsonCommands {
		cmdName := strings.Join(cmd, " ")
		t.Run("json_"+cmdName, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := cli.Execute(cmd, &stdout, &stderr, "v0.1.0")
			if err != nil {
				t.Fatalf("expected command %q to succeed, got err: %v", cmdName, err)
			}
			if !json.Valid(stdout.Bytes()) {
				t.Fatalf("expected valid JSON output for command %q, got:\n%s", cmdName, stdout.String())
			}
		})
	}
}

func TestCLI_VCSCommitGovernance(t *testing.T) {
	t.Parallel()

	queryJSON := func(args ...string) []byte {
		t.Helper()
		var stdout, stderr bytes.Buffer
		if err := cli.Execute(args, &stdout, &stderr, "v0.1.0"); err != nil {
			t.Fatalf("query %q failed: %v", strings.Join(args, " "), err)
		}
		return stdout.Bytes()
	}

	var builder knowledge.RoleDefinition
	if err := json.Unmarshal(queryJSON("role", "builder"), &builder); err != nil {
		t.Fatalf("failed to decode builder role: %v", err)
	}
	if !slices.Contains(builder.Responsibilities, "Produce minimal, reviewable working copy diffs accompanied by reproduction and green unit tests.") {
		t.Error("expected builder responsibility to require verified working copy diffs and green tests")
	}
	if !slices.Contains(builder.Boundaries, "Do not execute VCS commit commands, modify commit history, or alter branch/revision pointers; hand off verified working copy diffs to Planner for VCS governance.") {
		t.Error("expected builder boundary to prohibit direct VCS history changes")
	}

	var planner knowledge.RoleDefinition
	if err := json.Unmarshal(queryJSON("role", "planner"), &planner); err != nil {
		t.Fatalf("failed to decode planner role: %v", err)
	}
	if !slices.Contains(planner.Responsibilities, "Exclusively own VCS history and version control governance, creating atomic Conventional Commits from verified Builder diffs and managing revision progression.") {
		t.Error("expected planner responsibility to own VCS history")
	}

	var build knowledge.Flow
	if err := json.Unmarshal(queryJSON("flow", "build"), &build); err != nil {
		t.Fatalf("failed to decode build flow: %v", err)
	}
	step6Index := slices.IndexFunc(build.Steps, func(step knowledge.FlowStep) bool {
		return step.Index == 6
	})
	if step6Index < 0 {
		t.Fatal("expected build flow to define Step 6")
	}
	step6 := build.Steps[step6Index]
	if step6.Actor != knowledge.RolePlanner {
		t.Errorf("expected build flow Step 6 actor planner, got %q", step6.Actor)
	}
	if step6.Action != "Inspect working copy diff against in-scope boundaries, verify local test execution, and hand off to Reviewer without committing." {
		t.Errorf("expected build flow Step 6 to hand off to Reviewer without committing, got %q", step6.Action)
	}

	skill, err := os.ReadFile("../../SKILL.md")
	if err != nil {
		t.Fatalf("failed to read SKILL.md: %v", err)
	}
	for _, expected := range []string{
		"Version Control Governance is exclusively owned by Planner",
		"Builder delivers verified working copy diffs",
		"Governed Commit Flow",
		"delegating to the active VCS/policy mechanism capabilities",
	} {
		if !strings.Contains(string(skill), expected) {
			t.Errorf("expected SKILL.md to document %q", expected)
		}
	}
}

func TestCLI_PlanGateAndTDDBaseline(t *testing.T) {
	t.Parallel()

	queryJSON := func(args ...string) []byte {
		t.Helper()
		var stdout, stderr bytes.Buffer
		if err := cli.Execute(args, &stdout, &stderr, "v0.1.0"); err != nil {
			t.Fatalf("query %q failed: %v", strings.Join(args, " "), err)
		}
		return stdout.Bytes()
	}

	var planner knowledge.RoleDefinition
	if err := json.Unmarshal(queryJSON("role", "planner"), &planner); err != nil {
		t.Fatalf("failed to decode planner role: %v", err)
	}
	if !slices.Contains(planner.Boundaries, "DO NOT self-approve task plans or bypass the Plan-Review Gate; task plans must be independently evaluated and approved by Reviewer before dispatching to Builder. Only an explicit user instruction may override this gate; Planner cannot waive it autonomously.") {
		t.Error("expected planner boundary to prohibit bypassing the Plan-Review Gate")
	}

	var reviewer knowledge.RoleDefinition
	if err := json.Unmarshal(queryJSON("role", "reviewer"), &reviewer); err != nil {
		t.Fatalf("failed to decode reviewer role: %v", err)
	}
	if !slices.Contains(reviewer.Responsibilities, "Exclusively hold the gate approval authority for task build-plans and review-plans during the plan flow.") {
		t.Error("expected reviewer to hold plan gate approval authority")
	}

	var builder knowledge.RoleDefinition
	if err := json.Unmarshal(queryJSON("role", "builder"), &builder); err != nil {
		t.Fatalf("failed to decode builder role: %v", err)
	}
	if !slices.Contains(builder.Boundaries, "DO NOT modify application code to address review findings without first establishing an expectedly failing TDD reproduction test.") {
		t.Error("expected builder boundary to require a failing TDD reproduction test")
	}

	var rules []knowledge.Rule
	if err := json.Unmarshal(queryJSON("rule", "explain", "tdd-reproduction"), &rules); err != nil {
		t.Fatalf("failed to decode tdd-reproduction rule: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected one tdd-reproduction rule, got %d", len(rules))
	}
	for _, expected := range []string{
		"Builder must establish a Red (failing) reproduction test before any application-code changes.",
		"Planner verifies the reproduction failure is an assertion failure expressing the reported defect, not a compile, fixture, or timeout failure.",
		"Planner records a baseline reference (revision ID or snapshot) and the complete frozen test dependency surface (test files, helpers, fixtures, mocks, golden files, generated inputs, and pre-existing dependencies the reproduction test depends on).",
		"Builder must not alter any frozen path during the fix phase.",
		"Planner compares all frozen paths to the baseline reference before accepting the fix handoff.",
		"If the audit detects any diff in frozen paths, Planner stops acceptance and resolves or escalates; the original baseline reference is preserved as evidence.",
		"Reviewer independently validates the final result once all tests are green.",
	} {
		if !slices.Contains(rules[0].Guidelines, expected) {
			t.Errorf("expected tdd-reproduction guideline %q", expected)
		}
	}
}

func TestCLI_LivingAgentsAndCommitFlow(t *testing.T) {
	t.Parallel()

	queryJSON := func(args ...string) []byte {
		t.Helper()
		var stdout, stderr bytes.Buffer
		if err := cli.Execute(args, &stdout, &stderr, "v0.1.0"); err != nil {
			t.Fatalf("query %q failed: %v", strings.Join(args, " "), err)
		}
		return stdout.Bytes()
	}

	// 1. Verify agents-md artifact
	var agentsMd knowledge.Artifact
	if err := json.Unmarshal(queryJSON("artifact", "agents-md"), &agentsMd); err != nil {
		t.Fatalf("failed to decode agents-md artifact: %v", err)
	}
	if agentsMd.Name != "agents-md" || agentsMd.Path != "AGENTS.md" || agentsMd.Owner != knowledge.RolePlanner {
		t.Errorf("unexpected agents-md metadata: %+v", agentsMd)
	}
	if len(agentsMd.Sections) != 5 {
		t.Fatalf("expected 5 sections in agents-md, got %d", len(agentsMd.Sections))
	}
	expectedSections := []string{
		"Architectural Topology & Jurisdictions",
		"Global Operational Invariants",
		"Builder Precautions & Gotchas",
		"Reviewer Precautions & Checklist",
		"Active State & In-Flight Context",
	}
	for i, expectedSec := range expectedSections {
		if agentsMd.Sections[i].Name != expectedSec {
			t.Errorf("expected section %d to be %q, got %q", i, expectedSec, agentsMd.Sections[i].Name)
		}
	}

	// 2. Verify commit flow
	var commit knowledge.Flow
	if err := json.Unmarshal(queryJSON("flow", "commit"), &commit); err != nil {
		t.Fatalf("failed to decode commit flow: %v", err)
	}
	if len(commit.Steps) != 9 {
		t.Fatalf("expected 9 steps in commit flow, got %d", len(commit.Steps))
	}

	// Verify all actors and sequential / conditional wiring
	for _, step := range commit.Steps {
		expectedActor := knowledge.RolePlanner
		if step.Index == 5 {
			expectedActor = knowledge.RoleReviewer
		}
		if step.Actor != expectedActor {
			t.Errorf("step %d expected actor %q, got %q", step.Index, expectedActor, step.Actor)
		}
		// VCS-neutral invariant: Flow step actions must NOT contain raw jj or git command syntax
		for _, forbidden := range []string{"jj ", "git ", "git commit", "jj describe", "jj new"} {
			if strings.Contains(step.Action, forbidden) {
				t.Errorf("step %d action violates VCS-neutral invariant by containing raw command %q: %s", step.Index, forbidden, step.Action)
			}
		}
	}

	// Check Step 5 visibility rejection loop
	step5 := commit.Steps[4]
	if step5.Actor != knowledge.RoleReviewer {
		t.Errorf("expected Step 5 actor to be reviewer, got %s", step5.Actor)
	}
	step5Conditions := make(map[string]int)
	for _, c := range step5.Conditions {
		step5Conditions[c.When] = c.Then
	}
	if step5Conditions["AGENTS_REVIEW_PASS"] != 6 {
		t.Errorf("expected Step 5 AGENTS_REVIEW_PASS -> 6, got %d", step5Conditions["AGENTS_REVIEW_PASS"])
	}
	if step5Conditions["BARRIER_LEAK"] != 4 {
		t.Errorf("expected Step 5 BARRIER_LEAK -> 4, got %d", step5Conditions["BARRIER_LEAK"])
	}

	// Check Step 6 secret scan failure loop
	step6 := commit.Steps[5]
	step6Conditions := make(map[string]int)
	for _, c := range step6.Conditions {
		step6Conditions[c.When] = c.Then
	}
	if step6Conditions["SCAN_CLEAN"] != 7 {
		t.Errorf("expected Step 6 SCAN_CLEAN -> 7, got %d", step6Conditions["SCAN_CLEAN"])
	}
	if step6Conditions["SCAN_FAILED"] != 4 {
		t.Errorf("expected Step 6 SCAN_FAILED -> 4, got %d", step6Conditions["SCAN_FAILED"])
	}

	// Check Step 7 authorization denial loop
	step7 := commit.Steps[6]
	step7Conditions := make(map[string]int)
	for _, c := range step7.Conditions {
		step7Conditions[c.When] = c.Then
	}
	if step7Conditions["AUTHORIZATION_GRANTED"] != 8 {
		t.Errorf("expected Step 7 AUTHORIZATION_GRANTED -> 8, got %d", step7Conditions["AUTHORIZATION_GRANTED"])
	}
	if step7Conditions["AUTHORIZATION_DENIED"] != 2 {
		t.Errorf("expected Step 7 AUTHORIZATION_DENIED -> 2, got %d", step7Conditions["AUTHORIZATION_DENIED"])
	}

	// Check Step 8 publication condition
	step8 := commit.Steps[7]
	if len(step8.Conditions) != 1 || step8.Conditions[0].When != "PUBLICATION_AUTHORIZED" || step8.Conditions[0].Then != 9 {
		t.Errorf("expected Step 8 to have condition PUBLICATION_AUTHORIZED -> 9, got %+v", step8.Conditions)
	}

	// Check Step 9 is terminal
	step9 := commit.Steps[8]
	if len(step9.Conditions) != 0 {
		t.Errorf("expected Step 9 to be terminal with 0 conditions, got %d", len(step9.Conditions))
	}

	// 3. Verify agents-md-single-writer rule
	var singleWriterRules []knowledge.Rule
	if err := json.Unmarshal(queryJSON("rule", "explain", "agents-md-single-writer"), &singleWriterRules); err != nil {
		t.Fatalf("failed to decode agents-md-single-writer rule: %v", err)
	}
	if len(singleWriterRules) != 1 {
		t.Fatalf("expected 1 agents-md-single-writer rule, got %d", len(singleWriterRules))
	}
	singleWriterGuidelines := singleWriterRules[0].Guidelines
	for _, expected := range []string{
		"Planner is the sole author and curator of AGENTS.md; Builder and Reviewer must never edit AGENTS.md directly.",
		"The Active State section must record baseline provenance with 'Observed-At: <UTC timestamp> @ <base-revision-id>', dirty status, recent milestones, and next pickup item.",
		"Receiving Planners cold-starting a session must execute fresh VCS status and log commands to revalidate mutable ground truth before planning or executing tasks.",
		"Reviewer must conduct narrow visibility and blind-barrier checks on AGENTS.md during the commit flow (BARRIER_LEAK returns to Planner for redaction).",
	} {
		if !slices.Contains(singleWriterGuidelines, expected) {
			t.Errorf("expected agents-md-single-writer guideline %q", expected)
		}
	}

	// 4. Verify commit-authority-separation rule
	var authorityRules []knowledge.Rule
	if err := json.Unmarshal(queryJSON("rule", "explain", "commit-authority-separation"), &authorityRules); err != nil {
		t.Fatalf("failed to decode commit-authority-separation rule: %v", err)
	}
	if len(authorityRules) != 1 {
		t.Fatalf("expected 1 commit-authority-separation rule, got %d", len(authorityRules))
	}
	authorityGuidelines := authorityRules[0].Guidelines
	for _, expected := range []string{
		"Commit authorization permits local revision sealing only; it does not grant authority to publish to remote repositories.",
		"Publishing to remote repositories requires separate and explicit human publication authorization.",
		"If commit authorization is denied (AUTHORIZATION_DENIED), Planner must return to Step 2 awaiting renewed user intent; autonomous re-drafting is strictly forbidden.",
	} {
		if !slices.Contains(authorityGuidelines, expected) {
			t.Errorf("expected commit-authority-separation guideline %q", expected)
		}
	}
}

func TestCLI_InterfaceStabilityContractTesting(t *testing.T) {
	t.Parallel()

	queryJSON := func(args ...string) []byte {
		t.Helper()
		var stdout, stderr bytes.Buffer
		if err := cli.Execute(args, &stdout, &stderr, "v0.1.0"); err != nil {
			t.Fatalf("query %q failed: %v", strings.Join(args, " "), err)
		}
		return stdout.Bytes()
	}

	var rules []knowledge.Rule
	if err := json.Unmarshal(queryJSON("rule", "explain", "interface-stability-contract-testing"), &rules); err != nil {
		t.Fatalf("failed to decode interface-stability-contract-testing rule: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected one interface-stability-contract-testing rule, got %d", len(rules))
	}
	expectedGuidelines := []string{
		"A build plan must identify all affected boundary symbols, endpoints, schemas, files, or consumer contracts, or explicitly state that no external boundary is affected.",
		"Interface changes require a plan amendment before implementation, identifying affected consumers and compatibility or migration handling.",
		"Contract tests must assert observable input/output, side effects, errors, or interoperability at the boundary, not internal implementation details or mere absence of failure.",
		"A contract test must fail under at least one plausible violating implementation; Reviewer assesses falsifiability through targeted variation where feasible.",
		"Unexpected cross-boundary dependencies require Planner escalation; Builder must not unilaterally expand scope.",
		"Contract tests are distinct from TDD reproduction tests: TDD reproduction is mandatory for validated review findings; contract tests are required when boundary behavior is added, changed, or insufficiently protected.",
	}
	if len(rules[0].Guidelines) != len(expectedGuidelines) {
		t.Fatalf("expected %d interface-stability-contract-testing guidelines, got %d", len(expectedGuidelines), len(rules[0].Guidelines))
	}
	for _, expected := range expectedGuidelines {
		if !slices.Contains(rules[0].Guidelines, expected) {
			t.Errorf("expected interface-stability-contract-testing guideline %q", expected)
		}
	}

	var planner knowledge.RoleDefinition
	if err := json.Unmarshal(queryJSON("role", "planner"), &planner); err != nil {
		t.Fatalf("failed to decode planner role: %v", err)
	}
	if !slices.Contains(planner.Responsibilities, "Declare all component boundary symbols, endpoints, schemas, or consumer contracts affected by a build plan, or explicitly state no external boundary is affected.") {
		t.Error("expected planner responsibility to declare affected component boundaries")
	}
	if !slices.Contains(planner.Boundaries, "DO NOT permit implementation to alter public component interfaces without prior plan amendment identifying affected consumers and compatibility or migration handling.") {
		t.Error("expected planner boundary to require plan amendment for public interface changes")
	}

	var builder knowledge.RoleDefinition
	if err := json.Unmarshal(queryJSON("role", "builder"), &builder); err != nil {
		t.Fatalf("failed to decode builder role: %v", err)
	}
	if !slices.Contains(builder.Responsibilities, "When a component boundary or externally observable behavior is added, changed, or lacks protection, supply contract tests that assert observable input/output, side effects, errors, or interoperability at the boundary—not internal implementation details or mere absence of failure.") {
		t.Error("expected builder responsibility to require meaningful contract tests")
	}
	for _, expected := range []string{
		"DO NOT supply vacuous or tautological contract tests in place of substantive behavioral assertions.",
		"Escalate to Planner rather than unilaterally widening scope when unexpected cross-boundary dependencies are encountered.",
	} {
		if !slices.Contains(builder.Boundaries, expected) {
			t.Errorf("expected builder boundary %q", expected)
		}
	}

	var reviewer knowledge.RoleDefinition
	if err := json.Unmarshal(queryJSON("role", "reviewer"), &reviewer); err != nil {
		t.Fatalf("failed to decode reviewer role: %v", err)
	}
	if !slices.Contains(reviewer.Responsibilities, "Independently audit component interface stability and verify that contract tests assert genuine behavioral invariants—failing under at least one plausible violating implementation—distinct from TDD bug-fix reproductions.") {
		t.Error("expected reviewer responsibility to audit falsifiable contract tests")
	}
}

func TestCLI_ScoutReconnaissance(t *testing.T) {
	t.Parallel()

	queryJSON := func(args ...string) []byte {
		t.Helper()
		var stdout, stderr bytes.Buffer
		if err := cli.Execute(args, &stdout, &stderr, "v0.1.0"); err != nil {
			t.Fatalf("query %q failed: %v", strings.Join(args, " "), err)
		}
		return stdout.Bytes()
	}

	var scout knowledge.RoleDefinition
	if err := json.Unmarshal(queryJSON("role", "scout"), &scout); err != nil {
		t.Fatalf("failed to decode scout role: %v", err)
	}
	if scout.Title != "Scout" {
		t.Errorf("expected scout title %q, got %q", "Scout", scout.Title)
	}
	for _, expected := range []string{
		"Conduct read-only exploration and topological mapping of repositories, identifying directory structure, entry points, build graphs, and component boundaries.",
		"Deliver structured reconnaissance survey artifacts to Planner without modifying repository files or persistent documentation.",
		"Extract factual code symbols, dependency relationships, and toolchain configurations to accelerate cold-start onboarding.",
	} {
		if !slices.Contains(scout.Responsibilities, expected) {
			t.Errorf("expected scout responsibility %q", expected)
		}
	}
	for _, expected := range []string{
		"DO NOT edit, create, or delete repository files; exploration is strictly read-only.",
		"DO NOT edit AGENTS.md directly; enforce the Single-Writer Principle by delivering survey findings to Planner.",
		"DO NOT read, search for, or request task-specific review plans, reviewer-only tests, or verification artifacts.",
		"DO NOT participate in build planning, code implementation, or VCS mutation.",
		"Report only factual, verifiable codebase topography supported by concrete evidence rather than speculative design prescriptions.",
	} {
		if !slices.Contains(scout.Boundaries, expected) {
			t.Errorf("expected scout boundary %q", expected)
		}
	}
	if len(scout.Communication.Targets) != 1 || scout.Communication.Targets[0] != knowledge.RolePlanner {
		t.Errorf("expected scout communication target planner, got %v", scout.Communication.Targets)
	}

	var planner knowledge.RoleDefinition
	if err := json.Unmarshal(queryJSON("role", "planner"), &planner); err != nil {
		t.Fatalf("failed to decode planner role: %v", err)
	}
	if !slices.Contains(planner.Communication.Targets, knowledge.RoleScout) {
		t.Errorf("expected planner communication targets to include scout, got %v", planner.Communication.Targets)
	}

	var survey knowledge.Artifact
	if err := json.Unmarshal(queryJSON("artifact", "scout-survey"), &survey); err != nil {
		t.Fatalf("failed to decode scout-survey artifact: %v", err)
	}
	if survey.Owner != knowledge.RoleScout || survey.Type != "message" {
		t.Errorf("unexpected scout-survey metadata: %+v", survey)
	}
	if len(survey.Visibility) != 2 || survey.Visibility[0] != knowledge.RolePlanner || survey.Visibility[1] != knowledge.RoleScout {
		t.Errorf("expected scout-survey visibility [planner scout], got %v", survey.Visibility)
	}
	expectedFields := []string{"id", "provenance", "repository_topology", "module_boundaries", "build_and_toolchains", "evidence", "uncertainties"}
	if len(survey.Fields) != len(expectedFields) {
		t.Fatalf("expected %d scout-survey fields, got %d", len(expectedFields), len(survey.Fields))
	}
	for i, expected := range expectedFields {
		if survey.Fields[i].Name != expected || !survey.Fields[i].Required {
			t.Errorf("expected required scout-survey field %q, got %+v", expected, survey.Fields[i])
		}
	}

	var init knowledge.Flow
	if err := json.Unmarshal(queryJSON("flow", "init"), &init); err != nil {
		t.Fatalf("failed to decode init flow: %v", err)
	}
	if len(init.Steps) != 9 {
		t.Fatalf("expected 9 init steps, got %d", len(init.Steps))
	}
	step1 := init.Steps[0]
	conditions := make(map[string]int)
	for _, condition := range step1.Conditions {
		conditions[condition.When] = condition.Then
	}
	if conditions["DIRECT_SURVEY"] != 3 || conditions["SCOUT_RECON_REQUIRED"] != 2 {
		t.Errorf("unexpected init step 1 conditions: %v", conditions)
	}
	if init.Steps[1].Actor != knowledge.RoleScout {
		t.Errorf("expected init step 2 actor scout, got %q", init.Steps[1].Actor)
	}
	if init.Steps[3].Actor != knowledge.RoleReviewer {
		t.Errorf("expected init step 4 actor reviewer, got %q", init.Steps[3].Actor)
	}
	step8 := init.Steps[7]
	consensusConditions := make(map[string]int)
	for _, condition := range step8.Conditions {
		consensusConditions[condition.When] = condition.Then
	}
	if consensusConditions["QUESTIONS_RAISED"] != 4 || consensusConditions["NO_QUESTIONS_RAISED"] != 9 {
		t.Errorf("unexpected init step 8 conditions: %v", consensusConditions)
	}

	var step1Query knowledge.FlowStep
	if err := json.Unmarshal(queryJSON("flow", "init", "--step", "1"), &step1Query); err != nil {
		t.Fatalf("failed to decode init step 1: %v", err)
	}
	step1QueryConditions := make(map[string]int)
	for _, condition := range step1Query.Conditions {
		step1QueryConditions[condition.When] = condition.Then
	}
	if step1Query.Index != 1 || step1QueryConditions["DIRECT_SURVEY"] != 3 {
		t.Errorf("unexpected init step 1 query: %+v", step1Query)
	}
	var step2Query knowledge.FlowStep
	if err := json.Unmarshal(queryJSON("flow", "init", "--step", "2"), &step2Query); err != nil {
		t.Fatalf("failed to decode init step 2: %v", err)
	}
	if step2Query.Index != 2 || step2Query.Actor != knowledge.RoleScout {
		t.Errorf("unexpected init step 2 query: %+v", step2Query)
	}

	var rules []knowledge.Rule
	if err := json.Unmarshal(queryJSON("rule", "explain", "scout-recon-read-only"), &rules); err != nil {
		t.Fatalf("failed to decode scout-recon-read-only rule: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected one scout-recon-read-only rule, got %d", len(rules))
	}
	expectedGuidelines := []string{
		"Scout performs strictly read-only reconnaissance; creating, editing, or deleting repository files is strictly prohibited.",
		"Scout delivers transient survey findings with provenance and evidence directly to Planner; Scout must never write or edit AGENTS.md directly, preserving the Single-Writer Principle.",
		"Scout must not read, search for, or request task-specific review plans, reviewer-only tests, or verification artifacts.",
		"Planner validates Scout findings and evidence against live repository ground truth before synthesizing them into AGENTS.md or task plans.",
		"Model tiering recommends Scout >= Reviewer >= Planner >= Builder as advisory capacity routing guidance, with model selection scaling by task domain, scope, uncertainty, and risk.",
	}
	if len(rules[0].Guidelines) != len(expectedGuidelines) {
		t.Fatalf("expected %d scout-recon-read-only guidelines, got %d", len(expectedGuidelines), len(rules[0].Guidelines))
	}
	for _, expected := range expectedGuidelines {
		if !slices.Contains(rules[0].Guidelines, expected) {
			t.Errorf("expected scout-recon-read-only guideline %q", expected)
		}
	}
}

func TestCLI_GoldenErrorMatrix(t *testing.T) {
	t.Parallel()

	errorCommands := [][]string{
		{"--language", "--transport"},
		{"role", "nonexistent"},
		{"role", "builder", "--responsibility", "--boundary"},
		{"role", "--responsibility"},
		{"flow", "nonexistent"},
		{"flow", "init", "--step", "999"},
		{"flow", "--step", "1"},
		{"artifact", "nonexistent"},
		{"rule", "unexpected"},
		{"rule", "explain"},
		{"rule", "explain", "nonexistent"},
	}

	for _, cmd := range errorCommands {
		cmdName := strings.Join(cmd, " ")
		t.Run("error_"+cmdName, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := cli.Execute(cmd, &stdout, &stderr, "v0.1.0")
			if err == nil {
				t.Fatalf("expected command %q to fail with error, got nil", cmdName)
			}
		})
	}
}
