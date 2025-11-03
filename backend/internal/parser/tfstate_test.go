// Package parser provides utilities for parsing and transforming input data.
// It handles data normalization, validation, and conversion between formats.
package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTfstate_Valid(t *testing.T) {
	input := []byte(`{
		"version": 4,
		"terraform_version": "1.5.0",
		"serial": 1,
		"lineage": "abc-123",
		"resources": [
			{
				"mode": "managed",
				"type": "aws_s3_bucket",
				"name": "my_bucket",
				"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
				"instances": [
					{
						"schema_version": 0,
						"attributes": {
							"bucket": "my-test-bucket",
							"region": "us-east-1"
						}
					}
				]
			}
		]
	}`)

	state, err := ParseTfstate(input)

	require.NoError(t, err)
	assert.Equal(t, 4, state.Version)
	assert.Equal(t, "1.5.0", state.TerraformVersion)
	assert.Len(t, state.Resources, 1)
	assert.Equal(t, "aws_s3_bucket", state.Resources[0].Type)
	assert.Equal(t, "my_bucket", state.Resources[0].Name)
}

func TestParseTfstate_Empty(t *testing.T) {
	_, err := ParseTfstate([]byte{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty tfstate data")
}

func TestParseTfstate_InvalidJSON(t *testing.T) {
	_, err := ParseTfstate([]byte(`{invalid json`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal")
}

func TestParseTfstate_MissingVersion(t *testing.T) {
	input := []byte(`{
		"terraform_version": "1.5.0",
		"resources": []
	}`)

	_, err := ParseTfstate(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing version")
}

func TestParseTfstate_MissingTerraformVersion(t *testing.T) {
	input := []byte(`{
		"version": 4,
		"resources": []
	}`)

	_, err := ParseTfstate(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing terraform_version")
}
