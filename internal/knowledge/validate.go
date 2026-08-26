package knowledge

import (
	"errors"
	"fmt"
	"strings"
)

// Validate performs comprehensive semantic and referential integrity checks across the entire knowledge store.
// All detected violations are collected and returned using errors.Join.
func Validate(k *Knowledge) error {
	var errs []error

	// 1. Validate Config
	if len(k.Config.Languages) == 0 {
		errs = append(errs, errors.New("config.languages cannot be empty"))
	}
	if k.Config.Transport == "" {
		errs = append(errs, errors.New("config.transport cannot be empty"))
	}
	if !strings.Contains(k.Config.Communication.RolePrefixTemplate, "{role}") {
		errs = append(errs, fmt.Errorf("config.communication.role_prefix_template %q must contain {role}", k.Config.Communication.RolePrefixTemplate))
	}
	if !strings.Contains(k.Config.Communication.FlowPrefixTemplate, "{role}") || !strings.Contains(k.Config.Communication.FlowPrefixTemplate, "{flow}") {
		errs = append(errs, fmt.Errorf("config.communication.flow_prefix_template %q must contain {role} and {flow}", k.Config.Communication.FlowPrefixTemplate))
	}

	// 2. Validate Roles
	roleNames := make(map[Role]bool, len(k.roleList))
	for _, r := range k.roleList {
		if r.Name == "" {
			errs = append(errs, errors.New("role name cannot be empty"))
			continue
		}
		if roleNames[r.Name] {
			errs = append(errs, fmt.Errorf("duplicate role name: %q", r.Name))
		}
		roleNames[r.Name] = true

		if r.Description == "" {
			errs = append(errs, fmt.Errorf("role %q description cannot be empty", r.Name))
		}
		if len(r.Responsibilities) == 0 {
			errs = append(errs, fmt.Errorf("role %q responsibilities cannot be empty", r.Name))
		}
		if len(r.Boundaries) == 0 {
			errs = append(errs, fmt.Errorf("role %q boundaries cannot be empty", r.Name))
		}
	}

	for _, r := range k.roleList {
		for _, target := range r.Communication.Targets {
			if !roleNames[target] {
				errs = append(errs, fmt.Errorf("role %q communication target %q does not exist", r.Name, target))
			}
		}
	}

	// 3. Validate Artifacts
	artifactNames := make(map[string]bool, len(k.artifactList))
	for _, a := range k.artifactList {
		if a.Name == "" {
			errs = append(errs, errors.New("artifact name cannot be empty"))
			continue
		}
		if artifactNames[a.Name] {
			errs = append(errs, fmt.Errorf("duplicate artifact name: %q", a.Name))
		}
		artifactNames[a.Name] = true

		if !roleNames[a.Owner] {
			errs = append(errs, fmt.Errorf("artifact %q owner %q does not exist", a.Name, a.Owner))
		}
		if len(a.Visibility) == 0 {
			errs = append(errs, fmt.Errorf("artifact %q visibility cannot be empty", a.Name))
		}
		for _, v := range a.Visibility {
			if !roleNames[v] {
				errs = append(errs, fmt.Errorf("artifact %q visibility role %q does not exist", a.Name, v))
			}
		}

		if a.Type != "document" && a.Type != "message" {
			errs = append(errs, fmt.Errorf("artifact %q has invalid type %q (must be 'document' or 'message')", a.Name, a.Type))
		}
		if a.Type == "document" && len(a.Sections) == 0 {
			errs = append(errs, fmt.Errorf("document artifact %q must define at least one section", a.Name))
		}
		if a.Type == "message" && len(a.Fields) == 0 {
			errs = append(errs, fmt.Errorf("message artifact %q must define at least one field", a.Name))
		}
	}

	// 4. Validate Flows
	flowNames := make(map[string]bool, len(k.flowList))
	for _, f := range k.flowList {
		if f.Name == "" {
			errs = append(errs, errors.New("flow name cannot be empty"))
			continue
		}
		if flowNames[f.Name] {
			errs = append(errs, fmt.Errorf("duplicate flow name: %q", f.Name))
		}
		flowNames[f.Name] = true

		if len(f.Steps) == 0 {
			errs = append(errs, fmt.Errorf("flow %q must contain at least one step", f.Name))
			continue
		}

		stepIndices := make(map[int]bool, len(f.Steps))
		maxIdx := 0
		for _, s := range f.Steps {
			if s.Index <= 0 {
				errs = append(errs, fmt.Errorf("flow %q step index %d must be positive", f.Name, s.Index))
			}
			if stepIndices[s.Index] {
				errs = append(errs, fmt.Errorf("flow %q has duplicate step index: %d", f.Name, s.Index))
			}
			stepIndices[s.Index] = true
			if s.Index > maxIdx {
				maxIdx = s.Index
			}

			if !roleNames[s.Actor] {
				errs = append(errs, fmt.Errorf("flow %q step %d actor %q does not exist", f.Name, s.Index, s.Actor))
			}
			if strings.TrimSpace(s.Action) == "" {
				errs = append(errs, fmt.Errorf("flow %q step %d action cannot be empty", f.Name, s.Index))
			}
		}

		for _, s := range f.Steps {
			whens := make(map[string]bool, len(s.Conditions))
			for _, c := range s.Conditions {
				if strings.TrimSpace(c.When) == "" {
					errs = append(errs, fmt.Errorf("flow %q step %d condition 'when' cannot be empty", f.Name, s.Index))
				}
				if whens[c.When] {
					errs = append(errs, fmt.Errorf("flow %q step %d duplicate condition trigger: %q", f.Name, s.Index, c.When))
				}
				whens[c.When] = true

				if !stepIndices[c.Then] {
					errs = append(errs, fmt.Errorf("flow %q step %d condition target step %d does not exist", f.Name, s.Index, c.Then))
				}
			}

			// Sequential validation: if no conditions and not the maximum index, step index + 1 must exist
			if len(s.Conditions) == 0 && s.Index < maxIdx {
				if !stepIndices[s.Index+1] {
					errs = append(errs, fmt.Errorf("flow %q step %d has no conditions but missing sequential next step %d", f.Name, s.Index, s.Index+1))
				}
			}
		}
	}

	// 5. Validate Rules
	ruleIDs := make(map[string]bool, len(k.ruleList))
	for _, r := range k.ruleList {
		if r.ID == "" {
			errs = append(errs, errors.New("rule id cannot be empty"))
			continue
		}
		if ruleIDs[r.ID] {
			errs = append(errs, fmt.Errorf("duplicate rule id: %q", r.ID))
		}
		ruleIDs[r.ID] = true

		if r.Title == "" {
			errs = append(errs, fmt.Errorf("rule %q title cannot be empty", r.ID))
		}
		if r.Summary == "" {
			errs = append(errs, fmt.Errorf("rule %q summary cannot be empty", r.ID))
		}
		if r.Details == "" {
			errs = append(errs, fmt.Errorf("rule %q details cannot be empty", r.ID))
		}
		if len(r.Guidelines) == 0 {
			errs = append(errs, fmt.Errorf("rule %q guidelines cannot be empty", r.ID))
		}
	}

	return errors.Join(errs...)
}
