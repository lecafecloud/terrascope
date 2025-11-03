// Package parser provides utilities for parsing and transforming input data.
// It handles data normalization, validation, and conversion between formats.
package parser

import (
	"encoding/json"
	"fmt"

	"github.com/terrascope/core/internal/models"
)

func ParseTfstate(data []byte) (*models.TerraformState, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty tfstate data")
	}

	var state models.TerraformState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tfstate: %w", err)
	}

	if state.Version == 0 {
		return nil, fmt.Errorf("invalid tfstate: missing version field")
	}

	if state.TerraformVersion == "" {
		return nil, fmt.Errorf("invalid tfstate: missing terraform_version field")
	}

	return &state, nil
}
