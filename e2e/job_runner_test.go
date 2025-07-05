//go:build e2e

package e2e

import (
	"testing"

	"github.com/hashicorp/nomad/api"
	"github.com/stretchr/testify/assert"
)

// UT2.1: Create JobRunner with API client
func TestNewJobRunner_CreatesWithClient(t *testing.T) {
	client := &api.Client{} // mock client
	runner := NewJobRunner(client)
	assert.NotNil(t, runner)
	assert.Equal(t, client, runner.client)
}

// UT2.3: Validate job specification format
func TestValidateJobSpec_ChecksRequiredFields(t *testing.T) {
	runner := NewJobRunner(nil)

	// Valid job spec
	validJob := &api.Job{
		ID:   stringPtr("test"),
		Name: stringPtr("test"),
		Type: stringPtr("batch"),
	}
	err := runner.ValidateJobSpec(validJob)
	assert.NoError(t, err)

	// Invalid job spec (missing ID)
	invalidJob := &api.Job{
		Name: stringPtr("test"),
		Type: stringPtr("batch"),
	}
	err = runner.ValidateJobSpec(invalidJob)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID is required")
}

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}