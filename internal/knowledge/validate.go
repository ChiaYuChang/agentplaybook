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
		if r.Name == RoleUser {
			errs = append(errs, errors.New("user cannot be registered as an internal role"))
		}
		if roleNames[r.Name] {
			errs = append(errs, fmt.Errorf("duplicate role name: %q", r.Name))
		}
		roleNames[r.Name] = true

		if r.Category != "core" && r.Category != "companion" {
			errs = append(errs, fmt.Errorf("role %q has invalid category %q (must be 'core' or 'companion')", r.Name, r.Category))
		}

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
		targetSet := make(map[Role]bool, len(r.Communication.Targets))
		for _, target := range r.Communication.Targets {
			if targetSet[target] {
				errs = append(errs, fmt.Errorf("role %q has duplicate communication target %q", r.Name, target))
			}
			if target == RoleUser {
				if r.Name != RoleNavigator && r.Name != RoleCartographer {
					errs = append(errs, fmt.Errorf("role %q cannot communicate with external user (permitted exclusively for companion roles navigator and cartographer)", r.Name))
				}
			} else if !roleNames[target] {
				errs = append(errs, fmt.Errorf("role %q communication target %q does not exist", r.Name, target))
			}
			targetSet[target] = true
		}

		switch r.Name {
		case RolePlanner:
			if len(r.Communication.Targets) != 5 {
				errs = append(errs, fmt.Errorf("planner communication targets must contain exactly 5 targets, got %d", len(r.Communication.Targets)))
			}
			for _, required := range []Role{RoleBuilder, RoleReviewer, RoleScout, RoleNavigator, RoleCartographer} {
				if !targetSet[required] {
					errs = append(errs, fmt.Errorf("planner communication targets must include %q", required))
				}
			}
		case RoleNavigator:
			if len(r.Communication.Targets) != 3 {
				errs = append(errs, fmt.Errorf("navigator communication targets must contain exactly 3 targets ('user', 'planner', and 'cartographer'), got %d", len(r.Communication.Targets)))
			}
			for _, required := range []Role{RoleUser, RolePlanner, RoleCartographer} {
				if !targetSet[required] {
					errs = append(errs, fmt.Errorf("navigator communication targets must include %q", required))
				}
			}
			for target := range targetSet {
				if target != RoleUser && target != RolePlanner && target != RoleCartographer {
					errs = append(errs, fmt.Errorf("navigator communication target %q is not permitted (must be 'user', 'planner', or 'cartographer')", target))
				}
			}
		case RoleCartographer:
			if len(r.Communication.Targets) != 3 {
				errs = append(errs, fmt.Errorf("cartographer communication targets must contain exactly 3 targets ('user', 'planner', and 'navigator'), got %d", len(r.Communication.Targets)))
			}
			for _, required := range []Role{RoleUser, RolePlanner, RoleNavigator} {
				if !targetSet[required] {
					errs = append(errs, fmt.Errorf("cartographer communication targets must include %q", required))
				}
			}
			for target := range targetSet {
				if target != RoleUser && target != RolePlanner && target != RoleNavigator {
					errs = append(errs, fmt.Errorf("cartographer communication target %q is not permitted (must be 'user', 'planner', or 'navigator')", target))
				}
			}
		case RoleBuilder, RoleReviewer, RoleScout:
			if len(r.Communication.Targets) != 1 || !targetSet[RolePlanner] {
				errs = append(errs, fmt.Errorf("role %q communication targets must contain strictly 'planner'", r.Name))
			}
		}
	}

	// 3. Validate Artifacts
	allowedNavigatorArtifacts := map[string]bool{
		"agents-md":             true,
		"sub-review-resolution": true,
		"review-resolution":     true,
		"diagram-brief":         true,
		"diagram-completion":    true,
	}
	allowedCartographerArtifacts := map[string]bool{
		"agents-md":             true,
		"sub-review-resolution": true,
		"review-resolution":     true,
		"diagram-brief":         true,
		"diagram-completion":    true,
	}

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

		if a.Owner == RoleUser {
			errs = append(errs, fmt.Errorf("artifact %q owner cannot be user", a.Name))
		} else if a.Owner == RoleNavigator {
			errs = append(errs, fmt.Errorf("artifact %q owner cannot be companion role navigator", a.Name))
		} else if a.Owner == RoleCartographer && a.Name != "diagram-completion" {
			errs = append(errs, fmt.Errorf("artifact %q owner cannot be companion role cartographer (permitted exclusively for diagram-completion)", a.Name))
		} else if !roleNames[a.Owner] {
			errs = append(errs, fmt.Errorf("artifact %q owner %q does not exist", a.Name, a.Owner))
		}

		if len(a.Visibility) == 0 {
			errs = append(errs, fmt.Errorf("artifact %q visibility cannot be empty", a.Name))
		}
		for _, v := range a.Visibility {
			if v == RoleUser {
				errs = append(errs, fmt.Errorf("artifact %q visibility cannot include user", a.Name))
			} else if !roleNames[v] {
				errs = append(errs, fmt.Errorf("artifact %q visibility role %q does not exist", a.Name, v))
			} else if v == RoleNavigator && !allowedNavigatorArtifacts[a.Name] {
				errs = append(errs, fmt.Errorf("artifact %q is not in the settled allowlist and cannot include companion role navigator in visibility", a.Name))
			} else if v == RoleCartographer && !allowedCartographerArtifacts[a.Name] {
				errs = append(errs, fmt.Errorf("artifact %q is not in the settled allowlist and cannot include companion role cartographer in visibility", a.Name))
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

			if s.Actor == RoleUser {
				errs = append(errs, fmt.Errorf("flow %q step %d actor cannot be user", f.Name, s.Index))
			} else if s.Actor == RoleNavigator {
				errs = append(errs, fmt.Errorf("flow %q step %d actor cannot be companion role navigator", f.Name, s.Index))
			} else if s.Actor == RoleCartographer && f.Name != "cartography" {
				errs = append(errs, fmt.Errorf("flow %q step %d actor cannot be companion role cartographer (permitted exclusively in cartography flow)", f.Name, s.Index))
			} else if !roleNames[s.Actor] {
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
