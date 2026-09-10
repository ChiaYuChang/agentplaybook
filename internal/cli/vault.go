package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	ErrVaultUnbound    = errors.New("vault directory exists but is unbound (missing or malformed .vault-binding.json)")
	ErrVaultCollision  = errors.New("vault directory belongs to a different repository or location")
	ErrVaultSplitBrain = errors.New("vault split-brain: project directory and legacy directories both present (SPLIT_BRAIN)")
	validSlugRegex     = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)
)

type VaultResolveOptions struct {
	Adopt  bool
	Rebind bool
}

type VaultPaths struct {
	VaultRoot   string `json:"vault_root"`
	ProjectName string `json:"project_name"`
	ProjectDir  string `json:"project_dir"`
	PlanDir     string `json:"plan_dir"`
	E2EDir      string `json:"e2e_dir"`
}

type VaultBindingDescriptor struct {
	RepositoryID   string `json:"repository_id"`
	RepositoryRoot string `json:"repository_root"`
	UpdatedAt      string `json:"updated_at"`
}

// CanonicalRepoID deterministically derives a collision-safe repository identifier.
// For remote-backed repositories, it canonicalizes remote origin URL to host/owner/repo
// (lowercased, stripped of credentials, .git, and trailing slashes).
// For local-only repositories, it binds to local:<abs_repo_root>.
func CanonicalRepoID(repoRoot string) (string, error) {
	absRepo, err := filepath.Abs(filepath.Clean(repoRoot))
	if err != nil {
		return "", fmt.Errorf("failed to determine absolute repository root: %w", err)
	}

	remoteURL := getGitRemoteURL(absRepo)
	if remoteURL != "" {
		canonical := NormalizeRemoteURL(remoteURL)
		if canonical != "" {
			return canonical, nil
		}
	}

	return "local:" + absRepo, nil
}

// NormalizeRemoteURL normalizes a Git remote URL to a credential-free, lowercase host/path.
func NormalizeRemoteURL(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return ""
	}

	cleanGitPath := func(p string) string {
		p = strings.TrimPrefix(p, "/")
		p = strings.TrimRight(p, "/")
		p = strings.TrimSuffix(p, ".git")
		p = strings.TrimRight(p, "/")
		return p
	}

	// 1. Handle scheme-based URLs (http://, https://, ssh://, git://)
	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err == nil && u.Host != "" {
			host := u.Host
			if h, _, err := net.SplitHostPort(host); err == nil {
				host = h
			}
			path := cleanGitPath(u.Path)
			if host != "" && path != "" {
				return host + "/" + path
			}
		}

		// Fallback manual parsing if url.Parse fails
		idx := strings.Index(raw, "://")
		rest := raw[idx+3:]
		if atIdx := strings.Index(rest, "@"); atIdx != -1 {
			rest = rest[atIdx+1:]
		}
		parts := strings.SplitN(rest, "/", 2)
		host := parts[0]
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		path := ""
		if len(parts) > 1 {
			path = parts[1]
		}
		path = cleanGitPath(path)
		if host != "" && path != "" {
			return host + "/" + path
		}
	}

	// 2. Handle SCP-like SSH syntax: [user@]host:path (does not contain "://")
	if strings.Contains(raw, ":") {
		part := raw
		if atIdx := strings.Index(part, "@"); atIdx != -1 {
			part = part[atIdx+1:]
		}
		colonIdx := strings.Index(part, ":")
		host := part[:colonIdx]
		path := cleanGitPath(part[colonIdx+1:])
		if host != "" && path != "" {
			return host + "/" + path
		}
	}

	// 3. Fallback for host/path
	return cleanGitPath(raw)
}

func getGitRemoteURL(repoRoot string) string {
	// First try git command if git is available
	cmd := exec.Command("git", "-C", repoRoot, "config", "--get", "remote.origin.url")
	if out, err := cmd.Output(); err == nil {
		u := strings.TrimSpace(string(out))
		if u != "" {
			return u
		}
	}

	// Fallback to parsing .git directly from filesystem
	dotGit := filepath.Join(repoRoot, ".git")
	fi, err := os.Lstat(dotGit)
	if err != nil {
		// Also check if .jj exists and has colocated git or store git
		jjGit := filepath.Join(repoRoot, ".jj", "repo", "store", "git")
		if _, err := os.Stat(jjGit); err == nil {
			if u := parseGitConfigFile(filepath.Join(jjGit, "config")); u != "" {
				return u
			}
		}
		return ""
	}

	var configPath string
	if fi.IsDir() {
		configPath = filepath.Join(dotGit, "config")
	} else {
		// Linked worktree (.git is a file containing "gitdir: <path>")
		content, err := os.ReadFile(dotGit)
		if err == nil {
			line := strings.TrimSpace(string(content))
			if strings.HasPrefix(line, "gitdir:") {
				gitDir := strings.TrimSpace(strings.TrimPrefix(line, "gitdir:"))
				if !filepath.IsAbs(gitDir) {
					gitDir = filepath.Join(repoRoot, gitDir)
				}
				gitDir = filepath.Clean(gitDir)
				commonDirFile := filepath.Join(gitDir, "commondir")
				if cBytes, err := os.ReadFile(commonDirFile); err == nil {
					cDir := strings.TrimSpace(string(cBytes))
					if !filepath.IsAbs(cDir) {
						cDir = filepath.Join(gitDir, cDir)
					}
					configPath = filepath.Join(filepath.Clean(cDir), "config")
				} else {
					configPath = filepath.Join(gitDir, "config")
				}
			}
		}
	}

	if configPath != "" {
		if u := parseGitConfigFile(configPath); u != "" {
			return u
		}
	}

	return ""
}

func parseGitConfigFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	currentSection := ""
	originURL := ""
	firstRemoteURL := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") || line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(line[1 : len(line)-1])
			continue
		}
		if strings.HasPrefix(currentSection, "remote ") || currentSection == "remote" {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				if strings.EqualFold(key, "url") {
					if strings.Contains(currentSection, "origin") {
						originURL = val
					} else if firstRemoteURL == "" {
						firstRemoteURL = val
					}
				}
			}
		}
	}

	if originURL != "" {
		return originURL
	}
	return firstRemoteURL
}

// FindRepoRoot discovers the canonical repository root by searching upward for .jj or .git.
func FindRepoRoot(startDir string) (string, error) {
	curr, err := filepath.Abs(filepath.Clean(startDir))
	if err != nil {
		return "", err
	}

	for {
		// Check .jj directory
		jjPath := filepath.Join(curr, ".jj")
		if fi, err := os.Lstat(jjPath); err == nil && fi.IsDir() {
			return curr, nil
		}

		// Check .git directory or file (linked worktree)
		gitPath := filepath.Join(curr, ".git")
		if _, err := os.Lstat(gitPath); err == nil {
			return curr, nil
		}

		parent := filepath.Dir(curr)
		if parent == curr || parent == "" {
			break
		}
		curr = parent
	}

	return filepath.Abs(filepath.Clean(startDir))
}

func checkNoSymlinks(path string) error {
	cleanPath := filepath.Clean(path)
	vol := filepath.VolumeName(cleanPath)
	rest := cleanPath[len(vol):]
	parts := strings.Split(rest, string(filepath.Separator))

	curr := vol + string(filepath.Separator)
	for _, part := range parts {
		if part == "" {
			continue
		}
		curr = filepath.Join(curr, part)
		fi, err := os.Lstat(curr)
		if err != nil {
			if os.IsNotExist(err) {
				break
			}
			return err
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("vault path component cannot be a symlink: %s", curr)
		}
	}
	return nil
}

func parseProjectFromAgentsMD(agentsMDPath string) string {
	data, err := os.ReadFile(agentsMDPath)
	if err != nil {
		return ""
	}
	content := string(data)
	lines := strings.Split(content, "\n")
	splitVault := func(s string) []string {
		return strings.FieldsFunc(s, func(r rune) bool {
			return r == ' ' || r == '`' || r == '|' || r == '/' || r == '\\' || r == '\t' || r == '"' || r == '\'' || r == '{' || r == '}' || r == ',' || r == ':' || r == ';'
		})
	}
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "scaffolding vault") || strings.Contains(lower, "vault:") {
			// Project-first marker: .agentplaybook/<project>[/plan|/e2e].
			if idx := strings.Index(line, ".agentplaybook/"); idx != -1 {
				rem := line[idx+len(".agentplaybook/"):]
				if !strings.Contains(rem, "{") {
					fields := splitVault(rem)
					if len(fields) >= 1 {
						first := strings.TrimRight(fields[0], ".,:;")
						if (first == "plan" || first == "e2e") && len(fields) >= 2 {
							second := strings.TrimRight(fields[1], ".,:;")
							if validSlugRegex.MatchString(second) {
								return second
							}
						} else if first != "" && first != "plan" && first != "e2e" {
							if validSlugRegex.MatchString(first) {
								return first
							}
						}
					}
				}
			}
			// Legacy fallback: /plan/<project>
			idx := strings.Index(line, "/plan/")
			if idx != -1 {
				rem := line[idx+len("/plan/"):]
				fields := strings.FieldsFunc(rem, func(r rune) bool {
					return r == ' ' || r == '`' || r == '|' || r == '/' || r == '\\' || r == '\t'
				})
				if len(fields) > 0 {
					slug := strings.TrimRight(fields[0], ".,:;")
					if validSlugRegex.MatchString(slug) {
						return slug
					}
				}
			}
		}
	}
	return ""
}

// migrationFaultStage reads the test-only fault-injection override for the
// legacy migration transaction. Empty means no fault. Recognized values:
// "after-first-rename", "after-second-rename", "fail-root-write",
// "after-root-write", "fail-child-cleanup".
func migrationFaultStage() string {
	return strings.TrimSpace(os.Getenv("AGENTPLAYBOOK_VAULT_MIGRATION_FAULT"))
}

// legacyLayoutDirs returns the previous type-first directories for a project.
// Kept strictly for read-only reconciliation and reversible migration.
func legacyLayoutDirs(absVaultRoot, projectName string) (string, string) {
	return filepath.Join(absVaultRoot, "plan", projectName), filepath.Join(absVaultRoot, "e2e", projectName)
}

// ResolveVault determines the canonical Out-of-Tree Shadow Scaffolding Vault paths.
// Project-first layout: <root>/<project> owns one root binding; plan and e2e
// are subdirectories. Legacy type-first directories are reconciled read-only
// and migrated transactionally on first contact.
func ResolveVault(projectRoot string, opts VaultResolveOptions) (*VaultPaths, error) {
	if projectRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		projectRoot = cwd
	}

	absProjectRoot, err := filepath.Abs(filepath.Clean(projectRoot))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute project root: %w", err)
	}

	// 1. Vault Root Resolution Precedence
	vaultRoot := os.Getenv("AGENTPLAYBOOK_VAULT_ROOT")
	if vaultRoot == "" {
		vaultRoot = os.Getenv("AGENTPLAYBOOK_HOME")
	}
	if vaultRoot == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to determine user home directory: %w", err)
		}
		vaultRoot = filepath.Join(homeDir, ".agentplaybook")
	}
	absVaultRoot, err := filepath.Abs(filepath.Clean(vaultRoot))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute vault root: %w", err)
	}

	// 2. Symlink Rejection on Vault Root
	if err := checkNoSymlinks(absVaultRoot); err != nil {
		return nil, err
	}

	// 3. Project Name Resolution Precedence
	var projectName string
	if val, ok := os.LookupEnv("AGENTPLAYBOOK_PROJECT_NAME"); ok {
		projectName = val
	} else {
		agentsMDPath := filepath.Join(absProjectRoot, "AGENTS.md")
		projectName = parseProjectFromAgentsMD(agentsMDPath)
		if projectName == "" {
			repoRoot, err := FindRepoRoot(absProjectRoot)
			if err != nil {
				repoRoot = absProjectRoot
			}
			projectName = filepath.Base(repoRoot)
		}
	}

	// 4. Project Slug Validation
	if !validSlugRegex.MatchString(projectName) || projectName == "." || projectName == ".." {
		return nil, fmt.Errorf("invalid project name %q: must match ^[a-zA-Z0-9][a-zA-Z0-9._-]*$", projectName)
	}

	repoRoot, err := FindRepoRoot(absProjectRoot)
	if err != nil {
		repoRoot = absProjectRoot
	}

	projectDir := filepath.Join(absVaultRoot, projectName)
	planDir := filepath.Join(projectDir, "plan")
	e2eDir := filepath.Join(projectDir, "e2e")
	legacyPlanDir, legacyE2EDir := legacyLayoutDirs(absVaultRoot, projectName)

	// Symlink check on project target (existing components only).
	if err := checkNoSymlinks(projectDir); err != nil {
		return nil, err
	}

	// 5. Absolute Containment: project directory must be outside project root.
	isOutside := func(rel string) bool {
		return rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))
	}
	if rel, errRel := filepath.Rel(absProjectRoot, projectDir); errRel == nil && !isOutside(rel) {
		return nil, fmt.Errorf("vault project directory %s cannot be inside project root %s", projectDir, absProjectRoot)
	}

	repoID, err := CanonicalRepoID(repoRoot)
	if err != nil {
		return nil, err
	}

	paths := &VaultPaths{
		VaultRoot:   absVaultRoot,
		ProjectName: projectName,
		ProjectDir:  projectDir,
		PlanDir:     planDir,
		E2EDir:      e2eDir,
	}

	projectExists := dirExists(projectDir)
	legacyPlanExists := dirExists(legacyPlanDir)
	legacyE2EExists := dirExists(legacyE2EDir)

	// Split-brain: new project dir plus any legacy dir is fail-closed, no mutation.
	if projectExists && (legacyPlanExists || legacyE2EExists) {
		return nil, fmt.Errorf("%w: %s coexists with legacy type-first directories", ErrVaultSplitBrain, projectDir)
	}

	// Legacy reconciliation + migration when only legacy layout is present.
	if !projectExists && (legacyPlanExists || legacyE2EExists) {
		if err := reconcileAndMigrateLegacy(paths, legacyPlanDir, legacyE2EDir, repoRoot, repoID, opts); err != nil {
			return nil, err
		}
		return paths, nil
	}

	// 6. Provenance, Collision, Relocation & Rebind Handling (single root binding).
	if projectExists {
		adopt := opts.Adopt || os.Getenv("AGENTPLAYBOOK_VAULT_ADOPT") == "1"
		existingBinding, bindingErr := readVaultBinding(projectDir)
		if bindingErr != nil {
			if !adopt {
				return nil, ErrVaultUnbound
			}
			return paths, nil
		}
		if existingBinding.RepositoryID == repoID {
			if existingBinding.RepositoryRoot != repoRoot {
				if err := UpdateVaultDescriptors(paths, repoRoot, repoID); err != nil {
					return nil, err
				}
			}
		} else {
			if opts.Rebind {
				if err := UpdateVaultDescriptors(paths, repoRoot, repoID); err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("%w: vault belongs to %s, current is %s", ErrVaultCollision, existingBinding.RepositoryID, repoID)
			}
		}
	}

	return paths, nil
}

// reconcileAndMigrateLegacy validates legacy descriptors read-only before any
// mutation, then runs the ordered reversible migration transaction.
func reconcileAndMigrateLegacy(paths *VaultPaths, legacyPlanDir, legacyE2EDir, repoRoot, repoID string, opts VaultResolveOptions) error {
	legacyPlanExists := dirExists(legacyPlanDir)
	legacyE2EExists := dirExists(legacyE2EDir)
	adopt := opts.Adopt || os.Getenv("AGENTPLAYBOOK_VAULT_ADOPT") == "1"

	var (
		planBinding *VaultBindingDescriptor
		planErr     error
		e2eBinding  *VaultBindingDescriptor
		e2eErr      error
	)
	if legacyPlanExists {
		planBinding, planErr = readVaultBinding(legacyPlanDir)
	}
	if legacyE2EExists {
		e2eBinding, e2eErr = readVaultBinding(legacyE2EDir)
	}

	// Malformed/missing asymmetry fails closed without adopt; zero mutation so far.
	if (legacyPlanExists && planErr != nil) || (legacyE2EExists && e2eErr != nil) {
		if !adopt {
			return ErrVaultUnbound
		}
	}

	// Valid-but-different legacy IDs disagree: fail closed, zero mutation.
	if legacyPlanExists && legacyE2EExists && planErr == nil && e2eErr == nil {
		if planBinding.RepositoryID != e2eBinding.RepositoryID {
			return fmt.Errorf("%w: legacy plan vault (%s) and e2e vault (%s) belong to different repositories", ErrVaultCollision, planBinding.RepositoryID, e2eBinding.RepositoryID)
		}
	}

	// Either legacy ID mismatching current repo without rebind fails closed.
	for _, b := range []*VaultBindingDescriptor{planBinding, e2eBinding} {
		if b == nil {
			continue
		}
		if b.RepositoryID != repoID && !opts.Rebind {
			return fmt.Errorf("%w: legacy vault belongs to %s, current is %s", ErrVaultCollision, b.RepositoryID, repoID)
		}
	}

	return migrateLegacyVault(paths, legacyPlanDir, legacyE2EDir, repoRoot, repoID)
}

// migrateLegacyVault executes the ordered reversible migration transaction:
// backup, symlink-check sources, rename legacy dirs, write single root binding,
// then remove relocated child descriptors. Any stage failure rolls back fully.
func migrateLegacyVault(paths *VaultPaths, legacyPlanDir, legacyE2EDir, repoRoot, repoID string) error {
	legacyPlanExists := dirExists(legacyPlanDir)
	legacyE2EExists := dirExists(legacyE2EDir)

	// (a) Backup both legacy descriptors in memory; fail closed on unreadable non-NotExist.
	planBackup, planHadBackupErr := readBindingBytes(filepath.Join(legacyPlanDir, ".vault-binding.json"))
	if legacyPlanExists && planHadBackupErr != nil && !os.IsNotExist(planHadBackupErr) {
		return fmt.Errorf("cannot read pre-existing legacy plan vault binding for safe rollback: %w", planHadBackupErr)
	}
	e2eBackup, e2eHadBackupErr := readBindingBytes(filepath.Join(legacyE2EDir, ".vault-binding.json"))
	if legacyE2EExists && e2eHadBackupErr != nil && !os.IsNotExist(e2eHadBackupErr) {
		return fmt.Errorf("cannot read pre-existing legacy e2e vault binding for safe rollback: %w", e2eHadBackupErr)
	}
	planHadDesc := legacyPlanExists && planHadBackupErr == nil
	e2eHadDesc := legacyE2EExists && e2eHadBackupErr == nil

	// Symlink-check every migration source before any rename.
	if legacyPlanExists {
		if err := checkNoSymlinks(legacyPlanDir); err != nil {
			return err
		}
	}
	if legacyE2EExists {
		if err := checkNoSymlinks(legacyE2EDir); err != nil {
			return err
		}
	}

	var createdDirs []string
	type renameOp struct{ oldPath, newPath string }
	var completedRenames []renameOp

	mkdirTrack := func(dir string) error {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			var toCreate []string
			curr := dir
			for {
				if _, err := os.Stat(curr); err == nil {
					break
				}
				toCreate = append([]string{curr}, toCreate...)
				parent := filepath.Dir(curr)
				if parent == curr || parent == "" {
					break
				}
				curr = parent
			}
			if err := os.MkdirAll(dir, 0750); err != nil {
				return err
			}
			for _, d := range toCreate {
				_ = os.Chmod(d, 0750)
				createdDirs = append(createdDirs, d)
			}
		}
		return nil
	}

	rollback := func(trigger error) error {
		var rbErrs []error
		// Reverse completed renames.
		for i := len(completedRenames) - 1; i >= 0; i-- {
			op := completedRenames[i]
			if err := os.Rename(op.newPath, op.oldPath); err != nil && !os.IsNotExist(err) {
				rbErrs = append(rbErrs, fmt.Errorf("rollback rename %s -> %s: %w", op.newPath, op.oldPath, err))
			}
		}
		// Restore descriptor backups (including relocated child descriptors).
		if planHadDesc {
			if err := os.WriteFile(filepath.Join(legacyPlanDir, ".vault-binding.json"), planBackup, 0640); err != nil {
				rbErrs = append(rbErrs, fmt.Errorf("rollback restore legacy plan binding: %w", err))
			}
		}
		if e2eHadDesc {
			if err := os.WriteFile(filepath.Join(legacyE2EDir, ".vault-binding.json"), e2eBackup, 0640); err != nil {
				rbErrs = append(rbErrs, fmt.Errorf("rollback restore legacy e2e binding: %w", err))
			}
		}
		// Remove root binding written by this transaction.
		if err := os.Remove(filepath.Join(paths.ProjectDir, ".vault-binding.json")); err != nil && !os.IsNotExist(err) {
			rbErrs = append(rbErrs, fmt.Errorf("rollback remove root binding: %w", err))
		}
		// Remove created dirs in reverse (fresh sides + project dir).
		for i := len(createdDirs) - 1; i >= 0; i-- {
			if err := os.Remove(createdDirs[i]); err != nil && !os.IsNotExist(err) {
				rbErrs = append(rbErrs, fmt.Errorf("rollback remove dir %s: %w", createdDirs[i], err))
			}
		}
		if len(rbErrs) == 0 {
			return trigger
		}
		return errors.Join(append([]error{trigger}, rbErrs...)...)
	}

	injected := func(stage string) error {
		return fmt.Errorf("injected migration fault at %s (AGENTPLAYBOOK_VAULT_MIGRATION_FAULT=%s)", stage, stage)
	}
	fault := migrationFaultStage()

	// Create project root before renames.
	if err := mkdirTrack(paths.ProjectDir); err != nil {
		return rollback(fmt.Errorf("failed to create vault project directory %s: %w", paths.ProjectDir, err))
	}

	// (b) Rename legacy dirs (same-filesystem rename, no copy).
	firstRenameDone := false
	if legacyPlanExists {
		if err := os.Rename(legacyPlanDir, paths.PlanDir); err != nil {
			return rollback(fmt.Errorf("failed to migrate legacy plan directory: %w", err))
		}
		completedRenames = append(completedRenames, renameOp{oldPath: legacyPlanDir, newPath: paths.PlanDir})
		firstRenameDone = true
		if fault == "after-first-rename" {
			return rollback(injected(fault))
		}
	}
	if legacyE2EExists {
		if err := os.Rename(legacyE2EDir, paths.E2EDir); err != nil {
			return rollback(fmt.Errorf("failed to migrate legacy e2e directory: %w", err))
		}
		completedRenames = append(completedRenames, renameOp{oldPath: legacyE2EDir, newPath: paths.E2EDir})
		if fault == "after-second-rename" {
			return rollback(injected(fault))
		}
	}
	_ = firstRenameDone

	// Missing side created fresh under the same transaction.
	if !legacyPlanExists {
		if err := mkdirTrack(paths.PlanDir); err != nil {
			return rollback(fmt.Errorf("failed to create plan vault directory %s: %w", paths.PlanDir, err))
		}
	}
	if !legacyE2EExists {
		if err := mkdirTrack(paths.E2EDir); err != nil {
			return rollback(fmt.Errorf("failed to create e2e vault directory %s: %w", paths.E2EDir, err))
		}
	}

	// (c) Write single root binding (0640); only after renames.
	if fault == "fail-root-write" {
		return rollback(injected(fault))
	}
	now := time.Now().UTC().Format(time.RFC3339)
	desc := VaultBindingDescriptor{
		RepositoryID:   repoID,
		RepositoryRoot: repoRoot,
		UpdatedAt:      now,
	}
	data, err := json.MarshalIndent(desc, "", "  ")
	if err != nil {
		return rollback(err)
	}
	data = append(data, '\n')
	rootDesc := filepath.Join(paths.ProjectDir, ".vault-binding.json")
	if err := os.WriteFile(rootDesc, data, 0640); err != nil {
		return rollback(fmt.Errorf("failed to write vault root binding: %w", err))
	}
	if fault == "after-root-write" {
		return rollback(injected(fault))
	}

	// (d) Only after root binding is durable, remove relocated child descriptors.
	if fault == "fail-child-cleanup" {
		return rollback(injected(fault))
	}
	for _, childDesc := range []string{
		filepath.Join(paths.PlanDir, ".vault-binding.json"),
		filepath.Join(paths.E2EDir, ".vault-binding.json"),
	} {
		if err := os.Remove(childDesc); err != nil && !os.IsNotExist(err) {
			return rollback(fmt.Errorf("failed to remove relocated child descriptor %s: %w", childDesc, err))
		}
	}

	return nil
}

func readBindingBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func dirExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

func readVaultBinding(dir string) (*VaultBindingDescriptor, error) {
	bindingPath := filepath.Join(dir, ".vault-binding.json")
	data, err := os.ReadFile(bindingPath)
	if err != nil {
		return nil, err
	}
	var desc VaultBindingDescriptor
	if err := json.Unmarshal(data, &desc); err != nil {
		return nil, err
	}
	if desc.RepositoryID == "" {
		return nil, errors.New("empty repository_id in descriptor")
	}
	return &desc, nil
}

// EnsureVaultDirectories creates the project-first vault directories with 0750
// permissions and writes the single root .vault-binding.json (0640) with
// all-or-nothing rollback on failure.
func EnsureVaultDirectories(paths *VaultPaths, repoRoot string, repoID string, opts VaultResolveOptions) error {
	if paths == nil {
		return errors.New("nil vault paths")
	}
	if paths.ProjectDir == "" {
		// Backfill derived paths for callers constructing VaultPaths manually.
		if paths.VaultRoot != "" && paths.ProjectName != "" {
			paths.ProjectDir = filepath.Join(paths.VaultRoot, paths.ProjectName)
		} else {
			return errors.New("vault project directory is not resolved")
		}
	}
	if paths.PlanDir == "" {
		paths.PlanDir = filepath.Join(paths.ProjectDir, "plan")
	}
	if paths.E2EDir == "" {
		paths.E2EDir = filepath.Join(paths.ProjectDir, "e2e")
	}

	if repoRoot == "" {
		if r, err := FindRepoRoot("."); err == nil {
			repoRoot = r
		}
	}
	if repoID == "" && repoRoot != "" {
		if id, err := CanonicalRepoID(repoRoot); err == nil {
			repoID = id
		}
	}

	rootDesc := filepath.Join(paths.ProjectDir, ".vault-binding.json")

	// Pre-read existing root descriptor before any mutation.
	rootBackup, rootHadBackupErr := os.ReadFile(rootDesc)
	if rootHadBackupErr != nil && !os.IsNotExist(rootHadBackupErr) {
		return fmt.Errorf("cannot read pre-existing vault root binding for safe rollback: %w", rootHadBackupErr)
	}

	var createdDirs []string

	mkdir := func(dir string) error {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			var toCreate []string
			curr := dir
			for {
				if _, err := os.Stat(curr); err == nil {
					break
				}
				toCreate = append([]string{curr}, toCreate...)
				parent := filepath.Dir(curr)
				if parent == curr || parent == "" {
					break
				}
				curr = parent
			}
			if err := os.MkdirAll(dir, 0750); err != nil {
				return err
			}
			for _, d := range toCreate {
				_ = os.Chmod(d, 0750)
				createdDirs = append(createdDirs, d)
			}
		}
		return nil
	}

	rollback := func() {
		if rootHadBackupErr == nil {
			_ = os.WriteFile(rootDesc, rootBackup, 0640)
		} else {
			_ = os.Remove(rootDesc)
		}
		for i := len(createdDirs) - 1; i >= 0; i-- {
			_ = os.Remove(createdDirs[i])
		}
	}

	if err := mkdir(paths.ProjectDir); err != nil {
		rollback()
		return fmt.Errorf("failed to create vault project directory %s: %w", paths.ProjectDir, err)
	}

	if err := mkdir(paths.PlanDir); err != nil {
		rollback()
		return fmt.Errorf("failed to create plan vault directory %s: %w", paths.PlanDir, err)
	}

	if err := mkdir(paths.E2EDir); err != nil {
		rollback()
		return fmt.Errorf("failed to create e2e vault directory %s: %w", paths.E2EDir, err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	desc := VaultBindingDescriptor{
		RepositoryID:   repoID,
		RepositoryRoot: repoRoot,
		UpdatedAt:      now,
	}
	data, err := json.MarshalIndent(desc, "", "  ")
	if err != nil {
		rollback()
		return err
	}
	data = append(data, '\n')

	if err := os.WriteFile(rootDesc, data, 0640); err != nil {
		rollback()
		return fmt.Errorf("failed to write vault root binding: %w", err)
	}

	return nil
}

// UpdateVaultDescriptors writes the single root .vault-binding.json descriptor
// with rollback restoring the prior descriptor on failure.
func UpdateVaultDescriptors(paths *VaultPaths, repoRoot string, repoID string) error {
	if paths == nil || paths.ProjectDir == "" {
		return errors.New("vault project directory is not resolved")
	}
	rootDesc := filepath.Join(paths.ProjectDir, ".vault-binding.json")

	rootBackup, rootHadBackupErr := os.ReadFile(rootDesc)
	if rootHadBackupErr != nil && !os.IsNotExist(rootHadBackupErr) {
		return fmt.Errorf("cannot read pre-existing vault root binding for safe rollback: %w", rootHadBackupErr)
	}

	rollback := func() {
		if rootHadBackupErr == nil {
			_ = os.WriteFile(rootDesc, rootBackup, 0640)
		} else {
			_ = os.Remove(rootDesc)
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	desc := VaultBindingDescriptor{
		RepositoryID:   repoID,
		RepositoryRoot: repoRoot,
		UpdatedAt:      now,
	}
	data, err := json.MarshalIndent(desc, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	if err := os.WriteFile(rootDesc, data, 0640); err != nil {
		rollback()
		return fmt.Errorf("failed to update vault root binding: %w", err)
	}

	return nil
}
