package knowledge

import (
	"slices"
)

// Roles returns all registered role definitions in stable declaration order.
func (k *Knowledge) Roles() []RoleDefinition {
	return slices.Clone(k.roleList)
}

// Role returns the role definition for the given role identifier, if present.
func (k *Knowledge) Role(name string) (RoleDefinition, bool) {
	r, ok := k.roles[Role(name)]
	return r, ok
}

// Flows returns all registered flows in stable declaration order.
func (k *Knowledge) Flows() []Flow {
	return slices.Clone(k.flowList)
}

// Flow returns the flow definition for the given flow name, if present.
func (k *Knowledge) Flow(name string) (Flow, bool) {
	f, ok := k.flows[name]
	return f, ok
}

// FlowStep returns the specific step index from a named flow, if both exist.
func (k *Knowledge) FlowStep(flowName string, stepIndex int) (FlowStep, bool) {
	fl, ok := k.flows[flowName]
	if !ok {
		return FlowStep{}, false
	}
	for _, s := range fl.Steps {
		if s.Index == stepIndex {
			return s, true
		}
	}
	return FlowStep{}, false
}

// Artifacts returns all registered artifact definitions in stable declaration order.
func (k *Knowledge) Artifacts() []Artifact {
	return slices.Clone(k.artifactList)
}

// Artifact returns the artifact definition for the given artifact name, if present.
func (k *Knowledge) Artifact(name string) (Artifact, bool) {
	a, ok := k.artifacts[name]
	return a, ok
}

// Rules returns all registered rules in stable declaration order.
func (k *Knowledge) Rules() []Rule {
	return slices.Clone(k.ruleList)
}

// Rule returns the rule definition for the given rule ID, if present.
func (k *Knowledge) Rule(id string) (Rule, bool) {
	r, ok := k.rules[id]
	return r, ok
}
