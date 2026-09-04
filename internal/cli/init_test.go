package cli_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ChiaYuChang/agentplaybook/internal/cli"
)

func TestCLI_Init_DefaultStdout(t *testing.T) {
	t.Chdir(t.TempDir())

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"init"}, &stdout, &stderr, "dev")
	if err != nil {
		t.Fatalf("expected init without flags to succeed, got: %v\nStderr: %s", err, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("expected empty stderr, got: %s", stderr.String())
	}

	outStr := stdout.String()
	if !strings.Contains(outStr, "# AGENTS.md") {
		t.Errorf("expected stdout to contain markdown header")
	}
	if !strings.Contains(outStr, "Peer-Session Primacy over Subagents") {
		t.Errorf("expected stdout to contain Peer-Session Primacy invariant")
	}

	// Verify zero filesystem writes occurred
	if _, err := os.Stat("AGENTS.md"); !os.IsNotExist(err) {
		t.Errorf("expected AGENTS.md not to exist on disk in default stdout mode")
	}
}

func TestCLI_Init_ForceWithoutFile_Error(t *testing.T) {
	t.Chdir(t.TempDir())

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"init", "--force"}, &stdout, &stderr, "dev")
	if err == nil {
		t.Fatal("expected error with --force without --file, got nil")
	}

	expectedErr := "--force requires --file"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestCLI_Init_File_Default(t *testing.T) {
	t.Chdir(t.TempDir())

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"init", "--file", "AGENTS.md"}, &stdout, &stderr, "dev")
	if err != nil {
		t.Fatalf("expected init --file to succeed, got: %v\nStderr: %s", err, stderr.String())
	}

	expectedMsg := "Initialized standard AGENTS.md at AGENTS.md\n"
	if stdout.String() != expectedMsg {
		t.Errorf("expected stdout %q, got %q", expectedMsg, stdout.String())
	}

	info, err := os.Stat("AGENTS.md")
	if err != nil {
		t.Fatalf("failed to stat created AGENTS.md: %v", err)
	}
	if info.Mode().Perm() != 0644 {
		t.Errorf("expected created AGENTS.md permissions to be 0644, got %o", info.Mode().Perm())
	}

	content, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("failed to read created AGENTS.md: %v", err)
	}
	if !strings.Contains(string(content), "Peer-Session Primacy over Subagents") {
		t.Errorf("expected AGENTS.md to contain Peer-Session Primacy invariant")
	}
}

func TestCLI_Init_Collision(t *testing.T) {
	t.Chdir(t.TempDir())

	initialContent := "pre-existing content"
	if err := os.WriteFile("AGENTS.md", []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to write dummy AGENTS.md: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"init", "--file", "AGENTS.md"}, &stdout, &stderr, "dev")
	if err == nil {
		t.Fatal("expected collision error, got nil")
	}

	expectedErr := "AGENTS.md already exists; use --force to overwrite"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}

	// Verify existing file was not modified
	data, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("failed to read AGENTS.md: %v", err)
	}
	if string(data) != initialContent {
		t.Errorf("file was unexpectedly modified: got %q", string(data))
	}
}

func TestCLI_Init_Force(t *testing.T) {
	t.Chdir(t.TempDir())

	initialContent := "pre-existing content"
	// Create dummy file with overly permissive mode (0777) to verify --force strictly enforces 0644
	if err := os.WriteFile("AGENTS.md", []byte(initialContent), 0777); err != nil {
		t.Fatalf("failed to write dummy AGENTS.md: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"init", "--file", "AGENTS.md", "--force"}, &stdout, &stderr, "dev")
	if err != nil {
		t.Fatalf("expected --force init to succeed, got: %v\nStderr: %s", err, stderr.String())
	}

	expectedMsg := "Initialized standard AGENTS.md at AGENTS.md\n"
	if stdout.String() != expectedMsg {
		t.Errorf("expected stdout %q, got %q", expectedMsg, stdout.String())
	}

	info, err := os.Stat("AGENTS.md")
	if err != nil {
		t.Fatalf("failed to stat overwritten AGENTS.md: %v", err)
	}
	if info.Mode().Perm() != 0644 {
		t.Errorf("expected overwritten AGENTS.md permissions to be 0644, got %o", info.Mode().Perm())
	}

	data, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("failed to read overwritten AGENTS.md: %v", err)
	}
	if !strings.Contains(string(data), "Peer-Session Primacy over Subagents") {
		t.Errorf("expected overwritten content to contain Peer-Session Primacy invariant")
	}
}

func TestCLI_Init_CustomPath(t *testing.T) {
	t.Chdir(t.TempDir())

	customPath := filepath.Join("docs", "nested", "CUSTOM_AGENTS.md")
	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"init", "--file", customPath}, &stdout, &stderr, "dev")
	if err != nil {
		t.Fatalf("expected custom path init to succeed, got: %v\nStderr: %s", err, stderr.String())
	}

	expectedMsg := "Initialized standard AGENTS.md at " + customPath + "\n"
	if stdout.String() != expectedMsg {
		t.Errorf("expected stdout %q, got %q", expectedMsg, stdout.String())
	}

	info, err := os.Stat(customPath)
	if err != nil {
		t.Fatalf("failed to stat custom path file: %v", err)
	}
	if info.Mode().Perm() != 0644 {
		t.Errorf("expected custom path permissions to be 0644, got %o", info.Mode().Perm())
	}

	data, err := os.ReadFile(customPath)
	if err != nil {
		t.Fatalf("failed to read custom path file: %v", err)
	}
	if !strings.Contains(string(data), "Peer-Session Primacy over Subagents") {
		t.Errorf("expected custom path file to contain Peer-Session Primacy invariant")
	}
}

type failingWriter struct{}

func (failingWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("simulated write failure")
}

func TestCLI_Init_WriterErrorPropagation(t *testing.T) {
	t.Parallel()

	err := cli.WriteLivingMemoryTemplate(failingWriter{}, false)
	if err == nil {
		t.Fatal("expected write error, got nil")
	}
	if !strings.Contains(err.Error(), "simulated write failure") {
		t.Errorf("expected simulated write failure, got: %v", err)
	}
}

func TestCLI_Init_ConfirmationWriterFailure(t *testing.T) {
	t.Chdir(t.TempDir())

	var stderr bytes.Buffer
	err := cli.Execute([]string{"init", "--file", "AGENTS.md"}, failingWriter{}, &stderr, "dev")
	if err == nil {
		t.Fatal("expected confirmation write error, got nil")
	}
	if !strings.Contains(err.Error(), "simulated write failure") {
		t.Errorf("expected simulated write failure, got: %v", err)
	}

	// Verify file was created on disk before confirmation failed
	if _, err := os.Stat("AGENTS.md"); err != nil {
		t.Errorf("expected AGENTS.md to have been created before confirmation failed: %v", err)
	}
}

func TestCLI_Init_Minimal_Stdout(t *testing.T) {
	t.Chdir(t.TempDir())

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"init", "--minimal"}, &stdout, &stderr, "dev")
	if err != nil {
		t.Fatalf("expected init --minimal to succeed, got: %v\nStderr: %s", err, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("expected empty stderr, got: %s", stderr.String())
	}

	expected := cli.MinimalLivingMemoryTemplate()
	if stdout.String() != expected {
		t.Errorf("expected stdout to match MinimalLivingMemoryTemplate exactly")
	}

	// Verify zero filesystem writes occurred
	if _, err := os.Stat("AGENTS.md"); !os.IsNotExist(err) {
		t.Errorf("expected AGENTS.md not to exist on disk in minimal stdout mode")
	}
}

func TestCLI_Init_Minimal_ShortFlag(t *testing.T) {
	t.Chdir(t.TempDir())

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"init", "-m"}, &stdout, &stderr, "dev")
	if err != nil {
		t.Fatalf("expected init -m to succeed, got: %v\nStderr: %s", err, stderr.String())
	}

	expected := cli.MinimalLivingMemoryTemplate()
	if stdout.String() != expected {
		t.Errorf("expected init -m stdout to match MinimalLivingMemoryTemplate")
	}
}

func TestCLI_Init_Minimal_ForceWithoutFile_Error(t *testing.T) {
	t.Chdir(t.TempDir())

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"init", "--minimal", "--force"}, &stdout, &stderr, "dev")
	if err == nil {
		t.Fatal("expected error with --force without --file, got nil")
	}

	expectedErr := "--force requires --file"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}

	// Verify zero filesystem writes occurred
	if _, err := os.Stat("AGENTS.md"); !os.IsNotExist(err) {
		t.Errorf("expected AGENTS.md not to exist on disk")
	}
}

func TestCLI_Init_Minimal_File_Default(t *testing.T) {
	t.Chdir(t.TempDir())

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"init", "--minimal", "--file", "AGENTS.md"}, &stdout, &stderr, "dev")
	if err != nil {
		t.Fatalf("expected init --minimal --file to succeed, got: %v\nStderr: %s", err, stderr.String())
	}

	expectedMsg := "Initialized standard AGENTS.md at AGENTS.md\n"
	if stdout.String() != expectedMsg {
		t.Errorf("expected stdout %q, got %q", expectedMsg, stdout.String())
	}

	info, err := os.Stat("AGENTS.md")
	if err != nil {
		t.Fatalf("failed to stat created AGENTS.md: %v", err)
	}
	if info.Mode().Perm() != 0644 {
		t.Errorf("expected created AGENTS.md permissions to be 0644, got %o", info.Mode().Perm())
	}

	data, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("failed to read created AGENTS.md: %v", err)
	}
	if string(data) != cli.MinimalLivingMemoryTemplate() {
		t.Errorf("created AGENTS.md content does not match MinimalLivingMemoryTemplate")
	}
}

func TestCLI_Init_Minimal_Collision(t *testing.T) {
	t.Chdir(t.TempDir())

	initialContent := "pre-existing content"
	if err := os.WriteFile("AGENTS.md", []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to write dummy AGENTS.md: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"init", "--minimal", "--file", "AGENTS.md"}, &stdout, &stderr, "dev")
	if err == nil {
		t.Fatal("expected collision error, got nil")
	}

	expectedErr := "AGENTS.md already exists; use --force to overwrite"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}

	data, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("failed to read AGENTS.md: %v", err)
	}
	if string(data) != initialContent {
		t.Errorf("file was unexpectedly modified: got %q", string(data))
	}
}

func TestCLI_Init_Minimal_Force(t *testing.T) {
	t.Chdir(t.TempDir())

	initialContent := "pre-existing content"
	if err := os.WriteFile("AGENTS.md", []byte(initialContent), 0777); err != nil {
		t.Fatalf("failed to write dummy AGENTS.md: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := cli.Execute([]string{"init", "--minimal", "--file", "AGENTS.md", "--force"}, &stdout, &stderr, "dev")
	if err != nil {
		t.Fatalf("expected --force init to succeed, got: %v\nStderr: %s", err, stderr.String())
	}

	info, err := os.Stat("AGENTS.md")
	if err != nil {
		t.Fatalf("failed to stat overwritten AGENTS.md: %v", err)
	}
	if info.Mode().Perm() != 0644 {
		t.Errorf("expected overwritten AGENTS.md permissions to be 0644, got %o", info.Mode().Perm())
	}

	data, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("failed to read overwritten AGENTS.md: %v", err)
	}
	if string(data) != cli.MinimalLivingMemoryTemplate() {
		t.Errorf("overwritten AGENTS.md content does not match MinimalLivingMemoryTemplate")
	}
}

func TestCLI_Init_Minimal_WriterErrorPropagation(t *testing.T) {
	t.Parallel()

	err := cli.WriteLivingMemoryTemplate(failingWriter{}, true)
	if err == nil {
		t.Fatal("expected write error, got nil")
	}
	if !strings.Contains(err.Error(), "simulated write failure") {
		t.Errorf("expected simulated write failure, got: %v", err)
	}
}

func TestCLI_Init_Minimal_OutputSizeBudget(t *testing.T) {
	t.Parallel()

	minimal := cli.MinimalLivingMemoryTemplate()
	standard := cli.DefaultLivingMemoryTemplate()

	lines := strings.Split(strings.TrimSpace(minimal), "\n")
	if len(lines) > 50 {
		t.Errorf("minimal template exceeds line budget (<=50 lines): got %d lines", len(lines))
	}

	byteLen := len([]byte(minimal))
	if byteLen > 2500 {
		t.Errorf("minimal template exceeds byte budget (<=2500 bytes): got %d bytes", byteLen)
	}

	if byteLen >= len([]byte(standard))/4 {
		t.Errorf("minimal template does not achieve >75%% size reduction (<25%% of standard): minimal=%d, standard=%d, max_allowed=%d", byteLen, len([]byte(standard)), len([]byte(standard))/4)
	}
}

func TestCLI_Init_Minimal_Invariants(t *testing.T) {
	t.Parallel()

	minimal := cli.MinimalLivingMemoryTemplate()

	// 1. Strict ASCII byte check
	for i, b := range []byte(minimal) {
		if b > 127 {
			t.Errorf("minimal template contains non-ASCII byte %d at index %d", b, i)
		}
	}

	// 2. Exact substring assertions
	requiredStrings := []string{
		"Peer-Session Primacy",
		"invoke_subagent",
		"Blind Barrier",
		"herdr",
		"planner", "reviewer", "builder", "scout", "navigator", "cartographer",
		"init", "plan", "blueprint", "build", "review", "commit", "cartography", "session-handoff",
		"sub-review-resolution",
		"Please send this requirement directly to Planner",
		"[Source: <path> | Observed: <rev> @ <timestamp>]",
		"idle/done only",
		"discard on arrival",
		"no retry/queue",
		"max 1 in-flight",
		"<500 chars",
		"static fallback",
		"docs/diagrams/<safe-name>.html",
		"traversal prohibited",
		"diagram-completion",
		"file URI",
		"single-sentence digest",
		"node/edge statistics",
		"<100 tokens",
		"<=60 words",
		"<=250 runes",
		"zero inline markup",
		"Taste Gate",
		"<=12 nodes",
		"<=12 transitions",
		"ADVISORY_ISSUED",
		"async non-blocking",
		"diagram-design",
		"PLAN_PASS",
		"REVIEW_PASS",
	}

	for _, s := range requiredStrings {
		if !strings.Contains(minimal, s) {
			t.Errorf("minimal template missing required invariant %q", s)
		}
	}
}

func TestCLI_Init_TemplateContent(t *testing.T) {
	t.Parallel()

	template := cli.DefaultLivingMemoryTemplate()

	// 1. Verify ASCII-only requirement
	for i, b := range []byte(template) {
		if b > 127 {
			t.Errorf("template contains non-ASCII byte %d at index %d", b, i)
		}
	}

	// 2. Verify mandatory structural strings
	requiredStrings := []string{
		"AgentPlaybook v0.3.4 Living Memory Blueprint",
		"Peer-Session Primacy over Subagents",
		"invoke_subagent",
		"Blind Barrier",
		"cartographer",
		"navigator",
		"session-handoff",
		"init",
		"plan",
		"blueprint",
		"build",
		"review",
		"commit",
		"cartography",
		"planner",
		"reviewer",
		"builder",
		"scout",
	}

	for _, s := range requiredStrings {
		if !strings.Contains(template, s) {
			t.Errorf("template missing required string %q", s)
		}
	}

	// 3. Verify synchronized release header against scripts/VERSION
	versionBytes, err := os.ReadFile("../../scripts/VERSION")
	if err != nil {
		t.Fatalf("failed to read scripts/VERSION: %v", err)
	}
	currentVer := strings.TrimSpace(string(versionBytes))
	expectedHeader := "AgentPlaybook " + currentVer + " Living Memory Blueprint"
	if !strings.Contains(template, expectedHeader) {
		t.Errorf("template does not contain synchronized release header %q", expectedHeader)
	}
}
