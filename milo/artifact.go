package milo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateArtifactExtension checks if the artifact file has a .jar extension
func ValidateArtifactExtension(artifactPath string) error {
	// Extract filename from path
	filename := filepath.Base(artifactPath)

	// Check if it ends with .jar
	if !strings.HasSuffix(strings.ToLower(filename), ".jar") {
		return fmt.Errorf("Artifact must be a .jar file, got: %s", filename)
	}

	return nil
}

// FindArtifactInTaskDir finds the first .jar file in the task's local directory
func FindArtifactInTaskDir(taskDirPath string) (string, error) {
	localDir := filepath.Join(taskDirPath, "local")

	// Read the local directory to find JAR files
	entries, err := os.ReadDir(localDir)
	if err != nil {
		return "", fmt.Errorf("failed to read local directory: %v", err)
	}

	// Find the first JAR file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(entry.Name()), ".jar") {
			return filepath.Join(localDir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("no JAR file found in task directory")
}

// CheckArtifactExists verifies that the artifact file exists on the filesystem
func CheckArtifactExists(artifactPath string) bool {
	_, err := os.Stat(artifactPath)
	return err == nil
}

// ValidateArtifactExists checks if the artifact file exists and returns appropriate error
func ValidateArtifactExists(artifactPath string) error {
	if !CheckArtifactExists(artifactPath) {
		return fmt.Errorf("Failed to download artifact: file not found")
	}
	return nil
}
