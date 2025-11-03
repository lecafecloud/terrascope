// Package models defines the core data structures and database interaction logic.
// It includes entity definitions and methods for persistence and validation.
package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerraformStateUnmarshal(t *testing.T) {
	t.Run("minimal valid state", func(t *testing.T) {
		jsonData := `{
			"version": 4,
			"terraform_version": "1.5.0",
			"serial": 1,
			"lineage": "abc-123",
			"resources": []
		}`

		var state TerraformState
		err := json.Unmarshal([]byte(jsonData), &state)

		require.NoError(t, err)
		assert.Equal(t, 4, state.Version)
		assert.Equal(t, "1.5.0", state.TerraformVersion)
		assert.Equal(t, 1, state.Serial)
		assert.Equal(t, "abc-123", state.Lineage)
		assert.Empty(t, state.Resources)
	})

	t.Run("state with outputs", func(t *testing.T) {
		jsonData := `{
			"version": 4,
			"terraform_version": "1.5.0",
			"serial": 1,
			"lineage": "abc-123",
			"outputs": {
				"bucket_name": {
					"value": "my-bucket",
					"type": "string"
				},
				"instance_count": {
					"value": 3,
					"type": "number",
					"sensitive": false
				}
			},
			"resources": []
		}`

		var state TerraformState
		err := json.Unmarshal([]byte(jsonData), &state)

		require.NoError(t, err)
		assert.Len(t, state.Outputs, 2)
		assert.Equal(t, "my-bucket", state.Outputs["bucket_name"].Value)
		assert.Equal(t, float64(3), state.Outputs["instance_count"].Value)
		assert.False(t, state.Outputs["instance_count"].Sensitive)
	})

	t.Run("state with sensitive output", func(t *testing.T) {
		jsonData := `{
			"version": 4,
			"terraform_version": "1.5.0",
			"serial": 1,
			"lineage": "abc-123",
			"outputs": {
				"db_password": {
					"value": "secret",
					"type": "string",
					"sensitive": true
				}
			},
			"resources": []
		}`

		var state TerraformState
		err := json.Unmarshal([]byte(jsonData), &state)

		require.NoError(t, err)
		assert.True(t, state.Outputs["db_password"].Sensitive)
	})

	t.Run("state with managed resource", func(t *testing.T) {
		jsonData := `{
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
								"id": "my-bucket",
								"bucket": "my-bucket",
								"arn": "arn:aws:s3:::my-bucket"
							}
						}
					]
				}
			]
		}`

		var state TerraformState
		err := json.Unmarshal([]byte(jsonData), &state)

		require.NoError(t, err)
		assert.Len(t, state.Resources, 1)

		res := state.Resources[0]
		assert.Equal(t, "managed", res.Mode)
		assert.Equal(t, "aws_s3_bucket", res.Type)
		assert.Equal(t, "assets", res.Name)
		assert.Equal(t, "provider[\"registry.terraform.io/hashicorp/aws\"]", res.Provider)
		assert.Len(t, res.Instances, 1)
		assert.Equal(t, "my-bucket", res.Instances[0].Attributes["id"])
	})

	t.Run("state with data source", func(t *testing.T) {
		jsonData := `{
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
					"instances": [
						{
							"schema_version": 0,
							"attributes": {
								"id": "ami-123456",
								"name": "ubuntu-20.04"
							}
						}
					]
				}
			]
		}`

		var state TerraformState
		err := json.Unmarshal([]byte(jsonData), &state)

		require.NoError(t, err)
		assert.Equal(t, "data", state.Resources[0].Mode)
		assert.Equal(t, "aws_ami", state.Resources[0].Type)
	})

	t.Run("resource with module", func(t *testing.T) {
		jsonData := `{
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
					"module": "module.app",
					"instances": [
						{
							"schema_version": 1,
							"attributes": {
								"id": "i-1234567890"
							}
						}
					]
				}
			]
		}`

		var state TerraformState
		err := json.Unmarshal([]byte(jsonData), &state)

		require.NoError(t, err)
		assert.Equal(t, "module.app", state.Resources[0].Module)
	})

	t.Run("resource with depends_on", func(t *testing.T) {
		jsonData := `{
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
					"depends_on": [
						"aws_vpc.main",
						"aws_security_group.web"
					],
					"instances": [
						{
							"schema_version": 1,
							"attributes": {
								"id": "i-123"
							}
						}
					]
				}
			]
		}`

		var state TerraformState
		err := json.Unmarshal([]byte(jsonData), &state)

		require.NoError(t, err)
		assert.Len(t, state.Resources[0].DependsOn, 2)
		assert.Contains(t, state.Resources[0].DependsOn, "aws_vpc.main")
		assert.Contains(t, state.Resources[0].DependsOn, "aws_security_group.web")
	})

	t.Run("resource with multiple instances", func(t *testing.T) {
		jsonData := `{
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
								"id": "subnet-1"
							},
							"index_key": 0
						},
						{
							"schema_version": 1,
							"attributes": {
								"id": "subnet-2"
							},
							"index_key": 1
						}
					]
				}
			]
		}`

		var state TerraformState
		err := json.Unmarshal([]byte(jsonData), &state)

		require.NoError(t, err)
		assert.Len(t, state.Resources[0].Instances, 2)
		assert.Equal(t, float64(0), state.Resources[0].Instances[0].IndexKey)
		assert.Equal(t, float64(1), state.Resources[0].Instances[1].IndexKey)
	})

	t.Run("instance with dependencies", func(t *testing.T) {
		jsonData := `{
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
								"id": "subnet-1"
							},
							"dependencies": [
								"aws_vpc.main"
							]
						}
					]
				}
			]
		}`

		var state TerraformState
		err := json.Unmarshal([]byte(jsonData), &state)

		require.NoError(t, err)
		assert.Len(t, state.Resources[0].Instances[0].Dependencies, 1)
		assert.Equal(t, "aws_vpc.main", state.Resources[0].Instances[0].Dependencies[0])
	})

	t.Run("instance with private field", func(t *testing.T) {
		jsonData := `{
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
					"instances": [
						{
							"schema_version": 1,
							"attributes": {
								"id": "i-123"
							},
							"private": "eyJzY2hlbWFfdmVyc2lvbiI6IjEifQ=="
						}
					]
				}
			]
		}`

		var state TerraformState
		err := json.Unmarshal([]byte(jsonData), &state)

		require.NoError(t, err)
		assert.Equal(t, "eyJzY2hlbWFfdmVyc2lvbiI6IjEifQ==", state.Resources[0].Instances[0].Private)
	})
}

func TestTerraformStateMarshal(t *testing.T) {
	t.Run("marshal minimal state", func(t *testing.T) {
		state := TerraformState{
			Version:          4,
			TerraformVersion: "1.5.0",
			Serial:           1,
			Lineage:          "abc-123",
			Resources:        []ResourceState{},
		}

		data, err := json.Marshal(state)
		require.NoError(t, err)

		var decoded TerraformState
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, state.Version, decoded.Version)
		assert.Equal(t, state.TerraformVersion, decoded.TerraformVersion)
		assert.Equal(t, state.Serial, decoded.Serial)
		assert.Equal(t, state.Lineage, decoded.Lineage)
	})

	t.Run("marshal state with resources", func(t *testing.T) {
		state := TerraformState{
			Version:          4,
			TerraformVersion: "1.5.0",
			Serial:           1,
			Lineage:          "abc-123",
			Resources: []ResourceState{
				{
					Mode:     "managed",
					Type:     "aws_s3_bucket",
					Name:     "assets",
					Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
					Instances: []ResourceInstance{
						{
							SchemaVersion: 0,
							Attributes: map[string]any{
								"id":     "my-bucket",
								"bucket": "my-bucket",
							},
						},
					},
				},
			},
		}

		data, err := json.Marshal(state)
		require.NoError(t, err)

		var decoded TerraformState
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Len(t, decoded.Resources, 1)
		assert.Equal(t, "aws_s3_bucket", decoded.Resources[0].Type)
		assert.Equal(t, "my-bucket", decoded.Resources[0].Instances[0].Attributes["id"])
	})

	t.Run("omitempty fields are omitted when empty", func(t *testing.T) {
		state := TerraformState{
			Version:          4,
			TerraformVersion: "1.5.0",
			Serial:           1,
			Lineage:          "abc-123",
			Resources:        []ResourceState{},
		}

		data, err := json.Marshal(state)
		require.NoError(t, err)

		jsonString := string(data)
		assert.NotContains(t, jsonString, "outputs")
	})

	t.Run("omitempty fields are included when present", func(t *testing.T) {
		state := TerraformState{
			Version:          4,
			TerraformVersion: "1.5.0",
			Serial:           1,
			Lineage:          "abc-123",
			Outputs: map[string]Output{
				"bucket_name": {
					Value: "my-bucket",
					Type:  "string",
				},
			},
			Resources: []ResourceState{},
		}

		data, err := json.Marshal(state)
		require.NoError(t, err)

		jsonString := string(data)
		assert.Contains(t, jsonString, "outputs")
		assert.Contains(t, jsonString, "bucket_name")
	})
}

func TestResourceStateFields(t *testing.T) {
	t.Run("all fields present", func(t *testing.T) {
		jsonData := `{
			"mode": "managed",
			"type": "aws_instance",
			"name": "web",
			"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
			"module": "module.app",
			"depends_on": ["aws_vpc.main"],
			"instances": [
				{
					"schema_version": 1,
					"attributes": {"id": "i-123"},
					"dependencies": ["aws_subnet.private"],
					"index_key": 0,
					"private": "base64string"
				}
			]
		}`

		var res ResourceState
		err := json.Unmarshal([]byte(jsonData), &res)

		require.NoError(t, err)
		assert.Equal(t, "managed", res.Mode)
		assert.Equal(t, "aws_instance", res.Type)
		assert.Equal(t, "web", res.Name)
		assert.Equal(t, "module.app", res.Module)
		assert.Len(t, res.DependsOn, 1)
		assert.Len(t, res.Instances, 1)
		assert.Len(t, res.Instances[0].Dependencies, 1)
	})
}

func TestOutputFields(t *testing.T) {
	t.Run("output with all fields", func(t *testing.T) {
		jsonData := `{
			"value": "sensitive-data",
			"type": "string",
			"sensitive": true
		}`

		var output Output
		err := json.Unmarshal([]byte(jsonData), &output)

		require.NoError(t, err)
		assert.Equal(t, "sensitive-data", output.Value)
		assert.Equal(t, "string", output.Type)
		assert.True(t, output.Sensitive)
	})

	t.Run("output without sensitive field", func(t *testing.T) {
		jsonData := `{
			"value": "public-data",
			"type": "string"
		}`

		var output Output
		err := json.Unmarshal([]byte(jsonData), &output)

		require.NoError(t, err)
		assert.False(t, output.Sensitive)
	})
}
