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
	ErrVaultUnbound   = errors.New("vault directory exists but is unbound (missing or malformed .vault-binding.json)")
	ErrVaultCollision = errors.New("vault directory belongs to a different repository or location")
	validSlugRegex    = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)
)

type VaultResolveOptions struct {
	Adopt  bool
	Rebind bool
}

type VaultPaths struct {
	VaultRoot   string `json:"vault_root"`
	ProjectName string `json:"project_name"`
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
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "scaffolding vault") || strings.Contains(lower, "vault:") {
			// Look for /plan/<project> or /e2e/<project>
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

// ResolveVault determines the canonical Out-of-Tree Shadow Scaffolding Vault paths.
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

	planDir := filepath.Join(absVaultRoot, "plan", projectName)
	e2eDir := filepath.Join(absVaultRoot, "e2e", projectName)

	// Symlink checks on targets
	if err := checkNoSymlinks(planDir); err != nil {
		return nil, err
	}
	if err := checkNoSymlinks(e2eDir); err != nil {
		return nil, err
	}

	// 5. Absolute Containment: vault paths must strictly be outside project root
	isOutside := func(rel string) bool {
		return rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))
	}
	relPlan, errPlan := filepath.Rel(absProjectRoot, planDir)
	relE2E, errE2E := filepath.Rel(absProjectRoot, e2eDir)
	if errPlan == nil && !isOutside(relPlan) {
		return nil, fmt.Errorf("vault plan directory %s cannot be inside project root %s", planDir, absProjectRoot)
	}
	if errE2E == nil && !isOutside(relE2E) {
		return nil, fmt.Errorf("vault e2e directory %s cannot be inside project root %s", e2eDir, absProjectRoot)
	}

	repoID, err := CanonicalRepoID(repoRoot)
	if err != nil {
		return nil, err
	}

	// 6. Provenance, Collision, Relocation & Rebind Handling
	planExists := dirExists(planDir)
	e2eExists := dirExists(e2eDir)

	if planExists || e2eExists {
		var (
			planBinding *VaultBindingDescriptor
			planErr     error
			e2eBinding  *VaultBindingDescriptor
			e2eErr      error
		)
		if planExists {
			planBinding, planErr = readVaultBinding(planDir)
		}
		if e2eExists {
			e2eBinding, e2eErr = readVaultBinding(e2eDir)
		}

		adopt := opts.Adopt || os.Getenv("AGENTPLAYBOOK_VAULT_ADOPT") == "1"
		if (planExists && planErr != nil) || (e2eExists && e2eErr != nil) {
			if !adopt {
				return nil, ErrVaultUnbound
			}
		}

		var existingBinding *VaultBindingDescriptor
		if planExists && planErr == nil {
			existingBinding = planBinding
		} else if e2eExists && e2eErr == nil {
			existingBinding = e2eBinding
		}

		if existingBinding != nil {
			if existingBinding.RepositoryID == repoID {
				// Matching repository_id: authorize relocation
				if existingBinding.RepositoryRoot != repoRoot {
					paths := &VaultPaths{
						VaultRoot:   absVaultRoot,
						ProjectName: projectName,
						PlanDir:     planDir,
						E2EDir:      e2eDir,
					}
					if err := UpdateVaultDescriptors(paths, repoRoot, repoID); err != nil {
						return nil, err
					}
				}
			} else {
				// Mismatching repository_id
				if opts.Rebind {
					paths := &VaultPaths{
						VaultRoot:   absVaultRoot,
						ProjectName: projectName,
						PlanDir:     planDir,
						E2EDir:      e2eDir,
					}
					if err := UpdateVaultDescriptors(paths, repoRoot, repoID); err != nil {
						return nil, err
					}
				} else {
					return nil, fmt.Errorf("%w: vault belongs to %s, current is %s", ErrVaultCollision, existingBinding.RepositoryID, repoID)
				}
			}
		}
	}

	return &VaultPaths{
		VaultRoot:   absVaultRoot,
		ProjectName: projectName,
		PlanDir:     planDir,
		E2EDir:      e2eDir,
	}, nil
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

// EnsureVaultDirectories creates the plan and e2e vault directories with 0750 permissions
// and writes .vault-binding.json descriptors using all-or-nothing rollback on failure.
func EnsureVaultDirectories(paths *VaultPaths, repoRoot string, repoID string, opts VaultResolveOptions) error {
	if paths == nil {
		return errors.New("nil vault paths")
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

	planDesc := filepath.Join(paths.PlanDir, ".vault-binding.json")
	e2eDesc := filepath.Join(paths.E2EDir, ".vault-binding.json")

	// Pre-read existing descriptors before any mutation.
	// Fail closed on any read error that is not os.IsNotExist — we cannot
	// safely restore a descriptor whose content we cannot read.
	planBackup, planHadBackupErr := os.ReadFile(planDesc)
	if planHadBackupErr != nil && !os.IsNotExist(planHadBackupErr) {
		return fmt.Errorf("cannot read pre-existing plan vault binding for safe rollback: %w", planHadBackupErr)
	}
	e2eBackup, e2eHadBackupErr := os.ReadFile(e2eDesc)
	if e2eHadBackupErr != nil && !os.IsNotExist(e2eHadBackupErr) {
		return fmt.Errorf("cannot read pre-existing e2e vault binding for safe rollback: %w", e2eHadBackupErr)
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
		if planHadBackupErr == nil {
			_ = os.WriteFile(planDesc, planBackup, 0640)
		} else {
			_ = os.Remove(planDesc)
		}
		if e2eHadBackupErr == nil {
			_ = os.WriteFile(e2eDesc, e2eBackup, 0640)
		} else {
			_ = os.Remove(e2eDesc)
		}
		for i := len(createdDirs) - 1; i >= 0; i-- {
			_ = os.Remove(createdDirs[i])
		}
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

	if err := os.WriteFile(planDesc, data, 0640); err != nil {
		rollback()
		return fmt.Errorf("failed to write plan vault binding: %w", err)
	}

	if err := os.WriteFile(e2eDesc, data, 0640); err != nil {
		rollback()
		return fmt.Errorf("failed to write e2e vault binding: %w", err)
	}

	return nil
}

// UpdateVaultDescriptors writes updated .vault-binding.json descriptors to both PlanDir and E2EDir
// with atomic paired rollback if either write fails.
func UpdateVaultDescriptors(paths *VaultPaths, repoRoot string, repoID string) error {
	planDesc := filepath.Join(paths.PlanDir, ".vault-binding.json")
	e2eDesc := filepath.Join(paths.E2EDir, ".vault-binding.json")

	// Pre-read both descriptors before any mutation.
	// Fail closed on any non-NotExist error — we cannot safely restore what we cannot read.
	planBackup, planHadBackupErr := os.ReadFile(planDesc)
	if planHadBackupErr != nil && !os.IsNotExist(planHadBackupErr) {
		return fmt.Errorf("cannot read pre-existing plan vault binding for safe rollback: %w", planHadBackupErr)
	}
	e2eBackup, e2eHadBackupErr := os.ReadFile(e2eDesc)
	if e2eHadBackupErr != nil && !os.IsNotExist(e2eHadBackupErr) {
		return fmt.Errorf("cannot read pre-existing e2e vault binding for safe rollback: %w", e2eHadBackupErr)
	}

	rollback := func() {
		if planHadBackupErr == nil {
			_ = os.WriteFile(planDesc, planBackup, 0640)
		} else {
			_ = os.Remove(planDesc)
		}
		if e2eHadBackupErr == nil {
			_ = os.WriteFile(e2eDesc, e2eBackup, 0640)
		} else {
			_ = os.Remove(e2eDesc)
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

	if err := os.WriteFile(planDesc, data, 0640); err != nil {
		rollback()
		return fmt.Errorf("failed to update plan vault binding: %w", err)
	}

	if err := os.WriteFile(e2eDesc, data, 0640); err != nil {
		rollback()
		return fmt.Errorf("failed to update e2e vault binding: %w", err)
	}

	return nil
}
