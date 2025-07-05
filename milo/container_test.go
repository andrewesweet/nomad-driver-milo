package milo

import (
	"testing"
)

func TestGenerateCrunCommand(t *testing.T) {
	bundlePath := "/tmp/container-bundle"
	containerID := "milo-task-12345"

	cmd := GenerateCrunCommand(bundlePath, containerID)

	expected := []string{"crun", "run", "--bundle", bundlePath, containerID}

	if len(cmd) != len(expected) {
		t.Fatalf("Expected command length %d, got %d", len(expected), len(cmd))
	}

	for i, arg := range expected {
		if cmd[i] != arg {
			t.Errorf("Expected argument %d to be %q, got %q", i, arg, cmd[i])
		}
	}
}

func TestCreateOCISpec(t *testing.T) {
	javaHome := "/usr/lib/jvm/java-17"
	jarPath := "/app/hello-world.jar"
	taskDir := "/tmp/nomad-task"

	spec, err := CreateOCISpec(javaHome, jarPath, taskDir)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if spec == nil {
		t.Fatal("Expected spec to not be nil")
	}

	// Verify process configuration
	if spec.Process == nil {
		t.Fatal("Expected process to not be nil")
	}

	if len(spec.Process.Args) == 0 {
		t.Fatal("Expected process args to not be empty")
	}

	// Should contain java command
	expectedArgs := []string{"java", "-jar", "/app/hello-world.jar"}
	if len(spec.Process.Args) != len(expectedArgs) {
		t.Fatalf("Expected %d process args, got %d", len(expectedArgs), len(spec.Process.Args))
	}

	for i, arg := range expectedArgs {
		if spec.Process.Args[i] != arg {
			t.Errorf("Expected process arg %d to be %q, got %q", i, arg, spec.Process.Args[i])
		}
	}

	// Verify root filesystem
	if spec.Root == nil {
		t.Fatal("Expected root to not be nil")
	}

	if spec.Root.Path != "rootfs" {
		t.Errorf("Expected root path to be 'rootfs', got %q", spec.Root.Path)
	}

	// Verify mounts include Java runtime
	foundJavaMount := false
	for _, mount := range spec.Mounts {
		if mount.Source == javaHome && mount.Destination == "/usr/lib/jvm/java" {
			foundJavaMount = true
			break
		}
	}

	if !foundJavaMount {
		t.Error("Expected Java runtime mount not found")
	}
	
	// Verify Linux namespace configuration for crun
	if spec.Linux == nil {
		t.Fatal("Expected Linux configuration to be present for crun compatibility")
	}
	
	// Verify required namespaces are present
	expectedNamespaces := []string{"pid", "ipc", "uts", "mount"}
	if len(spec.Linux.Namespaces) != len(expectedNamespaces) {
		t.Errorf("Expected %d namespaces, got %d", len(expectedNamespaces), len(spec.Linux.Namespaces))
	}
	
	// Check each namespace type
	namespaceMap := make(map[string]bool)
	for _, ns := range spec.Linux.Namespaces {
		namespaceMap[string(ns.Type)] = true
	}
	
	for _, expectedNS := range expectedNamespaces {
		if !namespaceMap[expectedNS] {
			t.Errorf("Expected namespace %q not found", expectedNS)
		}
	}
}

func TestCreateContainerBundle(t *testing.T) {
	javaHome := "/usr/lib/jvm/java-17"
	jarPath := "/app/hello-world.jar"
	taskDir := "/tmp/nomad-task"
	bundlePath := "/tmp/test-bundle"

	// Create the OCI spec first
	spec, err := CreateOCISpec(javaHome, jarPath, taskDir)
	if err != nil {
		t.Fatalf("Failed to create OCI spec: %v", err)
	}

	// Test creating the bundle
	err = CreateContainerBundle(bundlePath, spec)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}
