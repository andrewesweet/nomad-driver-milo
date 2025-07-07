package integration

import (
	"strings"
	"testing"

	"github.com/andrewesweet/nomad-driver-milo/tests/helpers"
	"github.com/stretchr/testify/require"
)

// Test Scenario 1: Successful JAR Execution
func TestSuccessfulJARExecution(t *testing.T) {
	// Given a host with Java runtime installed
	// And a test JAR file exists at the configured location
	// And the JAR when executed prints exactly:
	//   "Hello from Java!"
	//   "Milo driver test complete"
	// And the JAR exits with code 0

	// When the user executes: nomad job run test-job.nomad
	jobID := helpers.RunJobAndWait(t, "../../tests/fixtures/jobs/basic.nomad", "dead")
	defer helpers.StopJob(t, jobID)

	// Then the job status should show "dead (success)"
	// (Already verified by RunJobAndWait)

	// And the task exit code should be 0
	exitCode := helpers.GetTaskExitCode(t, jobID, "java-app")
	require.Equal(t, 0, exitCode, "Task should exit with code 0")

	// And running nomad logs should output exactly:
	//   "Hello from Java!"
	//   "Milo driver test complete"
	logs := helpers.GetLogs(t, jobID, "java-app")
	logs = strings.TrimSpace(logs)
	expectedOutput := "Hello from Java!\nMilo driver test complete"
	require.Equal(t, expectedOutput, logs, "Output should match expected")
}
