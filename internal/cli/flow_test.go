package cli_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ChiaYuChang/agentplaybook/internal/cli"
	"github.com/ChiaYuChang/agentplaybook/internal/knowledge"
)

func TestFlow_BareDiscovery(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"flow"}, &stdout, &stderr, "dev")
	if err != nil {
		t.Fatalf("expected bare 'flow' to succeed, got: %v", err)
	}

	out := stdout.String()
	for _, expected := range []string{"init", "plan", "blueprint", "build", "review", "commit", "cartography", "session-handoff", "e2e"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected flow %q in discovery output, got: %s", expected, out)
		}
	}
}

func TestFlow_QueryFull(t *testing.T) {
	t.Parallel()

	// 1. init flow
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"flow", "init"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("querying flow init failed: %v", err)
		}

		var f knowledge.Flow
		if err := json.Unmarshal(stdout.Bytes(), &f); err != nil {
			t.Fatalf("failed to decode JSON response: %v\nRaw: %s", err, stdout.String())
		}

		if f.Name != "init" {
			t.Errorf("expected flow name 'init', got %q", f.Name)
		}
		if len(f.Steps) != 9 {
			t.Errorf("expected 9 steps in init, got %d", len(f.Steps))
		}
		step1 := f.Steps[0]
		conditions := make(map[string]int)
		for _, c := range step1.Conditions {
			conditions[c.When] = c.Then
		}
		if conditions["DIRECT_SURVEY"] != 3 {
			t.Errorf("expected DIRECT_SURVEY to target step 3, got %d", conditions["DIRECT_SURVEY"])
		}
		if conditions["SCOUT_RECON_REQUIRED"] != 2 {
			t.Errorf("expected SCOUT_RECON_REQUIRED to target step 2, got %d", conditions["SCOUT_RECON_REQUIRED"])
		}
	}

	// 2. plan flow (reviewability and Track B)
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"flow", "plan"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("querying flow plan failed: %v", err)
		}

		var f knowledge.Flow
		if err := json.Unmarshal(stdout.Bytes(), &f); err != nil {
			t.Fatalf("failed to decode JSON response: %v\nRaw: %s", err, stdout.String())
		}
		if len(f.Steps) != 5 {
			t.Fatalf("expected 5 steps in plan flow, got %d", len(f.Steps))
		}
		if !strings.Contains(f.Steps[0].Action, "Hierarchical Blueprint") {
			t.Errorf("expected plan step 1 action to mention Hierarchical Blueprint, got: %s", f.Steps[0].Action)
		}
		if !strings.Contains(f.Steps[2].Action, "Counterfactual Decomposition Challenge") {
			t.Errorf("expected plan step 3 action to mention Counterfactual Decomposition Challenge, got: %s", f.Steps[2].Action)
		}
	}

	// 3. blueprint flow (12 steps)
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"flow", "blueprint"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("querying flow blueprint failed: %v", err)
		}

		var f knowledge.Flow
		if err := json.Unmarshal(stdout.Bytes(), &f); err != nil {
			t.Fatalf("failed to decode JSON response: %v\nRaw: %s", err, stdout.String())
		}
		if len(f.Steps) != 12 {
			t.Fatalf("expected 12 steps in blueprint flow, got %d", len(f.Steps))
		}
		if f.Steps[0].Actor != knowledge.RolePlanner || f.Steps[1].Actor != knowledge.RoleReviewer {
			t.Errorf("unexpected actors in blueprint steps 1-2: %s, %s", f.Steps[0].Actor, f.Steps[1].Actor)
		}
		if !strings.Contains(f.Steps[5].Action, "When Track B is selected, assert the pinned baseline identity") {
			t.Errorf("expected blueprint step 6 action to mention Track B pinned baseline identity assertion, got: %s", f.Steps[5].Action)
		}
		bpStep6Conditions := make(map[string]int)
		for _, c := range f.Steps[5].Conditions {
			bpStep6Conditions[c.When] = c.Then
		}
		if bpStep6Conditions["BASELINE_STALE"] != 2 {
			t.Errorf("expected blueprint step 6 BASELINE_STALE condition to point to step 2, got: %v", bpStep6Conditions)
		}
	}

	// 4. review flow (severities, Track A/B, IMPLEMENTATION_REVIEW_PASS, and terminal Step 6)
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"flow", "review"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("querying flow review failed: %v", err)
		}

		var f knowledge.Flow
		if err := json.Unmarshal(stdout.Bytes(), &f); err != nil {
			t.Fatalf("failed to decode JSON response: %v\nRaw: %s", err, stdout.String())
		}
		if len(f.Steps) != 8 {
			t.Fatalf("expected 8 steps in review flow, got %d", len(f.Steps))
		}
		if f.Steps[0].Actor != knowledge.RolePlanner || len(f.Steps[0].Conditions) != 0 {
			t.Errorf("expected review step 1 to be planner handoff with 0 conditions, got actor %q, conditions: %v", f.Steps[0].Actor, f.Steps[0].Conditions)
		}
		if f.Steps[1].Actor != knowledge.RoleReviewer {
			t.Errorf("expected review step 2 actor to be reviewer, got %q", f.Steps[1].Actor)
		}
		step2Action := f.Steps[1].Action
		if !strings.Contains(step2Action, "When Track B is selected, assert the pinned baseline identity") {
			t.Errorf("expected review step 2 action to mention Track B pinned baseline identity assertion, got: %s", step2Action)
		}
		reviewStep2Conditions := make(map[string]int)
		for _, c := range f.Steps[1].Conditions {
			reviewStep2Conditions[c.When] = c.Then
		}
		if reviewStep2Conditions["BASELINE_STALE"] != 8 || reviewStep2Conditions["REVIEW_PASS"] != 6 || reviewStep2Conditions["IMPLEMENTATION_REVIEW_PASS"] != 6 || reviewStep2Conditions["FINDINGS_REPORTED"] != 3 {
			t.Errorf("expected review step 2 conditions [BASELINE_STALE: 8, REVIEW_PASS: 6, IMPLEMENTATION_REVIEW_PASS: 6, FINDINGS_REPORTED: 3], got: %v", reviewStep2Conditions)
		}
		if !strings.Contains(step2Action, "formal severities (Blocker, Major, Minor, Other)") {
			t.Errorf("expected review step 2 to classify formal severities, got: %s", step2Action)
		}
		for _, reqPhrase := range []string{
			"unresolved Blocker blocks REVIEW_PASS",
			"resolved or arbitrated Blocker and resolved or waived Major may pass",
			"Minor and Other do not block",
		} {
			if !strings.Contains(step2Action, reqPhrase) {
				t.Errorf("expected review step 2 action to contain %q, got: %s", reqPhrase, step2Action)
			}
		}

		if !strings.Contains(f.Steps[3].Action, "Track A failing reproduction test") || !strings.Contains(f.Steps[3].Action, "static/specification evidence") {
			t.Errorf("expected review step 4 to distinguish Track A from static evidence, got: %s", f.Steps[3].Action)
		}
		if !strings.Contains(f.Steps[5].Action, "review-resolution") || !strings.Contains(f.Steps[5].Action, "E2E_REQUIRED") || !strings.Contains(f.Steps[5].Action, "E2E_NOT_REQUIRED") || !f.Steps[5].Terminal || len(f.Steps[5].Conditions) != 0 {
			t.Errorf("expected review step 6 to be terminal with E2E classification and review-resolution synthesis, got: %+v", f.Steps[5])
		}
		if !strings.Contains(f.Steps[6].Action, "arbitrate Blocker") || !strings.Contains(f.Steps[6].Action, "record waiver with rationale for Major") {
			t.Errorf("expected review step 7 to handle Blocker arbitration and Major waivers, got: %s", f.Steps[6].Action)
		}
	}

	// 5. e2e flow (12 steps, admission gate, remediation loop, terminal steps 11 and 12)
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"flow", "e2e"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("querying flow e2e failed: %v", err)
		}

		var f knowledge.Flow
		if err := json.Unmarshal(stdout.Bytes(), &f); err != nil {
			t.Fatalf("failed to decode JSON response: %v\nRaw: %s", err, stdout.String())
		}
		if len(f.Steps) != 12 {
			t.Fatalf("expected 12 steps in e2e flow, got %d", len(f.Steps))
		}

		// Step 5 conditions (all 8 outcomes)
		step5Conditions := make(map[string]int)
		for _, c := range f.Steps[4].Conditions {
			step5Conditions[c.When] = c.Then
		}
		expectedStep5 := map[string]int{
			"E2E_RESULT_PASS":            6,
			"E2E_PRODUCT_FAILURE":        7,
			"E2E_BUILD_FAILURE":          7,
			"E2E_ENV_BLOCKED":            9,
			"E2E_CLARIFICATION_REQUIRED": 9,
			"E2E_TIMEOUT":                10,
			"E2E_CANCELLED":              10,
			"E2E_INVALID_EVIDENCE":       10,
		}
		for when, then := range expectedStep5 {
			if step5Conditions[when] != then {
				t.Errorf("expected step 5 condition %q -> %d, got %d", when, then, step5Conditions[when])
			}
		}

		// Step 6 Admission Gate
		step6Conditions := make(map[string]int)
		for _, c := range f.Steps[5].Conditions {
			step6Conditions[c.When] = c.Then
		}
		if step6Conditions["E2E_EVIDENCE_ADMITTED"] != 11 || step6Conditions["E2E_EVIDENCE_STALE"] != 1 || step6Conditions["E2E_INVALID_EVIDENCE"] != 10 {
			t.Errorf("unexpected step 6 admission conditions: %v", step6Conditions)
		}

		// Step 7 Remediation Dispatch
		if len(f.Steps[6].Conditions) != 1 || f.Steps[6].Conditions[0].When != "REMEDIATION_DISPATCHED" || f.Steps[6].Conditions[0].Then != 8 {
			t.Errorf("unexpected step 7 conditions: %v", f.Steps[6].Conditions)
		}

		// Step 8 Reviewer Affected Implementation Review
		step8Conditions := make(map[string]int)
		for _, c := range f.Steps[7].Conditions {
			step8Conditions[c.When] = c.Then
		}
		if step8Conditions["IMPLEMENTATION_REVIEW_PASS"] != 1 || step8Conditions["FINDINGS_REPORTED"] != 7 {
			t.Errorf("unexpected step 8 conditions: %v", step8Conditions)
		}

		// Step 9 Environment/Clarification Resolution
		step9Conditions := make(map[string]int)
		for _, c := range f.Steps[8].Conditions {
			step9Conditions[c.When] = c.Then
		}
		if step9Conditions["E2E_INPUTS_AMENDED"] != 1 || step9Conditions["CLARIFICATION_RESOLVED"] != 1 || step9Conditions["PACKAGE_DEFECT_ESCALATED"] != 7 {
			t.Errorf("unexpected step 9 conditions: %v", step9Conditions)
		}

		// Step 10 Timeout/Cancellation Arbitration
		step10Conditions := make(map[string]int)
		for _, c := range f.Steps[9].Conditions {
			step10Conditions[c.When] = c.Then
		}
		if step10Conditions["E2E_RETRY_AUTHORIZED"] != 1 || step10Conditions["PACKAGE_DEFECT_ESCALATED"] != 7 || step10Conditions["E2E_EXECUTION_HALTED"] != 12 {
			t.Errorf("unexpected step 10 conditions: %v", step10Conditions)
		}

		// Terminal Steps 11 and 12
		if !f.Steps[10].Terminal || len(f.Steps[10].Conditions) != 0 {
			t.Errorf("expected step 11 to be terminal with zero conditions, got: %+v", f.Steps[10])
		}
		if !f.Steps[11].Terminal || len(f.Steps[11].Conditions) != 0 {
			t.Errorf("expected step 12 to be terminal with zero conditions, got: %+v", f.Steps[11])
		}
	}
}

func TestFlow_StepFlag(t *testing.T) {
	t.Parallel()

	// 1. Valid step
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"flow", "init", "--step", "2"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("--step 2 failed: %v", err)
		}

		var s knowledge.FlowStep
		if err := json.Unmarshal(stdout.Bytes(), &s); err != nil {
			t.Fatalf("failed to decode step JSON: %v", err)
		}
		if s.Index != 2 || s.Actor != "scout" {
			t.Errorf("unexpected step data: %+v", s)
		}
	}

	// 2. Existing reviewer step remains at step 4
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"flow", "init", "--step", "4"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("--step 4 failed: %v", err)
		}

		var s knowledge.FlowStep
		if err := json.Unmarshal(stdout.Bytes(), &s); err != nil {
			t.Fatalf("failed to decode step JSON: %v", err)
		}
		if s.Index != 4 || s.Actor != "reviewer" {
			t.Errorf("unexpected step data: %+v", s)
		}
	}

	// 3. e2e flow steps: step 1, 5, 6, 8, 10, 11, 12
	for _, tc := range []struct {
		step     string
		actor    knowledge.Role
		terminal bool
	}{
		{"1", knowledge.RolePlanner, false},
		{"5", knowledge.RoleVerifier, false},
		{"6", knowledge.RolePlanner, false},
		{"8", knowledge.RoleReviewer, false},
		{"10", knowledge.RolePlanner, false},
		{"11", knowledge.RolePlanner, true},
		{"12", knowledge.RolePlanner, true},
	} {
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"flow", "e2e", "--step", tc.step}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("e2e --step %s failed: %v", tc.step, err)
		}
		var s knowledge.FlowStep
		if err := json.Unmarshal(stdout.Bytes(), &s); err != nil {
			t.Fatalf("failed to decode step %s JSON: %v", tc.step, err)
		}
		if s.Actor != tc.actor {
			t.Errorf("step %s actor mismatch: expected %s, got %s", tc.step, tc.actor, s.Actor)
		}
		if s.Terminal != tc.terminal {
			t.Errorf("step %s terminal mismatch: expected %v, got %v", tc.step, tc.terminal, s.Terminal)
		}
	}

	// 4. Out of range step
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"flow", "init", "--step", "999"}, &stdout, &stderr, "dev")
		if err == nil {
			t.Error("expected error for non-existent step, got nil")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 5. Step flag without flow name
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"flow", "--step", "1"}, &stdout, &stderr, "dev")
		if err == nil {
			t.Error("expected error when --step used without flow name")
		}
	}
}

func TestFlow_Unknown(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"flow", "nonexistent"}, &stdout, &stderr, "dev")
	if err == nil {
		t.Error("expected error for unknown flow, got nil")
	}
	if !strings.Contains(err.Error(), "unknown flow") {
		t.Errorf("unexpected error message: %v", err)
	}
}
