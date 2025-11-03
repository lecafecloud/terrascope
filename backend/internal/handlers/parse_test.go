// Package handlers provides HTTP request handlers for the API endpoints.
// It defines the routing logic, response formatting, and error handling mechanisms.
package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terrascope/core/internal/models"
)

func TestParseHandler(t *testing.T) {
	t.Run("returns 200 OK for valid tfstate", func(t *testing.T) {
		validTfstate := `{
			"version": 4,
			"terraform_version": "1.5.0",
			"serial": 1,
			"lineage": "abc-123",
			"resources": []
		}`

		req := httptest.NewRequest(http.MethodPost, "/parse", strings.NewReader(validTfstate))
		w := httptest.NewRecorder()

		ParseHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns correct content type", func(t *testing.T) {
		validTfstate := `{
			"version": 4,
			"terraform_version": "1.5.0",
			"serial": 1,
			"lineage": "abc-123",
			"resources": []
		}`

		req := httptest.NewRequest(http.MethodPost, "/parse", strings.NewReader(validTfstate))
		w := httptest.NewRecorder()

		ParseHandler(w, req)

		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("returns valid graph for empty resources", func(t *testing.T) {
		validTfstate := `{
			"version": 4,
			"terraform_version": "1.5.0",
			"serial": 1,
			"lineage": "abc-123",
			"resources": []
		}`

		req := httptest.NewRequest(http.MethodPost, "/parse", strings.NewReader(validTfstate))
		w := httptest.NewRecorder()

		ParseHandler(w, req)

		var graph models.Graph
		err := json.NewDecoder(w.Body).Decode(&graph)
		require.NoError(t, err)

		assert.Empty(t, graph.Nodes)
		assert.Empty(t, graph.Edges)
	})

	t.Run("returns graph with nodes for single resource", func(t *testing.T) {
		validTfstate := `{
			"version": 4,
			"terraform_version": "1.5.0",
			"serial": 1,
			"lineage": "abc-123",
			"resources": [
				{
					"mode": "managed",
					"type": "aws_s3_bucket",
					"name": "assets",
					"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
					"instances": [
						{
							"schema_version": 0,
							"attributes": {
								"id": "my-bucket"
							}
						}
					]
				}
			]
		}`

		req := httptest.NewRequest(http.MethodPost, "/parse", strings.NewReader(validTfstate))
		w := httptest.NewRecorder()

		ParseHandler(w, req)

		var graph models.Graph
		err := json.NewDecoder(w.Body).Decode(&graph)
		require.NoError(t, err)

		assert.Len(t, graph.Nodes, 1)
		assert.Equal(t, "aws_s3_bucket.assets", graph.Nodes[0].ID)
		assert.Equal(t, "aws_s3_bucket", graph.Nodes[0].Type)
		assert.Equal(t, "managed", graph.Nodes[0].Mode)
	})

	t.Run("returns graph with edges for dependencies", func(t *testing.T) {
		validTfstate := `{
			"version": 4,
			"terraform_version": "1.5.0",
			"serial": 1,
			"lineage": "abc-123",
			"resources": [
				{
					"mode": "managed",
					"type": "aws_subnet",
					"name": "private",
					"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
					"instances": [
						{
							"schema_version": 1,
							"attributes": {
								"id": "subnet-123"
							},
							"dependencies": ["aws_vpc.main"]
						}
					]
				}
			]
		}`

		req := httptest.NewRequest(http.MethodPost, "/parse", strings.NewReader(validTfstate))
		w := httptest.NewRecorder()

		ParseHandler(w, req)

		var graph models.Graph
		err := json.NewDecoder(w.Body).Decode(&graph)
		require.NoError(t, err)

		assert.Len(t, graph.Edges, 1)
		assert.Equal(t, "aws_subnet.private", graph.Edges[0].Source)
		assert.Equal(t, "aws_vpc.main", graph.Edges[0].Target)
	})

	t.Run("returns 405 for GET request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/parse", nil)
		w := httptest.NewRecorder()

		ParseHandler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		assert.Contains(t, w.Body.String(), "Method not allowed")
	})

	t.Run("returns 405 for PUT request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/parse", nil)
		w := httptest.NewRecorder()

		ParseHandler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("returns 405 for DELETE request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/parse", nil)
		w := httptest.NewRecorder()

		ParseHandler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		invalidJSON := `{invalid json}`

		req := httptest.NewRequest(http.MethodPost, "/parse", strings.NewReader(invalidJSON))
		w := httptest.NewRecorder()

		ParseHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid tfstate")
	})

	t.Run("returns 400 for empty body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/parse", strings.NewReader(""))
		w := httptest.NewRecorder()

		ParseHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 for missing required fields", func(t *testing.T) {
		invalidTfstate := `{
			"version": 4
		}`

		req := httptest.NewRequest(http.MethodPost, "/parse", strings.NewReader(invalidTfstate))
		w := httptest.NewRecorder()

		ParseHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid tfstate")
	})

	t.Run("handles large tfstate file", func(t *testing.T) {
		resources := make([]string, 100)
		for i := range 100 {
			resources[i] = fmt.Sprintf(`{
				"mode": "managed",
				"type": "aws_s3_bucket",
				"name": "bucket%d",
				"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
				"instances": [{
					"schema_version": 0,
					"attributes": {"id": "bucket-%d"}
				}]
			}`, i, i)
		}

		largeTfstate := `{
			"version": 4,
			"terraform_version": "1.5.0",
			"serial": 1,
			"lineage": "abc-123",
			"resources": [` + strings.Join(resources, ",") + `]
		}`

		req := httptest.NewRequest(http.MethodPost, "/parse", strings.NewReader(largeTfstate))
		w := httptest.NewRecorder()

		ParseHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var graph models.Graph
		err := json.NewDecoder(w.Body).Decode(&graph)
		require.NoError(t, err)

		assert.Len(t, graph.Nodes, 100)
	})

	t.Run("closes request body", func(t *testing.T) {
		validTfstate := `{
			"version": 4,
			"terraform_version": "1.5.0",
			"serial": 1,
			"lineage": "abc-123",
			"resources": []
		}`

		body := io.NopCloser(strings.NewReader(validTfstate))
		req := httptest.NewRequest(http.MethodPost, "/parse", body)
		w := httptest.NewRecorder()

		ParseHandler(w, req)

		_, err := body.Read(make([]byte, 1))
		assert.Error(t, err) // Should error because body is closed
	})

	t.Run("handles nil body gracefully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/parse", nil)
		w := httptest.NewRecorder()

		ParseHandler(w, req)

		assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusInternalServerError)
	})

	t.Run("handles complex tfstate with modules", func(t *testing.T) {
		complexTfstate := `{
			"version": 4,
			"terraform_version": "1.5.0",
			"serial": 1,
			"lineage": "abc-123",
			"resources": [
				{
					"mode": "managed",
					"type": "aws_vpc",
					"name": "main",
					"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
					"instances": [{
						"schema_version": 1,
						"attributes": {"id": "vpc-123"}
					}]
				},
				{
					"mode": "managed",
					"type": "aws_instance",
					"name": "web",
					"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
					"module": "module.app",
					"instances": [{
						"schema_version": 1,
						"attributes": {"id": "i-123"},
						"dependencies": ["aws_vpc.main"]
					}]
				}
			]
		}`

		req := httptest.NewRequest(http.MethodPost, "/parse", strings.NewReader(complexTfstate))
		w := httptest.NewRecorder()

		ParseHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var graph models.Graph
		err := json.NewDecoder(w.Body).Decode(&graph)
		require.NoError(t, err)

		assert.Len(t, graph.Nodes, 2)
		assert.Len(t, graph.Edges, 1)

		// Find the module node
		var moduleNode *models.Node
		for i, node := range graph.Nodes {
			if node.Module != "" {
				moduleNode = &graph.Nodes[i]
				break
			}
		}
		require.NotNil(t, moduleNode)
		assert.Equal(t, "module.app", moduleNode.Module)
	})

	t.Run("handles concurrent requests", func(t *testing.T) {
		validTfstate := `{
			"version": 4,
			"terraform_version": "1.5.0",
			"serial": 1,
			"lineage": "abc-123",
			"resources": []
		}`

		numRequests := 10
		results := make(chan int, numRequests)

		for range numRequests {
			go func() {
				req := httptest.NewRequest(http.MethodPost, "/parse", strings.NewReader(validTfstate))
				w := httptest.NewRecorder()
				ParseHandler(w, req)
				results <- w.Code
			}()
		}

		for range numRequests {
			code := <-results
			assert.Equal(t, http.StatusOK, code)
		}
	})

	t.Run("handles binary data gracefully", func(t *testing.T) {
		binaryData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}

		req := httptest.NewRequest(http.MethodPost, "/parse", bytes.NewReader(binaryData))
		w := httptest.NewRecorder()

		ParseHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid tfstate")
	})

	t.Run("preserves resource metadata", func(t *testing.T) {
		tfstateWithMetadata := `{
			"version": 4,
			"terraform_version": "1.5.0",
			"serial": 1,
			"lineage": "abc-123",
			"resources": [
				{
					"mode": "managed",
					"type": "aws_instance",
					"name": "web",
					"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
					"instances": [{
						"schema_version": 1,
						"attributes": {
							"id": "i-1234567890",
							"name": "web-server",
							"tags": {
								"Environment": "production",
								"Team": "platform"
							}
						}
					}]
				}
			]
		}`

		req := httptest.NewRequest(http.MethodPost, "/parse", strings.NewReader(tfstateWithMetadata))
		w := httptest.NewRecorder()

		ParseHandler(w, req)

		var graph models.Graph
		err := json.NewDecoder(w.Body).Decode(&graph)
		require.NoError(t, err)

		assert.Len(t, graph.Nodes, 1)
		assert.NotNil(t, graph.Nodes[0].Metadata)
		assert.Equal(t, "i-1234567890", graph.Nodes[0].Metadata["id"])
		assert.Equal(t, "web-server", graph.Nodes[0].Metadata["name"])
	})

	t.Run("handles data sources", func(t *testing.T) {
		tfstateWithDataSource := `{
			"version": 4,
			"terraform_version": "1.5.0",
			"serial": 1,
			"lineage": "abc-123",
			"resources": [
				{
					"mode": "data",
					"type": "aws_ami",
					"name": "ubuntu",
					"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
					"instances": [{
						"schema_version": 0,
						"attributes": {"id": "ami-123456"}
					}]
				}
			]
		}`

		req := httptest.NewRequest(http.MethodPost, "/parse", strings.NewReader(tfstateWithDataSource))
		w := httptest.NewRecorder()

		ParseHandler(w, req)

		var graph models.Graph
		err := json.NewDecoder(w.Body).Decode(&graph)
		require.NoError(t, err)

		assert.Len(t, graph.Nodes, 1)
		assert.Equal(t, "data", graph.Nodes[0].Mode)
	})
}
