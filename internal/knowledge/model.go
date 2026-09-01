package knowledge

// Config defines the global workflow configuration.
type Config struct {
	Languages     []string            `json:"languages"`
	Transport     string              `json:"transport"`
	Communication CommunicationConfig `json:"communication"`
}

// CommunicationConfig defines message prefix templates.
type CommunicationConfig struct {
	RolePrefixTemplate string `json:"role_prefix_template"`
	FlowPrefixTemplate string `json:"flow_prefix_template"`
}

// Role represents a participant role identifier.
type Role string

const (
	RolePlanner   Role = "planner"
	RoleBuilder   Role = "builder"
	RoleReviewer  Role = "reviewer"
	RoleScout     Role = "scout"
	RoleNavigator Role = "navigator"
	RoleUser      Role = "user"
)

// RoleDefinition defines the identity and boundaries of a participant role.
type RoleDefinition struct {
	Name             Role              `json:"name"`
	Title            string            `json:"title"`
	Category         string            `json:"category"`
	Description      string            `json:"description"`
	Responsibilities []string          `json:"responsibilities"`
	Boundaries       []string          `json:"boundaries"`
	Communication    RoleCommunication `json:"communication"`
}

// RoleCommunication specifies the communication targets for a role.
type RoleCommunication struct {
	Targets []Role `json:"targets"`
}

// Flow defines an end-to-end multi-agent orchestration state-machine.
type Flow struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Steps       []FlowStep `json:"steps"`
}

// FlowStep defines a discrete step within a flow.
type FlowStep struct {
	Index      int         `json:"index"`
	Actor      Role        `json:"actor"`
	Action     string      `json:"action"`
	Conditions []Condition `json:"conditions,omitempty"`
}

// Condition defines a semantic event condition and its transition target step index.
type Condition struct {
	When string `json:"when"`
	Then int    `json:"then"`
}

// Artifact defines an information contract for a produced, exchanged, or persisted object.
type Artifact struct {
	Name          string            `json:"name"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Owner         Role              `json:"owner"`
	Visibility    []Role            `json:"visibility"`
	Path          string            `json:"path,omitempty"`
	PathVariables map[string]string `json:"path_variables,omitempty"`
	Type          string            `json:"type"` // "document" or "message"
	Sections      []ArtifactSection `json:"sections,omitempty"`
	Fields        []ArtifactField   `json:"fields,omitempty"`
}

// ArtifactSection defines a required or optional section in a document artifact.
type ArtifactSection struct {
	Name        string `json:"name"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// ArtifactField defines a required or optional field in a message artifact.
type ArtifactField struct {
	Name        string `json:"name"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// Rule defines a contextual operational policy, invariant, or protocol.
type Rule struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Category   string   `json:"category"`
	Summary    string   `json:"summary"`
	Details    string   `json:"details"`
	Guidelines []string `json:"guidelines"`
}

// Knowledge represents the immutable aggregate knowledge store.
type Knowledge struct {
	Config Config

	roles        map[Role]RoleDefinition
	roleList     []RoleDefinition
	flows        map[string]Flow
	flowList     []Flow
	artifacts    map[string]Artifact
	artifactList []Artifact
	rules        map[string]Rule
	ruleList     []Rule
}
