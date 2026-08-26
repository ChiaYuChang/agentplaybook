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
	for _, expected := range []string{"repo-summary", "build-plan", "review-plan", "review-findings"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected artifact %q in discovery output, got: %s", expected, out)
		}
	}
}

func TestArtifact_Query(t *testing.T) {
	t.Parallel()

	// 1. Document artifact (build-plan)
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

	// 2. Message artifact (review-findings)
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
		if len(a.Fields) == 0 {
			t.Error("expected non-empty fields for review-findings")
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
