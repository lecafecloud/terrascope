// Package models defines the core data structures and database interaction logic.
// It includes entity definitions and methods for persistence and validation.
package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphUnmarshal(t *testing.T) {
	t.Run("empty graph", func(t *testing.T) {
		jsonData := `{
			"nodes": [],
			"edges": []
		}`

		var graph Graph
		err := json.Unmarshal([]byte(jsonData), &graph)

		require.NoError(t, err)
		assert.Empty(t, graph.Nodes)
		assert.Empty(t, graph.Edges)
	})

	t.Run("graph with nodes and edges", func(t *testing.T) {
		jsonData := `{
			"nodes": [
				{
					"id": "aws_vpc.main",
					"type": "aws_vpc",
					"mode": "managed",
					"provider": "aws"
				},
				{
					"id": "aws_subnet.private",
					"type": "aws_subnet",
					"mode": "managed",
					"provider": "aws"
				}
			],
			"edges": [
				{
					"source": "aws_subnet.private",
					"target": "aws_vpc.main",
					"type": "implicit"
				}
			]
		}`

		var graph Graph
		err := json.Unmarshal([]byte(jsonData), &graph)

		require.NoError(t, err)
		assert.Len(t, graph.Nodes, 2)
		assert.Len(t, graph.Edges, 1)
		assert.Equal(t, "aws_vpc.main", graph.Nodes[0].ID)
		assert.Equal(t, "aws_subnet.private", graph.Edges[0].Source)
	})

	t.Run("graph with stats", func(t *testing.T) {
		jsonData := `{
			"nodes": [],
			"edges": [],
			"stats": {
				"total_nodes": 10,
				"total_edges": 5,
				"resources_by_type": {
					"aws_vpc": 1,
					"aws_subnet": 3
				},
				"resources_by_mode": {
					"managed": 8,
					"data": 2
				}
			}
		}`

		var graph Graph
		err := json.Unmarshal([]byte(jsonData), &graph)

		require.NoError(t, err)
		assert.Equal(t, 10, graph.Stats.TotalNodes)
		assert.Equal(t, 5, graph.Stats.TotalEdges)
		assert.Equal(t, 1, graph.Stats.ResourcesByType["aws_vpc"])
		assert.Equal(t, 8, graph.Stats.ResourcesByMode["managed"])
	})
}

func TestGraphMarshal(t *testing.T) {
	t.Run("marshal empty graph", func(t *testing.T) {
		graph := Graph{
			Nodes: []Node{},
			Edges: []Edge{},
		}

		data, err := json.Marshal(graph)
		require.NoError(t, err)

		var decoded Graph
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Empty(t, decoded.Nodes)
		assert.Empty(t, decoded.Edges)
	})

	t.Run("marshal graph with data", func(t *testing.T) {
		graph := Graph{
			Nodes: []Node{
				{
					ID:       "aws_vpc.main",
					Type:     "aws_vpc",
					Mode:     "managed",
					Provider: "aws",
				},
			},
			Edges: []Edge{
				{
					Source: "aws_subnet.private",
					Target: "aws_vpc.main",
					Type:   "implicit",
				},
			},
			Stats: &Stats{
				TotalNodes: 1,
				TotalEdges: 1,
			},
		}

		data, err := json.Marshal(graph)
		require.NoError(t, err)

		var decoded Graph
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Len(t, decoded.Nodes, 1)
		assert.Len(t, decoded.Edges, 1)
		assert.Equal(t, 1, decoded.Stats.TotalNodes)
	})

	t.Run("omitempty stats when not set", func(t *testing.T) {
		graph := Graph{
			Nodes: []Node{},
			Edges: []Edge{},
		}

		data, err := json.Marshal(graph)
		require.NoError(t, err)

		jsonString := string(data)
		assert.NotContains(t, jsonString, "stats")
	})
}

func TestNodeUnmarshal(t *testing.T) {
	t.Run("node with all required fields", func(t *testing.T) {
		jsonData := `{
			"id": "aws_s3_bucket.assets",
			"type": "aws_s3_bucket",
			"mode": "managed",
			"provider": "aws"
		}`

		var node Node
		err := json.Unmarshal([]byte(jsonData), &node)

		require.NoError(t, err)
		assert.Equal(t, "aws_s3_bucket.assets", node.ID)
		assert.Equal(t, "aws_s3_bucket", node.Type)
		assert.Equal(t, "managed", node.Mode)
		assert.Equal(t, "aws", node.Provider)
	})

	t.Run("node with module", func(t *testing.T) {
		jsonData := `{
			"id": "module.app.aws_instance.web",
			"type": "aws_instance",
			"mode": "managed",
			"provider": "aws",
			"module": "module.app"
		}`

		var node Node
		err := json.Unmarshal([]byte(jsonData), &node)

		require.NoError(t, err)
		assert.Equal(t, "module.app", node.Module)
	})

	t.Run("node with metadata", func(t *testing.T) {
		jsonData := `{
			"id": "aws_instance.web",
			"type": "aws_instance",
			"mode": "managed",
			"provider": "aws",
			"metadata": {
				"id": "i-1234567890",
				"name": "web-server",
				"tags": {
					"Environment": "production"
				}
			}
		}`

		var node Node
		err := json.Unmarshal([]byte(jsonData), &node)

		require.NoError(t, err)
		assert.NotNil(t, node.Metadata)
		assert.Equal(t, "i-1234567890", node.Metadata["id"])
		assert.Equal(t, "web-server", node.Metadata["name"])

		tags, ok := node.Metadata["tags"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "production", tags["Environment"])
	})

	t.Run("data source node", func(t *testing.T) {
		jsonData := `{
			"id": "data.aws_ami.ubuntu",
			"type": "aws_ami",
			"mode": "data",
			"provider": "aws"
		}`

		var node Node
		err := json.Unmarshal([]byte(jsonData), &node)

		require.NoError(t, err)
		assert.Equal(t, "data", node.Mode)
	})
}

func TestNodeMarshal(t *testing.T) {
	t.Run("marshal node without optional fields", func(t *testing.T) {
		node := Node{
			ID:       "aws_vpc.main",
			Type:     "aws_vpc",
			Mode:     "managed",
			Provider: "aws",
		}

		data, err := json.Marshal(node)
		require.NoError(t, err)

		jsonString := string(data)
		assert.NotContains(t, jsonString, "module")
		assert.NotContains(t, jsonString, "metadata")
	})

	t.Run("marshal node with all fields", func(t *testing.T) {
		node := Node{
			ID:       "module.app.aws_instance.web",
			Type:     "aws_instance",
			Mode:     "managed",
			Provider: "aws",
			Module:   "module.app",
			Metadata: map[string]any{
				"id":   "i-123",
				"name": "web-server",
			},
		}

		data, err := json.Marshal(node)
		require.NoError(t, err)

		var decoded Node
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, node.ID, decoded.ID)
		assert.Equal(t, node.Module, decoded.Module)
		assert.Equal(t, "i-123", decoded.Metadata["id"])
	})
}

func TestEdgeUnmarshal(t *testing.T) {
	t.Run("implicit dependency edge", func(t *testing.T) {
		jsonData := `{
			"source": "aws_subnet.private",
			"target": "aws_vpc.main",
			"type": "implicit"
		}`

		var edge Edge
		err := json.Unmarshal([]byte(jsonData), &edge)

		require.NoError(t, err)
		assert.Equal(t, "aws_subnet.private", edge.Source)
		assert.Equal(t, "aws_vpc.main", edge.Target)
		assert.Equal(t, "implicit", edge.Type)
	})

	t.Run("explicit depends_on edge", func(t *testing.T) {
		jsonData := `{
			"source": "aws_instance.web",
			"target": "aws_security_group.web",
			"type": "depends_on"
		}`

		var edge Edge
		err := json.Unmarshal([]byte(jsonData), &edge)

		require.NoError(t, err)
		assert.Equal(t, "depends_on", edge.Type)
	})
}

func TestEdgeMarshal(t *testing.T) {
	t.Run("marshal edge", func(t *testing.T) {
		edge := Edge{
			Source: "aws_subnet.private",
			Target: "aws_vpc.main",
			Type:   "implicit",
		}

		data, err := json.Marshal(edge)
		require.NoError(t, err)

		var decoded Edge
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, edge.Source, decoded.Source)
		assert.Equal(t, edge.Target, decoded.Target)
		assert.Equal(t, edge.Type, decoded.Type)
	})
}

func TestStatsUnmarshal(t *testing.T) {
	t.Run("stats with all fields", func(t *testing.T) {
		jsonData := `{
			"total_nodes": 15,
			"total_edges": 12,
			"resources_by_type": {
				"aws_vpc": 1,
				"aws_subnet": 3,
				"aws_instance": 5
			},
			"resources_by_mode": {
				"managed": 13,
				"data": 2
			}
		}`

		var stats Stats
		err := json.Unmarshal([]byte(jsonData), &stats)

		require.NoError(t, err)
		assert.Equal(t, 15, stats.TotalNodes)
		assert.Equal(t, 12, stats.TotalEdges)
		assert.Len(t, stats.ResourcesByType, 3)
		assert.Len(t, stats.ResourcesByMode, 2)
		assert.Equal(t, 1, stats.ResourcesByType["aws_vpc"])
		assert.Equal(t, 13, stats.ResourcesByMode["managed"])
	})

	t.Run("stats with minimal fields", func(t *testing.T) {
		jsonData := `{
			"total_nodes": 5,
			"total_edges": 3
		}`

		var stats Stats
		err := json.Unmarshal([]byte(jsonData), &stats)

		require.NoError(t, err)
		assert.Equal(t, 5, stats.TotalNodes)
		assert.Equal(t, 3, stats.TotalEdges)
		assert.Nil(t, stats.ResourcesByType)
		assert.Nil(t, stats.ResourcesByMode)
	})
}

func TestStatsMarshal(t *testing.T) {
	t.Run("marshal stats with all fields", func(t *testing.T) {
		stats := Stats{
			TotalNodes: 10,
			TotalEdges: 8,
			ResourcesByType: map[string]int{
				"aws_vpc":    1,
				"aws_subnet": 3,
			},
			ResourcesByMode: map[string]int{
				"managed": 8,
				"data":    2,
			},
		}

		data, err := json.Marshal(stats)
		require.NoError(t, err)

		var decoded Stats
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, stats.TotalNodes, decoded.TotalNodes)
		assert.Equal(t, stats.TotalEdges, decoded.TotalEdges)
		assert.Equal(t, stats.ResourcesByType["aws_vpc"], decoded.ResourcesByType["aws_vpc"])
		assert.Equal(t, stats.ResourcesByMode["managed"], decoded.ResourcesByMode["managed"])
	})

	t.Run("omitempty maps when nil", func(t *testing.T) {
		stats := Stats{
			TotalNodes: 5,
			TotalEdges: 3,
		}

		data, err := json.Marshal(stats)
		require.NoError(t, err)

		jsonString := string(data)
		assert.NotContains(t, jsonString, "resources_by_type")
		assert.NotContains(t, jsonString, "resources_by_mode")
	})
}

func TestCompleteGraphRoundTrip(t *testing.T) {
	t.Run("full graph serialization roundtrip", func(t *testing.T) {
		original := Graph{
			Nodes: []Node{
				{
					ID:       "aws_vpc.main",
					Type:     "aws_vpc",
					Mode:     "managed",
					Provider: "aws",
					Metadata: map[string]any{
						"id":   "vpc-123",
						"cidr": "10.0.0.0/16",
					},
				},
				{
					ID:       "module.app.aws_instance.web",
					Type:     "aws_instance",
					Mode:     "managed",
					Provider: "aws",
					Module:   "module.app",
					Metadata: map[string]any{
						"id": "i-123",
					},
				},
				{
					ID:       "data.aws_ami.ubuntu",
					Type:     "aws_ami",
					Mode:     "data",
					Provider: "aws",
				},
			},
			Edges: []Edge{
				{
					Source: "module.app.aws_instance.web",
					Target: "aws_vpc.main",
					Type:   "implicit",
				},
				{
					Source: "module.app.aws_instance.web",
					Target: "data.aws_ami.ubuntu",
					Type:   "implicit",
				},
			},
			Stats: &Stats{
				TotalNodes: 3,
				TotalEdges: 2,
				ResourcesByType: map[string]int{
					"aws_vpc":      1,
					"aws_instance": 1,
					"aws_ami":      1,
				},
				ResourcesByMode: map[string]int{
					"managed": 2,
					"data":    1,
				},
			},
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded Graph
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Len(t, decoded.Nodes, 3)
		assert.Len(t, decoded.Edges, 2)
		assert.Equal(t, original.Stats.TotalNodes, decoded.Stats.TotalNodes)
		assert.Equal(t, original.Stats.TotalEdges, decoded.Stats.TotalEdges)

		assert.Equal(t, "aws_vpc.main", decoded.Nodes[0].ID)
		assert.Equal(t, "module.app", decoded.Nodes[1].Module)
		assert.Equal(t, "data", decoded.Nodes[2].Mode)

		assert.Equal(t, "vpc-123", decoded.Nodes[0].Metadata["id"])

		assert.Equal(t, "module.app.aws_instance.web", decoded.Edges[0].Source)
		assert.Equal(t, "implicit", decoded.Edges[0].Type)
	})
}
