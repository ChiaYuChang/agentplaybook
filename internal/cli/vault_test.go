package cli_test

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ChiaYuChang/agentplaybook/internal/cli"
)

func TestCanonicalRepoID_RemoteNormalization(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		rawURL   string
		expected string
	}{
		{
			name:     "standard https with .git",
			rawURL:   "https://github.com/owner/repo.git",
			expected: "github.com/owner/repo",
		},
		{
			name:     "https without .git",
			rawURL:   "https://github.com/owner/repo",
			expected: "github.com/owner/repo",
		},
		{
			name:     "https with credentials",
			rawURL:   "https://user:secret123@github.com/owner/repo.git",
			expected: "github.com/owner/repo",
		},
		{
			name:     "ssh scp-style with user",
			rawURL:   "git@github.com:owner/repo.git",
			expected: "github.com/owner/repo",
		},
		{
			name:     "ssh scp-style without user",
			rawURL:   "github.com:owner/repo.git",
			expected: "github.com/owner/repo",
		},
		{
			name:     "ssh scheme with user and port",
			rawURL:   "ssh://git@github.com:22/owner/repo.git",
			expected: "github.com/owner/repo",
		},
		{
			name:     "git scheme",
			rawURL:   "git://github.com/owner/repo.git",
			expected: "github.com/owner/repo",
		},
		{
			name:     "gitlab nested subgroups with trailing slash",
			rawURL:   "https://gitlab.example.com/group/subgroup/repo.git/",
			expected: "gitlab.example.com/group/subgroup/repo",
		},
		{
			name:     "mixed case normalized to lowercase",
			rawURL:   "https://GitHub.com/Owner/Repo.GIT",
			expected: "github.com/owner/repo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := cli.NormalizeRemoteURL(tc.rawURL)
			if got != tc.expected {
				t.Errorf("NormalizeRemoteURL(%q) = %q; want %q", tc.rawURL, got, tc.expected)
			}
		})
	}
}

func TestCanonicalRepoID_LocalIsolation(t *testing.T) {
	tempBase := t.TempDir()
	dirA := filepath.Join(tempBase, "team-a", "api")
	dirB := filepath.Join(tempBase, "team-b", "api")

	if err := os.MkdirAll(dirA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dirB, 0755); err != nil {
		t.Fatal(err)
	}

	idA, err := cli.CanonicalRepoID(dirA)
	if err != nil {
		t.Fatalf("CanonicalRepoID(dirA) failed: %v", err)
	}
	idB, err := cli.CanonicalRepoID(dirB)
	if err != nil {
		t.Fatalf("CanonicalRepoID(dirB) failed: %v", err)
	}

	if idA == idB {
		t.Errorf("expected distinct local IDs for unrelated repos with same basename, got identical: %q", idA)
	}
	if !strings.HasPrefix(idA, "local:") || !strings.HasPrefix(idB, "local:") {
		t.Errorf("expected local: prefix, got idA=%q, idB=%q", idA, idB)
	}
}

func TestCanonicalRepoID_GitConfigParsing(t *testing.T) {
	repoDir := t.TempDir()
	dotGit := filepath.Join(repoDir, ".git")
	if err := os.MkdirAll(dotGit, 0755); err != nil {
		t.Fatal(err)
	}

	configContent := `[core]
	repositoryformatversion = 0
	filemode = true
[remote "origin"]
	url = https://token:secret@github.com/myorg/myrepo.git
	fetch = +refs/heads/*:refs/remotes/origin/*
`
	if err := os.WriteFile(filepath.Join(dotGit, "config"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	id, err := cli.CanonicalRepoID(repoDir)
	if err != nil {
		t.Fatalf("CanonicalRepoID failed: %v", err)
	}
	expected := "github.com/myorg/myrepo"
	if id != expected {
		t.Errorf("CanonicalRepoID = %q; want %q", id, expected)
	}
}

func TestFindRepoRoot(t *testing.T) {
	t.Parallel()

	t.Run("git directory", func(t *testing.T) {
		tempDir := t.TempDir()
		repoRoot := filepath.Join(tempDir, "repo")
		subDir := filepath.Join(repoRoot, "sub", "pkg")
		if err := os.MkdirAll(filepath.Join(repoRoot, ".git"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatal(err)
		}

		found, err := cli.FindRepoRoot(subDir)
		if err != nil {
			t.Fatalf("FindRepoRoot failed: %v", err)
		}
		if found != repoRoot {
			t.Errorf("FindRepoRoot(%q) = %q; want %q", subDir, found, repoRoot)
		}
	})

	t.Run("git worktree file", func(t *testing.T) {
		tempDir := t.TempDir()
		mainRepo := filepath.Join(tempDir, "main-repo")
		mainGit := filepath.Join(mainRepo, ".git")
		worktreeGitDir := filepath.Join(mainGit, "worktrees", "wt1")
		worktreeDir := filepath.Join(tempDir, "wt1")

		if err := os.MkdirAll(worktreeGitDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(worktreeDir, 0755); err != nil {
			t.Fatal(err)
		}

		// .git regular file pointing to main repo worktree gitdir
		dotGitFile := filepath.Join(worktreeDir, ".git")
		if err := os.WriteFile(dotGitFile, []byte("gitdir: "+worktreeGitDir+"\n"), 0644); err != nil {
			t.Fatal(err)
		}

		// main config in mainGit
		configContent := `[remote "origin"]
	url = git@github.com:acme/project.git
`
		if err := os.WriteFile(filepath.Join(worktreeGitDir, "commondir"), []byte("../../\n"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(mainGit, "config"), []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		found, err := cli.FindRepoRoot(worktreeDir)
		if err != nil {
			t.Fatalf("FindRepoRoot failed: %v", err)
		}
		if found != worktreeDir {
			t.Errorf("FindRepoRoot worktree = %q; want %q", found, worktreeDir)
		}

		id, err := cli.CanonicalRepoID(worktreeDir)
		if err != nil {
			t.Fatalf("CanonicalRepoID worktree failed: %v", err)
		}
		if id != "github.com/acme/project" {
			t.Errorf("CanonicalRepoID worktree = %q; want github.com/acme/project", id)
		}
	})

	t.Run("jj directory", func(t *testing.T) {
		tempDir := t.TempDir()
		repoRoot := filepath.Join(tempDir, "jj-repo")
		subDir := filepath.Join(repoRoot, "src")
		if err := os.MkdirAll(filepath.Join(repoRoot, ".jj"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatal(err)
		}

		found, err := cli.FindRepoRoot(subDir)
		if err != nil {
			t.Fatalf("FindRepoRoot failed: %v", err)
		}
		if found != repoRoot {
			t.Errorf("FindRepoRoot jj = %q; want %q", found, repoRoot)
		}
	})

	t.Run("fallback when no VCS directory", func(t *testing.T) {
		tempDir := t.TempDir()
		found, err := cli.FindRepoRoot(tempDir)
		if err != nil {
			t.Fatalf("FindRepoRoot fallback failed: %v", err)
		}
		if found != tempDir {
			t.Errorf("FindRepoRoot fallback = %q; want %q", found, tempDir)
		}
	})
}

func TestResolveVault_PrecedenceAndNormalization(t *testing.T) {
	tempHome := t.TempDir()
	vaultCustom := t.TempDir()
	projectDir := t.TempDir()

	t.Setenv("HOME", tempHome)

	// Default fallback: ~/.agentplaybook
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", "")
	t.Setenv("AGENTPLAYBOOK_HOME", "")
	paths, err := cli.ResolveVault(projectDir, cli.VaultResolveOptions{})
	if err != nil {
		t.Fatalf("ResolveVault default failed: %v", err)
	}
	expectedDefaultRoot := filepath.Join(tempHome, ".agentplaybook")
	if paths.VaultRoot != expectedDefaultRoot {
		t.Errorf("paths.VaultRoot = %q; want %q", paths.VaultRoot, expectedDefaultRoot)
	}

	// AGENTPLAYBOOK_HOME precedence
	customHome := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_HOME", customHome)
	paths, err = cli.ResolveVault(projectDir, cli.VaultResolveOptions{})
	if err != nil {
		t.Fatalf("ResolveVault AGENTPLAYBOOK_HOME failed: %v", err)
	}
	if paths.VaultRoot != customHome {
		t.Errorf("paths.VaultRoot = %q; want %q", paths.VaultRoot, customHome)
	}

	// AGENTPLAYBOOK_VAULT_ROOT highest precedence
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultCustom)
	paths, err = cli.ResolveVault(projectDir, cli.VaultResolveOptions{})
	if err != nil {
		t.Fatalf("ResolveVault AGENTPLAYBOOK_VAULT_ROOT failed: %v", err)
	}
	if paths.VaultRoot != vaultCustom {
		t.Errorf("paths.VaultRoot = %q; want %q", paths.VaultRoot, vaultCustom)
	}

	// Normalization of relative paths
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", "./relative-vault-root")
	paths, err = cli.ResolveVault(projectDir, cli.VaultResolveOptions{})
	if err != nil {
		t.Fatalf("ResolveVault relative path failed: %v", err)
	}
	if !filepath.IsAbs(paths.VaultRoot) {
		t.Errorf("paths.VaultRoot %q is not absolute", paths.VaultRoot)
	}
}

func TestResolveVault_AbsoluteContainment(t *testing.T) {
	projectDir := t.TempDir()

	t.Run("nested_dot_agentplaybook", func(t *testing.T) {
		nestedVault := filepath.Join(projectDir, ".agentplaybook")
		t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", nestedVault)
		_, err := cli.ResolveVault(projectDir, cli.VaultResolveOptions{})
		if err == nil {
			t.Fatal("expected error for vault nested inside project root, got nil")
		}
		if !strings.Contains(err.Error(), "cannot be inside project root") {
			t.Errorf("expected containment error, got: %v", err)
		}
	})

	t.Run("nested_dotdot_prefix_directory", func(t *testing.T) {
		// Regression test: directory starting with '..' like '..vault' inside project root
		dotDotVault := filepath.Join(projectDir, "..vault")
		t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", dotDotVault)
		_, err := cli.ResolveVault(projectDir, cli.VaultResolveOptions{})
		if err == nil {
			t.Fatal("expected error for vault with '..'-prefixed dir inside project root, got nil")
		}
		if !strings.Contains(err.Error(), "cannot be inside project root") {
			t.Errorf("expected containment error for dot-dot prefix dir, got: %v", err)
		}
	})
}

func TestResolveVault_SymlinkRejection(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping symlink tests on windows")
	}

	tempDir := t.TempDir()
	realVault := filepath.Join(tempDir, "real-vault")
	symlinkVault := filepath.Join(tempDir, "symlink-vault")
	projectDir := filepath.Join(tempDir, "my-project")

	if err := os.MkdirAll(realVault, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realVault, symlinkVault); err != nil {
		t.Fatal(err)
	}

	// Symlinked vault root fails closed
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", symlinkVault)
	_, err := cli.ResolveVault(projectDir, cli.VaultResolveOptions{})
	if err == nil {
		t.Fatal("expected error for symlinked vault root, got nil")
	}
	if !strings.Contains(err.Error(), "vault path component cannot be a symlink") {
		t.Errorf("expected symlink rejection error, got: %v", err)
	}
}

func TestResolveVault_ProjectSlugValidation(t *testing.T) {
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)
	projectDir := t.TempDir()

	invalidSlugs := []string{
		"../escape",
		"foo/bar",
		"foo\\bar",
		"has space",
		"-leading-dash",
		".dotstart",
		"",
	}

	for _, slug := range invalidSlugs {
		t.Run("slug_"+slug, func(t *testing.T) {
			t.Setenv("AGENTPLAYBOOK_PROJECT_NAME", slug)
			_, err := cli.ResolveVault(projectDir, cli.VaultResolveOptions{})
			if err == nil {
				t.Fatalf("expected error for invalid slug %q, got nil", slug)
			}
			if !strings.Contains(err.Error(), "invalid project name") {
				t.Errorf("expected invalid project name error, got: %v", err)
			}
		})
	}
}

func TestEnsureVaultDirectories_AtomicRollback(t *testing.T) {
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)

	paths := &cli.VaultPaths{
		VaultRoot:   vaultRoot,
		ProjectName: "atomic-test",
		PlanDir:     filepath.Join(vaultRoot, "plan", "atomic-test"),
		// Force E2EDir to fail by creating a regular file at its parent location
		E2EDir: filepath.Join(vaultRoot, "e2e-blocked-file", "atomic-test"),
	}

	// Create a regular file where directory is supposed to be
	blockedParent := filepath.Join(vaultRoot, "e2e-blocked-file")
	if err := os.WriteFile(blockedParent, []byte("blocker"), 0644); err != nil {
		t.Fatal(err)
	}

	err := cli.EnsureVaultDirectories(paths, "local:/test/repo", "local:/test/repo", cli.VaultResolveOptions{})
	if err == nil {
		t.Fatal("expected EnsureVaultDirectories to fail, got nil")
	}

	// Verify PlanDir was rolled back and does not exist on disk
	if _, err := os.Stat(paths.PlanDir); !os.IsNotExist(err) {
		t.Errorf("expected PlanDir %s to be rolled back, but it exists", paths.PlanDir)
	}
}

func TestEnsureVaultDirectories_PermissionsAndBinding(t *testing.T) {
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)
	projectDir := t.TempDir()

	paths, err := cli.ResolveVault(projectDir, cli.VaultResolveOptions{})
	if err != nil {
		t.Fatalf("ResolveVault failed: %v", err)
	}

	repoRoot := projectDir
	repoID := "local:" + repoRoot
	if err := cli.EnsureVaultDirectories(paths, repoRoot, repoID, cli.VaultResolveOptions{}); err != nil {
		t.Fatalf("EnsureVaultDirectories failed: %v", err)
	}

	// Verify plan directory mode 0750
	planInfo, err := os.Stat(paths.PlanDir)
	if err != nil {
		t.Fatalf("failed to stat PlanDir: %v", err)
	}
	if planInfo.Mode().Perm() != 0750 {
		t.Errorf("PlanDir mode = %o; want 0750", planInfo.Mode().Perm())
	}

	// Verify e2e directory mode 0750
	e2eInfo, err := os.Stat(paths.E2EDir)
	if err != nil {
		t.Fatalf("failed to stat E2EDir: %v", err)
	}
	if e2eInfo.Mode().Perm() != 0750 {
		t.Errorf("E2EDir mode = %o; want 0750", e2eInfo.Mode().Perm())
	}

	// Verify descriptor in plan directory
	planDescPath := filepath.Join(paths.PlanDir, ".vault-binding.json")
	data, err := os.ReadFile(planDescPath)
	if err != nil {
		t.Fatalf("failed to read plan descriptor: %v", err)
	}
	var desc cli.VaultBindingDescriptor
	if err := json.Unmarshal(data, &desc); err != nil {
		t.Fatalf("failed to unmarshal descriptor: %v", err)
	}
	if desc.RepositoryID != repoID {
		t.Errorf("descriptor.RepositoryID = %q; want %q", desc.RepositoryID, repoID)
	}
	if desc.RepositoryRoot != repoRoot {
		t.Errorf("descriptor.RepositoryRoot = %q; want %q", desc.RepositoryRoot, repoRoot)
	}

	// Verify idempotency: running again succeeds without error
	if err := cli.EnsureVaultDirectories(paths, repoRoot, repoID, cli.VaultResolveOptions{}); err != nil {
		t.Errorf("second EnsureVaultDirectories failed: %v", err)
	}
}

func TestResolveVault_UnboundAndAdopt(t *testing.T) {
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)
	projectDir := t.TempDir()
	projectName := filepath.Base(projectDir)

	planDir := filepath.Join(vaultRoot, "plan", projectName)
	if err := os.MkdirAll(planDir, 0750); err != nil {
		t.Fatal(err)
	}
	// PlanDir exists without .vault-binding.json

	// 1. Without adopt: fails closed with ErrVaultUnbound
	_, err := cli.ResolveVault(projectDir, cli.VaultResolveOptions{Adopt: false})
	if err == nil {
		t.Fatal("expected ErrVaultUnbound, got nil")
	}
	if !errors.Is(err, cli.ErrVaultUnbound) {
		t.Errorf("expected errors.Is ErrVaultUnbound, got: %v", err)
	}

	// 2. With adopt: succeeds
	paths, err := cli.ResolveVault(projectDir, cli.VaultResolveOptions{Adopt: true})
	if err != nil {
		t.Fatalf("ResolveVault with adopt failed: %v", err)
	}
	if paths == nil {
		t.Fatal("expected non-nil paths with adopt")
	}

	// 3. EnsureVaultDirectories binds the adopted directory
	repoRoot := projectDir
	repoID := "local:" + repoRoot
	if err := cli.EnsureVaultDirectories(paths, repoRoot, repoID, cli.VaultResolveOptions{Adopt: true}); err != nil {
		t.Fatalf("EnsureVaultDirectories with adopt failed: %v", err)
	}

	// 4. Now subsequent ResolveVault without adopt flag succeeds
	paths2, err := cli.ResolveVault(projectDir, cli.VaultResolveOptions{Adopt: false})
	if err != nil {
		t.Fatalf("ResolveVault after adoption failed: %v", err)
	}
	if paths2 == nil {
		t.Fatal("expected non-nil paths after adoption")
	}
}

func TestResolveVault_Relocation(t *testing.T) {
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)

	oldRepoRoot := t.TempDir()
	newRepoRoot := t.TempDir()

	// Simulate remote repository with identical remote.origin.url in both old and new roots
	setupFakeGitRepo(t, oldRepoRoot, "https://github.com/myorg/relocated-project.git")
	setupFakeGitRepo(t, newRepoRoot, "https://github.com/myorg/relocated-project.git")

	t.Setenv("AGENTPLAYBOOK_PROJECT_NAME", "relocated-project")

	// 1. Initialize vault at old location
	paths, err := cli.ResolveVault(oldRepoRoot, cli.VaultResolveOptions{})
	if err != nil {
		t.Fatalf("initial ResolveVault failed: %v", err)
	}
	repoID := "github.com/myorg/relocated-project"
	if err := cli.EnsureVaultDirectories(paths, oldRepoRoot, repoID, cli.VaultResolveOptions{}); err != nil {
		t.Fatalf("initial EnsureVaultDirectories failed: %v", err)
	}

	// 2. Resolve at new location: matching remote ID authorizes automatic relocation
	pathsNew, err := cli.ResolveVault(newRepoRoot, cli.VaultResolveOptions{})
	if err != nil {
		t.Fatalf("ResolveVault at new location failed: %v", err)
	}
	if pathsNew == nil {
		t.Fatal("expected non-nil paths")
	}

	// Verify descriptor was updated to newRepoRoot
	descData, err := os.ReadFile(filepath.Join(pathsNew.PlanDir, ".vault-binding.json"))
	if err != nil {
		t.Fatal(err)
	}
	var desc cli.VaultBindingDescriptor
	if err := json.Unmarshal(descData, &desc); err != nil {
		t.Fatal(err)
	}
	if desc.RepositoryRoot != newRepoRoot {
		t.Errorf("descriptor.RepositoryRoot = %q; want %q", desc.RepositoryRoot, newRepoRoot)
	}
	if desc.RepositoryID != repoID {
		t.Errorf("descriptor.RepositoryID = %q; want %q", desc.RepositoryID, repoID)
	}
}

func TestResolveVault_CollisionAndRebind(t *testing.T) {
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)

	oldLocalRepo := t.TempDir()
	newLocalRepo := t.TempDir()

	t.Setenv("AGENTPLAYBOOK_PROJECT_NAME", "shared-slug")

	// 1. Initialize vault for old local repo
	paths, err := cli.ResolveVault(oldLocalRepo, cli.VaultResolveOptions{})
	if err != nil {
		t.Fatalf("initial ResolveVault failed: %v", err)
	}
	oldID := "local:" + oldLocalRepo
	if err := cli.EnsureVaultDirectories(paths, oldLocalRepo, oldID, cli.VaultResolveOptions{}); err != nil {
		t.Fatalf("initial EnsureVaultDirectories failed: %v", err)
	}

	// 2. Resolve at new local repo: mismatching local ID triggers ErrVaultCollision
	_, err = cli.ResolveVault(newLocalRepo, cli.VaultResolveOptions{Rebind: false})
	if err == nil {
		t.Fatal("expected ErrVaultCollision, got nil")
	}
	if !errors.Is(err, cli.ErrVaultCollision) {
		t.Errorf("expected ErrVaultCollision, got: %v", err)
	}

	// 3. Authorize explicit rebind: succeeds and updates descriptors
	pathsRebound, err := cli.ResolveVault(newLocalRepo, cli.VaultResolveOptions{Rebind: true})
	if err != nil {
		t.Fatalf("ResolveVault with rebind failed: %v", err)
	}
	if pathsRebound == nil {
		t.Fatal("expected non-nil paths on rebind")
	}

	newID := "local:" + newLocalRepo
	descData, err := os.ReadFile(filepath.Join(pathsRebound.PlanDir, ".vault-binding.json"))
	if err != nil {
		t.Fatal(err)
	}
	var desc cli.VaultBindingDescriptor
	if err := json.Unmarshal(descData, &desc); err != nil {
		t.Fatal(err)
	}
	if desc.RepositoryRoot != newLocalRepo {
		t.Errorf("rebound RepositoryRoot = %q; want %q", desc.RepositoryRoot, newLocalRepo)
	}
	if desc.RepositoryID != newID {
		t.Errorf("rebound RepositoryID = %q; want %q", desc.RepositoryID, newID)
	}
}

func TestUpdateVaultDescriptors_PairedRollback(t *testing.T) {
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)

	paths := &cli.VaultPaths{
		VaultRoot:   vaultRoot,
		ProjectName: "rollback-test",
		PlanDir:     filepath.Join(vaultRoot, "plan", "rollback-test"),
		E2EDir:      filepath.Join(vaultRoot, "e2e", "rollback-test"),
	}

	if err := cli.EnsureVaultDirectories(paths, "/initial/repo", "local:/initial/repo", cli.VaultResolveOptions{}); err != nil {
		t.Fatal(err)
	}

	planDescPath := filepath.Join(paths.PlanDir, ".vault-binding.json")
	initialPlanBytes, err := os.ReadFile(planDescPath)
	if err != nil {
		t.Fatal(err)
	}

	// Replace E2E descriptor with a directory: causes a non-NotExist read error.
	// With the new fail-closed pre-read, UpdateVaultDescriptors returns before any mutation.
	e2eDescPath := filepath.Join(paths.E2EDir, ".vault-binding.json")
	_ = os.Remove(e2eDescPath)
	if err := os.MkdirAll(e2eDescPath, 0755); err != nil {
		t.Fatal(err)
	}

	err = cli.UpdateVaultDescriptors(paths, "/updated/repo", "local:/updated/repo")
	if err == nil {
		t.Fatal("expected UpdateVaultDescriptors to fail, got nil")
	}

	// Assert Plan descriptor was NOT mutated (fail-early: no writes occurred)
	currentPlanBytes, err := os.ReadFile(planDescPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(currentPlanBytes) != string(initialPlanBytes) {
		t.Errorf("PlanDir descriptor was mutated despite fail-early!\ngot:\n%s\nwant:\n%s", string(currentPlanBytes), string(initialPlanBytes))
	}

	// Assert E2E blocker entry (directory) is still intact — zero mutation occurred
	e2eInfo, err := os.Stat(e2eDescPath)
	if err != nil {
		t.Fatalf("E2E blocker entry was removed or inaccessible after fail-early: %v", err)
	}
	if !e2eInfo.IsDir() {
		t.Errorf("E2E blocker entry is no longer a directory; unexpected mutation occurred")
	}
}

func TestEnsureVaultDirectories_RollbackPreservesExistingPlanDescriptor(t *testing.T) {
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)

	paths := &cli.VaultPaths{
		VaultRoot:   vaultRoot,
		ProjectName: "preserve-test",
		PlanDir:     filepath.Join(vaultRoot, "plan", "preserve-test"),
		E2EDir:      filepath.Join(vaultRoot, "e2e", "preserve-test"),
	}

	// 1. Initial setup: successfully bind PlanDir and E2EDir with original descriptor
	initialRepo := "/original/repo"
	initialID := "local:" + initialRepo
	if err := cli.EnsureVaultDirectories(paths, initialRepo, initialID, cli.VaultResolveOptions{}); err != nil {
		t.Fatalf("initial EnsureVaultDirectories failed: %v", err)
	}

	planDescPath := filepath.Join(paths.PlanDir, ".vault-binding.json")
	originalPlanBytes, err := os.ReadFile(planDescPath)
	if err != nil {
		t.Fatalf("failed to read original plan descriptor: %v\n", err)
	}

	// 2. Induce failure on reading the E2E descriptor:
	// Replace E2E descriptor file with a directory so os.ReadFile fails (not-exist → is-dir error).
	// This triggers the new fail-closed pre-read path, ensuring no mutation occurs.
	e2eDescPath := filepath.Join(paths.E2EDir, ".vault-binding.json")
	_ = os.Remove(e2eDescPath)
	if err := os.MkdirAll(e2eDescPath, 0755); err != nil {
		t.Fatalf("failed to create blocker directory at e2eDescPath: %v", err)
	}

	// 3. Attempt EnsureVaultDirectories with new repo values; must fail because E2E descriptor
	// cannot be safely read for backup (directory at descriptor path → non-NotExist error).
	newRepo := "/attempted/repo"
	newID := "local:" + newRepo
	err = cli.EnsureVaultDirectories(paths, newRepo, newID, cli.VaultResolveOptions{})
	if err == nil {
		t.Fatal("expected EnsureVaultDirectories to fail when E2E descriptor is blocked, got nil")
	}

	// 4. Assert that the pre-existing Plan descriptor was NOT destroyed or removed,
	// and was preserved with its exact original content (fail-early: no writes occurred).
	preservedPlanBytes, err := os.ReadFile(planDescPath)
	if err != nil {
		t.Fatalf("pre-existing plan descriptor was destroyed or missing after fail-early: %v", err)
	}
	if string(preservedPlanBytes) != string(originalPlanBytes) {
		t.Errorf("Plan descriptor was modified despite fail-early!\ngot:\n%s\nwant original:\n%s",
			string(preservedPlanBytes), string(originalPlanBytes))
	}

	// 5. Assert that the E2E blocker (directory we planted) is still a directory — no mutation occurred.
	e2eInfo, err := os.Stat(e2eDescPath)
	if err != nil {
		t.Fatalf("E2E blocker entry was removed or became inaccessible: %v", err)
	}
	if !e2eInfo.IsDir() {
		t.Errorf("E2E blocker entry is no longer a directory; unexpected mutation occurred")
	}
}

func setupFakeGitRepo(t *testing.T, dir string, remoteURL string) {
	t.Helper()
	dotGit := filepath.Join(dir, ".git")
	if err := os.MkdirAll(dotGit, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := `[remote "origin"]
	url = ` + remoteURL + `
`
	if err := os.WriteFile(filepath.Join(dotGit, "config"), []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}
}

// Ensure mock exec git works without external git dependency
var _ = exec.Command
