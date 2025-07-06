package milo

import (
	"os"
	"path/filepath"
	"testing"
)

// === CYCLE 8: Scan Java installation paths ===
// Phase: RED
func TestScanJavaInstallationPaths_FindsJava(t *testing.T) {
	// Given a system with Java installed in a common location
	// We'll create a mock java executable for testing
	tempDir := t.TempDir()
	javaPath := filepath.Join(tempDir, "bin", "java")

	// Create the directory structure
	err := os.MkdirAll(filepath.Dir(javaPath), 0755)
	if err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}

	// Create a mock java executable
	err = os.WriteFile(javaPath, []byte("#!/bin/bash\necho 'java version 17'"), 0600)
	if err != nil {
		t.Fatalf("failed to create mock java executable: %v", err)
	}

	// When we scan for Java installations
	javaPaths := ScanJavaInstallationPaths([]string{tempDir})

	// Then we should find the Java installation
	if len(javaPaths) == 0 {
		t.Error("expected to find Java installation, but none found")
	}

	expected := tempDir
	if javaPaths[0] != expected {
		t.Errorf("expected Java path %q, got %q", expected, javaPaths[0])
	}
}

// === CYCLE 9: Check JAVA_HOME environment ===
// Phase: RED
func TestCheckJavaHomeEnvironment_ReturnsPath(t *testing.T) {
	// Given JAVA_HOME is set
	tempDir := t.TempDir()
	javaPath := filepath.Join(tempDir, "bin", "java")

	// Create the directory structure and java executable
	err := os.MkdirAll(filepath.Dir(javaPath), 0755)
	if err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}

	err = os.WriteFile(javaPath, []byte("#!/bin/bash\necho 'java version 17'"), 0600)
	if err != nil {
		t.Fatalf("failed to create mock java executable: %v", err)
	}

	// Set JAVA_HOME temporarily
	originalJavaHome := os.Getenv("JAVA_HOME")
	os.Setenv("JAVA_HOME", tempDir)
	defer os.Setenv("JAVA_HOME", originalJavaHome)

	// When we check JAVA_HOME
	javaHome, found := CheckJavaHomeEnvironment()

	// Then we should get the JAVA_HOME path
	if !found {
		t.Error("expected to find JAVA_HOME, but not found")
	}

	if javaHome != tempDir {
		t.Errorf("expected JAVA_HOME %q, got %q", tempDir, javaHome)
	}
}

// === CYCLE 10: Validate Java executable ===
// Phase: RED
func TestValidateJavaExecutable_ValidJava(t *testing.T) {
	// Given a valid Java installation directory
	tempDir := t.TempDir()
	javaPath := filepath.Join(tempDir, "bin", "java")

	// Create the directory structure and java executable
	err := os.MkdirAll(filepath.Dir(javaPath), 0755)
	if err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}

	err = os.WriteFile(javaPath, []byte("#!/bin/bash\necho 'java version 17'"), 0600)
	if err != nil {
		t.Fatalf("failed to create mock java executable: %v", err)
	}

	// When we validate the Java executable
	valid := ValidateJavaExecutable(tempDir)

	// Then it should be valid
	if !valid {
		t.Error("expected Java executable to be valid, but validation failed")
	}
}

// === CYCLE 11: Missing Java error generation ===
// Phase: RED
func TestFormatMissingJavaError_GeneratesMessage(t *testing.T) {
	// When we format a missing Java error
	err := FormatMissingJavaError()

	// Then we should get the expected error message
	expectedMsg := "Error: No Java runtime found on host. Please install Java to use Milo driver."
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}

// === CYCLE 12: Complete Java detection pipeline ===
// Phase: RED
func TestDetectJavaRuntime_FindsJava(t *testing.T) {
	// Given a system with Java installed
	tempDir := t.TempDir()
	javaPath := filepath.Join(tempDir, "bin", "java")

	// Create the directory structure and java executable
	err := os.MkdirAll(filepath.Dir(javaPath), 0755)
	if err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}

	err = os.WriteFile(javaPath, []byte("#!/bin/bash\necho 'java version 17'"), 0600)
	if err != nil {
		t.Fatalf("failed to create mock java executable: %v", err)
	}

	// When we detect Java runtime (using custom search paths for testing)
	javaHome, err := DetectJavaRuntime([]string{tempDir})

	// Then we should find the Java installation
	if err != nil {
		t.Errorf("expected to find Java, got error: %v", err)
	}

	if javaHome != tempDir {
		t.Errorf("expected Java home %q, got %q", tempDir, javaHome)
	}
}

// === CYCLE 13: Java detection failure ===
// Phase: RED
func TestDetectJavaRuntime_NoJavaFound(t *testing.T) {
	// Given a system with no Java installed
	tempDir := t.TempDir()

	// When we detect Java runtime in empty directories
	_, err := DetectJavaRuntime([]string{tempDir})

	// Then we should get a missing Java error
	if err == nil {
		t.Error("expected error when no Java found, got nil")
	}

	expectedMsg := "Error: No Java runtime found on host. Please install Java to use Milo driver."
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}
