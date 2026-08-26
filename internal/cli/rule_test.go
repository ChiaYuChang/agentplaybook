package cli_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ChiaYuChang/workflow/internal/cli"
	"github.com/ChiaYuChang/workflow/internal/knowledge"
)

func TestRule_BareDiscovery(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"rule"}, &stdout, &stderr, "dev")
	if err != nil {
		t.Fatalf("expected bare 'rule' to succeed, got: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "workflow rule - Operational Policies") {
		t.Errorf("unexpected bare output: %s", out)
	}
	if !strings.Contains(out, "list") || !strings.Contains(out, "explain") {
		t.Errorf("expected list and explain in catalog: %s", out)
	}
}

func TestRule_List(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"rule", "list"}, &stdout, &stderr, "dev")
	if err != nil {
		t.Fatalf("'rule list' failed: %v", err)
	}

	type summaryItem struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Category string `json:"category"`
		Summary  string `json:"summary"`
	}

	var items []summaryItem
	if err := json.Unmarshal(stdout.Bytes(), &items); err != nil {
		t.Fatalf("failed to decode JSON response: %v\nRaw: %s", err, stdout.String())
	}

	if len(items) < 7 {
		t.Errorf("expected at least 7 rule items, got %d", len(items))
	}

	foundAntiCheating := false
	for _, item := range items {
		if item.ID == "anti-cheating" {
			foundAntiCheating = true
			break
		}
	}
	if !foundAntiCheating {
		t.Error("expected anti-cheating in rule list")
	}
}

func TestRule_Explain(t *testing.T) {
	t.Parallel()

	// 1. Single rule
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"rule", "explain", "anti-cheating"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("explain anti-cheating failed: %v", err)
		}

		var rules []knowledge.Rule
		if err := json.Unmarshal(stdout.Bytes(), &rules); err != nil {
			t.Fatalf("failed to decode JSON response: %v", err)
		}
		if len(rules) != 1 {
			t.Fatalf("expected 1 rule, got %d", len(rules))
		}
		if rules[0].ID != "anti-cheating" {
			t.Errorf("expected ID 'anti-cheating', got %q", rules[0].ID)
		}
	}

	// 2. Multiple rules
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"rule", "explain", "anti-cheating", "atomic-change-units"}, &stdout, &stderr, "dev")
		if err != nil {
			t.Fatalf("explain multiple rules failed: %v", err)
		}

		var rules []knowledge.Rule
		if err := json.Unmarshal(stdout.Bytes(), &rules); err != nil {
			t.Fatalf("failed to decode JSON response: %v", err)
		}
		if len(rules) != 2 {
			t.Fatalf("expected 2 rules, got %d", len(rules))
		}
	}

	// 3. Explain without args
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"rule", "explain"}, &stdout, &stderr, "dev")
		if err == nil {
			t.Error("expected error when explain invoked without args")
		}
	}

	// 4. Explain unknown rule
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"rule", "explain", "nonexistent"}, &stdout, &stderr, "dev")
		if err == nil {
			t.Error("expected error for unknown rule, got nil")
		}
		if !strings.Contains(err.Error(), "unknown rule") {
			t.Errorf("unexpected error message: %v", err)
		}
	}
}
