package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ChiaYuChang/workflow/internal/cli"
)

func TestRoot_BareDiscovery(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{}, &stdout, &stderr, "v1.0.0-test")
	if err != nil {
		t.Fatalf("expected bare execution to succeed, got: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "workflow - Multi-Agent Collaboration Manual CLI") {
		t.Errorf("unexpected bare output: %s", out)
	}
	if !strings.Contains(out, "Knowledge Domains:") {
		t.Errorf("expected knowledge domains in output, got: %s", out)
	}
}

func TestRoot_GlobalFlags(t *testing.T) {
	t.Parallel()

	// 1. --language
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"--language"}, &stdout, &stderr, "v1.0.0-test")
		if err != nil {
			t.Fatalf("--language failed: %v", err)
		}
		if !strings.Contains(stdout.String(), "zh-TW") || !strings.Contains(stdout.String(), "en-US") {
			t.Errorf("unexpected --language output: %s", stdout.String())
		}
	}

	// 2. --transport
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"--transport"}, &stdout, &stderr, "v1.0.0-test")
		if err != nil {
			t.Fatalf("--transport failed: %v", err)
		}
		if strings.TrimSpace(stdout.String()) != "herdr" {
			t.Errorf("unexpected --transport output: %s", stdout.String())
		}
	}

	// 3. --version
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"--version"}, &stdout, &stderr, "v1.0.0-test")
		if err != nil {
			t.Fatalf("--version failed: %v", err)
		}
		if strings.TrimSpace(stdout.String()) != "v1.0.0-test" {
			t.Errorf("unexpected --version output: %s", stdout.String())
		}
	}

	// 4. Conflicting flags
	{
		var stdout, stderr bytes.Buffer
		err := cli.Execute([]string{"--language", "--transport"}, &stdout, &stderr, "v1.0.0-test")
		if err == nil {
			t.Error("expected error with multiple flags, got nil")
		}
	}
}
