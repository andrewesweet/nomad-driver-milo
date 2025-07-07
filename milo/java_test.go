package milo

import (
	"os"
	"path/filepath"
	"strings"
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
	// When we format a missing Java error with search paths
	searchPaths := []string{"/usr/lib/jvm/java-17", "/opt/java"}
	err := FormatMissingJavaError(searchPaths)

	// Then we should get a MissingJavaError with simple Error() and detailed Detailed() messages
	errMsg := err.Error()
	if errMsg != "No Java runtime found on host" {
		t.Errorf("expected error to be 'No Java runtime found on host', got %q", errMsg)
	}
	
	// Check that it's a MissingJavaError and verify detailed message
	mjErr, ok := err.(*MissingJavaError)
	if !ok {
		t.Errorf("expected error to be a *MissingJavaError, got %T", err)
	} else {
		detailedMsg := mjErr.Detailed()
		if !strings.Contains(detailedMsg, "Error: No Java runtime found on host") {
			t.Errorf("expected detailed message to contain 'Error: No Java runtime found on host', got %q", detailedMsg)
		}
		if !strings.Contains(detailedMsg, "Searched locations:") {
			t.Errorf("expected detailed message to contain 'Searched locations:', got %q", detailedMsg)
		}
		if !strings.Contains(detailedMsg, "/usr/lib/jvm/java-17") {
			t.Errorf("expected detailed message to contain searched path, got %q", detailedMsg)
		}
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

	// Check that the error message contains the key parts
	errMsg := err.Error()
	if !strings.Contains(errMsg, "No Java runtime found on host") {
		t.Errorf("expected error to contain 'No Java runtime found on host', got %q", errMsg)
	}
}
