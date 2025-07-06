package milo

import (
	"os"
	"path/filepath"
	"testing"
)

// === CYCLE 1: Validate artifact extension ===
// Phase: RED
func TestValidateArtifactExtension_RejectsNonJar(t *testing.T) {
	// Given a non-JAR file path
	artifactPath := "/tmp/my-script.py"

	// When we validate the artifact extension
	err := ValidateArtifactExtension(artifactPath)

	// Then we should get an error for non-JAR files
	if err == nil {
		t.Error("expected error for non-JAR file, got nil")
	}

	expectedMsg := "Error: Artifact must be a .jar file, got: my-script.py"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}

// === CYCLE 2: Validate JAR extension passes ===
// Phase: RED
func TestValidateArtifactExtension_AcceptsJar(t *testing.T) {
	// Given a JAR file path
	artifactPath := "/tmp/my-app.jar"

	// When we validate the artifact extension
	err := ValidateArtifactExtension(artifactPath)

	// Then we should get no error for JAR files
	if err != nil {
		t.Errorf("expected no error for JAR file, got %v", err)
	}
}

// === CYCLE 3: Find artifact files in task directory ===
// Phase: RED
func TestFindArtifactInTaskDir(t *testing.T) {
	// Given a task directory with a local subdirectory containing a JAR file
	taskDirPath := "/tmp/test-task-find"
	localDir := filepath.Join(taskDirPath, "local")
	jarPath := filepath.Join(localDir, "test-app.jar")

	// Create the directory structure
	err := os.MkdirAll(localDir, 0755)
	if err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}
	defer os.RemoveAll(taskDirPath)

	// Create a test JAR file
	content := "PK\x03\x04" // JAR file magic bytes
	err = os.WriteFile(jarPath, []byte(content), 0600)
	if err != nil {
		t.Fatalf("failed to create test JAR file: %v", err)
	}

	// When we look for artifact files in the task directory
	artifactPath, err := FindArtifactInTaskDir(taskDirPath)

	// Then we should find the JAR file
	if err != nil {
		t.Errorf("expected no error finding artifact, got %v", err)
	}

	if artifactPath != jarPath {
		t.Errorf("expected artifact path %q, got %q", jarPath, artifactPath)
	}
}

// === CYCLE 5: Check if artifact file exists ===
// Phase: RED
func TestCheckArtifactExists_FileExists(t *testing.T) {
	// Given an existing artifact file
	artifactPath := "/tmp/test-app.jar"

	// Create the file for testing
	content := "fake jar content"
	err := os.WriteFile(artifactPath, []byte(content), 0600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer os.Remove(artifactPath)

	// When we check if the artifact exists
	exists := CheckArtifactExists(artifactPath)

	// Then it should return true
	if !exists {
		t.Error("expected artifact to exist, but CheckArtifactExists returned false")
	}
}

// === CYCLE 6: Check if artifact file is missing ===
// Phase: RED
func TestCheckArtifactExists_FileMissing(t *testing.T) {
	// Given a non-existent artifact file
	artifactPath := "/tmp/non-existent.jar"

	// Ensure the file doesn't exist
	os.Remove(artifactPath)

	// When we check if the artifact exists
	exists := CheckArtifactExists(artifactPath)

	// Then it should return false
	if exists {
		t.Error("expected artifact to not exist, but CheckArtifactExists returned true")
	}
}

// === CYCLE 7: Generate missing file error ===
// Phase: RED
func TestValidateArtifactExists_ReturnsError(t *testing.T) {
	// Given a non-existent artifact file
	artifactPath := "/tmp/missing-app.jar"

	// Ensure the file doesn't exist
	os.Remove(artifactPath)

	// When we validate the artifact exists
	err := ValidateArtifactExists(artifactPath)

	// Then we should get an error
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}

	expectedMsg := "Error: Failed to download artifact: file not found"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}
