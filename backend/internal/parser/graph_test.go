// Package parser provides utilities for parsing and transforming input data.
// It handles data normalization, validation, and conversion between formats.
package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terrascope/core/internal/models"
)

func TestBuildGraph(t *testing.T) {
	t.Run("empty state returns empty graph", func(t *testing.T) {
		state := &models.TerraformState{
			Resources: []models.ResourceState{},
		}

		graph := BuildGraph(state)

		assert.NotNil(t, graph)
		assert.Empty(t, graph.Nodes)
		assert.Empty(t, graph.Edges)
	})

	t.Run("single resource creates single node", func(t *testing.T) {
		state := &models.TerraformState{
			Resources: []models.ResourceState{
				{
					Type:     "aws_s3_bucket",
					Name:     "assets",
					Mode:     "managed",
					Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
					Module:   "",
					Instances: []models.ResourceInstance{
						{
							Attributes: map[string]any{
								"id":   "my-bucket",
								"name": "assets-bucket",
							},
						},
					},
				},
			},
		}

		graph := BuildGraph(state)

		assert.Len(t, graph.Nodes, 1)
		assert.Equal(t, "aws_s3_bucket.assets", graph.Nodes[0].ID)
		assert.Equal(t, "aws_s3_bucket", graph.Nodes[0].Type)
		assert.Equal(t, "managed", graph.Nodes[0].Mode)
		assert.Equal(t, "aws", graph.Nodes[0].Provider)
		assert.Empty(t, graph.Edges)
	})

	t.Run("resource with module prefix", func(t *testing.T) {
		state := &models.TerraformState{
			Resources: []models.ResourceState{
				{
					Type:     "aws_instance",
					Name:     "web",
					Mode:     "managed",
					Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
					Module:   "module.app",
					Instances: []models.ResourceInstance{
						{
							Attributes: map[string]any{
								"id": "i-1234567890",
							},
						},
					},
				},
			},
		}

		graph := BuildGraph(state)

		assert.Len(t, graph.Nodes, 1)
		assert.Equal(t, "module.app.aws_instance.web", graph.Nodes[0].ID)
		assert.Equal(t, "module.app", graph.Nodes[0].Module)
	})

	t.Run("resource with multiple instances creates indexed nodes", func(t *testing.T) {
		state := &models.TerraformState{
			Resources: []models.ResourceState{
				{
					Type:     "aws_subnet",
					Name:     "private",
					Mode:     "managed",
					Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
					Instances: []models.ResourceInstance{
						{
							Attributes: map[string]any{"id": "subnet-1"},
							IndexKey:   intPtr(0),
						},
						{
							Attributes: map[string]any{"id": "subnet-2"},
							IndexKey:   intPtr(1),
						},
					},
				},
			},
		}

		graph := BuildGraph(state)

		assert.Len(t, graph.Nodes, 2)
		assert.Equal(t, "aws_subnet.private.[0]", graph.Nodes[0].ID)
		assert.Equal(t, "aws_subnet.private.[1]", graph.Nodes[1].ID)
	})

	t.Run("duplicate nodes are not added", func(t *testing.T) {
		state := &models.TerraformState{
			Resources: []models.ResourceState{
				{
					Type:     "aws_vpc",
					Name:     "main",
					Mode:     "managed",
					Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
					Instances: []models.ResourceInstance{
						{Attributes: map[string]any{"id": "vpc-1"}},
					},
				},
				{
					Type:     "aws_vpc",
					Name:     "main",
					Mode:     "managed",
					Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
					Instances: []models.ResourceInstance{
						{Attributes: map[string]any{"id": "vpc-1"}},
					},
				},
			},
		}

		graph := BuildGraph(state)

		assert.Len(t, graph.Nodes, 1)
	})

	t.Run("explicit depends_on creates edges", func(t *testing.T) {
		state := &models.TerraformState{
			Resources: []models.ResourceState{
				{
					Type:      "aws_instance",
					Name:      "web",
					Mode:      "managed",
					Provider:  "provider[\"registry.terraform.io/hashicorp/aws\"]",
					DependsOn: []string{"aws_vpc.main"},
					Instances: []models.ResourceInstance{
						{Attributes: map[string]any{"id": "i-123"}},
					},
				},
			},
		}

		graph := BuildGraph(state)

		assert.Len(t, graph.Edges, 1)
		assert.Equal(t, "aws_instance.web", graph.Edges[0].Source)
		assert.Equal(t, "aws_vpc.main", graph.Edges[0].Target)
		assert.Equal(t, "depends_on", graph.Edges[0].Type)
	})

	t.Run("implicit dependencies create edges", func(t *testing.T) {
		state := &models.TerraformState{
			Resources: []models.ResourceState{
				{
					Type:     "aws_subnet",
					Name:     "private",
					Mode:     "managed",
					Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
					Instances: []models.ResourceInstance{
						{
							Attributes:   map[string]any{"id": "subnet-1"},
							Dependencies: []string{"aws_vpc.main"},
						},
					},
				},
			},
		}

		graph := BuildGraph(state)

		assert.Len(t, graph.Edges, 1)
		assert.Equal(t, "aws_subnet.private", graph.Edges[0].Source)
		assert.Equal(t, "aws_vpc.main", graph.Edges[0].Target)
		assert.Equal(t, "implicit", graph.Edges[0].Type)
	})

	t.Run("both explicit and implicit dependencies", func(t *testing.T) {
		state := &models.TerraformState{
			Resources: []models.ResourceState{
				{
					Type:      "aws_instance",
					Name:      "web",
					Mode:      "managed",
					Provider:  "provider[\"registry.terraform.io/hashicorp/aws\"]",
					DependsOn: []string{"aws_security_group.web"},
					Instances: []models.ResourceInstance{
						{
							Attributes:   map[string]any{"id": "i-123"},
							Dependencies: []string{"aws_subnet.private"},
						},
					},
				},
			},
		}

		graph := BuildGraph(state)

		assert.Len(t, graph.Edges, 2)

		edgeTypes := make(map[string]bool)
		for _, edge := range graph.Edges {
			edgeTypes[edge.Type] = true
		}

		assert.True(t, edgeTypes["depends_on"])
		assert.True(t, edgeTypes["implicit"])
	})

	t.Run("data source mode", func(t *testing.T) {
		state := &models.TerraformState{
			Resources: []models.ResourceState{
				{
					Type:     "aws_ami",
					Name:     "ubuntu",
					Mode:     "data",
					Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
					Instances: []models.ResourceInstance{
						{Attributes: map[string]any{"id": "ami-123"}},
					},
				},
			},
		}

		graph := BuildGraph(state)

		assert.Len(t, graph.Nodes, 1)
		assert.Equal(t, "data", graph.Nodes[0].Mode)
		assert.Equal(t, "data", graph.Nodes[0].Metadata["mode"])
	})
}

func TestBuildNodeID(t *testing.T) {
	t.Run("simple resource without module", func(t *testing.T) {
		res := models.ResourceState{
			Type: "aws_s3_bucket",
			Name: "assets",
			Instances: []models.ResourceInstance{
				{Attributes: map[string]any{}},
			},
		}

		id := buildNodeID(res, res.Instances[0], 0)

		assert.Equal(t, "aws_s3_bucket.assets", id)
	})

	t.Run("resource with module", func(t *testing.T) {
		res := models.ResourceState{
			Type:   "aws_instance",
			Name:   "web",
			Module: "module.app.module.compute",
			Instances: []models.ResourceInstance{
				{Attributes: map[string]any{}},
			},
		}

		id := buildNodeID(res, res.Instances[0], 0)

		assert.Equal(t, "module.app.module.compute.aws_instance.web", id)
	})

	t.Run("resource with multiple instances adds index", func(t *testing.T) {
		res := models.ResourceState{
			Type: "aws_subnet",
			Name: "private",
			Instances: []models.ResourceInstance{
				{Attributes: map[string]any{}},
				{Attributes: map[string]any{}},
			},
		}

		id0 := buildNodeID(res, res.Instances[0], 0)
		id1 := buildNodeID(res, res.Instances[1], 1)

		assert.Equal(t, "aws_subnet.private.[0]", id0)
		assert.Equal(t, "aws_subnet.private.[1]", id1)
	})

	t.Run("single instance does not add index", func(t *testing.T) {
		res := models.ResourceState{
			Type: "aws_vpc",
			Name: "main",
			Instances: []models.ResourceInstance{
				{Attributes: map[string]any{}},
			},
		}

		id := buildNodeID(res, res.Instances[0], 0)

		assert.Equal(t, "aws_vpc.main", id)
	})

	t.Run("multiple instances with IndexKey use the key instead of index", func(t *testing.T) {
		res := models.ResourceState{
			Type: "aws_security_group",
			Name: "sg",
			Instances: []models.ResourceInstance{
				{IndexKey: "frontend", Attributes: map[string]any{}},
				{IndexKey: "backend", Attributes: map[string]any{}},
			},
		}

		id0 := buildNodeID(res, res.Instances[0], 0)
		id1 := buildNodeID(res, res.Instances[1], 1)

		assert.Equal(t, "aws_security_group.sg.[frontend]", id0)
		assert.Equal(t, "aws_security_group.sg.[backend]", id1)
	})
}

func TestExtractProviderName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard AWS provider",
			input:    "provider[\"registry.terraform.io/hashicorp/aws\"]",
			expected: "aws",
		},
		{
			name:     "GCP provider",
			input:    "provider[\"registry.terraform.io/hashicorp/google\"]",
			expected: "google",
		},
		{
			name:     "Azure provider",
			input:    "provider[\"registry.terraform.io/hashicorp/azurerm\"]",
			expected: "azurerm",
		},
		{
			name:     "custom provider with namespace",
			input:    "provider[\"registry.terraform.io/mycorp/custom\"]",
			expected: "custom",
		},
		{
			name:     "short format",
			input:    "provider[\"aws\"]",
			expected: "aws",
		},
		{
			name:     "already clean name",
			input:    "aws",
			expected: "aws",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractProviderName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildMetadata(t *testing.T) {
	t.Run("includes mode", func(t *testing.T) {
		res := models.ResourceState{Mode: "managed"}
		instance := models.ResourceInstance{Attributes: map[string]any{}}

		metadata := buildMetadata(res, instance)

		assert.Equal(t, "managed", metadata["mode"])
	})

	t.Run("includes id when present", func(t *testing.T) {
		res := models.ResourceState{Mode: "managed"}
		instance := models.ResourceInstance{
			Attributes: map[string]any{
				"id": "resource-123",
			},
		}

		metadata := buildMetadata(res, instance)

		assert.Equal(t, "resource-123", metadata["id"])
	})

	t.Run("includes name when present", func(t *testing.T) {
		res := models.ResourceState{Mode: "managed"}
		instance := models.ResourceInstance{
			Attributes: map[string]any{
				"name": "my-resource",
			},
		}

		metadata := buildMetadata(res, instance)

		assert.Equal(t, "my-resource", metadata["name"])
	})

	t.Run("includes arn when present", func(t *testing.T) {
		res := models.ResourceState{Mode: "managed"}
		instance := models.ResourceInstance{
			Attributes: map[string]any{
				"arn": "arn:aws:s3:::my-bucket",
			},
		}

		metadata := buildMetadata(res, instance)

		assert.Equal(t, "arn:aws:s3:::my-bucket", metadata["arn"])
	})

	t.Run("includes tags when present", func(t *testing.T) {
		res := models.ResourceState{Mode: "managed"}
		tags := map[string]any{
			"Environment": "production",
			"Owner":       "platform-team",
		}
		instance := models.ResourceInstance{
			Attributes: map[string]any{
				"tags": tags,
			},
		}

		metadata := buildMetadata(res, instance)

		assert.Equal(t, tags, metadata["tags"])
	})

	t.Run("includes index_key when present", func(t *testing.T) {
		res := models.ResourceState{Mode: "managed"}
		indexKey := 42
		instance := models.ResourceInstance{
			Attributes: map[string]any{},
			IndexKey:   &indexKey,
		}

		metadata := buildMetadata(res, instance)

		assert.Equal(t, &indexKey, metadata["index_key"])
	})

	t.Run("omits optional fields when not present", func(t *testing.T) {
		res := models.ResourceState{Mode: "data"}
		instance := models.ResourceInstance{
			Attributes: map[string]any{},
		}

		metadata := buildMetadata(res, instance)

		assert.Equal(t, "data", metadata["mode"])
		assert.NotContains(t, metadata, "id")
		assert.NotContains(t, metadata, "name")
		assert.NotContains(t, metadata, "arn")
		assert.NotContains(t, metadata, "tags")
		assert.NotContains(t, metadata, "index_key")
	})

	t.Run("handles all fields together", func(t *testing.T) {
		res := models.ResourceState{Mode: "managed"}
		indexKey := 1
		instance := models.ResourceInstance{
			Attributes: map[string]any{
				"id":   "i-123",
				"name": "web-server",
				"arn":  "arn:aws:ec2:us-east-1:123456789012:instance/i-123",
				"tags": map[string]any{
					"Name": "web-server",
				},
			},
			IndexKey: &indexKey,
		}

		metadata := buildMetadata(res, instance)

		assert.Equal(t, "managed", metadata["mode"])
		assert.Equal(t, "i-123", metadata["id"])
		assert.Equal(t, "web-server", metadata["name"])
		assert.Equal(t, "arn:aws:ec2:us-east-1:123456789012:instance/i-123", metadata["arn"])
		assert.NotNil(t, metadata["tags"])
		assert.Equal(t, &indexKey, metadata["index_key"])
	})
}

func intPtr(i int) *int {
	return &i
}
