package cli_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ChiaYuChang/workflow/internal/cli"
)

func TestCLI_GoldenDiscoveryMatrix(t *testing.T) {
	t.Parallel()

	discoveryCommands := [][]string{
		{},
		{"role"},
		{"flow"},
		{"artifact"},
		{"rule"},
	}

	for _, cmd := range discoveryCommands {
		cmdName := strings.Join(cmd, " ")
		if cmdName == "" {
			cmdName = "<root>"
		}
		t.Run("discovery_"+cmdName, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := cli.Execute(cmd, &stdout, &stderr, "v0.1.0")
			if err != nil {
				t.Fatalf("expected discovery command %q to succeed with exit 0, got err: %v", cmdName, err)
			}
			if stdout.Len() == 0 {
				t.Errorf("expected non-empty output for discovery command %q", cmdName)
			}
			if stderr.Len() != 0 {
				t.Errorf("expected empty stderr for discovery command %q, got: %s", cmdName, stderr.String())
			}
		})
	}
}

func TestCLI_GoldenJSONMatrix(t *testing.T) {
	t.Parallel()

	jsonCommands := [][]string{
		{"--language"},
		{"role", "planner"},
		{"role", "builder"},
		{"role", "reviewer"},
		{"role", "builder", "--responsibility"},
		{"role", "builder", "--boundary"},
		{"role", "builder", "--communication"},
		{"flow", "init"},
		{"flow", "init", "--step", "1"},
		{"flow", "plan"},
		{"flow", "build"},
		{"flow", "review"},
		{"artifact", "repo-summary"},
		{"artifact", "build-plan"},
		{"artifact", "review-plan"},
		{"artifact", "review-findings"},
		{"rule", "list"},
		{"rule", "explain", "anti-cheating"},
		{"rule", "explain", "atomic-change-units"},
		{"rule", "explain", "mandatory-alignment"},
		{"rule", "explain", "tdd-reproduction"},
	}

	for _, cmd := range jsonCommands {
		cmdName := strings.Join(cmd, " ")
		t.Run("json_"+cmdName, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := cli.Execute(cmd, &stdout, &stderr, "v0.1.0")
			if err != nil {
				t.Fatalf("expected command %q to succeed, got err: %v", cmdName, err)
			}
			if !json.Valid(stdout.Bytes()) {
				t.Fatalf("expected valid JSON output for command %q, got:\n%s", cmdName, stdout.String())
			}
		})
	}
}

func TestCLI_GoldenErrorMatrix(t *testing.T) {
	t.Parallel()

	errorCommands := [][]string{
		{"--language", "--transport"},
		{"role", "nonexistent"},
		{"role", "builder", "--responsibility", "--boundary"},
		{"role", "--responsibility"},
		{"flow", "nonexistent"},
		{"flow", "init", "--step", "999"},
		{"flow", "--step", "1"},
		{"artifact", "nonexistent"},
		{"rule", "explain"},
		{"rule", "explain", "nonexistent"},
	}

	for _, cmd := range errorCommands {
		cmdName := strings.Join(cmd, " ")
		t.Run("error_"+cmdName, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := cli.Execute(cmd, &stdout, &stderr, "v0.1.0")
			if err == nil {
				t.Fatalf("expected command %q to fail with error, got nil", cmdName)
			}
		})
	}
}
