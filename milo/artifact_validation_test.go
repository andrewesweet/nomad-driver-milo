package milo

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// === AUGMENTED CODING TDD/ATDD Mode ACTIVATED ===
// I will maintain code quality and separate structural from behavioral changes.

// === CYCLE 1: Case-insensitive extension check ===
// Phase: RED
// Time: Starting now
// Test from plan.md: Test 1 - Case-insensitive extension check (.jar, .JAR, .Jar)
func TestValidateArtifactExtension_CaseInsensitive(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "lowercase .jar",
			filename: "app.jar",
			wantErr:  false,
		},
		{
			name:     "uppercase .JAR",
			filename: "APP.JAR",
			wantErr:  false,
		},
		{
			name:     "mixed case .Jar",
			filename: "MyApp.Jar",
			wantErr:  false,
		},
		{
			name:     "weird case .jAr",
			filename: "weird.jAr",
			wantErr:  false,
		},
		{
			name:     "non-jar file",
			filename: "script.py",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateArtifactExtension(tc.filename)
			if tc.wantErr {
				require.Error(t, err, "expected error for %s", tc.filename)
			} else {
				require.NoError(t, err, "expected no error for %s", tc.filename)
			}
		})
	}
}

// === CYCLE 2: Detailed error message for non-JAR files ===
// Phase: RED
// Test from plan.md: Test 2 - Detailed error message for non-JAR files
func TestValidateArtifactExtension_DetailedErrorMessage(t *testing.T) {
	// Given a non-JAR file
	filename := "script.py"

	// When we validate the extension
	err := ValidateArtifactExtensionEnhanced(filename)

	// Then we should get a detailed error message
	require.Error(t, err)

	// Check error contains all required parts
	errMsg := err.Error()
	require.Contains(t, errMsg, "Error:")
	require.Contains(t, errMsg, "Expected:")
	require.Contains(t, errMsg, "Got:")
	require.Contains(t, errMsg, "Suggestion:")

	// Check specific content
	require.Contains(t, errMsg, "script.py")
	require.Contains(t, errMsg, ".jar")
}

// === CYCLE 3: Valid JAR passes ZIP structure check ===
// Phase: RED
// Test from plan.md: Test 4 - Valid JAR passes ZIP structure check
func TestValidateJARIntegrity_ValidJAR(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	jarPath := filepath.Join(tmpDir, "valid.jar")

	// Create a valid JAR file
	createValidTestJAR(t, jarPath)

	// When we validate the JAR integrity
	err := ValidateJARIntegrity(jarPath)

	// Then it should pass without error
	require.NoError(t, err)
}

// Helper to create a valid test JAR
func createValidTestJAR(t *testing.T, path string) {
	file, err := os.Create(path)
	require.NoError(t, err)
	defer file.Close()

	writer := zip.NewWriter(file)
	defer writer.Close()

	// Add META-INF directory
	_, err = writer.Create("META-INF/")
	require.NoError(t, err)

	// Add manifest
	manifest, err := writer.Create("META-INF/MANIFEST.MF")
	require.NoError(t, err)
	_, err = manifest.Write([]byte("Manifest-Version: 1.0\nMain-Class: Main\n"))
	require.NoError(t, err)

	// Add a class file
	class, err := writer.Create("Main.class")
	require.NoError(t, err)
	_, err = class.Write([]byte("fake class content"))
	require.NoError(t, err)
}

// === CYCLE 4: Corrupt JAR fails with specific error ===
// Phase: RED
// Test from plan.md: Test 5 - Corrupt JAR fails with specific error
func TestValidateJARIntegrity_CorruptJAR(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	jarPath := filepath.Join(tmpDir, "corrupt.jar")

	// Create a corrupt JAR (truncated ZIP)
	data := []byte{0x50, 0x4B, 0x03, 0x04, 0x14, 0x00, 0x00, 0x00} // ZIP header then truncated
	err := os.WriteFile(jarPath, data, 0600)
	require.NoError(t, err)

	// When we validate the JAR integrity
	err = ValidateJARIntegrity(jarPath)

	// Then it should fail with specific error
	require.Error(t, err)
	errMsg := err.Error()
	require.Contains(t, errMsg, "Invalid JAR file format")
	require.Contains(t, errMsg, "Corrupted or incomplete file")
	require.Contains(t, errMsg, "Re-download")
}

// === CYCLE 5: HTML error page disguised as JAR fails ===
// Phase: RED
// Test from plan.md: Test 6 - HTML error page disguised as JAR fails
func TestValidateJARIntegrity_HTMLAsJAR(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	jarPath := filepath.Join(tmpDir, "error.jar")

	// Create an HTML file disguised as JAR
	html := `<html><head><title>404 Not Found</title></head>
<body><h1>Not Found</h1><p>The requested artifact was not found.</p></body></html>`
	err := os.WriteFile(jarPath, []byte(html), 0600)
	require.NoError(t, err)

	// When we validate the JAR integrity
	err = ValidateJARIntegrity(jarPath)

	// Then it should fail with specific error
	require.Error(t, err)
	errMsg := err.Error()
	require.Contains(t, errMsg, "Invalid JAR file format")
}

// === CYCLE 6: Find and validate single JAR ===
// Phase: RED
// Test from plan.md: Test 8 - Single JAR found successfully
func TestFindAndValidateArtifact_SingleJAR(t *testing.T) {
	// Create a task directory structure
	tmpDir := t.TempDir()
	taskDir := filepath.Join(tmpDir, "task")
	localDir := filepath.Join(taskDir, "local")
	err := os.MkdirAll(localDir, 0755)
	require.NoError(t, err)

	// Create a single valid JAR
	jarPath := filepath.Join(localDir, "app.jar")
	createValidTestJAR(t, jarPath)

	// When we find and validate artifacts
	validator := NewArtifactValidator(taskDir)
	artifactPath, err := validator.FindAndValidateArtifact()

	// Then it should succeed and return the JAR path
	require.NoError(t, err)
	require.Equal(t, jarPath, artifactPath)
}

// === CYCLE 7: Multiple JARs trigger error ===
// Phase: RED
// Test from plan.md: Test 9 - Multiple JARs trigger specific error
func TestFindAndValidateArtifact_MultipleJARs(t *testing.T) {
	// Create a task directory structure
	tmpDir := t.TempDir()
	taskDir := filepath.Join(tmpDir, "task")
	localDir := filepath.Join(taskDir, "local")
	err := os.MkdirAll(localDir, 0755)
	require.NoError(t, err)

	// Create multiple JARs
	createValidTestJAR(t, filepath.Join(localDir, "app.jar"))
	createValidTestJAR(t, filepath.Join(localDir, "lib.jar"))
	createValidTestJAR(t, filepath.Join(localDir, "util.jar"))

	// When we find and validate artifacts
	validator := NewArtifactValidator(taskDir)
	_, err = validator.FindAndValidateArtifact()

	// Then it should fail with specific error
	require.Error(t, err)
	errMsg := err.Error()
	require.Contains(t, errMsg, "Multiple JAR files found")
	require.Contains(t, errMsg, "3 JAR files")
	require.Contains(t, errMsg, "app.jar")
}

// === CYCLE 8: No JARs found triggers error ===
// Phase: RED
// Test from plan.md: Test 10 - No JARs found triggers specific error
func TestFindAndValidateArtifact_NoJARs(t *testing.T) {
	// Create a task directory structure
	tmpDir := t.TempDir()
	taskDir := filepath.Join(tmpDir, "task")
	localDir := filepath.Join(taskDir, "local")
	err := os.MkdirAll(localDir, 0755)
	require.NoError(t, err)

	// Create non-JAR files
	err = os.WriteFile(filepath.Join(localDir, "config.properties"), []byte("config"), 0600)
	require.NoError(t, err)

	// When we find and validate artifacts
	validator := NewArtifactValidator(taskDir)
	_, err = validator.FindAndValidateArtifact()

	// Then it should fail with specific error
	require.Error(t, err)
	errMsg := err.Error()
	require.Contains(t, errMsg, "Artifact file not found")
	require.Contains(t, errMsg, "No JAR files found")
}
