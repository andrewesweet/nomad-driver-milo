package features

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andrewesweet/nomad-driver-milo/e2e/shared"
	"github.com/hashicorp/nomad/api"
)

// RealBDDTestContext provides a test context for BDD testing with real Nomad integration
// It uses the shared.LiveNomadServer to run tests against a real Nomad instance
type RealBDDTestContext struct {
	t              *testing.T
	nomadServer    *shared.LiveNomadServer
	nomadClient    *api.Client
	tempFiles      map[string]string // filename -> path
	nomadJobFiles  map[string]string // filename -> content
	currentJobID   string
	jobAllocations map[string]*api.Allocation
	expectedOutput string
	origJavaHome   string // Backup of original JAVA_HOME
	testJavaHome   string // Test-specific JAVA_HOME
	lastExitCode   int
	lastOutput     string
	lastTaskEvents []string
	tempDir        string            // Temporary directory for test files
	artifactPaths  map[string]string // artifact name -> local path
	artifactServer *shared.ArtifactServer
}

// NewRealBDDTestContext creates a new BDD test context with real Nomad server
func NewRealBDDTestContext(t *testing.T) *RealBDDTestContext {
	return &RealBDDTestContext{
		t:              t,
		tempFiles:      make(map[string]string),
		nomadJobFiles:  make(map[string]string),
		jobAllocations: make(map[string]*api.Allocation),
		artifactPaths:  make(map[string]string),
	}
}

// setup initializes the test context by starting Nomad server and creating resources
func (ctx *RealBDDTestContext) setup() error {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "real-bdd-test-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	ctx.tempDir = tempDir

	// Backup original environment variables
	ctx.origJavaHome = os.Getenv("JAVA_HOME")

	// Start the artifact server using tests/fixtures directory
	fixturesDir := filepath.Join("..", "tests", "fixtures")
	ctx.artifactServer = shared.NewArtifactServer(fixturesDir)

	// Start real Nomad server
	ctx.nomadServer = shared.NewLiveNomadServer()
	if err := ctx.nomadServer.Start(); err != nil {
		return fmt.Errorf("failed to start Nomad server: %v", err)
	}

	// Get the API client from the server
	ctx.nomadClient = ctx.nomadServer.GetClient()
	if ctx.nomadClient == nil {
		return fmt.Errorf("failed to get Nomad client")
	}

	return nil
}

// cleanup cleans up all resources including temporary files and stops servers
func (ctx *RealBDDTestContext) cleanup() {
	// Restore original environment variables
	if ctx.origJavaHome != "" {
		os.Setenv("JAVA_HOME", ctx.origJavaHome)
	} else {
		os.Unsetenv("JAVA_HOME")
	}

	// Stop any running jobs
	if ctx.nomadClient != nil && ctx.currentJobID != "" {
		ctx.nomadClient.Jobs().Deregister(ctx.currentJobID, false, nil)
	}

	// Stop servers
	if ctx.artifactServer != nil {
		ctx.artifactServer.Close()
	}
	if ctx.nomadServer != nil {
		ctx.nomadServer.Stop()
	}

	// Remove temporary files
	for _, path := range ctx.tempFiles {
		os.Remove(path)
	}

	// Remove temporary directory
	if ctx.tempDir != "" {
		os.RemoveAll(ctx.tempDir)
	}
}

// setTestJavaPath sets a controlled Java path for testing
func (ctx *RealBDDTestContext) setTestJavaPath(javaPath string) {
	ctx.testJavaHome = javaPath
	os.Setenv("JAVA_HOME", javaPath)
}

// removeJavaPath removes Java from the environment to simulate no Java installation
func (ctx *RealBDDTestContext) removeJavaPath() {
	ctx.testJavaHome = ""
	os.Unsetenv("JAVA_HOME")
}

// createTempFile creates a temporary file with the given content
func (ctx *RealBDDTestContext) createTempFile(filename, content string) (string, error) {
	path := filepath.Join(ctx.tempDir, filename)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("failed to create temp file %s: %v", filename, err)
	}
	ctx.tempFiles[filename] = path
	return path, nil
}

// createJarFile creates a test JAR file
func (ctx *RealBDDTestContext) createJarFile(filename string) (string, error) {
	// For real tests, we should use actual JAR files
	// Copy the test hello-world.jar to our temp directory
	testJarPath := filepath.Join("..", "tests", "fixtures", "hello-world.jar")
	testJarContent, err := os.ReadFile(testJarPath)
	if err != nil {
		// If test JAR doesn't exist, create a minimal one
		jarContent := "PK\x03\x04\x14\x00\x00\x00\x08\x00"
		path, err := ctx.createTempFile(filename, jarContent)
		if err != nil {
			return "", err
		}
		// Track the artifact path for later mapping
		ctx.artifactPaths[filename] = path
		return path, nil
	}
	path, err := ctx.createTempFile(filename, string(testJarContent))
	if err != nil {
		return "", err
	}
	// Track the artifact path for later mapping
	ctx.artifactPaths[filename] = path
	return path, nil
}

// createJobFile creates a Nomad job file with the given content
func (ctx *RealBDDTestContext) createJobFile(filename, content string) error {
	_, err := ctx.createTempFile(filename, content)
	if err != nil {
		return err
	}
	ctx.nomadJobFiles[filename] = content
	return nil
}

// submitJob submits a job to the real Nomad server
func (ctx *RealBDDTestContext) submitJob(jobID, jarPath string) error {
	ctx.currentJobID = jobID

	// If we have a JAR file, copy it to artifact server directory and serve it
	var artifactURL string
	if jarPath != "" && ctx.artifactServer != nil {
		jarName := filepath.Base(jarPath)
		// Copy JAR to artifact server directory
		destPath := filepath.Join(ctx.artifactServer.BaseDir(), jarName)
		srcData, err := os.ReadFile(jarPath)
		if err != nil {
			return fmt.Errorf("failed to read JAR file: %v", err)
		}
		if err := os.WriteFile(destPath, srcData, 0644); err != nil {
			return fmt.Errorf("failed to copy JAR to artifact dir: %v", err)
		}
		artifactURL = ctx.artifactServer.URL()
	}

	// Submit job based on the type
	if artifactURL != "" {
		jarName := filepath.Base(jarPath)
		return ctx.nomadServer.SubmitJarJobFromHTTP(jobID, jarName, artifactURL)
	}

	// For jobs without artifacts, create a minimal job
	job := &api.Job{
		ID:   &jobID,
		Name: &jobID,
		Type: stringToPtr("batch"),
		TaskGroups: []*api.TaskGroup{
			{
				Name: stringToPtr("app"),
				Tasks: []*api.Task{
					{
						Name:   "java-app",
						Driver: "milo",
						Config: map[string]interface{}{
							"dummy": "",
						},
					},
				},
			},
		},
	}

	_, _, err := ctx.nomadClient.Jobs().Register(job, nil)
	return err
}

// waitForJobCompletion waits for a real job to complete
func (ctx *RealBDDTestContext) waitForJobCompletion(jobID string, timeout time.Duration) error {
	err := ctx.nomadServer.WaitForJobCompletion(jobID, timeout)
	if err != nil {
		// Job didn't complete - get and display server logs for debugging
		serverLogs := ctx.getNomadServerLogs()
		if serverLogs != "" {
			ctx.t.Logf("Nomad server logs:\n%s", serverLogs)
		}
		
		// Also get job status and task events for additional context
		if status, statusErr := ctx.getJobStatus(jobID); statusErr == nil {
			ctx.t.Logf("Job status: %+v", status)
		}
		
		if events, eventsErr := ctx.getTaskEvents(jobID, "java-app"); eventsErr == nil {
			ctx.t.Logf("Task events: %v", events)
		}
	}
	return err
}

// getJobStatus retrieves the status of a real job
func (ctx *RealBDDTestContext) getJobStatus(jobID string) (*api.JobSummary, error) {
	return ctx.nomadServer.GetJobStatus(jobID)
}

// getJobLogs retrieves the logs for a specific task in a job
func (ctx *RealBDDTestContext) getJobLogs(jobID, taskName string) (string, error) {
	return ctx.nomadServer.GetJobLogs(jobID, taskName)
}

// getTaskExitCode retrieves the exit code for a specific task
func (ctx *RealBDDTestContext) getTaskExitCode(jobID, taskName string) (int, error) {
	return ctx.nomadServer.GetTaskExitCode(jobID, taskName)
}

// getTaskEvents retrieves the events for a specific task
func (ctx *RealBDDTestContext) getTaskEvents(jobID, taskName string) ([]string, error) {
	// Get allocations for the job
	allocs, _, err := ctx.nomadClient.Jobs().Allocations(jobID, false, nil)
	if err != nil {
		return nil, err
	}

	if len(allocs) == 0 {
		return nil, fmt.Errorf("no allocations found for job %s", jobID)
	}

	// Get task state from first allocation
	alloc, _, err := ctx.nomadClient.Allocations().Info(allocs[0].ID, nil)
	if err != nil {
		return nil, err
	}

	var events []string
	if taskState, exists := alloc.TaskStates[taskName]; exists {
		for _, event := range taskState.Events {
			events = append(events, fmt.Sprintf("%s: %s", event.Type, event.DisplayMessage))
		}
	}

	ctx.lastTaskEvents = events
	return events, nil
}

// isJobSuccessful checks if a job completed successfully
func (ctx *RealBDDTestContext) isJobSuccessful(jobID string) (bool, error) {
	status, err := ctx.getJobStatus(jobID)
	if err != nil {
		return false, err
	}

	for _, taskGroup := range status.Summary {
		if taskGroup.Failed > 0 || taskGroup.Lost > 0 {
			return false, nil
		}
		if taskGroup.Complete == 0 {
			return false, nil
		}
	}

	return true, nil
}

// executeNomadCommand executes a real Nomad command
func (ctx *RealBDDTestContext) executeNomadCommand(command string) (string, error) {
	if strings.Contains(command, "nomad job run") {
		// Extract job file from command
		parts := strings.Fields(command)
		if len(parts) < 4 {
			return "", fmt.Errorf("invalid job run command")
		}
		jobFile := parts[3]

		// Get job content and extract job ID
		content, exists := ctx.nomadJobFiles[jobFile]
		if !exists {
			return "", fmt.Errorf("job file %s not found", jobFile)
		}

		// Parse and submit the job using the real API
		jobID := ctx.extractJobIDFromContent(content)
		if jobID == "" {
			return "", fmt.Errorf("could not extract job ID from %s", jobFile)
		}

		// Check if the job has artifact sources with file:// URLs
		if strings.Contains(content, "source = \"file://") && ctx.artifactServer != nil {
			// Process the content to replace file:// URLs with HTTP URLs
			modifiedContent := content
			
			// Find all file:// URLs in the content
			lines := strings.Split(content, "\n")
			for i, line := range lines {
				if strings.Contains(line, "source = \"file://") {
					// Extract the file path from the file:// URL
					start := strings.Index(line, "file://") + 7
					end := strings.Index(line[start:], "\"")
					if end > 0 {
						filePath := line[start : start+end]
						fileName := filepath.Base(filePath)
						
						// Copy the file to artifact server directory
						destPath := filepath.Join(ctx.artifactServer.BaseDir(), fileName)
						
						// Check if the file exists at the specified path
						if _, err := os.Stat(filePath); err == nil {
							// File exists, copy it
							srcData, err := os.ReadFile(filePath)
							if err == nil {
								os.WriteFile(destPath, srcData, 0644)
							}
						} else {
							// File doesn't exist at the path, check artifactPaths
							for name, path := range ctx.artifactPaths {
								if name == fileName || path == filePath {
									srcData, err := os.ReadFile(path)
									if err == nil {
										os.WriteFile(destPath, srcData, 0644)
									}
									break
								}
							}
						}
						
						// Replace the file:// URL with HTTP URL
						httpURL := fmt.Sprintf("%s/%s", ctx.artifactServer.URL(), fileName)
						lines[i] = strings.Replace(line, fmt.Sprintf("file://%s", filePath), httpURL, 1)
					}
				}
			}
			modifiedContent = strings.Join(lines, "\n")
			content = modifiedContent
		}

		// Submit the job with the modified content
		// Since we can't easily parse HCL to create an api.Job struct,
		// we'll extract the JAR path and use submitJob
		var jarPath string
		if strings.Contains(content, ctx.artifactServer.URL()) {
			// Extract the JAR name from the HTTP URL
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.Contains(line, "source = \"") && strings.Contains(line, ctx.artifactServer.URL()) {
					// Extract the JAR file name
					url := line[strings.Index(line, "\"")+1 : strings.LastIndex(line, "\"")]
					jarPath = filepath.Join(ctx.artifactServer.BaseDir(), filepath.Base(url))
					break
				}
			}
		}

		err := ctx.submitJob(jobID, jarPath)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Job %s submitted successfully", jobID), nil
	}

	if strings.Contains(command, "nomad logs") {
		// Extract job ID from command
		parts := strings.Fields(command)
		if len(parts) < 3 {
			return "", fmt.Errorf("invalid logs command")
		}
		jobID := parts[2]

		// Get real logs
		logs, err := ctx.getJobLogs(jobID, "java-app")
		if err != nil {
			return "", err
		}
		return logs, nil
	}

	return "", fmt.Errorf("unsupported command: %s", command)
}

// extractJobIDFromContent extracts the job ID from a Nomad job file content
func (ctx *RealBDDTestContext) extractJobIDFromContent(content string) string {
	// Simple extraction (could be more robust)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, "job \"") {
			start := strings.Index(line, "job \"") + 5
			end := strings.Index(line[start:], "\"")
			if end > 0 {
				return line[start : start+end]
			}
		}
	}
	return ""
}

// getNomadServerLogs retrieves the Nomad server logs
func (ctx *RealBDDTestContext) getNomadServerLogs() string {
	if ctx.nomadServer == nil {
		return ""
	}
	
	dataDir := ctx.nomadServer.GetDataDir()
	if dataDir == "" {
		return ""
	}
	
	logPath := filepath.Join(dataDir, "nomad.log")
	logContent, err := os.ReadFile(logPath)
	if err != nil {
		return fmt.Sprintf("Failed to read server logs: %v", err)
	}
	
	// Return last 100 lines or full content if less
	lines := strings.Split(string(logContent), "\n")
	if len(lines) > 100 {
		return "... (truncated)\n" + strings.Join(lines[len(lines)-100:], "\n")
	}
	return string(logContent)
}

// Helper function to convert string to pointer
func stringToPtr(s string) *string {
	return &s
}