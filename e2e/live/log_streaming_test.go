//go:build live_e2e

package live

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Test 13: Static logs display correctly via `nomad logs`
func TestLogStreaming_StaticLogs(t *testing.T) {
	require := require.New(t)
	jobID := generateTestJobID(t)

	// Create a simple JAR that prints and exits
	createTestJar(t, "static-log-test.jar", `
public class Main {
    public static void main(String[] args) {
        System.out.println("Line 1: Starting application");
        System.out.println("Line 2: Processing data");
        System.err.println("Line 3: Warning message");
        System.out.println("Line 4: Completed successfully");
    }
}`)

	// Create job specification
	jobSpec := fmt.Sprintf(`
job "%s" {
  type = "batch"
  datacenters = ["dc1"]
  
  group "test" {
    task "log-test" {
      driver = "milo"
      
      config {
        dummy = ""
      }
      
      artifact {
        source = "%s/static-log-test.jar"
        destination = "local/"
      }
      
      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}`, jobID, artifactServerURL)

	// Write job file
	jobFile := filepath.Join(t.TempDir(), "job.hcl")
	require.NoError(os.WriteFile(jobFile, []byte(jobSpec), 0644))

	// Submit job
	cmd := exec.Command("nomad", "job", "run", jobFile)
	output, err := cmd.CombinedOutput()
	require.NoError(err, "Failed to submit job: %s", output)

	// Wait for job to complete
	waitForJobComplete(t, jobID, 30*time.Second)

	// Get logs
	cmd = exec.Command("nomad", "logs", jobID, "log-test")
	stdout, err := cmd.Output()
	require.NoError(err)

	// Get stderr logs
	cmd = exec.Command("nomad", "logs", "-stderr", jobID, "log-test")
	stderr, err := cmd.Output()
	require.NoError(err)

	// Verify stdout contains expected lines
	stdoutStr := string(stdout)
	require.Contains(stdoutStr, "Line 1: Starting application")
	require.Contains(stdoutStr, "Line 2: Processing data")
	require.Contains(stdoutStr, "Line 4: Completed successfully")
	require.NotContains(stdoutStr, "Line 3: Warning message") // This should be in stderr

	// Verify stderr contains warning
	stderrStr := string(stderr)
	require.Contains(stderrStr, "Line 3: Warning message")

	// Cleanup
	exec.Command("nomad", "job", "stop", "-purge", jobID).Run()
}

// Test 14: Real-time streaming works with `nomad logs -f`
func TestLogStreaming_RealTimeStreaming(t *testing.T) {
	require := require.New(t)
	jobID := generateTestJobID(t)

	// Create a long-running JAR that prints periodically
	createTestJar(t, "streaming-test.jar", `
public class Main {
    public static void main(String[] args) throws InterruptedException {
        System.out.println("Starting application...");
        System.out.flush();
        
        for (int i = 1; i <= 5; i++) {
            Thread.sleep(1000);
            System.out.println("Processing... " + i);
            System.out.flush();
        }
        
        System.out.println("Completed!");
    }
}`)

	// Create job specification
	jobSpec := fmt.Sprintf(`
job "%s" {
  type = "batch"
  datacenters = ["dc1"]
  
  group "test" {
    task "stream-test" {
      driver = "milo"
      
      config {
        dummy = ""
      }
      
      artifact {
        source = "%s/streaming-test.jar"
        destination = "local/"
      }
      
      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}`, jobID, artifactServerURL)

	// Write job file
	jobFile := filepath.Join(t.TempDir(), "job.hcl")
	require.NoError(os.WriteFile(jobFile, []byte(jobSpec), 0644))

	// Submit job
	cmd := exec.Command("nomad", "job", "run", jobFile)
	output, err := cmd.CombinedOutput()
	require.NoError(err, "Failed to submit job: %s", output)

	// Wait for job to start
	waitForJobRunning(t, jobID, 10*time.Second)

	// Start streaming logs
	cmd = exec.Command("nomad", "logs", "-f", jobID, "stream-test")
	stdout, err := cmd.StdoutPipe()
	require.NoError(err)
	
	require.NoError(cmd.Start())

	// Read streaming output
	streamedLines := make([]string, 0)
	lineChan := make(chan string)
	doneChan := make(chan bool)

	go func() {
		buf := make([]byte, 1024)
		accumulated := ""
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				break
			}
			accumulated += string(buf[:n])
			// Split by newlines
			lines := strings.Split(accumulated, "\n")
			// Keep the last incomplete line
			accumulated = lines[len(lines)-1]
			// Send complete lines
			for i := 0; i < len(lines)-1; i++ {
				if lines[i] != "" {
					lineChan <- lines[i]
				}
			}
		}
		close(doneChan)
	}()

	// Collect lines with timeout
	timeout := time.After(15 * time.Second)
	for {
		select {
		case line := <-lineChan:
			streamedLines = append(streamedLines, line)
			if strings.Contains(line, "Completed!") {
				cmd.Process.Kill()
				goto verify
			}
		case <-doneChan:
			goto verify
		case <-timeout:
			cmd.Process.Kill()
			goto verify
		}
	}

verify:
	// Verify we got real-time streaming
	require.GreaterOrEqual(len(streamedLines), 3, "Should have received multiple streamed lines")
	
	// Verify order
	foundStarting := false
	foundProcessing := false
	for _, line := range streamedLines {
		if strings.Contains(line, "Starting application") {
			foundStarting = true
		}
		if strings.Contains(line, "Processing...") && foundStarting {
			foundProcessing = true
		}
	}
	require.True(foundStarting, "Should have seen 'Starting application'")
	require.True(foundProcessing, "Should have seen 'Processing...' after start")

	// Cleanup
	exec.Command("nomad", "job", "stop", "-purge", jobID).Run()
}

// Test 16: Logs continue streaming after task completes
func TestLogStreaming_AfterCompletion(t *testing.T) {
	require := require.New(t)
	jobID := generateTestJobID(t)

	// Create a JAR that prints and exits quickly
	createTestJar(t, "completion-test.jar", `
public class Main {
    public static void main(String[] args) {
        System.out.println("Task starting");
        System.out.println("Task completing");
        System.out.println("Final message");
    }
}`)

	// Create job specification
	jobSpec := fmt.Sprintf(`
job "%s" {
  type = "batch"
  datacenters = ["dc1"]
  
  group "test" {
    task "completion-test" {
      driver = "milo"
      
      config {
        dummy = ""
      }
      
      artifact {
        source = "%s/completion-test.jar"
        destination = "local/"
      }
      
      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}`, jobID, artifactServerURL)

	// Write job file
	jobFile := filepath.Join(t.TempDir(), "job.hcl")
	require.NoError(os.WriteFile(jobFile, []byte(jobSpec), 0644))

	// Submit job
	cmd := exec.Command("nomad", "job", "run", jobFile)
	output, err := cmd.CombinedOutput()
	require.NoError(err, "Failed to submit job: %s", output)

	// Wait for job to complete
	waitForJobComplete(t, jobID, 30*time.Second)

	// Verify we can still get logs after completion
	cmd = exec.Command("nomad", "logs", jobID, "completion-test")
	stdout, err := cmd.Output()
	require.NoError(err, "Should be able to get logs after task completion")

	// Verify all messages are present
	stdoutStr := string(stdout)
	require.Contains(stdoutStr, "Task starting")
	require.Contains(stdoutStr, "Task completing")
	require.Contains(stdoutStr, "Final message")

	// Cleanup
	exec.Command("nomad", "job", "stop", "-purge", jobID).Run()
}

// Helper function to create test JAR files
func createTestJar(t *testing.T, jarName string, javaCode string) {
	tmpDir := t.TempDir()
	
	// Write Java source
	javaFile := filepath.Join(tmpDir, "Main.java")
	require.NoError(t, os.WriteFile(javaFile, []byte(javaCode), 0644))

	// Compile Java code
	cmd := exec.Command("javac", javaFile)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to compile Java: %s", output)

	// Create JAR with manifest
	manifestContent := "Main-Class: Main\n"
	manifestFile := filepath.Join(tmpDir, "MANIFEST.MF")
	require.NoError(t, os.WriteFile(manifestFile, []byte(manifestContent), 0644))

	// Create JAR
	jarPath := filepath.Join("../../test-artifacts", jarName)
	cmd = exec.Command("jar", "cfm", jarPath, manifestFile, "-C", tmpDir, "Main.class")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Failed to create JAR: %s", output)
}

// Helper to wait for job to reach running state
func waitForJobRunning(t *testing.T, jobID string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		cmd := exec.Command("nomad", "job", "status", jobID)
		output, err := cmd.Output()
		if err == nil && strings.Contains(string(output), "running") {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("Job %s did not reach running state within %v", jobID, timeout)
}

// Helper to wait for job completion
func waitForJobComplete(t *testing.T, jobID string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		cmd := exec.Command("nomad", "job", "status", jobID)
		output, err := cmd.Output()
		if err == nil && (strings.Contains(string(output), "complete") || strings.Contains(string(output), "dead")) {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("Job %s did not complete within %v", jobID, timeout)
}