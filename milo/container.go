package milo

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/opencontainers/runtime-spec/specs-go"
)

// GenerateCrunCommand generates the command arguments for running a container with crun
func GenerateCrunCommand(bundlePath, containerID string) []string {
	return []string{"crun", "run", "--bundle", bundlePath, containerID}
}

// CreateOCISpec creates an OCI runtime specification for executing a JAR file
func CreateOCISpec(javaHome, jarPath, taskDir string) (*specs.Spec, error) {
	spec := &specs.Spec{
		Version: "1.0.0",
		Process: &specs.Process{
			Terminal: false,
			Args:     []string{"java", "-jar", jarPath},
			Env:      []string{"PATH=/usr/bin:/bin", "JAVA_HOME=/usr/lib/jvm/java"},
			Cwd:      "/app",
		},
		Root: &specs.Root{
			Path:     "rootfs",
			Readonly: false,
		},
		Linux: &specs.Linux{
			Namespaces: []specs.LinuxNamespace{
				{Type: "pid"},
				{Type: "ipc"},
				{Type: "uts"},
				{Type: "mount"},
			},
		},
		Mounts: []specs.Mount{
			{
				Destination: "/usr/lib/jvm/java",
				Source:      javaHome,
				Type:        "bind",
				Options:     []string{"bind", "ro"},
			},
			{
				Destination: "/app",
				Source:      taskDir,
				Type:        "bind",
				Options:     []string{"bind"},
			},
		},
	}

	return spec, nil
}

// CreateContainerBundle creates a container bundle directory with config.json and rootfs
func CreateContainerBundle(bundlePath string, spec *specs.Spec) error {
	// Create the bundle directory
	if err := os.MkdirAll(bundlePath, 0755); err != nil {
		return err
	}

	// Create the rootfs directory
	rootfsPath := filepath.Join(bundlePath, "rootfs")
	if err := os.MkdirAll(rootfsPath, 0755); err != nil {
		return err
	}

	// Write the OCI spec as config.json
	configPath := filepath.Join(bundlePath, "config.json")
	configData, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, configData, 0600)
}
