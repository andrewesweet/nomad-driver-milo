package milo

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ValidationError provides structured error information
type ValidationError struct {
	What       string
	Expected   string
	Got        string
	Suggestion string
}

// Error implements the error interface with formatted message
func (e ValidationError) Error() string {
	// Special cases for BDD-compliant messages
	switch e.What {
	case "Artifact must be a JAR file":
		return fmt.Sprintf("Error: Artifact must be a .jar file, got: %s", e.Got)
	case "Failed to download artifact":
		return "Error: Failed to download artifact: file not found"
	default:
		return fmt.Sprintf("Error: %s", e.What)
	}
}

// ValidateArtifactExtensionEnhanced checks if the artifact file has a .jar extension with detailed errors
func ValidateArtifactExtensionEnhanced(artifactPath string) error {
	// Extract filename from path
	filename := filepath.Base(artifactPath)

	// Check if it ends with .jar (case-insensitive)
	if !strings.HasSuffix(strings.ToLower(filename), ".jar") {
		return ValidationError{
			What:       "Artifact must be a JAR file",
			Expected:   "File with .jar extension",
			Got:        filename,
			Suggestion: "Ensure your artifact points to a compiled Java JAR file",
		}
	}

	return nil
}

// ValidateJARIntegrity checks if the file is a valid JAR (ZIP) file
func ValidateJARIntegrity(jarPath string) error {
	// Try to open as ZIP
	reader, err := zip.OpenReader(jarPath)
	if err != nil {
		return ValidationError{
			What:       "Invalid JAR file format",
			Expected:   "Valid Java archive (ZIP format)",
			Got:        "Corrupted or incomplete file",
			Suggestion: "Re-download the artifact or verify the source file",
		}
	}
	defer reader.Close()

	// Check if it's empty
	if len(reader.File) == 0 {
		return ValidationError{
			What:       "JAR file is empty",
			Expected:   "JAR containing Java classes",
			Got:        "0 files in archive",
			Suggestion: "Check the artifact source and build process",
		}
	}

	// Check for JAR structure (at least META-INF directory)
	hasMetaInf := false
	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, "META-INF/") {
			hasMetaInf = true
			break
		}
	}

	if !hasMetaInf {
		return ValidationError{
			What:       "Not a valid JAR file",
			Expected:   "JAR with META-INF directory",
			Got:        "ZIP file without JAR structure",
			Suggestion: "Ensure the artifact is a compiled Java JAR file",
		}
	}

	return nil
}

// ArtifactValidator provides comprehensive artifact validation
type ArtifactValidator struct {
	taskDir string
}

// NewArtifactValidator creates a new validator for the given task directory
func NewArtifactValidator(taskDir string) *ArtifactValidator {
	return &ArtifactValidator{
		taskDir: taskDir,
	}
}

// FindAndValidateArtifact finds JAR files in the task directory and validates them
func (v *ArtifactValidator) FindAndValidateArtifact() (string, error) {
	// Look in the local directory (where Nomad places artifacts)
	localDir := filepath.Join(v.taskDir, "local")

	// Find all JAR files
	jarFiles := []string{}
	err := filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".jar") {
			jarFiles = append(jarFiles, path)
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to search for artifacts: %w", err)
	}

	// Check how many JARs we found
	switch len(jarFiles) {
	case 0:
		return "", ValidationError{
			What:       "Failed to download artifact",
			Expected:   "JAR file in task directory",
			Got:        "No JAR files found in local/",
			Suggestion: "Check that your artifact was downloaded successfully",
		}
	case 1:
		// Validate the single JAR
		jarPath := jarFiles[0]
		if err := ValidateJARIntegrity(jarPath); err != nil {
			return "", err
		}
		return jarPath, nil
	default:
		// Multiple JARs found
		sort.Strings(jarFiles)
		fileNames := make([]string, len(jarFiles))
		for i, path := range jarFiles {
			fileNames[i] = filepath.Base(path)
		}

		return "", ValidationError{
			What:       "Multiple JAR files found",
			Expected:   "Single JAR file to execute",
			Got:        fmt.Sprintf("%d JAR files (%s)", len(jarFiles), strings.Join(fileNames, ", ")),
			Suggestion: "Use a single artifact or specify which JAR to execute",
		}
	}
}
