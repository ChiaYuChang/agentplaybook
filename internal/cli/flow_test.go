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
	for _, expected := range []string{"init", "plan", "blueprint", "build", "review", "commit", "session-handoff"} {
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

	// 4. review flow (severities, Track A/B, and review-resolution synthesis)
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
		if reviewStep2Conditions["BASELINE_STALE"] != 8 || reviewStep2Conditions["REVIEW_PASS"] != 6 || reviewStep2Conditions["FINDINGS_REPORTED"] != 3 {
			t.Errorf("expected review step 2 conditions [BASELINE_STALE: 8, REVIEW_PASS: 6, FINDINGS_REPORTED: 3], got: %v", reviewStep2Conditions)
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
		if !strings.Contains(f.Steps[5].Action, "review-resolution") || !strings.Contains(f.Steps[5].Action, "resolved or arbitrated") || !strings.Contains(f.Steps[5].Action, "resolved or waived") {
			t.Errorf("expected review step 6 to confirm resolved/arbitrated Blockers, resolved/waived Majors, and review-resolution synthesis, got: %s", f.Steps[5].Action)
		}
		if !strings.Contains(f.Steps[6].Action, "arbitrate Blocker") || !strings.Contains(f.Steps[6].Action, "record waiver with rationale for Major") {
			t.Errorf("expected review step 7 to handle Blocker arbitration and Major waivers, got: %s", f.Steps[6].Action)
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

	// 2. Out of range step
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

	// 3. Step flag without flow name
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
