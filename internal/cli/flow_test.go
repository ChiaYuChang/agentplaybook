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
	for _, expected := range []string{"init", "plan", "build", "review"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected flow %q in discovery output, got: %s", expected, out)
		}
	}
}

func TestFlow_QueryFull(t *testing.T) {
	t.Parallel()

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
	if len(f.Steps) != 7 {
		t.Errorf("expected 7 steps in init, got %d", len(f.Steps))
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
		if s.Index != 2 || s.Actor != "reviewer" {
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
