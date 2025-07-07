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

// BDD Scenario from implementation plan:
// Real-time streaming of logs from a long-running Java application
func TestBDD_RealTimeLogStreaming(t *testing.T) {
	require := require.New(t)

	// Given a host with Java runtime installed
	// (verified in TestMain)
	
	// And a test JAR file that prints periodically
	createStreamingTestJar(t)
	
	// And a Nomad job file contains the streaming test configuration
	jobID := generateTestJobID(t)
	jobSpec := fmt.Sprintf(`
job "%s" {
  type = "service"
  datacenters = ["dc1"]
  
  group "app" {
    task "java-app" {
      driver = "milo"
      
      config {
        dummy = ""
      }
      
      artifact {
        source = "%s/long-running.jar"
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
	jobFile := filepath.Join(t.TempDir(), "streaming-test.nomad")
	require.NoError(os.WriteFile(jobFile, []byte(jobSpec), 0644))

	// When the user executes: `nomad job run streaming-test.nomad`
	cmd := exec.Command("nomad", "job", "run", jobFile)
	output, err := cmd.CombinedOutput()
	require.NoError(err, "Failed to submit job: %s", output)

	// And waits 5 seconds
	time.Sleep(5 * time.Second)

	// And executes: `nomad logs -f streaming-test java-app`
	logsCmd := exec.Command("nomad", "logs", "-f", jobID, "java-app")
	stdout, err := logsCmd.StdoutPipe()
	require.NoError(err)
	
	require.NoError(logsCmd.Start())

	// Then the log output should show expected lines
	expectedLines := []string{
		"Starting application...",
		"Processing...",
		"Processing...",
	}
	
	foundLines := make(map[string]int)
	lastProcessingTime := time.Now()
	
	// Read logs for up to 10 seconds
	timeout := time.After(10 * time.Second)
	lineChan := make(chan string)
	
	go func() {
		buf := make([]byte, 1024)
		accumulated := ""
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				return
			}
			accumulated += string(buf[:n])
			lines := strings.Split(accumulated, "\n")
			accumulated = lines[len(lines)-1]
			for i := 0; i < len(lines)-1; i++ {
				if lines[i] != "" {
					lineChan <- lines[i]
				}
			}
		}
	}()

	processingCount := 0
	
collectLogs:
	for {
		select {
		case line := <-lineChan:
			t.Logf("Received line: %s", line)
			
			// Count occurrences
			for _, expected := range expectedLines {
				if strings.Contains(line, expected) {
					foundLines[expected]++
				}
			}
			
			// Track timing of "Processing..." messages
			if strings.Contains(line, "Processing...") {
				processingCount++
				if processingCount > 1 {
					elapsed := time.Since(lastProcessingTime)
					// Verify ~2 second interval (allow 1.5-2.5 seconds)
					require.True(elapsed >= 1500*time.Millisecond && elapsed <= 2500*time.Millisecond,
						"Processing interval should be ~2 seconds, got %v", elapsed)
				}
				lastProcessingTime = time.Now()
			}
			
			// Once we have enough data, verify and exit
			if processingCount >= 3 {
				break collectLogs
			}
			
		case <-timeout:
			break collectLogs
		}
	}
	
	// Kill the logs command
	logsCmd.Process.Kill()
	
	// Verify we saw the expected output
	require.GreaterOrEqual(foundLines["Starting application..."], 1, "Should see 'Starting application...'")
	require.GreaterOrEqual(foundLines["Processing..."], 2, "Should see at least 2 'Processing...' lines")
	
	// And new "Processing..." lines should appear every 2 seconds
	// (verified above during collection)
	
	// And the task status should show "running"
	statusCmd := exec.Command("nomad", "job", "status", jobID)
	statusOutput, err := statusCmd.Output()
	require.NoError(err)
	require.Contains(string(statusOutput), "running", "Task should be running")
	
	// When the user executes: `nomad job stop streaming-test`
	stopCmd := exec.Command("nomad", "job", "stop", jobID)
	stopOutput, err := stopCmd.CombinedOutput()
	require.NoError(err, "Failed to stop job: %s", stopOutput)
	
	// Then the task should terminate within 5 seconds
	terminated := false
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		statusCmd := exec.Command("nomad", "job", "status", jobID)
		statusOutput, err := statusCmd.Output()
		if err != nil || strings.Contains(string(statusOutput), "dead") {
			terminated = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	require.True(terminated, "Task should terminate within 5 seconds")
	
	// Cleanup
	exec.Command("nomad", "job", "stop", "-purge", jobID).Run()
}

// Create the long-running test JAR as specified in the BDD scenario
func createStreamingTestJar(t *testing.T) {
	javaCode := `
public class Main {
    public static void main(String[] args) throws InterruptedException {
        System.out.println("Starting application...");
        System.out.flush();
        
        // Run until terminated
        while (true) {
            Thread.sleep(2000);
            System.out.println("Processing...");
            System.out.flush();
        }
    }
}`

	tmpDir := t.TempDir()
	
	// Write Java source
	javaFile := filepath.Join(tmpDir, "Main.java")
	require.NoError(t, os.WriteFile(javaFile, []byte(javaCode), 0644))

	// Compile
	cmd := exec.Command("javac", javaFile)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to compile: %s", output)

	// Create manifest
	manifestContent := "Main-Class: Main\n"
	manifestFile := filepath.Join(tmpDir, "MANIFEST.MF")
	require.NoError(t, os.WriteFile(manifestFile, []byte(manifestContent), 0644))

	// Create JAR
	jarPath := filepath.Join("../../test-artifacts", "long-running.jar")
	cmd = exec.Command("jar", "cfm", jarPath, manifestFile, "-C", tmpDir, "Main.class")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Failed to create JAR: %s", output)
}