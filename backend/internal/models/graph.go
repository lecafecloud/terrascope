// Package models defines the core data structures and database interaction logic.
// It includes entity definitions and methods for persistence and validation.
package models

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
	Stats *Stats `json:"stats,omitempty"`
}

type Node struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Mode     string         `json:"mode"`
	Provider string         `json:"provider"`
	Module   string         `json:"module,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
}

type Stats struct {
	TotalNodes      int            `json:"total_nodes"`
	TotalEdges      int            `json:"total_edges"`
	ResourcesByType map[string]int `json:"resources_by_type,omitempty"`
	ResourcesByMode map[string]int `json:"resources_by_mode,omitempty"`
}
