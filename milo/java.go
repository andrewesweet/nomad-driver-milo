package milo

import (
	"fmt"
	"os"
	"path/filepath"
)

// MissingJavaError represents an error when Java runtime is not found
type MissingJavaError struct {
	SearchPaths []string
}

// Error returns the simple BDD-compliant error message
func (e *MissingJavaError) Error() string {
	return "No Java runtime found on host"
}

// Detailed returns the full detailed error message with search paths and instructions
func (e *MissingJavaError) Detailed() string {
	msg := "Error: No Java runtime found on host. Please install Java to use Milo driver.\n\n"
	msg += "Searched locations:\n"

	// Check JAVA_HOME first
	if javaHome := os.Getenv("JAVA_HOME"); javaHome != "" {
		msg += fmt.Sprintf("- JAVA_HOME: %s (invalid or not found)\n", javaHome)
	} else {
		msg += "- JAVA_HOME environment variable (not set)\n"
	}

	// List all searched paths
	for _, path := range e.SearchPaths {
		msg += fmt.Sprintf("- %s (not found)\n", path)
	}

	msg += "\nTo fix:\n"
	msg += "1. Install Java: sudo apt install openjdk-17-jdk\n"
	msg += "2. Or set JAVA_HOME to existing installation"

	return msg
}

// ScanJavaInstallationPaths scans common Java installation directories
func ScanJavaInstallationPaths(searchPaths []string) []string {
	var javaInstallations []string

	for _, basePath := range searchPaths {
		// Check if there's a bin/java executable
		javaExe := filepath.Join(basePath, "bin", "java")
		if _, err := os.Stat(javaExe); err == nil {
			javaInstallations = append(javaInstallations, basePath)
		}
	}

	return javaInstallations
}

// CheckJavaHomeEnvironment checks if JAVA_HOME is set and valid
func CheckJavaHomeEnvironment() (string, bool) {
	javaHome := os.Getenv("JAVA_HOME")
	if javaHome == "" {
		return "", false
	}

	// Verify that JAVA_HOME points to a valid Java installation
	javaExe := filepath.Join(javaHome, "bin", "java")
	if _, err := os.Stat(javaExe); err != nil {
		return "", false
	}

	return javaHome, true
}

// ValidateJavaExecutable checks if a directory contains a valid Java executable
func ValidateJavaExecutable(javaDir string) bool {
	javaExe := filepath.Join(javaDir, "bin", "java")
	_, err := os.Stat(javaExe)
	return err == nil
}

// FormatMissingJavaError creates a MissingJavaError with the provided search paths
func FormatMissingJavaError(searchPaths []string) error {
	return &MissingJavaError{
		SearchPaths: searchPaths,
	}
}

// DetectJavaRuntime attempts to find a Java runtime on the system
func DetectJavaRuntime(searchPaths []string) (string, error) {
	// First, check JAVA_HOME
	if javaHome, found := CheckJavaHomeEnvironment(); found {
		return javaHome, nil
	}

	// If JAVA_HOME not set or invalid, scan common installation paths
	javaInstallations := ScanJavaInstallationPaths(searchPaths)
	if len(javaInstallations) > 0 {
		// Return the first valid installation found
		return javaInstallations[0], nil
	}

	// No Java runtime found
	return "", FormatMissingJavaError(searchPaths)
}
