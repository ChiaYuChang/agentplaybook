package knowledge

import (
	"encoding/json"
	"fmt"

	"github.com/ChiaYuChang/workflow/internal/data"
)

// Load unmarshals the embedded JSON datasets, indexes them into maps, validates all cross-references,
// and returns an immutable Knowledge aggregate.
func Load() (*Knowledge, error) {
	var cfg Config
	if err := json.Unmarshal(data.ConfigJSON, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config.json: %w", err)
	}

	var roles []RoleDefinition
	if err := json.Unmarshal(data.RolesJSON, &roles); err != nil {
		return nil, fmt.Errorf("unmarshal roles.json: %w", err)
	}

	var flows []Flow
	if err := json.Unmarshal(data.FlowsJSON, &flows); err != nil {
		return nil, fmt.Errorf("unmarshal flows.json: %w", err)
	}

	var artifacts []Artifact
	if err := json.Unmarshal(data.ArtifactsJSON, &artifacts); err != nil {
		return nil, fmt.Errorf("unmarshal artifacts.json: %w", err)
	}

	var rules []Rule
	if err := json.Unmarshal(data.RulesJSON, &rules); err != nil {
		return nil, fmt.Errorf("unmarshal rules.json: %w", err)
	}

	roleMap := make(map[Role]RoleDefinition, len(roles))
	for _, r := range roles {
		roleMap[r.Name] = r
	}

	flowMap := make(map[string]Flow, len(flows))
	for _, f := range flows {
		flowMap[f.Name] = f
	}

	artifactMap := make(map[string]Artifact, len(artifacts))
	for _, a := range artifacts {
		artifactMap[a.Name] = a
	}

	ruleMap := make(map[string]Rule, len(rules))
	for _, r := range rules {
		ruleMap[r.ID] = r
	}

	k := &Knowledge{
		Config:       cfg,
		roles:        roleMap,
		roleList:     roles,
		flows:        flowMap,
		flowList:     flows,
		artifacts:    artifactMap,
		artifactList: artifacts,
		rules:        ruleMap,
		ruleList:     rules,
	}

	if err := Validate(k); err != nil {
		return nil, err
	}

	return k, nil
}
