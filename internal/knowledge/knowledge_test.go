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
	if len(roles) != 6 {
		t.Fatalf("expected 6 roles, got %d", len(roles))
	}
	for _, expected := range []string{"planner", "builder", "reviewer", "scout", "navigator", "cartographer"} {
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

	// Verify planner communication targets include navigator and cartographer
	planner, ok := k.Role("planner")
	if !ok {
		t.Fatalf("expected planner role to exist")
	}
	if planner.Category != "core" {
		t.Errorf("expected planner category 'core', got %q", planner.Category)
	}
	foundNavTarget, foundCartTarget := false, false
	for _, target := range planner.Communication.Targets {
		if target == knowledge.RoleNavigator {
			foundNavTarget = true
		}
		if target == knowledge.RoleCartographer {
			foundCartTarget = true
		}
	}
	if !foundNavTarget {
		t.Errorf("expected planner communication targets to include navigator, got %v", planner.Communication.Targets)
	}
	if !foundCartTarget {
		t.Errorf("expected planner communication targets to include cartographer, got %v", planner.Communication.Targets)
	}

	// 3. Verify Flows
	flows := k.Flows()
	if len(flows) != 8 {
		t.Fatalf("expected 8 flows, got %d", len(flows))
	}
	for _, expected := range []string{"init", "plan", "blueprint", "build", "review", "commit", "session-handoff", "cartography"} {
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
	if len(cartFlow.Steps) != 5 {
		t.Fatalf("expected 5 steps in cartography flow, got %d", len(cartFlow.Steps))
	}
	if cartFlow.Steps[0].Actor != knowledge.RolePlanner || cartFlow.Steps[1].Actor != knowledge.RolePlanner {
		t.Errorf("expected cartography steps 1 and 2 actor to be planner")
	}
	if cartFlow.Steps[2].Actor != knowledge.RoleCartographer || cartFlow.Steps[3].Actor != knowledge.RoleCartographer || cartFlow.Steps[4].Actor != knowledge.RoleCartographer {
		t.Errorf("expected cartography steps 3, 4, 5 actor to be cartographer")
	}
	cartStep3Conditions := make(map[string]int)
	for _, c := range cartFlow.Steps[2].Conditions {
		cartStep3Conditions[c.When] = c.Then
	}
	if cartStep3Conditions["DIAGRAM_APPROVED"] != 4 || cartStep3Conditions["ADVISORY_ISSUED"] != 5 {
		t.Errorf("unexpected cartography step 3 conditions: %v", cartStep3Conditions)
	}

	// 4. Verify Artifacts
	artifacts := k.Artifacts()
	if len(artifacts) != 12 {
		t.Fatalf("expected 12 artifacts, got %d", len(artifacts))
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
	if len(reviewResolution.Visibility) != 5 {
		t.Errorf("expected review-resolution visibility to include 5 roles, got %v", reviewResolution.Visibility)
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

	// 5. Verify Rules
	rules := k.Rules()
	if len(rules) < 26 {
		t.Fatalf("expected at least 26 rules, got %d", len(rules))
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

func setRoles(k *knowledge.Knowledge, roles []knowledge.RoleDefinition) {
	k.SetRolesForTest(roles)
}

func setArtifacts(k *knowledge.Knowledge, artifacts []knowledge.Artifact) {
	k.SetArtifactsForTest(artifacts)
}

func setFlows(k *knowledge.Knowledge, flows []knowledge.Flow) {
	k.SetFlowsForTest(flows)
}
