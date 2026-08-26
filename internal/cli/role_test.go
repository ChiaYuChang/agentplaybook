package cli_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ChiaYuChang/workflow/internal/cli"
	"github.com/ChiaYuChang/workflow/internal/knowledge"
)

func TestRole_BareDiscovery(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"role"}, &stdout, &stderr, "dev")
	if err != nil {
		t.Fatalf("expected bare 'role' to succeed, got: %v", err)
	}

	out := stdout.String()
	for _, expected := range []string{"planner", "builder", "reviewer"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected role %q in discovery output, got: %s", expected, out)
		}
	}
}

func TestRole_QueryFull(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"role", "builder"}, &stdout, &stderr, "dev")
	if err != nil {
		t.Fatalf("querying role builder failed: %v", err)
	}

	var r knowledge.RoleDefinition
	if err := json.Unmarshal(stdout.Bytes(), &r); err != nil {
		t.Fatalf("failed to decode JSON response: %v\nRaw: %s", err, stdout.String())
	}

	if r.Name != "builder" {
		t.Errorf("expected role name 'builder', got %q", r.Name)
	}
	if len(r.Responsibilities) == 0 {
		t.Error("expected non-empty responsibilities")
	}
	if len(r.Boundaries) == 0 {
		t.Error("expected non-empty boundaries")
	}
}

func TestRole_Selectors(t *testing.T) {
	t.Parallel()

	// 1. --responsibility
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"role", "builder", "--responsibility"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("--responsibility failed: %v", err)
		}
		var list []string
		if err := json.Unmarshal(stdout.Bytes(), &list); err != nil {
			t.Fatalf("failed to decode responsibilities: %v", err)
		}
		if len(list) == 0 {
			t.Error("expected responsibilities list to be non-empty")
		}
	}

	// 2. --boundary
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"role", "builder", "--boundary"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("--boundary failed: %v", err)
		}
		var list []string
		if err := json.Unmarshal(stdout.Bytes(), &list); err != nil {
			t.Fatalf("failed to decode boundaries: %v", err)
		}
		if len(list) == 0 {
			t.Error("expected boundaries list to be non-empty")
		}
	}

	// 3. Mutually exclusive selectors
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"role", "builder", "--responsibility", "--boundary"}, &stdout, &stderr, "dev")
		if err == nil {
			t.Error("expected error when multiple selectors are specified")
		}
	}

	// 4. Selector without role name
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"role", "--responsibility"}, &stdout, &stderr, "dev")
		if err == nil {
			t.Error("expected error when selector specified without role name")
		}
	}
}

func TestRole_Unknown(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"role", "nonexistent"}, &stdout, &stderr, "dev")
	if err == nil {
		t.Error("expected error for unknown role, got nil")
	}
	if !strings.Contains(err.Error(), "unknown role") {
		t.Errorf("unexpected error message: %v", err)
	}
}
