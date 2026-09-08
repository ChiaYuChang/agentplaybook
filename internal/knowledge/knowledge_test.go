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
	if len(roles) != 7 {
		t.Fatalf("expected 7 roles, got %d", len(roles))
	}
	for _, expected := range []string{"planner", "builder", "reviewer", "scout", "navigator", "cartographer", "verifier"} {
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

	// Verify blueprint flow has 12 steps and deterministic routing
	blueprintFlow, ok := k.Flow("blueprint")
	if !ok {
		t.Fatalf("expected blueprint flow to exist")
	}
	if len(blueprintFlow.Steps) != 12 {
		t.Fatalf("expected 12 steps in blueprint flow, got %d", len(blueprintFlow.Steps))
	}
	bpStep2, _ := k.FlowStep("blueprint", 2)
	bpStep2Conditions := make(map[string]int)
	for _, c := range bpStep2.Conditions {
		bpStep2Conditions[c.When] = c.Then
	}
	if bpStep2Conditions["BLUEPRINT_PASS"] != 3 || bpStep2Conditions["BLUEPRINT_REJECT"] != 1 {
		t.Errorf("unexpected blueprint step 2 conditions: %v", bpStep2Conditions)
	}
	bpStep6, _ := k.FlowStep("blueprint", 6)
	bpStep6Conditions := make(map[string]int)
	for _, c := range bpStep6.Conditions {
		bpStep6Conditions[c.When] = c.Then
	}
	if bpStep6Conditions["SUBPLAN_REVIEW_FINDINGS"] != 7 || bpStep6Conditions["SUBPLAN_REVIEW_SATISFIED"] != 8 || bpStep6Conditions["BLUEPRINT_REVIEW_REQUIRED"] != 2 || bpStep6Conditions["BASELINE_STALE"] != 2 {
		t.Errorf("unexpected blueprint step 6 conditions: %v", bpStep6Conditions)
	}
	bpStep9, _ := k.FlowStep("blueprint", 9)
	bpStep9Conditions := make(map[string]int)
	for _, c := range bpStep9.Conditions {
		bpStep9Conditions[c.When] = c.Then
	}
	if bpStep9Conditions["SUBPLAN_REVIEW_PASS_MORE_SUBPLANS"] != 3 || bpStep9Conditions["SUBPLAN_REVIEW_PASS_ALL_COMPLETED"] != 10 || bpStep9Conditions["RESOLUTION_REJECTED"] != 8 {
		t.Errorf("unexpected blueprint step 9 conditions: %v", bpStep9Conditions)
	}
	bpStep11, _ := k.FlowStep("blueprint", 11)
	bpStep11Conditions := make(map[string]int)
	for _, c := range bpStep11.Conditions {
		bpStep11Conditions[c.When] = c.Then
	}
	if bpStep11Conditions["FEATURE_REVIEW_PASS"] != 12 || bpStep11Conditions["FEATURE_REVIEW_REJECT"] != 3 || bpStep11Conditions["DEPENDENT_EVIDENCE_STALE"] != 2 {
		t.Errorf("unexpected blueprint step 11 conditions: %v", bpStep11Conditions)
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

	// Verify e2e flow (12 steps)
	e2eFlow, ok := k.Flow("e2e")
	if !ok {
		t.Fatalf("expected e2e flow to exist")
	}
	if len(e2eFlow.Steps) != 12 {
		t.Fatalf("expected 12 steps in e2e flow, got %d", len(e2eFlow.Steps))
	}
	if e2eFlow.Steps[0].Actor != knowledge.RolePlanner || e2eFlow.Steps[1].Actor != knowledge.RolePlanner {
		t.Errorf("expected e2e steps 1-2 actor to be planner")
	}
	if e2eFlow.Steps[2].Actor != knowledge.RoleVerifier || e2eFlow.Steps[3].Actor != knowledge.RoleVerifier || e2eFlow.Steps[4].Actor != knowledge.RoleVerifier {
		t.Errorf("expected e2e steps 3-5 actor to be verifier")
	}
	if e2eFlow.Steps[5].Actor != knowledge.RolePlanner || e2eFlow.Steps[6].Actor != knowledge.RolePlanner {
		t.Errorf("expected e2e steps 6-7 actor to be planner")
	}
	if e2eFlow.Steps[7].Actor != knowledge.RoleReviewer {
		t.Errorf("expected e2e step 8 actor to be reviewer")
	}
	if e2eFlow.Steps[8].Actor != knowledge.RolePlanner || e2eFlow.Steps[9].Actor != knowledge.RolePlanner || e2eFlow.Steps[10].Actor != knowledge.RolePlanner || e2eFlow.Steps[11].Actor != knowledge.RolePlanner {
		t.Errorf("expected e2e steps 9-12 actor to be planner")
	}
	if !e2eFlow.Steps[10].Terminal || len(e2eFlow.Steps[10].Conditions) != 0 {
		t.Errorf("expected e2e step 11 to be terminal: true with 0 conditions")
	}
	if !e2eFlow.Steps[11].Terminal || len(e2eFlow.Steps[11].Conditions) != 0 {
		t.Errorf("expected e2e step 12 to be terminal: true with 0 conditions")
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
		"scout-survey",
		"diagram-brief",
		"diagram-completion",
		"diagram-clarification-request",
		"e2e-brief",
		"e2e-report",
		"e2e-test-spec",
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
	if len(buildPlan.Visibility) != 3 {
		t.Errorf("expected build-plan visibility to have 3 roles, got %v", buildPlan.Visibility)
	}
	if len(buildPlan.Sections) != 7 {
		t.Errorf("expected build-plan to have 7 sections, got %d", len(buildPlan.Sections))
	}
	subBuildPlan, _ := k.Artifact("sub-build-plan")
	if len(subBuildPlan.Sections) != 7 {
		t.Errorf("expected sub-build-plan to have 7 sections, got %d", len(subBuildPlan.Sections))
	}
	reviewPlan, _ := k.Artifact("review-plan")
	if len(reviewPlan.Sections) != 6 {
		t.Errorf("expected review-plan to have 6 sections, got %d", len(reviewPlan.Sections))
	}
	subReviewPlan, _ := k.Artifact("sub-review-plan")
	if len(subReviewPlan.Sections) != 6 {
		t.Errorf("expected sub-review-plan to have 6 sections, got %d", len(subReviewPlan.Sections))
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
	if len(reviewResolution.Visibility) != 6 {
		t.Errorf("expected review-resolution visibility to include 6 roles, got %v", reviewResolution.Visibility)
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
	if len(e2eBrief.Fields) != 6 {
		t.Errorf("expected e2e-brief to have 6 fields, got %d", len(e2eBrief.Fields))
	}

	e2eReport, _ := k.Artifact("e2e-report")
	if e2eReport.Owner != knowledge.RoleVerifier || e2eReport.Type != "message" {
		t.Errorf("unexpected e2e-report metadata: %+v", e2eReport)
	}
	if len(e2eReport.Visibility) != 3 || e2eReport.Visibility[0] != knowledge.RolePlanner || e2eReport.Visibility[1] != knowledge.RoleVerifier || e2eReport.Visibility[2] != knowledge.RoleReviewer {
		t.Errorf("expected e2e-report visibility [planner verifier reviewer], got %v", e2eReport.Visibility)
	}
	if len(e2eReport.Fields) != 14 {
		t.Errorf("expected e2e-report to have 14 fields, got %d", len(e2eReport.Fields))
	}

	// Verify e2e-test-spec and e2e-clarification-request metadata
	e2eTestSpec, _ := k.Artifact("e2e-test-spec")
	if e2eTestSpec.Owner != knowledge.RoleBuilder || e2eTestSpec.Type != "document" {
		t.Errorf("unexpected e2e-test-spec metadata: %+v", e2eTestSpec)
	}
	if len(e2eTestSpec.Visibility) != 4 || e2eTestSpec.Visibility[0] != knowledge.RolePlanner || e2eTestSpec.Visibility[1] != knowledge.RoleReviewer || e2eTestSpec.Visibility[2] != knowledge.RoleBuilder || e2eTestSpec.Visibility[3] != knowledge.RoleVerifier {
		t.Errorf("expected e2e-test-spec visibility [planner reviewer builder verifier], got %v", e2eTestSpec.Visibility)
	}
	if len(e2eTestSpec.Sections) != 5 {
		t.Errorf("expected e2e-test-spec to have 5 sections, got %d", len(e2eTestSpec.Sections))
	}

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
	if len(rules) < 34 {
		t.Fatalf("expected at least 34 rules, got %d", len(rules))
	}
	for _, expected := range []string{
		"anti-cheating",
		"mandatory-alignment",
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

	// 4. Reachability test: e2e Step 11 is Terminal: true, has zero conditions, and cannot reach Step 12
	{
		e2eFlow, ok := k.Flow("e2e")
		if !ok {
			t.Fatalf("expected e2e flow to exist")
		}
		var step11 *knowledge.FlowStep
		for _, s := range e2eFlow.Steps {
			if s.Index == 11 {
				sCopy := s
				step11 = &sCopy
				break
			}
		}
		if step11 == nil {
			t.Fatalf("expected e2e step 11 to exist")
		}
		if !step11.Terminal {
			t.Errorf("expected e2e step 11 to be terminal: true")
		}
		if len(step11.Conditions) != 0 {
			t.Errorf("expected e2e step 11 to have zero conditions, got %d", len(step11.Conditions))
		}
		for _, c := range step11.Conditions {
			if c.Then == 12 {
				t.Errorf("e2e step 11 must not transition to step 12")
			}
		}
	}

	// 5. Reachability test: e2e Step 12 is Terminal: true, has zero conditions, and cannot reach Step 1, 2, or 11
	{
		e2eFlow, ok := k.Flow("e2e")
		if !ok {
			t.Fatalf("expected e2e flow to exist")
		}
		var step12 *knowledge.FlowStep
		for _, s := range e2eFlow.Steps {
			if s.Index == 12 {
				sCopy := s
				step12 = &sCopy
				break
			}
		}
		if step12 == nil {
			t.Fatalf("expected e2e step 12 to exist")
		}
		if !step12.Terminal {
			t.Errorf("expected e2e step 12 to be terminal: true")
		}
		if len(step12.Conditions) != 0 {
			t.Errorf("expected e2e step 12 to have zero conditions, got %d", len(step12.Conditions))
		}
		for _, c := range step12.Conditions {
			if c.Then == 1 || c.Then == 2 || c.Then == 11 {
				t.Errorf("e2e step 12 must not transition to step %d", c.Then)
			}
		}
	}
}

func TestValidateE2EReport(t *testing.T) {
	t.Parallel()

	validBrief := func() knowledge.E2EBriefContext {
		return knowledge.E2EBriefContext{
			RunID:        "e2e-run-20260908T220000Z-7f8a9b",
			CandidateRef: "0123456789abcdef0123456789abcdef01234567",
			TestPackage:  "e2e/core-flow",
			Mode:         "build-and-run",
			TimeoutMS:    180000,
		}
	}

	validCoverage := func() knowledge.E2EScenarioCoverageContext {
		return knowledge.E2EScenarioCoverageContext{
			SelectedScenarioIDs: []string{"SC-01", "SC-02", "SC-03"},
			SkippedScenarioIDs:  []string{"SC-03"},
			AllowedSkipIDs:      []string{"SC-03"},
		}
	}

	validReport := func() knowledge.E2EReportPayload {
		return knowledge.E2EReportPayload{
			RunID:              "e2e-run-20260908T220000Z-7f8a9b",
			CandidateRef:       "0123456789abcdef0123456789abcdef01234567",
			StageOutcome:       "PASS",
			BuildOutcome:       "PASS",
			ExecutionOutcome:   "PASS",
			ScenariosSelected:  3,
			ScenariosPassed:    2,
			ScenariosFailed:    0,
			ScenariosSkipped:   1,
			ResultCompleteness: "COMPLETE",
			TestedArtifactRef:  "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 [bin/agentplaybook]",
			DurationMS:         12500,
			Diagnostic:         "All 2 selected scenarios passed with 1 authorized skip.",
			EvidenceURI:        "file:///tmp/e2e-sandbox-7f8a9b/evidence.tar.gz",
		}
	}

	// 1. Success test: build-and-run
	{
		brief := validBrief()
		report := validReport()
		cov := validCoverage()
		if err := knowledge.ValidateE2EReport(brief, report, cov); err != nil {
			t.Fatalf("expected valid report to pass: %v", err)
		}
		// Deterministic token budget test
		tokenCount := knowledge.EstimateTokenCount(knowledge.FormatE2EReport(report))
		if tokenCount >= 150 {
			t.Errorf("expected token count < 150, got %d", tokenCount)
		}
	}

	// 2. Success test: run-only
	{
		brief := validBrief()
		brief.Mode = "run-only"
		report := validReport()
		report.BuildOutcome = "NOT_RUN"
		cov := validCoverage()
		if err := knowledge.ValidateE2EReport(brief, report, cov); err != nil {
			t.Fatalf("expected valid run-only report to pass: %v", err)
		}
	}

	// 3. Pre-provisioning failure test with UNAVAILABLE evidence and NONE artifact
	{
		brief := validBrief()
		report := validReport()
		report.StageOutcome = "ENV_BLOCKED"
		report.BuildOutcome = "ENV_ERROR"
		report.ExecutionOutcome = "NOT_RUN"
		report.ResultCompleteness = "NONE"
		report.TestedArtifactRef = "NONE"
		report.EvidenceURI = "UNAVAILABLE"
		report.ScenariosSelected = 1
		report.ScenariosPassed = 0
		report.ScenariosFailed = 0
		report.ScenariosSkipped = 1
		cov := knowledge.E2EScenarioCoverageContext{
			SelectedScenarioIDs: []string{"SC-01"},
			SkippedScenarioIDs:  []string{"SC-01"},
			AllowedSkipIDs:      []string{"SC-01"},
		}
		if err := knowledge.ValidateE2EReport(brief, report, cov); err != nil {
			t.Fatalf("expected pre-provisioning failure report to pass: %v", err)
		}
	}

	// 4. Mode mismatch: build-and-run mode with build_outcome: NOT_RUN
	{
		brief := validBrief()
		brief.Mode = "build-and-run"
		report := validReport()
		report.BuildOutcome = "NOT_RUN"
		cov := validCoverage()
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for build-and-run mode with build_outcome NOT_RUN")
		}
	}

	// 5. Zero selection: scenarios_selected: 0
	{
		brief := validBrief()
		report := validReport()
		report.ScenariosSelected = 0
		cov := validCoverage()
		cov.SelectedScenarioIDs = nil
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for scenarios_selected == 0")
		}
	}

	// 6. Count reconciliation mismatch
	{
		brief := validBrief()
		report := validReport()
		report.ScenariosSelected = 5
		report.ScenariosPassed = 3
		report.ScenariosFailed = 1
		report.ScenariosSkipped = 0 // 3+1+0 != 5
		cov := validCoverage()
		cov.SelectedScenarioIDs = []string{"S1", "S2", "S3", "S4", "S5"}
		cov.SkippedScenarioIDs = nil
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for count reconciliation mismatch")
		}
	}

	// 7. Duplicate scenario IDs in SelectedScenarioIDs or SkippedScenarioIDs
	{
		brief := validBrief()
		report := validReport()
		cov := validCoverage()
		cov.SelectedScenarioIDs = []string{"SC-01", "SC-01", "SC-02"}
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for duplicate SelectedScenarioIDs")
		}

		cov2 := validCoverage()
		cov2.SkippedScenarioIDs = []string{"SC-03", "SC-03"}
		report.ScenariosSkipped = 2
		report.ScenariosPassed = 1
		if err := knowledge.ValidateE2EReport(brief, report, cov2); err == nil {
			t.Error("expected error for duplicate SkippedScenarioIDs")
		}
	}

	// 8. Skipped scenario ID absent from SelectedScenarioIDs
	{
		brief := validBrief()
		report := validReport()
		cov := validCoverage()
		cov.SkippedScenarioIDs = []string{"SC-UNKNOWN"}
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error when skipped scenario ID is not in SelectedScenarioIDs")
		}
	}

	// 9. Unauthorized skip IDs for stage_outcome == PASS
	{
		brief := validBrief()
		report := validReport()
		cov := validCoverage()
		cov.AllowedSkipIDs = []string{"SC-OTHER"} // SC-03 is skipped but not authorized
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for unauthorized skipped scenario with PASS stage outcome")
		}
	}

	// 10. BUILD_FAILURE misclassification
	{
		// ENV_ERROR misclassified as BUILD_FAILURE
		brief := validBrief()
		report := validReport()
		report.StageOutcome = "BUILD_FAILURE"
		report.BuildOutcome = "ENV_ERROR"
		report.ExecutionOutcome = "NOT_RUN"
		report.TestedArtifactRef = "NONE"
		cov := validCoverage()
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error when build ENV_ERROR is misclassified as BUILD_FAILURE")
		}

		// TIMEOUT misclassified as BUILD_FAILURE
		report.BuildOutcome = "TIMEOUT"
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error when build TIMEOUT is misclassified as BUILD_FAILURE")
		}
	}

	// 11. Contradictory states
	{
		// BUILD_FAILURE with execution_outcome: PASS
		brief := validBrief()
		report := validReport()
		report.StageOutcome = "BUILD_FAILURE"
		report.BuildOutcome = "FAIL"
		report.ExecutionOutcome = "PASS"
		report.TestedArtifactRef = "NONE"
		cov := validCoverage()
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for BUILD_FAILURE with execution_outcome: PASS")
		}

		// PRODUCT_FAILURE with scenarios_failed: 0
		report = validReport()
		report.StageOutcome = "PRODUCT_FAILURE"
		report.BuildOutcome = "PASS"
		report.ExecutionOutcome = "FAIL"
		report.ScenariosFailed = 0
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for PRODUCT_FAILURE with scenarios_failed: 0")
		}

		// TIMEOUT with execution_outcome: PASS
		report = validReport()
		report.StageOutcome = "TIMEOUT"
		report.BuildOutcome = "TIMEOUT"
		report.ExecutionOutcome = "PASS"
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for TIMEOUT with execution_outcome: PASS")
		}

		// CANCELLED with execution_outcome: PASS
		report = validReport()
		report.StageOutcome = "CANCELLED"
		report.BuildOutcome = "CANCELLED"
		report.ExecutionOutcome = "PASS"
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for CANCELLED with execution_outcome: PASS")
		}

		// ENV_BLOCKED with build_outcome: PASS & execution_outcome: PASS
		report = validReport()
		report.StageOutcome = "ENV_BLOCKED"
		report.BuildOutcome = "PASS"
		report.ExecutionOutcome = "PASS"
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for ENV_BLOCKED with build_outcome: PASS & execution_outcome: PASS")
		}

		// CLARIFICATION_REQUIRED with execution_outcome: PASS
		report = validReport()
		report.StageOutcome = "CLARIFICATION_REQUIRED"
		report.BuildOutcome = "NOT_RUN"
		report.ExecutionOutcome = "PASS"
		report.ResultCompleteness = "NONE"
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for CLARIFICATION_REQUIRED with execution_outcome: PASS")
		}

		// INVALID_EVIDENCE with stage_outcome: PASS
		report = validReport()
		report.StageOutcome = "PASS"
		report.ExecutionOutcome = "INVALID_EVIDENCE"
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for INVALID_EVIDENCE with stage_outcome: PASS")
		}
	}

	// 12. Identity mismatch
	{
		brief := validBrief()
		report := validReport()
		cov := validCoverage()

		report.RunID = "different-run-id"
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for RunID mismatch")
		}

		report = validReport()
		report.CandidateRef = "fedcba9876543210fedcba9876543210fedcba98"
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for CandidateRef mismatch")
		}
	}

	// 13. Absent provenance (run-only)
	{
		brief := validBrief()
		brief.Mode = "run-only"
		report := validReport()
		report.BuildOutcome = "NOT_RUN"
		cov := validCoverage()

		report.TestedArtifactRef = "N/A"
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for tested_artifact_ref: N/A")
		}

		report.TestedArtifactRef = "bin/agentplaybook" // missing sha256:
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for tested_artifact_ref without sha256 prefix")
		}

		report.TestedArtifactRef = "sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945" // bare digest without provenance descriptor
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for bare digest tested_artifact_ref without provenance descriptor")
		}
	}

	// 14. Budget overflow (token count >= 150)
	{
		brief := validBrief()
		report := validReport()
		cov := validCoverage()
		report.Diagnostic = strings.Repeat("Extremely verbose diagnostic message detailing multiple failure traces and long logs. ", 15)
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error when diagnostic overflows 150-token budget")
		}
	}

	// 15. CandidateRef syntax validation
	{
		brief := validBrief()
		report := validReport()
		cov := validCoverage()

		// Brief CandidateRef symbolic "main"
		brief.CandidateRef = "main"
		report.CandidateRef = "main"
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for symbolic CandidateRef 'main'")
		}

		// Brief CandidateRef short hex
		brief = validBrief()
		brief.CandidateRef = "0123456789abcdef"
		report.CandidateRef = "0123456789abcdef"
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for short hex CandidateRef")
		}

		// Report CandidateRef invalid syntax
		brief = validBrief()
		report = validReport()
		report.CandidateRef = "not-valid-hex-length-and-characters-!!!"
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for invalid report CandidateRef syntax")
		}
	}

	// 16. TestPackage syntax validation
	{
		brief := validBrief()
		report := validReport()
		cov := validCoverage()

		brief.TestPackage = "../outside"
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for TestPackage directory traversal '../outside'")
		}

		brief.TestPackage = "pkg/core"
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for TestPackage outside e2e/ 'pkg/core'")
		}

		brief.TestPackage = "e2e/../escape"
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for TestPackage traversal in e2e/ 'e2e/../escape'")
		}
	}

	// 17. EvidenceURI syntax and UNAVAILABLE sentinel restrictions
	{
		brief := validBrief()
		report := validReport()
		cov := validCoverage()

		// Invalid EvidenceURI syntax
		report.EvidenceURI = "not-a-uri"
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for invalid EvidenceURI 'not-a-uri'")
		}

		// UNAVAILABLE prohibited on PASS
		report = validReport()
		report.EvidenceURI = "UNAVAILABLE"
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for EvidenceURI UNAVAILABLE with PASS")
		}

		// UNAVAILABLE prohibited on PRODUCT_FAILURE
		report = validReport()
		report.StageOutcome = "PRODUCT_FAILURE"
		report.BuildOutcome = "PASS"
		report.ExecutionOutcome = "FAIL"
		report.ScenariosFailed = 1
		report.ScenariosPassed = 1
		report.ScenariosSkipped = 1
		report.EvidenceURI = "UNAVAILABLE"
		if err := knowledge.ValidateE2EReport(brief, report, cov); err == nil {
			t.Error("expected error for EvidenceURI UNAVAILABLE with PRODUCT_FAILURE")
		}

		// UNAVAILABLE prohibited when scenarios were executed (scenarios_passed > 0)
		report = validReport()
		report.StageOutcome = "ENV_BLOCKED"
		report.BuildOutcome = "ENV_ERROR"
		report.ExecutionOutcome = "ENV_ERROR"
		report.ResultCompleteness = "NONE"
		report.TestedArtifactRef = "NONE"
		report.EvidenceURI = "UNAVAILABLE"
		report.ScenariosSelected = 2
		report.ScenariosPassed = 1
		report.ScenariosFailed = 0
		report.ScenariosSkipped = 1
		covExecuted := knowledge.E2EScenarioCoverageContext{
			SelectedScenarioIDs: []string{"SC-01", "SC-02"},
			SkippedScenarioIDs:  []string{"SC-02"},
			AllowedSkipIDs:      []string{"SC-02"},
		}
		if err := knowledge.ValidateE2EReport(brief, report, covExecuted); err == nil {
			t.Error("expected error for EvidenceURI UNAVAILABLE when scenarios_passed > 0")
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
