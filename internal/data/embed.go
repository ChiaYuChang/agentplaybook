package data

import _ "embed"

// ConfigJSON contains the embedded global workflow configuration.
//
//go:embed config.json
var ConfigJSON []byte

// RolesJSON contains the embedded role definitions.
//
//go:embed roles.json
var RolesJSON []byte

// FlowsJSON contains the embedded workflow procedure state-machines.
//
//go:embed flows.json
var FlowsJSON []byte

// ArtifactsJSON contains the embedded artifact schemas and contracts.
//
//go:embed artifacts.json
var ArtifactsJSON []byte

// RulesJSON contains the embedded operational policies and invariants.
//
//go:embed rules.json
var RulesJSON []byte
