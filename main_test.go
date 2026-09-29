package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ChiaYuChang/agentplaybook/internal/version"
)

func TestMain_EmbeddedVersionOutsideCheckout(t *testing.T) {
	binaryPath := filepath.Join(t.TempDir(), "agentplaybook")

	build := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	versionCommand := exec.Command(binaryPath, "--version")
	versionCommand.Dir = t.TempDir()
	output, err := versionCommand.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary outside checkout: %v\n%s", err, output)
	}
	if got, want := strings.TrimSpace(string(output)), version.Release(); got != want {
		t.Fatalf("--version = %q, want embedded version %q", got, want)
	}

	initCommand := exec.Command(binaryPath, "init")
	initCommand.Dir = t.TempDir()
	output, err = initCommand.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run init outside checkout: %v\n%s", err, output)
	}
	expectedHeader := "AgentPlaybook " + version.Release() + " Living Memory Blueprint"
	if !strings.Contains(string(output), expectedHeader) {
		t.Fatalf("init output missing embedded version header %q", expectedHeader)
	}
}
