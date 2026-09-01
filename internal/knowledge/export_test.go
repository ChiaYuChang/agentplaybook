package knowledge

// SetRolesForTest sets the roleList and roles map on Knowledge for validation tests.
func (k *Knowledge) SetRolesForTest(roles []RoleDefinition) {
	k.roleList = roles
	k.roles = make(map[Role]RoleDefinition, len(roles))
	for _, r := range roles {
		k.roles[r.Name] = r
	}
}

// SetArtifactsForTest sets the artifactList and artifacts map on Knowledge for validation tests.
func (k *Knowledge) SetArtifactsForTest(artifacts []Artifact) {
	k.artifactList = artifacts
	k.artifacts = make(map[string]Artifact, len(artifacts))
	for _, a := range artifacts {
		k.artifacts[a.Name] = a
	}
}

// SetFlowsForTest sets the flowList and flows map on Knowledge for validation tests.
func (k *Knowledge) SetFlowsForTest(flows []Flow) {
	k.flowList = flows
	k.flows = make(map[string]Flow, len(flows))
	for _, f := range flows {
		k.flows[f.Name] = f
	}
}
