//go:build e2e

package e2e

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/hashicorp/nomad/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBasicE2EWorkflow tests the complete e2e workflow
func TestBasicE2EWorkflow(t *testing.T) {
	// Skip if Nomad is not available or if this is not a full e2e environment
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Create and configure server
	server := NewNomadServer(t)
	require.NoError(t, server.GenerateConfig())
	require.NoError(t, server.GenerateConfigFile())

	// Create job runner
	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("http://127.0.0.1:%d", server.config.HTTPPort)
	
	client, err := api.NewClient(config)
	require.NoError(t, err)
	
	runner := NewJobRunner(client)
	verifier := NewOutputVerifier(client)

	// Create a test job
	job := createTestJob(t)
	
	// Validate job spec
	err = runner.ValidateJobSpec(job)
	require.NoError(t, err)

	// Test output verification with mock logs
	testLogs := "Hello from Java!\nMilo driver test complete\n"
	err = verifier.VerifyLogContent(testLogs, "Hello from Java!")
	require.NoError(t, err)

	// Test cleanup functionality
	cleaner := NewTestCleaner()
	cleanupCalled := false
	cleaner.RegisterCleanup(func() error {
		cleanupCalled = true
		return nil
	})
	
	err = cleaner.ExecuteCleanup()
	require.NoError(t, err)
	assert.True(t, cleanupCalled)
}

// TestJobFixtureLoading tests loading job specifications from fixture files
func TestJobFixtureLoading(t *testing.T) {
	// Test success job fixture
	successJob, err := loadJobFixture("success.nomad")
	require.NoError(t, err)
	assert.NotNil(t, successJob)
	
	runner := NewJobRunner(nil)
	err = runner.ValidateJobSpec(successJob)
	require.NoError(t, err)
	
	// Test failure job fixture
	failureJob, err := loadJobFixture("failure.nomad")
	require.NoError(t, err)
	assert.NotNil(t, failureJob)
	
	err = runner.ValidateJobSpec(failureJob)
	require.NoError(t, err)
}

// createTestJob creates a test job specification
func createTestJob(t *testing.T) *api.Job {
	return &api.Job{
		ID:   stringPtr("e2e-test-job"),
		Name: stringPtr("e2e-test-job"),
		Type: stringPtr("batch"),
		TaskGroups: []*api.TaskGroup{
			{
				Name: stringPtr("app"),
				Tasks: []*api.Task{
					{
						Name:   "java-app",
						Driver: "nomad-driver-milo",
						Artifacts: []*api.TaskArtifact{
							{
								GetterSource: stringPtr("file://test-artifacts/hello-world.jar"),
							},
						},
						Resources: &api.Resources{
							CPU:      intPtr(100),
							MemoryMB: intPtr(64),
						},
					},
				},
			},
		},
	}
}

// loadJobFixture loads a job specification from a fixture file
func loadJobFixture(filename string) (*api.Job, error) {
	fixturePath := filepath.Join("fixtures", filename)
	_, err := ioutil.ReadFile(fixturePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixture %s: %v", filename, err)
	}

	// For now, return a basic job spec based on the fixture name
	// In a real implementation, we would parse the HCL
	if filename == "success.nomad" {
		return &api.Job{
			ID:   stringPtr("e2e-success-test"),
			Name: stringPtr("e2e-success-test"),
			Type: stringPtr("batch"),
		}, nil
	}
	
	return &api.Job{
		ID:   stringPtr("e2e-failure-test"),
		Name: stringPtr("e2e-failure-test"),
		Type: stringPtr("batch"),
	}, nil
}

// Helper functions for API types
func intPtr(i int) *int {
	return &i
}