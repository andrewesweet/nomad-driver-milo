package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andrewesweet/nomad-driver-milo/milo"
	"github.com/stretchr/testify/require"
)

// Test Scenario 4: Missing Java Runtime
func TestMissingJavaRuntime(t *testing.T) {
	// This test verifies the Java detection logic without actually running a Nomad job
	// since we can't easily simulate a missing Java runtime in the integration environment

	// Given a host with no Java runtime installed
	// (simulated by providing non-existent paths)
	nonExistentPaths := []string{
		"/nonexistent/java/path1",
		"/nonexistent/java/path2",
	}

	// Save and clear JAVA_HOME to simulate no environment variable
	originalJavaHome := os.Getenv("JAVA_HOME")
	os.Unsetenv("JAVA_HOME")
	defer os.Setenv("JAVA_HOME", originalJavaHome)

	// When the driver attempts to detect Java
	_, err := milo.DetectJavaRuntime(nonExistentPaths)

	// Then it should return an error containing:
	//   "Error: No Java runtime found on host. Please install Java to use Milo driver."
	require.Error(t, err)
	require.Contains(t, err.Error(), "Error: No Java runtime found on host")
	require.Contains(t, err.Error(), "Please install Java to use Milo driver")
	require.Contains(t, err.Error(), "JAVA_HOME environment variable (not set)")
	require.Contains(t, err.Error(), "/nonexistent/java/path1 (not found)")
	require.Contains(t, err.Error(), "/nonexistent/java/path2 (not found)")
	require.Contains(t, err.Error(), "sudo apt install openjdk-17-jdk")
}

// Additional test for successful Java detection
func TestJavaRuntimeDetection(t *testing.T) {
	// Create a mock Java installation for testing
	tempDir := t.TempDir()
	javaDir := filepath.Join(tempDir, "java")
	javaBin := filepath.Join(javaDir, "bin")
	require.NoError(t, os.MkdirAll(javaBin, 0755))

	// Create mock java executable
	javaPath := filepath.Join(javaBin, "java")
	require.NoError(t, os.WriteFile(javaPath, []byte("#!/bin/bash\necho 'mock java'"), 0600))
	require.NoError(t, os.Chmod(javaPath, 0700))

	// Test detection with the mock installation
	detectedJava, err := milo.DetectJavaRuntime([]string{javaDir})
	require.NoError(t, err)
	require.Equal(t, javaDir, detectedJava)

	// Test JAVA_HOME detection
	originalJavaHome := os.Getenv("JAVA_HOME")
	os.Setenv("JAVA_HOME", javaDir)
	defer os.Setenv("JAVA_HOME", originalJavaHome)

	detectedJava, err = milo.DetectJavaRuntime([]string{"/some/other/path"})
	require.NoError(t, err)
	require.Equal(t, javaDir, detectedJava)
}
