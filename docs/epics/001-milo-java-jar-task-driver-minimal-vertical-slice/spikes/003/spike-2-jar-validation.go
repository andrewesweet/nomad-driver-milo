package main

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
)

// Spike 2: Test different JAR validation methods
func main() {
	fmt.Println("=== Spike 2: JAR File Validation Methods ===")
	fmt.Println()

	// Create test files
	testDir := "/tmp/jar-validation-spike"
	err := os.RemoveAll(testDir)
	if err != nil {
		fmt.Printf("Error removing test dir: %v\n", err)
	}
	err = os.MkdirAll(testDir, 0755)
	if err != nil {
		fmt.Printf("Error creating test dir: %v\n", err)
		return
	}

	// Test 1: Valid JAR file
	validJar := filepath.Join(testDir, "valid.jar")
	createValidJar(validJar)

	// Test 2: Corrupt JAR (truncated)
	corruptJar := filepath.Join(testDir, "corrupt.jar")
	createCorruptJar(corruptJar)

	// Test 3: HTML error page saved as .jar
	htmlJar := filepath.Join(testDir, "error.jar")
	createHTMLAsJar(htmlJar)

	// Test 4: Empty file
	emptyJar := filepath.Join(testDir, "empty.jar")
	err = os.WriteFile(emptyJar, []byte{}, 0600)
	if err != nil {
		fmt.Printf("Error creating empty jar: %v\n", err)
	}

	// Test validation methods
	files := []string{validJar, corruptJar, htmlJar, emptyJar}

	for _, file := range files {
		fmt.Printf("\nValidating: %s\n", filepath.Base(file))
		fmt.Printf("  Magic bytes check: %v\n", checkMagicBytes(file))
		fmt.Printf("  ZIP structure check: %v\n", checkZipStructure(file))
		fmt.Printf("  Manifest check: %v\n", checkManifest(file))
	}

	// Summary
	fmt.Println("\n=== Validation Method Recommendations ===")
	fmt.Println("1. Magic bytes (PK\\x03\\x04) - Fast but not sufficient alone")
	fmt.Println("2. ZIP structure validation - Most reliable, catches corrupt files")
	fmt.Println("3. Manifest check - Good for ensuring it's a JAR, not just ZIP")
	fmt.Println("4. Recommended: Combine ZIP structure + check for META-INF/")

	// Cleanup
	err = os.RemoveAll(testDir)
	if err != nil {
		fmt.Printf("Error cleaning up: %v\n", err)
	}
}

// Check if file starts with ZIP magic bytes
func checkMagicBytes(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}
	defer file.Close()

	magic := make([]byte, 4)
	_, err = file.Read(magic)
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	// ZIP files start with PK\x03\x04 (0x504B0304)
	if len(magic) >= 4 && magic[0] == 0x50 && magic[1] == 0x4B && magic[2] == 0x03 && magic[3] == 0x04 {
		return "PASS - Has ZIP magic bytes"
	}
	return fmt.Sprintf("FAIL - Wrong magic bytes: %x", magic)
}

// Validate ZIP structure
func checkZipStructure(path string) string {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Sprintf("FAIL - Invalid ZIP: %v", err)
	}
	defer reader.Close()

	fileCount := len(reader.File)
	if fileCount == 0 {
		return "FAIL - ZIP has no files"
	}

	return fmt.Sprintf("PASS - Valid ZIP with %d files", fileCount)
}

// Check for JAR manifest
func checkManifest(path string) string {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return "FAIL - Cannot read as ZIP"
	}
	defer reader.Close()

	hasManifest := false
	hasMetaInf := false

	for _, file := range reader.File {
		if file.Name == "META-INF/" {
			hasMetaInf = true
		}
		if file.Name == "META-INF/MANIFEST.MF" {
			hasManifest = true
		}
	}

	if hasManifest {
		return "PASS - Has JAR manifest"
	} else if hasMetaInf {
		return "WARN - Has META-INF/ but no manifest"
	}
	return "FAIL - No META-INF directory"
}

// Helper functions to create test files
func createValidJar(path string) {
	writer, err := os.Create(path)
	if err != nil {
		fmt.Printf("Error creating jar file: %v\n", err)
		return
	}
	defer writer.Close()

	zipWriter := zip.NewWriter(writer)
	defer zipWriter.Close()

	// Add META-INF/
	_, err = zipWriter.Create("META-INF/")
	if err != nil {
		fmt.Printf("Error creating META-INF: %v\n", err)
		return
	}

	// Add manifest
	manifest, err := zipWriter.Create("META-INF/MANIFEST.MF")
	if err != nil {
		fmt.Printf("Error creating manifest: %v\n", err)
		return
	}
	_, err = manifest.Write([]byte("Manifest-Version: 1.0\n"))
	if err != nil {
		fmt.Printf("Error writing manifest: %v\n", err)
	}

	// Add a class file
	class, err := zipWriter.Create("Main.class")
	if err != nil {
		fmt.Printf("Error creating class: %v\n", err)
		return
	}
	_, err = class.Write([]byte("fake class content"))
	if err != nil {
		fmt.Printf("Error writing class: %v\n", err)
	}
}

func createCorruptJar(path string) {
	// Start with valid ZIP header, then truncate
	data := []byte{0x50, 0x4B, 0x03, 0x04, 0x14, 0x00, 0x00, 0x00}
	err := os.WriteFile(path, data, 0600)
	if err != nil {
		fmt.Printf("Error creating corrupt jar: %v\n", err)
	}
}

func createHTMLAsJar(path string) {
	html := `<html><head><title>404 Not Found</title></head>
<body><h1>Not Found</h1></body></html>`
	err := os.WriteFile(path, []byte(html), 0600)
	if err != nil {
		fmt.Printf("Error creating HTML jar: %v\n", err)
	}
}
