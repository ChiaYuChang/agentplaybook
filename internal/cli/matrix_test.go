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
		{"role", "builder", "--responsibility"},
		{"role", "builder", "--boundary"},
		{"role", "builder", "--communication"},
		{"flow", "init"},
		{"flow", "init", "--step", "1"},
		{"flow", "plan"},
		{"flow", "build"},
		{"flow", "review"},
		{"artifact", "repo-summary"},
		{"artifact", "build-plan"},
		{"artifact", "review-plan"},
		{"artifact", "review-findings"},
		{"rule", "list"},
		{"rule", "explain", "anti-cheating"},
		{"rule", "explain", "atomic-change-units"},
		{"rule", "explain", "mandatory-alignment"},
		{"rule", "explain", "tdd-reproduction"},
		{"rule", "explain", "ephemeral-communication-buffers"},
		{"rule", "explain", "event-driven-transport-coordination"},
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
	if step6.Action != "Inspect the working copy diff against declared in-scope boundaries, execute the atomic Conventional Commit under Planner VCS governance, and advance to the next step or conclude build." {
		t.Errorf("expected build flow Step 6 to use Planner VCS governance, got %q", step6.Action)
	}

	skill, err := os.ReadFile("../../SKILL.md")
	if err != nil {
		t.Fatalf("failed to read SKILL.md: %v", err)
	}
	for _, expected := range []string{
		"Git: Planner stages in-scope files and runs `git commit -m '...'` (which advances the active branch).",
		"Jujutsu: Planner describes the finalized revision with `jj describe -m '...'` and opens the next revision with `jj new`.",
		"`jj bookmark set <name> -r @-`",
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
