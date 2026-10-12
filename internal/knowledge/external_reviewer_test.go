package knowledge_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/ChiaYuChang/agentplaybook/internal/knowledge"
)

// externalVisibilityAllowlist is the only set of artifacts External Reviewer may see.
var externalVisibilityAllowlist = []string{
	"agents-md",
	"build-plan",
	"blueprint-plan",
	"review-resolution",
	"sub-review-resolution",
	"external-review-report",
}

func cloneCatalog(t *testing.T) (*knowledge.Knowledge, []knowledge.RoleDefinition, []knowledge.Artifact, []knowledge.Flow) {
	t.Helper()
	k, err := knowledge.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	roles := append([]knowledge.RoleDefinition(nil), k.Roles()...)
	for i := range roles {
		roles[i].Communication.Targets = slices.Clone(roles[i].Communication.Targets)
	}
	artifacts := append([]knowledge.Artifact(nil), k.Artifacts()...)
	for i := range artifacts {
		artifacts[i].Visibility = slices.Clone(artifacts[i].Visibility)
	}
	flows := append([]knowledge.Flow(nil), k.Flows()...)
	for i := range flows {
		flows[i].Steps = slices.Clone(flows[i].Steps)
	}
	invalid := *k
	return &invalid, roles, artifacts, flows
}

// EXT-01/EXT-02: distinct identity, planner-only transport, existing reviewer unchanged.
func TestExternalReviewer_CatalogIdentity(t *testing.T) {
	t.Parallel()

	k, err := knowledge.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	ext, ok := k.Role("external-reviewer")
	if !ok {
		t.Fatalf("expected external-reviewer role to exist")
	}
	if ext.Name != knowledge.RoleExternalReviewer {
		t.Errorf("expected role name %q, got %q", knowledge.RoleExternalReviewer, ext.Name)
	}
	if ext.Category != "core" {
		t.Errorf("expected external-reviewer category core, got %q", ext.Category)
	}
	if len(ext.Communication.Targets) != 1 || ext.Communication.Targets[0] != knowledge.RolePlanner {
		t.Errorf("external-reviewer must communicate exactly with planner, got %v", ext.Communication.Targets)
	}

	planner, _ := k.Role("planner")
	if len(planner.Communication.Targets) != 7 || !slices.Contains(planner.Communication.Targets, knowledge.RoleExternalReviewer) {
		t.Errorf("planner must keep six existing targets plus external-reviewer, got %v", planner.Communication.Targets)
	}
	for _, existing := range []knowledge.Role{knowledge.RoleBuilder, knowledge.RoleReviewer, knowledge.RoleScout, knowledge.RoleNavigator, knowledge.RoleCartographer, knowledge.RoleVerifier} {
		if !slices.Contains(planner.Communication.Targets, existing) {
			t.Errorf("planner lost existing target %q: %v", existing, planner.Communication.Targets)
		}
	}

	internal, ok := k.Role("reviewer")
	if !ok || internal.Name != knowledge.RoleReviewer {
		t.Fatalf("existing reviewer role must remain (Internal Reviewer)")
	}
	if len(internal.Communication.Targets) != 1 || internal.Communication.Targets[0] != knowledge.RolePlanner {
		t.Errorf("reviewer targets must remain [planner], got %v", internal.Communication.Targets)
	}
}

// EXT-02: external-reviewer visible only on the settled allowlist; reviewer keeps private plan access.
func TestExternalReviewer_ArtifactVisibilityAllowlist(t *testing.T) {
	t.Parallel()

	k, err := knowledge.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	for _, a := range k.Artifacts() {
		visible := slices.Contains(a.Visibility, knowledge.RoleExternalReviewer)
		allowed := slices.Contains(externalVisibilityAllowlist, a.Name)
		if visible && !allowed {
			t.Errorf("artifact %q must not be visible to external-reviewer", a.Name)
		}
		if a.Owner == knowledge.RoleExternalReviewer && a.Name != "external-review-report" {
			t.Errorf("external-reviewer must own no artifact other than external-review-report, owns %q", a.Name)
		}
	}
	for _, name := range []string{"agents-md", "build-plan", "blueprint-plan", "review-resolution", "sub-review-resolution", "external-review-report"} {
		a, ok := k.Artifact(name)
		if !ok {
			t.Fatalf("expected artifact %q", name)
		}
		if !slices.Contains(a.Visibility, knowledge.RoleExternalReviewer) {
			t.Errorf("artifact %q must be visible to external-reviewer, got %v", name, a.Visibility)
		}
	}
	for _, private := range []string{"review-plan", "review-findings", "sub-review-plan", "sub-build-plan", "e2e-brief", "e2e-report", "e2e-clarification-request", "scout-survey"} {
		a, ok := k.Artifact(private)
		if !ok {
			t.Fatalf("expected artifact %q", private)
		}
		if slices.Contains(a.Visibility, knowledge.RoleExternalReviewer) {
			t.Errorf("private artifact %q leaked to external-reviewer", private)
		}
	}
	// Existing Internal Reviewer retains private plan and findings access.
	for _, private := range []string{"review-plan", "review-findings"} {
		a, _ := k.Artifact(private)
		if !slices.Contains(a.Visibility, knowledge.RoleReviewer) {
			t.Errorf("Internal Reviewer must keep access to %q, got %v", private, a.Visibility)
		}
	}
}

// EXT-03: report contract is owned by external-reviewer and visible to exactly planner + external-reviewer.
func TestExternalReviewReport_Contract(t *testing.T) {
	t.Parallel()

	k, err := knowledge.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	a, ok := k.Artifact("external-review-report")
	if !ok {
		t.Fatalf("expected external-review-report artifact")
	}
	if a.Type != "message" || a.Owner != knowledge.RoleExternalReviewer {
		t.Errorf("expected message owned by external-reviewer, got type=%q owner=%q", a.Type, a.Owner)
	}
	if len(a.Visibility) != 2 || !slices.Contains(a.Visibility, knowledge.RolePlanner) || !slices.Contains(a.Visibility, knowledge.RoleExternalReviewer) {
		t.Errorf("report visibility must be exactly planner and external-reviewer, got %v", a.Visibility)
	}
	required := map[string]bool{
		"candidate_ref":           true,
		"scope":                   true,
		"outcome":                 true,
		"evidence":                true,
		"findings":                false,
		"verification_assessment": true,
	}
	if len(a.Fields) != len(required) {
		t.Fatalf("expected %d fields, got %d: %+v", len(required), len(a.Fields), a.Fields)
	}
	for _, f := range a.Fields {
		want, ok := required[f.Name]
		if !ok {
			t.Errorf("unexpected field %q", f.Name)
			continue
		}
		if f.Required != want {
			t.Errorf("field %q required=%v, want %v", f.Name, f.Required, want)
		}
	}
	desc := strings.ToLower(a.Description)
	for _, phrase := range []string{"does not authorize", "push", "fixed"} {
		if !strings.Contains(desc, phrase) {
			t.Errorf("report description must state %q, got: %s", phrase, a.Description)
		}
	}
}

// EXT-02: validator rejects every boundary violation for the new identity.
func TestValidate_ExternalReviewerBoundaries(t *testing.T) {
	t.Parallel()

	// Baseline: catalog must be valid.
	{
		k, roles, artifacts, flows := cloneCatalog(t)
		k.SetRolesForTest(roles)
		k.SetArtifactsForTest(artifacts)
		k.SetFlowsForTest(flows)
		if err := knowledge.Validate(k); err != nil {
			t.Fatalf("baseline catalog must validate: %v", err)
		}
	}

	roleMutations := map[string]func([]knowledge.RoleDefinition){
		"external-reviewer targets builder": func(rs []knowledge.RoleDefinition) {
			setTargets(rs, knowledge.RoleExternalReviewer, knowledge.RolePlanner, knowledge.RoleBuilder)
		},
		"external-reviewer targets user": func(rs []knowledge.RoleDefinition) {
			setTargets(rs, knowledge.RoleExternalReviewer, knowledge.RoleUser)
		},
		"external-reviewer targets reviewer only": func(rs []knowledge.RoleDefinition) {
			setTargets(rs, knowledge.RoleExternalReviewer, knowledge.RoleReviewer)
		},
		"external-reviewer targets nobody": func(rs []knowledge.RoleDefinition) {
			setTargets(rs, knowledge.RoleExternalReviewer)
		},
		"planner drops external-reviewer": func(rs []knowledge.RoleDefinition) {
			setTargets(rs, knowledge.RolePlanner, knowledge.RoleBuilder, knowledge.RoleReviewer, knowledge.RoleScout, knowledge.RoleNavigator, knowledge.RoleCartographer, knowledge.RoleVerifier)
		},
		"planner has duplicate external-reviewer": func(rs []knowledge.RoleDefinition) {
			setTargets(rs, knowledge.RolePlanner, knowledge.RoleBuilder, knowledge.RoleReviewer, knowledge.RoleScout, knowledge.RoleNavigator, knowledge.RoleCartographer, knowledge.RoleVerifier, knowledge.RoleExternalReviewer, knowledge.RoleExternalReviewer)
		},
		"planner drops reviewer but keeps count": func(rs []knowledge.RoleDefinition) {
			setTargets(rs, knowledge.RolePlanner, knowledge.RoleBuilder, knowledge.RoleExternalReviewer, knowledge.RoleScout, knowledge.RoleNavigator, knowledge.RoleCartographer, knowledge.RoleVerifier)
		},
		"navigator targets external-reviewer": func(rs []knowledge.RoleDefinition) {
			setTargets(rs, knowledge.RoleNavigator, knowledge.RoleUser, knowledge.RolePlanner, knowledge.RoleCartographer, knowledge.RoleExternalReviewer)
		},
		"external-reviewer invalid category": func(rs []knowledge.RoleDefinition) {
			for i := range rs {
				if rs[i].Name == knowledge.RoleExternalReviewer {
					rs[i].Category = "bogus"
				}
			}
		},
	}
	for name, mutate := range roleMutations {
		t.Run("role/"+name, func(t *testing.T) {
			t.Parallel()
			k, roles, _, _ := cloneCatalog(t)
			mutate(roles)
			k.SetRolesForTest(roles)
			if err := knowledge.Validate(k); err == nil {
				t.Errorf("expected validation error")
			}
		})
	}

	artifactMutations := map[string]func([]knowledge.Artifact){
		"external-reviewer sees review-plan": func(as []knowledge.Artifact) {
			addVisibility(as, "review-plan", knowledge.RoleExternalReviewer)
		},
		"external-reviewer sees review-findings": func(as []knowledge.Artifact) {
			addVisibility(as, "review-findings", knowledge.RoleExternalReviewer)
		},
		"external-reviewer sees sub-review-plan": func(as []knowledge.Artifact) {
			addVisibility(as, "sub-review-plan", knowledge.RoleExternalReviewer)
		},
		"external-reviewer sees e2e-report": func(as []knowledge.Artifact) {
			addVisibility(as, "e2e-report", knowledge.RoleExternalReviewer)
		},
		"external-reviewer sees scout-survey": func(as []knowledge.Artifact) {
			addVisibility(as, "scout-survey", knowledge.RoleExternalReviewer)
		},
		"external-reviewer owns build-plan": func(as []knowledge.Artifact) {
			setOwner(as, "build-plan", knowledge.RoleExternalReviewer)
		},
		"external-reviewer owns review-findings": func(as []knowledge.Artifact) {
			setOwner(as, "review-findings", knowledge.RoleExternalReviewer)
		},
		"reviewer owns external-review-report": func(as []knowledge.Artifact) {
			setOwner(as, "external-review-report", knowledge.RoleReviewer)
		},
		"planner owns external-review-report": func(as []knowledge.Artifact) {
			setOwner(as, "external-review-report", knowledge.RolePlanner)
		},
		"report visible to builder": func(as []knowledge.Artifact) {
			addVisibility(as, "external-review-report", knowledge.RoleBuilder)
		},
		"report visible to reviewer": func(as []knowledge.Artifact) {
			addVisibility(as, "external-review-report", knowledge.RoleReviewer)
		},
		"report visible to navigator": func(as []knowledge.Artifact) {
			addVisibility(as, "external-review-report", knowledge.RoleNavigator)
		},
		"report duplicate planner recipient": func(as []knowledge.Artifact) {
			setVisibility(as, "external-review-report", knowledge.RolePlanner, knowledge.RolePlanner)
		},
		"report planner only": func(as []knowledge.Artifact) {
			setVisibility(as, "external-review-report", knowledge.RolePlanner)
		},
		"report external-reviewer only": func(as []knowledge.Artifact) {
			setVisibility(as, "external-review-report", knowledge.RoleExternalReviewer)
		},
		"report duplicate external recipient": func(as []knowledge.Artifact) {
			setVisibility(as, "external-review-report", knowledge.RolePlanner, knowledge.RoleExternalReviewer, knowledge.RoleExternalReviewer)
		},
	}
	for name, mutate := range artifactMutations {
		t.Run("artifact/"+name, func(t *testing.T) {
			t.Parallel()
			k, _, artifacts, _ := cloneCatalog(t)
			mutate(artifacts)
			k.SetArtifactsForTest(artifacts)
			if err := knowledge.Validate(k); err == nil {
				t.Errorf("expected validation error")
			}
		})
	}

	// External Reviewer is a user-requested message identity, never a standing flow gate actor.
	for _, flowName := range []string{"plan", "build", "review", "commit", "blueprint", "e2e"} {
		t.Run("flow/"+flowName, func(t *testing.T) {
			t.Parallel()
			k, _, _, flows := cloneCatalog(t)
			for i := range flows {
				if flows[i].Name == flowName {
					flows[i].Steps[0].Actor = knowledge.RoleExternalReviewer
				}
			}
			k.SetFlowsForTest(flows)
			if err := knowledge.Validate(k); err == nil {
				t.Errorf("expected validation error when external-reviewer acts in flow %q", flowName)
			}
		})
	}
}

func setTargets(rs []knowledge.RoleDefinition, name knowledge.Role, targets ...knowledge.Role) {
	for i := range rs {
		if rs[i].Name == name {
			rs[i].Communication.Targets = targets
		}
	}
}

func addVisibility(as []knowledge.Artifact, name string, role knowledge.Role) {
	for i := range as {
		if as[i].Name == name {
			as[i].Visibility = append(as[i].Visibility, role)
		}
	}
}

func setVisibility(as []knowledge.Artifact, name string, roles ...knowledge.Role) {
	for i := range as {
		if as[i].Name == name {
			as[i].Visibility = roles
		}
	}
}

func setOwner(as []knowledge.Artifact, name string, owner knowledge.Role) {
	for i := range as {
		if as[i].Name == name {
			as[i].Owner = owner
		}
	}
}
