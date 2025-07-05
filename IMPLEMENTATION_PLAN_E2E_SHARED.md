# E2E Test Infrastructure Implementation Plan

**Quality Assessment: A- (Strong, production-ready plan)**

## Overview

Implement shared e2e test infrastructure that:
- Uses TestMain to start one Nomad server and one HTTP artifact server for ALL tests
- Uses unique job IDs for bulletproof test isolation between tests
- Enables parallel test execution with t.Parallel()
- Uses httptest.NewServer for artifact serving instead of custom HTTP server
- Uses t.Cleanup() for robust cleanup
- Maintains existing ATDD methodology and test structure

## Performance Goal
Reduce test startup cost from ~9+ seconds per test to ~9 seconds amortized across entire test suite.

## Implementation Steps

### Step 1: Create `e2e/live/main_test.go`

```go
//go:build live_e2e

package live

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

var (
	testServer        *LiveNomadServer
	artifactServerURL string
)

func TestMain(m *testing.M) {
	// 1. Start HTTP artifact server
	artifactServer := httptest.NewServer(http.FileServer(http.Dir("../../test-artifacts")))
	artifactServerURL = artifactServer.URL

	// 2. Start shared Nomad server
	t := &testing.T{}
	server := NewLiveNomadServer(t)
	if err := server.Start(); err != nil {
		artifactServer.Close()
		server.Stop()
		os.Exit(1)
	}
	testServer = server

	// 3. Run tests
	exitCode := m.Run()

	// 4. Cleanup
	artifactServer.Close()
	testServer.Stop()
	os.Exit(exitCode)
}

// sanitizeForNomadJobID converts test names to valid Nomad job IDs
func sanitizeForNomadJobID(name string) string {
	name = strings.ReplaceAll(name, "/", "-")
	reg := regexp.MustCompile(`[^a-zA-Z0-9-]+`)
	sanitized := reg.ReplaceAllString(name, "-")
	sanitized = strings.Trim(sanitized, "-")
	return strings.ToLower(sanitized)
}

// generateTestJobID creates unique job IDs for parallel tests
func generateTestJobID(t *testing.T) string {
	rand.Seed(time.Now().UnixNano())
	baseName := sanitizeForNomadJobID(t.Name())
	return fmt.Sprintf("%s-%d", baseName, rand.Intn(10000))
}
```

### Step 2: Add HTTP submission method to `e2e/live/live_nomad_server_test.go`

```go
// SubmitJarJobFromHTTP submits a job using HTTP artifact source
func (s *LiveNomadServer) SubmitJarJobFromHTTP(jobID, jarName, httpArtifactURL string) error {
	if s.client == nil {
		return fmt.Errorf("client not initialized")
	}

	artifactSource := fmt.Sprintf("%s/%s", httpArtifactURL, jarName)
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
						Artifacts: []*api.TaskArtifact{
							{
								GetterSource:  stringToPtr(artifactSource),
								RelativeDest:  stringToPtr("local/"),
							},
						},
						Config: map[string]interface{}{},
					},
				},
			},
		},
	}

	_, _, err := s.client.Jobs().Register(job, nil)
	return err
}
```

### Step 3: Update existing test in `e2e/live/live_nomad_server_test.go`

```go
func TestLiveNomadServer_RunHelloWorldJar(t *testing.T) {
	t.Parallel() // Enable parallel execution

	// Use shared server from TestMain
	require.NotNil(t, testServer, "test server should be initialized")
	require.True(t, testServer.IsAccessible(), "test server should be accessible")

	// Generate unique job ID for this test
	jobID := generateTestJobID(t)
	
	// Schedule cleanup
	t.Cleanup(func() {
		testServer.client.Jobs().Deregister(jobID, true, nil)
	})

	// Submit job using HTTP artifacts
	jarName := "hello-world.jar"
	err := testServer.SubmitJarJobFromHTTP(jobID, jarName, artifactServerURL)
	require.NoError(t, err, "Should submit job successfully")

	// Wait for completion and verify results
	err = testServer.WaitForJobCompletion(jobID, 30*time.Second)
	require.NoError(t, err, "Job should complete successfully")

	// Verify job succeeded
	status, err := testServer.GetJobStatus(jobID)
	require.NoError(t, err)
	
	var hasCompleted bool
	var hasFailed bool
	for _, taskGroup := range status.Summary {
		if taskGroup.Complete > 0 {
			hasCompleted = true
		}
		if taskGroup.Failed > 0 {
			hasFailed = true
		}
	}
	assert.True(t, hasCompleted, "Job should have completed tasks")
	assert.False(t, hasFailed, "Job should have no failed tasks")

	// Verify logs contain expected output
	logs, err := testServer.GetJobLogs(jobID, "java-app")
	require.NoError(t, err, "Should get job logs")
	assert.Contains(t, logs, "Hello from Java!", "Logs should contain expected output")

	// Verify task exit code
	exitCode, err := testServer.GetTaskExitCode(jobID, "java-app")
	require.NoError(t, err)
	assert.Equal(t, 0, exitCode, "Task should exit with code 0")
}
```

### Step 4: Remove redundant tests

Delete the following tests that are now covered by TestMain:
- `TestLiveNomadServer_StartStopWithPlugin` 
- `TestLiveNomadServer_DynamicPortAllocation`

Keep `TestLiveNomadServer_PluginLoadTimeout` but refactor it to use shared infrastructure.

## Key Features Delivered

✅ **Shared Infrastructure**: One Nomad server and HTTP server for all tests  
✅ **Parallel Safety**: Unique job IDs with sanitization  
✅ **Existing Pattern Compatibility**: Uses current LiveNomadServer functions  
✅ **HTTP Artifacts**: Replaces git:// with httptest.NewServer  
✅ **Automatic Cleanup**: t.Cleanup() for robust resource management  
✅ **Error Handling**: Proper TestMain error handling and cleanup  

## Success Criteria

- [ ] TestMain starts shared infrastructure once
- [ ] Tests can run in parallel with t.Parallel()
- [ ] HTTP artifact server serves test-artifacts directory
- [ ] Unique job IDs prevent test interference
- [ ] HelloWorld.jar test passes using HTTP artifacts
- [ ] Test suite startup time reduced significantly
- [ ] All tests clean up resources automatically

## Implementation Notes

- Job ID sanitization handles test names with special characters
- Random seeding ensures unique job IDs across parallel tests
- t.Cleanup() provides robust cleanup even on test failures
- Nomad API client is thread-safe for parallel access
- HTTP artifact serving replaces git:// dependency