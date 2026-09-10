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

func TestResolveVault_ProjectFirstLayout(t *testing.T) {
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)
	t.Setenv("AGENTPLAYBOOK_PROJECT_NAME", "prism")

	projectDir := t.TempDir()
	paths, err := cli.ResolveVault(projectDir, cli.VaultResolveOptions{})
	if err != nil {
		t.Fatalf("ResolveVault failed: %v", err)
	}
	if paths.ProjectDir != filepath.Join(vaultRoot, "prism") {
		t.Errorf("ProjectDir = %q; want %q", paths.ProjectDir, filepath.Join(vaultRoot, "prism"))
	}
	if paths.PlanDir != filepath.Join(vaultRoot, "prism", "plan") {
		t.Errorf("PlanDir = %q; want Join(vaultRoot, name, plan)", paths.PlanDir)
	}
	if paths.E2EDir != filepath.Join(vaultRoot, "prism", "e2e") {
		t.Errorf("E2EDir = %q; want Join(vaultRoot, name, e2e)", paths.E2EDir)
	}
}

func TestEnsureVaultDirectories_AtomicRollback(t *testing.T) {
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)

	// Block project directory creation with a regular file: root write must fail.
	projectDir := filepath.Join(vaultRoot, "atomic-test")
	if err := os.WriteFile(projectDir, []byte("blocker"), 0644); err != nil {
		t.Fatal(err)
	}
	paths := &cli.VaultPaths{
		VaultRoot:   vaultRoot,
		ProjectName: "atomic-test",
		ProjectDir:  projectDir,
		PlanDir:     filepath.Join(projectDir, "plan"),
		E2EDir:      filepath.Join(projectDir, "e2e"),
	}

	err := cli.EnsureVaultDirectories(paths, "local:/test/repo", "local:/test/repo", cli.VaultResolveOptions{})
	if err == nil {
		t.Fatal("expected EnsureVaultDirectories to fail, got nil")
	}

	// Verify no vault directories were created (any stat error means absent).
	if _, err := os.Stat(paths.PlanDir); err == nil {
		t.Errorf("expected PlanDir %s to be rolled back, but it exists", paths.PlanDir)
	}
	// ProjectDir path remains the blocker file, not a directory.
	if fi, err := os.Stat(projectDir); err != nil || fi.IsDir() {
		t.Errorf("expected ProjectDir blocker file to remain, stat err=%v", err)
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

	// Verify project directory mode 0750
	projInfo, err := os.Stat(paths.ProjectDir)
	if err != nil {
		t.Fatalf("failed to stat ProjectDir: %v", err)
	}
	if projInfo.Mode().Perm() != 0750 {
		t.Errorf("ProjectDir mode = %o; want 0750", projInfo.Mode().Perm())
	}

	// Verify single root descriptor; no child descriptors.
	rootDescPath := filepath.Join(paths.ProjectDir, ".vault-binding.json")
	data, err := os.ReadFile(rootDescPath)
	if err != nil {
		t.Fatalf("failed to read root descriptor: %v", err)
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
	if _, err := os.Stat(filepath.Join(paths.PlanDir, ".vault-binding.json")); !os.IsNotExist(err) {
		t.Errorf("expected no child descriptor under plan dir")
	}
	if _, err := os.Stat(filepath.Join(paths.E2EDir, ".vault-binding.json")); !os.IsNotExist(err) {
		t.Errorf("expected no child descriptor under e2e dir")
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

	projVaultDir := filepath.Join(vaultRoot, projectName)
	if err := os.MkdirAll(projVaultDir, 0750); err != nil {
		t.Fatal(err)
	}
	// ProjectDir exists without .vault-binding.json

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

	// Verify single root descriptor was updated to newRepoRoot
	descData, err := os.ReadFile(filepath.Join(pathsNew.ProjectDir, ".vault-binding.json"))
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
	descData, err := os.ReadFile(filepath.Join(pathsRebound.ProjectDir, ".vault-binding.json"))
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

func TestUpdateVaultDescriptors_SingleRollback(t *testing.T) {
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)

	projectDir := filepath.Join(vaultRoot, "rollback-test")
	paths := &cli.VaultPaths{
		VaultRoot:   vaultRoot,
		ProjectName: "rollback-test",
		ProjectDir:  projectDir,
		PlanDir:     filepath.Join(projectDir, "plan"),
		E2EDir:      filepath.Join(projectDir, "e2e"),
	}

	if err := cli.EnsureVaultDirectories(paths, "/initial/repo", "local:/initial/repo", cli.VaultResolveOptions{}); err != nil {
		t.Fatal(err)
	}

	rootDescPath := filepath.Join(projectDir, ".vault-binding.json")
	initialRootBytes, err := os.ReadFile(rootDescPath)
	if err != nil {
		t.Fatal(err)
	}

	// Successful update changes root binding.
	if err := cli.UpdateVaultDescriptors(paths, "/updated/repo", "local:/updated/repo"); err != nil {
		t.Fatalf("UpdateVaultDescriptors failed: %v", err)
	}
	updatedBytes, err := os.ReadFile(rootDescPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(updatedBytes) == string(initialRootBytes) {
		t.Errorf("expected root descriptor to change after update")
	}

	// Replace root descriptor with a directory: fail-closed pre-read, zero mutation.
	_ = os.Remove(rootDescPath)
	if err := os.MkdirAll(rootDescPath, 0755); err != nil {
		t.Fatal(err)
	}
	err = cli.UpdateVaultDescriptors(paths, "/blocked/repo", "local:/blocked/repo")
	if err == nil {
		t.Fatal("expected UpdateVaultDescriptors to fail, got nil")
	}
	info, err := os.Stat(rootDescPath)
	if err != nil {
		t.Fatalf("root blocker entry was removed after fail-early: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("root blocker entry is no longer a directory; unexpected mutation occurred")
	}
}

func TestEnsureVaultDirectories_RollbackPreservesRootDescriptor(t *testing.T) {
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)

	projectDir := filepath.Join(vaultRoot, "preserve-test")
	paths := &cli.VaultPaths{
		VaultRoot:   vaultRoot,
		ProjectName: "preserve-test",
		ProjectDir:  projectDir,
		PlanDir:     filepath.Join(projectDir, "plan"),
		E2EDir:      filepath.Join(projectDir, "e2e"),
	}

	initialRepo := "/original/repo"
	initialID := "local:" + initialRepo
	if err := cli.EnsureVaultDirectories(paths, initialRepo, initialID, cli.VaultResolveOptions{}); err != nil {
		t.Fatalf("initial EnsureVaultDirectories failed: %v", err)
	}

	rootDescPath := filepath.Join(projectDir, ".vault-binding.json")
	originalRootBytes, err := os.ReadFile(rootDescPath)
	if err != nil {
		t.Fatalf("failed to read original root descriptor: %v\n", err)
	}

	// Block root descriptor with a directory: pre-read fails closed.
	_ = os.Remove(rootDescPath)
	if err := os.MkdirAll(rootDescPath, 0755); err != nil {
		t.Fatalf("failed to create blocker directory: %v", err)
	}

	err = cli.EnsureVaultDirectories(paths, "/attempted/repo", "local:/attempted/repo", cli.VaultResolveOptions{})
	if err == nil {
		t.Fatal("expected EnsureVaultDirectories to fail when root descriptor is blocked, got nil")
	}

	info, err := os.Stat(rootDescPath)
	if err != nil {
		t.Fatalf("root blocker entry was removed: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("root blocker entry is no longer a directory; unexpected mutation occurred")
	}
	_ = originalRootBytes
}

func writeLegacyBinding(t *testing.T, dir, repoID, repoRoot string) []byte {
	t.Helper()
	if err := os.MkdirAll(dir, 0750); err != nil {
		t.Fatal(err)
	}
	desc := cli.VaultBindingDescriptor{RepositoryID: repoID, RepositoryRoot: repoRoot, UpdatedAt: "2026-01-01T00:00:00Z"}
	data, err := json.MarshalIndent(desc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(dir, ".vault-binding.json"), data, 0640); err != nil {
		t.Fatal(err)
	}
	return data
}

func TestMigrateLegacy_MatchMatch(t *testing.T) {
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)
	repoRoot := t.TempDir()
	repoID := "local:" + repoRoot
	t.Setenv("AGENTPLAYBOOK_PROJECT_NAME", "prism")

	legacyPlan := filepath.Join(vaultRoot, "plan", "prism")
	legacyE2E := filepath.Join(vaultRoot, "e2e", "prism")
	writeLegacyBinding(t, legacyPlan, repoID, repoRoot)
	writeLegacyBinding(t, legacyE2E, repoID, repoRoot)
	if err := os.WriteFile(filepath.Join(legacyPlan, "notes.txt"), []byte("plan-content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacyE2E, "data.db"), []byte("e2e-content"), 0644); err != nil {
		t.Fatal(err)
	}

	paths, err := cli.ResolveVault(repoRoot, cli.VaultResolveOptions{})
	if err != nil {
		t.Fatalf("ResolveVault migration failed: %v", err)
	}
	if paths.ProjectDir != filepath.Join(vaultRoot, "prism") {
		t.Errorf("ProjectDir = %q", paths.ProjectDir)
	}
	// Legacy dirs gone, project-first present.
	if _, err := os.Stat(legacyPlan); !os.IsNotExist(err) {
		t.Errorf("expected legacy plan dir to be migrated away")
	}
	if _, err := os.Stat(legacyE2E); !os.IsNotExist(err) {
		t.Errorf("expected legacy e2e dir to be migrated away")
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "prism", "plan", "notes.txt")); err != nil {
		t.Errorf("migrated plan content missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "prism", "e2e", "data.db")); err != nil {
		t.Errorf("migrated e2e content missing: %v", err)
	}
	// Single root binding, no child descriptors.
	if _, err := os.Stat(filepath.Join(vaultRoot, "prism", ".vault-binding.json")); err != nil {
		t.Errorf("root binding missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "prism", "plan", ".vault-binding.json")); !os.IsNotExist(err) {
		t.Errorf("expected no child descriptor under plan")
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "prism", "e2e", ".vault-binding.json")); !os.IsNotExist(err) {
		t.Errorf("expected no child descriptor under e2e")
	}
}

func TestMigrateLegacy_MatchMismatch(t *testing.T) {
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)
	repoRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_PROJECT_NAME", "prism")

	writeLegacyBinding(t, filepath.Join(vaultRoot, "plan", "prism"), "github.com/o/a", "/old/a")
	writeLegacyBinding(t, filepath.Join(vaultRoot, "e2e", "prism"), "github.com/o/b", "/old/b")

	_, err := cli.ResolveVault(repoRoot, cli.VaultResolveOptions{})
	if err == nil {
		t.Fatal("expected disagreement failure, got nil")
	}
	if !errors.Is(err, cli.ErrVaultCollision) {
		t.Errorf("expected ErrVaultCollision, got: %v", err)
	}
	// Zero mutation.
	if _, err := os.Stat(filepath.Join(vaultRoot, "prism")); !os.IsNotExist(err) {
		t.Errorf("expected no project dir after disagreement")
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "plan", "prism")); err != nil {
		t.Errorf("legacy plan dir must remain: %v", err)
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "e2e", "prism")); err != nil {
		t.Errorf("legacy e2e dir must remain: %v", err)
	}
}

func TestMigrateLegacy_ValidMalformed(t *testing.T) {
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)
	repoRoot := t.TempDir()
	repoID := "local:" + repoRoot
	t.Setenv("AGENTPLAYBOOK_PROJECT_NAME", "prism")

	writeLegacyBinding(t, filepath.Join(vaultRoot, "plan", "prism"), repoID, repoRoot)
	badDir := filepath.Join(vaultRoot, "e2e", "prism")
	if err := os.MkdirAll(badDir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(badDir, ".vault-binding.json"), []byte("{bad json"), 0640); err != nil {
		t.Fatal(err)
	}

	_, err := cli.ResolveVault(repoRoot, cli.VaultResolveOptions{})
	if err == nil {
		t.Fatal("expected ErrVaultUnbound without adopt, got nil")
	}
	if !errors.Is(err, cli.ErrVaultUnbound) {
		t.Errorf("expected ErrVaultUnbound, got: %v", err)
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "prism")); !os.IsNotExist(err) {
		t.Errorf("expected zero mutation without adopt")
	}

	paths, err := cli.ResolveVault(repoRoot, cli.VaultResolveOptions{Adopt: true})
	if err != nil {
		t.Fatalf("adopt migration failed: %v", err)
	}
	if paths.ProjectDir != filepath.Join(vaultRoot, "prism") {
		t.Errorf("ProjectDir = %q", paths.ProjectDir)
	}
}

func TestMigrateLegacy_ValidMissing(t *testing.T) {
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)
	repoRoot := t.TempDir()
	repoID := "local:" + repoRoot
	t.Setenv("AGENTPLAYBOOK_PROJECT_NAME", "prism")

	writeLegacyBinding(t, filepath.Join(vaultRoot, "plan", "prism"), repoID, repoRoot)
	if err := os.MkdirAll(filepath.Join(vaultRoot, "e2e", "prism"), 0750); err != nil {
		t.Fatal(err)
	}
	// e2e dir exists without descriptor.

	_, err := cli.ResolveVault(repoRoot, cli.VaultResolveOptions{})
	if err == nil {
		t.Fatal("expected ErrVaultUnbound without adopt, got nil")
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "prism")); !os.IsNotExist(err) {
		t.Errorf("expected zero mutation without adopt")
	}

	if _, err := cli.ResolveVault(repoRoot, cli.VaultResolveOptions{Adopt: true}); err != nil {
		t.Fatalf("adopt migration failed: %v", err)
	}
}

func TestMigrateLegacy_MismatchVsCurrent(t *testing.T) {
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)
	repoRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_PROJECT_NAME", "prism")

	writeLegacyBinding(t, filepath.Join(vaultRoot, "plan", "prism"), "github.com/o/other", "/elsewhere")
	writeLegacyBinding(t, filepath.Join(vaultRoot, "e2e", "prism"), "github.com/o/other", "/elsewhere")

	_, err := cli.ResolveVault(repoRoot, cli.VaultResolveOptions{})
	if err == nil {
		t.Fatal("expected collision without rebind, got nil")
	}
	if !errors.Is(err, cli.ErrVaultCollision) {
		t.Errorf("expected ErrVaultCollision, got: %v", err)
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "prism")); !os.IsNotExist(err) {
		t.Errorf("expected zero mutation without rebind")
	}

	paths, err := cli.ResolveVault(repoRoot, cli.VaultResolveOptions{Rebind: true})
	if err != nil {
		t.Fatalf("rebind migration failed: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(paths.ProjectDir, ".vault-binding.json"))
	if err != nil {
		t.Fatal(err)
	}
	var desc cli.VaultBindingDescriptor
	if err := json.Unmarshal(data, &desc); err != nil {
		t.Fatal(err)
	}
	if desc.RepositoryRoot != repoRoot {
		t.Errorf("rebound root = %q; want %q", desc.RepositoryRoot, repoRoot)
	}
}

func TestMigrateLegacy_SingleSide(t *testing.T) {
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)
	repoRoot := t.TempDir()
	repoID := "local:" + repoRoot
	t.Setenv("AGENTPLAYBOOK_PROJECT_NAME", "solo")

	writeLegacyBinding(t, filepath.Join(vaultRoot, "plan", "solo"), repoID, repoRoot)

	paths, err := cli.ResolveVault(repoRoot, cli.VaultResolveOptions{})
	if err != nil {
		t.Fatalf("single-side migration failed: %v", err)
	}
	if _, err := os.Stat(paths.PlanDir); err != nil {
		t.Errorf("migrated plan dir missing: %v", err)
	}
	if _, err := os.Stat(paths.E2EDir); err != nil {
		t.Errorf("fresh e2e dir missing: %v", err)
	}
}

func TestMigrateLegacy_SplitBrain(t *testing.T) {
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)
	repoRoot := t.TempDir()
	repoID := "local:" + repoRoot
	t.Setenv("AGENTPLAYBOOK_PROJECT_NAME", "prism")

	projDir := filepath.Join(vaultRoot, "prism")
	if err := os.MkdirAll(filepath.Join(projDir, "plan"), 0750); err != nil {
		t.Fatal(err)
	}
	writeLegacyBinding(t, projDir, repoID, repoRoot)
	legacyPlan := filepath.Join(vaultRoot, "plan", "prism")
	writeLegacyBinding(t, legacyPlan, repoID, repoRoot)
	origRoot, _ := os.ReadFile(filepath.Join(projDir, ".vault-binding.json"))

	_, err := cli.ResolveVault(repoRoot, cli.VaultResolveOptions{})
	if err == nil {
		t.Fatal("expected SPLIT_BRAIN failure, got nil")
	}
	if !strings.Contains(err.Error(), "SPLIT_BRAIN") {
		t.Errorf("expected SPLIT_BRAIN error, got: %v", err)
	}
	cur, _ := os.ReadFile(filepath.Join(projDir, ".vault-binding.json"))
	if string(cur) != string(origRoot) {
		t.Errorf("root binding mutated during split-brain")
	}
	if _, err := os.Stat(legacyPlan); err != nil {
		t.Errorf("legacy dir must remain: %v", err)
	}
}

func TestMigrateLegacy_SymlinkSources(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping symlink tests on windows")
	}
	t.Run("legacy plan symlink", func(t *testing.T) {
		vaultRoot := t.TempDir()
		t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)
		t.Setenv("AGENTPLAYBOOK_VAULT_MIGRATION_FAULT", "")
		repoRoot := t.TempDir()
		repoID := "local:" + repoRoot
		t.Setenv("AGENTPLAYBOOK_PROJECT_NAME", "prism")

		realPlan := filepath.Join(vaultRoot, "real-plan")
		writeLegacyBinding(t, realPlan, repoID, repoRoot)
		legacyPlan := filepath.Join(vaultRoot, "plan", "prism")
		if err := os.MkdirAll(filepath.Join(vaultRoot, "plan"), 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(realPlan, legacyPlan); err != nil {
			t.Fatal(err)
		}
		writeLegacyBinding(t, filepath.Join(vaultRoot, "e2e", "prism"), repoID, repoRoot)

		_, err := cli.ResolveVault(repoRoot, cli.VaultResolveOptions{})
		if err == nil {
			t.Fatal("expected symlink failure, got nil")
		}
		if !strings.Contains(err.Error(), "symlink") {
			t.Errorf("expected symlink error, got: %v", err)
		}
		if _, err := os.Stat(filepath.Join(vaultRoot, "prism")); !os.IsNotExist(err) {
			t.Errorf("expected zero mutation on symlink source")
		}
	})

	t.Run("legacy e2e symlink", func(t *testing.T) {
		vaultRoot := t.TempDir()
		t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)
		t.Setenv("AGENTPLAYBOOK_VAULT_MIGRATION_FAULT", "")
		repoRoot := t.TempDir()
		repoID := "local:" + repoRoot
		t.Setenv("AGENTPLAYBOOK_PROJECT_NAME", "prism")

		writeLegacyBinding(t, filepath.Join(vaultRoot, "plan", "prism"), repoID, repoRoot)
		realE2E := filepath.Join(vaultRoot, "real-e2e")
		writeLegacyBinding(t, realE2E, repoID, repoRoot)
		legacyE2E := filepath.Join(vaultRoot, "e2e", "prism")
		if err := os.MkdirAll(filepath.Join(vaultRoot, "e2e"), 0750); err != nil {
			t.Fatal(err)
		}
		// Remove dir created by helper parent? helper created legacyE2E? No, helper created realE2E only.
		if err := os.Symlink(realE2E, legacyE2E); err != nil {
			t.Fatal(err)
		}

		_, err := cli.ResolveVault(repoRoot, cli.VaultResolveOptions{})
		if err == nil {
			t.Fatal("expected symlink failure, got nil")
		}
		if !strings.Contains(err.Error(), "symlink") {
			t.Errorf("expected symlink error, got: %v", err)
		}
	})

	t.Run("project dir symlink", func(t *testing.T) {
		vaultRoot := t.TempDir()
		t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)
		t.Setenv("AGENTPLAYBOOK_VAULT_MIGRATION_FAULT", "")
		repoRoot := t.TempDir()
		t.Setenv("AGENTPLAYBOOK_PROJECT_NAME", "prism")

		realDir := filepath.Join(vaultRoot, "real-proj")
		if err := os.MkdirAll(realDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(realDir, filepath.Join(vaultRoot, "prism")); err != nil {
			t.Fatal(err)
		}
		_, err := cli.ResolveVault(repoRoot, cli.VaultResolveOptions{})
		if err == nil {
			t.Fatal("expected symlink failure, got nil")
		}
		if !strings.Contains(err.Error(), "symlink") {
			t.Errorf("expected symlink error, got: %v", err)
		}
	})
}

func TestMigrateLegacy_FaultInjection(t *testing.T) {
	setup := func(t *testing.T) (string, string, string, []byte, []byte) {
		t.Helper()
		vaultRoot := t.TempDir()
		t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)
		repoRoot := t.TempDir()
		repoID := "local:" + repoRoot
		t.Setenv("AGENTPLAYBOOK_PROJECT_NAME", "prism")
		planData := writeLegacyBinding(t, filepath.Join(vaultRoot, "plan", "prism"), repoID, repoRoot)
		e2eData := writeLegacyBinding(t, filepath.Join(vaultRoot, "e2e", "prism"), repoID, repoRoot)
		return vaultRoot, repoRoot, repoID, planData, e2eData
	}
	assertRolledBack := func(t *testing.T, vaultRoot string, wantPlan, wantE2E []byte) {
		t.Helper()
		if _, err := os.Stat(filepath.Join(vaultRoot, "prism")); !os.IsNotExist(err) {
			t.Errorf("expected project dir to be rolled back")
		}
		planDesc, err := os.ReadFile(filepath.Join(vaultRoot, "plan", "prism", ".vault-binding.json"))
		if err != nil {
			t.Errorf("legacy plan descriptor must be restored: %v", err)
		}
		e2eDesc, err := os.ReadFile(filepath.Join(vaultRoot, "e2e", "prism", ".vault-binding.json"))
		if err != nil {
			t.Errorf("legacy e2e descriptor must be restored: %v", err)
		}
		if string(planDesc) != string(wantPlan) {
			t.Errorf("plan descriptor corrupted by rollback:\ngot:\n%s\nwant:\n%s", string(planDesc), string(wantPlan))
		}
		if string(e2eDesc) != string(wantE2E) {
			t.Errorf("e2e descriptor corrupted by rollback:\ngot:\n%s\nwant:\n%s", string(e2eDesc), string(wantE2E))
		}
	}

	for _, stage := range []string{"after-first-rename", "after-second-rename", "fail-root-write", "after-root-write", "fail-child-cleanup"} {
		t.Run("fault_"+stage, func(t *testing.T) {
			vaultRoot, repoRoot, _, wantPlan, wantE2E := setup(t)
			t.Setenv("AGENTPLAYBOOK_VAULT_MIGRATION_FAULT", stage)
			_, err := cli.ResolveVault(repoRoot, cli.VaultResolveOptions{})
			if err == nil {
				t.Fatalf("expected injected fault %q to fail, got nil", stage)
			}
			if !strings.Contains(err.Error(), "injected migration fault") {
				t.Errorf("expected injected fault error, got: %v", err)
			}
			assertRolledBack(t, vaultRoot, wantPlan, wantE2E)
		})
	}
}

func TestParseProjectFromAgentsMD_ProjectFirst(t *testing.T) {
	dir := t.TempDir()
	newMD := "# AGENTS.md\n- **Scaffolding Vault**: plan: `~/.agentplaybook/prism/plan` | e2e: `~/.agentplaybook/prism/e2e`.\n"
	legacyMD := "# AGENTS.md\n- **Scaffolding Vault**: plan: `~/.agentplaybook/plan/legacy` | e2e: `~/.agentplaybook/e2e/legacy`.\n"
	_ = newMD
	_ = legacyMD
	// Exercise via ResolveVault with AGENTS.md present and no env override.
	vaultRoot := t.TempDir()
	t.Setenv("AGENTPLAYBOOK_VAULT_ROOT", vaultRoot)
	repoRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoRoot, "AGENTS.md"), []byte(newMD), 0644); err != nil {
		t.Fatal(err)
	}
	paths, err := cli.ResolveVault(repoRoot, cli.VaultResolveOptions{})
	if err != nil {
		t.Fatalf("ResolveVault with new marker failed: %v", err)
	}
	if paths.ProjectName != "prism" {
		t.Errorf("ProjectName = %q; want prism", paths.ProjectName)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "AGENTS.md"), []byte(legacyMD), 0644); err != nil {
		t.Fatal(err)
	}
	paths, err = cli.ResolveVault(repoRoot, cli.VaultResolveOptions{})
	if err != nil {
		t.Fatalf("ResolveVault with legacy marker failed: %v", err)
	}
	if paths.ProjectName != "legacy" {
		t.Errorf("ProjectName = %q; want legacy", paths.ProjectName)
	}
	_ = dir
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
