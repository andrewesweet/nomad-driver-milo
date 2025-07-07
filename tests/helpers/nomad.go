package helpers

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// RunJobAndWait runs a Nomad job and waits for it to reach the expected status
func RunJobAndWait(t *testing.T, jobFile string, expectedStatus string) string {
	// Submit the job
	cmd := exec.Command("nomad", "job", "run", jobFile)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to submit job: %s", string(output))

	// Extract job ID from output
	jobID := extractJobID(string(output))
	require.NotEmpty(t, jobID, "Failed to extract job ID from output")

	// Wait for job to reach expected status
	waitForJobStatus(t, jobID, expectedStatus)

	return jobID
}

// extractJobID extracts the job ID from nomad job run output
func extractJobID(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Job registration successful") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "ID:" && i+1 < len(parts) {
					return parts[i+1]
				}
			}
		}
	}
	return ""
}

// waitForJobStatus waits for a job to reach the expected status
func waitForJobStatus(t *testing.T, jobID, expectedStatus string) {
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		status := getJobStatus(t, jobID)
		if strings.Contains(status, expectedStatus) {
			return
		}
		time.Sleep(time.Second)
	}
	t.Fatalf("Job %s did not reach status %s within timeout", jobID, expectedStatus)
}

// getJobStatus gets the current status of a job
func getJobStatus(t *testing.T, jobID string) string {
	cmd := exec.Command("nomad", "job", "status", jobID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Job might not exist yet
		return ""
	}
	return string(output)
}

// GetTaskExitCode extracts the exit code from a task
func GetTaskExitCode(t *testing.T, jobID, taskName string) int {
	status := getJobStatus(t, jobID)
	lines := strings.Split(status, "\n")

	// Look for the task status section
	inAllocations := false
	for _, line := range lines {
		if strings.Contains(line, "Allocations") {
			inAllocations = true
			continue
		}
		if inAllocations && strings.Contains(line, taskName) && strings.Contains(line, "Exit Code:") {
			// Extract exit code
			parts := strings.Split(line, "Exit Code:")
			if len(parts) >= 2 {
				codePart := strings.TrimSpace(parts[1])
				// The exit code is the first field after "Exit Code:"
				codeFields := strings.Fields(codePart)
				if len(codeFields) > 0 {
					var exitCode int
					_, err := fmt.Sscanf(codeFields[0], "%d", &exitCode)
					require.NoError(t, err)
					return exitCode
				}
			}
		}
	}

	t.Fatalf("Could not find exit code for task %s in job %s", taskName, jobID)
	return -1
}

// GetLogs gets the logs for a task
func GetLogs(t *testing.T, jobID, taskName string) string {
	cmd := exec.Command("nomad", "alloc", "logs", "-job", jobID, taskName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try to get stderr logs if stdout fails
		cmd = exec.Command("nomad", "alloc", "logs", "-stderr", "-job", jobID, taskName)
		stderrOutput, _ := cmd.CombinedOutput()
		return string(output) + "\n" + string(stderrOutput)
	}
	return string(output)
}

// StreamLogs streams logs from a task
func StreamLogs(t *testing.T, jobID, taskName string) (chan string, func()) {
	cmd := exec.Command("nomad", "alloc", "logs", "-f", "-job", jobID, taskName)

	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)

	err = cmd.Start()
	require.NoError(t, err)

	logChan := make(chan string, 100)
	done := make(chan struct{})

	// Read logs in background
	go func() {
		buf := make([]byte, 1024)
		var leftover []byte
		for {
			select {
			case <-done:
				close(logChan)
				return
			default:
				n, err := stdout.Read(buf)
				if err != nil {
					close(logChan)
					return
				}
				if n > 0 {
					data := append(leftover, buf[:n]...)
					lines := bytes.Split(data, []byte{'\n'})

					// Process complete lines
					for i := 0; i < len(lines)-1; i++ {
						if len(lines[i]) > 0 {
							logChan <- string(lines[i])
						}
					}

					// Save incomplete line for next iteration
					leftover = lines[len(lines)-1]
				}
			}
		}
	}()

	cleanup := func() {
		close(done)
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}

	return logChan, cleanup
}

// StopJob stops a running job
func StopJob(t *testing.T, jobID string) {
	cmd := exec.Command("nomad", "job", "stop", "-purge", jobID)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to stop job: %s", string(output))
}
