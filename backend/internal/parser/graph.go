// Package parser provides utilities for parsing and transforming input data.
// It handles data normalization, validation, and conversion between formats.
package parser

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/terrascope/core/internal/models"
)

func BuildGraph(state *models.TerraformState) *models.Graph {
	graph := &models.Graph{
		Nodes: []models.Node{},
		Edges: []models.Edge{},
	}
	nodeMap := make(map[string]bool)

	for _, res := range state.Resources {
		for i, instance := range res.Instances {
			nodeID := buildNodeID(res, instance, i)

			if nodeMap[nodeID] {
				continue
			}

			node := models.Node{
				ID:       nodeID,
				Type:     res.Type,
				Mode:     res.Mode,
				Provider: extractProviderName(res.Provider),
				Module:   res.Module,
				Metadata: buildMetadata(res, instance),
			}

			graph.Nodes = append(graph.Nodes, node)
			nodeMap[nodeID] = true
			deps := collectDependencies(res.DependsOn, instance.Dependencies)

			for target, edgeType := range deps {
				graph.Edges = append(graph.Edges, models.Edge{
					Source: nodeID,
					Target: target,
					Type:   edgeType,
				})
			}
		}
	}

	return graph
}

func buildNodeID(res models.ResourceState, instance models.ResourceInstance, instanceIndex int) string {
	parts := []string{}

	if res.Module != "" {
		parts = append(parts, res.Module)
	}

	parts = append(parts, res.Type, res.Name)

	if len(res.Instances) > 1 {
		key := instance.IndexKey
		if key != nil {
			val := reflect.ValueOf(key)
			if val.Kind() == reflect.Ptr && !val.IsNil() {
				val = val.Elem()
			}
			parts = append(parts, fmt.Sprintf("[%v]", val.Interface()))
		} else {
			parts = append(parts, fmt.Sprintf("[%d]", instanceIndex))
		}
	}

	return strings.Join(parts, ".")
}

func extractProviderName(providerString string) string {
	providerString = strings.TrimPrefix(providerString, "provider[\"")
	providerString = strings.TrimSuffix(providerString, "\"]")

	parts := strings.Split(providerString, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return providerString
}

func buildMetadata(res models.ResourceState, instance models.ResourceInstance) map[string]any {
	metadata := map[string]any{
		"mode": res.Mode,
	}

	if id, ok := instance.Attributes["id"]; ok {
		metadata["id"] = id
	}

	if name, ok := instance.Attributes["name"]; ok {
		metadata["name"] = name
	}

	if arn, ok := instance.Attributes["arn"]; ok {
		metadata["arn"] = arn
	}

	if tags, ok := instance.Attributes["tags"].(map[string]any); ok {
		metadata["tags"] = tags
	}

	if instance.IndexKey != nil {
		metadata["index_key"] = instance.IndexKey
	}

	return metadata
}

func collectDependencies(explicit, implicit []string) map[string]string {
	deps := make(map[string]string)

	for _, dep := range explicit {
		deps[dep] = "depends_on"
	}

	for _, dep := range implicit {
		if _, exists := deps[dep]; !exists {
			deps[dep] = "implicit"
		}
	}

	return deps
}
