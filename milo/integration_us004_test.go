package milo

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/base"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/stretchr/testify/require"
)

// Integration test for User Story 004: Java Runtime Detection and Exit Code Propagation
func TestUserStory004_JavaDetectionIntegration(t *testing.T) {
	logger := hclog.NewNullLogger()
	d := NewPlugin(logger).(*MiloDriverPlugin)
	d.config = &Config{}

	// Setup minimal nomadConfig to avoid nil pointer
	d.nomadConfig = &base.ClientDriverConfig{
		ClientMaxPort: 10000,
		ClientMinPort: 9000,
	}

	// Test Case 1: Missing Java Runtime
	t.Run("missing java runtime", func(t *testing.T) {
		// Use non-existent paths to simulate no Java
		nonExistentPaths := []string{"/nonexistent/java/path"}

		// Create temporary directory for test
		allocDir := t.TempDir()
		taskName := "test-no-java"
		taskDir := filepath.Join(allocDir, taskName)
		localDir := filepath.Join(taskDir, "local")
		require.NoError(t, os.MkdirAll(localDir, 0755))

		// Create a test JAR file
		jarPath := filepath.Join(localDir, "test.jar")
		createValidTestJAR(t, jarPath)

		// Prepare task config
		taskCfg := &drivers.TaskConfig{
			ID:       "test-no-java",
			Name:     taskName,
			AllocDir: allocDir,
		}

		// Encode driver config
		driverConfig := map[string]interface{}{
			"dummy": "",
		}
		require.NoError(t, taskCfg.EncodeConcreteDriverConfig(&driverConfig))

		// Test Java detection directly with non-existent paths
		_, err := DetectJavaRuntime(nonExistentPaths)
		require.Error(t, err)
		require.Equal(t, "No Java runtime found on host", err.Error())
		
		// Check that it's a MissingJavaError and verify detailed message
		mjErr, ok := err.(*MissingJavaError)
		require.True(t, ok, "expected error to be a *MissingJavaError")
		
		detailedMsg := mjErr.Detailed()
		require.Contains(t, detailedMsg, "Error: No Java runtime found on host")
		require.Contains(t, detailedMsg, "Searched locations:")
		require.Contains(t, detailedMsg, "/nonexistent/java/path")
	})

	// Test Case 2: Java Runtime Detection Success
	t.Run("java runtime detection success", func(t *testing.T) {
		// Create a mock Java installation
		tempDir := t.TempDir()
		javaPath := filepath.Join(tempDir, "bin", "java")
		require.NoError(t, os.MkdirAll(filepath.Dir(javaPath), 0755))
		require.NoError(t, os.WriteFile(javaPath, []byte("#!/bin/bash\necho 'mock java'"), 0600))
		require.NoError(t, os.Chmod(javaPath, 0700))

		// Test detection
		detectedJava, err := DetectJavaRuntime([]string{tempDir})
		require.NoError(t, err)
		require.Equal(t, tempDir, detectedJava)
	})

	// Test Case 3: JAVA_HOME Environment Variable
	t.Run("JAVA_HOME environment detection", func(t *testing.T) {
		// Create a mock Java installation
		tempDir := t.TempDir()
		javaPath := filepath.Join(tempDir, "bin", "java")
		require.NoError(t, os.MkdirAll(filepath.Dir(javaPath), 0755))
		require.NoError(t, os.WriteFile(javaPath, []byte("#!/bin/bash\necho 'mock java'"), 0600))
		require.NoError(t, os.Chmod(javaPath, 0700))

		// Set JAVA_HOME
		originalJavaHome := os.Getenv("JAVA_HOME")
		os.Setenv("JAVA_HOME", tempDir)
		defer os.Setenv("JAVA_HOME", originalJavaHome)

		// Test detection - should use JAVA_HOME first
		detectedJava, err := DetectJavaRuntime([]string{"/some/other/path"})
		require.NoError(t, err)
		require.Equal(t, tempDir, detectedJava)
	})
}

// Integration test for exit code propagation with real process execution
func TestUserStory004_ExitCodePropagationIntegration(t *testing.T) {
	t.Skip("Covered by TestExitCodePropagation tests")
	// Skip if we don't have proper environment
	if _, err := os.Stat("/bin/sh"); err != nil {
		t.Skip("Requires /bin/sh for integration test")
	}

	testCases := []struct {
		name          string
		scriptContent string
		expectedCode  int
		expectError   bool
	}{
		{
			name:          "successful execution",
			scriptContent: "#!/bin/sh\necho 'Success'\nexit 0",
			expectedCode:  0,
			expectError:   false,
		},
		{
			name:          "general error",
			scriptContent: "#!/bin/sh\necho 'Error occurred' >&2\nexit 1",
			expectedCode:  1,
			expectError:   false,
		},
		{
			name:          "specific exit code 42",
			scriptContent: "#!/bin/sh\necho 'Application error' >&2\nexit 42",
			expectedCode:  42,
			expectError:   false,
		},
		{
			name:          "command not found",
			scriptContent: "#!/bin/sh\nexit 127",
			expectedCode:  127,
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test script
			tempDir := t.TempDir()
			scriptPath := filepath.Join(tempDir, "test.sh")
			require.NoError(t, os.WriteFile(scriptPath, []byte(tc.scriptContent), 0600))
			require.NoError(t, os.Chmod(scriptPath, 0700))

			// Execute via taskHandle
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			cmd := exec.Command(scriptPath)
			err := cmd.Start()
			require.NoError(t, err)

			h := &taskHandle{
				logger:     hclog.NewNullLogger(),
				cmd:        cmd,
				pid:        cmd.Process.Pid,
				taskConfig: &drivers.TaskConfig{ID: fmt.Sprintf("test-%s", tc.name)},
				procState:  drivers.TaskStateRunning,
				startedAt:  time.Now(),
				ctx:        ctx,
				cancelFunc: cancel,
				waitCh:     make(chan struct{}),
			}

			// Run and wait
			go h.run()

			select {
			case <-h.waitCh:
				// Process completed
			case <-time.After(10 * time.Second):
				t.Fatal("timeout waiting for process")
			}

			// Give a tiny bit more time for state to update
			time.Sleep(10 * time.Millisecond)

			// Verify exit code
			status := h.TaskStatus()
			require.Equal(t, drivers.TaskStateExited, status.State)
			require.NotNil(t, status.ExitResult)
			require.Equal(t, tc.expectedCode, status.ExitResult.ExitCode)

			if tc.expectError {
				require.Error(t, status.ExitResult.Err)
			} else {
				require.NoError(t, status.ExitResult.Err)
			}
		})
	}
}

// Test the enhanced error message format
func TestUserStory004_EnhancedErrorMessage(t *testing.T) {
	// Test with various scenarios
	testCases := []struct {
		name          string
		searchPaths   []string
		javaHomeSet   bool
		javaHomeValue string
		expectedInMsg []string
	}{
		{
			name:        "no JAVA_HOME set",
			searchPaths: []string{"/usr/lib/jvm/java-17", "/opt/java"},
			javaHomeSet: false,
			expectedInMsg: []string{
				"JAVA_HOME environment variable (not set)",
				"/usr/lib/jvm/java-17 (not found)",
				"/opt/java (not found)",
				"sudo apt install openjdk-17-jdk",
			},
		},
		{
			name:          "invalid JAVA_HOME set",
			searchPaths:   []string{"/usr/lib/jvm/java-11"},
			javaHomeSet:   true,
			javaHomeValue: "/invalid/java/home",
			expectedInMsg: []string{
				"JAVA_HOME: /invalid/java/home (invalid or not found)",
				"/usr/lib/jvm/java-11 (not found)",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set/unset JAVA_HOME
			originalJavaHome := os.Getenv("JAVA_HOME")
			if tc.javaHomeSet {
				os.Setenv("JAVA_HOME", tc.javaHomeValue)
			} else {
				os.Unsetenv("JAVA_HOME")
			}
			defer os.Setenv("JAVA_HOME", originalJavaHome)

			// Generate error
			err := FormatMissingJavaError(tc.searchPaths)
			require.Error(t, err)

			// Verify simple error message
			require.Equal(t, "No Java runtime found on host", err.Error())
			
			// Check detailed message
			mjErr, ok := err.(*MissingJavaError)
			require.True(t, ok, "expected error to be a *MissingJavaError")
			
			detailedMsg := mjErr.Detailed()
			for _, expected := range tc.expectedInMsg {
				require.Contains(t, detailedMsg, expected, "Detailed message should contain: %s", expected)
			}
		})
	}
}
