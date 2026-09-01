package knowledge_test

import (
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
	if len(roles) != 5 {
		t.Fatalf("expected 5 roles, got %d", len(roles))
	}
	for _, expected := range []string{"planner", "builder", "reviewer", "scout", "navigator"} {
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
	if len(nav.Communication.Targets) != 2 || nav.Communication.Targets[0] != knowledge.RoleUser || nav.Communication.Targets[1] != knowledge.RolePlanner {
		t.Errorf("expected navigator communication targets [user planner], got %v", nav.Communication.Targets)
	}

	// Verify planner communication targets include navigator
	planner, ok := k.Role("planner")
	if !ok {
		t.Fatalf("expected planner role to exist")
	}
	if planner.Category != "core" {
		t.Errorf("expected planner category 'core', got %q", planner.Category)
	}
	foundNavTarget := false
	for _, target := range planner.Communication.Targets {
		if target == knowledge.RoleNavigator {
			foundNavTarget = true
			break
		}
	}
	if !foundNavTarget {
		t.Errorf("expected planner communication targets to include navigator, got %v", planner.Communication.Targets)
	}

	// 3. Verify Flows
	flows := k.Flows()
	if len(flows) != 7 {
		t.Fatalf("expected 7 flows, got %d", len(flows))
	}
	for _, expected := range []string{"init", "plan", "blueprint", "build", "review", "commit", "session-handoff"} {
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

	// 4. Verify Artifacts
	artifacts := k.Artifacts()
	if len(artifacts) != 10 {
		t.Fatalf("expected 10 artifacts, got %d", len(artifacts))
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
	} {
		a, ok := k.Artifact(expected)
		if !ok {
			t.Errorf("expected artifact %q to exist", expected)
		}
		if a.Name != expected {
			t.Errorf("expected artifact name %q, got %q", expected, a.Name)
		}
	}

	// Regression check: build-plan and review-findings visibility
	buildPlan, _ := k.Artifact("build-plan")
	if len(buildPlan.Visibility) != 3 {
		t.Errorf("expected build-plan visibility to have 3 roles, got %v", buildPlan.Visibility)
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
	if len(reviewResolution.Visibility) != 4 {
		t.Errorf("expected review-resolution visibility to include 4 roles, got %v", reviewResolution.Visibility)
	}

	// 5. Verify Rules
	rules := k.Rules()
	if len(rules) < 22 {
		t.Fatalf("expected at least 22 rules, got %d", len(rules))
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
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RoleBuilder, knowledge.RoleReviewer, knowledge.RoleScout}
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Planner omits navigator from communication targets")
		}
	}

	// 7. Planner targeting external user directly
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RolePlanner {
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RoleBuilder, knowledge.RoleReviewer, knowledge.RoleScout, knowledge.RoleUser}
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Planner targets external user directly")
		}
	}

	// 8. Navigator omitting user or planner
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RoleNavigator {
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RolePlanner}
			}
		}
		setRoles(&invalid, roles)
		if err := knowledge.Validate(&invalid); err == nil {
			t.Error("expected error when Navigator omits user from communication targets")
		}
	}

	// 8. Navigator targeting builder (violating star topology)
	{
		invalid := *k
		roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
		for i, r := range roles {
			if r.Name == knowledge.RoleNavigator {
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RoleUser, knowledge.RolePlanner, knowledge.RoleBuilder}
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
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RoleUser, knowledge.RolePlanner, knowledge.RolePlanner}
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
				roles[i].Communication.Targets = []knowledge.Role{knowledge.RoleBuilder, knowledge.RoleReviewer, knowledge.RoleScout, knowledge.RoleNavigator, knowledge.RoleNavigator}
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

	// 10. Navigator in in-flight artifact visibility
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

	// 11. Navigator in non-allowlisted artifact visibility
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

	// 12. Navigator as flow actor
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

	// 13. User as artifact owner or flow actor
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

func setRoles(k *knowledge.Knowledge, roles []knowledge.RoleDefinition) {
	k.SetRolesForTest(roles)
}

func setArtifacts(k *knowledge.Knowledge, artifacts []knowledge.Artifact) {
	k.SetArtifactsForTest(artifacts)
}

func setFlows(k *knowledge.Knowledge, flows []knowledge.Flow) {
	k.SetFlowsForTest(flows)
}
