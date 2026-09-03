package knowledge

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var diagramPathRegex = regexp.MustCompile(`^docs/diagrams/[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)*\.html$`)

// ValidateDiagramPath enforces that targetPath is a safe, relative path within docs/diagrams/.
// It strictly matches ^docs/diagrams/[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)*\.html$
// allowing dots for versioning (e.g. overview.v2.html) while strictly rejecting directory traversal (..),
// backslashes, leading slashes, and non-HTML extensions.
func ValidateDiagramPath(targetPath string) error {
	if targetPath == "" {
		return errors.New("diagram path cannot be empty")
	}
	if strings.Contains(targetPath, "\\") {
		return errors.New("diagram path cannot contain backslashes")
	}
	if strings.HasPrefix(targetPath, "/") {
		return errors.New("diagram path cannot be an absolute path with leading slash")
	}
	if strings.Contains(targetPath, "..") {
		return errors.New("diagram path cannot contain directory traversal '..'")
	}
	if !strings.HasSuffix(targetPath, ".html") {
		return fmt.Errorf("diagram path %q must have .html extension", targetPath)
	}
	if !diagramPathRegex.MatchString(targetPath) {
		return fmt.Errorf("diagram path %q must match pattern ^docs/diagrams/[a-zA-Z0-9_-]+(\\.[a-zA-Z0-9_-]+)*\\.html$", targetPath)
	}
	return nil
}

// EstimateTokenCount estimates token count using a deterministic heuristic:
// each whitespace-separated word contributes at least 1 token, plus 1 token
// per 4 characters beyond the first (subwords = (len(word) + 3) / 4).
// This deterministic subword estimator defines the canonical programmatic metric for the <100-token digest contract across rules and artifacts.
func EstimateTokenCount(s string) int {
	tokens := 0
	for _, word := range strings.Fields(s) {
		subwords := (len(word) + 3) / 4
		if subwords < 1 {
			subwords = 1
		}
		tokens += subwords
	}
	return tokens
}

// ValidateDiagramCompletion enforces that uri points strictly to docs/diagrams/<safe-name>.html
// (supporting repo-relative paths and file://docs/diagrams/<name>.html;
// rejecting external schemes like http://, https://, absolute file:/// paths, or path traversal ..).
// It also enforces that digest is non-empty, contains zero raw HTML/SVG markup ('<' or '>'),
// contains no newlines, is a single sentence, contains <= 60 words, <= 250 characters, and < 100 tokens.
// This deterministic subword estimator defines the canonical programmatic metric for the <100-token digest contract across rules and artifacts.
func ValidateDiagramCompletion(uri string, digest string) error {
	if strings.TrimSpace(uri) == "" {
		return errors.New("diagram completion URI cannot be empty")
	}
	if strings.Contains(uri, "..") {
		return errors.New("diagram completion URI cannot contain directory traversal '..'")
	}
	if strings.Contains(uri, "\\") {
		return errors.New("diagram completion URI cannot contain backslashes")
	}

	if strings.Contains(uri, "://") {
		if !strings.HasPrefix(uri, "file://") {
			return fmt.Errorf("diagram completion URI %q has invalid scheme (only file:// is permitted)", uri)
		}
		pathPart := strings.TrimPrefix(uri, "file://")
		if strings.HasPrefix(pathPart, "/") {
			return fmt.Errorf("diagram completion URI %q: absolute file:/// URIs are prohibited; use repo-relative path docs/diagrams/<name>.html or file://docs/diagrams/<name>.html", uri)
		}
		// Relative file path: file://docs/diagrams/...
		if err := ValidateDiagramPath(pathPart); err != nil {
			return fmt.Errorf("diagram completion URI %q: %w", uri, err)
		}
	} else {
		// Repo-relative path
		if strings.HasPrefix(uri, "/") {
			return fmt.Errorf("diagram completion URI %q cannot be an absolute path without file:// scheme", uri)
		}
		if err := ValidateDiagramPath(uri); err != nil {
			return fmt.Errorf("diagram completion URI %q: %w", uri, err)
		}
	}

	if strings.TrimSpace(digest) == "" {
		return errors.New("diagram completion digest cannot be empty")
	}
	if strings.Contains(digest, "\n") || strings.Contains(digest, "\r") {
		return errors.New("diagram completion digest must not contain newline characters")
	}
	if strings.Contains(digest, "<") || strings.Contains(digest, ">") {
		return errors.New("diagram completion digest must not contain HTML/SVG markup ('<' or '>')")
	}
	if len([]rune(digest)) > 250 {
		return fmt.Errorf("diagram completion digest exceeds maximum length of 250 characters (got %d)", len([]rune(digest)))
	}

	words := strings.Fields(digest)
	if len(words) > 60 {
		return fmt.Errorf("diagram completion digest exceeds maximum word limit of 60 words (got %d)", len(words))
	}

	tokenCount := EstimateTokenCount(digest)
	if tokenCount >= 100 {
		return fmt.Errorf("diagram completion digest exceeds limit of <100 tokens (got %d tokens)", tokenCount)
	}

	trimmedSentence := strings.TrimRight(strings.TrimSpace(digest), ".!?")
	if strings.ContainsAny(trimmedSentence, ".!?") {
		return errors.New("diagram completion digest must be a single sentence")
	}

	return nil
}
