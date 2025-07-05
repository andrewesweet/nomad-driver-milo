package milo

import (
	"fmt"
	"os"
	"path/filepath"
)

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

// FormatMissingJavaError generates a user-friendly error for missing Java
func FormatMissingJavaError() error {
	return fmt.Errorf("No Java runtime found on host. Please install Java to use Milo driver.")
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
	return "", FormatMissingJavaError()
}
