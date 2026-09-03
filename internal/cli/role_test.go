package cli_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ChiaYuChang/agentplaybook/internal/cli"
	"github.com/ChiaYuChang/agentplaybook/internal/knowledge"
)

func TestRole_BareDiscovery(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"role"}, &stdout, &stderr, "dev")
	if err != nil {
		t.Fatalf("expected bare 'role' to succeed, got: %v", err)
	}

	out := stdout.String()
	for _, expected := range []string{"planner", "builder", "reviewer", "scout", "navigator", "cartographer"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected role %q in discovery output, got: %s", expected, out)
		}
	}
}

func TestRole_Navigator(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"role", "navigator"}, &stdout, &stderr, "dev")
	if err != nil {
		t.Fatalf("querying role navigator failed: %v", err)
	}

	var r knowledge.RoleDefinition
	if err := json.Unmarshal(stdout.Bytes(), &r); err != nil {
		t.Fatalf("failed to decode JSON response: %v\nRaw: %s", err, stdout.String())
	}

	if r.Name != "navigator" {
		t.Errorf("expected role name 'navigator', got %q", r.Name)
	}
	if r.Category != "companion" {
		t.Errorf("expected category 'companion', got %q", r.Category)
	}
	if len(r.Responsibilities) == 0 {
		t.Error("expected non-empty responsibilities")
	}
	if len(r.Boundaries) == 0 {
		t.Error("expected non-empty boundaries")
	}
	if len(r.Communication.Targets) != 3 || r.Communication.Targets[0] != knowledge.RoleUser || r.Communication.Targets[1] != knowledge.RolePlanner || r.Communication.Targets[2] != knowledge.RoleCartographer {
		t.Errorf("expected navigator communication targets [user planner cartographer], got %v", r.Communication.Targets)
	}

	// Test selectors
	for _, flag := range []string{"--responsibility", "--boundary", "--communication"} {
		stdout.Reset()
		stderr.Reset()
		if err := cli.Execute([]string{"role", "navigator", flag}, &stdout, &stderr, "dev"); err != nil {
			t.Fatalf("querying role navigator %s failed: %v", flag, err)
		}
		if !json.Valid(stdout.Bytes()) {
			t.Errorf("expected valid JSON for flag %s, got: %s", flag, stdout.String())
		}
	}
}

func TestRole_Cartographer(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"role", "cartographer"}, &stdout, &stderr, "dev")
	if err != nil {
		t.Fatalf("querying role cartographer failed: %v", err)
	}

	var r knowledge.RoleDefinition
	if err := json.Unmarshal(stdout.Bytes(), &r); err != nil {
		t.Fatalf("failed to decode JSON response: %v\nRaw: %s", err, stdout.String())
	}

	if r.Name != "cartographer" {
		t.Errorf("expected role name 'cartographer', got %q", r.Name)
	}
	if r.Category != "companion" {
		t.Errorf("expected category 'companion', got %q", r.Category)
	}
	if r.Title != "System Cartographer & Visual Architect" {
		t.Errorf("unexpected cartographer title: %q", r.Title)
	}
	if len(r.Responsibilities) == 0 {
		t.Error("expected non-empty responsibilities")
	}
	if len(r.Boundaries) == 0 {
		t.Error("expected non-empty boundaries")
	}
	if len(r.Communication.Targets) != 3 || r.Communication.Targets[0] != knowledge.RoleUser || r.Communication.Targets[1] != knowledge.RolePlanner || r.Communication.Targets[2] != knowledge.RoleNavigator {
		t.Errorf("expected cartographer communication targets [user planner navigator], got %v", r.Communication.Targets)
	}

	// Test selectors
	for _, flag := range []string{"--responsibility", "--boundary", "--communication"} {
		stdout.Reset()
		stderr.Reset()
		if err := cli.Execute([]string{"role", "cartographer", flag}, &stdout, &stderr, "dev"); err != nil {
			t.Fatalf("querying role cartographer %s failed: %v", flag, err)
		}
		if !json.Valid(stdout.Bytes()) {
			t.Errorf("expected valid JSON for flag %s, got: %s", flag, stdout.String())
		}
	}
}

func TestCLI_Role_Cartographer(t *testing.T) {
	TestRole_Cartographer(t)
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
