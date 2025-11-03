// Package models defines the core data structures and database interaction logic.
// It includes entity definitions and methods for persistence and validation.
package models

type TerraformState struct {
	Version          int               `json:"version"`
	TerraformVersion string            `json:"terraform_version"`
	Serial           int               `json:"serial"`
	Lineage          string            `json:"lineage"`
	Outputs          map[string]Output `json:"outputs,omitempty"`
	Resources        []ResourceState   `json:"resources"`
}

type Output struct {
	Value     any  `json:"value"`
	Type      any  `json:"type"`
	Sensitive bool `json:"sensitive,omitempty"`
}

type ResourceState struct {
	Mode      string             `json:"mode"`
	Type      string             `json:"type"`
	Name      string             `json:"name"`
	Provider  string             `json:"provider"`
	Module    string             `json:"module,omitempty"`
	Instances []ResourceInstance `json:"instances"`
	DependsOn []string           `json:"depends_on,omitempty"`
}

type ResourceInstance struct {
	SchemaVersion  int               `json:"schema_version"`
	Attributes     map[string]any    `json:"attributes"`
	AttributesFlat map[string]string `json:"attributes_flat,omitempty"`
	Private        string            `json:"private,omitempty"`
	Dependencies   []string          `json:"dependencies,omitempty"`
	IndexKey       any               `json:"index_key,omitempty"`
}
