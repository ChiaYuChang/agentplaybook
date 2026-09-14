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
	for _, expected := range []string{"init", "plan", "blueprint", "build", "review", "commit", "cartography", "session-handoff", "e2e", "navigator-cartography"} {
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

	// 2. plan flow (independent concept/method and Interaction Contract gates)
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
		if len(f.Steps) != 8 {
			t.Fatalf("expected 8 steps in plan flow, got %d", len(f.Steps))
		}
		if !strings.Contains(f.Steps[0].Action, "Hierarchical Blueprint") {
			t.Errorf("expected plan step 1 action to mention Hierarchical Blueprint, got: %s", f.Steps[0].Action)
		}
		if !strings.Contains(f.Steps[2].Action, "Counterfactual Decomposition Challenge") {
			t.Errorf("expected plan step 3 action to mention Counterfactual Decomposition Challenge, got: %s", f.Steps[2].Action)
		}
		phase1Conditions := make(map[string]int)
		for _, condition := range f.Steps[2].Conditions {
			phase1Conditions[condition.When] = condition.Then
		}
		if phase1Conditions["PLAN_REVIEW_PASS"] != 5 || phase1Conditions["PLAN_REVIEW_REJECT"] != 4 {
			t.Errorf("unexpected phase-1 plan gate transitions: %v", phase1Conditions)
		}
		if f.Steps[4].Actor != knowledge.RolePlanner || !strings.Contains(f.Steps[4].Action, "scenario-driven interaction boundaries") {
			t.Errorf("expected step 5 to draft interaction contracts after concept pass, got: %+v", f.Steps[4])
		}
		contractConditions := make(map[string]int)
		for _, condition := range f.Steps[5].Conditions {
			contractConditions[condition.When] = condition.Then
		}
		if f.Steps[5].Actor != knowledge.RoleReviewer || contractConditions["CONTRACT_REVIEW_PASS"] != 8 || contractConditions["CONTRACT_REVIEW_REJECT"] != 7 {
			t.Errorf("unexpected contract review gate: %+v, transitions=%v", f.Steps[5], contractConditions)
		}
		revisionConditions := make(map[string]int)
		for _, condition := range f.Steps[6].Conditions {
			revisionConditions[condition.When] = condition.Then
		}
		if revisionConditions["CONTRACT_ONLY_REVISION"] != 6 || revisionConditions["FOUNDATION_OR_SCOPE_CHANGED"] != 2 {
			t.Errorf("unexpected contract revision transitions: %v", revisionConditions)
		}
		if !strings.Contains(f.Steps[7].Action, "PLAN_REVIEW_PASS and CONTRACT_REVIEW_PASS") {
			t.Errorf("expected finalization to require both approvals, got: %s", f.Steps[7].Action)
		}
	}

	// 3. blueprint flow (17 steps; shared and JIT contract gates before Builder)
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
		if len(f.Steps) != 17 {
			t.Fatalf("expected 17 steps in blueprint flow, got %d", len(f.Steps))
		}
		if f.Steps[0].Actor != knowledge.RolePlanner || f.Steps[1].Actor != knowledge.RoleReviewer {
			t.Errorf("unexpected actors in blueprint steps 1-2: %s, %s", f.Steps[0].Actor, f.Steps[1].Actor)
		}
		blueprintGateConditions := make(map[string]int)
		for _, condition := range f.Steps[1].Conditions {
			blueprintGateConditions[condition.When] = condition.Then
		}
		if blueprintGateConditions["BLUEPRINT_PASS"] != 3 || blueprintGateConditions["BLUEPRINT_REJECT"] != 1 {
			t.Errorf("unexpected Blueprint Gate transitions: %v", blueprintGateConditions)
		}
		if f.Steps[2].Actor != knowledge.RoleReviewer || !strings.Contains(f.Steps[2].Action, "blueprint-level shared Interaction Contracts") {
			t.Errorf("expected step 3 to separately review shared blueprint contracts after Blueprint Gate, got: %+v", f.Steps[2])
		}
		sharedConditions := make(map[string]int)
		for _, condition := range f.Steps[2].Conditions {
			sharedConditions[condition.When] = condition.Then
		}
		if sharedConditions["SHARED_CONTRACT_REVIEW_PASS"] != 5 || sharedConditions["SHARED_CONTRACT_REVIEW_REJECT"] != 4 {
			t.Errorf("unexpected shared contract gate transitions: %v", sharedConditions)
		}
		sharedRevisionConditions := make(map[string]int)
		for _, condition := range f.Steps[3].Conditions {
			sharedRevisionConditions[condition.When] = condition.Then
		}
		if sharedRevisionConditions["SHARED_CONTRACT_AMENDED"] != 2 || sharedRevisionConditions["BLUEPRINT_CONCEPT_CHANGED"] != 2 {
			t.Errorf("shared contract amendments must re-enter Blueprint Gate before contract review: %v", sharedRevisionConditions)
		}
		if !strings.Contains(f.Steps[3].Action, "invalidates prior approvals for dependent sub-plans") {
			t.Errorf("shared contract amendment must invalidate dependent approvals, got: %s", f.Steps[3].Action)
		}
		if !strings.Contains(f.Steps[4].Action, "Reassess every dependent sub-plan through Sub-Plan and contract gates") {
			t.Errorf("shared contract approval must restart dependent JIT assessments, got: %s", f.Steps[4].Action)
		}
		if f.Steps[5].Actor != knowledge.RoleReviewer || !strings.Contains(f.Steps[5].Action, "Sub-Plan Gate") {
			t.Errorf("expected step 6 to remain the JIT Sub-Plan Gate, got: %+v", f.Steps[5])
		}
		if f.Steps[6].Actor != knowledge.RolePlanner || f.Steps[7].Actor != knowledge.RoleReviewer || f.Steps[9].Actor != knowledge.RoleBuilder {
			t.Errorf("expected JIT contract design/review before Builder handoff, got steps 7-10: %+v", f.Steps[6:10])
		}
		contractConditions := make(map[string]int)
		for _, condition := range f.Steps[7].Conditions {
			contractConditions[condition.When] = condition.Then
		}
		if contractConditions["CONTRACT_REVIEW_PASS"] != 10 || contractConditions["CONTRACT_REVIEW_REJECT"] != 9 {
			t.Errorf("unexpected JIT contract gate transitions: %v", contractConditions)
		}
		if !strings.Contains(f.Steps[9].Action, "SHARED_CONTRACT_REVIEW_PASS") {
			t.Errorf("Builder handoff must require shared contract approval, got: %s", f.Steps[9].Action)
		}
		revisionConditions := make(map[string]int)
		for _, condition := range f.Steps[8].Conditions {
			revisionConditions[condition.When] = condition.Then
		}
		if revisionConditions["CONTRACT_ONLY_REVISION"] != 8 || revisionConditions["SUBPLAN_CONCEPT_CHANGED"] != 6 || revisionConditions["SHARED_CONTRACT_CHANGED"] != 2 {
			t.Errorf("unexpected contract revision transitions: %v", revisionConditions)
		}
		if !strings.Contains(f.Steps[10].Action, "When Track B is selected, assert the pinned baseline identity") {
			t.Errorf("expected blueprint step 11 action to retain Track B pinned baseline identity assertion, got: %s", f.Steps[10].Action)
		}
		bpStep11Conditions := make(map[string]int)
		for _, c := range f.Steps[10].Conditions {
			bpStep11Conditions[c.When] = c.Then
		}
		if bpStep11Conditions["BASELINE_STALE"] != 2 || bpStep11Conditions["SUBPLAN_REVIEW_FINDINGS"] != 12 || bpStep11Conditions["SUBPLAN_REVIEW_SATISFIED"] != 13 {
			t.Errorf("expected blueprint step 11 code-review transitions to remain intact, got: %v", bpStep11Conditions)
		}
		remediationConditions := make(map[string]int)
		for _, condition := range f.Steps[11].Conditions {
			remediationConditions[condition.When] = condition.Then
		}
		if remediationConditions["REMEDIATION_DISPATCHED"] != 10 || remediationConditions["CONTRACT_ONLY_REVISION"] != 7 ||
			remediationConditions["SUBPLAN_CONCEPT_CHANGED"] != 6 || remediationConditions["SHARED_CONTRACT_CHANGED"] != 2 {
			t.Errorf("unexpected remediation routing for contract changes: %v", remediationConditions)
		}
		if !strings.Contains(f.Steps[11].Action, "shared-contract amendments invalidate all dependent sub-plan approvals") {
			t.Errorf("code-remediation shared-contract change must invalidate dependent approvals, got: %s", f.Steps[11].Action)
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

	// 10. navigator-cartography flow
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"flow", "navigator-cartography"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("querying flow navigator-cartography failed: %v", err)
		}

		var f knowledge.Flow
		if err := json.Unmarshal(stdout.Bytes(), &f); err != nil {
			t.Fatalf("failed to decode JSON response: %v\nRaw: %s", err, stdout.String())
		}

		if f.Name != "navigator-cartography" {
			t.Errorf("expected flow name 'navigator-cartography', got %q", f.Name)
		}
		if len(f.Steps) != 6 {
			t.Fatalf("expected 6 steps in navigator-cartography, got %d", len(f.Steps))
		}
		if f.Steps[0].Actor != knowledge.RoleNavigator || f.Steps[1].Actor != knowledge.RoleNavigator {
			t.Errorf("expected steps 1 and 2 actor navigator, got %s and %s", f.Steps[0].Actor, f.Steps[1].Actor)
		}
		if f.Steps[2].Actor != knowledge.RoleCartographer || f.Steps[3].Actor != knowledge.RoleCartographer || f.Steps[4].Actor != knowledge.RoleCartographer || f.Steps[5].Actor != knowledge.RoleCartographer {
			t.Errorf("expected steps 3-6 actor cartographer")
		}

		// Step 3 Taste Gate conditions
		step3Conditions := make(map[string]int)
		for _, c := range f.Steps[2].Conditions {
			step3Conditions[c.When] = c.Then
		}
		if step3Conditions["CLARIFICATION_REQUIRED"] != 4 || step3Conditions["DIAGRAM_APPROVED"] != 5 || step3Conditions["ADVISORY_ISSUED"] != 6 {
			t.Errorf("unexpected step 3 conditions: %v", step3Conditions)
		}

		// Step 4 Clarification Inquiry condition
		step4Conditions := make(map[string]int)
		for _, c := range f.Steps[3].Conditions {
			step4Conditions[c.When] = c.Then
		}
		if step4Conditions["CLARIFICATION_RESOLVED"] != 3 {
			t.Errorf("unexpected step 4 conditions: %v", step4Conditions)
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

	// 4. navigator-cartography flow steps: step 1, 2, 3, 4, 5, 6
	for _, tc := range []struct {
		step     string
		actor    knowledge.Role
		terminal bool
	}{
		{"1", knowledge.RoleNavigator, false},
		{"2", knowledge.RoleNavigator, false},
		{"3", knowledge.RoleCartographer, false},
		{"4", knowledge.RoleCartographer, false},
		{"5", knowledge.RoleCartographer, false},
		{"6", knowledge.RoleCartographer, false},
	} {
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"flow", "navigator-cartography", "--step", tc.step}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("navigator-cartography --step %s failed: %v", tc.step, err)
		}
		var s knowledge.FlowStep
		if err := json.Unmarshal(stdout.Bytes(), &s); err != nil {
			t.Fatalf("failed to decode step %s JSON: %v", tc.step, err)
		}
		if s.Actor != tc.actor {
			t.Errorf("navigator-cartography step %s actor mismatch: expected %s, got %s", tc.step, tc.actor, s.Actor)
		}
		if s.Terminal != tc.terminal {
			t.Errorf("navigator-cartography step %s terminal mismatch: expected %v, got %v", tc.step, tc.terminal, s.Terminal)
		}
	}

	// 5. Out of range step
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

	// 6. Step flag without flow name
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
