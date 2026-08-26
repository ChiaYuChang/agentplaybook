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
	if len(roles) != 3 {
		t.Fatalf("expected 3 roles, got %d", len(roles))
	}
	for _, expected := range []string{"planner", "builder", "reviewer"} {
		r, ok := k.Role(expected)
		if !ok {
			t.Errorf("expected role %q to exist", expected)
		}
		if string(r.Name) != expected {
			t.Errorf("expected role name %q, got %q", expected, r.Name)
		}
	}

	// 3. Verify Flows
	flows := k.Flows()
	if len(flows) != 4 {
		t.Fatalf("expected 4 flows, got %d", len(flows))
	}
	for _, expected := range []string{"init", "plan", "build", "review"} {
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
	if step2.Index != 2 || step2.Actor != "reviewer" {
		t.Errorf("unexpected step 2 data: %+v", step2)
	}
	if _, ok := k.FlowStep("init", 999); ok {
		t.Errorf("expected non-existent step to return false")
	}
	if _, ok := k.FlowStep("unknown-flow", 1); ok {
		t.Errorf("expected non-existent flow to return false")
	}

	// Regression check: review flow has 8 steps and step 7 redirects to step 8
	reviewFlow, _ := k.Flow("review")
	if len(reviewFlow.Steps) != 8 {
		t.Errorf("expected 8 steps in review flow, got %d", len(reviewFlow.Steps))
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

	// 4. Verify Artifacts
	artifacts := k.Artifacts()
	if len(artifacts) != 4 {
		t.Fatalf("expected 4 artifacts, got %d", len(artifacts))
	}
	for _, expected := range []string{"repo-summary", "build-plan", "review-plan", "review-findings"} {
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
	if len(reviewFindings.Visibility) != 2 {
		t.Errorf("expected review-findings visibility to have 2 roles, got %v", reviewFindings.Visibility)
	}

	// 5. Verify Rules
	rules := k.Rules()
	if len(rules) < 7 {
		t.Fatalf("expected at least 7 rules, got %d", len(rules))
	}
	for _, expected := range []string{"anti-cheating", "mandatory-alignment", "tdd-reproduction", "atomic-change-units"} {
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
