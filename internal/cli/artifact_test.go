package cli_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ChiaYuChang/agentplaybook/internal/cli"
	"github.com/ChiaYuChang/agentplaybook/internal/knowledge"
)

func TestArtifact_BareDiscovery(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"artifact"}, &stdout, &stderr, "dev")
	if err != nil {
		t.Fatalf("expected bare 'artifact' to succeed, got: %v", err)
	}

	out := stdout.String()
	for _, expected := range []string{
		"agents-md",
		"build-plan",
		"review-plan",
		"blueprint-plan",
		"sub-build-plan",
		"sub-review-plan",
		"sub-review-resolution",
		"review-findings",
		"scout-survey",
		"review-resolution",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected artifact %q in discovery output, got: %s", expected, out)
		}
	}

	expectedRows := []struct {
		name         string
		owner        string
		artifactType string
		description  string
	}{
		{"agents-md", "planner", "document", "Living operational memory"},
		{"build-plan", "planner", "document", "Task-specific implementation plan"},
		{"review-plan", "planner", "document", "Reviewer-only verification plan"},
		{"blueprint-plan", "planner", "document", "Hierarchical architecture plan"},
		{"sub-build-plan", "planner", "document", "Task-specific implementation plan for an individual sub-plan unit"},
		{"sub-review-plan", "planner", "document", "Reviewer-only verification plan used for independent validation during sub-plan"},
		{"sub-review-resolution", "planner", "document", "Planner-owned post-review synthesis for an individual sub-plan"},
		{"review-findings", "reviewer", "message", "Structured findings reported by Reviewer"},
		{"review-resolution", "planner", "document", "Planner-owned post-review synthesis"},
		{"scout-survey", "scout", "message", "Transient message artifact containing read-only architectural topography"},
	}
	for _, expected := range expectedRows {
		found := false
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, expected.name) &&
				strings.Contains(line, expected.owner) &&
				strings.Contains(line, expected.artifactType) &&
				strings.Contains(line, expected.description) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected catalog row for %q to include owner, type, and description, got: %s", expected.name, out)
		}
	}
}

func TestArtifact_Query(t *testing.T) {
	t.Parallel()

	// 1. Document artifact (agents-md)
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"artifact", "agents-md"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("querying agents-md failed: %v", err)
		}

		var a knowledge.Artifact
		if err := json.Unmarshal(stdout.Bytes(), &a); err != nil {
			t.Fatalf("failed to decode JSON response: %v\nRaw: %s", err, stdout.String())
		}
		if a.Name != "agents-md" {
			t.Errorf("expected artifact name 'agents-md', got %q", a.Name)
		}
		if len(a.Sections) != 5 {
			t.Errorf("expected 5 sections for agents-md, got %d", len(a.Sections))
		}
	}

	// 2. Document artifact (build-plan)
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"artifact", "build-plan"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("querying build-plan failed: %v", err)
		}

		var a knowledge.Artifact
		if err := json.Unmarshal(stdout.Bytes(), &a); err != nil {
			t.Fatalf("failed to decode JSON response: %v\nRaw: %s", err, stdout.String())
		}
		if a.Name != "build-plan" {
			t.Errorf("expected artifact name 'build-plan', got %q", a.Name)
		}
		if len(a.Sections) == 0 {
			t.Error("expected non-empty sections for build-plan")
		}
	}

	// 3. Document artifact (review-plan with optional Track B section)
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"artifact", "review-plan"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("querying review-plan failed: %v", err)
		}

		var a knowledge.Artifact
		if err := json.Unmarshal(stdout.Bytes(), &a); err != nil {
			t.Fatalf("failed to decode JSON response: %v\nRaw: %s", err, stdout.String())
		}
		if a.Name != "review-plan" {
			t.Errorf("expected artifact name 'review-plan', got %q", a.Name)
		}
		foundTrackB := false
		for _, s := range a.Sections {
			if s.Name == "Track B Definition" {
				foundTrackB = true
				if s.Required {
					t.Error("expected Track B Definition to be optional (required=false)")
				}
				for _, reqField := range []string{"action", "boundary", "baseline_identity", "environment", "metrics", "sampling", "criteria"} {
					if !strings.Contains(s.Description, reqField) {
						t.Errorf("expected Track B Definition description to contain field %q, got: %s", reqField, s.Description)
					}
				}
			}
		}
		if !foundTrackB {
			t.Error("expected review-plan to include Track B Definition section")
		}
	}

	// 4. Blueprint artifact (blueprint-plan)
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"artifact", "blueprint-plan"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("querying blueprint-plan failed: %v", err)
		}

		var a knowledge.Artifact
		if err := json.Unmarshal(stdout.Bytes(), &a); err != nil {
			t.Fatalf("failed to decode JSON response: %v\nRaw: %s", err, stdout.String())
		}
		if a.Name != "blueprint-plan" || a.Owner != knowledge.RolePlanner || a.Type != "document" {
			t.Errorf("unexpected blueprint-plan metadata: %+v", a)
		}
		if a.Path != "plan/{timestamp}/{slug}.blueprint.md" {
			t.Errorf("expected blueprint path 'plan/{timestamp}/{slug}.blueprint.md', got %q", a.Path)
		}
		if len(a.Sections) != 4 {
			t.Errorf("expected 4 sections for blueprint-plan, got %d", len(a.Sections))
		}
	}

	// 5. Sub-build plan artifact (sub-build-plan)
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"artifact", "sub-build-plan"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("querying sub-build-plan failed: %v", err)
		}

		var a knowledge.Artifact
		if err := json.Unmarshal(stdout.Bytes(), &a); err != nil {
			t.Fatalf("failed to decode JSON response: %v\nRaw: %s", err, stdout.String())
		}
		if a.Name != "sub-build-plan" || a.Owner != knowledge.RolePlanner || a.Type != "document" {
			t.Errorf("unexpected sub-build-plan metadata: %+v", a)
		}
		if a.Path != "plan/{timestamp}/sub/{slug}.build.md" {
			t.Errorf("expected sub-build-plan path 'plan/{timestamp}/sub/{slug}.build.md', got %q", a.Path)
		}
	}

	// 6. Sub-review plan artifact (sub-review-plan)
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"artifact", "sub-review-plan"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("querying sub-review-plan failed: %v", err)
		}

		var a knowledge.Artifact
		if err := json.Unmarshal(stdout.Bytes(), &a); err != nil {
			t.Fatalf("failed to decode JSON response: %v\nRaw: %s", err, stdout.String())
		}
		if a.Name != "sub-review-plan" || a.Owner != knowledge.RolePlanner || a.Type != "document" {
			t.Errorf("unexpected sub-review-plan metadata: %+v", a)
		}
		if a.Path != "plan/{timestamp}/sub/{slug}.review.md" {
			t.Errorf("expected sub-review-plan path 'plan/{timestamp}/sub/{slug}.review.md', got %q", a.Path)
		}
		if len(a.Visibility) != 2 || a.Visibility[0] != knowledge.RolePlanner || a.Visibility[1] != knowledge.RoleReviewer {
			t.Errorf("expected sub-review-plan visibility [planner reviewer], got %v", a.Visibility)
		}
	}

	// 7. Sub-review resolution artifact (sub-review-resolution)
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"artifact", "sub-review-resolution"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("querying sub-review-resolution failed: %v", err)
		}

		var a knowledge.Artifact
		if err := json.Unmarshal(stdout.Bytes(), &a); err != nil {
			t.Fatalf("failed to decode JSON response: %v\nRaw: %s", err, stdout.String())
		}
		if a.Name != "sub-review-resolution" || a.Owner != knowledge.RolePlanner || a.Type != "document" {
			t.Errorf("unexpected sub-review-resolution metadata: %+v", a)
		}
		if a.Path != "plan/{timestamp}/sub/{slug}.resolution.md" {
			t.Errorf("expected sub-review-resolution path 'plan/{timestamp}/sub/{slug}.resolution.md', got %q", a.Path)
		}
		if len(a.Sections) != 5 {
			t.Errorf("expected 5 sections in sub-review-resolution, got %d", len(a.Sections))
		}
	}

	// 8. Message artifact (review-findings)
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"artifact", "review-findings"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("querying review-findings failed: %v", err)
		}

		var a knowledge.Artifact
		if err := json.Unmarshal(stdout.Bytes(), &a); err != nil {
			t.Fatalf("failed to decode JSON response: %v\nRaw: %s", err, stdout.String())
		}
		if a.Name != "review-findings" {
			t.Errorf("expected artifact name 'review-findings', got %q", a.Name)
		}
		if len(a.Visibility) != 2 || a.Visibility[0] != knowledge.RolePlanner || a.Visibility[1] != knowledge.RoleReviewer {
			t.Errorf("expected review-findings visibility strictly [planner reviewer], got %v", a.Visibility)
		}
		if len(a.Fields) == 0 {
			t.Error("expected non-empty fields for review-findings")
		}

		fieldMap := make(map[string]knowledge.ArtifactField)
		for _, f := range a.Fields {
			fieldMap[f.Name] = f
		}
		for _, reqField := range []string{"id", "location", "severity", "evidence_mode", "evidence", "description", "expected_behavior"} {
			f, ok := fieldMap[reqField]
			if !ok || !f.Required {
				t.Errorf("expected required field %q in review-findings", reqField)
			}
		}
		if f, ok := fieldMap["reproduction_scenario"]; !ok || f.Required {
			t.Error("expected conditional/optional reproduction_scenario in review-findings")
		}
		if strings.Contains(fieldMap["severity"].Description, "Critical") {
			t.Error("expected severity description to exclude ambiguous Critical terminology")
		}
	}

	// 9. Document artifact (review-resolution)
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"artifact", "review-resolution"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("querying review-resolution failed: %v", err)
		}

		var a knowledge.Artifact
		if err := json.Unmarshal(stdout.Bytes(), &a); err != nil {
			t.Fatalf("failed to decode JSON response: %v\nRaw: %s", err, stdout.String())
		}
		if a.Name != "review-resolution" || a.Owner != knowledge.RolePlanner || a.Type != "document" {
			t.Errorf("unexpected review-resolution metadata: %+v", a)
		}
		if a.Path != "plan/{timestamp}/{slug}.resolution.md" {
			t.Errorf("expected path 'plan/{timestamp}/{slug}.resolution.md', got %q", a.Path)
		}
		if a.PathVariables["timestamp"] == "" || a.PathVariables["slug"] == "" {
			t.Errorf("expected path_variables timestamp and slug, got: %+v", a.PathVariables)
		}
		if len(a.Visibility) != 5 || a.Visibility[0] != knowledge.RolePlanner || a.Visibility[1] != knowledge.RoleBuilder || a.Visibility[2] != knowledge.RoleReviewer || a.Visibility[3] != knowledge.RoleNavigator || a.Visibility[4] != knowledge.RoleCartographer {
			t.Errorf("expected visibility [planner builder reviewer navigator cartographer], got %v", a.Visibility)
		}
		for _, reqPhrase := range []string{
			"sanitization",
			"review-plan criteria",
			"hidden test fixtures",
			"private inspection methods",
		} {
			if !strings.Contains(a.Description, reqPhrase) {
				t.Errorf("expected review-resolution description to contain %q, got: %s", reqPhrase, a.Description)
			}
		}

		expectedSections := []string{
			"Outcome",
			"Resolved Findings",
			"Deviations & Rationales",
			"Residual Risks",
			"Verification Evidence",
		}
		if len(a.Sections) != len(expectedSections) {
			t.Fatalf("expected %d sections in review-resolution, got %d", len(expectedSections), len(a.Sections))
		}
		for i, expectedSec := range expectedSections {
			if a.Sections[i].Name != expectedSec || !a.Sections[i].Required {
				t.Errorf("expected required section %q at index %d, got %+v", expectedSec, i, a.Sections[i])
			}
		}
	}
}

func TestArtifact_Unknown(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"artifact", "nonexistent"}, &stdout, &stderr, "dev")
	if err == nil {
		t.Error("expected error for unknown artifact, got nil")
	}
	if !strings.Contains(err.Error(), "unknown artifact") {
		t.Errorf("unexpected error message: %v", err)
	}
}
