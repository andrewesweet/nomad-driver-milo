package integration

import (
	"testing"
	"time"

	"github.com/andrewesweet/nomad-driver-milo/tests/helpers"
	"github.com/stretchr/testify/require"
)

// Test Scenario 2: Real-time Log Streaming
func TestRealTimeLogStreaming(t *testing.T) {
	// Given a test JAR file exists at the configured location
	// And the JAR when executed:
	//   - Prints "Starting application..." immediately
	//   - Prints "Processing..." every 2 seconds
	//   - Runs until terminated

	// When the user executes: nomad job run streaming-test.nomad
	jobID := helpers.RunJobAndWait(t, "../../tests/fixtures/jobs/streaming.nomad", "running")
	defer helpers.StopJob(t, jobID)

	// And waits 5 seconds
	// And executes: nomad logs -f streaming-test java-app
	logChan, cleanup := helpers.StreamLogs(t, jobID, "java-app")
	defer cleanup()

	// Then the log output should show:
	//   "Starting application..."
	//   "Processing..."
	//   "Processing..."
	// And new "Processing..." lines should appear every 2 seconds

	// Collect logs for 6 seconds
	var logs []string
	timeout := time.After(6 * time.Second)

	for {
		select {
		case log, ok := <-logChan:
			if !ok {
				t.Fatal("Log stream closed unexpectedly")
			}
			logs = append(logs, log)
		case <-timeout:
			goto validate
		}
	}

validate:
	// Verify we got the expected logs
	require.GreaterOrEqual(t, len(logs), 3, "Should have at least 3 log lines")
	require.Equal(t, "Starting application...", logs[0], "First log should be startup message")

	// Count "Processing..." messages
	processingCount := 0
	for _, log := range logs[1:] {
		if log == "Processing..." {
			processingCount++
		}
	}
	require.GreaterOrEqual(t, processingCount, 2, "Should have at least 2 'Processing...' messages")
}
