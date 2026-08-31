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
		{"flow", "blueprint"},
		{"flow", "blueprint", "--step", "1"},
		{"flow", "blueprint", "--step", "2"},
		{"flow", "build"},
		{"flow", "review"},
		{"flow", "commit"},
		{"flow", "session-handoff"},
		{"flow", "session-handoff", "--step", "1"},
		{"flow", "session-handoff", "--step", "2"},
		{"artifact", "agents-md"},
		{"artifact", "build-plan"},
		{"artifact", "review-plan"},
		{"artifact", "blueprint-plan"},
		{"artifact", "sub-build-plan"},
		{"artifact", "sub-review-plan"},
		{"artifact", "sub-review-resolution"},
		{"artifact", "review-findings"},
		{"artifact", "scout-survey"},
		{"artifact", "review-resolution"},
		{"rule", "list"},
		{"rule", "explain", "anti-cheating"},
		{"rule", "explain", "coherent-plan-units"},
		{"rule", "explain", "anti-rubber-stamp-plan-gate"},
		{"rule", "explain", "evidence-proportional-persistence"},
		{"rule", "explain", "mandatory-alignment"},
		{"rule", "explain", "tdd-reproduction"},
		{"rule", "explain", "repo-context-storage"},
		{"rule", "explain", "agents-md-single-writer"},
		{"rule", "explain", "acceptance-publication-authority"},
		{"rule", "explain", "ephemeral-communication-buffers"},
		{"rule", "explain", "event-driven-transport-coordination"},
		{"rule", "explain", "interface-stability-contract-testing"},
		{"rule", "explain", "scout-recon-read-only"},
		{"rule", "explain", "session-handoff-audit"},
		{"rule", "explain", "planner-reviewability"},
		{"rule", "explain", "review-severity-semantics"},
		{"rule", "explain", "track-b-action-differential-verification"},
		{"rule", "explain", "out-of-tree-baseline-mirror"},
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
		"Track A verification applies to local behavioral changes and bug fixes, requiring a Red (failing) reproduction test before modifying application code.",
		"Planner verifies the reproduction failure is an assertion failure expressing the reported defect, not a compile, fixture, or timeout failure.",
		"Planner records a baseline reference (revision ID or snapshot) and the complete frozen test dependency surface (test files, helpers, fixtures, mocks, golden files, generated inputs, and pre-existing dependencies the reproduction test depends on).",
		"Builder must not alter any frozen path during the fix phase.",
		"Planner compares all frozen paths to the baseline reference before accepting the fix handoff.",
		"If the audit detects any diff in frozen paths, Planner stops acceptance and resolves or escalates; the original baseline reference is preserved as evidence.",
		"Non-behavioral findings (documentation, lint, static policy, design) use static/specification evidence rather than artificial failing tests.",
		"Reviewer independently validates final results using Track A green tests or verified static evidence.",
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
		"AGENTS.md must be authored in concise US English (en-US); any non-ASCII domain term must carry an explicit adjacent inline rationale.",
		"Reviewer audits AGENTS.md during the commit flow for secret leaks, unauthorized non-ASCII text lacking inline justification, and excessive verbosity or bloat.",
	} {
		if !slices.Contains(singleWriterGuidelines, expected) {
			t.Errorf("expected agents-md-single-writer guideline %q", expected)
		}
	}

	// 4. Verify acceptance-publication-authority rule
	var authorityRules []knowledge.Rule
	if err := json.Unmarshal(queryJSON("rule", "explain", "acceptance-publication-authority"), &authorityRules); err != nil {
		t.Fatalf("failed to decode acceptance-publication-authority rule: %v", err)
	}
	if len(authorityRules) != 1 {
		t.Fatalf("expected 1 acceptance-publication-authority rule, got %d", len(authorityRules))
	}
	authorityGuidelines := authorityRules[0].Guidelines
	for _, expected := range []string{
		"The revision lifecycle strictly distinguishes WORKING, ACCEPTED, and PUBLISHED states.",
		"Milestone acceptance seals candidate revisions locally and enforces Finalization Equivalence (tree(Final) == tree(Verified)).",
		"Publishing to remote repositories requires explicit, separate human publication authorization.",
		"If authorization is denied (AUTHORIZATION_DENIED), Planner returns to Step 2 awaiting renewed user intent; autonomous re-drafting is strictly forbidden.",
	} {
		if !slices.Contains(authorityGuidelines, expected) {
			t.Errorf("expected acceptance-publication-authority guideline %q", expected)
		}
	}
}

func TestCLI_AgentsLanguageConciseness(t *testing.T) {
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
	if !slices.Contains(planner.Responsibilities, "Author and maintain AGENTS.md in telegraphic, token-dense style (omitting filler words, conversational prose, and redundant grammar), maximizing token efficiency for LLM parsing while keeping exact technical terms and code symbols in concise US English (en-US); any non-ASCII domain term must carry an explicit adjacent inline rationale.") {
		t.Error("expected planner responsibility to require telegraphic, token-dense AGENTS.md content with inline rationale")
	}
	if !slices.Contains(planner.Boundaries, "DO NOT permit AGENTS.md to accumulate conversational filler, grammatical fluff, redundant explanations, or unverified operational narratives; enforce telegraphic density.") {
		t.Error("expected planner boundary to enforce telegraphic density")
	}

	var reviewer knowledge.RoleDefinition
	if err := json.Unmarshal(queryJSON("role", "reviewer"), &reviewer); err != nil {
		t.Fatalf("failed to decode reviewer role: %v", err)
	}
	if !slices.Contains(reviewer.Responsibilities, "Provide public, independently observable operational caveats on commit, and conduct narrow visibility, language consistency (verifying en-US ASCII purity unless inline domain rationale is provided), telegraphic conciseness (flagging conversational fluff or human-facing essays), and blind-barrier checks on AGENTS.md.") {
		t.Error("expected reviewer responsibility to audit AGENTS.md telegraphic conciseness")
	}

	telegraphicMessageContract := "Communicate using telegraphic Caveman mode for all inter-agent messages (including [Planner], [Reviewer], [Builder], [Scout], or [<role>] prefixed exchanges): omit conversational filler, pleasantries, and polite framing in favor of compact, structured technical fragments."
	ephemeralCavemanMandate := "Inter-agent communication requires telegraphic (caveman) compression for all exchanges (including [Planner], [Reviewer], [Builder], [Scout], or [<role>] prefixed messages): drop conversational filler, pleasantries, and polite framing; communicate in compact, structured technical fragments to minimize transport token consumption."
	for _, roleName := range []string{"planner", "builder", "reviewer", "scout"} {
		var role knowledge.RoleDefinition
		if err := json.Unmarshal(queryJSON("role", roleName), &role); err != nil {
			t.Fatalf("failed to decode %s role: %v", roleName, err)
		}
		if !slices.Contains(role.Responsibilities, telegraphicMessageContract) {
			t.Errorf("expected %s responsibility to use the unified telegraphic Caveman contract", roleName)
		}
	}

	var rules []knowledge.Rule
	if err := json.Unmarshal(queryJSON("rule", "explain", "agents-md-single-writer"), &rules); err != nil {
		t.Fatalf("failed to decode agents-md-single-writer rule: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected one agents-md-single-writer rule, got %d", len(rules))
	}
	for _, expected := range []string{
		"Planner is the sole author and curator of AGENTS.md; Builder and Reviewer must never edit AGENTS.md directly.",
		"AGENTS.md contains 5 canonical sections: Architectural Topology & Jurisdictions, Global Operational Invariants, Builder Precautions & Gotchas, Reviewer Precautions & Checklist, and Active State & In-Flight Context.",
		"The Active State section must record baseline provenance with 'Observed-At: <UTC timestamp> @ <base-revision-id>', dirty status, recent milestones, and next pickup item.",
		"Receiving Planners cold-starting a session must execute fresh VCS status and log commands to revalidate mutable ground truth before planning or executing tasks.",
		"Reviewer must conduct narrow visibility and blind-barrier checks on AGENTS.md during the commit flow (BARRIER_LEAK returns to Planner for redaction).",
		"AGENTS.md must be authored in concise US English (en-US); any non-ASCII domain term must carry an explicit adjacent inline rationale.",
		"Reviewer audits AGENTS.md during the commit flow for secret leaks, unauthorized non-ASCII text lacking inline justification, and excessive verbosity or bloat.",
		"AGENTS.md is machine-facing operational memory written in telegraphic, token-dense style: drop articles, conversational filler, and narrative pleasantries; fragments are expected; technical terms and code symbols must remain exact.",
	} {
		if !slices.Contains(rules[0].Guidelines, expected) {
			t.Errorf("expected agents-md-single-writer guideline %q", expected)
		}
	}

	var communicationRules []knowledge.Rule
	if err := json.Unmarshal(queryJSON("rule", "explain", "ephemeral-communication-buffers"), &communicationRules); err != nil {
		t.Fatalf("failed to decode ephemeral-communication-buffers rule: %v", err)
	}
	if len(communicationRules) != 1 {
		t.Fatalf("expected one ephemeral-communication-buffers rule, got %d", len(communicationRules))
	}
	if !slices.Contains(communicationRules[0].Guidelines, ephemeralCavemanMandate) {
		t.Error("expected ephemeral communication rule to require one unified Caveman mandate")
	}
}

func TestCLI_RoleContextLifecycle(t *testing.T) {
	t.Parallel()

	queryJSON := func(args ...string) []byte {
		t.Helper()
		var stdout, stderr bytes.Buffer
		if err := cli.Execute(args, &stdout, &stderr, "v0.1.0"); err != nil {
			t.Fatalf("query %q failed: %v", strings.Join(args, " "), err)
		}
		return stdout.Bytes()
	}

	var reviewer knowledge.RoleDefinition
	if err := json.Unmarshal(queryJSON("role", "reviewer"), &reviewer); err != nil {
		t.Fatalf("failed to decode reviewer role: %v", err)
	}
	if !slices.Contains(reviewer.Responsibilities, "Self-evaluate context window utilization after completing a commit flow, executing compaction or session cleanup when context usage exceeds 50% or when accumulated review history bloats high-tier model operating costs.") {
		t.Error("expected reviewer responsibility to govern high-tier context compaction")
	}

	var builder knowledge.RoleDefinition
	if err := json.Unmarshal(queryJSON("role", "builder"), &builder); err != nil {
		t.Fatalf("failed to decode builder role: %v", err)
	}
	if !slices.Contains(builder.Responsibilities, "Operate as a stateless, easily replaceable worker; when context window is bloated or model quotas are reached, discard the session and allow Planner to spawn a fresh instance from the approved plan rather than spending tokens on compaction.") {
		t.Error("expected builder responsibility to govern stateless replaceability")
	}

	var rules []knowledge.Rule
	if err := json.Unmarshal(queryJSON("rule", "explain", "role-context-lifecycle"), &rules); err != nil {
		t.Fatalf("failed to decode role-context-lifecycle rule: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected one role-context-lifecycle rule, got %d", len(rules))
	}
	if rules[0].ID != "role-context-lifecycle" {
		t.Fatalf("expected role-context-lifecycle rule, got %q", rules[0].ID)
	}
	if !strings.Contains(rules[0].Details, "approved <slug>.plan.md") {
		t.Error("expected role-context-lifecycle details to reference the approved plan artifact convention")
	}
	for _, expected := range []string{
		"Reviewer must self-evaluate context window utilization after completing a commit flow, executing compaction when usage exceeds 50% or when accumulated review history bloats high-tier model operating costs.",
		"Builder operates as a stateless, disposable worker; bloated Builder sessions are replaced with fresh instances rather than spending tokens on compaction.",
		"Planner retains orchestration history and manages living memory transitions across ephemeral agent lifecycles.",
	} {
		if !slices.Contains(rules[0].Guidelines, expected) {
			t.Errorf("expected role-context-lifecycle guideline %q", expected)
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

func TestCLI_SessionHandoffAuditAndStartupChecklist(t *testing.T) {
	t.Parallel()

	queryJSON := func(args ...string) []byte {
		t.Helper()
		var stdout, stderr bytes.Buffer
		if err := cli.Execute(args, &stdout, &stderr, "v0.1.0"); err != nil {
			t.Fatalf("query %q failed: %v", strings.Join(args, " "), err)
		}
		return stdout.Bytes()
	}

	// 1. Verify Planner role startup responsibilities & boundary
	var planner knowledge.RoleDefinition
	if err := json.Unmarshal(queryJSON("role", "planner"), &planner); err != nil {
		t.Fatalf("failed to decode planner role: %v", err)
	}
	for _, expected := range []string{
		"On Planner startup or workflow initiation, inspect active same-workspace peer sessions through the active harness transport before other workflow actions.",
		"Prioritize discovering and dispatching to existing dedicated Reviewer, Builder, or Scout peer sessions in the same workspace rather than spawning nested subagents.",
	} {
		if !slices.Contains(planner.Responsibilities, expected) {
			t.Errorf("expected planner responsibility %q", expected)
		}
	}
	if !slices.Contains(planner.Boundaries, "DO NOT skip active workspace peer-session discovery or spawn nested subagents when dedicated peer role sessions exist.") {
		t.Error("expected planner boundary to enforce peer-session discovery and prioritize dedicated peer roles")
	}

	// 2. Verify session-handoff-audit rule
	var rules []knowledge.Rule
	if err := json.Unmarshal(queryJSON("rule", "explain", "session-handoff-audit"), &rules); err != nil {
		t.Fatalf("failed to decode session-handoff-audit rule: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 session-handoff-audit rule, got %d", len(rules))
	}
	rule := rules[0]
	if rule.ID != "session-handoff-audit" || rule.Category != "protocol" {
		t.Errorf("unexpected rule metadata: %+v", rule)
	}
	if rule.Title == "" || rule.Summary == "" || rule.Details == "" {
		t.Errorf("expected non-empty rule title, summary, and details")
	}

	// Verify all 9 State & Topology Anchor fields in rule
	anchorFields := []string{"target_roots", "tier_topology", "role_assignments", "baseline_revision", "dirty_status", "recent_milestones", "utc_plan_paths", "next_pickup", "evidence"}
	for _, field := range anchorFields {
		if !strings.Contains(rule.Details, field) {
			t.Errorf("expected session-handoff-audit details to contain anchor field %q", field)
		}
	}

	// Verify all 4 mandatory audit dimensions in rule
	dimensions := []string{"Plan Understanding", "Progress & State Understanding", "Architectural & Governance Decisions", "Permissions & Path-Scoping Invariants"}
	for _, dim := range dimensions {
		if !strings.Contains(rule.Details, dim) {
			t.Errorf("expected session-handoff-audit details to contain dimension %q", dim)
		}
		foundInGuideline := false
		for _, g := range rule.Guidelines {
			if strings.Contains(g, dim) {
				foundInGuideline = true
				break
			}
		}
		if !foundInGuideline {
			t.Errorf("expected session-handoff-audit guidelines to contain dimension %q", dim)
		}
	}

	// Verify release cadence & peer dispatch in rule
	for _, expectedPhrase := range []string{
		"daily bugfix/patch commits roll on main without SemVer tag bumps",
		"SemVer tags and release publication are reserved for consolidated milestones",
		"inspect active same-workspace peer sessions through the active harness transport",
		"prioritize dispatch to existing dedicated Reviewer, Builder, or Scout peer sessions rather than spawning nested subagents",
	} {
		found := strings.Contains(rule.Details, expectedPhrase)
		if !found {
			for _, g := range rule.Guidelines {
				if strings.Contains(g, expectedPhrase) {
					found = true
					break
				}
			}
		}
		if !found {
			t.Errorf("expected session-handoff-audit rule to contain phrase %q", expectedPhrase)
		}
	}

	// 3. Verify init flow step 1 startup checklist
	var initFlow knowledge.Flow
	if err := json.Unmarshal(queryJSON("flow", "init"), &initFlow); err != nil {
		t.Fatalf("failed to decode init flow: %v", err)
	}
	initStep1 := initFlow.Steps[0]
	startupPrefix := "First inspect active same-workspace peer sessions through the active harness transport; prioritize discovery and dispatch to existing dedicated Reviewer, Builder, or Scout peer sessions rather than spawning nested subagents; then continue to repository scale assessment or anchor capture"
	if !strings.HasPrefix(initStep1.Action, startupPrefix) {
		t.Errorf("expected init step 1 action to start with startup checklist prefix %q, got: %s", startupPrefix, initStep1.Action)
	}

	// 4. Verify session-handoff flow
	var handoffFlow knowledge.Flow
	if err := json.Unmarshal(queryJSON("flow", "session-handoff"), &handoffFlow); err != nil {
		t.Fatalf("failed to decode session-handoff flow: %v", err)
	}
	if len(handoffFlow.Steps) != 8 {
		t.Fatalf("expected 8 steps in session-handoff flow, got %d", len(handoffFlow.Steps))
	}
	for _, step := range handoffFlow.Steps {
		if step.Actor != knowledge.RolePlanner {
			t.Errorf("step %d expected actor planner, got %q", step.Index, step.Actor)
		}
		for _, forbidden := range []string{"jj ", "git ", "git commit", "jj describe", "jj new", "sh ", "bash ", "herdr", "pane:", "cli:", "mcp"} {
			if strings.Contains(step.Action, forbidden) {
				t.Errorf("step %d action violates VCS/transport-neutral invariant by containing raw command or transport token %q: %s", step.Index, forbidden, step.Action)
			}
		}
	}

	if !strings.HasPrefix(handoffFlow.Steps[0].Action, startupPrefix) {
		t.Errorf("expected session-handoff step 1 action to start with startup checklist prefix %q, got: %s", startupPrefix, handoffFlow.Steps[0].Action)
	}

	// Verify exact condition counts, triggers, and targets across all 8 steps
	stepConditions := func(step knowledge.FlowStep) map[string]int {
		m := make(map[string]int)
		for _, c := range step.Conditions {
			m[c.When] = c.Then
		}
		return m
	}

	// Step 1: exactly 2 conditions
	s1 := handoffFlow.Steps[0]
	if len(s1.Conditions) != 2 {
		t.Errorf("expected step 1 to have exactly 2 conditions, got %d", len(s1.Conditions))
	}
	s1Cond := stepConditions(s1)
	if s1Cond["ANCHOR_CAPTURED"] != 2 || s1Cond["ANCHOR_INVALID"] != 1 {
		t.Errorf("unexpected step 1 conditions: %v", s1Cond)
	}

	// Step 2: exactly 2 conditions
	s2 := handoffFlow.Steps[1]
	if len(s2.Conditions) != 2 {
		t.Errorf("expected step 2 to have exactly 2 conditions, got %d", len(s2.Conditions))
	}
	s2Cond := stepConditions(s2)
	if s2Cond["GROUND_TRUTH_CONFIRMED"] != 3 || s2Cond["GROUND_TRUTH_CONFLICT"] != 1 {
		t.Errorf("unexpected step 2 conditions: %v", s2Cond)
	}

	// Step 3: exactly 2 conditions
	s3 := handoffFlow.Steps[2]
	if len(s3.Conditions) != 2 {
		t.Errorf("expected step 3 to have exactly 2 conditions, got %d", len(s3.Conditions))
	}
	s3Cond := stepConditions(s3)
	if s3Cond["ROUND_1_COMPLETE"] != 4 || s3Cond["AUDIT_BLOCKED"] != 3 {
		t.Errorf("unexpected step 3 conditions: %v", s3Cond)
	}
	if !strings.Contains(s3.Action, "Readiness Audit Round 1") {
		t.Errorf("expected step 3 action to mention Readiness Audit Round 1")
	}
	for _, dim := range dimensions {
		if !strings.Contains(s3.Action, dim) {
			t.Errorf("expected step 3 action to explicitly name dimension %q", dim)
		}
	}

	// Step 4: exactly 2 conditions
	s4 := handoffFlow.Steps[3]
	if len(s4.Conditions) != 2 {
		t.Errorf("expected step 4 to have exactly 2 conditions, got %d", len(s4.Conditions))
	}
	s4Cond := stepConditions(s4)
	if s4Cond["ROUND_2_COMPLETE"] != 5 || s4Cond["AUDIT_BLOCKED"] != 3 {
		t.Errorf("unexpected step 4 conditions: %v", s4Cond)
	}
	if !strings.Contains(s4.Action, "Readiness Audit Round 2") {
		t.Errorf("expected step 4 action to mention Readiness Audit Round 2")
	}
	for _, dim := range dimensions {
		if !strings.Contains(s4.Action, dim) {
			t.Errorf("expected step 4 action to explicitly name dimension %q", dim)
		}
	}

	// Step 5: exactly 2 conditions
	s5 := handoffFlow.Steps[4]
	if len(s5.Conditions) != 2 {
		t.Errorf("expected step 5 to have exactly 2 conditions, got %d", len(s5.Conditions))
	}
	s5Cond := stepConditions(s5)
	if s5Cond["ROUND_3_COMPLETE"] != 6 || s5Cond["QUESTIONS_REMAIN"] != 4 {
		t.Errorf("unexpected step 5 conditions: %v", s5Cond)
	}
	if !strings.Contains(s5.Action, "Readiness Audit Round 3") {
		t.Errorf("expected step 5 action to mention Readiness Audit Round 3")
	}
	for _, dim := range dimensions {
		if !strings.Contains(s5.Action, dim) {
			t.Errorf("expected step 5 action to explicitly name dimension %q", dim)
		}
	}

	// Step 6: exactly 2 conditions
	s6 := handoffFlow.Steps[5]
	if len(s6.Conditions) != 2 {
		t.Errorf("expected step 6 to have exactly 2 conditions, got %d", len(s6.Conditions))
	}
	s6Cond := stepConditions(s6)
	if s6Cond["TAKEOVER_PERMISSIONS_CLEAR"] != 7 || s6Cond["SCOPE_OR_BARRIER_VIOLATION"] != 3 {
		t.Errorf("unexpected step 6 conditions: %v", s6Cond)
	}

	// Step 7: exactly 2 conditions
	s7 := handoffFlow.Steps[6]
	if len(s7.Conditions) != 2 {
		t.Errorf("expected step 7 to have exactly 2 conditions, got %d", len(s7.Conditions))
	}
	s7Cond := stepConditions(s7)
	if s7Cond["TAKEOVER_READY"] != 8 || s7Cond["READINESS_BLOCKED"] != 4 {
		t.Errorf("unexpected step 7 conditions: %v", s7Cond)
	}

	// Step 8: exactly 0 conditions (terminal)
	s8 := handoffFlow.Steps[7]
	if len(s8.Conditions) != 0 {
		t.Errorf("expected step 8 to be terminal with 0 conditions, got %d", len(s8.Conditions))
	}

	// Test single step queries for session-handoff
	var step1Query knowledge.FlowStep
	if err := json.Unmarshal(queryJSON("flow", "session-handoff", "--step", "1"), &step1Query); err != nil {
		t.Fatalf("failed to decode session-handoff step 1: %v", err)
	}
	if step1Query.Index != 1 || step1Query.Actor != knowledge.RolePlanner {
		t.Errorf("unexpected session-handoff step 1 query: %+v", step1Query)
	}

	// 5. Verify documentation synchronization across README.md and SKILL.md
	for _, docPath := range []string{"../../README.md", "../../SKILL.md"} {
		content, err := os.ReadFile(docPath)
		if err != nil {
			t.Fatalf("failed to read doc file %q: %v", docPath, err)
		}
		docStr := string(content)
		for _, required := range []string{
			"session-handoff",
			"session-handoff-audit",
			"active workspace peer-session discovery",
			"dedicated Reviewer, Builder, or Scout peer sessions",
			"active harness transport",
			"nested subagent",
		} {
			if !strings.Contains(docStr, required) {
				t.Errorf("expected doc file %q to contain %q", docPath, required)
			}
		}
	}
}

func TestCLI_AIReviewerSpecThreeTierIntegration(t *testing.T) {
	t.Parallel()

	queryJSON := func(args ...string) []byte {
		t.Helper()
		var stdout, stderr bytes.Buffer
		if err := cli.Execute(args, &stdout, &stderr, "v0.1.0"); err != nil {
			t.Fatalf("query %q failed: %v", strings.Join(args, " "), err)
		}
		return stdout.Bytes()
	}

	// 1. Pillar 1: Planner reviewability & coverage
	{
		var rules []knowledge.Rule
		if err := json.Unmarshal(queryJSON("rule", "explain", "planner-reviewability"), &rules); err != nil {
			t.Fatalf("failed to decode planner-reviewability rule: %v", err)
		}
		if len(rules) != 1 {
			t.Fatalf("expected 1 planner-reviewability rule, got %d", len(rules))
		}
		r := rules[0]
		for _, signal := range []string{"coupling", "cross-cutting", "mixed concerns", "verification heterogeneity"} {
			if !strings.Contains(r.Details, signal) {
				t.Errorf("expected planner-reviewability details to contain signal %q", signal)
			}
		}
		if !strings.Contains(r.Details, "Review Plan") {
			t.Error("expected planner-reviewability to mention Review Plan verification path")
		}
		// Falsifiable check: rejects rigid >400 LOC threshold while preserving signals and coverage
		if !strings.Contains(r.Summary, "no >400 LOC hard limit") {
			t.Errorf("expected planner-reviewability summary to reject >400 LOC limit, got: %s", r.Summary)
		}
		if !strings.Contains(r.Details, "not an arbitrary or rigid line-count threshold like >400 LOC") {
			t.Errorf("expected planner-reviewability details to reject rigid line count, got: %s", r.Details)
		}
		foundNoRigidThreshold := false
		for _, g := range r.Guidelines {
			if strings.Contains(g, "rather than a rigid line-count threshold") {
				foundNoRigidThreshold = true
			}
			if strings.Contains(g, "must not exceed") && strings.Contains(g, "LOC") {
				t.Errorf("falsifiable violation: guideline must not impose hard LOC limit: %s", g)
			}
		}
		if !foundNoRigidThreshold {
			t.Error("expected planner-reviewability guidelines to explicitly reject a rigid line-count threshold")
		}
	}

	// 2. Pillar 2: Severity semantics
	{
		var rules []knowledge.Rule
		if err := json.Unmarshal(queryJSON("rule", "explain", "review-severity-semantics"), &rules); err != nil {
			t.Fatalf("failed to decode review-severity-semantics rule: %v", err)
		}
		if len(rules) != 1 {
			t.Fatalf("expected 1 review-severity-semantics rule, got %d", len(rules))
		}
		r := rules[0]
		for _, sev := range []string{"Blocker", "Major", "Minor", "Other"} {
			if !strings.Contains(r.Details, sev) {
				t.Errorf("expected review-severity-semantics details to contain severity %q", sev)
			}
		}
		// Negative check: Critical is excluded
		if strings.Contains(r.Summary, "Critical") {
			t.Error("expected review-severity-semantics summary to exclude Critical")
		}
		for _, g := range r.Guidelines {
			if strings.Contains(g, "Critical") && !strings.Contains(g, "strictly excluded") {
				t.Errorf("expected Critical to be marked strictly excluded in guidelines, got: %s", g)
			}
		}
	}

	// 3. Pillar 3: Track A local behavioral verification
	{
		var rules []knowledge.Rule
		if err := json.Unmarshal(queryJSON("rule", "explain", "tdd-reproduction"), &rules); err != nil {
			t.Fatalf("failed to decode tdd-reproduction rule: %v", err)
		}
		if len(rules) != 1 {
			t.Fatalf("expected 1 tdd-reproduction rule, got %d", len(rules))
		}
		r := rules[0]
		if !strings.Contains(r.Title, "Track A") {
			t.Errorf("expected tdd-reproduction title to include Track A, got: %s", r.Title)
		}
		if !strings.Contains(r.Details, "static/specification evidence") {
			t.Error("expected tdd-reproduction to mention static/specification evidence for non-behavioral findings")
		}
	}

	// 4. Pillar 4: Track B action differential verification
	{
		var rules []knowledge.Rule
		if err := json.Unmarshal(queryJSON("rule", "explain", "track-b-action-differential-verification"), &rules); err != nil {
			t.Fatalf("failed to decode track-b-action-differential-verification rule: %v", err)
		}
		if len(rules) != 1 {
			t.Fatalf("expected 1 track-b-action-differential-verification rule, got %d", len(rules))
		}
		r := rules[0]
		for _, required := range []string{"action", "boundary", "baseline_identity", "environment", "metrics", "sampling", "criteria", "median", "p95", "variance"} {
			if !strings.Contains(r.Details, required) {
				t.Errorf("expected track-b rule details to contain %q", required)
			}
		}
		if !strings.Contains(r.Details, "unexplained persistent differential") {
			t.Error("expected track-b rule to require investigation of unexplained persistent differential")
		}
	}

	// 5. Pillar 5: Out-of-tree baseline mirror
	{
		var rules []knowledge.Rule
		if err := json.Unmarshal(queryJSON("rule", "explain", "out-of-tree-baseline-mirror"), &rules); err != nil {
			t.Fatalf("failed to decode out-of-tree-baseline-mirror rule: %v", err)
		}
		if len(rules) != 1 {
			t.Fatalf("expected 1 out-of-tree-baseline-mirror rule, got %d", len(rules))
		}
		r := rules[0]
		for _, required := range []string{
			"persistent",
			"reconstructible",
			"BASELINE_STALE",
			"ACCEPTED",
			"PUBLISHED",
			"repository_identity",
			"baseline_identity",
			"creation",
			"synchronization",
			"advancement",
			"retirement",
			"post-seal equivalence",
			"forbidden on REVIEW_PASS",
			"commit success",
			"publication",
		} {
			if !strings.Contains(r.Details, required) {
				t.Errorf("expected out-of-tree-baseline-mirror details to contain %q", required)
			}
		}

		foundAdvancementCondition := false
		for _, g := range r.Guidelines {
			if strings.Contains(g, "post-seal equivalence") && strings.Contains(g, "strictly forbidden on REVIEW_PASS, commit success, or publication") {
				foundAdvancementCondition = true
			}
		}
		if !foundAdvancementCondition {
			t.Error("expected out-of-tree-baseline-mirror guideline to assert post-seal equivalence and forbidden advancement triggers")
		}
	}

	// 6. Pillar 6: Structured artifacts & resolution
	{
		var planArtifact knowledge.Artifact
		if err := json.Unmarshal(queryJSON("artifact", "review-plan"), &planArtifact); err != nil {
			t.Fatalf("failed to decode review-plan artifact: %v", err)
		}
		foundTrackB := false
		for _, s := range planArtifact.Sections {
			if s.Name == "Track B Definition" {
				foundTrackB = true
				for _, reqField := range []string{"action", "boundary", "baseline_identity", "environment", "metrics", "sampling", "criteria"} {
					if !strings.Contains(s.Description, reqField) {
						t.Errorf("expected Track B section in review-plan to mention field %q", reqField)
					}
				}
			}
		}
		if !foundTrackB {
			t.Error("expected review-plan to include Track B Definition section")
		}

		var resolution knowledge.Artifact
		if err := json.Unmarshal(queryJSON("artifact", "review-resolution"), &resolution); err != nil {
			t.Fatalf("failed to decode review-resolution artifact: %v", err)
		}
		if resolution.Name != "review-resolution" || resolution.Owner != knowledge.RolePlanner || resolution.Type != "document" {
			t.Errorf("unexpected review-resolution metadata: %+v", resolution)
		}
		if resolution.Path != "plan/{timestamp}/{slug}.resolution.md" {
			t.Errorf("expected resolution path 'plan/{timestamp}/{slug}.resolution.md', got %q", resolution.Path)
		}
		if resolution.PathVariables["timestamp"] == "" || resolution.PathVariables["slug"] == "" {
			t.Errorf("expected path_variables timestamp and slug in review-resolution, got %+v", resolution.PathVariables)
		}
		if len(resolution.Visibility) != 3 || resolution.Visibility[0] != knowledge.RolePlanner || resolution.Visibility[1] != knowledge.RoleBuilder || resolution.Visibility[2] != knowledge.RoleReviewer {
			t.Errorf("expected resolution visibility [planner builder reviewer], got %v", resolution.Visibility)
		}
		for _, reqPhrase := range []string{"sanitization", "review-plan criteria", "hidden test fixtures", "private inspection methods"} {
			if !strings.Contains(resolution.Description, reqPhrase) {
				t.Errorf("expected resolution description to contain %q, got: %s", reqPhrase, resolution.Description)
			}
		}

		expectedSections := []string{"Outcome", "Resolved Findings", "Deviations & Rationales", "Residual Risks", "Verification Evidence"}
		for i, sec := range expectedSections {
			if resolution.Sections[i].Name != sec {
				t.Errorf("expected section %d to be %q, got %q", i, sec, resolution.Sections[i].Name)
			}
		}

		var findings knowledge.Artifact
		if err := json.Unmarshal(queryJSON("artifact", "review-findings"), &findings); err != nil {
			t.Fatalf("failed to decode review-findings artifact: %v", err)
		}
		if len(findings.Visibility) != 2 || findings.Visibility[0] != knowledge.RolePlanner || findings.Visibility[1] != knowledge.RoleReviewer {
			t.Errorf("expected review-findings visibility strictly [planner reviewer], got %v", findings.Visibility)
		}
		if slices.Contains(findings.Visibility, knowledge.RoleBuilder) || slices.Contains(findings.Visibility, knowledge.RoleScout) {
			t.Errorf("review-findings must strictly exclude builder and scout, got: %v", findings.Visibility)
		}
		fieldMap := make(map[string]knowledge.ArtifactField)
		for _, f := range findings.Fields {
			fieldMap[f.Name] = f
		}
		if !fieldMap["evidence_mode"].Required || !fieldMap["evidence"].Required {
			t.Error("expected evidence_mode and evidence to be required in review-findings")
		}
		if fieldMap["reproduction_scenario"].Required {
			t.Error("expected reproduction_scenario to be conditional/optional in review-findings")
		}
	}

	// 7. Roles & Flows assertions
	{
		var planner knowledge.RoleDefinition
		if err := json.Unmarshal(queryJSON("role", "planner"), &planner); err != nil {
			t.Fatalf("failed to decode planner role: %v", err)
		}
		foundReviewability := false
		for _, resp := range planner.Responsibilities {
			if strings.Contains(resp, "reviewable implementation units") {
				foundReviewability = true
				break
			}
		}
		if !foundReviewability {
			t.Error("expected planner responsibility for reviewable implementation units")
		}

		var reviewer knowledge.RoleDefinition
		if err := json.Unmarshal(queryJSON("role", "reviewer"), &reviewer); err != nil {
			t.Fatalf("failed to decode reviewer role: %v", err)
		}
		foundCoverage := false
		for _, resp := range reviewer.Responsibilities {
			if strings.Contains(resp, "verification coverage") {
				foundCoverage = true
				break
			}
		}
		if !foundCoverage {
			t.Error("expected reviewer responsibility for verification coverage audit")
		}

		var planFlow knowledge.Flow
		if err := json.Unmarshal(queryJSON("flow", "plan"), &planFlow); err != nil {
			t.Fatalf("failed to decode plan flow: %v", err)
		}
		for _, step := range planFlow.Steps {
			for _, forbidden := range []string{"jj ", "git ", "git commit", "jj describe", "jj new", "sh -c", "bash ", "herdr", "pane:", "cli:", "mcp"} {
				if strings.Contains(step.Action, forbidden) {
					t.Errorf("plan flow step %d violates command neutrality: %s", step.Index, step.Action)
				}
			}
		}

		var reviewFlow knowledge.Flow
		if err := json.Unmarshal(queryJSON("flow", "review"), &reviewFlow); err != nil {
			t.Fatalf("failed to decode review flow: %v", err)
		}
		for _, step := range reviewFlow.Steps {
			for _, forbidden := range []string{"jj ", "git ", "git commit", "jj describe", "jj new", "sh -c", "bash ", "herdr", "pane:", "cli:", "mcp"} {
				if strings.Contains(step.Action, forbidden) {
					t.Errorf("review flow step %d violates command neutrality: %s", step.Index, step.Action)
				}
			}
		}
	}

	// 8. Independent Documentation Assertions
	for _, docPath := range []string{"../../README.md", "../../SKILL.md"} {
		content, err := os.ReadFile(docPath)
		if err != nil {
			t.Fatalf("failed to read doc file %q: %v", docPath, err)
		}
		docStr := string(content)
		for _, required := range []string{
			"planner-reviewability",
			"review-severity-semantics",
			"track-b-action-differential-verification",
			"out-of-tree-baseline-mirror",
			"review-resolution",
			"Track A",
			"Track B",
			"Blocker",
			"Major",
			"Minor",
			"Other",
		} {
			if !strings.Contains(docStr, required) {
				t.Errorf("expected doc file %q to contain six-pillar term %q", docPath, required)
			}
		}
	}
}

func TestCLI_V030_BlueprintAndGovernance(t *testing.T) {
	t.Parallel()

	queryJSON := func(args ...string) []byte {
		t.Helper()
		var stdout, stderr bytes.Buffer
		if err := cli.Execute(args, &stdout, &stderr, "v0.3.0"); err != nil {
			t.Fatalf("query %q failed: %v", strings.Join(args, " "), err)
		}
		return stdout.Bytes()
	}

	// 1. Artifacts & Scout Isolation
	{
		for _, artName := range []string{
			"blueprint-plan",
			"sub-build-plan",
			"sub-review-plan",
			"sub-review-resolution",
			"review-resolution",
			"build-plan",
			"review-plan",
		} {
			var a knowledge.Artifact
			if err := json.Unmarshal(queryJSON("artifact", artName), &a); err != nil {
				t.Fatalf("failed to decode %s: %v", artName, err)
			}
			if slices.Contains(a.Visibility, knowledge.RoleScout) {
				t.Errorf("scout must be strictly excluded from in-flight task artifact %q", artName)
			}
		}

		var bp knowledge.Artifact
		if err := json.Unmarshal(queryJSON("artifact", "blueprint-plan"), &bp); err != nil {
			t.Fatalf("failed to decode blueprint-plan: %v", err)
		}
		if bp.Path != "plan/{timestamp}/{slug}.blueprint.md" {
			t.Errorf("unexpected blueprint-plan path: %s", bp.Path)
		}

		var rf knowledge.Artifact
		if err := json.Unmarshal(queryJSON("artifact", "review-findings"), &rf); err != nil {
			t.Fatalf("failed to decode review-findings: %v", err)
		}
		if len(rf.Visibility) != 2 || rf.Visibility[0] != knowledge.RolePlanner || rf.Visibility[1] != knowledge.RoleReviewer {
			t.Errorf("expected review-findings visibility strictly [planner reviewer], got %v", rf.Visibility)
		}
		if slices.Contains(rf.Visibility, knowledge.RoleBuilder) || slices.Contains(rf.Visibility, knowledge.RoleScout) {
			t.Errorf("review-findings must strictly exclude builder and scout, got: %v", rf.Visibility)
		}
	}

	// 2. Rules Evolution & Legacy Pruning
	{
		// Assert legacy rules are removed
		var stdout, stderr bytes.Buffer
		if err := cli.Execute([]string{"rule", "explain", "atomic-change-units"}, &stdout, &stderr, "dev"); err == nil {
			t.Error("expected error querying removed rule atomic-change-units, got nil")
		}
		stdout.Reset()
		stderr.Reset()
		if err := cli.Execute([]string{"rule", "explain", "commit-authority-separation"}, &stdout, &stderr, "dev"); err == nil {
			t.Error("expected error querying removed rule commit-authority-separation, got nil")
		}

		// Assert coherent-plan-units
		var coherentRules []knowledge.Rule
		if err := json.Unmarshal(queryJSON("rule", "explain", "coherent-plan-units"), &coherentRules); err != nil {
			t.Fatalf("failed to decode coherent-plan-units rule: %v", err)
		}
		if len(coherentRules) != 1 {
			t.Fatalf("expected 1 coherent-plan-units rule, got %d", len(coherentRules))
		}
		for _, term := range []string{"root intent", "coupling", "cross-cutting", "mixed concerns", "verification heterogeneity"} {
			if !strings.Contains(coherentRules[0].Details, term) {
				t.Errorf("expected coherent-plan-units to contain %q", term)
			}
		}

		// Assert anti-rubber-stamp-plan-gate
		var antiStampRules []knowledge.Rule
		if err := json.Unmarshal(queryJSON("rule", "explain", "anti-rubber-stamp-plan-gate"), &antiStampRules); err != nil {
			t.Fatalf("failed to decode anti-rubber-stamp-plan-gate rule: %v", err)
		}
		if len(antiStampRules) != 1 {
			t.Fatalf("expected 1 anti-rubber-stamp-plan-gate rule, got %d", len(antiStampRules))
		}
		for _, term := range []string{"SPLIT_ATTEMPT", "SPLIT_REJECTED_BECAUSE", "Counterfactual Decomposition Challenge"} {
			if !strings.Contains(antiStampRules[0].Details, term) {
				t.Errorf("expected anti-rubber-stamp-plan-gate to contain %q", term)
			}
		}

		// Assert evidence-proportional-persistence
		var persistenceRules []knowledge.Rule
		if err := json.Unmarshal(queryJSON("rule", "explain", "evidence-proportional-persistence"), &persistenceRules); err != nil {
			t.Fatalf("failed to decode evidence-proportional-persistence rule: %v", err)
		}
		if len(persistenceRules) != 1 {
			t.Fatalf("expected 1 evidence-proportional-persistence rule, got %d", len(persistenceRules))
		}
		for _, term := range []string{"proportional persistence", "sub-review-resolution", "review-resolution", "AGENTS.md"} {
			if !strings.Contains(persistenceRules[0].Details, term) {
				t.Errorf("expected evidence-proportional-persistence to contain %q", term)
			}
		}

		// Assert acceptance-publication-authority
		var acceptanceRules []knowledge.Rule
		if err := json.Unmarshal(queryJSON("rule", "explain", "acceptance-publication-authority"), &acceptanceRules); err != nil {
			t.Fatalf("failed to decode acceptance-publication-authority rule: %v", err)
		}
		if len(acceptanceRules) != 1 {
			t.Fatalf("expected 1 acceptance-publication-authority rule, got %d", len(acceptanceRules))
		}
		for _, term := range []string{"WORKING", "ACCEPTED", "PUBLISHED", "Finalization Equivalence", "AUTHORIZATION_DENIED"} {
			if !strings.Contains(acceptanceRules[0].Details, term) {
				t.Errorf("expected acceptance-publication-authority to contain %q", term)
			}
		}
	}

	// 3. Blueprint Flow & Two-Tier Gating
	{
		var bpFlow knowledge.Flow
		if err := json.Unmarshal(queryJSON("flow", "blueprint"), &bpFlow); err != nil {
			t.Fatalf("failed to decode blueprint flow: %v", err)
		}
		if len(bpFlow.Steps) != 12 {
			t.Fatalf("expected 12 steps in blueprint flow, got %d", len(bpFlow.Steps))
		}

		// Assert Blueprint Gate (step 2)
		if bpFlow.Steps[1].Actor != knowledge.RoleReviewer || !strings.Contains(bpFlow.Steps[1].Action, "Blueprint Gate") {
			t.Errorf("expected Step 2 to be Reviewer Blueprint Gate, got: %+v", bpFlow.Steps[1])
		}

		// Assert Sub-Plan Gate (step 4)
		if bpFlow.Steps[3].Actor != knowledge.RoleReviewer || !strings.Contains(bpFlow.Steps[3].Action, "Sub-Plan Gate") {
			t.Errorf("expected Step 4 to be Reviewer Sub-Plan Gate, got: %+v", bpFlow.Steps[3])
		}

		// Assert Mediation (step 7)
		if bpFlow.Steps[6].Actor != knowledge.RolePlanner || !strings.Contains(bpFlow.Steps[6].Action, "Mediate and sanitize") {
			t.Errorf("expected Step 7 to be Planner Mediation, got: %+v", bpFlow.Steps[6])
		}

		// Assert Sub-Resolution Synthesis & Verification (steps 8 & 9)
		if bpFlow.Steps[7].Actor != knowledge.RolePlanner || !strings.Contains(bpFlow.Steps[7].Action, "sub-review-resolution") {
			t.Errorf("expected Step 8 to be Planner Sub-Resolution Synthesis, got: %+v", bpFlow.Steps[7])
		}
		if bpFlow.Steps[8].Actor != knowledge.RoleReviewer || !strings.Contains(bpFlow.Steps[8].Action, "sub-review-resolution") {
			t.Errorf("expected Step 9 to be Reviewer Sub-Resolution Verification, got: %+v", bpFlow.Steps[8])
		}

		// Assert Feature Composition Gate (step 11)
		if bpFlow.Steps[10].Actor != knowledge.RoleReviewer || !strings.Contains(bpFlow.Steps[10].Action, "Feature Composition Gate") {
			t.Errorf("expected Step 11 to be Reviewer Feature Composition Gate, got: %+v", bpFlow.Steps[10])
		}

		// Assert BASELINE_STALE fail-closed routes and Track B action text
		if !strings.Contains(bpFlow.Steps[5].Action, "When Track B is selected, assert the pinned baseline identity") {
			t.Errorf("expected blueprint step 6 action to mention Track B pinned baseline identity, got: %s", bpFlow.Steps[5].Action)
		}
		bpStep6Conditions := make(map[string]int)
		for _, c := range bpFlow.Steps[5].Conditions {
			bpStep6Conditions[c.When] = c.Then
		}
		if bpStep6Conditions["BASELINE_STALE"] != 2 {
			t.Errorf("expected blueprint step 6 BASELINE_STALE condition to point to step 2, got: %v", bpStep6Conditions)
		}

		var revFlow knowledge.Flow
		if err := json.Unmarshal(queryJSON("flow", "review"), &revFlow); err != nil {
			t.Fatalf("failed to decode review flow: %v", err)
		}
		if revFlow.Steps[0].Actor != knowledge.RolePlanner || len(revFlow.Steps[0].Conditions) != 0 {
			t.Errorf("expected review step 1 to be planner handoff with 0 conditions, got actor %q, conditions: %v", revFlow.Steps[0].Actor, revFlow.Steps[0].Conditions)
		}
		if revFlow.Steps[1].Actor != knowledge.RoleReviewer {
			t.Errorf("expected review step 2 actor to be reviewer, got %q", revFlow.Steps[1].Actor)
		}
		if !strings.Contains(revFlow.Steps[1].Action, "When Track B is selected, assert the pinned baseline identity") {
			t.Errorf("expected review step 2 action to mention Track B pinned baseline identity, got: %s", revFlow.Steps[1].Action)
		}
		revStep2Conditions := make(map[string]int)
		for _, c := range revFlow.Steps[1].Conditions {
			revStep2Conditions[c.When] = c.Then
		}
		if revStep2Conditions["BASELINE_STALE"] != 8 || revStep2Conditions["REVIEW_PASS"] != 6 || revStep2Conditions["FINDINGS_REPORTED"] != 3 {
			t.Errorf("expected review step 2 conditions [BASELINE_STALE: 8, REVIEW_PASS: 6, FINDINGS_REPORTED: 3], got: %v", revStep2Conditions)
		}
	}

	// 4. Obsolete Script Removal
	{
		if _, err := os.Stat("../../scripts/run-workflow.sh"); !os.IsNotExist(err) {
			t.Errorf("expected scripts/run-workflow.sh to be deleted, but it still exists (err: %v)", err)
		}
	}

	// 5. Tier 3 Data Layer VCS Neutrality
	{
		for _, flowName := range []string{"init", "plan", "blueprint", "build", "review", "commit", "session-handoff"} {
			var f knowledge.Flow
			if err := json.Unmarshal(queryJSON("flow", flowName), &f); err != nil {
				t.Fatalf("failed to decode flow %s: %v", flowName, err)
			}
			for _, step := range f.Steps {
				for _, forbidden := range []string{"jj ", "git ", "git commit", "jj describe", "jj new", "sh -c", "bash ", "herdr", "pane:", "cli:", "mcp"} {
					if strings.Contains(step.Action, forbidden) {
						t.Errorf("flow %s step %d violates command neutrality: %s", flowName, step.Index, step.Action)
					}
				}
			}
		}
	}

	// 6. Independent Documentation Synchronization
	{
		for _, docPath := range []string{"../../README.md", "../../SKILL.md"} {
			content, err := os.ReadFile(docPath)
			if err != nil {
				t.Fatalf("failed to read doc file %q: %v", docPath, err)
			}
			docStr := string(content)
			for _, required := range []string{
				"blueprint-plan",
				"sub-build-plan",
				"sub-review-plan",
				"sub-review-resolution",
				"coherent-plan-units",
				"anti-rubber-stamp-plan-gate",
				"evidence-proportional-persistence",
				"acceptance-publication-authority",
				"Blueprint Gate",
				"Sub-Plan Gate",
				"Feature Composition Gate",
				"Finalization Equivalence",
			} {
				if !strings.Contains(docStr, required) {
					t.Errorf("expected doc file %q to contain v0.3.0 term %q", docPath, required)
				}
			}
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
