package integration

import (
	"testing"

	"github.com/andrewesweet/nomad-driver-milo/tests/helpers"
	"github.com/stretchr/testify/require"
)

// Test Scenario 3: Invalid File Extension
func TestInvalidFileExtension(t *testing.T) {
	// Given a Python script exists at the configured location

	// When the user executes: nomad job run invalid-test.nomad
	jobID := helpers.RunJobAndWait(t, "../../tests/fixtures/jobs/invalid.nomad", "dead")
	defer helpers.StopJob(t, jobID)

	// Then the job status should show "dead (failed)"
	// (Already verified by RunJobAndWait)

	// And running nomad logs should contain:
	//   "Error: Artifact must be a .jar file, got: HelloWorld.java"
	logs := helpers.GetLogs(t, jobID, "java-app")
	require.Contains(t, logs, "Error: Artifact validation failed", "Should contain validation error")
	require.Contains(t, logs, "Expected: A file with .jar extension", "Should contain expected format")
	require.Contains(t, logs, "Got: HelloWorld.java", "Should contain actual filename")
}
