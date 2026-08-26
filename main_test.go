package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain_LinkerVersionOverride(t *testing.T) {
	binaryPath := filepath.Join(t.TempDir(), "agentplaybook")

	build := exec.Command("go", "build", "-ldflags", "-X main.version=v9.9.9", "-o", binaryPath, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("failed to build version override binary: %v\n%s", err, output)
	}

	output, err := exec.Command(binaryPath, "--version").CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run version override binary: %v\n%s", err, output)
	}
	if got := strings.TrimSpace(string(output)); got != "v9.9.9" {
		t.Fatalf("expected linker-injected version v9.9.9, got %q", got)
	}
}
