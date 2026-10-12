package knowledge_test

import (
	"strings"
	"testing"

	"github.com/ChiaYuChang/agentplaybook/internal/knowledge"
)

func TestLoad_Success(t *testing.T) {
	t.Parallel()

	k, err := knowledge.Load()
	if err != nil {
		t.Fatalf("Load() failed unexpectedly: %v", err)
	}

	// 1. Verify Config
	if len(k.Config.Languages) != 2 {
		t.Errorf("expected 2 languages, got %v", k.Config.Languages)
	}
	if k.Config.Transport != "herdr" {
		t.Errorf("expected transport 'herdr', got %q", k.Config.Transport)
	}

	// 2. Verify Roles
	roles := k.Roles()
	if len(roles) != 8 {
		t.Fatalf("expected 8 roles, got %d", len(roles))
	}
	for _, expected := range []string{"planner", "builder", "reviewer", "scout", "navigator", "cartographer", "verifier", "external-reviewer"} {
		r, ok := k.Role(expected)
		if !ok {
			t.Errorf("expected role %q to exist", expected)
		}
		if string(r.Name) != expected {
			t.Errorf("expected role name %q, got %q", expected, r.Name)
		}
	}

	// Verify navigator role metadata
	nav, ok := k.Role("navigator")
	if !ok {
		t.Fatalf("expected navigator role to exist")
	}
	if nav.Category != "companion" {
		t.Errorf("expected navigator category 'companion', got %q", nav.Category)
	}
	if len(nav.Communication.Targets) != 3 || nav.Communication.Targets[0] != knowledge.RoleUser || nav.Communication.Targets[1] != knowledge.RolePlanner || nav.Communication.Targets[2] != knowledge.RoleCartographer {
		t.Errorf("expected navigator communication targets [user planner cartographer], got %v", nav.Communication.Targets)
	}

	// Verify cartographer role metadata
	cart, ok := k.Role("cartographer")
	if !ok {
		t.Fatalf("expected cartographer role to exist")
	}
	if cart.Category != "companion" {
		t.Errorf("expected cartographer category 'companion', got %q", cart.Category)
	}
	if len(cart.Communication.Targets) != 3 || cart.Communication.Targets[0] != knowledge.RoleUser || cart.Communication.Targets[1] != knowledge.RolePlanner || cart.Communication.Targets[2] != knowledge.RoleNavigator {
		t.Errorf("expected cartographer communication targets [user planner navigator], got %v", cart.Communication.Targets)
	}

	// Verify verifier role metadata
	ver, ok := k.Role("verifier")
	if !ok {
		t.Fatalf("expected verifier role to exist")
	}
	if ver.Category != "core" {
		t.Errorf("expected verifier category 'core', got %q", ver.Category)
	}
	if len(ver.Communication.Targets) != 1 || ver.Communication.Targets[0] != knowledge.RolePlanner {
		t.Errorf("expected verifier communication targets [planner], got %v", ver.Communication.Targets)
	}

	// Verify planner communication targets include navigator, cartographer, and verifier
	planner, ok := k.Role("planner")
	if !ok {
		t.Fatalf("expected planner role to exist")
	}
	if planner.Category != "core" {
		t.Errorf("expected planner category 'core', got %q", planner.Category)
	}
	foundNavTarget, foundCartTarget, foundVerTarget := false, false, false
	for _, target := range planner.Communication.Targets {
		if target == knowledge.RoleNavigator {
			foundNavTarget = true
		}
		if target == knowledge.RoleCartographer {
			foundCartTarget = true
		}
		if target == knowledge.RoleVerifier {
			foundVerTarget = true
		}
	}
	if !foundNavTarget {
		t.Errorf("expected planner communication targets to include navigator, got %v", planner.Communication.Targets)
	}
	if !foundCartTarget {
		t.Errorf("expected planner communication targets to include cartographer, got %v", planner.Communication.Targets)
	}
	if !foundVerTarget {
		t.Errorf("expected planner communication targets to include verifier, got %v", planner.Communication.Targets)
	}

	// 3. Verify Flows
	flows := k.Flows()
	if len(flows) != 10 {
		t.Fatalf("expected 10 flows, got %d", len(flows))
	}
	for _, expected := range []string{"init", "plan", "blueprint", "build", "review", "commit", "session-handoff", "cartography", "e2e", "navigator-cartography"} {
		f, ok := k.Flow(expected)
		if !ok {
			t.Errorf("expected flow %q to exist", expected)
		}
		if f.Name != expected {
			t.Errorf("expected flow name %q, got %q", expected, f.Name)
		}
	}

	// Test FlowStep query
	step2, ok := k.FlowStep("init", 2)
	if !ok {
		t.Errorf("expected init step 2 to exist")
	}
	if step2.Index != 2 || step2.Actor != knowledge.RoleScout {
		t.Errorf("unexpected step 2 data: %+v", step2)
	}
	initFlow, _ := k.Flow("init")
	if len(initFlow.Steps) != 9 {
		t.Errorf("expected 9 steps in init flow, got %d", len(initFlow.Steps))
	}
	step1, _ := k.FlowStep("init", 1)
	conditions := make(map[string]int)
	for _, c := range step1.Conditions {
		conditions[c.When] = c.Then
	}
	if conditions["DIRECT_SURVEY"] != 3 {
		t.Errorf("expected DIRECT_SURVEY to point to step 3, got %d", conditions["DIRECT_SURVEY"])
	}
	if conditions["SCOUT_RECON_REQUIRED"] != 2 {
		t.Errorf("expected SCOUT_RECON_REQUIRED to point to step 2, got %d", conditions["SCOUT_RECON_REQUIRED"])
	}
	step4, _ := k.FlowStep("init", 4)
	if step4.Actor != knowledge.RoleReviewer {
		t.Errorf("expected init step 4 actor reviewer, got %q", step4.Actor)
	}
	step8, _ := k.FlowStep("init", 8)
	consensusConditions := make(map[string]int)
	for _, c := range step8.Conditions {
		consensusConditions[c.When] = c.Then
	}
	if consensusConditions["QUESTIONS_RAISED"] != 4 {
		t.Errorf("expected QUESTIONS_RAISED to point to step 4, got %d", consensusConditions["QUESTIONS_RAISED"])
	}
	if consensusConditions["NO_QUESTIONS_RAISED"] != 9 {
		t.Errorf("expected NO_QUESTIONS_RAISED to point to step 9, got %d", consensusConditions["NO_QUESTIONS_RAISED"])
	}
	if _, ok := k.FlowStep("init", 999); ok {
		t.Errorf("expected non-existent step to return false")
	}
	if _, ok := k.FlowStep("unknown-flow", 1); ok {
		t.Errorf("expected non-existent flow to return false")
	}

	// Verify single-approval plan flow and deterministic gate routing.
	planFlow, ok := k.Flow("plan")
	if !ok {
		t.Fatalf("expected plan flow to exist")
	}
	if len(planFlow.Steps) != 7 {
		t.Fatalf("expected 7 steps in plan flow, got %d", len(planFlow.Steps))
	}
	planStep5, _ := k.FlowStep("plan", 5)
	planStep5Conditions := make(map[string]int)
	for _, c := range planStep5.Conditions {
		planStep5Conditions[c.When] = c.Then
	}
	if planStep5Conditions["PLAN_REVIEW_PASS"] != 7 || planStep5Conditions["PLAN_REVIEW_REJECT"] != 6 {
		t.Errorf("unexpected plan gate conditions: %v", planStep5Conditions)
	}
	planStep6, _ := k.FlowStep("plan", 6)
	planStep6Conditions := make(map[string]int)
	for _, c := range planStep6.Conditions {
		planStep6Conditions[c.When] = c.Then
	}
	if planStep6Conditions["PLANS_REVISED"] != 4 {
		t.Errorf("unexpected plan revision routing: %v", planStep6Conditions)
	}

	// Verify direct build entry documents internal PLAN_REVIEW_PASS.
	buildFlow, ok := k.Flow("build")
	if !ok {
		t.Fatalf("expected build flow to exist")
	}
	if len(buildFlow.Steps) == 0 || !strings.Contains(buildFlow.Steps[0].Action, "PLAN_REVIEW_PASS") {
		t.Errorf("expected build entry to require PLAN_REVIEW_PASS, got: %+v", buildFlow)
	}

	// Verify blueprint concept/shared-contract approvals and JIT gate ordering.
	blueprintFlow, ok := k.Flow("blueprint")
	if !ok {
		t.Fatalf("expected blueprint flow to exist")
	}
	if len(blueprintFlow.Steps) != 17 {
		t.Fatalf("expected 17 steps in blueprint flow, got %d", len(blueprintFlow.Steps))
	}
	bpStep2, _ := k.FlowStep("blueprint", 2)
	bpStep2Conditions := make(map[string]int)
	for _, c := range bpStep2.Conditions {
		bpStep2Conditions[c.When] = c.Then
	}
	if bpStep2Conditions["BLUEPRINT_PASS"] != 3 || bpStep2Conditions["BLUEPRINT_REJECT"] != 1 {
		t.Errorf("unexpected blueprint step 2 conditions: %v", bpStep2Conditions)
	}
	bpStep3, _ := k.FlowStep("blueprint", 3)
	bpStep3Conditions := make(map[string]int)
	for _, c := range bpStep3.Conditions {
		bpStep3Conditions[c.When] = c.Then
	}
	if bpStep3.Actor != knowledge.RoleReviewer || bpStep3Conditions["SHARED_CONTRACT_REVIEW_PASS"] != 5 || bpStep3Conditions["SHARED_CONTRACT_REVIEW_REJECT"] != 4 {
		t.Errorf("unexpected post-Blueprint-Gate shared contract review: %+v conditions=%v", bpStep3, bpStep3Conditions)
	}
	bpStep4, _ := k.FlowStep("blueprint", 4)
	bpStep4Conditions := make(map[string]int)
	for _, c := range bpStep4.Conditions {
		bpStep4Conditions[c.When] = c.Then
	}
	if bpStep4Conditions["SHARED_CONTRACT_AMENDED"] != 2 || bpStep4Conditions["BLUEPRINT_CONCEPT_CHANGED"] != 2 {
		t.Errorf("every shared contract amendment must re-enter Blueprint Gate: %v", bpStep4Conditions)
	}
	bpStep5, _ := k.FlowStep("blueprint", 5)
	if !strings.Contains(bpStep4.Action, "invalidates prior approvals for dependent sub-plans") ||
		!strings.Contains(bpStep5.Action, "Reassess every dependent sub-plan through Sub-Plan and contract gates") {
		t.Errorf("shared-contract amendment must invalidate and reassess dependent sub-plans: step4=%q step5=%q", bpStep4.Action, bpStep5.Action)
	}
	bpStep6, _ := k.FlowStep("blueprint", 6)
	bpStep6Conditions := make(map[string]int)
	for _, c := range bpStep6.Conditions {
		bpStep6Conditions[c.When] = c.Then
	}
	if bpStep6.Actor != knowledge.RoleReviewer || !strings.Contains(bpStep6.Action, "Sub-Plan Gate") || bpStep6Conditions["SUBPLAN_PLAN_PASS"] != 7 || bpStep6Conditions["SUBPLAN_PLAN_REJECT"] != 5 {
		t.Errorf("unexpected JIT Sub-Plan Gate: %+v conditions=%v", bpStep6, bpStep6Conditions)
	}
	bpStep8, _ := k.FlowStep("blueprint", 8)
	bpStep8Conditions := make(map[string]int)
	for _, c := range bpStep8.Conditions {
		bpStep8Conditions[c.When] = c.Then
	}
	if bpStep8Conditions["CONTRACT_REVIEW_PASS"] != 10 || bpStep8Conditions["CONTRACT_REVIEW_REJECT"] != 9 {
		t.Errorf("unexpected JIT contract review conditions: %v", bpStep8Conditions)
	}
	bpStep9, _ := k.FlowStep("blueprint", 9)
	bpStep9Conditions := make(map[string]int)
	for _, c := range bpStep9.Conditions {
		bpStep9Conditions[c.When] = c.Then
	}
	if bpStep9Conditions["CONTRACT_ONLY_REVISION"] != 8 || bpStep9Conditions["SUBPLAN_CONCEPT_CHANGED"] != 6 || bpStep9Conditions["SHARED_CONTRACT_CHANGED"] != 2 {
		t.Errorf("unexpected JIT contract revision routing: %v", bpStep9Conditions)
	}
	bpStep10, _ := k.FlowStep("blueprint", 10)
	if bpStep10.Actor != knowledge.RoleBuilder || !strings.Contains(bpStep10.Action, "BLUEPRINT_PASS, SHARED_CONTRACT_REVIEW_PASS, SUBPLAN_PLAN_PASS, and CONTRACT_REVIEW_PASS") {
		t.Errorf("expected Builder handoff only after all four approvals, got: %+v", bpStep10)
	}
	bpStep11, _ := k.FlowStep("blueprint", 11)
	bpStep11Conditions := make(map[string]int)
	for _, c := range bpStep11.Conditions {
		bpStep11Conditions[c.When] = c.Then
	}
	if bpStep11Conditions["SUBPLAN_REVIEW_FINDINGS"] != 12 || bpStep11Conditions["SUBPLAN_REVIEW_SATISFIED"] != 13 || bpStep11Conditions["BLUEPRINT_REVIEW_REQUIRED"] != 2 || bpStep11Conditions["BASELINE_STALE"] != 2 {
		t.Errorf("unexpected blueprint step 11 code-review conditions: %v", bpStep11Conditions)
	}
	bpStep12, _ := k.FlowStep("blueprint", 12)
	bpStep12Conditions := make(map[string]int)
	for _, c := range bpStep12.Conditions {
		bpStep12Conditions[c.When] = c.Then
	}
	if bpStep12Conditions["REMEDIATION_DISPATCHED"] != 10 || bpStep12Conditions["CONTRACT_ONLY_REVISION"] != 7 ||
		bpStep12Conditions["SUBPLAN_CONCEPT_CHANGED"] != 6 || bpStep12Conditions["SHARED_CONTRACT_CHANGED"] != 2 {
		t.Errorf("unexpected blueprint remediation routing: %v", bpStep12Conditions)
	}
	if !strings.Contains(bpStep12.Action, "shared-contract amendments invalidate all dependent sub-plan approvals") {
		t.Errorf("shared-contract code remediation must invalidate dependent approvals: %q", bpStep12.Action)
	}
	bpStep14, _ := k.FlowStep("blueprint", 14)
	bpStep14Conditions := make(map[string]int)
	for _, c := range bpStep14.Conditions {
		bpStep14Conditions[c.When] = c.Then
	}
	if bpStep14Conditions["SUBPLAN_REVIEW_PASS_MORE_SUBPLANS"] != 5 || bpStep14Conditions["SUBPLAN_REVIEW_PASS_ALL_COMPLETED"] != 15 || bpStep14Conditions["RESOLUTION_REJECTED"] != 13 {
		t.Errorf("unexpected blueprint step 14 conditions: %v", bpStep14Conditions)
	}
	bpStep16, _ := k.FlowStep("blueprint", 16)
	bpStep16Conditions := make(map[string]int)
	for _, c := range bpStep16.Conditions {
		bpStep16Conditions[c.When] = c.Then
	}
	if bpStep16Conditions["FEATURE_REVIEW_PASS"] != 17 || bpStep16Conditions["FEATURE_REVIEW_REJECT"] != 5 || bpStep16Conditions["DEPENDENT_EVIDENCE_STALE"] != 2 {
		t.Errorf("unexpected blueprint step 16 conditions: %v", bpStep16Conditions)
	}

	// Regression check: review flow has 8 steps, step 1 is planner handoff, and step 2 has BASELINE_STALE condition
	reviewFlow, _ := k.Flow("review")
	if len(reviewFlow.Steps) != 8 {
		t.Errorf("expected 8 steps in review flow, got %d", len(reviewFlow.Steps))
	}
	reviewStep1, _ := k.FlowStep("review", 1)
	if reviewStep1.Actor != knowledge.RolePlanner || len(reviewStep1.Conditions) != 0 {
		t.Errorf("expected review step 1 to be planner handoff with 0 conditions, got actor %q, conditions: %v", reviewStep1.Actor, reviewStep1.Conditions)
	}
	reviewStep2, _ := k.FlowStep("review", 2)
	if reviewStep2.Actor != knowledge.RoleReviewer {
		t.Errorf("expected review step 2 to be reviewer, got %q", reviewStep2.Actor)
	}
	reviewStep2Conditions := make(map[string]int)
	for _, c := range reviewStep2.Conditions {
		reviewStep2Conditions[c.When] = c.Then
	}
	if reviewStep2Conditions["BASELINE_STALE"] != 8 || reviewStep2Conditions["REVIEW_PASS"] != 6 || reviewStep2Conditions["FINDINGS_REPORTED"] != 3 {
		t.Errorf("expected review step 2 conditions [BASELINE_STALE: 8, REVIEW_PASS: 6, FINDINGS_REPORTED: 3], got: %v", reviewStep2Conditions)
	}
	step7, _ := k.FlowStep("review", 7)
	foundPlanUpdate := false
	for _, c := range step7.Conditions {
		if c.When == "PLAN_UPDATE_REQUIRED" {
			foundPlanUpdate = true
			if c.Then != 8 {
				t.Errorf("expected PLAN_UPDATE_REQUIRED to point to step 8, got %d", c.Then)
			}
		}
	}
	if !foundPlanUpdate {
		t.Errorf("expected PLAN_UPDATE_REQUIRED condition on review step 7")
	}

	// Regression check: commit flow has 9 steps
	commitFlow, ok := k.Flow("commit")
	if !ok {
		t.Fatalf("expected commit flow to exist")
	}
	if len(commitFlow.Steps) != 9 {
		t.Errorf("expected 9 steps in commit flow, got %d", len(commitFlow.Steps))
	}

	// Regression check: session-handoff flow has 8 steps
	handoffFlow, ok := k.Flow("session-handoff")
	if !ok {
		t.Fatalf("expected session-handoff flow to exist")
	}
	if len(handoffFlow.Steps) != 8 {
		t.Errorf("expected 8 steps in session-handoff flow, got %d", len(handoffFlow.Steps))
	}
	handoffStep1, _ := k.FlowStep("session-handoff", 1)
	handoffStep1Conditions := make(map[string]int)
	for _, c := range handoffStep1.Conditions {
		handoffStep1Conditions[c.When] = c.Then
	}
	if handoffStep1Conditions["ANCHOR_CAPTURED"] != 2 || handoffStep1Conditions["ANCHOR_INVALID"] != 1 {
		t.Errorf("unexpected session-handoff step 1 conditions: %v", handoffStep1Conditions)
	}

	// Verify cartography flow
	cartFlow, ok := k.Flow("cartography")
	if !ok {
		t.Fatalf("expected cartography flow to exist")
	}
	if len(cartFlow.Steps) != 6 {
		t.Fatalf("expected 6 steps in cartography flow, got %d", len(cartFlow.Steps))
	}
	if cartFlow.Steps[0].Actor != knowledge.RolePlanner || cartFlow.Steps[1].Actor != knowledge.RolePlanner {
		t.Errorf("expected cartography steps 1 and 2 actor to be planner")
	}
	if cartFlow.Steps[2].Actor != knowledge.RoleCartographer || cartFlow.Steps[3].Actor != knowledge.RoleCartographer || cartFlow.Steps[4].Actor != knowledge.RoleCartographer || cartFlow.Steps[5].Actor != knowledge.RoleCartographer {
		t.Errorf("expected cartography steps 3, 4, 5, 6 actor to be cartographer")
	}
	cartStep3Conditions := make(map[string]int)
	for _, c := range cartFlow.Steps[2].Conditions {
		cartStep3Conditions[c.When] = c.Then
	}
	if cartStep3Conditions["CLARIFICATION_REQUIRED"] != 4 || cartStep3Conditions["DIAGRAM_APPROVED"] != 5 || cartStep3Conditions["ADVISORY_ISSUED"] != 6 {
		t.Errorf("unexpected cartography step 3 conditions: %v", cartStep3Conditions)
	}
	cartStep4Conditions := make(map[string]int)
	for _, c := range cartFlow.Steps[3].Conditions {
		cartStep4Conditions[c.When] = c.Then
	}
	if cartStep4Conditions["CLARIFICATION_RESOLVED"] != 3 {
		t.Errorf("unexpected cartography step 4 conditions: %v", cartStep4Conditions)
	}

	// Verify navigator-cartography flow (6 steps)
	navCartFlow, ok := k.Flow("navigator-cartography")
	if !ok {
		t.Fatalf("expected navigator-cartography flow to exist")
	}
	if len(navCartFlow.Steps) != 6 {
		t.Fatalf("expected 6 steps in navigator-cartography flow, got %d", len(navCartFlow.Steps))
	}
	if navCartFlow.Steps[0].Actor != knowledge.RoleNavigator || navCartFlow.Steps[1].Actor != knowledge.RoleNavigator {
		t.Errorf("expected navigator-cartography steps 1 and 2 actor to be navigator")
	}
	if navCartFlow.Steps[2].Actor != knowledge.RoleCartographer || navCartFlow.Steps[3].Actor != knowledge.RoleCartographer || navCartFlow.Steps[4].Actor != knowledge.RoleCartographer || navCartFlow.Steps[5].Actor != knowledge.RoleCartographer {
		t.Errorf("expected navigator-cartography steps 3, 4, 5, 6 actor to be cartographer")
	}
	navCartStep3Conditions := make(map[string]int)
	for _, c := range navCartFlow.Steps[2].Conditions {
		navCartStep3Conditions[c.When] = c.Then
	}
	if navCartStep3Conditions["CLARIFICATION_REQUIRED"] != 4 || navCartStep3Conditions["DIAGRAM_APPROVED"] != 5 || navCartStep3Conditions["ADVISORY_ISSUED"] != 6 {
		t.Errorf("unexpected navigator-cartography step 3 conditions: %v", navCartStep3Conditions)
	}
	navCartStep4Conditions := make(map[string]int)
	for _, c := range navCartFlow.Steps[3].Conditions {
		navCartStep4Conditions[c.When] = c.Then
	}
	if navCartStep4Conditions["CLARIFICATION_RESOLVED"] != 3 {
		t.Errorf("unexpected navigator-cartography step 4 conditions: %v", navCartStep4Conditions)
	}

	// Verify e2e flow (8 steps)
	e2eFlow, ok := k.Flow("e2e")
	if !ok {
		t.Fatalf("expected e2e flow to exist")
	}
	if len(e2eFlow.Steps) != 8 {
		t.Fatalf("expected 8 steps in e2e flow, got %d", len(e2eFlow.Steps))
	}
	if e2eFlow.Steps[0].Actor != knowledge.RolePlanner {
		t.Errorf("expected e2e step 1 actor to be planner")
	}
	if e2eFlow.Steps[1].Actor != knowledge.RoleVerifier || e2eFlow.Steps[2].Actor != knowledge.RoleVerifier {
		t.Errorf("expected e2e steps 2-3 actor to be verifier")
	}
	if e2eFlow.Steps[3].Actor != knowledge.RolePlanner || e2eFlow.Steps[4].Actor != knowledge.RolePlanner {
		t.Errorf("expected e2e steps 4-5 actor to be planner")
	}
	if e2eFlow.Steps[5].Actor != knowledge.RoleBuilder {
		t.Errorf("expected e2e step 6 actor to be builder")
	}
	if e2eFlow.Steps[6].Actor != knowledge.RoleReviewer {
		t.Errorf("expected e2e step 7 actor to be reviewer")
	}
	if e2eFlow.Steps[7].Actor != knowledge.RolePlanner {
		t.Errorf("expected e2e step 8 actor to be planner")
	}
	if !e2eFlow.Steps[4].Terminal || len(e2eFlow.Steps[4].Conditions) != 0 {
		t.Errorf("expected e2e step 5 to be terminal: true with 0 conditions")
	}
	if !e2eFlow.Steps[7].Terminal || len(e2eFlow.Steps[7].Conditions) != 0 {
		t.Errorf("expected e2e step 8 to be terminal: true with 0 conditions")
	}

	// 4. Verify Artifacts
	artifacts := k.Artifacts()
	if len(artifacts) != 17 {
		t.Fatalf("expected 17 artifacts, got %d", len(artifacts))
	}
	for _, expected := range []string{
		"agents-md",
		"build-plan",
		"review-plan",
		"blueprint-plan",
		"sub-build-plan",
		"sub-review-plan",
		"sub-review-resolution",
		"review-findings",
		"review-resolution",
		"external-review-report",
		"scout-survey",
		"diagram-brief",
		"diagram-completion",
		"diagram-clarification-request",
		"e2e-brief",
		"e2e-report",
		"e2e-clarification-request",
	} {
		a, ok := k.Artifact(expected)
		if !ok {
			t.Errorf("expected artifact %q to exist", expected)
		}
		if a.Name != expected {
			t.Errorf("expected artifact name %q, got %q", expected, a.Name)
		}
	}

	// Regression check: plan artifact section counts and visibility
	buildPlan, _ := k.Artifact("build-plan")
	if len(buildPlan.Visibility) != 4 {
		t.Errorf("expected build-plan visibility to have 4 roles, got %v", buildPlan.Visibility)
	}
	if len(buildPlan.Sections) != 10 {
		t.Errorf("expected build-plan to have 10 sections, got %d", len(buildPlan.Sections))
	}
	subBuildPlan, _ := k.Artifact("sub-build-plan")
	if len(subBuildPlan.Sections) != 8 {
		t.Errorf("expected sub-build-plan to have 8 sections, got %d", len(subBuildPlan.Sections))
	}
	reviewPlan, _ := k.Artifact("review-plan")
	if len(reviewPlan.Sections) != 7 {
		t.Errorf("expected review-plan to have 7 sections, got %d", len(reviewPlan.Sections))
	}
	subReviewPlan, _ := k.Artifact("sub-review-plan")
	if len(subReviewPlan.Sections) != 7 {
		t.Errorf("expected sub-review-plan to have 7 sections, got %d", len(subReviewPlan.Sections))
	}
	reviewFindings, _ := k.Artifact("review-findings")
	if len(reviewFindings.Visibility) != 2 || reviewFindings.Visibility[0] != knowledge.RolePlanner || reviewFindings.Visibility[1] != knowledge.RoleReviewer {
		t.Errorf("expected review-findings visibility strictly [planner reviewer], got %v", reviewFindings.Visibility)
	}
	scoutSurvey, _ := k.Artifact("scout-survey")
	if scoutSurvey.Owner != knowledge.RoleScout || scoutSurvey.Type != "message" {
		t.Errorf("unexpected scout-survey metadata: %+v", scoutSurvey)
	}
	if len(scoutSurvey.Visibility) != 2 || scoutSurvey.Visibility[0] != knowledge.RolePlanner || scoutSurvey.Visibility[1] != knowledge.RoleScout {
		t.Errorf("expected scout-survey visibility [planner scout], got %v", scoutSurvey.Visibility)
	}
	if len(scoutSurvey.Fields) != 7 {
		t.Errorf("expected scout-survey to have 7 fields, got %d", len(scoutSurvey.Fields))
	}
	reviewResolution, _ := k.Artifact("review-resolution")
	if reviewResolution.Owner != knowledge.RolePlanner || reviewResolution.Type != "document" {
		t.Errorf("unexpected review-resolution metadata: %+v", reviewResolution)
	}
	if len(reviewResolution.Sections) != 5 {
		t.Errorf("expected review-resolution to have 5 sections, got %d", len(reviewResolution.Sections))
	}
	if len(reviewResolution.Visibility) != 7 {
		t.Errorf("expected review-resolution visibility to include 7 roles, got %v", reviewResolution.Visibility)
	}

	// Verify diagram-brief and diagram-completion metadata
	diagramBrief, _ := k.Artifact("diagram-brief")
	if diagramBrief.Owner != knowledge.RolePlanner || diagramBrief.Type != "message" {
		t.Errorf("unexpected diagram-brief metadata: %+v", diagramBrief)
	}
	if len(diagramBrief.Visibility) != 3 || diagramBrief.Visibility[0] != knowledge.RolePlanner || diagramBrief.Visibility[1] != knowledge.RoleNavigator || diagramBrief.Visibility[2] != knowledge.RoleCartographer {
		t.Errorf("expected diagram-brief visibility [planner navigator cartographer], got %v", diagramBrief.Visibility)
	}
	if len(diagramBrief.Fields) != 5 {
		t.Errorf("expected diagram-brief to have 5 fields, got %d", len(diagramBrief.Fields))
	}

	diagramCompletion, _ := k.Artifact("diagram-completion")
	if diagramCompletion.Owner != knowledge.RoleCartographer || diagramCompletion.Type != "message" {
		t.Errorf("unexpected diagram-completion metadata: %+v", diagramCompletion)
	}
	if len(diagramCompletion.Visibility) != 3 || diagramCompletion.Visibility[0] != knowledge.RolePlanner || diagramCompletion.Visibility[1] != knowledge.RoleNavigator || diagramCompletion.Visibility[2] != knowledge.RoleCartographer {
		t.Errorf("expected diagram-completion visibility [planner navigator cartographer], got %v", diagramCompletion.Visibility)
	}
	if len(diagramCompletion.Fields) != 3 {
		t.Errorf("expected diagram-completion to have 3 fields, got %d", len(diagramCompletion.Fields))
	}

	diagramClarification, _ := k.Artifact("diagram-clarification-request")
	if diagramClarification.Owner != knowledge.RoleCartographer || diagramClarification.Type != "message" {
		t.Errorf("unexpected diagram-clarification-request metadata: %+v", diagramClarification)
	}
	if len(diagramClarification.Visibility) != 3 || diagramClarification.Visibility[0] != knowledge.RolePlanner || diagramClarification.Visibility[1] != knowledge.RoleNavigator || diagramClarification.Visibility[2] != knowledge.RoleCartographer {
		t.Errorf("expected diagram-clarification-request visibility [planner navigator cartographer], got %v", diagramClarification.Visibility)
	}
	if len(diagramClarification.Fields) != 4 {
		t.Errorf("expected diagram-clarification-request to have 4 fields, got %d", len(diagramClarification.Fields))
	}

	// Verify e2e-brief and e2e-report metadata
	e2eBrief, _ := k.Artifact("e2e-brief")
	if e2eBrief.Owner != knowledge.RolePlanner || e2eBrief.Type != "message" {
		t.Errorf("unexpected e2e-brief metadata: %+v", e2eBrief)
	}
	if len(e2eBrief.Visibility) != 2 || e2eBrief.Visibility[0] != knowledge.RolePlanner || e2eBrief.Visibility[1] != knowledge.RoleVerifier {
		t.Errorf("expected e2e-brief visibility [planner verifier], got %v", e2eBrief.Visibility)
	}
	if len(e2eBrief.Fields) != 20 {
		t.Errorf("expected e2e-brief to have 20 fields, got %d", len(e2eBrief.Fields))
	}

	e2eReport, _ := k.Artifact("e2e-report")
	if e2eReport.Owner != knowledge.RoleVerifier || e2eReport.Type != "message" {
		t.Errorf("unexpected e2e-report metadata: %+v", e2eReport)
	}
	if len(e2eReport.Visibility) != 3 || e2eReport.Visibility[0] != knowledge.RolePlanner || e2eReport.Visibility[1] != knowledge.RoleVerifier || e2eReport.Visibility[2] != knowledge.RoleReviewer {
		t.Errorf("expected e2e-report visibility [planner verifier reviewer], got %v", e2eReport.Visibility)
	}
	if len(e2eReport.Fields) != 23 {
		t.Errorf("expected e2e-report to have 23 fields, got %d", len(e2eReport.Fields))
	}

	// Verify e2e-clarification-request metadata
	e2eClarification, _ := k.Artifact("e2e-clarification-request")
	if e2eClarification.Owner != knowledge.RoleVerifier || e2eClarification.Type != "message" {
		t.Errorf("unexpected e2e-clarification-request metadata: %+v", e2eClarification)
	}
	if len(e2eClarification.Visibility) != 2 || e2eClarification.Visibility[0] != knowledge.RolePlanner || e2eClarification.Visibility[1] != knowledge.RoleVerifier {
		t.Errorf("expected e2e-clarification-request visibility [planner verifier], got %v", e2eClarification.Visibility)
	}
	if len(e2eClarification.Fields) != 4 {
		t.Errorf("expected e2e-clarification-request to have 4 fields, got %d", len(e2eClarification.Fields))
	}

	// 5. Verify Rules
	rules := k.Rules()
	if len(rules) < 35 {
		t.Fatalf("expected at least 35 rules, got %d", len(rules))
	}
	for _, expected := range []string{
		"anti-cheating",
		"bounded-builder-autonomy",
		"user-requested-external-review",
		"tdd-reproduction",
		"coherent-plan-units",
		"anti-rubber-stamp-plan-gate",
		"evidence-proportional-persistence",
		"agents-md-single-writer",
		"acceptance-publication-authority",
		"role-context-lifecycle",
		"session-handoff-audit",
		"planner-reviewability",
		"review-severity-semantics",
		"track-b-action-differential-verification",
		"out-of-tree-baseline-mirror",
		"navigator-read-only-companion",
		"companion-query-zero-side-effect",
		"planner-source-restricted-response",
		"target-state-gated-inquiry",
		"cartographer-visual-architect-boundary",
		"cartography-zero-context-pollution",
		"cartography-taste-gate-advisory",
		"cartography-asynchronous-decoupling",
		"cartography-brief-self-sufficiency",
		"cartography-clarification-inquiry",
		"peer-session-transport-primacy",
		"e2e-sandbox-isolation",
		"e2e-zero-log-pollution",
		"e2e-package-self-sufficiency",
		"e2e-clarification-inquiry",
		"e2e-immutable-evidence-binding",
		"e2e-lifecycle-admission",
		"e2e-secrets-hygiene",
		"scaffolding-vault-isolation",
	} {
		r, ok := k.Rule(expected)
		if !ok {
			t.Errorf("expected rule %q to exist", expected)
		}
		if r.ID != expected {
			t.Errorf("expected rule ID %q, got %q", expected, r.ID)
		}
	}
}

func TestValidate_Errors(t *testing.T) {
	t.Parallel()

	k, err := knowledge.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Mutate config languages to empty
	invalid := *k
	invalid.Config.Languages = nil
	if err := knowledge.Validate(&invalid); err == nil {
		t.Error("expected error when languages is empty")
	}
}

func TestValidate_RoleCategoryAndUserConstraints(t *testing.T) {
	t.Parallel()

	k, err := knowledge.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// 1. Invalid role category
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		roles[0].Category = "invalid_category"
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error for invalid role category")
		}
	}

	// 2. User role registered as internal role
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		roles = append(roles, knowledge.RoleDefinition{
			Name:             knowledge.RoleUser,
			Title:            "User",
			Category:         "core",
			Description:      "Human user",
			Responsibilities: []string{"Direct project"},
			Boundaries:       []string{"Do not edit internals directly"},
		})
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when RoleUser is registered as an internal role")
		}
	}

	// 3. Builder targeting user directly
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RoleBuilder {
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RoleUser}
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Builder targets user directly")
		}
	}

	// 4. Builder targeting reviewer or multiple roles
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RoleBuilder {
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RolePlanner, knowledge.RoleReviewer}
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Builder targets roles other than strictly planner")
		}
	}

	// 5. Reviewer / Scout targeting unauthorized roles
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RoleScout {
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RoleNavigator}
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Scout targets Navigator")
		}
	}

	// 6. Planner omitting navigator target
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RolePlanner {
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RoleBuilder, knowledge.RoleReviewer, knowledge.RoleScout, knowledge.RoleCartographer}
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Planner omits navigator from communication targets")
		}
	}

	// 6b. Planner omitting cartographer target
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RolePlanner {
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RoleBuilder, knowledge.RoleReviewer, knowledge.RoleScout, knowledge.RoleNavigator}
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Planner omits cartographer from communication targets")
		}
	}

	// 7. Planner targeting external user directly
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RolePlanner {
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RoleBuilder, knowledge.RoleReviewer, knowledge.RoleScout, knowledge.RoleNavigator, knowledge.RoleCartographer, knowledge.RoleUser}
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Planner targets external user directly")
		}
	}

	// 8. Navigator omitting user, planner, or cartographer
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RoleNavigator {
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RolePlanner, knowledge.RoleCartographer}
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Navigator omits user from communication targets")
		}
	}

	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RoleNavigator {
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RoleUser, knowledge.RolePlanner}
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Navigator omits cartographer from communication targets")
		}
	}

	// 8b. Navigator targeting builder (violating star topology)
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RoleNavigator {
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RoleUser, knowledge.RolePlanner, knowledge.RoleCartographer, knowledge.RoleBuilder}
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Navigator targets Builder directly")
		}
	}

	// 9. Navigator duplicate communication targets
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RoleNavigator {
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RoleUser, knowledge.RolePlanner, knowledge.RoleCartographer, knowledge.RoleCartographer}
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Navigator has duplicate communication targets")
		}
	}

	// 10. Planner duplicate communication targets
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RolePlanner {
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RoleBuilder, knowledge.RoleReviewer, knowledge.RoleScout, knowledge.RoleNavigator, knowledge.RoleCartographer, knowledge.RoleCartographer}
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Planner has duplicate communication targets")
		}
	}

	// 11. Navigator as artifact owner
	{
		invalid := *k
		artifacts := append([]knowledge.Artifact(nil), k.Artifacts()...)
		artifacts = append(artifacts, knowledge.Artifact{
			Name:        "navigator-notes",
			Title:       "Navigator Notes",
			Description: "Notes by navigator",
			Owner:       knowledge.RoleNavigator,
			Visibility:  []knowledge.Role{knowledge.RolePlanner},
			Type:        "document",
			Sections:    []knowledge.ArtifactSection{{Name: "Notes", Required: true, Description: "Notes content"}},
		})
		setArtifacts(&invalid, artifacts)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Navigator is an artifact owner")
		}
	}

	// 12. Navigator in in-flight artifact visibility
	{
		invalid := *k
		artifacts := append([]knowledge.Artifact(nil), k.Artifacts()...)
		for i, a := range artifacts {
			if a.Name == "build-plan" {
				artifacts[i].Visibility = append(artifacts[i].Visibility, knowledge.RoleNavigator)
			}
		}
		setArtifacts(&invalid, artifacts)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Navigator is in build-plan visibility")
		}
	}

	// 13. Navigator in non-allowlisted artifact visibility
	{
		invalid := *k
		artifacts := append([]knowledge.Artifact(nil), k.Artifacts()...)
		artifacts = append(artifacts, knowledge.Artifact{
			Name:        "unsettled-notes",
			Title:       "Unsettled Notes",
			Description: "Custom non-allowlisted document",
			Owner:       knowledge.RolePlanner,
			Visibility:  []knowledge.Role{knowledge.RolePlanner, knowledge.RoleNavigator},
			Type:        "document",
			Sections:    []knowledge.ArtifactSection{{Name: "Content", Required: true, Description: "Content"}},
		})
		setArtifacts(&invalid, artifacts)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Navigator is in non-allowlisted artifact visibility")
		}
	}

	// 14. Navigator as flow actor
	{
		invalid := *k
		flows := append([]knowledge.Flow(nil), k.Flows()...)
		flows = append(flows, knowledge.Flow{
			Name:        "navigator-flow",
			Description: "Invalid flow with navigator actor",
			Steps: []knowledge.FlowStep{
				{
					Index:  1,
					Actor:  knowledge.RoleNavigator,
					Action: "Do something",
				},
			},
		})
		setFlows(&invalid, flows)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Navigator is a flow actor")
		}
	}

	// 15. User as artifact owner or flow actor
	{
		invalid := *k
		artifacts := append([]knowledge.Artifact(nil), k.Artifacts()...)
		artifacts = append(artifacts, knowledge.Artifact{
			Name:        "user-artifact",
			Title:       "User Artifact",
			Description: "Artifact owned by user",
			Owner:       knowledge.RoleUser,
			Visibility:  []knowledge.Role{knowledge.RolePlanner},
			Type:        "document",
			Sections:    []knowledge.ArtifactSection{{Name: "Content", Required: true, Description: "Content"}},
		})
		setArtifacts(&invalid, artifacts)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when User is an artifact owner")
		}
	}
}

func TestValidate_CartographerConstraints(t *testing.T) {
	t.Parallel()

	k, err := knowledge.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// 1. Cartographer omitting user, planner, or navigator
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RoleCartographer {
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RolePlanner, knowledge.RoleNavigator}
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Cartographer omits user from communication targets")
		}
	}

	// 2. Cartographer targeting builder directly (violating star topology)
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RoleCartographer {
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RoleUser, knowledge.RolePlanner, knowledge.RoleBuilder}
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Cartographer targets builder directly")
		}
	}

	// 3. Cartographer duplicate communication targets
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RoleCartographer {
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RoleUser, knowledge.RolePlanner, knowledge.RolePlanner}
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Cartographer has duplicate communication targets")
		}
	}

	// 4. Cartographer as artifact owner of non-diagram-completion artifact
	{
		invalid := *k
		artifacts := append([]knowledge.Artifact(nil), k.Artifacts()...)
		artifacts = append(artifacts, knowledge.Artifact{
			Name:        "cartographer-code-plan",
			Title:       "Cartographer Code Plan",
			Description: "Invalid artifact owned by cartographer",
			Owner:       knowledge.RoleCartographer,
			Visibility:  []knowledge.Role{knowledge.RolePlanner},
			Type:        "document",
			Sections:    []knowledge.ArtifactSection{{Name: "Plan", Required: true, Description: "Plan content"}},
		})
		setArtifacts(&invalid, artifacts)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Cartographer owns a non-diagram-completion artifact")
		}
	}

	// 5. Cartographer in in-flight artifact visibility (build-plan, review-findings, etc.)
	for _, inFlight := range []string{"build-plan", "review-plan", "blueprint-plan", "review-findings", "scout-survey"} {
		invalid := *k
		artifacts := append([]knowledge.Artifact(nil), k.Artifacts()...)
		for i, a := range artifacts {
			if a.Name == inFlight {
				artifacts[i].Visibility = append(artifacts[i].Visibility, knowledge.RoleCartographer)
			}
		}
		setArtifacts(&invalid, artifacts)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Errorf("expected error when Cartographer is in %s visibility", inFlight)
		}
	}

	// 6. Cartographer in non-allowlisted artifact visibility
	{
		invalid := *k
		artifacts := append([]knowledge.Artifact(nil), k.Artifacts()...)
		artifacts = append(artifacts, knowledge.Artifact{
			Name:        "custom-internal-spec",
			Title:       "Custom Internal Spec",
			Description: "Non-allowlisted document",
			Owner:       knowledge.RolePlanner,
			Visibility:  []knowledge.Role{knowledge.RolePlanner, knowledge.RoleCartographer},
			Type:        "document",
			Sections:    []knowledge.ArtifactSection{{Name: "Content", Required: true, Description: "Content"}},
		})
		setArtifacts(&invalid, artifacts)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Cartographer is in non-allowlisted artifact visibility")
		}
	}

	// 7. Cartographer as flow actor in engineering flows (build, init, plan, review, commit)
	for _, engFlow := range []string{"init", "plan", "blueprint", "build", "review", "commit"} {
		invalid := *k
		flows := append([]knowledge.Flow(nil), k.Flows()...)
		for i, f := range flows {
			if f.Name == engFlow {
				flows[i].Steps = append(flows[i].Steps, knowledge.FlowStep{
					Index:  len(f.Steps) + 1,
					Actor:  knowledge.RoleCartographer,
					Action: "Unauthorized action by cartographer",
				})
			}
		}
		setFlows(&invalid, flows)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Errorf("expected error when Cartographer is an actor in %s flow", engFlow)
		}
	}
}

func TestValidateDiagramPath(t *testing.T) {
	t.Parallel()

	validPaths := []string{
		"docs/diagrams/system-topology.html",
		"docs/diagrams/code-development-lifecycle.html",
		"docs/diagrams/overview.v2.html",
		"docs/diagrams/flow_123-abc.html",
		"docs/diagrams/arch.v1.2.3.html",
	}
	for _, p := range validPaths {
		if err := knowledge.ValidateDiagramPath(p); err != nil {
			t.Errorf("expected path %q to be valid, got err: %v", p, err)
		}
	}

	invalidPaths := []struct {
		path string
		desc string
	}{
		{"", "empty path"},
		{"../etc/passwd", "directory traversal"},
		{"docs/diagrams/../escape.html", "directory traversal within path"},
		{"docs/diagrams/foo.svg", "non-html extension svg"},
		{"docs/diagrams/foo.png", "non-html extension png"},
		{"diagrams/foo.html", "missing docs/ prefix"},
		{"/docs/diagrams/foo.html", "leading slash absolute path"},
		{"docs/diagrams/sub/foo.html", "nested directory inside diagrams"},
		{"docs\\diagrams\\foo.html", "backslashes"},
		{"docs/diagrams/foo..html", "double dots"},
		{"docs/diagrams/.html", "missing base name"},
		{"docs/diagrams/foo$.html", "invalid character $"},
	}
	for _, tc := range invalidPaths {
		if err := knowledge.ValidateDiagramPath(tc.path); err == nil {
			t.Errorf("expected path %q (%s) to be rejected, got nil", tc.path, tc.desc)
		}
	}
}

func TestEstimateTokenCount(t *testing.T) {
	t.Parallel()

	if tokens := knowledge.EstimateTokenCount(""); tokens != 0 {
		t.Errorf("expected 0 tokens for empty string, got %d", tokens)
	}

	// 1 word of 4 letters: (4+3)/4 = 1 token
	if tokens := knowledge.EstimateTokenCount("word"); tokens != 1 {
		t.Errorf("expected 1 token for 'word', got %d", tokens)
	}

	// 1 word of 5 letters: (5+3)/4 = 2 tokens
	if tokens := knowledge.EstimateTokenCount("hello"); tokens != 2 {
		t.Errorf("expected 2 tokens for 'hello', got %d", tokens)
	}

	// 1 word of 20 letters: (20+3)/4 = 5 tokens
	if tokens := knowledge.EstimateTokenCount("antidisestablishment"); tokens != 5 {
		t.Errorf("expected 5 tokens for 20-letter word, got %d", tokens)
	}

	// 20 words of 20 letters: 20 * 5 = 100 tokens
	longWords := strings.TrimSpace(strings.Repeat("antidisestablishment ", 20))
	if tokens := knowledge.EstimateTokenCount(longWords); tokens != 100 {
		t.Errorf("expected 100 tokens for 20 20-letter words, got %d", tokens)
	}
}

func TestValidateDiagramCompletion(t *testing.T) {
	t.Parallel()

	validCases := []struct {
		uri    string
		digest string
	}{
		{
			uri:    "docs/diagrams/system-topology.html",
			digest: "System topology diagram illustrating 5 microservices and event queues.",
		},
		{
			uri:    "file://docs/diagrams/code-development-lifecycle.html",
			digest: "Lifecycle sequence showing state machine progression across 6 roles.",
		},
		{
			uri:    "docs/diagrams/overview.v2.html",
			digest: "High-level overview mapping out architectural layers and boundaries.",
		},
	}
	for _, tc := range validCases {
		if err := knowledge.ValidateDiagramCompletion(tc.uri, tc.digest); err != nil {
			t.Errorf("expected valid diagram completion (%s, %s), got err: %v", tc.uri, tc.digest, err)
		}
	}

	// Assert that absolute file:/// URIs are directly rejected with the specific error message
	absURI := "file:///absolute/path/docs/diagrams/overview.html"
	err := knowledge.ValidateDiagramCompletion(absURI, "Valid single-sentence digest.")
	if err == nil {
		t.Errorf("expected absolute file:/// URI %q to be rejected, got nil", absURI)
	} else {
		expectedSubstr := "absolute file:/// URIs are prohibited; use repo-relative path docs/diagrams/<name>.html or file://docs/diagrams/<name>.html"
		if !strings.Contains(err.Error(), expectedSubstr) {
			t.Errorf("expected error to contain %q, got: %v", expectedSubstr, err)
		}
	}

	invalidCases := []struct {
		uri    string
		digest string
		desc   string
	}{
		{"", "Valid digest.", "empty URI"},
		{"http://evil.com/diagram.html", "Valid digest.", "external http URI"},
		{"https://example.com/diagram.html", "Valid digest.", "external https URI"},
		{"../escape.html", "Valid digest.", "traversal URI"},
		{"file://docs/diagrams/../escape.html", "Valid digest.", "file URI traversal"},
		{"file:///etc/passwd", "Valid digest.", "file URI outside diagrams"},
		{"file:///outside/repo/docs/diagrams/name.html", "Valid digest.", "file URI outside repo working directory"},
		{"file:///any/other/root/docs/diagrams/name.html", "Valid digest.", "file URI with arbitrary root outside repo"},
		{"/docs/diagrams/foo.html", "Valid digest.", "absolute URI without file:// scheme"},
		{"docs/diagrams/foo.svg", "Valid digest.", "non-html URI"},
		{"docs/diagrams/foo.html", "", "empty digest"},
		{"docs/diagrams/foo.html", "<svg>alert(1)</svg>", "raw SVG markup in digest"},
		{"docs/diagrams/foo.html", "<div>Invalid</div>", "raw HTML markup in digest"},
		{"docs/diagrams/foo.html", "Has < opening tag.", "opening tag in digest"},
		{"docs/diagrams/foo.html", "Has > closing tag.", "closing tag in digest"},
		{"docs/diagrams/foo.html", "Line 1.\nLine 2.", "newline in digest"},
		{"docs/diagrams/foo.html", "Line 1.\rLine 2.", "carriage return in digest"},
		{"docs/diagrams/foo.html", "First sentence. Second sentence.", "multi-sentence digest with period"},
		{"docs/diagrams/foo.html", "First sentence! Second sentence.", "multi-sentence digest with exclamation mark"},
		{"docs/diagrams/foo.html", "Question one? Question two.", "multi-sentence digest with question mark"},
		{
			uri:    "docs/diagrams/foo.html",
			digest: strings.Repeat("A", 251) + ".",
			desc:   "digest exceeding 250 characters",
		},
		{
			uri:    "docs/diagrams/foo.html",
			digest: strings.TrimSpace(strings.Repeat("word ", 61)) + ".",
			desc:   "digest exceeding 60 words limit",
		},
		{
			uri:    "docs/diagrams/foo.html",
			digest: strings.TrimSpace(strings.Repeat("antidisestablishment ", 20)) + ".",
			desc:   "digest exceeding 100 tokens limit with subword-heavy words",
		},
	}
	for _, tc := range invalidCases {
		if err := knowledge.ValidateDiagramCompletion(tc.uri, tc.digest); err == nil {
			t.Errorf("expected invalid diagram completion (%s, %s) [%s] to fail, got nil", tc.uri, tc.digest, tc.desc)
		}
	}
}

func TestValidate_VerifierConstraints(t *testing.T) {
	t.Parallel()

	k, err := knowledge.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// 1. Verifier category must be core
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RoleVerifier {
				roles[i].Category = "invalid_cat"
				break
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error for invalid verifier category")
		}
	}

	// 2. Verifier communication target cannot be builder or user
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RoleVerifier {
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RoleBuilder}
				break
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when verifier communicates with builder")
		}
	}

	// 3. Planner communication targets must include verifier
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RolePlanner {
				// Drop verifier
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RoleBuilder, knowledge.RoleReviewer, knowledge.RoleScout, knowledge.RoleNavigator, knowledge.RoleCartographer}
				break
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when planner omits verifier")
		}
	}

	// 4. Verifier cannot own artifacts other than e2e-report
	{
		invalid := *k
		artifacts := append([]knowledge.Artifact(nil), k.Artifacts()...)
		for i, a := range artifacts {
			if a.Name == "build-plan" {
				artifacts[i].Owner = knowledge.RoleVerifier
				break
			}
		}
		setArtifacts(&invalid, artifacts)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when verifier owns build-plan")
		}
	}

	// 5. Verifier cannot be an actor in flow build or flow review
	{
		invalid := *k
		flows := append([]knowledge.Flow(nil), k.Flows()...)
		for i, f := range flows {
			if f.Name == "build" {
				steps := append([]knowledge.FlowStep(nil), f.Steps...)
				steps[0].Actor = knowledge.RoleVerifier
				flows[i].Steps = steps
				break
			}
		}
		setFlows(&invalid, flows)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when verifier is actor in flow build")
		}
	}

	// 6. Verifier cannot be visible on unapproved artifacts (e.g. build-plan)
	{
		invalid := *k
		artifacts := append([]knowledge.Artifact(nil), k.Artifacts()...)
		for i, a := range artifacts {
			if a.Name == "build-plan" {
				artifacts[i].Visibility = append(artifacts[i].Visibility, knowledge.RoleVerifier)
				break
			}
		}
		setArtifacts(&invalid, artifacts)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when verifier is visible on build-plan")
		}
	}
}

func setRoles(k *knowledge.Knowledge, roles []knowledge.RoleDefinition) {
	k.SetRolesForTest(roles)
}

func setArtifacts(k *knowledge.Knowledge, artifacts []knowledge.Artifact) {
	k.SetArtifactsForTest(artifacts)
}

func setFlows(k *knowledge.Knowledge, flows []knowledge.Flow) {
	k.SetFlowsForTest(flows)
}

func TestValidate_TerminalStepsAndReachability(t *testing.T) {
	t.Parallel()

	k, err := knowledge.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// 1. Terminal step declaring conditions must fail validation
	{
		invalid := *k
		flows := append([]knowledge.Flow(nil), k.Flows()...)
		for i, f := range flows {
			if f.Name == "review" {
				steps := append([]knowledge.FlowStep(nil), f.Steps...)
				for j, s := range steps {
					if s.Index == 6 {
						steps[j].Terminal = true
						steps[j].Conditions = []knowledge.Condition{{When: "INVALID_CONDITION", Then: 7}}
						break
					}
				}
				flows[i].Steps = steps
				break
			}
		}
		setFlows(&invalid, flows)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when a terminal step defines conditions")
		}
	}

	// 2. Synthetic flow test: non-maximum step with Terminal: true suppresses sequential next step requirement; Terminal: false fails
	{
		// Valid case: Step 2 has Terminal: true, missing step 3 is permitted
		valid := *k
		validFlows := append([]knowledge.Flow(nil), k.Flows()...)
		validFlows = append(validFlows, knowledge.Flow{
			Name:        "test-synthetic-terminal",
			Description: "Synthetic flow to test terminal step sequential suppression",
			Steps: []knowledge.FlowStep{
				{Index: 1, Actor: knowledge.RolePlanner, Action: "Initial step"},
				{Index: 2, Actor: knowledge.RolePlanner, Action: "Terminal intermediate step", Terminal: true},
				{Index: 4, Actor: knowledge.RolePlanner, Action: "Max step", Terminal: true},
			},
		})
		setFlows(&valid, validFlows)
		if err := knowledge.Validate(&valid); err != nil {
			t.Errorf("expected synthetic flow with Terminal: true on intermediate step to pass validation, got: %v", err)
		}

		// Invalid case: Step 2 has Terminal: false, missing step 3 triggers validation error
		invalid := *k
		invalidFlows := append([]knowledge.Flow(nil), k.Flows()...)
		invalidFlows = append(invalidFlows, knowledge.Flow{
			Name:        "test-synthetic-terminal",
			Description: "Synthetic flow to test terminal step sequential suppression",
			Steps: []knowledge.FlowStep{
				{Index: 1, Actor: knowledge.RolePlanner, Action: "Initial step"},
				{Index: 2, Actor: knowledge.RolePlanner, Action: "Non-terminal intermediate step", Terminal: false},
				{Index: 4, Actor: knowledge.RolePlanner, Action: "Max step", Terminal: true},
			},
		})
		setFlows(&invalid, invalidFlows)
		err := knowledge.Validate(&invalid)
		if err == nil {
			t.Error("expected error when a non-terminal non-maximum step lacks sequential next step")
		} else if !strings.Contains(err.Error(), "has no conditions but missing sequential next step 3") {
			t.Errorf("expected error to contain 'has no conditions but missing sequential next step 3', got: %v", err)
		}
	}

	// 3. Reachability test: review Step 6 is Terminal: true, has zero conditions, and cannot reach Step 7
	{
		reviewFlow, ok := k.Flow("review")
		if !ok {
			t.Fatalf("expected review flow to exist")
		}
		var step6 *knowledge.FlowStep
		for _, s := range reviewFlow.Steps {
			if s.Index == 6 {
				sCopy := s
				step6 = &sCopy
				break
			}
		}
		if step6 == nil {
			t.Fatalf("expected review step 6 to exist")
		}
		if !step6.Terminal {
			t.Errorf("expected review step 6 to be terminal: true")
		}
		if len(step6.Conditions) != 0 {
			t.Errorf("expected review step 6 to have zero conditions, got %d", len(step6.Conditions))
		}
		for _, c := range step6.Conditions {
			if c.Then == 7 {
				t.Errorf("review step 6 must not transition to step 7")
			}
		}
	}

	// 4. Reachability test: e2e Step 5 is Terminal: true, has zero conditions, and cannot reach Step 6
	{
		e2eFlow, ok := k.Flow("e2e")
		if !ok {
			t.Fatalf("expected e2e flow to exist")
		}
		var step5 *knowledge.FlowStep
		for _, s := range e2eFlow.Steps {
			if s.Index == 5 {
				sCopy := s
				step5 = &sCopy
				break
			}
		}
		if step5 == nil {
			t.Fatalf("expected e2e step 5 to exist")
		}
		if !step5.Terminal {
			t.Errorf("expected e2e step 5 to be terminal: true")
		}
		if len(step5.Conditions) != 0 {
			t.Errorf("expected e2e step 5 to have zero conditions, got %d", len(step5.Conditions))
		}
		for _, c := range step5.Conditions {
			if c.Then == 6 {
				t.Errorf("e2e step 5 must not transition to step 6")
			}
		}
	}

	// 5. Reachability test: e2e Step 8 is Terminal: true, has zero conditions, and cannot reach Step 1, 2, or 5
	{
		e2eFlow, ok := k.Flow("e2e")
		if !ok {
			t.Fatalf("expected e2e flow to exist")
		}
		var step8 *knowledge.FlowStep
		for _, s := range e2eFlow.Steps {
			if s.Index == 8 {
				sCopy := s
				step8 = &sCopy
				break
			}
		}
		if step8 == nil {
			t.Fatalf("expected e2e step 8 to exist")
		}
		if !step8.Terminal {
			t.Errorf("expected e2e step 8 to be terminal: true")
		}
		if len(step8.Conditions) != 0 {
			t.Errorf("expected e2e step 8 to have zero conditions, got %d", len(step8.Conditions))
		}
		for _, c := range step8.Conditions {
			if c.Then == 1 || c.Then == 2 || c.Then == 5 {
				t.Errorf("e2e step 8 must not transition to step %d", c.Then)
			}
		}
	}
}

func validTestBrief() knowledge.E2EBriefContext {
	return knowledge.E2EBriefContext{
		RunID:             "e2e-run-20261012T220000Z-7f8a9b",
		RepoRef:           "github.com/ChiaYuChang/agentplaybook",
		CandidateRef:      "0123456789abcdef0123456789abcdef01234567",
		TestPackage:       "e2e/core-flow",
		Mode:              "build-and-run",
		Command:           []string{"go", "test", "-count=1", "./e2e/core-flow/..."},
		BuildCommand:      []string{"go", "build", "-o", "bin/agentplaybook", "./cmd/agentplaybook"},
		TimeoutMS:         180000,
		GoVersion:         "go1.24.0",
		Platforms:         []string{"linux/amd64"},
		DependencyLockRef: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		Environment: []knowledge.E2EEnvironmentInput{
			{Name: "APP_ENV", Source: "LITERAL", Value: "test"},
		},
		ExpectedOutputs: []string{"test-report.json"},
		Scenarios: []knowledge.E2EScenarioContract{
			{ID: "SC-01", Preconditions: "Clean DB", Steps: []string{"Step 1"}, ExpectedResult: "Pass", Verification: "Assert 200"},
			{ID: "SC-02", Preconditions: "Clean DB", Steps: []string{"Step 2"}, ExpectedResult: "Pass", Verification: "Assert 200"},
			{ID: "SC-03", Preconditions: "Clean DB", Steps: []string{"Step 3"}, ExpectedResult: "Skip", Verification: "Assert skip"},
		},
		AllowedSkipIDs: []string{"SC-03"},
		Services:       []knowledge.E2EServiceRequirement{},
		CleanupObligations: []knowledge.E2ECleanupObligation{
			{ResourceID: "test-data", Kind: "TEST_DATA", Owner: "verifier", Action: "rm -rf"},
		},
		CredentialPolicy: []knowledge.E2ECredentialPolicyEntry{
			{Name: "test-token", GenerationMode: "THROWAWAY", GenerationInstruction: "generate ephemeral token"},
		},
		CredentialAuthorization: nil,
		PriorCleanupResolution:  nil,
	}
}

func validTestBriefState() knowledge.E2EBriefValidationContext {
	return knowledge.E2EBriefValidationContext{
		RunIDUnique:                     true,
		CredentialAuthorizationVerified: true,
		PriorCleanup:                    knowledge.E2EPriorCleanupContext{Required: false},
	}
}

func validTestReport(brief knowledge.E2EBriefContext) knowledge.E2EReportPayload {
	attestation, _ := knowledge.ComputeCanonicalSecretsAttestation(brief.CredentialPolicy, brief.CredentialAuthorization)
	bmd := knowledge.ComputeBuildMaterialDigest(brief.RepoRef, brief.CandidateRef, "sha256:1111111111111111111111111111111111111111111111111111111111111111", brief.BuildCommand, brief.GoVersion, brief.Platforms, brief.DependencyLockRef)
	artifactRef := knowledge.FormatTestedArtifactRef("2222222222222222222222222222222222222222222222222222222222222222", brief.CandidateRef, bmd, "FRESH")

	mat := validTestMaterial(brief)
	execRef := knowledge.ComputeExecutionMaterialRef(brief, mat)

	return knowledge.E2EReportPayload{
		RunID:             brief.RunID,
		RepoRef:           brief.RepoRef,
		CandidateRef:      brief.CandidateRef,
		TestPackage:       brief.TestPackage,
		StageOutcome:      "PASS",
		BuildOutcome:      "PASS",
		ExecutionOutcome:  "PASS",
		ScenariosSelected: 3,
		ScenariosPassed:   2,
		ScenariosFailed:   0,
		ScenariosSkipped:  1,
		ScenariosNotRun:   0,
		ScenariosUnknown:  0,
		ScenarioResults: []knowledge.E2EScenarioResult{
			{ScenarioID: "SC-01", Outcome: "PASS"},
			{ScenarioID: "SC-02", Outcome: "PASS"},
			{ScenarioID: "SC-03", Outcome: "SKIP"},
		},
		ResultCompleteness:   "COMPLETE",
		TestedArtifactRef:    artifactRef,
		ExecutionMaterialRef: execRef,
		DurationMS:           12500,
		Diagnostic:           "All selected scenarios passed with 1 authorized skip.",
		EvidenceURI:          "file:///tmp/e2e-run-7f8a9b/evidence.tar.gz",
		CleanupStatus:        "CONFIRMED",
		CleanupEvidenceURI:   "file:///tmp/e2e-run-7f8a9b/cleanup.json",
		SecretsAttestation:   attestation,
	}
}

func validTestReportState(report knowledge.E2EReportPayload) knowledge.E2EReportValidationContext {
	provState := "SUCCEEDED"
	if strings.HasPrefix(report.SecretsAttestation, "NOT_PROVISIONED") {
		provState = "NOT_STARTED"
	}
	return knowledge.E2EReportValidationContext{
		RunIDUnique:                     true,
		CredentialAuthorizationVerified: true,
		PriorCleanupResolved:            true,
		PreflightOutcome:                "READY",
		BuildProcessState:               "PASS",
		ExecutionProcessState:           "PASS",
		ExecutionStarted:                true,
		ServiceStartupAttempted:         false,
		CredentialProvisioningState:     provState,
		ResourcesAllocated:              true,
		AllocatedResourceIDs:            []string{"test-data"},
		UnresolvedResourceIDs:           nil,
		CleanupRequired:                 true,
		CleanupStatus:                   "CONFIRMED",
		CleanupRecordVerified:           true,
		Finalization: knowledge.E2EReportFinalizationContext{
			State:                   "FIRST_FINALIZATION",
			Authority:               "VERIFIER",
			IncomingReportDigest:    knowledge.ComputeReportDigest(report),
			TerminalReportImmutable: true,
		},
	}
}

func validTestMaterial(brief knowledge.E2EBriefContext) knowledge.E2EExecutionMaterialContext {
	treeRef := "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	bmd := knowledge.ComputeBuildMaterialDigest(brief.RepoRef, brief.CandidateRef, treeRef, brief.BuildCommand, brief.GoVersion, brief.Platforms, brief.DependencyLockRef)

	return knowledge.E2EExecutionMaterialContext{
		RepoRef:                   brief.RepoRef,
		CandidateTreeRef:          treeRef,
		SourceTreeBeforeRef:       treeRef,
		SourceTreeAfterRef:        treeRef,
		CheckoutCleanBefore:       true,
		CheckoutCleanAfter:        true,
		InputProtectionMode:       "READ_ONLY_MATERIALIZATION",
		ProtectionProofRef:        "ro-proof-12345",
		ProtectedIntervalComplete: true,
		ProtectionProofVerified:   true,
		TestPackagePath:           brief.TestPackage,
		TestPackageTreeRef:        "sha256:3333333333333333333333333333333333333333333333333333333333333333",
		Command:                   brief.Command,
		BuildCommand:              brief.BuildCommand,
		GoVersion:                 brief.GoVersion,
		Platforms:                 brief.Platforms,
		DependencyLockRef:         brief.DependencyLockRef,
		ServiceRefs:               nil,
		FreshTestExecution:        true,
		ArtifactSHA256:            "2222222222222222222222222222222222222222222222222222222222222222",
		ArtifactCandidateRef:      brief.CandidateRef,
		ArtifactBuildMaterialRef:  "sha256:" + bmd,
		ArtifactSource:            "FRESH",
	}
}

func TestValidateE2EReport(t *testing.T) {
	t.Parallel()

	brief := validTestBrief()
	report := validTestReport(brief)
	state := validTestReportState(report)

	// 1. Success test: build-and-run
	if err := knowledge.ValidateE2EReport(brief, report, state); err != nil {
		t.Fatalf("expected valid build-and-run report to pass: %v", err)
	}

	// 2. Deterministic Format and Digest assertions
	formatted := knowledge.FormatE2EReport(report)
	if !strings.HasPrefix(formatted, "run_id: ") {
		t.Errorf("expected formatted report to start with run_id, got: %s", formatted)
	}
	if !strings.Contains(formatted, "SC-01: PASS") {
		t.Errorf("expected formatted report to contain scenario results, got: %s", formatted)
	}
	digest := knowledge.ComputeReportDigest(report)
	if len(digest) != 64 {
		t.Errorf("expected 64-char sha256 hex digest, got %d chars: %s", len(digest), digest)
	}

	// 3. Success test: run-only
	{
		runOnlyBrief := brief
		runOnlyBrief.Mode = "run-only"
		runOnlyReport := report
		runOnlyReport.BuildOutcome = "NOT_RUN"
		runOnlyState := validTestReportState(runOnlyReport)
		runOnlyState.BuildProcessState = "NOT_STARTED"
		runOnlyState.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(runOnlyReport)
		if err := knowledge.ValidateE2EReport(runOnlyBrief, runOnlyReport, runOnlyState); err != nil {
			t.Fatalf("expected valid run-only report to pass: %v", err)
		}
	}

	// 4. Preflight AUTHORIZATION_DENIED test
	{
		deniedBrief := brief
		deniedBrief.CredentialPolicy = []knowledge.E2ECredentialPolicyEntry{
			{Name: "prod-db-token", GenerationMode: "REAL_ONLY"},
		}
		deniedBrief.CredentialAuthorization = []knowledge.E2ECredentialAuthorization{
			{
				Name:             "prod-db-token",
				Purpose:          "Integration test",
				Lifetime:         "SINGLE_RUN",
				AuthorizedBy:     "operator",
				AuthorizedAt:     "2026-10-12T20:00:00Z",
				AuthorizationRef: "auth-12345",
			},
		}

		deniedReport := report
		deniedReport.StageOutcome = "ENV_BLOCKED"
		deniedReport.BuildOutcome = "NOT_RUN"
		deniedReport.ExecutionOutcome = "ENV_ERROR"
		deniedReport.ScenariosSelected = 3
		deniedReport.ScenariosPassed = 0
		deniedReport.ScenariosFailed = 0
		deniedReport.ScenariosSkipped = 0
		deniedReport.ScenariosNotRun = 3
		deniedReport.ScenariosUnknown = 0
		deniedReport.ScenarioResults = []knowledge.E2EScenarioResult{
			{ScenarioID: "SC-01", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-02", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-03", Outcome: "NOT_RUN"},
		}
		deniedReport.ResultCompleteness = "NONE"
		deniedReport.TestedArtifactRef = "NONE"
		deniedReport.ExecutionMaterialRef = "UNAVAILABLE"
		deniedReport.EvidenceURI = "UNAVAILABLE"
		deniedReport.CleanupStatus = "CONFIRMED"
		deniedReport.CleanupEvidenceURI = "NOT_REQUIRED"
		deniedReport.SecretsAttestation = "ENV_BLOCKED:AUTHORIZATION_DENIED"

		deniedState := knowledge.E2EReportValidationContext{
			RunIDUnique:                     true,
			CredentialAuthorizationVerified: true,
			PriorCleanupResolved:            true,
			PreflightOutcome:                "AUTHORIZATION_DENIED",
			BuildProcessState:               "NOT_STARTED",
			ExecutionProcessState:           "NOT_STARTED",
			ExecutionStarted:                false,
			ServiceStartupAttempted:         false,
			CredentialProvisioningState:     "NOT_STARTED",
			ResourcesAllocated:              false,
			CleanupRequired:                 false,
			CleanupStatus:                   "CONFIRMED",
			Finalization: knowledge.E2EReportFinalizationContext{
				State:                   "FIRST_FINALIZATION",
				Authority:               "VERIFIER",
				IncomingReportDigest:    knowledge.ComputeReportDigest(deniedReport),
				TerminalReportImmutable: true,
			},
		}

		if err := knowledge.ValidateE2EReport(deniedBrief, deniedReport, deniedState); err != nil {
			t.Fatalf("expected AUTHORIZATION_DENIED report to pass: %v", err)
		}

		// Negative: AUTHORIZATION_DENIED with attempted execution
		invalidDeniedState := deniedState
		invalidDeniedState.ExecutionStarted = true
		if err := knowledge.ValidateE2EReport(deniedBrief, deniedReport, invalidDeniedState); err == nil {
			t.Error("expected error when AUTHORIZATION_DENIED has executionStarted true")
		}
	}

	// 5. Negative: ledger uniqueness violation
	{
		badState := validTestReportState(report)
		badState.RunIDUnique = false
		if err := knowledge.ValidateE2EReport(brief, report, badState); err == nil {
			t.Error("expected error when run_id is not unique in ledger")
		}
	}

	// 6. Negative: prior cleanup unconfirmed
	{
		badState := validTestReportState(report)
		badState.PriorCleanupResolved = false
		if err := knowledge.ValidateE2EReport(brief, report, badState); err == nil {
			t.Error("expected error when prior cleanup is not resolved")
		}
	}

	// 7. Negative: correlation mismatch
	{
		mismatchReport := report
		mismatchReport.RunID = "different-run-id"
		mismatchState := validTestReportState(mismatchReport)
		if err := knowledge.ValidateE2EReport(brief, mismatchReport, mismatchState); err == nil {
			t.Error("expected error for run_id mismatch")
		}

		mismatchReport = report
		mismatchReport.CandidateRef = "ffffffffffffffffffffffffffffffffffffffff"
		mismatchState = validTestReportState(mismatchReport)
		if err := knowledge.ValidateE2EReport(brief, mismatchReport, mismatchState); err == nil {
			t.Error("expected error for candidate_ref mismatch")
		}
	}

	// 8. Negative: diagnostic multi-line or length > 160 bytes
	{
		multiLineReport := report
		multiLineReport.Diagnostic = "Line 1\nLine 2"
		multiLineState := validTestReportState(multiLineReport)
		if err := knowledge.ValidateE2EReport(brief, multiLineReport, multiLineState); err == nil {
			t.Error("expected error for multi-line diagnostic")
		}

		longReport := report
		longReport.Diagnostic = strings.Repeat("a", 161)
		longState := validTestReportState(longReport)
		if err := knowledge.ValidateE2EReport(brief, longReport, longState); err == nil {
			t.Error("expected error for diagnostic exceeding 160 bytes")
		}
	}

	// 9. Negative: count reconciliation mismatch
	{
		badCountReport := report
		badCountReport.ScenariosPassed = 1 // 1 + 0 + 1 + 0 + 0 = 2 != 3
		badCountState := validTestReportState(badCountReport)
		if err := knowledge.ValidateE2EReport(brief, badCountReport, badCountState); err == nil {
			t.Error("expected error for scenario count reconciliation mismatch")
		}
	}

	// 10. Negative: unauthorized skip on PASS
	{
		badSkipBrief := brief
		badSkipBrief.AllowedSkipIDs = nil // no skips allowed
		badSkipState := validTestReportState(report)
		if err := knowledge.ValidateE2EReport(badSkipBrief, report, badSkipState); err == nil {
			t.Error("expected error when unauthorized skip is present with PASS outcome")
		}
	}

	// 11. Negative: UNCONFIRMED cleanup status
	{
		unconfirmedReport := report
		unconfirmedReport.CleanupStatus = "UNCONFIRMED"
		unconfirmedState := validTestReportState(unconfirmedReport)
		unconfirmedState.CleanupStatus = "UNCONFIRMED"
		unconfirmedState.UnresolvedResourceIDs = []string{"dangling-volume"}
		// UNCONFIRMED with PASS stage outcome must fail
		if err := knowledge.ValidateE2EReport(brief, unconfirmedReport, unconfirmedState); err == nil {
			t.Error("expected error when stage outcome is PASS but cleanup is UNCONFIRMED")
		}
	}

	// 12. Negative: invalid SecretsAttestation
	{
		badAttestReport := report
		badAttestReport.SecretsAttestation = "ATTESTED:THROWAWAY sha256:0000000000000000000000000000000000000000000000000000000000000000"
		badAttestState := validTestReportState(badAttestReport)
		if err := knowledge.ValidateE2EReport(brief, badAttestReport, badAttestState); err == nil {
			t.Error("expected error for secrets attestation hash mismatch")
		}
	}
}

func TestEvaluateE2EAdmission(t *testing.T) {
	t.Parallel()

	brief := validTestBrief()
	report := validTestReport(brief)
	mat := validTestMaterial(brief)
	state := validTestReportState(report)

	// 1. Success test: complete fresh PASS admission
	if err := knowledge.EvaluateE2EAdmission(brief, report, mat, state); err != nil {
		t.Fatalf("expected valid admission to pass: %v", err)
	}

	// 2. Denied: non-PASS stage outcome
	{
		failReport := report
		failReport.StageOutcome = "PRODUCT_FAILURE"
		failReport.ExecutionOutcome = "FAIL"
		failReport.ScenariosFailed = 1
		failReport.ScenariosPassed = 1
		failReport.ScenarioResults[0].Outcome = "FAIL"
		failState := validTestReportState(failReport)
		failState.ExecutionProcessState = "FAIL"
		if err := knowledge.EvaluateE2EAdmission(brief, failReport, mat, failState); err == nil {
			t.Error("expected admission denied for non-PASS outcome")
		}
	}

	// 3. Denied: unconfirmed cleanup
	{
		unconfirmedReport := report
		unconfirmedReport.CleanupStatus = "UNCONFIRMED"
		unconfirmedState := validTestReportState(unconfirmedReport)
		unconfirmedState.CleanupStatus = "UNCONFIRMED"
		unconfirmedState.UnresolvedResourceIDs = []string{"res-leak"}
		if err := knowledge.EvaluateE2EAdmission(brief, unconfirmedReport, mat, unconfirmedState); err == nil {
			t.Error("expected admission denied for UNCONFIRMED cleanup")
		}
	}

	// 4. Denied: transient source mutation (SourceTreeAfterRef != CandidateTreeRef)
	{
		mutatedMat := mat
		mutatedMat.SourceTreeAfterRef = "sha256:9999999999999999999999999999999999999999999999999999999999999999"
		if err := knowledge.EvaluateE2EAdmission(brief, report, mutatedMat, state); err == nil {
			t.Error("expected admission denied for source tree mutation")
		}
	}

	// 5. Denied: checkout dirty before or after
	{
		dirtyMat := mat
		dirtyMat.CheckoutCleanBefore = false
		if err := knowledge.EvaluateE2EAdmission(brief, report, dirtyMat, state); err == nil {
			t.Error("expected admission denied when checkout is not clean before execution")
		}
	}

	// 6. Denied: unverified protection proof
	{
		unverifiedMat := mat
		unverifiedMat.ProtectionProofVerified = false
		if err := knowledge.EvaluateE2EAdmission(brief, report, unverifiedMat, state); err == nil {
			t.Error("expected admission denied when protection proof is not verified")
		}
	}

	// 7. Denied: non-fresh execution (cached test verdicts)
	{
		cachedMat := mat
		cachedMat.FreshTestExecution = false
		if err := knowledge.EvaluateE2EAdmission(brief, report, cachedMat, state); err == nil {
			t.Error("expected admission denied for non-fresh test execution")
		}
	}

	// 8. Denied: command missing -count=1
	{
		noCountMat := mat
		noCountMat.Command = []string{"go", "test", "./e2e/core-flow/..."}
		noCountBrief := brief
		noCountBrief.Command = noCountMat.Command
		if err := knowledge.EvaluateE2EAdmission(noCountBrief, report, noCountMat, state); err == nil {
			t.Error("expected admission denied when test command lacks -count=1")
		}
	}

	// 9. Denied: artifact provenance mismatch
	{
		mismatchMat := mat
		mismatchMat.ArtifactCandidateRef = "ffffffffffffffffffffffffffffffffffffffff"
		if err := knowledge.EvaluateE2EAdmission(brief, report, mismatchMat, state); err == nil {
			t.Error("expected admission denied for artifact candidate ref mismatch")
		}
	}

	// 10. Denied: execution material ref mismatch
	{
		mismatchReport := report
		mismatchReport.ExecutionMaterialRef = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
		mismatchState := validTestReportState(mismatchReport)
		if err := knowledge.EvaluateE2EAdmission(brief, mismatchReport, mat, mismatchState); err == nil {
			t.Error("expected admission denied for execution_material_ref mismatch")
		}
	}
}

func TestValidate_NavigatorCartographyFlowActors(t *testing.T) {
	t.Parallel()

	k, err := knowledge.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// 1. Positive: default loaded knowledge is completely valid
	if err := knowledge.Validate(k); err != nil {
		t.Fatalf("expected loaded knowledge with navigator-cartography to be valid: %v", err)
	}

	// 2. Negative: Navigator actor in any other flow fails validation
	for _, f := range k.Flows() {
		if f.Name == "navigator-cartography" {
			continue
		}
		invalid := *k
		flows := append([]knowledge.Flow(nil), k.Flows()...)
		for i, fl := range flows {
			if fl.Name == f.Name {
				flows[i].Steps = append(flows[i].Steps, knowledge.FlowStep{
					Index:  len(fl.Steps) + 1,
					Actor:  knowledge.RoleNavigator,
					Action: "Unauthorized action by navigator",
				})
			}
		}
		setFlows(&invalid, flows)
		err := knowledge.Validate(&invalid)
		if err == nil {
			t.Errorf("expected error when Navigator is an actor in flow %q", f.Name)
		} else if !strings.Contains(err.Error(), "actor cannot be companion role navigator (permitted exclusively in navigator-cartography flow)") {
			t.Errorf("expected error for flow %q to contain specific navigator error, got: %v", f.Name, err)
		}
	}

	// 3. Negative: Cartographer actor in any flow other than cartography and navigator-cartography fails validation
	for _, f := range k.Flows() {
		if f.Name == "cartography" || f.Name == "navigator-cartography" {
			continue
		}
		invalid := *k
		flows := append([]knowledge.Flow(nil), k.Flows()...)
		for i, fl := range flows {
			if fl.Name == f.Name {
				flows[i].Steps = append(flows[i].Steps, knowledge.FlowStep{
					Index:  len(fl.Steps) + 1,
					Actor:  knowledge.RoleCartographer,
					Action: "Unauthorized action by cartographer",
				})
			}
		}
		setFlows(&invalid, flows)
		err := knowledge.Validate(&invalid)
		if err == nil {
			t.Errorf("expected error when Cartographer is an actor in flow %q", f.Name)
		} else if !strings.Contains(err.Error(), "actor cannot be companion role cartographer (permitted exclusively in cartography and navigator-cartography flows)") {
			t.Errorf("expected error for flow %q to contain specific cartographer error, got: %v", f.Name, err)
		}
	}

	// 4. Artifact allowlist: Navigator visibility with unallowed artifact fails validation
	{
		invalid := *k
		artifacts := append([]knowledge.Artifact(nil), k.Artifacts()...)
		artifacts = append(artifacts, knowledge.Artifact{
			Name:        "unallowed-for-navigator",
			Title:       "Unallowed Artifact",
			Description: "Artifact unallowed for navigator",
			Owner:       knowledge.RolePlanner,
			Visibility:  []knowledge.Role{knowledge.RolePlanner, knowledge.RoleNavigator},
			Type:        "document",
			Sections:    []knowledge.ArtifactSection{{Name: "Content", Required: true, Description: "Content"}},
		})
		setArtifacts(&invalid, artifacts)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Navigator is in non-allowlisted artifact visibility")
		}
	}
}

func TestExecutionMaterial_GoldenCanonicalJSON(t *testing.T) {
	t.Parallel()

	brief := knowledge.E2EBriefContext{
		RunID:             "run-1",
		RepoRef:           "github.com/org/repo",
		CandidateRef:      "0123456789abcdef0123456789abcdef01234567",
		TestPackage:       "e2e/test",
		Mode:              "build-and-run",
		Command:           []string{"go", "test", "-count=1", "./..."},
		BuildCommand:      []string{"go", "build", "./..."},
		TimeoutMS:         60000,
		GoVersion:         "go1.24.0",
		Platforms:         []string{"linux/amd64"},
		DependencyLockRef: "NONE",
		Environment:       []knowledge.E2EEnvironmentInput{},
		ExpectedOutputs:   []string{},
		Scenarios: []knowledge.E2EScenarioContract{
			{
				ID:             "S1",
				Preconditions:  "none",
				Steps:          []string{"step1"},
				ExpectedResult: "ok",
				Verification:   "assert",
			},
		},
		AllowedSkipIDs:     []string{},
		Services:           nil,
		CleanupObligations: []knowledge.E2ECleanupObligation{},
		CredentialPolicy:   []knowledge.E2ECredentialPolicyEntry{},
	}

	mat := knowledge.E2EExecutionMaterialContext{
		RepoRef:             "github.com/org/repo",
		CandidateTreeRef:    "sha256:1111111111111111111111111111111111111111111111111111111111111111",
		DependencyLockRef:   "NONE",
		GoVersion:           "go1.24.0",
		InputProtectionMode: "READ_ONLY_MATERIALIZATION",
		Platforms:           []string{"linux/amd64"},
		ProtectionProofRef:  "proof-1",
		ServiceRefs:         nil,
		TestPackagePath:     "e2e/test",
		TestPackageTreeRef:  "sha256:2222222222222222222222222222222222222222222222222222222222222222",
	}

	expectedGoldenJSON := `{"brief":{"allowed_skip_ids":[],"build_command":["go","build","./..."],"candidate_ref":"0123456789abcdef0123456789abcdef01234567","cleanup_obligations":[],"command":["go","test","-count=1","./..."],"credential_authorization":null,"credential_policy":[],"dependency_lock_ref":"NONE","environment":[],"expected_outputs":[],"go_version":"go1.24.0","mode":"build-and-run","platforms":["linux/amd64"],"prior_cleanup_resolution":null,"repo_ref":"github.com/org/repo","run_id":"run-1","scenarios":[{"expected_result":"ok","id":"S1","preconditions":"none","steps":["step1"],"verification":"assert"}],"services":null,"test_package":"e2e/test","timeout_ms":60000},"candidate_tree_ref":"sha256:1111111111111111111111111111111111111111111111111111111111111111","dependency_lock_ref":"NONE","go_version":"go1.24.0","input_protection_mode":"READ_ONLY_MATERIALIZATION","platforms":["linux/amd64"],"protection_proof_ref":"proof-1","repo_ref":"github.com/org/repo","service_refs":null,"test_package_path":"e2e/test","test_package_tree_ref":"sha256:2222222222222222222222222222222222222222222222222222222222222222"}`
	expectedGoldenDigest := "sha256:4c6fc89b8f7505ac9a60858a9cd0855d452519dfa517885f4dbf85a00665ac17"

	// 1. Direct CanonicalJSON test
	cem := knowledge.CanonicalExecutionMaterial{
		Brief:               brief,
		CandidateTreeRef:    mat.CandidateTreeRef,
		DependencyLockRef:   mat.DependencyLockRef,
		GoVersion:           mat.GoVersion,
		InputProtectionMode: mat.InputProtectionMode,
		Platforms:           mat.Platforms,
		ProtectionProofRef:  mat.ProtectionProofRef,
		RepoRef:             mat.RepoRef,
		ServiceRefs:         mat.ServiceRefs,
		TestPackagePath:     mat.TestPackagePath,
		TestPackageTreeRef:  mat.TestPackageTreeRef,
	}
	canonicalBytes, err := knowledge.CanonicalJSON(cem)
	if err != nil {
		t.Fatalf("CanonicalJSON failed: %v", err)
	}
	if string(canonicalBytes) != expectedGoldenJSON {
		t.Errorf("CanonicalJSON bytes mismatch:\nexpected: %s\ngot:      %s", expectedGoldenJSON, string(canonicalBytes))
	}

	// 2. Computed execution material ref test
	ref := knowledge.ComputeExecutionMaterialRef(brief, mat)
	if ref != expectedGoldenDigest {
		t.Errorf("ComputeExecutionMaterialRef mismatch: expected %q, got %q", expectedGoldenDigest, ref)
	}
}

func TestValidateE2EBrief_NilVsEmptyListsAndPathConfinement(t *testing.T) {
	t.Parallel()

	state := validTestBriefState()

	// 1. Omitted CredentialPolicy (nil) is rejected; explicit empty [] is accepted
	{
		b := validTestBrief()
		b.CredentialPolicy = nil
		if err := knowledge.ValidateE2EBrief(b, state); err == nil {
			t.Error("expected error for nil CredentialPolicy")
		} else if !strings.Contains(err.Error(), "credential_policy is required") {
			t.Errorf("unexpected error: %v", err)
		}

		b.CredentialPolicy = []knowledge.E2ECredentialPolicyEntry{}
		if err := knowledge.ValidateE2EBrief(b, state); err != nil {
			t.Errorf("expected explicit empty CredentialPolicy to pass, got: %v", err)
		}
	}

	// 2. Omitted ExpectedOutputs (nil) is rejected; explicit empty [] is accepted
	{
		b := validTestBrief()
		b.ExpectedOutputs = nil
		if err := knowledge.ValidateE2EBrief(b, state); err == nil {
			t.Error("expected error for nil ExpectedOutputs")
		} else if !strings.Contains(err.Error(), "expected_outputs is required") {
			t.Errorf("unexpected error: %v", err)
		}

		b.ExpectedOutputs = []string{}
		if err := knowledge.ValidateE2EBrief(b, state); err != nil {
			t.Errorf("expected explicit empty ExpectedOutputs to pass, got: %v", err)
		}
	}

	// 3. Omitted CleanupObligations (nil) is rejected; explicit empty [] without services is accepted
	{
		b := validTestBrief()
		b.CleanupObligations = nil
		if err := knowledge.ValidateE2EBrief(b, state); err == nil {
			t.Error("expected error for nil CleanupObligations")
		} else if !strings.Contains(err.Error(), "cleanup_obligations is required") {
			t.Errorf("unexpected error: %v", err)
		}

		b.CleanupObligations = []knowledge.E2ECleanupObligation{}
		if err := knowledge.ValidateE2EBrief(b, state); err != nil {
			t.Errorf("expected explicit empty CleanupObligations to pass, got: %v", err)
		}
	}

	// 4. Scenario steps validation: non-empty steps required
	{
		b := validTestBrief()
		b.Scenarios[0].Steps = []string{""}
		if err := knowledge.ValidateE2EBrief(b, state); err == nil {
			t.Error("expected error for empty step string in scenario")
		}

		b.Scenarios[0].Steps = []string{"   "}
		if err := knowledge.ValidateE2EBrief(b, state); err == nil {
			t.Error("expected error for whitespace step string in scenario")
		}
	}

	// 5. ExpectedOutputs path confinement: reject absolute and directory traversal
	{
		b := validTestBrief()
		b.ExpectedOutputs = []string{"/etc/passwd"}
		if err := knowledge.ValidateE2EBrief(b, state); err == nil {
			t.Error("expected error for absolute path in ExpectedOutputs")
		}

		b.ExpectedOutputs = []string{"../out.json"}
		if err := knowledge.ValidateE2EBrief(b, state); err == nil {
			t.Error("expected error for directory traversal in ExpectedOutputs")
		}

		b.ExpectedOutputs = []string{"artifacts/../out.json"}
		if err := knowledge.ValidateE2EBrief(b, state); err == nil {
			t.Error("expected error for directory traversal in ExpectedOutputs")
		}

		b.ExpectedOutputs = []string{"~/out.json"}
		if err := knowledge.ValidateE2EBrief(b, state); err == nil {
			t.Error("expected error for home-relative path in ExpectedOutputs")
		}

		b.ExpectedOutputs = []string{"reports/e2e-summary.json"}
		if err := knowledge.ValidateE2EBrief(b, state); err != nil {
			t.Errorf("expected clean relative path to pass, got: %v", err)
		}
	}
}

func TestValidateE2EReport_LifecycleFailClosed(t *testing.T) {
	t.Parallel()

	brief := validTestBrief()

	// 1. Unknown / invalid PreflightOutcome
	{
		report := validTestReport(brief)
		state := validTestReportState(report)
		state.PreflightOutcome = "UNKNOWN"
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error for invalid PreflightOutcome UNKNOWN")
		}
	}

	// 2. PASS requires PreflightOutcome READY
	{
		report := validTestReport(brief)
		state := validTestReportState(report)
		state.PreflightOutcome = "AUTHORIZATION_DENIED"
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error for PASS report with PreflightOutcome AUTHORIZATION_DENIED")
		}
	}

	// 3. PASS requires ExecutionStarted == true
	{
		report := validTestReport(brief)
		state := validTestReportState(report)
		state.ExecutionStarted = false
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error for PASS report with ExecutionStarted == false")
		}
	}

	// 4. Provisioning / allocation attempted implies CleanupRequired == true
	{
		report := validTestReport(brief)
		state := validTestReportState(report)
		state.ServiceStartupAttempted = true
		state.CleanupRequired = false
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error when ServiceStartupAttempted is true but CleanupRequired is false")
		}

		state = validTestReportState(report)
		state.CredentialProvisioningState = "SUCCEEDED"
		state.CleanupRequired = false
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error when CredentialProvisioningState is not NOT_STARTED but CleanupRequired is false")
		}

		state = validTestReportState(report)
		state.ResourcesAllocated = true
		state.CleanupRequired = false
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error when ResourcesAllocated is true but CleanupRequired is false")
		}
	}

	// 5. CleanupRequired == true prohibits NOT_REQUIRED evidence URI and requires verified cleanup record
	{
		report := validTestReport(brief)
		report.CleanupEvidenceURI = "NOT_REQUIRED"
		state := validTestReportState(report)
		state.CleanupRequired = true
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error when CleanupRequired is true but CleanupEvidenceURI is NOT_REQUIRED")
		}

		report = validTestReport(brief)
		state = validTestReportState(report)
		state.CleanupRequired = true
		state.CleanupRecordVerified = false
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error when CleanupRequired is true but CleanupRecordVerified is false")
		}
	}

	// 6. ResourcesAllocated consistency with AllocatedResourceIDs
	{
		report := validTestReport(brief)
		state := validTestReportState(report)
		state.ResourcesAllocated = true
		state.AllocatedResourceIDs = nil
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error when ResourcesAllocated is true but AllocatedResourceIDs is empty")
		}

		state = validTestReportState(report)
		state.ResourcesAllocated = false
		state.AllocatedResourceIDs = []string{"res-1"}
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error when ResourcesAllocated is false but AllocatedResourceIDs is non-empty")
		}
	}

	// 7. CleanupStatus consistency
	{
		report := validTestReport(brief)
		state := validTestReportState(report)
		state.CleanupStatus = "CONFIRMED"
		state.UnresolvedResourceIDs = []string{"res-1"}
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error when CleanupStatus is CONFIRMED but UnresolvedResourceIDs is non-empty")
		}

		report = validTestReport(brief)
		report.CleanupStatus = "UNCONFIRMED"
		state = validTestReportState(report)
		state.CleanupStatus = "UNCONFIRMED"
		state.UnresolvedResourceIDs = nil
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error when CleanupStatus is UNCONFIRMED but UnresolvedResourceIDs is empty")
		}
	}
}

func TestValidateE2EReport_ProviderDenialAuthorizationRecord(t *testing.T) {
	t.Parallel()

	// 1. Negative: PreflightOutcome AUTHORIZATION_DENIED without authorization record in brief
	{
		brief := validTestBrief()
		report := validTestReport(brief)
		report.StageOutcome = "ENV_BLOCKED"
		report.BuildOutcome = "NOT_RUN"
		report.ExecutionOutcome = "ENV_ERROR"
		report.ResultCompleteness = "NONE"
		report.TestedArtifactRef = "NONE"
		report.ExecutionMaterialRef = "UNAVAILABLE"
		report.EvidenceURI = "UNAVAILABLE"
		report.CleanupStatus = "CONFIRMED"
		report.CleanupEvidenceURI = "NOT_REQUIRED"
		report.SecretsAttestation = "ENV_BLOCKED:AUTHORIZATION_DENIED"
		report.ScenariosPassed = 0
		report.ScenariosFailed = 0
		report.ScenariosSkipped = 0
		report.ScenariosNotRun = report.ScenariosSelected
		report.ScenariosUnknown = 0
		report.ScenarioResults = []knowledge.E2EScenarioResult{
			{ScenarioID: "SC-01", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-02", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-03", Outcome: "NOT_RUN"},
		}

		state := knowledge.E2EReportValidationContext{
			RunIDUnique:                     true,
			CredentialAuthorizationVerified: true,
			PriorCleanupResolved:            true,
			PreflightOutcome:                "AUTHORIZATION_DENIED",
			BuildProcessState:               "NOT_STARTED",
			ExecutionProcessState:           "NOT_STARTED",
			ExecutionStarted:                false,
			CredentialProvisioningState:     "NOT_STARTED",
			CleanupStatus:                   "CONFIRMED",
			Finalization: knowledge.E2EReportFinalizationContext{
				State:                   "FIRST_FINALIZATION",
				Authority:               "VERIFIER",
				IncomingReportDigest:    knowledge.ComputeReportDigest(report),
				TerminalReportImmutable: true,
			},
		}

		// brief has THROWAWAY credential only, no REAL_ONLY, no authorization record
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error for AUTHORIZATION_DENIED when brief lacks REAL_ONLY credential authorization records")
		}
	}

	// 2. Negative: PreflightOutcome AUTHORIZATION_DENIED with unverified controller context
	{
		brief := validTestBrief()
		brief.CredentialPolicy = []knowledge.E2ECredentialPolicyEntry{
			{Name: "prod-db-token", GenerationMode: "REAL_ONLY"},
		}
		brief.CredentialAuthorization = []knowledge.E2ECredentialAuthorization{
			{
				Name:             "prod-db-token",
				Purpose:          "Integration test",
				Lifetime:         "SINGLE_RUN",
				AuthorizedBy:     "operator",
				AuthorizedAt:     "2026-10-12T20:00:00Z",
				AuthorizationRef: "auth-12345",
			},
		}
		report := validTestReport(brief)
		report.StageOutcome = "ENV_BLOCKED"
		report.BuildOutcome = "NOT_RUN"
		report.ExecutionOutcome = "ENV_ERROR"
		report.ResultCompleteness = "NONE"
		report.TestedArtifactRef = "NONE"
		report.ExecutionMaterialRef = "UNAVAILABLE"
		report.EvidenceURI = "UNAVAILABLE"
		report.CleanupStatus = "CONFIRMED"
		report.CleanupEvidenceURI = "NOT_REQUIRED"
		report.SecretsAttestation = "ENV_BLOCKED:AUTHORIZATION_DENIED"
		report.ScenariosPassed = 0
		report.ScenariosFailed = 0
		report.ScenariosSkipped = 0
		report.ScenariosNotRun = report.ScenariosSelected
		report.ScenariosUnknown = 0
		report.ScenarioResults = []knowledge.E2EScenarioResult{
			{ScenarioID: "SC-01", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-02", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-03", Outcome: "NOT_RUN"},
		}

		state := knowledge.E2EReportValidationContext{
			RunIDUnique:                     true,
			CredentialAuthorizationVerified: false, // unverified!
			PriorCleanupResolved:            true,
			PreflightOutcome:                "AUTHORIZATION_DENIED",
			BuildProcessState:               "NOT_STARTED",
			ExecutionProcessState:           "NOT_STARTED",
			ExecutionStarted:                false,
			CredentialProvisioningState:     "NOT_STARTED",
			CleanupStatus:                   "CONFIRMED",
			Finalization: knowledge.E2EReportFinalizationContext{
				State:                   "FIRST_FINALIZATION",
				Authority:               "VERIFIER",
				IncomingReportDigest:    knowledge.ComputeReportDigest(report),
				TerminalReportImmutable: true,
			},
		}
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error for AUTHORIZATION_DENIED when CredentialAuthorizationVerified is false")
		}
	}
}

func TestEvaluateE2EAdmission_ServiceBindingAndCommandInvocation(t *testing.T) {
	t.Parallel()

	brief := validTestBrief()
	mat := validTestMaterial(brief)
	report := validTestReport(brief)
	state := validTestReportState(report)

	// 1. ServiceRefs mismatch when brief declares services
	{
		svcBrief := brief
		svcBrief.Services = []knowledge.E2EServiceRequirement{
			{
				ID:               "postgres",
				Kind:             "TESTCONTAINER",
				ImageOrVersion:   "postgres:16-alpine",
				ReadyCondition:   "port 5432 open",
				StartupTimeoutMS: 30000,
			},
		}
		svcBrief.CleanupObligations = append(svcBrief.CleanupObligations, knowledge.E2ECleanupObligation{
			ResourceID: "postgres", Kind: "SERVICE", Owner: "verifier", Action: "stop",
		})

		svcMat := mat
		svcMat.ServiceRefs = nil // count mismatch
		if err := knowledge.EvaluateE2EAdmission(svcBrief, report, svcMat, state); err == nil {
			t.Error("expected admission denied when service_refs count does not match brief services count")
		}

		svcMat.ServiceRefs = []string{"postgres=postgres:15-alpine"} // version mismatch
		if err := knowledge.EvaluateE2EAdmission(svcBrief, report, svcMat, state); err == nil {
			t.Error("expected admission denied when service ref version does not match brief image/version")
		}
	}

	// 2. Command not a genuine Go test command
	{
		nonGoMat := mat
		nonGoMat.Command = []string{"bash", "-c", "echo -count=1"}
		nonGoBrief := brief
		nonGoBrief.Command = nonGoMat.Command
		if err := knowledge.EvaluateE2EAdmission(nonGoBrief, report, nonGoMat, state); err == nil {
			t.Error("expected admission denied when command executable is not go")
		}

		nonTestMat := mat
		nonTestMat.Command = []string{"go", "run", "-count=1", "main.go"}
		nonTestBrief := brief
		nonTestBrief.Command = nonTestMat.Command
		if err := knowledge.EvaluateE2EAdmission(nonTestBrief, report, nonTestMat, state); err == nil {
			t.Error("expected admission denied when subcommand is not test")
		}
	}

	// 3. Command lacks distinct -count=1 flag (mere substring)
	{
		substrMat := mat
		substrMat.Command = []string{"go", "test", "-tags=-count=1", "./..."}
		substrBrief := brief
		substrBrief.Command = substrMat.Command
		if err := knowledge.EvaluateE2EAdmission(substrBrief, report, substrMat, state); err == nil {
			t.Error("expected admission denied when -count=1 is a mere flag substring")
		}
	}
}

func TestValidateE2EReport_RunnerLossMappings(t *testing.T) {
	t.Parallel()

	brief := validTestBrief()
	mat := validTestMaterial(brief)
	policyDigest := knowledge.ComputePolicyDigest(brief.CredentialPolicy)

	// 1. Runner loss before build: NOT_STARTED -> NOT_RUN, attestation NOT_PROVISIONED
	{
		report := validTestReport(brief)
		report.StageOutcome = "RUNNER_LOST"
		report.BuildOutcome = "NOT_RUN"
		report.ExecutionOutcome = "NOT_RUN"
		report.ScenariosSelected = 3
		report.ScenariosPassed = 0
		report.ScenariosFailed = 0
		report.ScenariosSkipped = 0
		report.ScenariosNotRun = 3
		report.ScenariosUnknown = 0
		report.ScenarioResults = []knowledge.E2EScenarioResult{
			{ScenarioID: "SC-01", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-02", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-03", Outcome: "NOT_RUN"},
		}
		report.ResultCompleteness = "NONE"
		report.TestedArtifactRef = "NONE"
		report.ExecutionMaterialRef = "UNAVAILABLE"
		report.EvidenceURI = "UNAVAILABLE"
		report.SecretsAttestation = "NOT_PROVISIONED sha256:" + policyDigest
		report.CleanupEvidenceURI = "NOT_REQUIRED"
		report.Diagnostic = "Runner lost before build commenced."

		state := knowledge.E2EReportValidationContext{
			RunIDUnique:                     true,
			CredentialAuthorizationVerified: true,
			PriorCleanupResolved:            true,
			PreflightOutcome:                "READY",
			BuildProcessState:               "NOT_STARTED",
			ExecutionProcessState:           "NOT_STARTED",
			ExecutionStarted:                false,
			ServiceStartupAttempted:         false,
			CredentialProvisioningState:     "NOT_STARTED",
			ResourcesAllocated:              false,
			CleanupRequired:                 false,
			CleanupStatus:                   "CONFIRMED",
			CleanupRecordVerified:           false,
			Finalization: knowledge.E2EReportFinalizationContext{
				State:                   "FIRST_FINALIZATION",
				Authority:               "CONTROLLER_RUNNER_LOST",
				IncomingReportDigest:    knowledge.ComputeReportDigest(report),
				TerminalReportImmutable: true,
			},
		}

		if err := knowledge.ValidateE2EReport(brief, report, state); err != nil {
			t.Fatalf("expected valid pre-build RUNNER_LOST report to pass: %v", err)
		}
		if err := knowledge.EvaluateE2EAdmission(brief, report, mat, state); err == nil {
			t.Error("expected EvaluateE2EAdmission to deny admission for RUNNER_LOST report")
		}
	}

	// 2. Runner loss during build: build RUNNING -> UNKNOWN, execution NOT_STARTED -> NOT_RUN
	{
		report := validTestReport(brief)
		report.StageOutcome = "RUNNER_LOST"
		report.BuildOutcome = "UNKNOWN"
		report.ExecutionOutcome = "NOT_RUN"
		report.ScenariosSelected = 3
		report.ScenariosPassed = 0
		report.ScenariosFailed = 0
		report.ScenariosSkipped = 0
		report.ScenariosNotRun = 3
		report.ScenariosUnknown = 0
		report.ScenarioResults = []knowledge.E2EScenarioResult{
			{ScenarioID: "SC-01", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-02", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-03", Outcome: "NOT_RUN"},
		}
		report.ResultCompleteness = "NONE"
		report.TestedArtifactRef = "NONE"
		report.ExecutionMaterialRef = "UNAVAILABLE"
		report.EvidenceURI = "UNAVAILABLE"
		report.SecretsAttestation = "NOT_PROVISIONED sha256:" + policyDigest
		report.CleanupEvidenceURI = "NOT_REQUIRED"
		report.Diagnostic = "Runner lost during binary compilation."

		state := knowledge.E2EReportValidationContext{
			RunIDUnique:                     true,
			CredentialAuthorizationVerified: true,
			PriorCleanupResolved:            true,
			PreflightOutcome:                "READY",
			BuildProcessState:               "RUNNING",
			ExecutionProcessState:           "NOT_STARTED",
			ExecutionStarted:                false,
			ServiceStartupAttempted:         false,
			CredentialProvisioningState:     "NOT_STARTED",
			ResourcesAllocated:              false,
			AllocatedResourceIDs:            nil,
			CleanupRequired:                 false,
			CleanupStatus:                   "CONFIRMED",
			Finalization: knowledge.E2EReportFinalizationContext{
				State:                   "FIRST_FINALIZATION",
				Authority:               "CONTROLLER_RUNNER_LOST",
				IncomingReportDigest:    knowledge.ComputeReportDigest(report),
				TerminalReportImmutable: true,
			},
		}

		if err := knowledge.ValidateE2EReport(brief, report, state); err != nil {
			t.Fatalf("expected valid in-build RUNNER_LOST report to pass: %v", err)
		}
	}

	// 3. Runner loss during execution: execution RUNNING -> UNKNOWN, in-flight scenario UNKNOWN, attestation UNVERIFIED:RUNNER_LOST
	{
		report := validTestReport(brief)
		report.StageOutcome = "RUNNER_LOST"
		report.BuildOutcome = "PASS"
		report.ExecutionOutcome = "UNKNOWN"
		report.ScenariosSelected = 3
		report.ScenariosPassed = 1
		report.ScenariosFailed = 0
		report.ScenariosSkipped = 0
		report.ScenariosNotRun = 1
		report.ScenariosUnknown = 1
		report.ScenarioResults = []knowledge.E2EScenarioResult{
			{ScenarioID: "SC-01", Outcome: "PASS"},
			{ScenarioID: "SC-02", Outcome: "UNKNOWN"},
			{ScenarioID: "SC-03", Outcome: "NOT_RUN"},
		}
		report.ResultCompleteness = "PARTIAL"
		report.SecretsAttestation = "UNVERIFIED:RUNNER_LOST"
		report.Diagnostic = "Runner terminated abruptly during scenario execution."

		state := knowledge.E2EReportValidationContext{
			RunIDUnique:                     true,
			CredentialAuthorizationVerified: true,
			PriorCleanupResolved:            true,
			PreflightOutcome:                "READY",
			BuildProcessState:               "PASS",
			ExecutionProcessState:           "RUNNING",
			ExecutionStarted:                true,
			ServiceStartupAttempted:         false,
			CredentialProvisioningState:     "UNKNOWN",
			ResourcesAllocated:              true,
			AllocatedResourceIDs:            []string{"test-data"},
			CleanupRequired:                 true,
			CleanupStatus:                   "CONFIRMED",
			CleanupRecordVerified:           true,
			Finalization: knowledge.E2EReportFinalizationContext{
				State:                   "FIRST_FINALIZATION",
				Authority:               "CONTROLLER_RUNNER_LOST",
				IncomingReportDigest:    knowledge.ComputeReportDigest(report),
				TerminalReportImmutable: true,
			},
		}

		if err := knowledge.ValidateE2EReport(brief, report, state); err != nil {
			t.Fatalf("expected valid in-execution RUNNER_LOST report to pass: %v", err)
		}
		if err := knowledge.EvaluateE2EAdmission(brief, report, mat, state); err == nil {
			t.Error("expected EvaluateE2EAdmission to deny admission for in-execution RUNNER_LOST")
		}
	}
}

func TestValidateE2EReport_FinalizationReplayAndConflicts(t *testing.T) {
	t.Parallel()

	brief := validTestBrief()
	report := validTestReport(brief)
	reportDigest := knowledge.ComputeReportDigest(report)

	// 1. Success: FIRST_FINALIZATION with VERIFIER authority
	{
		state := validTestReportState(report)
		state.Finalization = knowledge.E2EReportFinalizationContext{
			State:                   "FIRST_FINALIZATION",
			Authority:               "VERIFIER",
			IncomingReportDigest:    reportDigest,
			TerminalReportImmutable: true,
		}
		if err := knowledge.ValidateE2EReport(brief, report, state); err != nil {
			t.Fatalf("expected valid FIRST_FINALIZATION to pass: %v", err)
		}
	}

	// 2. Negative: FIRST_FINALIZATION with CONTROLLER_RUNNER_LOST for normal non-RUNNER_LOST report
	{
		state := validTestReportState(report)
		state.Finalization = knowledge.E2EReportFinalizationContext{
			State:                   "FIRST_FINALIZATION",
			Authority:               "CONTROLLER_RUNNER_LOST",
			IncomingReportDigest:    reportDigest,
			TerminalReportImmutable: true,
		}
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error when non-RUNNER_LOST report uses CONTROLLER_RUNNER_LOST authority")
		}
	}

	// 3. Success: DUPLICATE_IDENTICAL replay
	{
		state := validTestReportState(report)
		state.Finalization = knowledge.E2EReportFinalizationContext{
			State:                           "DUPLICATE_IDENTICAL",
			Authority:                       "VERIFIER",
			ExistingTerminalReportAuthority: "VERIFIER",
			ExistingTerminalReportDigest:    reportDigest,
			IncomingReportDigest:            reportDigest,
			TerminalReportImmutable:         true,
		}
		if err := knowledge.ValidateE2EReport(brief, report, state); err != nil {
			t.Fatalf("expected idempotent DUPLICATE_IDENTICAL replay to pass: %v", err)
		}
	}

	// 4. Negative: DUPLICATE_IDENTICAL with mismatched digest
	{
		state := validTestReportState(report)
		state.Finalization = knowledge.E2EReportFinalizationContext{
			State:                           "DUPLICATE_IDENTICAL",
			Authority:                       "VERIFIER",
			ExistingTerminalReportAuthority: "VERIFIER",
			ExistingTerminalReportDigest:    "0000000000000000000000000000000000000000000000000000000000000000",
			IncomingReportDigest:            reportDigest,
			TerminalReportImmutable:         true,
		}
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error for DUPLICATE_IDENTICAL digest mismatch")
		}
	}

	// 5. Negative: DUPLICATE_IDENTICAL with mismatched authority
	{
		state := validTestReportState(report)
		state.Finalization = knowledge.E2EReportFinalizationContext{
			State:                           "DUPLICATE_IDENTICAL",
			Authority:                       "VERIFIER",
			ExistingTerminalReportAuthority: "CONTROLLER_RUNNER_LOST",
			ExistingTerminalReportDigest:    reportDigest,
			IncomingReportDigest:            reportDigest,
			TerminalReportImmutable:         true,
		}
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error for DUPLICATE_IDENTICAL authority mismatch")
		}
	}

	// 6. Negative: CONFLICTING state
	{
		state := validTestReportState(report)
		state.Finalization = knowledge.E2EReportFinalizationContext{
			State:                   "CONFLICTING",
			Authority:               "VERIFIER",
			IncomingReportDigest:    reportDigest,
			TerminalReportImmutable: true,
		}
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error for CONFLICTING finalization state")
		}
	}

	// 7. Negative: LATE_AFTER_FINAL state
	{
		state := validTestReportState(report)
		state.Finalization = knowledge.E2EReportFinalizationContext{
			State:                   "LATE_AFTER_FINAL",
			Authority:               "VERIFIER",
			IncomingReportDigest:    reportDigest,
			TerminalReportImmutable: true,
		}
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error for LATE_AFTER_FINAL finalization state")
		}
	}

	// 8. Negative: TerminalReportImmutable == false
	{
		state := validTestReportState(report)
		state.Finalization = knowledge.E2EReportFinalizationContext{
			State:                   "FIRST_FINALIZATION",
			Authority:               "VERIFIER",
			IncomingReportDigest:    reportDigest,
			TerminalReportImmutable: false,
		}
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error when TerminalReportImmutable is false")
		}
	}
}

func TestValidateE2EReport_CredentialProvisioningStates(t *testing.T) {
	t.Parallel()

	brief := validTestBrief()
	policyDigest := knowledge.ComputePolicyDigest(brief.CredentialPolicy)

	// 1. NOT_STARTED: requires NOT_PROVISIONED sha256:<hex>
	{
		report := validTestReport(brief)
		report.StageOutcome = "ENV_BLOCKED"
		report.BuildOutcome = "ENV_ERROR"
		report.ExecutionOutcome = "NOT_RUN"
		report.ScenariosPassed = 0
		report.ScenariosFailed = 0
		report.ScenariosSkipped = 0
		report.ScenariosNotRun = 3
		report.ScenariosUnknown = 0
		report.ResultCompleteness = "NONE"
		report.TestedArtifactRef = "NONE"
		report.ExecutionMaterialRef = "UNAVAILABLE"
		report.EvidenceURI = "UNAVAILABLE"
		report.CleanupEvidenceURI = "NOT_REQUIRED"
		report.SecretsAttestation = "NOT_PROVISIONED sha256:" + policyDigest
		report.ScenarioResults = []knowledge.E2EScenarioResult{
			{ScenarioID: "SC-01", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-02", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-03", Outcome: "NOT_RUN"},
		}

		state := validTestReportState(report)
		state.BuildProcessState = "ENV_ERROR"
		state.ExecutionProcessState = "NOT_STARTED"
		state.ExecutionStarted = false
		state.CredentialProvisioningState = "NOT_STARTED"
		state.ResourcesAllocated = false
		state.AllocatedResourceIDs = nil
		state.CleanupRequired = false
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)

		if err := knowledge.ValidateE2EReport(brief, report, state); err != nil {
			t.Fatalf("expected valid NOT_STARTED provisioning report to pass: %v", err)
		}
	}

	// 2. Empty policy: NOT_STARTED maps to NOT_PROVISIONED; SUCCEEDED and FAILED are rejected
	{
		emptyBrief := brief
		emptyBrief.CredentialPolicy = []knowledge.E2ECredentialPolicyEntry{}
		report := validTestReport(emptyBrief)
		emptyDigest := knowledge.ComputePolicyDigest(emptyBrief.CredentialPolicy)
		report.SecretsAttestation = "NOT_PROVISIONED sha256:" + emptyDigest
		state := validTestReportState(report)
		state.CredentialProvisioningState = "NOT_STARTED"
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)

		// Positive: empty-policy with NOT_STARTED passes
		if err := knowledge.ValidateE2EReport(emptyBrief, report, state); err != nil {
			t.Fatalf("expected valid empty-policy NOT_STARTED report to pass: %v", err)
		}

		// Negative: empty-policy with SUCCEEDED must be rejected
		badSucceededState := state
		badSucceededState.CredentialProvisioningState = "SUCCEEDED"
		if err := knowledge.ValidateE2EReport(emptyBrief, report, badSucceededState); err == nil {
			t.Error("expected error when CredentialPolicy is empty but CredentialProvisioningState is SUCCEEDED")
		}

		// Negative: empty-policy with FAILED must be rejected
		badFailedState := state
		badFailedState.CredentialProvisioningState = "FAILED"
		if err := knowledge.ValidateE2EReport(emptyBrief, report, badFailedState); err == nil {
			t.Error("expected error when CredentialPolicy is empty but CredentialProvisioningState is FAILED")
		}
	}

	// 3. FAILED: requires ENV_BLOCKED:PROVISIONING_FAILED with non-empty policy
	{
		report := validTestReport(brief)
		report.StageOutcome = "ENV_BLOCKED"
		report.BuildOutcome = "NOT_RUN"
		report.ExecutionOutcome = "ENV_ERROR"
		report.ScenariosPassed = 0
		report.ScenariosFailed = 0
		report.ScenariosSkipped = 0
		report.ScenariosNotRun = 3
		report.ScenariosUnknown = 0
		report.ResultCompleteness = "NONE"
		report.TestedArtifactRef = "NONE"
		report.ExecutionMaterialRef = "UNAVAILABLE"
		report.EvidenceURI = "UNAVAILABLE"
		report.SecretsAttestation = "ENV_BLOCKED:PROVISIONING_FAILED"
		report.ScenarioResults = []knowledge.E2EScenarioResult{
			{ScenarioID: "SC-01", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-02", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-03", Outcome: "NOT_RUN"},
		}

		state := validTestReportState(report)
		state.BuildProcessState = "NOT_STARTED"
		state.ExecutionProcessState = "ENV_ERROR"
		state.ExecutionStarted = false
		state.CredentialProvisioningState = "FAILED"
		state.ResourcesAllocated = false
		state.AllocatedResourceIDs = nil
		state.CleanupRequired = true
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)

		if err := knowledge.ValidateE2EReport(brief, report, state); err != nil {
			t.Fatalf("expected valid FAILED provisioning report to pass: %v", err)
		}

		// Negative: FAILED with empty policy must fail
		emptyBrief := brief
		emptyBrief.CredentialPolicy = []knowledge.E2ECredentialPolicyEntry{}
		if err := knowledge.ValidateE2EReport(emptyBrief, report, state); err == nil {
			t.Error("expected error when CredentialProvisioningState is FAILED but CredentialPolicy is empty")
		}
	}

	// 4. UNKNOWN: requires UNVERIFIED:RUNNER_LOST and StageOutcome RUNNER_LOST
	{
		report := validTestReport(brief)
		report.StageOutcome = "RUNNER_LOST"
		report.BuildOutcome = "PASS"
		report.ExecutionOutcome = "UNKNOWN"
		report.ScenariosPassed = 1
		report.ScenariosFailed = 0
		report.ScenariosSkipped = 0
		report.ScenariosUnknown = 1
		report.ScenariosNotRun = 1
		report.ScenarioResults = []knowledge.E2EScenarioResult{
			{ScenarioID: "SC-01", Outcome: "PASS"},
			{ScenarioID: "SC-02", Outcome: "UNKNOWN"},
			{ScenarioID: "SC-03", Outcome: "NOT_RUN"},
		}
		report.ResultCompleteness = "PARTIAL"
		report.SecretsAttestation = "UNVERIFIED:RUNNER_LOST"

		state := validTestReportState(report)
		state.BuildProcessState = "PASS"
		state.ExecutionProcessState = "RUNNING"
		state.CredentialProvisioningState = "UNKNOWN"
		state.Finalization = knowledge.E2EReportFinalizationContext{
			State:                   "FIRST_FINALIZATION",
			Authority:               "CONTROLLER_RUNNER_LOST",
			IncomingReportDigest:    knowledge.ComputeReportDigest(report),
			TerminalReportImmutable: true,
		}

		if err := knowledge.ValidateE2EReport(brief, report, state); err != nil {
			t.Fatalf("expected valid UNKNOWN provisioning report with RUNNER_LOST to pass: %v", err)
		}

		// Negative: UNKNOWN provisioning with stage outcome PASS must fail
		badReport := report
		badReport.StageOutcome = "PASS"
		badState := state
		badState.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(badReport)
		if err := knowledge.ValidateE2EReport(brief, badReport, badState); err == nil {
			t.Error("expected error when CredentialProvisioningState is UNKNOWN but StageOutcome is PASS")
		}
	}

	// 5. Positive non-empty throwaway policy: SUCCEEDED maps to ATTESTED:THROWAWAY
	{
		throwawayBrief := brief
		throwawayBrief.CredentialPolicy = []knowledge.E2ECredentialPolicyEntry{
			{Name: "ephemeral-token", GenerationMode: "THROWAWAY", GenerationInstruction: "create token"},
		}
		throwawayBrief.CredentialAuthorization = nil
		report := validTestReport(throwawayBrief)
		state := validTestReportState(report)
		state.CredentialProvisioningState = "SUCCEEDED"
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)

		if err := knowledge.ValidateE2EReport(throwawayBrief, report, state); err != nil {
			t.Fatalf("expected valid throwaway SUCCEEDED provisioning report to pass: %v", err)
		}
	}

	// 6. Positive non-empty mixed policy: SUCCEEDED maps to ATTESTED:MIXED
	{
		mixedBrief := brief
		mixedBrief.CredentialPolicy = []knowledge.E2ECredentialPolicyEntry{
			{Name: "db-secret", GenerationMode: "REAL_ONLY"},
		}
		mixedBrief.CredentialAuthorization = []knowledge.E2ECredentialAuthorization{
			{
				Name:             "db-secret",
				Purpose:          "Integration",
				Lifetime:         "SINGLE_RUN",
				AuthorizedBy:     "operator",
				AuthorizedAt:     "2026-10-12T20:00:00Z",
				AuthorizationRef: "auth-mixed-1",
			},
		}
		report := validTestReport(mixedBrief)
		state := validTestReportState(report)
		state.CredentialProvisioningState = "SUCCEEDED"
		state.CredentialAuthorizationVerified = true
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)

		if err := knowledge.ValidateE2EReport(mixedBrief, report, state); err != nil {
			t.Fatalf("expected valid mixed SUCCEEDED provisioning report to pass: %v", err)
		}
	}
}

func TestValidateE2EReport_TruthfulLifecycleFactsAndCompleteness(t *testing.T) {
	t.Parallel()

	brief := validTestBrief()

	// 1. RUNNER_LOST report when all scenario results are durable and passing
	{
		report := validTestReport(brief)
		report.StageOutcome = "RUNNER_LOST"
		report.BuildOutcome = "PASS"
		report.ExecutionOutcome = "PASS"
		report.ScenariosPassed = len(brief.Scenarios)
		report.ScenariosFailed = 0
		report.ScenariosSkipped = 0
		report.ScenariosNotRun = 0
		report.ScenariosUnknown = 0
		report.ScenarioResults = []knowledge.E2EScenarioResult{
			{ScenarioID: "SC-01", Outcome: "PASS"},
			{ScenarioID: "SC-02", Outcome: "PASS"},
			{ScenarioID: "SC-03", Outcome: "PASS"},
		}
		report.ResultCompleteness = "COMPLETE"
		report.CleanupStatus = "CONFIRMED"

		state := validTestReportState(report)
		state.BuildProcessState = "PASS"
		state.ExecutionProcessState = "PASS"
		state.ExecutionStarted = true
		state.CleanupStatus = "CONFIRMED"
		state.Finalization.Authority = "CONTROLLER_RUNNER_LOST"
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)

		// Report validation must succeed (truthful durable facts preserved)
		if err := knowledge.ValidateE2EReport(brief, report, state); err != nil {
			t.Fatalf("expected structurally consistent RUNNER_LOST with passing scenarios to pass: %v", err)
		}

		// Admission must fail closed
		mat := validTestMaterial(brief)
		if err := knowledge.EvaluateE2EAdmission(brief, report, mat, state); err == nil {
			t.Error("expected EvaluateE2EAdmission to deny RUNNER_LOST even with passing scenarios")
		} else if !strings.Contains(err.Error(), "RUNNER_LOST") {
			t.Errorf("unexpected admission error: %v", err)
		}
	}

	// 2. RUNNER_LOST report when execution never started after successful build
	{
		report := validTestReport(brief)
		report.StageOutcome = "RUNNER_LOST"
		report.BuildOutcome = "PASS"
		report.ExecutionOutcome = "NOT_RUN"
		report.ScenariosPassed = 0
		report.ScenariosFailed = 0
		report.ScenariosSkipped = 0
		report.ScenariosNotRun = len(brief.Scenarios)
		report.ScenariosUnknown = 0
		report.ScenarioResults = []knowledge.E2EScenarioResult{
			{ScenarioID: "SC-01", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-02", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-03", Outcome: "NOT_RUN"},
		}
		report.ResultCompleteness = "NONE"
		report.CleanupStatus = "CONFIRMED"

		state := validTestReportState(report)
		state.BuildProcessState = "PASS"
		state.ExecutionProcessState = "NOT_STARTED"
		state.ExecutionStarted = false
		state.CleanupStatus = "CONFIRMED"
		state.Finalization.Authority = "CONTROLLER_RUNNER_LOST"
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)

		if err := knowledge.ValidateE2EReport(brief, report, state); err != nil {
			t.Fatalf("expected RUNNER_LOST after successful build before execution start to pass: %v", err)
		}

		mat := validTestMaterial(brief)
		if err := knowledge.EvaluateE2EAdmission(brief, report, mat, state); err == nil {
			t.Error("expected EvaluateE2EAdmission to deny non-started execution RUNNER_LOST")
		}
	}

	// 3. TIMEOUT with all scenario states known (ResultCompleteness COMPLETE)
	{
		report := validTestReport(brief)
		report.StageOutcome = "TIMEOUT"
		report.BuildOutcome = "PASS"
		report.ExecutionOutcome = "TIMEOUT"
		report.ScenariosPassed = 2
		report.ScenariosFailed = 1
		report.ScenariosSkipped = 0
		report.ScenariosNotRun = 0
		report.ScenariosUnknown = 0
		report.ScenarioResults = []knowledge.E2EScenarioResult{
			{ScenarioID: "SC-01", Outcome: "PASS"},
			{ScenarioID: "SC-02", Outcome: "PASS"},
			{ScenarioID: "SC-03", Outcome: "FAIL"},
		}
		report.ResultCompleteness = "COMPLETE"

		state := validTestReportState(report)
		state.BuildProcessState = "PASS"
		state.ExecutionProcessState = "TIMEOUT"
		state.ExecutionStarted = true
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)

		if err := knowledge.ValidateE2EReport(brief, report, state); err != nil {
			t.Fatalf("expected TIMEOUT with COMPLETE completeness to pass: %v", err)
		}

		mat := validTestMaterial(brief)
		if err := knowledge.EvaluateE2EAdmission(brief, report, mat, state); err == nil {
			t.Error("expected EvaluateE2EAdmission to deny TIMEOUT report")
		}
	}

	// 4. CANCELLED with all scenario states known (ResultCompleteness COMPLETE)
	{
		report := validTestReport(brief)
		report.StageOutcome = "CANCELLED"
		report.BuildOutcome = "PASS"
		report.ExecutionOutcome = "CANCELLED"
		report.ScenariosPassed = 1
		report.ScenariosFailed = 0
		report.ScenariosSkipped = 0
		report.ScenariosNotRun = 2
		report.ScenariosUnknown = 0
		report.ScenarioResults = []knowledge.E2EScenarioResult{
			{ScenarioID: "SC-01", Outcome: "PASS"},
			{ScenarioID: "SC-02", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-03", Outcome: "NOT_RUN"},
		}
		report.ResultCompleteness = "COMPLETE"

		state := validTestReportState(report)
		state.BuildProcessState = "PASS"
		state.ExecutionProcessState = "CANCELLED"
		state.ExecutionStarted = true
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)

		if err := knowledge.ValidateE2EReport(brief, report, state); err != nil {
			t.Fatalf("expected CANCELLED with COMPLETE completeness to pass: %v", err)
		}

		mat := validTestMaterial(brief)
		if err := knowledge.EvaluateE2EAdmission(brief, report, mat, state); err == nil {
			t.Error("expected EvaluateE2EAdmission to deny CANCELLED report")
		}
	}
}

func TestValidateE2EAdmission_CredentialProvisioningRequirements(t *testing.T) {
	t.Parallel()

	brief := validTestBrief()
	mat := validTestMaterial(brief)

	// 1. Nonempty policy with NOT_STARTED passes report validation (with NOT_PROVISIONED attestation) but is denied at admission
	{
		report := validTestReport(brief)
		report.SecretsAttestation = "NOT_PROVISIONED sha256:" + knowledge.ComputePolicyDigest(brief.CredentialPolicy)
		state := validTestReportState(report)
		state.CredentialProvisioningState = "NOT_STARTED"
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)

		if err := knowledge.EvaluateE2EAdmission(brief, report, mat, state); err == nil {
			t.Error("expected EvaluateE2EAdmission to reject credential state NOT_STARTED for nonempty policy")
		} else if !strings.Contains(err.Error(), "SUCCEEDED required for non-empty credential policy") {
			t.Errorf("unexpected error for state NOT_STARTED: %v", err)
		}
	}

	// 2. Empty policy with NOT_STARTED is admitted
	{
		emptyBrief := validTestBrief()
		emptyBrief.CredentialPolicy = []knowledge.E2ECredentialPolicyEntry{}
		emptyReport := validTestReport(emptyBrief)
		emptyReport.SecretsAttestation = "NOT_PROVISIONED sha256:" + knowledge.ComputePolicyDigest(emptyBrief.CredentialPolicy)
		emptyState := validTestReportState(emptyReport)
		emptyState.CredentialProvisioningState = "NOT_STARTED"
		emptyState.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(emptyReport)
		emptyMat := validTestMaterial(emptyBrief)

		if err := knowledge.EvaluateE2EAdmission(emptyBrief, emptyReport, emptyMat, emptyState); err != nil {
			t.Fatalf("expected empty credential policy with NOT_STARTED to pass admission: %v", err)
		}
	}

	// 3. Empty policy with SUCCEEDED is rejected
	{
		emptyBrief := validTestBrief()
		emptyBrief.CredentialPolicy = []knowledge.E2ECredentialPolicyEntry{}
		emptyReport := validTestReport(emptyBrief)
		emptyReport.SecretsAttestation = "NOT_PROVISIONED sha256:" + knowledge.ComputePolicyDigest(emptyBrief.CredentialPolicy)
		emptyState := validTestReportState(emptyReport)
		emptyState.CredentialProvisioningState = "SUCCEEDED"
		emptyState.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(emptyReport)
		emptyMat := validTestMaterial(emptyBrief)

		if err := knowledge.EvaluateE2EAdmission(emptyBrief, emptyReport, emptyMat, emptyState); err == nil {
			t.Error("expected EvaluateE2EAdmission to reject credential state SUCCEEDED for empty policy")
		}
	}
}

func TestValidateE2EReport_AuthorizationDeniedNameMatching(t *testing.T) {
	t.Parallel()

	brief := validTestBrief()
	brief.CredentialPolicy = []knowledge.E2ECredentialPolicyEntry{
		{Name: "db-secret", GenerationMode: "REAL_ONLY"},
	}
	// Mismatched name: authorization has "api-token", policy has "db-secret"
	brief.CredentialAuthorization = []knowledge.E2ECredentialAuthorization{
		{
			Name:             "api-token",
			Purpose:          "Integration",
			Lifetime:         "SINGLE_RUN",
			AuthorizedBy:     "operator",
			AuthorizedAt:     "2026-10-12T20:00:00Z",
			AuthorizationRef: "auth-1",
		},
	}

	report := validTestReport(brief)
	report.StageOutcome = "ENV_BLOCKED"
	report.BuildOutcome = "NOT_RUN"
	report.ExecutionOutcome = "ENV_ERROR"
	report.ResultCompleteness = "NONE"
	report.TestedArtifactRef = "NONE"
	report.ExecutionMaterialRef = "UNAVAILABLE"
	report.EvidenceURI = "UNAVAILABLE"
	report.CleanupStatus = "CONFIRMED"
	report.CleanupEvidenceURI = "NOT_REQUIRED"
	report.SecretsAttestation = "ENV_BLOCKED:AUTHORIZATION_DENIED"
	report.ScenariosPassed = 0
	report.ScenariosFailed = 0
	report.ScenariosSkipped = 0
	report.ScenariosUnknown = 0
	report.ScenariosNotRun = len(brief.Scenarios)
	report.ScenarioResults = []knowledge.E2EScenarioResult{
		{ScenarioID: "SC-01", Outcome: "NOT_RUN"},
		{ScenarioID: "SC-02", Outcome: "NOT_RUN"},
		{ScenarioID: "SC-03", Outcome: "NOT_RUN"},
	}

	state := validTestReportState(report)
	state.PreflightOutcome = "AUTHORIZATION_DENIED"
	state.CredentialAuthorizationVerified = true
	state.BuildProcessState = "NOT_STARTED"
	state.ExecutionProcessState = "NOT_STARTED"
	state.ExecutionStarted = false
	state.ServiceStartupAttempted = false
	state.CredentialProvisioningState = "NOT_STARTED"
	state.ResourcesAllocated = false
	state.AllocatedResourceIDs = nil
	state.CleanupRequired = false
	state.CleanupStatus = "CONFIRMED"
	state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)

	// Name mismatch should be rejected
	if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
		t.Error("expected error when authorization name does not match REAL_ONLY credential name")
	} else if !strings.Contains(err.Error(), "does not match declared REAL_ONLY credential") {
		t.Errorf("unexpected error: %v", err)
	}

	// Correct matching name passes
	brief.CredentialAuthorization[0].Name = "db-secret"
	state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)
	if err := knowledge.ValidateE2EReport(brief, report, state); err != nil {
		t.Fatalf("expected valid AUTHORIZATION_DENIED report with matching name to pass: %v", err)
	}
}

func TestValidateE2EBrief_ServicesNilVsEmpty(t *testing.T) {
	t.Parallel()

	state := validTestBriefState()
	b := validTestBrief()

	b.Services = nil
	if err := knowledge.ValidateE2EBrief(b, state); err == nil {
		t.Error("expected error for brief with Services == nil")
	} else if !strings.Contains(err.Error(), "services is required (use explicit empty list [] if none)") {
		t.Errorf("unexpected error: %v", err)
	}

	b.Services = []knowledge.E2EServiceRequirement{}
	if err := knowledge.ValidateE2EBrief(b, state); err != nil {
		t.Errorf("expected brief with Services == [] to pass: %v", err)
	}
}

func TestValidateGoTestCommand_EffectiveFlagSemantics(t *testing.T) {
	t.Parallel()

	briefState := validTestBriefState()

	// 1. Non-test command
	{
		b := validTestBrief()
		b.Command = []string{"pytest", "test", "cases"}
		if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
			t.Error("expected error for non-go test command")
		} else if !strings.Contains(err.Error(), "test command executable must be 'go'") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 2. Go subcommand not 'test'
	{
		b := validTestBrief()
		b.Command = []string{"go", "run", "main.go"}
		if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
			t.Error("expected error for go run")
		} else if !strings.Contains(err.Error(), "test command subcommand must be 'test'") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 3. Missing -count=1 flag
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "./..."}
		if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
			t.Error("expected error for go test without -count=1")
		} else if !strings.Contains(err.Error(), "must include distinct -count=1 flag") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 4. Later overriding -count value (-count=1 ... -count=2)
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "-count=1", "./...", "-count=2"}
		if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
			t.Error("expected error for overridden -count flag")
		} else if !strings.Contains(err.Error(), "effective -count flag value must be '1'") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 5. Flags after -args that mimic test flags
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "./...", "-args", "-count=1"}
		if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
			t.Error("expected error when -count=1 appears only after -args")
		} else if !strings.Contains(err.Error(), "must include distinct -count=1 flag before -args") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 6. Valid -count=1 before -args with unrelated valid forwarded flags after -args
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-v", "--env=staging", "custom-arg"}
		if err := knowledge.ValidateE2EBrief(b, briefState); err != nil {
			t.Errorf("expected valid ordinary command with unrelated forwarded flags to pass: %v", err)
		}
	}

	// 7. Forwarded -test.count=1 maintains repeat protection
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-test.count=1"}
		if err := knowledge.ValidateE2EBrief(b, briefState); err != nil {
			t.Errorf("expected forwarded -test.count=1 to pass: %v", err)
		}
	}

	// 8. Forwarded -count=2 override rejected at brief preflight
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-count=2"}
		if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
			t.Error("expected error for forwarded -count=2 override")
		} else if !strings.Contains(err.Error(), "overrides effective test count") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 9. Forwarded -test.count=2 override rejected at brief preflight
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-test.count=2"}
		if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
			t.Error("expected error for forwarded -test.count=2 override")
		} else if !strings.Contains(err.Error(), "overrides effective test count") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 10. Forwarded -test.count 2 (two args) rejected at brief preflight
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-test.count", "2"}
		if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
			t.Error("expected error for forwarded -test.count 2 override")
		} else if !strings.Contains(err.Error(), "overrides effective test count") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 11. Forwarded -count=0 override rejected at brief preflight
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-count=0"}
		if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
			t.Error("expected error for forwarded -count=0 override")
		} else if !strings.Contains(err.Error(), "overrides effective test count") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 12. Forwarded test flags after '--' are treated as positional arguments
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "-count=1", "./...", "-args", "--", "-test.count=2"}
		if err := knowledge.ValidateE2EBrief(b, briefState); err != nil {
			t.Errorf("expected -test.count=2 after '--' to be treated as positional and pass: %v", err)
		}
	}

	// 13. Valid -count 1 (two arguments)
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "-count", "1", "./..."}
		if err := knowledge.ValidateE2EBrief(b, briefState); err != nil {
			t.Errorf("expected valid -count 1 to pass: %v", err)
		}
	}

	// 14. EvaluateE2EAdmission enforces driver-side check
	{
		b := validTestBrief()
		report := validTestReport(b)
		state := validTestReportState(report)
		mat := validTestMaterial(b)
		mat.Command = []string{"go", "test", "-count=1", "./...", "-count=0"}

		if err := knowledge.EvaluateE2EAdmission(b, report, mat, state); err == nil {
			t.Error("expected EvaluateE2EAdmission to reject overridden -count=0")
		} else if !strings.Contains(err.Error(), "effective -count flag value must be '1'") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 15. EvaluateE2EAdmission enforces forwarded test-count override check
	{
		b := validTestBrief()
		report := validTestReport(b)
		state := validTestReportState(report)
		mat := validTestMaterial(b)
		mat.Command = []string{"go", "test", "-count=1", "./...", "-args", "-test.count=2"}

		if err := knowledge.EvaluateE2EAdmission(b, report, mat, state); err == nil {
			t.Error("expected EvaluateE2EAdmission to reject forwarded -test.count=2")
		} else if !strings.Contains(err.Error(), "overrides effective test count") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 16. EvaluateE2EAdmission accepts unrelated forwarded flags
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-v", "--env=staging"}
		report := validTestReport(b)
		state := validTestReportState(report)
		mat := validTestMaterial(b)
		mat.Command = b.Command

		if err := knowledge.EvaluateE2EAdmission(b, report, mat, state); err != nil {
			t.Errorf("expected EvaluateE2EAdmission to accept unrelated forwarded flags: %v", err)
		}
	}

	// 17. Known native value-taking flag consumes value even if beginning with '-' at brief preflight
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-test.outputdir", "-test.count=2"}
		if err := knowledge.ValidateE2EBrief(b, briefState); err != nil {
			t.Errorf("expected -test.outputdir consuming -test.count=2 as directory value to pass: %v", err)
		}
	}

	// 18. Known native value-taking flag consumes non-dash value, then rejects subsequent count override
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-test.outputdir", "outdir", "-test.count=2"}
		if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
			t.Error("expected error for count override following -test.outputdir outdir")
		} else if !strings.Contains(err.Error(), "overrides effective test count to \"2\"") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 19. Known native value-taking flag missing required argument at end of command
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-test.outputdir"}
		if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
			t.Error("expected error for -test.outputdir missing argument")
		} else if !strings.Contains(err.Error(), "requires an argument") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 20. First positional argument stops flag parsing, preserving subsequent tokens as non-flags at brief preflight
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "-count=1", "./...", "-args", "my_target_arg", "-test.count=2"}
		if err := knowledge.ValidateE2EBrief(b, briefState); err != nil {
			t.Errorf("expected positional arg to stop flag parsing and pass: %v", err)
		}
	}

	// 21. Direct count override before positional argument is rejected at brief preflight
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-test.count=2", "my_target_arg"}
		if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
			t.Error("expected error for direct count override before positional arg")
		} else if !strings.Contains(err.Error(), "overrides effective test count to \"2\"") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 22. Known native value-taking flag consuming '-' value is admitted at EvaluateE2EAdmission
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-test.outputdir", "-test.count=2"}
		report := validTestReport(b)
		state := validTestReportState(report)
		mat := validTestMaterial(b)
		mat.Command = b.Command

		if err := knowledge.EvaluateE2EAdmission(b, report, mat, state); err != nil {
			t.Errorf("expected EvaluateE2EAdmission to admit count-preserving -test.outputdir form: %v", err)
		}
	}

	// 23. Positional argument stopping flag parsing is admitted at EvaluateE2EAdmission
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "-count=1", "./...", "-args", "my_target_arg", "-test.count=2"}
		report := validTestReport(b)
		state := validTestReportState(report)
		mat := validTestMaterial(b)
		mat.Command = b.Command

		if err := knowledge.EvaluateE2EAdmission(b, report, mat, state); err != nil {
			t.Errorf("expected EvaluateE2EAdmission to admit count-preserving positional arg form: %v", err)
		}
	}

	// 24. Direct count override before positional argument is rejected at EvaluateE2EAdmission
	{
		b := validTestBrief()
		b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-test.count=2", "my_target_arg"}
		report := validTestReport(b)
		state := validTestReportState(report)
		mat := validTestMaterial(b)
		mat.Command = b.Command

		if err := knowledge.EvaluateE2EAdmission(b, report, mat, state); err == nil {
			t.Error("expected EvaluateE2EAdmission to reject direct count override before positional arg")
		} else if !strings.Contains(err.Error(), "overrides effective test count to \"2\"") {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

func TestValidateE2EReport_LosslessSerializationAndControlCharacters(t *testing.T) {
	t.Parallel()

	brief := validTestBrief()

	// 1. Control character / newline in RunID
	{
		b := brief
		b.RunID = "run-1\n"
		report := validTestReport(b)
		report.RunID = "run-1\n"
		state := validTestReportState(report)
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)

		if err := knowledge.ValidateE2EReport(b, report, state); err == nil {
			t.Error("expected error for newline in RunID")
		} else if !strings.Contains(err.Error(), "field run_id contains illegal control character") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 2. Control character in ScenarioID
	{
		report := validTestReport(brief)
		report.ScenarioResults[0].ScenarioID = "SC-01\x00"
		state := validTestReportState(report)
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)

		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error for control char in ScenarioID")
		} else if !strings.Contains(err.Error(), "contains illegal control character") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 3. Invalid UTF-8 in TestPackage
	{
		b := brief
		b.TestPackage = "e2e/\xff\xfe"
		report := validTestReport(b)
		report.TestPackage = "e2e/\xff\xfe"
		state := validTestReportState(report)
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)

		if err := knowledge.ValidateE2EReport(b, report, state); err == nil {
			t.Error("expected error for invalid UTF-8 in TestPackage")
		} else if !strings.Contains(err.Error(), "contains invalid UTF-8") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 4. Lossless serialization
	{
		report := validTestReport(brief)
		formatted := knowledge.FormatE2EReport(report)
		if !strings.Contains(formatted, "run_id: "+report.RunID) {
			t.Errorf("FormatE2EReport missing exact RunID: %s", formatted)
		}
		if !strings.Contains(formatted, "candidate_ref: "+report.CandidateRef) {
			t.Errorf("FormatE2EReport missing exact CandidateRef: %s", formatted)
		}
	}
}

func TestValidateE2EReport_FinalizationAuthorityClosedEnum(t *testing.T) {
	t.Parallel()

	brief := validTestBrief()
	report := validTestReport(brief)

	// 1. FIRST_FINALIZATION with unknown authority
	{
		state := validTestReportState(report)
		state.Finalization.Authority = "UNKNOWN"
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)

		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error for unknown authority in FIRST_FINALIZATION")
		} else if !strings.Contains(err.Error(), "invalid finalization authority") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 2. DUPLICATE_IDENTICAL where both incoming and stored are identical unknown authority
	{
		state := validTestReportState(report)
		state.Finalization.State = "DUPLICATE_IDENTICAL"
		state.Finalization.Authority = "UNKNOWN"
		state.Finalization.ExistingTerminalReportAuthority = "UNKNOWN"
		state.Finalization.ExistingTerminalReportDigest = knowledge.ComputeReportDigest(report)
		state.Finalization.IncomingReportDigest = state.Finalization.ExistingTerminalReportDigest

		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error for identical unknown authority in DUPLICATE_IDENTICAL")
		} else if !strings.Contains(err.Error(), "invalid finalization authority") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 3. DUPLICATE_IDENTICAL where ExistingTerminalReportAuthority is unknown
	{
		state := validTestReportState(report)
		state.Finalization.State = "DUPLICATE_IDENTICAL"
		state.Finalization.Authority = "VERIFIER"
		state.Finalization.ExistingTerminalReportAuthority = "UNKNOWN"
		state.Finalization.ExistingTerminalReportDigest = knowledge.ComputeReportDigest(report)
		state.Finalization.IncomingReportDigest = state.Finalization.ExistingTerminalReportDigest

		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error for unknown ExistingTerminalReportAuthority")
		} else if !strings.Contains(err.Error(), "invalid existing terminal report authority") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 4. DUPLICATE_IDENTICAL with CONTROLLER_RUNNER_LOST for both passes
	{
		runnerLostReport := report
		runnerLostReport.StageOutcome = "RUNNER_LOST"
		state := validTestReportState(runnerLostReport)
		state.Finalization.State = "DUPLICATE_IDENTICAL"
		state.Finalization.Authority = "CONTROLLER_RUNNER_LOST"
		state.Finalization.ExistingTerminalReportAuthority = "CONTROLLER_RUNNER_LOST"
		digest := knowledge.ComputeReportDigest(runnerLostReport)
		state.Finalization.ExistingTerminalReportDigest = digest
		state.Finalization.IncomingReportDigest = digest

		if err := knowledge.ValidateE2EReport(brief, runnerLostReport, state); err != nil {
			t.Fatalf("expected valid DUPLICATE_IDENTICAL with CONTROLLER_RUNNER_LOST to pass: %v", err)
		}
	}
}

func TestValidateE2E_PlannerRemediationRegressions(t *testing.T) {
	t.Parallel()

	// 1. Provider-denial report exception:
	// Preflight AUTHORIZATION_DENIED keeps both process states NOT_STARTED,
	// allows approved report shape (ENV_BLOCKED, NOT_RUN, ENV_ERROR, NOT_RUN, NONE, attestation ENV_BLOCKED:AUTHORIZATION_DENIED).
	// Validates structurally but never admits. Mismatched auth names or non-NOT_STARTED process states rejected.
	{
		brief := validTestBrief()
		brief.CredentialPolicy = []knowledge.E2ECredentialPolicyEntry{
			{Name: "prod-db-token", GenerationMode: "REAL_ONLY"},
		}
		brief.CredentialAuthorization = []knowledge.E2ECredentialAuthorization{
			{
				Name:             "prod-db-token",
				Purpose:          "Integration test",
				Lifetime:         "SINGLE_RUN",
				AuthorizedBy:     "operator",
				AuthorizedAt:     "2026-10-12T20:00:00Z",
				AuthorizationRef: "auth-12345",
			},
		}
		report := validTestReport(brief)
		report.StageOutcome = "ENV_BLOCKED"
		report.BuildOutcome = "NOT_RUN"
		report.ExecutionOutcome = "ENV_ERROR"
		report.ResultCompleteness = "NONE"
		report.TestedArtifactRef = "NONE"
		report.ExecutionMaterialRef = "UNAVAILABLE"
		report.EvidenceURI = "UNAVAILABLE"
		report.CleanupStatus = "CONFIRMED"
		report.CleanupEvidenceURI = "NOT_REQUIRED"
		report.SecretsAttestation = "ENV_BLOCKED:AUTHORIZATION_DENIED"
		report.ScenariosPassed = 0
		report.ScenariosFailed = 0
		report.ScenariosSkipped = 0
		report.ScenariosNotRun = report.ScenariosSelected
		report.ScenariosUnknown = 0
		report.ScenarioResults = []knowledge.E2EScenarioResult{
			{ScenarioID: "SC-01", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-02", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-03", Outcome: "NOT_RUN"},
		}

		state := validTestReportState(report)
		state.PreflightOutcome = "AUTHORIZATION_DENIED"
		state.CredentialAuthorizationVerified = true
		state.BuildProcessState = "NOT_STARTED"
		state.ExecutionProcessState = "NOT_STARTED"
		state.ExecutionStarted = false
		state.ServiceStartupAttempted = false
		state.CredentialProvisioningState = "NOT_STARTED"
		state.ResourcesAllocated = false
		state.AllocatedResourceIDs = nil
		state.CleanupRequired = false
		state.CleanupStatus = "CONFIRMED"
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)

		// Positive: validates structurally
		if err := knowledge.ValidateE2EReport(brief, report, state); err != nil {
			t.Fatalf("expected valid AUTHORIZATION_DENIED report with both process states NOT_STARTED to pass: %v", err)
		}

		// Negative: admission MUST be denied
		mat := validTestMaterial(brief)
		if err := knowledge.EvaluateE2EAdmission(brief, report, mat, state); err == nil {
			t.Error("expected EvaluateE2EAdmission to reject AUTHORIZATION_DENIED preflight")
		} else if !strings.Contains(err.Error(), "preflight AUTHORIZATION_DENIED cannot be admitted") {
			t.Errorf("unexpected admission error: %v", err)
		}

		// Negative: Fabricated ExecutionProcessState == "ENV_ERROR" must be rejected
		badState := state
		badState.ExecutionProcessState = "ENV_ERROR"
		if err := knowledge.ValidateE2EReport(brief, report, badState); err == nil {
			t.Error("expected error when AUTHORIZATION_DENIED has execution_process_state ENV_ERROR")
		}

		// Positive: Runner-loss after durable AUTHORIZATION_DENIED preflight preserves blocked attestation and NOT_STARTED process states
		runnerLostReport := report
		runnerLostReport.StageOutcome = "RUNNER_LOST"
		runnerLostReport.BuildOutcome = "NOT_RUN"
		runnerLostReport.ExecutionOutcome = "NOT_RUN"                            // unstarted phase preserved as NOT_RUN
		runnerLostReport.SecretsAttestation = "ENV_BLOCKED:AUTHORIZATION_DENIED" // known authorization block preserved
		runnerLostState := state
		runnerLostState.Finalization.Authority = "CONTROLLER_RUNNER_LOST"
		runnerLostState.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(runnerLostReport)

		if err := knowledge.ValidateE2EReport(brief, runnerLostReport, runnerLostState); err != nil {
			t.Fatalf("expected valid runner-loss report after AUTHORIZATION_DENIED to pass: %v", err)
		}

		// Negative: Runner-loss after AUTHORIZATION_DENIED is strictly non-admissible
		if err := knowledge.EvaluateE2EAdmission(brief, runnerLostReport, mat, runnerLostState); err == nil {
			t.Error("expected admission to be denied for runner-loss after AUTHORIZATION_DENIED")
		}

		// Negative: Fabricating ExecutionOutcome == ENV_ERROR under RUNNER_LOST must be rejected (must be NOT_RUN for unstarted phase)
		badLossReport := runnerLostReport
		badLossReport.ExecutionOutcome = "ENV_ERROR"
		badLossState := runnerLostState
		badLossState.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(badLossReport)
		if err := knowledge.ValidateE2EReport(brief, badLossReport, badLossState); err == nil {
			t.Error("expected error when RUNNER_LOST after AUTHORIZATION_DENIED fabricates ExecutionOutcome ENV_ERROR")
		} else if !strings.Contains(err.Error(), "requires execution_outcome NOT_RUN") {
			t.Errorf("unexpected error: %v", err)
		}

		// Negative: Erasing known authorization block under RUNNER_LOST must be rejected
		badAttestReport := runnerLostReport
		badAttestReport.SecretsAttestation = "UNVERIFIED:RUNNER_LOST"
		badAttestState := runnerLostState
		badAttestState.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(badAttestReport)
		if err := knowledge.ValidateE2EReport(brief, badAttestReport, badAttestState); err == nil {
			t.Error("expected error when RUNNER_LOST after AUTHORIZATION_DENIED erases authorization block")
		} else if !strings.Contains(err.Error(), "requires secrets_attestation ENV_BLOCKED:AUTHORIZATION_DENIED") {
			t.Errorf("unexpected error: %v", err)
		}

		// Negative: Normal AUTHORIZATION_DENIED attempting to use ExecutionOutcome == NOT_RUN (must be ENV_ERROR)
		badNormalReport := report
		badNormalReport.ExecutionOutcome = "NOT_RUN"
		badNormalState := state
		badNormalState.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(badNormalReport)
		if err := knowledge.ValidateE2EReport(brief, badNormalReport, badNormalState); err == nil {
			t.Error("expected error when normal AUTHORIZATION_DENIED report uses ExecutionOutcome NOT_RUN")
		} else if !strings.Contains(err.Error(), "requires execution_outcome ENV_ERROR") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 2. Go test command effective-argv validation:
	{
		briefState := validTestBriefState()

		// Positive: -test.count 1 alias
		{
			b := validTestBrief()
			b.Command = []string{"go", "test", "-test.count", "1", "./..."}
			if err := knowledge.ValidateE2EBrief(b, briefState); err != nil {
				t.Errorf("expected go test -test.count 1 to pass: %v", err)
			}
		}

		// Positive: -count=2 followed by overriding -test.count=1
		{
			b := validTestBrief()
			b.Command = []string{"go", "test", "-count=2", "./...", "-test.count=1"}
			if err := knowledge.ValidateE2EBrief(b, briefState); err != nil {
				t.Errorf("expected go test with overriding -test.count=1 to pass: %v", err)
			}
		}

		// Negative: -run consumes -count=1 as regex pattern
		{
			b := validTestBrief()
			b.Command = []string{"go", "test", "-run", "-count=1", "./..."}
			if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
				t.Error("expected error when -run flag consumes -count=1")
			} else if !strings.Contains(err.Error(), "must include distinct -count=1 flag") {
				t.Errorf("unexpected error: %v", err)
			}
		}

		// Negative: unknown flag without '=' is rejected as ambiguous
		{
			b := validTestBrief()
			b.Command = []string{"go", "test", "-unknown_custom_flag", "arg", "-count=1", "./..."}
			if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
				t.Error("expected error for ambiguous unknown flag without '='")
			} else if !strings.Contains(err.Error(), "ambiguous or unrecognized flag") {
				t.Errorf("unexpected error: %v", err)
			}
		}

		// Negative: '--' terminates flag parsing, so -count=1 afterwards is a package argument
		{
			b := validTestBrief()
			b.Command = []string{"go", "test", "--", "-count=1"}
			if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
				t.Error("expected error when -count=1 appears after '--'")
			} else if !strings.Contains(err.Error(), "must include distinct -count=1 flag") {
				t.Errorf("unexpected error: %v", err)
			}
		}

		// Positive: unrelated forwarded flags after -args preserved at brief preflight
		{
			b := validTestBrief()
			b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-v", "--flag=val", "myarg"}
			if err := knowledge.ValidateE2EBrief(b, briefState); err != nil {
				t.Errorf("expected unrelated forwarded flags after -args to pass: %v", err)
			}
		}

		// Positive: forwarded -test.count=1 maintains repeat protection at brief preflight
		{
			b := validTestBrief()
			b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-test.count=1"}
			if err := knowledge.ValidateE2EBrief(b, briefState); err != nil {
				t.Errorf("expected forwarded -test.count=1 to pass: %v", err)
			}
		}

		// Negative: forwarded -test.count=2 rejected at brief preflight
		{
			b := validTestBrief()
			b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-test.count=2"}
			if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
				t.Error("expected error when forwarded -test.count=2 is passed")
			} else if !strings.Contains(err.Error(), "overrides effective test count to \"2\", disabling repeat protection") {
				t.Errorf("unexpected error: %v", err)
			}
		}

		// Negative: forwarded -test.count 2 (space-separated) rejected at brief preflight
		{
			b := validTestBrief()
			b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-test.count", "2"}
			if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
				t.Error("expected error when forwarded -test.count 2 is passed")
			} else if !strings.Contains(err.Error(), "overrides effective test count to \"2\", disabling repeat protection") {
				t.Errorf("unexpected error: %v", err)
			}
		}

		// Negative: forwarded -count=0 rejected at brief preflight
		{
			b := validTestBrief()
			b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-count=0"}
			if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
				t.Error("expected error when forwarded -count=0 is passed")
			} else if !strings.Contains(err.Error(), "overrides effective test count to \"0\", disabling repeat protection") {
				t.Errorf("unexpected error: %v", err)
			}
		}

		// Positive: unrelated forwarded flags accepted at admission
		{
			b := validTestBrief()
			b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-v", "--env=staging"}
			report := validTestReport(b)
			state := validTestReportState(report)
			mat := validTestMaterial(b)
			mat.Command = b.Command
			if err := knowledge.EvaluateE2EAdmission(b, report, mat, state); err != nil {
				t.Errorf("expected admission to accept unrelated forwarded flags: %v", err)
			}
		}

		// Negative: forwarded -test.count=2 rejected at admission
		{
			b := validTestBrief()
			report := validTestReport(b)
			state := validTestReportState(report)
			mat := validTestMaterial(b)
			mat.Command = []string{"go", "test", "-count=1", "./...", "-args", "-test.count=2"}
			if err := knowledge.EvaluateE2EAdmission(b, report, mat, state); err == nil {
				t.Error("expected admission to reject forwarded -test.count=2 override")
			} else if !strings.Contains(err.Error(), "overrides effective test count to \"2\", disabling repeat protection") {
				t.Errorf("unexpected error: %v", err)
			}
		}

		// Positive: known value-taking flag consuming '-' value passes at brief preflight and admission
		{
			b := validTestBrief()
			b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-test.outputdir", "-test.count=2"}
			if err := knowledge.ValidateE2EBrief(b, briefState); err != nil {
				t.Errorf("expected -test.outputdir consuming -test.count=2 to pass at brief preflight: %v", err)
			}

			report := validTestReport(b)
			state := validTestReportState(report)
			mat := validTestMaterial(b)
			mat.Command = b.Command
			if err := knowledge.EvaluateE2EAdmission(b, report, mat, state); err != nil {
				t.Errorf("expected -test.outputdir consuming -test.count=2 to pass at admission: %v", err)
			}
		}

		// Positive: positional argument stopping flag parsing passes at brief preflight and admission
		{
			b := validTestBrief()
			b.Command = []string{"go", "test", "-count=1", "./...", "-args", "my_positional_target", "-test.count=2"}
			if err := knowledge.ValidateE2EBrief(b, briefState); err != nil {
				t.Errorf("expected positional arg stopping flag parsing to pass at brief preflight: %v", err)
			}

			report := validTestReport(b)
			state := validTestReportState(report)
			mat := validTestMaterial(b)
			mat.Command = b.Command
			if err := knowledge.EvaluateE2EAdmission(b, report, mat, state); err != nil {
				t.Errorf("expected positional arg stopping flag parsing to pass at admission: %v", err)
			}
		}

		// Negative: direct count override before positional arg rejected at brief preflight and admission
		{
			b := validTestBrief()
			b.Command = []string{"go", "test", "-count=1", "./...", "-args", "-test.count=2", "my_positional_target"}
			if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
				t.Error("expected direct count override before positional arg to fail at brief preflight")
			} else if !strings.Contains(err.Error(), "overrides effective test count to \"2\"") {
				t.Errorf("unexpected error: %v", err)
			}

			report := validTestReport(b)
			state := validTestReportState(report)
			mat := validTestMaterial(b)
			mat.Command = b.Command
			if err := knowledge.EvaluateE2EAdmission(b, report, mat, state); err == nil {
				t.Error("expected direct count override before positional arg to fail at admission")
			} else if !strings.Contains(err.Error(), "overrides effective test count to \"2\"") {
				t.Errorf("unexpected error: %v", err)
			}
		}
	}

	// 3. Runner loss preserves durable terminal process outcomes:
	{
		brief := validTestBrief()
		report := validTestReport(brief)
		report.StageOutcome = "RUNNER_LOST"
		report.BuildOutcome = "PASS"
		report.ExecutionOutcome = "PASS"
		state := validTestReportState(report)
		state.BuildProcessState = "PASS"
		state.ExecutionProcessState = "PASS"
		state.Finalization.Authority = "CONTROLLER_RUNNER_LOST"
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)

		// Positive: preserves durable execution outcome PASS
		if err := knowledge.ValidateE2EReport(brief, report, state); err != nil {
			t.Fatalf("expected RUNNER_LOST report with durable PASS process outcome to pass: %v", err)
		}

		// Negative: mutating known PASS to INVALID_EVIDENCE is rejected
		mutatedReport := report
		mutatedReport.ExecutionOutcome = "INVALID_EVIDENCE"
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(mutatedReport)
		if err := knowledge.ValidateE2EReport(brief, mutatedReport, state); err == nil {
			t.Error("expected error when report mutates known PASS process outcome to INVALID_EVIDENCE")
		} else if !strings.Contains(err.Error(), "does not match durable execution_process_state PASS") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 4. Cleanup obligation bounds:
	{
		brief := validTestBrief()
		report := validTestReport(brief)
		state := validTestReportState(report)

		// Negative: ExecutionStarted == true but CleanupRequired == false is rejected
		state.ExecutionStarted = true
		state.CleanupRequired = false
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error when ExecutionStarted is true but CleanupRequired is false")
		} else if !strings.Contains(err.Error(), "requires cleanup_required to be true") {
			t.Errorf("unexpected error: %v", err)
		}

		// Negative: ResourcesAllocated == true but CleanupRequired == false is rejected
		state.ExecutionStarted = false
		state.BuildProcessState = "ENV_ERROR"
		state.ExecutionProcessState = "NOT_STARTED"
		report.StageOutcome = "ENV_BLOCKED"
		report.BuildOutcome = "ENV_ERROR"
		report.ExecutionOutcome = "NOT_RUN"
		report.ResultCompleteness = "NONE"
		report.TestedArtifactRef = "NONE"
		report.EvidenceURI = "UNAVAILABLE"
		report.CleanupEvidenceURI = "NOT_REQUIRED"
		report.ScenariosPassed = 0
		report.ScenariosNotRun = len(brief.Scenarios)
		report.ScenarioResults = []knowledge.E2EScenarioResult{
			{ScenarioID: "SC-01", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-02", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-03", Outcome: "NOT_RUN"},
		}
		state.ResourcesAllocated = true
		state.AllocatedResourceIDs = []string{"res-1"}
		state.CleanupRequired = false
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error when ResourcesAllocated is true but CleanupRequired is false")
		} else if !strings.Contains(err.Error(), "requires cleanup_required to be true") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 5. Credential-policy shape: nil rejected at all boundaries; [] admitted
	{
		b := validTestBrief()
		report := validTestReport(b)
		mat := validTestMaterial(b)
		state := validTestReportState(report)

		bNil := b
		bNil.CredentialPolicy = nil

		// Brief boundary
		briefState := validTestBriefState()
		if err := knowledge.ValidateE2EBrief(bNil, briefState); err == nil {
			t.Error("expected ValidateE2EBrief to reject nil CredentialPolicy")
		} else if !strings.Contains(err.Error(), "credential_policy is required") {
			t.Errorf("unexpected error: %v", err)
		}

		// Report boundary
		if err := knowledge.ValidateE2EReport(bNil, report, state); err == nil {
			t.Error("expected ValidateE2EReport to reject nil CredentialPolicy")
		} else if !strings.Contains(err.Error(), "credential_policy is required (use explicit empty list [] if none)") {
			t.Errorf("unexpected error: %v", err)
		}

		// Admission boundary
		if err := knowledge.EvaluateE2EAdmission(bNil, report, mat, state); err == nil {
			t.Error("expected EvaluateE2EAdmission to reject nil CredentialPolicy")
		} else if !strings.Contains(err.Error(), "credential_policy is required (use explicit empty list [] if none)") {
			t.Errorf("unexpected error: %v", err)
		}

		// Positive: explicit empty slice [] admitted with empty-policy digest
		bEmpty := b
		bEmpty.CredentialPolicy = []knowledge.E2ECredentialPolicyEntry{}
		rEmpty := validTestReport(bEmpty)
		emptyDigest := knowledge.ComputePolicyDigest(bEmpty.CredentialPolicy)
		rEmpty.SecretsAttestation = "NOT_PROVISIONED sha256:" + emptyDigest
		sEmpty := validTestReportState(rEmpty)
		sEmpty.CredentialProvisioningState = "NOT_STARTED"
		sEmpty.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(rEmpty)
		mEmpty := validTestMaterial(bEmpty)

		if err := knowledge.ValidateE2EBrief(bEmpty, briefState); err != nil {
			t.Fatalf("expected ValidateE2EBrief to accept empty slice []: %v", err)
		}
		if err := knowledge.ValidateE2EReport(bEmpty, rEmpty, sEmpty); err != nil {
			t.Fatalf("expected ValidateE2EReport to accept empty slice []: %v", err)
		}
		if err := knowledge.EvaluateE2EAdmission(bEmpty, rEmpty, mEmpty, sEmpty); err != nil {
			t.Fatalf("expected EvaluateE2EAdmission to admit empty slice []: %v", err)
		}
	}

	// 6. Credential attestation precedence & generator blocks:
	{
		emptyBrief := validTestBrief()
		emptyBrief.CredentialPolicy = []knowledge.E2ECredentialPolicyEntry{}
		report := validTestReport(emptyBrief)
		state := validTestReportState(report)
		state.CredentialProvisioningState = "NOT_STARTED"

		// Negative: empty policy claiming GENERATOR_UNAVAILABLE is rejected
		report.SecretsAttestation = "ENV_BLOCKED:GENERATOR_UNAVAILABLE"
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)
		if err := knowledge.ValidateE2EReport(emptyBrief, report, state); err == nil {
			t.Error("expected error for empty policy with GENERATOR_UNAVAILABLE attestation")
		} else if !strings.Contains(err.Error(), "empty credential policy requires secrets_attestation") {
			t.Errorf("unexpected error: %v", err)
		}

		// Positive: non-empty policy with durable generator block preserved
		nonEmptyBrief := validTestBrief()
		genReport := validTestReport(nonEmptyBrief)
		genReport.StageOutcome = "ENV_BLOCKED"
		genReport.BuildOutcome = "ENV_ERROR"
		genReport.ExecutionOutcome = "NOT_RUN"
		genReport.ResultCompleteness = "NONE"
		genReport.TestedArtifactRef = "NONE"
		genReport.ExecutionMaterialRef = "UNAVAILABLE"
		genReport.EvidenceURI = "UNAVAILABLE"
		genReport.CleanupEvidenceURI = "NOT_REQUIRED"
		genReport.SecretsAttestation = "ENV_BLOCKED:GENERATOR_UNAVAILABLE"
		genReport.ScenariosPassed = 0
		genReport.ScenariosFailed = 0
		genReport.ScenariosSkipped = 0
		genReport.ScenariosNotRun = len(nonEmptyBrief.Scenarios)
		genReport.ScenariosUnknown = 0
		genReport.ScenarioResults = []knowledge.E2EScenarioResult{
			{ScenarioID: "SC-01", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-02", Outcome: "NOT_RUN"},
			{ScenarioID: "SC-03", Outcome: "NOT_RUN"},
		}
		genState := validTestReportState(genReport)
		genState.BuildProcessState = "ENV_ERROR"
		genState.ExecutionProcessState = "NOT_STARTED"
		genState.ExecutionStarted = false
		genState.CredentialProvisioningState = "NOT_STARTED"
		genState.ResourcesAllocated = false
		genState.AllocatedResourceIDs = nil
		genState.CleanupRequired = false
		genState.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(genReport)

		if err := knowledge.ValidateE2EReport(nonEmptyBrief, genReport, genState); err != nil {
			t.Fatalf("expected non-empty policy with durable generator block to pass: %v", err)
		}
	}

	// 7. Identity congruence at brief preflight:
	{
		briefState := validTestBriefState()

		// RunID with Unicode NEL
		b := validTestBrief()
		b.RunID = "run\u0085id"
		if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
			t.Error("expected ValidateE2EBrief to reject RunID with Unicode NEL")
		} else if !strings.Contains(err.Error(), "contains illegal control character") {
			t.Errorf("unexpected error: %v", err)
		}

		// Scenario ID with ASCII control character
		b = validTestBrief()
		b.Scenarios[0].ID = "SC\x0001"
		if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
			t.Error("expected ValidateE2EBrief to reject scenario ID with control char")
		} else if !strings.Contains(err.Error(), "contains illegal control character") {
			t.Errorf("unexpected error: %v", err)
		}

		// Scenario ID with Unicode Line Separator
		b = validTestBrief()
		b.Scenarios[0].ID = "SC\u202801"
		if err := knowledge.ValidateE2EBrief(b, briefState); err == nil {
			t.Error("expected ValidateE2EBrief to reject scenario ID with Line Separator")
		} else if !strings.Contains(err.Error(), "contains illegal control character") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 8. Single-line diagnostics & control / line-separator characters:
	{
		brief := validTestBrief()

		// Unicode NEL (\u0085) in Diagnostic
		report := validTestReport(brief)
		report.Diagnostic = "Test failed\u0085with error"
		state := validTestReportState(report)
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error for Unicode NEL in Diagnostic")
		} else if !strings.Contains(err.Error(), "field diagnostic contains illegal control character") {
			t.Errorf("unexpected error: %v", err)
		}

		// Unicode Line Separator (\u2028) in Diagnostic
		report.Diagnostic = "Test failed\u2028with error"
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error for Unicode Line Separator in Diagnostic")
		} else if !strings.Contains(err.Error(), "field diagnostic contains illegal control character") {
			t.Errorf("unexpected error: %v", err)
		}

		// Unicode Paragraph Separator (\u2029) in Diagnostic
		report.Diagnostic = "Test failed\u2029with error"
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error for Unicode Paragraph Separator in Diagnostic")
		} else if !strings.Contains(err.Error(), "field diagnostic contains illegal control character") {
			t.Errorf("unexpected error: %v", err)
		}

		// ASCII LF (\n) in Diagnostic
		report.Diagnostic = "Test failed\nwith error"
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)
		if err := knowledge.ValidateE2EReport(brief, report, state); err == nil {
			t.Error("expected error for ASCII newline in Diagnostic")
		} else if !strings.Contains(err.Error(), "field diagnostic contains illegal control character") {
			t.Errorf("unexpected error: %v", err)
		}

		// Positive: Printable UTF-8 non-ASCII characters pass without alteration
		report.Diagnostic = "测试诊断信息: 所有测试用例执行完毕"
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)
		if err := knowledge.ValidateE2EReport(brief, report, state); err != nil {
			t.Fatalf("expected printable UTF-8 diagnostic to pass: %v", err)
		}
	}

	// 10. Deterministic validation error order:
	{
		brief := validTestBrief()
		b := brief
		b.RunID = "bad\x01id"

		report := validTestReport(b)
		report.RunID = "bad\x01id"
		report.Diagnostic = "bad\x02diag"
		state := validTestReportState(report)
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)

		err := knowledge.ValidateE2EReport(b, report, state)
		if err == nil {
			t.Fatal("expected error for multiple invalid fields")
		}
		if !strings.Contains(err.Error(), "field run_id contains illegal control character") {
			t.Errorf("expected deterministic error on run_id first, got: %v", err)
		}

		// Both diagnostic and evidence_uri contain control character: diagnostic must be reported first
		report = validTestReport(brief)
		report.Diagnostic = "bad\x02diag"
		report.EvidenceURI = "bad\x03uri"
		state = validTestReportState(report)
		state.Finalization.IncomingReportDigest = knowledge.ComputeReportDigest(report)

		err = knowledge.ValidateE2EReport(brief, report, state)
		if err == nil {
			t.Fatal("expected error for multiple invalid fields")
		}
		if !strings.Contains(err.Error(), "field diagnostic contains illegal control character") {
			t.Errorf("expected deterministic error on diagnostic first, got: %v", err)
		}
	}
}
