package integration

import (
	"strings"
	"testing"

	"github.com/andrewesweet/nomad-driver-milo/tests/helpers"
	"github.com/stretchr/testify/require"
)

// Test Scenario 5: Exit Code Propagation
func TestExitCodePropagation(t *testing.T) {
	// Given a test JAR file exists at the configured location
	// And the JAR when executed:
	//   - Prints "Application encountered an error"
	//   - Exits with code 42

	// When the user executes: nomad job run exit-code-test.nomad
	jobID := helpers.RunJobAndWait(t, "../../tests/fixtures/jobs/exit-code.nomad", "dead")
	defer helpers.StopJob(t, jobID)

	// Then the task exit code should be 42
	exitCode := helpers.GetTaskExitCode(t, jobID, "java-app")
	require.Equal(t, 42, exitCode, "Task should exit with code 42")

	// And the job status should show "dead (failed)"
	// (Already verified by RunJobAndWait)

	// Verify the output contains the error message
	logs := helpers.GetLogs(t, jobID, "java-app")
	logs = strings.TrimSpace(logs)
	require.Equal(t, "Application encountered an error", logs, "Output should contain error message")
}
