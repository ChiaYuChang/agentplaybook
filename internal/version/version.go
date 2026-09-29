package version

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var embedded string

// Release returns the canonical release version embedded in the binary.
func Release() string {
	return releaseValue(embedded)
}

func releaseValue(contents string) string {
	value, ok := parseRelease(contents)
	if !ok {
		panic("agentplaybook: invalid embedded release version")
	}
	return value
}

func parseRelease(contents string) (string, bool) {
	if !strings.HasSuffix(contents, "\n") {
		return "", false
	}
	contents = strings.TrimSuffix(contents, "\n")

	components, ok := strings.CutPrefix(contents, "v")
	if !ok {
		return "", false
	}
	major, remaining, ok := strings.Cut(components, ".")
	if !ok {
		return "", false
	}
	minor, patch, ok := strings.Cut(remaining, ".")
	if !ok || strings.Contains(patch, ".") {
		return "", false
	}
	if !validNumber(major) || !validNumber(minor) || !validNumber(patch) {
		return "", false
	}
	return contents, true
}

func validNumber(component string) bool {
	if component == "" || len(component) > 1 && component[0] == '0' {
		return false
	}
	for i := range len(component) {
		if component[i] < '0' || component[i] > '9' {
			return false
		}
	}
	return true
}
